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
	svc.bindCodexSessionTransport(store, 9, "session_hash_metrics", OpenAIUpstreamTransportResponsesWebsocketV2, true, true)
	svc.bindCodexSessionTransport(store, 9, "session_hash_metrics", OpenAIUpstreamTransportResponsesWebsocketV2, false, true)
	svc.bindCodexSessionTransport(store, 9, "session_hash_metrics", OpenAIUpstreamTransportHTTPSSE, false, true)

	afterBind := SnapshotOpenAICodexCompatibilityMetrics()
	require.GreaterOrEqual(t, afterBind.SessionTransportWarmupIgnoredTotal, before.SessionTransportWarmupIgnoredTotal+1)
	require.GreaterOrEqual(t, afterBind.SessionTransportBindWSTotal, before.SessionTransportBindWSTotal+1)
	require.GreaterOrEqual(t, afterBind.SessionTransportBindHTTPTotal, before.SessionTransportBindHTTPTotal+1)
	require.GreaterOrEqual(t, afterBind.SessionTransportHTTPDowngradeTotal, before.SessionTransportHTTPDowngradeTotal+1)

	gin.SetMode(gin.TestMode)
	c := newTestGinContext(t)
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 codex_cli_rs/0.200.0")
	c.Request.Header.Set("originator", "codex_vscode")
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

	ObserveOpenAICodexRequestProfile(CodexRequestProfile{
		NativeClient:   true,
		OfficialClient: true,
		CodexVersion:   "0.1.0",
		WireAPI:        CodexWireAPIResponsesHTTP,
	})
	ObserveOpenAICodexRequestProfile(CodexRequestProfile{
		NativeClient:   true,
		OfficialClient: true,
		CodexVersion:   "0.1.0",
		WireAPI:        CodexWireAPIResponsesWebSocket,
		Warmup:         true,
	})
	ObserveOpenAICodexSchedulingDecision(CodexRequestProfile{
		NativeClient:   true,
		OfficialClient: true,
		CodexVersion:   "0.1.0",
		WireAPI:        CodexWireAPIResponsesHTTP,
	}, OpenAIAccountScheduleDecision{StickyPreviousHit: true})
	ObserveOpenAICodexSchedulingDecision(CodexRequestProfile{
		NativeClient:   true,
		OfficialClient: true,
		CodexVersion:   "0.1.0",
		WireAPI:        CodexWireAPIResponsesWebSocket,
	}, OpenAIAccountScheduleDecision{StickySessionHit: true})

	decision := CodexRecoveryPolicy{}.Apply(map[string]any{
		"previous_response_id": "resp_metrics_1",
	}, CodexRecoveryPolicyInput{
		Reason:                    codexRecoveryReasonPreviousResponseNotFound,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportHTTPSSE,
	})
	require.True(t, decision.Applied)

	afterObserve := SnapshotOpenAICodexCompatibilityMetrics()
	require.GreaterOrEqual(t, afterObserve.Summary.OfficialRequestTotal, before.Summary.OfficialRequestTotal+2)
	require.GreaterOrEqual(t, afterObserve.Summary.OfficialResponsesHTTPRequests, before.Summary.OfficialResponsesHTTPRequests+1)
	require.GreaterOrEqual(t, afterObserve.Summary.OfficialResponsesWebSocketRequests, before.Summary.OfficialResponsesWebSocketRequests+1)
	require.GreaterOrEqual(t, afterObserve.Summary.ChainSelectionTotal, before.Summary.ChainSelectionTotal+2)
	require.GreaterOrEqual(t, afterObserve.Summary.ChainHitTotal, before.Summary.ChainHitTotal+2)
	require.GreaterOrEqual(t, afterObserve.Summary.ChainPreviousResponseHitTotal, before.Summary.ChainPreviousResponseHitTotal+1)
	require.GreaterOrEqual(t, afterObserve.Summary.ChainSessionHitTotal, before.Summary.ChainSessionHitTotal+1)
	require.GreaterOrEqual(t, afterObserve.Summary.MinimalRewriteAppliedTotal, before.Summary.MinimalRewriteAppliedTotal+1)
	require.Greater(t, afterObserve.Summary.ChainHitRate, 0.0)
	require.Greater(t, afterObserve.Summary.SessionHTTPFallbackHitRate, 0.0)
	require.Greater(t, afterObserve.Summary.SessionTransportHTTPDowngradeRate, 0.0)
	require.Greater(t, afterObserve.Summary.MinimalRewriteRate, 0.0)

	versionBefore := codexVersionMetricTotal(before, "0.1.0")
	versionAfter := codexVersionMetricTotal(afterObserve, "0.1.0")
	require.GreaterOrEqual(t, versionAfter.OfficialRequestTotal, versionBefore.OfficialRequestTotal+2)
	require.GreaterOrEqual(t, versionAfter.ResponsesHTTPRequestsTotal, versionBefore.ResponsesHTTPRequestsTotal+1)
	require.GreaterOrEqual(t, versionAfter.ResponsesWebSocketRequestsTotal, versionBefore.ResponsesWebSocketRequestsTotal+1)
	require.GreaterOrEqual(t, versionAfter.WarmupTotal, versionBefore.WarmupTotal+1)
	require.GreaterOrEqual(t, versionAfter.ChainSelectionTotal, versionBefore.ChainSelectionTotal+2)
	require.GreaterOrEqual(t, versionAfter.ChainHitTotal, versionBefore.ChainHitTotal+2)
	require.Greater(t, versionAfter.ChainHitRate, 0.0)

	httpBefore := codexTransportMetricTotal(before, string(CodexWireAPIResponsesHTTP))
	httpAfter := codexTransportMetricTotal(afterObserve, string(CodexWireAPIResponsesHTTP))
	require.GreaterOrEqual(t, httpAfter.OfficialRequestTotal, httpBefore.OfficialRequestTotal+1)
	require.GreaterOrEqual(t, httpAfter.ChainSelectionTotal, httpBefore.ChainSelectionTotal+1)
	require.GreaterOrEqual(t, httpAfter.ChainHitTotal, httpBefore.ChainHitTotal+1)
	require.GreaterOrEqual(t, httpAfter.MinimalRewriteAppliedTotal, httpBefore.MinimalRewriteAppliedTotal+1)

	wsBefore := codexTransportMetricTotal(before, string(CodexWireAPIResponsesWebSocket))
	wsAfter := codexTransportMetricTotal(afterObserve, string(CodexWireAPIResponsesWebSocket))
	require.GreaterOrEqual(t, wsAfter.OfficialRequestTotal, wsBefore.OfficialRequestTotal+1)
	require.GreaterOrEqual(t, wsAfter.WarmupTotal, wsBefore.WarmupTotal+1)
	require.GreaterOrEqual(t, wsAfter.ChainSelectionTotal, wsBefore.ChainSelectionTotal+1)
	require.GreaterOrEqual(t, wsAfter.ChainHitTotal, wsBefore.ChainHitTotal+1)
}

