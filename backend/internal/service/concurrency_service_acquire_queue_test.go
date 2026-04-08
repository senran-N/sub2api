package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type acquireQueueTestCache struct {
	accountAcquireFn        func(context.Context, int64, int, string) (bool, error)
	accountAcquireOrQueueFn func(context.Context, int64, int, int, string) (bool, bool, error)
	accountReleaseCalls     atomic.Int32
	accountWaitAllowed      bool
	accountWaitErr          error

	userAcquireFn        func(context.Context, int64, int, string) (bool, error)
	userAcquireOrQueueFn func(context.Context, int64, int, int, string) (bool, bool, error)
	userReleaseCalls     atomic.Int32
	userWaitAllowed      bool
	userWaitErr          error
}

var _ ConcurrencyCache = (*acquireQueueTestCache)(nil)

func (c *acquireQueueTestCache) AcquireAccountSlot(ctx context.Context, accountID int64, maxConcurrency int, requestID string) (bool, error) {
	if c.accountAcquireFn != nil {
		return c.accountAcquireFn(ctx, accountID, maxConcurrency, requestID)
	}
	return false, nil
}

func (c *acquireQueueTestCache) AcquireAccountSlotOrEnqueueWait(ctx context.Context, accountID int64, maxConcurrency int, maxWait int, requestID string) (bool, bool, error) {
	if c.accountAcquireOrQueueFn != nil {
		return c.accountAcquireOrQueueFn(ctx, accountID, maxConcurrency, maxWait, requestID)
	}
	return false, false, nil
}

func (c *acquireQueueTestCache) ReleaseAccountSlot(context.Context, int64, string) error {
	c.accountReleaseCalls.Add(1)
	return nil
}

func (c *acquireQueueTestCache) GetAccountConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}

func (c *acquireQueueTestCache) GetAccountConcurrencyBatch(context.Context, []int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}

func (c *acquireQueueTestCache) IncrementAccountWaitCount(context.Context, int64, int) (bool, error) {
	return c.accountWaitAllowed, c.accountWaitErr
}

func (c *acquireQueueTestCache) DecrementAccountWaitCount(context.Context, int64) error {
	return nil
}

func (c *acquireQueueTestCache) GetAccountWaitingCount(context.Context, int64) (int, error) {
	return 0, nil
}

func (c *acquireQueueTestCache) AcquireUserSlot(ctx context.Context, userID int64, maxConcurrency int, requestID string) (bool, error) {
	if c.userAcquireFn != nil {
		return c.userAcquireFn(ctx, userID, maxConcurrency, requestID)
	}
	return false, nil
}

func (c *acquireQueueTestCache) AcquireUserSlotOrEnqueueWait(ctx context.Context, userID int64, maxConcurrency int, maxWait int, requestID string) (bool, bool, error) {
	if c.userAcquireOrQueueFn != nil {
		return c.userAcquireOrQueueFn(ctx, userID, maxConcurrency, maxWait, requestID)
	}
	return false, false, nil
}

func (c *acquireQueueTestCache) ReleaseUserSlot(context.Context, int64, string) error {
	c.userReleaseCalls.Add(1)
	return nil
}

func (c *acquireQueueTestCache) GetUserConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}

func (c *acquireQueueTestCache) IncrementWaitCount(context.Context, int64, int) (bool, error) {
	return c.userWaitAllowed, c.userWaitErr
}

func (c *acquireQueueTestCache) DecrementWaitCount(context.Context, int64) error {
	return nil
}

func (c *acquireQueueTestCache) GetAccountsLoadBatch(context.Context, []AccountWithConcurrency) (map[int64]*AccountLoadInfo, error) {
	return map[int64]*AccountLoadInfo{}, nil
}

func (c *acquireQueueTestCache) GetUsersLoadBatch(context.Context, []UserWithConcurrency) (map[int64]*UserLoadInfo, error) {
	return map[int64]*UserLoadInfo{}, nil
}

func (c *acquireQueueTestCache) CleanupExpiredAccountSlots(context.Context, int64) error {
	return nil
}

func (c *acquireQueueTestCache) CleanupStaleProcessSlots(context.Context, string) error {
	return nil
}

func TestAcquireUserSlotOrQueue_AcquiresAndReleases(t *testing.T) {
	cache := &acquireQueueTestCache{
		userAcquireOrQueueFn: func(context.Context, int64, int, int, string) (bool, bool, error) {
			return true, false, nil
		},
	}
	svc := NewConcurrencyService(cache)

	result, err := svc.AcquireUserSlotOrQueue(context.Background(), 11, 2, 5)
	require.NoError(t, err)
	require.True(t, result.Acquired)
	require.True(t, result.QueueAllowed)
	require.NotNil(t, result.ReleaseFunc)

	result.ReleaseFunc()
	require.Equal(t, int32(1), cache.userReleaseCalls.Load())
}

func TestAcquireAccountSlotOrQueue_EnqueuesWhenBusy(t *testing.T) {
	cache := &acquireQueueTestCache{
		accountAcquireOrQueueFn: func(context.Context, int64, int, int, string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewConcurrencyService(cache)

	result, err := svc.AcquireAccountSlotOrQueue(context.Background(), 21, 1, 3)
	require.NoError(t, err)
	require.False(t, result.Acquired)
	require.True(t, result.QueueAllowed)
	require.True(t, result.WaitCounted)
	require.Nil(t, result.ReleaseFunc)
}

func TestAcquireAccountSlotOrQueue_QueueFull(t *testing.T) {
	cache := &acquireQueueTestCache{
		accountAcquireOrQueueFn: func(context.Context, int64, int, int, string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewConcurrencyService(cache)

	result, err := svc.AcquireAccountSlotOrQueue(context.Background(), 22, 1, 3)
	require.NoError(t, err)
	require.False(t, result.Acquired)
	require.False(t, result.QueueAllowed)
	require.False(t, result.WaitCounted)
}

func TestAcquireUserSlotOrQueue_FallbackAllowsWaitWhenIncrementFails(t *testing.T) {
	cache := &acquireQueueTestCache{
		userAcquireOrQueueFn: func(context.Context, int64, int, int, string) (bool, bool, error) {
			return false, false, errors.New("script failed")
		},
		userAcquireFn: func(context.Context, int64, int, string) (bool, error) {
			return false, nil
		},
		userWaitAllowed: false,
		userWaitErr:     errors.New("wait increment failed"),
	}
	svc := NewConcurrencyService(cache)

	result, err := svc.AcquireUserSlotOrQueue(context.Background(), 12, 2, 5)
	require.NoError(t, err)
	require.False(t, result.Acquired)
	require.True(t, result.QueueAllowed)
	require.False(t, result.WaitCounted)
}
