package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/usagestats"
)

// TrendDataPoint represents a single point in trend data
type TrendDataPoint = usagestats.TrendDataPoint

// ModelStat represents usage statistics for a single model
type ModelStat = usagestats.ModelStat

// UserUsageTrendPoint represents user usage trend data point
type UserUsageTrendPoint = usagestats.UserUsageTrendPoint

// UserSpendingRankingItem represents a user spending ranking row.
type UserSpendingRankingItem = usagestats.UserSpendingRankingItem
type UserSpendingRankingResponse = usagestats.UserSpendingRankingResponse

// APIKeyUsageTrendPoint represents API key usage trend data point
type APIKeyUsageTrendPoint = usagestats.APIKeyUsageTrendPoint

func buildAPIKeyUsageTrendQuery(granularity string) string {
	dateFormat := safeDateFormat(granularity)
	return fmt.Sprintf(`
		WITH scoped AS (
			SELECT
				TO_CHAR(u.created_at, '%s') as date,
				u.api_key_id,
				COALESCE(k.name, '') as key_name,
				COUNT(*) as requests,
				COALESCE(SUM(u.input_tokens + u.output_tokens + u.cache_creation_tokens + u.cache_read_tokens), 0) as tokens
			FROM usage_logs u
			LEFT JOIN api_keys k ON u.api_key_id = k.id
			WHERE u.created_at >= $1 AND u.created_at < $2
			GROUP BY date, u.api_key_id, k.name
		),
		ranked_keys AS (
			SELECT
				api_key_id,
				ROW_NUMBER() OVER (ORDER BY SUM(tokens) DESC, api_key_id ASC) AS rank
			FROM scoped
			GROUP BY api_key_id
		)
		SELECT
			scoped.date,
			scoped.api_key_id,
			scoped.key_name,
			scoped.requests,
			scoped.tokens
		FROM scoped
		INNER JOIN ranked_keys ON ranked_keys.api_key_id = scoped.api_key_id
		WHERE ranked_keys.rank <= $3
		ORDER BY scoped.date ASC, scoped.tokens DESC, scoped.api_key_id ASC
	`, dateFormat)
}

func buildUserUsageTrendQuery(granularity string) string {
	dateFormat := safeDateFormat(granularity)
	return fmt.Sprintf(`
		WITH scoped AS (
			SELECT
				TO_CHAR(u.created_at, '%s') as date,
				u.user_id,
				COALESCE(us.email, '') as email,
				COALESCE(us.username, '') as username,
				COUNT(*) as requests,
				COALESCE(SUM(u.input_tokens + u.output_tokens + u.cache_creation_tokens + u.cache_read_tokens), 0) as tokens,
				COALESCE(SUM(u.total_cost), 0) as cost,
				COALESCE(SUM(u.actual_cost), 0) as actual_cost
			FROM usage_logs u
			LEFT JOIN users us ON u.user_id = us.id
			WHERE u.created_at >= $1 AND u.created_at < $2
			GROUP BY date, u.user_id, us.email, us.username
		),
		ranked_users AS (
			SELECT
				user_id,
				ROW_NUMBER() OVER (ORDER BY SUM(tokens) DESC, user_id ASC) AS rank
			FROM scoped
			GROUP BY user_id
		)
		SELECT
			scoped.date,
			scoped.user_id,
			scoped.email,
			scoped.username,
			scoped.requests,
			scoped.tokens,
			scoped.cost,
			scoped.actual_cost
		FROM scoped
		INNER JOIN ranked_users ON ranked_users.user_id = scoped.user_id
		WHERE ranked_users.rank <= $3
		ORDER BY scoped.date ASC, scoped.tokens DESC, scoped.user_id ASC
	`, dateFormat)
}