func TestObserveOpenAICodexSchedulingDecision_WarmupSkipped(t *testing.T) {
	before := SnapshotOpenAICodexCompatibilityMetrics()

	ObserveOpenAICodexRequestProfile(CodexRequestProfile{
		NativeClient:   true,
		OfficialClient: true,
		CodexVersion:   "warmup-skip",
		WireAPI:        CodexWireAPIResponsesWebSocket,
		Warmup:         true,
	})
	ObserveOpenAICodexSchedulingDecision(CodexRequestProfile{
		NativeClient:   true,
		OfficialClient: true,
		CodexVersion:   "warmup-skip",
		WireAPI:        CodexWireAPIResponsesWebSocket,
		Warmup:         true,
	}, OpenAIAccountScheduleDecision{StickySessionHit: true})

	after := SnapshotOpenAICodexCompatibilityMetrics()

	require.GreaterOrEqual(t, after.Summary.OfficialRequestTotal, before.Summary.OfficialRequestTotal+1)
	require.Equal(t, before.Summary.ChainSelectionTotal, after.Summary.ChainSelectionTotal)
	require.Equal(t, before.Summary.ChainHitTotal, after.Summary.ChainHitTotal)
	require.Equal(t, before.Summary.ChainSessionHitTotal, after.Summary.ChainSessionHitTotal)

	versionBefore := codexVersionMetricTotal(before, "warmup-skip")
	versionAfter := codexVersionMetricTotal(after, "warmup-skip")
	require.GreaterOrEqual(t, versionAfter.OfficialRequestTotal, versionBefore.OfficialRequestTotal+1)
	require.GreaterOrEqual(t, versionAfter.WarmupTotal, versionBefore.WarmupTotal+1)
	require.Equal(t, versionBefore.ChainSelectionTotal, versionAfter.ChainSelectionTotal)
	require.Equal(t, versionBefore.ChainHitTotal, versionAfter.ChainHitTotal)

	transportBefore := codexTransportMetricTotal(before, string(CodexWireAPIResponsesWebSocket))
	transportAfter := codexTransportMetricTotal(after, string(CodexWireAPIResponsesWebSocket))
	require.GreaterOrEqual(t, transportAfter.OfficialRequestTotal, transportBefore.OfficialRequestTotal+1)
	require.GreaterOrEqual(t, transportAfter.WarmupTotal, transportBefore.WarmupTotal+1)
	require.Equal(t, transportBefore.ChainSelectionTotal, transportAfter.ChainSelectionTotal)
	require.Equal(t, transportBefore.ChainHitTotal, transportAfter.ChainHitTotal)
}

