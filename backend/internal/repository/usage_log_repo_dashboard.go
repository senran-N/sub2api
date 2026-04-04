package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/timezone"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"github.com/senran-N/sub2api/internal/service"
	"golang.org/x/sync/errgroup"
)

// DashboardStats 仪表盘统计
type DashboardStats = usagestats.DashboardStats

func (r *usageLogRepository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}
	now := timezone.Now()
	todayStart := timezone.Today()

	if !r.shouldRunConcurrentReadQueries() {
		if err := r.fillDashboardEntityStats(ctx, stats, todayStart, now); err != nil {
			return nil, err
		}
		if err := r.fillDashboardUsageStatsAggregated(ctx, stats, todayStart, now); err != nil {
			return nil, err
		}
		rpm, tpm, err := r.getPerformanceStats(ctx, 0)
		if err != nil {
			return nil, err
		}
		stats.Rpm = rpm
		stats.Tpm = tpm
		return stats, nil
	}

	var (
		rpm int64
		tpm int64
	)
	queryGroup, queryCtx := errgroup.WithContext(ctx)
	queryGroup.Go(func() error {
		return r.fillDashboardEntityStats(queryCtx, stats, todayStart, now)
	})
	queryGroup.Go(func() error {
		return r.fillDashboardUsageStatsAggregated(queryCtx, stats, todayStart, now)
	})
	queryGroup.Go(func() error {
		var queryErr error
		rpm, tpm, queryErr = r.getPerformanceStats(queryCtx, 0)
		return queryErr
	})
	if err := queryGroup.Wait(); err != nil {
		return nil, err
	}
	stats.Rpm = rpm
	stats.Tpm = tpm

	return stats, nil
}

func (r *usageLogRepository) GetDashboardStatsWithRange(ctx context.Context, start, end time.Time) (*DashboardStats, error) {
	startUTC := start.UTC()
	endUTC := end.UTC()
	if !endUTC.After(startUTC) {
		return nil, errors.New("统计时间范围无效")
	}

	stats := &DashboardStats{}
	now := timezone.Now()
	todayStart := timezone.Today()

	if !r.shouldRunConcurrentReadQueries() {
		if err := r.fillDashboardEntityStats(ctx, stats, todayStart, now); err != nil {
			return nil, err
		}
		if err := r.fillDashboardUsageStatsFromUsageLogs(ctx, stats, startUTC, endUTC, todayStart, now); err != nil {
			return nil, err
		}
		rpm, tpm, err := r.getPerformanceStats(ctx, 0)
		if err != nil {
			return nil, err
		}
		stats.Rpm = rpm
		stats.Tpm = tpm
		return stats, nil
	}

	var (
		rpm int64
		tpm int64
	)
	queryGroup, queryCtx := errgroup.WithContext(ctx)
	queryGroup.Go(func() error {
		return r.fillDashboardEntityStats(queryCtx, stats, todayStart, now)
	})
	queryGroup.Go(func() error {
		return r.fillDashboardUsageStatsFromUsageLogs(queryCtx, stats, startUTC, endUTC, todayStart, now)
	})
	queryGroup.Go(func() error {
		var queryErr error
		rpm, tpm, queryErr = r.getPerformanceStats(queryCtx, 0)
		return queryErr
	})
	if err := queryGroup.Wait(); err != nil {
		return nil, err
	}
	stats.Rpm = rpm
	stats.Tpm = tpm

	return stats, nil
}

