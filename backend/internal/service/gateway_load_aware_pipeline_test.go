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

func TestCollectExcludedAccountIDs_ReturnsSortedIDs(t *testing.T) {
	excludedIDs := map[int64]struct{}{
		9: {},
		2: {},
		5: {},
	}

	ids := collectExcludedAccountIDs(excludedIDs)

	require.Equal(t, []int64{2, 5, 9}, ids)
}

func TestSelectAccountWithLoadBatch_PrefersRoutedSelection(t *testing.T) {
	cache := &mockGatewayCacheForPlatform{}
	concurrencyCache := &mockConcurrencyCache{
		loadMap: map[int64]*AccountLoadInfo{
			1: {AccountID: 1, LoadRate: 90},
			2: {AccountID: 2, LoadRate: 10},
		},
	}
	svc := &GatewayService{
		cache:              cache,
		cfg:                testConfig(),
		concurrencyService: NewConcurrencyService(concurrencyCache),
	}

	state := &loadAwareSchedulingState{
		platform:    PlatformAnthropic,
		preferOAuth: false,
		accounts: []Account{
			{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 5, Priority: 1},
			{ID: 2, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 5, Priority: 1},
		},
		useMixed: false,
		accountByID: map[int64]*Account{
			1: {ID: 1, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 5, Priority: 1},
			2: {ID: 2, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 5, Priority: 1},
		},
		routingAccountIDs: []int64{1, 2},
	}

	result, err := svc.selectAccountWithLoadBatch(&loadAwareBatchSelectionInput{
		ctx:              context.Background(),
		sessionHash:      "route-batch",
		requestedModel:   "claude-3-5-sonnet-20241022",
		excludedIDs:      map[int64]struct{}{},
		schedulingConfig: config.GatewaySchedulingConfig{StickySessionWaitTimeout: time.Second, StickySessionMaxWaiting: 1},
		schedulingState:  state,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, int64(2), result.Account.ID)
	require.Equal(t, int64(2), cache.sessionBindings["route-batch"])
}

func TestSelectAccountWithLoadBatch_FallsBackWhenRoutedSelectionMisses(t *testing.T) {
	cache := &mockGatewayCacheForPlatform{}
	concurrencyCache := &mockConcurrencyCache{
		loadBatchErr:   errors.New("boom"),
		acquireResults: map[int64]bool{7: true},
	}
	svc := &GatewayService{
		cache:              cache,
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
	state := &loadAwareSchedulingState{
		platform:          PlatformAnthropic,
		preferOAuth:       false,
		accounts:          []Account{account},
		useMixed:          false,
		accountByID:       map[int64]*Account{account.ID: &account},
		routingAccountIDs: nil,
	}

	result, err := svc.selectAccountWithLoadBatch(&loadAwareBatchSelectionInput{
		ctx:            context.Background(),
		sessionHash:    "fallback-batch",
		requestedModel: "claude-3-5-sonnet-20241022",
		excludedIDs:    map[int64]struct{}{},
		schedulingConfig: config.GatewaySchedulingConfig{
			FallbackWaitTimeout: time.Second,
			FallbackMaxWaiting:  10,
		},
		schedulingState: state,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, account.ID, result.Account.ID)
	require.True(t, result.Acquired)
}
