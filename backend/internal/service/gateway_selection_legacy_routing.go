package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type legacyRoutedSelectionInput struct {
	groupID             *int64
	sessionHash         string
	requestedModel      string
	excludedIDs         map[int64]struct{}
	plan                *gatewaySelectionPlan
	debugBeginLabel     string
	debugStickyHitLabel string
	debugSelectLabel    string
	bindSticky          bool
}

func (s *GatewayService) loadSchedulingGroup(ctx context.Context, groupID *int64) *Group {
	if groupID == nil || s.groupRepo == nil {
		return nil
	}
	group, _ := s.groupRepo.GetByID(ctx, *groupID)
	return group
}

func (s *GatewayService) prefetchSelectionSignals(ctx context.Context, accounts []Account) context.Context {
	ctx = s.withWindowCostPrefetch(ctx, accounts)
	return s.withRPMPrefetch(ctx, accounts)
}

func (s *GatewayService) prefetchSelectionSignalsFromPointers(ctx context.Context, accounts []*Account) context.Context {
	if ctx == nil || len(accounts) == 0 {
		return ctx
	}
	if s == nil {
		return ctx
	}
	if s.rpmCache == nil && (s.sessionLimitCache == nil || s.usageLogRepo == nil) {
		return ctx
	}
	return s.prefetchSelectionSignals(ctx, derefAccounts(accounts))
}

func (s *GatewayService) tryLegacyRoutedSelection(
	ctx context.Context,
	input *legacyRoutedSelectionInput,
) (context.Context, []Account, bool, *Account, error) {
	if input == nil || input.plan == nil || !input.plan.routingEnabled() {
		return ctx, nil, false, nil, nil
	}
	routingAccountIDs := input.plan.routingAccountIDs

	if s.debugModelRoutingEnabled() {
		logger.LegacyPrintf("service.gateway", input.debugBeginLabel,
			derefGroupID(input.groupID),
			input.requestedModel,
			input.plan.platform,
			shortSessionHash(input.sessionHash),
			routingAccountIDs,
		)
	}

	if input.sessionHash != "" && s.cache != nil {
		accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(input.groupID), input.sessionHash)
		if err == nil && accountID > 0 && containsInt64(routingAccountIDs, accountID) {
			if account, ok := s.tryStickySessionAccount(ctx, input.groupID, input.sessionHash, input.requestedModel, input.excludedIDs, input.plan.stickyPlatformCheck); ok {
				if s.debugModelRoutingEnabled() {
					logger.LegacyPrintf("service.gateway", input.debugStickyHitLabel,
						derefGroupID(input.groupID),
						input.requestedModel,
						shortSessionHash(input.sessionHash),
						accountID,
					)
				}
				return ctx, nil, false, account, nil
			}
		}
	}

	accounts := s.loadSelectionAccountsByID(ctx, routingAccountIDs)
	ctx = s.prefetchSelectionSignals(ctx, accounts)
	filterParams := *input.plan.newCandidateFilterParams(ctx, input.requestedModel, input.excludedIDs, true)
	filterParams.ctx = ctx

	selected := s.selectBestCandidateWithStickyPolicy(
		ctx,
		input.groupID,
		input.sessionHash,
		accounts,
		&filterParams,
		input.plan.oauthTieBreaker,
		input.bindSticky,
	)
	if selected != nil {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf("service.gateway", input.debugSelectLabel,
				derefGroupID(input.groupID),
				input.requestedModel,
				shortSessionHash(input.sessionHash),
				selected.ID,
			)
		}
		return ctx, nil, false, selected, nil
	}

	logger.LegacyPrintf("service.gateway", "[ModelRouting] No routed accounts available for model=%s, falling back to normal selection", input.requestedModel)
	return ctx, nil, false, nil, nil
}
