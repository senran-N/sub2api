package service

import (
	"context"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

func isOpenAIAccountBaseEligible(account *Account) bool {
	return account != nil && account.IsOpenAI() && account.IsSchedulable()
}

func isOpenAIAccountRuntimeEligible(account *Account, requestedModel string) bool {
	if account == nil || !account.IsOpenAI() {
		return false
	}
	if !account.IsSchedulable() {
		return false
	}
	if oauthSelectionCredentialIssue(account) != "" {
		return false
	}
	if requestedModel != "" && !isOpenAIAccountModelEligible(account, requestedModel) {
		return false
	}
	return true
}

func isOpenAIAccountModelEligible(account *Account, requestedModel string) bool {
	if requestedModel == "" {
		return true
	}
	if account == nil {
		return false
	}
	if account.IsModelSupported(requestedModel) {
		return true
	}
	mappedModel, matched := resolveMappedModelWithOpenAIReasoningFallback(account, requestedModel)
	return matched && strings.TrimSpace(mappedModel) != ""
}

// isOpenAIAccountExcluded reports whether the account is in exclusion set.
func isOpenAIAccountExcluded(excludedIDs map[int64]struct{}, accountID int64) bool {
	if excludedIDs == nil {
		return false
	}
	_, excluded := excludedIDs[accountID]
	return excluded
}

// filterSchedulableOpenAICandidates filters accounts that can participate in OpenAI scheduling.
func filterSchedulableOpenAICandidates(
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
) []*Account {
	candidates := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if isOpenAIAccountExcluded(excludedIDs, account.ID) {
			continue
		}
		if !isOpenAIAccountRuntimeEligible(account, requestedModel) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

func filterSchedulableOpenAIAccountPointers(
	accounts []*Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
) []*Account {
	candidates := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		if isOpenAIAccountExcluded(excludedIDs, account.ID) {
			continue
		}
		if !isOpenAIAccountRuntimeEligible(account, requestedModel) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

// normalizeOpenAIWaitLoadMap converts nil load map to an empty map for wait-plan scoring.
func normalizeOpenAIWaitLoadMap(loadMap map[int64]*AccountLoadInfo) map[int64]*AccountLoadInfo {
	if loadMap != nil {
		return loadMap
	}
	return map[int64]*AccountLoadInfo{}
}

// accountLoadInfoOrDefault returns the load info for account or a zero-value fallback.
func accountLoadInfoOrDefault(loadMap map[int64]*AccountLoadInfo, accountID int64) *AccountLoadInfo {
	if loadMap != nil {
		if info := loadMap[accountID]; info != nil {
			return info
		}
	}
	return &AccountLoadInfo{AccountID: accountID}
}

// shouldPreferOpenAIWaitCandidate applies wait-plan candidate ordering:
// waiting count -> load rate -> priority -> LRU.
func shouldPreferOpenAIWaitCandidate(
	current *Account,
	currentLoad *AccountLoadInfo,
	best *Account,
	bestLoad *AccountLoadInfo,
) bool {
	if currentLoad.WaitingCount != bestLoad.WaitingCount {
		return currentLoad.WaitingCount < bestLoad.WaitingCount
	}
	if currentLoad.LoadRate != bestLoad.LoadRate {
		return currentLoad.LoadRate < bestLoad.LoadRate
	}
	if current.Priority != best.Priority {
		return current.Priority < best.Priority
	}
	switch {
	case current.LastUsedAt == nil && best.LastUsedAt != nil:
		return true
	case current.LastUsedAt != nil && best.LastUsedAt == nil:
		return false
	case current.LastUsedAt == nil && best.LastUsedAt == nil:
		return false
	default:
		return current.LastUsedAt.Before(*best.LastUsedAt)
	}
}

// selectBestOpenAIWaitCandidate picks the best account for fallback wait plan.
func selectBestOpenAIWaitCandidate(candidates []*Account, loadMap map[int64]*AccountLoadInfo) *Account {
	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	bestLoad := accountLoadInfoOrDefault(loadMap, best.ID)
	for i := 1; i < len(candidates); i++ {
		current := candidates[i]
		currentLoad := accountLoadInfoOrDefault(loadMap, current.ID)
		if shouldPreferOpenAIWaitCandidate(current, currentLoad, best, bestLoad) {
			best = current
			bestLoad = currentLoad
		}
	}

	return best
}

type openAIStickySessionResolvePolicy struct {
	deleteOnLookupMiss bool
	refreshTTLOnHit    bool
	recheckOnResolve   bool
}

type openAIIndexedCandidateScope struct {
	groupID           *int64
	requestedModel    string
	requiredTransport OpenAIUpstreamTransport
	requirePrivacy    bool
}

type openAIImmediateSelectionRequest struct {
	groupID             *int64
	sessionHash         string
	requestedModel      string
	requiredTransport   OpenAIUpstreamTransport
	bindSticky          bool
	recheckAfterAcquire bool
}

type openAIWaitCandidate struct {
	account  *Account
	loadInfo *AccountLoadInfo
}

type openAIStickyResolvedSelectionSpec struct {
	account        *Account
	accountID      int64
	cfg            config.GatewaySchedulingConfig
	finalize       func(account *Account) *Account
	onSelected     func(account *Account, acquired *AcquireResult) *AccountSelectionResult
	onFinalizeMiss func()
	stickyWaitPlan func(account *Account) (*AccountSelectionResult, bool)
}

func (s *OpenAIGatewayService) resolveOpenAIStickySessionAccount(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	stickyAccountID int64,
	policy openAIStickySessionResolvePolicy,
) (*Account, int64) {
	if sessionHash == "" {
		return nil, 0
	}

	accountID := stickyAccountID
	if accountID <= 0 {
		var err error
		accountID, err = s.getStickySessionAccountID(ctx, groupID, sessionHash)
		if err != nil || accountID <= 0 {
			return nil, 0
		}
	}

	if isOpenAIAccountExcluded(excludedIDs, accountID) {
		return nil, 0
	}

	account, err := s.getSchedulableAccount(ctx, accountID)
	if err != nil {
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingSoftMiss("lookup_error"), accountID, sessionHash, "")
		return nil, 0
	}
	if account == nil {
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingSoftMiss("lookup_miss"), accountID, sessionHash, "")
		return nil, 0
	}

	disposition := classifyStickyBindingDisposition(account, requestedModel)
	switch disposition.Outcome {
	case stickyBindingOutcomeHardInvalidate:
		_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
		return nil, 0
	case stickyBindingOutcomeSoftMiss:
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
		return nil, 0
	}
	if !isOpenAIAccountBaseEligible(account) {
		disposition = newStickyBindingHardInvalidate("platform_mismatch")
		_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
		return nil, 0
	}
	if !isOpenAIAccountModelEligible(account, requestedModel) {
		disposition = newStickyBindingHardInvalidate("model_unsupported")
		_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
		return nil, 0
	}

	if policy.recheckOnResolve {
		account, disposition = s.recheckStickyBoundOpenAIAccountFromDB(ctx, account, requestedModel)
		if account == nil {
			if disposition.Outcome == stickyBindingOutcomeUsable {
				disposition = newStickyBindingSoftMiss("db_recheck_miss")
			}
			if disposition.Outcome == stickyBindingOutcomeHardInvalidate {
				_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
			}
			recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
			return nil, 0
		}
	}

	if policy.refreshTTLOnHit {
		_ = s.refreshStickySessionTTL(ctx, groupID, sessionHash, s.openAIWSSessionStickyTTL())
	}

	return account, accountID
}

func (s *OpenAIGatewayService) recheckStickyBoundOpenAIAccountFromDB(
	ctx context.Context,
	account *Account,
	requestedModel string,
) (*Account, stickyBindingDisposition) {
	if account == nil {
		return nil, newStickyBindingSoftMiss("db_recheck_miss")
	}
	if s == nil || s.schedulerSnapshot == nil || s.accountRepo == nil {
		return account, stickyBindingDisposition{}
	}

	latest, err := s.accountRepo.GetByID(ctx, account.ID)
	if err != nil || latest == nil {
		return nil, newStickyBindingSoftMiss("db_recheck_miss")
	}

	disposition := classifyStickyBindingDisposition(latest, requestedModel)
	if disposition.Outcome != stickyBindingOutcomeUsable {
		return nil, disposition
	}
	if !isOpenAIAccountBaseEligible(latest) {
		return nil, newStickyBindingHardInvalidate("platform_mismatch")
	}
	if !isOpenAIAccountModelEligible(latest, requestedModel) {
		return nil, newStickyBindingHardInvalidate("model_unsupported")
	}
	return latest, stickyBindingDisposition{}
}

func (s *OpenAIGatewayService) buildOpenAIIndexedCandidatePager(
	ctx context.Context,
	scope openAIIndexedCandidateScope,
) (*schedulerIndexedAccountPager, error) {
	if s == nil || s.schedulerSnapshot == nil {
		return nil, nil
	}

	sources, err := buildRequestedModelCapabilitySources(
		ctx,
		s.schedulerSnapshot,
		scope.groupID,
		PlatformOpenAI,
		false,
		scope.requestedModel,
	)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		sources = []SchedulerCapabilityIndex{{Kind: SchedulerCapabilityIndexAll}}
		if scope.requiredTransport != OpenAIUpstreamTransportAny && scope.requiredTransport != OpenAIUpstreamTransportHTTPSSE {
			sources = []SchedulerCapabilityIndex{{Kind: SchedulerCapabilityIndexOpenAIWS}}
		} else if scope.requirePrivacy {
			sources = []SchedulerCapabilityIndex{{Kind: SchedulerCapabilityIndexPrivacySet}}
		}
	}
	return newSchedulerIndexedAccountPager(s.schedulerSnapshot, scope.groupID, PlatformOpenAI, false, sources), nil
}

