package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ForwardCountTokens 转发 count_tokens 请求到上游 API
// 特点：不记录使用量、仅支持非流式响应
func (s *GatewayService) ForwardCountTokens(ctx context.Context, c *gin.Context, account *Account, parsed *ParsedRequest) error {
	if parsed == nil {
		s.countTokensError(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return fmt.Errorf("parse request: empty request")
	}

	if handled, err := s.handleCountTokensSpecialCases(ctx, c, account, parsed); handled {
		return err
	}

	body, reqModel, shouldMimicClaudeCode := s.prepareCountTokensRequestBody(ctx, c, account, parsed)

	resp, respBody, err := s.executeCountTokensUpstream(ctx, c, account, body, reqModel, shouldMimicClaudeCode)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return s.handleCountTokensUpstreamError(ctx, c, account, resp, respBody)
	}

	c.Data(resp.StatusCode, "application/json", respBody)
	return nil
}
