package service

import (
	"testing"

	"github.com/senran-N/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGatewayService_ApplyCountTokensModelMapping_APIKeyUsesAccountMapping(t *testing.T) {
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

	gotBody, gotModel := svc.applyCountTokensModelMapping(account, "claude-sonnet-4-20250514", "claude-sonnet-4-20250514", body)

	require.Equal(t, "claude-sonnet-4-5-20241022", gotModel)
	require.Equal(t, "claude-sonnet-4-5-20241022", gjson.GetBytes(gotBody, "model").String())
}

func TestGatewayService_ApplyCountTokensModelMapping_UpstreamUsesReasoningVariantFallback(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Name:     "upstream-account",
		Platform: PlatformAnthropic,
		Type:     AccountTypeUpstream,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.4": "claude-sonnet-4-5-20241022",
			},
		},
	}
	body := []byte(`{"model":"gpt-5.4-xhigh","messages":[]}`)

	gotBody, gotModel := svc.applyCountTokensModelMapping(account, "gpt-5.4-xhigh", "gpt-5.4-xhigh", body)

	require.Equal(t, "claude-sonnet-4-5-20241022", gotModel)
	require.Equal(t, "claude-sonnet-4-5-20241022", gjson.GetBytes(gotBody, "model").String())
}

func TestGatewayService_ApplyCountTokensModelMapping_OAuthUsesClaudeNormalization(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Name:     "oauth-account",
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"claude-sonnet-4-5","messages":[]}`)

	gotBody, gotModel := svc.applyCountTokensModelMapping(account, "claude-sonnet-4-5", "claude-sonnet-4-5", body)

	require.Equal(t, "claude-sonnet-4-5-20250929", gotModel)
	require.Equal(t, "claude-sonnet-4-5-20250929", gjson.GetBytes(gotBody, "model").String())
}

func TestGatewayService_ResolveCountTokensProxyURL_SkipsDirectProxyWhenCustomBaseURLEnabled(t *testing.T) {
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

	require.Empty(t, svc.resolveCountTokensProxyURL(account))
}

func TestGatewayService_ResolveCountTokensProxyURL_UsesProxyForStandardUpstream(t *testing.T) {
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

	require.Equal(t, "socks5://127.0.0.1:1080", svc.resolveCountTokensProxyURL(account))
}
