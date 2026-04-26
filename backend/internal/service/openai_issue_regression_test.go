package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type errAfterDataReadCloser struct {
	data []byte
	read bool
}

func (r *errAfterDataReadCloser) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.ErrUnexpectedEOF
	}
	r.read = true
	return copy(p, r.data), nil
}

func (r *errAfterDataReadCloser) Close() error {
	return nil
}

func TestExtractCodexFinalResponse_RebuildsEmptyOutputFromSSE(t *testing.T) {
	body := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_rebuild","object":"response","model":"gpt-5.4","status":"in_progress","output":[]}}`,
		`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_rebuild","type":"message","role":"assistant","content":[],"status":"in_progress"}}`,
		`data: {"type":"response.content_part.done","output_index":0,"content_index":0,"item_id":"msg_rebuild","part":{"type":"output_text","text":"TEST_OK"}}`,
		`data: {"type":"response.completed","response":{"id":"resp_rebuild","object":"response","model":"gpt-5.4","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":2}}}`,
		`data: [DONE]`,
	}, "\n")

	finalResp, ok := extractCodexFinalResponse(body)
	require.True(t, ok)
	require.Equal(t, "TEST_OK", gjson.GetBytes(finalResp, "output.0.content.0.text").String())
}

func TestHandleChatBufferedStreamingResponse_ForcesJSONContentTypeAndRepairsOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-chat"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_chat","object":"response","model":"gpt-5.4","status":"in_progress","output":[]}}`,
			`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_chat","type":"message","role":"assistant","content":[],"status":"in_progress"}}`,
			`data: {"type":"response.content_part.done","output_index":0,"content_index":0,"item_id":"msg_chat","part":{"type":"output_text","text":"hello world"}}`,
			`data: {"type":"response.completed","response":{"id":"resp_chat","object":"response","model":"gpt-5.4","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2}}}`,
			`data: [DONE]`,
		}, "\n"))),
	}

	svc := &OpenAIGatewayService{}
	result, err := svc.handleChatBufferedStreamingResponse(resp, c, "gpt-5.4", "gpt-5.4", time.Now())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.Contains(t, rec.Body.String(), `"content":"hello world"`)
}

func TestHandleAnthropicBufferedStreamingResponse_RebuildsOutputFromDeltasOnResponseDone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-anthropic"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_messages","object":"response","model":"gpt-5.4","status":"in_progress","output":[]}}`,
			`data: {"type":"response.output_text.delta","delta":"hello world"}`,
			`data: {"type":"response.done","response":{"id":"resp_messages","object":"response","model":"gpt-5.4","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2}}}`,
			`data: [DONE]`,
		}, "\n"))),
	}

	svc := &OpenAIGatewayService{}
	result, err := svc.handleAnthropicBufferedStreamingResponse(resp, c, "gpt-5.4", "gpt-5.4", "gpt-5.4", time.Now())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 3, result.Usage.InputTokens)
	require.Equal(t, 2, result.Usage.OutputTokens)
	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.Equal(t, "gpt-5.4", gjson.GetBytes(rec.Body.Bytes(), "model").String())
	require.Equal(t, "hello world", gjson.GetBytes(rec.Body.Bytes(), "content.0.text").String())
}

func TestOpenAIStreamEventIsTerminal_AcceptsIncompleteAndCancelledVariants(t *testing.T) {
	require.True(t, openAIStreamEventIsTerminal(`{"type":"response.incomplete"}`))
	require.True(t, openAIStreamEventIsTerminal(`{"type":"response.cancelled"}`))
	require.True(t, openAIStreamEventIsTerminal(`{"type":"response.canceled"}`))
}

func TestOpenAIStreamingReadErrorClosesOpenEventBeforeInjectedErrorEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       &errAfterDataReadCloser{data: []byte("data: {\"type\":\"response.output_item.added\",\"item\":{\"type\":\"message\"},\"output_index\":0}\n")},
	}

	svc := &OpenAIGatewayService{}
	_, err := svc.handleStreamingResponse(context.Background(), resp, c, &Account{ID: 1}, time.Now(), "model", "model")
	require.Error(t, err)
	require.True(t, errors.Is(err, io.ErrUnexpectedEOF) || strings.Contains(err.Error(), "stream read error"))
	require.Contains(t, rec.Body.String(), "\"stream_read_error\"")
	require.Contains(t, rec.Body.String(), "response.output_item.added")
	require.Contains(t, rec.Body.String(), "}\n\ndata: {\"type\":\"error\"")
}

