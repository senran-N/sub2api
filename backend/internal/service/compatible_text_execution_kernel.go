package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const OpenAICompatibleSameAccountRetryDelay = 500 * time.Millisecond

type CompatibleTextExecutionOutcome string

const (
	CompatibleTextExecutionSucceeded         CompatibleTextExecutionOutcome = "succeeded"
	CompatibleTextExecutionSelectionError    CompatibleTextExecutionOutcome = "selection_error"
	CompatibleTextExecutionNoAvailable       CompatibleTextExecutionOutcome = "no_available_account"
	CompatibleTextExecutionAcquireBlocked    CompatibleTextExecutionOutcome = "acquire_blocked"
	CompatibleTextExecutionForwardError      CompatibleTextExecutionOutcome = "forward_error"
	CompatibleTextExecutionFailoverExhausted CompatibleTextExecutionOutcome = "failover_exhausted"
	CompatibleTextExecutionCanceled          CompatibleTextExecutionOutcome = "canceled"
)

type CompatibleTextExecutionForwardFunc func(context.Context, *Account, string) (*OpenAIForwardResult, error)
type CompatibleTextExecutionAcquireFunc func(context.Context, string, *AccountSelectionResult) (func(), bool)
type CompatibleTextExecutionAccountValidator func(*Account) bool

type CompatibleTextExecutionAttemptFeedback struct {
	Account          *Account
	ForwardModelHint string
	FailoverErr      *UpstreamFailoverError
}

type CompatibleTextExecutionHooks struct {
	ResetFallbackModel           func()
	ResolveSelectionFallback     func(context.Context, string, error) string
	OnFallbackSelected           func(string)
	OnAccountSelecting           func(string, int)
	OnAccountSelectFailed        func(error, int)
	OnPreviousResponseSelected   func(*Account)
	OnScheduleDecision           func(OpenAIAccountScheduleDecision)
	OnAccountSelected            func(*Account)
	OnAccountRejected            func(*Account)
	OnRoutingLatency             func(time.Duration)
	OnForwardLatency             func(time.Duration, *OpenAIForwardResult)
	OnFailoverAttemptFailed      func(CompatibleTextExecutionAttemptFeedback)
	ResolveCodexFailoverRecovery func(*Account, *UpstreamFailoverError, bool) CodexRecoveryDecision
	RecordCodexRecoverySwitch    func(*Account, *UpstreamFailoverError, bool)
	OnSameAccountRetry           func(*Account, *UpstreamFailoverError, OpenAIPoolFailoverDecision)
	OnFailoverSwitch             func(*Account, *UpstreamFailoverError, OpenAIPoolFailoverDecision)
	OnCompleted                  func(*Account, int)
}

type CompatibleTextExecutionRequest struct {
	ProtocolFamily            CompatibleGatewayProtocolFamily
	Provider                  string
	APIKey                    *APIKey
	Body                      []byte
	ReqModel                  string
	ReqStream                 bool
	SchedulingModel           string
	PreferredMappedModel      string
	PreviousResponseID        string
	SessionHash               string
	ChannelUsage              ChannelUsageFields
	RoutingStart              time.Time
	SchedulerContext          context.Context
	CodexProfile              *CodexRequestProfile
	UseSelectionFallback      bool
	IncludeRequestPayloadHash bool
	MaxAccountSwitches        int
	RequiredTransport         OpenAIUpstreamTransport
	ValidateAccount           CompatibleTextExecutionAccountValidator
	Forward                   CompatibleTextExecutionForwardFunc
	AcquireAccountSlot        CompatibleTextExecutionAcquireFunc
	Hooks                     CompatibleTextExecutionHooks
}

type CompatibleTextExecutionResult struct {
	Outcome             CompatibleTextExecutionOutcome
	Account             *Account
	ForwardResult       *OpenAIForwardResult
	Err                 error
	InitialSelectionErr error
	LastFailoverErr     *UpstreamFailoverError
	FailoverErr         *UpstreamFailoverError
	FailedAccountIDs    map[int64]struct{}
	SwitchCount         int
	SessionHash         string
	ForwardModelHint    string
}

type CompatibleTextExecutionKernel struct {
	gatewayService *OpenAIGatewayService
}

func NewCompatibleTextExecutionKernel(gatewayService *OpenAIGatewayService) *CompatibleTextExecutionKernel {
	return &CompatibleTextExecutionKernel{gatewayService: gatewayService}
}

