package service

import (
	"context"
	"time"
)

func (s *GatewayService) trySelectStickyAccountWithoutRouting(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	stickyAccountID int64,
	excludedIDs map[int64]struct{},
	accountByID map[int64]*Account,
	platform string,
	useMixed bool,
	waitTimeout time.Duration,
	maxWaiting int,
) (*AccountSelectionResult, bool) {
	if sessionHash == "" || stickyAccountID <= 0 {
		return nil, false
	}
	if _, excluded := excludedIDs[stickyAccountID]; excluded {
		return nil, false
	}

	account, ok := accountByID[stickyAccountID]
	if !ok {
		return nil, false
	}

	clearSticky := shouldClearStickySession(account, requestedModel)
	if clearSticky && s.cache != nil {
		_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
	}
	if clearSticky {
		return nil, false
	}

	if !s.isStickyAccountFullySchedulable(ctx, account, groupID, requestedModel, true,
		func(a *Account) bool { return s.isAccountAllowedForPlatform(a, platform, useMixed) }) {
		return nil, false
	}

	if result, ok := s.tryAcquireAndMaybeBindSelection(ctx, nil, sessionHash, account, false); ok {
		return result, true
	}

	if result, _, ok := s.tryBuildAccountWaitPlan(ctx, account, sessionHash, waitTimeout, maxWaiting); ok {
		return result, true
	}

	return nil, false
}
