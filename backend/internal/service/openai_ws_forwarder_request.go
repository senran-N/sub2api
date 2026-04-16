package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func (s *OpenAIGatewayService) buildOpenAIResponsesWSURL(account *Account) (string, error) {
	if account == nil {
		return "", errors.New("account is nil")
	}
	var targetURL string
	upstreamTarget := newOpenAIResponsesUpstreamTarget(openaiPlatformAPIURL)
	switch account.Type {
	case AccountTypeOAuth:
		targetURL = chatgptCodexURL
	case AccountTypeAPIKey, AccountTypeUpstream:
		baseURL := account.GetOpenAIBaseURL()
		if baseURL == "" {
			targetURL = openaiPlatformAPIURL
		} else {
			validatedURL, err := s.validateUpstreamBaseURL(baseURL)
			if err != nil {
				return "", err
			}
			upstreamTarget = newOpenAIResponsesUpstreamTargetWithOptions(
				validatedURL,
				account.GetCompatibleAuthMode(""),
				account.GetCompatibleEndpointOverride("responses"),
			)
			targetURL = upstreamTarget.URL
		}
	default:
		targetURL = openaiPlatformAPIURL
	}

	parsed, err := url.Parse(strings.TrimSpace(targetURL))
	if err != nil {
		return "", fmt.Errorf("invalid target url: %w", err)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "https":
		parsed.Scheme = "wss"
	case "http":
		parsed.Scheme = "ws"
	case "wss", "ws":
	default:
		return "", fmt.Errorf("unsupported scheme for ws: %s", parsed.Scheme)
	}
	return parsed.String(), nil
}

