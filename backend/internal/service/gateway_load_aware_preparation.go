package service

import (
	"context"
	"sort"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type loadAwareSelectionScope struct {
	ctx             context.Context
	group           *Group
	groupID         *int64
	stickyAccountID int64
}

type loadAwareSchedulingState struct {
	ctx               context.Context
	platform          string
	preferOAuth       bool
	accounts          []Account
	useMixed          bool
	accountByID       map[int64]*Account
	routingAccountIDs []int64
}

func (s *GatewayService) prepareLoadAwareSelectionScope(ctx context.Context, groupID *int64, sessionHash string) (*loadAwareSelectionScope, error) {
	group, resolvedGroupID, err := s.checkClaudeCodeRestriction(ctx, groupID)
	if err != nil {
		return nil, err
	}

	scopedCtx := s.withGroupContext(ctx, group)
	return &loadAwareSelectionScope{
		ctx:             scopedCtx,
		group:           group,
		groupID:         resolvedGroupID,
		stickyAccountID: s.resolveStickyAccountID(scopedCtx, resolvedGroupID, sessionHash),
	}, nil
}

func (s *GatewayService) resolveStickyAccountID(ctx context.Context, groupID *int64, sessionHash string) int64 {
	if stickyAccountID := prefetchedStickyAccountIDFromContext(ctx, groupID); stickyAccountID > 0 {
		return stickyAccountID
	}
	if sessionHash == "" || s.cache == nil {
		return 0
	}

	accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
	if err != nil || accountID <= 0 {
		return 0
	}
	return accountID
}

func (s *GatewayService) prepareLoadAwareSchedulingState(
	ctx context.Context,
	groupID *int64,
	group *Group,
	requestedModel string,
	sessionHash string,
	stickyAccountID int64,
) (*loadAwareSchedulingState, error) {
	platform, hasForcePlatform, err := s.resolvePlatform(ctx, groupID, group)
	if err != nil {
		return nil, err
	}

	if s.debugModelRoutingEnabled() && platform == PlatformAnthropic && requestedModel != "" {
		logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] load-aware enabled: group_id=%v model=%s session=%s platform=%s", derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), platform)
	}

	accounts, useMixed, err := s.listSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	return &loadAwareSchedulingState{
		ctx:               s.prefetchSelectionSignals(ctx, accounts),
		platform:          platform,
		preferOAuth:       platform == PlatformGemini,
		accounts:          accounts,
		useMixed:          useMixed,
		accountByID:       buildAccountIndexByID(accounts),
		routingAccountIDs: s.resolveLoadAwareRoutingAccountIDs(group, requestedModel, sessionHash, stickyAccountID),
	}, nil
}

func buildAccountIndexByID(accounts []Account) map[int64]*Account {
	accountByID := make(map[int64]*Account, len(accounts))
	for i := range accounts {
		accountByID[accounts[i].ID] = &accounts[i]
	}
	return accountByID
}

func (s *GatewayService) resolveLoadAwareRoutingAccountIDs(group *Group, requestedModel string, sessionHash string, stickyAccountID int64) []int64 {
	if group == nil || requestedModel == "" || group.Platform != PlatformAnthropic {
		return nil
	}

	routingAccountIDs := group.GetRoutingAccountIDs(requestedModel)
	if s.debugModelRoutingEnabled() {
		logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] context group routing: group_id=%d model=%s enabled=%v rules=%d matched_ids=%v session=%s sticky_account=%d",
			group.ID, requestedModel, group.ModelRoutingEnabled, len(group.ModelRouting), routingAccountIDs, shortSessionHash(sessionHash), stickyAccountID)
		if len(routingAccountIDs) == 0 && group.ModelRoutingEnabled && len(group.ModelRouting) > 0 {
			logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] context group routing miss: group_id=%d model=%s patterns(sample)=%v", group.ID, requestedModel, sampleSortedRoutingPatterns(group.ModelRouting))
		}
	}

	return routingAccountIDs
}

func sampleSortedRoutingPatterns(modelRouting map[string][]int64) []string {
	if len(modelRouting) == 0 {
		return nil
	}

	keys := make([]string, 0, len(modelRouting))
	for pattern := range modelRouting {
		keys = append(keys, pattern)
	}
	sort.Strings(keys)

	const maxKeys = 20
	if len(keys) > maxKeys {
		keys = keys[:maxKeys]
	}
	return keys
}
