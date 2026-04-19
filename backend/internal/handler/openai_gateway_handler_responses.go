package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Responses handles OpenAI Responses API endpoint.
// POST /openai/v1/responses
func (h *OpenAIGatewayHandler) Responses(c *gin.Context) {
	compatibleTextHandlerFromOpenAIHandler(h).Responses(c)
}

func (h *OpenAIGatewayHandler) logOpenAIRemoteCompactOutcome(c *gin.Context, startedAt time.Time) {
	compatibleTextHandlerFromOpenAIHandler(h).logOpenAIRemoteCompactOutcome(c, startedAt)
}

func isOpenAIRemoteCompactPath(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}
	normalizedPath := strings.TrimRight(strings.TrimSpace(c.Request.URL.Path), "/")
	return strings.HasSuffix(normalizedPath, "/responses/compact")
}

func (h *OpenAIGatewayHandler) recoverResponsesPanic(c *gin.Context, streamStarted *bool) {
	recovered := recover()
	if recovered == nil {
		return
	}
	compatibleTextHandlerFromOpenAIHandler(h).handleRecoveredResponsesPanic(c, streamStarted, recovered)
}
