package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/timezone"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"golang.org/x/sync/errgroup"
)

// addWindowStats 为 usage 数据添加窗口期统计，使用独立缓存与上游 API 缓存隔离。
func (s *AccountUsageService) addWindowStats(ctx context.Context, account *Account, usage *UsageInfo) {
	if usage.FiveHour == nil && usage.SevenDay == nil && usage.SevenDaySonnet == nil {
		return
	}

	windowStats := s.cachedWindowStats(account.ID)
	if windowStats == nil {
		stats, err := s.usageLogRepo.GetAccountWindowStats(ctx, account.ID, account.GetCurrentWindowStartTime())
		if err != nil {
			log.Printf("Failed to get window stats for account %d: %v", account.ID, err)
			return
		}

		windowStats = &WindowStats{
			Requests:     stats.Requests,
			Tokens:       stats.Tokens,
			Cost:         stats.Cost,
			StandardCost: stats.StandardCost,
			UserCost:     stats.UserCost,
		}
		s.cache.windowStatsCache.Store(account.ID, &windowStatsCache{
			stats:     windowStats,
			timestamp: time.Now(),
		})
	}

	if usage.FiveHour != nil {
		usage.FiveHour.WindowStats = windowStats
	}
}

func (s *AccountUsageService) cachedWindowStats(accountID int64) *WindowStats {
	if s == nil || s.cache == nil {
		return nil
	}
	cached, ok := s.cache.windowStatsCache.Load(accountID)
	if !ok {
		return nil
	}
	entry, ok := cached.(*windowStatsCache)
	if !ok || time.Since(entry.timestamp) >= windowStatsCacheTTL {
		return nil
	}
	return entry.stats
}

// GetTodayStats 获取账号今日统计。
func (s *AccountUsageService) GetTodayStats(ctx context.Context, accountID int64) (*WindowStats, error) {
	stats, err := s.usageLogRepo.GetAccountTodayStats(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get today stats failed: %w", err)
	}
	return windowStatsFromAccountStats(stats), nil
}

// GetTodayStatsBatch 批量获取账号今日统计，优先走批量 SQL，失败时回退单账号查询。
func (s *AccountUsageService) GetTodayStatsBatch(ctx context.Context, accountIDs []int64) (map[int64]*WindowStats, error) {
	uniqueIDs := collectUniquePositiveAccountIDs(accountIDs)
	result := make(map[int64]*WindowStats, len(uniqueIDs))
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	startTime := timezone.Today()
	if batchReader, ok := s.usageLogRepo.(accountWindowStatsBatchReader); ok {
		statsByAccount, err := batchReader.GetAccountWindowStatsBatch(ctx, uniqueIDs, startTime)
		if err == nil {
			for _, accountID := range uniqueIDs {
				result[accountID] = windowStatsFromAccountStats(statsByAccount[accountID])
			}
			return result, nil
		}
	}

	if err := loadAccountWindowStatsIndividually(ctx, s.usageLogRepo, uniqueIDs, startTime, func(accountID int64, stats *usagestats.AccountStats) {
		result[accountID] = windowStatsFromAccountStats(stats)
	}); err != nil {
		return nil, err
	}

	for _, accountID := range uniqueIDs {
		if _, ok := result[accountID]; !ok {
			result[accountID] = &WindowStats{}
		}
	}
	return result, nil
}

// GetAccountWindowStatsBatch groups accounts by current window start and
// fetches usage stats in batches to keep callers off the per-account SQL path.
func (s *AccountUsageService) GetAccountWindowStatsBatch(ctx context.Context, accounts []Account) (map[int64]*usagestats.AccountStats, error) {
	result := make(map[int64]*usagestats.AccountStats)
	if s == nil || s.usageLogRepo == nil || len(accounts) == 0 {
		return result, nil
	}

	groupedAccountIDs := make(map[int64][]int64)
	windowStartTimes := make(map[int64]time.Time)
	seen := make(map[int64]struct{}, len(accounts))

	for i := range accounts {
		account := &accounts[i]
		if account == nil || account.ID <= 0 || !account.IsAnthropicOAuthOrSetupToken() || account.GetWindowCostLimit() <= 0 {
			continue
		}
		if _, exists := seen[account.ID]; exists {
			continue
		}
		seen[account.ID] = struct{}{}
		startTime := account.GetCurrentWindowStartTime()
		startKey := startTime.Unix()
		groupedAccountIDs[startKey] = append(groupedAccountIDs[startKey], account.ID)
		windowStartTimes[startKey] = startTime
		result[account.ID] = &usagestats.AccountStats{}
	}

	if len(result) == 0 {
		return result, nil
	}

	batchReader, hasBatch := s.usageLogRepo.(accountWindowStatsBatchReader)
	for startKey, accountIDs := range groupedAccountIDs {
		startTime := windowStartTimes[startKey]
		if hasBatch {
			statsByAccount, err := batchReader.GetAccountWindowStatsBatch(ctx, accountIDs, startTime)
			if err == nil {
				for _, accountID := range accountIDs {
					if stats, ok := statsByAccount[accountID]; ok && stats != nil {
						result[accountID] = stats
					}
				}
				continue
			}
		}

		if err := loadAccountWindowStatsIndividually(ctx, s.usageLogRepo, accountIDs, startTime, func(accountID int64, stats *usagestats.AccountStats) {
			if stats != nil {
				result[accountID] = stats
			}
		}); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func collectUniquePositiveAccountIDs(accountIDs []int64) []int64 {
	uniqueIDs := make([]int64, 0, len(accountIDs))
	seen := make(map[int64]struct{}, len(accountIDs))
	for _, accountID := range accountIDs {
		if accountID <= 0 {
			continue
		}
		if _, exists := seen[accountID]; exists {
			continue
		}
		seen[accountID] = struct{}{}
		uniqueIDs = append(uniqueIDs, accountID)
	}
	return uniqueIDs
}

func loadAccountWindowStatsIndividually(
	ctx context.Context,
	repo UsageLogRepository,
	accountIDs []int64,
	startTime time.Time,
	handle func(accountID int64, stats *usagestats.AccountStats),
) error {
	var mu sync.Mutex
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(8)

	for _, accountID := range accountIDs {
		id := accountID
		group.Go(func() error {
			stats, err := repo.GetAccountWindowStats(groupCtx, id, startTime)
			if err != nil {
				return nil
			}
			mu.Lock()
			handle(id, stats)
			mu.Unlock()
			return nil
		})
	}

	return group.Wait()
}

func windowStatsFromAccountStats(stats *usagestats.AccountStats) *WindowStats {
	if stats == nil {
		return &WindowStats{}
	}
	return &WindowStats{
		Requests:     stats.Requests,
		Tokens:       stats.Tokens,
		Cost:         stats.Cost,
		StandardCost: stats.StandardCost,
		UserCost:     stats.UserCost,
	}
}

func (s *AccountUsageService) GetAccountUsageStats(ctx context.Context, accountID int64, startTime, endTime time.Time) (*usagestats.AccountUsageStatsResponse, error) {
	stats, err := s.usageLogRepo.GetAccountUsageStats(ctx, accountID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get account usage stats failed: %w", err)
	}
	return stats, nil
}

// GetAccountWindowStats 获取账号在指定时间窗口内的使用统计。
func (s *AccountUsageService) GetAccountWindowStats(ctx context.Context, accountID int64, startTime time.Time) (*usagestats.AccountStats, error) {
	return s.usageLogRepo.GetAccountWindowStats(ctx, accountID, startTime)
}
