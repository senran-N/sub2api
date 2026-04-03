//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSelectLoadAwareFallbackFlow_StickyHitWithoutRouting(t *testing.T) {
	concurrencyCache := &mockConcurrencyCache{}
	svc := &GatewayService{
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(concurrencyCache),
	}

	account := Account{
		ID:          5,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 4,
	}

	result, err := svc.selectLoadAwareFallbackFlow(&loadAwareFallbackSelectionInput{
		ctx:             context.Background(),
		sessionHash:     "sticky-session",
		requestedModel:  "claude-3-5-sonnet-20241022",
		stickyAccountID: account.ID,
		excludedIDs:     map[int64]struct{}{},
		accountByID: map[int64]*Account{
			account.ID: &account,
		},
		platform:         PlatformAnthropic,
		useMixed:         false,
		accounts:         []Account{account},
		preferOAuth:      false,
		schedulingConfig: config.GatewaySchedulingConfig{StickySessionWaitTimeout: time.Second, StickySessionMaxWaiting: 1},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, account.ID, result.Account.ID)
	require.True(t, result.Acquired)
	require.Equal(t, 0, concurrencyCache.loadBatchCalls)
}

func TestSelectLoadAwareFallbackFlow_LoadBatchErrorFallsBackToLegacyOrder(t *testing.T) {
	concurrencyCache := &mockConcurrencyCache{
		loadBatchErr:   errors.New("boom"),
		acquireResults: map[int64]bool{7: true},
	}
	svc := &GatewayService{
		cache:              &mockGatewayCacheForPlatform{},
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(concurrencyCache),
	}

	account := Account{
		ID:          7,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 4,
		Priority:    1,
	}

	result, err := svc.selectLoadAwareFallbackFlow(&loadAwareFallbackSelectionInput{
		ctx:            context.Background(),
		sessionHash:    "legacy-fallback",
		requestedModel: "claude-3-5-sonnet-20241022",
		excludedIDs:    map[int64]struct{}{},
		accountByID:    map[int64]*Account{account.ID: &account},
		platform:       PlatformAnthropic,
		useMixed:       false,
		accounts:       []Account{account},
		preferOAuth:    false,
		schedulingConfig: config.GatewaySchedulingConfig{
			FallbackWaitTimeout: time.Second,
			FallbackMaxWaiting:  10,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, account.ID, result.Account.ID)
	require.True(t, result.Acquired)
	require.Equal(t, 1, concurrencyCache.loadBatchCalls)
}

func TestSelectLoadAwareFallbackFlow_NoCandidatesReturnsError(t *testing.T) {
	svc := &GatewayService{
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(&mockConcurrencyCache{}),
	}

	account := Account{
		ID:          9,
		Platform:    PlatformAnthropic,
		Status:      StatusDisabled,
		Schedulable: true,
	}

	result, err := svc.selectLoadAwareFallbackFlow(&loadAwareFallbackSelectionInput{
		ctx:            context.Background(),
		requestedModel: "claude-3-5-sonnet-20241022",
		excludedIDs:    map[int64]struct{}{},
		accountByID:    map[int64]*Account{account.ID: &account},
		platform:       PlatformAnthropic,
		useMixed:       false,
		accounts:       []Account{account},
		schedulingConfig: config.GatewaySchedulingConfig{
			FallbackWaitTimeout: time.Second,
			FallbackMaxWaiting:  10,
		},
	})

	require.ErrorIs(t, err, ErrNoAvailableAccounts)
	require.Nil(t, result)
}
