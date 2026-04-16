package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"
)

func (s *OpenAIGatewayService) ForwardCompatiblePassthrough(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	requestMeta openAICompatiblePassthroughRequestMeta,
	defaultMappedModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()

	reqModel := requestMeta.Model
	forwardBody := body
	mappedModel := resolveOpenAIForwardModel(account, reqModel, defaultMappedModel)
	mappedModel = normalizeOpenAIModelForUpstream(account, mappedModel)

	if requestMeta.JSONBody && reqModel != "" && mappedModel != "" && mappedModel != reqModel {
		patchedBody, err := sjson.SetBytes(body, "model", mappedModel)
		if err != nil {
			return nil, fmt.Errorf("patch passthrough model: %w", err)
		}
		forwardBody = patchedBody
	}

	var reasoningEffort *string
	if value := normalizeOpenAIReasoningEffort(requestMeta.ReasoningEffort); value != "" {
		reasoningEffort = &value
	}

	result, err := s.forwardOpenAIPassthrough(
		ctx,
		c,
		account,
		forwardBody,
		reqModel,
		reasoningEffort,
		requestMeta.Stream,
		startTime,
	)
	if err != nil {
		return nil, err
	}
	if result != nil && mappedModel != "" && mappedModel != reqModel {
		result.UpstreamModel = mappedModel
	}
	return result, nil
}