func (k *CompatibleTextExecutionKernel) Execute(ctx context.Context, req CompatibleTextExecutionRequest) CompatibleTextExecutionResult {
	if ctx == nil {
		ctx = context.Background()
	}
	result := CompatibleTextExecutionResult{
		Outcome:          CompatibleTextExecutionForwardError,
		FailedAccountIDs: make(map[int64]struct{}),
		SessionHash:      req.SessionHash,
	}
	if k == nil || k.gatewayService == nil || req.Forward == nil || req.AcquireAccountSlot == nil {
		result.Err = errors.New("compatible text execution kernel is not configured")
		return result
	}
	if req.SchedulerContext == nil {
		req.SchedulerContext = ctx
	}
	if req.Provider == "" {
		req.Provider = PlatformOpenAI
	}
	if req.MaxAccountSwitches <= 0 {
		req.MaxAccountSwitches = 3
	}

	switchCount := 0
	sameAccountRetryCount := make(map[int64]int)
	var lastFailoverErr *UpstreamFailoverError
	fallbackModel := ""

	for {
		fallbackModel = ""
		if req.Hooks.ResetFallbackModel != nil {
			req.Hooks.ResetFallbackModel()
		}

		selectionModel := req.selectionModel()
		if req.Hooks.OnAccountSelecting != nil {
			req.Hooks.OnAccountSelecting(selectionModel, len(result.FailedAccountIDs))
		}

		selection, scheduleDecision, err := k.gatewayService.SelectAccountWithScheduler(
			req.SchedulerContext,
			req.apiKeyGroupID(),
			req.PreviousResponseID,
			result.SessionHash,
			selectionModel,
			result.FailedAccountIDs,
			req.requiredTransport(),
		)
		if err != nil {
			if req.Hooks.OnAccountSelectFailed != nil {
				req.Hooks.OnAccountSelectFailed(err, len(result.FailedAccountIDs))
			}
			selected, decision, selectedFallback, fallbackErr, ok := k.trySelectionFallback(ctx, req, selectionModel, result.FailedAccountIDs, err)
			if !ok {
				result.Outcome = CompatibleTextExecutionSelectionError
				result.Err = err
				result.InitialSelectionErr = err
				result.LastFailoverErr = lastFailoverErr
				return result
			}
			if fallbackErr != nil {
				result.Outcome = CompatibleTextExecutionSelectionError
				result.Err = fallbackErr
				result.InitialSelectionErr = err
				result.LastFailoverErr = lastFailoverErr
				return result
			}
			selection = selected
			scheduleDecision = decision
			fallbackModel = selectedFallback
		}

		if selection == nil || selection.Account == nil {
			result.Outcome = CompatibleTextExecutionNoAvailable
			result.LastFailoverErr = lastFailoverErr
			return result
		}

		account := selection.Account
		if req.ProtocolFamily == CompatibleGatewayProtocolFamilyResponses && strings.TrimSpace(req.PreviousResponseID) != "" && req.Hooks.OnPreviousResponseSelected != nil {
			req.Hooks.OnPreviousResponseSelected(account)
		}
		if req.Hooks.OnScheduleDecision != nil {
			req.Hooks.OnScheduleDecision(scheduleDecision)
		}

		if req.ValidateAccount != nil && !req.ValidateAccount(account) {
			result.FailedAccountIDs[account.ID] = struct{}{}
			if req.Hooks.OnAccountRejected != nil {
				req.Hooks.OnAccountRejected(account)
			}
			continue
		}

		result.SessionHash = EnsureOpenAIPoolModeSessionHash(result.SessionHash, account)
		req.SessionHash = result.SessionHash
		if req.Hooks.OnAccountSelected != nil {
			req.Hooks.OnAccountSelected(account)
		}

		release, acquired := req.AcquireAccountSlot(ctx, result.SessionHash, selection)
		if !acquired {
			result.Outcome = CompatibleTextExecutionAcquireBlocked
			result.Account = account
			return result
		}

		if req.Hooks.OnRoutingLatency != nil && !req.RoutingStart.IsZero() {
			req.Hooks.OnRoutingLatency(time.Since(req.RoutingStart))
		}
		forwardStart := time.Now()
		defaultMappedModel := req.forwardDefaultMappedModel(fallbackModel)
		forwardModelHint := req.forwardModelHint(defaultMappedModel, selectionModel)
		forwardResult, forwardErr := req.Forward(ctx, account, defaultMappedModel)
		forwardDuration := time.Since(forwardStart)
		if release != nil {
			release()
		}
		if req.Hooks.OnForwardLatency != nil {
			req.Hooks.OnForwardLatency(forwardDuration, forwardResult)
		}

		if forwardErr != nil {
			var failoverErr *UpstreamFailoverError
			if errors.As(forwardErr, &failoverErr) {
				k.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
				if req.Hooks.OnFailoverAttemptFailed != nil {
					req.Hooks.OnFailoverAttemptFailed(CompatibleTextExecutionAttemptFeedback{
						Account:          account,
						ForwardModelHint: forwardModelHint,
						FailoverErr:      failoverErr,
					})
				}
				lastFailoverErr = failoverErr
				decision := k.applyFailoverPolicy(ctx, req, account, failoverErr, sameAccountRetryCount, result.FailedAccountIDs, &switchCount)
				if decision.SameAccountRetry {
					if req.Hooks.OnSameAccountRetry != nil {
						req.Hooks.OnSameAccountRetry(account, failoverErr, decision)
					}
					if err := sleepWithContext(ctx, OpenAICompatibleSameAccountRetryDelay); err != nil {
						result.Outcome = CompatibleTextExecutionCanceled
						result.Account = account
						result.Err = err
						result.LastFailoverErr = lastFailoverErr
						result.FailoverErr = failoverErr
						result.SwitchCount = switchCount
						result.ForwardModelHint = forwardModelHint
						return result
					}
					continue
				}
				if decision.Action == RuntimeFailoverExhausted {
					result.Outcome = CompatibleTextExecutionFailoverExhausted
					result.Account = account
					result.Err = forwardErr
					result.LastFailoverErr = lastFailoverErr
					result.FailoverErr = failoverErr
					result.SwitchCount = switchCount
					result.ForwardModelHint = forwardModelHint
					return result
				}
				if req.Hooks.OnFailoverSwitch != nil {
					req.Hooks.OnFailoverSwitch(account, failoverErr, decision)
				}
				continue
			}

			k.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			result.Outcome = CompatibleTextExecutionForwardError
			result.Account = account
			result.Err = forwardErr
			result.SwitchCount = switchCount
			result.ForwardModelHint = forwardModelHint
			return result
		}

		k.reportSuccess(ctx, req, account, forwardResult)
		if req.Hooks.OnCompleted != nil {
			req.Hooks.OnCompleted(account, switchCount)
		}
		result.Outcome = CompatibleTextExecutionSucceeded
		result.Account = account
		result.ForwardResult = forwardResult
		result.SwitchCount = switchCount
		result.ForwardModelHint = forwardModelHint
		return result
	}
}

