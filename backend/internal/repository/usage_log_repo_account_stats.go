package repository

import (
	"context"
	"time"

	"github.com/lib/pq"
	"github.com/senran-N/sub2api/internal/pkg/timezone"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
)

// GetAccountTodayStats 获取账号今日统计
func (r *usageLogRepository) GetAccountTodayStats(ctx context.Context, accountID int64) (*usagestats.AccountStats, error) {
	today := timezone.Today()

	query := `
		SELECT
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as tokens,
			COALESCE(SUM(total_cost * COALESCE(account_rate_multiplier, 1)), 0) as cost,
			COALESCE(SUM(total_cost), 0) as standard_cost,
			COALESCE(SUM(actual_cost), 0) as user_cost
		FROM usage_logs
		WHERE account_id = $1 AND created_at >= $2
	`

	stats := &usagestats.AccountStats{}
	if err := scanSingleRow(
		ctx,
		r.sql,
		query,
		[]any{accountID, today},
		&stats.Requests,
		&stats.Tokens,
		&stats.Cost,
		&stats.StandardCost,
		&stats.UserCost,
	); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetAccountWindowStats 获取账号时间窗口内的统计
func (r *usageLogRepository) GetAccountWindowStats(ctx context.Context, accountID int64, startTime time.Time) (*usagestats.AccountStats, error) {
	query := `
		SELECT
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as tokens,
			COALESCE(SUM(total_cost * COALESCE(account_rate_multiplier, 1)), 0) as cost,
			COALESCE(SUM(total_cost), 0) as standard_cost,
			COALESCE(SUM(actual_cost), 0) as user_cost
		FROM usage_logs
		WHERE account_id = $1 AND created_at >= $2
	`

	stats := &usagestats.AccountStats{}
	if err := scanSingleRow(
		ctx,
		r.sql,
		query,
		[]any{accountID, startTime},
		&stats.Requests,
		&stats.Tokens,
		&stats.Cost,
		&stats.StandardCost,
		&stats.UserCost,
	); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetAccountWindowStatsBatch 批量获取同一窗口起点下多个账号的统计数据。
// 返回 map[accountID]*AccountStats，未命中的账号会返回零值统计，便于上层直接复用。
func (r *usageLogRepository) GetAccountWindowStatsBatch(ctx context.Context, accountIDs []int64, startTime time.Time) (map[int64]*usagestats.AccountStats, error) {
	result := make(map[int64]*usagestats.AccountStats, len(accountIDs))
	if len(accountIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT
			account_id,
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as tokens,
			COALESCE(SUM(total_cost * COALESCE(account_rate_multiplier, 1)), 0) as cost,
			COALESCE(SUM(total_cost), 0) as standard_cost,
			COALESCE(SUM(actual_cost), 0) as user_cost
		FROM usage_logs
		WHERE account_id = ANY($1) AND created_at >= $2
		GROUP BY account_id
	`
	rows, err := r.sql.QueryContext(ctx, query, pq.Array(accountIDs), startTime)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var accountID int64
		stats := &usagestats.AccountStats{}
		if err := rows.Scan(
			&accountID,
			&stats.Requests,
			&stats.Tokens,
			&stats.Cost,
			&stats.StandardCost,
			&stats.UserCost,
		); err != nil {
			return nil, err
		}
		result[accountID] = stats
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, accountID := range accountIDs {
		if _, ok := result[accountID]; !ok {
			result[accountID] = &usagestats.AccountStats{}
		}
	}
	return result, nil
}
