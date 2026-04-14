package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCodexNativeMutationPolicy_ResolveOAuthSessionHeaders_PreservesHeaderIDs(t *testing.T) {
	policy := NewCodexNativeMutationPolicy(CodexRequestProfile{
		Headers: CodexRequestHeaderProfile{
			SessionID:      "sess-cli",
			ConversationID: "conv-cli",
		},
	})

	resolution := policy.ResolveOAuthSessionHeaders("pcache_123", "compact-seed", true)
	require.Equal(t, "sess-cli", resolution.SessionID)
	require.Equal(t, "conv-cli", resolution.ConversationID)
	require.Equal(t, "header_session_id", resolution.SessionSource)
	require.Equal(t, "header_conversation_id", resolution.ConversationSource)
}

func TestCodexNativeMutationPolicy_ResolveSessionHeaders_PrefersHeadersThenPromptCacheKey(t *testing.T) {
	t.Run("conversation header backfills session id", func(t *testing.T) {
		policy := NewCodexNativeMutationPolicy(CodexRequestProfile{
			Headers: CodexRequestHeaderProfile{
				ConversationID: "conv-cli",
			},
		})

		resolution := policy.ResolveSessionHeaders("pcache_123")
		require.Equal(t, "conv-cli", resolution.SessionID)
		require.Equal(t, "conv-cli", resolution.ConversationID)
		require.Equal(t, "header_conversation_id", resolution.SessionSource)
		require.Equal(t, "header_conversation_id", resolution.ConversationSource)
	})

	t.Run("prompt cache key is only used when headers are absent", func(t *testing.T) {
		policy := NewCodexNativeMutationPolicy(CodexRequestProfile{})

		resolution := policy.ResolveSessionHeaders("pcache_123")
		require.Equal(t, "pcache_123", resolution.SessionID)
		require.Empty(t, resolution.ConversationID)
		require.Equal(t, "prompt_cache_key", resolution.SessionSource)
		require.Equal(t, "none", resolution.ConversationSource)
	})
}

func TestCodexNativeMutationPolicy_ResolveRequestSessionID_PrefersHeadersThenBody(t *testing.T) {
	policy := NewCodexNativeMutationPolicy(CodexRequestProfile{
		Headers: CodexRequestHeaderProfile{
			ConversationID: "conv-cli",
		},
		Body: CodexRequestBodyProfile{
			PromptCacheKey: "pcache_123",
		},
	})
	require.Equal(t, "conv-cli", policy.ResolveRequestSessionID())

	policy = NewCodexNativeMutationPolicy(CodexRequestProfile{
		Body: CodexRequestBodyProfile{
			PromptCacheKey: "pcache_123",
		},
	})
	require.Equal(t, "pcache_123", policy.ResolveRequestSessionID())
}

func TestCodexNativeMutationPolicy_ResolveUserAgent_PreservesOfficialCompositeUA(t *testing.T) {
	policy := NewCodexNativeMutationPolicy(CodexRequestProfile{
		OfficialClient: true,
		Headers: CodexRequestHeaderProfile{
			UserAgent: "Mozilla/5.0 codex_cli_rs/0.200.0",
		},
	})

	require.Equal(
		t,
		"Mozilla/5.0 codex_cli_rs/0.200.0",
		policy.ResolveUserAgent(&Account{Type: AccountTypeOAuth}, false, true),
	)
}

