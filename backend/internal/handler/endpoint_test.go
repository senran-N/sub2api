package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

// ──────────────────────────────────────────────────────────
// NormalizeInboundEndpoint
// ──────────────────────────────────────────────────────────

func TestNormalizeInboundEndpoint(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// Direct canonical paths.
		{"/v1/messages", EndpointMessages},
		{"/v1/chat/completions", EndpointChatCompletions},
		{"/v1/responses", EndpointResponses},
		{"/v1/embeddings", EndpointEmbeddings},
		{"/v1/moderations", EndpointModerations},
		{"/v1/images/generations", EndpointImages},
		{"/v1/audio/speech", EndpointAudioSpeech},
		{"/v1/audio/transcriptions", EndpointAudioTranscribe},
		{"/v1/audio/translations", EndpointAudioTranslate},
		{"/v1/tts", EndpointTTS},
		{"/v1/stt", EndpointSTT},
		{"/v1/realtime/client_secrets", EndpointRealtimeSecret},
		{"/v1/videos/job_123", EndpointVideos},
		{"/v1beta/models", EndpointGeminiModels},

		// Prefixed paths (antigravity, openai).
		{"/antigravity/v1/messages", EndpointMessages},
		{"/openai/v1/responses", EndpointResponses},
		{"/openai/v1/responses/compact", EndpointResponses},
		{"/antigravity/v1beta/models/gemini:generateContent", EndpointGeminiModels},

		// Gin route patterns with wildcards.
		{"/v1beta/models/*modelAction", EndpointGeminiModels},
		{"/v1/responses/*subpath", EndpointResponses},

		// Unknown path is returned as-is.
		{"/v1/embeddings", "/v1/embeddings"},
		{"", ""},
		{"  /v1/messages  ", EndpointMessages},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			require.Equal(t, tt.want, NormalizeInboundEndpoint(tt.path))
		})
	}
}

// ──────────────────────────────────────────────────────────
// DeriveUpstreamEndpoint
// ──────────────────────────────────────────────────────────

func TestDeriveUpstreamEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		inbound  string
		rawPath  string
		platform string
		want     string
	}{
		// Anthropic.
		{"anthropic messages", EndpointMessages, "/v1/messages", service.PlatformAnthropic, EndpointMessages},

		// Gemini.
		{"gemini models", EndpointGeminiModels, "/v1beta/models/gemini:gen", service.PlatformGemini, EndpointGeminiModels},

		// OpenAI — always /v1/responses.
		{"openai responses root", EndpointResponses, "/v1/responses", service.PlatformOpenAI, EndpointResponses},
		{"openai responses compact", EndpointResponses, "/openai/v1/responses/compact", service.PlatformOpenAI, "/v1/responses/compact"},
		{"openai responses nested", EndpointResponses, "/openai/v1/responses/compact/detail", service.PlatformOpenAI, "/v1/responses/compact/detail"},
		{"openai from messages", EndpointMessages, "/v1/messages", service.PlatformOpenAI, EndpointResponses},
		{"openai from completions", EndpointChatCompletions, "/v1/chat/completions", service.PlatformOpenAI, EndpointResponses},
		{"openai passthrough images", EndpointImages, "/v1/images/generations", service.PlatformOpenAI, EndpointResponses},
		{"openai passthrough audio", EndpointAudioTranscribe, "/v1/audio/transcriptions", service.PlatformOpenAI, EndpointResponses},
		{"openai passthrough videos", EndpointVideos, "/v1/videos/job_123", service.PlatformOpenAI, EndpointResponses},
		{"grok from messages", EndpointMessages, "/grok/v1/messages", service.PlatformGrok, EndpointResponses},
		{"grok responses compact", EndpointResponses, "/grok/v1/responses/compact", service.PlatformGrok, "/v1/responses/compact"},

		// Antigravity — uses inbound to pick Claude vs Gemini upstream.
		{"antigravity claude", EndpointMessages, "/antigravity/v1/messages", service.PlatformAntigravity, EndpointMessages},
		{"antigravity gemini", EndpointGeminiModels, "/antigravity/v1beta/models", service.PlatformAntigravity, EndpointGeminiModels},

		// Unknown platform — passthrough.
		{"unknown platform", "/v1/embeddings", "/v1/embeddings", "unknown", "/v1/embeddings"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, DeriveUpstreamEndpoint(tt.inbound, tt.rawPath, tt.platform))
		})
	}
}

