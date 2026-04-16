//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTrySelectStickyAccountWithoutRouting_ReturnsStickyHit(t *testing.T) {
	svc := &GatewayService{
		cfg: testConfig(),
	}

	accountByID := map[int64]*Account{
		1: {
			ID:          1,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 5,
		},
	}

	result, ok := svc.trySelectStickyAccountWithoutRouting(
		context.Background(),
		nil,
		"sticky",
		"claude-3-5-sonnet-20241022",
		1,
		nil,
		accountByID,
		PlatformAnthropic,
		false,
		time.Second,
		1,
	)

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, int64(1), result.Account.ID)
	require.True(t, result.Acquired)
}

func TestTrySelectStickyAccountWithoutRouting_ClearsUnschedulableBinding(t *testing.T) {
	cache := &mockGatewayCacheForPlatform{
		sessionBindings: map[string]int64{"sticky": 1},
	}
	svc := &GatewayService{
		cache: cache,
		cfg:   testConfig(),
	}

	accountByID := map[int64]*Account{
		1: {
			ID:          1,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Status:      StatusDisabled,
			Schedulable: true,
		},
	}

	result, ok := svc.trySelectStickyAccountWithoutRouting(
		context.Background(),
		nil,
		"sticky",
		"claude-3-5-sonnet-20241022",
		1,
		nil,
		accountByID,
		PlatformAnthropic,
		false,
		time.Second,
		1,
	)

	require.False(t, ok)
	require.Nil(t, result)
	require.Equal(t, 1, cache.deletedSessions["sticky"])
}

func TestTrySelectStickyAccountWithoutRouting_KeepsCodexSnapshotBinding(t *testing.T) {
	now := time.Now().UTC()
	usedPercent := 100.0
	resetAfter := 3600
	windowMinutes := 10080
	cache := &mockGatewayCacheForPlatform{
		sessionBindings: map[string]int64{"sticky": 1},
	}
	svc := &GatewayService{
		cache: cache,
		cfg:   testConfig(),
	}

	accountByID := map[int64]*Account{
		1: {
			ID:          1,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Credentials: map[string]any{
				"access_token": "token",
				"expires_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
			Extra: buildCodexUsageExtraUpdates(&OpenAICodexUsageSnapshot{
				PrimaryUsedPercent:       &usedPercent,
				PrimaryResetAfterSeconds: &resetAfter,
				PrimaryWindowMinutes:     &windowMinutes,
				UpdatedAt:                now.Format(time.RFC3339),
			}, now),
		},
	}

	result, ok := svc.trySelectStickyAccountWithoutRouting(
		context.Background(),
		nil,
		"sticky",
		"gpt-5.1",
		1,
		nil,
		accountByID,
		PlatformOpenAI,
		false,
		time.Second,
		1,
	)

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, int64(1), result.Account.ID)
	require.True(t, result.Acquired)
	require.Empty(t, cache.deletedSessions)
	require.Nil(t, accountByID[1].RateLimitResetAt)
}
