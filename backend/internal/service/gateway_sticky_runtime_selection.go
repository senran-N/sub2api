package service

import (
	"context"
	"time"
)

func (s *GatewayService) trySelectResolvedStickyAccount(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	account *Account,
	bindSticky bool,
	waitTimeout time.Duration,
	maxWaiting int,
) (*AccountSelectionResult, string, bool) {
	if account == nil {
		return nil, "", false
	}
	return trySelectStickyRuntimeSelection(stickyRuntimeSelectionSpec{
		tryAcquire: func() (*AccountSelectionResult, string, bool) {
			result, ok := s.tryAcquireAndMaybeBindSelection(ctx, groupID, sessionHash, account, bindSticky)
			return result, "", ok
		},
		buildWaitPlan: func() (*AccountSelectionResult, string, bool) {
			return s.tryBuildAccountWaitPlan(ctx, account, sessionHash, waitTimeout, maxWaiting)
		},
	})
}
