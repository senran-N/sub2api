package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type compatibleGatewayProtocolHandler interface {
	Messages(*gin.Context)
	Responses(*gin.Context)
	ResponsesWebSocket(*gin.Context)
	ChatCompletions(*gin.Context)
	Passthrough(*gin.Context)
	Images(*gin.Context)
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
	// Compatibility fallback for tests/helpers that call without a request.
	return service.ResolveCompatibleGatewayPlatform(context.Background(), platform)
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

func (h *CompatibleGatewayHandler) GetModel(c *gin.Context) {
	groupID, _ := resolveGatewayModelListingContext(c)
	platform := h.compatibleGatewayPlatform(c)
	if writeCompatibleGatewayModelResponse(
		c,
		strings.TrimSpace(c.Param("model")),
		groupID,
		platform,
		h.compatibleUpstreamModelsService,
		h.gatewayService,
	) {
		return
	}
	writeCompatibleGatewayModelNotFound(c, c.Param("model"))
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

func (h *CompatibleGatewayHandler) Images(c *gin.Context) {
	protocolHandler := h.protocolHandler(c)
	if protocolHandler == nil {
		writeCompatibleGatewayMisconfigured(c, h.compatibleGatewayPlatform(c))
		return
	}
	protocolHandler.Images(c)
}

func writeCompatibleGatewayModelResponse(
	c *gin.Context,
	modelID string,
	groupID *int64,
	platform string,
	compatibleUpstreamModelsService *service.CompatibleUpstreamModelsService,
	gatewayService *service.GatewayService,
) bool {
	if compatibleUpstreamModelsService != nil {
		discoveredModels, err := compatibleUpstreamModelsService.DiscoverGroupModels(c.Request.Context(), groupID, platform)
		if err == nil {
			if model, ok := service.LookupCompatibleGatewayDiscoveredModel(discoveredModels, modelID, platform); ok {
				c.JSON(http.StatusOK, model)
				return true
			}
		}
	}

	if gatewayService != nil {
		availableModels := gatewayService.GetAvailableModels(c.Request.Context(), groupID, platform)
		if model, ok := service.LookupCompatibleGatewayMappedModel(availableModels, modelID, platform); ok {
			c.JSON(http.StatusOK, model)
			return true
		}
	}

	if model, ok := service.LookupCompatibleGatewayDefaultModel(modelID, platform); ok {
		c.JSON(http.StatusOK, model)
		return true
	}

	return false
}

func writeCompatibleGatewayModelNotFound(c *gin.Context, modelID string) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"message": fmt.Sprintf("Model %q not found", strings.TrimSpace(modelID)),
			"type":    "invalid_request_error",
		},
	})
}
