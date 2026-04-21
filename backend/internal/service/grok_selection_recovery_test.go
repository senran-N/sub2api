//go:build unit

package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/stretchr/testify/require"
)

type grokSelectionRecoverySyncerStub struct {
	calls []int64
}

func (s *grokSelectionRecoverySyncerStub) SyncAccount(_ context.Context, account *Account) error {
	if account != nil {
		s.calls = append(s.calls, account.ID)
	}
	return nil
}

type grokSelectionRecoveryMutatingSyncerStub struct {
	calls []int64
}

func (s *grokSelectionRecoveryMutatingSyncerStub) SyncAccount(_ context.Context, account *Account) error {
	if account == nil {
		return nil
	}
	s.calls = append(s.calls, account.ID)
	account.Extra = map[string]any{
		"grok": map[string]any{
			"tier": map[string]any{
				"normalized": "heavy",
				"source":     "on_demand_sync",
			},
			"quota_windows": map[string]any{
				"heavy": map[string]any{
					"remaining": 3,
					"total":     3,
					"source":    "sync",
				},
			},
		},
	}
	return nil
}

type grokSelectionRecoveryProberStub struct {
	calls []int64
}

func (s *grokSelectionRecoveryProberStub) ProbeAccount(_ context.Context, account *Account) error {
	if account != nil {
		s.calls = append(s.calls, account.ID)
	}
	return nil
}

type grokSelectionRecoveryMutationStub struct {
	repo  *mockAccountRepoForPlatform
	calls int
}

func (s *grokSelectionRecoveryMutationStub) RecoverOnDemand(_ context.Context, accounts []Account, _ string) bool {
	s.calls++
	for i := range accounts {
		if accounts[i].ID != 71 {
			continue
		}
		accounts[i].Extra = map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "heavy",
					"source":     "on_demand_recovery",
				},
				"quota_windows": map[string]any{
					"heavy": map[string]any{
						"remaining": 5,
						"total":     5,
						"source":    "sync",
					},
				},
			},
		}
	}
	if s.repo != nil {
		s.repo.applyAccountUpdate(71, func(account *Account) {
			account.Extra = map[string]any{
				"grok": map[string]any{
					"tier": map[string]any{
						"normalized": "heavy",
						"source":     "on_demand_recovery",
					},
					"quota_windows": map[string]any{
						"heavy": map[string]any{
							"remaining": 5,
							"total":     5,
							"source":    "sync",
						},
					},
				},
			}
		})
	}
	return true
}

func TestGrokOnDemandRecoveryServiceRecoverOnDemandThrottlesAndSplitsAccountKinds(t *testing.T) {
	syncer := &grokSelectionRecoverySyncerStub{}
	prober := &grokSelectionRecoveryProberStub{}
	svc := NewGrokOnDemandRecoveryService(syncer, prober, nil)

	now := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	accounts := []Account{
		{
			ID:          1,
			Platform:    PlatformGrok,
			Type:        AccountTypeSession,
			Status:      StatusActive,
			Schedulable: true,
		},
		{
			ID:          2,
			Platform:    PlatformGrok,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
		},
		{
			ID:          3,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
		},
	}

	attempted := svc.RecoverOnDemand(context.Background(), accounts, "grok-4.20-heavy")
	require.True(t, attempted)
	require.Equal(t, []int64{1}, syncer.calls)
	require.Equal(t, []int64{2}, prober.calls)

	attempted = svc.RecoverOnDemand(context.Background(), accounts, "grok-4.20-heavy")
	require.False(t, attempted)
	require.Equal(t, []int64{1}, syncer.calls)
	require.Equal(t, []int64{2}, prober.calls)

	now = now.Add(grokOnDemandRecoveryMinInterval + time.Second)
	attempted = svc.RecoverOnDemand(context.Background(), accounts, "grok-4.20-heavy")
	require.True(t, attempted)
	require.Equal(t, []int64{1, 1}, syncer.calls)
	require.Equal(t, []int64{2, 2}, prober.calls)
}

func TestGrokOnDemandRecoveryServiceRecoverOnDemandMutatesOriginalAccounts(t *testing.T) {
	syncer := &grokSelectionRecoveryMutatingSyncerStub{}
	svc := NewGrokOnDemandRecoveryService(syncer, nil, nil)

	accounts := []Account{
		{
			ID:          11,
			Platform:    PlatformGrok,
			Type:        AccountTypeSession,
			Status:      StatusActive,
			Schedulable: true,
		},
	}

	attempted := svc.RecoverOnDemand(context.Background(), accounts, "grok-4.20-heavy")
	require.True(t, attempted)
	require.Equal(t, []int64{11}, syncer.calls)
	require.Equal(t, grok.TierHeavy, accounts[0].GrokTierState().Normalized)
	require.Equal(t, 3, accounts[0].grokQuotaWindow("heavy").Remaining)
}

func TestGrokGatewayServiceHandleResponses_OnDemandRecoveryCanRestoreSessionModelAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"grok-4.20-heavy","input":"hello"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request = c.Request.WithContext(WithGrokSessionTextRuntimeAllowed(context.Background()))

	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, strings.Join([]string{
				`{"result":{"response":{"token":"answer","messageTag":"final"}}}`,
				`{"result":{"response":{"finalMetadata":{"stop_reason":"end_turn"}}}}`,
			}, "\n")),
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          71,
				Name:        "grok-session-unknown-tier",
				Platform:    PlatformGrok,
				Type:        AccountTypeSession,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"session_token": "session-cookie",
				},
			},
		},
	}
	recovery := &grokSelectionRecoveryMutationStub{repo: repo}
	gatewayService := &GatewayService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg:          testConfig(),
	}
	gatewayService.SetGrokSelectionRecovery(recovery)
	svc := NewGrokGatewayService(gatewayService, nil)

	handled := svc.HandleResponses(c, nil, body)
	require.True(t, handled)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, recovery.calls)
	require.Len(t, upstream.requests, 1)
	require.Contains(t, upstream.requests[0].URL.String(), "/rest/app-chat/conversations/new")
	require.Len(t, repo.runtimeStates, 1)
	require.Equal(t, "success", repo.runtimeStates[0]["last_outcome"])
}
