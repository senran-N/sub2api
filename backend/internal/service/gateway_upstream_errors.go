package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
)

// UpstreamFailoverError indicates an upstream error that should trigger account failover.
type UpstreamFailoverError struct {
	StatusCode             int
	ResponseBody           []byte      // 上游响应体，用于错误透传规则匹配
	ResponseHeaders        http.Header // 上游响应头，用于透传 cf-ray/cf-mitigated/content-type 等诊断信息
	ForceCacheBilling      bool        // Antigravity 粘性会话切换时设为 true
	RetryableOnSameAccount bool        // 临时性错误（如 Google 间歇性 400、空响应），应在同一账号上重试 N 次再切换
}

func (e *UpstreamFailoverError) Error() string {
	return fmt.Sprintf("upstream error: %d (failover)", e.StatusCode)
}

type upstreamStreamEventError struct {
	statusCode   int
	responseBody []byte
}

func (e *upstreamStreamEventError) Error() string {
	return "have error in stream"
}

func (e *upstreamStreamEventError) StatusCode() int {
	if e == nil {
		return 0
	}
	return e.statusCode
}

func (e *upstreamStreamEventError) ResponseBody() []byte {
	if e == nil || len(e.responseBody) == 0 {
		return nil
	}
	out := make([]byte, len(e.responseBody))
	copy(out, e.responseBody)
	return out
}

func (s *GatewayService) shouldRectifySignatureError(ctx context.Context, account *Account, respBody []byte) bool {
	if account.Type == AccountTypeAPIKey {
		settings, err := s.settingService.GetRectifierSettings(ctx)
		if err != nil || !settings.Enabled || !settings.APIKeySignatureEnabled {
			return false
		}
		if s.isThinkingBlockSignatureError(respBody) {
			return true
		}
		return matchSignaturePatterns(respBody, settings.APIKeySignaturePatterns)
	}

	return s.isThinkingBlockSignatureError(respBody) && s.settingService.IsSignatureRectifierEnabled(ctx)
}

// isSignatureErrorPattern 仅做模式匹配，不检查开关。
func (s *GatewayService) isSignatureErrorPattern(ctx context.Context, account *Account, respBody []byte) bool {
	if s.isThinkingBlockSignatureError(respBody) {
		return true
	}
	if account.Type == AccountTypeAPIKey {
		settings, err := s.settingService.GetRectifierSettings(ctx)
		if err != nil {
			return false
		}
		return matchSignaturePatterns(respBody, settings.APIKeySignaturePatterns)
	}
	return false
}

