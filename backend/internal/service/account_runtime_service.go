package service

import (
	"context"
	"time"
)

// AccountRuntimeMetrics aggregates short-lived runtime signals used by the
// admin accounts API without leaking cache/database orchestration into
// handlers.
type AccountRuntimeMetrics struct {
	ConcurrencyCounts map[int64]int
	WindowCosts       map[int64]float64
	ActiveSessions    map[int64]int
	RPMCounts         map[int64]int
}

// AccountRuntimeService collects runtime-only account metrics from caches and
// batched usage queries. It intentionally keeps the HTTP layer free from
// orchestration logic.
type AccountRuntimeService struct {
	accountUsageService *AccountUsageService
	concurrencyService  *ConcurrencyService
	sessionLimitCache   SessionLimitCache
	rpmCache            RPMCache
}

func NewAccountRuntimeService(
	accountUsageService *AccountUsageService,
	concurrencyService *ConcurrencyService,
	sessionLimitCache SessionLimitCache,
	rpmCache RPMCache,
) *AccountRuntimeService {
	return &AccountRuntimeService{
		accountUsageService: accountUsageService,
		concurrencyService:  concurrencyService,
		sessionLimitCache:   sessionLimitCache,
		rpmCache:            rpmCache,
	}
}

func (s *AccountRuntimeService) CollectAccountMetrics(ctx context.Context, accounts []Account) AccountRuntimeMetrics {
	metrics := AccountRuntimeMetrics{
		ConcurrencyCounts: make(map[int64]int),
		WindowCosts:       make(map[int64]float64),
		ActiveSessions:    make(map[int64]int),
		RPMCounts:         make(map[int64]int),
	}
	if len(accounts) == 0 {
		return metrics
	}

	accountIDs := make([]int64, 0, len(accounts))
	windowCostScopes := make(map[int64]time.Time, len(accounts))
	sessionLimitIDs := make([]int64, 0, len(accounts))
	rpmIDs := make([]int64, 0, len(accounts))
	sessionIdleTimeouts := make(map[int64]time.Duration)
	missingWindowCostAccounts := make([]Account, 0, len(accounts))

	for i := range accounts {
		account := &accounts[i]
		if account == nil || account.ID <= 0 {
			continue
		}
		accountIDs = append(accountIDs, account.ID)
		if !account.IsAnthropicOAuthOrSetupToken() {
			continue
		}
		if account.GetWindowCostLimit() > 0 {
			windowCostScopes[account.ID] = account.GetCurrentWindowStartTime()
			missingWindowCostAccounts = append(missingWindowCostAccounts, *account)
		}
		if account.GetMaxSessions() > 0 {
			sessionLimitIDs = append(sessionLimitIDs, account.ID)
			sessionIdleTimeouts[account.ID] = time.Duration(account.GetSessionIdleTimeoutMinutes()) * time.Minute
		}
		if account.GetBaseRPM() > 0 {
			rpmIDs = append(rpmIDs, account.ID)
		}
	}

	if len(accountIDs) > 0 && s.concurrencyService != nil {
		if counts, err := s.concurrencyService.GetAccountConcurrencyBatch(ctx, accountIDs); err == nil && counts != nil {
			metrics.ConcurrencyCounts = counts
		}
	}

	if len(rpmIDs) > 0 && s.rpmCache != nil {
		if counts, err := s.rpmCache.GetRPMBatch(ctx, rpmIDs); err == nil && counts != nil {
			metrics.RPMCounts = counts
		}
	}

	if len(sessionLimitIDs) > 0 && s.sessionLimitCache != nil {
		if sessions, err := s.sessionLimitCache.GetActiveSessionCountBatch(ctx, sessionLimitIDs, sessionIdleTimeouts); err == nil && sessions != nil {
			metrics.ActiveSessions = sessions
		}
	}

	if len(windowCostScopes) == 0 {
		return metrics
	}

	if s.sessionLimitCache != nil {
		if cachedCosts, err := s.sessionLimitCache.GetWindowCostBatch(ctx, windowCostScopes); err == nil && cachedCosts != nil {
			for accountID, cost := range cachedCosts {
				metrics.WindowCosts[accountID] = cost
			}
		}
	}

	if len(missingWindowCostAccounts) == 0 || s.accountUsageService == nil {
		return metrics
	}

	filteredAccounts := missingWindowCostAccounts[:0]
	for _, account := range missingWindowCostAccounts {
		if _, ok := metrics.WindowCosts[account.ID]; ok {
			continue
		}
		filteredAccounts = append(filteredAccounts, account)
	}
	if len(filteredAccounts) == 0 {
		return metrics
	}

	statsByAccount, err := s.accountUsageService.GetAccountWindowStatsBatch(ctx, filteredAccounts)
	if err != nil {
		return metrics
	}

	for _, account := range filteredAccounts {
		cost := 0.0
		if stats := statsByAccount[account.ID]; stats != nil {
			cost = stats.StandardCost
		}
		metrics.WindowCosts[account.ID] = cost
		if s.sessionLimitCache != nil {
			_ = s.sessionLimitCache.SetWindowCost(ctx, account.ID, account.GetCurrentWindowStartTime(), cost)
		}
	}

	return metrics
}
