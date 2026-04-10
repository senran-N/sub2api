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

func TestTryAcquireOrBuildStickyWaitPlan_Acquired(t *testing.T) {
	account := &Account{ID: 90, Concurrency: 3}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 13 * time.Second,
		StickySessionMaxWaiting:  5,
	}
	concurrencyService := NewConcurrencyService(stubConcurrencyCache{
		waitCounts: map[int64]int{90: 999},
	})
	releaseCalled := false
	onAcquireCalled := false

	result, ok := tryAcquireOrBuildStickyWaitPlan(
		context.Background(),
		account,
		account.ID,
		cfg,
		concurrencyService,
		func(context.Context, int64, int) (*AcquireResult, error) {
			return &AcquireResult{
				Acquired: true,
				ReleaseFunc: func() {
					releaseCalled = true
				},
			}, nil
		},
		func(account *Account, acquired *AcquireResult) *AccountSelectionResult {
			onAcquireCalled = true
			return newAcquiredAccountSelection(account, acquired.ReleaseFunc)
		},
	)

	require.True(t, ok)
	require.True(t, onAcquireCalled)
	require.NotNil(t, result)
	require.True(t, result.Acquired)
	require.NotNil(t, result.ReleaseFunc)
	result.ReleaseFunc()
	require.True(t, releaseCalled)
}

func TestTryAcquireOrBuildStickyWaitPlan_AcquireFailedReturnsStickyWaitPlan(t *testing.T) {
	account := &Account{ID: 91, Concurrency: 2}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 17 * time.Second,
		StickySessionMaxWaiting:  4,
	}
	concurrencyService := NewConcurrencyService(stubConcurrencyCache{
		waitCounts: map[int64]int{91: 1000},
	})
	onAcquireCalled := false

	result, ok := tryAcquireOrBuildStickyWaitPlan(
		context.Background(),
		account,
		account.ID,
		cfg,
		concurrencyService,
		func(context.Context, int64, int) (*AcquireResult, error) {
			return &AcquireResult{Acquired: false}, nil
		},
		func(*Account, *AcquireResult) *AccountSelectionResult {
			onAcquireCalled = true
			return nil
		},
	)

	require.True(t, ok)
	require.False(t, onAcquireCalled)
	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, int64(91), result.WaitPlan.AccountID)
	require.Equal(t, 17*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 4, result.WaitPlan.MaxWaiting)
}

func TestTryAcquireOrBuildStickyWaitPlan_AcquireFailedWithoutConcurrencyService(t *testing.T) {
	account := &Account{ID: 92, Concurrency: 1}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 19 * time.Second,
		StickySessionMaxWaiting:  2,
	}

	result, ok := tryAcquireOrBuildStickyWaitPlan(
		context.Background(),
		account,
		account.ID,
		cfg,
		nil,
		func(context.Context, int64, int) (*AcquireResult, error) {
			return &AcquireResult{Acquired: false}, nil
		},
		func(*Account, *AcquireResult) *AccountSelectionResult {
			return nil
		},
	)

	require.False(t, ok)
	require.Nil(t, result)
}
