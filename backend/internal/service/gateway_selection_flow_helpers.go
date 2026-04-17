package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// derefGroupID safely dereferences *int64 to int64, returning 0 if nil
func derefGroupID(groupID *int64) int64 {
	if groupID == nil {
		return 0
	}
	return *groupID
}

func prefetchedStickyGroupIDFromContext(ctx context.Context) (int64, bool) {
	return PrefetchedStickyGroupIDFromContext(ctx)
}

func prefetchedStickyAccountIDFromContext(ctx context.Context, groupID *int64) int64 {
	prefetchedGroupID, ok := prefetchedStickyGroupIDFromContext(ctx)
	if !ok || prefetchedGroupID != derefGroupID(groupID) {
		return 0
	}
	if accountID, ok := PrefetchedStickyAccountIDFromContext(ctx); ok && accountID > 0 {
		return accountID
	}
	return 0
}

func buildRoutingSet(accountIDs []int64) map[int64]struct{} {
	if len(accountIDs) == 0 {
		return nil
	}
	routingSet := make(map[int64]struct{}, len(accountIDs))
	for _, id := range accountIDs {
		if id > 0 {
			routingSet[id] = struct{}{}
		}
	}
	return routingSet
}

func (s *GatewayService) listAccountsForSelection(ctx context.Context, groupID *int64, platform string, allowForcePlatform bool) ([]Account, error) {
	hasForcePlatform := false
	if allowForcePlatform {
		forcePlatform, ok := ctx.Value(ctxkey.ForcePlatform).(string)
		hasForcePlatform = ok && forcePlatform != ""
	}
	accounts, _, err := s.listSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *GatewayService) tryStickySessionAccount(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	platformCheck func(*Account) bool,
) (*Account, bool) {
	if sessionHash == "" || s.cache == nil {
		return nil, false
	}

	accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
	if err != nil || accountID <= 0 {
		return nil, false
	}
	if _, excluded := excludedIDs[accountID]; excluded {
		return nil, false
	}

	account, err := s.getSchedulableAccount(ctx, accountID)
	if err != nil {
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, stickyBindingDisposition{
			Outcome: stickyBindingOutcomeSoftMiss,
			Reason:  "lookup_error",
		}, accountID, sessionHash, "")
		return nil, false
	}

	disposition := classifyStickyBindingDisposition(account, requestedModel)
	switch disposition.Outcome {
	case stickyBindingOutcomeHardInvalidate:
		_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
		return nil, false
	case stickyBindingOutcomeSoftMiss:
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
		return nil, false
	}
	if !s.isStickyAccountFullySchedulable(ctx, account, groupID, requestedModel, true, platformCheck) {
		return nil, false
	}

	return account, true
}

func (s *GatewayService) forEachIndexedSelectionBatch(
	ctx context.Context,
	groupID *int64,
	pager *schedulerIndexedAccountPager,
	schedGroup *Group,
	pageSize int,
	visit func(pageCtx context.Context, batch []*Account) (bool, error),
) (bool, error) {
	if pager == nil || visit == nil {
		return false, nil
	}
	if pageSize <= 0 {
		pageSize = 1
	}

	scopedFound := false
	for {
		batch, hasMore, err := pager.NextRefs(ctx, pageSize)
		if err != nil {
			return scopedFound, err
		}
		if len(batch) == 0 {
			if !hasMore {
				return scopedFound, nil
			}
			continue
		}

		scopedFound = true
		batch = s.filterSelectionBatchByIndexedCapabilitiesFromPointers(ctx, groupID, pager.platform, pager.hasForcePlatform, batch, schedGroup)
		if len(batch) > 0 {
			pageCtx := s.prefetchSelectionSignalsFromPointers(ctx, batch)
			stop, err := visit(pageCtx, batch)
			if err != nil {
				return scopedFound, err
			}
			if stop {
				return scopedFound, nil
			}
		}

		if !hasMore {
			return scopedFound, nil
		}
	}
}

func (s *GatewayService) selectBestCandidateFromPagerWithStickyPolicy(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	pager *schedulerIndexedAccountPager,
	filterParams *candidateFilterParams,
	oauthTieBreaker func(a, b *Account) bool,
	pageSize int,
	bindSticky bool,
) (*Account, bool, error) {
	if pager == nil || filterParams == nil {
		return nil, false, nil
	}
	selected, scopedFound, err := s.selectBestCandidateFromPager(ctx, groupID, pager, filterParams, oauthTieBreaker, pageSize)
	if err != nil {
		return nil, false, err
	}
	return s.finalizeSelectedGatewayCandidate(ctx, groupID, sessionHash, selected, bindSticky), scopedFound, nil
}

func (s *GatewayService) selectBestCandidateFromPager(
	ctx context.Context,
	groupID *int64,
	pager *schedulerIndexedAccountPager,
	filterParams *candidateFilterParams,
	oauthTieBreaker func(a, b *Account) bool,
	pageSize int,
) (*Account, bool, error) {
	if pager == nil || filterParams == nil {
		return nil, false, nil
	}
	if pageSize <= 0 {
		pageSize = 1
	}

	var selected *Account
	scopedFound, err := s.forEachIndexedSelectionBatch(ctx, groupID, pager, filterParams.schedGroup, pageSize, func(scopedCtx context.Context, batch []*Account) (bool, error) {
		params := *filterParams
		params.ctx = scopedCtx
		candidates := s.filterCandidatePointers(batch, &params)
		pageBest := selectBestByPriorityAndLastUsed(candidates, oauthTieBreaker)
		if s.isBetterSelectionCandidate(pageBest, selected, oauthTieBreaker) {
			selected = pageBest
		}
		return false, nil
	})
	if err != nil {
		return nil, false, err
	}
	return selected, scopedFound, nil
}

func (s *GatewayService) selectBestCandidateWithStickyPolicy(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	accounts []Account,
	filterParams *candidateFilterParams,
	oauthTieBreaker func(a, b *Account) bool,
	bindSticky bool,
) *Account {
	selected := s.selectBestCandidate(accounts, filterParams, oauthTieBreaker)
	return s.finalizeSelectedGatewayCandidate(ctx, groupID, sessionHash, selected, bindSticky)
}

func (s *GatewayService) selectBestCandidate(
	accounts []Account,
	filterParams *candidateFilterParams,
	oauthTieBreaker func(a, b *Account) bool,
) *Account {
	candidates := s.filterCandidates(accounts, filterParams)
	shuffleCandidatesByPriority(candidates)
	return selectBestByPriorityAndLastUsed(candidates, oauthTieBreaker)
}

func (s *GatewayService) finalizeSelectedGatewayCandidate(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	selected *Account,
	bindSticky bool,
) *Account {
	if selected == nil {
		return nil
	}
	if bindSticky {
		s.bindStickySelection(ctx, groupID, sessionHash, selected.ID)
	}
	return selected
}

func (s *GatewayService) bindStickySelection(ctx context.Context, groupID *int64, sessionHash string, accountID int64) {
	if sessionHash == "" || s.cache == nil || accountID <= 0 {
		return
	}
	if err := s.BindStickySession(ctx, groupID, sessionHash, accountID); err != nil {
		logger.LegacyPrintf("service.gateway", "set session account failed: session=%s account_id=%d err=%v", sessionHash, accountID, err)
	}
}
