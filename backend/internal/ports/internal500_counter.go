package ports

import "context"

// Internal500CounterCache tracks consecutive INTERNAL 500 failures for an account.
type Internal500CounterCache interface {
	IncrementInternal500Count(ctx context.Context, accountID int64) (int64, error)
	ResetInternal500Count(ctx context.Context, accountID int64) error
}
