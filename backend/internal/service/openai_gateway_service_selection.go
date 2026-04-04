package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

// SelectAccount selects an OpenAI account with sticky session support.
func (s *OpenAIGatewayService) SelectAccount(ctx context.Context, groupID *int64, sessionHash string) (*Account, error) {
	return s.SelectAccountForModel(ctx, groupID, sessionHash, "")
}

// SelectAccountForModel selects an account supporting the requested model.
func (s *OpenAIGatewayService) SelectAccountForModel(ctx context.Context, groupID *int64, sessionHash string, requestedModel string) (*Account, error) {
	return s.SelectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, nil)
}

// SelectAccountForModelWithExclusions selects an account supporting the requested model while excluding specified accounts.
func (s *OpenAIGatewayService) SelectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*Account, error) {
	return s.selectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, excludedIDs, 0)
}

func (s *OpenAIGatewayService) selectAccountForModelWithExclusions(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, stickyAccountID int64) (*Account, error) {
	if account := s.tryStickySessionHit(ctx, groupID, sessionHash, requestedModel, excludedIDs, stickyAccountID); account != nil {
		return account, nil
	}

	accounts, err := s.listSchedulableAccounts(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}

	selected := s.selectBestAccount(ctx, accounts, requestedModel, excludedIDs)
	if selected == nil {
		if requestedModel != "" {
			return nil, fmt.Errorf("no available OpenAI accounts supporting model: %s", requestedModel)
		}
		return nil, errors.New("no available OpenAI accounts")
	}

	if sessionHash != "" {
		_ = s.setStickySessionAccountID(ctx, groupID, sessionHash, selected.ID, s.openAIWSSessionStickyTTL())
	}

	return selected, nil
}

// tryStickySessionHit tries to resolve a sticky session to a currently usable account.
func (s *OpenAIGatewayService) tryStickySessionHit(ctx context.Context, groupID *int64, sessionHash, requestedModel string, excludedIDs map[int64]struct{}, stickyAccountID int64) *Account {
	if sessionHash == "" {
		return nil
	}

	accountID := stickyAccountID
	if accountID <= 0 {
		var err error
		accountID, err = s.getStickySessionAccountID(ctx, groupID, sessionHash)
		if err != nil || accountID <= 0 {
			return nil
		}
	}

	if _, excluded := excludedIDs[accountID]; excluded {
		return nil
	}

	account, err := s.getSchedulableAccount(ctx, accountID)
	if err != nil {
		return nil
	}
	if shouldClearStickySession(account, requestedModel) {
		_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		return nil
	}
	if !account.IsSchedulable() || !account.IsOpenAI() {
		return nil
	}
	if requestedModel != "" && !account.IsModelSupported(requestedModel) {
		return nil
	}

	account = s.recheckSelectedOpenAIAccountFromDB(ctx, account, requestedModel)
	if account == nil {
		_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
		return nil
	}

	_ = s.refreshStickySessionTTL(ctx, groupID, sessionHash, s.openAIWSSessionStickyTTL())
	return account
}

func (s *OpenAIGatewayService) selectBestAccount(ctx context.Context, accounts []Account, requestedModel string, excludedIDs map[int64]struct{}) *Account {
	var selected *Account

	for i := range accounts {
		account := &accounts[i]
		if _, excluded := excludedIDs[account.ID]; excluded {
			continue
		}

		fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, account, requestedModel)
		if fresh == nil {
			continue
		}
		fresh = s.recheckSelectedOpenAIAccountFromDB(ctx, fresh, requestedModel)
		if fresh == nil {
			continue
		}

		if selected == nil || s.isBetterAccount(fresh, selected) {
			selected = fresh
		}
	}

	return selected
}

