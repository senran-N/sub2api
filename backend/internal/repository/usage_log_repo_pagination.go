package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/senran-N/sub2api/internal/service"
)

func (r *usageLogRepository) listUsageLogPageIDs(ctx context.Context, whereClause string, args []any, limit, offset int) (ids []int64, err error) {
	if limit <= 0 {
		return []int64{}, nil
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	queryArgs := append(append([]any{}, args...), limit, offset)
	query := fmt.Sprintf("SELECT id FROM usage_logs %s ORDER BY id DESC LIMIT $%d OFFSET $%d", whereClause, limitPos, offsetPos)

	rows, err := r.sql.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			ids = nil
		}
	}()

	return scanInt64Rows(rows, limit)
}

func (r *usageLogRepository) queryUsageLogsByIDsOrdered(ctx context.Context, ids []int64) (logs []service.UsageLog, err error) {
	if len(ids) == 0 {
		return []service.UsageLog{}, nil
	}

	query := fmt.Sprintf(
		"SELECT %s FROM usage_logs WHERE id = ANY($1::bigint[]) ORDER BY array_position($1::bigint[], id)",
		usageLogSelectColumns,
	)
	rows, err := r.sql.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			logs = nil
		}
	}()

	logs = make([]service.UsageLog, 0, len(ids))
	for rows.Next() {
		log, scanErr := scanUsageLog(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		logs = append(logs, *log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func scanInt64Rows(rows *sql.Rows, capacity int) ([]int64, error) {
	ids := make([]int64, 0, capacity)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