func (r *usageLogRepository) fillDashboardEntityStats(ctx context.Context, stats *DashboardStats, todayUTC, now time.Time) error {
	userStatsQuery := `
		SELECT
			COUNT(*) as total_users,
			COUNT(CASE WHEN created_at >= $1 THEN 1 END) as today_new_users
		FROM users
		WHERE deleted_at IS NULL
	`
	apiKeyStatsQuery := `
		SELECT
			COUNT(*) as total_api_keys,
			COUNT(CASE WHEN status = $1 THEN 1 END) as active_api_keys
		FROM api_keys
		WHERE deleted_at IS NULL
	`
	accountStatsQuery := `
		SELECT
			COUNT(*) as total_accounts,
			COUNT(CASE WHEN status = $1 AND schedulable = true THEN 1 END) as normal_accounts,
			COUNT(CASE WHEN status = $2 THEN 1 END) as error_accounts,
			COUNT(CASE WHEN rate_limited_at IS NOT NULL AND rate_limit_reset_at > $3 THEN 1 END) as ratelimit_accounts,
			COUNT(CASE WHEN overload_until IS NOT NULL AND overload_until > $4 THEN 1 END) as overload_accounts
		FROM accounts
		WHERE deleted_at IS NULL
	`
	if !r.shouldRunConcurrentReadQueries() {
		if err := scanSingleRow(
			ctx,
			r.sql,
			userStatsQuery,
			[]any{todayUTC},
			&stats.TotalUsers,
			&stats.TodayNewUsers,
		); err != nil {
			return err
		}
		if err := scanSingleRow(
			ctx,
			r.sql,
			apiKeyStatsQuery,
			[]any{service.StatusActive},
			&stats.TotalAPIKeys,
			&stats.ActiveAPIKeys,
		); err != nil {
			return err
		}
		return scanSingleRow(
			ctx,
			r.sql,
			accountStatsQuery,
			[]any{service.StatusActive, service.StatusError, now, now},
			&stats.TotalAccounts,
			&stats.NormalAccounts,
			&stats.ErrorAccounts,
			&stats.RateLimitAccounts,
			&stats.OverloadAccounts,
		)
	}

	queryGroup, queryCtx := errgroup.WithContext(ctx)
	queryGroup.Go(func() error {
		return scanSingleRow(
			queryCtx,
			r.sql,
			userStatsQuery,
			[]any{todayUTC},
			&stats.TotalUsers,
			&stats.TodayNewUsers,
		)
	})
	queryGroup.Go(func() error {
		return scanSingleRow(
			queryCtx,
			r.sql,
			apiKeyStatsQuery,
			[]any{service.StatusActive},
			&stats.TotalAPIKeys,
			&stats.ActiveAPIKeys,
		)
	})
	queryGroup.Go(func() error {
		return scanSingleRow(
			queryCtx,
			r.sql,
			accountStatsQuery,
			[]any{service.StatusActive, service.StatusError, now, now},
			&stats.TotalAccounts,
			&stats.NormalAccounts,
			&stats.ErrorAccounts,
			&stats.RateLimitAccounts,
			&stats.OverloadAccounts,
		)
	})
	return queryGroup.Wait()
}

