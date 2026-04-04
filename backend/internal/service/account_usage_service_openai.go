package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	httppool "github.com/senran-N/sub2api/internal/pkg/httpclient"
	openaipkg "github.com/senran-N/sub2api/internal/pkg/openai"
)

func (s *AccountUsageService) getOpenAIUsage(ctx context.Context, account *Account) (*UsageInfo, error) {
	now := time.Now()
	usage := &UsageInfo{UpdatedAt: &now}

	if account == nil {
		return usage, nil
	}
	syncOpenAICodexRateLimitFromExtra(ctx, s.accountRepo, account, now)

	if progress := buildCodexUsageProgressFromExtra(account.Extra, "5h", now); progress != nil {
		usage.FiveHour = progress
	}
	if progress := buildCodexUsageProgressFromExtra(account.Extra, "7d", now); progress != nil {
		usage.SevenDay = progress
	}

	if shouldRefreshOpenAICodexSnapshot(account, usage, now) && s.shouldProbeOpenAICodexSnapshot(account.ID, now) {
		if updates, resetAt, err := s.probeOpenAICodexSnapshot(ctx, account); err == nil && (len(updates) > 0 || resetAt != nil) {
			mergeAccountExtra(account, updates)
			if resetAt != nil {
				account.RateLimitResetAt = resetAt
			}
			if usage.UpdatedAt == nil {
				usage.UpdatedAt = &now
			}
			if progress := buildCodexUsageProgressFromExtra(account.Extra, "5h", now); progress != nil {
				usage.FiveHour = progress
			}
			if progress := buildCodexUsageProgressFromExtra(account.Extra, "7d", now); progress != nil {
				usage.SevenDay = progress
			}
		}
	}

	if s.usageLogRepo == nil {
		return usage, nil
	}

	if stats, err := s.usageLogRepo.GetAccountWindowStats(ctx, account.ID, now.Add(-5*time.Hour)); err == nil {
		if usage.FiveHour == nil {
			usage.FiveHour = &UsageProgress{Utilization: 0}
		}
		usage.FiveHour.WindowStats = windowStatsFromAccountStats(stats)
	}

	if stats, err := s.usageLogRepo.GetAccountWindowStats(ctx, account.ID, now.Add(-7*24*time.Hour)); err == nil {
		if usage.SevenDay == nil {
			usage.SevenDay = &UsageProgress{Utilization: 0}
		}
		usage.SevenDay.WindowStats = windowStatsFromAccountStats(stats)
	}

	return usage, nil
}

func shouldRefreshOpenAICodexSnapshot(account *Account, usage *UsageInfo, now time.Time) bool {
	if account == nil {
		return false
	}
	if usage == nil {
		return true
	}
	if usage.FiveHour == nil || usage.SevenDay == nil {
		return true
	}
	if account.IsRateLimited() {
		return true
	}
	return isOpenAICodexSnapshotStale(account, now)
}

func isOpenAICodexSnapshotStale(account *Account, now time.Time) bool {
	if account == nil || !account.IsOpenAIOAuth() || !account.IsOpenAIResponsesWebSocketV2Enabled() {
		return false
	}
	if account.Extra == nil {
		return true
	}
	raw, ok := account.Extra["codex_usage_updated_at"]
	if !ok {
		return true
	}
	timestamp, err := parseTime(fmt.Sprint(raw))
	if err != nil {
		return true
	}
	return now.Sub(timestamp) >= openAIProbeCacheTTL
}

func (s *AccountUsageService) shouldProbeOpenAICodexSnapshot(accountID int64, now time.Time) bool {
	if s == nil || s.cache == nil || accountID <= 0 {
		return true
	}
	if cached, ok := s.cache.openAIProbeCache.Load(accountID); ok {
		if timestamp, ok := cached.(time.Time); ok && now.Sub(timestamp) < openAIProbeCacheTTL {
			return false
		}
	}
	s.cache.openAIProbeCache.Store(accountID, now)
	return true
}

