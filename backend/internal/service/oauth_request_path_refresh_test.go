package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type requestPathTokenCacheStub struct {
	mu           sync.Mutex
	tokens       map[string]string
	lockAcquired bool
}

func newRequestPathTokenCacheStub() *requestPathTokenCacheStub {
	return &requestPathTokenCacheStub{
		tokens:       make(map[string]string),
		lockAcquired: true,
	}
}

func (s *requestPathTokenCacheStub) GetAccessToken(context.Context, string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tokens["openai:account:100"], nil
}

func (s *requestPathTokenCacheStub) SetAccessToken(_ context.Context, cacheKey string, token string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[cacheKey] = token
	return nil
}

func (s *requestPathTokenCacheStub) DeleteAccessToken(_ context.Context, cacheKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, cacheKey)
	return nil
}

func (s *requestPathTokenCacheStub) AcquireRefreshLock(context.Context, string, time.Duration) (bool, error) {
	return s.lockAcquired, nil
}

func (s *requestPathTokenCacheStub) ReleaseRefreshLock(context.Context, string) error {
	return nil
}

type requestPathAccountRepoStub struct {
	AccountRepository
	account       *Account
	tempCalls     int
	errorCalls    int
	lastTempUntil time.Time
	lastReason    string
}

func (s *requestPathAccountRepoStub) GetByID(context.Context, int64) (*Account, error) {
	return s.account, nil
}

func (s *requestPathAccountRepoStub) SetTempUnschedulable(_ context.Context, _ int64, until time.Time, reason string) error {
	s.tempCalls++
	s.lastTempUntil = until
	s.lastReason = reason
	return nil
}

func (s *requestPathAccountRepoStub) SetError(context.Context, int64, string) error {
	s.errorCalls++
	return nil
}

type requestPathRefreshExecutorStub struct {
	err          error
	refreshCalls int
}

func (s *requestPathRefreshExecutorStub) CanRefresh(*Account) bool {
	return true
}

func (s *requestPathRefreshExecutorStub) NeedsRefresh(*Account, time.Duration) bool {
	return true
}

func (s *requestPathRefreshExecutorStub) Refresh(context.Context, *Account) (map[string]any, error) {
	s.refreshCalls++
	return nil, s.err
}

func (s *requestPathRefreshExecutorStub) CacheKey(account *Account) string {
	return OpenAITokenCacheKey(account)
}

func TestOpenAITokenProvider_RequestPathLockHeldFailsOverWithoutOldToken(t *testing.T) {
	cache := newRequestPathTokenCacheStub()
	cache.lockAcquired = false
	account := &Account{
		ID:       100,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "old-token",
			"expires_at":   time.Now().Add(time.Minute).Format(time.RFC3339),
		},
	}
	executor := &requestPathRefreshExecutorStub{}
	provider := NewOpenAITokenProvider(nil, cache, nil)
	provider.SetRefreshAPI(NewOAuthRefreshAPI(nil, cache), executor)
	provider.SetRequestPathRefreshSettings(OAuthRequestPathRefreshSettings{
		RequestTimeout:               time.Second,
		LockWaitTimeout:              10 * time.Millisecond,
		TransientTempUnschedDuration: time.Minute,
	})

	token, err := provider.GetAccessToken(context.Background(), account)
	if err == nil {
		t.Fatal("expected lock-held failover error")
	}
	if token != "" {
		t.Fatalf("token=%q, want empty", token)
	}
	var failoverErr *UpstreamFailoverError
	if !errors.As(err, &failoverErr) {
		t.Fatalf("error %T does not wrap UpstreamFailoverError", err)
	}
	if failoverErr.FailureReason != "oauth_token_refresh_lock_held" {
		t.Fatalf("failure reason=%q", failoverErr.FailureReason)
	}
	if executor.refreshCalls != 0 {
		t.Fatalf("refresh calls=%d, want 0", executor.refreshCalls)
	}
}

func TestOAuthRefreshAPI_LocalLockContentionReturnsLockHeld(t *testing.T) {
	account := &Account{
		ID:       100,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":  "old-token",
			"refresh_token": "refresh-token",
			"expires_at":    time.Now().Add(time.Minute).Format(time.RFC3339),
		},
	}
	repo := &requestPathAccountRepoStub{account: account}
	executor := &requestPathRefreshExecutorStub{}
	api := NewOAuthRefreshAPI(repo, nil)

	cacheKey := executor.CacheKey(account)
	localMu := api.getLocalLock(cacheKey)
	localMu.Lock()
	defer localMu.Unlock()

	result, err := api.RefreshIfNeeded(context.Background(), account, executor, openAITokenRefreshSkew)
	if err != nil {
		t.Fatalf("RefreshIfNeeded error=%v", err)
	}
	if result == nil || !result.LockHeld {
		t.Fatalf("LockHeld=%v, want true", result != nil && result.LockHeld)
	}
	if executor.refreshCalls != 0 {
		t.Fatalf("refresh calls=%d, want 0", executor.refreshCalls)
	}
}

func TestOpenAITokenProvider_RequestPathRefreshErrorTempUnschedulesAndFailsOver(t *testing.T) {
	cache := newRequestPathTokenCacheStub()
	account := &Account{
		ID:       100,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":  "old-token",
			"refresh_token": "refresh-token",
			"expires_at":    time.Now().Add(time.Minute).Format(time.RFC3339),
		},
	}
	repo := &requestPathAccountRepoStub{account: account}
	executor := &requestPathRefreshExecutorStub{err: errors.New("network timeout")}
	provider := NewOpenAITokenProvider(repo, cache, nil)
	provider.SetRefreshAPI(NewOAuthRefreshAPI(repo, cache), executor)
	provider.SetRequestPathRefreshSettings(OAuthRequestPathRefreshSettings{
		RequestTimeout:               time.Second,
		LockWaitTimeout:              10 * time.Millisecond,
		TransientTempUnschedDuration: time.Minute,
	})

	token, err := provider.GetAccessToken(context.Background(), account)
	if err == nil {
		t.Fatal("expected refresh failure")
	}
	if token != "" {
		t.Fatalf("token=%q, want empty", token)
	}
	var failoverErr *UpstreamFailoverError
	if !errors.As(err, &failoverErr) {
		t.Fatalf("error %T does not wrap UpstreamFailoverError", err)
	}
	if repo.tempCalls != 1 {
		t.Fatalf("temp calls=%d, want 1", repo.tempCalls)
	}
	if repo.errorCalls != 0 {
		t.Fatalf("error calls=%d, want 0", repo.errorCalls)
	}
	if time.Until(repo.lastTempUntil) <= 0 {
		t.Fatalf("temp until=%v, want future", repo.lastTempUntil)
	}
}
