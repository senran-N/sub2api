//go:build unit

package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteGrokRequestScopedFailover_ReturnsLastFailoverAfterPrepareExhaustion(t *testing.T) {
	selectionErr := errors.New("no compatible grok media accounts")
	accountID := int64(41)
	prepareCalls := 0

	err := executeGrokRequestScopedFailover(
		func(excludedIDs map[int64]struct{}) (*grokRequestScopedAttempt[struct{}], error) {
			prepareCalls++
			if prepareCalls == 1 {
				return &grokRequestScopedAttempt[struct{}]{
					account: &Account{ID: accountID},
				}, nil
			}
			require.Contains(t, excludedIDs, accountID)
			return nil, selectionErr
		},
		func(*grokRequestScopedAttempt[struct{}]) error {
			return &UpstreamFailoverError{
				StatusCode:   http.StatusTooManyRequests,
				ResponseBody: []byte(`{"error":{"message":"rate limit reached"}}`),
			}
		},
		func(err error) *UpstreamFailoverError {
			var failoverErr *UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				return failoverErr
			}
			return nil
		},
		func(err error) bool {
			return err == selectionErr
		},
	)

	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusTooManyRequests, failoverErr.StatusCode)
	require.Equal(t, 2, prepareCalls)
}

func TestExecuteGrokRequestScopedFailover_WrapsPrepareErrorsWithPhase(t *testing.T) {
	prepareErr := errors.New("requested model unavailable:grok-4")

	err := executeGrokRequestScopedFailover(
		func(map[int64]struct{}) (*grokRequestScopedAttempt[struct{}], error) {
			return nil, prepareErr
		},
		func(*grokRequestScopedAttempt[struct{}]) error {
			t.Fatal("execute should not be called")
			return nil
		},
		func(error) *UpstreamFailoverError { return nil },
		nil,
	)

	var scopedErr *grokRequestScopedExecutionError
	require.ErrorAs(t, err, &scopedErr)
	require.Equal(t, grokRequestScopedErrorPhasePrepare, scopedErr.phase)
	require.ErrorIs(t, err, prepareErr)
}

func TestExecuteGrokRequestScopedFailover_WrapsNonFailoverExecutionErrorsWithPhase(t *testing.T) {
	executeErr := errors.New("attachment upload failed")

	err := executeGrokRequestScopedFailover(
		func(map[int64]struct{}) (*grokRequestScopedAttempt[struct{}], error) {
			return &grokRequestScopedAttempt[struct{}]{
				account: &Account{ID: 7},
			}, nil
		},
		func(*grokRequestScopedAttempt[struct{}]) error {
			return executeErr
		},
		func(error) *UpstreamFailoverError { return nil },
		nil,
	)

	var scopedErr *grokRequestScopedExecutionError
	require.ErrorAs(t, err, &scopedErr)
	require.Equal(t, grokRequestScopedErrorPhaseExecute, scopedErr.phase)
	require.ErrorIs(t, err, executeErr)
}
