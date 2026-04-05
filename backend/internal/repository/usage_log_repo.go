package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	dbent "github.com/senran-N/sub2api/ent"
	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/senran-N/sub2api/internal/service"
)

const usageLogSelectColumns = "id, user_id, api_key_id, account_id, request_id, model, requested_model, upstream_model, group_id, subscription_id, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, cache_creation_5m_tokens, cache_creation_1h_tokens, image_output_tokens, image_output_cost, input_cost, output_cost, cache_creation_cost, cache_read_cost, total_cost, actual_cost, rate_multiplier, account_rate_multiplier, billing_type, request_type, stream, openai_ws_mode, duration_ms, first_token_ms, user_agent, ip_address, image_count, image_size, media_type, service_tier, reasoning_effort, inbound_endpoint, upstream_endpoint, cache_ttl_overridden, channel_id, model_mapping_chain, billing_tier, billing_mode, created_at"

// usageLogInsertArgTypes must stay in the same order as:
//  1. prepareUsageLogInsert().args
//  2. every INSERT/CTE VALUES column list in this file
//  3. execUsageLogInsertNoResult placeholder positions
//  4. scanUsageLog selected column order (via usageLogSelectColumns)
//
// When adding a usage_logs column, update all of those call sites together.
var usageLogInsertArgTypes = [...]string{
	"bigint",      // user_id
	"bigint",      // api_key_id
	"bigint",      // account_id
	"text",        // request_id
	"text",        // model
	"text",        // requested_model
	"text",        // upstream_model
	"bigint",      // group_id
	"bigint",      // subscription_id
	"integer",     // input_tokens
	"integer",     // output_tokens
	"integer",     // cache_creation_tokens
	"integer",     // cache_read_tokens
	"integer",     // cache_creation_5m_tokens
	"integer",     // cache_creation_1h_tokens
	"integer",     // image_output_tokens
	"numeric",     // image_output_cost
	"numeric",     // input_cost
	"numeric",     // output_cost
	"numeric",     // cache_creation_cost
	"numeric",     // cache_read_cost
	"numeric",     // total_cost
	"numeric",     // actual_cost
	"numeric",     // rate_multiplier
	"numeric",     // account_rate_multiplier
	"smallint",    // billing_type
	"smallint",    // request_type
	"boolean",     // stream
	"boolean",     // openai_ws_mode
	"integer",     // duration_ms
	"integer",     // first_token_ms
	"text",        // user_agent
	"text",        // ip_address
	"integer",     // image_count
	"text",        // image_size
	"text",        // media_type
	"text",        // service_tier
	"text",        // reasoning_effort
	"text",        // inbound_endpoint
	"text",        // upstream_endpoint
	"boolean",     // cache_ttl_overridden
	"bigint",      // channel_id
	"text",        // model_mapping_chain
	"text",        // billing_tier
	"text",        // billing_mode
	"timestamptz", // created_at
}

const rawUsageLogModelColumn = "model"

// rawUsageLogModelColumn preserves the exact stored usage_logs.model semantics for direct filters.
// Historical rows may contain upstream/billing model values, while newer rows store requested_model.
// Requested/upstream/mapping analytics must use resolveModelDimensionExpression instead.

// dateFormatWhitelist 将 granularity 参数映射为 PostgreSQL TO_CHAR 格式字符串，防止外部输入直接拼入 SQL
var dateFormatWhitelist = map[string]string{
	"hour":  "YYYY-MM-DD HH24:00",
	"day":   "YYYY-MM-DD",
	"week":  "IYYY-IW",
	"month": "YYYY-MM",
}

// safeDateFormat 根据白名单获取 dateFormat，未匹配时返回默认值
func safeDateFormat(granularity string) string {
	if f, ok := dateFormatWhitelist[granularity]; ok {
		return f
	}
	return "YYYY-MM-DD"
}

// appendRawUsageLogModelWhereCondition keeps direct model filters on the raw model column for backward
// compatibility with historical rows. Requested/upstream analytics must use
// resolveModelDimensionExpression instead.
func appendRawUsageLogModelWhereCondition(conditions []string, args []any, model string) ([]string, []any) {
	return appendRawUsageLogModelWhereConditionWithColumn(conditions, args, rawUsageLogModelColumn, model)
}

