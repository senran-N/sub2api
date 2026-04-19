package handler

import (
	"context"
	"net/http"

	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type compatibleGatewayProtocolHandler interface {
	Messages(*gin.Context)
	Responses(*gin.Context)
	ResponsesWebSocket(*gin.Context)
	ChatCompletions(*gin.Context)
	Passthrough(*gin.Context)
}

// CompatibleGatewayHandler owns the shared OpenAI-compatible HTTP entrypoints.
// OpenAI remains one consumer of this layer rather than the control-plane owner.
type CompatibleGatewayHandler struct {
	sharedRuntime                   *CompatibleGatewayRuntimeHandler
	grokGateway                     *GrokGatewayHandler
	gatewayService                  *service.GatewayService
	compatibleUpstreamModelsService *service.CompatibleUpstreamModelsService
}

func NewCompatibleGatewayHandler(
	sharedRuntime *CompatibleGatewayRuntimeHandler,
	grokGateway *GrokGatewayHandler,
	gatewayService *service.GatewayService,
	compatibleUpstreamModelsService *service.CompatibleUpstreamModelsService,
) *CompatibleGatewayHandler {
	return &CompatibleGatewayHandler{
		sharedRuntime:                   sharedRuntime,
		grokGateway:                     grokGateway,
		gatewayService:                  gatewayService,
		compatibleUpstreamModelsService: compatibleUpstreamModelsService,
	}
}

func writeCompatibleGatewayMisconfigured(c *gin.Context, platform string) {
	if c == nil {
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"type":    "api_error",
			"message": "compatible gateway handler is not configured for platform: " + platform,
		},
	})
}

func (h *CompatibleGatewayHandler) compatibleGatewayPlatform(c *gin.Context) string {
	_, platform := resolveGatewayModelListingContext(c)
	if c != nil && c.Request != nil {
		return service.ResolveCompatibleGatewayPlatform(c.Request.Context(), platform)
	}
	return service.ResolveCompatibleGatewayPlatform(context.TODO(), platform)
}

func (h *CompatibleGatewayHandler) protocolHandler(c *gin.Context) compatibleGatewayProtocolHandler {
	var protocolHandler compatibleGatewayProtocolHandler
	switch h.compatibleGatewayPlatform(c) {
	case service.PlatformGrok:
		if h.grokGateway != nil {
			protocolHandler = h.grokGateway
		}
	case service.PlatformOpenAI:
		if h.sharedRuntime != nil {
			protocolHandler = h.sharedRuntime
		}
	}
	if protocolHandler != nil {
		return protocolHandler
	}
	if h.sharedRuntime != nil {
		return h.sharedRuntime
	}
	if h.grokGateway != nil {
		return h.grokGateway
	}
	return nil
}

func (h *CompatibleGatewayHandler) Models(c *gin.Context) {
	groupID, _ := resolveGatewayModelListingContext(c)
	platform := h.compatibleGatewayPlatform(c)

	if h.compatibleUpstreamModelsService != nil {
		discoveredModels, err := h.compatibleUpstreamModelsService.DiscoverGroupModels(c.Request.Context(), groupID, platform)
		if err == nil && len(discoveredModels) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"object": "list",
				"data":   service.BuildCompatibleGatewayDiscoveredModels(discoveredModels, platform),
			})
			return
		}
	}

	if h.gatewayService != nil {
		availableModels := h.gatewayService.GetAvailableModels(c.Request.Context(), groupID, platform)
		if len(availableModels) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"object": "list",
				"data":   service.BuildCompatibleGatewayMappedModels(availableModels, platform),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   service.DefaultCompatibleGatewayModels(platform),
	})
}

func (h *CompatibleGatewayHandler) Messages(c *gin.Context) {
	protocolHandler := h.protocolHandler(c)
	if protocolHandler == nil {
		writeCompatibleGatewayMisconfigured(c, h.compatibleGatewayPlatform(c))
		return
	}
	protocolHandler.Messages(c)
}

func (h *CompatibleGatewayHandler) Responses(c *gin.Context) {
	protocolHandler := h.protocolHandler(c)
	if protocolHandler == nil {
		writeCompatibleGatewayMisconfigured(c, h.compatibleGatewayPlatform(c))
		return
	}
	protocolHandler.Responses(c)
}

func (h *CompatibleGatewayHandler) ResponsesWebSocket(c *gin.Context) {
	protocolHandler := h.protocolHandler(c)
	if protocolHandler == nil {
		writeCompatibleGatewayMisconfigured(c, h.compatibleGatewayPlatform(c))
		return
	}
	protocolHandler.ResponsesWebSocket(c)
}

func (h *CompatibleGatewayHandler) ChatCompletions(c *gin.Context) {
	protocolHandler := h.protocolHandler(c)
	if protocolHandler == nil {
		writeCompatibleGatewayMisconfigured(c, h.compatibleGatewayPlatform(c))
		return
	}
	protocolHandler.ChatCompletions(c)
}

func (h *CompatibleGatewayHandler) Passthrough(c *gin.Context) {
	protocolHandler := h.protocolHandler(c)
	if protocolHandler == nil {
		writeCompatibleGatewayMisconfigured(c, h.compatibleGatewayPlatform(c))
		return
	}
	protocolHandler.Passthrough(c)
}
