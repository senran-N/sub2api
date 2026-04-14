package service

import (
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/openai"
)

// CodexNativeMutationPolicy centralizes the conservative upstream defaults for
// official Codex traffic so HTTP forward, passthrough, and WS header builders
// stop re-implementing slightly different fallback rules.
type CodexNativeMutationPolicy struct {
	Profile CodexRequestProfile
}

func NewCodexNativeMutationPolicy(profile CodexRequestProfile) CodexNativeMutationPolicy {
	return CodexNativeMutationPolicy{Profile: profile}
}

func (p CodexNativeMutationPolicy) ResolveAccept(fallback string) string {
	if value := strings.TrimSpace(p.Profile.Headers.Accept); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func (p CodexNativeMutationPolicy) ResolveAcceptLanguage() string {
	return strings.TrimSpace(p.Profile.Headers.AcceptLanguage)
}

func (p CodexNativeMutationPolicy) ResolveOpenAIBeta(fallback string) string {
	if value := strings.TrimSpace(p.Profile.Headers.OpenAIBeta); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func (p CodexNativeMutationPolicy) ResolveOriginator() string {
	if value := strings.TrimSpace(p.Profile.Headers.Originator); value != "" {
		return value
	}
	if p.Profile.OfficialClient {
		return "codex_cli_rs"
	}
	return "opencode"
}

func (p CodexNativeMutationPolicy) ResolveUserAgent(account *Account, forceCodexCLI bool, forceOAuthCodexFallback bool) string {
	if account != nil {
		if customUA := strings.TrimSpace(account.GetOpenAIUserAgent()); customUA != "" {
			return customUA
		}
	}

	if userAgent := strings.TrimSpace(p.Profile.Headers.UserAgent); userAgent != "" {
		if !forceOAuthCodexFallback || p.Profile.OfficialClient || openai.IsCodexCLIRequest(userAgent) {
			return userAgent
		}
	}

	if forceCodexCLI || p.Profile.OfficialClient || (account != nil && account.Type == AccountTypeOAuth && forceOAuthCodexFallback) {
		return codexCLIUserAgent
	}
	return ""
}

func (p CodexNativeMutationPolicy) ResolveVersion(fallback string) string {
	if value := strings.TrimSpace(p.Profile.Headers.Version); value != "" {
		return value
	}
	if value := strings.TrimSpace(p.Profile.CodexVersion); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func (p CodexNativeMutationPolicy) ResolveRequestSessionID() string {
	if sessionID := strings.TrimSpace(p.Profile.Headers.SessionID); sessionID != "" {
		return sessionID
	}
	if conversationID := strings.TrimSpace(p.Profile.Headers.ConversationID); conversationID != "" {
		return conversationID
	}
	return strings.TrimSpace(p.Profile.Body.PromptCacheKey)
}

func (p CodexNativeMutationPolicy) ResolveSessionHeaders(promptCacheKey string) openAIWSSessionHeaderResolution {
	return p.resolveSessionHeaders(promptCacheKey, "", false, false)
}

func (p CodexNativeMutationPolicy) ResolveOAuthSessionHeaders(promptCacheKey, compactSessionID string, includeConversationPromptFallback bool) openAIWSSessionHeaderResolution {
	return p.resolveSessionHeaders(promptCacheKey, compactSessionID, includeConversationPromptFallback, p.Profile.CompactPath)
}

func (p CodexNativeMutationPolicy) resolveSessionHeaders(
	promptCacheKey, compactSessionID string,
	includeConversationPromptFallback bool,
	enableCompactPathFallback bool,
) openAIWSSessionHeaderResolution {
	resolution := openAIWSSessionHeaderResolution{
		SessionSource:      "none",
		ConversationSource: "none",
	}

	if sessionID := strings.TrimSpace(p.Profile.Headers.SessionID); sessionID != "" {
		resolution.SessionID = sessionID
		resolution.SessionSource = "header_session_id"
	}
	if conversationID := strings.TrimSpace(p.Profile.Headers.ConversationID); conversationID != "" {
		resolution.ConversationID = conversationID
		resolution.ConversationSource = "header_conversation_id"
		if resolution.SessionID == "" {
			resolution.SessionID = conversationID
			resolution.SessionSource = "header_conversation_id"
		}
	}

	if enableCompactPathFallback {
		if resolution.SessionID == "" {
			if compactSessionID = strings.TrimSpace(compactSessionID); compactSessionID != "" {
				resolution.SessionID = compactSessionID
				resolution.SessionSource = "compact_path_fallback"
			}
		}
		return resolution
	}

	if promptCacheKey = strings.TrimSpace(promptCacheKey); promptCacheKey != "" {
		if resolution.SessionID == "" {
			resolution.SessionID = promptCacheKey
			resolution.SessionSource = "prompt_cache_key"
		}
		if includeConversationPromptFallback && resolution.ConversationID == "" {
			resolution.ConversationID = promptCacheKey
			resolution.ConversationSource = "prompt_cache_key"
		}
	}

	return resolution
}
