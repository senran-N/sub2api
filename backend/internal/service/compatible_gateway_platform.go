package service

import (
	"context"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
)

func NormalizeCompatibleGatewayPlatform(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case PlatformOpenAI:
		return PlatformOpenAI
	case PlatformGrok:
		return PlatformGrok
	default:
		return ""
	}
}

func IsCompatibleGatewayPlatform(platform string) bool {
	return NormalizeCompatibleGatewayPlatform(platform) != ""
}

func CompatibleGatewayDefaultBaseURL(platform string) string {
	switch NormalizeCompatibleGatewayPlatform(platform) {
	case PlatformOpenAI:
		return "https://api.openai.com"
	case PlatformGrok:
		return "https://api.x.ai"
	default:
		return ""
	}
}

func DetectCompatibleGatewayPlatform(ctx context.Context, fallback string) string {
	if ctx != nil {
		if forced, ok := ctx.Value(ctxkey.ForcePlatform).(string); ok {
			if platform := NormalizeCompatibleGatewayPlatform(forced); platform != "" {
				return platform
			}
		}
		if group, ok := ctx.Value(ctxkey.Group).(*Group); ok && IsGroupContextValid(group) {
			if platform := NormalizeCompatibleGatewayPlatform(group.Platform); platform != "" {
				return platform
			}
		}
	}
	return NormalizeCompatibleGatewayPlatform(fallback)
}

func ResolveCompatibleGatewayPlatform(ctx context.Context, fallback string) string {
	if platform := DetectCompatibleGatewayPlatform(ctx, fallback); platform != "" {
		return platform
	}
	return PlatformOpenAI
}
