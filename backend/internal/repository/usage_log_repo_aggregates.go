package repository

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/timezone"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
)

func (r *usageLogRepository) GetUserStatsAggregated(ctx context.Context, userID int64, startTime, endTime time.Time) (*usagestats.UsageStats, error) {
	return r.queryUsageStatsAggregate(ctx, startTime, endTime, usageLogScopedFilters{UserID: userID}, usageLogScopedFilterOptions{})
}

func (r *usageLogRepository) GetAPIKeyStatsAggregated(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) (*usagestats.UsageStats, error) {
	return r.queryUsageStatsAggregate(ctx, startTime, endTime, usageLogScopedFilters{APIKeyID: apiKeyID}, usageLogScopedFilterOptions{})
}

// GetAccountStatsAggregated 使用 SQL 聚合统计账号使用数据
func (r *usageLogRepository) GetAccountStatsAggregated(ctx context.Context, accountID int64, startTime, endTime time.Time) (*usagestats.UsageStats, error) {
	return r.queryUsageStatsAggregate(ctx, startTime, endTime, usageLogScopedFilters{AccountID: accountID}, usageLogScopedFilterOptions{})
}

// GetModelStatsAggregated 使用 SQL 聚合统计模型使用数据
func (r *usageLogRepository) GetModelStatsAggregated(ctx context.Context, modelName string, startTime, endTime time.Time) (*usagestats.UsageStats, error) {
	return r.queryUsageStatsAggregate(
		ctx,
		startTime,
		endTime,
		usageLogScopedFilters{Model: modelName},
		usageLogScopedFilterOptions{IncludeModel: true},
	)
}

// UsageStats represents usage statistics
type UsageStats = usagestats.UsageStats

// GetDailyStatsAggregated 使用 SQL 聚合统计用户的每日使用数据
func (r *usageLogRepository) GetDailyStatsAggregated(ctx context.Context, userID int64, startTime, endTime time.Time) (result []map[string]any, err error) {
	tzName := resolveUsageStatsTimezone()
	query := `
		SELECT
			TO_CHAR(created_at AT TIME ZONE $4, 'YYYY-MM-DD') as date,
			COUNT(*) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_input_tokens,
			COALESCE(SUM(output_tokens), 0) as total_output_tokens,
			COALESCE(SUM(cache_creation_tokens + cache_read_tokens), 0) as total_cache_tokens,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(SUM(actual_cost), 0) as total_actual_cost,
			COALESCE(AVG(COALESCE(duration_ms, 0)), 0) as avg_duration_ms
		FROM usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		GROUP BY 1
		ORDER BY 1
	`

	rows, err := r.sql.QueryContext(ctx, query, userID, startTime, endTime, tzName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			result = nil
		}
	}()

	result = make([]map[string]any, 0)
	for rows.Next() {
		var (
			date              string
			totalRequests     int64
			totalInputTokens  int64
			totalOutputTokens int64
			totalCacheTokens  int64
			totalCost         float64
			totalActualCost   float64
			avgDurationMs     float64
		)
		if err = rows.Scan(
			&date,
			&totalRequests,
			&totalInputTokens,
			&totalOutputTokens,
			&totalCacheTokens,
			&totalCost,
			&totalActualCost,
			&avgDurationMs,
		); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{
			"date":                date,
			"total_requests":      totalRequests,
			"total_input_tokens":  totalInputTokens,
			"total_output_tokens": totalOutputTokens,
			"total_cache_tokens":  totalCacheTokens,
			"total_tokens":        totalInputTokens + totalOutputTokens + totalCacheTokens,
			"total_cost":          totalCost,
			"total_actual_cost":   totalActualCost,
			"average_duration_ms": avgDurationMs,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// resolveUsageStatsTimezone 获取用于 SQL 分组的时区名称。
func resolveUsageStatsTimezone() string {
	tzName := timezone.Name()
	if tzName != "" && tzName != "Local" {
		return tzName
	}
	if envTZ := strings.TrimSpace(os.Getenv("TZ")); envTZ != "" {
		return envTZ
	}
	return "UTC"
}

func buildUsageStatsAggregateQuery(startTime, endTime time.Time, scopedFilters usageLogScopedFilters, options usageLogScopedFilterOptions) (string, []any) {
	query := `
		SELECT
			COUNT(*) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_input_tokens,
			COALESCE(SUM(output_tokens), 0) as total_output_tokens,
			COALESCE(SUM(cache_creation_tokens + cache_read_tokens), 0) as total_cache_tokens,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(SUM(actual_cost), 0) as total_actual_cost,
			COALESCE(AVG(COALESCE(duration_ms, 0)), 0) as avg_duration_ms
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
	`
	args := []any{startTime, endTime}
	extraConditions, args := appendUsageLogScopedConditions(nil, args, scopedFilters, options)
	query = appendConditionsToQuery(query, extraConditions)
	return query, args
}

func (r *usageLogRepository) queryUsageStatsAggregate(ctx context.Context, startTime, endTime time.Time, scopedFilters usageLogScopedFilters, options usageLogScopedFilterOptions) (*UsageStats, error) {
	query, args := buildUsageStatsAggregateQuery(startTime, endTime, scopedFilters, options)
	stats := &UsageStats{}
	if err := scanSingleRow(
		ctx,
		r.sql,
		query,
		args,
		&stats.TotalRequests,
		&stats.TotalInputTokens,
		&stats.TotalOutputTokens,
		&stats.TotalCacheTokens,
		&stats.TotalCost,
		&stats.TotalActualCost,
		&stats.AverageDurationMs,
	); err != nil {
		return nil, err
	}
	stats.TotalTokens = stats.TotalInputTokens + stats.TotalOutputTokens + stats.TotalCacheTokens
	return stats, nil
}