func TestResolveCodexTransportState_NonOfficialMetricsSkipped(t *testing.T) {
	before := SnapshotOpenAICodexCompatibilityMetrics()

	store := NewOpenAIWSStateStore(nil)
	svc := &OpenAIGatewayService{
		cfg:                &config.Config{},
		cache:              &stubGatewayCache{},
		openaiWSStateStore: store,
	}

	gin.SetMode(gin.TestMode)
	c := newTestGinContext(t)
	groupID := int64(19)
	c.Set("api_key", &APIKey{GroupID: &groupID})

	sessionHash := DeriveSessionHashFromSeed("pcache_non_official_metrics")
	store.BindSessionTransport(groupID, sessionHash, OpenAIUpstreamTransportHTTPSSE, time.Minute)
	store.MarkSessionTransportFallback(groupID, sessionHash, time.Minute)
	svc.cfg.Gateway.OpenAIWS.FallbackCooldownSeconds = 30
	svc.markOpenAIWSFallbackCooling(42, "upgrade_required")

	state := svc.resolveCodexTransportState(c, codexTransportStateInput{
		AccountID:      42,
		PromptCacheKey: "pcache_non_official_metrics",
	})
	require.Equal(t, OpenAIUpstreamTransportHTTPSSE, state.PreferredTransport)
	require.True(t, state.PreferredHTTPFallback)
	require.False(t, state.TrackCompatibilityMetrics)

	after := SnapshotOpenAICodexCompatibilityMetrics()
	require.Equal(t, before.TransportFallbackCoolingHitTotal, after.TransportFallbackCoolingHitTotal)
	require.Equal(t, before.SessionPreferredTransportHitTotal, after.SessionPreferredTransportHitTotal)
	require.Equal(t, before.SessionPreferredTransportHTTPHitTotal, after.SessionPreferredTransportHTTPHitTotal)
	require.Equal(t, before.SessionHTTPFallbackHitTotal, after.SessionHTTPFallbackHitTotal)
}

func TestObserveOpenAICodexCompatibilityMetrics_ForceCodexCLICompatibilitySkipped(t *testing.T) {
	before := SnapshotOpenAICodexCompatibilityMetrics()

	profile := CodexRequestProfile{
		OfficialClient:       true,
		OfficialClientReason: CodexOfficialClientReasonForceCodexCLI,
		NativeClient:         false,
		CodexVersion:         "force-codex-cli",
		WireAPI:              CodexWireAPIResponsesHTTP,
	}

	ObserveOpenAICodexRequestProfile(profile)
	ObserveOpenAICodexSchedulingDecision(profile, OpenAIAccountScheduleDecision{StickyPreviousHit: true})

	after := SnapshotOpenAICodexCompatibilityMetrics()
	require.Equal(t, before.Summary.OfficialRequestTotal, after.Summary.OfficialRequestTotal)
	require.Equal(t, before.Summary.ChainSelectionTotal, after.Summary.ChainSelectionTotal)
	require.Equal(t, before.Summary.ChainHitTotal, after.Summary.ChainHitTotal)
	require.Equal(t, codexVersionMetricTotal(before, "force-codex-cli"), codexVersionMetricTotal(after, "force-codex-cli"))
	require.Equal(t, codexTransportMetricTotal(before, string(CodexWireAPIResponsesHTTP)), codexTransportMetricTotal(after, string(CodexWireAPIResponsesHTTP)))
}

func codexVersionMetricTotal(
	snapshot OpenAICodexCompatibilityMetricsSnapshot,
	version string,
) OpenAICodexCompatibilityVersionMetricsSnapshot {
	for _, item := range snapshot.ByVersion {
		if item.Version == version {
			return item
		}
	}
	return OpenAICodexCompatibilityVersionMetricsSnapshot{}
}

func codexTransportMetricTotal(
	snapshot OpenAICodexCompatibilityMetricsSnapshot,
	transport string,
) OpenAICodexCompatibilityTransportMetricsSnapshot {
	for _, item := range snapshot.ByTransport {
		if item.Transport == transport {
			return item
		}
	}
	return OpenAICodexCompatibilityTransportMetricsSnapshot{}
}

func newTestGinContext(t *testing.T) *gin.Context {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	return c
}
