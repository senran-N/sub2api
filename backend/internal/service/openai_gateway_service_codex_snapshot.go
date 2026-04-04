package service

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// OpenAICodexUsageSnapshot represents Codex API usage limits from response headers.
type OpenAICodexUsageSnapshot struct {
	PrimaryUsedPercent          *float64 `json:"primary_used_percent,omitempty"`
	PrimaryResetAfterSeconds    *int     `json:"primary_reset_after_seconds,omitempty"`
	PrimaryWindowMinutes        *int     `json:"primary_window_minutes,omitempty"`
	SecondaryUsedPercent        *float64 `json:"secondary_used_percent,omitempty"`
	SecondaryResetAfterSeconds  *int     `json:"secondary_reset_after_seconds,omitempty"`
	SecondaryWindowMinutes      *int     `json:"secondary_window_minutes,omitempty"`
	PrimaryOverSecondaryPercent *float64 `json:"primary_over_secondary_percent,omitempty"`
	UpdatedAt                   string   `json:"updated_at,omitempty"`
}

// NormalizedCodexLimits contains normalized 5h/7d rate limit data.
type NormalizedCodexLimits struct {
	Used5hPercent   *float64
	Reset5hSeconds  *int
	Window5hMinutes *int
	Used7dPercent   *float64
	Reset7dSeconds  *int
	Window7dMinutes *int
}

// Normalize converts primary/secondary fields to canonical 5h/7d fields.
func (s *OpenAICodexUsageSnapshot) Normalize() *NormalizedCodexLimits {
	if s == nil {
		return nil
	}

	result := &NormalizedCodexLimits{}
	primaryMins := 0
	secondaryMins := 0
	hasPrimaryWindow := false
	hasSecondaryWindow := false

	if s.PrimaryWindowMinutes != nil {
		primaryMins = *s.PrimaryWindowMinutes
		hasPrimaryWindow = true
	}
	if s.SecondaryWindowMinutes != nil {
		secondaryMins = *s.SecondaryWindowMinutes
		hasSecondaryWindow = true
	}

	use5hFromPrimary := false
	use7dFromPrimary := false
	switch {
	case hasPrimaryWindow && hasSecondaryWindow:
		if primaryMins < secondaryMins {
			use5hFromPrimary = true
		} else {
			use7dFromPrimary = true
		}
	case hasPrimaryWindow:
		if primaryMins <= 360 {
			use5hFromPrimary = true
		} else {
			use7dFromPrimary = true
		}
	case hasSecondaryWindow:
		if secondaryMins <= 360 {
			use7dFromPrimary = true
		} else {
			use5hFromPrimary = true
		}
	default:
		use7dFromPrimary = true
	}

	if use5hFromPrimary {
		result.Used5hPercent = s.PrimaryUsedPercent
		result.Reset5hSeconds = s.PrimaryResetAfterSeconds
		result.Window5hMinutes = s.PrimaryWindowMinutes
		result.Used7dPercent = s.SecondaryUsedPercent
		result.Reset7dSeconds = s.SecondaryResetAfterSeconds
		result.Window7dMinutes = s.SecondaryWindowMinutes
	} else if use7dFromPrimary {
		result.Used7dPercent = s.PrimaryUsedPercent
		result.Reset7dSeconds = s.PrimaryResetAfterSeconds
		result.Window7dMinutes = s.PrimaryWindowMinutes
		result.Used5hPercent = s.SecondaryUsedPercent
		result.Reset5hSeconds = s.SecondaryResetAfterSeconds
		result.Window5hMinutes = s.SecondaryWindowMinutes
	}

	return result
}