func TestOpenAIBuildUpstreamRequestOAuthCompact_UsesStableUpstreamPersona(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", bytes.NewReader([]byte(`{"model":"gpt-5"}`)))
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 codex_cli_rs/0.200.0")
	c.Request.Header.Set("originator", "codex_vscode")
	c.Request.Header.Set("version", "0.200.0")
	c.Request.Header.Set("OpenAI-Beta", "responses=experimental,foo=bar")
	c.Request.Header.Set("session_id", "sess-cli")
	c.Request.Header.Set("conversation_id", "conv-cli")
	c.Request.Header.Set("Accept", "application/json")

	svc := &OpenAIGatewayService{}
	account := &Account{
		ID:          31,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	req, err := svc.buildUpstreamRequest(context.Background(), c, account, []byte(`{"model":"gpt-5"}`), "token", false, "pcache_123")
	require.NoError(t, err)
	require.Equal(t, codexCLIUserAgent, req.Header.Get("User-Agent"))
	require.Equal(t, "codex_cli_rs", req.Header.Get("originator"))
	require.Equal(t, codexCLIVersion, req.Header.Get("Version"))
	require.Equal(t, "responses=experimental", req.Header.Get("OpenAI-Beta"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), req.Header.Get("session_id"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "conv-cli"), req.Header.Get("conversation_id"))
	require.Equal(t, resolveOpenAICodexUpstreamWindowID(account.ID), req.Header.Get(openAICodexMetadataWindowIDKey))
	require.Equal(t, resolveOpenAICodexUpstreamInstallationID(account.ID), req.Header.Get(openAICodexMetadataInstallationIDKey))
}

func TestOpenAIBuildUpstreamRequestOpenAIPassthroughOAuthCompact_UsesStableUpstreamPersona(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", bytes.NewReader([]byte(`{"model":"gpt-5"}`)))
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 codex_cli_rs/0.200.0")
	c.Request.Header.Set("originator", "codex_vscode")
	c.Request.Header.Set("version", "0.200.0")
	c.Request.Header.Set("OpenAI-Beta", "responses=experimental,foo=bar")
	c.Request.Header.Set("session_id", "sess-cli")
	c.Request.Header.Set("conversation_id", "conv-cli")
	c.Request.Header.Set("Accept", "application/json")

	svc := &OpenAIGatewayService{}
	account := &Account{
		ID:          32,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, []byte(`{"model":"gpt-5"}`), "token")
	require.NoError(t, err)
	require.Equal(t, codexCLIUserAgent, req.Header.Get("User-Agent"))
	require.Equal(t, "codex_cli_rs", req.Header.Get("originator"))
	require.Equal(t, codexCLIVersion, req.Header.Get("Version"))
	require.Equal(t, "responses=experimental", req.Header.Get("OpenAI-Beta"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), req.Header.Get("session_id"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "conv-cli"), req.Header.Get("conversation_id"))
	require.Equal(t, resolveOpenAICodexUpstreamWindowID(account.ID), req.Header.Get(openAICodexMetadataWindowIDKey))
	require.Equal(t, resolveOpenAICodexUpstreamInstallationID(account.ID), req.Header.Get(openAICodexMetadataInstallationIDKey))
}

func TestOpenAIBuildOpenAIWSHeadersOAuth_UsesStableUpstreamPersona(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 codex_cli_rs/0.200.0")
	c.Request.Header.Set("originator", "codex_vscode")
	c.Request.Header.Set("OpenAI-Beta", "responses=experimental,foo=bar")
	c.Request.Header.Set("Accept-Language", "zh-CN")
	c.Request.Header.Set("session_id", "sess-cli")
	c.Request.Header.Set("conversation_id", "conv-cli")

	svc := &OpenAIGatewayService{}
	account := &Account{
		ID:          33,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	headers, resolution := svc.buildOpenAIWSHeaders(c, account, "token", OpenAIWSProtocolDecision{Transport: OpenAIUpstreamTransportResponsesWebsocketV2}, "", "", "pcache_123")
	require.Equal(t, codexCLIUserAgent, headers.Get("User-Agent"))
	require.Equal(t, "codex_cli_rs", headers.Get("originator"))
	require.Equal(t, openAIWSBetaV2Value, headers.Get("OpenAI-Beta"))
	require.Empty(t, headers.Get("Accept-Language"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), headers.Get("session_id"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "conv-cli"), headers.Get("conversation_id"))
	require.Equal(t, resolveOpenAICodexUpstreamWindowID(account.ID), headers.Get(openAICodexMetadataWindowIDKey))
	require.Equal(t, "header_session_id", resolution.SessionSource)
	require.Equal(t, "header_conversation_id", resolution.ConversationSource)
}

func TestOpenAIBuildUpstreamRequestOAuthNativeClient_UsesSessionIDOnlyAndCodexHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader([]byte(`{"model":"gpt-5.4","prompt_cache_key":"pcache_123"}`)))
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.120.0")
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set("session_id", "sess-cli")
	c.Request.Header.Set("x-codex-beta-features", "feature_a,feature_b")

	svc := &OpenAIGatewayService{}
	account := &Account{
		ID:          34,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	req, err := svc.buildUpstreamRequest(context.Background(), c, account, []byte(`{"model":"gpt-5.4","prompt_cache_key":"pcache_123"}`), "token", true, "pcache_123")
	require.NoError(t, err)
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), req.Header.Get("session_id"))
	require.Empty(t, req.Header.Get("conversation_id"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), req.Header.Get("x-client-request-id"))
	require.Empty(t, req.Header.Get("x-codex-beta-features"))
}

func TestOpenAIBuildUpstreamRequestOpenAIPassthroughOAuthNativeClient_UsesSessionIDOnlyAndCodexHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader([]byte(`{"model":"gpt-5.4","prompt_cache_key":"pcache_123"}`)))
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.120.0")
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set("session_id", "sess-cli")
	c.Request.Header.Set("x-codex-beta-features", "feature_a,feature_b")

	svc := &OpenAIGatewayService{}
	account := &Account{
		ID:          35,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, []byte(`{"model":"gpt-5.4","prompt_cache_key":"pcache_123"}`), "token")
	require.NoError(t, err)
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), req.Header.Get("session_id"))
	require.Empty(t, req.Header.Get("conversation_id"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), req.Header.Get("x-client-request-id"))
	require.Empty(t, req.Header.Get("x-codex-beta-features"))
}

