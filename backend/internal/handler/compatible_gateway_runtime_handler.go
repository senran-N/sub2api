package handler

import (
	"context"

	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type compatibleGatewayPassthroughRuntime interface {
	ResponsesWebSocket(*gin.Context)
	Passthrough(*gin.Context)
}

// CompatibleGatewayRuntimeHandler is the shared protocol runtime for
// OpenAI-compatible request execution. Today it reuses the existing
// OpenAI-compatible implementation, but the shared control-plane depends on a
// neutral protocol runtime seam instead of an OpenAI-owned handler type.
type CompatibleGatewayRuntimeHandler struct {
	textRuntime        *CompatibleGatewayTextHandler
	passthroughRuntime compatibleGatewayPassthroughRuntime
}

func NewCompatibleGatewayRuntimeHandler(
	textRuntime *CompatibleGatewayTextHandler,
	passthroughRuntime compatibleGatewayPassthroughRuntime,
) *CompatibleGatewayRuntimeHandler {
	return &CompatibleGatewayRuntimeHandler{
		textRuntime:        textRuntime,
		passthroughRuntime: passthroughRuntime,
	}
}

func (h *CompatibleGatewayRuntimeHandler) runtimePlatform(c *gin.Context) string {
	_, platform := resolveGatewayModelListingContext(c)
	if c != nil && c.Request != nil {
		return service.ResolveCompatibleGatewayPlatform(c.Request.Context(), platform)
	}
	// Compatibility fallback for tests/helpers that call without a request.
	return service.ResolveCompatibleGatewayPlatform(context.Background(), platform)
}

func (h *CompatibleGatewayRuntimeHandler) Messages(c *gin.Context) {
	if h == nil || h.textRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, h.runtimePlatform(c))
		return
	}
	h.textRuntime.Messages(c)
}

func (h *CompatibleGatewayRuntimeHandler) Responses(c *gin.Context) {
	if h == nil || h.textRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, h.runtimePlatform(c))
		return
	}
	h.textRuntime.Responses(c)
}

func (h *CompatibleGatewayRuntimeHandler) ResponsesWebSocket(c *gin.Context) {
	if h == nil || h.passthroughRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, h.runtimePlatform(c))
		return
	}
	h.passthroughRuntime.ResponsesWebSocket(c)
}

func (h *CompatibleGatewayRuntimeHandler) ChatCompletions(c *gin.Context) {
	if h == nil || h.textRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, h.runtimePlatform(c))
		return
	}
	h.textRuntime.ChatCompletions(c)
}

func (h *CompatibleGatewayRuntimeHandler) Passthrough(c *gin.Context) {
	if h == nil || h.passthroughRuntime == nil {
		writeCompatibleGatewayMisconfigured(c, h.runtimePlatform(c))
		return
	}
	h.passthroughRuntime.Passthrough(c)
}
