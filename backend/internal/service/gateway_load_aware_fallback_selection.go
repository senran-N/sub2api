package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/config"
)

type loadAwareFallbackSelectionInput struct {
	ctx               context.Context
	groupID           *int64
	sessionHash       string
	requestedModel    string
	stickyAccountID   int64
	excludedIDs       map[int64]struct{}
	accountByID       map[int64]*Account
	platform          string
	useMixed          bool
	accounts          []Account
	preferOAuth       bool
	routingAccountIDs []int64
	schedulingConfig  config.GatewaySchedulingConfig
}

func (s *GatewayService) selectLoadAwareFallbackFlow(input *loadAwareFallbackSelectionInput) (*AccountSelectionResult, error) {
	if input == nil {
		return nil, ErrNoAvailableAccounts
	}

	if len(input.routingAccountIDs) == 0 {
		if result, ok := s.trySelectStickyAccountWithoutRouting(
			input.ctx,
			input.groupID,
			input.sessionHash,
			input.requestedModel,
			input.stickyAccountID,
			input.excludedIDs,
			input.accountByID,
			input.platform,
			input.useMixed,
			input.schedulingConfig.StickySessionWaitTimeout,
			input.schedulingConfig.StickySessionMaxWaiting,
		); ok {
			return result, nil
		}
	}

	candidates := s.filterLoadAwareCandidates(
		input.ctx,
		input.accounts,
		input.requestedModel,
		input.platform,
		input.useMixed,
		input.excludedIDs,
	)
	if len(candidates) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	loadMap, err := s.concurrencyService.GetAccountsLoadBatch(input.ctx, buildAccountLoadRequests(candidates))
	if err != nil {
		if result, ok := s.tryAcquireByLegacyOrder(input.ctx, candidates, input.groupID, input.sessionHash, input.preferOAuth); ok {
			return result, nil
		}
	} else {
		if result, ok := s.selectLoadAwareAvailableAccount(
			input.ctx,
			buildAvailableAccountLoads(candidates, loadMap),
			input.groupID,
			input.sessionHash,
			input.preferOAuth,
		); ok {
			return result, nil
		}
	}

	if result, ok := s.selectFallbackWaitPlan(
		input.ctx,
		candidates,
		input.sessionHash,
		input.preferOAuth,
		input.schedulingConfig.FallbackSelectionMode,
		input.schedulingConfig.FallbackWaitTimeout,
		input.schedulingConfig.FallbackMaxWaiting,
	); ok {
		return result, nil
	}

	return nil, ErrNoAvailableAccounts
}
