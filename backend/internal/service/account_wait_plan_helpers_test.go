package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestTryBuildStickySessionWaitPlan_UsesStickyConfigWhenQueueHasCapacity(t *testing.T) {
	account := &Account{ID: 42, Concurrency: 8}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 7 * time.Second,
		StickySessionMaxWaiting:  3,
	}
	concurrencyService := NewConcurrencyService(stubConcurrencyCache{
		waitCounts: map[int64]int{42: 2},
	})

	result, ok := tryBuildStickySessionWaitPlan(context.Background(), account, cfg, concurrencyService)

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, int64(42), result.WaitPlan.AccountID)
	require.Equal(t, 7*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 3, result.WaitPlan.MaxWaiting)
}

func TestTryBuildStickySessionWaitPlan_RejectsFullQueue(t *testing.T) {
	account := &Account{ID: 42, Concurrency: 8}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 7 * time.Second,
		StickySessionMaxWaiting:  3,
	}
	concurrencyService := NewConcurrencyService(stubConcurrencyCache{
		waitCounts: map[int64]int{42: 3},
	})

	result, ok := tryBuildStickySessionWaitPlan(context.Background(), account, cfg, concurrencyService)

	require.False(t, ok)
	require.Nil(t, result)
}

func TestBuildStickyAwareFallbackWaitPlan_FallsBackForNonStickyAccount(t *testing.T) {
	account := &Account{ID: 7, Concurrency: 4}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 9 * time.Second,
		StickySessionMaxWaiting:  2,
		FallbackWaitTimeout:      3 * time.Second,
		FallbackMaxWaiting:       10,
	}
	concurrencyService := NewConcurrencyService(stubConcurrencyCache{
		waitCounts: map[int64]int{7: 0},
	})

	result := buildStickyAwareFallbackWaitPlan(context.Background(), account, 42, cfg, concurrencyService)

	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, 3*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 10, result.WaitPlan.MaxWaiting)
}

func TestBuildStickySessionWaitPlanIfConcurrencyEnabled_AlwaysBuildsWhenServiceExists(t *testing.T) {
	account := &Account{ID: 77, Concurrency: 2}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 11 * time.Second,
		StickySessionMaxWaiting:  1,
	}
	concurrencyService := NewConcurrencyService(stubConcurrencyCache{
		waitCounts: map[int64]int{77: 999},
	})

	result, ok := buildStickySessionWaitPlanIfConcurrencyEnabled(account, cfg, concurrencyService)

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, int64(77), result.WaitPlan.AccountID)
	require.Equal(t, 11*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 1, result.WaitPlan.MaxWaiting)
}
