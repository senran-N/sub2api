package ports

import (
	"context"

	"github.com/senran-N/sub2api/internal/domain"
)

// ConcurrencyCache stores runtime slot and queue state used by schedulers.
type ConcurrencyCache interface {
	AcquireAccountSlot(ctx context.Context, accountID int64, maxConcurrency int, requestID string) (bool, error)
	ReleaseAccountSlot(ctx context.Context, accountID int64, requestID string) error
	GetAccountConcurrency(ctx context.Context, accountID int64) (int, error)
	GetAccountConcurrencyBatch(ctx context.Context, accountIDs []int64) (map[int64]int, error)

	IncrementAccountWaitCount(ctx context.Context, accountID int64, maxWait int) (bool, error)
	DecrementAccountWaitCount(ctx context.Context, accountID int64) error
	GetAccountWaitingCount(ctx context.Context, accountID int64) (int, error)

	AcquireUserSlot(ctx context.Context, userID int64, maxConcurrency int, requestID string) (bool, error)
	ReleaseUserSlot(ctx context.Context, userID int64, requestID string) error
	GetUserConcurrency(ctx context.Context, userID int64) (int, error)

	IncrementWaitCount(ctx context.Context, userID int64, maxWait int) (bool, error)
	DecrementWaitCount(ctx context.Context, userID int64) error

	GetAccountsLoadBatch(ctx context.Context, accounts []domain.AccountWithConcurrency) (map[int64]*domain.AccountLoadInfo, error)
	GetUsersLoadBatch(ctx context.Context, users []domain.UserWithConcurrency) (map[int64]*domain.UserLoadInfo, error)

	CleanupExpiredAccountSlots(ctx context.Context, accountID int64) error
	CleanupStaleProcessSlots(ctx context.Context, activeRequestPrefix string) error
}
