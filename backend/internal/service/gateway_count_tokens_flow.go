package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GatewayService) prepareCountTokensRequestBody(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *ParsedRequest,
) ([]byte, string, bool) {
	body := StripEmptyTextBlocks(parsed.Body)
	reqModel := parsed.Model

	isClaudeCode := isClaudeCodeRequest(ctx, c, parsed)
	shouldMimicClaudeCode := account.IsOAuth() && !isClaudeCode
	if shouldMimicClaudeCode {
		normalizeOpts := claudeOAuthNormalizeOptions{stripSystemCacheControl: true}
		body, reqModel = normalizeClaudeOAuthRequestBody(body, reqModel, normalizeOpts)
	}

	body, reqModel = s.applyCountTokensModelMapping(account, parsed.Model, reqModel, body)
	return body, reqModel, shouldMimicClaudeCode
}

func (s *GatewayService) applyCountTokensModelMapping(
	account *Account,
	originalModel string,
	requestModel string,
	body []byte,
) ([]byte, string) {
	if requestModel == "" {
		return body, requestModel
	}

	mappedModel := requestModel
	mappingSource := ""
	if account.Type == AccountTypeAPIKey {
		mappedModel = account.GetMappedModel(requestModel)
		if mappedModel != requestModel {
			mappingSource = "account"
		}
	}
	if mappingSource == "" && account.Platform == PlatformAnthropic && account.Type != AccountTypeAPIKey {
		normalized := claude.NormalizeModelID(requestModel)
		if normalized != requestModel {
			mappedModel = normalized
			mappingSource = "prefix"
		}
	}
	if mappedModel == requestModel {
		return body, requestModel
	}

	logger.LegacyPrintf(
		"service.gateway",
		"CountTokens model mapping applied: %s -> %s (account: %s, source=%s)",
		originalModel,
		mappedModel,
		account.Name,
		mappingSource,
	)
	return s.replaceModelInBody(body, mappedModel), mappedModel
}

func (s *GatewayService) resolveCountTokensProxyURL(account *Account) string {
	if account == nil || account.ProxyID == nil || account.Proxy == nil {
		return ""
	}
	if account.IsCustomBaseURLEnabled() && account.GetCustomBaseURL() != "" {
		return ""
	}
	return account.Proxy.URL()
}

func (s *GatewayService) executeCountTokensUpstream(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	reqModel string,
	shouldMimicClaudeCode bool,
) (*http.Response, []byte, error) {
	token, tokenType, err := s.GetAccessToken(ctx, account)
	if err != nil {
		s.countTokensError(c, http.StatusBadGateway, "upstream_error", "Failed to get access token")
		return nil, nil, err
	}

	upstreamReq, err := s.buildCountTokensRequest(ctx, c, account, body, token, tokenType, reqModel, shouldMimicClaudeCode)
	if err != nil {
		s.countTokensError(c, http.StatusInternalServerError, "api_error", "Failed to build request")
		return nil, nil, err
	}

	proxyURL := s.resolveCountTokensProxyURL(account)
	resp, respBody, err := s.doCountTokensUpstreamRequest(c, account, upstreamReq, proxyURL)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode == http.StatusBadRequest && s.shouldRectifySignatureError(ctx, account, respBody) {
		logger.LegacyPrintf("service.gateway", "Account %d: detected thinking block signature error on count_tokens, retrying with filtered thinking blocks", account.ID)

		filteredBody := FilterThinkingBlocksForRetry(body)
		retryReq, buildErr := s.buildCountTokensRequest(ctx, c, account, filteredBody, token, tokenType, reqModel, shouldMimicClaudeCode)
		if buildErr == nil {
			retryResp, retryBody, retryErr := s.doCountTokensUpstreamRequest(c, account, retryReq, proxyURL)
			if retryErr == nil {
				resp = retryResp
				respBody = retryBody
			}
		}
	}

	return resp, respBody, nil
}

func (s *GatewayService) doCountTokensUpstreamRequest(
	c *gin.Context,
	account *Account,
	upstreamReq *http.Request,
	proxyURL string,
) (*http.Response, []byte, error) {
	resp, err := s.httpUpstream.DoWithTLS(upstreamReq, proxyURL, account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		setOpsUpstreamError(c, 0, sanitizeUpstreamErrorMessage(err.Error()), "")
		s.countTokensError(c, http.StatusBadGateway, "upstream_error", "Request failed")
		return nil, nil, fmt.Errorf("upstream request failed: %w", err)
	}

	maxReadBytes := resolveUpstreamResponseReadLimit(s.cfg)
	respBody, err := readUpstreamResponseBodyLimited(resp.Body, maxReadBytes)
	_ = resp.Body.Close()
	if err != nil {
		if errors.Is(err, ErrUpstreamResponseBodyTooLarge) {
			setOpsUpstreamError(c, http.StatusBadGateway, "upstream response too large", "")
			s.countTokensError(c, http.StatusBadGateway, "upstream_error", "Upstream response too large")
			return nil, nil, err
		}
		s.countTokensError(c, http.StatusBadGateway, "upstream_error", "Failed to read response")
		return nil, nil, err
	}

	return resp, respBody, nil
}

func (s *GatewayService) handleCountTokensUpstreamError(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	resp *http.Response,
	respBody []byte,
) error {
	s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, respBody)

	upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
	upstreamDetail := ""
	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
		if maxBytes <= 0 {
			maxBytes = 2048
		}
		upstreamDetail = truncateString(string(respBody), maxBytes)
	}
	setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, upstreamDetail)

	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		logger.LegacyPrintf(
			"service.gateway",
			"count_tokens upstream error %d (account=%d platform=%s type=%s): %s",
			resp.StatusCode,
			account.ID,
			account.Platform,
			account.Type,
			truncateForLog(respBody, s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes),
		)
	}

	errMsg := "Upstream request failed"
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		errMsg = "Rate limit exceeded"
	case 529:
		errMsg = "Service overloaded"
	}
	s.countTokensError(c, resp.StatusCode, "upstream_error", errMsg)
	if upstreamMsg == "" {
		return fmt.Errorf("upstream error: %d", resp.StatusCode)
	}
	return fmt.Errorf("upstream error: %d message=%s", resp.StatusCode, upstreamMsg)
}
