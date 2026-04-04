package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"golang.org/x/sync/errgroup"
)

// GetGlobalStats gets usage statistics for all users within a time range
func (r *usageLogRepository) GetGlobalStats(ctx context.Context, startTime, endTime time.Time) (*UsageStats, error) {
	return r.queryUsageStatsAggregate(ctx, startTime, endTime, usageLogScopedFilters{}, usageLogScopedFilterOptions{})
}

// GetStatsWithFilters gets usage statistics with optional filters
func (r *usageLogRepository) GetStatsWithFilters(ctx context.Context, filters UsageLogFilters) (*UsageStats, error) {
	conditions, args := appendUsageLogScopedConditions(
		make([]string, 0, 9),
		make([]any, 0, 9),
		usageLogScopedFiltersFromUsageLogFilters(filters),
		usageLogScopedFilterOptions{IncludeModel: true, IncludeTime: true},
	)

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_input_tokens,
			COALESCE(SUM(output_tokens), 0) as total_output_tokens,
			COALESCE(SUM(cache_creation_tokens + cache_read_tokens), 0) as total_cache_tokens,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(SUM(actual_cost), 0) as total_actual_cost,
			COALESCE(SUM(total_cost * COALESCE(account_rate_multiplier, 1)), 0) as total_account_cost,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms
		FROM usage_logs
		%s
	`, buildWhere(conditions))

	stats := &UsageStats{}

	start := time.Unix(0, 0).UTC()
	if filters.StartTime != nil {
		start = *filters.StartTime
	}
	end := time.Now().UTC()
	if filters.EndTime != nil {
		end = *filters.EndTime
	}

	var (
		totalRequests     int64
		totalInputTokens  int64
		totalOutputTokens int64
		totalCacheTokens  int64
		totalCost         float64
		totalActualCost   float64
		totalAccountCost  float64
		averageDuration   float64
		endpointStats     usageLogEndpointStatsResult
	)
	loadAggregate := func(queryCtx context.Context) error {
		return scanSingleRow(
			queryCtx,
			r.sql,
			query,
			args,
			&totalRequests,
			&totalInputTokens,
			&totalOutputTokens,
			&totalCacheTokens,
			&totalCost,
			&totalActualCost,
			&totalAccountCost,
			&averageDuration,
		)
	}
	loadEndpointStats := func(queryCtx context.Context) {
		endpointStats = r.collectEndpointStatsWithFilters(queryCtx, start, end, filters.UserID, filters.APIKeyID, filters.AccountID, filters.GroupID, filters.Model, filters.RequestType, filters.Stream, filters.BillingType, true)
	}
	if !r.shouldRunConcurrentReadQueries() {
		if err := loadAggregate(ctx); err != nil {
			return nil, err
		}
		loadEndpointStats(ctx)
	} else {
		queryGroup, queryCtx := errgroup.WithContext(ctx)
		queryGroup.Go(func() error {
			return loadAggregate(queryCtx)
		})
		queryGroup.Go(func() error {
			loadEndpointStats(queryCtx)
			return nil
		})
		if err := queryGroup.Wait(); err != nil {
			return nil, err
		}
	}

	stats.TotalRequests = totalRequests
	stats.TotalInputTokens = totalInputTokens
	stats.TotalOutputTokens = totalOutputTokens
	stats.TotalCacheTokens = totalCacheTokens
	stats.TotalCost = totalCost
	stats.TotalActualCost = totalActualCost
	stats.AverageDurationMs = averageDuration
	if filters.AccountID > 0 {
		stats.TotalAccountCost = &totalAccountCost
	}
	stats.TotalTokens = stats.TotalInputTokens + stats.TotalOutputTokens + stats.TotalCacheTokens
	stats.Endpoints = endpointStats.Endpoints
	stats.UpstreamEndpoints = endpointStats.UpstreamEndpoints
	stats.EndpointPaths = endpointStats.EndpointPaths

	return stats, nil
}

// EndpointStat represents endpoint usage statistics row.
type EndpointStat = usagestats.EndpointStat

type usageLogEndpointStatsResult struct {
	Endpoints         []EndpointStat
	UpstreamEndpoints []EndpointStat
	EndpointPaths     []EndpointStat
}

type usageLogEndpointStatsBucketRow struct {
	Kind EndpointStatKind
	Stat EndpointStat
}

type EndpointStatKind string

const (
	endpointStatKindInbound  EndpointStatKind = "inbound"
	endpointStatKindUpstream EndpointStatKind = "upstream"
	endpointStatKindPath     EndpointStatKind = "path"
)

func (r *usageLogRepository) collectEndpointStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8, includePaths bool) usageLogEndpointStatsResult {
	result := usageLogEndpointStatsResult{
		Endpoints:         []EndpointStat{},
		UpstreamEndpoints: []EndpointStat{},
	}
	if includePaths {
		result.EndpointPaths = []EndpointStat{}
	}

	rows, err := r.queryEndpointStatsByKindsWithFilters(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, model, requestType, stream, billingType, includePaths)
	if err != nil {
		logger.LegacyPrintf("repository.usage_log", "queryEndpointStatsByKindsWithFilters failed: %v", err)
		return result
	}
	result = groupEndpointStatsByKind(rows, includePaths)
	return result
}

func buildEndpointDimensionValues(includePaths bool) string {
	values := []string{
		fmt.Sprintf("('%s', COALESCE(NULLIF(TRIM(ul.inbound_endpoint), ''), 'unknown'))", endpointStatKindInbound),
		fmt.Sprintf("('%s', COALESCE(NULLIF(TRIM(ul.upstream_endpoint), ''), 'unknown'))", endpointStatKindUpstream),
	}
	if includePaths {
		values = append(values, fmt.Sprintf(
			"('%s', CONCAT(COALESCE(NULLIF(TRIM(ul.inbound_endpoint), ''), 'unknown'), ' -> ', COALESCE(NULLIF(TRIM(ul.upstream_endpoint), ''), 'unknown')))",
			endpointStatKindPath,
		))
	}
	return strings.Join(values, ",")
}

func (r *usageLogRepository) queryEndpointStatsByKindsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8, includePaths bool) (results []usageLogEndpointStatsBucketRow, err error) {
	query := fmt.Sprintf(`
		SELECT
			dim.stat_type,
			dim.endpoint,
			COUNT(*) AS requests,
			COALESCE(SUM(ul.input_tokens + ul.output_tokens + ul.cache_creation_tokens + ul.cache_read_tokens), 0) AS total_tokens,
			COALESCE(SUM(ul.total_cost), 0) as cost,
			%s
		FROM usage_logs ul
		CROSS JOIN LATERAL (VALUES %s) AS dim(stat_type, endpoint)
		WHERE ul.created_at >= $1 AND ul.created_at < $2
	`, usageLogActualCostAggregateExpr("ul", userID, apiKeyID, accountID), buildEndpointDimensionValues(includePaths))

	args := []any{startTime, endTime}
	extraConditions, args := appendUsageLogScopedConditions(nil, args, usageLogScopedFilters{
		UserID:      userID,
		APIKeyID:    apiKeyID,
		AccountID:   accountID,
		GroupID:     groupID,
		Model:       model,
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
	}, usageLogScopedFilterOptions{IncludeModel: true, TableAlias: "ul"})
	query = appendConditionsToQuery(query, extraConditions)
	query += " GROUP BY dim.stat_type, dim.endpoint ORDER BY dim.stat_type ASC, requests DESC, dim.endpoint ASC"

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
		}
	}()

	results = make([]usageLogEndpointStatsBucketRow, 0)
	for rows.Next() {
		var row usageLogEndpointStatsBucketRow
		if err := rows.Scan(&row.Kind, &row.Stat.Endpoint, &row.Stat.Requests, &row.Stat.TotalTokens, &row.Stat.Cost, &row.Stat.ActualCost); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func groupEndpointStatsByKind(rows []usageLogEndpointStatsBucketRow, includePaths bool) usageLogEndpointStatsResult {
	result := usageLogEndpointStatsResult{
		Endpoints:         []EndpointStat{},
		UpstreamEndpoints: []EndpointStat{},
	}
	if includePaths {
		result.EndpointPaths = []EndpointStat{}
	}

	for _, row := range rows {
		switch row.Kind {
		case endpointStatKindInbound:
			result.Endpoints = append(result.Endpoints, row.Stat)
		case endpointStatKindUpstream:
			result.UpstreamEndpoints = append(result.UpstreamEndpoints, row.Stat)
		case endpointStatKindPath:
			if includePaths {
				result.EndpointPaths = append(result.EndpointPaths, row.Stat)
			}
		}
	}
	return result
}

func (r *usageLogRepository) getEndpointStatsByColumnWithFilters(ctx context.Context, endpointColumn string, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) (results []EndpointStat, err error) {
	query := fmt.Sprintf(`
		SELECT
			COALESCE(NULLIF(TRIM(%s), ''), 'unknown') AS endpoint,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) AS total_tokens,
			COALESCE(SUM(total_cost), 0) as cost,
			%s
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
	`, endpointColumn, usageLogActualCostAggregateExpr("", userID, apiKeyID, accountID))

	args := []any{startTime, endTime}
	extraConditions, args := appendUsageLogScopedConditions(nil, args, usageLogScopedFilters{
		UserID:      userID,
		APIKeyID:    apiKeyID,
		AccountID:   accountID,
		GroupID:     groupID,
		Model:       model,
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
	}, usageLogScopedFilterOptions{IncludeModel: true})
	query = appendConditionsToQuery(query, extraConditions)
	query += " GROUP BY endpoint ORDER BY requests DESC"

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
		}
	}()

	results = make([]EndpointStat, 0)
	for rows.Next() {
		var row EndpointStat
		if err := rows.Scan(&row.Endpoint, &row.Requests, &row.TotalTokens, &row.Cost, &row.ActualCost); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *usageLogRepository) getEndpointPathStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) (results []EndpointStat, err error) {
	query := fmt.Sprintf(`
		SELECT
			CONCAT(
				COALESCE(NULLIF(TRIM(inbound_endpoint), ''), 'unknown'),
				' -> ',
				COALESCE(NULLIF(TRIM(upstream_endpoint), ''), 'unknown')
			) AS endpoint,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) AS total_tokens,
			COALESCE(SUM(total_cost), 0) as cost,
			%s
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
	`, usageLogActualCostAggregateExpr("", userID, apiKeyID, accountID))

	args := []any{startTime, endTime}
	extraConditions, args := appendUsageLogScopedConditions(nil, args, usageLogScopedFilters{
		UserID:      userID,
		APIKeyID:    apiKeyID,
		AccountID:   accountID,
		GroupID:     groupID,
		Model:       model,
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
	}, usageLogScopedFilterOptions{IncludeModel: true})
	query = appendConditionsToQuery(query, extraConditions)
	query += " GROUP BY endpoint ORDER BY requests DESC"

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
		}
	}()

	results = make([]EndpointStat, 0)
	for rows.Next() {
		var row EndpointStat
		if err := rows.Scan(&row.Endpoint, &row.Requests, &row.TotalTokens, &row.Cost, &row.ActualCost); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetEndpointStatsWithFilters returns inbound endpoint statistics with optional filters.
func (r *usageLogRepository) GetEndpointStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) ([]EndpointStat, error) {
	return r.getEndpointStatsByColumnWithFilters(ctx, "inbound_endpoint", startTime, endTime, userID, apiKeyID, accountID, groupID, model, requestType, stream, billingType)
}

// GetUpstreamEndpointStatsWithFilters returns upstream endpoint statistics with optional filters.
func (r *usageLogRepository) GetUpstreamEndpointStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) ([]EndpointStat, error) {
	return r.getEndpointStatsByColumnWithFilters(ctx, "upstream_endpoint", startTime, endTime, userID, apiKeyID, accountID, groupID, model, requestType, stream, billingType)
}
