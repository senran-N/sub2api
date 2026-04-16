//go:build unit

package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var _ OpsRepository = (*stubOpsRepo)(nil)

type stubOpsRepo struct {
	OpsRepository
	overview *OpsDashboardOverview
	err      error
}

func (s *stubOpsRepo) GetDashboardOverview(ctx context.Context, filter *OpsDashboardFilter) (*OpsDashboardOverview, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.overview != nil {
		return s.overview, nil
	}
	return &OpsDashboardOverview{}, nil
}

func TestComputeGroupAvailableRatio(t *testing.T) {
	t.Parallel()

	t.Run("正常情况: 10个账号, 8个可用 = 80%", func(t *testing.T) {
		t.Parallel()

		got := computeGroupAvailableRatio(&GroupAvailability{
			TotalAccounts:  10,
			AvailableCount: 8,
		})
		require.InDelta(t, 80.0, got, 0.0001)
	})

	t.Run("边界情况: TotalAccounts = 0 应返回 0", func(t *testing.T) {
		t.Parallel()

		got := computeGroupAvailableRatio(&GroupAvailability{
			TotalAccounts:  0,
			AvailableCount: 8,
		})
		require.Equal(t, 0.0, got)
	})

	t.Run("边界情况: AvailableCount = 0 应返回 0%", func(t *testing.T) {
		t.Parallel()

		got := computeGroupAvailableRatio(&GroupAvailability{
			TotalAccounts:  10,
			AvailableCount: 0,
		})
		require.Equal(t, 0.0, got)
	})
}

func TestCountAccountsByCondition(t *testing.T) {
	t.Parallel()

	t.Run("测试限流账号统计: acc.IsRateLimited", func(t *testing.T) {
		t.Parallel()

		accounts := map[int64]*AccountAvailability{
			1: {IsRateLimited: true},
			2: {IsRateLimited: false},
			3: {IsRateLimited: true},
		}

		got := countAccountsByCondition(accounts, func(acc *AccountAvailability) bool {
			return acc.IsRateLimited
		})
		require.Equal(t, int64(2), got)
	})

	t.Run("测试错误账号统计（排除临时不可调度）: acc.HasError && acc.TempUnschedulableUntil == nil", func(t *testing.T) {
		t.Parallel()

		until := time.Now().UTC().Add(5 * time.Minute)
		accounts := map[int64]*AccountAvailability{
			1: {HasError: true},
			2: {HasError: true, TempUnschedulableUntil: &until},
			3: {HasError: false},
		}

		got := countAccountsByCondition(accounts, func(acc *AccountAvailability) bool {
			return acc.HasError && acc.TempUnschedulableUntil == nil
		})
		require.Equal(t, int64(1), got)
	})

	t.Run("边界情况: 空 map 应返回 0", func(t *testing.T) {
		t.Parallel()

		got := countAccountsByCondition(map[int64]*AccountAvailability{}, func(acc *AccountAvailability) bool {
			return acc.IsRateLimited
		})
		require.Equal(t, int64(0), got)
	})
}

func TestComputeRuleMetricNewIndicators(t *testing.T) {
	t.Parallel()

	groupID := int64(101)
	platform := "openai"

	availability := &OpsAccountAvailability{
		Group: &GroupAvailability{
			GroupID:        groupID,
			TotalAccounts:  10,
			AvailableCount: 8,
		},
		Accounts: map[int64]*AccountAvailability{
			1: {IsRateLimited: true},
			2: {IsRateLimited: true},
			3: {HasError: true},
			4: {HasError: true, TempUnschedulableUntil: timePtr(time.Now().UTC().Add(2 * time.Minute))},
			5: {HasError: false, IsRateLimited: false},
		},
	}

	opsService := &OpsService{
		getAccountAvailability: func(_ context.Context, _ string, _ *int64) (*OpsAccountAvailability, error) {
			return availability, nil
		},
	}

	svc := &OpsAlertEvaluatorService{
		opsService: opsService,
		opsRepo:    &stubOpsRepo{overview: &OpsDashboardOverview{}},
	}

	start := time.Now().UTC().Add(-5 * time.Minute)
	end := time.Now().UTC()
	ctx := context.Background()

	tests := []struct {
		name       string
		metricType string
		groupID    *int64
		wantValue  float64
		wantOK     bool
	}{
		{
			name:       "group_available_accounts",
			metricType: "group_available_accounts",
			groupID:    &groupID,
			wantValue:  8,
			wantOK:     true,
		},
		{
			name:       "group_available_ratio",
			metricType: "group_available_ratio",
			groupID:    &groupID,
			wantValue:  80.0,
			wantOK:     true,
		},
		{
			name:       "account_rate_limited_count",
			metricType: "account_rate_limited_count",
			groupID:    nil,
			wantValue:  2,
			wantOK:     true,
		},
		{
			name:       "account_error_count",
			metricType: "account_error_count",
			groupID:    nil,
			wantValue:  1,
			wantOK:     true,
		},
		{
			name:       "group_available_accounts without group_id returns false",
			metricType: "group_available_accounts",
			groupID:    nil,
			wantValue:  0,
			wantOK:     false,
		},
		{
			name:       "group_available_ratio without group_id returns false",
			metricType: "group_available_ratio",
			groupID:    nil,
			wantValue:  0,
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := &OpsAlertRule{
				MetricType: tt.metricType,
			}
			gotValue, gotOK := svc.computeRuleMetric(ctx, rule, nil, start, end, platform, tt.groupID)
			require.Equal(t, tt.wantOK, gotOK)
			if !tt.wantOK {
				return
			}
			require.InDelta(t, tt.wantValue, gotValue, 0.0001)
		})
	}
}

