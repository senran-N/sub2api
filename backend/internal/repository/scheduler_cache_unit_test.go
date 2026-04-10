package repository

import (
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildSchedulerMetadataAccount_StripsNonSchedulingFields(t *testing.T) {
	resetAt := time.Now().Add(10 * time.Minute).UTC()
	rate := 1.5
	loadFactor := 7
	account := service.Account{
		ID:               101,
		Name:             "sched-meta",
		Platform:         service.PlatformGemini,
		Type:             service.AccountTypeOAuth,
		Status:           service.StatusActive,
		Schedulable:      true,
		Concurrency:      3,
		Priority:         2,
		RateMultiplier:   &rate,
		LoadFactor:       &loadFactor,
		RateLimitResetAt: &resetAt,
		Credentials: map[string]any{
			"api_key":       "kept",
			"project_id":    "project-kept",
			"oauth_type":    "ai_studio",
			"model_mapping": map[string]any{"gemini-*": "gemini-2.5-pro"},
			"access_token":  "strip-me",
		},
		Extra: map[string]any{
			"mixed_scheduling":             true,
			"window_cost_limit":            12.5,
			"max_sessions":                 8,
			"session_idle_timeout_minutes": 15,
			"model_rate_limits": map[string]any{
				"gemini-2.5-pro": map[string]any{"reset_at": resetAt.Format(time.RFC3339)},
			},
			"quota_daily_limit": 1000,
			"quota_daily_used":  10,
			"unrelated_secret":  "strip-me",
		},
		ErrorMessage: "strip-me",
	}

	meta := buildSchedulerMetadataAccount(account)

	require.Equal(t, account.ID, meta.ID)
	require.Equal(t, account.Platform, meta.Platform)
	require.Equal(t, account.Type, meta.Type)
	require.Equal(t, account.Priority, meta.Priority)
	require.Equal(t, account.Concurrency, meta.Concurrency)
	require.Equal(t, "kept", meta.Credentials["api_key"])
	require.Equal(t, "project-kept", meta.Credentials["project_id"])
	require.Equal(t, "ai_studio", meta.Credentials["oauth_type"])
	require.Contains(t, meta.Credentials, "model_mapping")
	require.NotContains(t, meta.Credentials, "access_token")
	require.Equal(t, true, meta.Extra["mixed_scheduling"])
	require.Equal(t, 12.5, meta.Extra["window_cost_limit"])
	require.Equal(t, 8, meta.Extra["max_sessions"])
	require.Contains(t, meta.Extra, "model_rate_limits")
	require.Contains(t, meta.Extra, "quota_daily_limit")
	require.NotContains(t, meta.Extra, "unrelated_secret")
	require.Empty(t, meta.ErrorMessage)
}
