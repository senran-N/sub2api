package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

const (
	defaultOAuthRequestPathRefreshTimeout = 3 * time.Second
	defaultOAuthLockWaitTimeout           = 250 * time.Millisecond
	// tokenRefreshTempUnschedDuration is the default request/background
	// quarantine duration after retryable token refresh failures.
	tokenRefreshTempUnschedDuration = 10 * time.Minute
)

var (
	errOAuthRefreshLockHeld = errors.New("oauth refresh lock held")
)

type OAuthRefreshFailureKind string

const (
	OAuthRefreshFailureNone           OAuthRefreshFailureKind = ""
	OAuthRefreshFailurePermanent      OAuthRefreshFailureKind = "permanent"
	OAuthRefreshFailureTransient      OAuthRefreshFailureKind = "transient"
	OAuthRefreshFailureLockHeld       OAuthRefreshFailureKind = "lock_held"
	OAuthRefreshFailureNoRefreshToken OAuthRefreshFailureKind = "no_refresh_token"
	OAuthRefreshFailurePersist        OAuthRefreshFailureKind = "persist_failed"
	OAuthRefreshFailureCanceled       OAuthRefreshFailureKind = "canceled"
)

type OAuthRequestPathRefreshSettings struct {
	RequestTimeout               time.Duration
	LockWaitTimeout              time.Duration
	TransientTempUnschedDuration time.Duration
}

func DefaultOAuthRequestPathRefreshSettings() OAuthRequestPathRefreshSettings {
	return OAuthRequestPathRefreshSettings{
		RequestTimeout:               defaultOAuthRequestPathRefreshTimeout,
		LockWaitTimeout:              defaultOAuthLockWaitTimeout,
		TransientTempUnschedDuration: tokenRefreshTempUnschedDuration,
	}
}

func OAuthRequestPathRefreshSettingsFromConfig(cfg *config.Config) OAuthRequestPathRefreshSettings {
	settings := DefaultOAuthRequestPathRefreshSettings()
	if cfg == nil {
		return settings
	}
	if cfg.TokenRefresh.RequestPathTimeoutSeconds > 0 {
		settings.RequestTimeout = time.Duration(cfg.TokenRefresh.RequestPathTimeoutSeconds) * time.Second
	}
	if cfg.TokenRefresh.LockWaitTimeoutMilliseconds > 0 {
		settings.LockWaitTimeout = time.Duration(cfg.TokenRefresh.LockWaitTimeoutMilliseconds) * time.Millisecond
	}
	if cfg.TokenRefresh.TransientTempUnschedMinutes > 0 {
		settings.TransientTempUnschedDuration = time.Duration(cfg.TokenRefresh.TransientTempUnschedMinutes) * time.Minute
	}
	return normalizeOAuthRequestPathRefreshSettings(settings)
}

func normalizeOAuthRequestPathRefreshSettings(settings OAuthRequestPathRefreshSettings) OAuthRequestPathRefreshSettings {
	defaults := DefaultOAuthRequestPathRefreshSettings()
	if settings.RequestTimeout <= 0 {
		settings.RequestTimeout = defaults.RequestTimeout
	}
	if settings.LockWaitTimeout <= 0 {
		settings.LockWaitTimeout = defaults.LockWaitTimeout
	}
	if settings.TransientTempUnschedDuration <= 0 {
		settings.TransientTempUnschedDuration = defaults.TransientTempUnschedDuration
	}
	return settings
}

func classifyOAuthRefreshFailure(err error) OAuthRefreshFailureKind {
	if err == nil {
		return OAuthRefreshFailureNone
	}
	if errors.Is(err, errOAuthRefreshLockHeld) {
		return OAuthRefreshFailureLockHeld
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return OAuthRefreshFailureCanceled
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no refresh token available") ||
		strings.Contains(msg, "refresh token not found") ||
		strings.Contains(msg, "missing refresh_token") ||
		strings.Contains(msg, "missing refresh token"):
		return OAuthRefreshFailureNoRefreshToken
	case strings.Contains(msg, "invalid_grant") ||
		strings.Contains(msg, "invalid_client") ||
		strings.Contains(msg, "unauthorized_client") ||
		strings.Contains(msg, "access_denied") ||
		strings.Contains(msg, "missing_project_id"):
		return OAuthRefreshFailurePermanent
	case strings.Contains(msg, "oauth refresh succeeded but db update failed") ||
		strings.Contains(msg, "failed to save credentials") ||
		strings.Contains(msg, "db update failed") ||
		strings.Contains(msg, "persistence failed"):
		return OAuthRefreshFailurePersist
	default:
		return OAuthRefreshFailureTransient
	}
}

