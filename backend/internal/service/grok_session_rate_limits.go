package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"
)

const (
	grokSessionRateLimitsEndpoint  = "/rest/rate-limits"
	grokSessionRateLimitsBodyLimit = 64 * 1024
	grokTierSourceUsageAPI         = "usage_api"
)

type grokSessionRateLimitsResponse struct {
	WindowSizeSeconds int  `json:"windowSizeSeconds"`
	RemainingQueries  *int `json:"remainingQueries"`
	TotalQueries      *int `json:"totalQueries"`
}

type grokSessionRateLimitsResult struct {
	QuotaWindows map[string]any
	StatusCode   int
	Tier         grok.Tier
}

type grokSessionRateLimitsError struct {
	WindowName   string
	StatusCode   int
	ResponseBody string
	Cause        error
}

func (e *grokSessionRateLimitsError) Error() string {
	if e == nil {
		return ""
	}
	window := strings.TrimSpace(e.WindowName)
	if window == "" {
		window = grok.QuotaWindowAuto
	}
	if e.Cause != nil {
		return fmt.Sprintf("grok rate-limits fetch failed for %s: %v", window, e.Cause)
	}
	body := strings.TrimSpace(e.ResponseBody)
	if body == "" {
		return fmt.Sprintf("grok rate-limits API returned %d for %s", e.StatusCode, window)
	}
	return fmt.Sprintf("grok rate-limits API returned %d for %s: %s", e.StatusCode, window, body)
}

func (e *grokSessionRateLimitsError) InvalidCredentials() bool {
	if e == nil {
		return false
	}
	if e.StatusCode != http.StatusBadRequest && e.StatusCode != http.StatusUnauthorized && e.StatusCode != http.StatusForbidden {
		return false
	}
	return grokSessionRateLimitsInvalidCredentialsBody(e.ResponseBody)
}

func grokSessionRateLimitsInvalidCredentialsBody(body string) bool {
	text := strings.ToLower(strings.TrimSpace(body))
	if text == "" {
		return false
	}
	return strings.Contains(text, "invalid-credentials") ||
		strings.Contains(text, "bad-credentials") ||
		strings.Contains(text, "failed to look up session id") ||
		strings.Contains(text, "blocked-user") ||
		strings.Contains(text, "email-domain-rejected") ||
		strings.Contains(text, "session not found") ||
		strings.Contains(text, "account suspended") ||
		strings.Contains(text, "token revoked") ||
		strings.Contains(text, "token expired")
}

func grokRateLimitWindowNames() []string {
	return []string{
		grok.QuotaWindowAuto,
		grok.QuotaWindowFast,
		grok.QuotaWindowExpert,
		grok.QuotaWindowHeavy,
	}
}

func grokRateLimitDefaultWindowSeconds(windowName string) int {
	switch strings.TrimSpace(windowName) {
	case grok.QuotaWindowExpert:
		return 36000
	case grok.QuotaWindowHeavy:
		return 7200
	default:
		return 72000
	}
}

func (s *GrokQuotaSyncService) shouldFetchLiveSessionQuota(account *Account) bool {
	return s != nil &&
		account != nil &&
		account.Type == AccountTypeSession &&
		s.httpUpstream != nil &&
		strings.TrimSpace(account.GetGrokSessionToken()) != ""
}

func (s *GrokQuotaSyncService) fetchLiveSessionQuota(
	ctx context.Context,
	account *Account,
	now time.Time,
) (*grokSessionRateLimitsResult, error) {
	if !s.shouldFetchLiveSessionQuota(account) {
		return nil, nil
	}

	runtimeSettings := DefaultGrokRuntimeSettings()
	if s != nil && s.settingSvc != nil {
		runtimeSettings = s.settingSvc.GetGrokRuntimeSettings(ctx)
	}

	target, err := resolveGrokTransportTargetWithSettings(account, nil, runtimeSettings)
	if err != nil {
		return nil, err
	}
	if target.Kind != grokTransportKindSession {
		return nil, fmt.Errorf("unsupported grok session transport: %s", target.Kind)
	}

	target.URL, target.SessionBaseURL, err = buildGrokSessionTransportURLForPathWithSettings(
		nil,
		runtimeSettings,
		grokSessionRateLimitsEndpoint,
	)
	if err != nil {
		return nil, err
	}

	result := &grokSessionRateLimitsResult{
		QuotaWindows: make(map[string]any, 4),
		StatusCode:   http.StatusOK,
	}

	var firstErr error
	for _, windowName := range grokRateLimitWindowNames() {
		window, statusCode, fetchErr := s.fetchLiveSessionQuotaWindow(ctx, account, target, windowName, now)
		if fetchErr != nil {
			if firstErr == nil {
				firstErr = fetchErr
			}
			var rateErr *grokSessionRateLimitsError
			if errors.As(fetchErr, &rateErr) && rateErr != nil && rateErr.InvalidCredentials() {
				return nil, rateErr
			}
			continue
		}
		if statusCode > 0 {
			result.StatusCode = statusCode
		}
		result.QuotaWindows[windowName] = window
	}

	if len(result.QuotaWindows) == 0 {
		return nil, firstErr
	}
	result.Tier = grokInferTierFromQuotaWindows(result.QuotaWindows)
	return result, nil
}