func (r *usageLogRepository) fillDashboardUsageStatsAggregated(ctx context.Context, stats *DashboardStats, todayUTC, now time.Time) error {
	combinedDailyStatsQuery := `
		SELECT
			COALESCE(SUM(total_requests), 0) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_input_tokens,
			COALESCE(SUM(output_tokens), 0) as total_output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) as total_cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) as total_cache_read_tokens,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(SUM(actual_cost), 0) as total_actual_cost,
			COALESCE(SUM(total_duration_ms), 0) as total_duration_ms,
			COALESCE(SUM(total_requests) FILTER (WHERE bucket_date = $1::date), 0) AS today_requests,
			COALESCE(SUM(input_tokens) FILTER (WHERE bucket_date = $1::date), 0) AS today_input_tokens,
			COALESCE(SUM(output_tokens) FILTER (WHERE bucket_date = $1::date), 0) AS today_output_tokens,
			COALESCE(SUM(cache_creation_tokens) FILTER (WHERE bucket_date = $1::date), 0) AS today_cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens) FILTER (WHERE bucket_date = $1::date), 0) AS today_cache_read_tokens,
			COALESCE(SUM(total_cost) FILTER (WHERE bucket_date = $1::date), 0) AS today_cost,
			COALESCE(SUM(actual_cost) FILTER (WHERE bucket_date = $1::date), 0) AS today_actual_cost,
			COALESCE(MAX(active_users) FILTER (WHERE bucket_date = $1::date), 0) AS active_users
		FROM usage_dashboard_daily
	`
	var totalDurationMs int64

	hourlyActiveQuery := `
		SELECT active_users
		FROM usage_dashboard_hourly
		WHERE bucket_start = $1
	`
	hourStart := now.In(timezone.Location()).Truncate(time.Hour)
	if !r.shouldRunConcurrentReadQueries() {
		if err := scanSingleRow(
			ctx,
			r.sql,
			combinedDailyStatsQuery,
			[]any{todayUTC},
			&stats.TotalRequests,
			&stats.TotalInputTokens,
			&stats.TotalOutputTokens,
			&stats.TotalCacheCreationTokens,
			&stats.TotalCacheReadTokens,
			&stats.TotalCost,
			&stats.TotalActualCost,
			&totalDurationMs,
			&stats.TodayRequests,
			&stats.TodayInputTokens,
			&stats.TodayOutputTokens,
			&stats.TodayCacheCreationTokens,
			&stats.TodayCacheReadTokens,
			&stats.TodayCost,
			&stats.TodayActualCost,
			&stats.ActiveUsers,
		); err != nil {
			return err
		}
		if err := scanSingleRow(ctx, r.sql, hourlyActiveQuery, []any{hourStart}, &stats.HourlyActiveUsers); err != nil && err != sql.ErrNoRows {
			return err
		}
	} else {
		queryGroup, queryCtx := errgroup.WithContext(ctx)
		queryGroup.Go(func() error {
			return scanSingleRow(
				queryCtx,
				r.sql,
				combinedDailyStatsQuery,
				[]any{todayUTC},
				&stats.TotalRequests,
				&stats.TotalInputTokens,
				&stats.TotalOutputTokens,
				&stats.TotalCacheCreationTokens,
				&stats.TotalCacheReadTokens,
				&stats.TotalCost,
				&stats.TotalActualCost,
				&totalDurationMs,
				&stats.TodayRequests,
				&stats.TodayInputTokens,
				&stats.TodayOutputTokens,
				&stats.TodayCacheCreationTokens,
				&stats.TodayCacheReadTokens,
				&stats.TodayCost,
				&stats.TodayActualCost,
				&stats.ActiveUsers,
			)
		})
		queryGroup.Go(func() error {
			err := scanSingleRow(queryCtx, r.sql, hourlyActiveQuery, []any{hourStart}, &stats.HourlyActiveUsers)
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		})
		if err := queryGroup.Wait(); err != nil {
			return err
		}
	}

	stats.TotalTokens = stats.TotalInputTokens + stats.TotalOutputTokens + stats.TotalCacheCreationTokens + stats.TotalCacheReadTokens
	if stats.TotalRequests > 0 {
		stats.AverageDurationMs = float64(totalDurationMs) / float64(stats.TotalRequests)
	}
	stats.TodayTokens = stats.TodayInputTokens + stats.TodayOutputTokens + stats.TodayCacheCreationTokens + stats.TodayCacheReadTokens

	return nil
}

