package service

import (
	"context"
	"fmt"
)

type legacyAccountSelectionInput struct {
	ctx                   context.Context
	groupID               *int64
	sessionHash           string
	requestedModel        string
	excludedIDs           map[int64]struct{}
	plan                  *gatewaySelectionPlan
	debugBeginLabel       string
	debugStickyHitLabel   string
	debugSelectLabel      string
	bindStickyOnSelection bool
}

func (s *GatewayService) selectAccountWithLegacyScheduling(input *legacyAccountSelectionInput) (*Account, error) {
	if input == nil {
		return nil, ErrNoAvailableAccounts
	}

	if input.plan == nil {
		return nil, ErrNoAvailableAccounts
	}
	schedGroup := input.plan.schedGroup

	var accounts []Account
	accountsLoaded := false

	ctx, accounts, accountsLoaded, selected, err := s.tryLegacyRoutedSelection(input.ctx, &legacyRoutedSelectionInput{
		groupID:             input.groupID,
		sessionHash:         input.sessionHash,
		requestedModel:      input.requestedModel,
		excludedIDs:         input.excludedIDs,
		plan:                input.plan,
		debugBeginLabel:     input.debugBeginLabel,
		debugStickyHitLabel: input.debugStickyHitLabel,
		debugSelectLabel:    input.debugSelectLabel,
		bindSticky:          input.bindStickyOnSelection,
	})
	if err != nil {
		return nil, err
	}
	if selected != nil {
		return s.hydrateSelectedAccount(ctx, selected)
	}

	if account, ok := s.tryStickySessionAccount(ctx, input.groupID, input.sessionHash, input.requestedModel, input.excludedIDs, input.plan.stickyPlatformCheck); ok {
		return account, nil
	}

	if !accountsLoaded {
		if s.shouldUseIndexedCandidateSource() {
			pager, pagerErr := s.buildGatewayIndexedCandidatePager(ctx, input.groupID, input.plan.platform, input.plan.hasForcePlatform, input.requestedModel, schedGroup)
			if pagerErr != nil {
				return nil, fmt.Errorf("query accounts failed: %w", pagerErr)
			}
			pageSize := snapshotPageSizeOrDefault(s.cfg)
			selected, modelScopedFound, pagerErr := s.selectBestCandidateFromPagerWithStickyPolicy(
				ctx,
				input.groupID,
				input.sessionHash,
				pager,
				input.plan.newCandidateFilterParams(ctx, input.requestedModel, input.excludedIDs, false),
				input.plan.oauthTieBreaker,
				pageSize,
				input.bindStickyOnSelection,
			)
			if pagerErr != nil {
				return nil, fmt.Errorf("query accounts failed: %w", pagerErr)
			}
			if selected != nil {
				return s.hydrateSelectedAccount(ctx, selected)
			}
			if input.requestedModel != "" && !modelScopedFound {
				return nil, fmt.Errorf("%w supporting model: %s", ErrNoAvailableAccounts, input.requestedModel)
			}
			diagnosticPager, pagerErr := s.buildGatewayIndexedCandidatePager(ctx, input.groupID, input.plan.platform, input.plan.hasForcePlatform, input.requestedModel, schedGroup)
			if pagerErr != nil {
				return nil, fmt.Errorf("query accounts failed: %w", pagerErr)
			}
			stats, pagerErr := s.logDetailedSelectionFailureFromPager(
				ctx,
				input.groupID,
				input.sessionHash,
				input.requestedModel,
				input.plan.platform,
				input.excludedIDs,
				input.plan.failureUseMixed,
				schedGroup,
				diagnosticPager,
				pageSize,
			)
			if pagerErr != nil {
				return nil, fmt.Errorf("query accounts failed: %w", pagerErr)
			}
			if input.requestedModel != "" {
				return nil, fmt.Errorf("%w supporting model: %s (%s)", ErrNoAvailableAccounts, input.requestedModel, summarizeSelectionFailureStats(stats))
			}
			return nil, ErrNoAvailableAccounts
		}

		accounts, err = s.listAccountsForSelection(ctx, input.groupID, input.plan.platform, input.plan.hasForcePlatform)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
	}

	ctx = s.prefetchSelectionSignals(ctx, accounts)
	selected = s.selectBestCandidateWithStickyPolicy(
		ctx,
		input.groupID,
		input.sessionHash,
		accounts,
		input.plan.newCandidateFilterParams(ctx, input.requestedModel, input.excludedIDs, false),
		input.plan.oauthTieBreaker,
		input.bindStickyOnSelection,
	)
	if selected != nil {
		return s.hydrateSelectedAccount(ctx, selected)
	}

	stats := s.logDetailedSelectionFailure(ctx, input.groupID, input.sessionHash, input.requestedModel, input.plan.platform, accounts, input.excludedIDs, input.plan.failureUseMixed)
	if input.requestedModel != "" {
		return nil, fmt.Errorf("%w supporting model: %s (%s)", ErrNoAvailableAccounts, input.requestedModel, summarizeSelectionFailureStats(stats))
	}
	return nil, ErrNoAvailableAccounts
}
