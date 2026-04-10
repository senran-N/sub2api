package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

// accountWithLoad 账号与负载信息的组合，用于负载感知调度
type accountWithLoad struct {
	account  *Account
	loadInfo *AccountLoadInfo
}

type AccountWaitPlan struct {
	AccountID      int64
	MaxConcurrency int
	Timeout        time.Duration
	MaxWaiting     int
}

type AccountSelectionResult struct {
	Account     *Account
	Acquired    bool
	ReleaseFunc func()
	WaitPlan    *AccountWaitPlan // nil means no wait allowed
}

func buildAccountLoadRequests(candidates []*Account) []AccountWithConcurrency {
	accountLoads := make([]AccountWithConcurrency, 0, len(candidates))
	for _, acc := range candidates {
		accountLoads = append(accountLoads, AccountWithConcurrency{
			ID:             acc.ID,
			MaxConcurrency: acc.EffectiveLoadFactor(),
		})
	}
	return accountLoads
}

func buildAvailableAccountLoads(candidates []*Account, loadMap map[int64]*AccountLoadInfo) []accountWithLoad {
	available := make([]accountWithLoad, 0, len(candidates))
	for _, acc := range candidates {
		loadInfo := loadMap[acc.ID]
		if loadInfo == nil {
			loadInfo = &AccountLoadInfo{AccountID: acc.ID}
		}
		if loadInfo.LoadRate < 100 {
			available = append(available, accountWithLoad{
				account:  acc,
				loadInfo: loadInfo,
			})
		}
	}
	return available
}

func removeAccountWithLoadByID(available []accountWithLoad, accountID int64) []accountWithLoad {
	filtered := make([]accountWithLoad, 0, max(0, len(available)-1))
	for _, acc := range available {
		if acc.account.ID != accountID {
			filtered = append(filtered, acc)
		}
	}
	return filtered
}

func newAcquiredAccountSelection(account *Account, releaseFunc func()) *AccountSelectionResult {
	return &AccountSelectionResult{
		Account:     account,
		Acquired:    true,
		ReleaseFunc: releaseFunc,
	}
}

func newWaitPlanAccountSelection(account *Account, timeout time.Duration, maxWaiting int) *AccountSelectionResult {
	return &AccountSelectionResult{
		Account: account,
		WaitPlan: &AccountWaitPlan{
			AccountID:      account.ID,
			MaxConcurrency: account.Concurrency,
			Timeout:        timeout,
			MaxWaiting:     maxWaiting,
		},
	}
}

func (s *GatewayService) tryAcquireAccountSlot(ctx context.Context, accountID int64, maxConcurrency int) (*AcquireResult, error) {
	if s.concurrencyService == nil {
		return &AcquireResult{Acquired: true, ReleaseFunc: func() {}}, nil
	}
	return s.concurrencyService.AcquireAccountSlot(ctx, accountID, maxConcurrency)
}

func (s *GatewayService) tryAcquireByLegacyOrder(ctx context.Context, candidates []*Account, groupID *int64, sessionHash string, preferOAuth bool) (*AccountSelectionResult, bool) {
	ordered := append([]*Account(nil), candidates...)
	sortAccountsByPriorityAndLastUsed(ordered, preferOAuth)

	for _, acc := range ordered {
		if result, ok := s.tryAcquireAndMaybeBindSelection(ctx, groupID, sessionHash, acc, true); ok {
			return result, true
		}
	}

	return nil, false
}

func (s *GatewayService) schedulingConfig() config.GatewaySchedulingConfig {
	if s.cfg != nil {
		return s.cfg.Gateway.Scheduling
	}
	return config.GatewaySchedulingConfig{
		StickySessionMaxWaiting:  3,
		StickySessionWaitTimeout: 45 * time.Second,
		FallbackWaitTimeout:      30 * time.Second,
		FallbackMaxWaiting:       100,
		LoadBatchEnabled:         true,
		SlotCleanupInterval:      30 * time.Second,
	}
}

func (s *GatewayService) tryAcquireAndMaybeBindSelection(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	account *Account,
	bindSticky bool,
) (*AccountSelectionResult, bool) {
	result, err := s.tryAcquireAccountSlot(ctx, account.ID, account.Concurrency)
	if err != nil || !result.Acquired {
		return nil, false
	}
	hydrated := s.hydrateSelectedAccountOrNil(ctx, account)
	if hydrated == nil {
		if result.ReleaseFunc != nil {
			result.ReleaseFunc()
		}
		return nil, false
	}
	if !s.checkAndRegisterSession(ctx, hydrated, sessionHash) {
		result.ReleaseFunc()
		return nil, false
	}
	if bindSticky {
		s.bindStickySelection(ctx, groupID, sessionHash, hydrated.ID)
	}
	return newAcquiredAccountSelection(hydrated, result.ReleaseFunc), true
}

func (s *GatewayService) tryBuildAccountWaitPlan(
	ctx context.Context,
	account *Account,
	sessionHash string,
	timeout time.Duration,
	maxWaiting int,
) (*AccountSelectionResult, string, bool) {
	if s.concurrencyService != nil {
		waitingCount, _ := s.concurrencyService.GetAccountWaitingCount(ctx, account.ID)
		if waitingCount >= maxWaiting {
			return nil, "wait_queue_full", false
		}
	}

	if !s.checkAndRegisterSession(ctx, account, sessionHash) {
		return nil, "session_limit", false
	}

	hydrated := s.hydrateSelectedAccountOrNil(ctx, account)
	if hydrated == nil {
		return nil, "hydrate_miss", false
	}

	return newWaitPlanAccountSelection(hydrated, timeout, maxWaiting), "", true
}

func (s *GatewayService) selectLoadAwareAvailableAccount(
	ctx context.Context,
	available []accountWithLoad,
	groupID *int64,
	sessionHash string,
	preferOAuth bool,
) (*AccountSelectionResult, bool) {
	if len(available) == 0 {
		return nil, false
	}

	sortAccountsByPriorityLoadAndLastUsed(available, preferOAuth)
	for _, item := range available {
		if result, ok := s.tryAcquireAndMaybeBindSelection(ctx, groupID, sessionHash, item.account, true); ok {
			return result, true
		}
	}

	return nil, false
}

func (s *GatewayService) selectFallbackWaitPlan(
	ctx context.Context,
	candidates []*Account,
	sessionHash string,
	preferOAuth bool,
	mode string,
	timeout time.Duration,
	maxWaiting int,
) (*AccountSelectionResult, bool) {
	ordered := append([]*Account(nil), candidates...)
	s.sortCandidatesForFallback(ordered, preferOAuth, mode)

	for _, acc := range ordered {
		if result, _, ok := s.tryBuildAccountWaitPlan(ctx, acc, sessionHash, timeout, maxWaiting); ok {
			return result, true
		}
	}

	return nil, false
}
