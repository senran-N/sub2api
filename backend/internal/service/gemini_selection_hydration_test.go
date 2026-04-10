package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGeminiMessagesCompatService_SelectAccountForModelWithExclusions_HydratesSnapshotSelection(t *testing.T) {
	ctx := context.Background()
	full := &Account{
		ID:          9201,
		Name:        "hydrated-gemini",
		Platform:    PlatformGemini,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    1,
		Credentials: map[string]any{
			"api_key": "kept",
			"secret":  "full-only-secret",
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
		Credentials: map[string]any{
			"api_key": "kept",
		},
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{snapshot},
		accountsByID:     map[int64]*Account{full.ID: full},
	}
	svc := &GeminiMessagesCompatService{
		accountRepo:       stubOpenAIAccountRepo{accounts: []Account{*full}},
		cache:             &stubGatewayCache{},
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
	}

	account, err := svc.SelectAccountForModelWithExclusions(ctx, nil, "", "gemini-2.5-flash", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, full.ID, account.ID)
	require.Equal(t, "full-only-secret", account.Credentials["secret"])
}

func TestGeminiMessagesCompatService_SelectAccountForAIStudioEndpoints_HydratesSnapshotSelection(t *testing.T) {
	ctx := context.Background()
	full := &Account{
		ID:          9202,
		Name:        "hydrated-gemini-aistudio",
		Platform:    PlatformGemini,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    1,
		Credentials: map[string]any{
			"api_key": "kept",
			"secret":  "full-only-secret",
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
		Credentials: map[string]any{
			"api_key": "kept",
		},
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{snapshot},
		accountsByID:     map[int64]*Account{full.ID: full},
	}
	svc := &GeminiMessagesCompatService{
		accountRepo:       stubOpenAIAccountRepo{accounts: []Account{*full}},
		cache:             &stubGatewayCache{},
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
	}

	account, err := svc.SelectAccountForAIStudioEndpoints(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, full.ID, account.ID)
	require.Equal(t, "full-only-secret", account.Credentials["secret"])
}

func TestGeminiMessagesCompatService_SelectAccountForAIStudioEndpoints_UsesCapabilityIndexPaging(t *testing.T) {
	ctx := context.Background()
	first := &Account{
		ID:          9401,
		Name:        "gemini-aistudio-oauth",
		Platform:    PlatformGemini,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    5,
	}
	second := &Account{
		ID:          9402,
		Name:        "gemini-aistudio-key",
		Platform:    PlatformGemini,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    9,
		Credentials: map[string]any{
			"api_key": "kept",
		},
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{first, second},
		accountsByID:     map[int64]*Account{first.ID: first, second.ID: second},
	}
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.SnapshotPageSize = 1
	svc := &GeminiMessagesCompatService{
		accountRepo:       stubOpenAIAccountRepo{accounts: []Account{*first, *second}},
		cache:             &stubGatewayCache{},
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
		cfg:               cfg,
	}

	account, err := svc.SelectAccountForAIStudioEndpoints(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, second.ID, account.ID)
	require.Equal(t, []snapshotIndexPageCall{
		{kind: SchedulerCapabilityIndexAll, value: "", offset: 0, limit: 1},
		{kind: SchedulerCapabilityIndexAll, value: "", offset: 1, limit: 1},
	}, snapshotCache.indexPageCalls)
}

func TestGeminiMessagesCompatService_SelectAccountForModelWithExclusions_UsesCapabilityIndexPaging(t *testing.T) {
	ctx := context.Background()
	first := &Account{
		ID:          9301,
		Name:        "gemini-page-1",
		Platform:    PlatformGemini,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    5,
	}
	second := &Account{
		ID:          9302,
		Name:        "gemini-page-2",
		Platform:    PlatformGemini,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    1,
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{first, second},
		accountsByID:     map[int64]*Account{first.ID: first, second.ID: second},
	}
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.SnapshotPageSize = 1
	svc := &GeminiMessagesCompatService{
		accountRepo:       stubOpenAIAccountRepo{accounts: []Account{*first, *second}},
		cache:             &stubGatewayCache{},
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
		cfg:               cfg,
	}

	account, err := svc.SelectAccountForModelWithExclusions(ctx, nil, "", "gemini-2.5-flash", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, second.ID, account.ID)
	require.Equal(t, []snapshotIndexPageCall{
		{kind: SchedulerCapabilityIndexModelAny, value: "", offset: 0, limit: 1},
		{kind: SchedulerCapabilityIndexModelExact, value: "gemini-2.5-flash", offset: 0, limit: 1},
		{kind: SchedulerCapabilityIndexModelAny, value: "", offset: 1, limit: 1},
	}, snapshotCache.indexPageCalls)
}
