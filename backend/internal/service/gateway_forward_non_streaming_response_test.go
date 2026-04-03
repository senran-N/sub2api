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
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGatewayService_HandleNonStreamingResponse_RewritesCachedTokensAndModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	svc := &GatewayService{
		cfg:                  &config.Config{},
		rateLimitService:     &RateLimitService{},
		responseHeaderFilter: compileResponseHeaderFilter(&config.Config{}),
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(`{"model":"claude-3-opus-20240229","usage":{"input_tokens":12,"output_tokens":7,"cache_creation":{"ephemeral_5m_input_tokens":2,"ephemeral_1h_input_tokens":3},"cached_tokens":4}}`)),
	}

	usage, err := svc.handleNonStreamingResponse(context.Background(), resp, c, &Account{}, "claude-3-5-sonnet-latest", "claude-3-opus-20240229")

	require.NoError(t, err)
	require.NotNil(t, usage)
	require.Equal(t, 12, usage.InputTokens)
	require.Equal(t, 7, usage.OutputTokens)
	require.Equal(t, 4, usage.CacheReadInputTokens)
	require.Equal(t, 2, usage.CacheCreation5mTokens)
	require.Equal(t, 3, usage.CacheCreation1hTokens)
	require.Contains(t, recorder.Body.String(), `"model":"claude-3-5-sonnet-latest"`)
	require.Contains(t, recorder.Body.String(), `"cache_read_input_tokens":4`)
}

func TestGatewayService_HandleNonStreamingResponse_ResponseTooLargeWritesGatewayError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	svc := &GatewayService{
		cfg: &config.Config{
			Gateway: config.GatewayConfig{
				UpstreamResponseReadMaxBytes: 8,
			},
		},
		rateLimitService:     &RateLimitService{},
		responseHeaderFilter: compileResponseHeaderFilter(&config.Config{}),
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"usage":{"input_tokens":123456}}`))),
	}

	usage, err := svc.handleNonStreamingResponse(context.Background(), resp, c, &Account{}, "model", "model")

	require.Nil(t, usage)
	require.ErrorIs(t, err, ErrUpstreamResponseBodyTooLarge)
	require.Equal(t, http.StatusBadGateway, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Upstream response too large")
}
