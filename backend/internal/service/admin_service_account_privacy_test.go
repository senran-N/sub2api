//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type privacyAccountRepoStub struct {
	accountRepoStub
	updateExtraErr   error
	updateExtraID    int64
	updateExtraCalls int
	updateExtraValue map[string]any
}

func (s *privacyAccountRepoStub) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	s.updateExtraCalls++
	s.updateExtraID = id
	s.updateExtraValue = make(map[string]any, len(updates))
	for key, value := range updates {
		s.updateExtraValue[key] = value
	}
	return s.updateExtraErr
}

type privacyProxyRepoStub struct {
	proxyRepoStub
	getByIDErr error
	proxies    map[int64]*Proxy
}

func (s *privacyProxyRepoStub) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	return s.proxies[id], nil
}

func TestAdminServiceResolveAccountProxyURL(t *testing.T) {
	proxyID := int64(42)
	svc := &adminServiceImpl{
		proxyRepo: &privacyProxyRepoStub{
			proxies: map[int64]*Proxy{
				proxyID: {
					Protocol: "http",
					Host:     "127.0.0.1",
					Port:     8080,
				},
			},
		},
	}

	require.Equal(t, "http://127.0.0.1:8080", svc.resolveAccountProxyURL(context.Background(), &proxyID))
	require.Equal(t, "", svc.resolveAccountProxyURL(context.Background(), nil))
}

func TestAdminServicePersistAccountPrivacyModeSyncsInMemoryOnSuccess(t *testing.T) {
	repo := &privacyAccountRepoStub{}
	svc := &adminServiceImpl{accountRepo: repo}
	account := &Account{ID: 7}

	got := svc.persistAccountPrivacyMode(
		context.Background(),
		account,
		"strict",
		"force_update_openai_privacy_mode_failed",
		setAccountPrivacyMode,
	)

	require.Equal(t, "strict", got)
	require.Equal(t, 1, repo.updateExtraCalls)
	require.Equal(t, int64(7), repo.updateExtraID)
	require.Equal(t, "strict", repo.updateExtraValue["privacy_mode"])
	require.Equal(t, "strict", account.Extra["privacy_mode"])
}

func TestAdminServicePersistAccountPrivacyModeSkipsInMemorySyncOnFailure(t *testing.T) {
	repo := &privacyAccountRepoStub{updateExtraErr: errors.New("db down")}
	svc := &adminServiceImpl{accountRepo: repo}
	account := &Account{
		ID: 8,
		Extra: map[string]any{
			"existing": "value",
		},
	}

	got := svc.persistAccountPrivacyMode(
		context.Background(),
		account,
		"strict",
		"force_update_openai_privacy_mode_failed",
		setAccountPrivacyMode,
	)

	require.Equal(t, "strict", got)
	require.Equal(t, 1, repo.updateExtraCalls)
	require.Equal(t, "value", account.Extra["existing"])
	_, exists := account.Extra["privacy_mode"]
	require.False(t, exists)
}
