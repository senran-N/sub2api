package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/senran-N/sub2api/internal/pkg/httputil"
	servermiddleware "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// GrokGatewayHandler is the Grok-owned compatible control-plane entrypoint.
// Today it reuses the shared compatible runtime for apikey/upstream accounts.
// Future provider-owned session transport should branch here instead of growing
// more Grok ownership inside OpenAI-named handlers.
type GrokGatewayHandler struct {
	sharedRuntime                   *CompatibleGatewayRuntimeHandler
	gatewayService                  *service.GatewayService
	grokGatewayService              *service.GrokGatewayService
	compatibleUpstreamModelsService *service.CompatibleUpstreamModelsService
}

func NewGrokGatewayHandler(
	sharedRuntime *CompatibleGatewayRuntimeHandler,
	gatewayService *service.GatewayService,
	grokGatewayService *service.GrokGatewayService,
	compatibleUpstreamModelsService *service.CompatibleUpstreamModelsService,
) *GrokGatewayHandler {
	return &GrokGatewayHandler{
		sharedRuntime:                   sharedRuntime,
		gatewayService:                  gatewayService,
		grokGatewayService:              grokGatewayService,
		compatibleUpstreamModelsService: compatibleUpstreamModelsService,
	}
}

func (h *GrokGatewayHandler) Models(c *gin.Context) {
	h.withForcedGrokPlatform(c)
	groupID, _ := resolveGatewayModelListingContext(c)

	if h.compatibleUpstreamModelsService != nil {
		discoveredModels, err := h.compatibleUpstreamModelsService.DiscoverGroupModels(c.Request.Context(), groupID, service.PlatformGrok)
		if err == nil && len(discoveredModels) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"object": "list",
				"data":   service.BuildCompatibleGatewayDiscoveredModels(discoveredModels, service.PlatformGrok),
			})
			return
		}
	}

	if h.gatewayService != nil {
		availableModels := h.gatewayService.GetAvailableModels(c.Request.Context(), groupID, service.PlatformGrok)
		if len(availableModels) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"object": "list",
				"data":   service.BuildCompatibleGatewayMappedModels(availableModels, service.PlatformGrok),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   service.DefaultCompatibleGatewayModels(service.PlatformGrok),
	})
}

func (h *GrokGatewayHandler) Messages(c *gin.Context) {
	h.withForcedGrokTextRuntime(c)
	if h == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	body, groupID, ok := h.prepareTextRouteBody(c)
	if !ok {
		return
	}
	if h.grokGatewayService != nil && h.grokGatewayService.HandleMessages(c, groupID, body) {
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if h.sharedRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	h.sharedRuntime.Messages(c)
}

func (h *GrokGatewayHandler) Responses(c *gin.Context) {
	h.withForcedGrokTextRuntime(c)
	if h == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	body, groupID, ok := h.prepareTextRouteBody(c)
	if !ok {
		return
	}

	if h.grokGatewayService != nil && h.grokGatewayService.HandleResponses(c, groupID, body) {
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if h.sharedRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	h.sharedRuntime.Responses(c)
}

func (h *GrokGatewayHandler) ResponsesWebSocket(c *gin.Context) {
	h.withForcedGrokPlatform(c)
	if h == nil || h.sharedRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	h.sharedRuntime.ResponsesWebSocket(c)
}

func (h *GrokGatewayHandler) ChatCompletions(c *gin.Context) {
	h.withForcedGrokTextRuntime(c)
	if h == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	body, groupID, ok := h.prepareTextRouteBody(c)
	if !ok {
		return
	}
	if h.grokGatewayService != nil && h.grokGatewayService.HandleChatCompletions(c, groupID, body) {
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if h.sharedRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	h.sharedRuntime.ChatCompletions(c)
}

func (h *GrokGatewayHandler) Passthrough(c *gin.Context) {
	h.withForcedGrokPlatform(c)
	if h == nil || h.sharedRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	h.sharedRuntime.Passthrough(c)
}

func (h *GrokGatewayHandler) Images(c *gin.Context) {
	h.withForcedGrokPlatform(c)
	if h == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	body, groupID, ok := h.preparePassthroughRouteBody(c)
	if !ok {
		return
	}
	if h.grokGatewayService != nil && h.grokGatewayService.HandleImages(c, groupID, body) {
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if h.sharedRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	h.sharedRuntime.Passthrough(c)
}

func (h *GrokGatewayHandler) Videos(c *gin.Context) {
	h.withForcedGrokPlatform(c)
	if h == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	body, groupID, ok := h.preparePassthroughRouteBody(c)
	if !ok {
		return
	}
	if h.grokGatewayService != nil && h.grokGatewayService.HandleVideos(c, groupID, body) {
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if h.sharedRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	h.sharedRuntime.Passthrough(c)
}

func (h *GrokGatewayHandler) MediaAssetContent(c *gin.Context) {
	if h == nil || h.grokGatewayService == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return
	}
	if h.grokGatewayService.HandleMediaAssetContent(c, c.Param("assetID")) {
		return
	}
	writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
}

func (h *GrokGatewayHandler) withForcedGrokPlatform(c *gin.Context) {
	h.withForcedGrokRequestContext(c, false)
}

func (h *GrokGatewayHandler) withForcedGrokTextRuntime(c *gin.Context) {
	h.withForcedGrokRequestContext(c, true)
}

func (h *GrokGatewayHandler) withForcedGrokRequestContext(c *gin.Context, allowSessionTextRuntime bool) {
	if c == nil || c.Request == nil {
		return
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.ForcePlatform, service.PlatformGrok)
	if allowSessionTextRuntime {
		ctx = service.WithGrokSessionTextRuntimeAllowed(ctx)
	}
	c.Request = c.Request.WithContext(ctx)
	c.Set(string(servermiddleware.ContextKeyForcePlatform), service.PlatformGrok)
}

func (h *GrokGatewayHandler) prepareTextRouteBody(c *gin.Context) ([]byte, *int64, bool) {
	if c == nil || c.Request == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return nil, nil, false
	}

	body, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "invalid_request_error",
				"message": "Failed to read request body",
			},
		})
		return nil, nil, false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	apiKey, _ := servermiddleware.GetAPIKeyFromContext(c)
	if apiKey == nil {
		return body, nil, true
	}
	return body, apiKey.GroupID, true
}

func (h *GrokGatewayHandler) preparePassthroughRouteBody(c *gin.Context) ([]byte, *int64, bool) {
	if c == nil || c.Request == nil {
		writeCompatibleGatewayMisconfigured(c, service.PlatformGrok)
		return nil, nil, false
	}

	body, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "invalid_request_error",
				"message": "Failed to read request body",
			},
		})
		return nil, nil, false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	apiKey, _ := servermiddleware.GetAPIKeyFromContext(c)
	if apiKey == nil {
		return body, nil, true
	}
	return body, apiKey.GroupID, true
}
