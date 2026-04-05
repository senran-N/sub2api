package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestResolveUsageBillingRequestID_PrefersClientRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxkey.ClientRequestID, " client-123 ")
	ctx = context.WithValue(ctx, ctxkey.RequestID, "local-456")

	got := resolveUsageBillingRequestID(ctx, "upstream-789")

	require.Equal(t, "client:client-123", got)
}

func TestResolveUsageBillingPayloadFingerprint_FallsBackToLocalRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxkey.RequestID, " local-456 ")

	got := resolveUsageBillingPayloadFingerprint(ctx, "")

	require.Equal(t, "local:local-456", got)
}

func TestBuildUsageBillingCommand_PopulatesCostsAndMetadata(t *testing.T) {
	groupID := int64(7)
	subscriptionID := int64(11)
	mediaType := "image"
	serviceTier := "priority"
	reasoningEffort := "high"

	cmd := buildUsageBillingCommand("req-1", &UsageLog{
		Model:               "claude-sonnet-4-5",
		BillingType:         BillingTypeBalance,
		InputTokens:         10,
		OutputTokens:        20,
		CacheCreationTokens: 3,
		CacheReadTokens:     4,
		ImageCount:          1,
		MediaType:           &mediaType,
		ServiceTier:         &serviceTier,
		ReasoningEffort:     &reasoningEffort,
		SubscriptionID:      &subscriptionID,
	}, &postUsageBillingParams{
		Cost: &CostBreakdown{
			TotalCost:  6,
			ActualCost: 5,
		},
		APIKey: &APIKey{
			ID:          101,
			Quota:       99,
			GroupID:     &groupID,
			RateLimit5h: 10,
		},
		User: &User{ID: 202},
		Account: &Account{
			ID:   303,
			Type: AccountTypeAPIKey,
			Extra: map[string]any{
				"quota_limit": 10,
			},
		},
		APIKeyService:         &testAPIKeyQuotaUpdater{},
		AccountRateMultiplier: 1.5,
		RequestPayloadHash:    " hash-1 ",
	})

	require.NotNil(t, cmd)
	require.Equal(t, "req-1", cmd.RequestID)
	require.Equal(t, int64(101), cmd.APIKeyID)
	require.Equal(t, int64(202), cmd.UserID)
	require.Equal(t, int64(303), cmd.AccountID)
	require.Equal(t, AccountTypeAPIKey, cmd.AccountType)
	require.Equal(t, "hash-1", cmd.RequestPayloadHash)
	require.Equal(t, "claude-sonnet-4-5", cmd.Model)
	require.Equal(t, BillingTypeBalance, cmd.BillingType)
	require.Equal(t, 10, cmd.InputTokens)
	require.Equal(t, 20, cmd.OutputTokens)
	require.Equal(t, 3, cmd.CacheCreationTokens)
	require.Equal(t, 4, cmd.CacheReadTokens)
	require.Equal(t, 1, cmd.ImageCount)
	require.Equal(t, "image", cmd.MediaType)
	require.Equal(t, "priority", cmd.ServiceTier)
	require.Equal(t, "high", cmd.ReasoningEffort)
	require.Equal(t, 5.0, cmd.BalanceCost)
	require.Equal(t, 5.0, cmd.APIKeyQuotaCost)
	require.Equal(t, 5.0, cmd.APIKeyRateLimitCost)
	require.Equal(t, 9.0, cmd.AccountQuotaCost)
	require.Equal(t, &subscriptionID, cmd.SubscriptionID)
}

type testAPIKeyQuotaUpdater struct{}

func (u *testAPIKeyQuotaUpdater) UpdateQuotaUsed(context.Context, int64, float64) error {
	return nil
}

func (u *testAPIKeyQuotaUpdater) UpdateRateLimitUsage(context.Context, int64, float64) error {
	return nil
}

func TestPostUsageBilling_AllowsNilCost(t *testing.T) {
	deferred := &DeferredService{}

	postUsageBilling(context.Background(), &postUsageBillingParams{
		Cost:    nil,
		Account: &Account{ID: 303},
	}, &billingDeps{
		deferredService:     deferred,
		billingCacheService: &BillingCacheService{},
	})

	_, ok := deferred.lastUsedUpdates.Load(int64(303))
	require.True(t, ok, "nil cost should still schedule last_used update")
}

func TestFinalizePostUsageBilling_WithNilCostStillSchedulesLastUsedUpdate(t *testing.T) {
	deferred := &DeferredService{}

	finalizePostUsageBilling(&postUsageBillingParams{
		Cost:    nil,
		Account: &Account{ID: 303},
	}, &billingDeps{
		deferredService:     deferred,
		billingCacheService: &BillingCacheService{},
	})

	_, ok := deferred.lastUsedUpdates.Load(int64(303))
	require.True(t, ok, "nil cost should still schedule last_used update")
}

func TestFinalizePostUsageBilling_AllowsNilServices(t *testing.T) {
	finalizePostUsageBilling(&postUsageBillingParams{
		Cost:    &CostBreakdown{ActualCost: 1, TotalCost: 2},
		User:    &User{ID: 1},
		APIKey:  &APIKey{ID: 2},
		Account: &Account{ID: 3},
	}, &billingDeps{})
}
