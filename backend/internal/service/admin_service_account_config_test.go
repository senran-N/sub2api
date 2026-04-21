//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type accountProxyValidationProxyRepoStub struct {
	proxyRepoStub
	proxies      map[int64]*Proxy
	getByIDCalls []int64
	getByIDErr   error
}

func (s *accountProxyValidationProxyRepoStub) GetByID(_ context.Context, id int64) (*Proxy, error) {
	s.getByIDCalls = append(s.getByIDCalls, id)
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	if proxy, ok := s.proxies[id]; ok {
		return proxy, nil
	}
	return nil, ErrProxyNotFound
}

func TestNormalizeAccountLoadFactor(t *testing.T) {
	t.Run("nil and non-positive values clear load factor", func(t *testing.T) {
		zero := 0
		negative := -3

		gotNil, err := normalizeAccountLoadFactor(nil)
		require.NoError(t, err)
		require.Nil(t, gotNil)

		gotZero, err := normalizeAccountLoadFactor(&zero)
		require.NoError(t, err)
		require.Nil(t, gotZero)

		gotNegative, err := normalizeAccountLoadFactor(&negative)
		require.NoError(t, err)
		require.Nil(t, gotNegative)
	})

	t.Run("positive values are preserved", func(t *testing.T) {
		value := 12

		got, err := normalizeAccountLoadFactor(&value)
		require.NoError(t, err)
		require.Same(t, &value, got)
	})

	t.Run("values above the limit are rejected", func(t *testing.T) {
		value := maxAccountLoadFactor + 1

		got, err := normalizeAccountLoadFactor(&value)
		require.ErrorContains(t, err, "load_factor must be <= 10000")
		require.Nil(t, got)
	})
}

func TestNormalizeAccountExpiresAt(t *testing.T) {
	require.Nil(t, normalizeAccountExpiresAt(nil))

	zero := int64(0)
	require.Nil(t, normalizeAccountExpiresAt(&zero))

	value := int64(1710000000)
	got := normalizeAccountExpiresAt(&value)
	require.NotNil(t, got)
	require.Equal(t, time.Unix(value, 0), *got)
}

func TestApplyAccountProxyID(t *testing.T) {
	originalProxyID := int64(9)
	account := &Account{
		ProxyID: &originalProxyID,
		Proxy:   &Proxy{ID: originalProxyID},
	}

	applyAccountProxyID(account, nil)
	require.NotNil(t, account.ProxyID)
	require.NotNil(t, account.Proxy)

	clearProxyID := int64(0)
	applyAccountProxyID(account, &clearProxyID)
	require.Nil(t, account.ProxyID)
	require.Nil(t, account.Proxy)

	newProxyID := int64(11)
	account.Proxy = &Proxy{ID: newProxyID}
	applyAccountProxyID(account, &newProxyID)
	require.Equal(t, &newProxyID, account.ProxyID)
	require.Nil(t, account.Proxy)
}

func TestAdminServiceCreateAccountRejectsMissingProxyBeforeCreate(t *testing.T) {
	accountRepo := &groupBindingRuleAccountRepoStub{}
	svc := &adminServiceImpl{
		accountRepo: accountRepo,
		proxyRepo:   &accountProxyValidationProxyRepoStub{},
	}
	missingProxyID := int64(999)

	account, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                 "missing-proxy",
		Platform:             PlatformAnthropic,
		Type:                 AccountTypeOAuth,
		Credentials:          map[string]any{},
		ProxyID:              &missingProxyID,
		SkipDefaultGroupBind: true,
	})

	require.Nil(t, account)
	require.ErrorIs(t, err, ErrProxyNotFound)
	require.Zero(t, accountRepo.createCalls)
}

func TestAdminServiceUpdateAccountRejectsMissingProxyBeforeUpdate(t *testing.T) {
	accountRepo := &groupBindingRuleAccountRepoStub{
		getByIDAccount: &Account{
			ID:       7,
			Platform: PlatformAnthropic,
			Type:     AccountTypeOAuth,
		},
	}
	svc := &adminServiceImpl{
		accountRepo: accountRepo,
		proxyRepo:   &accountProxyValidationProxyRepoStub{},
	}
	missingProxyID := int64(999)

	account, err := svc.UpdateAccount(context.Background(), 7, &UpdateAccountInput{
		ProxyID: &missingProxyID,
	})

	require.Nil(t, account)
	require.ErrorIs(t, err, ErrProxyNotFound)
	require.Zero(t, accountRepo.updateCalls)
}
