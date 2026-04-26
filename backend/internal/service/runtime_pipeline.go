package service

import (
	"context"
	"errors"
)

type RuntimePipeline struct{}

func NewRuntimePipeline() *RuntimePipeline {
	return &RuntimePipeline{}
}

func (p *RuntimePipeline) Execute(ctx context.Context, req RuntimePipelineRequest) RuntimePipelineResult {
	if ctx == nil {
		ctx = context.Background()
	}
	if p == nil || req.Select == nil || req.AcquireSlot == nil || req.Forward == nil {
		return RuntimePipelineResult{
			Outcome:          RuntimePipelineMisconfigured,
			Subject:          req.Subject,
			Session:          req.Session,
			Err:              errors.New("runtime pipeline is not configured"),
			FailedAccountIDs: map[int64]struct{}{},
		}
	}

	tempUnscheduler := req.TempUnscheduler
	if tempUnscheduler == nil {
		tempUnscheduler = noopRuntimeTempUnscheduler{}
	}
	failoverState := NewRuntimeFailoverState(req.MaxSwitches, req.Session.HasBound)
	state := RuntimePipelineState{
		Subject:          req.Subject,
		Session:          req.Session,
		SelectionRequest: req.Selection,
		FailoverState:    failoverState,
	}

	for {
		selectionRequest := state.SelectionRequest
		selectionRequest.ExcludedIDs = failoverState.FailedAccountIDs
		selection, decision, err := req.Select(ctx, selectionRequest)
		if err != nil {
			selectionFailure := failoverState.HandleSelectionError(ctx, err)
			switch selectionFailure.Outcome {
			case RuntimeSelectionFailureInitialUnavailable:
				return p.result(req, state, RuntimePipelineSelectionInitialUnavailable, err, selectionFailure, RuntimeForwardFailoverResult{})
			case RuntimeSelectionFailureRetry:
				if req.Hooks.OnSelectionRetry != nil {
					if nextCtx := req.Hooks.OnSelectionRetry(ctx, &state); nextCtx != nil {
						ctx = nextCtx
					}
				}
				continue
			case RuntimeSelectionFailureCanceled:
				return p.result(req, state, RuntimePipelineSelectionRetryCanceled, err, selectionFailure, RuntimeForwardFailoverResult{})
			default:
				return p.result(req, state, RuntimePipelineSelectionExhausted, err, selectionFailure, RuntimeForwardFailoverResult{})
			}
		}
		if selection == nil || selection.Account == nil {
			return p.result(req, state, RuntimePipelineNoAvailableAccount, nil, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		}

		state.Selection = selection
		state.SelectionDecision = decision
		state.Account = selection.Account
		if req.Hooks.OnSelectionDecision != nil {
			req.Hooks.OnSelectionDecision(ctx, &state, decision)
		}
		if aborted := runRuntimePipelineHook(ctx, &state, req.Hooks.OnAccountSelected, RuntimePipelineForwardHookAborted); aborted.Abort {
			return p.result(req, state, aborted.Outcome, aborted.Err, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		}
		if aborted := runRuntimePipelineHook(ctx, &state, req.Hooks.BeforeAcquireSlot, RuntimePipelineForwardHookAborted); aborted.Abort {
			return p.result(req, state, aborted.Outcome, aborted.Err, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		}

		slot := req.AcquireSlot(ctx, &state)
		state.Slot = slot
		state.AccountRelease = slot.ReleaseFunc
		switch slot.Outcome {
		case RuntimeAccountSlotSucceeded:
		case RuntimeAccountSlotQueueFull:
			return p.result(req, state, RuntimePipelineAccountSlotQueueFull, slot.Err, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		case RuntimeAccountSlotAcquireError:
			return p.result(req, state, RuntimePipelineAccountSlotAcquireError, slot.Err, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		case RuntimeAccountSlotWaitAcquireError:
			return p.result(req, state, RuntimePipelineAccountSlotWaitAcquireError, slot.Err, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		default:
			return p.result(req, state, RuntimePipelineAccountSlotUnavailable, slot.Err, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		}

		admission := RuntimeAdmissionResult{
			Outcome: RuntimeAdmissionSucceeded,
			Account: state.Account,
		}
		if req.Admit != nil {
			admission = req.Admit(ctx, &state)
			if admission.Outcome == "" {
				admission.Outcome = RuntimeAdmissionSucceeded
			}
			if admission.Account == nil {
				admission.Account = state.Account
			}
		}
		state.Admission = admission
		state.WindowCostReservation = admission.WindowCostReservation
		if admission.Outcome != RuntimeAdmissionSucceeded {
			cleanup := admission.Cleanup
			cleanup.Account = firstRuntimeCleanupAccount(cleanup.Account, state.Account)
			cleanup.FailedAccountIDs = firstRuntimeCleanupFailedIDs(cleanup.FailedAccountIDs, failoverState.FailedAccountIDs)
			if cleanup.AccountRelease == nil {
				cleanup.AccountRelease = state.AccountRelease
			}
			CleanupRuntimeAdmissionDenied(cleanup)
			continue
		}

		if aborted := runRuntimePipelineHook(ctx, &state, req.Hooks.BeforeForward, RuntimePipelineForwardHookAborted); aborted.Abort {
			if state.AccountRelease != nil {
				state.AccountRelease()
			}
			return p.result(req, state, aborted.Outcome, aborted.Err, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
		}

		forward := req.Forward(ctx, &state)
		state.Forward = forward
		if forward.Err == nil {
			result := p.result(req, state, RuntimePipelineSucceeded, nil, RuntimeSelectionFailureResult{}, RuntimeForwardFailoverResult{})
			if req.Hooks.OnSuccess != nil {
				req.Hooks.OnSuccess(ctx, result)
			}
			return result
		}

		failover := failoverState.HandleForwardError(
			ctx,
			tempUnscheduler,
			state.Account.ID,
			state.Account.Platform,
			forward.Err,
			forward.ResponseStarted,
		)
		if failover.Handled {
			switch failover.Action {
			case RuntimeFailoverContinue:
				continue
			case RuntimeFailoverCanceled:
				return p.result(req, state, RuntimePipelineFailoverCanceled, forward.Err, RuntimeSelectionFailureResult{}, failover)
			default:
				return p.result(req, state, RuntimePipelineFailoverExhausted, forward.Err, RuntimeSelectionFailureResult{}, failover)
			}
		}
		return p.result(req, state, RuntimePipelineForwardError, forward.Err, RuntimeSelectionFailureResult{}, failover)
	}
}

type noopRuntimeTempUnscheduler struct{}

func (noopRuntimeTempUnscheduler) TempUnscheduleRetryableError(context.Context, int64, *UpstreamFailoverError) {
}

func (p *RuntimePipeline) result(
	req RuntimePipelineRequest,
	state RuntimePipelineState,
	outcome RuntimePipelineOutcome,
	err error,
	selectionFailure RuntimeSelectionFailureResult,
	failover RuntimeForwardFailoverResult,
) RuntimePipelineResult {
	return RuntimePipelineResult{
		Outcome:           outcome,
		Subject:           req.Subject,
		Session:           state.Session,
		Account:           state.Account,
		ForwardResult:     state.Forward.Result,
		Err:               err,
		SelectionFailure:  selectionFailure,
		Slot:              state.Slot,
		Admission:         state.Admission,
		Forward:           state.Forward,
		Failover:          failover,
		FailedAccountIDs:  copyRuntimeFailedAccountIDs(state.FailoverState),
		SwitchCount:       runtimePipelineSwitchCount(state.FailoverState),
		ForceCacheBilling: runtimePipelineForceCacheBilling(state.FailoverState),
	}
}

func runRuntimePipelineHook(
	ctx context.Context,
	state *RuntimePipelineState,
	hook RuntimePipelineHookFunc,
	defaultOutcome RuntimePipelineOutcome,
) RuntimePipelineHookResult {
	if hook == nil {
		return RuntimePipelineHookResult{}
	}
	result := hook(ctx, state)
	if result.Abort && result.Outcome == "" {
		result.Outcome = defaultOutcome
	}
	return result
}

func firstRuntimeCleanupAccount(primary, fallback *Account) *Account {
	if primary != nil {
		return primary
	}
	return fallback
}

func firstRuntimeCleanupFailedIDs(primary, fallback map[int64]struct{}) map[int64]struct{} {
	if primary != nil {
		return primary
	}
	return fallback
}

func copyRuntimeFailedAccountIDs(state *RuntimeFailoverState) map[int64]struct{} {
	copied := make(map[int64]struct{})
	if state == nil {
		return copied
	}
	for id := range state.FailedAccountIDs {
		copied[id] = struct{}{}
	}
	return copied
}

func runtimePipelineSwitchCount(state *RuntimeFailoverState) int {
	if state == nil {
		return 0
	}
	return state.SwitchCount
}

func runtimePipelineForceCacheBilling(state *RuntimeFailoverState) bool {
	return state != nil && state.ForceCacheBilling
}
