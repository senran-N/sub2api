package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GatewayService) handleCountTokensSpecialCases(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *ParsedRequest,
) (bool, error) {
	if account != nil && account.IsAnthropicAPIKeyPassthroughEnabled() {
		return true, s.forwardCountTokensAnthropicAPIKeyPassthrough(ctx, c, account, s.resolveCountTokensPassthroughBody(account, parsed))
	}
	if account != nil && account.IsBedrock() {
		s.countTokensError(c, http.StatusNotFound, "not_found_error", "count_tokens endpoint is not supported for Bedrock")
		return true, nil
	}
	if account != nil && account.Platform == PlatformAntigravity {
		s.countTokensError(c, http.StatusNotFound, "not_found_error", "count_tokens endpoint is not supported for this platform")
		return true, nil
	}
	return false, nil
}

func (s *GatewayService) resolveCountTokensPassthroughBody(account *Account, parsed *ParsedRequest) []byte {
	if account == nil || parsed == nil {
		return nil
	}

	body := parsed.Body
	if reqModel := parsed.Model; reqModel != "" {
		if mappedModel, matched := resolveMappedModelWithOpenAIReasoningFallback(account, reqModel); matched && mappedModel != reqModel {
			body = s.replaceModelInBody(body, mappedModel)
			logger.LegacyPrintf("service.gateway", "CountTokens passthrough model mapping: %s -> %s (account: %s)", reqModel, mappedModel, account.Name)
		}
	}
	return body
}
