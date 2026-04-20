package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupSettingModelCatalogRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewSettingHandler(nil, nil, nil, nil)
	router.GET("/api/v1/admin/model-catalog", handler.GetModelCatalog)
	return router
}

func TestSettingHandlerGetModelCatalog_Grok(t *testing.T) {
	router := setupSettingModelCatalogRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-catalog?platform=grok", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			Platform string `json:"platform"`
			Models   []struct {
				ID             string   `json:"id"`
				DisplayName    string   `json:"display_name"`
				Capability     string   `json:"capability"`
				ProtocolFamily string   `json:"protocol_family"`
				RequiredTier   string   `json:"required_tier"`
				Aliases        []string `json:"aliases"`
			} `json:"models"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "grok", resp.Data.Platform)
	require.NotEmpty(t, resp.Data.Models)
	require.Equal(t, "grok-3", resp.Data.Models[0].ID)
	require.Equal(t, "Grok 3", resp.Data.Models[0].DisplayName)
	require.Equal(t, "chat", resp.Data.Models[0].Capability)
	require.Equal(t, "responses", resp.Data.Models[0].ProtocolFamily)
	require.Equal(t, "basic", resp.Data.Models[0].RequiredTier)
	require.Contains(t, resp.Data.Models[0].Aliases, "grok-4.20-auto")
}

func TestSettingHandlerGetModelCatalog_UnsupportedPlatform(t *testing.T) {
	router := setupSettingModelCatalogRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-catalog?platform=openai", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
