//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCloneExcludedAccountIDs_CreatesIndependentCopy(t *testing.T) {
	original := map[int64]struct{}{
		1: {},
		2: {},
	}

	cloned := cloneExcludedAccountIDs(original)
	delete(cloned, 1)
	cloned[3] = struct{}{}

	require.Len(t, original, 2)
	_, hasOriginalOne := original[1]
	require.True(t, hasOriginalOne)
	_, hasOriginalThree := original[3]
	require.False(t, hasOriginalThree)
}

func TestBuildNonLoadBatchWaitPlan_UsesStickyWaitConfigForStickyAccount(t *testing.T) {
	cache := &stubConcurrencyCacheForTest{waitCount: 1}
	svc := &GatewayService{
		concurrencyService: NewConcurrencyService(cache),
	}
	account := &Account{ID: 9, Concurrency: 4}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 5 * time.Second,
		StickySessionMaxWaiting:  3,
		FallbackWaitTimeout:      2 * time.Second,
		FallbackMaxWaiting:       10,
	}

	result := svc.buildNonLoadBatchWaitPlan(context.Background(), account, account.ID, cfg)

	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, account.ID, result.WaitPlan.AccountID)
	require.Equal(t, 5*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 3, result.WaitPlan.MaxWaiting)
}

func TestBuildNonLoadBatchWaitPlan_FallsBackWhenStickyQueueIsFull(t *testing.T) {
	cache := &stubConcurrencyCacheForTest{waitCount: 3}
	svc := &GatewayService{
		concurrencyService: NewConcurrencyService(cache),
	}
	account := &Account{ID: 9, Concurrency: 4}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 5 * time.Second,
		StickySessionMaxWaiting:  3,
		FallbackWaitTimeout:      2 * time.Second,
		FallbackMaxWaiting:       10,
	}

	result := svc.buildNonLoadBatchWaitPlan(context.Background(), account, account.ID, cfg)

	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, account.ID, result.WaitPlan.AccountID)
	require.Equal(t, 2*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 10, result.WaitPlan.MaxWaiting)
}
