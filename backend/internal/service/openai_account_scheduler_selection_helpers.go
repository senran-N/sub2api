package service

import "context"

type openAILoadBalancePreparation struct {
	filtered                []*Account
	loadReq                 []AccountWithConcurrency
	requestedModelAvailable bool
}

func normalizeOpenAISchedulerTopK(configuredTopK int, candidateCount int) int {
	if configuredTopK > candidateCount {
		configuredTopK = candidateCount
	}
	if configuredTopK <= 0 {
		return 1
	}
	return configuredTopK
}

func selectionDecisionTopK(configuredTopK int, candidateCount int) int {
	if candidateCount <= 0 {
		return 0
	}
	return normalizeOpenAISchedulerTopK(configuredTopK, candidateCount)
}

func (s *defaultOpenAIAccountScheduler) prepareLoadBalanceCandidates(
	_ context.Context,
	req OpenAIAccountScheduleRequest,
	accounts []Account,
	schedGroup *Group,
) ([]*Account, []AccountWithConcurrency) {
	prepared := s.prepareLoadBalanceCandidatePage(req, accounts, schedGroup)
	return prepared.filtered, prepared.loadReq
}

func (s *defaultOpenAIAccountScheduler) prepareLoadBalanceCandidatePage(
	req OpenAIAccountScheduleRequest,
	accounts []Account,
	schedGroup *Group,
) openAILoadBalancePreparation {
	prepared := openAILoadBalancePreparation{
		filtered: make([]*Account, 0, len(accounts)),
		loadReq:  make([]AccountWithConcurrency, 0, len(accounts)),
	}

	for i := range accounts {
		account := &accounts[i]
		if !account.IsOpenAI() {
			continue
		}
		if req.RequestedModel == "" || isOpenAIAccountModelEligible(account, req.RequestedModel) {
			prepared.requestedModelAvailable = true
		}
		if isOpenAIAccountExcluded(req.ExcludedIDs, account.ID) {
			continue
		}
		if schedGroup != nil && schedGroup.RequirePrivacySet && !account.IsPrivacySet() {
			continue
		}
		if !s.isAccountTransportCompatible(account, req.RequiredTransport) {
			continue
		}
		if !isOpenAIAccountRuntimeEligible(account, req.RequestedModel) {
			continue
		}

		prepared.filtered = append(prepared.filtered, account)
		prepared.loadReq = append(prepared.loadReq, AccountWithConcurrency{
			ID:             account.ID,
			MaxConcurrency: account.EffectiveLoadFactor(),
		})
	}

	return prepared
}

func (s *defaultOpenAIAccountScheduler) prepareLoadBalanceCandidatePointers(
	req OpenAIAccountScheduleRequest,
	accounts []*Account,
	schedGroup *Group,
) openAILoadBalancePreparation {
	prepared := openAILoadBalancePreparation{
		filtered: make([]*Account, 0, len(accounts)),
		loadReq:  make([]AccountWithConcurrency, 0, len(accounts)),
	}

	for _, account := range accounts {
		if account == nil || !account.IsOpenAI() {
			continue
		}
		if req.RequestedModel == "" || isOpenAIAccountModelEligible(account, req.RequestedModel) {
			prepared.requestedModelAvailable = true
		}
		if isOpenAIAccountExcluded(req.ExcludedIDs, account.ID) {
			continue
		}
		if schedGroup != nil && schedGroup.RequirePrivacySet && !account.IsPrivacySet() {
			continue
		}
		if !s.isAccountTransportCompatible(account, req.RequiredTransport) {
			continue
		}
		if !isOpenAIAccountRuntimeEligible(account, req.RequestedModel) {
			continue
		}

		prepared.filtered = append(prepared.filtered, account)
		prepared.loadReq = append(prepared.loadReq, AccountWithConcurrency{
			ID:             account.ID,
			MaxConcurrency: account.EffectiveLoadFactor(),
		})
	}

	return prepared
}