// ParseCodexRateLimitHeaders extracts Codex usage limits from response headers.
// Exported for use in ratelimit_service when handling OpenAI 429 responses.
func ParseCodexRateLimitHeaders(headers http.Header) *OpenAICodexUsageSnapshot {
	snapshot := &OpenAICodexUsageSnapshot{}
	hasData := false

	parseFloat := func(key string) *float64 {
		if value := headers.Get(key); value != "" {
			if parsed, err := strconv.ParseFloat(value, 64); err == nil {
				return &parsed
			}
		}
		return nil
	}
	parseInt := func(key string) *int {
		if value := headers.Get(key); value != "" {
			if parsed, err := strconv.Atoi(value); err == nil {
				return &parsed
			}
		}
		return nil
	}

	if value := parseFloat("x-codex-primary-used-percent"); value != nil {
		snapshot.PrimaryUsedPercent = value
		hasData = true
	}
	if value := parseInt("x-codex-primary-reset-after-seconds"); value != nil {
		snapshot.PrimaryResetAfterSeconds = value
		hasData = true
	}
	if value := parseInt("x-codex-primary-window-minutes"); value != nil {
		snapshot.PrimaryWindowMinutes = value
		hasData = true
	}
	if value := parseFloat("x-codex-secondary-used-percent"); value != nil {
		snapshot.SecondaryUsedPercent = value
		hasData = true
	}
	if value := parseInt("x-codex-secondary-reset-after-seconds"); value != nil {
		snapshot.SecondaryResetAfterSeconds = value
		hasData = true
	}
	if value := parseInt("x-codex-secondary-window-minutes"); value != nil {
		snapshot.SecondaryWindowMinutes = value
		hasData = true
	}
	if value := parseFloat("x-codex-primary-over-secondary-limit-percent"); value != nil {
		snapshot.PrimaryOverSecondaryPercent = value
		hasData = true
	}
	if !hasData {
		return nil
	}

	snapshot.UpdatedAt = time.Now().Format(time.RFC3339)
	return snapshot
}

func codexSnapshotBaseTime(snapshot *OpenAICodexUsageSnapshot, fallback time.Time) time.Time {
	if snapshot == nil || snapshot.UpdatedAt == "" {
		return fallback
	}
	baseTime, err := time.Parse(time.RFC3339, snapshot.UpdatedAt)
	if err != nil {
		return fallback
	}
	return baseTime
}

func codexResetAtRFC3339(baseTime time.Time, resetAfterSeconds *int) *string {
	if resetAfterSeconds == nil {
		return nil
	}

	seconds := *resetAfterSeconds
	if seconds < 0 {
		seconds = 0
	}
	resetAt := baseTime.Add(time.Duration(seconds) * time.Second).Format(time.RFC3339)
	return &resetAt
}

func buildCodexUsageExtraUpdates(snapshot *OpenAICodexUsageSnapshot, fallbackNow time.Time) map[string]any {
	if snapshot == nil {
		return nil
	}

	baseTime := codexSnapshotBaseTime(snapshot, fallbackNow)
	updates := make(map[string]any)

	if snapshot.PrimaryUsedPercent != nil {
		updates["codex_primary_used_percent"] = *snapshot.PrimaryUsedPercent
	}
	if snapshot.PrimaryResetAfterSeconds != nil {
		updates["codex_primary_reset_after_seconds"] = *snapshot.PrimaryResetAfterSeconds
	}
	if snapshot.PrimaryWindowMinutes != nil {
		updates["codex_primary_window_minutes"] = *snapshot.PrimaryWindowMinutes
	}
	if snapshot.SecondaryUsedPercent != nil {
		updates["codex_secondary_used_percent"] = *snapshot.SecondaryUsedPercent
	}
	if snapshot.SecondaryResetAfterSeconds != nil {
		updates["codex_secondary_reset_after_seconds"] = *snapshot.SecondaryResetAfterSeconds
	}
	if snapshot.SecondaryWindowMinutes != nil {
		updates["codex_secondary_window_minutes"] = *snapshot.SecondaryWindowMinutes
	}
	if snapshot.PrimaryOverSecondaryPercent != nil {
		updates["codex_primary_over_secondary_percent"] = *snapshot.PrimaryOverSecondaryPercent
	}
	updates["codex_usage_updated_at"] = baseTime.Format(time.RFC3339)

	if normalized := snapshot.Normalize(); normalized != nil {
		if normalized.Used5hPercent != nil {
			updates["codex_5h_used_percent"] = *normalized.Used5hPercent
		}
		if normalized.Reset5hSeconds != nil {
			updates["codex_5h_reset_after_seconds"] = *normalized.Reset5hSeconds
		}
		if normalized.Window5hMinutes != nil {
			updates["codex_5h_window_minutes"] = *normalized.Window5hMinutes
		}
		if normalized.Used7dPercent != nil {
			updates["codex_7d_used_percent"] = *normalized.Used7dPercent
		}
		if normalized.Reset7dSeconds != nil {
			updates["codex_7d_reset_after_seconds"] = *normalized.Reset7dSeconds
		}
		if normalized.Window7dMinutes != nil {
			updates["codex_7d_window_minutes"] = *normalized.Window7dMinutes
		}
		if resetAt := codexResetAtRFC3339(baseTime, normalized.Reset5hSeconds); resetAt != nil {
			updates["codex_5h_reset_at"] = *resetAt
		}
		if resetAt := codexResetAtRFC3339(baseTime, normalized.Reset7dSeconds); resetAt != nil {
			updates["codex_7d_reset_at"] = *resetAt
		}
	}

	return updates
}