func TestPrepareOpenAIForwardRequest_StripsUnsupportedUserField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	account := &Account{
		Name:     "test-openai",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"gpt-5.4","input":"hello","user":"user_123"}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "", "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.NotContains(t, string(prepared.body), `"user":"user_123"`)
	require.NotContains(t, prepared.reqBody, "user")
}

func TestPrepareOpenAIForwardRequest_OfficialCodexPreservesMissingInstructions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.104.0")

	account := &Account{
		Name:     "test-openai",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"gpt-5.4","input":"hello"}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "", "", true, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.False(t, gjson.GetBytes(prepared.body, "instructions").Exists())
	_, hasInstructions := prepared.reqBody["instructions"]
	require.False(t, hasInstructions)
}

func TestPrepareOpenAIForwardRequest_OfficialCodexPreservesReasoningMinimal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.117.0")

	account := &Account{
		Name:     "test-openai",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"gpt-5.4","reasoning":{"effort":"minimal"},"input":"hello"}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "", "", true, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "minimal", gjson.GetBytes(prepared.body, "reasoning.effort").String())
	reasoning, ok := prepared.reqBody["reasoning"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "minimal", reasoning["effort"])
}

func TestPrepareOpenAIForwardRequest_OfficialCodexPreservesNativeModelIdentifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.117.0")

	account := &Account{
		Name:     "test-openai",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"codex-mini-latest","stream":true,"store":false,"input":"hello"}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "codex-mini-latest", true, "", "", true, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "codex-mini-latest", gjson.GetBytes(prepared.body, "model").String())
	require.Equal(t, "codex-mini-latest", prepared.reqBody["model"])
}

func TestPrepareOpenAIForwardRequest_ForceCodexCLICompatStillInjectsInstructions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "curl/8.0")

	account := &Account{
		Name:     "test-openai",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"gpt-5.4","input":"hello"}`)

	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{ForceCodexCLI: true}}}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "", "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, defaultOpenAICodexInstructions, gjson.GetBytes(prepared.body, "instructions").String())
	require.Equal(t, defaultOpenAICodexInstructions, prepared.reqBody["instructions"])
}

func TestPrepareOpenAIForwardRequest_APIKeyPreservesMappedCustomModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	account := &Account{
		Name:     "test-openai-api-key",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"custom-original-model": "custom/upstream-model",
			},
		},
	}
	body := []byte(`{"model":"custom-original-model","input":"hello"}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "custom-original-model", false, "", "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "custom/upstream-model", gjson.GetBytes(prepared.body, "model").String())
}

