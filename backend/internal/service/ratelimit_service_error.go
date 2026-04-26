package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
)

type ErrorPolicyResult int

const (
	ErrorPolicyNone ErrorPolicyResult = iota
	ErrorPolicySkipped
	ErrorPolicyMatched
	ErrorPolicyTempUnscheduled
)

type AccountHealthAction string

const (
	AccountHealthActionNone        AccountHealthAction = "none"
	AccountHealthActionSetError    AccountHealthAction = "set_error"
	AccountHealthActionOAuth401    AccountHealthAction = "oauth_401"
	AccountHealthActionRateLimit   AccountHealthAction = "rate_limit"
	AccountHealthActionOverload    AccountHealthAction = "overload"
	AccountHealthActionCustomError AccountHealthAction = "custom_error"
	AccountHealthActionLogOnly     AccountHealthAction = "log_only"
)

type AccountHealthDecision struct {
	Action          AccountHealthAction
	FailureKind     string
	ErrorMessage    string
	UpstreamMessage string
	ShouldDisable   bool
}

func (s *RateLimitService) CheckErrorPolicy(ctx context.Context, account *Account, statusCode int, responseBody []byte) ErrorPolicyResult {
	if account.IsCustomErrorCodesEnabled() {
		if account.ShouldHandleErrorCode(statusCode) {
			return ErrorPolicyMatched
		}
		slog.Info("account_error_code_skipped", "account_id", account.ID, "status_code", statusCode)
		return ErrorPolicySkipped
	}
	if account.IsPoolMode() {
		return ErrorPolicySkipped
	}
	if s.tryTempUnschedulable(ctx, account, statusCode, responseBody) {
		return ErrorPolicyTempUnscheduled
	}
	return ErrorPolicyNone
}

func (s *RateLimitService) HandleUpstreamError(ctx context.Context, account *Account, statusCode int, headers http.Header, responseBody []byte) (shouldDisable bool) {
	customErrorCodesEnabled := account.IsCustomErrorCodesEnabled()
	if account.IsPoolMode() && !customErrorCodesEnabled {
		slog.Info("pool_mode_error_skipped", "account_id", account.ID, "status_code", statusCode)
		return false
	}
	if !account.ShouldHandleErrorCode(statusCode) {
		slog.Info("account_error_code_skipped", "account_id", account.ID, "status_code", statusCode)
		return false
	}
	if statusCode != http.StatusUnauthorized && s.tryTempUnschedulable(ctx, account, statusCode, responseBody) {
		return true
	}

	decision := s.classifyAccountHealthDecision(account, statusCode, responseBody, customErrorCodesEnabled)

	if statusCode == http.StatusForbidden {
		logger.LegacyPrintf(
			"service.ratelimit",
			"[HandleUpstreamErrorRaw] account_id=%d platform=%s type=%s status=403 request_id=%s cf_ray=%s upstream_msg=%s raw_body=%s",
			account.ID,
			account.Platform,
			account.Type,
			strings.TrimSpace(headers.Get("x-request-id")),
			strings.TrimSpace(headers.Get("cf-ray")),
			decision.UpstreamMessage,
			truncateForLog(responseBody, 1024),
		)
	}

	s.logAccountHealthDecision(account, statusCode, headers, decision)

	switch decision.Action {
	case AccountHealthActionSetError:
		s.handleAuthError(ctx, account, decision.ErrorMessage)
		return decision.ShouldDisable
	case AccountHealthActionOAuth401:
		s.handleOAuth401(ctx, account, decision.UpstreamMessage)
		return decision.ShouldDisable
	case AccountHealthActionRateLimit:
		s.handle429(ctx, account, headers, responseBody)
		return false
	case AccountHealthActionOverload:
		s.handle529(ctx, account)
		return false
	case AccountHealthActionCustomError:
		message := decision.ErrorMessage
		if message == "" {
			message = "Custom error code triggered"
		}
		s.handleCustomErrorCode(ctx, account, statusCode, message)
		return decision.ShouldDisable
	case AccountHealthActionLogOnly:
		if statusCode >= http.StatusInternalServerError {
			slog.Warn("account_upstream_error", "account_id", account.ID, "status_code", statusCode)
		}
		return false
	default:
		return false
	}
}

