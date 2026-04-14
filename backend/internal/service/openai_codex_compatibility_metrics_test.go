package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSnapshotOpenAICodexCompatibilityMetrics(t *testing.T) {
	before := SnapshotOpenAICodexCompatibilityMetrics()

	store := NewOpenAIWSStateStore(nil)
	svc := &OpenAIGatewayService{cfg: &config.Config{}}
	svc.bindCodexSessionTransport(store, 9, "session_hash_metrics", OpenAIUpstreamTransportResponsesWebsocketV2, true)
	svc.bindCodexSessionTransport(store, 9, "session_hash_metrics", OpenAIUpstreamTransportResponsesWebsocketV2, false)
	svc.bindCodexSessionTransport(store, 9, "session_hash_metrics", OpenAIUpstreamTransportHTTPSSE, false)

	afterBind := SnapshotOpenAICodexCompatibilityMetrics()
	require.GreaterOrEqual(t, afterBind.SessionTransportWarmupIgnoredTotal, before.SessionTransportWarmupIgnoredTotal+1)
	require.GreaterOrEqual(t, afterBind.SessionTransportBindWSTotal, before.SessionTransportBindWSTotal+1)
	require.GreaterOrEqual(t, afterBind.SessionTransportBindHTTPTotal, before.SessionTransportBindHTTPTotal+1)
	require.GreaterOrEqual(t, afterBind.SessionTransportHTTPDowngradeTotal, before.SessionTransportHTTPDowngradeTotal+1)

	gin.SetMode(gin.TestMode)
	c := newTestGinContext(t)
	groupID := int64(9)
	c.Set("api_key", &APIKey{GroupID: &groupID})

	sessionHash := DeriveSessionHashFromSeed("pcache_metrics_state")
	store.BindSessionTransport(groupID, sessionHash, OpenAIUpstreamTransportHTTPSSE, time.Minute)
	store.MarkSessionTransportFallback(groupID, sessionHash, time.Minute)

	svc.cfg.Gateway.OpenAIWS.FallbackCooldownSeconds = 30
	svc.markOpenAIWSFallbackCooling(22, "upgrade_required")
	svc.openaiWSStateStore = store

	state := svc.resolveCodexTransportState(c, codexTransportStateInput{
		AccountID:      22,
		PromptCacheKey: "pcache_metrics_state",
	})
	require.Equal(t, OpenAIUpstreamTransportHTTPSSE, state.PreferredTransport)
	require.True(t, state.PreferredHTTPFallback)

	afterResolve := SnapshotOpenAICodexCompatibilityMetrics()
	require.GreaterOrEqual(t, afterResolve.TransportFallbackCoolingHitTotal, before.TransportFallbackCoolingHitTotal+1)
	require.GreaterOrEqual(t, afterResolve.SessionPreferredTransportHitTotal, before.SessionPreferredTransportHitTotal+1)
	require.GreaterOrEqual(t, afterResolve.SessionPreferredTransportHTTPHitTotal, before.SessionPreferredTransportHTTPHitTotal+1)
	require.GreaterOrEqual(t, afterResolve.SessionHTTPFallbackHitTotal, before.SessionHTTPFallbackHitTotal+1)
}

func newTestGinContext(t *testing.T) *gin.Context {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	return c
}