func (r *usageLogRepository) fillDashboardUsageStatsFromUsageLogs(ctx context.Context, stats *DashboardStats, startUTC, endUTC, todayUTC, now time.Time) error {
	todayEnd := todayUTC.Add(24 * time.Hour)
	hourStart := now.UTC().Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)
	combinedStatsQuery := `
		WITH scoped AS (
			SELECT
				user_id,
				created_at,
				input_tokens,
				output_tokens,
				cache_creation_tokens,
				cache_read_tokens,
				total_cost,
				actual_cost,
				COALESCE(duration_ms, 0) AS duration_ms
			FROM usage_logs
			WHERE created_at >= LEAST($1::timestamptz, $3::timestamptz, $5::timestamptz)
				AND created_at < GREATEST($2::timestamptz, $4::timestamptz, $6::timestamptz)
		)
		SELECT
			COUNT(*) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz) AS total_requests,
			COALESCE(SUM(input_tokens) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz), 0) AS total_input_tokens,
			COALESCE(SUM(output_tokens) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz), 0) AS total_output_tokens,
			COALESCE(SUM(cache_creation_tokens) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz), 0) AS total_cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz), 0) AS total_cache_read_tokens,
			COALESCE(SUM(total_cost) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz), 0) AS total_cost,
			COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz), 0) AS total_actual_cost,
			COALESCE(SUM(duration_ms) FILTER (WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz), 0) AS total_duration_ms,
			COUNT(*) FILTER (WHERE created_at >= $3::timestamptz AND created_at < $4::timestamptz) AS today_requests,
			COALESCE(SUM(input_tokens) FILTER (WHERE created_at >= $3::timestamptz AND created_at < $4::timestamptz), 0) AS today_input_tokens,
			COALESCE(SUM(output_tokens) FILTER (WHERE created_at >= $3::timestamptz AND created_at < $4::timestamptz), 0) AS today_output_tokens,
			COALESCE(SUM(cache_creation_tokens) FILTER (WHERE created_at >= $3::timestamptz AND created_at < $4::timestamptz), 0) AS today_cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens) FILTER (WHERE created_at >= $3::timestamptz AND created_at < $4::timestamptz), 0) AS today_cache_read_tokens,
			COALESCE(SUM(total_cost) FILTER (WHERE created_at >= $3::timestamptz AND created_at < $4::timestamptz), 0) AS today_cost,
			COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $3::timestamptz AND created_at < $4::timestamptz), 0) AS today_actual_cost,
			COUNT(DISTINCT CASE WHEN created_at >= $3::timestamptz AND created_at < $4::timestamptz THEN user_id END) AS active_users,
			COUNT(DISTINCT CASE WHEN created_at >= $5::timestamptz AND created_at < $6::timestamptz THEN user_id END) AS hourly_active_users
		FROM scoped
	`
	var totalDurationMs int64
	if err := scanSingleRow(
		ctx,
		r.sql,
		combinedStatsQuery,
		[]any{startUTC, endUTC, todayUTC, todayEnd, hourStart, hourEnd},
		&stats.TotalRequests,
		&stats.TotalInputTokens,
		&stats.TotalOutputTokens,
		&stats.TotalCacheCreationTokens,
		&stats.TotalCacheReadTokens,
		&stats.TotalCost,
		&stats.TotalActualCost,
		&totalDurationMs,
		&stats.TodayRequests,
		&stats.TodayInputTokens,
		&stats.TodayOutputTokens,
		&stats.TodayCacheCreationTokens,
		&stats.TodayCacheReadTokens,
		&stats.TodayCost,
		&stats.TodayActualCost,
		&stats.ActiveUsers,
		&stats.HourlyActiveUsers,
	); err != nil {
		return err
	}
	stats.TotalTokens = stats.TotalInputTokens + stats.TotalOutputTokens + stats.TotalCacheCreationTokens + stats.TotalCacheReadTokens
	if stats.TotalRequests > 0 {
		stats.AverageDurationMs = float64(totalDurationMs) / float64(stats.TotalRequests)
	}

	stats.TodayTokens = stats.TodayInputTokens + stats.TodayOutputTokens + stats.TodayCacheCreationTokens + stats.TodayCacheReadTokens

	return nil
}

// UserDashboardStats 用户仪表盘统计
type UserDashboardStats = usagestats.UserDashboardStats

type dashboardUsageAggregate struct {
	TotalRequests            int64
	TotalInputTokens         int64
	TotalOutputTokens        int64
	TotalCacheCreationTokens int64
	TotalCacheReadTokens     int64
	TotalCost                float64
	TotalActualCost          float64
	AverageDurationMs        float64
	TodayRequests            int64
	TodayInputTokens         int64
	TodayOutputTokens        int64
	TodayCacheCreationTokens int64
	TodayCacheReadTokens     int64
	TodayCost                float64
	TodayActualCost          float64
}

type dashboardAPIKeyCounts struct {
	TotalAPIKeys  int64
	ActiveAPIKeys int64
}

func (r *usageLogRepository) shouldRunConcurrentReadQueries() bool {
	return r != nil && r.db != nil
}

