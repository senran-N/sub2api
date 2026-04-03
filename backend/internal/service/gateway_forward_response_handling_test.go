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

func TestGatewayService_HandleForwardUpstreamResponse_SuccessSkipsHandling(t *testing.T) {
	svc := &GatewayService{}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}

	result, err, handled := svc.handleForwardUpstreamResponse(context.Background(), nil, &Account{}, resp)

	require.False(t, handled)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestGatewayService_FinalizeForwardSuccess_NonStreamingBuildsForwardResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	svc := &GatewayService{
		cfg:                  &config.Config{},
		rateLimitService:     &RateLimitService{},
		responseHeaderFilter: compileResponseHeaderFilter(&config.Config{}),
	}
	account := &Account{ID: 42, Platform: PlatformAnthropic}
	preparation := &forwardRequestPreparation{
		reqModel:      "claude-sonnet-4-5-20250929",
		originalModel: "claude-sonnet-4-5",
		reqStream:     false,
	}
	parsed := &ParsedRequest{}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Request-Id": []string{"req-forward-success"},
		},
		Body: io.NopCloser(strings.NewReader(`{"usage":{"input_tokens":7,"output_tokens":11}}`)),
	}

	result, err := svc.finalizeForwardSuccess(context.Background(), c, account, resp, parsed, preparation, time.Now().Add(-50*time.Millisecond))

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "req-forward-success", result.RequestID)
	require.Equal(t, "claude-sonnet-4-5", result.Model)
	require.Equal(t, "claude-sonnet-4-5-20250929", result.UpstreamModel)
	require.False(t, result.Stream)
	require.Equal(t, 7, result.Usage.InputTokens)
	require.Equal(t, 11, result.Usage.OutputTokens)
	require.Equal(t, http.StatusOK, recorder.Code)
}
