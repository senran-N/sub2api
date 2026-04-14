package handler

import "github.com/senran-N/sub2api/internal/service"

type openAIPoolFailoverDecision struct {
	Action           FailoverAction
	SameAccountRetry bool
	RetryCount       int
	RetryLimit       int
	SwitchCount      int
}

func applyOpenAIPoolFailoverPolicy(
	account *service.Account,
	failoverErr *service.UpstreamFailoverError,
	sameAccountRetryCount map[int64]int,
	failedAccountIDs map[int64]struct{},
	switchCount *int,
	maxAccountSwitches int,
	recordSwitch func(),
) openAIPoolFailoverDecision {
	decision := openAIPoolFailoverDecision{Action: FailoverExhausted}
	if account == nil || failoverErr == nil || switchCount == nil {
		return decision
	}

	decision.RetryLimit = account.GetPoolModeRetryCount()
	decision.SwitchCount = *switchCount

	if shouldExhaustFailoverImmediately(failoverErr) {
		if failedAccountIDs != nil {
			failedAccountIDs[account.ID] = struct{}{}
		}
		return decision
	}

	if failoverErr.RetryableOnSameAccount && sameAccountRetryCount != nil && sameAccountRetryCount[account.ID] < decision.RetryLimit {
		sameAccountRetryCount[account.ID]++
		decision.Action = FailoverContinue
		decision.SameAccountRetry = true
		decision.RetryCount = sameAccountRetryCount[account.ID]
		return decision
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
	decision.Action = FailoverContinue
	decision.SwitchCount = *switchCount
	return decision
}