func (s *GrokQuotaSyncService) fetchLiveSessionQuotaWindow(
	ctx context.Context,
	account *Account,
	target grokTransportTarget,
	windowName string,
	now time.Time,
) (map[string]any, int, error) {
	if s == nil || s.httpUpstream == nil {
		return nil, 0, nil
	}

	payloadBytes, err := json.Marshal(map[string]string{
		"modelName": strings.TrimSpace(windowName),
	})
	if err != nil {
		return nil, 0, &grokSessionRateLimitsError{
			WindowName: windowName,
			Cause:      fmt.Errorf("marshal payload: %w", err),
		}
	}

	req, err := newGrokSessionJSONRequest(
		ctx,
		http.MethodPost,
		target,
		payloadBytes,
		"application/json, text/plain, */*",
	)
	if err != nil {
		return nil, 0, &grokSessionRateLimitsError{
			WindowName: windowName,
			Cause:      fmt.Errorf("build request: %w", err),
		}
	}

	resp, err := s.httpUpstream.DoWithTLS(
		req,
		accountTestProxyURL(account),
		account.ID,
		account.Concurrency,
		s.tlsProfile(account),
	)
	if err != nil {
		return nil, 0, &grokSessionRateLimitsError{
			WindowName: windowName,
			Cause:      fmt.Errorf("request failed: %w", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, grokSessionRateLimitsBodyLimit))
	if readErr != nil {
		return nil, resp.StatusCode, &grokSessionRateLimitsError{
			WindowName: windowName,
			StatusCode: resp.StatusCode,
			Cause:      fmt.Errorf("read response: %w", readErr),
		}
	}

	bodyText := strings.TrimSpace(string(body))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, resp.StatusCode, &grokSessionRateLimitsError{
			WindowName:   windowName,
			StatusCode:   resp.StatusCode,
			ResponseBody: bodyText,
		}
	}

	var payload grokSessionRateLimitsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, resp.StatusCode, &grokSessionRateLimitsError{
			WindowName:   windowName,
			StatusCode:   resp.StatusCode,
			ResponseBody: bodyText,
			Cause:        fmt.Errorf("decode response: %w", err),
		}
	}
	if payload.RemainingQueries == nil {
		return nil, resp.StatusCode, &grokSessionRateLimitsError{
			WindowName:   windowName,
			StatusCode:   resp.StatusCode,
			ResponseBody: bodyText,
			Cause:        fmt.Errorf("response missing remainingQueries"),
		}
	}

	total := *payload.RemainingQueries
	if payload.TotalQueries != nil {
		total = *payload.TotalQueries
	}
	windowSeconds := payload.WindowSizeSeconds
	if windowSeconds <= 0 {
		windowSeconds = grokRateLimitDefaultWindowSeconds(windowName)
	}

	window := map[string]any{
		"remaining":      *payload.RemainingQueries,
		"total":          total,
		"window_seconds": windowSeconds,
		"source":         grok.QuotaSourceLive,
	}
	if windowSeconds > 0 {
		window["reset_at"] = now.Add(time.Duration(windowSeconds) * time.Second).UTC().Format(time.RFC3339)
	}

	return window, resp.StatusCode, nil
}

func (s *GrokQuotaSyncService) tlsProfile(account *Account) *tlsfingerprint.Profile {
	if s == nil || s.tlsFPProfileService == nil {
		return nil
	}
	return s.tlsFPProfileService.ResolveTLSProfile(account)
}
