package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

type schedulerOutboxTestCache struct {
	SchedulerCache
	watermark         int64
	bucketCalls       []SchedulerBucket
	watermarkCtxPtrs  []string
	setWatermarkCalls int
}

func (c *schedulerOutboxTestCache) GetSnapshot(context.Context, SchedulerBucket) ([]*Account, bool, error) {
	return nil, false, nil
}

func (c *schedulerOutboxTestCache) SetSnapshot(context.Context, SchedulerBucket, []Account) error {
	return nil
}

func (c *schedulerOutboxTestCache) GetAccount(context.Context, int64) (*Account, error) {
	return nil, nil
}

func (c *schedulerOutboxTestCache) SetAccount(context.Context, *Account) error {
	return nil
}

func (c *schedulerOutboxTestCache) DeleteAccount(context.Context, int64) error {
	return nil
}

func (c *schedulerOutboxTestCache) UpdateLastUsed(context.Context, map[int64]time.Time) error {
	return nil
}

func (c *schedulerOutboxTestCache) TryLockBucket(_ context.Context, bucket SchedulerBucket, _ time.Duration) (bool, error) {
	c.bucketCalls = append(c.bucketCalls, bucket)
	return false, nil
}

func (c *schedulerOutboxTestCache) ListBuckets(context.Context) ([]SchedulerBucket, error) {
	return nil, nil
}

func (c *schedulerOutboxTestCache) GetOutboxWatermark(context.Context) (int64, error) {
	return c.watermark, nil
}

func (c *schedulerOutboxTestCache) SetOutboxWatermark(ctx context.Context, id int64) error {
	c.setWatermarkCalls++
	c.watermarkCtxPtrs = append(c.watermarkCtxPtrs, fmt.Sprintf("%p", ctx))
	if c.setWatermarkCalls == 1 {
		return context.DeadlineExceeded
	}
	c.watermark = id
	return nil
}

type schedulerOutboxTestRepo struct {
	SchedulerOutboxRepository
	events []domain.SchedulerOutboxEvent
}

func (r *schedulerOutboxTestRepo) ListAfter(context.Context, int64, int) ([]domain.SchedulerOutboxEvent, error) {
	return r.events, nil
}

func (r *schedulerOutboxTestRepo) MaxID(context.Context) (int64, error) {
	if len(r.events) == 0 {
		return 0, nil
	}
	return r.events[len(r.events)-1].ID, nil
}

type schedulerOutboxTestAccountRepo struct {
	AccountRepository
	accounts map[int64]*Account
}

func (r *schedulerOutboxTestAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	account, ok := r.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func TestSchedulerSnapshotService_PollOutboxDedupesGroupPlatformRebuildsWithinBatch(t *testing.T) {
	accountID1 := int64(1)
	accountID2 := int64(2)
	groupID11 := int64(11)
	groupID22 := int64(22)

	cache := &schedulerOutboxTestCache{}
	repo := &schedulerOutboxTestRepo{
		events: []domain.SchedulerOutboxEvent{
			{ID: 101, EventType: SchedulerOutboxEventAccountChanged, AccountID: &accountID1, Payload: map[string]any{"group_ids": []any{groupID11, groupID22}}},
			{ID: 102, EventType: SchedulerOutboxEventAccountChanged, AccountID: &accountID2, Payload: map[string]any{"group_ids": []any{groupID11, groupID22}}},
		},
	}
	accountRepo := &schedulerOutboxTestAccountRepo{
		accounts: map[int64]*Account{
			accountID1: {ID: accountID1, Platform: PlatformOpenAI, GroupIDs: []int64{groupID11, groupID22}},
			accountID2: {ID: accountID2, Platform: PlatformOpenAI, GroupIDs: []int64{groupID11, groupID22}},
		},
	}

	svc := &SchedulerSnapshotService{
		cache:       cache,
		outboxRepo:  repo,
		accountRepo: accountRepo,
	}

	svc.pollOutbox()

	require.Len(t, cache.bucketCalls, 4)
	require.ElementsMatch(t, []SchedulerBucket{
		{GroupID: groupID11, Platform: PlatformOpenAI, Mode: SchedulerModeSingle},
		{GroupID: groupID11, Platform: PlatformOpenAI, Mode: SchedulerModeForced},
		{GroupID: groupID22, Platform: PlatformOpenAI, Mode: SchedulerModeSingle},
		{GroupID: groupID22, Platform: PlatformOpenAI, Mode: SchedulerModeForced},
	}, cache.bucketCalls)
	require.Equal(t, int64(102), cache.watermark)
}

func TestSchedulerSnapshotService_PollOutboxUsesFreshContextForWatermarkRetry(t *testing.T) {
	cache := &schedulerOutboxTestCache{}
	repo := &schedulerOutboxTestRepo{
		events: []domain.SchedulerOutboxEvent{
			{ID: 7, EventType: SchedulerOutboxEventAccountLastUsed},
		},
	}

	svc := &SchedulerSnapshotService{
		cache:      cache,
		outboxRepo: repo,
	}

	svc.pollOutbox()

	require.Equal(t, 2, cache.setWatermarkCalls)
	require.Len(t, cache.watermarkCtxPtrs, 2)
	require.NotEqual(t, cache.watermarkCtxPtrs[0], cache.watermarkCtxPtrs[1])
	require.Equal(t, int64(7), cache.watermark)
}
