package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const antigravityQuotaFetchTimeout = 30 * time.Second

func (s *AccountUsageService) getGeminiUsage(ctx context.Context, account *Account) (*UsageInfo, error) {
	now := time.Now()
	usage := &UsageInfo{UpdatedAt: &now}

	if s.geminiQuotaService == nil || s.usageLogRepo == nil {
		return usage, nil
	}

	quota, ok := s.geminiQuotaService.QuotaForAccount(ctx, account)
	if !ok {
		return usage, nil
	}

	dayStart := geminiDailyWindowStart(now)
	dayStats, err := s.usageLogRepo.GetModelStatsWithFilters(ctx, dayStart, now, 0, 0, account.ID, 0, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get gemini usage stats failed: %w", err)
	}
	dayTotals := geminiAggregateUsage(dayStats)
	dailyResetAt := geminiDailyResetTime(now)

	if quota.SharedRPD > 0 {
		totalRequests := dayTotals.ProRequests + dayTotals.FlashRequests
		totalTokens := dayTotals.ProTokens + dayTotals.FlashTokens
		totalCost := dayTotals.ProCost + dayTotals.FlashCost
		usage.GeminiSharedDaily = buildGeminiUsageProgress(totalRequests, quota.SharedRPD, dailyResetAt, totalTokens, totalCost, now)
	} else {
		usage.GeminiProDaily = buildGeminiUsageProgress(dayTotals.ProRequests, quota.ProRPD, dailyResetAt, dayTotals.ProTokens, dayTotals.ProCost, now)
		usage.GeminiFlashDaily = buildGeminiUsageProgress(dayTotals.FlashRequests, quota.FlashRPD, dailyResetAt, dayTotals.FlashTokens, dayTotals.FlashCost, now)
	}

	minuteStart := now.Truncate(time.Minute)
	minuteResetAt := minuteStart.Add(time.Minute)
	minuteStats, err := s.usageLogRepo.GetModelStatsWithFilters(ctx, minuteStart, now, 0, 0, account.ID, 0, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get gemini minute usage stats failed: %w", err)
	}
	minuteTotals := geminiAggregateUsage(minuteStats)

	if quota.SharedRPM > 0 {
		totalRequests := minuteTotals.ProRequests + minuteTotals.FlashRequests
		totalTokens := minuteTotals.ProTokens + minuteTotals.FlashTokens
		totalCost := minuteTotals.ProCost + minuteTotals.FlashCost
		usage.GeminiSharedMinute = buildGeminiUsageProgress(totalRequests, quota.SharedRPM, minuteResetAt, totalTokens, totalCost, now)
	} else {
		usage.GeminiProMinute = buildGeminiUsageProgress(minuteTotals.ProRequests, quota.ProRPM, minuteResetAt, minuteTotals.ProTokens, minuteTotals.ProCost, now)
		usage.GeminiFlashMinute = buildGeminiUsageProgress(minuteTotals.FlashRequests, quota.FlashRPM, minuteResetAt, minuteTotals.FlashTokens, minuteTotals.FlashCost, now)
	}

	return usage, nil
}

