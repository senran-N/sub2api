package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type anthropicRequestBuildState struct {
	body        []byte
	headers     http.Header
	fingerprint *Fingerprint
	enableFP    bool
	enableMPT   bool
}

func (s *GatewayService) resolveAnthropicUpstreamURL(account *Account, defaultURL, path string) (string, error) {
	if account.Type == AccountTypeAPIKey {
		baseURL := account.GetBaseURL()
		if baseURL == "" {
			return defaultURL, nil
		}
		validatedURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return "", err
		}
		return validatedURL + path + "?beta=true", nil
	}

	if account.IsCustomBaseURLEnabled() {
		customURL := account.GetCustomBaseURL()
		if customURL == "" {
			return "", fmt.Errorf("custom_base_url is enabled but not configured for account %d", account.ID)
		}
		validatedURL, err := s.validateUpstreamBaseURL(customURL)
		if err != nil {
			return "", err
		}
		return s.buildCustomRelayURL(validatedURL, path, account), nil
	}

	return defaultURL, nil
}

func extractClientHeaders(c *gin.Context) http.Header {
	if c != nil && c.Request != nil {
		return c.Request.Header
	}
	return http.Header{}
}

func (s *GatewayService) prepareAnthropicRequestBuildState(
	ctx context.Context,
	account *Account,
	body []byte,
	clientHeaders http.Header,
) anthropicRequestBuildState {
	state := anthropicRequestBuildState{
		body:    body,
		headers: clientHeaders,
	}

	if s.settingService != nil {
		state.enableFP, state.enableMPT = s.settingService.GetGatewayForwardingSettings(ctx)
	} else {
		state.enableFP = true
	}

	if !account.IsOAuth() || s.identityService == nil {
		return state
	}

	fingerprint, err := s.identityService.GetOrCreateFingerprint(ctx, account.ID, clientHeaders)
	if err != nil {
		logger.LegacyPrintf("service.gateway", "Warning: failed to get fingerprint for account %d: %v", account.ID, err)
		return state
	}

	if state.enableFP {
		state.fingerprint = fingerprint
	}
	if state.enableMPT {
		return state
	}

	accountUUID := account.GetExtraString("account_uuid")
	if accountUUID == "" || fingerprint.ClientID == "" {
		return state
	}

	if newBody, err := s.identityService.RewriteUserIDWithMasking(
		ctx,
		body,
		account,
		accountUUID,
		fingerprint.ClientID,
		fingerprint.UserAgent,
	); err == nil && len(newBody) > 0 {
		state.body = newBody
	}

	return state
}

func applyAllowedClientHeaders(dst http.Header, clientHeaders http.Header) {
	for key, values := range clientHeaders {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		if !allowedHeaders[lowerKey] {
			continue
		}
		wireKey := resolveWireCasing(key)
		for _, value := range values {
			addHeaderRaw(dst, wireKey, value)
		}
	}
}

func applyAnthropicAuthHeader(dst http.Header, token, tokenType string) {
	if tokenType == "oauth" {
		setHeaderRaw(dst, "authorization", "Bearer "+token)
		return
	}
	setHeaderRaw(dst, "x-api-key", token)
}

func ensureAnthropicBaseHeaders(req *http.Request, tokenType string) {
	if getHeaderRaw(req.Header, "content-type") == "" {
		setHeaderRaw(req.Header, "content-type", "application/json")
	}
	if getHeaderRaw(req.Header, "anthropic-version") == "" {
		setHeaderRaw(req.Header, "anthropic-version", "2023-06-01")
	}
	if tokenType == "oauth" {
		applyClaudeOAuthHeaderDefaults(req)
	}
}