func TestPrepareOpenAIForwardRequest_PreservesReasoningVariantFallbackMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	account := &Account{
		Name:     "test-openai-oauth",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.4": "gpt-5.3-codex-spark",
			},
		},
	}
	body := []byte(`{"model":"gpt-5.4-xhigh","input":"hello"}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4-xhigh", false, "", "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "gpt-5.3-codex-spark-xhigh", gjson.GetBytes(prepared.body, "model").String())
	require.Equal(t, "gpt-5.3-codex-spark-xhigh", prepared.reqBody["model"])
}

func TestPrepareOpenAIForwardRequest_AppliesDefaultMappedModelForReasoningVariant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	account := &Account{
		Name:     "test-openai-oauth",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"gpt-5.4-xhigh","input":"hello"}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4-xhigh", false, "", "gpt-5.2", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "gpt-5.2-xhigh", gjson.GetBytes(prepared.body, "model").String())
	require.Equal(t, "gpt-5.2-xhigh", prepared.reqBody["model"])
}

func TestPrepareOpenAIForwardRequest_AppliesCompactModelMappingOnlyOnCompactPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{
		Name:     "test-openai-oauth",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"compact_model_mapping": map[string]any{
				"gpt-5.4": "gpt-5.4-compact",
			},
		},
	}
	body := []byte(`{"model":"gpt-5.4","input":"hello"}`)
	svc := &OpenAIGatewayService{}

	rec := httptest.NewRecorder()
	regularCtx, _ := gin.CreateTestContext(rec)
	regularCtx.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	regularPrepared, err := svc.prepareOpenAIForwardRequest(regularCtx, account, body, "gpt-5.4", false, "", "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.Equal(t, "gpt-5.4", gjson.GetBytes(regularPrepared.body, "model").String())

	compactRec := httptest.NewRecorder()
	compactCtx, _ := gin.CreateTestContext(compactRec)
	compactCtx.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", nil)
	compactPrepared, err := svc.prepareOpenAIForwardRequest(compactCtx, account, body, "gpt-5.4", false, "", "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.Equal(t, "gpt-5.4-compact", gjson.GetBytes(compactPrepared.body, "model").String())
	require.Equal(t, "gpt-5.4-compact", compactPrepared.reqBody["model"])
}

func TestSetOpenAICompatPromptCacheSessionID_UsesIsolatedPromptCacheKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	req := httptest.NewRequest(http.MethodPost, "https://chatgpt.com/backend-api/codex/responses", nil)
	account := &Account{ID: 101, Type: AccountTypeOAuth}
	setOpenAICompatPromptCacheSessionID(c, account, req, "pc-shared")

	isolatedPromptCacheKey := isolateOpenAISessionID(account.ID, "pc-shared")
	require.Equal(t, generateSessionUUID(isolatedPromptCacheKey), req.Header.Get("session_id"))
}

func TestForwardAsChatCompletions_OAuthPromptCacheKeyKeepsIsolatedSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	requestBody := []byte(`{"model":"gpt-5.4","messages":[{"role":"user","content":"hello"}],"stream":false}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(requestBody))
	c.Set("api_key", &APIKey{ID: 101})

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/event-stream"},
				"x-request-id": []string{"rid-chat-isolated"},
			},
			Body: io.NopCloser(strings.NewReader(strings.Join([]string{
				`data: {"type":"response.created","response":{"id":"resp_chat_isolated","object":"response","model":"gpt-5.4","status":"in_progress","output":[]}}`,
				`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_chat_isolated","type":"message","role":"assistant","content":[],"status":"in_progress"}}`,
				`data: {"type":"response.content_part.done","output_index":0,"content_index":0,"item_id":"msg_chat_isolated","part":{"type":"output_text","text":"hello back"}}`,
				`data: {"type":"response.completed","response":{"id":"resp_chat_isolated","object":"response","model":"gpt-5.4","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2}}}`,
				`data: [DONE]`,
			}, "\n"))),
		},
	}

	svc := &OpenAIGatewayService{httpUpstream: upstream}
	account := &Account{
		ID:          1,
		Name:        "oauth-chat",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":       "oauth-token",
			"chatgpt_account_id": "chatgpt-acc",
		},
	}

	promptCacheKey := "pc-shared"
	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, requestBody, promptCacheKey, "")
	require.NoError(t, err)
	require.NotNil(t, result)

	isolatedPromptCacheKey := isolateOpenAISessionID(account.ID, promptCacheKey)
	require.Equal(t, isolatedPromptCacheKey, upstream.lastReq.Header.Get("conversation_id"))
	require.Equal(t, generateSessionUUID(isolatedPromptCacheKey), upstream.lastReq.Header.Get("session_id"))
	require.NotEqual(t, generateSessionUUID(promptCacheKey), upstream.lastReq.Header.Get("session_id"))
}

func TestParseOpenAIWSIngressClientPayload_APIKeyPreservesMappedCustomModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	account := &Account{
		Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"custom-original-model": "custom/upstream-model",
			},
		},
	}

	svc := &OpenAIGatewayService{}
	parsed, err := svc.parseOpenAIWSIngressClientPayload(c, account, []byte(`{"model":"custom-original-model","stream":false}`))
	require.NoError(t, err)
	require.Equal(t, "custom/upstream-model", gjson.GetBytes(parsed.payloadRaw, "model").String())
}