func (s *RateLimitService) classifyAccountHealthDecision(account *Account, statusCode int, responseBody []byte, customErrorCodesEnabled bool) AccountHealthDecision {
	upstreamMessage := strings.TrimSpace(extractUpstreamErrorMessage(responseBody))
	upstreamMessage = sanitizeUpstreamErrorMessage(upstreamMessage)
	if upstreamMessage != "" {
		upstreamMessage = truncateForLog([]byte(upstreamMessage), 512)
	}

	decision := AccountHealthDecision{
		Action:          AccountHealthActionNone,
		FailureKind:     "none",
		UpstreamMessage: upstreamMessage,
	}

	switch statusCode {
	case http.StatusBadRequest:
		lowerMessage := strings.ToLower(upstreamMessage)
		switch {
		case strings.Contains(lowerMessage, "organization has been disabled"):
			decision.Action = AccountHealthActionSetError
			decision.FailureKind = "organization_disabled"
			decision.ErrorMessage = "Organization disabled (400): " + upstreamMessage
			decision.ShouldDisable = true
		case strings.Contains(lowerMessage, "identity verification is required"):
			decision.Action = AccountHealthActionSetError
			decision.FailureKind = "identity_verification_required"
			decision.ErrorMessage = "Identity verification required (400): " + upstreamMessage
			decision.ShouldDisable = true
		case account.Platform == PlatformAnthropic && strings.Contains(lowerMessage, "credit balance"):
			decision.Action = AccountHealthActionSetError
			decision.FailureKind = "credit_balance_exhausted"
			decision.ErrorMessage = "Credit balance exhausted (400): " + upstreamMessage
			decision.ShouldDisable = true
		}
	case http.StatusUnauthorized:
		openAIErrorCode := extractUpstreamErrorCode(responseBody)
		if account.Platform == PlatformOpenAI && (openAIErrorCode == "token_invalidated" || openAIErrorCode == "token_revoked") {
			message := "Token revoked (401): account authentication permanently revoked"
			if upstreamMessage != "" {
				message = "Token revoked (401): " + upstreamMessage
			}
			decision.Action = AccountHealthActionSetError
			decision.FailureKind = openAIErrorCode
			decision.ErrorMessage = message
			decision.ShouldDisable = true
			return decision
		}
		if account.Type == AccountTypeOAuth && account.Platform != PlatformAntigravity {
			decision.Action = AccountHealthActionOAuth401
			decision.FailureKind = "oauth_401"
			decision.ShouldDisable = true
			return decision
		}
		message := "Authentication failed (401): invalid or expired credentials"
		if upstreamMessage != "" {
			message = "Authentication failed (401): " + upstreamMessage
		}
		decision.Action = AccountHealthActionSetError
		decision.FailureKind = "auth_failed"
		decision.ErrorMessage = message
		decision.ShouldDisable = true
	case http.StatusPaymentRequired:
		if account.Platform == PlatformOpenAI && gjson.GetBytes(responseBody, "detail.code").String() == "deactivated_workspace" {
			decision.Action = AccountHealthActionSetError
			decision.FailureKind = "workspace_deactivated"
			decision.ErrorMessage = "Workspace deactivated (402): workspace has been deactivated"
			decision.ShouldDisable = true
			return decision
		}
		message := "Payment required (402): insufficient balance or billing issue"
		if upstreamMessage != "" {
			message = "Payment required (402): " + upstreamMessage
		}
		decision.Action = AccountHealthActionSetError
		decision.FailureKind = "payment_required"
		decision.ErrorMessage = message
		decision.ShouldDisable = true
	case http.StatusForbidden:
		return s.classify403Decision(account, upstreamMessage, responseBody)
	case http.StatusTooManyRequests:
		decision.Action = AccountHealthActionRateLimit
		decision.FailureKind = "rate_limited"
	case 529:
		decision.Action = AccountHealthActionOverload
		decision.FailureKind = "overloaded"
	default:
		if customErrorCodesEnabled {
			message := "Custom error code triggered"
			if upstreamMessage != "" {
				message = upstreamMessage
			}
			decision.Action = AccountHealthActionCustomError
			decision.FailureKind = "custom_error_code"
			decision.ErrorMessage = message
			decision.ShouldDisable = true
			return decision
		}
		if statusCode >= http.StatusInternalServerError {
			decision.Action = AccountHealthActionLogOnly
			decision.FailureKind = "upstream_5xx"
		}
	}
	return decision
}

