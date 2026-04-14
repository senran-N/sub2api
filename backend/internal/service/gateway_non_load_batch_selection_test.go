//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type sessionLimitPerAccountStub struct {
	allowedByAccount map[int64]bool
	calls            []int64
}

func (s *sessionLimitPerAccountStub) RegisterSession(ctx context.Context, accountID int64, sessionUUID string, maxSessions int, idleTimeout time.Duration) (bool, error) {
	s.calls = append(s.calls, accountID)
	allowed, ok := s.allowedByAccount[accountID]
	if !ok {
		return true, nil
	}
	return allowed, nil
}

func (s *sessionLimitPerAccountStub) RefreshSession(ctx context.Context, accountID int64, sessionUUID string, idleTimeout time.Duration) error {
	return nil
}

func (s *sessionLimitPerAccountStub) GetActiveSessionCount(ctx context.Context, accountID int64) (int, error) {
	return 0, nil
}

func (s *sessionLimitPerAccountStub) GetActiveSessionCountBatch(ctx context.Context, accountIDs []int64, idleTimeouts map[int64]time.Duration) (map[int64]int, error) {
	return nil, nil
}

func (s *sessionLimitPerAccountStub) IsSessionActive(ctx context.Context, accountID int64, sessionUUID string) (bool, error) {
	return false, nil
}

func (s *sessionLimitPerAccountStub) GetWindowCost(ctx context.Context, accountID int64, windowStart time.Time) (float64, bool, error) {
	return 0, false, nil
}

func (s *sessionLimitPerAccountStub) SetWindowCost(ctx context.Context, accountID int64, windowStart time.Time, cost float64) error {
	return nil
}

func (s *sessionLimitPerAccountStub) GetWindowCostBatch(ctx context.Context, accountWindows map[int64]time.Time) (map[int64]float64, error) {
	return nil, nil
}

func (s *sessionLimitPerAccountStub) ReserveWindowCost(ctx context.Context, accountID int64, windowStart time.Time, reservationID string, cost float64, limit float64, ttl time.Duration) (bool, float64, error) {
	return true, 0, nil
}

func (s *sessionLimitPerAccountStub) ReleaseWindowCost(ctx context.Context, accountID int64, windowStart time.Time, reservationID string) error {
	return nil
}

func TestCloneExcludedAccountIDs_CreatesIndependentCopy(t *testing.T) {
	original := map[int64]struct{}{
		1: {},
		2: {},
	}

	cloned := cloneExcludedAccountIDs(original)
	delete(cloned, 1)
	cloned[3] = struct{}{}

	require.Len(t, original, 2)
	_, hasOriginalOne := original[1]
	require.True(t, hasOriginalOne)
	_, hasOriginalThree := original[3]
	require.False(t, hasOriginalThree)
}

func TestBuildStickyAwareFallbackWaitPlan_UsesStickyWaitConfigForStickyAccount(t *testing.T) {
	cache := &stubConcurrencyCacheForTest{waitCount: 1}
	concurrencyService := NewConcurrencyService(cache)
	account := &Account{ID: 9, Concurrency: 4}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 5 * time.Second,
		StickySessionMaxWaiting:  3,
		FallbackWaitTimeout:      2 * time.Second,
		FallbackMaxWaiting:       10,
	}

	result := buildStickyAwareFallbackWaitPlan(context.Background(), account, account.ID, cfg, concurrencyService)

	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, account.ID, result.WaitPlan.AccountID)
	require.Equal(t, 5*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 3, result.WaitPlan.MaxWaiting)
}

func TestBuildStickyAwareFallbackWaitPlan_FallsBackWhenStickyQueueIsFull(t *testing.T) {
	cache := &stubConcurrencyCacheForTest{waitCount: 3}
	concurrencyService := NewConcurrencyService(cache)
	account := &Account{ID: 9, Concurrency: 4}
	cfg := config.GatewaySchedulingConfig{
		StickySessionWaitTimeout: 5 * time.Second,
		StickySessionMaxWaiting:  3,
		FallbackWaitTimeout:      2 * time.Second,
		FallbackMaxWaiting:       10,
	}

	result := buildStickyAwareFallbackWaitPlan(context.Background(), account, account.ID, cfg, concurrencyService)

	require.NotNil(t, result)
	require.NotNil(t, result.WaitPlan)
	require.Equal(t, account.ID, result.WaitPlan.AccountID)
	require.Equal(t, 2*time.Second, result.WaitPlan.Timeout)
	require.Equal(t, 10, result.WaitPlan.MaxWaiting)
}

func TestSelectAccountWithoutLoadBatch_ClearsRejectedStickyBindingAndBindsReplacement(t *testing.T) {
	cache := &mockGatewayCacheForPlatform{
		sessionBindings: map[string]int64{"sticky-session": 1},
	}
	sessionLimit := &sessionLimitPerAccountStub{
		allowedByAccount: map[int64]bool{
			1: false,
			2: true,
		},
	}
	repo := &mockAccountRepoForPlatform{
		accounts: []Account{
			{
				ID:          1,
				Platform:    PlatformAnthropic,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    1,
				Extra: map[string]any{
					"max_sessions": 1,
				},
			},
			{
				ID:          2,
				Platform:    PlatformAnthropic,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    2,
				Extra: map[string]any{
					"max_sessions": 1,
				},
			},
		},
		accountsByID: map[int64]*Account{},
	}
	for i := range repo.accounts {
		repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
	}
	svc := &GatewayService{
		accountRepo:       repo,
		cache:             cache,
		sessionLimitCache: sessionLimit,
		cfg:               testConfig(),
	}

	result, err := svc.selectAccountWithoutLoadBatch(
		&gatewaySelectionScope{
			ctx:             context.Background(),
			groupID:         nil,
			platform:        PlatformAnthropic,
			useMixed:        true,
			preferOAuth:     false,
			stickyAccountID: 1,
		},
		&gatewaySelectionPlan{
			platform:            PlatformAnthropic,
			useMixed:            true,
			preferOAuth:         false,
			stickyPlatformCheck: func(a *Account) bool { return a.Platform == PlatformAnthropic },
			platformFilter:      func(a *Account) bool { return a.Platform == PlatformAnthropic },
			oauthTieBreaker:     func(a, b *Account) bool { return false },
		},
		"sticky-session",
		"claude-3-5-sonnet-20241022",
		nil,
		config.GatewaySchedulingConfig{},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Account)
	require.Equal(t, int64(2), result.Account.ID)
	require.Equal(t, []int64{1, 2}, sessionLimit.calls)
	require.Equal(t, 1, cache.deletedSessions["sticky-session"])
	require.Equal(t, int64(2), cache.sessionBindings["sticky-session"])
}
