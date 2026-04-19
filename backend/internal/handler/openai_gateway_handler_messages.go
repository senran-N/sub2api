package handler

import "github.com/gin-gonic/gin"

// Messages handles Anthropic Messages API requests routed to OpenAI platform.
// POST /v1/messages (when group platform is OpenAI)
func (h *OpenAIGatewayHandler) Messages(c *gin.Context) {
	compatibleTextHandlerFromOpenAIHandler(h).Messages(c)
}