func (s *OpenAIGatewayService) isBetterAccount(candidate, current *Account) bool {
	if candidate.Priority < current.Priority {
		return true
	}
	if candidate.Priority > current.Priority {
		return false
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

// SelectAccountWithLoadAwareness selects an account with load-awareness and wait plan.
func (s *OpenAIGatewayService) SelectAccountWithLoadAwareness(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*AccountSelectionResult, error) {
	cfg := s.schedulingConfig()
	stickyTTL := s.openAIWSSessionStickyTTL()

	var stickyAccountID int64
	if sessionHash != "" && s.cache != nil {
		if accountID, err := s.getStickySessionAccountID(ctx, groupID, sessionHash); err == nil {
			stickyAccountID = accountID
		}
	}

	if s.concurrencyService == nil || !cfg.LoadBatchEnabled {
		account, err := s.selectAccountForModelWithExclusions(ctx, groupID, sessionHash, requestedModel, excludedIDs, stickyAccountID)
		if err != nil {
			return nil, err
		}

		result, err := s.tryAcquireAccountSlot(ctx, account.ID, account.Concurrency)
		if err == nil && result.Acquired {
			return &AccountSelectionResult{
				Account:     account,
				Acquired:    true,
				ReleaseFunc: result.ReleaseFunc,
			}, nil
		}

		if stickyAccountID > 0 && stickyAccountID == account.ID && s.concurrencyService != nil {
			waitingCount, _ := s.concurrencyService.GetAccountWaitingCount(ctx, account.ID)
			if waitingCount < cfg.StickySessionMaxWaiting {
				return &AccountSelectionResult{
					Account: account,
					WaitPlan: &AccountWaitPlan{
						AccountID:      account.ID,
						MaxConcurrency: account.Concurrency,
						Timeout:        cfg.StickySessionWaitTimeout,
						MaxWaiting:     cfg.StickySessionMaxWaiting,
					},
				}, nil
			}
		}

		return &AccountSelectionResult{
			Account: account,
			WaitPlan: &AccountWaitPlan{
				AccountID:      account.ID,
				MaxConcurrency: account.Concurrency,
				Timeout:        cfg.FallbackWaitTimeout,
				MaxWaiting:     cfg.FallbackMaxWaiting,
			},
		}, nil
	}

	accounts, err := s.listSchedulableAccounts(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	isExcluded := func(accountID int64) bool {
		if excludedIDs == nil {
			return false
		}
		_, excluded := excludedIDs[accountID]
		return excluded
	}

	if sessionHash != "" {
		accountID := stickyAccountID
		if accountID > 0 && !isExcluded(accountID) {
			account, err := s.getSchedulableAccount(ctx, accountID)
			if err == nil {
				clearSticky := shouldClearStickySession(account, requestedModel)
				if clearSticky {
					_ = s.deleteStickySessionAccountID(ctx, groupID, sessionHash)
				}
				if !clearSticky && account.IsSchedulable() && account.IsOpenAI() &&
					(requestedModel == "" || account.IsModelSupported(requestedModel)) {
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
							return &AccountSelectionResult{
								Account:     verified,
								Acquired:    true,
								ReleaseFunc: result.ReleaseFunc,
							}, nil
						}
					}

					waitingCount, _ := s.concurrencyService.GetAccountWaitingCount(ctx, accountID)
					if waitingCount < cfg.StickySessionMaxWaiting {
						return &AccountSelectionResult{
							Account: account,
							WaitPlan: &AccountWaitPlan{
								AccountID:      accountID,
								MaxConcurrency: account.Concurrency,
								Timeout:        cfg.StickySessionWaitTimeout,
								MaxWaiting:     cfg.StickySessionMaxWaiting,
							},
						}, nil
					}
				}
			}
		}
	}

	candidates := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if isExcluded(account.ID) {
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
	if len(candidates) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	accountLoads := make([]AccountWithConcurrency, 0, len(candidates))
	for _, account := range candidates {
		accountLoads = append(accountLoads, AccountWithConcurrency{
			ID:             account.ID,
			MaxConcurrency: account.EffectiveLoadFactor(),
		})
	}

	loadMap, err := s.concurrencyService.GetAccountsLoadBatch(ctx, accountLoads)
	if err != nil {
		ordered := append([]*Account(nil), candidates...)
		sortAccountsByPriorityAndLastUsed(ordered, false)
		for _, account := range ordered {
			fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, account, requestedModel)
			if fresh == nil {
				continue
			}
			result, acquireErr := s.tryAcquireAccountSlot(ctx, fresh.ID, fresh.Concurrency)
			if acquireErr == nil && result.Acquired {
				if sessionHash != "" {
					_ = s.setStickySessionAccountID(ctx, groupID, sessionHash, fresh.ID, stickyTTL)
				}
				return &AccountSelectionResult{
					Account:     fresh,
					Acquired:    true,
					ReleaseFunc: result.ReleaseFunc,
				}, nil
			}
		}
	} else {
		available := make([]accountWithLoad, 0, len(candidates))
		for _, account := range candidates {
			loadInfo := loadMap[account.ID]
			if loadInfo == nil {
				loadInfo = &AccountLoadInfo{AccountID: account.ID}
			}
			if loadInfo.LoadRate < 100 {
				available = append(available, accountWithLoad{
					account:  account,
					loadInfo: loadInfo,
				})
			}
		}

		if len(available) > 0 {
			sort.SliceStable(available, func(i, j int) bool {
				left := available[i]
				right := available[j]
				if left.account.Priority != right.account.Priority {
					return left.account.Priority < right.account.Priority
				}
				if left.loadInfo.LoadRate != right.loadInfo.LoadRate {
					return left.loadInfo.LoadRate < right.loadInfo.LoadRate
				}
				switch {
				case left.account.LastUsedAt == nil && right.account.LastUsedAt != nil:
					return true
				case left.account.LastUsedAt != nil && right.account.LastUsedAt == nil:
					return false
				case left.account.LastUsedAt == nil && right.account.LastUsedAt == nil:
					return false
				default:
					return left.account.LastUsedAt.Before(*right.account.LastUsedAt)
				}
			})
			shuffleWithinSortGroups(available)

			for _, item := range available {
				fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, item.account, requestedModel)
				if fresh == nil {
					continue
				}
				result, acquireErr := s.tryAcquireAccountSlot(ctx, fresh.ID, fresh.Concurrency)
				if acquireErr == nil && result.Acquired {
					if sessionHash != "" {
						_ = s.setStickySessionAccountID(ctx, groupID, sessionHash, fresh.ID, stickyTTL)
					}
					return &AccountSelectionResult{
						Account:     fresh,
						Acquired:    true,
						ReleaseFunc: result.ReleaseFunc,
					}, nil
				}
			}
		}
	}

	waitLoadMap := loadMap
	if waitLoadMap == nil {
		waitLoadMap = map[int64]*AccountLoadInfo{}
	}
	bestWaitIdx := 0
	for i := 1; i < len(candidates); i++ {
		current := candidates[i]
		best := candidates[bestWaitIdx]
		currentLoad := waitLoadMap[current.ID]
		bestLoad := waitLoadMap[best.ID]
		if currentLoad == nil {
			currentLoad = &AccountLoadInfo{AccountID: current.ID}
		}
		if bestLoad == nil {
			bestLoad = &AccountLoadInfo{AccountID: best.ID}
		}
		if currentLoad.WaitingCount != bestLoad.WaitingCount {
			if currentLoad.WaitingCount < bestLoad.WaitingCount {
				bestWaitIdx = i
			}
			continue
		}
		if currentLoad.LoadRate != bestLoad.LoadRate {
			if currentLoad.LoadRate < bestLoad.LoadRate {
				bestWaitIdx = i
			}
			continue
		}
		if current.Priority != best.Priority {
			if current.Priority < best.Priority {
				bestWaitIdx = i
			}
			continue
		}
		switch {
		case current.LastUsedAt == nil && best.LastUsedAt != nil:
			bestWaitIdx = i
		case current.LastUsedAt != nil && best.LastUsedAt == nil:
		case current.LastUsedAt != nil && best.LastUsedAt != nil && current.LastUsedAt.Before(*best.LastUsedAt):
			bestWaitIdx = i
		}
	}

	fresh := s.resolveFreshSchedulableOpenAIAccount(ctx, candidates[bestWaitIdx], requestedModel)
	if fresh != nil {
		return &AccountSelectionResult{
			Account: fresh,
			WaitPlan: &AccountWaitPlan{
				AccountID:      fresh.ID,
				MaxConcurrency: fresh.Concurrency,
				Timeout:        cfg.FallbackWaitTimeout,
				MaxWaiting:     cfg.FallbackMaxWaiting,
			},
		}, nil
	}

	return nil, ErrNoAvailableAccounts
}

