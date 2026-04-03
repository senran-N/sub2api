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
	"github.com/tidwall/gjson"
)

func TestGatewayService_ResolveCountTokensPassthroughBody_AppliesAccountModelMapping(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Name:     "passthrough-account",
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"claude-sonnet-4-*": "claude-sonnet-4-5-20241022",
			},
		},
	}
	parsed := &ParsedRequest{
		Body:  []byte(`{"model":"claude-sonnet-4-20250514","messages":[]}`),
		Model: "claude-sonnet-4-20250514",
	}

	body := svc.resolveCountTokensPassthroughBody(account, parsed)

	require.Equal(t, "claude-sonnet-4-5-20241022", gjson.GetBytes(body, "model").String())
}

func TestGatewayService_HandleCountTokensSpecialCases_BedrockReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	svc := &GatewayService{}
	handled, err := svc.handleCountTokensSpecialCases(context.Background(), c, &Account{
		Platform: PlatformAnthropic,
		Type:     AccountTypeBedrock,
	}, &ParsedRequest{Body: []byte(`{}`)})

	require.True(t, handled)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.JSONEq(t, `{
		"type":"error",
		"error":{"type":"not_found_error","message":"count_tokens endpoint is not supported for Bedrock"}
	}`, recorder.Body.String())
}

func TestGatewayService_HandleCountTokensSpecialCases_AntigravityReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	svc := &GatewayService{}
	handled, err := svc.handleCountTokensSpecialCases(context.Background(), c, &Account{
		Platform: PlatformAntigravity,
		Type:     AccountTypeOAuth,
	}, &ParsedRequest{Body: []byte(`{}`)})

	require.True(t, handled)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.JSONEq(t, `{
		"type":"error",
		"error":{"type":"not_found_error","message":"count_tokens endpoint is not supported for this platform"}
	}`, recorder.Body.String())
}

func TestGatewayService_HandleCountTokensSpecialCases_AnthropicPassthroughUsesMappedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)

	upstream := &anthropicHTTPUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"input_tokens":12}`)),
		},
	}
	svc := &GatewayService{
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
		httpUpstream: upstream,
	}
	account := &Account{
		ID:          321,
		Name:        "passthrough-route",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "upstream-key",
			"base_url": "https://api.anthropic.com",
			"model_mapping": map[string]any{
				"claude-sonnet-4-*": "claude-sonnet-4-5-20241022",
			},
		},
		Extra:       map[string]any{"anthropic_passthrough": true},
		Status:      StatusActive,
		Schedulable: true,
	}
	parsed := &ParsedRequest{
		Body:  []byte(`{"model":"claude-sonnet-4-20250514","messages":[]}`),
		Model: "claude-sonnet-4-20250514",
	}

	handled, err := svc.handleCountTokensSpecialCases(context.Background(), c, account, parsed)

	require.True(t, handled)
	require.NoError(t, err)
	require.Equal(t, "claude-sonnet-4-5-20241022", gjson.GetBytes(upstream.lastBody, "model").String())
}
