package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/service"
)

// ──────────────────────────────────────────────────────────
// Canonical inbound / upstream endpoint paths.
// All normalization and derivation reference this single set
// of constants — add new paths HERE when a new API surface
// is introduced.
// ──────────────────────────────────────────────────────────

const (
	EndpointMessages        = "/v1/messages"
	EndpointChatCompletions = "/v1/chat/completions"
	EndpointResponses       = "/v1/responses"
	EndpointEmbeddings      = "/v1/embeddings"
	EndpointModerations     = "/v1/moderations"
	EndpointImages          = "/v1/images"
	EndpointVideos          = "/v1/videos"
	EndpointAudioSpeech     = "/v1/audio/speech"
	EndpointAudioTranscribe = "/v1/audio/transcriptions"
	EndpointAudioTranslate  = "/v1/audio/translations"
	EndpointTTS             = "/v1/tts"
	EndpointSTT             = "/v1/stt"
	EndpointRealtimeSecret  = "/v1/realtime/client_secrets"
	EndpointGeminiModels    = "/v1beta/models"
)

// gin.Context keys used by the middleware and helpers below.
const (
	ctxKeyInboundEndpoint = "_gateway_inbound_endpoint"
)

// ──────────────────────────────────────────────────────────
// Normalization functions
// ──────────────────────────────────────────────────────────

// NormalizeInboundEndpoint maps a raw request path (which may carry
// prefixes like /antigravity and /openai) to its canonical form.
//
//	"/antigravity/v1/messages"   → "/v1/messages"
//	"/v1/chat/completions"       → "/v1/chat/completions"
//	"/openai/v1/responses/foo"   → "/v1/responses"
//	"/v1beta/models/gemini:gen"  → "/v1beta/models"
func NormalizeInboundEndpoint(path string) string {
	path = strings.TrimSpace(path)
	switch {
	case strings.Contains(path, EndpointChatCompletions):
		return EndpointChatCompletions
	case strings.Contains(path, EndpointMessages):
		return EndpointMessages
	case strings.Contains(path, EndpointResponses):
		return EndpointResponses
	case strings.Contains(path, EndpointEmbeddings):
		return EndpointEmbeddings
	case strings.Contains(path, EndpointModerations):
		return EndpointModerations
	case strings.Contains(path, EndpointAudioSpeech):
		return EndpointAudioSpeech
	case strings.Contains(path, EndpointAudioTranscribe):
		return EndpointAudioTranscribe
	case strings.Contains(path, EndpointAudioTranslate):
		return EndpointAudioTranslate
	case strings.Contains(path, EndpointRealtimeSecret):
		return EndpointRealtimeSecret
	case strings.Contains(path, EndpointTTS):
		return EndpointTTS
	case strings.Contains(path, EndpointSTT):
		return EndpointSTT
	case strings.Contains(path, EndpointImages):
		return EndpointImages
	case strings.Contains(path, EndpointVideos):
		return EndpointVideos
	case strings.Contains(path, EndpointGeminiModels):
		return EndpointGeminiModels
	default:
		return path
	}
}

// DeriveUpstreamEndpoint determines the upstream endpoint from the
// account platform and the normalized inbound endpoint.
//
// Platform-specific rules:
//   - OpenAI defaults to /v1/responses (with optional subpath such as
//     /v1/responses/compact preserved from the raw URL). Context-aware
//     passthrough exceptions are handled by GetUpstreamEndpoint.
//   - Anthropic  → /v1/messages
//   - Gemini     → /v1beta/models
//   - Antigravity routes may target either Claude or Gemini, so the
//     inbound endpoint is used to distinguish.
func DeriveUpstreamEndpoint(inbound, rawRequestPath, platform string) string {
	inbound = strings.TrimSpace(inbound)

	switch platform {
	case service.PlatformOpenAI:
		// OpenAI forwards everything to the Responses API.
		// Preserve subresource suffix (e.g. /v1/responses/compact).
		if suffix := responsesSubpathSuffix(rawRequestPath); suffix != "" {
			return EndpointResponses + suffix
		}
		return EndpointResponses

	case service.PlatformAnthropic:
		return EndpointMessages

	case service.PlatformGemini:
		return EndpointGeminiModels

	case service.PlatformAntigravity:
		// Antigravity accounts serve both Claude and Gemini.
		if inbound == EndpointGeminiModels {
			return EndpointGeminiModels
		}
		return EndpointMessages
	}

	// Unknown platform — fall back to inbound.
	return inbound
}

