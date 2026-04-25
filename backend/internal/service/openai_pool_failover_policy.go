package service

import "net/http"

type OpenAIPoolFailoverDecision struct {
	Action           RuntimeFailoverAction
	SameAccountRetry bool
	RetryCount       int
	RetryLimit       int
	SwitchCount      int
}

func ApplyOpenAIPoolFailoverPolicy(
	account *Account,
	failoverErr *UpstreamFailoverError,
	codexDecision CodexRecoveryDecision,
	hasSessionAffinity bool,
	sameAccountRetryCount map[int64]int,
	failedAccountIDs map[int64]struct{},
	switchCount *int,
	maxAccountSwitches int,
	tempUnscheduleRetryable func(),
	recordSwitch func(),
) OpenAIPoolFailoverDecision {
	decision := OpenAIPoolFailoverDecision{Action: RuntimeFailoverExhausted}
	if account == nil || failoverErr == nil || switchCount == nil {
		return decision
	}

	decision.RetryLimit = account.GetPoolModeRetryCount()
	decision.SwitchCount = *switchCount

	forceSwitchAccount := codexDecision.SwitchAccount

	if codexDecision.ExhaustFailover {
		if failedAccountIDs != nil {
			failedAccountIDs[account.ID] = struct{}{}
		}
		return decision
	}

	if !forceSwitchAccount && shouldExhaustOpenAIPoolFailoverImmediately(failoverErr, hasSessionAffinity) {
		if failedAccountIDs != nil {
			failedAccountIDs[account.ID] = struct{}{}
		}
		return decision
	}

	if !forceSwitchAccount && failoverErr.RetryableOnSameAccount && sameAccountRetryCount != nil && sameAccountRetryCount[account.ID] < decision.RetryLimit {
		sameAccountRetryCount[account.ID]++
		decision.Action = RuntimeFailoverContinue
		decision.SameAccountRetry = true
		decision.RetryCount = sameAccountRetryCount[account.ID]
		return decision
	}

	if !forceSwitchAccount && shouldPreserveOpenAIPoolBoundSessionOnRateLimit(failoverErr, hasSessionAffinity) {
		if failedAccountIDs != nil {
			failedAccountIDs[account.ID] = struct{}{}
		}
		return decision
	}

	if !forceSwitchAccount && failoverErr.RetryableOnSameAccount && tempUnscheduleRetryable != nil {
		tempUnscheduleRetryable()
	}

	if failedAccountIDs != nil {
		failedAccountIDs[account.ID] = struct{}{}
	}
	if *switchCount >= maxAccountSwitches {
		return decision
	}

	if recordSwitch != nil {
		recordSwitch()
	}
	*switchCount++
	decision.Action = RuntimeFailoverContinue
	decision.SwitchCount = *switchCount
	return decision
}

func shouldExhaustOpenAIPoolFailoverImmediately(failoverErr *UpstreamFailoverError, hasSessionAffinity bool) bool {
	if failoverErr == nil {
		return false
	}
	switch failoverErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return true
	case http.StatusTooManyRequests:
		return !hasSessionAffinity
	default:
		return false
	}
}

func shouldPreserveOpenAIPoolBoundSessionOnRateLimit(failoverErr *UpstreamFailoverError, hasSessionAffinity bool) bool {
	return hasSessionAffinity && failoverErr != nil && failoverErr.StatusCode == http.StatusTooManyRequests
}