func (k *CompatibleTextExecutionKernel) trySelectionFallback(
	ctx context.Context,
	req CompatibleTextExecutionRequest,
	selectionModel string,
	failedAccountIDs map[int64]struct{},
	initialErr error,
) (*AccountSelectionResult, OpenAIAccountScheduleDecision, string, error, bool) {
	if !req.UseSelectionFallback || len(failedAccountIDs) != 0 || req.Hooks.ResolveSelectionFallback == nil {
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
		failedAccountIDs,
		req.requiredTransport(),
	)
	return selection, decision, fallbackModel, err, true
}

func (k *CompatibleTextExecutionKernel) applyFailoverPolicy(
	ctx context.Context,
	req CompatibleTextExecutionRequest,
	account *Account,
	failoverErr *UpstreamFailoverError,
	sameAccountRetryCount map[int64]int,
	failedAccountIDs map[int64]struct{},
	switchCount *int,
) OpenAIPoolFailoverDecision {
	codexFailoverDecision := CodexRecoveryDecision{}
	stickyContext := req.SessionHash != ""
	if req.ProtocolFamily == CompatibleGatewayProtocolFamilyResponses {
		stickyContext = stickyContext || strings.TrimSpace(req.PreviousResponseID) != ""
		if req.CodexProfile != nil && req.CodexProfile.NativeClient && req.Hooks.ResolveCodexFailoverRecovery != nil {
			codexFailoverDecision = req.Hooks.ResolveCodexFailoverRecovery(account, failoverErr, stickyContext)
		}
	}
	return ApplyOpenAIPoolFailoverPolicy(
		account,
		failoverErr,
		codexFailoverDecision,
		stickyContext,
		sameAccountRetryCount,
		failedAccountIDs,
		switchCount,
		req.MaxAccountSwitches,
		func() {
			k.gatewayService.TempUnscheduleRetryableError(ctx, account.ID, failoverErr)
		},
		func() {
			k.gatewayService.RecordOpenAIAccountSwitch()
			if req.ProtocolFamily == CompatibleGatewayProtocolFamilyResponses {
				if req.CodexProfile != nil && req.CodexProfile.NativeClient && codexFailoverDecision.SwitchAccount && req.Hooks.RecordCodexRecoverySwitch != nil {
					req.Hooks.RecordCodexRecoverySwitch(account, failoverErr, true)
				}
				return
			}
			if req.Hooks.RecordCodexRecoverySwitch != nil {
				req.Hooks.RecordCodexRecoverySwitch(account, failoverErr, false)
			}
		},
	)
}

