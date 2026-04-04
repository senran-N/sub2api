package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/senran-N/sub2api/internal/pkg/timezone"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"github.com/senran-N/sub2api/internal/service"
)

// GetGeminiUsageTotalsBatch 批量聚合 Gemini 账号在窗口内的 Pro/Flash 请求与用量。
// 模型分类规则与 service.geminiModelClassFromName 一致：model 包含 flash/lite 视为 flash，其余视为 pro。
func (r *usageLogRepository) GetGeminiUsageTotalsBatch(ctx context.Context, accountIDs []int64, startTime, endTime time.Time) (map[int64]service.GeminiUsageTotals, error) {
	result := make(map[int64]service.GeminiUsageTotals, len(accountIDs))
	if len(accountIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT
			account_id,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(model, '')) LIKE '%flash%' OR LOWER(COALESCE(model, '')) LIKE '%lite%' THEN 1 ELSE 0 END), 0) AS flash_requests,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(model, '')) LIKE '%flash%' OR LOWER(COALESCE(model, '')) LIKE '%lite%' THEN 0 ELSE 1 END), 0) AS pro_requests,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(model, '')) LIKE '%flash%' OR LOWER(COALESCE(model, '')) LIKE '%lite%' THEN (input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens) ELSE 0 END), 0) AS flash_tokens,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(model, '')) LIKE '%flash%' OR LOWER(COALESCE(model, '')) LIKE '%lite%' THEN 0 ELSE (input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens) END), 0) AS pro_tokens,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(model, '')) LIKE '%flash%' OR LOWER(COALESCE(model, '')) LIKE '%lite%' THEN actual_cost ELSE 0 END), 0) AS flash_cost,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(model, '')) LIKE '%flash%' OR LOWER(COALESCE(model, '')) LIKE '%lite%' THEN 0 ELSE actual_cost END), 0) AS pro_cost
		FROM usage_logs
		WHERE account_id = ANY($1) AND created_at >= $2 AND created_at < $3
		GROUP BY account_id
	`
	rows, err := r.sql.QueryContext(ctx, query, pq.Array(accountIDs), startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var accountID int64
		var totals service.GeminiUsageTotals
		if err := rows.Scan(
			&accountID,
			&totals.FlashRequests,
			&totals.ProRequests,
			&totals.FlashTokens,
			&totals.ProTokens,
			&totals.FlashCost,
			&totals.ProCost,
		); err != nil {
			return nil, err
		}
		result[accountID] = totals
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, accountID := range accountIDs {
		if _, ok := result[accountID]; !ok {
			result[accountID] = service.GeminiUsageTotals{}
		}
	}
	return result, nil
}

// BatchUserUsageStats represents usage stats for a single user
type BatchUserUsageStats = usagestats.BatchUserUsageStats

// BatchAPIKeyUsageStats represents usage stats for a single API key
type BatchAPIKeyUsageStats = usagestats.BatchAPIKeyUsageStats

func normalizePositiveInt64IDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

type batchUsageCostSummary struct {
	ID              int64
	TotalActualCost float64
	TodayActualCost float64
}

func normalizeBatchUsageRange(startTime, endTime time.Time) (time.Time, time.Time) {
	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, 0, -30)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}
	return startTime, endTime
}

func buildBatchUsageCostQuery(idColumn string) string {
	switch idColumn {
	case "user_id", "api_key_id":
	default:
		panic("unsupported batch usage id column: " + idColumn)
	}
	return fmt.Sprintf(`
		SELECT
			%s,
			COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $2 AND created_at < $3), 0) as total_cost,
			COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $4), 0) as today_cost
		FROM usage_logs
		WHERE %s = ANY($1)
		  AND created_at >= LEAST($2, $4)
		GROUP BY %s
	`, idColumn, idColumn, idColumn)
}

func (r *usageLogRepository) loadBatchUsageCostSummaries(ctx context.Context, idColumn string, ids []int64, startTime, endTime time.Time) (map[int64]batchUsageCostSummary, error) {
	result := make(map[int64]batchUsageCostSummary, len(ids))
	for _, id := range ids {
		result[id] = batchUsageCostSummary{ID: id}
	}

	startTime, endTime = normalizeBatchUsageRange(startTime, endTime)
	today := timezone.Today()
	rows, err := r.sql.QueryContext(ctx, buildBatchUsageCostQuery(idColumn), pq.Array(ids), startTime, endTime, today)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var row batchUsageCostSummary
		if err := rows.Scan(&row.ID, &row.TotalActualCost, &row.TodayActualCost); err != nil {
			return nil, err
		}
		result[row.ID] = row
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// GetBatchUserUsageStats gets today and total actual_cost for multiple users within a time range.
// If startTime is zero, defaults to 30 days ago.
func (r *usageLogRepository) GetBatchUserUsageStats(ctx context.Context, userIDs []int64, startTime, endTime time.Time) (map[int64]*BatchUserUsageStats, error) {
	result := make(map[int64]*BatchUserUsageStats)
	normalizedUserIDs := normalizePositiveInt64IDs(userIDs)
	if len(normalizedUserIDs) == 0 {
		return result, nil
	}

	summaries, err := r.loadBatchUsageCostSummaries(ctx, "user_id", normalizedUserIDs, startTime, endTime)
	if err != nil {
		return nil, err
	}
	for _, id := range normalizedUserIDs {
		summary := summaries[id]
		result[id] = &BatchUserUsageStats{
			UserID:          id,
			TotalActualCost: summary.TotalActualCost,
			TodayActualCost: summary.TodayActualCost,
		}
	}
	return result, nil
}

// GetBatchAPIKeyUsageStats gets today and total actual_cost for multiple API keys within a time range.
// If startTime is zero, defaults to 30 days ago.
func (r *usageLogRepository) GetBatchAPIKeyUsageStats(ctx context.Context, apiKeyIDs []int64, startTime, endTime time.Time) (map[int64]*BatchAPIKeyUsageStats, error) {
	result := make(map[int64]*BatchAPIKeyUsageStats)
	normalizedAPIKeyIDs := normalizePositiveInt64IDs(apiKeyIDs)
	if len(normalizedAPIKeyIDs) == 0 {
		return result, nil
	}

	summaries, err := r.loadBatchUsageCostSummaries(ctx, "api_key_id", normalizedAPIKeyIDs, startTime, endTime)
	if err != nil {
		return nil, err
	}
	for _, id := range normalizedAPIKeyIDs {
		summary := summaries[id]
		result[id] = &BatchAPIKeyUsageStats{
			APIKeyID:        id,
			TotalActualCost: summary.TotalActualCost,
			TodayActualCost: summary.TodayActualCost,
		}
	}
	return result, nil
}
