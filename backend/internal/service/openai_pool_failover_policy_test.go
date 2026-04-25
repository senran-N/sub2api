package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyOpenAIPoolFailoverPolicyServiceSameAccountRetry(t *testing.T) {
	account := &Account{ID: 21}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 0

	decision := ApplyOpenAIPoolFailoverPolicy(
		account,
		&UpstreamFailoverError{StatusCode: http.StatusBadGateway, RetryableOnSameAccount: true},
		CodexRecoveryDecision{},
		false,
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() {},
		func() {},
	)

	require.Equal(t, RuntimeFailoverContinue, decision.Action)
	require.True(t, decision.SameAccountRetry)
	require.Equal(t, 1, decision.RetryCount)
	require.Empty(t, failedAccountIDs)
	require.Zero(t, switchCount)
}

func TestApplyOpenAIPoolFailoverPolicyServiceSwitchesAfterRetryBudget(t *testing.T) {
	account := &Account{ID: 22}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := map[int64]int{22: account.GetPoolModeRetryCount()}
	switchCount := 0
	tempUnscheduleCalls := 0
	switchCalls := 0

	decision := ApplyOpenAIPoolFailoverPolicy(
		account,
		&UpstreamFailoverError{StatusCode: http.StatusBadGateway, RetryableOnSameAccount: true},
		CodexRecoveryDecision{},
		false,
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() { tempUnscheduleCalls++ },
		func() { switchCalls++ },
	)

	require.Equal(t, RuntimeFailoverContinue, decision.Action)
	require.False(t, decision.SameAccountRetry)
	require.Equal(t, 1, decision.SwitchCount)
	require.Contains(t, failedAccountIDs, int64(22))
	require.Equal(t, 1, tempUnscheduleCalls)
	require.Equal(t, 1, switchCalls)
}

func TestApplyOpenAIPoolFailoverPolicyServiceImmediateExhaustStatuses(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			account := &Account{ID: 23}
			failedAccountIDs := make(map[int64]struct{})
			sameAccountRetryCount := make(map[int64]int)
			switchCount := 1
			switchCalls := 0

			decision := ApplyOpenAIPoolFailoverPolicy(
				account,
				&UpstreamFailoverError{StatusCode: status, RetryableOnSameAccount: true},
				CodexRecoveryDecision{},
				false,
				sameAccountRetryCount,
				failedAccountIDs,
				&switchCount,
				3,
				func() {},
				func() { switchCalls++ },
			)

			require.Equal(t, RuntimeFailoverExhausted, decision.Action)
			require.False(t, decision.SameAccountRetry)
			require.Equal(t, 1, switchCount)
			require.Contains(t, failedAccountIDs, account.ID)
			require.Empty(t, sameAccountRetryCount)
			require.Zero(t, switchCalls)
		})
	}
}

func TestApplyOpenAIPoolFailoverPolicyServiceExhaustsWhenSwitchBudgetReached(t *testing.T) {
	account := &Account{ID: 24}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 2
	switchCalls := 0
	tempUnscheduleCalls := 0

	decision := ApplyOpenAIPoolFailoverPolicy(
		account,
		&UpstreamFailoverError{StatusCode: http.StatusBadGateway},
		CodexRecoveryDecision{},
		false,
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		2,
		func() { tempUnscheduleCalls++ },
		func() { switchCalls++ },
	)

	require.Equal(t, RuntimeFailoverExhausted, decision.Action)
	require.Equal(t, 2, decision.SwitchCount)
	require.Equal(t, 2, switchCount)
	require.Contains(t, failedAccountIDs, account.ID)
	require.Zero(t, switchCalls)
	require.Zero(t, tempUnscheduleCalls)
}

func TestApplyOpenAIPoolFailoverPolicyServiceCodexExhaustDecision(t *testing.T) {
	account := &Account{ID: 25}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 0
	switchCalls := 0

	decision := ApplyOpenAIPoolFailoverPolicy(
		account,
		&UpstreamFailoverError{StatusCode: http.StatusBadGateway},
		CodexRecoveryDecision{
			Applied:         true,
			Action:          "exhaust_failover",
			ExhaustFailover: true,
			Reason:          "failover",
		},
		false,
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() {},
		func() { switchCalls++ },
	)

	require.Equal(t, RuntimeFailoverExhausted, decision.Action)
	require.Contains(t, failedAccountIDs, account.ID)
	require.Zero(t, switchCalls)
}

func TestApplyOpenAIPoolFailoverPolicyServiceCodexSwitchOverridesImmediate429Exhaust(t *testing.T) {
	account := &Account{ID: 26}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 0
	switchCalls := 0

	decision := ApplyOpenAIPoolFailoverPolicy(
		account,
		&UpstreamFailoverError{StatusCode: http.StatusTooManyRequests, RetryableOnSameAccount: true},
		CodexRecoveryDecision{
			Applied:       true,
			Action:        "switch_account",
			SwitchAccount: true,
			Reason:        "failover",
		},
		true,
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() {},
		func() { switchCalls++ },
	)

	require.Equal(t, RuntimeFailoverContinue, decision.Action)
	require.False(t, decision.SameAccountRetry)
	require.Equal(t, 1, decision.SwitchCount)
	require.Equal(t, 1, switchCount)
	require.Contains(t, failedAccountIDs, account.ID)
	require.Empty(t, sameAccountRetryCount)
	require.Equal(t, 1, switchCalls)
}

func TestApplyOpenAIPoolFailoverPolicyServiceBoundSession429RetriesThenExhausts(t *testing.T) {
	account := &Account{ID: 27}
	failedAccountIDs := make(map[int64]struct{})
	sameAccountRetryCount := make(map[int64]int)
	switchCount := 0
	switchCalls := 0
	tempUnscheduleCalls := 0

	for i := 1; i <= account.GetPoolModeRetryCount(); i++ {
		decision := ApplyOpenAIPoolFailoverPolicy(
			account,
			&UpstreamFailoverError{StatusCode: http.StatusTooManyRequests, RetryableOnSameAccount: true},
			CodexRecoveryDecision{},
			true,
			sameAccountRetryCount,
			failedAccountIDs,
			&switchCount,
			3,
			func() { tempUnscheduleCalls++ },
			func() { switchCalls++ },
		)

		require.Equal(t, RuntimeFailoverContinue, decision.Action)
		require.True(t, decision.SameAccountRetry)
		require.Equal(t, i, decision.RetryCount)
	}

	decision := ApplyOpenAIPoolFailoverPolicy(
		account,
		&UpstreamFailoverError{StatusCode: http.StatusTooManyRequests, RetryableOnSameAccount: true},
		CodexRecoveryDecision{},
		true,
		sameAccountRetryCount,
		failedAccountIDs,
		&switchCount,
		3,
		func() { tempUnscheduleCalls++ },
		func() { switchCalls++ },
	)

	require.Equal(t, RuntimeFailoverExhausted, decision.Action)
	require.False(t, decision.SameAccountRetry)
	require.Equal(t, 0, switchCount)
	require.Contains(t, failedAccountIDs, account.ID)
	require.Zero(t, switchCalls)
	require.Zero(t, tempUnscheduleCalls)
}
