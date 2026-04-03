package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetryBackoffDelay(t *testing.T) {
	require.Equal(t, retryBaseDelay, retryBackoffDelay(0))
	require.Equal(t, retryBaseDelay, retryBackoffDelay(1))
	require.Equal(t, 600*time.Millisecond, retryBackoffDelay(2))
	require.Equal(t, retryMaxDelay, retryBackoffDelay(10))
}

func TestSleepWithContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sleepWithContext(ctx, time.Second)

	require.ErrorIs(t, err, context.Canceled)
}

func TestGatewayServiceShouldRetryUpstreamError(t *testing.T) {
	svc := &GatewayService{}

	oauthAccount := &Account{Type: AccountTypeOAuth}
	require.True(t, svc.shouldRetryUpstreamError(oauthAccount, 403))
	require.False(t, svc.shouldRetryUpstreamError(oauthAccount, 429))

	apiKeyAccount := &Account{
		Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"custom_error_codes_enabled": true,
			"custom_error_codes":         []any{float64(401), float64(403)},
		},
	}
	require.True(t, svc.shouldRetryUpstreamError(apiKeyAccount, 429))
	require.False(t, svc.shouldRetryUpstreamError(apiKeyAccount, 401))
}
