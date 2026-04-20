package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"
)

// GetUsage 获取账号使用量。
// OAuth 账号调用上游 API，Setup Token 基于本地窗口数据估算，其他平台走各自 provider。
func (s *AccountUsageService) GetUsage(ctx context.Context, accountID int64) (*UsageInfo, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account failed: %w", err)
	}

	usage, err := s.getPlatformUsage(ctx, account)
	if err == nil {
		s.tryClearRecoverableAccountError(ctx, account)
	}
	return usage, err
}

func (s *AccountUsageService) getPlatformUsage(ctx context.Context, account *Account) (*UsageInfo, error) {
	switch {
	case account.Platform == PlatformOpenAI && account.Type == AccountTypeOAuth:
		return s.getOpenAIUsage(ctx, account)
	case account.Platform == PlatformGemini:
		return s.getGeminiUsage(ctx, account)
	case NormalizeCompatibleGatewayPlatform(account.Platform) == PlatformGrok:
		return s.getGrokUsage(ctx, account)
	case account.Platform == PlatformAntigravity:
		return s.getAntigravityUsage(ctx, account)
	case account.CanGetUsage():
		return s.getAnthropicOAuthUsage(ctx, account)
	case account.Type == AccountTypeSetupToken:
		usage := s.estimateSetupTokenUsage(account)
		s.addWindowStats(ctx, account, usage)
		return usage, nil
	default:
		return nil, fmt.Errorf("account type %s does not support usage query", account.Type)
	}
}

func (s *AccountUsageService) getAnthropicOAuthUsage(ctx context.Context, account *Account) (*UsageInfo, error) {
	apiResp, err := s.loadAnthropicUsageResponse(ctx, account)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	usage := s.buildUsageInfo(apiResp, &now)
	s.addWindowStats(ctx, account, usage)
	s.syncActiveToPassive(ctx, account.ID, usage)
	return usage, nil
}

