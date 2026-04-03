package service

import (
	"context"
	"fmt"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type legacyRoutedSelectionInput struct {
	groupID             *int64
	sessionHash         string
	requestedModel      string
	excludedIDs         map[int64]struct{}
	routingAccountIDs   []int64
	platform            string
	allowForcePlatform  bool
	debugBeginLabel     string
	debugStickyHitLabel string
	debugSelectLabel    string
	stickyPlatformCheck func(*Account) bool
	filterParams        *candidateFilterParams
	oauthTieBreaker     func(a, b *Account) bool
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

func (s *GatewayService) tryLegacyRoutedSelection(
	ctx context.Context,
	input *legacyRoutedSelectionInput,
) (context.Context, []Account, bool, *Account, error) {
	if input == nil || len(input.routingAccountIDs) == 0 {
		return ctx, nil, false, nil, nil
	}

	if s.debugModelRoutingEnabled() {
		logger.LegacyPrintf("service.gateway", input.debugBeginLabel,
			derefGroupID(input.groupID),
			input.requestedModel,
			input.platform,
			shortSessionHash(input.sessionHash),
			input.routingAccountIDs,
		)
	}

	if input.sessionHash != "" && s.cache != nil {
		accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(input.groupID), input.sessionHash)
		if err == nil && accountID > 0 && containsInt64(input.routingAccountIDs, accountID) {
			if account, ok := s.tryStickySessionAccount(ctx, input.groupID, input.sessionHash, input.requestedModel, input.excludedIDs, input.stickyPlatformCheck); ok {
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

	accounts, err := s.listAccountsForSelection(ctx, input.groupID, input.platform, input.allowForcePlatform)
	if err != nil {
		return ctx, nil, false, nil, fmt.Errorf("query accounts failed: %w", err)
	}

	ctx = s.prefetchSelectionSignals(ctx, accounts)
	filterParams := *input.filterParams
	filterParams.ctx = ctx

	selected := s.selectBestCandidateAndBindSession(ctx, input.groupID, input.sessionHash, accounts, &filterParams, input.oauthTieBreaker)
	if selected != nil {
		if s.debugModelRoutingEnabled() {
			logger.LegacyPrintf("service.gateway", input.debugSelectLabel,
				derefGroupID(input.groupID),
				input.requestedModel,
				shortSessionHash(input.sessionHash),
				selected.ID,
			)
		}
		return ctx, accounts, true, selected, nil
	}

	logger.LegacyPrintf("service.gateway", "[ModelRouting] No routed accounts available for model=%s, falling back to normal selection", input.requestedModel)
	return ctx, accounts, true, nil, nil
}
