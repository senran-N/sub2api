package service

import "context"

// RuntimeSubject is the authenticated caller state consumed by runtime kernels.
// HTTP handlers should derive it from middleware context and then pass the
// value through instead of letting lower stages depend on Gin.
type RuntimeSubject struct {
	UserID      int64
	APIKeyID    int64
	GroupID     *int64
	Concurrency int
	Provider    string
	Protocol    GatewayProtocol
}

// RuntimeSessionState is the service-owned sticky/session state shared across
// selection, account-slot acquisition, admission, and failover decisions.
type RuntimeSessionState struct {
	SessionHash    string
	SessionKey     string
	BoundAccountID int64
	HasBound       bool
}

// RuntimeAdmissionResult describes the provider-neutral admission decision for
// a selected account. Denied admissions are cleaned up by RuntimePipeline using
// the cleanup request returned here.
type RuntimeAdmissionResult struct {
	Outcome               RuntimeAdmissionOutcome
	Account               *Account
	WindowCostReservation *WindowCostReservation
	Cleanup               RuntimeAdmissionCleanupRequest
	Err                   error
}

// RuntimeForwardResult is the normalized result of a single forward attempt.
// Forward attempt cleanup should be performed by ExecuteRuntimeForwardAttempt
// before this value is returned to RuntimePipeline.
type RuntimeForwardResult struct {
	Result          *ForwardResult
	Err             error
	ResponseStarted bool
	Attempt         RuntimeForwardAttemptResult
}

type RuntimePipelineOutcome string

const (
	RuntimePipelineSucceeded                   RuntimePipelineOutcome = "succeeded"
	RuntimePipelineSelectionInitialUnavailable RuntimePipelineOutcome = "selection_initial_unavailable"
	RuntimePipelineSelectionRetryCanceled      RuntimePipelineOutcome = "selection_retry_canceled"
	RuntimePipelineSelectionExhausted          RuntimePipelineOutcome = "selection_exhausted"
	RuntimePipelineNoAvailableAccount          RuntimePipelineOutcome = "no_available_account"
	RuntimePipelineAccountSlotQueueFull        RuntimePipelineOutcome = "account_slot_queue_full"
	RuntimePipelineAccountSlotAcquireError     RuntimePipelineOutcome = "account_slot_acquire_error"
	RuntimePipelineAccountSlotWaitAcquireError RuntimePipelineOutcome = "account_slot_wait_acquire_error"
	RuntimePipelineAccountSlotUnavailable      RuntimePipelineOutcome = "account_slot_unavailable"
	RuntimePipelineForwardError                RuntimePipelineOutcome = "forward_error"
	RuntimePipelineFailoverExhausted           RuntimePipelineOutcome = "failover_exhausted"
	RuntimePipelineFailoverCanceled            RuntimePipelineOutcome = "failover_canceled"
	RuntimePipelineForwardHookAborted          RuntimePipelineOutcome = "forward_hook_aborted"
	RuntimePipelineMisconfigured               RuntimePipelineOutcome = "misconfigured"
)

// ProtocolRenderer is intentionally small: rendering remains an adapter concern,
// while runtime kernels expose protocol-neutral outcomes.
type ProtocolRenderer interface {
	RenderRuntimePipelineOutcome(context.Context, RuntimePipelineResult)
}

type RuntimePipelineSelectFunc func(context.Context, SelectionRequest) (*AccountSelectionResult, SelectionDecision, error)
type RuntimePipelineAccountSlotFunc func(context.Context, *RuntimePipelineState) RuntimeAccountSlotResult
type RuntimePipelineAdmissionFunc func(context.Context, *RuntimePipelineState) RuntimeAdmissionResult
type RuntimePipelineForwardFunc func(context.Context, *RuntimePipelineState) RuntimeForwardResult
type RuntimePipelineHookFunc func(context.Context, *RuntimePipelineState) RuntimePipelineHookResult
type RuntimePipelineSuccessFunc func(context.Context, RuntimePipelineResult)

type RuntimePipelineHookResult struct {
	Abort   bool
	Outcome RuntimePipelineOutcome
	Err     error
}

type RuntimePipelineHooks struct {
	OnSelectionDecision func(context.Context, *RuntimePipelineState, SelectionDecision)
	OnSelectionRetry    func(context.Context, *RuntimePipelineState) context.Context
	OnAccountSelected   RuntimePipelineHookFunc
	BeforeAcquireSlot   RuntimePipelineHookFunc
	BeforeForward       RuntimePipelineHookFunc
	OnSuccess           RuntimePipelineSuccessFunc
}

type RuntimePipelineRequest struct {
	Subject         RuntimeSubject
	Session         RuntimeSessionState
	Selection       SelectionRequest
	MaxSwitches     int
	TempUnscheduler TempUnscheduler
	Select          RuntimePipelineSelectFunc
	AcquireSlot     RuntimePipelineAccountSlotFunc
	Admit           RuntimePipelineAdmissionFunc
	Forward         RuntimePipelineForwardFunc
	Hooks           RuntimePipelineHooks
}

type RuntimePipelineState struct {
	Subject               RuntimeSubject
	Session               RuntimeSessionState
	SelectionRequest      SelectionRequest
	Selection             *AccountSelectionResult
	SelectionDecision     SelectionDecision
	Account               *Account
	Slot                  RuntimeAccountSlotResult
	AccountRelease        RuntimeForwardReleaseFunc
	Admission             RuntimeAdmissionResult
	WindowCostReservation *WindowCostReservation
	Forward               RuntimeForwardResult
	FailoverState         *RuntimeFailoverState
}

type RuntimePipelineResult struct {
	Outcome           RuntimePipelineOutcome
	Subject           RuntimeSubject
	Session           RuntimeSessionState
	Account           *Account
	ForwardResult     *ForwardResult
	Err               error
	SelectionFailure  RuntimeSelectionFailureResult
	Slot              RuntimeAccountSlotResult
	Admission         RuntimeAdmissionResult
	Forward           RuntimeForwardResult
	Failover          RuntimeForwardFailoverResult
	FailedAccountIDs  map[int64]struct{}
	SwitchCount       int
	ForceCacheBilling bool
}
