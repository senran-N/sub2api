package service

import (
	"context"
	"testing"

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
