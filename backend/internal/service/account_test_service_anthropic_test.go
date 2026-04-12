package service

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/stretchr/testify/require"
)

func TestAccountTestService_AnthropicUpstreamUsesModelMappingAndEndpointOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader("data: {\"type\":\"message_stop\"}\n\n"))

	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}
	account := &Account{
		ID:          77,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeUpstream,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "sk-ant-test",
			"base_url": "https://proxy.example.com/anthropic",
			"endpoint_overrides": map[string]any{
				"messages": "custom/messages?source=test",
			},
			"model_mapping": map[string]any{
				claude.DefaultTestModel: "claude-sonnet-4-6",
			},
		},
	}

	err := svc.testClaudeAccountConnection(ctx, account, "", "")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://proxy.example.com/anthropic/custom/messages?source=test", upstream.requests[0].URL.String())
	require.Equal(t, "sk-ant-test", upstream.requests[0].Header.Get("X-Api-Key"))
	require.Contains(t, string(readAnthropicAccountTestRequestBody(t, upstream.requests[0])), `"model":"claude-sonnet-4-6"`)
}

func readAnthropicAccountTestRequestBody(t *testing.T, req *http.Request) []byte {
	t.Helper()
	if req == nil || req.Body == nil {
		return nil
	}
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	req.Body = io.NopCloser(strings.NewReader(string(body)))
	return body
}
