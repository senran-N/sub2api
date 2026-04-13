//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGeminiMessagesCompatService_SelectAccountForModelWithExclusions_SkipsExpiredOAuthWithoutRefreshToken(t *testing.T) {
	now := time.Now().UTC()
	repo := &mockAccountRepoForGemini{
		accounts: []Account{
			{
				ID:          1,
				Platform:    PlatformGemini,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    1,
				Credentials: map[string]any{
					"access_token": "expired-token",
					"expires_at":   now.Add(-time.Minute).Format(time.RFC3339),
				},
			},
			{
				ID:          2,
				Platform:    PlatformGemini,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    2,
				Credentials: map[string]any{
					"access_token": "healthy-token",
					"expires_at":   now.Add(time.Hour).Format(time.RFC3339),
				},
			},
		},
		accountsByID: map[int64]*Account{},
	}
	for i := range repo.accounts {
		repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
	}

	svc := &GeminiMessagesCompatService{
		accountRepo: repo,
		groupRepo:   &mockGroupRepoForGemini{groups: map[int64]*Group{}},
		cache:       &mockGatewayCacheForGemini{},
	}

	account, err := svc.SelectAccountForModelWithExclusions(context.Background(), nil, "", "gemini-2.5-flash", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, int64(2), account.ID)
}

func TestGeminiMessagesCompatService_SelectAccountForAIStudioEndpoints_SkipsExpiredOAuthWithoutRefreshToken(t *testing.T) {
	now := time.Now().UTC()
	first := &Account{
		ID:          9403,
		Name:        "gemini-aistudio-expired",
		Platform:    PlatformGemini,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Priority:    1,
		Credentials: map[string]any{
			"access_token": "expired-token",
			"expires_at":   now.Add(-time.Minute).Format(time.RFC3339),
		},
	}
	second := &Account{
		ID:          9404,
		Name:        "gemini-aistudio-healthy",
		Platform:    PlatformGemini,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Priority:    2,
		Credentials: map[string]any{
			"access_token": "healthy-token",
			"expires_at":   now.Add(time.Hour).Format(time.RFC3339),
		},
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{first, second},
		accountsByID: map[int64]*Account{
			first.ID:  first,
			second.ID: second,
		},
	}
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.SnapshotPageSize = 1
	svc := &GeminiMessagesCompatService{
		accountRepo:       &mockAccountRepoForGemini{},
		cache:             &mockGatewayCacheForGemini{},
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
		cfg:               cfg,
	}

	account, err := svc.SelectAccountForAIStudioEndpoints(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, second.ID, account.ID)
}