func matchSignaturePatterns(respBody []byte, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	bodyLower := strings.ToLower(string(respBody))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if strings.Contains(bodyLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func isClaudeCodeCredentialScopeError(msg string) bool {
	m := strings.ToLower(strings.TrimSpace(msg))
	if m == "" {
		return false
	}
	return strings.Contains(m, "only authorized for use with claude code") &&
		strings.Contains(m, "cannot be used for other api requests")
}

func (s *GatewayService) isThinkingBlockSignatureError(respBody []byte) bool {
	message := strings.ToLower(strings.TrimSpace(extractUpstreamErrorMessage(respBody)))
	if message == "" {
		return false
	}

	if strings.Contains(message, "signature") {
		return true
	}

	if strings.Contains(message, "expected") && (strings.Contains(message, "thinking") || strings.Contains(message, "redacted_thinking")) {
		logger.LegacyPrintf("service.gateway", "[SignatureCheck] Detected thinking block type error")
		return true
	}

	if strings.Contains(message, "cannot be modified") && (strings.Contains(message, "thinking") || strings.Contains(message, "redacted_thinking")) {
		logger.LegacyPrintf("service.gateway", "[SignatureCheck] Detected thinking block modification error")
		return true
	}

	if strings.Contains(message, "non-empty content") || strings.Contains(message, "empty content") ||
		strings.Contains(message, "content blocks must be non-empty") {
		logger.LegacyPrintf("service.gateway", "[SignatureCheck] Detected empty content error")
		return true
	}

	return false
}

func (s *GatewayService) shouldFailoverOn400(respBody []byte) bool {
	message := strings.ToLower(strings.TrimSpace(extractUpstreamErrorMessage(respBody)))
	if message == "" {
		return false
	}

	if strings.Contains(message, "anthropic-beta") ||
		strings.Contains(message, "beta feature") ||
		strings.Contains(message, "requires beta") {
		return true
	}

	if strings.Contains(message, "thinking") || strings.Contains(message, "thought_signature") || strings.Contains(message, "signature") {
		return true
	}
	if strings.Contains(message, "tool_use") || strings.Contains(message, "tool_result") || strings.Contains(message, "tools") {
		return true
	}

	return false
}

func ExtractUpstreamErrorMessage(body []byte) string {
	return extractUpstreamErrorMessage(body)
}

func extractUpstreamErrorMessage(body []byte) string {
	if message := gjson.GetBytes(body, "error.message").String(); strings.TrimSpace(message) != "" {
		inner := strings.TrimSpace(message)
		if strings.HasPrefix(inner, "{") {
			if nested := gjson.Get(inner, "error.message").String(); strings.TrimSpace(nested) != "" {
				return nested
			}
		}
		return message
	}

	if detail := gjson.GetBytes(body, "detail").String(); strings.TrimSpace(detail) != "" {
		return detail
	}

	return gjson.GetBytes(body, "message").String()
}

func extractUpstreamErrorCode(body []byte) string {
	if code := strings.TrimSpace(gjson.GetBytes(body, "error.code").String()); code != "" {
		return code
	}

	inner := strings.TrimSpace(gjson.GetBytes(body, "error.message").String())
	if !strings.HasPrefix(inner, "{") {
		return ""
	}

	if code := strings.TrimSpace(gjson.Get(inner, "error.code").String()); code != "" {
		return code
	}
	if lastBrace := strings.LastIndex(inner, "}"); lastBrace >= 0 {
		if code := strings.TrimSpace(gjson.Get(inner[:lastBrace+1], "error.code").String()); code != "" {
			return code
		}
	}
	return ""
}

func isCountTokensUnsupported404(statusCode int, body []byte) bool {
	if statusCode != http.StatusNotFound {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(extractUpstreamErrorMessage(body)))
	if message == "" {
		return false
	}
	if strings.Contains(message, "/v1/messages/count_tokens") {
		return true
	}
	return strings.Contains(message, "count_tokens") && strings.Contains(message, "not found")
}

func (s *GatewayService) handleErrorResponse(ctx context.Context, resp *http.Response, c *gin.Context, account *Account) (*ForwardResult, error) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))

	logger.LegacyPrintf("service.gateway", "[Forward] Upstream error (non-retryable): Account=%d(%s) Status=%d RequestID=%s Body=%s",
		account.ID, account.Name, resp.StatusCode, resp.Header.Get("x-request-id"), truncateString(string(body), 1000))

	upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(body))
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)

	if isClaudeCodeCredentialScopeError(upstreamMsg) && c != nil {
		if value, ok := c.Get(claudeMimicDebugInfoKey); ok {
			if line, ok := value.(string); ok && strings.TrimSpace(line) != "" {
				logger.LegacyPrintf("service.gateway", "[ClaudeMimicDebugOnError] status=%d request_id=%s %s",
					resp.StatusCode,
					resp.Header.Get("x-request-id"),
					line,
				)
			}
		}
	}

	upstreamDetail := ""
	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
		if maxBytes <= 0 {
			maxBytes = 2048
		}
		upstreamDetail = truncateString(string(body), maxBytes)
	}
	setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, upstreamDetail)
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:           account.Platform,
		AccountID:          account.ID,
		UpstreamStatusCode: resp.StatusCode,
		UpstreamRequestID:  resp.Header.Get("x-request-id"),
		Kind:               "http_error",
		Message:            upstreamMsg,
		Detail:             upstreamDetail,
	})

	shouldDisable := false
	if s.rateLimitService != nil {
		shouldDisable = s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, body)
	}
	if shouldDisable {
		return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: body}
	}

	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		logger.LegacyPrintf("service.gateway",
			"Upstream error %d (account=%d platform=%s type=%s): %s",
			resp.StatusCode,
			account.ID,
			account.Platform,
			account.Type,
			truncateForLog(body, s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes),
		)
	}

	if status, errType, errMsg, matched := applyErrorPassthroughRule(
		c,
		account.Platform,
		resp.StatusCode,
		body,
		http.StatusBadGateway,
		"upstream_error",
		"Upstream request failed",
	); matched {
		c.JSON(status, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    errType,
				"message": errMsg,
			},
		})

		summary := upstreamMsg
		if summary == "" {
			summary = errMsg
		}
		if summary == "" {
			return nil, fmt.Errorf("upstream error: %d (passthrough rule matched)", resp.StatusCode)
		}
		return nil, fmt.Errorf("upstream error: %d (passthrough rule matched) message=%s", resp.StatusCode, summary)
	}

	var (
		errType    string
		errMsg     string
		statusCode int
	)
	switch resp.StatusCode {
	case 400:
		c.Data(http.StatusBadRequest, "application/json", body)
		summary := upstreamMsg
		if summary == "" {
			summary = truncateForLog(body, 512)
		}
		if summary == "" {
			return nil, fmt.Errorf("upstream error: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("upstream error: %d message=%s", resp.StatusCode, summary)
	case 401:
		statusCode = http.StatusBadGateway
		errType = "upstream_error"
		errMsg = "Upstream authentication failed, please contact administrator"
	case 403:
		statusCode = http.StatusBadGateway
		errType = "upstream_error"
		errMsg = "Upstream access forbidden, please contact administrator"
	case 429:
		statusCode = http.StatusTooManyRequests
		errType = "rate_limit_error"
		errMsg = "Upstream rate limit exceeded, please retry later"
	case 529:
		statusCode = http.StatusServiceUnavailable
		errType = "overloaded_error"
		errMsg = "Upstream service overloaded, please retry later"
	case 500, 502, 503, 504:
		statusCode = http.StatusBadGateway
		errType = "upstream_error"
		errMsg = "Upstream service temporarily unavailable"
	default:
		statusCode = http.StatusBadGateway
		errType = "upstream_error"
		errMsg = "Upstream request failed"
	}

	c.JSON(statusCode, gin.H{
		"type": "error",
		"error": gin.H{
			"type":    errType,
			"message": errMsg,
		},
	})

	if upstreamMsg == "" {
		return nil, fmt.Errorf("upstream error: %d", resp.StatusCode)
	}
	return nil, fmt.Errorf("upstream error: %d message=%s", resp.StatusCode, upstreamMsg)
}

func (s *GatewayService) handleRetryExhaustedSideEffects(ctx context.Context, resp *http.Response, account *Account) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	statusCode := resp.StatusCode

	if account.IsOAuth() && statusCode == http.StatusForbidden {
		s.rateLimitService.HandleUpstreamError(ctx, account, statusCode, resp.Header, body)
		logger.LegacyPrintf("service.gateway", "Account %d: marked as error after %d retries for status %d", account.ID, maxRetryAttempts, statusCode)
		return
	}

	logger.LegacyPrintf("service.gateway", "Account %d: upstream error %d after %d retries (not marking account)", account.ID, statusCode, maxRetryAttempts)
}