func codexUsagePercentExhausted(value *float64) bool {
	return value != nil && *value >= 100-1e-9
}

func codexRateLimitResetAtFromSnapshot(snapshot *OpenAICodexUsageSnapshot, fallbackNow time.Time) *time.Time {
	if snapshot == nil {
		return nil
	}
	normalized := snapshot.Normalize()
	if normalized == nil {
		return nil
	}

	baseTime := codexSnapshotBaseTime(snapshot, fallbackNow)
	if codexUsagePercentExhausted(normalized.Used7dPercent) && normalized.Reset7dSeconds != nil {
		resetAt := baseTime.Add(time.Duration(*normalized.Reset7dSeconds) * time.Second)
		return &resetAt
	}
	if codexUsagePercentExhausted(normalized.Used5hPercent) && normalized.Reset5hSeconds != nil {
		resetAt := baseTime.Add(time.Duration(*normalized.Reset5hSeconds) * time.Second)
		return &resetAt
	}
	return nil
}

func codexRateLimitResetAtFromExtra(extra map[string]any, now time.Time) *time.Time {
	if len(extra) == 0 {
		return nil
	}
	if progress := buildCodexUsageProgressFromExtra(extra, "7d", now); progress != nil && codexUsagePercentExhausted(&progress.Utilization) && progress.ResetsAt != nil && now.Before(*progress.ResetsAt) {
		resetAt := progress.ResetsAt.UTC()
		return &resetAt
	}
	if progress := buildCodexUsageProgressFromExtra(extra, "5h", now); progress != nil && codexUsagePercentExhausted(&progress.Utilization) && progress.ResetsAt != nil && now.Before(*progress.ResetsAt) {
		resetAt := progress.ResetsAt.UTC()
		return &resetAt
	}
	return nil
}

func applyOpenAICodexRateLimitFromExtra(account *Account, now time.Time) (*time.Time, bool) {
	if account == nil || !account.IsOpenAI() {
		return nil, false
	}
	resetAt := codexRateLimitResetAtFromExtra(account.Extra, now)
	if resetAt == nil {
		return nil, false
	}
	if account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt) && !account.RateLimitResetAt.Before(*resetAt) {
		return account.RateLimitResetAt, false
	}
	account.RateLimitResetAt = resetAt
	return resetAt, true
}

func syncOpenAICodexRateLimitFromExtra(ctx context.Context, repo AccountRepository, account *Account, now time.Time) *time.Time {
	resetAt, changed := applyOpenAICodexRateLimitFromExtra(account, now)
	if !changed || resetAt == nil || repo == nil || account == nil || account.ID <= 0 {
		return resetAt
	}
	_ = repo.SetRateLimited(ctx, account.ID, *resetAt)
	return resetAt
}

// updateCodexUsageSnapshot saves the Codex usage snapshot to account Extra and rate-limit state.
func (s *OpenAIGatewayService) updateCodexUsageSnapshot(ctx context.Context, accountID int64, snapshot *OpenAICodexUsageSnapshot) {
	if snapshot == nil || s == nil || s.accountRepo == nil {
		return
	}

	now := time.Now()
	updates := buildCodexUsageExtraUpdates(snapshot, now)
	resetAt := codexRateLimitResetAtFromSnapshot(snapshot, now)
	if len(updates) == 0 && resetAt == nil {
		return
	}

	shouldPersistUpdates := len(updates) > 0 && s.getCodexSnapshotThrottle().Allow(accountID, now)
	if !shouldPersistUpdates && resetAt == nil {
		return
	}

	go func() {
		updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shouldPersistUpdates {
			_ = s.accountRepo.UpdateExtra(updateCtx, accountID, updates)
		}
		if resetAt != nil {
			_ = s.accountRepo.SetRateLimited(updateCtx, accountID, *resetAt)
		}
	}()
}

func (s *OpenAIGatewayService) UpdateCodexUsageSnapshotFromHeaders(ctx context.Context, accountID int64, headers http.Header) {
	if accountID <= 0 || headers == nil {
		return
	}
	if snapshot := ParseCodexRateLimitHeaders(headers); snapshot != nil {
		s.updateCodexUsageSnapshot(ctx, accountID, snapshot)
	}
}
