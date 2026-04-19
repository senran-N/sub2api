package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/ctxkey"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	servermiddleware "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCompatibleGatewayHandlerModels_DefaultsGrokModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/grok/v1/models", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.ForcePlatform, service.PlatformGrok))
	c.Request = req
	c.Set(string(servermiddleware.ContextKeyForcePlatform), service.PlatformGrok)

	h := NewCompatibleGatewayHandler(NewCompatibleGatewayRuntimeHandler(&CompatibleGatewayTextHandler{}, &OpenAIGatewayHandler{}), nil, nil, nil)
	h.Models(c)

	require.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Object string       `json:"object"`
		Data   []grok.Model `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Equal(t, "list", response.Object)
	require.NotEmpty(t, response.Data)
	require.Equal(t, "grok-3", response.Data[0].ID)
	require.Equal(t, "xai", response.Data[0].OwnedBy)
}

func TestGrokGatewayHandlerModels_ForceBindsPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/grok/v1/models", nil)

	h := NewGrokGatewayHandler(NewCompatibleGatewayRuntimeHandler(&CompatibleGatewayTextHandler{}, &OpenAIGatewayHandler{}), nil, nil, nil)
	h.Models(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, service.PlatformGrok, c.GetString(string(servermiddleware.ContextKeyForcePlatform)))
	require.Equal(t, service.PlatformGrok, c.Request.Context().Value(ctxkey.ForcePlatform))
	require.False(t, service.AllowsGrokSessionTextRuntime(c.Request.Context()))

	var response struct {
		Object string       `json:"object"`
		Data   []grok.Model `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Equal(t, "list", response.Object)
	require.NotEmpty(t, response.Data)
	require.Equal(t, "grok-3", response.Data[0].ID)
}

func TestGrokGatewayHandlerMessages_ForceBindsPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/grok/v1/messages", nil)

	h := NewGrokGatewayHandler(NewCompatibleGatewayRuntimeHandler(&CompatibleGatewayTextHandler{}, &OpenAIGatewayHandler{}), nil, nil, nil)
	h.Messages(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Equal(t, service.PlatformGrok, c.GetString(string(servermiddleware.ContextKeyForcePlatform)))
	require.Equal(t, service.PlatformGrok, c.Request.Context().Value(ctxkey.ForcePlatform))
	require.True(t, service.AllowsGrokSessionTextRuntime(c.Request.Context()))
}

func TestGrokGatewayHandlerTextRoutes_EnableSessionTextRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		path   string
		invoke func(*GrokGatewayHandler, *gin.Context)
		method string
	}{
		{
			name:   "responses",
			path:   "/grok/v1/responses",
			method: http.MethodPost,
			invoke: (*GrokGatewayHandler).Responses,
		},
		{
			name:   "chat_completions",
			path:   "/grok/v1/chat/completions",
			method: http.MethodPost,
			invoke: (*GrokGatewayHandler).ChatCompletions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, tt.path, nil)

			h := NewGrokGatewayHandler(nil, nil, nil, nil)
			tt.invoke(h, c)

			require.Equal(t, service.PlatformGrok, c.Request.Context().Value(ctxkey.ForcePlatform))
			require.True(t, service.AllowsGrokSessionTextRuntime(c.Request.Context()))
		})
	}
}

func TestGrokGatewayHandlerNonTextRoutes_KeepSessionRuntimeDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		path   string
		method string
		invoke func(*GrokGatewayHandler, *gin.Context)
	}{
		{
			name:   "responses_websocket",
			path:   "/grok/v1/responses",
			method: http.MethodGet,
			invoke: (*GrokGatewayHandler).ResponsesWebSocket,
		},
		{
			name:   "passthrough",
			path:   "/grok/v1/images/generations",
			method: http.MethodPost,
			invoke: (*GrokGatewayHandler).Passthrough,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, tt.path, nil)

			h := NewGrokGatewayHandler(nil, nil, nil, nil)
			tt.invoke(h, c)

			require.Equal(t, service.PlatformGrok, c.Request.Context().Value(ctxkey.ForcePlatform))
			require.False(t, service.AllowsGrokSessionTextRuntime(c.Request.Context()))
		})
	}
}

func TestBuildNativeGatewayMappedModels_UsesClaudeModelShape(t *testing.T) {
	models := buildNativeGatewayMappedModels([]string{"claude-opus-4-6"})

	typed, ok := models.([]claude.Model)
	require.True(t, ok)
	require.Len(t, typed, 1)
	require.Equal(t, "claude-opus-4-6", typed[0].ID)
	require.Equal(t, "model", typed[0].Type)
}
