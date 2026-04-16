//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type openAIAccountTestRepo struct {
	mockAccountRepoForGemini
	updatedExtra  map[string]any
	rateLimitedID int64
	rateLimitedAt *time.Time
	setErrorID    int64
	setErrorMsg   string
}

func (r *openAIAccountTestRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.updatedExtra = updates
	return nil
}

func (r *openAIAccountTestRepo) SetRateLimited(_ context.Context, id int64, resetAt time.Time) error {
	r.rateLimitedID = id
	r.rateLimitedAt = &resetAt
	return nil
}

func (r *openAIAccountTestRepo) SetError(_ context.Context, id int64, errorMsg string) error {
	r.setErrorID = id
	r.setErrorMsg = errorMsg
	return nil
}

func TestAccountTestService_OpenAISuccessPersistsSnapshotFromHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, recorder := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader(`data: {"type":"response.completed"}

`))
	resp.Header.Set("x-codex-primary-used-percent", "88")
	resp.Header.Set("x-codex-primary-reset-after-seconds", "604800")
	resp.Header.Set("x-codex-primary-window-minutes", "10080")
	resp.Header.Set("x-codex-secondary-used-percent", "42")
	resp.Header.Set("x-codex-secondary-reset-after-seconds", "18000")
	resp.Header.Set("x-codex-secondary-window-minutes", "300")

	repo := &openAIAccountTestRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
	account := &Account{
		ID:          89,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4", "")
	require.NoError(t, err)
	require.NotEmpty(t, repo.updatedExtra)
	require.Equal(t, 42.0, repo.updatedExtra["codex_5h_used_percent"])
	require.Equal(t, 88.0, repo.updatedExtra["codex_7d_used_percent"])
	require.Contains(t, recorder.Body.String(), "test_complete")
}

func TestAccountTestService_OpenAI429PersistsSnapshotWithoutRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusTooManyRequests, `{"error":{"type":"usage_limit_reached","message":"limit reached"}}`)
	resp.Header.Set("x-codex-primary-used-percent", "100")
	resp.Header.Set("x-codex-primary-reset-after-seconds", "604800")
	resp.Header.Set("x-codex-primary-window-minutes", "10080")
	resp.Header.Set("x-codex-secondary-used-percent", "100")
	resp.Header.Set("x-codex-secondary-reset-after-seconds", "18000")
	resp.Header.Set("x-codex-secondary-window-minutes", "300")

	repo := &openAIAccountTestRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
	account := &Account{
		ID:          88,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4", "")
	require.Error(t, err)
	require.NotEmpty(t, repo.updatedExtra)
	require.Equal(t, 100.0, repo.updatedExtra["codex_5h_used_percent"])
	require.Zero(t, repo.rateLimitedID)
	require.Nil(t, repo.rateLimitedAt)
	require.Nil(t, account.RateLimitResetAt)
}

func TestAccountTestService_OpenAI401MarksAccountAsAuthFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusUnauthorized, `{"detail":"Unauthorized"}`)

	repo := &openAIAccountTestRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
	account := &Account{
		ID:          87,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4", "")
	require.Error(t, err)
	require.Equal(t, int64(87), repo.setErrorID)
	require.Equal(t, `Authentication failed (401): {"detail":"Unauthorized"}`, repo.setErrorMsg)
}

func TestAccountTestService_OpenAIAzureAPIKeyUsesResponsesEndpointAndAPIKeyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader("data: [DONE]\n\n"))

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}
	account := &Account{
		ID:          86,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "azure-key",
			"base_url": "https://demo.cognitiveservices.azure.com/openai?api-version=2025-04-01-preview",
		},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://demo.cognitiveservices.azure.com/openai/responses?api-version=2025-04-01-preview", upstream.requests[0].URL.String())
	require.Equal(t, "azure-key", upstream.requests[0].Header.Get("Api-Key"))
	require.Empty(t, upstream.requests[0].Header.Get("Authorization"))
}

func TestAccountTestService_OpenAIAPIKeyProbeAppliesWildcardModelMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader("data: [DONE]\n\n"))

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}
	account := &Account{
		ID:          85,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "test-key",
			"model_mapping": map[string]any{
				"gpt-5.4*": "gpt-5.3-codex",
			},
		},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4-mini", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "gpt-5.3-codex", requestBodyModel(t, upstream.requests[0]))
}

func TestAccountTestService_OpenAIAPIKeyProbeUsesReasoningVariantBaseMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader("data: [DONE]\n\n"))

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}
	account := &Account{
		ID:          84,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "test-key",
			"model_mapping": map[string]any{
				"gpt-5.4": "gpt-5.3-codex",
			},
		},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4-xhigh", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "gpt-5.3-codex-xhigh", requestBodyModel(t, upstream.requests[0]))
}

func requestBodyModel(t *testing.T, req *http.Request) string {
	t.Helper()
	require.NotNil(t, req)
	require.NotNil(t, req.Body)

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	var payload struct {
		Model string `json:"model"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload.Model
}
