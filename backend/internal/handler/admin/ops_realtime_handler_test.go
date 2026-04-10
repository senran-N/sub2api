package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type opsRealtimeTrafficResponse struct {
	Code int `json:"code"`
	Data struct {
		Enabled              bool                                 `json:"enabled"`
		Timestamp            time.Time                            `json:"timestamp"`
		RuntimeObservability service.RuntimeObservabilitySnapshot `json:"runtime_observability"`
	} `json:"data"`
}

func newOpsRealtimeTestRouter(handler *OpsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ops/realtime-traffic", handler.GetRealtimeTrafficSummary)
	return r
}

func TestOpsRealtimeTrafficDisabledIncludesRuntimeObservability(t *testing.T) {
	settingRepo := newTestSettingRepo()
	require.NoError(t, settingRepo.Set(nil, service.SettingKeyOpsRealtimeMonitoringEnabled, "false"))

	svc := service.NewOpsService(nil, settingRepo, &config.Config{
		Ops: config.OpsConfig{Enabled: true},
	}, nil, nil, nil, nil, nil, nil, nil, nil)

	handler := NewOpsHandler(svc)
	router := newOpsRealtimeTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/ops/realtime-traffic?window=1min", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var payload opsRealtimeTrafficResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)
	require.False(t, payload.Data.Enabled)
	require.False(t, payload.Data.Timestamp.IsZero())
	require.Equal(t, service.SnapshotRuntimeObservability(), payload.Data.RuntimeObservability)
}
