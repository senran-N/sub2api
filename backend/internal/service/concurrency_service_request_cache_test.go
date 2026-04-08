package service

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type requestCacheTestConcurrencyCache struct {
	mu             sync.Mutex
	loadBatchCalls int
	requestedIDs   [][]int64
	loadBatchFunc  func([]AccountWithConcurrency) map[int64]*AccountLoadInfo
}

var _ ConcurrencyCache = (*requestCacheTestConcurrencyCache)(nil)

func (c *requestCacheTestConcurrencyCache) AcquireAccountSlot(context.Context, int64, int, string) (bool, error) {
	return false, nil
}

func (c *requestCacheTestConcurrencyCache) AcquireAccountSlotOrEnqueueWait(ctx context.Context, accountID int64, maxConcurrency int, maxWait int, requestID string) (bool, bool, error) {
	acquired, err := c.AcquireAccountSlot(ctx, accountID, maxConcurrency, requestID)
	return acquired, !acquired, err
}

func (c *requestCacheTestConcurrencyCache) ReleaseAccountSlot(context.Context, int64, string) error {
	return nil
}

func (c *requestCacheTestConcurrencyCache) GetAccountConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}

func (c *requestCacheTestConcurrencyCache) GetAccountConcurrencyBatch(_ context.Context, accountIDs []int64) (map[int64]int, error) {
	result := make(map[int64]int, len(accountIDs))
	for _, accountID := range accountIDs {
		result[accountID] = 0
	}
	return result, nil
}

func (c *requestCacheTestConcurrencyCache) IncrementAccountWaitCount(context.Context, int64, int) (bool, error) {
	return false, nil
}

func (c *requestCacheTestConcurrencyCache) DecrementAccountWaitCount(context.Context, int64) error {
	return nil
}

func (c *requestCacheTestConcurrencyCache) GetAccountWaitingCount(context.Context, int64) (int, error) {
	return 0, nil
}

func (c *requestCacheTestConcurrencyCache) AcquireUserSlot(context.Context, int64, int, string) (bool, error) {
	return false, nil
}

func (c *requestCacheTestConcurrencyCache) AcquireUserSlotOrEnqueueWait(ctx context.Context, userID int64, maxConcurrency int, maxWait int, requestID string) (bool, bool, error) {
	acquired, err := c.AcquireUserSlot(ctx, userID, maxConcurrency, requestID)
	return acquired, !acquired, err
}

func (c *requestCacheTestConcurrencyCache) ReleaseUserSlot(context.Context, int64, string) error {
	return nil
}

func (c *requestCacheTestConcurrencyCache) GetUserConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}

func (c *requestCacheTestConcurrencyCache) IncrementWaitCount(context.Context, int64, int) (bool, error) {
	return false, nil
}

func (c *requestCacheTestConcurrencyCache) DecrementWaitCount(context.Context, int64) error {
	return nil
}

func (c *requestCacheTestConcurrencyCache) GetAccountsLoadBatch(_ context.Context, accounts []AccountWithConcurrency) (map[int64]*AccountLoadInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.loadBatchCalls++
	ids := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}
	c.requestedIDs = append(c.requestedIDs, ids)

	if c.loadBatchFunc == nil {
		return map[int64]*AccountLoadInfo{}, nil
	}
	return c.loadBatchFunc(accounts), nil
}

func (c *requestCacheTestConcurrencyCache) GetUsersLoadBatch(context.Context, []UserWithConcurrency) (map[int64]*UserLoadInfo, error) {
	return map[int64]*UserLoadInfo{}, nil
}

func (c *requestCacheTestConcurrencyCache) CleanupExpiredAccountSlots(context.Context, int64) error {
	return nil
}

func (c *requestCacheTestConcurrencyCache) CleanupStaleProcessSlots(context.Context, string) error {
	return nil
}

func TestGetAccountsLoadBatch_RequestScopedCacheReusesSnapshot(t *testing.T) {
	cache := &requestCacheTestConcurrencyCache{
		loadBatchFunc: func(accounts []AccountWithConcurrency) map[int64]*AccountLoadInfo {
			result := make(map[int64]*AccountLoadInfo, len(accounts))
			for _, account := range accounts {
				switch account.ID {
				case 1:
					result[1] = &AccountLoadInfo{AccountID: 1, CurrentConcurrency: 3, WaitingCount: 1}
				case 2:
					result[2] = &AccountLoadInfo{AccountID: 2, CurrentConcurrency: 2, WaitingCount: 0}
				}
			}
			return result
		},
	}
	svc := NewConcurrencyService(cache)
	ctx := WithRequestAccountLoadCache(context.Background())

	first, err := svc.GetAccountsLoadBatch(ctx, []AccountWithConcurrency{
		{ID: 1, MaxConcurrency: 5},
		{ID: 2, MaxConcurrency: 4},
	})
	require.NoError(t, err)
	require.Equal(t, 1, cache.loadBatchCalls)
	require.Equal(t, 80, first[1].LoadRate)
	require.Equal(t, 50, first[2].LoadRate)

	second, err := svc.GetAccountsLoadBatch(ctx, []AccountWithConcurrency{
		{ID: 1, MaxConcurrency: 10},
	})
	require.NoError(t, err)
	require.Equal(t, 1, cache.loadBatchCalls)
	require.Equal(t, 40, second[1].LoadRate)
	require.Equal(t, 3, second[1].CurrentConcurrency)
	require.Equal(t, 1, second[1].WaitingCount)
}