func (s *AccountUsageService) loadAnthropicUsageResponse(ctx context.Context, account *Account) (*ClaudeUsageResponse, error) {
	if cachedResp, cachedErr, ok := s.cachedAnthropicUsageEntry(account.ID); ok {
		if cachedErr != nil {
			return nil, cachedErr
		}
		return cachedResp, nil
	}

	if cached := s.cachedAnthropicUsageResponse(account.ID); cached != nil {
		return cached, nil
	}

	jitter := time.Duration(rand.Int64N(int64(apiQueryMaxJitter)))
	select {
	case <-time.After(jitter):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	flightKey := fmt.Sprintf("usage:%d", account.ID)
	result, err, _ := s.cache.apiFlight.Do(flightKey, func() (any, error) {
		if cachedResp, cachedErr, ok := s.cachedAnthropicUsageEntry(account.ID); ok {
			if cachedErr != nil {
				return nil, cachedErr
			}
			return cachedResp, nil
		}

		resp, fetchErr := s.fetchOAuthUsageRaw(ctx, account)
		if fetchErr != nil {
			s.cache.apiCache.Store(account.ID, &apiUsageCache{
				err:       fetchErr,
				timestamp: time.Now(),
			})
			return nil, fetchErr
		}

		s.cache.apiCache.Store(account.ID, &apiUsageCache{
			response:  resp,
			timestamp: time.Now(),
		})
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	response, _ := result.(*ClaudeUsageResponse)
	return response, nil
}

func (s *AccountUsageService) cachedAnthropicUsageResponse(accountID int64) *ClaudeUsageResponse {
	response, _, ok := s.cachedAnthropicUsageEntry(accountID)
	if !ok {
		return nil
	}
	return response
}

func (s *AccountUsageService) cachedAnthropicUsageEntry(accountID int64) (*ClaudeUsageResponse, error, bool) {
	if s == nil || s.cache == nil {
		return nil, nil, false
	}

	cached, ok := s.cache.apiCache.Load(accountID)
	if !ok {
		return nil, nil, false
	}

	entry, ok := cached.(*apiUsageCache)
	if !ok {
		return nil, nil, false
	}

	age := time.Since(entry.timestamp)
	if entry.err != nil && age < apiErrorCacheTTL {
		return nil, entry.err, true
	}
	if entry.response != nil && age < apiCacheTTL {
		return entry.response, nil, true
	}
	return nil, nil, false
}

// GetPassiveUsage 从 Account.Extra 中的被动采样数据构建 UsageInfo，不调用外部 API。
// 仅适用于 Anthropic OAuth / SetupToken 账号。
func (s *AccountUsageService) GetPassiveUsage(ctx context.Context, accountID int64) (*UsageInfo, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account failed: %w", err)
	}

	if !account.IsAnthropicOAuthOrSetupToken() {
		return nil, fmt.Errorf("passive usage only supported for Anthropic OAuth/SetupToken accounts")
	}

	info := s.estimateSetupTokenUsage(account)
	info.Source = "passive"

	if raw, ok := account.Extra["passive_usage_sampled_at"]; ok {
		if str, ok := raw.(string); ok {
			if sampledAt, parseErr := time.Parse(time.RFC3339, str); parseErr == nil {
				info.UpdatedAt = &sampledAt
			}
		}
	}

	util7d := parseExtraFloat64(account.Extra["passive_usage_7d_utilization"])
	reset7dRaw := parseExtraFloat64(account.Extra["passive_usage_7d_reset"])
	if util7d > 0 || reset7dRaw > 0 {
		var resetAt *time.Time
		var remaining int
		if reset7dRaw > 0 {
			resetTime := time.Unix(int64(reset7dRaw), 0)
			resetAt = &resetTime
			remaining = int(time.Until(resetTime).Seconds())
			if remaining < 0 {
				remaining = 0
			}
		}
		info.SevenDay = &UsageProgress{
			Utilization:      util7d * 100,
			ResetsAt:         resetAt,
			RemainingSeconds: remaining,
		}
	}

	s.addWindowStats(ctx, account, info)
	return info, nil
}

// syncActiveToPassive 将主动查询的最新数据回写到 Extra 被动缓存，
// 这样下次被动加载时能看到最新值。
func (s *AccountUsageService) syncActiveToPassive(ctx context.Context, accountID int64, usage *UsageInfo) {
	extraUpdates := make(map[string]any, 4)

	if usage.FiveHour != nil {
		extraUpdates["session_window_utilization"] = usage.FiveHour.Utilization / 100
	}
	if usage.SevenDay != nil {
		extraUpdates["passive_usage_7d_utilization"] = usage.SevenDay.Utilization / 100
		if usage.SevenDay.ResetsAt != nil {
			extraUpdates["passive_usage_7d_reset"] = usage.SevenDay.ResetsAt.Unix()
		}
	}

	if len(extraUpdates) == 0 {
		return
	}

	extraUpdates["passive_usage_sampled_at"] = time.Now().UTC().Format(time.RFC3339)
	if err := s.accountRepo.UpdateExtra(ctx, accountID, extraUpdates); err != nil {
		slog.Warn("sync_active_to_passive_failed", "account_id", accountID, "error", err)
	}
}

func (s *AccountUsageService) tryClearRecoverableAccountError(ctx context.Context, account *Account) {
	if account == nil || account.Status != StatusError {
		return
	}

	message := strings.ToLower(strings.TrimSpace(account.ErrorMessage))
	if message == "" {
		return
	}

	if !strings.Contains(message, "token refresh failed") &&
		!strings.Contains(message, "invalid_client") &&
		!strings.Contains(message, "missing_project_id") &&
		!strings.Contains(message, "unauthenticated") {
		return
	}

	if err := s.accountRepo.ClearError(ctx, account.ID); err != nil {
		log.Printf("[usage] failed to clear recoverable account error for account %d: %v", account.ID, err)
		return
	}

	account.Status = StatusActive
	account.ErrorMessage = ""
}