func (s *OpenAIGatewayService) filterOpenAIBatchBySnapshotMembershipFromPointers(
	ctx context.Context,
	groupID *int64,
	accounts []*Account,
	index SchedulerCapabilityIndex,
) []*Account {
	if len(accounts) == 0 || s == nil || s.schedulerSnapshot == nil {
		return accounts
	}
	accountIDs := make([]int64, 0, len(accounts))
	for i := range accounts {
		if accounts[i] != nil {
			accountIDs = append(accountIDs, accounts[i].ID)
		}
	}
	matches, _, err := s.schedulerSnapshot.MatchSchedulableAccountsCapability(ctx, groupID, PlatformOpenAI, false, index, accountIDs)
	if err != nil {
		return accounts
	}
	filtered := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		if account != nil && matches[account.ID] {
			filtered = append(filtered, account)
		}
	}
	return filtered
}

func (s *OpenAIGatewayService) filterOpenAIBatchByIndexedCapabilitiesFromPointers(
	ctx context.Context,
	accounts []*Account,
	scope openAIIndexedCandidateScope,
) []*Account {
	if len(accounts) == 0 || s == nil || s.schedulerSnapshot == nil {
		return accounts
	}

	filtered := accounts
	if scope.requirePrivacy {
		filtered = s.filterOpenAIBatchBySnapshotMembershipFromPointers(ctx, scope.groupID, filtered, SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexPrivacySet})
	}
	if scope.requiredTransport != OpenAIUpstreamTransportAny && scope.requiredTransport != OpenAIUpstreamTransportHTTPSSE {
		filtered = s.filterOpenAIBatchBySnapshotMembershipFromPointers(ctx, scope.groupID, filtered, SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexOpenAIWS})
	}
	return filtered
}

