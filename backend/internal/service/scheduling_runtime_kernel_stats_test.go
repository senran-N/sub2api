package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnapshotSchedulingRuntimeKernelMetrics(t *testing.T) {
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	require.Equal(t, SchedulingRuntimeKernelMetricsSnapshot{}, before)

	defaultSchedulingRuntimeKernelStats.indexPageFetches.Add(3)
	defaultSchedulingRuntimeKernelStats.indexReturnedAccounts.Add(9)
	defaultSchedulingRuntimeKernelStats.runtimeAcquireAttempts.Add(2)
	defaultSchedulingRuntimeKernelStats.runtimeAcquireSuccess.Add(1)
	defaultSchedulingRuntimeKernelStats.runtimeWaitPlanAttempts.Add(4)

	after := SnapshotSchedulingRuntimeKernelMetrics()
	require.Equal(t, int64(3), after.IndexPageFetches)
	require.Equal(t, int64(9), after.IndexReturnedAccounts)
	require.Equal(t, int64(2), after.RuntimeAcquireAttempts)
	require.Equal(t, int64(1), after.RuntimeAcquireSuccess)
	require.Equal(t, int64(4), after.RuntimeWaitPlanAttempts)
}