// ──────────────────────────────────────────────────────────
// responsesSubpathSuffix
// ──────────────────────────────────────────────────────────

func TestResponsesSubpathSuffix(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"/v1/responses", ""},
		{"/v1/responses/", ""},
		{"/v1/responses/compact", "/compact"},
		{"/openai/v1/responses/compact/detail", "/compact/detail"},
		{"/v1/messages", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			require.Equal(t, tt.want, responsesSubpathSuffix(tt.raw))
		})
	}
}

// ──────────────────────────────────────────────────────────
// InboundEndpointMiddleware + context helpers
// ──────────────────────────────────────────────────────────

func TestInboundEndpointMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(InboundEndpointMiddleware())

	var captured string
	router.POST("/v1/messages", func(c *gin.Context) {
		captured = GetInboundEndpoint(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, EndpointMessages, captured)
}

func TestGetInboundEndpoint_FallbackWithoutMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/antigravity/v1/messages", nil)

	// Middleware did not run — fallback to normalizing c.Request.URL.Path.
	got := GetInboundEndpoint(c)
	require.Equal(t, EndpointMessages, got)
}

func TestGetUpstreamEndpoint_FullFlow(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses/compact", nil)

	// Simulate middleware.
	c.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(c.Request.URL.Path))

	got := GetUpstreamEndpoint(c, service.PlatformOpenAI)
	require.Equal(t, "/v1/responses/compact", got)
}

func TestGetUpstreamEndpoint_OpenAIPassthroughChatCompletions(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(c.Request.URL.Path))
	c.Set("openai_passthrough", true)

	got := GetUpstreamEndpoint(c, service.PlatformOpenAI)
	require.Equal(t, EndpointChatCompletions, got)
}

func TestGetUpstreamEndpoint_OpenAIPassthroughCompatibleEndpoints(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "/v1/images/generations", want: EndpointImages},
		{path: "/v1/audio/transcriptions", want: EndpointAudioTranscribe},
		{path: "/v1/tts", want: EndpointTTS},
		{path: "/v1/videos/job_123", want: EndpointVideos},
		{path: "/v1/realtime/client_secrets", want: EndpointRealtimeSecret},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)
		c.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(c.Request.URL.Path))
		c.Set("openai_passthrough", true)

		got := GetUpstreamEndpoint(c, service.PlatformOpenAI)
		require.Equal(t, tt.want, got)
	}
}

func TestGetUpstreamEndpoint_GrokPassthroughCompatibleEndpoints(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "/grok/v1/images/generations", want: EndpointImages},
		{path: "/grok/v1/chat/completions", want: EndpointChatCompletions},
		{path: "/grok/v1/videos/job_123", want: EndpointVideos},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)
		c.Set(ctxKeyInboundEndpoint, NormalizeInboundEndpoint(c.Request.URL.Path))
		c.Set("openai_passthrough", true)

		got := GetUpstreamEndpoint(c, service.PlatformGrok)
		require.Equal(t, tt.want, got)
	}
}

func TestGuessPlatformFromPath(t *testing.T) {
	require.Equal(t, service.PlatformOpenAI, guessPlatformFromPath("/v1/chat/completions"))
	require.Equal(t, service.PlatformOpenAI, guessPlatformFromPath("/openai/v1/responses/compact"))
	require.Equal(t, service.PlatformGrok, guessPlatformFromPath("/grok/v1/messages"))
	require.Equal(t, service.PlatformGemini, guessPlatformFromPath("/v1beta/models/gemini:generateContent"))
}
