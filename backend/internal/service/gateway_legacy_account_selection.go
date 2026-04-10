package service

import (
	"context"
	"fmt"
)

type legacyAccountSelectionInput struct {
	ctx                 context.Context
	groupID             *int64
	sessionHash         string
	requestedModel      string
	excludedIDs         map[int64]struct{}
	platform            string
	allowForcePlatform  bool
	debugBeginLabel     string
	debugStickyHitLabel string
	debugSelectLabel    string
	stickyPlatformCheck func(*Account) bool
	platformFilter      func(*Account) bool
	oauthTieBreaker     func(a, b *Account) bool
	failureUseMixed     bool
}

func (s *GatewayService) selectAccountWithLegacyScheduling(input *legacyAccountSelectionInput) (*Account, error) {
	if input == nil {
		return nil, ErrNoAvailableAccounts
	}

	routingAccountIDs := s.routingAccountIDsForRequest(input.ctx, input.groupID, input.requestedModel, input.platform)
	schedGroup := s.loadSchedulingGroup(input.ctx, input.groupID)

	var accounts []Account
	accountsLoaded := false

	ctx, accounts, accountsLoaded, selected, err := s.tryLegacyRoutedSelection(input.ctx, &legacyRoutedSelectionInput{
		groupID:             input.groupID,
		sessionHash:         input.sessionHash,
		requestedModel:      input.requestedModel,
		excludedIDs:         input.excludedIDs,
		routingAccountIDs:   routingAccountIDs,
		platform:            input.platform,
		allowForcePlatform:  input.allowForcePlatform,
		debugBeginLabel:     input.debugBeginLabel,
		debugStickyHitLabel: input.debugStickyHitLabel,
		debugSelectLabel:    input.debugSelectLabel,
		stickyPlatformCheck: input.stickyPlatformCheck,
		filterParams: &candidateFilterParams{
			ctx:            input.ctx,
			requestedModel: input.requestedModel,
			excludedIDs:    input.excludedIDs,
			routingSet:     buildRoutingSet(routingAccountIDs),
			schedGroup:     schedGroup,
			platformFilter: input.platformFilter,
		},
		oauthTieBreaker: input.oauthTieBreaker,
	})
	if err != nil {
		return nil, err
	}
	if selected != nil {
		return s.hydrateSelectedAccount(ctx, selected)
	}

	if account, ok := s.tryStickySessionAccount(ctx, input.groupID, input.sessionHash, input.requestedModel, input.excludedIDs, input.stickyPlatformCheck); ok {
		return account, nil
	}

	if !accountsLoaded {
		accounts, err = s.listAccountsForSelection(ctx, input.groupID, input.platform, input.allowForcePlatform)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
	}

	ctx = s.prefetchSelectionSignals(ctx, accounts)
	selected = s.selectBestCandidateAndBindSession(ctx, input.groupID, input.sessionHash, accounts, &candidateFilterParams{
		ctx:            ctx,
		requestedModel: input.requestedModel,
		excludedIDs:    input.excludedIDs,
		schedGroup:     schedGroup,
		platformFilter: input.platformFilter,
	}, input.oauthTieBreaker)
	if selected != nil {
		return s.hydrateSelectedAccount(ctx, selected)
	}

	stats := s.logDetailedSelectionFailure(ctx, input.groupID, input.sessionHash, input.requestedModel, input.platform, accounts, input.excludedIDs, input.failureUseMixed)
	if input.requestedModel != "" {
		return nil, fmt.Errorf("%w supporting model: %s (%s)", ErrNoAvailableAccounts, input.requestedModel, summarizeSelectionFailureStats(stats))
	}
	return nil, ErrNoAvailableAccounts
}
