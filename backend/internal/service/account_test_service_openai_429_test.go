package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type openAI429ProbeRepo struct {
	stubOpenAIAccountRepo
	updatedExtra  map[string]any
	rateLimitedID int64
	rateLimitedAt *time.Time
	clearErrorID  int64
}

func (r *openAI429ProbeRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.updatedExtra = updates
	return nil
}

func (r *openAI429ProbeRepo) SetRateLimited(_ context.Context, id int64, resetAt time.Time) error {
	r.rateLimitedID = id
	r.rateLimitedAt = &resetAt
	return nil
}

func (r *openAI429ProbeRepo) ClearError(_ context.Context, id int64) error {
	r.clearErrorID = id
	return nil
}

func TestAccountTestService_OpenAI429PersistsSnapshotAndRateLimitState_Runtime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusTooManyRequests, `{"error":{"type":"usage_limit_reached","message":"limit reached","resets_at":1777283883}}`)
	defer func() {
		_ = resp.Body.Close()
	}()
	resp.Header.Set("x-codex-primary-used-percent", "100")
	resp.Header.Set("x-codex-primary-reset-after-seconds", "604800")
	resp.Header.Set("x-codex-primary-window-minutes", "10080")
	resp.Header.Set("x-codex-secondary-used-percent", "100")
	resp.Header.Set("x-codex-secondary-reset-after-seconds", "18000")
	resp.Header.Set("x-codex-secondary-window-minutes", "300")

	repo := &openAI429ProbeRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
	account := &Account{
		ID:          88,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusError,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testCompatibleGatewayAPIKeyConnection(ctx, account, "gpt-5.4", "")
	require.Error(t, err)
	require.NotEmpty(t, repo.updatedExtra)
	require.Equal(t, 100.0, repo.updatedExtra["codex_5h_used_percent"])
	require.Equal(t, account.ID, repo.rateLimitedID)
	require.NotNil(t, repo.rateLimitedAt)
	require.Equal(t, account.ID, repo.clearErrorID)
	require.Equal(t, StatusActive, account.Status)
	require.NotNil(t, account.RateLimitResetAt)
}
