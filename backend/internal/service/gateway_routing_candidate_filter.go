package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type routedCandidateFilterStats struct {
	FilteredExcluded    int
	FilteredMissing     int
	FilteredUnsched     int
	FilteredPlatform    int
	FilteredModelScope  int
	FilteredModelMap    int
	FilteredWindowCost  int
	ModelScopeSkippedID []int64
}

type routedCandidateFilterResult struct {
	SelectionCtx context.Context
	Candidates   []*Account
	Stats        routedCandidateFilterStats
}

func (s *GatewayService) filterRoutedCandidates(
	ctx context.Context,
	accountByID map[int64]*Account,
	routingAccountIDs []int64,
	requestedModel string,
	platform string,
	useMixed bool,
	excludedIDs map[int64]struct{},
) routedCandidateFilterResult {
	result := routedCandidateFilterResult{
		SelectionCtx: ctx,
		Candidates:   make([]*Account, 0, len(routingAccountIDs)),
	}
	if len(routingAccountIDs) == 0 {
		return result
	}

	loadedByID := make(map[int64]*Account, len(routingAccountIDs))
	prefetchAccounts := make([]Account, 0, len(routingAccountIDs))

	for _, routingAccountID := range routingAccountIDs {
		if _, excluded := excludedIDs[routingAccountID]; excluded {
			result.Stats.FilteredExcluded++
			continue
		}

		account, ok := s.resolveSelectionAccountByID(ctx, accountByID, routingAccountID)
		if !ok {
			result.Stats.FilteredMissing++
			continue
		}
		if !s.isAccountSchedulableForSelection(account) {
			result.Stats.FilteredUnsched++
			continue
		}
		loadedByID[routingAccountID] = account
		prefetchAccounts = append(prefetchAccounts, *account)
	}

	result.SelectionCtx = s.prefetchSelectionSignals(ctx, prefetchAccounts)

	for _, routingAccountID := range routingAccountIDs {
		if _, excluded := excludedIDs[routingAccountID]; excluded {
			continue
		}

		account, ok := loadedByID[routingAccountID]
		if !ok {
			continue
		}

		if !s.isAccountAllowedForPlatform(account, platform, useMixed) {
			result.Stats.FilteredPlatform++
			continue
		}
		if requestedModel != "" && !s.isModelSupportedByAccountWithContext(result.SelectionCtx, account, requestedModel) {
			result.Stats.FilteredModelMap++
			continue
		}
		if !s.isAccountSchedulableForModelSelection(result.SelectionCtx, account, requestedModel) {
			result.Stats.FilteredModelScope++
			result.Stats.ModelScopeSkippedID = append(result.Stats.ModelScopeSkippedID, account.ID)
			continue
		}
		if !s.isAccountSchedulableForQuota(account) {
			continue
		}
		if !s.isAccountSchedulableForWindowCost(result.SelectionCtx, account, false) {
			result.Stats.FilteredWindowCost++
			continue
		}
		if !s.isAccountSchedulableForRPM(result.SelectionCtx, account, false) {
			continue
		}

		result.Candidates = append(result.Candidates, account)
	}

	return result
}

func (s *GatewayService) logRoutedCandidateFilterResult(
	groupID *int64,
	requestedModel string,
	routingAccountIDs []int64,
	result routedCandidateFilterResult,
) {
	if !s.debugModelRoutingEnabled() {
		return
	}

	logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] routed candidates: group_id=%v model=%s routed=%d candidates=%d filtered(excluded=%d missing=%d unsched=%d platform=%d model_scope=%d model_mapping=%d window_cost=%d)",
		derefGroupID(groupID), requestedModel, len(routingAccountIDs), len(result.Candidates),
		result.Stats.FilteredExcluded, result.Stats.FilteredMissing, result.Stats.FilteredUnsched,
		result.Stats.FilteredPlatform, result.Stats.FilteredModelScope, result.Stats.FilteredModelMap, result.Stats.FilteredWindowCost)
	if len(result.Stats.ModelScopeSkippedID) > 0 {
		logger.LegacyPrintf("service.gateway", "[ModelRoutingDebug] model_rate_limited accounts skipped: group_id=%v model=%s account_ids=%v",
			derefGroupID(groupID), requestedModel, result.Stats.ModelScopeSkippedID)
	}
}
