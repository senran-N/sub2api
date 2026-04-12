package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
)

func (s *GatewayService) shouldUseIndexedCandidateSource() bool {
	return s != nil && s.schedulerSnapshot != nil
}

func (s *GatewayService) hasForcedSelectionPlatform(ctx context.Context, allowForcePlatform bool) bool {
	if !allowForcePlatform {
		return false
	}
	_, ok := forcePlatformFromContext(ctx)
	return ok
}

func forcePlatformFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	forcePlatform, ok := ctx.Value(ctxkey.ForcePlatform).(string)
	return forcePlatform, ok && forcePlatform != ""
}

func (s *GatewayService) buildGatewayIndexedCandidatePager(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	requestedModel string,
	schedGroup *Group,
) (*schedulerIndexedAccountPager, error) {
	if s == nil || s.schedulerSnapshot == nil {
		return nil, nil
	}
	sources, err := buildRequestedModelCapabilitySources(ctx, s.schedulerSnapshot, groupID, platform, hasForcePlatform, requestedModel)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		if schedGroup != nil && schedGroup.RequirePrivacySet {
			sources = []SchedulerCapabilityIndex{{Kind: SchedulerCapabilityIndexPrivacySet}}
		} else {
			sources = []SchedulerCapabilityIndex{{Kind: SchedulerCapabilityIndexAll}}
		}
	}
	return newSchedulerIndexedAccountPager(s.schedulerSnapshot, groupID, platform, hasForcePlatform, sources), nil
}

func (s *GatewayService) filterSelectionBatchByIndexedCapabilities(
	ctx context.Context,
	groupID *int64,
	platform string,
	hasForcePlatform bool,
	accounts []Account,
	schedGroup *Group,
) []Account {
	if len(accounts) == 0 || s == nil || s.schedulerSnapshot == nil || schedGroup == nil || !schedGroup.RequirePrivacySet {
		return accounts
	}
	accountIDs := make([]int64, 0, len(accounts))
	for i := range accounts {
		accountIDs = append(accountIDs, accounts[i].ID)
	}
	matches, _, err := s.schedulerSnapshot.MatchSchedulableAccountsCapability(
		ctx,
		groupID,
		platform,
		hasForcePlatform,
		SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexPrivacySet},
		accountIDs,
	)
	if err != nil {
		return accounts
	}
	filtered := make([]Account, 0, len(accounts))
	for i := range accounts {
		if matches[accounts[i].ID] {
			filtered = append(filtered, accounts[i])
		}
	}
	return filtered
}

func (s *GatewayService) loadSelectionAccountsByID(ctx context.Context, accountIDs []int64) []Account {
	if len(accountIDs) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(accountIDs))
	accounts := make([]Account, 0, len(accountIDs))
	accountByID := make(map[int64]*Account, len(accountIDs))
	for _, accountID := range accountIDs {
		if accountID <= 0 {
			continue
		}
		if _, exists := seen[accountID]; exists {
			continue
		}
		seen[accountID] = struct{}{}
		account, ok := s.resolveSelectionAccountByID(ctx, accountByID, accountID)
		if !ok {
			continue
		}
		accounts = append(accounts, *account)
	}
	return accounts
}

func (s *GatewayService) resolveSelectionAccountByID(ctx context.Context, accountByID map[int64]*Account, accountID int64) (*Account, bool) {
	if accountID <= 0 {
		return nil, false
	}
	if accountByID != nil {
		if account, ok := accountByID[accountID]; ok && account != nil {
			return account, true
		}
	}

	account, err := s.getSchedulableAccount(ctx, accountID)
	if err != nil || account == nil {
		return nil, false
	}
	if accountByID != nil {
		accountByID[accountID] = account
	}
	return account, true
}

func (s *GatewayService) resolveSelectionAccountFromCurrentMap(ctx context.Context, accountByID map[int64]*Account, accountID int64) (*Account, bool) {
	if accountID <= 0 {
		return nil, false
	}
	if accountByID != nil {
		if account, ok := accountByID[accountID]; ok && account != nil {
			return account, true
		}
		if len(accountByID) > 0 && !s.shouldUseIndexedCandidateSource() {
			return nil, false
		}
	}
	return s.resolveSelectionAccountByID(ctx, nil, accountID)
}

func (s *GatewayService) isBetterSelectionCandidate(
	candidate *Account,
	current *Account,
	oauthTieBreaker func(a, b *Account) bool,
) bool {
	if candidate == nil {
		return false
	}
	if current == nil {
		return true
	}
	if candidate.Priority != current.Priority {
		return candidate.Priority < current.Priority
	}
	switch {
	case candidate.LastUsedAt == nil && current.LastUsedAt != nil:
		return true
	case candidate.LastUsedAt != nil && current.LastUsedAt == nil:
		return false
	case candidate.LastUsedAt == nil && current.LastUsedAt == nil:
		return oauthTieBreaker != nil && oauthTieBreaker(candidate, current)
	default:
		return candidate.LastUsedAt.Before(*current.LastUsedAt)
	}
}
