//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type geminiTokenProviderRepoStub struct {
	mockAccountRepoForGemini
	account               *Account
	setTempUnschedCalls   int
	lastTempUnschedID     int64
	lastTempUnschedUntil  time.Time
	lastTempUnschedReason string
}

func (r *geminiTokenProviderRepoStub) GetByID(_ context.Context, _ int64) (*Account, error) {
	return r.account, nil
}

func (r *geminiTokenProviderRepoStub) SetTempUnschedulable(_ context.Context, id int64, until time.Time, reason string) error {
	r.setTempUnschedCalls++
	r.lastTempUnschedID = id
	r.lastTempUnschedUntil = until
	r.lastTempUnschedReason = reason
	return nil
}

type geminiTokenProviderTempUnschedCacheStub struct {
	setCalls      int
	lastAccountID int64
	lastState     *TempUnschedState
}

func (s *geminiTokenProviderTempUnschedCacheStub) SetTempUnsched(_ context.Context, accountID int64, state *TempUnschedState) error {
	s.setCalls++
	s.lastAccountID = accountID
	if state != nil {
		cloned := *state
		s.lastState = &cloned
	}
	return nil
}

func (s *geminiTokenProviderTempUnschedCacheStub) GetTempUnsched(context.Context, int64) (*TempUnschedState, error) {
	return nil, nil
}

func (s *geminiTokenProviderTempUnschedCacheStub) DeleteTempUnsched(context.Context, int64) error {
	return nil
}

func TestGeminiTokenProvider_GetAccessToken_RefreshErrorMarksTempUnschedulable(t *testing.T) {
	expiresAt := time.Now().Add(-time.Minute).Unix()
	account := &Account{
		ID:       42,
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":  "expired-token",
			"refresh_token": "refresh-token",
			"expires_at":    expiresAt,
		},
	}
	repo := &geminiTokenProviderRepoStub{account: account}
	tempCache := &geminiTokenProviderTempUnschedCacheStub{}
	refreshAPI := NewOAuthRefreshAPI(repo, &refreshAPICacheStub{lockResult: true})
	executor := &refreshAPIExecutorStub{
		needsRefresh: true,
		err:          errors.New("refresh failed"),
	}

	provider := NewGeminiTokenProvider(repo, nil, nil)
	provider.SetRefreshAPI(refreshAPI, executor)
	provider.SetRefreshPolicy(GeminiProviderRefreshPolicy())
	provider.SetTempUnschedCache(tempCache)

	token, err := provider.GetAccessToken(context.Background(), account)

	require.Error(t, err)
	require.Empty(t, token)
	require.Equal(t, 1, executor.refreshCalls)
	require.Equal(t, 1, repo.setTempUnschedCalls)
	require.Equal(t, account.ID, repo.lastTempUnschedID)
	require.Contains(t, repo.lastTempUnschedReason, "token refresh failed on request path")
	require.Contains(t, repo.lastTempUnschedReason, "refresh failed")
	require.WithinDuration(t, time.Now().Add(tokenRefreshTempUnschedDuration), repo.lastTempUnschedUntil, 5*time.Second)
	require.Equal(t, 1, tempCache.setCalls)
	require.Equal(t, account.ID, tempCache.lastAccountID)
	require.NotNil(t, tempCache.lastState)
	require.Equal(t, repo.lastTempUnschedReason, tempCache.lastState.ErrorMessage)
	require.InDelta(t, repo.lastTempUnschedUntil.Unix(), tempCache.lastState.UntilUnix, 5)
}

func TestGeminiTokenProvider_GetAccessToken_RefreshErrorWithExistingTokenPolicyDoesNotQuarantine(t *testing.T) {
	expiresAt := time.Now().Add(-time.Minute).Unix()
	account := &Account{
		ID:       43,
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":  "stale-but-usable",
			"refresh_token": "refresh-token",
			"expires_at":    expiresAt,
		},
	}
	repo := &geminiTokenProviderRepoStub{account: account}
	tempCache := &geminiTokenProviderTempUnschedCacheStub{}
	refreshAPI := NewOAuthRefreshAPI(repo, &refreshAPICacheStub{lockResult: true})
	executor := &refreshAPIExecutorStub{
		needsRefresh: true,
		err:          errors.New("refresh failed"),
	}

	provider := NewGeminiTokenProvider(repo, nil, nil)
	provider.SetRefreshAPI(refreshAPI, executor)
	provider.SetRefreshPolicy(ProviderRefreshPolicy{OnRefreshError: ProviderRefreshErrorUseExistingToken})
	provider.SetTempUnschedCache(tempCache)

	token, err := provider.GetAccessToken(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, "stale-but-usable", token)
	require.Equal(t, 0, repo.setTempUnschedCalls)
	require.Equal(t, 0, tempCache.setCalls)
}
