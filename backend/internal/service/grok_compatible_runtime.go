package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type grokCompatibleTextExecutor interface {
	ForwardResponses(context.Context, *gin.Context, *Account, []byte, string) (*OpenAIForwardResult, error)
	ForwardChatCompletions(context.Context, *gin.Context, *Account, []byte, string, string) (*OpenAIForwardResult, error)
	ForwardMessages(context.Context, *gin.Context, *Account, []byte, string, string) (*OpenAIForwardResult, error)
}

type GrokCompatibleRuntime struct {
	textExecutor            grokCompatibleTextExecutor
	feedbackRepo            AccountRepository
	feedbackOwnedByExecutor bool
}

func NewGrokCompatibleRuntime(textExecutor grokCompatibleTextExecutor, feedbackRepo AccountRepository) *GrokCompatibleRuntime {
	return NewGrokCompatibleRuntimeWithFeedbackOwnership(textExecutor, feedbackRepo, false)
}

func NewGrokCompatibleRuntimeWithFeedbackOwnership(
	textExecutor grokCompatibleTextExecutor,
	feedbackRepo AccountRepository,
	feedbackOwnedByExecutor bool,
) *GrokCompatibleRuntime {
	return &GrokCompatibleRuntime{
		textExecutor:            textExecutor,
		feedbackRepo:            feedbackRepo,
		feedbackOwnedByExecutor: feedbackOwnedByExecutor,
	}
}

func ProvideGrokCompatibleRuntime(compatibleTextRuntime *CompatibleGatewayTextRuntime) *GrokCompatibleRuntime {
	var feedbackRepo AccountRepository
	feedbackOwnedByExecutor := false
	if compatibleTextRuntime != nil {
		feedbackRepo = compatibleTextRuntime.FeedbackRepository()
		feedbackOwnedByExecutor = compatibleTextRuntime.OwnsCompatibleGatewayRuntimeFeedback()
	}
	return NewGrokCompatibleRuntimeWithFeedbackOwnership(compatibleTextRuntime, feedbackRepo, feedbackOwnedByExecutor)
}

func (r *GrokCompatibleRuntime) Execute(c *gin.Context, preparation *grokTextPreparation) error {
	if c == nil {
		return nil
	}
	if preparation == nil || preparation.account == nil {
		writeResponsesError(c, http.StatusServiceUnavailable, "api_error", "No available Grok compatible accounts")
		return nil
	}
	if r == nil || r.textExecutor == nil {
		writeGrokTextError(c, preparation.protocolFamily, http.StatusInternalServerError, "api_error", "Grok compatible text executor is not configured")
		return nil
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
	if !r.feedbackOwnedByExecutor {
		persistGrokRuntimeFeedbackToRepo(c.Request.Context(), r.feedbackRepo, GrokRuntimeFeedbackInput{
			Account:        preparation.account,
			RequestedModel: preparation.requestedModel,
			Result:         result,
			ProtocolFamily: preparation.protocolFamily,
			Err:            err,
		})
	}
	if err != nil && !c.Writer.Written() {
		var failoverErr *UpstreamFailoverError
		if errors.As(err, &failoverErr) {
			return err
		}
		writeGrokTextError(c, preparation.protocolFamily, http.StatusBadGateway, "api_error", "Grok upstream request failed")
	}
	return err
}
