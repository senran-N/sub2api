package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type compatibleTextKernelSchedulerReport struct {
	accountID    int64
	success      bool
	firstTokenMs *int
}

type compatibleTextKernelStubScheduler struct {
	selectFunc func(context.Context, OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error)

	selectCalls []OpenAIAccountScheduleRequest
	reports     []compatibleTextKernelSchedulerReport
	switches    int
}

func (s *compatibleTextKernelStubScheduler) Select(ctx context.Context, req OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
	s.selectCalls = append(s.selectCalls, req)
	if s.selectFunc == nil {
		return nil, OpenAIAccountScheduleDecision{}, errors.New("select not configured")
	}
	return s.selectFunc(ctx, req)
}

func (s *compatibleTextKernelStubScheduler) ReportResult(accountID int64, success bool, firstTokenMs *int) {
	s.reports = append(s.reports, compatibleTextKernelSchedulerReport{
		accountID:    accountID,
		success:      success,
		firstTokenMs: firstTokenMs,
	})
}

func (s *compatibleTextKernelStubScheduler) ReportSwitch() {
	s.switches++
}

func (s *compatibleTextKernelStubScheduler) SnapshotMetrics() OpenAIAccountSchedulerMetricsSnapshot {
	return OpenAIAccountSchedulerMetricsSnapshot{AccountSwitchTotal: int64(s.switches)}
}