func (r *usageLogRepository) loadDashboardUsageAggregate(ctx context.Context, scopeColumn string, scopeID int64, today time.Time) (dashboardUsageAggregate, error) {
	switch scopeColumn {
	case "user_id", "api_key_id":
	default:
		return dashboardUsageAggregate{}, fmt.Errorf("unsupported dashboard scope column: %s", scopeColumn)
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) AS total_requests,
			COALESCE(SUM(input_tokens), 0) AS total_input_tokens,
			COALESCE(SUM(output_tokens), 0) AS total_output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) AS total_cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) AS total_cache_read_tokens,
			COALESCE(SUM(total_cost), 0) AS total_cost,
			COALESCE(SUM(actual_cost), 0) AS total_actual_cost,
			COALESCE(AVG(duration_ms), 0) AS avg_duration_ms,
			COUNT(*) FILTER (WHERE created_at >= $2) AS today_requests,
			COALESCE(SUM(input_tokens) FILTER (WHERE created_at >= $2), 0) AS today_input_tokens,
			COALESCE(SUM(output_tokens) FILTER (WHERE created_at >= $2), 0) AS today_output_tokens,
			COALESCE(SUM(cache_creation_tokens) FILTER (WHERE created_at >= $2), 0) AS today_cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens) FILTER (WHERE created_at >= $2), 0) AS today_cache_read_tokens,
			COALESCE(SUM(total_cost) FILTER (WHERE created_at >= $2), 0) AS today_cost,
			COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $2), 0) AS today_actual_cost
		FROM usage_logs
		WHERE %s = $1
	`, scopeColumn)

	aggregate := dashboardUsageAggregate{}
	err := scanSingleRow(
		ctx,
		r.sql,
		query,
		[]any{scopeID, today},
		&aggregate.TotalRequests,
		&aggregate.TotalInputTokens,
		&aggregate.TotalOutputTokens,
		&aggregate.TotalCacheCreationTokens,
		&aggregate.TotalCacheReadTokens,
		&aggregate.TotalCost,
		&aggregate.TotalActualCost,
		&aggregate.AverageDurationMs,
		&aggregate.TodayRequests,
		&aggregate.TodayInputTokens,
		&aggregate.TodayOutputTokens,
		&aggregate.TodayCacheCreationTokens,
		&aggregate.TodayCacheReadTokens,
		&aggregate.TodayCost,
		&aggregate.TodayActualCost,
	)
	if err != nil {
		return dashboardUsageAggregate{}, err
	}
	return aggregate, nil
}

func (r *usageLogRepository) loadDashboardAPIKeyCounts(ctx context.Context, userID int64) (dashboardAPIKeyCounts, error) {
	query := `
		SELECT
			COUNT(*) AS total_api_keys,
			COUNT(*) FILTER (WHERE status = $2) AS active_api_keys
		FROM api_keys
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	counts := dashboardAPIKeyCounts{}
	err := scanSingleRow(
		ctx,
		r.sql,
		query,
		[]any{userID, service.StatusActive},
		&counts.TotalAPIKeys,
		&counts.ActiveAPIKeys,
	)
	if err != nil {
		return dashboardAPIKeyCounts{}, err
	}
	return counts, nil
}

func applyDashboardUsageAggregate(stats *UserDashboardStats, aggregate dashboardUsageAggregate) {
	stats.TotalRequests = aggregate.TotalRequests
	stats.TotalInputTokens = aggregate.TotalInputTokens
	stats.TotalOutputTokens = aggregate.TotalOutputTokens
	stats.TotalCacheCreationTokens = aggregate.TotalCacheCreationTokens
	stats.TotalCacheReadTokens = aggregate.TotalCacheReadTokens
	stats.TotalCost = aggregate.TotalCost
	stats.TotalActualCost = aggregate.TotalActualCost
	stats.AverageDurationMs = aggregate.AverageDurationMs
	stats.TotalTokens = stats.TotalInputTokens + stats.TotalOutputTokens + stats.TotalCacheCreationTokens + stats.TotalCacheReadTokens

	stats.TodayRequests = aggregate.TodayRequests
	stats.TodayInputTokens = aggregate.TodayInputTokens
	stats.TodayOutputTokens = aggregate.TodayOutputTokens
	stats.TodayCacheCreationTokens = aggregate.TodayCacheCreationTokens
	stats.TodayCacheReadTokens = aggregate.TodayCacheReadTokens
	stats.TodayCost = aggregate.TodayCost
	stats.TodayActualCost = aggregate.TodayActualCost
	stats.TodayTokens = stats.TodayInputTokens + stats.TodayOutputTokens + stats.TodayCacheCreationTokens + stats.TodayCacheReadTokens
}