func (s *OpenAIGatewayService) selectBestAccountFromIndexedSnapshot(
	ctx context.Context,
	groupID *int64,
	requestedModel string,
	excludedIDs map[int64]struct{},
) (*Account, bool, error) {
	pager, err := s.buildOpenAIIndexedCandidatePager(ctx, openAIIndexedCandidateScope{
		groupID:           groupID,
		requestedModel:    requestedModel,
		requiredTransport: OpenAIUpstreamTransportAny,
	})
	if err != nil {
		return nil, false, err
	}
	if pager == nil {
		return nil, false, nil
	}

	supported := strings.TrimSpace(requestedModel) == ""
	selected, supported, err := selectBestAccountFromIndexedSnapshotPager(
		ctx,
		pager,
		snapshotPageSizeOrDefault(s.cfg),
		supported,
		func(batch []*Account) (*Account, error) {
			candidates := filterSchedulableOpenAIAccountPointers(batch, requestedModel, excludedIDs)
			if len(candidates) == 0 {
				return nil, nil
			}

			ordered := append([]*Account(nil), candidates...)
			sortAccountsByPriorityAndLastUsed(ordered, false)
			for _, candidate := range ordered {
				fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, candidate, requestedModel)
				if fresh == nil {
					continue
				}
				fresh = s.recheckSelectedOpenAIAccountFromDB(ctx, fresh, requestedModel)
				if fresh == nil {
					continue
				}
				return fresh, nil
			}
			return nil, nil
		},
		isOpenAISelectionCandidateBetter,
	)
	if err != nil {
		return nil, supported, err
	}
	return selected, supported, nil
}

