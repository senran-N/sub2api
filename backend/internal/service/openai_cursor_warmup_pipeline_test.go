package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIForwardAsChatCompletionsOAuth_PreservesResponsesShapeBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	requestBody := []byte(`{
		"user":"cursor-user-1",
		"model":"gpt-5.4",
		"stream":false,
		"input":[
			{"role":"system","content":"You are GPT-5.4 running as a coding agent."},
			{"role":"user","content":"hello"}
		],
		"service_tier":"auto",
		"reasoning":{"effort":"high"},
		"prompt_cache_retention":"24h",
		"safety_identifier":"cursor-safe",
		"metadata":{"trace_id":"abc"},
		"stream_options":{"include_usage":true}
	}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(requestBody))

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/event-stream"},
				"x-request-id": []string{"rid-cursor-shape"},
			},
			Body: io.NopCloser(strings.NewReader(strings.Join([]string{
				`data: {"type":"response.created","response":{"id":"resp_cursor_shape","object":"response","model":"gpt-5.4","status":"in_progress","output":[]}}`,
				`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_cursor_shape","type":"message","role":"assistant","content":[],"status":"in_progress"}}`,
				`data: {"type":"response.content_part.done","output_index":0,"content_index":0,"item_id":"msg_cursor_shape","part":{"type":"output_text","text":"hello back"}}`,
				`data: {"type":"response.completed","response":{"id":"resp_cursor_shape","object":"response","model":"gpt-5.4","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2}}}`,
				`data: [DONE]`,
			}, "\n"))),
		},
	}

	svc := &OpenAIGatewayService{httpUpstream: upstream}
	account := &Account{
		ID:          1,
		Name:        "oauth-cursor-shape",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":       "oauth-token",
			"chatgpt_account_id": "chatgpt-acc",
		},
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, requestBody, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, upstream.lastReq)

	upstreamBody := upstream.lastBody
	require.Equal(t, "gpt-5.4", gjson.GetBytes(upstreamBody, "model").String())
	require.Equal(t, "cursor-user-1", gjson.GetBytes(upstreamBody, "user").String())
	require.Contains(t, gjson.GetBytes(upstreamBody, "instructions").String(), "You are GPT-5.4 running as a coding agent.")
	require.True(t, gjson.GetBytes(upstreamBody, "input").IsArray())
	require.Len(t, gjson.GetBytes(upstreamBody, "input").Array(), 1)
	require.Equal(t, "user", gjson.GetBytes(upstreamBody, "input.0.role").String())
	require.Equal(t, "hello", gjson.GetBytes(upstreamBody, "input.0.content").String())
	require.NotContains(t, string(upstreamBody), `"input":null`)

	for _, field := range cursorResponsesUnsupportedFields {
		require.False(t, gjson.GetBytes(upstreamBody, field).Exists(), "%s should be stripped before forwarding", field)
	}

	require.NotNil(t, result.ServiceTier)
	require.Equal(t, "auto", *result.ServiceTier)
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "high", *result.ReasoningEffort)
}
