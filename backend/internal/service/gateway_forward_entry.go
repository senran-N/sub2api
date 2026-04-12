package service

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GatewayService) handleForwardSpecialCases(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *ParsedRequest,
	startTime time.Time,
) (bool, *ForwardResult, error) {
	if account != nil && account.IsAnthropicAPIKeyPassthroughEnabled() {
		result, err := s.forwardAnthropicAPIKeyPassthroughWithInput(ctx, c, account, s.resolveForwardPassthroughInput(account, parsed, startTime))
		return true, result, err
	}
	if account != nil && account.IsBedrock() {
		result, err := s.forwardBedrock(ctx, c, account, parsed, startTime)
		return true, result, err
	}
	return false, nil, nil
}

func (s *GatewayService) resolveForwardPassthroughInput(
	account *Account,
	parsed *ParsedRequest,
	startTime time.Time,
) anthropicPassthroughForwardInput {
	body := parsed.Body
	requestModel := parsed.Model
	if requestModel != "" {
		if mappedModel, matched := resolveMappedModelWithOpenAIReasoningFallback(account, requestModel); matched && mappedModel != requestModel {
			body = s.replaceModelInBody(body, mappedModel)
			logger.LegacyPrintf("service.gateway", "Passthrough model mapping: %s -> %s (account: %s)", parsed.Model, mappedModel, account.Name)
			requestModel = mappedModel
		}
	}
	return anthropicPassthroughForwardInput{
		Body:          body,
		RequestModel:  requestModel,
		OriginalModel: parsed.Model,
		RequestStream: parsed.Stream,
		StartTime:     startTime,
	}
}

func (s *GatewayService) initializeForwardBetaPolicy(ctx context.Context, c *gin.Context, account *Account) error {
	if account == nil || account.Platform != PlatformAnthropic || c == nil {
		return nil
	}

	policy := s.evaluateBetaPolicy(ctx, c.GetHeader("anthropic-beta"), account)
	if policy.blockErr != nil {
		return policy.blockErr
	}
	filterSet := policy.filterSet
	if filterSet == nil {
		filterSet = map[string]struct{}{}
	}
	c.Set(betaPolicyFilterSetKey, filterSet)
	return nil
}
