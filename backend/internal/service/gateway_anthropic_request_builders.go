package service

import (
	"bytes"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *GatewayService) buildUpstreamRequest(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	token, tokenType, modelID string,
	reqStream bool,
	mimicClaudeCode bool,
) (*http.Request, error) {
	targetURL, err := s.resolveAnthropicUpstreamURL(account, claudeAPIURL, "/v1/messages")
	if err != nil {
		return nil, err
	}

	state := s.prepareAnthropicRequestBuildState(ctx, account, body, extractClientHeaders(c))
	body = state.body

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	policyFilterSet := s.getBetaPolicyFilterSet(ctx, c, account)
	applyAnthropicAuthHeader(req.Header, token, tokenType)
	applyAllowedClientHeaders(req.Header, state.headers)
	if state.fingerprint != nil {
		s.identityService.ApplyFingerprint(req, state.fingerprint)
	}
	ensureAnthropicBaseHeaders(req, tokenType)
	if mimicClaudeCode && tokenType == "oauth" && reqStream {
		setHeaderRaw(req.Header, "x-stainless-helper-method", "stream")
	}
	s.configureMessagesBetaHeader(req, body, account, tokenType, modelID, mimicClaudeCode, policyFilterSet)
	syncClaudeCodeSessionHeader(req, body, tokenType == "oauth")
	s.finalizeAnthropicRequestDebug(
		c,
		req,
		body,
		account,
		tokenType,
		mimicClaudeCode,
		state.enableFP,
		state.enableMPT,
		state.fingerprint != nil,
	)

	return req, nil
}

func (s *GatewayService) buildCountTokensRequest(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	token, tokenType, modelID string,
	mimicClaudeCode bool,
) (*http.Request, error) {
	targetURL, err := s.resolveAnthropicUpstreamURL(account, claudeAPICountTokensURL, "/v1/messages/count_tokens")
	if err != nil {
		return nil, err
	}

	state := s.prepareAnthropicRequestBuildState(ctx, account, body, extractClientHeaders(c))
	body = state.body

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	applyAnthropicAuthHeader(req.Header, token, tokenType)
	applyAllowedClientHeaders(req.Header, state.headers)
	if state.enableFP && state.fingerprint != nil {
		s.identityService.ApplyFingerprint(req, state.fingerprint)
	}
	ensureAnthropicBaseHeaders(req, tokenType)
	s.configureCountTokensBetaHeader(req, body, account, tokenType, modelID, mimicClaudeCode, s.getBetaPolicyFilterSet(ctx, c, account))
	syncClaudeCodeSessionHeader(req, body, tokenType == "oauth")
	s.finalizeAnthropicRequestDebug(
		c,
		req,
		body,
		account,
		tokenType,
		mimicClaudeCode,
		state.enableFP,
		state.enableMPT,
		state.enableFP && state.fingerprint != nil,
	)

	return req, nil
}
