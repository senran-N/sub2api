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
	feedbackRepo AccountRepository
}

func NewGrokCompatibleRuntime(textExecutor grokCompatibleTextExecutor, feedbackRepo AccountRepository) *GrokCompatibleRuntime {
	return &GrokCompatibleRuntime{textExecutor: textExecutor, feedbackRepo: feedbackRepo}
}

func ProvideGrokCompatibleRuntime(compatibleTextRuntime *CompatibleGatewayTextRuntime) *GrokCompatibleRuntime {
	var feedbackRepo AccountRepository
	if compatibleTextRuntime != nil && compatibleTextRuntime.openaiGatewayService != nil {
		feedbackRepo = compatibleTextRuntime.openaiGatewayService.accountRepo
	}
	return NewGrokCompatibleRuntime(compatibleTextRuntime, feedbackRepo)
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

	var (
		result *OpenAIForwardResult
		err    error
	)
	switch preparation.protocolFamily {
	case CompatibleGatewayProtocolFamilyChatCompletions:
		result, err = r.textExecutor.ForwardChatCompletions(c.Request.Context(), c, preparation.account, preparation.compatibleBody, "", "")
	case CompatibleGatewayProtocolFamilyMessages:
		result, err = r.textExecutor.ForwardMessages(c.Request.Context(), c, preparation.account, preparation.compatibleBody, "", "")
	default:
		result, err = r.textExecutor.ForwardResponses(c.Request.Context(), c, preparation.account, preparation.compatibleBody, "")
	}
	if _, delegated := r.textExecutor.(*CompatibleGatewayTextRuntime); !delegated {
		persistGrokRuntimeFeedbackToRepo(c.Request.Context(), r.feedbackRepo, GrokRuntimeFeedbackInput{
			Account:        preparation.account,
			RequestedModel: preparation.requestedModel,
			Result:         result,
			ProtocolFamily: preparation.protocolFamily,
			Err:            err,
		})
	}
	if err != nil && !c.Writer.Written() {
		writeGrokTextError(c, preparation.protocolFamily, http.StatusBadGateway, "api_error", "Grok upstream request failed")
	}
}
