package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/usagestats"
)

// GetUserModelStats 获取指定用户的模型统计
func (r *usageLogRepository) GetUserModelStats(ctx context.Context, userID int64, startTime, endTime time.Time) (results []ModelStat, err error) {
	return r.queryUsageModelStats(ctx, startTime, endTime, rawUsageLogModelColumn, usageLogActualCostAggregateExpr("", userID, 0, 0), usageLogScopedFilters{UserID: userID}, usageLogScopedFilterOptions{})
}

func buildUsageModelStatsQuery(startTime, endTime time.Time, modelExpr, actualCostExpr string, scopedFilters usageLogScopedFilters, options usageLogScopedFilterOptions) (string, []any) {
	query := fmt.Sprintf(`
		SELECT
			%s as model,
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) as cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as total_tokens,
			COALESCE(SUM(total_cost), 0) as cost,
			%s
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
	`, modelExpr, actualCostExpr)

	args := []any{startTime, endTime}
	extraConditions, args := appendUsageLogScopedConditions(nil, args, scopedFilters, options)
	query = appendConditionsToQuery(query, extraConditions)
	query += fmt.Sprintf(" GROUP BY %s ORDER BY total_tokens DESC", modelExpr)
	return query, args
}

func (r *usageLogRepository) queryUsageModelStats(ctx context.Context, startTime, endTime time.Time, modelExpr, actualCostExpr string, scopedFilters usageLogScopedFilters, options usageLogScopedFilterOptions) (results []ModelStat, err error) {
	query, args := buildUsageModelStatsQuery(startTime, endTime, modelExpr, actualCostExpr, scopedFilters, options)
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

	return scanModelStatsRows(rows)
}

func buildUsageGroupStatsQuery(startTime, endTime time.Time, scopedFilters usageLogScopedFilters, options usageLogScopedFilterOptions) (string, []any) {
	query := `
		SELECT
			COALESCE(ul.group_id, 0) as group_id,
			COALESCE(g.name, '') as group_name,
			COUNT(*) as requests,
			COALESCE(SUM(ul.input_tokens + ul.output_tokens + ul.cache_creation_tokens + ul.cache_read_tokens), 0) as total_tokens,
			COALESCE(SUM(ul.total_cost), 0) as cost,
			COALESCE(SUM(ul.actual_cost), 0) as actual_cost
		FROM usage_logs ul
		LEFT JOIN groups g ON g.id = ul.group_id
		WHERE ul.created_at >= $1 AND ul.created_at < $2
	`

	args := []any{startTime, endTime}
	extraConditions, args := appendUsageLogScopedConditions(nil, args, scopedFilters, options)
	query = appendConditionsToQuery(query, extraConditions)
	query += " GROUP BY ul.group_id, g.name ORDER BY total_tokens DESC"
	return query, args
}

// GetModelStatsWithFilters returns model statistics with optional filters
func (r *usageLogRepository) GetModelStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8) (results []ModelStat, err error) {
	return r.getModelStatsWithFiltersBySource(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType, usagestats.ModelSourceRequested)
}

// GetModelStatsWithFiltersBySource returns model statistics with optional filters and model source dimension.
// source: requested | upstream | mapping.
func (r *usageLogRepository) GetModelStatsWithFiltersBySource(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8, source string) (results []ModelStat, err error) {
	return r.getModelStatsWithFiltersBySource(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType, source)
}

func (r *usageLogRepository) getModelStatsWithFiltersBySource(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8, source string) (results []ModelStat, err error) {
	return r.queryUsageModelStats(
		ctx,
		startTime,
		endTime,
		resolveModelDimensionExpression(source),
		usageLogActualCostAggregateExpr("", userID, apiKeyID, accountID),
		usageLogScopedFilters{
			UserID:      userID,
			APIKeyID:    apiKeyID,
			AccountID:   accountID,
			GroupID:     groupID,
			RequestType: requestType,
			Stream:      stream,
			BillingType: billingType,
		},
		usageLogScopedFilterOptions{},
	)
}

// GetGroupStatsWithFilters returns group usage statistics with optional filters
func (r *usageLogRepository) GetGroupStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8) (results []usagestats.GroupStat, err error) {
	query, args := buildUsageGroupStatsQuery(startTime, endTime, usageLogScopedFilters{
		UserID:      userID,
		APIKeyID:    apiKeyID,
		AccountID:   accountID,
		GroupID:     groupID,
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
	}, usageLogScopedFilterOptions{TableAlias: "ul"})

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

	return scanGroupStatsRows(rows)
}

