package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type loadAwareRoutedSelectionInput struct {
	ctx             context.Context
	groupID         *int64
	sessionHash     string
	requestedModel  string
	stickyAccountID int64
	excludedIDs     map[int64]struct{}
	accountByID     map[int64]*Account
	plan            *gatewaySelectionPlan
	waitTimeout     time.Duration
	maxWaiting      int
}

func (s *GatewayService) trySelectLoadAwareRoutedAccount(input *loadAwareRoutedSelectionInput) (*AccountSelectionResult, bool) {
	if input == nil || input.plan == nil || !input.plan.routingEnabled() || s.concurrencyService == nil {
		return nil, false
	}
	routingAccountIDs := input.plan.routingAccountIDs

	routingFilter := s.filterRoutedCandidates(
		input.ctx,
		input.accountByID,
		routingAccountIDs,
		input.requestedModel,
		input.plan.platform,
		input.plan.useMixed,
		input.excludedIDs,
	)
	routingCandidates := routingFilter.Candidates
	s.logRoutedCandidateFilterResult(input.groupID, input.requestedModel, routingAccountIDs, routingFilter)
	if len(routingCandidates) == 0 {
		return nil, false
	}
	routingCtx := routingFilter.SelectionCtx

	if result, ok := s.trySelectRoutedStickyAccount(
		routingCtx,
		input.groupID,
		input.sessionHash,
		input.requestedModel,
		input.stickyAccountID,
		routingAccountIDs,
		input.excludedIDs,
		input.accountByID,
		input.plan.platform,
		input.plan.useMixed,
		input.waitTimeout,
		input.maxWaiting,
	); ok {
		return result, true
	}

	routingLoadMap, _ := s.concurrencyService.GetAccountsLoadBatch(routingCtx, buildAccountLoadRequests(routingCandidates))
	if result, ok := s.selectAvailableRoutedAccount(
		routingCtx,
		input.groupID,
		input.sessionHash,
		input.requestedModel,
		buildAvailableAccountLoads(routingCandidates, routingLoadMap),
		input.waitTimeout,
		input.maxWaiting,
	); ok {
		return result, true
	}

	logger.LegacyPrintf("service.gateway", "[ModelRouting] All routed accounts unavailable for model=%s, falling back to normal selection", input.requestedModel)
	return nil, false
}
