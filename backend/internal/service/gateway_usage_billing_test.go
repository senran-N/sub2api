package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type usageBillingCacheStub struct {
	mu                 sync.Mutex
	deductCalls        int
	subscriptionCalls  int
	rateLimitCalls     int
	lastDeductAmount   float64
	lastRateLimitCost  float64
	lastSubscriptionID struct {
		userID  int64
		groupID int64
		costUSD float64
	}
}

func (s *usageBillingCacheStub) GetUserBalance(context.Context, int64) (float64, error) {
	return 0, nil
}

func (s *usageBillingCacheStub) SetUserBalance(context.Context, int64, float64) error {
	return nil
}

func (s *usageBillingCacheStub) DeductUserBalance(_ context.Context, _ int64, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deductCalls++
	s.lastDeductAmount = amount
	return nil
}

func (s *usageBillingCacheStub) InvalidateUserBalance(context.Context, int64) error {
	return nil
}

func (s *usageBillingCacheStub) GetSubscriptionCache(context.Context, int64, int64) (*SubscriptionCacheData, error) {
	return nil, nil
}

func (s *usageBillingCacheStub) SetSubscriptionCache(context.Context, int64, int64, *SubscriptionCacheData) error {
	return nil
}

func (s *usageBillingCacheStub) UpdateSubscriptionUsage(_ context.Context, userID, groupID int64, cost float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptionCalls++
	s.lastSubscriptionID.userID = userID
	s.lastSubscriptionID.groupID = groupID
	s.lastSubscriptionID.costUSD = cost
	return nil
}

func (s *usageBillingCacheStub) InvalidateSubscriptionCache(context.Context, int64, int64) error {
	return nil
}

func (s *usageBillingCacheStub) GetAPIKeyRateLimit(context.Context, int64) (*APIKeyRateLimitCacheData, error) {
	return nil, nil
}

func (s *usageBillingCacheStub) SetAPIKeyRateLimit(context.Context, int64, *APIKeyRateLimitCacheData) error {
	return nil
}

func (s *usageBillingCacheStub) UpdateAPIKeyRateLimitUsage(_ context.Context, _ int64, cost float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rateLimitCalls++
	s.lastRateLimitCost = cost
	return nil
}

func (s *usageBillingCacheStub) InvalidateAPIKeyRateLimit(context.Context, int64) error {
	return nil
}

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

func TestBuildUsageBillingCommand_SubscriptionUsesActualCost(t *testing.T) {
	groupID := int64(7)
	subscriptionID := int64(42)

	tests := []struct {
		name               string
		totalCost          float64
		actualCost         float64
		isSubscriptionBill bool
		wantSubscription   float64
		wantBalance        float64
	}{
		{
			name:               "subscription honours 2x multiplier",
			totalCost:          1.0,
			actualCost:         2.0,
			isSubscriptionBill: true,
			wantSubscription:   2.0,
		},
		{
			name:               "subscription honours 0.5x multiplier",
			totalCost:          1.0,
			actualCost:         0.5,
			isSubscriptionBill: true,
			wantSubscription:   0.5,
		},
		{
			name:               "free subscription consumes no quota",
			totalCost:          1.0,
			actualCost:         0.0,
			isSubscriptionBill: true,
		},
		{
			name:               "balance billing still uses actual cost",
			totalCost:          1.0,
			actualCost:         2.0,
			isSubscriptionBill: false,
			wantBalance:        2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildUsageBillingCommand("req-1", nil, &postUsageBillingParams{
				Cost:               &CostBreakdown{TotalCost: tt.totalCost, ActualCost: tt.actualCost},
				User:               &User{ID: 1},
				APIKey:             &APIKey{ID: 2, GroupID: &groupID},
				Account:            &Account{ID: 3},
				Subscription:       &UserSubscription{ID: subscriptionID},
				IsSubscriptionBill: tt.isSubscriptionBill,
			})

			require.NotNil(t, cmd)
			require.InDelta(t, tt.wantSubscription, cmd.SubscriptionCost, 1e-12)
			require.InDelta(t, tt.wantBalance, cmd.BalanceCost, 1e-12)
		})
	}
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

func TestPostUsageBilling_QueuesBalanceAndRateLimitCacheExactlyOnce(t *testing.T) {
	cache := &usageBillingCacheStub{}
	billingCache := NewBillingCacheService(cache, nil, nil, nil, &config.Config{})
	defer billingCache.Stop()

	userRepo := &openAIRecordUsageUserRepoStub{}
	apiKeyQuota := &openAIRecordUsageAPIKeyQuotaStub{}
	deferred := &DeferredService{}

	postUsageBilling(context.Background(), &postUsageBillingParams{
		Cost: &CostBreakdown{
			ActualCost: 4.5,
			TotalCost:  4.5,
		},
		APIKey: &APIKey{
			ID:          101,
			Quota:       100,
			RateLimit5h: 10,
		},
		User:          &User{ID: 202},
		Account:       &Account{ID: 303, Type: AccountTypeOAuth},
		APIKeyService: apiKeyQuota,
	}, &billingDeps{
		userRepo:            userRepo,
		billingCacheService: billingCache,
		deferredService:     deferred,
	})

	require.Equal(t, 1, userRepo.deductCalls)
	require.Equal(t, 1, apiKeyQuota.quotaCalls)
	require.Equal(t, 1, apiKeyQuota.rateLimitCalls)
	require.Eventually(t, func() bool {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		return cache.deductCalls == 1 && cache.rateLimitCalls == 1
	}, 2*time.Second, 10*time.Millisecond)

	cache.mu.Lock()
	require.Equal(t, 4.5, cache.lastDeductAmount)
	require.Equal(t, 4.5, cache.lastRateLimitCost)
	cache.mu.Unlock()

	_, ok := deferred.lastUsedUpdates.Load(int64(303))
	require.True(t, ok)
}

func TestPostUsageBilling_QueuesSubscriptionCacheExactlyOnce(t *testing.T) {
	cache := &usageBillingCacheStub{}
	billingCache := NewBillingCacheService(cache, nil, nil, nil, &config.Config{})
	defer billingCache.Stop()

	groupID := int64(7)
	subRepo := &openAIRecordUsageSubRepoStub{}
	deferred := &DeferredService{}

	postUsageBilling(context.Background(), &postUsageBillingParams{
		Cost: &CostBreakdown{
			TotalCost:  6.25,
			ActualCost: 3.125,
		},
		IsSubscriptionBill: true,
		Subscription:       &UserSubscription{ID: 808},
		APIKey:             &APIKey{ID: 101, GroupID: &groupID},
		User:               &User{ID: 202},
		Account:            &Account{ID: 303, Type: AccountTypeOAuth},
	}, &billingDeps{
		userSubRepo:         subRepo,
		billingCacheService: billingCache,
		deferredService:     deferred,
	})

	require.Equal(t, 1, subRepo.incrementCalls)
	require.Eventually(t, func() bool {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		return cache.subscriptionCalls == 1
	}, 2*time.Second, 10*time.Millisecond)

	cache.mu.Lock()
	require.Equal(t, int64(202), cache.lastSubscriptionID.userID)
	require.Equal(t, int64(7), cache.lastSubscriptionID.groupID)
	require.Equal(t, 3.125, cache.lastSubscriptionID.costUSD)
	cache.mu.Unlock()

	_, ok := deferred.lastUsedUpdates.Load(int64(303))
	require.True(t, ok)
}
