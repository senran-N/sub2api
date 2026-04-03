package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGatewayService_HandleAnthropicPassthroughUpstreamResponse_SuccessSkipsHandling(t *testing.T) {
	svc := &GatewayService{}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}

	result, err, handled := svc.handleAnthropicPassthroughUpstreamResponse(context.Background(), nil, &Account{}, resp)

	require.False(t, handled)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestGatewayService_FinalizeAnthropicPassthroughSuccess_NonStreamingBuildsForwardResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	svc := &GatewayService{
		cfg:                  &config.Config{},
		rateLimitService:     &RateLimitService{},
		responseHeaderFilter: compileResponseHeaderFilter(&config.Config{}),
	}
	account := &Account{ID: 64, Platform: PlatformAnthropic}
	input := anthropicPassthroughForwardInput{
		OriginalModel: "claude-3-5-sonnet-latest",
		RequestModel:  "claude-3-opus-20240229",
		RequestStream: false,
		StartTime:     time.Now().Add(-80 * time.Millisecond),
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Request-Id": []string{"rid-pass-success"},
		},
		Body: io.NopCloser(strings.NewReader(`{"usage":{"input_tokens":5,"output_tokens":9}}`)),
	}

	result, err := svc.finalizeAnthropicPassthroughSuccess(context.Background(), c, account, resp, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "rid-pass-success", result.RequestID)
	require.Equal(t, "claude-3-5-sonnet-latest", result.Model)
	require.Equal(t, "claude-3-opus-20240229", result.UpstreamModel)
	require.False(t, result.Stream)
	require.Equal(t, 5, result.Usage.InputTokens)
	require.Equal(t, 9, result.Usage.OutputTokens)
	require.Equal(t, http.StatusOK, recorder.Code)
}