func (s *OpenAIGatewayService) listSchedulableAccounts(ctx context.Context, groupID *int64) ([]Account, error) {
	if s.schedulerSnapshot != nil {
		accounts, _, err := s.schedulerSnapshot.ListSchedulableAccounts(ctx, groupID, PlatformOpenAI, false)
		return accounts, err
	}

	var (
		accounts []Account
		err      error
	)
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		accounts, err = s.accountRepo.ListSchedulableByPlatform(ctx, PlatformOpenAI)
	} else if groupID != nil {
		accounts, err = s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, *groupID, PlatformOpenAI)
	} else {
		accounts, err = s.accountRepo.ListSchedulableUngroupedByPlatform(ctx, PlatformOpenAI)
	}
	if err != nil {
		return nil, fmt.Errorf("query accounts failed: %w", err)
	}
	return accounts, nil
}

func (s *OpenAIGatewayService) tryAcquireAccountSlot(ctx context.Context, accountID int64, maxConcurrency int) (*AcquireResult, error) {
	if s.concurrencyService == nil {
		return &AcquireResult{Acquired: true, ReleaseFunc: func() {}}, nil
	}
	return s.concurrencyService.AcquireAccountSlot(ctx, accountID, maxConcurrency)
}

func (s *OpenAIGatewayService) resolveFreshSchedulableOpenAIAccount(ctx context.Context, account *Account, requestedModel string) *Account {
	if account == nil {
		return nil
	}

	fresh := account
	if s.schedulerSnapshot != nil {
		current, err := s.getSchedulableAccount(ctx, account.ID)
		if err != nil || current == nil {
			return nil
		}
		fresh = current
	}
	if !fresh.IsSchedulable() || !fresh.IsOpenAI() {
		return nil
	}
	if requestedModel != "" && !fresh.IsModelSupported(requestedModel) {
		return nil
	}
	return fresh
}

