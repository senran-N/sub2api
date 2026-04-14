package handler

import (
	"net/http"
	"testing"

	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestApplyOpenAIPoolFailoverPolicy_ImmediateExhaustStatuses(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			account := &service.Account{ID: 11}
			failedAccountIDs := make(map[int64]struct{})
			sameAccountRetryCount := make(map[int64]int)
			switchCount := 1
			recordedSwitches := 0

			decision := applyOpenAIPoolFailoverPolicy(
				account,
				&service.UpstreamFailoverError{StatusCode: status, RetryableOnSameAccount: true},
				sameAccountRetryCount,
				failedAccountIDs,
				&switchCount,
				3,
				func() {},
				func() { recordedSwitches++ },
			)

			require.Equal(t, FailoverExhausted, decision.Action)
			require.False(t, decision.SameAccountRetry)
			require.Equal(t, 1, switchCount)
			require.Contains(t, failedAccountIDs, int64(11))
			require.Empty(t, sameAccountRetryCount)
			require.Zero(t, recordedSwitches)
		})
	}
}

func TestApplyOpenAIPoolFailoverPolicy_SameAccountRetry(t *testing.T) {
	account := &service.Account{ID: 12}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 0
	recordedSwitches := 0

	decision := applyOpenAIPoolFailoverPolicy(
		account,
		&service.UpstreamFailoverError{StatusCode: http.StatusBadGateway, RetryableOnSameAccount: true},
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() {},
		func() { recordedSwitches++ },
	)

	require.Equal(t, FailoverContinue, decision.Action)
	require.True(t, decision.SameAccountRetry)
	require.Equal(t, 1, decision.RetryCount)
	require.Equal(t, account.GetPoolModeRetryCount(), decision.RetryLimit)
	require.Equal(t, 0, switchCount)
	require.Empty(t, failedAccountIDs)
	require.Zero(t, recordedSwitches)
}

func TestApplyOpenAIPoolFailoverPolicy_SwitchAfterRetryBudget(t *testing.T) {
	account := &service.Account{ID: 13}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := map[int64]int{13: account.GetPoolModeRetryCount()}
	switchCount := 0
	recordedSwitches := 0
	tempUnschedCalls := 0

	decision := applyOpenAIPoolFailoverPolicy(
		account,
		&service.UpstreamFailoverError{StatusCode: http.StatusBadGateway, RetryableOnSameAccount: true},
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() { tempUnschedCalls++ },
		func() { recordedSwitches++ },
	)

	require.Equal(t, FailoverContinue, decision.Action)
	require.False(t, decision.SameAccountRetry)
	require.Equal(t, 1, decision.SwitchCount)
	require.Equal(t, 1, switchCount)
	require.Contains(t, failedAccountIDs, int64(13))
	require.Equal(t, 1, recordedSwitches)
	require.Equal(t, 1, tempUnschedCalls)
}

func TestApplyOpenAIPoolFailoverPolicy_ExhaustedWhenSwitchBudgetReached(t *testing.T) {
	account := &service.Account{ID: 14}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 2
	recordedSwitches := 0
	tempUnschedCalls := 0

	decision := applyOpenAIPoolFailoverPolicy(
		account,
		&service.UpstreamFailoverError{StatusCode: http.StatusBadGateway},
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		2,
		func() { tempUnschedCalls++ },
		func() { recordedSwitches++ },
	)

	require.Equal(t, FailoverExhausted, decision.Action)
	require.Equal(t, 2, decision.SwitchCount)
	require.Equal(t, 2, switchCount)
	require.Contains(t, failedAccountIDs, int64(14))
	require.Zero(t, recordedSwitches)
	require.Zero(t, tempUnschedCalls)
}

func TestApplyOpenAIPoolFailoverPolicy_SameAccountRetryDoesNotTempUnschedule(t *testing.T) {
	account := &service.Account{ID: 15}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 0
	tempUnschedCalls := 0

	decision := applyOpenAIPoolFailoverPolicy(
		account,
		&service.UpstreamFailoverError{StatusCode: http.StatusBadGateway, RetryableOnSameAccount: true},
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() { tempUnschedCalls++ },
		nil,
	)

	require.Equal(t, FailoverContinue, decision.Action)
	require.True(t, decision.SameAccountRetry)
	require.Zero(t, tempUnschedCalls)
}
