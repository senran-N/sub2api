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
	"github.com/senran-N/sub2api/internal/pkg/grok"
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

	err := svc.testCompatibleGatewayAPIKeyConnection(ctx, account, "gpt-5.4", "")
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

	err := svc.testCompatibleGatewayAPIKeyConnection(ctx, account, "gpt-5.4", "")
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

	err := svc.testCompatibleGatewayAPIKeyConnection(ctx, account, "gpt-5.4", "")
	require.Error(t, err)
	require.Equal(t, int64(87), repo.setErrorID)
	require.Equal(t, `Authentication failed (401): {"detail":"Unauthorized"}`, repo.setErrorMsg)
}

func TestAccountTestService_GrokAPIKeyUsesCompatibleProbeDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader("data: [DONE]\n\n"))

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	account := &Account{
		ID:          86,
		Platform:    PlatformGrok,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "xai-test-key",
		},
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "basic",
					"source":     "manual",
				},
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining": 9,
						"total":     20,
					},
				},
				"capabilities": map[string]any{
					"video": false,
				},
				"sync_state": map[string]any{
					"last_sync_at": "2026-04-19T00:00:00Z",
				},
			},
		},
	}
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}

	err := svc.TestAccountConnection(ctx, account.ID, "", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://api.x.ai/v1/responses", upstream.requests[0].URL.String())
	require.Equal(t, "Bearer xai-test-key", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, grok.DefaultTestModel, requestBodyModel(t, upstream.requests[0]))
	require.NotEmpty(t, repo.updatedExtra)

	grokExtra := grokExtraMap(repo.updatedExtra)
	require.Equal(t, AccountTypeAPIKey, grokExtra["auth_mode"])
	require.Equal(t, "basic", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "manual", getNestedGrokValue(grokExtra, "tier", "source"))
	require.Equal(t, "2026-04-19T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.NotEmpty(t, getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.NotEmpty(t, getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.Equal(t, 200, grokParseInt(getNestedGrokValue(grokExtra, "sync_state", "last_probe_status_code")))
	_, hasProbeError := grokNestedMap(grokExtra["sync_state"])["last_probe_error"]
	require.False(t, hasProbeError)

	quotaWindows := grokQuotaWindowsMap(grokExtra["quota_windows"])
	require.Equal(t, 9, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "remaining")))
	require.Equal(t, 20, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "total")))

	capabilities := grokNestedMap(grokExtra["capabilities"])
	require.Equal(t, false, capabilities["video"])
	require.ElementsMatch(t, []string{"chat", "image"}, grokParseStringSlice(capabilities["operations"]))
	require.ElementsMatch(t, []string{"grok-2-image", "grok-3", "grok-3-fast"}, grokParseStringSlice(capabilities["models"]))
}

func TestAccountTestService_GrokAPIKeyUsesConfiguredRuntimeOfficialBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader("data: [DONE]\n\n"))

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	account := &Account{
		ID:          860,
		Platform:    PlatformGrok,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "xai-test-key",
		},
	}
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		settingService: NewSettingService(&grokRuntimeSettingRepoStub{values: map[string]string{
			SettingKeyGrokOfficialBaseURL: "https://official.grok.example/api",
		}}, nil),
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}

	err := svc.TestAccountConnection(ctx, account.ID, "", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://official.grok.example/api/v1/responses", upstream.requests[0].URL.String())
}

func TestAccountTestService_GrokSessionProbeUsesProviderOwnedTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, recorder := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, `{"conversationId":"conv_123"}`)

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	account := &Account{
		ID:          85,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
	}
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
	}

	err := svc.TestAccountConnection(ctx, account.ID, "", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://grok.com/rest/app-chat/conversations/new", upstream.requests[0].URL.String())
	require.Equal(t, "sso=grok-session-token; sso-rw=grok-session-token", upstream.requests[0].Header.Get("Cookie"))
	require.Equal(t, grokSessionModeExpert, requestBodyModeID(t, upstream.requests[0]))
	require.Contains(t, recorder.Body.String(), "test_complete")
	require.NotEmpty(t, repo.updatedExtra)

	grokExtra := grokExtraMap(repo.updatedExtra)
	require.Equal(t, AccountTypeSession, grokExtra["auth_mode"])
	require.NotEmpty(t, getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.NotEmpty(t, getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.Equal(t, 200, grokParseInt(getNestedGrokValue(grokExtra, "sync_state", "last_probe_status_code")))

	capabilities := grokNestedMap(grokExtra["capabilities"])
	require.ElementsMatch(t, []string{"chat", "image", "image_edit", "video", "voice"}, grokParseStringSlice(capabilities["operations"]))
	require.ElementsMatch(t, []string{"grok-2-image", "grok-3", "grok-3-fast", "grok-4-fast-reasoning", "grok-4-voice", "grok-imagine-image", "grok-imagine-image-edit", "grok-imagine-image-pro", "grok-imagine-video"}, grokParseStringSlice(capabilities["models"]))
}

func TestAccountTestService_GrokSessionProbeUsesConfiguredRuntimeSessionBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, `{"conversationId":"conv_123"}`)
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	account := &Account{
		ID:          851,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
	}
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		settingService: NewSettingService(&grokRuntimeSettingRepoStub{values: map[string]string{
			SettingKeyGrokSessionBaseURL: "https://session.grok.example/root",
		}}, nil),
	}

	err := svc.TestAccountConnection(ctx, account.ID, "", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://session.grok.example/root/rest/app-chat/conversations/new", upstream.requests[0].URL.String())
	require.Equal(t, "https://session.grok.example/root", upstream.requests[0].Header.Get("Origin"))
	require.Equal(t, "https://session.grok.example/root/", upstream.requests[0].Header.Get("Referer"))
}

