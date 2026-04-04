package service

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type sessionWindowPlan struct {
	start             *time.Time
	end               *time.Time
	resetUsageSamples bool
}

// UpdateSessionWindow 从成功响应更新 5h 窗口状态。
func (s *RateLimitService) UpdateSessionWindow(ctx context.Context, account *Account, headers http.Header) {
	status := headers.Get("anthropic-ratelimit-unified-5h-status")
	if status == "" {
		return
	}

	now := time.Now()
	plan := s.buildSessionWindowPlan(account, headers, status, now)
	if plan.resetUsageSamples {
		s.clearSessionWindowUsageSamples(ctx, account.ID)
	}
	if err := s.accountRepo.UpdateSessionWindow(ctx, account.ID, plan.start, plan.end, status); err != nil {
		slog.Warn("session_window_update_failed", "account_id", account.ID, "error", err)
	}

	s.persistPassiveUsageSample(ctx, account.ID, headers, now)
	if status == "allowed" && account.IsRateLimited() {
		if err := s.ClearRateLimit(ctx, account.ID); err != nil {
			slog.Warn("rate_limit_clear_failed", "account_id", account.ID, "error", err)
		}
	}
}

func (s *RateLimitService) buildSessionWindowPlan(account *Account, headers http.Header, status string, now time.Time) sessionWindowPlan {
	needInitWindow := account.SessionWindowEnd == nil || now.After(*account.SessionWindowEnd)
	end := s.parseSessionWindowHeaderEnd(account.ID, headers, now)

	switch {
	case end != nil && (needInitWindow || account.SessionWindowEnd == nil || !end.Equal(*account.SessionWindowEnd)):
		start := end.Add(-5 * time.Hour)
		slog.Info("account_session_window_from_header", "account_id", account.ID, "window_start", start, "window_end", *end, "status", status)
		return sessionWindowPlan{
			start:             &start,
			end:               end,
			resetUsageSamples: needInitWindow,
		}
	case needInitWindow && (status == "allowed" || status == "allowed_warning"):
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		predictedEnd := start.Add(5 * time.Hour)
		slog.Info("account_session_window_initialized", "account_id", account.ID, "window_start", start, "window_end", predictedEnd, "status", status)
		return sessionWindowPlan{
			start:             &start,
			end:               &predictedEnd,
			resetUsageSamples: true,
		}
	default:
		return sessionWindowPlan{}
	}
}

func (s *RateLimitService) parseSessionWindowHeaderEnd(accountID int64, headers http.Header, now time.Time) *time.Time {
	resetValue := headers.Get("anthropic-ratelimit-unified-5h-reset")
	if resetValue == "" {
		return nil
	}

	resetUnix, err := strconv.ParseInt(resetValue, 10, 64)
	if err != nil {
		slog.Warn("account_session_window_header_parse_failed", "account_id", accountID, "raw_reset", resetValue, "error", err)
		return nil
	}
	if resetUnix > 1e11 {
		slog.Warn("account_session_window_header_millis_detected", "account_id", accountID, "raw_reset", resetValue)
		resetUnix /= 1000
	}

	end := time.Unix(resetUnix, 0)
	minAllowed := now.Add(-5 * time.Hour)
	maxAllowed := now.Add(7 * 24 * time.Hour)
	if end.Before(minAllowed) || end.After(maxAllowed) {
		slog.Warn("account_session_window_header_out_of_range", "account_id", accountID, "raw_reset", resetValue, "parsed_end", end)
		return nil
	}
	return &end
}

func (s *RateLimitService) clearSessionWindowUsageSamples(ctx context.Context, accountID int64) {
	_ = s.accountRepo.UpdateExtra(ctx, accountID, map[string]any{
		"session_window_utilization":   nil,
		"passive_usage_7d_utilization": nil,
		"passive_usage_7d_reset":       nil,
		"passive_usage_sampled_at":     nil,
	})
}

func (s *RateLimitService) persistPassiveUsageSample(ctx context.Context, accountID int64, headers http.Header, sampledAt time.Time) {
	extraUpdates := make(map[string]any, 4)
	if utilization := parseAnthropicUtilizationHeader(headers, "anthropic-ratelimit-unified-5h-utilization"); utilization != nil {
		extraUpdates["session_window_utilization"] = *utilization
	}
	if utilization := parseAnthropicUtilizationHeader(headers, "anthropic-ratelimit-unified-7d-utilization"); utilization != nil {
		extraUpdates["passive_usage_7d_utilization"] = *utilization
	}
	if resetAt := parseAnthropicResetHeader(headers, "anthropic-ratelimit-unified-7d-reset"); resetAt != nil {
		extraUpdates["passive_usage_7d_reset"] = resetAt.Unix()
	}
	if len(extraUpdates) == 0 {
		return
	}

	extraUpdates["passive_usage_sampled_at"] = sampledAt.UTC().Format(time.RFC3339)
	if err := s.accountRepo.UpdateExtra(ctx, accountID, extraUpdates); err != nil {
		slog.Warn("passive_usage_update_failed", "account_id", accountID, "error", err)
	}
}

func parseAnthropicUtilizationHeader(headers http.Header, key string) *float64 {
	value := headers.Get(key)
	if value == "" {
		return nil
	}
	utilization, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &utilization
}

func parseAnthropicResetHeader(headers http.Header, key string) *time.Time {
	value := headers.Get(key)
	if value == "" {
		return nil
	}

	resetUnix, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	if resetUnix > 1e11 {
		resetUnix /= 1000
	}

	resetAt := time.Unix(resetUnix, 0)
	return &resetAt
}