func (s *OpenAIGatewayService) buildOpenAIWSHeaders(
	c *gin.Context,
	body []byte,
	account *Account,
	token string,
	decision OpenAIWSProtocolDecision,
	turnState string,
	turnMetadata string,
	promptCacheKey string,
) (http.Header, openAIWSSessionHeaderResolution) {
	forceCodexCLI := s != nil && s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
	profile := GetCodexRequestProfile(c, body, forceCodexCLI)
	policy := NewCodexNativeMutationPolicy(profile)
	resolvedPromptCacheKey := strings.TrimSpace(promptCacheKey)
	if resolvedPromptCacheKey == "" {
		resolvedPromptCacheKey = strings.TrimSpace(profile.Body.PromptCacheKey)
	}

	headers := make(http.Header)
	upstreamTarget := newOpenAIResponsesUpstreamTarget(openaiPlatformAPIURL)
	if account != nil && (account.Type == AccountTypeAPIKey || account.Type == AccountTypeUpstream) {
		upstreamTarget = newOpenAIResponsesUpstreamTargetWithOptions(
			account.GetOpenAIBaseURL(),
			account.GetCompatibleAuthMode(""),
			account.GetCompatibleEndpointOverride("responses"),
		)
	}
	upstreamTarget.ApplyAuthHeader(headers, token)

	sessionResolution := policy.ResolveSessionHeaders(resolvedPromptCacheKey)
	if account != nil && account.Type == AccountTypeOAuth {
		sessionResolution = policy.ResolveOAuthSessionHeaders(resolvedPromptCacheKey, "", false)
	}
	if acceptLanguage := policy.ResolveAcceptLanguage(); acceptLanguage != "" && (account == nil || account.Type != AccountTypeOAuth) {
		headers.Set("accept-language", acceptLanguage)
	}
	if account != nil && account.Type == AccountTypeOAuth {
		isolatedSessionID := ""
		if sessionResolution.SessionID != "" {
			isolatedSessionID = resolveOpenAICodexUpstreamSessionHeaderValue(account.ID, sessionResolution.SessionID, sessionResolution.SessionSource)
			headers.Set("session_id", isolatedSessionID)
		}
		if sessionResolution.ConversationID != "" {
			headers.Set("conversation_id", resolveOpenAICodexUpstreamSessionHeaderValue(account.ID, sessionResolution.ConversationID, sessionResolution.ConversationSource))
		}
		if requestID := resolveOpenAICodexUpstreamClientRequestID(account.ID, profile.Headers.ClientRequestID, isolatedSessionID); requestID != "" {
			headers.Set("x-client-request-id", requestID)
		}
	} else {
		if sessionResolution.SessionID != "" {
			headers.Set("session_id", sessionResolution.SessionID)
		}
		if sessionResolution.ConversationID != "" {
			headers.Set("conversation_id", sessionResolution.ConversationID)
		}
		if requestID := policy.ResolveClientRequestID(sessionResolution.SessionID); requestID != "" {
			headers.Set("x-client-request-id", requestID)
		}
	}
	if state := strings.TrimSpace(turnState); state != "" {
		headers.Set(openAIWSTurnStateHeader, state)
	}
	metadataValue := strings.TrimSpace(turnMetadata)
	if account != nil && account.Type == AccountTypeOAuth {
		metadataValue = resolveOpenAICodexUpstreamTurnMetadata(account.ID, metadataValue)
	}
	if metadata := strings.TrimSpace(metadataValue); metadata != "" {
		headers.Set(openAIWSTurnMetadataHeader, metadata)
	}

	if account != nil && account.Type == AccountTypeOAuth {
		if chatgptAccountID := account.GetChatGPTAccountID(); chatgptAccountID != "" {
			headers.Set("chatgpt-account-id", chatgptAccountID)
		}
		headers.Set("originator", resolveOpenAICodexUpstreamOriginator(account))
		if acceptLanguage := resolveOpenAICodexUpstreamAcceptLanguage(account); acceptLanguage != "" {
			headers.Set("accept-language", acceptLanguage)
		} else {
			headers.Del("accept-language")
		}
		if windowID := resolveOpenAICodexUpstreamWindowID(account.ID); windowID != "" {
			headers.Set(openAICodexMetadataWindowIDKey, windowID)
		}
		if subagent := resolveOpenAICodexUpstreamSubagent(profile, body); subagent != "" {
			headers.Set(openAICodexMetadataSubagentKey, subagent)
		} else {
			headers.Del(openAICodexMetadataSubagentKey)
		}
		if parentThreadID := resolveOpenAICodexUpstreamParentThreadID(account.ID, profile, body); parentThreadID != "" {
			headers.Set(openAICodexMetadataParentThreadIDKey, parentThreadID)
		} else {
			headers.Del(openAICodexMetadataParentThreadIDKey)
		}
	}

	betaValue := openAIWSBetaV2Value
	if decision.Transport == OpenAIUpstreamTransportResponsesWebsocket {
		betaValue = openAIWSBetaV1Value
	}
	if account != nil && account.Type == AccountTypeOAuth {
		headers.Set("OpenAI-Beta", betaValue)
	} else {
		headers.Set("OpenAI-Beta", policy.ResolveOpenAIBeta(betaValue))
	}
	if betaFeatures := resolveOpenAICodexUpstreamBetaFeatures(account); betaFeatures != "" {
		headers.Set("x-codex-beta-features", betaFeatures)
	} else {
		headers.Del("x-codex-beta-features")
	}

	userAgent := ""
	if account != nil && account.Type == AccountTypeOAuth {
		userAgent = resolveOpenAICodexUpstreamUserAgent(account)
	} else {
		userAgent = policy.ResolveUserAgent(account, forceCodexCLI, true)
	}
	if userAgent != "" {
		headers.Set("user-agent", userAgent)
	}

	return headers, sessionResolution
}

func (s *OpenAIGatewayService) buildOpenAIWSCreatePayload(reqBody map[string]any, account *Account) map[string]any {
	// OpenAI WS Mode 协议：response.create 字段与 HTTP /responses 基本一致。
	// 保留 stream 字段（与 Codex CLI 一致），仅移除 background。
	payload := make(map[string]any, len(reqBody)+1)
	for k, v := range reqBody {
		payload[k] = v
	}

	delete(payload, "background")
	if _, exists := payload["stream"]; !exists {
		payload["stream"] = true
	}
	payload["type"] = "response.create"

	// OAuth 默认保持 store=false，避免误依赖服务端历史。
	if account != nil && account.Type == AccountTypeOAuth && !s.isOpenAIWSStoreRecoveryAllowed(account) {
		payload["store"] = false
	}
	return payload
}

