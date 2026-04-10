//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSchedulerSnapshotOutboxReplay(t *testing.T) {
	ctx := context.Background()
	rdb := testRedis(t)
	client := testEntClient(t)

	_, _ = integrationDB.ExecContext(ctx, "TRUNCATE scheduler_outbox")

	accountRepo := newAccountRepositoryWithSQL(client, integrationDB, nil)
	outboxRepo := NewSchedulerOutboxRepository(integrationDB)
	cache := NewSchedulerCache(rdb)

	cfg := &config.Config{
		RunMode: config.RunModeStandard,
		Gateway: config.GatewayConfig{
			Scheduling: config.GatewaySchedulingConfig{
				OutboxPollIntervalSeconds:  1,
				FullRebuildIntervalSeconds: 0,
				DbFallbackEnabled:          true,
			},
		},
	}

	account := &service.Account{
		Name:        "outbox-replay-" + time.Now().Format("150405.000000"),
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 3,
		Priority:    1,
		Credentials: map[string]any{},
		Extra:       map[string]any{},
	}
	require.NoError(t, accountRepo.Create(ctx, account))
	require.NoError(t, cache.SetAccount(ctx, account))

	svc := service.NewSchedulerSnapshotService(cache, outboxRepo, accountRepo, nil, cfg)
	svc.Start()
	t.Cleanup(svc.Stop)

	require.NoError(t, accountRepo.UpdateLastUsed(ctx, account.ID))
	updated, err := accountRepo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.LastUsedAt)
	expectedUnix := updated.LastUsedAt.Unix()

	require.Eventually(t, func() bool {
		cached, err := cache.GetAccount(ctx, account.ID)
		if err != nil || cached == nil || cached.LastUsedAt == nil {
			return false
		}
		return cached.LastUsedAt.Unix() == expectedUnix
	}, 5*time.Second, 100*time.Millisecond)
}

func TestSchedulerCache_GetSnapshotPage(t *testing.T) {
	ctx := context.Background()
	rdb := testRedis(t)

	cache := NewSchedulerCache(rdb)
	pager, ok := cache.(service.SchedulerCachePager)
	require.True(t, ok)

	bucket := service.SchedulerBucket{
		GroupID:  42,
		Platform: service.PlatformOpenAI,
		Mode:     service.SchedulerModeSingle,
	}
	accounts := []service.Account{
		{ID: 41001, Name: "page-1", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Status: service.StatusActive, Schedulable: true, Priority: 0, Concurrency: 1},
		{ID: 41002, Name: "page-2", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Status: service.StatusActive, Schedulable: true, Priority: 1, Concurrency: 1},
		{ID: 41003, Name: "page-3", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Status: service.StatusActive, Schedulable: true, Priority: 2, Concurrency: 1},
	}
	require.NoError(t, cache.SetSnapshot(ctx, bucket, accounts))

	firstPage, hit, hasMore, err := pager.GetSnapshotPage(ctx, bucket, 0, 2)
	require.NoError(t, err)
	require.True(t, hit)
	require.True(t, hasMore)
	require.Len(t, firstPage, 2)
	require.Equal(t, int64(41001), firstPage[0].ID)
	require.Equal(t, int64(41002), firstPage[1].ID)

	secondPage, hit, hasMore, err := pager.GetSnapshotPage(ctx, bucket, 2, 2)
	require.NoError(t, err)
	require.True(t, hit)
	require.False(t, hasMore)
	require.Len(t, secondPage, 1)
	require.Equal(t, int64(41003), secondPage[0].ID)
}
