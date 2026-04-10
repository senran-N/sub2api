//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTrySelectLoadAwareRoutedAccount_SelectsBestRoutedCandidate(t *testing.T) {
	cache := &mockGatewayCacheForPlatform{}
	concurrencyCache := &mockConcurrencyCache{
		loadMap: map[int64]*AccountLoadInfo{
			1: {AccountID: 1, LoadRate: 80},
			2: {AccountID: 2, LoadRate: 20},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			1: {
				ID:          1,
				Platform:    PlatformAnthropic,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Concurrency: 5,
				Priority:    1,
			},
			2: {
				ID:          2,
				Platform:    PlatformAnthropic,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Concurrency: 5,
				Priority:    1,
			},
		},
	}
	svc := &GatewayService{
		accountRepo:        repo,
		cache:              cache,
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(concurrencyCache),
	}

	result, ok := svc.trySelectLoadAwareRoutedAccount(&loadAwareRoutedSelectionInput{
		ctx:            context.Background(),
		sessionHash:    "route-session",
		requestedModel: "claude-3-5-sonnet-20241022",
		excludedIDs:    map[int64]struct{}{},
		accountByID:    map[int64]*Account{},
		plan: &gatewaySelectionPlan{
			platform:          PlatformAnthropic,
			useMixed:          false,
			routingAccountIDs: []int64{1, 2},
		},
		waitTimeout: time.Second,
		maxWaiting:  1,
	})

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, int64(2), result.Account.ID)
	require.Equal(t, int64(2), cache.sessionBindings["route-session"])
	require.Equal(t, 1, concurrencyCache.loadBatchCalls)
	require.Equal(t, 2, repo.getByIDCalls)
}

func TestTrySelectLoadAwareRoutedAccount_UsesStickyAccountWithoutPreloadedMap(t *testing.T) {
	repo := &mockAccountRepoForPlatform{
		accountsByID: map[int64]*Account{
			7: {
				ID:          7,
				Platform:    PlatformAnthropic,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Concurrency: 5,
				Priority:    1,
			},
		},
	}
	svc := &GatewayService{
		accountRepo:        repo,
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(&mockConcurrencyCache{}),
	}

	result, ok := svc.trySelectLoadAwareRoutedAccount(&loadAwareRoutedSelectionInput{
		ctx:             context.Background(),
		sessionHash:     "sticky-route",
		requestedModel:  "claude-3-5-sonnet-20241022",
		stickyAccountID: 7,
		excludedIDs:     map[int64]struct{}{},
		accountByID:     map[int64]*Account{},
		plan: &gatewaySelectionPlan{
			platform:          PlatformAnthropic,
			useMixed:          false,
			routingAccountIDs: []int64{7},
		},
		waitTimeout: time.Second,
		maxWaiting:  1,
	})

	require.True(t, ok)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, int64(7), result.Account.ID)
	require.Equal(t, 1, repo.getByIDCalls)
}

func TestTrySelectLoadAwareRoutedAccount_ReturnsFalseWithoutRoutingIDs(t *testing.T) {
	svc := &GatewayService{
		concurrencyService: NewConcurrencyService(&mockConcurrencyCache{}),
	}

	result, ok := svc.trySelectLoadAwareRoutedAccount(&loadAwareRoutedSelectionInput{
		ctx:  context.Background(),
		plan: &gatewaySelectionPlan{},
	})

	require.False(t, ok)
	require.Nil(t, result)
}
