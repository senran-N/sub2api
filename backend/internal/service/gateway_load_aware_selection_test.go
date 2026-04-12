//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
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

	cfg := gatewaySchedulingConfigOrDefault(svc.cfg)

	require.Equal(t, 3, cfg.StickySessionMaxWaiting)
	require.Equal(t, 45*time.Second, cfg.StickySessionWaitTimeout)
	require.Equal(t, 30*time.Second, cfg.FallbackWaitTimeout)
	require.Equal(t, 100, cfg.FallbackMaxWaiting)
	require.True(t, cfg.LoadBatchEnabled)
	require.Equal(t, 30*time.Second, cfg.SlotCleanupInterval)
}

func TestSchedulingConfig_FillsPartialValuesFromDefaults(t *testing.T) {
	input := &config.Config{RunMode: config.RunModeStandard}
	input.Gateway.Scheduling.LoadBatchEnabled = true

	cfg := gatewaySchedulingConfigOrDefault(input)

	require.Equal(t, 3, cfg.StickySessionMaxWaiting)
	require.Equal(t, 45*time.Second, cfg.StickySessionWaitTimeout)
	require.Equal(t, 30*time.Second, cfg.FallbackWaitTimeout)
	require.Equal(t, 100, cfg.FallbackMaxWaiting)
	require.True(t, cfg.LoadBatchEnabled)
	require.Equal(t, 30*time.Second, cfg.SlotCleanupInterval)
}

func TestSchedulingConfig_PreservesExplicitOverrides(t *testing.T) {
	input := &config.Config{RunMode: config.RunModeStandard}
	input.Gateway.Scheduling.StickySessionMaxWaiting = 9
	input.Gateway.Scheduling.StickySessionWaitTimeout = 12 * time.Second
	input.Gateway.Scheduling.FallbackWaitTimeout = 8 * time.Second
	input.Gateway.Scheduling.FallbackMaxWaiting = 77
	input.Gateway.Scheduling.LoadBatchEnabled = false
	input.Gateway.Scheduling.SlotCleanupInterval = 2 * time.Minute

	cfg := gatewaySchedulingConfigOrDefault(input)

	require.Equal(t, 9, cfg.StickySessionMaxWaiting)
	require.Equal(t, 12*time.Second, cfg.StickySessionWaitTimeout)
	require.Equal(t, 8*time.Second, cfg.FallbackWaitTimeout)
	require.Equal(t, 77, cfg.FallbackMaxWaiting)
	require.False(t, cfg.LoadBatchEnabled)
	require.Equal(t, 2*time.Minute, cfg.SlotCleanupInterval)
}

func TestResolveSelectionAccountFromCurrentMap_DoesNotFetchMissingAccountOutsideIndexedMode(t *testing.T) {
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			2: {ID: 2, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true},
		},
	}
	svc := &GatewayService{accountRepo: repo}

	account, ok := svc.resolveSelectionAccountFromCurrentMap(context.Background(), map[int64]*Account{
		1: {ID: 1, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true},
	}, 2)

	require.False(t, ok)
	require.Nil(t, account)
	require.Equal(t, 0, repo.getByIDCalls)
}

func TestAcquireAccountSlotWithConcurrencyService_ReturnsImmediateAcquireWhenServiceMissing(t *testing.T) {
	result, err := acquireAccountSlotWithConcurrencyService(context.Background(), nil, 42, 5)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Acquired)
	require.NotNil(t, result.ReleaseFunc)
	result.ReleaseFunc()
}