func isOpenAISelectionCandidateBetter(candidate *Account, current *Account) bool {
	if candidate == nil {
		return false
	}
	if current == nil {
		return true
	}
	if candidate.Priority != current.Priority {
		return candidate.Priority < current.Priority
	}
	switch {
	case candidate.LastUsedAt == nil && current.LastUsedAt != nil:
		return true
	case candidate.LastUsedAt != nil && current.LastUsedAt == nil:
		return false
	case candidate.LastUsedAt == nil && current.LastUsedAt == nil:
		return false
	default:
		return candidate.LastUsedAt.Before(*current.LastUsedAt)
	}
}

func (s *OpenAIGatewayService) isAccountTransportCompatible(account *Account, requiredTransport OpenAIUpstreamTransport) bool {
	if requiredTransport == OpenAIUpstreamTransportAny || requiredTransport == OpenAIUpstreamTransportHTTPSSE {
		return true
	}
	if s == nil || account == nil {
		return false
	}
	return s.getOpenAIWSProtocolResolver().Resolve(account).Transport == requiredTransport
}

func shouldAvoidOpenAIWSFallbackCooling(requiredTransport OpenAIUpstreamTransport) bool {
	switch requiredTransport {
	case OpenAIUpstreamTransportResponsesWebsocket, OpenAIUpstreamTransportResponsesWebsocketV2:
		return true
	default:
		return false
	}
}

func (s *OpenAIGatewayService) isOpenAITransportFallbackCooling(accountID int64, requiredTransport OpenAIUpstreamTransport) bool {
	if !shouldAvoidOpenAIWSFallbackCooling(requiredTransport) {
		return false
	}
	return s.isOpenAIWSFallbackCooling(accountID)
}

