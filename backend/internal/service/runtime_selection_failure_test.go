package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimeFailoverStateHandleSelectionErrorIgnoresNil(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)

	result := state.HandleSelectionError(context.Background(), nil)

	require.False(t, result.Handled)
	require.Empty(t, result.Outcome)
	require.Nil(t, result.Err)
}

func TestRuntimeFailoverStateHandleSelectionErrorInitialUnavailable(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)
	selectionErr := errors.New("no candidates")

	result := state.HandleSelectionError(context.Background(), selectionErr)

	require.True(t, result.Handled)
	require.Equal(t, RuntimeSelectionFailureInitialUnavailable, result.Outcome)
	require.Equal(t, RuntimeFailoverExhausted, result.Action)
	require.ErrorIs(t, result.Err, selectionErr)
	require.Nil(t, result.FailoverErr)
}

func TestRuntimeFailoverStateHandleSelectionErrorExhausted(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)
	state.FailedAccountIDs[11] = struct{}{}
	selectionErr := errors.New("excluded all candidates")

	result := state.HandleSelectionError(context.Background(), selectionErr)

	require.True(t, result.Handled)
	require.Equal(t, RuntimeSelectionFailureExhausted, result.Outcome)
	require.Equal(t, RuntimeFailoverExhausted, result.Action)
	require.ErrorIs(t, result.Err, selectionErr)
	require.Nil(t, result.FailoverErr)
}

func TestRuntimeFailoverStateHandleSelectionErrorCanceled(t *testing.T) {
	state := NewRuntimeFailoverState(2, false)
	state.FailedAccountIDs[12] = struct{}{}
	state.LastFailoverErr = &UpstreamFailoverError{StatusCode: http.StatusServiceUnavailable}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := state.HandleSelectionError(ctx, errors.New("retry selection"))

	require.True(t, result.Handled)
	require.Equal(t, RuntimeSelectionFailureCanceled, result.Outcome)
	require.Equal(t, RuntimeFailoverCanceled, result.Action)
	require.Equal(t, state.LastFailoverErr, result.FailoverErr)
}
