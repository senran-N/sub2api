package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGatewayServiceGetAccessToken_APIKey(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "sk-test"},
	}

	token, tokenType, err := svc.GetAccessToken(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, "sk-test", token)
	require.Equal(t, "apikey", tokenType)
}

func TestGatewayServiceGetAccessToken_AnthropicOAuthUsesProvider(t *testing.T) {
	svc := &GatewayService{
		claudeTokenProvider: &ClaudeTokenProvider{},
	}
	account := &Account{
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"access_token": "oauth-token"},
	}

	token, tokenType, err := svc.GetAccessToken(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, "oauth-token", token)
	require.Equal(t, "oauth", tokenType)
}

func TestGatewayServiceGetAccessToken_SetupTokenFallsBackToCredential(t *testing.T) {
	svc := &GatewayService{
		claudeTokenProvider: &ClaudeTokenProvider{},
	}
	account := &Account{
		Platform:    PlatformAnthropic,
		Type:        AccountTypeSetupToken,
		Credentials: map[string]any{"access_token": "setup-token"},
	}

	token, tokenType, err := svc.GetAccessToken(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, "setup-token", token)
	require.Equal(t, "oauth", tokenType)
}

func TestGatewayServiceGetAccessToken_UnsupportedType(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{Type: "unknown"}

	_, _, err := svc.GetAccessToken(context.Background(), account)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported account type")
}
