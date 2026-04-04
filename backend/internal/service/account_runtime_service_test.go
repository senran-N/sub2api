package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

func TestAccountUsageService_GetAccountWindowStatsBatch_GroupsByWindowStart(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)
	windowEnd := now.Add(time.Hour)
	otherStart := now.Add(-time.Hour)
	otherEnd := now.Add(2 * time.Hour)

	repo := &usageLogWindowBatchRepoStub{
		batchResult: map[int64]*usagestats.AccountStats{
			1: {StandardCost: 1.25},
			2: {StandardCost: 2.5},
			3: {StandardCost: 3.75},
		},
	}
	svc := &AccountUsageService{usageLogRepo: repo}

	accounts := []Account{
		{
			ID:                 1,
			Platform:           PlatformAnthropic,
			Type:               AccountTypeOAuth,
			Extra:              map[string]any{"window_cost_limit": 10},
			SessionWindowStart: &now,
			SessionWindowEnd:   &windowEnd,
		},
		{
			ID:                 2,
			Platform:           PlatformAnthropic,
			Type:               AccountTypeSetupToken,
			Extra:              map[string]any{"window_cost_limit": 10},
			SessionWindowStart: &now,
			SessionWindowEnd:   &windowEnd,
		},
		{
			ID:                 3,
			Platform:           PlatformAnthropic,
			Type:               AccountTypeOAuth,
			Extra:              map[string]any{"window_cost_limit": 10},
			SessionWindowStart: &otherStart,
			SessionWindowEnd:   &otherEnd,
		},
	}

	statsByAccount, err := svc.GetAccountWindowStatsBatch(context.Background(), accounts)
	require.NoError(t, err)
	require.EqualValues(t, 2, repo.batchCalls.Load(), "accounts should be grouped by active window start")
	require.Zero(t, repo.singleCalls.Load(), "batch provider should avoid per-account fallback on success")
	require.InDelta(t, 1.25, statsByAccount[1].StandardCost, 1e-9)
	require.InDelta(t, 2.5, statsByAccount[2].StandardCost, 1e-9)
	require.InDelta(t, 3.75, statsByAccount[3].StandardCost, 1e-9)
}

func TestAccountRuntimeService_CollectAccountMetrics_UsesCachedWindowCostAndBackfillsMisses(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)
	windowEnd := now.Add(time.Hour)

	repo := &usageLogWindowBatchRepoStub{
		batchResult: map[int64]*usagestats.AccountStats{
			2: {StandardCost: 4.5},
		},
	}
	usageSvc := &AccountUsageService{usageLogRepo: repo}
	sessionCache := &sessionLimitCacheHotpathStub{
		batchData: map[int64]float64{
			1: 1.5,
		},
	}
	runtimeSvc := NewAccountRuntimeService(usageSvc, nil, sessionCache, nil)

	accounts := []Account{
		{
			ID:                 1,
			Platform:           PlatformAnthropic,
			Type:               AccountTypeOAuth,
			Extra:              map[string]any{"window_cost_limit": 10},
			SessionWindowStart: &now,
			SessionWindowEnd:   &windowEnd,
		},
		{
			ID:                 2,
			Platform:           PlatformAnthropic,
			Type:               AccountTypeOAuth,
			Extra:              map[string]any{"window_cost_limit": 10},
			SessionWindowStart: &now,
			SessionWindowEnd:   &windowEnd,
		},
	}

	metrics := runtimeSvc.CollectAccountMetrics(context.Background(), accounts)
	require.InDelta(t, 1.5, metrics.WindowCosts[1], 1e-9)
	require.InDelta(t, 4.5, metrics.WindowCosts[2], 1e-9)
	require.EqualValues(t, 1, repo.batchCalls.Load(), "only uncached accounts should hit the batch stats path")
	require.Equal(t, map[int64]float64{2: 4.5}, sessionCache.setData, "freshly fetched costs should be written back to cache")
}
