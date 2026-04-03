package service

import (
	"context"
	"time"

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

// shouldClearStickySession 检查账号是否处于不可调度状态，需要清理粘性会话绑定。
// 当账号状态为错误、禁用、不可调度、处于临时不可调度期间，
// 或请求的模型处于限流状态时，返回 true。
func shouldClearStickySession(account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	if account.Status == StatusError || account.Status == StatusDisabled || !account.Schedulable {
		return true
	}
	if account.TempUnschedulableUntil != nil && time.Now().Before(*account.TempUnschedulableUntil) {
		return true
	}
	if remaining := account.GetRateLimitRemainingTimeWithContext(context.Background(), requestedModel); remaining > 0 {
		return true
	}
	return false
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
		return nil, false
	}

	if shouldClearStickySession(account, requestedModel) {
		_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
		return nil, false
	}
	if !s.isStickyAccountFullySchedulable(ctx, account, groupID, requestedModel, true, platformCheck) {
		return nil, false
	}

	return account, true
}

func (s *GatewayService) selectBestCandidateAndBindSession(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	accounts []Account,
	filterParams *candidateFilterParams,
	oauthTieBreaker func(a, b *Account) bool,
) *Account {
	candidates := s.filterCandidates(accounts, filterParams)
	shuffleCandidatesByPriority(candidates)
	selected := selectBestByPriorityAndLastUsed(candidates, oauthTieBreaker)
	if selected == nil {
		return nil
	}
	s.bindStickySelection(ctx, groupID, sessionHash, selected.ID)
	return selected
}

func (s *GatewayService) bindStickySelection(ctx context.Context, groupID *int64, sessionHash string, accountID int64) {
	if sessionHash == "" || s.cache == nil || accountID <= 0 {
		return
	}
	if err := s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), sessionHash, accountID, stickySessionTTL); err != nil {
		logger.LegacyPrintf("service.gateway", "set session account failed: session=%s account_id=%d err=%v", sessionHash, accountID, err)
	}
}
