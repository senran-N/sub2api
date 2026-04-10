package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GatewayService) selectAvailableRoutedAccount(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	available []accountWithLoad,
	waitTimeout time.Duration,
	maxWaiting int,
) (*AccountSelectionResult, bool) {
	if len(available) == 0 {
		return nil, false
	}

	sortAccountsByPriorityLoadAndLastUsed(available, false)
	ordered := make([]*Account, 0, len(available))
	for _, item := range available {
		ordered = append(ordered, item.account)
	}

	result, _, ok := selectFirstOrderedRuntimeSelection(ordered, func(account *Account) (*AccountSelectionResult, error, bool) {
		result, ok := s.tryAcquireAndMaybeBindSelection(ctx, groupID, sessionHash, account, true)
		return result, nil, ok
	})
	if ok {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] routed select: group_id=%v model=%s session=%s account=%d",
				derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), result.Account.ID)
		}
		return result, true
	}

	result, ok = selectFirstOrderedWaitPlan(ordered, func(account *Account) (*AccountSelectionResult, string, bool) {
		return s.tryBuildAccountWaitPlan(ctx, account, sessionHash, waitTimeout, maxWaiting)
	})
	if ok {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] routed wait: group_id=%v model=%s session=%s account=%d",
				derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), result.Account.ID)
		}
		return result, true
	}

	return nil, false
}
