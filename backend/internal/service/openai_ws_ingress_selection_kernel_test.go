package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIWSIngressSelectionKernelUsesFallbackAndAcquiresWaitPlan(t *testing.T) {
	initialErr := errors.New("no websocket account for requested model")
	account := &Account{ID: 401, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 3}
	scheduler := &compatibleTextKernelStubScheduler{}
	scheduler.selectFunc = func(_ context.Context, req OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		require.Equal(t, OpenAIUpstreamTransportResponsesWebsocketV2, req.RequiredTransport)
		switch req.RequestedModel {
		case "requested-model":
			return nil, OpenAIAccountScheduleDecision{}, initialErr
		case "fallback-model":
			return &AccountSelectionResult{
				Account: account,
				WaitPlan: &AccountWaitPlan{
					AccountID:      account.ID,
					MaxConcurrency: 5,
				},
			}, OpenAIAccountScheduleDecision{Layer: "load_balance", SelectedAccountID: account.ID}, nil
		default:
			return nil, OpenAIAccountScheduleDecision{}, errors.New("unexpected model")
		}
	}
	kernel := NewOpenAIWSIngressSelectionKernel(&OpenAIGatewayService{openaiScheduler: scheduler})

	fallbackSelected := ""
	released := false
	result := kernel.Select(context.Background(), OpenAIWSIngressSelectionRequest{
		SchedulingModel: "requested-model",
		SessionHash:     "session-hash",
		AcquireAccountSlot: func(_ context.Context, accountID int64, maxConcurrency int) (func(), bool, error) {
			require.Equal(t, account.ID, accountID)
			require.Equal(t, 5, maxConcurrency)
			return func() { released = true }, true, nil
		},
		Hooks: OpenAIWSIngressSelectionHooks{
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

	require.Equal(t, OpenAIWSIngressSelectionSucceeded, result.Outcome)
	require.Same(t, account, result.Account)
	require.Equal(t, "requested-model", result.InitialSelectionModel)
	require.Equal(t, "fallback-model", result.SchedulingModel)
	require.Equal(t, 5, result.AccountMaxConcurrency)
	require.Equal(t, "fallback-model", fallbackSelected)
	require.NotNil(t, result.AccountRelease)
	result.AccountRelease()
	require.True(t, released)
	require.Len(t, scheduler.selectCalls, 2)
	require.Equal(t, "requested-model", scheduler.selectCalls[0].RequestedModel)
	require.Equal(t, "fallback-model", scheduler.selectCalls[1].RequestedModel)
}

func TestOpenAIWSIngressSelectionKernelReportsBusyWhenNoWaitPlan(t *testing.T) {
	account := &Account{ID: 402, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 3}
	scheduler := &compatibleTextKernelStubScheduler{}
	scheduler.selectFunc = func(_ context.Context, req OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		return &AccountSelectionResult{Account: account}, OpenAIAccountScheduleDecision{SelectedAccountID: account.ID}, nil
	}
	kernel := NewOpenAIWSIngressSelectionKernel(&OpenAIGatewayService{openaiScheduler: scheduler})

	result := kernel.Select(context.Background(), OpenAIWSIngressSelectionRequest{
		SchedulingModel: "requested-model",
		SessionHash:     "session-hash",
	})

	require.Equal(t, OpenAIWSIngressSelectionAccountBusy, result.Outcome)
	require.Same(t, account, result.Account)
	require.Nil(t, result.AccountRelease)
}
