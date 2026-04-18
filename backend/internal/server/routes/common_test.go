package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newCommonRoutesTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterCommonRoutes(router)
	return router
}

func TestCommonRoutesTelemetryBatchReturnsOK(t *testing.T) {
	router := newCommonRoutesTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/event_logging/batch", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestCommonRoutesPolicyLimitsReturnsEmptyRestrictions(t *testing.T) {
	router := newCommonRoutesTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/claude_code/policy_limits", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, emptyClaudeCodePolicyLimitsETag, w.Header().Get("ETag"))

	var response struct {
		Restrictions map[string]any `json:"restrictions"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.NotNil(t, response.Restrictions)
	require.Empty(t, response.Restrictions)
}

func TestCommonRoutesPolicyLimitsReturnsNotModifiedForMatchingETag(t *testing.T) {
	router := newCommonRoutesTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/claude_code/policy_limits", nil)
	req.Header.Set("If-None-Match", emptyClaudeCodePolicyLimitsETag)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotModified, w.Code)
	require.Empty(t, w.Body.String())
}

func TestCommonRoutesSettingsReturnsEmptyObject(t *testing.T) {
	router := newCommonRoutesTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/claude_code/settings", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{}`, w.Body.String())
}

func TestCommonRoutesDomainInfoAllowsFetch(t *testing.T) {
	router := newCommonRoutesTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/web/domain_info?domain=example.com", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"can_fetch":true}`, w.Body.String())
}