func TestComputeRuntimeObservabilityMetric(t *testing.T) {
	t.Parallel()

	snapshot := RuntimeObservabilitySnapshot{
		Summary: RuntimeObservabilitySummary{
			SchedulingRuntimeKernel: SchedulingRuntimeKernelSummary{
				AcquireSuccessRate:        0.7,
				WaitPlanSuccessRate:       0.75,
				AvgFetchedAccountsPerPage: 50,
			},
			OpenAIAccountScheduler: RuntimeOpenAIAccountSchedulerSummary{
				StickyIntentMissRate:    0.12,
				NonStickyIntentShare:    0.35,
				IndexedLoadBalanceShare: 0.27,
			},
			Idempotency: RuntimeIdempotencySummary{
				AvgProcessingDurationMs: 30,
			},
		},
	}

	tests := []struct {
		name       string
		metricType string
		wantValue  float64
		wantOK     bool
	}{
		{name: "scheduler_acquire_success_rate", metricType: "scheduler_acquire_success_rate", wantValue: 70, wantOK: true},
		{name: "scheduler_wait_plan_success_rate", metricType: "scheduler_wait_plan_success_rate", wantValue: 75, wantOK: true},
		{name: "scheduler_index_page_density", metricType: "scheduler_index_page_density", wantValue: 50, wantOK: true},
		{name: "openai_scheduler_sticky_intent_miss_rate", metricType: "openai_scheduler_sticky_intent_miss_rate", wantValue: 12, wantOK: true},
		{name: "openai_scheduler_non_sticky_share", metricType: "openai_scheduler_non_sticky_share", wantValue: 35, wantOK: true},
		{name: "openai_scheduler_indexed_load_balance_share", metricType: "openai_scheduler_indexed_load_balance_share", wantValue: 27, wantOK: true},
		{name: "idempotency_processing_avg_ms", metricType: "idempotency_processing_avg_ms", wantValue: 30, wantOK: true},
		{name: "unknown", metricType: "nope", wantValue: 0, wantOK: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOK := computeRuntimeObservabilityMetric(tt.metricType, snapshot)
			require.Equal(t, tt.wantOK, gotOK)
			if !tt.wantOK {
				return
			}
			require.InDelta(t, tt.wantValue, gotValue, 0.0001)
		})
	}
}

func TestBuildOpsAlertDescriptionForRuntimeMetrics(t *testing.T) {
	t.Parallel()

	groupID := int64(42)

	tests := []struct {
		name       string
		rule       *OpsAlertRule
		value      float64
		wantParts  []string
		platform   string
		groupID    *int64
		windowMins int
	}{
		{
			name: "scheduler_acquire_success_rate",
			rule: &OpsAlertRule{
				MetricType: "scheduler_acquire_success_rate",
				Operator:   "<",
				Threshold:  75,
			},
			value:      63.5,
			windowMins: 5,
			platform:   "openai",
			groupID:    &groupID,
			wantParts: []string{
				"Unified scheduling kernel acquire success fell to 63.50%",
				"below the 75.00% threshold",
				"platform=openai group_id=42",
			},
		},
		{
			name: "scheduler_wait_plan_success_rate",
			rule: &OpsAlertRule{
				MetricType: "scheduler_wait_plan_success_rate",
				Operator:   "<",
				Threshold:  60,
			},
			value:      42,
			windowMins: 5,
			wantParts: []string{
				"wait-plan payoff dropped to 42.00%",
				"below the 60.00% threshold",
				"overall",
			},
		},
		{
			name: "scheduler_index_page_density",
			rule: &OpsAlertRule{
				MetricType: "scheduler_index_page_density",
				Operator:   "<",
				Threshold:  8,
			},
			value:      3.5,
			windowMins: 1,
			wantParts: []string{
				"index page density is 3.50 accounts/page",
				"below the 8.00 threshold",
			},
		},
		{
			name: "idempotency_processing_avg_ms",
			rule: &OpsAlertRule{
				MetricType: "idempotency_processing_avg_ms",
				Operator:   ">",
				Threshold:  80,
			},
			value:      140.25,
			windowMins: 1,
			wantParts: []string{
				"Idempotency processing averaged 140.25ms",
				"above the 80.00ms threshold",
			},
		},
		{
			name: "openai_scheduler_sticky_intent_miss_rate",
			rule: &OpsAlertRule{
				MetricType: "openai_scheduler_sticky_intent_miss_rate",
				Operator:   ">",
				Threshold:  10,
			},
			value:      18.5,
			windowMins: 3,
			wantParts: []string{
				"sticky-intent miss fallback climbed to 18.50%",
				"above the 10.00% threshold",
				"indexed snapshot and runtime acquire cost",
			},
		},
		{
			name: "openai_scheduler_non_sticky_share",
			rule: &OpsAlertRule{
				MetricType: "openai_scheduler_non_sticky_share",
				Operator:   ">",
				Threshold:  25,
			},
			value:      41,
			windowMins: 3,
			wantParts: []string{
				"non-sticky request share reached 41.00%",
				"above the 25.00% threshold",
			},
		},
		{
			name: "openai_scheduler_indexed_load_balance_share",
			rule: &OpsAlertRule{
				MetricType: "openai_scheduler_indexed_load_balance_share",
				Operator:   ">",
				Threshold:  20,
			},
			value:      33.3,
			windowMins: 3,
			wantParts: []string{
				"indexed load-balance share reached 33.30%",
				"above the 20.00% threshold",
				"instead of cheap sticky hits",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description := buildOpsAlertDescription(tt.rule, tt.value, tt.windowMins, tt.platform, tt.groupID)
			for _, part := range tt.wantParts {
				require.True(t, strings.Contains(description, part), "description=%q should contain %q", description, part)
			}
		})
	}
}
