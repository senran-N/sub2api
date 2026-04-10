//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildRoutingSet_IgnoresNonPositiveIDs(t *testing.T) {
	routingSet := buildRoutingSet([]int64{0, -1, 5, 9})

	require.Len(t, routingSet, 2)
	_, hasFive := routingSet[5]
	_, hasNine := routingSet[9]
	require.True(t, hasFive)
	require.True(t, hasNine)
}

func TestTryStickySessionAccount_ReturnsBoundAccount(t *testing.T) {
	ctx := context.Background()
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{ID: 1, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true},
		},
		accountsByID: map[int64]*Account{},
	}
	for i := range repo.accounts {
		repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
	}

	cache := &mockGatewayCacheForPlatform{
		sessionBindings: map[string]int64{"session-1": 1},
	}

	svc := &GatewayService{
		accountRepo: repo,
		cache:       cache,
		cfg:         testConfig(),
	}

	account, ok := svc.tryStickySessionAccount(ctx, nil, "session-1", "", nil, func(a *Account) bool {
		return a.Platform == PlatformAnthropic
	})

	require.True(t, ok)
	require.NotNil(t, account)
	require.Equal(t, int64(1), account.ID)
}

func TestTryStickySessionAccount_ClearsUnschedulableBinding(t *testing.T) {
	ctx := context.Background()
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{ID: 1, Platform: PlatformAnthropic, Status: StatusDisabled, Schedulable: true},
		},
		accountsByID: map[int64]*Account{},
	}
	for i := range repo.accounts {
		repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
	}

	cache := &mockGatewayCacheForPlatform{
		sessionBindings: map[string]int64{"session-1": 1},
	}

	svc := &GatewayService{
		accountRepo: repo,
		cache:       cache,
		cfg:         testConfig(),
	}

	account, ok := svc.tryStickySessionAccount(ctx, nil, "session-1", "", nil, func(a *Account) bool {
		return a.Platform == PlatformAnthropic
	})

	require.False(t, ok)
	require.Nil(t, account)
	require.Equal(t, 1, cache.deletedSessions["session-1"])
}

func TestSelectBestCandidateWithStickyPolicy_BindsSelectedAccount(t *testing.T) {
	ctx := context.Background()
	cache := &mockGatewayCacheForPlatform{}

	svc := &GatewayService{
		cache: cache,
		cfg:   testConfig(),
	}

	accounts := []Account{
		{ID: 1, Platform: PlatformOpenAI, Priority: 2, Status: StatusActive, Schedulable: true},
		{ID: 2, Platform: PlatformOpenAI, Priority: 1, Status: StatusActive, Schedulable: true},
	}

	selected := svc.selectBestCandidateWithStickyPolicy(ctx, nil, "session-1", accounts, &candidateFilterParams{
		ctx:         ctx,
		excludedIDs: map[int64]struct{}{},
	}, nil, true)

	require.NotNil(t, selected)
	require.Equal(t, int64(2), selected.ID)
	require.Equal(t, int64(2), cache.sessionBindings["session-1"])
}

func TestSortAccountsByPriorityLoadAndLastUsed_OrdersByPriorityThenLoadThenLRU(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Minute)
	later := now.Add(-time.Second)

	accounts := []accountWithLoad{
		{account: &Account{ID: 1, Priority: 2, LastUsedAt: &now}, loadInfo: &AccountLoadInfo{AccountID: 1, LoadRate: 10}},
		{account: &Account{ID: 2, Priority: 1, LastUsedAt: &later}, loadInfo: &AccountLoadInfo{AccountID: 2, LoadRate: 20}},
		{account: &Account{ID: 3, Priority: 1, LastUsedAt: &earlier}, loadInfo: &AccountLoadInfo{AccountID: 3, LoadRate: 10}},
	}

	sortAccountsByPriorityLoadAndLastUsed(accounts, false)

	require.Equal(t, int64(3), accounts[0].account.ID)
	require.Equal(t, int64(2), accounts[1].account.ID)
	require.Equal(t, int64(1), accounts[2].account.ID)
}
