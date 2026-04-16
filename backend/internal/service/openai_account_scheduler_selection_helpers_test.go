package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeOpenAISchedulerTopK(t *testing.T) {
	tests := []struct {
		name          string
		configuredTop int
		candidateCnt  int
		want          int
	}{
		{name: "clamps to candidate count", configuredTop: 10, candidateCnt: 3, want: 3},
		{name: "zero becomes one", configuredTop: 0, candidateCnt: 5, want: 1},
		{name: "negative becomes one", configuredTop: -1, candidateCnt: 5, want: 1},
		{name: "normal value unchanged", configuredTop: 2, candidateCnt: 5, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeOpenAISchedulerTopK(tt.configuredTop, tt.candidateCnt)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPrepareLoadBalanceCandidates_FiltersByExclusionAndModelAndTransport(t *testing.T) {
	req := OpenAIAccountScheduleRequest{
		RequestedModel:    "gpt-4",
		RequiredTransport: OpenAIUpstreamTransportAny,
		ExcludedIDs:       map[int64]struct{}{1: {}},
	}

	accounts := []Account{
		{ID: 1, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Concurrency: 3},
		{ID: 2, Platform: PlatformGemini, Status: StatusActive, Schedulable: true, Concurrency: 2},
		{
			ID:          3,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 2,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"gpt-3.5": "gpt-3.5"},
			},
		},
		{ID: 4, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Concurrency: 5},
	}

	svc := &OpenAIGatewayService{}
	scheduler := &defaultOpenAIAccountScheduler{service: svc}

	filtered, loadReq := scheduler.prepareLoadBalanceCandidates(context.Background(), req, accounts, nil)

	require.Len(t, filtered, 1)
	require.Equal(t, int64(4), filtered[0].ID)
	require.Len(t, loadReq, 1)
	require.Equal(t, int64(4), loadReq[0].ID)
	require.Equal(t, 5, loadReq[0].MaxConcurrency)
}

func TestPrepareLoadBalanceCandidates_PrivacyRequiredSkipsWithoutSideEffects(t *testing.T) {
	scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{}}

	req := OpenAIAccountScheduleRequest{
		RequiredTransport: OpenAIUpstreamTransportAny,
	}
	accounts := []Account{
		{ID: 100, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Concurrency: 1},
	}
	group := &Group{ID: 9, Name: "privacy-required", RequirePrivacySet: true}

	filtered, loadReq := scheduler.prepareLoadBalanceCandidates(context.Background(), req, accounts, group)

	require.Empty(t, filtered)
	require.Empty(t, loadReq)
}

func TestPrepareLoadBalanceCandidates_SkipsRuntimeUnsafeAccounts(t *testing.T) {
	now := time.Now().UTC()
	usedPercent := 100.0
	resetAfter := 3600
	windowMinutes := 10080
	codexExtra := buildCodexUsageExtraUpdates(&OpenAICodexUsageSnapshot{
		PrimaryUsedPercent:       &usedPercent,
		PrimaryResetAfterSeconds: &resetAfter,
		PrimaryWindowMinutes:     &windowMinutes,
		UpdatedAt:                now.Format(time.RFC3339),
	}, now)

	scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{}}
	req := OpenAIAccountScheduleRequest{
		RequestedModel:    "gpt-5.1",
		RequiredTransport: OpenAIUpstreamTransportAny,
	}
	accounts := []Account{
		{
			ID:          201,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 1,
			Credentials: map[string]any{
				"access_token": "oauth-token",
				"expires_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
			Extra: codexExtra,
		},
		{
			ID:          202,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 1,
			Credentials: map[string]any{
				"access_token": "expired-token",
				"expires_at":   now.Add(-time.Minute).Format(time.RFC3339),
			},
		},
		{
			ID:          203,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 2,
			Credentials: map[string]any{
				"access_token": "fresh-token",
				"expires_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
		},
	}

	filtered, loadReq := scheduler.prepareLoadBalanceCandidates(context.Background(), req, accounts, nil)

	require.Len(t, filtered, 2)
	require.Equal(t, int64(201), filtered[0].ID)
	require.Equal(t, int64(203), filtered[1].ID)
	require.Len(t, loadReq, 2)
	require.Equal(t, int64(201), loadReq[0].ID)
	require.Equal(t, int64(203), loadReq[1].ID)
	require.Nil(t, accounts[0].RateLimitResetAt)
}

func TestLoadSchedulerAccountLoads(t *testing.T) {
	ctx := context.Background()
	loadReq := []AccountWithConcurrency{{ID: 11, MaxConcurrency: 2}}

	t.Run("returns empty map when concurrency service missing", func(t *testing.T) {
		scheduler := &defaultOpenAIAccountScheduler{
			service: &OpenAIGatewayService{},
		}
		loadMap := scheduler.loadSchedulerAccountLoads(ctx, loadReq)
		require.NotNil(t, loadMap)
		require.Empty(t, loadMap)
	})

	t.Run("returns batch load map on success", func(t *testing.T) {
		scheduler := &defaultOpenAIAccountScheduler{
			service: &OpenAIGatewayService{
				concurrencyService: NewConcurrencyService(stubConcurrencyCache{
					loadMap: map[int64]*AccountLoadInfo{
						11: {AccountID: 11, LoadRate: 70, WaitingCount: 2},
					},
				}),
			},
		}
		loadMap := scheduler.loadSchedulerAccountLoads(ctx, loadReq)
		require.Len(t, loadMap, 1)
		require.Equal(t, 70, loadMap[11].LoadRate)
		require.Equal(t, 2, loadMap[11].WaitingCount)
	})

	t.Run("falls back to empty map on batch load error", func(t *testing.T) {
		scheduler := &defaultOpenAIAccountScheduler{
			service: &OpenAIGatewayService{
				concurrencyService: NewConcurrencyService(stubConcurrencyCache{
					loadBatchErr: errors.New("batch failed"),
				}),
			},
		}
		loadMap := scheduler.loadSchedulerAccountLoads(ctx, loadReq)
		require.NotNil(t, loadMap)
		require.Empty(t, loadMap)
	})
}