func (s *GatewayService) configureMessagesBetaHeader(
	req *http.Request,
	body []byte,
	account *Account,
	tokenType string,
	modelID string,
	mimicClaudeCode bool,
	policyFilterSet map[string]struct{},
) {
	effectiveDropSet := mergeDropSets(policyFilterSet)
	effectiveDropWithClaudeCodeSet := mergeDropSets(policyFilterSet, claude.ClaudeCodeBetaToken())

	if tokenType == "oauth" {
		if mimicClaudeCode {
			applyClaudeCodeMimicHeaders(req, getHeaderRaw(req.Header, "x-stainless-helper-method") == "stream")
			incomingBeta := getHeaderRaw(req.Header, "anthropic-beta")
			requiredBetas := []string{claude.OAuthBetaToken(), claude.InterleavedThinkingBetaToken()}
			setHeaderRaw(req.Header, "anthropic-beta", mergeAnthropicBetaDropping(requiredBetas, incomingBeta, effectiveDropWithClaudeCodeSet))
			return
		}

		clientBetaHeader := getHeaderRaw(req.Header, "anthropic-beta")
		setHeaderRaw(req.Header, "anthropic-beta", stripBetaTokensWithSet(s.getBetaHeader(modelID, clientBetaHeader), effectiveDropSet))
		return
	}

	if existingBeta := getHeaderRaw(req.Header, "anthropic-beta"); existingBeta != "" {
		setHeaderRaw(req.Header, "anthropic-beta", stripBetaTokensWithSet(existingBeta, effectiveDropSet))
		return
	}

	if s.cfg != nil && s.cfg.Gateway.InjectBetaForAPIKey && requestNeedsBetaFeatures(body) {
		if beta := defaultAPIKeyBetaHeader(body); beta != "" {
			setHeaderRaw(req.Header, "anthropic-beta", beta)
		}
	}
}

func (s *GatewayService) configureCountTokensBetaHeader(
	req *http.Request,
	body []byte,
	account *Account,
	tokenType string,
	modelID string,
	mimicClaudeCode bool,
	policyFilterSet map[string]struct{},
) {
	effectiveDropSet := mergeDropSets(policyFilterSet)

	if tokenType == "oauth" {
		if mimicClaudeCode {
			applyClaudeCodeMimicHeaders(req, false)
			incomingBeta := getHeaderRaw(req.Header, "anthropic-beta")
			requiredBetas := []string{
				claude.ClaudeCodeBetaToken(),
				claude.OAuthBetaToken(),
				claude.InterleavedThinkingBetaToken(),
				claude.TokenCountingBetaToken(),
			}
			setHeaderRaw(req.Header, "anthropic-beta", mergeAnthropicBetaDropping(requiredBetas, incomingBeta, effectiveDropSet))
			return
		}

		clientBetaHeader := getHeaderRaw(req.Header, "anthropic-beta")
		if clientBetaHeader == "" {
			setHeaderRaw(req.Header, "anthropic-beta", claude.CountTokensAnthropicBetaHeader())
			return
		}

		beta := s.getBetaHeader(modelID, clientBetaHeader)
		if !strings.Contains(beta, claude.TokenCountingBetaToken()) {
			beta += "," + claude.TokenCountingBetaToken()
		}
		setHeaderRaw(req.Header, "anthropic-beta", stripBetaTokensWithSet(beta, effectiveDropSet))
		return
	}

	if existingBeta := getHeaderRaw(req.Header, "anthropic-beta"); existingBeta != "" {
		setHeaderRaw(req.Header, "anthropic-beta", stripBetaTokensWithSet(existingBeta, effectiveDropSet))
		return
	}

	if s.cfg != nil && s.cfg.Gateway.InjectBetaForAPIKey && requestNeedsBetaFeatures(body) {
		if beta := defaultAPIKeyBetaHeader(body); beta != "" {
			setHeaderRaw(req.Header, "anthropic-beta", beta)
		}
	}
}

func (s *GatewayService) finalizeAnthropicRequestDebug(
	c *gin.Context,
	req *http.Request,
	body []byte,
	account *Account,
	tokenType string,
	mimicClaudeCode bool,
	enableFP bool,
	enableMPT bool,
	fingerprintApplied bool,
) {
	s.debugLogGatewaySnapshot("UPSTREAM_FORWARD", req.Header, body, map[string]string{
		"url":                 req.URL.String(),
		"token_type":          tokenType,
		"mimic_claude_code":   strconv.FormatBool(mimicClaudeCode),
		"fingerprint_applied": strconv.FormatBool(fingerprintApplied),
		"enable_fp":           strconv.FormatBool(enableFP),
		"enable_mpt":          strconv.FormatBool(enableMPT),
	})

	if c != nil && tokenType == "oauth" {
		c.Set(claudeMimicDebugInfoKey, buildClaudeMimicDebugLine(req, body, account, tokenType, mimicClaudeCode))
	}
	if s.debugClaudeMimicEnabled() {
		logClaudeMimicDebug(req, body, account, tokenType, mimicClaudeCode)
	}
}
