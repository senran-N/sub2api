package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractSessionID extracts the raw session ID from headers or body without hashing.
// Used by ForwardAsAnthropic to pass as prompt_cache_key for upstream cache.
func (s *OpenAIGatewayService) ExtractSessionID(c *gin.Context, body []byte) string {
	return resolveOpenAIRequestSessionID(c, body)
}

// GenerateSessionHash generates a sticky-session hash for OpenAI requests.
//
// Priority:
//  1. Header: session_id
//  2. Header: conversation_id
//  3. Body:   prompt_cache_key (opencode)
func (s *OpenAIGatewayService) GenerateSessionHash(c *gin.Context, body []byte) string {
	if c == nil {
		return ""
	}

	sessionID := resolveOpenAIRequestSessionID(c, body)
	if sessionID == "" {
		return ""
	}

	currentHash, legacyHash := deriveOpenAISessionHashes(sessionID)
	attachOpenAILegacySessionHashToGin(c, legacyHash)
	return currentHash
}

// GenerateSessionHashWithFallback 先按常规信号生成会话哈希；
// 当未携带 session_id/conversation_id/prompt_cache_key 时，使用 fallbackSeed 生成稳定哈希。
// 该方法用于 WS ingress，避免会话信号缺失时发生跨账号漂移。
func (s *OpenAIGatewayService) GenerateSessionHashWithFallback(c *gin.Context, body []byte, fallbackSeed string) string {
	sessionHash := s.GenerateSessionHash(c, body)
	if sessionHash != "" {
		return sessionHash
	}

	seed := strings.TrimSpace(fallbackSeed)
	if seed == "" {
		return ""
	}

	currentHash, legacyHash := deriveOpenAISessionHashes(seed)
	attachOpenAILegacySessionHashToGin(c, legacyHash)
	return currentHash
}

func BuildOpenAIWSIngressFallbackSessionSeed(apiKey *APIKey) string {
	if apiKey == nil {
		return ""
	}

	groupID := int64(0)
	if apiKey.GroupID != nil {
		groupID = *apiKey.GroupID
	}
	return fmt.Sprintf("openai_ws_ingress:%d:%d:%d", groupID, apiKey.UserID, apiKey.ID)
}

func (s *OpenAIGatewayService) GenerateOpenAIWSIngressSessionHash(c *gin.Context, body []byte) string {
	return s.GenerateSessionHashWithFallback(c, body, BuildOpenAIWSIngressFallbackSessionSeed(getOpenAIAPIKeyFromContext(c)))
}

func resolveOpenAIRequestSessionID(c *gin.Context, body []byte) string {
	if c == nil {
		return ""
	}

	sessionID := strings.TrimSpace(c.GetHeader("session_id"))
	if sessionID == "" {
		sessionID = strings.TrimSpace(c.GetHeader("conversation_id"))
	}
	if sessionID == "" && len(body) > 0 {
		sessionID = getOpenAIRequestMeta(c, body).PromptCacheKey
	}
	return sessionID
}

func resolveOpenAIUpstreamOriginator(c *gin.Context, isOfficialClient bool) string {
	if c != nil {
		if originator := strings.TrimSpace(c.GetHeader("originator")); originator != "" {
			return originator
		}
	}
	if isOfficialClient {
		return "codex_cli_rs"
	}
	return "opencode"
}

// BindStickySession sets session -> account binding with the configured sticky TTL.
func (s *OpenAIGatewayService) BindStickySession(ctx context.Context, groupID *int64, sessionHash string, accountID int64) error {
	if sessionHash == "" || accountID <= 0 {
		return nil
	}
	return s.setStickySessionAccountID(ctx, groupID, sessionHash, accountID, s.openAIWSSessionStickyTTL())
}