func appendRawUsageLogModelWhereConditionWithColumn(conditions []string, args []any, column string, model string) ([]string, []any) {
	if strings.TrimSpace(model) == "" {
		return conditions, args
	}
	conditions = append(conditions, fmt.Sprintf("%s = $%d", column, len(args)+1))
	args = append(args, model)
	return conditions, args
}

// appendRawUsageLogModelQueryFilter keeps direct model filters on the raw model column for backward
// compatibility with historical rows. Requested/upstream analytics must use
// resolveModelDimensionExpression instead.
func appendRawUsageLogModelQueryFilter(query string, args []any, model string) (string, []any) {
	if strings.TrimSpace(model) == "" {
		return query, args
	}
	query += fmt.Sprintf(" AND %s = $%d", rawUsageLogModelColumn, len(args)+1)
	args = append(args, model)
	return query, args
}

type usageLogRepository struct {
	client *dbent.Client
	sql    sqlExecutor
	db     *sql.DB

	createBatchOnce     sync.Once
	createBatchCh       chan usageLogCreateRequest
	bestEffortBatchOnce sync.Once
	bestEffortBatchCh   chan usageLogBestEffortRequest
	bestEffortRecent    *gocache.Cache
}

func NewUsageLogRepository(client *dbent.Client, sqlDB *sql.DB) service.UsageLogRepository {
	return newUsageLogRepositoryWithSQL(client, sqlDB)
}

func newUsageLogRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *usageLogRepository {
	// 使用 scanSingleRow 替代 QueryRowContext，保证 ent.Tx 作为 sqlExecutor 可用。
	repo := &usageLogRepository{client: client, sql: sqlq}
	if db, ok := sqlq.(*sql.DB); ok {
		repo.db = db
	}
	repo.bestEffortRecent = gocache.New(usageLogBestEffortRecentTTL, time.Minute)
	return repo
}

// getPerformanceStats 获取 RPM 和 TPM（近5分钟平均值，可选按用户过滤）
func (r *usageLogRepository) getPerformanceStats(ctx context.Context, userID int64) (rpm, tpm int64, err error) {
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	query := `
		SELECT
			COUNT(*) as request_count,
			COALESCE(SUM(input_tokens + output_tokens), 0) as token_count
		FROM usage_logs
		WHERE created_at >= $1`
	args := []any{fiveMinutesAgo}
	if userID > 0 {
		query += " AND user_id = $2"
		args = append(args, userID)
	}

	var requestCount int64
	var tokenCount int64
	if err := scanSingleRow(ctx, r.sql, query, args, &requestCount, &tokenCount); err != nil {
		return 0, 0, err
	}
	return requestCount / 5, tokenCount / 5, nil
}