func (k *CompatibleTextExecutionKernel) reportSuccess(ctx context.Context, req CompatibleTextExecutionRequest, account *Account, result *OpenAIForwardResult) {
	if result != nil {
		if req.ProtocolFamily == CompatibleGatewayProtocolFamilyResponses && account.Type == AccountTypeOAuth {
			k.gatewayService.UpdateCodexUsageSnapshotFromHeaders(ctx, account.ID, result.ResponseHeaders)
		}
		k.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, result.FirstTokenMs)
		return
	}
	k.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)
}

func (req CompatibleTextExecutionRequest) selectionModel() string {
	if preferred := strings.TrimSpace(req.PreferredMappedModel); preferred != "" {
		return preferred
	}
	return req.SchedulingModel
}

func (req CompatibleTextExecutionRequest) forwardDefaultMappedModel(fallbackModel string) string {
	if req.ProtocolFamily == CompatibleGatewayProtocolFamilyMessages {
		return strings.TrimSpace(req.PreferredMappedModel)
	}
	if fallbackModel == "" && req.ChannelUsage.ChannelMappedModel != "" {
		fallbackModel = req.ChannelUsage.ChannelMappedModel
	}
	return ResolveOpenAIForwardDefaultMappedModel(req.APIKey, fallbackModel)
}

func (req CompatibleTextExecutionRequest) forwardModelHint(defaultMappedModel, selectionModel string) string {
	if model := strings.TrimSpace(defaultMappedModel); model != "" {
		return model
	}
	return strings.TrimSpace(selectionModel)
}

func (req CompatibleTextExecutionRequest) apiKeyGroupID() *int64 {
	if req.APIKey == nil {
		return nil
	}
	return req.APIKey.GroupID
}

func (req CompatibleTextExecutionRequest) requiredTransport() OpenAIUpstreamTransport {
	if req.RequiredTransport == "" {
		return OpenAIUpstreamTransportAny
	}
	return req.RequiredTransport
}

func EnsureOpenAIPoolModeSessionHash(sessionHash string, account *Account) string {
	if sessionHash != "" || account == nil || !account.IsPoolMode() {
		return sessionHash
	}
	return "openai-pool-retry-" + uuid.NewString()
}

type CompatiblePassthroughExecutionRequest = CompatibleTextExecutionRequest
type CompatiblePassthroughExecutionResult = CompatibleTextExecutionResult
type CompatiblePassthroughExecutionHooks = CompatibleTextExecutionHooks

type CompatiblePassthroughExecutionKernel struct {
	textKernel *CompatibleTextExecutionKernel
}

func NewCompatiblePassthroughExecutionKernel(gatewayService *OpenAIGatewayService) *CompatiblePassthroughExecutionKernel {
	return &CompatiblePassthroughExecutionKernel{
		textKernel: NewCompatibleTextExecutionKernel(gatewayService),
	}
}

func (k *CompatiblePassthroughExecutionKernel) Execute(ctx context.Context, req CompatiblePassthroughExecutionRequest) CompatiblePassthroughExecutionResult {
	if k == nil || k.textKernel == nil {
		return CompatiblePassthroughExecutionResult{
			Outcome:          CompatibleTextExecutionForwardError,
			Err:              errors.New("compatible passthrough execution kernel is not configured"),
			FailedAccountIDs: make(map[int64]struct{}),
			SessionHash:      req.SessionHash,
		}
	}
	return k.textKernel.Execute(ctx, req)
}
