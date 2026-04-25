package service

import (
	"context"
	"errors"
	"strings"
)

type OpenAIWSIngressSelectionOutcome string

const (
	OpenAIWSIngressSelectionSucceeded     OpenAIWSIngressSelectionOutcome = "succeeded"
	OpenAIWSIngressSelectionError         OpenAIWSIngressSelectionOutcome = "selection_error"
	OpenAIWSIngressSelectionNoAvailable   OpenAIWSIngressSelectionOutcome = "no_available_account"
	OpenAIWSIngressSelectionAccountBusy   OpenAIWSIngressSelectionOutcome = "account_busy"
	OpenAIWSIngressSelectionAcquireError  OpenAIWSIngressSelectionOutcome = "account_acquire_error"
	OpenAIWSIngressSelectionNotConfigured OpenAIWSIngressSelectionOutcome = "not_configured"
)

type OpenAIWSIngressAccountSlotAcquireFunc func(context.Context, int64, int) (func(), bool, error)

type OpenAIWSIngressSelectionHooks struct {
	ResolveSelectionFallback  func(context.Context, string, error) string
	OnAccountSelectFailed     func(error)
	OnFallbackSelected        func(string)
	OnStickySessionBindFailed func(*Account, error)
}

type OpenAIWSIngressSelectionRequest struct {
	APIKey             *APIKey
	PreviousResponseID string
	SessionHash        string
	SchedulingModel    string
	SchedulerContext   context.Context
	CodexProfile       *CodexRequestProfile
	RequiredTransport  OpenAIUpstreamTransport
	AcquireAccountSlot OpenAIWSIngressAccountSlotAcquireFunc
	Hooks              OpenAIWSIngressSelectionHooks
}

type OpenAIWSIngressSelectionResult struct {
	Outcome               OpenAIWSIngressSelectionOutcome
	Account               *Account
	ScheduleDecision      OpenAIAccountScheduleDecision
	Err                   error
	InitialSelectionErr   error
	InitialSelectionModel string
	SchedulingModel       string
	AccountMaxConcurrency int
	AccountRelease        func()
}

type OpenAIWSIngressSelectionKernel struct {
	gatewayService *OpenAIGatewayService
}

func NewOpenAIWSIngressSelectionKernel(gatewayService *OpenAIGatewayService) *OpenAIWSIngressSelectionKernel {
	return &OpenAIWSIngressSelectionKernel{gatewayService: gatewayService}
}

func (k *OpenAIWSIngressSelectionKernel) Select(ctx context.Context, req OpenAIWSIngressSelectionRequest) OpenAIWSIngressSelectionResult {
	if ctx == nil {
		ctx = context.Background()
	}
	result := OpenAIWSIngressSelectionResult{
		Outcome:               OpenAIWSIngressSelectionNotConfigured,
		InitialSelectionModel: strings.TrimSpace(req.SchedulingModel),
		SchedulingModel:       strings.TrimSpace(req.SchedulingModel),
	}
	if k == nil || k.gatewayService == nil {
		result.Err = errors.New("openai websocket ingress selection kernel is not configured")
		return result
	}
	if req.SchedulerContext == nil {
		req.SchedulerContext = ctx
	}
	if req.CodexProfile != nil {
		req.SchedulerContext = WithOpenAICodexTransportPreference(req.SchedulerContext, req.CodexProfile.NativeClient)
	}

	selection, decision, err := k.gatewayService.SelectAccountWithScheduler(
		req.SchedulerContext,
		req.apiKeyGroupID(),
		req.PreviousResponseID,
		req.SessionHash,
		result.SchedulingModel,
		nil,
		req.requiredTransport(),
	)
	if err != nil {
		if req.Hooks.OnAccountSelectFailed != nil {
			req.Hooks.OnAccountSelectFailed(err)
		}
		selected, fallbackDecision, fallbackModel, fallbackErr, ok := k.trySelectionFallback(ctx, req, result.SchedulingModel, err)
		if !ok {
			result.Outcome = OpenAIWSIngressSelectionError
			result.Err = err
			result.InitialSelectionErr = err
			return result
		}
		if fallbackErr != nil {
			result.Outcome = OpenAIWSIngressSelectionError
			result.Err = fallbackErr
			result.InitialSelectionErr = err
			return result
		}
		selection = selected
		decision = fallbackDecision
		result.SchedulingModel = fallbackModel
	}

	if selection == nil || selection.Account == nil {
		result.Outcome = OpenAIWSIngressSelectionNoAvailable
		return result
	}

	account := selection.Account
	result.Account = account
	result.ScheduleDecision = decision
	result.AccountMaxConcurrency = account.Concurrency
	if selection.WaitPlan != nil && selection.WaitPlan.MaxConcurrency > 0 {
		result.AccountMaxConcurrency = selection.WaitPlan.MaxConcurrency
	}

	release, outcome, acquireErr := k.acquireInitialAccountSlot(ctx, req, selection)
	if outcome != OpenAIWSIngressSelectionSucceeded {
		result.Outcome = outcome
		result.Err = acquireErr
		return result
	}
	result.AccountRelease = release

	if err := k.gatewayService.BindStickySession(ctx, req.apiKeyGroupID(), req.SessionHash, account.ID); err != nil {
		if req.Hooks.OnStickySessionBindFailed != nil {
			req.Hooks.OnStickySessionBindFailed(account, err)
		}
	}
	if req.CodexProfile != nil {
		ObserveOpenAICodexSchedulingDecision(*req.CodexProfile, decision)
	}

	result.Outcome = OpenAIWSIngressSelectionSucceeded
	return result
}

