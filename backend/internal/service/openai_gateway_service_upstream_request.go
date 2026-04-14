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
				return nil, err
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

		apiKeyID := getAPIKeyIDFromContext(c)
		sessionResolution := policy.ResolveOAuthSessionHeaders(promptCacheKey, resolveOpenAICompactSessionID(c), true)
		req.Header.Set("OpenAI-Beta", policy.ResolveOpenAIBeta("responses=experimental"))
		req.Header.Set("originator", policy.ResolveOriginator())
		if profile.CompactPath {
			req.Header.Set("accept", policy.ResolveAccept("application/json"))
			req.Header.Set("version", policy.ResolveVersion(codexCLIVersion))
		} else {
			req.Header.Set("accept", policy.ResolveAccept("text/event-stream"))
		}
		if sessionResolution.SessionID != "" {
			req.Header.Set("session_id", isolateOpenAISessionID(apiKeyID, sessionResolution.SessionID))
		}
		if sessionResolution.ConversationID != "" {
			req.Header.Set("conversation_id", isolateOpenAISessionID(apiKeyID, sessionResolution.ConversationID))
		}
	}

	if userAgent := policy.ResolveUserAgent(account, forceCodexCLI, false); userAgent != "" {
		req.Header.Set("user-agent", userAgent)
	}
	if req.Header.Get("content-type") == "" {
		req.Header.Set("content-type", "application/json")
	}

	return req, nil
}

func setOpenAICompatPromptCacheSessionID(c *gin.Context, req *http.Request, promptCacheKey string) {
	if req == nil {
		return
	}
	promptCacheKey = strings.TrimSpace(promptCacheKey)
	if promptCacheKey == "" {
		return
	}

	apiKeyID := getAPIKeyIDFromContext(c)
	req.Header.Set("session_id", generateSessionUUID(isolateOpenAISessionID(apiKeyID, promptCacheKey)))
}
