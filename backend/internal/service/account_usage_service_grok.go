package service

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

func (s *AccountUsageService) getGrokUsage(ctx context.Context, account *Account) (*UsageInfo, error) {
	now := time.Now().UTC()
	usage := buildGrokUsageInfo(account, now)

	if account != nil && account.Type == AccountTypeSession && s != nil && s.grokQuotaSyncer != nil {
		syncCtx, cancel := context.WithTimeout(ctx, grokQuotaSyncTimeout)
		defer cancel()

		if err := s.grokQuotaSyncer.SyncAccount(syncCtx, account); err != nil {
			usage.Error = err.Error()
			return usage, nil
		}
		usage = buildGrokUsageInfo(account, now)
	}

	return usage, nil
}

func buildGrokUsageInfo(account *Account, now time.Time) *UsageInfo {
	updatedAt := now
	info := &UsageInfo{
		Source:    "active",
		UpdatedAt: &updatedAt,
	}

	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return info
	}

	if sync := account.grokSyncState(); sync.LastProbeAt != nil {
		info.UpdatedAt = sync.LastProbeAt
	} else if sync.LastSyncAt != nil {
		info.UpdatedAt = sync.LastSyncAt
	}

	windowNames := []string{
		grok.QuotaWindowAuto,
		grok.QuotaWindowFast,
		grok.QuotaWindowExpert,
		grok.QuotaWindowHeavy,
	}
	quotaWindows := make(map[string]*UsageProgress, len(windowNames))
	for _, name := range windowNames {
		progress := buildGrokUsageProgress(account.grokQuotaWindow(name), now)
		if progress == nil {
			continue
		}
		quotaWindows[name] = progress
	}
	if len(quotaWindows) > 0 {
		info.GrokQuotaWindows = quotaWindows
	}

	return info
}

func buildGrokUsageProgress(window grokQuotaWindowState, now time.Time) *UsageProgress {
	if !window.HasSignal || window.Total <= 0 {
		return nil
	}

	remaining := window.Remaining
	if remaining < 0 {
		remaining = 0
	}
	used := window.Total - remaining
	if used < 0 {
		used = 0
	}
	if used > window.Total {
		used = window.Total
	}

	var resetAt *time.Time
	remainingSeconds := 0
	if window.ResetAt != nil {
		resetValue := window.ResetAt.UTC()
		resetAt = &resetValue
		if now.Before(resetValue) {
			remainingSeconds = int(resetValue.Sub(now).Seconds())
		} else {
			used = 0
		}
	}

	utilization := 0.0
	if window.Total > 0 {
		utilization = (float64(used) / float64(window.Total)) * 100
	}

	return &UsageProgress{
		Utilization:      utilization,
		ResetsAt:         resetAt,
		RemainingSeconds: remainingSeconds,
		UsedRequests:     int64(used),
		LimitRequests:    int64(window.Total),
	}
}
