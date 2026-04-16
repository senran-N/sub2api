package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGatewayService_ApplyForwardModelMapping_APIKeyUsesAccountMapping(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Name:     "api-key-account",
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"claude-sonnet-4-*": "claude-sonnet-4-5-20241022",
			},
		},
	}
	body := []byte(`{"model":"claude-sonnet-4-20250514","messages":[]}`)

	gotBody, gotModel := svc.applyForwardModelMapping(account, "claude-sonnet-4-20250514", "claude-sonnet-4-20250514", body)

	require.Equal(t, "claude-sonnet-4-5-20241022", gotModel)
	require.Equal(t, "claude-sonnet-4-5-20241022", gjson.GetBytes(gotBody, "model").String())
}

func TestGatewayService_ApplyForwardModelMapping_APIKeyUsesReasoningVariantFallback(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Name:     "api-key-openai-compat",
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.4": "claude-sonnet-4-5-20250929",
			},
		},
	}
	body := []byte(`{"model":"gpt-5.4-xhigh","messages":[]}`)

	gotBody, gotModel := svc.applyForwardModelMapping(account, "gpt-5.4-xhigh", "gpt-5.4-xhigh", body)

	require.Equal(t, "claude-sonnet-4-5-20250929", gotModel)
	require.Equal(t, "claude-sonnet-4-5-20250929", gjson.GetBytes(gotBody, "model").String())
}

func TestGatewayService_ApplyForwardModelMapping_UpstreamUsesReasoningVariantFallback(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Name:     "upstream-openai-compat",
		Platform: PlatformAnthropic,
		Type:     AccountTypeUpstream,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.4": "claude-sonnet-4-5-20250929",
			},
		},
	}
	body := []byte(`{"model":"gpt-5.4-xhigh","messages":[]}`)

	gotBody, gotModel := svc.applyForwardModelMapping(account, "gpt-5.4-xhigh", "gpt-5.4-xhigh", body)

	require.Equal(t, "claude-sonnet-4-5-20250929", gotModel)
	require.Equal(t, "claude-sonnet-4-5-20250929", gjson.GetBytes(gotBody, "model").String())
}

func TestGatewayService_ApplyForwardModelMapping_OAuthUsesClaudeNormalization(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Name:     "oauth-account",
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"claude-sonnet-4-5","messages":[]}`)

	gotBody, gotModel := svc.applyForwardModelMapping(account, "claude-sonnet-4-5", "claude-sonnet-4-5", body)

	require.Equal(t, "claude-sonnet-4-5-20250929", gotModel)
	require.Equal(t, "claude-sonnet-4-5-20250929", gjson.GetBytes(gotBody, "model").String())
}

func TestGatewayService_ResolveForwardProxyURL_SkipsProxyWhenCustomBaseURLEnabled(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		ProxyID: func() *int64 {
			id := int64(1)
			return &id
		}(),
		Proxy: &domain.Proxy{
			Protocol: "http",
			Host:     "127.0.0.1",
			Port:     8080,
		},
		Extra: map[string]any{
			"custom_base_url_enabled": true,
			"custom_base_url":         "https://relay.example.com",
		},
	}

	require.Empty(t, svc.resolveForwardProxyURL(account))
}

func TestGatewayService_ResolveForwardProxyURL_UsesProxyForStandardUpstream(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		ProxyID: func() *int64 {
			id := int64(2)
			return &id
		}(),
		Proxy: &domain.Proxy{
			Protocol: "socks5",
			Host:     "127.0.0.1",
			Port:     1080,
		},
	}

	require.Equal(t, "socks5://127.0.0.1:1080", svc.resolveForwardProxyURL(account))
}

func TestGatewayService_NormalizeForwardOAuthRequestBody_ReusesSingleFingerprintLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storeGatewayForwardingCache(true, false, false, time.Minute)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set("User-Agent", "claude-cli/2.1.88 (external, cli)")

	cache := &identityCacheStub{
		fingerprint: &Fingerprint{
			ClientID:  "clientid123",
			UserAgent: "claude-cli/2.1.88 (external, cli)",
		},
	}
	svc := &GatewayService{
		identityService: NewIdentityService(cache),
		settingService:  NewSettingService(&gatewayForwardingRepoStub{}, &config.Config{}),
	}
	account := &Account{
		ID:       123,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"account_uuid": "acc-uuid",
		},
	}
	parsed := &ParsedRequest{
		Model: "claude-sonnet-4-5",
		Messages: []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "hello from cli",
					},
				},
			},
		},
	}
	body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":[{"type":"text","text":"hello from cli"}]}]}`)

	gotBody, gotModel := svc.normalizeForwardOAuthRequestBody(context.Background(), c, account, parsed, body, parsed.Model)

	require.Equal(t, "claude-sonnet-4-5-20250929", gotModel)
	require.Equal(t, 1, cache.getFingerprintCount)
	require.Equal(t, buildAttributionHeaderText(body, cache.fingerprint.UserAgent), gjson.GetBytes(gotBody, "system.0.text").String())

	metadataUserID := gjson.GetBytes(gotBody, "metadata.user_id").String()
	require.NotEmpty(t, metadataUserID)
	parsedUserID := ParseMetadataUserID(metadataUserID)
	require.NotNil(t, parsedUserID)
	require.True(t, parsedUserID.IsNewFormat)
	require.Equal(t, "clientid123", parsedUserID.DeviceID)
	require.Equal(t, "acc-uuid", parsedUserID.AccountUUID)
}