func TestOpenAIBuildOpenAIWSHeadersOAuthNativeClient_UsesSessionIDOnlyAndCodexHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.120.0")
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set("session_id", "sess-cli")
	c.Request.Header.Set("x-codex-beta-features", "feature_a,feature_b")

	svc := &OpenAIGatewayService{}
	account := &Account{
		ID:          36,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	headers, resolution := svc.buildOpenAIWSHeaders(c, account, "token", OpenAIWSProtocolDecision{Transport: OpenAIUpstreamTransportResponsesWebsocketV2}, "", "", "pcache_123")
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), headers.Get("session_id"))
	require.Empty(t, headers.Get("conversation_id"))
	require.Equal(t, isolateOpenAISessionID(account.ID, "sess-cli"), headers.Get("x-client-request-id"))
	require.Empty(t, headers.Get("x-codex-beta-features"))
	require.Equal(t, "header_session_id", resolution.SessionSource)
	require.Equal(t, "none", resolution.ConversationSource)
}

func TestPrepareOpenAIForwardRequest_OAuthVirtualizesBodyIdentityFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.120.0")
	c.Request.Header.Set("originator", "codex_cli_rs")

	account := &Account{
		ID:       37,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{
		"model":"gpt-5.4",
		"prompt_cache_key":"pcache_123",
		"client_metadata":{
			"x-codex-installation-id":"inst-local-1",
			"x-codex-window-id":"window-local-1",
			"x-codex-turn-metadata":"turn-local-1",
			"x-openai-subagent":"review",
			"ignored_key":"drop-me"
		},
		"input":[{"type":"input_text","text":"hello"}]
	}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "pcache_123", "", true, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)

	require.Equal(t, isolateOpenAISessionID(account.ID, "pcache_123"), gjson.GetBytes(prepared.body, "prompt_cache_key").String())
	require.Equal(t, isolateOpenAISessionID(account.ID, "turn-local-1"), gjson.GetBytes(prepared.body, "client_metadata.x-codex-turn-metadata").String())
	require.Equal(t, "review", gjson.GetBytes(prepared.body, "client_metadata.x-openai-subagent").String())
	require.Equal(t, resolveOpenAICodexUpstreamInstallationID(account.ID), gjson.GetBytes(prepared.body, "client_metadata.x-codex-installation-id").String())
	require.Equal(t, resolveOpenAICodexUpstreamWindowID(account.ID), gjson.GetBytes(prepared.body, "client_metadata.x-codex-window-id").String())
	require.False(t, gjson.GetBytes(prepared.body, "client_metadata.ignored_key").Exists())
}

func TestPrepareOpenAIForwardRequest_OAuthInjectsStableInstallationMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.120.0")
	c.Request.Header.Set("originator", "codex_cli_rs")

	account := &Account{
		ID:       38,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"gpt-5.4","prompt_cache_key":"pcache_456","input":[{"type":"input_text","text":"hello"}]}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "pcache_456", "", true, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, resolveOpenAICodexUpstreamInstallationID(account.ID), gjson.GetBytes(prepared.body, "client_metadata.x-codex-installation-id").String())
	require.False(t, gjson.GetBytes(prepared.body, "client_metadata.x-codex-window-id").Exists())
}

func TestOpenAIBuildUpstreamRequest_OAuthPropagatesSanitizedIdentityHeadersFromPreparedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.120.0")
	c.Request.Header.Set("originator", "codex_cli_rs")

	account := &Account{
		ID:          39,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}
	body := []byte(`{
		"model":"gpt-5.4",
		"prompt_cache_key":"pcache_789",
		"type":"response.create",
		"client_metadata":{
			"x-codex-parent-thread-id":"parent-thread-local",
			"x-openai-subagent":"review"
		},
		"input":[{"type":"input_text","text":"hello"}]
	}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "pcache_789", "", true, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)

	req, err := svc.buildUpstreamRequest(context.Background(), c, account, prepared.body, "token", false, prepared.promptCacheKey)
	require.NoError(t, err)
	require.Equal(t, "review", req.Header.Get(openAICodexMetadataSubagentKey))
	require.Equal(t, generateSessionUUID("openai_codex_parent_thread:39:parent-thread-local"), req.Header.Get(openAICodexMetadataParentThreadIDKey))
	require.Equal(t, resolveOpenAICodexUpstreamWindowID(account.ID), req.Header.Get(openAICodexMetadataWindowIDKey))
}

func TestOpenAIBuildOpenAIWSHeaders_OAuthPropagatesSanitizedSubagentHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.120.0")
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set(openAICodexMetadataSubagentKey, "review")
	c.Request.Header.Set(openAICodexMetadataParentThreadIDKey, "parent-thread-header")

	svc := &OpenAIGatewayService{}
	account := &Account{
		ID:          40,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	headers, _ := svc.buildOpenAIWSHeaders(c, account, "token", OpenAIWSProtocolDecision{Transport: OpenAIUpstreamTransportResponsesWebsocketV2}, "", "", "pcache_123")
	require.Equal(t, "review", headers.Get(openAICodexMetadataSubagentKey))
	require.Equal(t, generateSessionUUID("openai_codex_parent_thread:40:parent-thread-header"), headers.Get(openAICodexMetadataParentThreadIDKey))
}