func (r *usageLogRepository) listUsageLogsWithPagination(ctx context.Context, whereClause string, args []any, params pagination.PaginationParams) ([]service.UsageLog, *pagination.PaginationResult, error) {
	countQuery := "SELECT COUNT(*) FROM usage_logs " + whereClause
	var total int64
	if err := scanSingleRow(ctx, r.sql, countQuery, args, &total); err != nil {
		return nil, nil, err
	}

	pageIDs, err := r.listUsageLogPageIDs(ctx, whereClause, args, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	logs, err := r.queryUsageLogsByIDsOrdered(ctx, pageIDs)
	if err != nil {
		return nil, nil, err
	}
	return logs, paginationResultFromTotal(total, params), nil
}

func (r *usageLogRepository) listUsageLogsWithFastPagination(ctx context.Context, whereClause string, args []any, params pagination.PaginationParams) ([]service.UsageLog, *pagination.PaginationResult, error) {
	limit := params.Limit()
	offset := params.Offset()

	pageIDs, err := r.listUsageLogPageIDs(ctx, whereClause, args, limit+1, offset)
	if err != nil {
		return nil, nil, err
	}

	hasMore := false
	if len(pageIDs) > limit {
		hasMore = true
		pageIDs = pageIDs[:limit]
	}

	logs, err := r.queryUsageLogsByIDsOrdered(ctx, pageIDs)
	if err != nil {
		return nil, nil, err
	}

	total := int64(offset) + int64(len(logs))
	if hasMore {
		// 只保证“还有下一页”，避免对超大表做全量 COUNT(*)。
		total = int64(offset) + int64(limit) + 1
	}

	return logs, paginationResultFromTotal(total, params), nil
}

func scanTrendRows(rows *sql.Rows) ([]TrendDataPoint, error) {
	results := make([]TrendDataPoint, 0)
	for rows.Next() {
		var row TrendDataPoint
		if err := rows.Scan(
			&row.Date,
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

func buildWhere(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

func appendRequestTypeOrStreamWhereCondition(conditions []string, args []any, requestType *int16, stream *bool) ([]string, []any) {
	return appendRequestTypeOrStreamWhereConditionWithColumns(conditions, args, "request_type", "stream", "openai_ws_mode", requestType, stream)
}

func appendRequestTypeOrStreamWhereConditionWithColumns(conditions []string, args []any, requestTypeColumn, streamColumn, openAIWSModeColumn string, requestType *int16, stream *bool) ([]string, []any) {
	if requestType != nil {
		condition, conditionArgs := buildRequestTypeFilterConditionWithColumns(len(args)+1, requestTypeColumn, streamColumn, openAIWSModeColumn, *requestType)
		conditions = append(conditions, condition)
		args = append(args, conditionArgs...)
		return conditions, args
	}
	if stream != nil {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", streamColumn, len(args)+1))
		args = append(args, *stream)
	}
	return conditions, args
}

func appendRequestTypeOrStreamQueryFilter(query string, args []any, requestType *int16, stream *bool) (string, []any) {
	if requestType != nil {
		condition, conditionArgs := buildRequestTypeFilterCondition(len(args)+1, *requestType)
		query += " AND " + condition
		args = append(args, conditionArgs...)
		return query, args
	}
	if stream != nil {
		query += fmt.Sprintf(" AND stream = $%d", len(args)+1)
		args = append(args, *stream)
	}
	return query, args
}

// buildRequestTypeFilterCondition 在 request_type 过滤时兼容 legacy 字段，避免历史数据漏查。
func buildRequestTypeFilterCondition(startArgIndex int, requestType int16) (string, []any) {
	return buildRequestTypeFilterConditionWithColumns(startArgIndex, "request_type", "stream", "openai_ws_mode", requestType)
}

func buildRequestTypeFilterConditionWithColumns(startArgIndex int, requestTypeColumn, streamColumn, openAIWSModeColumn string, requestType int16) (string, []any) {
	normalized := service.RequestTypeFromInt16(requestType)
	requestTypeArg := int16(normalized)
	switch normalized {
	case service.RequestTypeSync:
		return fmt.Sprintf("(%s = $%d OR (%s = %d AND %s = FALSE AND %s = FALSE))", requestTypeColumn, startArgIndex, requestTypeColumn, int16(service.RequestTypeUnknown), streamColumn, openAIWSModeColumn), []any{requestTypeArg}
	case service.RequestTypeStream:
		return fmt.Sprintf("(%s = $%d OR (%s = %d AND %s = TRUE AND %s = FALSE))", requestTypeColumn, startArgIndex, requestTypeColumn, int16(service.RequestTypeUnknown), streamColumn, openAIWSModeColumn), []any{requestTypeArg}
	case service.RequestTypeWSV2:
		return fmt.Sprintf("(%s = $%d OR (%s = %d AND %s = TRUE))", requestTypeColumn, startArgIndex, requestTypeColumn, int16(service.RequestTypeUnknown), openAIWSModeColumn), []any{requestTypeArg}
	default:
		return fmt.Sprintf("%s = $%d", requestTypeColumn, startArgIndex), []any{requestTypeArg}
	}
}

func nullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullFloat64Ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	out := v.Float64
	return &out
}

func nullString(v *string) sql.NullString {
	if v == nil || *v == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func coalesceTrimmedString(v sql.NullString, fallback string) string {
	if v.Valid && strings.TrimSpace(v.String) != "" {
		return v.String
	}
	return fallback
}