func (s *defaultOpenAIAccountScheduler) buildIndexedSnapshotPager(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
	schedGroup *Group,
) (*schedulerIndexedAccountPager, error) {
	if s == nil || s.service == nil {
		return nil, nil
	}
	return s.service.buildOpenAIIndexedCandidatePager(ctx, openAIIndexedCandidateScope{
		groupID:           req.GroupID,
		requestedModel:    req.RequestedModel,
		requiredTransport: req.RequiredTransport,
		requirePrivacy:    schedGroup != nil && schedGroup.RequirePrivacySet,
	})
}

func (s *defaultOpenAIAccountScheduler) filterBatchByIndexedCapabilityPointers(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
	accounts []*Account,
	schedGroup *Group,
) []*Account {
	if len(accounts) == 0 || s == nil || s.service == nil {
		return accounts
	}
	return s.service.filterOpenAIBatchByIndexedCapabilitiesFromPointers(ctx, accounts, openAIIndexedCandidateScope{
		groupID:           req.GroupID,
		requestedModel:    req.RequestedModel,
		requiredTransport: req.RequiredTransport,
		requirePrivacy:    schedGroup != nil && schedGroup.RequirePrivacySet,
	})
}

func (s *defaultOpenAIAccountScheduler) loadSchedulerAccountLoads(
	ctx context.Context,
	loadReq []AccountWithConcurrency,
) map[int64]*AccountLoadInfo {
	loadMap := map[int64]*AccountLoadInfo{}
	if s.service.concurrencyService == nil {
		return loadMap
	}

	batchLoad, err := s.service.concurrencyService.GetAccountsLoadBatch(ctx, loadReq)
	if err == nil {
		return batchLoad
	}

	return loadMap
}

func (s *defaultOpenAIAccountScheduler) trySelectImmediateScheduledCandidate(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
	selectionOrder []openAIAccountCandidateScore,
) (*AccountSelectionResult, error, bool) {
	ordered := make([]*Account, 0, len(selectionOrder))
	for i := range selectionOrder {
		if selectionOrder[i].account != nil {
			ordered = append(ordered, selectionOrder[i].account)
		}
	}
	return s.service.tryAcquireImmediateOpenAISelectionFromOrderedAccounts(ctx, ordered, openAIImmediateSelectionRequest{
		groupID:             req.GroupID,
		sessionHash:         req.SessionHash,
		requestedModel:      req.RequestedModel,
		requiredTransport:   req.RequiredTransport,
		bindSticky:          true,
		recheckAfterAcquire: true,
	})
}

func (s *defaultOpenAIAccountScheduler) tryBuildScheduledWaitPlan(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
	candidates []openAIAccountCandidateScore,
) (*AccountSelectionResult, bool) {
	waitCandidate := chooseOpenAIWaitCandidate(candidates)
	if waitCandidate == nil {
		return nil, false
	}

	return s.service.tryBuildOpenAIWaitPlanSelection(
		ctx,
		waitCandidate.account,
		req.RequestedModel,
		req.RequiredTransport,
		gatewaySchedulingConfigOrDefault(s.service.cfg),
	)
}

func chooseBetterOpenAIWaitCandidate(
	best *openAIAccountCandidateScore,
	candidate *openAIAccountCandidateScore,
) *openAIAccountCandidateScore {
	if candidate == nil {
		return best
	}
	if best == nil {
		return candidate
	}
	selected := chooseOpenAIWaitCandidate([]openAIAccountCandidateScore{*best, *candidate})
	if selected == nil {
		return best
	}
	if selected.account != nil && candidate.account != nil && selected.account.ID == candidate.account.ID {
		return candidate
	}
	return best
}

func isBetterOpenAIWaitCandidate(
	candidate *openAIAccountCandidateScore,
	current *openAIAccountCandidateScore,
) bool {
	return chooseBetterOpenAIWaitCandidate(current, candidate) == candidate
}
