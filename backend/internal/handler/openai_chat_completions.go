package handler

import "github.com/gin-gonic/gin"

// ChatCompletions handles OpenAI Chat Completions API requests.
// POST /v1/chat/completions
func (h *OpenAIGatewayHandler) ChatCompletions(c *gin.Context) {
	compatibleTextHandlerFromOpenAIHandler(h).ChatCompletions(c)
}
