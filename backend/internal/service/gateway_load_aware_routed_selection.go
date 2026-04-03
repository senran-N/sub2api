package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type loadAwareRoutedSelectionInput struct {
	ctx               context.Context
	groupID           *int64
	sessionHash       string
	requestedModel    string
	stickyAccountID   int64
	excludedIDs       map[int64]struct{}
	accountByID       map[int64]*Account
	platform          string
	useMixed          bool
	routingAccountIDs []int64
	waitTimeout       time.Duration
	maxWaiting        int
}

func (s *GatewayService) trySelectLoadAwareRoutedAccount(input *loadAwareRoutedSelectionInput) (*AccountSelectionResult, bool) {
	if input == nil || len(input.routingAccountIDs) == 0 || s.concurrencyService == nil {
		return nil, false
	}

	routingFilter := s.filterRoutedCandidates(
		input.ctx,
		input.accountByID,
		input.routingAccountIDs,
		input.requestedModel,
		input.platform,
		input.useMixed,
		input.excludedIDs,
	)
	routingCandidates := routingFilter.Candidates
	s.logRoutedCandidateFilterResult(input.groupID, input.requestedModel, input.routingAccountIDs, routingFilter)
	if len(routingCandidates) == 0 {
		return nil, false
	}

	if result, ok := s.trySelectRoutedStickyAccount(
		input.ctx,
		input.groupID,
		input.sessionHash,
		input.requestedModel,
		input.stickyAccountID,
		input.routingAccountIDs,
		input.excludedIDs,
		input.accountByID,
		input.platform,
		input.useMixed,
		input.waitTimeout,
		input.maxWaiting,
	); ok {
		return result, true
	}

	routingLoadMap, _ := s.concurrencyService.GetAccountsLoadBatch(input.ctx, buildAccountLoadRequests(routingCandidates))
	if result, ok := s.selectAvailableRoutedAccount(
		input.ctx,
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
