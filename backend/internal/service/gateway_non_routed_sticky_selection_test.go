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
