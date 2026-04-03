//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildAvailableAccountLoads_UsesDefaultLoadInfoWhenMissing(t *testing.T) {
	candidates := []*Account{
		{ID: 1},
		{ID: 2},
	}

	available := buildAvailableAccountLoads(candidates, map[int64]*AccountLoadInfo{
		1: {AccountID: 1, LoadRate: 100},
	})

	require.Len(t, available, 1)
	require.Equal(t, int64(2), available[0].account.ID)
	require.NotNil(t, available[0].loadInfo)
	require.Equal(t, int64(2), available[0].loadInfo.AccountID)
	require.Zero(t, available[0].loadInfo.LoadRate)
}

func TestRemoveAccountWithLoadByID_RemovesSelectedAccount(t *testing.T) {
	available := []accountWithLoad{
		{account: &Account{ID: 1}, loadInfo: &AccountLoadInfo{AccountID: 1, LoadRate: 10}},
		{account: &Account{ID: 2}, loadInfo: &AccountLoadInfo{AccountID: 2, LoadRate: 20}},
		{account: &Account{ID: 3}, loadInfo: &AccountLoadInfo{AccountID: 3, LoadRate: 30}},
	}

	filtered := removeAccountWithLoadByID(available, 2)

	require.Len(t, filtered, 2)
	require.Equal(t, int64(1), filtered[0].account.ID)
	require.Equal(t, int64(3), filtered[1].account.ID)
}

func TestSchedulingConfig_ReturnsDefaultsWhenConfigNil(t *testing.T) {
	svc := &GatewayService{}

	cfg := svc.schedulingConfig()

	require.Equal(t, 3, cfg.StickySessionMaxWaiting)
	require.Equal(t, 45*time.Second, cfg.StickySessionWaitTimeout)
	require.Equal(t, 30*time.Second, cfg.FallbackWaitTimeout)
	require.Equal(t, 100, cfg.FallbackMaxWaiting)
	require.True(t, cfg.LoadBatchEnabled)
	require.Equal(t, 30*time.Second, cfg.SlotCleanupInterval)
}

func TestTryAcquireAccountSlot_ReturnsImmediateAcquireWhenConcurrencyServiceMissing(t *testing.T) {
	svc := &GatewayService{}

	result, err := svc.tryAcquireAccountSlot(context.Background(), 42, 5)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Acquired)
	require.NotNil(t, result.ReleaseFunc)
	result.ReleaseFunc()
}
