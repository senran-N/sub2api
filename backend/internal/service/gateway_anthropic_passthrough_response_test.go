package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestWriteAnthropicPassthroughResponseHeaders_DefaultFallback(t *testing.T) {
	dst := http.Header{}
	src := http.Header{
		"Content-Type": []string{"application/json"},
		"X-Request-Id": []string{"rid-123"},
		"X-Extra":      []string{"blocked"},
	}

	writeAnthropicPassthroughResponseHeaders(dst, src, nil)

	require.Equal(t, "application/json", dst.Get("Content-Type"))
	require.Equal(t, "rid-123", dst.Get("x-request-id"))
	require.Empty(t, dst.Get("x-extra"))
}

func TestWriteAnthropicPassthroughResponseHeaders_UsesCompiledFilter(t *testing.T) {
	filter := compileResponseHeaderFilter(&config.Config{
		Security: config.SecurityConfig{
			ResponseHeaders: config.ResponseHeaderConfig{
				Enabled:           true,
				AdditionalAllowed: []string{"x-trace-id"},
				ForceRemove:       []string{"x-request-id"},
			},
		},
	})

	dst := http.Header{}
	src := http.Header{
		"Content-Type": []string{"text/event-stream"},
		"x-request-id": []string{"rid-hidden"},
		"x-trace-id":   []string{"trace-visible"},
	}

	writeAnthropicPassthroughResponseHeaders(dst, src, filter)

	require.Equal(t, "text/event-stream", dst.Get("Content-Type"))
	require.Equal(t, "trace-visible", dst.Get("x-trace-id"))
	require.Empty(t, dst.Get("x-request-id"))
}

func TestParseClaudeUsageFromResponseBody_FallsBackToDetailedBuckets(t *testing.T) {
	body := []byte(`{"usage":{"input_tokens":21,"output_tokens":34,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"cached_tokens":13,"cache_creation":{"ephemeral_5m_input_tokens":5,"ephemeral_1h_input_tokens":8}}}`)

	got := parseClaudeUsageFromResponseBody(body)

	require.NotNil(t, got)
	require.Equal(t, 21, got.InputTokens)
	require.Equal(t, 34, got.OutputTokens)
	require.Equal(t, 13, got.CacheReadInputTokens)
	require.Equal(t, 13, got.CacheCreationInputTokens)
	require.Equal(t, 5, got.CacheCreation5mTokens)
	require.Equal(t, 8, got.CacheCreation1hTokens)
}

func TestExtractAnthropicSSEDataLine_TrimsLeadingWhitespace(t *testing.T) {
	data, ok := extractAnthropicSSEDataLine("data:\t  {\"type\":\"message_start\"}")

	require.True(t, ok)
	require.Equal(t, `{"type":"message_start"}`, data)
}

func TestGatewayService_HandleNonStreamingResponseAnthropicAPIKeyPassthrough_DefaultContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	svc := &GatewayService{}
	resp := &http.Response{
		StatusCode: http.StatusCreated,
		Header: http.Header{
			"X-Request-Id": []string{"rid-456"},
		},
		Body: io.NopCloser(strings.NewReader(`{"usage":{"cached_tokens":9}}`)),
	}

	usage, err := svc.handleNonStreamingResponseAnthropicAPIKeyPassthrough(context.Background(), resp, c, &Account{ID: 1})

	require.NoError(t, err)
	require.NotNil(t, usage)
	require.Equal(t, 9, usage.CacheReadInputTokens)
	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	require.Equal(t, "rid-456", recorder.Header().Get("x-request-id"))
	require.JSONEq(t, `{"usage":{"cached_tokens":9}}`, recorder.Body.String())
}
