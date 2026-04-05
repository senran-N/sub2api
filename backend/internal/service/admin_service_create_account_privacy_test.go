//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type createAccountPrivacyRepoStub struct {
	accountRepoStub
	createdAccounts []*Account
	updateExtraCh   chan map[string]any
}

func (s *createAccountPrivacyRepoStub) Create(_ context.Context, account *Account) error {
	account.ID = 123
	s.createdAccounts = append(s.createdAccounts, account)
	return nil
}

func (s *createAccountPrivacyRepoStub) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	if s.updateExtraCh != nil {
		select {
		case s.updateExtraCh <- updates:
		default:
		}
	}
	return nil
}

func TestAdminService_CreateAccount_OpenAIOAuthTriggersPrivacyEnsure(t *testing.T) {
	repo := &createAccountPrivacyRepoStub{
		updateExtraCh: make(chan map[string]any, 1),
	}
	svc := &adminServiceImpl{
		accountRepo: repo,
		privacyClientFactory: func(proxyURL string) (*req.Client, error) {
			return nil, errors.New("privacy client unavailable")
		},
	}

	account, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                  "openai-oauth",
		Platform:              PlatformOpenAI,
		Type:                  AccountTypeOAuth,
		Concurrency:           1,
		Credentials:           map[string]any{"access_token": "token-1"},
		SkipDefaultGroupBind:  true,
		SkipMixedChannelCheck: true,
	})
	require.NoError(t, err)
	require.NotNil(t, account)

	select {
	case updates := <-repo.updateExtraCh:
		require.Equal(t, PrivacyModeFailed, updates["privacy_mode"])
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected CreateAccount to trigger OpenAI privacy persistence")
	}
}