func TestGetAccountsLoadBatch_RequestScopedCacheFetchesOnlyMissingAccounts(t *testing.T) {
	cache := &requestCacheTestConcurrencyCache{
		loadBatchFunc: func(accounts []AccountWithConcurrency) map[int64]*AccountLoadInfo {
			result := make(map[int64]*AccountLoadInfo, len(accounts))
			for _, account := range accounts {
				result[account.ID] = &AccountLoadInfo{
					AccountID:          account.ID,
					CurrentConcurrency: int(account.ID),
					WaitingCount:       int(account.ID - 1),
				}
			}
			return result
		},
	}
	svc := NewConcurrencyService(cache)
	ctx := WithRequestAccountLoadCache(context.Background())

	first, err := svc.GetAccountsLoadBatch(ctx, []AccountWithConcurrency{
		{ID: 1, MaxConcurrency: 4},
	})
	require.NoError(t, err)
	require.Equal(t, 1, cache.loadBatchCalls)
	require.Equal(t, []int64{1}, cache.requestedIDs[0])
	require.Equal(t, 25, first[1].LoadRate)

	second, err := svc.GetAccountsLoadBatch(ctx, []AccountWithConcurrency{
		{ID: 1, MaxConcurrency: 8},
		{ID: 2, MaxConcurrency: 4},
	})
	require.NoError(t, err)
	require.Equal(t, 2, cache.loadBatchCalls)
	require.Equal(t, []int64{2}, cache.requestedIDs[1])
	require.Equal(t, 12, second[1].LoadRate)
	require.Equal(t, 75, second[2].LoadRate)
}

func BenchmarkGetAccountsLoadBatch_RequestScopedCacheHit(b *testing.B) {
	cache := &requestCacheTestConcurrencyCache{
		loadBatchFunc: func(accounts []AccountWithConcurrency) map[int64]*AccountLoadInfo {
			result := make(map[int64]*AccountLoadInfo, len(accounts))
			for _, account := range accounts {
				result[account.ID] = &AccountLoadInfo{
					AccountID:          account.ID,
					CurrentConcurrency: int(account.ID),
					WaitingCount:       1,
				}
			}
			return result
		},
	}
	svc := NewConcurrencyService(cache)
	ctx := WithRequestAccountLoadCache(context.Background())
	accounts := []AccountWithConcurrency{
		{ID: 1, MaxConcurrency: 4},
		{ID: 2, MaxConcurrency: 6},
		{ID: 3, MaxConcurrency: 8},
	}

	_, err := svc.GetAccountsLoadBatch(ctx, accounts)
	require.NoError(b, err)
	require.Equal(b, 1, cache.loadBatchCalls)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadMap, err := svc.GetAccountsLoadBatch(ctx, accounts)
		if err != nil {
			b.Fatal(err)
		}
		if len(loadMap) != len(accounts) {
			b.Fatalf("unexpected load map size: %d", len(loadMap))
		}
	}
}

func BenchmarkGetAccountsLoadBatch_WithoutRequestScopedCache(b *testing.B) {
	cache := &requestCacheTestConcurrencyCache{
		loadBatchFunc: func(accounts []AccountWithConcurrency) map[int64]*AccountLoadInfo {
			result := make(map[int64]*AccountLoadInfo, len(accounts))
			for _, account := range accounts {
				result[account.ID] = &AccountLoadInfo{
					AccountID:          account.ID,
					CurrentConcurrency: int(account.ID),
					WaitingCount:       1,
				}
			}
			return result
		},
	}
	svc := NewConcurrencyService(cache)
	accounts := []AccountWithConcurrency{
		{ID: 1, MaxConcurrency: 4},
		{ID: 2, MaxConcurrency: 6},
		{ID: 3, MaxConcurrency: 8},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadMap, err := svc.GetAccountsLoadBatch(context.Background(), accounts)
		if err != nil {
			b.Fatal(err)
		}
		if len(loadMap) != len(accounts) {
			b.Fatalf("unexpected load map size: %d", len(loadMap))
		}
	}
}