func (s *OpenAIGatewayService) tryAcquireImmediateOpenAISelection(
	ctx context.Context,
	account *Account,
	req openAIImmediateSelectionRequest,
) (*AccountSelectionResult, error, bool) {
	return tryAcquireRuntimeSelection(ctx, runtimeAcquireSelectionSpec{
		account: account,
		prepare: func(account *Account) *Account {
			fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, account, req.requestedModel)
			if fresh == nil || !s.isAccountTransportCompatible(fresh, req.requiredTransport) {
				return nil
			}
			return fresh
		},
		acquire: func(account *Account) (*AcquireResult, error) {
			return acquireAccountSlotWithConcurrencyService(ctx, s.concurrencyService, account.ID, account.Concurrency)
		},
		finalize: func(account *Account) *Account {
			verified := account
			if req.recheckAfterAcquire {
				verified = s.recheckSelectedOpenAIAccountFromDB(ctx, account, req.requestedModel)
			}
			if verified == nil || !s.isAccountTransportCompatible(verified, req.requiredTransport) {
				return nil
			}
			return verified
		},
		bind: func(account *Account) {
			if req.bindSticky && req.sessionHash != "" {
				_ = s.BindStickySessionIfUnbound(ctx, req.groupID, req.sessionHash, account.ID)
			}
		},
	})
}

func (s *OpenAIGatewayService) trySelectResolvedOpenAIStickyAccount(
	ctx context.Context,
	spec openAIStickyResolvedSelectionSpec,
) (*AccountSelectionResult, bool) {
	if s == nil || spec.account == nil {
		return nil, false
	}

	result, _, ok := trySelectStickyRuntimeSelection(stickyRuntimeSelectionSpec{
		tryAcquire: func() (*AccountSelectionResult, string, bool) {
			result, missReason, acquireErr, ok := tryAcquireRuntimeSelectionDetailed(ctx, runtimeAcquireSelectionSpec{
				account: spec.account,
				acquire: func(account *Account) (*AcquireResult, error) {
					return acquireAccountSlotWithConcurrencyService(ctx, s.concurrencyService, spec.accountID, account.Concurrency)
				},
				finalize: spec.finalize,
				onAcquired: func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
					if spec.onSelected != nil {
						return spec.onSelected(account, acquired)
					}
					return newOpenAIAcquiredSelection(account, acquired)
				},
			})
			if acquireErr != nil {
				return nil, missReason, false
			}
			return result, missReason, ok
		},
		onAcquireMiss: func(reason string) {
			if reason == runtimeAcquireMissFinalize && spec.onFinalizeMiss != nil {
				spec.onFinalizeMiss()
			}
		},
		buildWaitPlan: func() (*AccountSelectionResult, string, bool) {
			waitBuilder := spec.stickyWaitPlan
			if waitBuilder == nil {
				waitBuilder = func(account *Account) (*AccountSelectionResult, bool) {
					return tryBuildStickySessionWaitPlan(ctx, account, spec.cfg, s.concurrencyService)
				}
			}
			result, ok := waitBuilder(spec.account)
			return result, "", ok
		},
	})
	return result, ok
}

func (s *OpenAIGatewayService) tryAcquireImmediateOpenAISelectionFromOrderedAccounts(
	ctx context.Context,
	ordered []*Account,
	req openAIImmediateSelectionRequest,
) (*AccountSelectionResult, error, bool) {
	return selectFirstOrderedRuntimeSelection(ordered, func(account *Account) (*AccountSelectionResult, error, bool) {
		result, err, ok := s.tryAcquireImmediateOpenAISelection(ctx, account, req)
		if err != nil {
			return nil, err, false
		}
		return result, nil, ok
	})
}

func (s *OpenAIGatewayService) tryBuildOpenAIWaitPlanSelection(
	ctx context.Context,
	account *Account,
	requestedModel string,
	requiredTransport OpenAIUpstreamTransport,
	cfg config.GatewaySchedulingConfig,
) (*AccountSelectionResult, bool) {
	result, _, ok := tryBuildRuntimeWaitPlan(ctx, runtimeWaitPlanSpec{
		account: account,
		prepare: func(account *Account) *Account {
			fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, account, requestedModel)
			if fresh == nil || !s.isAccountTransportCompatible(fresh, requiredTransport) {
				return nil
			}
			return fresh
		},
		timeout:    cfg.FallbackWaitTimeout,
		maxWaiting: cfg.FallbackMaxWaiting,
	})
	return result, ok
}

