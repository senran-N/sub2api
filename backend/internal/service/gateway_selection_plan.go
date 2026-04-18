package service

import "context"

type gatewaySelectionPlan struct {
	platform            string
	hasForcePlatform    bool
	useMixed            bool
	preferOAuth         bool
	routingAccountIDs   []int64
	routingSet          map[int64]struct{}
	schedGroup          *Group
	stickyPlatformCheck func(*Account) bool
	platformFilter      func(*Account) bool
	oauthTieBreaker     func(a, b *Account) bool
	failureUseMixed     bool
}

func (p *gatewaySelectionPlan) routingEnabled() bool {
	return p != nil && len(p.routingAccountIDs) > 0
}

func (p *gatewaySelectionPlan) newCandidateFilterParams(
	ctx context.Context,
	requestedModel string,
	excludedIDs map[int64]struct{},
	includeRoutingSet bool,
) *candidateFilterParams {
	if p == nil {
		return nil
	}

	params := &candidateFilterParams{
		ctx:            ctx,
		requestedModel: requestedModel,
		excludedIDs:    excludedIDs,
		schedGroup:     p.schedGroup,
		platformFilter: p.platformFilter,
	}
	if includeRoutingSet {
		params.routingSet = p.routingSet
	}
	return params
}

func (p *gatewaySelectionPlan) legacyDebugLabels() (begin string, stickyHit string, selectLabel string) {
	if p != nil && p.useMixed {
		return "[ModelRoutingDebug] legacy mixed routed begin: group_id=%v model=%s platform=%s session=%s routed_ids=%v",
			"[ModelRoutingDebug] legacy mixed routed sticky hit: group_id=%v model=%s session=%s account=%d",
			"[ModelRoutingDebug] legacy mixed routed select: group_id=%v model=%s session=%s account=%d"
	}
	return "[ModelRoutingDebug] legacy routed begin: group_id=%v model=%s platform=%s session=%s routed_ids=%v",
		"[ModelRoutingDebug] legacy routed sticky hit: group_id=%v model=%s session=%s account=%d",
		"[ModelRoutingDebug] legacy routed select: group_id=%v model=%s session=%s account=%d"
}

func (s *GatewayService) buildGatewaySelectionPlan(scope *gatewaySelectionScope, requestedModel string) *gatewaySelectionPlan {
	if scope == nil {
		return nil
	}

	schedGroup := scope.group
	if scope.groupID != nil {
		if loadedGroup := s.loadSchedulingGroup(scope.ctx, scope.groupID); loadedGroup != nil {
			schedGroup = loadedGroup
		}
	}

	plan := &gatewaySelectionPlan{
		platform:         scope.platform,
		hasForcePlatform: scope.hasForcePlatform,
		useMixed:         scope.useMixed,
		preferOAuth:      scope.preferOAuth,
		schedGroup:       schedGroup,
	}

	if scope.useMixed {
		plan.failureUseMixed = true
		plan.stickyPlatformCheck = func(a *Account) bool {
			return a.Platform == scope.platform || (a.Platform == PlatformAntigravity && a.IsMixedSchedulingEnabled())
		}
		plan.platformFilter = func(a *Account) bool {
			return a.Platform != PlatformAntigravity || a.IsMixedSchedulingEnabled()
		}
		plan.oauthTieBreaker = func(a, b *Account) bool {
			return scope.preferOAuth && a.Platform == PlatformGemini && b.Platform == PlatformGemini && a.Type != b.Type && a.Type == AccountTypeOAuth
		}
	} else {
		plan.stickyPlatformCheck = func(a *Account) bool { return a.Platform == scope.platform }
		plan.oauthTieBreaker = func(a, b *Account) bool {
			return scope.preferOAuth && a.Type != b.Type && a.Type == AccountTypeOAuth
		}
	}

	plan.routingAccountIDs = s.resolveSelectionRoutingAccountIDs(schedGroup, requestedModel, scope.stickyAccountID)
	plan.routingSet = buildRoutingSet(plan.routingAccountIDs)
	return plan
}
