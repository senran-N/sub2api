package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/handler"
	servermiddleware "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func newGatewayRoutesTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	openaiGateway := &handler.OpenAIGatewayHandler{}
	compatibleRuntime := handler.NewCompatibleGatewayRuntimeHandler(&handler.CompatibleGatewayTextHandler{}, openaiGateway)
	grokGateway := handler.NewGrokGatewayHandler(compatibleRuntime, nil, nil, nil)
	compatibleGateway := handler.NewCompatibleGatewayHandler(compatibleRuntime, grokGateway, nil, nil)

	RegisterGatewayRoutes(
		router,
		&handler.Handlers{
			Gateway:           &handler.GatewayHandler{},
			CompatibleGateway: compatibleGateway,
			GrokGateway:       grokGateway,
		},
		servermiddleware.APIKeyAuthMiddleware(func(c *gin.Context) {
			groupID := int64(1)
			c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
				ID:      1,
				UserID:  1,
				GroupID: &groupID,
				Group: &service.Group{
					ID:       groupID,
					Platform: service.PlatformOpenAI,
				},
				User: &service.User{ID: 1},
			})
			c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 1, Concurrency: 1})
			c.Next()
		}),
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
	)

	return router
}

func TestGatewayRoutesOpenAIResponsesCompactPathIsRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	for _, path := range []string{"/v1/responses/compact", "/responses/compact"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-5"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit OpenAI responses handler", path)
	}
}

func TestGatewayRoutesGrokCompatPathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/grok/v1/models", body: ``},
		{method: http.MethodPost, path: "/grok/v1/chat/completions", body: `{"model":"grok-3"}`},
		{method: http.MethodPost, path: "/grok/v1/responses", body: `{"model":"grok-3"}`},
		{method: http.MethodPost, path: "/grok/v1/messages", body: `{"model":"grok-3"}`},
		{method: http.MethodPost, path: "/grok/v1/images/generations", body: `{"model":"grok-2-image"}`},
		{method: http.MethodGet, path: "/grok/v1/videos/job_123", body: ``},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		if tt.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit grok provider routes", tt.path)
	}
}

func TestGatewayRoutesOpenAICompatPassthroughPathsAreRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/v1/images/generations", body: `{"model":"grok-2-image"}`},
		{method: http.MethodPost, path: "/v1/audio/transcriptions", body: `{"model":"grok-4-voice"}`},
		{method: http.MethodPost, path: "/v1/tts", body: `{"model":"grok-4-voice"}`},
		{method: http.MethodPost, path: "/v1/stt", body: `{"model":"grok-4-voice"}`},
		{method: http.MethodPost, path: "/v1/embeddings", body: `{"model":"text-embedding-3-large"}`},
		{method: http.MethodPost, path: "/v1/moderations", body: `{"model":"omni-moderation-latest"}`},
		{method: http.MethodPost, path: "/v1/realtime/client_secrets", body: `{"model":"grok-4-fast-reasoning"}`},
		{method: http.MethodGet, path: "/v1/videos/job_123", body: ``},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		if tt.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should hit OpenAI passthrough handler", tt.path)
	}
}