// GetUserDashboardStats 获取用户专属的仪表盘统计
func (r *usageLogRepository) GetUserDashboardStats(ctx context.Context, userID int64) (*UserDashboardStats, error) {
	stats := &UserDashboardStats{}
	today := timezone.Today()

	var (
		aggregate dashboardUsageAggregate
		counts    dashboardAPIKeyCounts
		rpm       int64
		tpm       int64
	)

	if !r.shouldRunConcurrentReadQueries() {
		var err error
		counts, err = r.loadDashboardAPIKeyCounts(ctx, userID)
		if err != nil {
			return nil, err
		}
		aggregate, err = r.loadDashboardUsageAggregate(ctx, "user_id", userID, today)
		if err != nil {
			return nil, err
		}
		rpm, tpm, err = r.getPerformanceStats(ctx, userID)
		if err != nil {
			return nil, err
		}
	} else {
		queryGroup, queryCtx := errgroup.WithContext(ctx)
		queryGroup.Go(func() error {
			var queryErr error
			counts, queryErr = r.loadDashboardAPIKeyCounts(queryCtx, userID)
			return queryErr
		})
		queryGroup.Go(func() error {
			var queryErr error
			aggregate, queryErr = r.loadDashboardUsageAggregate(queryCtx, "user_id", userID, today)
			return queryErr
		})
		queryGroup.Go(func() error {
			var queryErr error
			rpm, tpm, queryErr = r.getPerformanceStats(queryCtx, userID)
			return queryErr
		})
		if err := queryGroup.Wait(); err != nil {
			return nil, err
		}
	}

	stats.TotalAPIKeys = counts.TotalAPIKeys
	stats.ActiveAPIKeys = counts.ActiveAPIKeys
	applyDashboardUsageAggregate(stats, aggregate)
	stats.Rpm = rpm
	stats.Tpm = tpm

	return stats, nil
}

// getPerformanceStatsByAPIKey 获取指定 API Key 的 RPM 和 TPM（近5分钟平均值）
func (r *usageLogRepository) getPerformanceStatsByAPIKey(ctx context.Context, apiKeyID int64) (rpm, tpm int64, err error) {
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	query := `
		SELECT
			COUNT(*) as request_count,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as token_count
		FROM usage_logs
		WHERE created_at >= $1 AND api_key_id = $2`
	args := []any{fiveMinutesAgo, apiKeyID}

	var requestCount int64
	var tokenCount int64
	if err := scanSingleRow(ctx, r.sql, query, args, &requestCount, &tokenCount); err != nil {
		return 0, 0, err
	}
	return requestCount / 5, tokenCount / 5, nil
}

// GetAPIKeyDashboardStats 获取指定 API Key 的仪表盘统计（按 api_key_id 过滤）
func (r *usageLogRepository) GetAPIKeyDashboardStats(ctx context.Context, apiKeyID int64) (*UserDashboardStats, error) {
	stats := &UserDashboardStats{
		TotalAPIKeys:  1,
		ActiveAPIKeys: 1,
	}
	today := timezone.Today()

	var (
		aggregate dashboardUsageAggregate
		rpm       int64
		tpm       int64
	)

	if !r.shouldRunConcurrentReadQueries() {
		var err error
		aggregate, err = r.loadDashboardUsageAggregate(ctx, "api_key_id", apiKeyID, today)
		if err != nil {
			return nil, err
		}
		rpm, tpm, err = r.getPerformanceStatsByAPIKey(ctx, apiKeyID)
		if err != nil {
			return nil, err
		}
	} else {
		queryGroup, queryCtx := errgroup.WithContext(ctx)
		queryGroup.Go(func() error {
			var queryErr error
			aggregate, queryErr = r.loadDashboardUsageAggregate(queryCtx, "api_key_id", apiKeyID, today)
			return queryErr
		})
		queryGroup.Go(func() error {
			var queryErr error
			rpm, tpm, queryErr = r.getPerformanceStatsByAPIKey(queryCtx, apiKeyID)
			return queryErr
		})
		if err := queryGroup.Wait(); err != nil {
			return nil, err
		}
	}

	applyDashboardUsageAggregate(stats, aggregate)
	stats.Rpm = rpm
	stats.Tpm = tpm

	return stats, nil
}