func setOpenAIWSTurnMetadata(payload map[string]any, turnMetadata string) {
	if len(payload) == 0 {
		return
	}
	metadata := strings.TrimSpace(turnMetadata)
	if metadata == "" {
		return
	}

	switch existing := payload["client_metadata"].(type) {
	case map[string]any:
		existing[openAIWSTurnMetadataHeader] = metadata
		payload["client_metadata"] = existing
	case map[string]string:
		next := make(map[string]any, len(existing)+1)
		for k, v := range existing {
			next[k] = v
		}
		next[openAIWSTurnMetadataHeader] = metadata
		payload["client_metadata"] = next
	default:
		payload["client_metadata"] = map[string]any{
			openAIWSTurnMetadataHeader: metadata,
		}
	}
}

func (s *OpenAIGatewayService) isOpenAIWSStoreRecoveryAllowed(account *Account) bool {
	if account != nil && account.IsOpenAIWSAllowStoreRecoveryEnabled() {
		return true
	}
	if s != nil && s.cfg != nil && s.cfg.Gateway.OpenAIWS.AllowStoreRecovery {
		return true
	}
	return false
}

func (s *OpenAIGatewayService) isOpenAIWSStoreDisabledInRequest(reqBody map[string]any, account *Account) bool {
	if account != nil && account.Type == AccountTypeOAuth && !s.isOpenAIWSStoreRecoveryAllowed(account) {
		return true
	}
	if len(reqBody) == 0 {
		return false
	}
	rawStore, ok := reqBody["store"]
	if !ok {
		return false
	}
	storeEnabled, ok := rawStore.(bool)
	if !ok {
		return false
	}
	return !storeEnabled
}

func (s *OpenAIGatewayService) isOpenAIWSStoreDisabledInRequestRaw(reqBody []byte, account *Account) bool {
	if account != nil && account.Type == AccountTypeOAuth && !s.isOpenAIWSStoreRecoveryAllowed(account) {
		return true
	}
	if len(reqBody) == 0 {
		return false
	}
	storeValue := gjson.GetBytes(reqBody, "store")
	if !storeValue.Exists() {
		return false
	}
	if storeValue.Type != gjson.True && storeValue.Type != gjson.False {
		return false
	}
	return !storeValue.Bool()
}

func (s *OpenAIGatewayService) openAIWSStoreDisabledConnMode() string {
	if s == nil || s.cfg == nil {
		return openAIWSStoreDisabledConnModeStrict
	}
	mode := strings.ToLower(strings.TrimSpace(s.cfg.Gateway.OpenAIWS.StoreDisabledConnMode))
	switch mode {
	case openAIWSStoreDisabledConnModeStrict, openAIWSStoreDisabledConnModeAdaptive, openAIWSStoreDisabledConnModeOff:
		return mode
	case "":
		// 兼容旧配置：仅配置了布尔开关时按旧语义推导。
		if s.cfg.Gateway.OpenAIWS.StoreDisabledForceNewConn {
			return openAIWSStoreDisabledConnModeStrict
		}
		return openAIWSStoreDisabledConnModeOff
	default:
		return openAIWSStoreDisabledConnModeStrict
	}
}

func shouldForceNewConnOnStoreDisabled(mode, lastFailureReason string) bool {
	switch mode {
	case openAIWSStoreDisabledConnModeOff:
		return false
	case openAIWSStoreDisabledConnModeAdaptive:
		reason := strings.TrimPrefix(strings.TrimSpace(lastFailureReason), "prewarm_")
		switch reason {
		case "policy_violation", "message_too_big", "auth_failed", "write_request", "write":
			return true
		default:
			return false
		}
	default:
		return true
	}
}