func TestCompatibleTextExecutionKernelUsesSelectionFallback(t *testing.T) {
	groupID := int64(7)
	initialErr := errors.New("no account for requested model")
	account := &Account{ID: 101, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	scheduler := &compatibleTextKernelStubScheduler{}
	scheduler.selectFunc = func(_ context.Context, req OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		switch req.RequestedModel {
		case "requested-model":
			return nil, OpenAIAccountScheduleDecision{}, initialErr
		case "fallback-model":
			return &AccountSelectionResult{Account: account}, OpenAIAccountScheduleDecision{SelectedAccountID: account.ID}, nil
		default:
			return nil, OpenAIAccountScheduleDecision{}, errors.New("unexpected model")
		}
	}
	kernel := NewCompatibleTextExecutionKernel(&OpenAIGatewayService{openaiScheduler: scheduler})

	fallbackSelected := ""
	forwardDefaultMappedModel := ""
	released := false
	result := kernel.Execute(context.Background(), CompatibleTextExecutionRequest{
		ProtocolFamily:       CompatibleGatewayProtocolFamilyResponses,
		APIKey:               &APIKey{GroupID: &groupID},
		SchedulingModel:      "requested-model",
		UseSelectionFallback: true,
		AcquireAccountSlot: func(_ context.Context, _ string, selection *AccountSelectionResult) (func(), bool) {
			require.Same(t, account, selection.Account)
			return func() { released = true }, true
		},
		Forward: func(_ context.Context, selected *Account, defaultMappedModel string) (*OpenAIForwardResult, error) {
			require.Same(t, account, selected)
			forwardDefaultMappedModel = defaultMappedModel
			firstTokenMs := 42
			return &OpenAIForwardResult{FirstTokenMs: &firstTokenMs}, nil
		},
		Hooks: CompatibleTextExecutionHooks{
			ResolveSelectionFallback: func(_ context.Context, selectionModel string, err error) string {
				require.Equal(t, "requested-model", selectionModel)
				require.ErrorIs(t, err, initialErr)
				return "fallback-model"
			},
			OnFallbackSelected: func(model string) {
				fallbackSelected = model
			},
		},
	})

	require.Equal(t, CompatibleTextExecutionSucceeded, result.Outcome)
	require.Same(t, account, result.Account)
	require.True(t, released)
	require.Equal(t, "fallback-model", fallbackSelected)
	require.Equal(t, "fallback-model", forwardDefaultMappedModel)
	require.Len(t, scheduler.selectCalls, 2)
	require.Equal(t, "requested-model", scheduler.selectCalls[0].RequestedModel)
	require.Equal(t, "fallback-model", scheduler.selectCalls[1].RequestedModel)
	require.Len(t, scheduler.reports, 1)
	require.Equal(t, account.ID, scheduler.reports[0].accountID)
	require.True(t, scheduler.reports[0].success)
	require.NotNil(t, scheduler.reports[0].firstTokenMs)
	require.Equal(t, 42, *scheduler.reports[0].firstTokenMs)
}

func TestCompatibleTextExecutionKernelClearsFallbackModelBetweenFailoverAttempts(t *testing.T) {
	initialErr := errors.New("no account for requested model")
	fallbackAccount := &Account{ID: 201, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	recoveredAccount := &Account{ID: 202, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	requestedSelectCount := 0
	scheduler := &compatibleTextKernelStubScheduler{}
	scheduler.selectFunc = func(_ context.Context, req OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		switch req.RequestedModel {
		case "requested-model":
			requestedSelectCount++
			if requestedSelectCount == 1 {
				return nil, OpenAIAccountScheduleDecision{}, initialErr
			}
			require.Contains(t, req.ExcludedIDs, fallbackAccount.ID)
			return &AccountSelectionResult{Account: recoveredAccount}, OpenAIAccountScheduleDecision{SelectedAccountID: recoveredAccount.ID}, nil
		case "fallback-model":
			return &AccountSelectionResult{Account: fallbackAccount}, OpenAIAccountScheduleDecision{SelectedAccountID: fallbackAccount.ID}, nil
		default:
			return nil, OpenAIAccountScheduleDecision{}, errors.New("unexpected model")
		}
	}
	kernel := NewCompatibleTextExecutionKernel(&OpenAIGatewayService{openaiScheduler: scheduler})

	var forwardedAccountIDs []int64
	var forwardedDefaultModels []string
	result := kernel.Execute(context.Background(), CompatibleTextExecutionRequest{
		ProtocolFamily:       CompatibleGatewayProtocolFamilyResponses,
		SchedulingModel:      "requested-model",
		UseSelectionFallback: true,
		AcquireAccountSlot: func(_ context.Context, _ string, selection *AccountSelectionResult) (func(), bool) {
			require.NotNil(t, selection)
			require.NotNil(t, selection.Account)
			return func() {}, true
		},
		Forward: func(_ context.Context, account *Account, defaultMappedModel string) (*OpenAIForwardResult, error) {
			forwardedAccountIDs = append(forwardedAccountIDs, account.ID)
			forwardedDefaultModels = append(forwardedDefaultModels, defaultMappedModel)
			if account.ID == fallbackAccount.ID {
				return nil, &UpstreamFailoverError{StatusCode: http.StatusInternalServerError}
			}
			return &OpenAIForwardResult{}, nil
		},
		Hooks: CompatibleTextExecutionHooks{
			ResolveSelectionFallback: func(_ context.Context, selectionModel string, err error) string {
				require.Equal(t, "requested-model", selectionModel)
				require.ErrorIs(t, err, initialErr)
				return "fallback-model"
			},
		},
	})

	require.Equal(t, CompatibleTextExecutionSucceeded, result.Outcome)
	require.Same(t, recoveredAccount, result.Account)
	require.Equal(t, []int64{fallbackAccount.ID, recoveredAccount.ID}, forwardedAccountIDs)
	require.Equal(t, []string{"fallback-model", ""}, forwardedDefaultModels)
	require.Len(t, scheduler.selectCalls, 3)
	require.Equal(t, "requested-model", scheduler.selectCalls[0].RequestedModel)
	require.Equal(t, "fallback-model", scheduler.selectCalls[1].RequestedModel)
	require.Equal(t, "requested-model", scheduler.selectCalls[2].RequestedModel)
	require.Equal(t, 1, scheduler.switches)
	require.Len(t, scheduler.reports, 2)
	require.Equal(t, compatibleTextKernelSchedulerReport{accountID: fallbackAccount.ID, success: false}, scheduler.reports[0])
	require.Equal(t, compatibleTextKernelSchedulerReport{accountID: recoveredAccount.ID, success: true}, scheduler.reports[1])
	require.Contains(t, result.FailedAccountIDs, fallbackAccount.ID)
}

func TestCompatiblePassthroughExecutionKernelRejectsIncompatibleAccountBeforeAcquire(t *testing.T) {
	incompatibleAccount := &Account{ID: 301, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	compatibleAccount := &Account{ID: 302, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	scheduler := &compatibleTextKernelStubScheduler{}
	scheduler.selectFunc = func(_ context.Context, req OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		if _, excluded := req.ExcludedIDs[incompatibleAccount.ID]; excluded {
			return &AccountSelectionResult{Account: compatibleAccount}, OpenAIAccountScheduleDecision{SelectedAccountID: compatibleAccount.ID}, nil
		}
		return &AccountSelectionResult{Account: incompatibleAccount}, OpenAIAccountScheduleDecision{SelectedAccountID: incompatibleAccount.ID}, nil
	}
	kernel := NewCompatiblePassthroughExecutionKernel(&OpenAIGatewayService{openaiScheduler: scheduler})

	var rejectedAccountID int64
	acquireCalls := 0
	forwardCalls := 0
	result := kernel.Execute(context.Background(), CompatiblePassthroughExecutionRequest{
		SchedulingModel: "grok-4",
		ValidateAccount: func(account *Account) bool {
			return account.ID == compatibleAccount.ID
		},
		AcquireAccountSlot: func(_ context.Context, _ string, selection *AccountSelectionResult) (func(), bool) {
			acquireCalls++
			require.Same(t, compatibleAccount, selection.Account)
			return func() {}, true
		},
		Forward: func(_ context.Context, account *Account, defaultMappedModel string) (*OpenAIForwardResult, error) {
			forwardCalls++
			require.Same(t, compatibleAccount, account)
			require.Empty(t, defaultMappedModel)
			return &OpenAIForwardResult{}, nil
		},
		Hooks: CompatiblePassthroughExecutionHooks{
			OnAccountRejected: func(account *Account) {
				rejectedAccountID = account.ID
			},
		},
	})

	require.Equal(t, CompatibleTextExecutionSucceeded, result.Outcome)
	require.Same(t, compatibleAccount, result.Account)
	require.Equal(t, incompatibleAccount.ID, rejectedAccountID)
	require.Contains(t, result.FailedAccountIDs, incompatibleAccount.ID)
	require.Equal(t, 1, acquireCalls)
	require.Equal(t, 1, forwardCalls)
	require.Len(t, scheduler.selectCalls, 2)
}
