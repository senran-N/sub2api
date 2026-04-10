package service

import (
	"context"
	"time"
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

func (s *GatewayService) tryAcquireByLegacyOrder(ctx context.Context, candidates []*Account, groupID *int64, sessionHash string, preferOAuth bool) (*AccountSelectionResult, bool) {
	ordered := append([]*Account(nil), candidates...)
	sortAccountsByPriorityAndLastUsed(ordered, preferOAuth)
	result, _, ok := selectFirstOrderedRuntimeSelection(ordered, func(account *Account) (*AccountSelectionResult, error, bool) {
		result, ok := s.tryAcquireAndMaybeBindSelection(ctx, groupID, sessionHash, account, true)
		return result, nil, ok
	})
	return result, ok
}

func (s *GatewayService) tryAcquireAndMaybeBindSelection(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	account *Account,
	bindSticky bool,
) (*AccountSelectionResult, bool) {
	result, err, ok := tryAcquireRuntimeSelection(ctx, s.buildGatewayRuntimeAcquireSpec(
		ctx,
		account,
		sessionHash,
		func(account *Account) *Account {
			return s.hydrateSelectedAccountOrNil(ctx, account)
		},
		s.buildGatewayStickyBindAdapter(ctx, groupID, sessionHash, bindSticky),
	))
	if err != nil || !ok {
		return nil, false
	}
	return result, true
}

func (s *GatewayService) tryBuildAccountWaitPlan(
	ctx context.Context,
	account *Account,
	sessionHash string,
	timeout time.Duration,
	maxWaiting int,
) (*AccountSelectionResult, string, bool) {
	return tryBuildRuntimeWaitPlan(ctx, s.buildGatewayRuntimeWaitPlanSpec(
		ctx,
		account,
		sessionHash,
		func(account *Account) *Account {
			return s.hydrateSelectedAccountOrNil(ctx, account)
		},
		timeout,
		maxWaiting,
	))
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
	ordered := make([]*Account, 0, len(available))
	for _, item := range available {
		ordered = append(ordered, item.account)
	}
	result, _, ok := selectFirstOrderedRuntimeSelection(ordered, func(account *Account) (*AccountSelectionResult, error, bool) {
		result, ok := s.tryAcquireAndMaybeBindSelection(ctx, groupID, sessionHash, account, true)
		return result, nil, ok
	})
	return result, ok
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
	return selectFirstOrderedWaitPlan(ordered, func(account *Account) (*AccountSelectionResult, string, bool) {
		return s.tryBuildAccountWaitPlan(ctx, account, sessionHash, timeout, maxWaiting)
	})
}
