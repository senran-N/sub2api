package service

import "context"

func normalizeOpenAISchedulerTopK(configuredTopK int, candidateCount int) int {
	if configuredTopK > candidateCount {
		configuredTopK = candidateCount
	}
	if configuredTopK <= 0 {
		return 1
	}
	return configuredTopK
}

func (s *defaultOpenAIAccountScheduler) prepareLoadBalanceCandidates(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
	accounts []Account,
	schedGroup *Group,
) ([]*Account, []AccountWithConcurrency) {
	filtered := make([]*Account, 0, len(accounts))
	loadReq := make([]AccountWithConcurrency, 0, len(accounts))

	for i := range accounts {
		account := &accounts[i]
		if isOpenAIAccountExcluded(req.ExcludedIDs, account.ID) {
			continue
		}
		if !account.IsSchedulable() || !account.IsOpenAI() {
			continue
		}
		if schedGroup != nil && schedGroup.RequirePrivacySet && !account.IsPrivacySet() {
			_ = s.service.accountRepo.SetError(
				ctx,
				account.ID,
				"Privacy not set, required by group ["+schedGroup.Name+"]",
			)
			continue
		}
		if req.RequestedModel != "" && !account.IsModelSupported(req.RequestedModel) {
			continue
		}
		if !s.isAccountTransportCompatible(account, req.RequiredTransport) {
			continue
		}

		filtered = append(filtered, account)
		loadReq = append(loadReq, AccountWithConcurrency{
			ID:             account.ID,
			MaxConcurrency: account.EffectiveLoadFactor(),
		})
	}

	return filtered, loadReq
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
	for i := 0; i < len(selectionOrder); i++ {
		candidate := selectionOrder[i]
		fresh := s.service.resolveFreshSchedulableOpenAIAccount(ctx, candidate.account, req.RequestedModel)
		if fresh == nil || !s.isAccountTransportCompatible(fresh, req.RequiredTransport) {
			continue
		}

		result, acquireErr := s.service.tryAcquireAccountSlot(ctx, fresh.ID, fresh.Concurrency)
		if acquireErr != nil {
			return nil, acquireErr, true
		}
		if result == nil || !result.Acquired {
			continue
		}

		verified := s.service.recheckSelectedOpenAIAccountFromDB(ctx, fresh, req.RequestedModel)
		if verified == nil || !s.isAccountTransportCompatible(verified, req.RequiredTransport) {
			if result.ReleaseFunc != nil {
				result.ReleaseFunc()
			}
			continue
		}
		if req.SessionHash != "" {
			_ = s.service.BindStickySession(ctx, req.GroupID, req.SessionHash, verified.ID)
		}
		return newAcquiredAccountSelection(verified, result.ReleaseFunc), nil, true
	}

	return nil, nil, false
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

	cfg := s.service.schedulingConfig()
	fresh := s.service.resolveFreshSchedulableOpenAIAccount(ctx, waitCandidate.account, req.RequestedModel)
	if fresh == nil || !s.isAccountTransportCompatible(fresh, req.RequiredTransport) {
		return nil, false
	}

	return newWaitPlanAccountSelection(fresh, cfg.FallbackWaitTimeout, cfg.FallbackMaxWaiting), true
}
