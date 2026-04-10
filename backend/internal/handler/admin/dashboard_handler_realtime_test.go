package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type dashboardRealtimeMetricsResponse struct {
	Code int `json:"code"`
	Data struct {
		ActiveRequests       int                                  `json:"active_requests"`
		RequestsPerMinute    int                                  `json:"requests_per_minute"`
		AverageResponseTime  int                                  `json:"average_response_time"`
		ErrorRate            float64                              `json:"error_rate"`
		Timestamp            time.Time                            `json:"timestamp"`
		RuntimeObservability service.RuntimeObservabilitySnapshot `json:"runtime_observability"`
	} `json:"data"`
}

func TestDashboardRealtimeMetricsIncludesRuntimeObservability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDashboardHandler(nil, nil)
	router := gin.New()
	router.GET("/admin/dashboard/realtime", handler.GetRealtimeMetrics)

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard/realtime", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var payload dashboardRealtimeMetricsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.Equal(t, 0, payload.Data.ActiveRequests)
	require.Equal(t, 0, payload.Data.RequestsPerMinute)
	require.Equal(t, 0, payload.Data.AverageResponseTime)
	require.Zero(t, payload.Data.ErrorRate)
	require.False(t, payload.Data.Timestamp.IsZero())
	require.Equal(t, service.SnapshotRuntimeObservability(), payload.Data.RuntimeObservability)
}
