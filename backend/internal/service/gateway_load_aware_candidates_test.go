//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFilterLoadAwareCandidates_AppliesEligibilityRules(t *testing.T) {
	svc := &GatewayService{
		sessionLimitCache: &sessionLimitRuntimeStub{
			windowHit:  true,
			windowCost: 20,
		},
		cfg: testConfig(),
	}

	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
		{
			ID:          2,
			Platform:    PlatformGemini,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
		{
			ID:          3,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"other-model": "other-model",
				},
			},
		},
		{
			ID:          4,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Extra: map[string]any{
				"window_cost_limit": 10.0,
			},
		},
		{
			ID:          5,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
		},
	}

	candidates := svc.filterLoadAwareCandidates(
		context.Background(),
		accounts,
		"claude-test",
		PlatformAnthropic,
		false,
		map[int64]struct{}{1: {}},
	)

	require.Len(t, candidates, 1)
	require.Equal(t, int64(5), candidates[0].ID)
}

func TestFilterLoadAwareCandidates_SkipsExpiredOAuthWithoutRefreshToken(t *testing.T) {
	svc := &GatewayService{cfg: testConfig()}
	expiredAt := time.Now().Add(-time.Minute).Format(time.RFC3339)

	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"access_token": "expired-token",
				"expires_at":   expiredAt,
			},
		},
		{
			ID:          2,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"access_token": "fresh-token",
				"expires_at":   time.Now().Add(time.Hour).Format(time.RFC3339),
			},
		},
	}

	candidates := svc.filterLoadAwareCandidates(
		context.Background(),
		accounts,
		"",
		PlatformAnthropic,
		false,
		nil,
	)

	require.Len(t, candidates, 1)
	require.Equal(t, int64(2), candidates[0].ID)
}

func TestFilterLoadAwareCandidates_SkipsOpenAICodexSnapshotRateLimitedAccounts(t *testing.T) {
	svc := &GatewayService{cfg: testConfig()}
	now := time.Now().UTC()
	usedPercent := 100.0
	resetAfter := 3600
	windowMinutes := 10080
	codexExtra := buildCodexUsageExtraUpdates(&OpenAICodexUsageSnapshot{
		PrimaryUsedPercent:       &usedPercent,
		PrimaryResetAfterSeconds: &resetAfter,
		PrimaryWindowMinutes:     &windowMinutes,
		UpdatedAt:                now.Format(time.RFC3339),
	}, now)

	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"access_token": "fresh-token",
				"expires_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
			Extra: codexExtra,
		},
		{
			ID:          2,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"access_token": "fresh-token-2",
				"expires_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
		},
	}

	candidates := svc.filterLoadAwareCandidates(
		context.Background(),
		accounts,
		"",
		PlatformOpenAI,
		false,
		nil,
	)

	require.Len(t, candidates, 1)
	require.Equal(t, int64(2), candidates[0].ID)
	require.NotNil(t, accounts[0].RateLimitResetAt)
}
