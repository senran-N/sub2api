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

func TestGatewayService_CountTokensError_WritesAnthropicErrorShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	svc := &GatewayService{}
	svc.countTokensError(c, http.StatusBadGateway, "upstream_error", "Request failed")

	require.Equal(t, http.StatusBadGateway, recorder.Code)
	require.JSONEq(t, `{
		"type":"error",
		"error":{"type":"upstream_error","message":"Request failed"}
	}`, recorder.Body.String())
}

func TestGatewayService_ForwardCountTokensAnthropicAPIKeyPassthrough_SuccessUsesFilteredHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)

	upstream := &anthropicHTTPUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusCreated,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-Request-Id": []string{"rid-count-pass"},
				"Set-Cookie":   []string{"secret=upstream"},
			},
			Body: io.NopCloser(strings.NewReader(`{"input_tokens":42}`)),
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

	err := svc.forwardCountTokensAnthropicAPIKeyPassthrough(
		context.Background(),
		c,
		newAnthropicAPIKeyAccountForTest(),
		[]byte(`{"model":"claude-sonnet-4-5-20250929"}`),
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	require.Equal(t, "rid-count-pass", recorder.Header().Get("x-request-id"))
	require.Empty(t, recorder.Header().Get("Set-Cookie"))
	require.JSONEq(t, `{"input_tokens":42}`, recorder.Body.String())
}

func TestGatewayService_ForwardCountTokensAnthropicAPIKeyPassthrough_Unsupported404ReturnsNotFoundWithoutError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)

	upstream := &anthropicHTTPUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(
				`{"error":{"message":"Not found: /v1/messages/count_tokens","type":"not_found_error"}}`,
			)),
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

	err := svc.forwardCountTokensAnthropicAPIKeyPassthrough(
		context.Background(),
		c,
		newAnthropicAPIKeyAccountForTest(),
		[]byte(`{"model":"claude-sonnet-4-5-20250929"}`),
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.JSONEq(t, `{
		"type":"error",
		"error":{"type":"not_found_error","message":"count_tokens endpoint is not supported by upstream"}
	}`, recorder.Body.String())
}

func TestGatewayService_ForwardCountTokensAnthropicAPIKeyPassthrough_RejectsNonAPIKeyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)

	svc := &GatewayService{}
	account := &Account{
		ID:          999,
		Name:        "wrong-token-type",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token": "oauth-token",
		},
		Status:      StatusActive,
		Schedulable: true,
	}

	err := svc.forwardCountTokensAnthropicAPIKeyPassthrough(context.Background(), c, account, []byte(`{}`))

	require.Error(t, err)
	require.Contains(t, err.Error(), "requires apikey token")
	require.Equal(t, http.StatusBadGateway, recorder.Code)
	require.JSONEq(t, `{
		"type":"error",
		"error":{"type":"upstream_error","message":"Invalid account token type"}
	}`, recorder.Body.String())
}