func (s *OpenAIGatewayService) recheckSelectedOpenAIAccountFromDB(ctx context.Context, account *Account, requestedModel string) *Account {
	if account == nil {
		return nil
	}
	if s.schedulerSnapshot == nil || s.accountRepo == nil {
		return account
	}

	latest, err := s.accountRepo.GetByID(ctx, account.ID)
	if err != nil || latest == nil {
		return nil
	}
	syncOpenAICodexRateLimitFromExtra(ctx, s.accountRepo, latest, time.Now())
	if !latest.IsSchedulable() || !latest.IsOpenAI() {
		return nil
	}
	if requestedModel != "" && !latest.IsModelSupported(requestedModel) {
		return nil
	}
	return latest
}

func (s *OpenAIGatewayService) getSchedulableAccount(ctx context.Context, accountID int64) (*Account, error) {
	var (
		account *Account
		err     error
	)
	if s.schedulerSnapshot != nil {
		account, err = s.schedulerSnapshot.GetAccount(ctx, accountID)
	} else {
		account, err = s.accountRepo.GetByID(ctx, accountID)
	}
	if err != nil || account == nil {
		return account, err
	}
	syncOpenAICodexRateLimitFromExtra(ctx, s.accountRepo, account, time.Now())
	return account, nil
}

func (s *OpenAIGatewayService) schedulingConfig() config.GatewaySchedulingConfig {
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
