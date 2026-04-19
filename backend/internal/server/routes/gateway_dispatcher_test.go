package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/handler"
	servermiddleware "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGatewayProtocolDispatcherModels_UsesCompatibleGatewayForGrok(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)

	groupID := int64(1)
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformGrok,
		},
	})

	dispatcher := newGatewayProtocolDispatcher(&handler.Handlers{
		Gateway: &handler.GatewayHandler{},
		CompatibleGateway: handler.NewCompatibleGatewayHandler(
			handler.NewCompatibleGatewayRuntimeHandler(&handler.CompatibleGatewayTextHandler{}, &handler.OpenAIGatewayHandler{}),
			nil,
			nil,
			nil,
		),
		GrokGateway: handler.NewGrokGatewayHandler(
			handler.NewCompatibleGatewayRuntimeHandler(&handler.CompatibleGatewayTextHandler{}, &handler.OpenAIGatewayHandler{}),
			nil,
			nil,
			nil,
		),
	})
	dispatcher.Models(c)

	require.Equal(t, http.StatusOK, w.Code)
	modelIDs := decodeGatewayModelIDs(t, w.Body.Bytes())
	require.Contains(t, modelIDs, "grok-3")
	require.NotContains(t, modelIDs, "claude-opus-4-6")
}

func TestGatewayProtocolDispatcherModels_UsesNativeGatewayForAnthropic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)

	groupID := int64(2)
	c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
		GroupID: &groupID,
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformAnthropic,
		},
	})

	dispatcher := newGatewayProtocolDispatcher(&handler.Handlers{
		Gateway: &handler.GatewayHandler{},
		CompatibleGateway: handler.NewCompatibleGatewayHandler(
			handler.NewCompatibleGatewayRuntimeHandler(&handler.CompatibleGatewayTextHandler{}, &handler.OpenAIGatewayHandler{}),
			nil,
			nil,
			nil,
		),
		GrokGateway: handler.NewGrokGatewayHandler(
			handler.NewCompatibleGatewayRuntimeHandler(&handler.CompatibleGatewayTextHandler{}, &handler.OpenAIGatewayHandler{}),
			nil,
			nil,
			nil,
		),
	})
	dispatcher.Models(c)

	require.Equal(t, http.StatusOK, w.Code)
	modelIDs := decodeGatewayModelIDs(t, w.Body.Bytes())
	require.Contains(t, modelIDs, "claude-opus-4-6")
	require.NotContains(t, modelIDs, "grok-3")
}

func decodeGatewayModelIDs(t *testing.T, body []byte) []string {
	t.Helper()

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &response))

	ids := make([]string, 0, len(response.Data))
	for _, model := range response.Data {
		ids = append(ids, model.ID)
	}
	return ids
}