func (s *RateLimitService) classify403Decision(account *Account, upstreamMsg string, responseBody []byte) AccountHealthDecision {
	decision := AccountHealthDecision{
		Action:          AccountHealthActionSetError,
		FailureKind:     "forbidden",
		UpstreamMessage: upstreamMsg,
		ShouldDisable:   true,
	}
	if account.Platform == PlatformAntigravity {
		switch classifyForbiddenType(string(responseBody)) {
		case forbiddenTypeValidation:
			message := "Validation required (403): account needs Google verification"
			if upstreamMsg != "" {
				message = "Validation required (403): " + upstreamMsg
			}
			if validationURL := extractValidationURL(string(responseBody)); validationURL != "" {
				message += " | validation_url: " + validationURL
			}
			decision.FailureKind = "validation_required"
			decision.ErrorMessage = message
			return decision
		case forbiddenTypeViolation:
			message := "Account violation (403): terms of service violation"
			if upstreamMsg != "" {
				message = "Account violation (403): " + upstreamMsg
			}
			decision.FailureKind = "account_violation"
			decision.ErrorMessage = message
			return decision
		}
	}
	message := "Access forbidden (403): account may be suspended or lack permissions"
	if upstreamMsg != "" {
		message = "Access forbidden (403): " + upstreamMsg
	}
	decision.ErrorMessage = message
	return decision
}

func (s *RateLimitService) logAccountHealthDecision(account *Account, statusCode int, headers http.Header, decision AccountHealthDecision) {
	if account == nil {
		return
	}
	var cooldownUntil any
	if account.RateLimitResetAt != nil {
		cooldownUntil = *account.RateLimitResetAt
	}
	slog.Info("account_health_decision",
		"account_id", account.ID,
		"platform", account.Platform,
		"status_code", statusCode,
		"decision", decision.Action,
		"failure_kind", decision.FailureKind,
		"cooldown_until", cooldownUntil,
		"request_id", strings.TrimSpace(headers.Get("x-request-id")),
	)
}

func (s *RateLimitService) handleOAuth401(ctx context.Context, account *Account, upstreamMessage string) {
	if s.tokenCacheInvalidator != nil {
		if err := s.tokenCacheInvalidator.InvalidateToken(ctx, account); err != nil {
			slog.Warn("oauth_401_invalidate_cache_failed", "account_id", account.ID, "error", err)
		}
	}
	if account.Credentials == nil {
		account.Credentials = make(map[string]any)
	}
	account.Credentials["expires_at"] = time.Now().Format(time.RFC3339)
	if err := persistAccountCredentials(ctx, s.accountRepo, account, account.Credentials); err != nil {
		slog.Warn("oauth_401_force_refresh_update_failed", "account_id", account.ID, "error", err)
	} else {
		slog.Info("oauth_401_force_refresh_set", "account_id", account.ID, "platform", account.Platform)
	}

	message := "Authentication failed (401): invalid or expired credentials"
	if upstreamMessage != "" {
		message = "OAuth 401: " + upstreamMessage
	}
	cooldownMinutes := 10
	if s.cfg != nil && s.cfg.RateLimit.OAuth401CooldownMinutes > 0 {
		cooldownMinutes = s.cfg.RateLimit.OAuth401CooldownMinutes
	}
	if cooldownMinutes <= 0 {
		cooldownMinutes = 10
	}
	now := time.Now()
	until := now.Add(time.Duration(cooldownMinutes) * time.Minute)
	state := &TempUnschedState{
		UntilUnix:       until.Unix(),
		TriggeredAtUnix: now.Unix(),
		StatusCode:      http.StatusUnauthorized,
		ErrorMessage:    message,
	}
	if err := s.accountRepo.SetTempUnschedulable(ctx, account.ID, until, marshalTempUnschedState(state)); err != nil {
		slog.Warn("oauth_401_set_temp_unschedulable_failed", "account_id", account.ID, "error", err)
		return
	}
	s.cacheTempUnschedState(ctx, account.ID, state)
	slog.Info("oauth_401_temp_unschedulable_set",
		"account_id", account.ID,
		"platform", account.Platform,
		"cooldown_until", until,
	)
}

