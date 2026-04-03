package ports

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
)

// IdempotencyRepository stores and updates idempotency records.
type IdempotencyRepository interface {
	CreateProcessing(ctx context.Context, record *domain.IdempotencyRecord) (bool, error)
	GetByScopeAndKeyHash(ctx context.Context, scope, keyHash string) (*domain.IdempotencyRecord, error)
	TryReclaim(ctx context.Context, id int64, fromStatus string, now, newLockedUntil, newExpiresAt time.Time) (bool, error)
	ExtendProcessingLock(ctx context.Context, id int64, requestFingerprint string, newLockedUntil, newExpiresAt time.Time) (bool, error)
	MarkSucceeded(ctx context.Context, id int64, responseStatus int, responseBody string, expiresAt time.Time) error
	MarkFailedRetryable(ctx context.Context, id int64, errorReason string, lockedUntil, expiresAt time.Time) error
	DeleteExpired(ctx context.Context, now time.Time, limit int) (int64, error)
}