func (s *GatewayService) handleFailoverSideEffects(ctx context.Context, resp *http.Response, account *Account) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, body)
}

func (s *GatewayService) handleRetryExhaustedError(ctx context.Context, resp *http.Response, c *gin.Context, account *Account) (*ForwardResult, error) {
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	s.handleRetryExhaustedSideEffects(ctx, resp, account)

	upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)

	if isClaudeCodeCredentialScopeError(upstreamMsg) && c != nil {
		if value, ok := c.Get(claudeMimicDebugInfoKey); ok {
			if line, ok := value.(string); ok && strings.TrimSpace(line) != "" {
				logger.LegacyPrintf("service.gateway", "[ClaudeMimicDebugOnError] status=%d request_id=%s %s",
					resp.StatusCode,
					resp.Header.Get("x-request-id"),
					line,
				)
			}
		}
	}

	upstreamDetail := ""
	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
		if maxBytes <= 0 {
			maxBytes = 2048
		}
		upstreamDetail = truncateString(string(respBody), maxBytes)
	}
	setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, upstreamDetail)
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:           account.Platform,
		AccountID:          account.ID,
		UpstreamStatusCode: resp.StatusCode,
		UpstreamRequestID:  resp.Header.Get("x-request-id"),
		Kind:               "retry_exhausted",
		Message:            upstreamMsg,
		Detail:             upstreamDetail,
	})

	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		logger.LegacyPrintf("service.gateway",
			"Upstream error %d retries_exhausted (account=%d platform=%s type=%s): %s",
			resp.StatusCode,
			account.ID,
			account.Platform,
			account.Type,
			truncateForLog(respBody, s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes),
		)
	}

	if status, errType, errMsg, matched := applyErrorPassthroughRule(
		c,
		account.Platform,
		resp.StatusCode,
		respBody,
		http.StatusBadGateway,
		"upstream_error",
		"Upstream request failed after retries",
	); matched {
		c.JSON(status, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    errType,
				"message": errMsg,
			},
		})

		summary := upstreamMsg
		if summary == "" {
			summary = errMsg
		}
		if summary == "" {
			return nil, fmt.Errorf("upstream error: %d (retries exhausted, passthrough rule matched)", resp.StatusCode)
		}
		return nil, fmt.Errorf("upstream error: %d (retries exhausted, passthrough rule matched) message=%s", resp.StatusCode, summary)
	}

	c.JSON(http.StatusBadGateway, gin.H{
		"type": "error",
		"error": gin.H{
			"type":    "upstream_error",
			"message": "Upstream request failed after retries",
		},
	})

	if upstreamMsg == "" {
		return nil, fmt.Errorf("upstream error: %d (retries exhausted)", resp.StatusCode)
	}
	return nil, fmt.Errorf("upstream error: %d (retries exhausted) message=%s", resp.StatusCode, upstreamMsg)
}
