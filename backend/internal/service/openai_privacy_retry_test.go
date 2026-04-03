//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestAdminService_EnsureOpenAIPrivacy_RetriesNonSuccessModes(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{PrivacyModeFailed, PrivacyModeCFBlocked} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			privacyCalls := 0
			svc := &adminServiceImpl{
				accountRepo: &mockAccountRepoForGemini{},
				privacyClientFactory: func(proxyURL string) (*req.Client, error) {
					privacyCalls++
					return nil, errors.New("factory failed")
				},
			}

			account := &Account{
				ID:       101,
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"access_token": "token-1",
				},
				Extra: map[string]any{
					"privacy_mode": mode,
				},
			}

			got := svc.EnsureOpenAIPrivacy(context.Background(), account)

			require.Equal(t, PrivacyModeFailed, got)
			require.Equal(t, 1, privacyCalls)
		})
	}
}

func TestTokenRefreshService_ensureOpenAIPrivacy_RetriesNonSuccessModes(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		TokenRefresh: config.TokenRefreshConfig{
			MaxRetries:          1,
			RetryBackoffSeconds: 0,
		},
	}

	for _, mode := range []string{PrivacyModeFailed, PrivacyModeCFBlocked} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			service := NewTokenRefreshService(&tokenRefreshAccountRepo{}, nil, nil, nil, nil, nil, nil, cfg, nil)
			privacyCalls := 0
			service.SetPrivacyDeps(func(proxyURL string) (*req.Client, error) {
				privacyCalls++
				return nil, errors.New("factory failed")
			}, nil)

			account := &Account{
				ID:       202,
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"access_token": "token-2",
				},
				Extra: map[string]any{
					"privacy_mode": mode,
				},
			}

			service.ensureOpenAIPrivacy(context.Background(), account)

			require.Equal(t, 1, privacyCalls)
		})
	}
}

type privacyTestProxyRepo struct {
	proxy *Proxy
}

func (s *privacyTestProxyRepo) Create(ctx context.Context, proxy *Proxy) error {
	panic("unexpected Create call")
}

func (s *privacyTestProxyRepo) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	if s.proxy == nil || s.proxy.ID != id {
		return nil, errors.New("proxy not found")
	}
	return s.proxy, nil
}

func (s *privacyTestProxyRepo) ListByIDs(ctx context.Context, ids []int64) ([]Proxy, error) {
	panic("unexpected ListByIDs call")
}

func (s *privacyTestProxyRepo) Update(ctx context.Context, proxy *Proxy) error {
	panic("unexpected Update call")
}

func (s *privacyTestProxyRepo) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}

func (s *privacyTestProxyRepo) List(ctx context.Context, params pagination.PaginationParams) ([]Proxy, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *privacyTestProxyRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]Proxy, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *privacyTestProxyRepo) ListWithFiltersAndAccountCount(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]ProxyWithAccountCount, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFiltersAndAccountCount call")
}

func (s *privacyTestProxyRepo) ListActive(ctx context.Context) ([]Proxy, error) {
	panic("unexpected ListActive call")
}

func (s *privacyTestProxyRepo) ListActiveWithAccountCount(ctx context.Context) ([]ProxyWithAccountCount, error) {
	panic("unexpected ListActiveWithAccountCount call")
}

func (s *privacyTestProxyRepo) ExistsByHostPortAuth(ctx context.Context, host string, port int, username, password string) (bool, error) {
	panic("unexpected ExistsByHostPortAuth call")
}

func (s *privacyTestProxyRepo) CountAccountsByProxyID(ctx context.Context, proxyID int64) (int64, error) {
	panic("unexpected CountAccountsByProxyID call")
}

func (s *privacyTestProxyRepo) ListAccountSummariesByProxyID(ctx context.Context, proxyID int64) ([]ProxyAccountSummary, error) {
	panic("unexpected ListAccountSummariesByProxyID call")
}

func TestTokenRefreshService_EnsureOpenAIPrivacy_UsesProxyURL(t *testing.T) {
	cfg := &config.Config{
		TokenRefresh: config.TokenRefreshConfig{
			MaxRetries:          1,
			RetryBackoffSeconds: 0,
		},
	}

	service := NewTokenRefreshService(&tokenRefreshAccountRepo{}, nil, nil, nil, nil, nil, nil, cfg, nil)
	proxyID := int64(88)
	capturedProxyURL := ""
	service.SetPrivacyDeps(func(proxyURL string) (*req.Client, error) {
		capturedProxyURL = proxyURL
		return nil, errors.New("factory failed")
	}, &privacyTestProxyRepo{
		proxy: &Proxy{
			ID:       proxyID,
			Protocol: "http",
			Host:     "127.0.0.1",
			Port:     8080,
		},
	})

	service.ensureOpenAIPrivacy(context.Background(), &Account{
		ID:       301,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		ProxyID:  &proxyID,
		Credentials: map[string]any{
			"access_token": "token-301",
		},
	})

	require.Equal(t, "http://127.0.0.1:8080", capturedProxyURL)
}
