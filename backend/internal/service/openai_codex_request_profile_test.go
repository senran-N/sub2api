package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetCodexRequestProfile_HTTPResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", strings.NewReader(`{"model":"gpt-5.3-codex","stream":true,"store":false,"prompt_cache_key":"cache_123","previous_response_id":"resp_prev_123","instructions":"be concise"}`))
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.104.0")
	c.Request.Header.Set("Accept", "text/event-stream")
	c.Request.Header.Set("Accept-Language", "zh-CN")
	c.Request.Header.Set("OpenAI-Beta", "responses=experimental")
	c.Request.Header.Set("session_id", "session_123")
	c.Request.Header.Set("conversation_id", "conversation_123")
	c.Request.Header.Set("originator", "codex_cli_rs")
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

	profile := GetCodexRequestProfile(c, []byte(`{"model":"gpt-5.3-codex","stream":true,"store":false,"prompt_cache_key":"cache_123","previous_response_id":"resp_prev_123","instructions":"be concise"}`), false)

	require.True(t, profile.OfficialClient)
	require.Equal(t, CodexOfficialClientReasonUserAgent, profile.OfficialClientReason)
	require.Equal(t, "0.104.0", profile.CodexVersion)
	require.Equal(t, CodexWireAPIResponsesHTTP, profile.WireAPI)
	require.Equal(t, OpenAIClientTransportHTTP, profile.ClientTransport)
	require.True(t, profile.TransportFallbackHTTP)
	require.False(t, profile.CompactPath)
	require.False(t, profile.Warmup)
	require.Equal(t, "gpt-5.3-codex", profile.Body.Model)
	require.True(t, profile.Body.StreamPresent)
	require.True(t, profile.Body.Stream)
	require.True(t, profile.Body.StorePresent)
	require.False(t, profile.Body.Store)
	require.Equal(t, "resp_prev_123", profile.Body.PreviousResponseID)
	require.Equal(t, "cache_123", profile.Body.PromptCacheKey)
	require.True(t, profile.Body.InstructionsPresent)
	require.True(t, profile.Continuation.DependsOnPriorResponse)
	require.Equal(t, OpenAIPreviousResponseIDKindResponseID, profile.Continuation.PreviousResponseIDKind)
}

func TestGetCodexRequestProfile_CompositeOfficialUserAgentVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", strings.NewReader(`{"model":"gpt-5.3-codex"}`))
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 codex_cli_rs/0.98.0 (Windows 10.0.19045; x86_64) unknown")
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

	profile := GetCodexRequestProfile(c, []byte(`{"model":"gpt-5.3-codex"}`), false)

	require.True(t, profile.OfficialClient)
	require.Equal(t, CodexOfficialClientReasonUserAgent, profile.OfficialClientReason)
	require.Equal(t, "0.98.0", profile.CodexVersion)
}

func TestGetCodexRequestProfile_WSResponsesWarmupAndContinuation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/openai/v1/responses", nil)
	c.Request.Header.Set("originator", "codex_chatgpt_desktop")
	c.Request.Header.Set(openAIWSTurnStateHeader, "turn_state_123")
	SetOpenAIClientTransport(c, OpenAIClientTransportWS)

	body := []byte(`{"type":"response.create","model":"gpt-5.3-codex","generate":false,"previous_response_id":"msg_prev_123","input":[{"type":"function_call_output","call_id":"call_1"}]}`)
	profile := GetCodexRequestProfile(c, body, false)

	require.True(t, profile.OfficialClient)
	require.Equal(t, CodexOfficialClientReasonOriginator, profile.OfficialClientReason)
	require.Equal(t, CodexWireAPIResponsesWebSocket, profile.WireAPI)
	require.Equal(t, OpenAIClientTransportWS, profile.ClientTransport)
	require.True(t, profile.Warmup)
	require.Equal(t, "response.create", profile.Body.RequestType)
	require.True(t, profile.Body.FunctionCallOutputPresent)
	require.Equal(t, "msg_prev_123", profile.Body.PreviousResponseID)
	require.True(t, profile.Continuation.HasTurnState)
	require.True(t, profile.Continuation.DependsOnPriorResponse)
	require.Equal(t, OpenAIPreviousResponseIDKindMessageID, profile.Continuation.PreviousResponseIDKind)
}

func TestGetCodexRequestProfile_CodexDesktopVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/openai/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 Codex Desktop/1.2.3")
	SetOpenAIClientTransport(c, OpenAIClientTransportWS)

	profile := GetCodexRequestProfile(c, []byte(`{"type":"response.create","model":"gpt-5.3-codex"}`), false)

	require.True(t, profile.OfficialClient)
	require.Equal(t, CodexOfficialClientReasonUserAgent, profile.OfficialClientReason)
	require.Equal(t, "1.2.3", profile.CodexVersion)
}

func TestGetCodexRequestProfile_CacheRespectsForceFlagAndBodyHash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "curl/8.0")
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

	firstBody := []byte(`{"model":"gpt-5.3-codex"}`)
	secondBody := []byte(`{"model":"gpt-5.4-codex"}`)

	headerOnlyProfile := GetCodexRequestProfile(c, nil, false)
	require.False(t, headerOnlyProfile.OfficialClient)
	require.Empty(t, headerOnlyProfile.Body.Model)

	profile := GetCodexRequestProfile(c, firstBody, false)
	require.False(t, profile.OfficialClient)
	require.Equal(t, "gpt-5.3-codex", profile.Body.Model)

	forcedProfile := GetCodexRequestProfile(c, firstBody, true)
	require.True(t, forcedProfile.OfficialClient)
	require.Equal(t, CodexOfficialClientReasonForceCodexCLI, forcedProfile.OfficialClientReason)

	updatedProfile := GetCodexRequestProfile(c, secondBody, true)
	require.Equal(t, "gpt-5.4-codex", updatedProfile.Body.Model)
}

func TestGetCodexRequestProfile_CacheInvalidatesOnHeaderChanges(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	c.Request.Header.Set("session_id", "session_123")
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

	profile := GetCodexRequestProfile(c, nil, false)
	require.Equal(t, "session_123", profile.Headers.SessionID)
	require.Empty(t, profile.Headers.ConversationID)

	c.Request.Header.Del("session_id")
	c.Request.Header.Set("conversation_id", "conversation_456")

	updatedProfile := GetCodexRequestProfile(c, nil, false)
	require.Empty(t, updatedProfile.Headers.SessionID)
	require.Equal(t, "conversation_456", updatedProfile.Headers.ConversationID)
}
