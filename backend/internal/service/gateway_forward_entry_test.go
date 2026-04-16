package service

import (
	"context"
	"encoding/json"
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

func TestGatewayService_ResolveForwardPassthroughInput_AppliesMappedModel(t *testing.T) {
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
		Body:   []byte(`{"model":"claude-sonnet-4-20250514","stream":true,"messages":[]}`),
		Model:  "claude-sonnet-4-20250514",
		Stream: true,
	}

	input := svc.resolveForwardPassthroughInput(account, parsed, time.Unix(123, 0))

	require.Equal(t, "claude-sonnet-4-5-20241022", input.RequestModel)
	require.Equal(t, "claude-sonnet-4-20250514", input.OriginalModel)
	require.True(t, input.RequestStream)
	require.Equal(t, int64(123), input.StartTime.Unix())
	require.Equal(t, "claude-sonnet-4-5-20241022", gjson.GetBytes(input.Body, "model").String())
}

func TestGatewayService_HandleForwardSpecialCases_PassthroughHandled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	upstream := &anthropicHTTPUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}, "x-request-id": []string{"rid-forward-pass"}},
			Body: io.NopCloser(strings.NewReader(
				`{"id":"msg_1","type":"message","usage":{"input_tokens":3,"output_tokens":2}}`,
			)),
		},
	}
	svc := &GatewayService{
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
		httpUpstream:         upstream,
		rateLimitService:     &RateLimitService{},
		responseHeaderFilter: compileResponseHeaderFilter(&config.Config{}),
	}
	account := newAnthropicAPIKeyAccountForTest()
	account.Credentials["model_mapping"] = map[string]any{
		"claude-3-5-sonnet-latest": "claude-3-opus-20240229",
	}
	parsed := &ParsedRequest{
		Body:  []byte(`{"model":"claude-3-5-sonnet-latest","messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]}`),
		Model: "claude-3-5-sonnet-latest",
	}

	handled, result, err := svc.handleForwardSpecialCases(context.Background(), c, account, parsed, time.Now())

	require.True(t, handled)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "claude-3-5-sonnet-latest", result.Model)
	require.Equal(t, "claude-3-opus-20240229", result.UpstreamModel)
	require.Equal(t, "claude-3-opus-20240229", gjson.GetBytes(upstream.lastBody, "model").String())
}

func TestGatewayService_HandleForwardSpecialCases_BedrockHandled(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Platform: PlatformAnthropic,
		Type:     AccountTypeBedrock,
	}
	parsed := &ParsedRequest{
		Body:  []byte(`{"model":"unsupported-bedrock-model","messages":[]}`),
		Model: "unsupported-bedrock-model",
	}

	handled, result, err := svc.handleForwardSpecialCases(context.Background(), nil, account, parsed, time.Now())

	require.True(t, handled)
	require.Nil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported bedrock model")
}

func TestGatewayService_InitializeForwardBetaPolicy_CachesFilterSet(t *testing.T) {
	settings := &BetaPolicySettings{
		Rules: []BetaPolicyRule{
			{
				BetaToken: "fine-grained-tool-streaming-2025-05-14",
				Action:    BetaPolicyActionFilter,
				Scope:     BetaPolicyScopeAll,
			},
		},
	}
	raw, err := json.Marshal(settings)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	svc := &GatewayService{
		settingService: NewSettingService(
			&betaPolicySettingRepoStub{values: map[string]string{
				SettingKeyBetaPolicySettings: string(raw),
			}},
			&config.Config{},
		),
	}

	err = svc.initializeForwardBetaPolicy(context.Background(), c, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, "")

	require.NoError(t, err)
	value, ok := c.Get(betaPolicyFilterSetKey)
	require.True(t, ok)
	filterSet, ok := value.(map[string]struct{})
	require.True(t, ok)
	_, exists := filterSet["fine-grained-tool-streaming-2025-05-14"]
	require.True(t, exists)
}

func TestGatewayService_InitializeForwardBetaPolicy_ReturnsBlockError(t *testing.T) {
	settings := &BetaPolicySettings{
		Rules: []BetaPolicyRule{
			{
				BetaToken:    "interleaved-thinking-2025-05-14",
				Action:       BetaPolicyActionBlock,
				Scope:        BetaPolicyScopeAll,
				ErrorMessage: "thinking blocked",
			},
		},
	}
	raw, err := json.Marshal(settings)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set("anthropic-beta", "interleaved-thinking-2025-05-14")

	svc := &GatewayService{
		settingService: NewSettingService(
			&betaPolicySettingRepoStub{values: map[string]string{
				SettingKeyBetaPolicySettings: string(raw),
			}},
			&config.Config{},
		),
	}

	err = svc.initializeForwardBetaPolicy(context.Background(), c, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, "")

	require.Error(t, err)
	require.Equal(t, "thinking blocked", err.Error())
	_, exists := c.Get(betaPolicyFilterSetKey)
	require.False(t, exists)
}

func TestGatewayService_InitializeForwardBetaPolicy_RespectsModelWhitelist(t *testing.T) {
	settings := &BetaPolicySettings{
		Rules: []BetaPolicyRule{
			{
				BetaToken:      "context-1m-2025-08-07",
				Action:         BetaPolicyActionFilter,
				Scope:          BetaPolicyScopeAll,
				ModelWhitelist: []string{"claude-opus-*"},
			},
		},
	}
	raw, err := json.Marshal(settings)
	require.NoError(t, err)

	svc := &GatewayService{
		settingService: NewSettingService(
			&betaPolicySettingRepoStub{values: map[string]string{
				SettingKeyBetaPolicySettings: string(raw),
			}},
			&config.Config{},
		),
	}

	makeContext := func() *gin.Context {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
		return c
	}

	opusContext := makeContext()
	err = svc.initializeForwardBetaPolicy(context.Background(), opusContext, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, "claude-opus-4-1")
	require.NoError(t, err)
	opusFilterSet, _ := opusContext.Get(betaPolicyFilterSetKey)
	_, exists := opusFilterSet.(map[string]struct{})["context-1m-2025-08-07"]
	require.True(t, exists)

	sonnetContext := makeContext()
	err = svc.initializeForwardBetaPolicy(context.Background(), sonnetContext, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, "claude-sonnet-4-1")
	require.NoError(t, err)
	sonnetFilterSet, _ := sonnetContext.Get(betaPolicyFilterSetKey)
	_, exists = sonnetFilterSet.(map[string]struct{})["context-1m-2025-08-07"]
	require.False(t, exists)
}