func (s *RateLimitService) GeminiCooldown(ctx context.Context, account *Account) time.Duration {
	if account == nil || s.geminiQuotaService == nil {
		return 5 * time.Minute
	}
	return s.geminiQuotaService.CooldownForAccount(ctx, account)
}

func (s *RateLimitService) handleAuthError(ctx context.Context, account *Account, errorMsg string) {
	if err := s.accountRepo.SetError(ctx, account.ID, errorMsg); err != nil {
		slog.Warn("account_set_error_failed", "account_id", account.ID, "error", err)
		return
	}
	slog.Warn("account_disabled_auth_error", "account_id", account.ID, "error", errorMsg)
}

func (s *RateLimitService) handleCustomErrorCode(ctx context.Context, account *Account, statusCode int, errorMsg string) {
	message := "Custom error code " + strconv.Itoa(statusCode) + ": " + errorMsg
	if err := s.accountRepo.SetError(ctx, account.ID, message); err != nil {
		slog.Warn("account_set_error_failed", "account_id", account.ID, "status_code", statusCode, "error", err)
		return
	}
	slog.Warn("account_disabled_custom_error", "account_id", account.ID, "status_code", statusCode, "error", errorMsg)
}

func (s *RateLimitService) handle429(ctx context.Context, account *Account, headers http.Header, responseBody []byte) {
	if account.Platform == PlatformOpenAI {
		s.persistOpenAICodexSnapshot(ctx, account, headers)
		if resetAt := s.calculateOpenAI429ResetTime(headers); resetAt != nil {
			s.persistRateLimitReset(ctx, account, *resetAt, "", nil)
			slog.Info("openai_account_rate_limited", "account_id", account.ID, "reset_at", *resetAt)
			return
		}
	}

	if result := calculateAnthropic429ResetTime(headers); result != nil {
		windowEnd := result.resetAt
		if result.fiveHourReset != nil {
			windowEnd = *result.fiveHourReset
		}
		windowStart := windowEnd.Add(-5 * time.Hour)
		s.persistRateLimitReset(ctx, account, result.resetAt, "rejected", &sessionWindowUpdate{
			start: &windowStart,
			end:   &windowEnd,
		})
		slog.Info("anthropic_account_rate_limited", "account_id", account.ID, "reset_at", result.resetAt, "reset_in", time.Until(result.resetAt).Truncate(time.Second))
		return
	}

	resetTimestamp := headers.Get("anthropic-ratelimit-unified-reset")
	if resetTimestamp == "" {
		if resetAt := s.platformSpecific429Reset(account, responseBody); resetAt != nil {
			s.persistRateLimitReset(ctx, account, *resetAt, "", nil)
			slog.Info("account_rate_limited", "account_id", account.ID, "platform", account.Platform, "reset_at", *resetAt, "reset_in", time.Until(*resetAt).Truncate(time.Second))
			return
		}
		if account.Platform == PlatformAnthropic {
			slog.Warn("rate_limit_429_no_reset_time_skipped", "account_id", account.ID, "platform", account.Platform, "reason", "no rate limit reset time in headers, likely not a real rate limit")
			return
		}
		defaultReset := time.Now().Add(5 * time.Minute)
		slog.Warn("rate_limit_no_reset_time", "account_id", account.ID, "platform", account.Platform, "using_default", "5m")
		s.persistRateLimitReset(ctx, account, defaultReset, "", nil)
		return
	}

	ts, err := strconv.ParseInt(resetTimestamp, 10, 64)
	if err != nil {
		slog.Warn("rate_limit_reset_parse_failed", "reset_timestamp", resetTimestamp, "error", err)
		s.persistRateLimitReset(ctx, account, time.Now().Add(5*time.Minute), "", nil)
		return
	}

	resetAt := time.Unix(ts, 0)
	windowStart := resetAt.Add(-5 * time.Hour)
	s.persistRateLimitReset(ctx, account, resetAt, "rejected", &sessionWindowUpdate{
		start: &windowStart,
		end:   &resetAt,
	})
	slog.Info("account_rate_limited", "account_id", account.ID, "reset_at", resetAt)
}

type sessionWindowUpdate struct {
	start *time.Time
	end   *time.Time
}