func scanAPIKeyUsageTrendRows(rows *sql.Rows) ([]APIKeyUsageTrendPoint, error) {
	results := make([]APIKeyUsageTrendPoint, 0)
	for rows.Next() {
		var row APIKeyUsageTrendPoint
		if err := rows.Scan(&row.Date, &row.APIKeyID, &row.KeyName, &row.Requests, &row.Tokens); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func scanUserUsageTrendRows(rows *sql.Rows) ([]UserUsageTrendPoint, error) {
	results := make([]UserUsageTrendPoint, 0)
	for rows.Next() {
		var row UserUsageTrendPoint
		if err := rows.Scan(&row.Date, &row.UserID, &row.Email, &row.Username, &row.Requests, &row.Tokens, &row.Cost, &row.ActualCost); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetAPIKeyUsageTrend returns usage trend data grouped by API key and date
func (r *usageLogRepository) GetAPIKeyUsageTrend(ctx context.Context, startTime, endTime time.Time, granularity string, limit int) (results []APIKeyUsageTrendPoint, err error) {
	rows, err := r.sql.QueryContext(ctx, buildAPIKeyUsageTrendQuery(granularity), startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		// 保持主错误优先；仅在无错误时回传 Close 失败。
		// 同时清空返回值，避免误用不完整结果。
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
		}
	}()
	return scanAPIKeyUsageTrendRows(rows)
}

// GetUserUsageTrend returns usage trend data grouped by user and date
func (r *usageLogRepository) GetUserUsageTrend(ctx context.Context, startTime, endTime time.Time, granularity string, limit int) (results []UserUsageTrendPoint, err error) {
	rows, err := r.sql.QueryContext(ctx, buildUserUsageTrendQuery(granularity), startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		// 保持主错误优先；仅在无错误时回传 Close 失败。
		// 同时清空返回值，避免误用不完整结果。
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			results = nil
		}
	}()
	return scanUserUsageTrendRows(rows)
}

// GetUserSpendingRanking returns user spending ranking aggregated within the time range.
func (r *usageLogRepository) GetUserSpendingRanking(ctx context.Context, startTime, endTime time.Time, limit int) (result *UserSpendingRankingResponse, err error) {
	if limit <= 0 {
		limit = 12
	}

	query := `
		WITH user_spend AS (
			SELECT
				u.user_id,
				COALESCE(us.email, '') as email,
				COALESCE(SUM(u.actual_cost), 0) as actual_cost,
				COUNT(*) as requests,
				COALESCE(SUM(u.input_tokens + u.output_tokens + u.cache_creation_tokens + u.cache_read_tokens), 0) as tokens
			FROM usage_logs u
			LEFT JOIN users us ON u.user_id = us.id
			WHERE u.created_at >= $1 AND u.created_at < $2
			GROUP BY u.user_id, us.email
		),
		ranked AS (
			SELECT
				user_id,
				email,
				actual_cost,
				requests,
				tokens,
				COALESCE(SUM(actual_cost) OVER (), 0) as total_actual_cost,
				COALESCE(SUM(requests) OVER (), 0) as total_requests,
				COALESCE(SUM(tokens) OVER (), 0) as total_tokens
			FROM user_spend
			ORDER BY actual_cost DESC, tokens DESC, user_id ASC
			LIMIT $3
		)
		SELECT
			user_id,
			email,
			actual_cost,
			requests,
			tokens,
			total_actual_cost,
			total_requests,
			total_tokens
		FROM ranked
		ORDER BY actual_cost DESC, tokens DESC, user_id ASC
	`

	rows, err := r.sql.QueryContext(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			result = nil
		}
	}()

	ranking := make([]UserSpendingRankingItem, 0)
	totalActualCost := 0.0
	totalRequests := int64(0)
	totalTokens := int64(0)
	for rows.Next() {
		var row UserSpendingRankingItem
		if err = rows.Scan(&row.UserID, &row.Email, &row.ActualCost, &row.Requests, &row.Tokens, &totalActualCost, &totalRequests, &totalTokens); err != nil {
			return nil, err
		}
		ranking = append(ranking, row)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &UserSpendingRankingResponse{
		Ranking:         ranking,
		TotalActualCost: totalActualCost,
		TotalRequests:   totalRequests,
		TotalTokens:     totalTokens,
	}, nil
}

// GetUserUsageTrendByUserID 获取指定用户的使用趋势
func (r *usageLogRepository) GetUserUsageTrendByUserID(ctx context.Context, userID int64, startTime, endTime time.Time, granularity string) (results []TrendDataPoint, err error) {
	return r.queryUsageTrendStats(ctx, startTime, endTime, granularity, usageLogScopedFilters{UserID: userID})
}

// GetUsageTrendWithFilters returns usage trend data with optional filters
func (r *usageLogRepository) GetUsageTrendWithFilters(ctx context.Context, startTime, endTime time.Time, granularity string, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) (results []TrendDataPoint, err error) {
	if shouldUsePreaggregatedTrend(granularity, userID, apiKeyID, accountID, groupID, model, requestType, stream, billingType) {
		aggregated, aggregatedErr := r.getUsageTrendFromAggregates(ctx, startTime, endTime, granularity)
		if aggregatedErr == nil && len(aggregated) > 0 {
			return aggregated, nil
		}
	}

	return r.queryUsageTrendStats(ctx, startTime, endTime, granularity, usageLogScopedFilters{
		UserID:      userID,
		APIKeyID:    apiKeyID,
		AccountID:   accountID,
		GroupID:     groupID,
		Model:       model,
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
	})
}

func shouldUsePreaggregatedTrend(granularity string, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) bool {
	if granularity != "day" && granularity != "hour" {
		return false
	}
	return userID == 0 &&
		apiKeyID == 0 &&
		accountID == 0 &&
		groupID == 0 &&
		model == "" &&
		requestType == nil &&
		stream == nil &&
		billingType == nil
}

func buildUsageTrendStatsQuery(startTime, endTime time.Time, granularity string, scopedFilters usageLogScopedFilters) (string, []any) {
	dateFormat := safeDateFormat(granularity)
	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(created_at, '%s') as date,
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) as cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as total_tokens,
			COALESCE(SUM(total_cost), 0) as cost,
			COALESCE(SUM(actual_cost), 0) as actual_cost
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
	`, dateFormat)

	args := []any{startTime, endTime}
	extraConditions, args := appendUsageLogScopedConditions(nil, args, scopedFilters, usageLogScopedFilterOptions{IncludeModel: true})
	query = appendConditionsToQuery(query, extraConditions)
	query += " GROUP BY date ORDER BY date ASC"
	return query, args
}

func (r *usageLogRepository) queryUsageTrendStats(ctx context.Context, startTime, endTime time.Time, granularity string, scopedFilters usageLogScopedFilters) (results []TrendDataPoint, err error) {
	query, args := buildUsageTrendStatsQuery(startTime, endTime, granularity, scopedFilters)
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

	return scanTrendRows(rows)
}

func (r *usageLogRepository) getUsageTrendFromAggregates(ctx context.Context, startTime, endTime time.Time, granularity string) (results []TrendDataPoint, err error) {
	dateFormat := safeDateFormat(granularity)
	query := ""
	args := []any{startTime, endTime}

	switch granularity {
	case "hour":
		query = fmt.Sprintf(`
			SELECT
				TO_CHAR(bucket_start, '%s') as date,
				total_requests as requests,
				input_tokens,
				output_tokens,
				cache_creation_tokens,
				cache_read_tokens,
				(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens) as total_tokens,
				total_cost as cost,
				actual_cost
			FROM usage_dashboard_hourly
			WHERE bucket_start >= $1 AND bucket_start < $2
			ORDER BY bucket_start ASC
		`, dateFormat)
	case "day":
		query = fmt.Sprintf(`
			SELECT
				TO_CHAR(bucket_date::timestamp, '%s') as date,
				total_requests as requests,
				input_tokens,
				output_tokens,
				cache_creation_tokens,
				cache_read_tokens,
				(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens) as total_tokens,
				total_cost as cost,
				actual_cost
			FROM usage_dashboard_daily
			WHERE bucket_date >= $1::date AND bucket_date < $2::date
			ORDER BY bucket_date ASC
		`, dateFormat)
	default:
		return nil, nil
	}

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

	return scanTrendRows(rows)
}