func isNonRetryableRefreshError(err error) bool {
	switch classifyOAuthRefreshFailure(err) {
	case OAuthRefreshFailurePermanent, OAuthRefreshFailureNoRefreshToken:
		return true
	default:
		return false
	}
}

func waitForCachedOAuthToken(ctx context.Context, cache GeminiTokenCache, cacheKey string, timeout time.Duration) (string, error) {
	if cache == nil || strings.TrimSpace(cacheKey) == "" {
		return "", nil
	}
	if timeout <= 0 {
		timeout = defaultOAuthLockWaitTimeout
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		token, err := cache.GetAccessToken(waitCtx, cacheKey)
		if err == nil && strings.TrimSpace(token) != "" {
			return token, nil
		}
		select {
		case <-waitCtx.Done():
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return "", nil
			}
			return "", waitCtx.Err()
		case <-ticker.C:
		}
	}
}

func requestPathRefreshContext(ctx context.Context, settings OAuthRequestPathRefreshSettings) (context.Context, context.CancelFunc) {
	settings = normalizeOAuthRequestPathRefreshSettings(settings)
	if settings.RequestTimeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, settings.RequestTimeout)
}

func markRequestPathRefreshFailure(
	accountRepo AccountRepository,
	tempUnschedCache TempUnschedCache,
	account *Account,
	refreshErr error,
	settings OAuthRequestPathRefreshSettings,
	provider string,
) {
	if accountRepo == nil || account == nil || refreshErr == nil {
		return
	}

	kind := classifyOAuthRefreshFailure(refreshErr)
	switch kind {
	case OAuthRefreshFailurePermanent, OAuthRefreshFailureNoRefreshToken:
		message := fmt.Sprintf("Token refresh failed (%s): %v", kind, refreshErr)
		// Persist the refresh outcome even if the client request has already ended.
		writeCtx, cancel := newTempUnschedWriteContext(context.Background())
		defer cancel()
		if err := accountRepo.SetError(writeCtx, account.ID, message); err != nil {
			slog.Warn("oauth_refresh_request_path_set_error_failed",
				"account_id", account.ID,
				"platform", account.Platform,
				"provider", provider,
				"failure_kind", kind,
				"error", err,
			)
			return
		}
		slog.Warn("oauth_refresh_request_path_set_error",
			"account_id", account.ID,
			"platform", account.Platform,
			"provider", provider,
			"failure_kind", kind,
		)
	case OAuthRefreshFailureLockHeld:
		return
	default:
		settings = normalizeOAuthRequestPathRefreshSettings(settings)
		now := time.Now()
		until := now.Add(settings.TransientTempUnschedDuration)
		reason := "token refresh failed on request path: " + refreshErr.Error()
		state := &TempUnschedState{
			UntilUnix:       until.Unix(),
			TriggeredAtUnix: now.Unix(),
			ErrorMessage:    reason,
		}
		// Persist the refresh outcome even if the client request has already ended.
		writeCtx, cancel := newTempUnschedWriteContext(context.Background())
		defer cancel()
		if err := accountRepo.SetTempUnschedulable(writeCtx, account.ID, until, marshalTempUnschedState(state)); err != nil {
			slog.Warn("oauth_refresh_request_path_temp_unsched_failed",
				"account_id", account.ID,
				"platform", account.Platform,
				"provider", provider,
				"failure_kind", kind,
				"error", err,
			)
			return
		}
		if tempUnschedCache != nil {
			if err := tempUnschedCache.SetTempUnsched(writeCtx, account.ID, state); err != nil {
				slog.Warn("oauth_refresh_request_path_temp_unsched_cache_failed",
					"account_id", account.ID,
					"platform", account.Platform,
					"provider", provider,
					"failure_kind", kind,
					"error", err,
				)
			}
		}
		slog.Warn("oauth_refresh_request_path_temp_unsched",
			"account_id", account.ID,
			"platform", account.Platform,
			"provider", provider,
			"failure_kind", kind,
			"cooldown_until", until,
		)
	}
}

func newOAuthRequestPathFailoverError(account *Account, kind OAuthRefreshFailureKind, err error) error {
	if kind == OAuthRefreshFailureNone {
		kind = classifyOAuthRefreshFailure(err)
	}
	message := "oauth token refresh failed"
	if kind == OAuthRefreshFailureLockHeld {
		message = "oauth token refresh lock held"
	}
	body, _ := json.Marshal(map[string]any{
		"error": map[string]any{
			"message": message,
			"type":    "oauth_refresh_" + string(kind),
		},
	})
	failureReason := "oauth_token_refresh_" + string(kind)
	return &UpstreamFailoverError{
		StatusCode:    http.StatusServiceUnavailable,
		ResponseBody:  body,
		FailureReason: failureReason,
	}
}