func (s *RateLimitService) persistRateLimitReset(ctx context.Context, account *Account, resetAt time.Time, status string, window *sessionWindowUpdate) {
	if err := s.accountRepo.SetRateLimited(ctx, account.ID, resetAt); err != nil {
		slog.Warn("rate_limit_set_failed", "account_id", account.ID, "error", err)
		return
	}
	if window == nil {
		return
	}
	if err := s.accountRepo.UpdateSessionWindow(ctx, account.ID, window.start, window.end, status); err != nil {
		slog.Warn("rate_limit_update_session_window_failed", "account_id", account.ID, "error", err)
	}
}

func (s *RateLimitService) platformSpecific429Reset(account *Account, responseBody []byte) *time.Time {
	switch account.Platform {
	case PlatformOpenAI:
		if resetAt := parseOpenAIRateLimitResetTime(responseBody); resetAt != nil {
			resetTime := time.Unix(*resetAt, 0)
			return &resetTime
		}
	case PlatformGemini, PlatformAntigravity:
		if resetAt := ParseGeminiRateLimitResetTime(responseBody); resetAt != nil {
			resetTime := time.Unix(*resetAt, 0)
			return &resetTime
		}
	}
	return nil
}

func (s *RateLimitService) calculateOpenAI429ResetTime(headers http.Header) *time.Time {
	snapshot := ParseCodexRateLimitHeaders(headers)
	if snapshot == nil {
		return nil
	}
	normalized := snapshot.Normalize()
	if normalized == nil {
		return nil
	}

	now := time.Now()
	is7dExhausted := normalized.Used7dPercent != nil && *normalized.Used7dPercent >= 100
	is5hExhausted := normalized.Used5hPercent != nil && *normalized.Used5hPercent >= 100

	if is7dExhausted && normalized.Reset7dSeconds != nil {
		resetAt := now.Add(time.Duration(*normalized.Reset7dSeconds) * time.Second)
		slog.Info("openai_429_7d_limit_exhausted", "reset_after_seconds", *normalized.Reset7dSeconds, "reset_at", resetAt)
		return &resetAt
	}
	if is5hExhausted && normalized.Reset5hSeconds != nil {
		resetAt := now.Add(time.Duration(*normalized.Reset5hSeconds) * time.Second)
		slog.Info("openai_429_5h_limit_exhausted", "reset_after_seconds", *normalized.Reset5hSeconds, "reset_at", resetAt)
		return &resetAt
	}

	var maxResetSeconds int
	if normalized.Reset7dSeconds != nil && *normalized.Reset7dSeconds > maxResetSeconds {
		maxResetSeconds = *normalized.Reset7dSeconds
	}
	if normalized.Reset5hSeconds != nil && *normalized.Reset5hSeconds > maxResetSeconds {
		maxResetSeconds = *normalized.Reset5hSeconds
	}
	if maxResetSeconds > 0 {
		resetAt := now.Add(time.Duration(maxResetSeconds) * time.Second)
		slog.Info("openai_429_using_max_reset", "max_reset_seconds", maxResetSeconds, "reset_at", resetAt)
		return &resetAt
	}

	return nil
}

type anthropic429Result struct {
	resetAt       time.Time
	fiveHourReset *time.Time
}

func calculateAnthropic429ResetTime(headers http.Header) *anthropic429Result {
	reset5hStr := headers.Get("anthropic-ratelimit-unified-5h-reset")
	reset7dStr := headers.Get("anthropic-ratelimit-unified-7d-reset")
	if reset5hStr == "" && reset7dStr == "" {
		return nil
	}

	var reset5h *time.Time
	var reset7d *time.Time
	if ts, err := strconv.ParseInt(reset5hStr, 10, 64); err == nil {
		t := time.Unix(ts, 0)
		reset5h = &t
	}
	if ts, err := strconv.ParseInt(reset7dStr, 10, 64); err == nil {
		t := time.Unix(ts, 0)
		reset7d = &t
	}

	is5hExceeded := isAnthropicWindowExceeded(headers, "5h")
	is7dExceeded := isAnthropicWindowExceeded(headers, "7d")
	slog.Info("anthropic_429_window_analysis", "is_5h_exceeded", is5hExceeded, "is_7d_exceeded", is7dExceeded, "reset_5h", reset5hStr, "reset_7d", reset7dStr)

	var chosen *time.Time
	switch {
	case is5hExceeded && is7dExceeded:
		chosen = reset7d
		if chosen == nil {
			chosen = reset5h
		}
	case is5hExceeded:
		chosen = reset5h
	case is7dExceeded:
		chosen = reset7d
	default:
		chosen = pickSooner(reset5h, reset7d)
	}

	if chosen == nil {
		return nil
	}
	return &anthropic429Result{resetAt: *chosen, fiveHourReset: reset5h}
}

