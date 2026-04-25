package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimeFailoverStateHandleForwardErrorIgnoresNonFailover(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)

	result := state.HandleForwardError(context.Background(), &mockTempUnscheduler{}, 10, PlatformAnthropic, errors.New("plain failure"), false)

	require.False(t, result.Handled)
	require.Nil(t, result.FailoverErr)
	require.Empty(t, state.FailedAccountIDs)
	require.Nil(t, state.LastFailoverErr)
}

func TestRuntimeFailoverStateHandleForwardErrorExhaustsAfterResponseStarted(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)
	failoverErr := newTestFailoverErr(http.StatusBadGateway, false, false)

	result := state.HandleForwardError(context.Background(), &mockTempUnscheduler{}, 10, PlatformAnthropic, failoverErr, true)

	require.True(t, result.Handled)
	require.Equal(t, RuntimeFailoverExhausted, result.Action)
	require.Equal(t, failoverErr, result.FailoverErr)
	require.True(t, result.ResponseStarted)
	require.Empty(t, state.FailedAccountIDs)
	require.Nil(t, state.LastFailoverErr)
}

func TestRuntimeFailoverStateHandleForwardErrorDelegatesToFailoverState(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)
	failoverErr := newTestFailoverErr(http.StatusBadGateway, false, false)

	result := state.HandleForwardError(context.Background(), &mockTempUnscheduler{}, 11, PlatformAnthropic, failoverErr, false)

	require.True(t, result.Handled)
	require.Equal(t, RuntimeFailoverContinue, result.Action)
	require.Equal(t, failoverErr, result.FailoverErr)
	require.False(t, result.ResponseStarted)
	require.Equal(t, 1, state.SwitchCount)
	require.Contains(t, state.FailedAccountIDs, int64(11))
	require.Equal(t, failoverErr, state.LastFailoverErr)
}

func TestRuntimeFailoverStateHandleForwardErrorReturnsLastFailoverErrOnExhausted(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)
	failoverErr := newTestFailoverErr(http.StatusUnauthorized, false, false)

	result := state.HandleForwardError(context.Background(), &mockTempUnscheduler{}, 12, PlatformAnthropic, failoverErr, false)

	require.True(t, result.Handled)
	require.Equal(t, RuntimeFailoverExhausted, result.Action)
	require.Equal(t, state.LastFailoverErr, result.FailoverErr)
	require.Equal(t, failoverErr, result.FailoverErr)
	require.Contains(t, state.FailedAccountIDs, int64(12))
}