func (s *OpenAIGatewayService) selectOpenAIAccountWithoutLoadBatch(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	stickyAccountID int64,
	cfg config.GatewaySchedulingConfig,
) (*AccountSelectionResult, error) {
	account, err := s.selectAccountForModelWithExclusions(
		ctx,
		groupID,
		sessionHash,
		requestedModel,
		excludedIDs,
		stickyAccountID,
	)
	if err != nil {
		return nil, err
	}

	if result, err, ok := tryAcquireRuntimeSelection(ctx, runtimeAcquireSelectionSpec{
		account: account,
		acquire: func(account *Account) (*AcquireResult, error) {
			return acquireAccountSlotWithConcurrencyService(ctx, s.concurrencyService, account.ID, account.Concurrency)
		},
	}); err == nil && ok {
		return result, nil
	}

	return buildStickyAwareFallbackWaitPlan(ctx, account, stickyAccountID, cfg, s.concurrencyService), nil
}

func (s *OpenAIGatewayService) trySelectOpenAIStickyLoadAwareAccount(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	stickyAccountID int64,
	stickyTTL time.Duration,
	cfg config.GatewaySchedulingConfig,
) (*AccountSelectionResult, bool) {
	if sessionHash == "" {
		return nil, false
	}

	account, accountID := s.resolveOpenAIStickySessionAccount(
		ctx,
		groupID,
		sessionHash,
		requestedModel,
		excludedIDs,
		stickyAccountID,
		openAIStickySessionResolvePolicy{
			deleteOnLookupMiss: false,
			refreshTTLOnHit:    false,
			recheckOnResolve:   false,
		},
	)
	if account == nil {
		return nil, false
	}

	finalizeDisposition := stickyBindingDisposition{}
	return s.trySelectResolvedOpenAIStickyAccount(ctx, openAIStickyResolvedSelectionSpec{
		account:   account,
		accountID: accountID,
		cfg:       cfg,
		finalize: func(account *Account) *Account {
			verified, disposition := s.recheckStickyBoundOpenAIAccountFromDB(ctx, account, requestedModel)
			finalizeDisposition = disposition
			return verified
		},
		onSelected: s.buildOpenAIStickyTTLSelectionAdapter(ctx, groupID, sessionHash, stickyTTL),
		onFinalizeMiss: func() {
			disposition := finalizeDisposition
			if disposition.Outcome == stickyBindingOutcomeUsable {
				disposition = newStickyBindingSoftMiss("db_recheck_miss")
			}
			if disposition.Outcome == stickyBindingOutcomeHardInvalidate {
				_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
			}
			recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, disposition, accountID, sessionHash, "")
			finalizeDisposition = stickyBindingDisposition{}
		},
	})
}

func (s *OpenAIGatewayService) trySelectOpenAILegacyFallbackAccount(
	ctx context.Context,
	candidates []*Account,
	groupID *int64,
	sessionHash string,
	requestedModel string,
) (*AccountSelectionResult, bool) {
	ordered := append([]*Account(nil), candidates...)
	sortAccountsByPriorityAndLastUsed(ordered, false)
	result, err, ok := s.tryAcquireImmediateOpenAISelectionFromOrderedAccounts(ctx, ordered, openAIImmediateSelectionRequest{
		groupID:             groupID,
		sessionHash:         sessionHash,
		requestedModel:      requestedModel,
		requiredTransport:   OpenAIUpstreamTransportAny,
		bindSticky:          true,
		recheckAfterAcquire: false,
	})
	if err != nil {
		return nil, false
	}
	if ok {
		return result, true
	}
	return nil, false
}