func isAnthropicWindowExceeded(headers http.Header, window string) bool {
	prefix := "anthropic-ratelimit-unified-" + window + "-"
	if surpassedThreshold := headers.Get(prefix + "surpassed-threshold"); strings.EqualFold(surpassedThreshold, "true") {
		return true
	}
	if utilizationValue := headers.Get(prefix + "utilization"); utilizationValue != "" {
		if utilization, err := strconv.ParseFloat(utilizationValue, 64); err == nil && utilization >= 1.0-1e-9 {
			return true
		}
	}
	return false
}

func pickSooner(a, b *time.Time) *time.Time {
	switch {
	case a != nil && b != nil:
		if a.Before(*b) {
			return a
		}
		return b
	case a != nil:
		return a
	default:
		return b
	}
}

func (s *RateLimitService) persistOpenAICodexSnapshot(ctx context.Context, account *Account, headers http.Header) {
	if s == nil || s.accountRepo == nil || account == nil || headers == nil {
		return
	}
	snapshot := ParseCodexRateLimitHeaders(headers)
	if snapshot == nil {
		return
	}
	updates := buildCodexUsageExtraUpdates(snapshot, time.Now())
	if len(updates) == 0 {
		return
	}
	if err := s.accountRepo.UpdateExtra(ctx, account.ID, updates); err != nil {
		slog.Warn("openai_codex_snapshot_persist_failed", "account_id", account.ID, "error", err)
	}
}

func parseOpenAIRateLimitResetTime(body []byte) *int64 {
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}

	errorObject, ok := parsed["error"].(map[string]any)
	if !ok {
		return nil
	}
	errorType, _ := errorObject["type"].(string)
	if errorType != "usage_limit_reached" && errorType != "rate_limit_exceeded" {
		return nil
	}

	if resetsAt, ok := errorObject["resets_at"].(float64); ok {
		ts := int64(resetsAt)
		return &ts
	}
	if resetsAt, ok := errorObject["resets_at"].(string); ok {
		if ts, err := strconv.ParseInt(resetsAt, 10, 64); err == nil {
			return &ts
		}
	}
	if resetsInSeconds, ok := errorObject["resets_in_seconds"].(float64); ok {
		ts := time.Now().Unix() + int64(resetsInSeconds)
		return &ts
	}
	if resetsInSeconds, ok := errorObject["resets_in_seconds"].(string); ok {
		if seconds, err := strconv.ParseInt(resetsInSeconds, 10, 64); err == nil {
			ts := time.Now().Unix() + seconds
			return &ts
		}
	}
	return nil
}

func (s *RateLimitService) handle529(ctx context.Context, account *Account) {
	var settings *OverloadCooldownSettings
	if s.settingService != nil {
		var err error
		settings, err = s.settingService.GetOverloadCooldownSettings(ctx)
		if err != nil {
			slog.Warn("overload_settings_read_failed", "account_id", account.ID, "error", err)
			settings = nil
		}
	}
	if settings == nil {
		cooldown := s.cfg.RateLimit.OverloadCooldownMinutes
		if cooldown <= 0 {
			cooldown = 10
		}
		settings = &OverloadCooldownSettings{Enabled: true, CooldownMinutes: cooldown}
	}
	if !settings.Enabled {
		slog.Info("account_529_ignored", "account_id", account.ID, "reason", "overload_cooldown_disabled")
		return
	}

	cooldownMinutes := settings.CooldownMinutes
	if cooldownMinutes <= 0 {
		cooldownMinutes = 10
	}
	until := time.Now().Add(time.Duration(cooldownMinutes) * time.Minute)
	if err := s.accountRepo.SetOverloaded(ctx, account.ID, until); err != nil {
		slog.Warn("overload_set_failed", "account_id", account.ID, "error", err)
		return
	}
	slog.Info("account_overloaded", "account_id", account.ID, "until", until)
}
