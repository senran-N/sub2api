package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayServiceGetAccessToken_GrokSessionRequiresProviderOwnedTransport(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
	}

	_, _, err := svc.GetAccessToken(context.Background(), account)

	require.EqualError(t, err, "unsupported compatible shared runtime account type for grok: session")
}
