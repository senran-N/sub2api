package service

import "sync/atomic"

type SchedulingRuntimeKernelMetricsSnapshot struct {
	IndexPageFetches         int64 `json:"index_page_fetches"`
	IndexFetchedAccounts     int64 `json:"index_fetched_accounts"`
	IndexReturnedBatches     int64 `json:"index_returned_batches"`
	IndexReturnedAccounts    int64 `json:"index_returned_accounts"`
	OrderedRuntimeProbes     int64 `json:"ordered_runtime_probes"`
	OrderedWaitPlanProbes    int64 `json:"ordered_wait_plan_probes"`
	RuntimeAcquireAttempts   int64 `json:"runtime_acquire_attempts"`
	RuntimeAcquireSuccess    int64 `json:"runtime_acquire_success"`
	RuntimeAcquireMisses     int64 `json:"runtime_acquire_misses"`
	RuntimeAcquireErrors     int64 `json:"runtime_acquire_errors"`
	RuntimeFinalizeMisses    int64 `json:"runtime_finalize_misses"`
	RuntimeSessionMisses     int64 `json:"runtime_session_misses"`
	RuntimeWaitPlanAttempts  int64 `json:"runtime_wait_plan_attempts"`
	RuntimeWaitPlanSuccess   int64 `json:"runtime_wait_plan_success"`
	RuntimeWaitPlanRejected  int64 `json:"runtime_wait_plan_rejected"`
	RuntimeWaitPlanFinalMiss int64 `json:"runtime_wait_plan_finalize_miss"`
}

type schedulingRuntimeKernelStats struct {
	indexPageFetches         atomic.Int64
	indexFetchedAccounts     atomic.Int64
	indexReturnedBatches     atomic.Int64
	indexReturnedAccounts    atomic.Int64
	orderedRuntimeProbes     atomic.Int64
	orderedWaitPlanProbes    atomic.Int64
	runtimeAcquireAttempts   atomic.Int64
	runtimeAcquireSuccess    atomic.Int64
	runtimeAcquireMisses     atomic.Int64
	runtimeAcquireErrors     atomic.Int64
	runtimeFinalizeMisses    atomic.Int64
	runtimeSessionMisses     atomic.Int64
	runtimeWaitPlanAttempts  atomic.Int64
	runtimeWaitPlanSuccess   atomic.Int64
	runtimeWaitPlanRejected  atomic.Int64
	runtimeWaitPlanFinalMiss atomic.Int64
}

var defaultSchedulingRuntimeKernelStats schedulingRuntimeKernelStats

func resetSchedulingRuntimeKernelStats() {
	defaultSchedulingRuntimeKernelStats = schedulingRuntimeKernelStats{}
}

func SnapshotSchedulingRuntimeKernelMetrics() SchedulingRuntimeKernelMetricsSnapshot {
	return SchedulingRuntimeKernelMetricsSnapshot{
		IndexPageFetches:         defaultSchedulingRuntimeKernelStats.indexPageFetches.Load(),
		IndexFetchedAccounts:     defaultSchedulingRuntimeKernelStats.indexFetchedAccounts.Load(),
		IndexReturnedBatches:     defaultSchedulingRuntimeKernelStats.indexReturnedBatches.Load(),
		IndexReturnedAccounts:    defaultSchedulingRuntimeKernelStats.indexReturnedAccounts.Load(),
		OrderedRuntimeProbes:     defaultSchedulingRuntimeKernelStats.orderedRuntimeProbes.Load(),
		OrderedWaitPlanProbes:    defaultSchedulingRuntimeKernelStats.orderedWaitPlanProbes.Load(),
		RuntimeAcquireAttempts:   defaultSchedulingRuntimeKernelStats.runtimeAcquireAttempts.Load(),
		RuntimeAcquireSuccess:    defaultSchedulingRuntimeKernelStats.runtimeAcquireSuccess.Load(),
		RuntimeAcquireMisses:     defaultSchedulingRuntimeKernelStats.runtimeAcquireMisses.Load(),
		RuntimeAcquireErrors:     defaultSchedulingRuntimeKernelStats.runtimeAcquireErrors.Load(),
		RuntimeFinalizeMisses:    defaultSchedulingRuntimeKernelStats.runtimeFinalizeMisses.Load(),
		RuntimeSessionMisses:     defaultSchedulingRuntimeKernelStats.runtimeSessionMisses.Load(),
		RuntimeWaitPlanAttempts:  defaultSchedulingRuntimeKernelStats.runtimeWaitPlanAttempts.Load(),
		RuntimeWaitPlanSuccess:   defaultSchedulingRuntimeKernelStats.runtimeWaitPlanSuccess.Load(),
		RuntimeWaitPlanRejected:  defaultSchedulingRuntimeKernelStats.runtimeWaitPlanRejected.Load(),
		RuntimeWaitPlanFinalMiss: defaultSchedulingRuntimeKernelStats.runtimeWaitPlanFinalMiss.Load(),
	}
}