func (s *AccountUsageService) probeOpenAICodexSnapshot(ctx context.Context, account *Account) (map[string]any, *time.Time, error) {
	if account == nil || !account.IsOAuth() {
		return nil, nil, nil
	}

	accessToken := account.GetOpenAIAccessToken()
	if accessToken == "" {
		return nil, nil, fmt.Errorf("no access token available")
	}

	payload := createOpenAITestPayload(openaipkg.DefaultTestModel, true, "")
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal openai probe payload: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, chatgptCodexURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("create openai probe request: %w", err)
	}
	req.Host = "chatgpt.com"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("OpenAI-Beta", "responses=experimental")
	req.Header.Set("Originator", "codex_cli_rs")
	req.Header.Set("Version", openAICodexProbeVersion)
	req.Header.Set("User-Agent", codexCLIUserAgent)
	if s.identityCache != nil {
		if fp, fpErr := s.identityCache.GetFingerprint(reqCtx, account.ID); fpErr == nil && fp != nil && strings.TrimSpace(fp.UserAgent) != "" {
			req.Header.Set("User-Agent", strings.TrimSpace(fp.UserAgent))
		}
	}
	if chatgptAccountID := account.GetChatGPTAccountID(); chatgptAccountID != "" {
		req.Header.Set("chatgpt-account-id", chatgptAccountID)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	client, err := httppool.GetClient(httppool.Options{
		ProxyURL:              proxyURL,
		Timeout:               15 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build openai probe client: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("openai codex probe request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	updates, resetAt, err := extractOpenAICodexProbeSnapshot(resp)
	if err != nil {
		return nil, nil, err
	}
	if len(updates) == 0 && resetAt == nil {
		return nil, nil, nil
	}

	s.persistOpenAICodexProbeSnapshot(account.ID, updates, resetAt)
	return updates, resetAt, nil
}

func (s *AccountUsageService) persistOpenAICodexProbeSnapshot(accountID int64, updates map[string]any, resetAt *time.Time) {
	if s == nil || s.accountRepo == nil || accountID <= 0 {
		return
	}
	if len(updates) == 0 && resetAt == nil {
		return
	}

	go func() {
		updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if len(updates) > 0 {
			_ = s.accountRepo.UpdateExtra(updateCtx, accountID, updates)
		}
		if resetAt != nil {
			_ = s.accountRepo.SetRateLimited(updateCtx, accountID, *resetAt)
		}
	}()
}

func extractOpenAICodexProbeSnapshot(resp *http.Response) (map[string]any, *time.Time, error) {
	if resp == nil {
		return nil, nil, nil
	}
	if snapshot := ParseCodexRateLimitHeaders(resp.Header); snapshot != nil {
		baseTime := time.Now()
		updates := buildCodexUsageExtraUpdates(snapshot, baseTime)
		resetAt := codexRateLimitResetAtFromSnapshot(snapshot, baseTime)
		if len(updates) > 0 {
			return updates, resetAt, nil
		}
		return nil, resetAt, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("openai codex probe returned status %d", resp.StatusCode)
	}
	return nil, nil, nil
}

func extractOpenAICodexProbeUpdates(resp *http.Response) (map[string]any, error) {
	updates, _, err := extractOpenAICodexProbeSnapshot(resp)
	return updates, err
}

func mergeAccountExtra(account *Account, updates map[string]any) {
	if account == nil || len(updates) == 0 {
		return
	}
	if account.Extra == nil {
		account.Extra = make(map[string]any, len(updates))
	}
	for key, value := range updates {
		account.Extra[key] = value
	}
}

func buildCodexUsageProgressFromExtra(extra map[string]any, window string, now time.Time) *UsageProgress {
	if len(extra) == 0 {
		return nil
	}

	var usedPercentKey string
	var resetAfterKey string
	var resetAtKey string

	switch window {
	case "5h":
		usedPercentKey = "codex_5h_used_percent"
		resetAfterKey = "codex_5h_reset_after_seconds"
		resetAtKey = "codex_5h_reset_at"
	case "7d":
		usedPercentKey = "codex_7d_used_percent"
		resetAfterKey = "codex_7d_reset_after_seconds"
		resetAtKey = "codex_7d_reset_at"
	default:
		return nil
	}

	usedRaw, ok := extra[usedPercentKey]
	if !ok {
		return nil
	}

	progress := &UsageProgress{Utilization: parseExtraFloat64(usedRaw)}
	if resetAtRaw, ok := extra[resetAtKey]; ok {
		if resetAt, err := parseTime(fmt.Sprint(resetAtRaw)); err == nil {
			progress.ResetsAt = &resetAt
			progress.RemainingSeconds = int(time.Until(resetAt).Seconds())
			if progress.RemainingSeconds < 0 {
				progress.RemainingSeconds = 0
			}
		}
	}
	if progress.ResetsAt == nil {
		if resetAfterSeconds := parseExtraInt(extra[resetAfterKey]); resetAfterSeconds > 0 {
			baseTime := now
			if updatedAtRaw, ok := extra["codex_usage_updated_at"]; ok {
				if updatedAt, err := parseTime(fmt.Sprint(updatedAtRaw)); err == nil {
					baseTime = updatedAt
				}
			}
			resetAt := baseTime.Add(time.Duration(resetAfterSeconds) * time.Second)
			progress.ResetsAt = &resetAt
			progress.RemainingSeconds = int(time.Until(resetAt).Seconds())
			if progress.RemainingSeconds < 0 {
				progress.RemainingSeconds = 0
			}
		}
	}

	if progress.ResetsAt != nil && !now.Before(*progress.ResetsAt) {
		progress.Utilization = 0
	}
	return progress
}