func TestAccountTestService_GrokSessionUnknownTierBootstrapsHighTierProbeBeforeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, recorder := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, `{"conversationId":"conv_123"}`)
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	account := &Account{
		ID:          185,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
	}
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
	}

	err := svc.TestAccountConnection(ctx, account.ID, "", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, grokSessionModeExpert, requestBodyModeID(t, upstream.requests[0]))
	require.Contains(t, recorder.Body.String(), "test_complete")

	grokExtra := grokExtraMap(repo.updatedExtra)
	require.ElementsMatch(t, []string{"chat", "image", "image_edit", "video", "voice"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, []string{"grok-2-image", "grok-3", "grok-3-fast", "grok-4-fast-reasoning", "grok-4-voice", "grok-imagine-image", "grok-imagine-image-edit", "grok-imagine-image-pro", "grok-imagine-video"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
}

func TestAccountTestService_GrokSessionUnknownTierFallsBackToDefaultProbeModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	upstream := &queuedHTTPUpstream{responses: []*http.Response{
		newJSONResponse(http.StatusForbidden, `{"error":"tier required"}`),
		newJSONResponse(http.StatusOK, `{"conversationId":"conv_123"}`),
	}}
	account := &Account{
		ID:          186,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
	}
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
	}

	err := svc.TestAccountConnection(ctx, account.ID, "", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, grokSessionModeExpert, requestBodyModeID(t, upstream.requests[0]))
	require.Equal(t, grokSessionModeAuto, requestBodyModeID(t, upstream.requests[1]))

	grokExtra := grokExtraMap(repo.updatedExtra)
	require.ElementsMatch(t, []string{"chat"}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "operations")))
	require.ElementsMatch(t, []string{grok.DefaultTestModel}, grokParseStringSlice(getNestedGrokValue(grokExtra, "capabilities", "models")))
}

func TestAccountTestService_GrokProbeFailurePersistsNormalizedState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusUnauthorized, `{"detail":"Unauthorized"}`)

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	account := &Account{
		ID:          84,
		Platform:    PlatformGrok,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "xai-test-key",
		},
		Extra: map[string]any{
			"grok": map[string]any{
				"tier": map[string]any{
					"normalized": "super",
					"source":     "sync",
				},
				"quota_windows": map[string]any{
					grok.QuotaWindowAuto: map[string]any{
						"remaining": 7,
						"total":     50,
					},
				},
				"capabilities": map[string]any{
					"models": []any{"grok-imagine-video"},
				},
				"sync_state": map[string]any{
					"last_sync_at":     "2026-04-19T00:00:00Z",
					"last_probe_ok_at": "2026-04-18T00:00:00Z",
				},
			},
		},
	}
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}

	err := svc.TestAccountConnection(ctx, account.ID, "grok-3-fast", "")
	require.Error(t, err)
	require.Equal(t, int64(84), repo.setErrorID)
	require.Contains(t, repo.setErrorMsg, "Authentication failed (401)")
	require.NotEmpty(t, repo.updatedExtra)

	grokExtra := grokExtraMap(repo.updatedExtra)
	require.Equal(t, AccountTypeAPIKey, grokExtra["auth_mode"])
	require.Equal(t, "super", getNestedGrokValue(grokExtra, "tier", "normalized"))
	require.Equal(t, "sync", getNestedGrokValue(grokExtra, "tier", "source"))
	require.Equal(t, "2026-04-19T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_sync_at"))
	require.Equal(t, "2026-04-18T00:00:00Z", getNestedGrokValue(grokExtra, "sync_state", "last_probe_ok_at"))
	require.NotEmpty(t, getNestedGrokValue(grokExtra, "sync_state", "last_probe_at"))
	require.NotEmpty(t, getNestedGrokValue(grokExtra, "sync_state", "last_probe_error_at"))
	require.Equal(t, 401, grokParseInt(getNestedGrokValue(grokExtra, "sync_state", "last_probe_status_code")))
	require.Contains(t, getStringFromMaps(grokNestedMap(grokExtra["sync_state"]), nil, "last_probe_error"), "API returned 401")

	quotaWindows := grokQuotaWindowsMap(grokExtra["quota_windows"])
	require.Equal(t, 7, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "remaining")))
	require.Equal(t, 50, grokParseInt(getNestedGrokValue(quotaWindows, grok.QuotaWindowAuto, "total")))

	capabilities := grokNestedMap(grokExtra["capabilities"])
	require.ElementsMatch(t, []string{"grok-imagine-video"}, grokParseStringSlice(capabilities["models"]))
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

	err := svc.testCompatibleGatewayAPIKeyConnection(ctx, account, "gpt-5.4", "")
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

	err := svc.testCompatibleGatewayAPIKeyConnection(ctx, account, "gpt-5.4-mini", "")
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

	err := svc.testCompatibleGatewayAPIKeyConnection(ctx, account, "gpt-5.4-xhigh", "")
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
		Model     string `json:"model"`
		ModelName string `json:"modelName"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	if payload.Model != "" {
		return payload.Model
	}
	return payload.ModelName
}

func requestBodyModeID(t *testing.T, req *http.Request) string {
	t.Helper()
	require.NotNil(t, req)
	require.NotNil(t, req.Body)

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	var payload struct {
		ModeID string `json:"modeId"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload.ModeID
}