// responsesSubpathSuffix extracts the part after "/responses" in a raw
// request path, e.g. "/openai/v1/responses/compact" → "/compact".
// Returns "" when there is no meaningful suffix.
func responsesSubpathSuffix(rawPath string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(rawPath), "/")
	idx := strings.LastIndex(trimmed, "/responses")
	if idx < 0 {
		return ""
	}
	suffix := trimmed[idx+len("/responses"):]
	if suffix == "" || suffix == "/" {
		return ""
	}
	if !strings.HasPrefix(suffix, "/") {
		return ""
	}
	return suffix
}

// ──────────────────────────────────────────────────────────
// Middleware
// ──────────────────────────────────────────────────────────

// InboundEndpointMiddleware normalizes the request path and stores the
// canonical inbound endpoint in gin.Context so that every handler in
// the chain can read it via GetInboundEndpoint.
//
// Apply this middleware to all gateway route groups.
func InboundEndpointMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" && c.Request != nil && c.Request.URL != nil {
			path = c.Request.URL.Path
		}
		c.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(path))
		c.Next()
	}
}

// ──────────────────────────────────────────────────────────
// Context helpers — used by handlers before building
// RecordUsageInput / RecordUsageLongContextInput.
// ──────────────────────────────────────────────────────────

// GetInboundEndpoint returns the canonical inbound endpoint stored by
// InboundEndpointMiddleware. If the middleware did not run (e.g. in
// tests), it falls back to normalizing c.FullPath() on the fly.
func GetInboundEndpoint(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyInboundEndpoint); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	// Fallback: normalize on the fly.
	path := ""
	if c != nil {
		path = c.FullPath()
		if path == "" && c.Request != nil && c.Request.URL != nil {
			path = c.Request.URL.Path
		}
	}
	return NormalizeInboundEndpoint(path)
}

// GetUpstreamEndpoint derives the upstream endpoint from the context
// and the account platform. Handlers call this after scheduling an
// account, passing account.Platform.
func GetUpstreamEndpoint(c *gin.Context, platform string) string {
	inbound := GetInboundEndpoint(c)
	rawPath := ""
	if c != nil && c.Request != nil && c.Request.URL != nil {
		rawPath = c.Request.URL.Path
	}
	if platform == service.PlatformOpenAI && isOpenAIPassthroughRequest(c) {
		return deriveOpenAICompatiblePassthroughEndpoint(inbound, rawPath)
	}
	return DeriveUpstreamEndpoint(inbound, rawPath, platform)
}

func deriveOpenAICompatiblePassthroughEndpoint(inbound, rawPath string) string {
	switch inbound {
	case EndpointResponses:
		if suffix := responsesSubpathSuffix(rawPath); suffix != "" {
			return EndpointResponses + suffix
		}
		return EndpointResponses
	case EndpointChatCompletions,
		EndpointEmbeddings,
		EndpointModerations,
		EndpointImages,
		EndpointVideos,
		EndpointAudioSpeech,
		EndpointAudioTranscribe,
		EndpointAudioTranslate,
		EndpointTTS,
		EndpointSTT,
		EndpointRealtimeSecret:
		return inbound
	default:
		return inbound
	}
}

func isOpenAIPassthroughRequest(c *gin.Context) bool {
	if c == nil {
		return false
	}
	value, ok := c.Get("openai_passthrough")
	if !ok {
		return false
	}
	passthrough, _ := value.(bool)
	return passthrough
}
