package ports

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
)

// TempUnschedCache stores temporary unschedulable state for accounts.
type TempUnschedCache interface {
	SetTempUnsched(ctx context.Context, accountID int64, state *domain.TempUnschedState) error
	GetTempUnsched(ctx context.Context, accountID int64) (*domain.TempUnschedState, error)
	DeleteTempUnsched(ctx context.Context, accountID int64) error
}

// TimeoutCounterCache tracks timeout counters within a rolling TTL window.
type TimeoutCounterCache interface {
	IncrementTimeoutCount(ctx context.Context, accountID int64, windowMinutes int) (int64, error)
	GetTimeoutCount(ctx context.Context, accountID int64) (int64, error)
	ResetTimeoutCount(ctx context.Context, accountID int64) error
	GetTimeoutCountTTL(ctx context.Context, accountID int64) (time.Duration, error)
}
