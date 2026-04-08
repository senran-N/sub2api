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

	for _, item := range available {
		if result, ok := s.tryAcquireAndMaybeBindSelection(ctx, groupID, sessionHash, item.account, true); ok {
			if s.debugModelRoutingEnabled() {
				logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] routed select: group_id=%v model=%s session=%s account=%d",
					derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), item.account.ID)
			}
			return result, true
		}
	}

	for _, item := range available {
		if result, _, ok := s.tryBuildAccountWaitPlan(ctx, item.account, sessionHash, waitTimeout, maxWaiting); ok {
			if s.debugModelRoutingEnabled() {
				logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] routed wait: group_id=%v model=%s session=%s account=%d",
					derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), item.account.ID)
			}
			return result, true
		}
	}

	return nil, false
}
