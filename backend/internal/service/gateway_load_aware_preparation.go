package service

import (
	"context"
	"sort"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type gatewaySelectionScope struct {
	ctx              context.Context
	group            *Group
	groupID          *int64
	platform         string
	hasForcePlatform bool
	useMixed         bool
	preferOAuth      bool
	stickyAccountID  int64
}

type loadAwareSchedulingState struct {
	ctx         context.Context
	accounts    []Account
	accountByID map[int64]*Account
	plan        *gatewaySelectionPlan
}

func (s *GatewayService) prepareGatewaySelectionScope(ctx context.Context, groupID *int64, sessionHash string) (*gatewaySelectionScope, error) {
	scopedCtx := ctx
	resolvedGroupID := groupID

	forcePlatform, hasForcePlatform := ctx.Value(ctxkey.ForcePlatform).(string)
	if !hasForcePlatform || forcePlatform == "" {
		hasForcePlatform = false
		forcePlatform = ""
	}

	var group *Group
	var err error
	if hasForcePlatform {
		resolvedGroupID = groupID
	} else if groupID != nil {
		group, resolvedGroupID, err = s.resolveGatewayGroup(ctx, groupID)
		if err != nil {
			return nil, err
		}
		scopedCtx = s.withGroupContext(ctx, group)
	}

	platform := forcePlatform
	if platform == "" {
		switch {
		case group != nil:
			platform = group.Platform
		default:
			platform = PlatformAnthropic
		}
	}

	return &gatewaySelectionScope{
		ctx:              scopedCtx,
		group:            group,
		groupID:          resolvedGroupID,
		platform:         platform,
		hasForcePlatform: hasForcePlatform,
		useMixed:         (platform == PlatformAnthropic || platform == PlatformGemini) && !hasForcePlatform,
		preferOAuth:      platform == PlatformGemini,
		stickyAccountID:  s.resolveStickyAccountID(scopedCtx, resolvedGroupID, sessionHash),
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

//nolint:unused // 保留给 load-aware 调度回归测试直接校验准备态，不改热路径调用方式。
func (s *GatewayService) prepareLoadAwareSchedulingState(
	ctx context.Context,
	groupID *int64,
	group *Group,
	requestedModel string,
	sessionHash string,
	stickyAccountID int64,
) (*loadAwareSchedulingState, error) {
	scope := &gatewaySelectionScope{
		ctx:             ctx,
		group:           group,
		groupID:         groupID,
		stickyAccountID: stickyAccountID,
	}
	platform, hasForcePlatform, err := s.resolvePlatform(ctx, groupID, group)
	if err != nil {
		return nil, err
	}
	scope.platform = platform
	scope.hasForcePlatform = hasForcePlatform
	scope.useMixed = (platform == PlatformAnthropic || platform == PlatformGemini) && !hasForcePlatform
	scope.preferOAuth = platform == PlatformGemini
	return s.buildLoadAwareSchedulingState(scope, s.buildGatewaySelectionPlan(scope, requestedModel), requestedModel, sessionHash)
}

func (s *GatewayService) buildLoadAwareSchedulingState(
	scope *gatewaySelectionScope,
	plan *gatewaySelectionPlan,
	requestedModel string,
	sessionHash string,
) (*loadAwareSchedulingState, error) {
	if scope == nil {
		return nil, ErrNoAvailableAccounts
	}
	if plan == nil {
		return nil, ErrNoAvailableAccounts
	}
	ctx := scope.ctx
	groupID := scope.groupID
	stickyAccountID := scope.stickyAccountID

	if s.debugModelRoutingEnabled() && plan.platform == PlatformAnthropic && requestedModel != "" {
		logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] load-aware enabled: group_id=%v model=%s session=%s platform=%s", derefGroupID(groupID), requestedModel, shortSessionHash(sessionHash), plan.platform)
	}

	if s.shouldUseIndexedCandidateSource() {
		var prefetchIDs []int64
		if stickyAccountID > 0 {
			prefetchIDs = append(prefetchIDs, stickyAccountID)
		}
		accounts := s.loadSelectionAccountsByID(ctx, prefetchIDs)
		return &loadAwareSchedulingState{
			ctx:         s.prefetchSelectionSignals(ctx, accounts),
			accounts:    nil,
			accountByID: buildAccountIndexByID(accounts),
			plan:        plan,
		}, nil
	}

	accounts, _, err := s.listSchedulableAccounts(ctx, groupID, plan.platform, plan.hasForcePlatform)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	return &loadAwareSchedulingState{
		ctx:         s.prefetchSelectionSignals(ctx, accounts),
		accounts:    accounts,
		accountByID: buildAccountIndexByID(accounts),
		plan:        plan,
	}, nil
}

func buildAccountIndexByID(accounts []Account) map[int64]*Account {
	accountByID := make(map[int64]*Account, len(accounts))
	for i := range accounts {
		accountByID[accounts[i].ID] = &accounts[i]
	}
	return accountByID
}

func (s *GatewayService) resolveSelectionRoutingAccountIDs(group *Group, requestedModel string, stickyAccountID int64) []int64 {
	if group == nil || requestedModel == "" || group.Platform != PlatformAnthropic {
		return nil
	}

	routingAccountIDs := group.GetRoutingAccountIDs(requestedModel)
	if s.debugModelRoutingEnabled() {
		logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] context group routing: group_id=%d model=%s enabled=%v rules=%d matched_ids=%v session=%s sticky_account=%d",
			group.ID, requestedModel, group.ModelRoutingEnabled, len(group.ModelRouting), routingAccountIDs, "-", stickyAccountID)
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
