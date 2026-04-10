package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildRuntimeObservabilitySummary(t *testing.T) {
	summary := buildRuntimeObservabilitySummary(
		SchedulingRuntimeKernelMetricsSnapshot{
			IndexPageFetches:        4,
			IndexFetchedAccounts:    200,
			IndexReturnedBatches:    5,
			IndexReturnedAccounts:   50,
			OrderedRuntimeProbes:    9,
			OrderedWaitPlanProbes:   3,
			RuntimeAcquireAttempts:  10,
			RuntimeAcquireSuccess:   7,
			RuntimeFinalizeMisses:   2,
			RuntimeSessionMisses:    1,
			RuntimeWaitPlanAttempts: 8,
			RuntimeWaitPlanSuccess:  6,
		},
		IdempotencyMetricsSnapshot{
			ClaimTotal:                20,
			ReplayTotal:               5,
			ConflictTotal:             3,
			ProcessingDurationCount:   4,
			ProcessingDurationTotalMs: 120,
		},
		OpenAICompatibilityFallbackMetricsSnapshot{
			SessionHashLegacyReadHitRate: 0.75,
			MetadataLegacyFallbackTotal:  11,
		},
	)

	require.InDelta(t, 50.0, summary.SchedulingRuntimeKernel.AvgFetchedAccountsPerPage, 0.0001)
	require.InDelta(t, 10.0, summary.SchedulingRuntimeKernel.AvgReturnedAccountsPerBatch, 0.0001)
	require.InDelta(t, 0.7, summary.SchedulingRuntimeKernel.AcquireSuccessRate, 0.0001)
	require.InDelta(t, 0.75, summary.SchedulingRuntimeKernel.WaitPlanSuccessRate, 0.0001)
	require.InDelta(t, 0.2, summary.SchedulingRuntimeKernel.FinalizeMissRate, 0.0001)
	require.InDelta(t, 0.1, summary.SchedulingRuntimeKernel.SessionMissRate, 0.0001)
	require.Equal(t, int64(12), summary.SchedulingRuntimeKernel.TotalRuntimeProbes)

	require.InDelta(t, 5.0/28.0, summary.Idempotency.ReplayShare, 0.0001)
	require.InDelta(t, 3.0/28.0, summary.Idempotency.ConflictShare, 0.0001)
	require.InDelta(t, 30.0, summary.Idempotency.AvgProcessingDurationMs, 0.0001)

	require.InDelta(t, 0.75, summary.OpenAICompatibilityFallback.SessionHashLegacyReadHitRate, 0.0001)
	require.Equal(t, int64(11), summary.OpenAICompatibilityFallback.MetadataLegacyFallbackTotal)
}
