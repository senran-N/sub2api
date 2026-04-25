package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteRuntimeForwardAttemptSuccessCleansUp(t *testing.T) {
	accountReleaseCalls := 0
	queueReleaseCalls := 0
	clearCalls := 0
	windowReleaseCalls := 0
	forwardResult := &ForwardResult{RequestID: "req_success"}
	reservation := &WindowCostReservation{AccountID: 10}

	result := ExecuteRuntimeForwardAttempt(context.Background(), RuntimeForwardAttemptRequest{
		Account: &Account{ID: 10},
		Forward: func(context.Context) (*ForwardResult, error) {
			return forwardResult, nil
		},
		WriterSize: func() int {
			return 100
		},
		AccountRelease: func() {
			accountReleaseCalls++
		},
		QueueRelease: func() {
			queueReleaseCalls++
		},
		ClearUpstreamAccepted: func() {
			clearCalls++
		},
		WindowCostReservation: reservation,
		ReleaseWindowCost: func(context.Context, *WindowCostReservation) error {
			windowReleaseCalls++
			return nil
		},
	})

	require.NoError(t, result.Err)
	require.Equal(t, forwardResult, result.Result)
	require.Equal(t, int64(10), result.Account.ID)
	require.Equal(t, 100, result.WriterSizeBefore)
	require.Equal(t, 100, result.WriterSizeAfter)
	require.False(t, result.ResponseStarted)
	require.NoError(t, result.WindowCostReleaseErr)
	require.Equal(t, 1, accountReleaseCalls)
	require.Equal(t, 1, queueReleaseCalls)
	require.Equal(t, 1, clearCalls)
	require.Zero(t, windowReleaseCalls)
}

func TestExecuteRuntimeForwardAttemptErrorBeforeResponseReleasesWindowCost(t *testing.T) {
	forwardErr := errors.New("upstream failed")
	releaseErr := errors.New("release failed")
	reservation := &WindowCostReservation{AccountID: 11}
	releasedReservation := (*WindowCostReservation)(nil)

	result := ExecuteRuntimeForwardAttempt(context.Background(), RuntimeForwardAttemptRequest{
		Forward: func(context.Context) (*ForwardResult, error) {
			return nil, forwardErr
		},
		WriterSize: func() int {
			return 0
		},
		WindowCostReservation: reservation,
		ReleaseWindowCost: func(_ context.Context, got *WindowCostReservation) error {
			releasedReservation = got
			return releaseErr
		},
	})

	require.ErrorIs(t, result.Err, forwardErr)
	require.False(t, result.ResponseStarted)
	require.Equal(t, reservation, releasedReservation)
	require.ErrorIs(t, result.WindowCostReleaseErr, releaseErr)
}

func TestExecuteRuntimeForwardAttemptErrorAfterResponseStartedKeepsWindowCost(t *testing.T) {
	forwardErr := errors.New("stream failed")
	writerSize := 0
	windowReleaseCalls := 0

	result := ExecuteRuntimeForwardAttempt(context.Background(), RuntimeForwardAttemptRequest{
		Forward: func(context.Context) (*ForwardResult, error) {
			writerSize = 32
			return nil, forwardErr
		},
		WriterSize: func() int {
			return writerSize
		},
		WindowCostReservation: &WindowCostReservation{AccountID: 12},
		ReleaseWindowCost: func(context.Context, *WindowCostReservation) error {
			windowReleaseCalls++
			return nil
		},
	})

	require.ErrorIs(t, result.Err, forwardErr)
	require.Equal(t, 0, result.WriterSizeBefore)
	require.Equal(t, 32, result.WriterSizeAfter)
	require.True(t, result.ResponseStarted)
	require.NoError(t, result.WindowCostReleaseErr)
	require.Zero(t, windowReleaseCalls)
}

func TestExecuteRuntimeForwardAttemptNilSafeCleanupAndMissingForward(t *testing.T) {
	result := ExecuteRuntimeForwardAttempt(context.Background(), RuntimeForwardAttemptRequest{})

	require.ErrorContains(t, result.Err, "runtime forward hook is not configured")
	require.Equal(t, 0, result.WriterSizeBefore)
	require.Equal(t, 0, result.WriterSizeAfter)
	require.False(t, result.ResponseStarted)
}
