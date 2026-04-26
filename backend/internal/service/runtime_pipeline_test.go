package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimePipelineSuccessOrdersCoreStages(t *testing.T) {
	var order []string
	released := false
	account := &Account{ID: 11, Platform: PlatformAnthropic}

	result := NewRuntimePipeline().Execute(context.Background(), RuntimePipelineRequest{
		Session: RuntimeSessionState{SessionHash: "session-1"},
		Selection: SelectionRequest{
			Provider:    PlatformAnthropic,
			Protocol:    GatewayProtocolMessages,
			Model:       "claude-test",
			SessionHash: "session-1",
		},
		MaxSwitches: 1,
		Select: func(_ context.Context, req SelectionRequest) (*AccountSelectionResult, SelectionDecision, error) {
			order = append(order, "selection")
			require.Equal(t, "session-1", req.SessionHash)
			require.Empty(t, req.ExcludedIDs)
			return &AccountSelectionResult{
					Account:     account,
					Acquired:    true,
					ReleaseFunc: func() { released = true },
				},
				SelectionDecision{Provider: req.Provider, Protocol: req.Protocol, Model: req.Model, AccountID: account.ID},
				nil
		},
		AcquireSlot: func(_ context.Context, state *RuntimePipelineState) RuntimeAccountSlotResult {
			order = append(order, "slot")
			require.Equal(t, account, state.Account)
			return RuntimeAccountSlotResult{
				Outcome:     RuntimeAccountSlotSucceeded,
				Account:     state.Account,
				ReleaseFunc: state.Selection.ReleaseFunc,
			}
		},
		Admit: func(_ context.Context, state *RuntimePipelineState) RuntimeAdmissionResult {
			order = append(order, "admission")
			return RuntimeAdmissionResult{Outcome: RuntimeAdmissionSucceeded, Account: state.Account}
		},
		Forward: func(ctx context.Context, state *RuntimePipelineState) RuntimeForwardResult {
			order = append(order, "forward")
			attempt := ExecuteRuntimeForwardAttempt(ctx, RuntimeForwardAttemptRequest{
				Account:        state.Account,
				Forward:        func(context.Context) (*ForwardResult, error) { return &ForwardResult{}, nil },
				AccountRelease: state.AccountRelease,
			})
			return RuntimeForwardResult{
				Result:          attempt.Result,
				Err:             attempt.Err,
				ResponseStarted: attempt.ResponseStarted,
				Attempt:         attempt,
			}
		},
	})

	require.Equal(t, RuntimePipelineSucceeded, result.Outcome)
	require.Equal(t, account, result.Account)
	require.True(t, released)
	require.Equal(t, []string{"selection", "slot", "admission", "forward"}, order)
}

func TestRuntimePipelineAdmissionDeniedCleansAndReselects(t *testing.T) {
	firstReleased := false
	accounts := []*Account{
		{ID: 1, Platform: PlatformAnthropic},
		{ID: 2, Platform: PlatformAnthropic},
	}
	selectCalls := 0

	result := NewRuntimePipeline().Execute(context.Background(), RuntimePipelineRequest{
		Selection:   SelectionRequest{Provider: PlatformAnthropic, Protocol: GatewayProtocolMessages, Model: "claude-test"},
		MaxSwitches: 2,
		Select: func(_ context.Context, req SelectionRequest) (*AccountSelectionResult, SelectionDecision, error) {
			if selectCalls == 1 {
				_, failed := req.ExcludedIDs[accounts[0].ID]
				require.True(t, failed)
			}
			account := accounts[selectCalls]
			selectCalls++
			return &AccountSelectionResult{Account: account, Acquired: true}, SelectionDecision{AccountID: account.ID}, nil
		},
		AcquireSlot: func(_ context.Context, state *RuntimePipelineState) RuntimeAccountSlotResult {
			release := func() {}
			if state.Account.ID == accounts[0].ID {
				release = func() { firstReleased = true }
			}
			return RuntimeAccountSlotResult{Outcome: RuntimeAccountSlotSucceeded, Account: state.Account, ReleaseFunc: release}
		},
		Admit: func(_ context.Context, state *RuntimePipelineState) RuntimeAdmissionResult {
			if state.Account.ID == accounts[0].ID {
				return RuntimeAdmissionResult{
					Outcome: RuntimeAdmissionRPMDenied,
					Account: state.Account,
					Cleanup: RuntimeAdmissionCleanupRequest{
						AccountRelease: state.AccountRelease,
					},
				}
			}
			return RuntimeAdmissionResult{Outcome: RuntimeAdmissionSucceeded, Account: state.Account}
		},
		Forward: func(ctx context.Context, state *RuntimePipelineState) RuntimeForwardResult {
			attempt := ExecuteRuntimeForwardAttempt(ctx, RuntimeForwardAttemptRequest{
				Account:        state.Account,
				Forward:        func(context.Context) (*ForwardResult, error) { return &ForwardResult{}, nil },
				AccountRelease: state.AccountRelease,
			})
			return RuntimeForwardResult{Result: attempt.Result, Err: attempt.Err, Attempt: attempt}
		},
	})

	require.Equal(t, RuntimePipelineSucceeded, result.Outcome)
	require.Equal(t, accounts[1], result.Account)
	require.True(t, firstReleased)
	require.Equal(t, 2, selectCalls)
	_, failed := result.FailedAccountIDs[accounts[0].ID]
	require.True(t, failed)
}

func TestRuntimePipelineResponseStartedFailoverExhaustsWithoutReselect(t *testing.T) {
	selectCalls := 0
	account := &Account{ID: 33, Platform: PlatformGemini}
	upstreamErr := &UpstreamFailoverError{StatusCode: http.StatusServiceUnavailable}

	result := NewRuntimePipeline().Execute(context.Background(), RuntimePipelineRequest{
		Selection:   SelectionRequest{Provider: PlatformGemini, Protocol: GatewayProtocolPassthrough, Model: "gemini-test"},
		MaxSwitches: 3,
		Select: func(context.Context, SelectionRequest) (*AccountSelectionResult, SelectionDecision, error) {
			selectCalls++
			return &AccountSelectionResult{Account: account, Acquired: true}, SelectionDecision{AccountID: account.ID}, nil
		},
		AcquireSlot: func(_ context.Context, state *RuntimePipelineState) RuntimeAccountSlotResult {
			return RuntimeAccountSlotResult{Outcome: RuntimeAccountSlotSucceeded, Account: state.Account}
		},
		Forward: func(context.Context, *RuntimePipelineState) RuntimeForwardResult {
			return RuntimeForwardResult{Err: upstreamErr, ResponseStarted: true}
		},
	})

	require.Equal(t, RuntimePipelineFailoverExhausted, result.Outcome)
	require.True(t, errors.Is(result.Err, upstreamErr))
	require.Equal(t, upstreamErr, result.Failover.FailoverErr)
	require.Equal(t, 1, selectCalls)
	require.Empty(t, result.FailedAccountIDs)
}
