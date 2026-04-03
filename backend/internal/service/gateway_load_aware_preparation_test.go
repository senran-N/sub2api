//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestResolveStickyAccountID_PrefetchSkipsCacheLookup(t *testing.T) {
	cache := &stickyGatewayCacheHotpathStub{stickyID: 99}
	svc := &GatewayService{cache: cache}

	ctx := context.WithValue(context.Background(), ctxkey.PrefetchedStickyAccountID, int64(42))
	ctx = context.WithValue(ctx, ctxkey.PrefetchedStickyGroupID, int64(0))

	stickyAccountID := svc.resolveStickyAccountID(ctx, nil, "session-1")

	require.Equal(t, int64(42), stickyAccountID)
	require.Equal(t, int64(0), cache.getCalls.Load())
}

func TestResolveStickyAccountID_FallsBackToCacheLookup(t *testing.T) {
	cache := &stickyGatewayCacheHotpathStub{stickyID: 99}
	svc := &GatewayService{cache: cache}

	stickyAccountID := svc.resolveStickyAccountID(context.Background(), nil, "session-1")

	require.Equal(t, int64(99), stickyAccountID)
	require.Equal(t, int64(1), cache.getCalls.Load())
}

func TestPrepareLoadAwareSelectionScope_UsesResolvedFallbackGroup(t *testing.T) {
	groupID := int64(10)
	fallbackID := int64(11)
	groupRepo := &mockGroupRepoForGateway{
		groups: map[int64]*Group{
			groupID: {
				ID:              groupID,
				Platform:        PlatformAnthropic,
				Status:          StatusActive,
				ClaudeCodeOnly:  true,
				FallbackGroupID: &fallbackID,
				Hydrated:        true,
			},
			fallbackID: {
				ID:       fallbackID,
				Platform: PlatformAnthropic,
				Status:   StatusActive,
				Hydrated: true,
			},
		},
	}
	svc := &GatewayService{groupRepo: groupRepo}

	scope, err := svc.prepareLoadAwareSelectionScope(context.Background(), &groupID, "session-1")

	require.NoError(t, err)
	require.NotNil(t, scope)
	require.NotNil(t, scope.groupID)
	require.Equal(t, fallbackID, *scope.groupID)
	require.NotNil(t, scope.group)
	require.Equal(t, fallbackID, scope.group.ID)
	require.Same(t, groupRepo.groups[fallbackID], scope.ctx.Value(ctxkey.Group))
	require.Equal(t, 0, groupRepo.getByIDCalls)
	require.Equal(t, 2, groupRepo.getByIDLiteCalls)
}

func TestPrepareLoadAwareSchedulingState_BuildsRoutingState(t *testing.T) {
	groupID := int64(7)
	group := &Group{
		ID:                  groupID,
		Platform:            PlatformAnthropic,
		Status:              StatusActive,
		Hydrated:            true,
		ModelRoutingEnabled: true,
		ModelRouting: map[string][]int64{
			"claude-sonnet-*": {2},
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{ID: 1, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true, Concurrency: 2},
			{ID: 2, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true, Concurrency: 3},
		},
		accountsByID: map[int64]*Account{},
	}
	for i := range repo.accounts {
		repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
	}

	svc := &GatewayService{
		accountRepo: repo,
		cfg:         testConfig(),
	}

	state, err := svc.prepareLoadAwareSchedulingState(context.Background(), &groupID, group, "claude-sonnet-4-20250514", "session-1", 2)

	require.NoError(t, err)
	require.NotNil(t, state)
	require.Equal(t, PlatformAnthropic, state.platform)
	require.False(t, state.preferOAuth)
	require.True(t, state.useMixed)
	require.Len(t, state.accounts, 2)
	require.Len(t, state.accountByID, 2)
	require.Equal(t, int64(1), state.accountByID[1].ID)
	require.Equal(t, int64(2), state.accountByID[2].ID)
	require.Equal(t, []int64{2}, state.routingAccountIDs)
}