func buildGeminiUsageProgress(used, limit int64, resetAt time.Time, tokens int64, cost float64, now time.Time) *UsageProgress {
	if limit <= 0 {
		return nil
	}

	remainingSeconds := int(resetAt.Sub(now).Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	resetCopy := resetAt
	return &UsageProgress{
		Utilization:      (float64(used) / float64(limit)) * 100,
		ResetsAt:         &resetCopy,
		RemainingSeconds: remainingSeconds,
		UsedRequests:     used,
		LimitRequests:    limit,
		WindowStats: &WindowStats{
			Requests: used,
			Tokens:   tokens,
			Cost:     cost,
		},
	}
}

// getAntigravityUsage 获取 Antigravity 账户额度。
func (s *AccountUsageService) getAntigravityUsage(ctx context.Context, account *Account) (*UsageInfo, error) {
	if s.antigravityQuotaFetcher == nil || !s.antigravityQuotaFetcher.CanFetch(account) {
		now := time.Now()
		return &UsageInfo{UpdatedAt: &now}, nil
	}

	if cached, ok := s.cache.antigravityCache.Load(account.ID); ok {
		if entry, ok := cached.(*antigravityUsageCache); ok {
			if time.Since(entry.timestamp) < antigravityCacheTTL(entry.usageInfo) {
				recalcAntigravityRemainingSeconds(entry.usageInfo)
				return entry.usageInfo, nil
			}
		}
	}

	flightKey := fmt.Sprintf("ag-usage:%d", account.ID)
	result, err, _ := s.cache.antigravityFlight.Do(flightKey, func() (any, error) {
		if cached, ok := s.cache.antigravityCache.Load(account.ID); ok {
			if entry, ok := cached.(*antigravityUsageCache); ok {
				if time.Since(entry.timestamp) < antigravityCacheTTL(entry.usageInfo) {
					recalcAntigravityRemainingSeconds(entry.usageInfo)
					return entry.usageInfo, nil
				}
			}
		}

		fetchCtx, cancel := context.WithTimeout(context.Background(), antigravityQuotaFetchTimeout)
		defer cancel()

		proxyURL := s.antigravityQuotaFetcher.GetProxyURL(fetchCtx, account)
		fetchResult, fetchErr := s.antigravityQuotaFetcher.FetchQuota(fetchCtx, account, proxyURL)
		if fetchErr != nil {
			degraded := buildAntigravityDegradedUsage(fetchErr)
			enrichUsageWithAccountError(degraded, account)
			s.cache.antigravityCache.Store(account.ID, &antigravityUsageCache{
				usageInfo: degraded,
				timestamp: time.Now(),
			})
			return degraded, nil
		}

		enrichUsageWithAccountError(fetchResult.UsageInfo, account)
		s.cache.antigravityCache.Store(account.ID, &antigravityUsageCache{
			usageInfo: fetchResult.UsageInfo,
			timestamp: time.Now(),
		})
		return fetchResult.UsageInfo, nil
	})
	if err != nil {
		return nil, err
	}

	usage, ok := result.(*UsageInfo)
	if ok && usage != nil {
		return usage, nil
	}

	now := time.Now()
	return &UsageInfo{UpdatedAt: &now}, nil
}

func recalcAntigravityRemainingSeconds(info *UsageInfo) {
	if info == nil || info.FiveHour == nil || info.FiveHour.ResetsAt == nil {
		return
	}

	remaining := int(time.Until(*info.FiveHour.ResetsAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	info.FiveHour.RemainingSeconds = remaining
}

func antigravityCacheTTL(info *UsageInfo) time.Duration {
	if info == nil {
		return antigravityErrorTTL
	}
	if info.IsForbidden {
		return apiCacheTTL
	}
	if info.ErrorCode != "" || info.Error != "" {
		return antigravityErrorTTL
	}
	return apiCacheTTL
}

func buildAntigravityDegradedUsage(err error) *UsageInfo {
	now := time.Now()
	errMessage := fmt.Sprintf("usage API error: %v", err)
	slog.Warn("antigravity usage fetch failed, returning degraded response", "error", err)

	info := &UsageInfo{
		UpdatedAt: &now,
		Error:     errMessage,
	}

	errText := err.Error()
	switch {
	case strings.Contains(errText, "HTTP 401"),
		strings.Contains(errText, "UNAUTHENTICATED"),
		strings.Contains(errText, "invalid_grant"):
		info.ErrorCode = errorCodeUnauthenticated
		info.NeedsReauth = true
	case strings.Contains(errText, "HTTP 429"):
		info.ErrorCode = errorCodeRateLimited
	default:
		info.ErrorCode = errorCodeNetworkError
	}

	return info
}

// enrichUsageWithAccountError 结合账号错误状态修正 UsageInfo。
func enrichUsageWithAccountError(info *UsageInfo, account *Account) {
	if info == nil || account == nil || account.Status != StatusError {
		return
	}

	errorMessage := strings.ToLower(account.ErrorMessage)
	if !strings.Contains(errorMessage, "403") &&
		!strings.Contains(errorMessage, "forbidden") &&
		!strings.Contains(errorMessage, "violation") &&
		!strings.Contains(errorMessage, "validation") {
		return
	}

	forbiddenType := classifyForbiddenType(account.ErrorMessage)
	info.IsForbidden = true
	info.ForbiddenType = forbiddenType
	info.ForbiddenReason = account.ErrorMessage
	info.NeedsVerify = forbiddenType == forbiddenTypeValidation
	info.IsBanned = forbiddenType == forbiddenTypeViolation
	info.ValidationURL = extractValidationURL(account.ErrorMessage)
	info.ErrorCode = errorCodeForbidden
	info.NeedsReauth = false
}
