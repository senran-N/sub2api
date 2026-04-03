package routes

import (
	"net/http"

	"github.com/senran-N/sub2api/internal/handler"
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
	if d.usesOpenAIProtocol(c) {
		d.handlers.OpenAIGateway.Messages(c)
		return
	}
	d.handlers.Gateway.Messages(c)
}

func (d gatewayProtocolDispatcher) CountTokens(c *gin.Context) {
	if d.usesOpenAIProtocol(c) {
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
	if d.usesOpenAIProtocol(c) {
		d.handlers.OpenAIGateway.Responses(c)
		return
	}
	d.handlers.Gateway.Responses(c)
}

func (d gatewayProtocolDispatcher) ChatCompletions(c *gin.Context) {
	if d.usesOpenAIProtocol(c) {
		d.handlers.OpenAIGateway.ChatCompletions(c)
		return
	}
	d.handlers.Gateway.ChatCompletions(c)
}

func (d gatewayProtocolDispatcher) usesOpenAIProtocol(c *gin.Context) bool {
	return d.groupPlatform(c) == service.PlatformOpenAI
}

func (d gatewayProtocolDispatcher) groupPlatform(c *gin.Context) string {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey.Group == nil {
		return ""
	}
	return apiKey.Group.Platform
}
