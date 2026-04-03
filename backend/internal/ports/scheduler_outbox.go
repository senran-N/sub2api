package ports

import (
	"context"

	"github.com/senran-N/sub2api/internal/domain"
)

// SchedulerOutboxRepository loads scheduler outbox events in ascending order.
type SchedulerOutboxRepository interface {
	ListAfter(ctx context.Context, afterID int64, limit int) ([]domain.SchedulerOutboxEvent, error)
	MaxID(ctx context.Context) (int64, error)
}
