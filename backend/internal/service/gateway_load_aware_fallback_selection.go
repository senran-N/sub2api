package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/config"
)

type loadAwareFallbackSelectionInput struct {
	ctx              context.Context
	groupID          *int64
	sessionHash      string
	requestedModel   string
	stickyAccountID  int64
	excludedIDs      map[int64]struct{}
	accountByID      map[int64]*Account
	accounts         []Account
	plan             *gatewaySelectionPlan
	schedulingConfig config.GatewaySchedulingConfig
}

func (s *GatewayService) selectLoadAwareFallbackFlow(input *loadAwareFallbackSelectionInput) (*AccountSelectionResult, error) {
	if input == nil {
		return nil, ErrNoAvailableAccounts
	}
	if input.plan == nil {
		return nil, ErrNoAvailableAccounts
	}

	if !input.plan.routingEnabled() {
		if result, ok := s.trySelectStickyAccountWithoutRouting(
			input.ctx,
			input.groupID,
			input.sessionHash,
			input.requestedModel,
			input.stickyAccountID,
			input.excludedIDs,
			input.accountByID,
			input.plan.platform,
			input.plan.useMixed,
			input.schedulingConfig.StickySessionWaitTimeout,
			input.schedulingConfig.StickySessionMaxWaiting,
		); ok {
			return result, nil
		}
	}

	if s.shouldUseIndexedCandidateSource() {
		return s.selectLoadAwareFallbackFromPager(input)
	}

	candidates := s.filterLoadAwareCandidates(
		input.ctx,
		input.accounts,
		input.requestedModel,
		input.plan.platform,
		input.plan.useMixed,
		input.excludedIDs,
	)
	if len(candidates) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	loadMap, err := s.concurrencyService.GetAccountsLoadBatch(input.ctx, buildAccountLoadRequests(candidates))
	if err != nil {
		if result, ok := s.tryAcquireByLegacyOrder(input.ctx, candidates, input.groupID, input.sessionHash, input.plan.preferOAuth); ok {
			return result, nil
		}
	} else {
		if result, ok := s.selectLoadAwareAvailableAccount(
			input.ctx,
			buildAvailableAccountLoads(candidates, loadMap),
			input.groupID,
			input.sessionHash,
			input.plan.preferOAuth,
		); ok {
			return result, nil
		}
	}

	if result, ok := s.selectFallbackWaitPlan(
		input.ctx,
		candidates,
		input.sessionHash,
		input.plan.preferOAuth,
		input.schedulingConfig.FallbackSelectionMode,
		input.schedulingConfig.FallbackWaitTimeout,
		input.schedulingConfig.FallbackMaxWaiting,
	); ok {
		return result, nil
	}

	return nil, ErrNoAvailableAccounts
}

func (s *GatewayService) selectLoadAwareFallbackFromPager(input *loadAwareFallbackSelectionInput) (*AccountSelectionResult, error) {
	if input == nil {
		return nil, ErrNoAvailableAccounts
	}
	if input.plan == nil {
		return nil, ErrNoAvailableAccounts
	}

	schedGroup := input.plan.schedGroup
	if schedGroup == nil {
		schedGroup = s.loadSchedulingGroup(input.ctx, input.groupID)
	}
	pager, err := s.buildGatewayIndexedCandidatePager(
		input.ctx,
		input.groupID,
		input.plan.platform,
		input.plan.hasForcePlatform,
		input.requestedModel,
		schedGroup,
	)
	if err != nil {
		return nil, err
	}
	if pager == nil {
		return nil, ErrNoAvailableAccounts
	}

	pageSize := snapshotPageSizeOrDefault(s.cfg)
	var selectedResult *AccountSelectionResult
	modelScopedFound, bestWait, err := executeIndexedRuntimeSelection(
		input.ctx,
		pager,
		pageSize,
		func(batch []Account) (bool, *Account, error) {
			batch = s.filterSelectionBatchByIndexedCapabilities(input.ctx, input.groupID, pager.platform, pager.hasForcePlatform, batch, schedGroup)
			if len(batch) == 0 {
				return false, nil, nil
			}
			pageCtx := s.prefetchSelectionSignals(input.ctx, batch)

			candidates := s.filterLoadAwareCandidates(
				pageCtx,
				batch,
				input.requestedModel,
				input.plan.platform,
				input.plan.useMixed,
				input.excludedIDs,
			)
			if len(candidates) == 0 {
				return false, nil, nil
			}

			loadMap, loadErr := s.concurrencyService.GetAccountsLoadBatch(pageCtx, buildAccountLoadRequests(candidates))
			if loadErr != nil {
				ordered := append([]*Account(nil), candidates...)
				sortAccountsByPriorityAndLastUsed(ordered, input.plan.preferOAuth)
				if result, _, ok := selectFirstOrderedRuntimeSelection(ordered, func(account *Account) (*AccountSelectionResult, error, bool) {
					result, ok := s.tryAcquireAndMaybeBindSelection(pageCtx, input.groupID, input.sessionHash, account, true)
					return result, nil, ok
				}); ok {
					selectedResult = result
					return true, nil, nil
				}
				if len(ordered) > 0 {
					return false, ordered[0], nil
				}
				return false, nil, nil
			}

			available := buildAvailableAccountLoads(candidates, loadMap)
			if result, ok := s.selectLoadAwareAvailableAccount(
				pageCtx,
				available,
				input.groupID,
				input.sessionHash,
				input.plan.preferOAuth,
			); ok {
				selectedResult = result
				return true, nil, nil
			}
			if len(candidates) > 0 {
				ordered := append([]*Account(nil), candidates...)
				s.sortCandidatesForFallback(ordered, input.plan.preferOAuth, input.schedulingConfig.FallbackSelectionMode)
				if len(ordered) > 0 {
					return false, ordered[0], nil
				}
			}
			return false, nil, nil
		},
		func(candidate, current *Account) bool {
			return s.isBetterSelectionCandidate(candidate, current, func(a, b *Account) bool {
				return input.plan.preferOAuth && a.Type != b.Type && a.Type == AccountTypeOAuth
			})
		},
	)
	if err != nil {
		return nil, err
	}
	if selectedResult != nil {
		return selectedResult, nil
	}

	if bestWait != nil {
		if result, _, ok := s.tryBuildAccountWaitPlan(
			input.ctx,
			bestWait,
			input.sessionHash,
			input.schedulingConfig.FallbackWaitTimeout,
			input.schedulingConfig.FallbackMaxWaiting,
		); ok {
			return result, nil
		}
	}

	if input.requestedModel != "" && !modelScopedFound {
		return nil, ErrNoAvailableAccounts
	}
	return nil, ErrNoAvailableAccounts
}
