package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
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

func TestOpenAIBuildUpstreamRequestOAuthCompact_PreservesClientFingerprint(t *testing.T) {
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
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	req, err := svc.buildUpstreamRequest(context.Background(), c, account, []byte(`{"model":"gpt-5"}`), "token", false, "pcache_123", true)
	require.NoError(t, err)
	require.Equal(t, "Mozilla/5.0 codex_cli_rs/0.200.0", req.Header.Get("User-Agent"))
	require.Equal(t, "codex_vscode", req.Header.Get("originator"))
	require.Equal(t, "0.200.0", req.Header.Get("Version"))
	require.Equal(t, "responses=experimental,foo=bar", req.Header.Get("OpenAI-Beta"))
	require.Equal(t, isolateOpenAISessionID(0, "sess-cli"), req.Header.Get("session_id"))
	require.Equal(t, isolateOpenAISessionID(0, "conv-cli"), req.Header.Get("conversation_id"))
}

func TestOpenAIBuildUpstreamRequestOpenAIPassthroughOAuthCompact_PreservesClientFingerprint(t *testing.T) {
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
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, account, []byte(`{"model":"gpt-5"}`), "token")
	require.NoError(t, err)
	require.Equal(t, "Mozilla/5.0 codex_cli_rs/0.200.0", req.Header.Get("User-Agent"))
	require.Equal(t, "codex_vscode", req.Header.Get("originator"))
	require.Equal(t, "0.200.0", req.Header.Get("Version"))
	require.Equal(t, "responses=experimental,foo=bar", req.Header.Get("OpenAI-Beta"))
	require.Equal(t, isolateOpenAISessionID(0, "sess-cli"), req.Header.Get("session_id"))
	require.Equal(t, isolateOpenAISessionID(0, "conv-cli"), req.Header.Get("conversation_id"))
}

func TestOpenAIBuildOpenAIWSHeadersOAuth_PreservesClientFingerprint(t *testing.T) {
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
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc"},
	}

	headers, resolution := svc.buildOpenAIWSHeaders(c, account, "token", OpenAIWSProtocolDecision{Transport: OpenAIUpstreamTransportResponsesWebsocketV2}, true, "", "", "pcache_123")
	require.Equal(t, "Mozilla/5.0 codex_cli_rs/0.200.0", headers.Get("User-Agent"))
	require.Equal(t, "codex_vscode", headers.Get("originator"))
	require.Equal(t, "responses=experimental,foo=bar", headers.Get("OpenAI-Beta"))
	require.Equal(t, "zh-CN", headers.Get("Accept-Language"))
	require.Equal(t, isolateOpenAISessionID(0, "sess-cli"), headers.Get("session_id"))
	require.Equal(t, isolateOpenAISessionID(0, "conv-cli"), headers.Get("conversation_id"))
	require.Equal(t, "header_session_id", resolution.SessionSource)
	require.Equal(t, "header_conversation_id", resolution.ConversationSource)
}
