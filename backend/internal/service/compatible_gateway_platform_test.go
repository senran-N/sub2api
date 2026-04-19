package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestDetectCompatibleGatewayPlatform_PrefersForcePlatformThenGroup(t *testing.T) {
	group := &Group{
		ID:       7,
		Platform: PlatformGrok,
		Status:   StatusActive,
		Hydrated: true,
	}
	ctx := context.WithValue(context.Background(), ctxkey.Group, group)
	require.Equal(t, PlatformGrok, DetectCompatibleGatewayPlatform(ctx, PlatformOpenAI))

	ctx = context.WithValue(ctx, ctxkey.ForcePlatform, PlatformOpenAI)
	require.Equal(t, PlatformOpenAI, DetectCompatibleGatewayPlatform(ctx, PlatformGrok))
}

func TestResolveCompatibleGatewayPlatform_DefaultsOpenAIWhenUnset(t *testing.T) {
	require.Equal(t, PlatformOpenAI, ResolveCompatibleGatewayPlatform(context.Background(), ""))
	require.Equal(t, PlatformGrok, ResolveCompatibleGatewayPlatform(context.Background(), PlatformGrok))
}
