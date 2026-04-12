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
		Body:       &errAfterDataReadCloser{data: []byte("data: {\"type\":\"response.in_progress\",\"response\":{\"id\":\"resp_boundary\"}}\n")},
	}

	svc := &OpenAIGatewayService{}
	_, err := svc.handleStreamingResponse(context.Background(), resp, c, &Account{ID: 1}, time.Now(), "model", "model")
	require.Error(t, err)
	require.True(t, errors.Is(err, io.ErrUnexpectedEOF) || strings.Contains(err.Error(), "stream read error"))
	require.Contains(t, rec.Body.String(), "\"stream_read_error\"")
	require.Contains(t, rec.Body.String(), "response.in_progress")
	require.Contains(t, rec.Body.String(), "}}\n\ndata: {\"type\":\"error\"")
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
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", false, "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.NotContains(t, string(prepared.body), `"user":"user_123"`)
	require.NotContains(t, prepared.reqBody, "user")
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
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "custom-original-model", false, "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "custom/upstream-model", gjson.GetBytes(prepared.body, "model").String())
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

	parsed, err := parseOpenAIWSIngressClientPayload(c, account, []byte(`{"model":"custom-original-model","stream":false}`))
	require.NoError(t, err)
	require.Equal(t, "custom/upstream-model", gjson.GetBytes(parsed.payloadRaw, "model").String())
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
	prepared, err := svc.prepareOpenAIForwardRequest(c, account, body, "gpt-5.4", true, "", false, openAIWSHTTPDecision("test"))
	require.NoError(t, err)
	require.NotNil(t, prepared)
	require.Equal(t, "resp_prev_http_1", gjson.GetBytes(prepared.body, "previous_response_id").String())
	require.Equal(t, "resp_prev_http_1", prepared.reqBody["previous_response_id"])
}
