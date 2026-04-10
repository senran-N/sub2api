package service

type SchedulingRuntimeKernelSummary struct {
	AvgFetchedAccountsPerPage   float64 `json:"avg_fetched_accounts_per_page"`
	AvgReturnedAccountsPerBatch float64 `json:"avg_returned_accounts_per_batch"`
	AcquireSuccessRate          float64 `json:"acquire_success_rate"`
	WaitPlanSuccessRate         float64 `json:"wait_plan_success_rate"`
	FinalizeMissRate            float64 `json:"finalize_miss_rate"`
	SessionMissRate             float64 `json:"session_miss_rate"`
	TotalRuntimeProbes          int64   `json:"total_runtime_probes"`
}

type RuntimeIdempotencySummary struct {
	ReplayShare             float64 `json:"replay_share"`
	ConflictShare           float64 `json:"conflict_share"`
	AvgProcessingDurationMs float64 `json:"avg_processing_duration_ms"`
}

type RuntimeCompatibilityFallbackSummary struct {
	SessionHashLegacyReadHitRate float64 `json:"session_hash_legacy_read_hit_rate"`
	MetadataLegacyFallbackTotal  int64   `json:"metadata_legacy_fallback_total"`
}

type RuntimeObservabilitySummary struct {
	SchedulingRuntimeKernel     SchedulingRuntimeKernelSummary      `json:"scheduling_runtime_kernel"`
	Idempotency                 RuntimeIdempotencySummary           `json:"idempotency"`
	OpenAICompatibilityFallback RuntimeCompatibilityFallbackSummary `json:"openai_compatibility_fallback"`
}

type RuntimeObservabilitySnapshot struct {
	SchedulingRuntimeKernel     SchedulingRuntimeKernelMetricsSnapshot     `json:"scheduling_runtime_kernel"`
	Idempotency                 IdempotencyMetricsSnapshot                 `json:"idempotency"`
	OpenAICompatibilityFallback OpenAICompatibilityFallbackMetricsSnapshot `json:"openai_compatibility_fallback"`
	Summary                     RuntimeObservabilitySummary                `json:"summary"`
}

func buildRuntimeObservabilitySummary(
	scheduling SchedulingRuntimeKernelMetricsSnapshot,
	idempotency IdempotencyMetricsSnapshot,
	compatibility OpenAICompatibilityFallbackMetricsSnapshot,
) RuntimeObservabilitySummary {
	idempotencyEvents := float64(idempotency.ClaimTotal + idempotency.ReplayTotal + idempotency.ConflictTotal)

	return RuntimeObservabilitySummary{
		SchedulingRuntimeKernel: SchedulingRuntimeKernelSummary{
			AvgFetchedAccountsPerPage:   ratioOfInt64(scheduling.IndexFetchedAccounts, scheduling.IndexPageFetches),
			AvgReturnedAccountsPerBatch: ratioOfInt64(scheduling.IndexReturnedAccounts, scheduling.IndexReturnedBatches),
			AcquireSuccessRate:          ratioOfInt64(scheduling.RuntimeAcquireSuccess, scheduling.RuntimeAcquireAttempts),
			WaitPlanSuccessRate:         ratioOfInt64(scheduling.RuntimeWaitPlanSuccess, scheduling.RuntimeWaitPlanAttempts),
			FinalizeMissRate:            ratioOfInt64(scheduling.RuntimeFinalizeMisses, scheduling.RuntimeAcquireAttempts),
			SessionMissRate:             ratioOfInt64(scheduling.RuntimeSessionMisses, scheduling.RuntimeAcquireAttempts),
			TotalRuntimeProbes:          scheduling.OrderedRuntimeProbes + scheduling.OrderedWaitPlanProbes,
		},
		Idempotency: RuntimeIdempotencySummary{
			ReplayShare:             ratioOfUint64(idempotency.ReplayTotal, idempotencyEvents),
			ConflictShare:           ratioOfUint64(idempotency.ConflictTotal, idempotencyEvents),
			AvgProcessingDurationMs: ratioOfFloat64(idempotency.ProcessingDurationTotalMs, float64(idempotency.ProcessingDurationCount)),
		},
		OpenAICompatibilityFallback: RuntimeCompatibilityFallbackSummary{
			SessionHashLegacyReadHitRate: compatibility.SessionHashLegacyReadHitRate,
			MetadataLegacyFallbackTotal:  compatibility.MetadataLegacyFallbackTotal,
		},
	}
}

func SnapshotRuntimeObservability() RuntimeObservabilitySnapshot {
	scheduling := SnapshotSchedulingRuntimeKernelMetrics()
	idempotency := GetIdempotencyMetricsSnapshot()
	compatibility := SnapshotOpenAICompatibilityFallbackMetrics()

	return RuntimeObservabilitySnapshot{
		SchedulingRuntimeKernel:     scheduling,
		Idempotency:                 idempotency,
		OpenAICompatibilityFallback: compatibility,
		Summary:                     buildRuntimeObservabilitySummary(scheduling, idempotency, compatibility),
	}
}

func ratioOfInt64(numerator, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func ratioOfUint64(numerator uint64, denominator float64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / denominator
}

func ratioOfFloat64(numerator, denominator float64) float64 {
	if denominator <= 0 {
		return 0
	}
	return numerator / denominator
}
