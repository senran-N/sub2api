package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type grokCompatibleTextExecutor interface {
	ForwardResponses(context.Context, *gin.Context, *Account, []byte, string) (*OpenAIForwardResult, error)
	ForwardChatCompletions(context.Context, *gin.Context, *Account, []byte, string, string) (*OpenAIForwardResult, error)
	ForwardMessages(context.Context, *gin.Context, *Account, []byte, string, string) (*OpenAIForwardResult, error)
}

type GrokCompatibleRuntime struct {
	textExecutor grokCompatibleTextExecutor
}

func NewGrokCompatibleRuntime(textExecutor grokCompatibleTextExecutor) *GrokCompatibleRuntime {
	return &GrokCompatibleRuntime{textExecutor: textExecutor}
}

func ProvideGrokCompatibleRuntime(compatibleTextRuntime *CompatibleGatewayTextRuntime) *GrokCompatibleRuntime {
	return NewGrokCompatibleRuntime(compatibleTextRuntime)
}

func (r *GrokCompatibleRuntime) Execute(c *gin.Context, preparation *grokTextPreparation) {
	if c == nil {
		return
	}
	if preparation == nil || preparation.account == nil {
		writeResponsesError(c, http.StatusServiceUnavailable, "api_error", "No available Grok compatible accounts")
		return
	}
	if r == nil || r.textExecutor == nil {
		writeGrokTextError(c, preparation.protocolFamily, http.StatusInternalServerError, "api_error", "Grok compatible text executor is not configured")
		return
	}

	var err error
	switch preparation.protocolFamily {
	case CompatibleGatewayProtocolFamilyChatCompletions:
		_, err = r.textExecutor.ForwardChatCompletions(c.Request.Context(), c, preparation.account, preparation.compatibleBody, "", "")
	case CompatibleGatewayProtocolFamilyMessages:
		_, err = r.textExecutor.ForwardMessages(c.Request.Context(), c, preparation.account, preparation.compatibleBody, "", "")
	default:
		_, err = r.textExecutor.ForwardResponses(c.Request.Context(), c, preparation.account, preparation.compatibleBody, "")
	}
	if err != nil && !c.Writer.Written() {
		writeGrokTextError(c, preparation.protocolFamily, http.StatusBadGateway, "api_error", "Grok upstream request failed")
	}
}
