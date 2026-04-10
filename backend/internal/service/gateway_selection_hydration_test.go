package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGatewayService_SelectAccountForModelWithExclusions_HydratesSnapshotSelection(t *testing.T) {
	ctx := context.Background()
	full := &Account{
		ID:          9101,
		Name:        "hydrated-gateway",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    1,
		Credentials: map[string]any{
			"access_token": "full-only-token",
		},
	}
	snapshot := &Account{
		ID:          full.ID,
		Name:        full.Name,
		Platform:    full.Platform,
		Type:        full.Type,
		Status:      full.Status,
		Schedulable: full.Schedulable,
		Concurrency: full.Concurrency,
		Priority:    full.Priority,
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{snapshot},
		accountsByID:     map[int64]*Account{full.ID: full},
	}
	svc := &GatewayService{
		accountRepo:       stubOpenAIAccountRepo{accounts: []Account{*full}},
		cache:             &stubGatewayCache{},
		cfg:               &config.Config{RunMode: config.RunModeStandard},
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
	}

	account, err := svc.SelectAccountForModelWithExclusions(ctx, nil, "", "claude-sonnet-4-5", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, full.ID, account.ID)
	require.Equal(t, "full-only-token", account.Credentials["access_token"])
}

func TestGatewayService_SelectAccountWithLoadAwareness_HydratesSnapshotSelection(t *testing.T) {
	ctx := context.Background()
	full := &Account{
		ID:          9102,
		Name:        "hydrated-gateway-load-aware",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    1,
		Credentials: map[string]any{
			"access_token": "full-only-token",
		},
	}
	snapshot := &Account{
		ID:          full.ID,
		Name:        full.Name,
		Platform:    full.Platform,
		Type:        full.Type,
		Status:      full.Status,
		Schedulable: full.Schedulable,
		Concurrency: full.Concurrency,
		Priority:    full.Priority,
	}
	cfg := &config.Config{RunMode: config.RunModeStandard}
	cfg.Gateway.Scheduling.LoadBatchEnabled = true
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{snapshot},
		accountsByID:     map[int64]*Account{full.ID: full},
	}
	svc := &GatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{*full}},
		cache:              &stubGatewayCache{},
		cfg:                cfg,
		schedulerSnapshot:  &SchedulerSnapshotService{cache: snapshotCache},
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
	}

	selection, err := svc.SelectAccountWithLoadAwareness(ctx, nil, "", "claude-sonnet-4-5", nil, "")
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.True(t, selection.Acquired)
	require.NotNil(t, selection.Account)
	require.Equal(t, full.ID, selection.Account.ID)
	require.Equal(t, "full-only-token", selection.Account.Credentials["access_token"])
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
}
