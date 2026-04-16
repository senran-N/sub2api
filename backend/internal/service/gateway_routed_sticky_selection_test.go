//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTrySelectRoutedStickyAccount_ReturnsStickyHit(t *testing.T) {
	svc := &GatewayService{
		cache: &mockGatewayCacheForPlatform{
			sessionBindings: map[string]int64{"session-1": 1},
		},
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

	result, ok := svc.trySelectRoutedStickyAccount(
		context.Background(),
		nil,
		"session-1",
		"claude-3-5-sonnet-20241022",
		1,
		[]int64{1, 2},
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
	require.True(t, result.Acquired)
	require.Equal(t, int64(1), result.Account.ID)
}

func TestTrySelectRoutedStickyAccount_ClearsMissingBinding(t *testing.T) {
	cache := &mockGatewayCacheForPlatform{
		sessionBindings: map[string]int64{"session-1": 1},
	}
	svc := &GatewayService{
		cache: cache,
		cfg:   testConfig(),
	}

	result, ok := svc.trySelectRoutedStickyAccount(
		context.Background(),
		nil,
		"session-1",
		"claude-3-5-sonnet-20241022",
		1,
		[]int64{1, 2},
		nil,
		map[int64]*Account{},
		PlatformAnthropic,
		false,
		time.Second,
		1,
	)

	require.False(t, ok)
	require.Nil(t, result)
	require.Equal(t, 1, cache.deletedSessions["session-1"])
}

func TestTrySelectRoutedStickyAccount_InvalidBindingPolicy(t *testing.T) {
	t.Run("rate limited account keeps binding", func(t *testing.T) {
		rateLimitedUntil := time.Now().Add(5 * time.Minute)
		cache := &mockGatewayCacheForPlatform{
			sessionBindings: map[string]int64{"session-1": 1},
		}
		svc := &GatewayService{
			cache: cache,
			cfg:   testConfig(),
		}

		result, ok := svc.trySelectRoutedStickyAccount(
			context.Background(),
			nil,
			"session-1",
			"claude-3-5-sonnet-20241022",
			1,
			[]int64{1},
			nil,
			map[int64]*Account{
				1: {
					ID:               1,
					Platform:         PlatformAnthropic,
					Type:             AccountTypeOAuth,
					Status:           StatusActive,
					Schedulable:      true,
					Concurrency:      5,
					RateLimitResetAt: &rateLimitedUntil,
				},
			},
			PlatformAnthropic,
			false,
			time.Second,
			1,
		)

		require.False(t, ok)
		require.Nil(t, result)
		require.Zero(t, cache.deletedSessions["session-1"])
		require.Equal(t, int64(1), cache.sessionBindings["session-1"])
	})

	t.Run("expired oauth without refresh token clears binding", func(t *testing.T) {
		cache := &mockGatewayCacheForPlatform{
			sessionBindings: map[string]int64{"session-1": 1},
		}
		svc := &GatewayService{
			cache: cache,
			cfg:   testConfig(),
		}

		result, ok := svc.trySelectRoutedStickyAccount(
			context.Background(),
			nil,
			"session-1",
			"claude-3-5-sonnet-20241022",
			1,
			[]int64{1},
			nil,
			map[int64]*Account{
				1: {
					ID:          1,
					Platform:    PlatformAnthropic,
					Type:        AccountTypeOAuth,
					Status:      StatusActive,
					Schedulable: true,
					Concurrency: 5,
					Credentials: map[string]any{
						"access_token": "expired-token",
						"expires_at":   time.Now().Add(-time.Minute).Format(time.RFC3339),
					},
				},
			},
			PlatformAnthropic,
			false,
			time.Second,
			1,
		)

		require.False(t, ok)
		require.Nil(t, result)
		require.Equal(t, 1, cache.deletedSessions["session-1"])
		_, exists := cache.sessionBindings["session-1"]
		require.False(t, exists)
	})
}