func (s *OpenAIGatewayService) trySelectOpenAILoadAwareAvailableAccount(
	ctx context.Context,
	candidates []*Account,
	loadMap map[int64]*AccountLoadInfo,
	groupID *int64,
	sessionHash string,
	requestedModel string,
) (*AccountSelectionResult, bool) {
	available := buildAvailableAccountLoads(candidates, loadMap)
	if len(available) == 0 {
		return nil, false
	}

	sortAccountsByPriorityLoadAndLastUsed(available, false)
	ordered := make([]*Account, 0, len(available))
	for _, item := range available {
		ordered = append(ordered, item.account)
	}
	result, err, ok := s.tryAcquireImmediateOpenAISelectionFromOrderedAccounts(ctx, ordered, openAIImmediateSelectionRequest{
		groupID:             groupID,
		sessionHash:         sessionHash,
		requestedModel:      requestedModel,
		requiredTransport:   OpenAIUpstreamTransportAny,
		bindSticky:          true,
		recheckAfterAcquire: false,
	})
	if err != nil {
		return nil, false
	}
	if ok {
		return result, true
	}
	return nil, false
}

func isBetterOpenAIWaitPlanCandidate(candidate, current *openAIWaitCandidate) bool {
	if candidate == nil || candidate.account == nil {
		return false
	}
	if current == nil || current.account == nil {
		return true
	}
	return shouldPreferOpenAIWaitCandidate(candidate.account, candidate.loadInfo, current.account, current.loadInfo)
}

func (s *OpenAIGatewayService) selectOpenAIAccountWithLoadAwarenessFromIndexedSnapshot(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requestedModel string,
	excludedIDs map[int64]struct{},
	cfg config.GatewaySchedulingConfig,
) (*AccountSelectionResult, bool, error) {
	pager, err := s.buildOpenAIIndexedCandidatePager(ctx, openAIIndexedCandidateScope{
		groupID:           groupID,
		requestedModel:    requestedModel,
		requiredTransport: OpenAIUpstreamTransportAny,
	})
	if err != nil {
		return nil, false, err
	}
	if pager == nil {
		return nil, false, nil
	}

	supported := strings.TrimSpace(requestedModel) == ""
	var selected *AccountSelectionResult
	modelScopedFound, bestWait, err := executeIndexedRuntimeSelection(
		ctx,
		pager,
		snapshotPageSizeOrDefault(s.cfg),
		func(batch []*Account) (bool, *openAIWaitCandidate, error) {
			supported = true
			candidates := filterSchedulableOpenAIAccountPointers(batch, requestedModel, excludedIDs)
			if len(candidates) == 0 {
				return false, nil, nil
			}

			loadMap, loadErr := s.concurrencyService.GetAccountsLoadBatch(ctx, buildAccountLoadRequests(candidates))
			if loadErr != nil {
				if result, ok := s.trySelectOpenAILegacyFallbackAccount(
					ctx,
					candidates,
					groupID,
					sessionHash,
					requestedModel,
				); ok {
					selected = result
					return true, nil, nil
				}
			} else {
				if result, ok := s.trySelectOpenAILoadAwareAvailableAccount(
					ctx,
					candidates,
					loadMap,
					groupID,
					sessionHash,
					requestedModel,
				); ok {
					selected = result
					return true, nil, nil
				}
			}

			waitCandidate := selectBestOpenAIWaitCandidate(candidates, loadMap)
			if waitCandidate == nil {
				return false, nil, nil
			}
			return false, &openAIWaitCandidate{
				account:  waitCandidate,
				loadInfo: accountLoadInfoOrDefault(loadMap, waitCandidate.ID),
			}, nil
		},
		isBetterOpenAIWaitPlanCandidate,
	)
	if err != nil {
		return nil, supported, err
	}
	supported = supported || modelScopedFound
	if selected != nil {
		return selected, supported, nil
	}
	if bestWait != nil && bestWait.account != nil {
		if waitPlan, ok := s.tryBuildOpenAIWaitPlanSelection(ctx, bestWait.account, requestedModel, OpenAIUpstreamTransportAny, cfg); ok {
			return waitPlan, supported, nil
		}
	}
	return nil, supported, nil
}
