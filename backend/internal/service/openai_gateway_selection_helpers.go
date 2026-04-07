package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

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
		if !account.IsSchedulable() {
			continue
		}
		if requestedModel != "" && !account.IsModelSupported(requestedModel) {
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
	if err != nil || account == nil {
		if policy.deleteOnLookupMiss {
			_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		}
		return nil, 0
	}
	if shouldClearStickySession(account, requestedModel) {
		_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		return nil, 0
	}
	if !account.IsSchedulable() || !account.IsOpenAI() {
		return nil, 0
	}
	if requestedModel != "" && !account.IsModelSupported(requestedModel) {
		return nil, 0
	}

	if policy.recheckOnResolve {
		account = s.recheckSelectedOpenAIAccountFromDB(ctx, account, requestedModel)
		if account == nil {
			_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
			return nil, 0
		}
	}

	if policy.refreshTTLOnHit {
		_ = s.refreshStickySessionTTL(ctx, groupID, sessionHash, s.openAIWSSessionStickyTTL())
	}

	return account, accountID
}

func (s *OpenAIGatewayService) lookupOpenAIStickyAccountID(ctx context.Context, groupID *int64, sessionHash string) int64 {
	if sessionHash == "" || s.cache == nil {
		return 0
	}

	accountID, err := s.getStickySessionAccountID(ctx, groupID, sessionHash)
	if err != nil {
		return 0
	}
	return accountID
}

func shouldUseOpenAILoadAwareSelection(
	concurrencyService *ConcurrencyService,
	cfg config.GatewaySchedulingConfig,
) bool {
	return concurrencyService != nil && cfg.LoadBatchEnabled
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

	result, err := s.tryAcquireAccountSlot(ctx, account.ID, account.Concurrency)
	if err == nil && result.Acquired {
		return newAcquiredAccountSelection(account, result.ReleaseFunc), nil
	}

	return s.buildOpenAINonLoadBatchWaitPlan(ctx, account, stickyAccountID, cfg), nil
}

func (s *OpenAIGatewayService) buildOpenAINonLoadBatchWaitPlan(
	ctx context.Context,
	account *Account,
	stickyAccountID int64,
	cfg config.GatewaySchedulingConfig,
) *AccountSelectionResult {
	if stickyAccountID > 0 && stickyAccountID == account.ID && s.concurrencyService != nil {
		waitingCount, _ := s.concurrencyService.GetAccountWaitingCount(ctx, account.ID)
		if waitingCount < cfg.StickySessionMaxWaiting {
			return newWaitPlanAccountSelection(account, cfg.StickySessionWaitTimeout, cfg.StickySessionMaxWaiting)
		}
	}

	return newWaitPlanAccountSelection(account, cfg.FallbackWaitTimeout, cfg.FallbackMaxWaiting)
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

	result, err := s.tryAcquireAccountSlot(ctx, accountID, account.Concurrency)
	if err == nil && result.Acquired {
		verified := s.recheckSelectedOpenAIAccountFromDB(ctx, account, requestedModel)
		if verified == nil {
			if result.ReleaseFunc != nil {
				result.ReleaseFunc()
			}
			_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		} else {
			_ = s.refreshStickySessionTTL(ctx, groupID, sessionHash, stickyTTL)
			return newAcquiredAccountSelection(verified, result.ReleaseFunc), true
		}
	}

	waitingCount, _ := s.concurrencyService.GetAccountWaitingCount(ctx, accountID)
	if waitingCount < cfg.StickySessionMaxWaiting {
		return newWaitPlanAccountSelection(account, cfg.StickySessionWaitTimeout, cfg.StickySessionMaxWaiting), true
	}

	return nil, false
}

func (s *OpenAIGatewayService) trySelectOpenAILegacyFallbackAccount(
	ctx context.Context,
	candidates []*Account,
	groupID *int64,
	sessionHash string,
	stickyTTL time.Duration,
	requestedModel string,
) (*AccountSelectionResult, bool) {
	ordered := append([]*Account(nil), candidates...)
	sortAccountsByPriorityAndLastUsed(ordered, false)

	for _, account := range ordered {
		if result, ok := s.tryAcquireFreshOpenAIAccountSelection(
			ctx,
			account,
			requestedModel,
			groupID,
			sessionHash,
			stickyTTL,
		); ok {
			return result, true
		}
	}

	return nil, false
}

func (s *OpenAIGatewayService) trySelectOpenAILoadAwareAvailableAccount(
	ctx context.Context,
	candidates []*Account,
	loadMap map[int64]*AccountLoadInfo,
	groupID *int64,
	sessionHash string,
	stickyTTL time.Duration,
	requestedModel string,
) (*AccountSelectionResult, bool) {
	available := buildAvailableAccountLoads(candidates, loadMap)
	if len(available) == 0 {
		return nil, false
	}

	sortAccountsByPriorityLoadAndLastUsed(available)
	for _, item := range available {
		if result, ok := s.tryAcquireFreshOpenAIAccountSelection(
			ctx,
			item.account,
			requestedModel,
			groupID,
			sessionHash,
			stickyTTL,
		); ok {
			return result, true
		}
	}

	return nil, false
}

func (s *OpenAIGatewayService) tryAcquireFreshOpenAIAccountSelection(
	ctx context.Context,
	account *Account,
	requestedModel string,
	groupID *int64,
	sessionHash string,
	stickyTTL time.Duration,
) (*AccountSelectionResult, bool) {
	fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, account, requestedModel)
	if fresh == nil {
		return nil, false
	}

	result, err := s.tryAcquireAccountSlot(ctx, fresh.ID, fresh.Concurrency)
	if err != nil || !result.Acquired {
		return nil, false
	}

	if sessionHash != "" {
		_ = s.setStickySessionAccountID(ctx, groupID, sessionHash, fresh.ID, stickyTTL)
	}

	return newAcquiredAccountSelection(fresh, result.ReleaseFunc), true
}

func (s *OpenAIGatewayService) selectOpenAILoadAwareCandidates(
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
) ([]*Account, error) {
	candidates := filterSchedulableOpenAICandidates(accounts, requestedModel, excludedIDs)
	if len(candidates) == 0 {
		return nil, ErrNoAvailableAccounts
	}
	return candidates, nil
}

func (s *OpenAIGatewayService) selectOpenAIWaitPlanCandidate(
	ctx context.Context,
	candidates []*Account,
	loadMap map[int64]*AccountLoadInfo,
	requestedModel string,
) *Account {
	waitCandidate := selectBestOpenAIWaitCandidate(candidates, normalizeOpenAIWaitLoadMap(loadMap))
	return s.resolveFreshSchedulableOpenAIAccount(ctx, waitCandidate, requestedModel)
}