func appendUserBreakdownDimensionConditions(query string, args []any, dim usagestats.UserBreakdownDimension) (string, []any) {
	if dim.GroupID > 0 {
		query += fmt.Sprintf(" AND ul.group_id = $%d", len(args)+1)
		args = append(args, dim.GroupID)
	}
	if dim.Model != "" {
		query += fmt.Sprintf(" AND %s = $%d", resolveModelDimensionExpression(dim.ModelType), len(args)+1)
		args = append(args, dim.Model)
	}
	if dim.Endpoint != "" {
		query += fmt.Sprintf(" AND %s = $%d", resolveEndpointColumn(dim.EndpointType), len(args)+1)
		args = append(args, dim.Endpoint)
	}
	return query, args
}

func buildUserBreakdownStatsQuery(startTime, endTime time.Time, dim usagestats.UserBreakdownDimension, limit int) (string, []any) {
	query := `
		SELECT
			COALESCE(ul.user_id, 0) as user_id,
			COALESCE(u.email, '') as email,
			COUNT(*) as requests,
			COALESCE(SUM(ul.input_tokens + ul.output_tokens + ul.cache_creation_tokens + ul.cache_read_tokens), 0) as total_tokens,
			COALESCE(SUM(ul.total_cost), 0) as cost,
			COALESCE(SUM(ul.actual_cost), 0) as actual_cost
		FROM usage_logs ul
		LEFT JOIN users u ON u.id = ul.user_id
		WHERE ul.created_at >= $1 AND ul.created_at < $2
	`
	args := []any{startTime, endTime}
	query, args = appendUserBreakdownDimensionConditions(query, args, dim)
	query += " GROUP BY ul.user_id, u.email ORDER BY actual_cost DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return query, args
}

// GetUserBreakdownStats returns per-user usage breakdown within a specific dimension.
func (r *usageLogRepository) GetUserBreakdownStats(ctx context.Context, startTime, endTime time.Time, dim usagestats.UserBreakdownDimension, limit int) (results []usagestats.UserBreakdownItem, err error) {
	query, args := buildUserBreakdownStatsQuery(startTime, endTime, dim, limit)
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
	return scanUserBreakdownRows(rows)
}

// GetAllGroupUsageSummary returns today's and cumulative actual_cost for every group.
func (r *usageLogRepository) GetAllGroupUsageSummary(ctx context.Context, todayStart time.Time) ([]usagestats.GroupUsageSummary, error) {
	query := `
		SELECT
			g.id AS group_id,
			COALESCE(summary.total_cost, 0) AS total_cost,
			COALESCE(summary.today_cost, 0) AS today_cost
		FROM groups g
		LEFT JOIN (
			SELECT
				group_id,
				COALESCE(SUM(actual_cost), 0) AS total_cost,
				COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $1), 0) AS today_cost
			FROM usage_logs
			WHERE group_id IS NOT NULL
			GROUP BY group_id
		) summary ON summary.group_id = g.id
		ORDER BY g.id ASC
	`

	rows, err := r.sql.QueryContext(ctx, query, todayStart)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var results []usagestats.GroupUsageSummary
	for rows.Next() {
		var row usagestats.GroupUsageSummary
		if err := rows.Scan(&row.GroupID, &row.TotalCost, &row.TodayCost); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func scanModelStatsRows(rows *sql.Rows) ([]ModelStat, error) {
	results := make([]ModelStat, 0)
	for rows.Next() {
		var row ModelStat
		if err := rows.Scan(
			&row.Model,
			&row.Requests,
			&row.InputTokens,
			&row.OutputTokens,
			&row.CacheCreationTokens,
			&row.CacheReadTokens,
			&row.TotalTokens,
			&row.Cost,
			&row.ActualCost,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func scanGroupStatsRows(rows *sql.Rows) ([]usagestats.GroupStat, error) {
	results := make([]usagestats.GroupStat, 0)
	for rows.Next() {
		var row usagestats.GroupStat
		if err := rows.Scan(&row.GroupID, &row.GroupName, &row.Requests, &row.TotalTokens, &row.Cost, &row.ActualCost); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func scanUserBreakdownRows(rows *sql.Rows) ([]usagestats.UserBreakdownItem, error) {
	results := make([]usagestats.UserBreakdownItem, 0)
	for rows.Next() {
		var row usagestats.UserBreakdownItem
		if err := rows.Scan(&row.UserID, &row.Email, &row.Requests, &row.TotalTokens, &row.Cost, &row.ActualCost); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
