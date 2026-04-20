package routes

import (
	"context"
	"net/http"
	"strings"

	"github.com/senran-N/sub2api/internal/handler"
	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// gatewayProtocolDispatcher centralizes protocol routing decisions so route
// registration stays declarative instead of embedding platform branching inline.
type gatewayProtocolDispatcher struct {
	handlers *handler.Handlers
}

func newGatewayProtocolDispatcher(handlers *handler.Handlers) gatewayProtocolDispatcher {
	return gatewayProtocolDispatcher{handlers: handlers}
}

func (d gatewayProtocolDispatcher) Messages(c *gin.Context) {
	if platform, ok := d.resolveCompatibleGatewayPlatform(c); ok {
		if platform == service.PlatformGrok {
			d.handlers.GrokGateway.Messages(c)
			return
		}
		d.handlers.CompatibleGateway.Messages(c)
		return
	}
	d.handlers.Gateway.Messages(c)
}

func (d gatewayProtocolDispatcher) Models(c *gin.Context) {
	if platform, ok := d.resolveCompatibleGatewayPlatform(c); ok {
		if platform == service.PlatformGrok {
			d.handlers.GrokGateway.Models(c)
			return
		}
		d.handlers.CompatibleGateway.Models(c)
		return
	}
	d.handlers.Gateway.Models(c)
}

func (d gatewayProtocolDispatcher) GetModel(c *gin.Context) {
	if platform, ok := d.resolveCompatibleGatewayPlatform(c); ok {
		if platform == service.PlatformGrok {
			d.handlers.GrokGateway.GetModel(c)
			return
		}
		d.handlers.CompatibleGateway.GetModel(c)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"type":    "not_found_error",
			"message": "The requested endpoint is not available for this platform",
		},
	})
}

func (d gatewayProtocolDispatcher) CountTokens(c *gin.Context) {
	if d.usesCompatibleGatewayProtocol(c) {
		c.JSON(http.StatusNotFound, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "not_found_error",
				"message": "Token counting is not supported for this platform",
			},
		})
		return
	}
	d.handlers.Gateway.CountTokens(c)
}

func (d gatewayProtocolDispatcher) Responses(c *gin.Context) {
	if platform, ok := d.resolveCompatibleGatewayPlatform(c); ok {
		if platform == service.PlatformGrok {
			d.handlers.GrokGateway.Responses(c)
			return
		}
		d.handlers.CompatibleGateway.Responses(c)
		return
	}
	d.handlers.Gateway.Responses(c)
}

func (d gatewayProtocolDispatcher) ChatCompletions(c *gin.Context) {
	if platform, ok := d.resolveCompatibleGatewayPlatform(c); ok {
		if platform == service.PlatformGrok {
			d.handlers.GrokGateway.ChatCompletions(c)
			return
		}
		d.handlers.CompatibleGateway.ChatCompletions(c)
		return
	}
	d.handlers.Gateway.ChatCompletions(c)
}

func (d gatewayProtocolDispatcher) OpenAICompatiblePassthrough(c *gin.Context) {
	if platform, ok := d.resolveCompatibleGatewayPlatform(c); ok {
		if platform == service.PlatformGrok {
			switch handler.GetInboundEndpoint(c) {
			case handler.EndpointImages:
				d.handlers.GrokGateway.Images(c)
			case handler.EndpointVideos:
				d.handlers.GrokGateway.Videos(c)
			default:
				d.handlers.GrokGateway.Passthrough(c)
			}
			return
		}
		d.handlers.CompatibleGateway.Passthrough(c)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"type":    "not_found_error",
			"message": "The requested endpoint is not available for this platform",
		},
	})
}

func (d gatewayProtocolDispatcher) resolveCompatibleGatewayPlatform(c *gin.Context) (string, bool) {
	platform := d.effectiveCompatiblePlatform(c)
	if platform == "" {
		return "", false
	}
	d.bindCompatiblePlatform(c, platform)
	return platform, true
}

func (d gatewayProtocolDispatcher) usesCompatibleGatewayProtocol(c *gin.Context) bool {
	_, ok := d.resolveCompatibleGatewayPlatform(c)
	return ok
}

func (d gatewayProtocolDispatcher) effectiveCompatiblePlatform(c *gin.Context) string {
	if forcedPlatform, ok := middleware.GetForcePlatformFromContext(c); ok {
		if platform := service.NormalizeCompatibleGatewayPlatform(forcedPlatform); platform != "" {
			return platform
		}
	}
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey.Group == nil {
		return ""
	}
	return service.NormalizeCompatibleGatewayPlatform(apiKey.Group.Platform)
}

func (d gatewayProtocolDispatcher) bindCompatiblePlatform(c *gin.Context, platform string) {
	if c == nil || c.Request == nil {
		return
	}
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.ForcePlatform, platform)
	c.Request = c.Request.WithContext(ctx)
	c.Set(string(middleware.ContextKeyForcePlatform), platform)
}
