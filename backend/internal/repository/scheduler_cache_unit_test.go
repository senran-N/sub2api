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
			"mixed_scheduling":                           true,
			"privacy_mode":                               "training_off",
			"window_cost_limit":                          12.5,
			"max_sessions":                               8,
			"session_idle_timeout_minutes":               15,
			"responses_websockets_v2_enabled":            true,
			"openai_apikey_responses_websockets_v2_mode": service.OpenAIWSIngressModeCtxPool,
			"openai_ws_force_http":                       true,
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
	require.Equal(t, "training_off", meta.Extra["privacy_mode"])
	require.Equal(t, 12.5, meta.Extra["window_cost_limit"])
	require.Equal(t, 8, meta.Extra["max_sessions"])
	require.Contains(t, meta.Extra, "model_rate_limits")
	require.Contains(t, meta.Extra, "quota_daily_limit")
	require.Equal(t, true, meta.Extra["responses_websockets_v2_enabled"])
	require.Equal(t, service.OpenAIWSIngressModeCtxPool, meta.Extra["openai_apikey_responses_websockets_v2_mode"])
	require.Equal(t, true, meta.Extra["openai_ws_force_http"])
	require.NotContains(t, meta.Extra, "unrelated_secret")
	require.Empty(t, meta.ErrorMessage)
}

func TestBuildSchedulerCapabilityIndices_BuildsModelPrivacyAndWSIndexes(t *testing.T) {
	bucket := service.SchedulerBucket{
		GroupID:  7,
		Platform: service.PlatformOpenAI,
		Mode:     service.SchedulerModeSingle,
	}
	accounts := []service.Account{
		{
			ID:          201,
			Platform:    service.PlatformOpenAI,
			Type:        service.AccountTypeOAuth,
			Status:      service.StatusActive,
			Schedulable: true,
			Concurrency: 2,
			Priority:    0,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.1": "gpt-5.1",
					"gpt-*":   "gpt-5.1",
				},
			},
			Extra: map[string]any{
				"privacy_mode":                    service.PrivacyModeTrainingOff,
				"responses_websockets_v2_enabled": true,
				"supported_models":                []any{"gpt-5.1", "gpt-*"},
			},
		},
		{
			ID:          202,
			Platform:    service.PlatformGemini,
			Type:        service.AccountTypeOAuth,
			Status:      service.StatusActive,
			Schedulable: true,
			Concurrency: 1,
			Priority:    1,
			Credentials: map[string]any{},
		},
	}

	build := buildSchedulerCapabilityIndices(bucket, "9", accounts)

	require.Contains(t, build.zsets, schedulerCapabilityIndexKey(bucket, "9", service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexPrivacySet}))
	require.Contains(t, build.zsets, schedulerCapabilityIndexKey(bucket, "9", service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexOpenAIWS}))
	require.Contains(t, build.zsets, schedulerCapabilityIndexKey(bucket, "9", service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelAny}))
	require.Contains(t, build.zsets, schedulerCapabilityIndexKey(bucket, "9", service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelExact, Value: "gpt-5.1"}))
	require.Contains(t, build.zsets, schedulerCapabilityIndexKey(bucket, "9", service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelPattern, Value: "gpt-*"}))
	require.Contains(t, build.values[schedulerIndexValuesKey(bucket, "9", service.SchedulerCapabilityIndexModelPattern)], "gpt-*")
}

func TestBuildSchedulerCapabilityIndices_OpenAIEmptySupportedModelsUsesModelAny(t *testing.T) {
	bucket := service.SchedulerBucket{
		GroupID:  7,
		Platform: service.PlatformOpenAI,
		Mode:     service.SchedulerModeSingle,
	}
	accounts := []service.Account{
		{
			ID:          203,
			Platform:    service.PlatformOpenAI,
			Type:        service.AccountTypeOAuth,
			Status:      service.StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.4": "gpt-5.4",
				},
			},
			Extra: map[string]any{
				"supported_models": []any{},
			},
		},
	}

	build := buildSchedulerCapabilityIndices(bucket, "10", accounts)

	require.Contains(t, build.zsets, schedulerCapabilityIndexKey(bucket, "10", service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelAny}))
	require.NotContains(t, build.zsets, schedulerCapabilityIndexKey(bucket, "10", service.SchedulerCapabilityIndex{Kind: service.SchedulerCapabilityIndexModelExact, Value: "gpt-5.4"}))
}
