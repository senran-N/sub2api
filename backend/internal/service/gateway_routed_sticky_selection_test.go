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
