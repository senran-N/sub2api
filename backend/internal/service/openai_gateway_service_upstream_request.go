package service

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *OpenAIGatewayService) buildUpstreamRequest(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	token string,
	isStream bool,
	promptCacheKey string,
) (*http.Request, error) {
	_ = isStream
	forceCodexCLI := s != nil && s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
	profile := GetCodexRequestProfile(c, body, forceCodexCLI)
	policy := NewCodexNativeMutationPolicy(profile)

	var targetURL string
	upstreamTarget := newCompatibleResponsesUpstreamTarget(openaiPlatformAPIURL)
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
				return nil, err
			}
			upstreamTarget = newCompatibleResponsesUpstreamTargetWithOptions(
				validatedURL,
				account.GetCompatibleAuthMode(""),
				account.GetCompatibleEndpointOverride("responses"),
			)
			targetURL = upstreamTarget.URL
		}
	default:
		targetURL = openaiPlatformAPIURL
	}
	targetURL = appendOpenAIResponsesRequestPathSuffix(targetURL, openAIResponsesRequestPathSuffix(c))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	upstreamTarget.ApplyAuthHeader(req.Header, token)

	if account.Type == AccountTypeOAuth {
		req.Host = "chatgpt.com"
		if chatgptAccountID := account.GetChatGPTAccountID(); chatgptAccountID != "" {
			req.Header.Set("chatgpt-account-id", chatgptAccountID)
		}
	}

	for key, values := range c.Request.Header {
		lowerKey := strings.ToLower(key)
		if !openaiAllowedHeaders[lowerKey] {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if account.Type == AccountTypeOAuth {
		req.Header.Del("conversation_id")
		req.Header.Del("session_id")
		req.Header.Del("accept-language")

		sessionResolution := policy.ResolveOAuthSessionHeaders(promptCacheKey, resolveOpenAICompactSessionID(c), !profile.NativeClient)
		req.Header.Set("OpenAI-Beta", openAICodexHTTPBetaValue)
		req.Header.Set("originator", resolveOpenAICodexUpstreamOriginator(account))
		if profile.CompactPath {
			req.Header.Set("accept", policy.ResolveAccept("application/json"))
			req.Header.Set("version", resolveOpenAICodexUpstreamVersion(account))
		} else {
			req.Header.Set("accept", policy.ResolveAccept("text/event-stream"))
		}
		isolatedSessionID := ""
		if sessionResolution.SessionID != "" {
			isolatedSessionID = resolveOpenAICodexUpstreamSessionHeaderValue(account.ID, sessionResolution.SessionID, sessionResolution.SessionSource)
			req.Header.Set("session_id", isolatedSessionID)
		}
		if sessionResolution.ConversationID != "" {
			req.Header.Set("conversation_id", resolveOpenAICodexUpstreamSessionHeaderValue(account.ID, sessionResolution.ConversationID, sessionResolution.ConversationSource))
		}
		if requestID := resolveOpenAICodexUpstreamClientRequestID(account.ID, profile.Headers.ClientRequestID, isolatedSessionID); requestID != "" {
			req.Header.Set("x-client-request-id", requestID)
		}
		if betaFeatures := resolveOpenAICodexUpstreamBetaFeatures(account); betaFeatures != "" {
			req.Header.Set("x-codex-beta-features", betaFeatures)
		} else {
			req.Header.Del("x-codex-beta-features")
		}
		if turnMetadata := resolveOpenAICodexUpstreamTurnMetadata(account.ID, profile.Headers.TurnMetadata); turnMetadata != "" {
			req.Header.Set(openAIWSTurnMetadataHeader, turnMetadata)
		} else {
			req.Header.Del(openAIWSTurnMetadataHeader)
		}
		if acceptLanguage := resolveOpenAICodexUpstreamAcceptLanguage(account); acceptLanguage != "" {
			req.Header.Set("accept-language", acceptLanguage)
		}
		if windowID := resolveOpenAICodexUpstreamWindowID(account.ID); windowID != "" {
			req.Header.Set(openAICodexMetadataWindowIDKey, windowID)
		}
		if subagent := resolveOpenAICodexUpstreamSubagent(profile, body); subagent != "" {
			req.Header.Set(openAICodexMetadataSubagentKey, subagent)
		} else {
			req.Header.Del(openAICodexMetadataSubagentKey)
		}
		if parentThreadID := resolveOpenAICodexUpstreamParentThreadID(account.ID, profile, body); parentThreadID != "" {
			req.Header.Set(openAICodexMetadataParentThreadIDKey, parentThreadID)
		} else {
			req.Header.Del(openAICodexMetadataParentThreadIDKey)
		}
		if profile.CompactPath {
			if installationID := resolveOpenAICodexUpstreamInstallationID(account.ID); installationID != "" {
				req.Header.Set(openAICodexMetadataInstallationIDKey, installationID)
			}
		} else {
			req.Header.Del(openAICodexMetadataInstallationIDKey)
		}
	}

	userAgent := ""
	if account != nil && account.Type == AccountTypeOAuth {
		userAgent = resolveOpenAICodexUpstreamUserAgent(account)
	} else {
		userAgent = policy.ResolveUserAgent(account, forceCodexCLI, false)
	}
	if userAgent != "" {
		req.Header.Set("user-agent", userAgent)
	}
	if req.Header.Get("content-type") == "" {
		req.Header.Set("content-type", "application/json")
	}

	return req, nil
}

func setOpenAICompatPromptCacheSessionID(c *gin.Context, account *Account, req *http.Request, promptCacheKey string) {
	if req == nil {
		return
	}
	promptCacheKey = strings.TrimSpace(promptCacheKey)
	if promptCacheKey == "" {
		return
	}

	accountID := int64(0)
	if account != nil {
		accountID = account.ID
	}
	req.Header.Set("session_id", generateSessionUUID(resolveOpenAICodexUpstreamSessionHeaderValue(accountID, promptCacheKey, "prompt_cache_key")))
}