func (k *OpenAIWSIngressSelectionKernel) trySelectionFallback(
	ctx context.Context,
	req OpenAIWSIngressSelectionRequest,
	selectionModel string,
	initialErr error,
) (*AccountSelectionResult, OpenAIAccountScheduleDecision, string, error, bool) {
	if strings.TrimSpace(selectionModel) == "" || req.Hooks.ResolveSelectionFallback == nil {
		return nil, OpenAIAccountScheduleDecision{}, "", nil, false
	}
	fallbackModel := strings.TrimSpace(req.Hooks.ResolveSelectionFallback(ctx, selectionModel, initialErr))
	if fallbackModel == "" || fallbackModel == selectionModel {
		return nil, OpenAIAccountScheduleDecision{}, "", nil, false
	}
	if req.Hooks.OnFallbackSelected != nil {
		req.Hooks.OnFallbackSelected(fallbackModel)
	}
	selection, decision, err := k.gatewayService.SelectAccountWithScheduler(
		req.SchedulerContext,
		req.apiKeyGroupID(),
		req.PreviousResponseID,
		req.SessionHash,
		fallbackModel,
		nil,
		req.requiredTransport(),
	)
	return selection, decision, fallbackModel, err, true
}

func (k *OpenAIWSIngressSelectionKernel) acquireInitialAccountSlot(
	ctx context.Context,
	req OpenAIWSIngressSelectionRequest,
	selection *AccountSelectionResult,
) (func(), OpenAIWSIngressSelectionOutcome, error) {
	if selection == nil {
		return nil, OpenAIWSIngressSelectionNoAvailable, nil
	}
	if selection.Acquired {
		return selection.ReleaseFunc, OpenAIWSIngressSelectionSucceeded, nil
	}
	if selection.WaitPlan == nil {
		return nil, OpenAIWSIngressSelectionAccountBusy, nil
	}
	if req.AcquireAccountSlot == nil {
		return nil, OpenAIWSIngressSelectionAcquireError, errors.New("openai websocket ingress account slot acquire hook is not configured")
	}
	release, acquired, err := req.AcquireAccountSlot(ctx, selection.WaitPlan.AccountID, selection.WaitPlan.MaxConcurrency)
	if err != nil {
		return nil, OpenAIWSIngressSelectionAcquireError, err
	}
	if !acquired {
		return nil, OpenAIWSIngressSelectionAccountBusy, nil
	}
	return release, OpenAIWSIngressSelectionSucceeded, nil
}

func (req OpenAIWSIngressSelectionRequest) apiKeyGroupID() *int64 {
	if req.APIKey == nil {
		return nil
	}
	return req.APIKey.GroupID
}

func (req OpenAIWSIngressSelectionRequest) requiredTransport() OpenAIUpstreamTransport {
	if req.RequiredTransport == "" {
		return OpenAIUpstreamTransportResponsesWebsocketV2
	}
	return req.RequiredTransport
}
