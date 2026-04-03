package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/ports"
)

type schedulerOutboxRepository struct {
	db *sql.DB
}

const schedulerOutboxDedupWindow = time.Second

func NewSchedulerOutboxRepository(db *sql.DB) ports.SchedulerOutboxRepository {
	return &schedulerOutboxRepository{db: db}
}

func (r *schedulerOutboxRepository) ListAfter(ctx context.Context, afterID int64, limit int) ([]domain.SchedulerOutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_type, account_id, group_id, payload, created_at
		FROM scheduler_outbox
		WHERE id > $1
		ORDER BY id ASC
		LIMIT $2
	`, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	events := make([]domain.SchedulerOutboxEvent, 0, limit)
	for rows.Next() {
		var (
			payloadRaw []byte
			accountID  sql.NullInt64
			groupID    sql.NullInt64
			event      domain.SchedulerOutboxEvent
		)
		if err := rows.Scan(&event.ID, &event.EventType, &accountID, &groupID, &payloadRaw, &event.CreatedAt); err != nil {
			return nil, err
		}
		if accountID.Valid {
			v := accountID.Int64
			event.AccountID = &v
		}
		if groupID.Valid {
			v := groupID.Int64
			event.GroupID = &v
		}
		if len(payloadRaw) > 0 {
			var payload map[string]any
			if err := json.Unmarshal(payloadRaw, &payload); err != nil {
				return nil, err
			}
			event.Payload = payload
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *schedulerOutboxRepository) MaxID(ctx context.Context) (int64, error) {
	var maxID int64
	if err := r.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(id), 0) FROM scheduler_outbox").Scan(&maxID); err != nil {
		return 0, err
	}
	return maxID, nil
}

func enqueueSchedulerOutbox(ctx context.Context, exec sqlExecutor, eventType string, accountID *int64, groupID *int64, payload any) error {
	if exec == nil {
		return nil
	}
	var payloadArg any
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		payloadArg = encoded
	}
	query := `
		INSERT INTO scheduler_outbox (event_type, account_id, group_id, payload)
		VALUES ($1, $2, $3, $4)
	`
	args := []any{eventType, accountID, groupID, payloadArg}
	if schedulerOutboxEventSupportsDedup(eventType) {
		query = `
			INSERT INTO scheduler_outbox (event_type, account_id, group_id, payload)
			SELECT $1, $2, $3, $4
			WHERE NOT EXISTS (
				SELECT 1
				FROM scheduler_outbox
				WHERE event_type = $1
					AND account_id IS NOT DISTINCT FROM $2
					AND group_id IS NOT DISTINCT FROM $3
					AND created_at >= NOW() - make_interval(secs => $5)
			)
		`
		args = append(args, schedulerOutboxDedupWindow.Seconds())
	}
	_, err := exec.ExecContext(ctx, query, args...)
	return err
}

func schedulerOutboxEventSupportsDedup(eventType string) bool {
	switch eventType {
	case domain.SchedulerOutboxEventAccountChanged,
		domain.SchedulerOutboxEventGroupChanged,
		domain.SchedulerOutboxEventFullRebuild:
		return true
	default:
		return false
	}
}