func TestParseOpenAIWSIngressClientPayload_PreservesReasoningVariantFallbackMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	account := &Account{
		Type: AccountTypeOAuth,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.4": "gpt-5.3-codex-spark",
			},
		},
	}

	svc := &OpenAIGatewayService{}
	parsed, err := svc.parseOpenAIWSIngressClientPayload(c, account, []byte(`{"type":"response.create","model":"gpt-5.4-xhigh","stream":false}`))
	require.NoError(t, err)
	require.Equal(t, "gpt-5.3-codex-spark-xhigh", gjson.GetBytes(parsed.payloadRaw, "model").String())
}

func TestParseOpenAIWSIngressClientPayload_DoesNotApplyGroupDefaultWithoutSelectionFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)

	groupID := int64(10)
	c.Set("api_key", &APIKey{
		GroupID: &groupID,
		Group:   &Group{ID: groupID, DefaultMappedModel: "gpt-5.2"},
	})

	account := &Account{
		Type: AccountTypeAPIKey,
	}

	svc := &OpenAIGatewayService{}
	parsed, err := svc.parseOpenAIWSIngressClientPayload(c, account, []byte(`{"type":"response.create","model":"gpt-5.4-xhigh","stream":false}`))
	require.NoError(t, err)
	require.Equal(t, "gpt-5.4-xhigh", gjson.GetBytes(parsed.payloadRaw, "model").String())
}

func TestParseOpenAIWSIngressClientPayload_AppliesExplicitSelectionFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
	AttachOpenAIWSSelectionFallbackModel(c, "gpt-5.2")

	account := &Account{
		Type: AccountTypeAPIKey,
	}

	svc := &OpenAIGatewayService{}
	parsed, err := svc.parseOpenAIWSIngressClientPayload(c, account, []byte(`{"type":"response.create","model":"gpt-5.4-xhigh","stream":false}`))
	require.NoError(t, err)
	require.Equal(t, "gpt-5.2-xhigh", gjson.GetBytes(parsed.payloadRaw, "model").String())
}

func TestParseOpenAIWSIngressClientPayload_PrefersChannelMappedModelOverGroupDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)

	groupID := int64(10)
	c.Set("api_key", &APIKey{
		GroupID: &groupID,
		Group:   &Group{ID: groupID, DefaultMappedModel: "gpt-5.2"},
	})

	channelService := newTestChannelService(makeStandardChannelRepo(Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{groupID},
		ModelMapping: map[string]map[string]string{
			PlatformOpenAI: {
				"gpt-5.4": "gpt-4.1",
			},
		},
	}, map[int64]string{groupID: PlatformOpenAI}))

	account := &Account{
		Type: AccountTypeAPIKey,
	}

	svc := &OpenAIGatewayService{channelService: channelService}
	parsed, err := svc.parseOpenAIWSIngressClientPayload(c, account, []byte(`{"type":"response.create","model":"gpt-5.4","stream":false}`))
	require.NoError(t, err)
	require.Equal(t, "gpt-4.1", gjson.GetBytes(parsed.payloadRaw, "model").String())
}

func TestSanitizeEmptyBase64InputImagesInOpenAIBody_DropsEmptyParts(t *testing.T) {
	body := []byte(`{"input":[{"type":"message","content":[{"type":"input_text","text":"hi"},{"type":"input_image","image_url":"data:image/png;base64,   "}]}]}`)

	sanitizedBody, sanitized, err := sanitizeEmptyBase64InputImagesInOpenAIBody(body)
	require.NoError(t, err)
	require.True(t, sanitized)
	require.False(t, bytes.Contains(sanitizedBody, []byte(`"input_image"`)))
	require.Equal(t, "hi", gjson.GetBytes(sanitizedBody, "input.0.content.0.text").String())
}

func TestPrepareOpenAIForwardRequest_HTTPPreservesPreviousResponseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	account := &Account{
		Name:     "test-openai-oauth",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"gpt-5.4","stream":true,"previous_response_id":"resp_prev_http_1","input":[{"type":"input_text","text":"hello"}]}`)

	svc := &OpenAIGatewayService{}
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", true, "", "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "resp_prev_http_1", gjson.GetBytes(prepared.body, "previous_response_id").String())
	require.Equal(t, "resp_prev_http_1", prepared.reqBody["previous_response_id"])
}
