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

func TestResolveCodexTransportState_PromptCacheFallbackAndSessionContinuation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.StoreDisabledConnMode = "adaptive"
	cfg.Gateway.OpenAIWS.FallbackCooldownSeconds = 30

	store := NewOpenAIWSStateStore(&stubGatewayCache{})
	svc := &OpenAIGatewayService{
		cfg:                cfg,
		cache:              &stubGatewayCache{},
		openaiWSStateStore: store,
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	groupID := int64(201)
	c.Set("api_key", &APIKey{GroupID: &groupID})

	sessionHash, _ := openAIWSSessionHashesFromID("pcache_turn_1")
	store.BindSessionTurnState(groupID, sessionHash, "turn_state_restored", time.Minute)
	store.BindSessionConn(groupID, sessionHash, "conn_session_1", time.Minute)
	store.BindSessionTransport(groupID, sessionHash, OpenAIUpstreamTransportHTTPSSE, time.Minute)
	svc.markOpenAIWSFallbackCooling(17, "upgrade_required")

	state := svc.resolveCodexTransportState(c, codexTransportStateInput{
		AccountID:             17,
		HasFunctionCallOutput: true,
		LastFailureReason:     "read_event",
		PromptCacheKey:        "pcache_turn_1",
		StoreDisabled:         true,
	})

	require.Equal(t, groupID, state.GroupID)
	require.Equal(t, sessionHash, state.SessionHash)
	require.True(t, state.HasExplicitContinuation)
	require.Equal(t, "turn_state_restored", state.TurnState)
	require.True(t, state.TurnStateRestored)
	require.Equal(t, "conn_session_1", state.PreferredConnID)
	require.Equal(t, codexTransportPreferredConnSourceSession, state.PreferredConnSource)
	require.Equal(t, OpenAIUpstreamTransportHTTPSSE, state.PreferredTransport)
	require.Equal(t, codexTransportPreferredTransportSourceSession, state.PreferredTransportSource)
	require.Equal(t, openAIWSStoreDisabledConnModeAdaptive, state.StoreDisabledConnMode)
	require.False(t, state.ForceNewConn)
	require.True(t, state.FallbackCooling)
	require.False(t, state.Warmup)
}

func TestResolveCodexTransportState_PrefersPreviousResponseConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := NewOpenAIWSStateStore(&stubGatewayCache{})
	svc := &OpenAIGatewayService{
		cfg:                &config.Config{},
		cache:              &stubGatewayCache{},
		openaiWSStateStore: store,
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	groupID := int64(202)
	c.Set("api_key", &APIKey{GroupID: &groupID})

	store.BindResponseConn("resp_prev_1", "conn_response_1", time.Minute)
	sessionHash, _ := openAIWSSessionHashesFromID("pcache_turn_2")
	store.BindSessionConn(groupID, sessionHash, "conn_session_2", time.Minute)

	state := svc.resolveCodexTransportState(c, codexTransportStateInput{
		PreviousResponseID: "resp_prev_1",
		PromptCacheKey:     "pcache_turn_2",
		StoreDisabled:      true,
	})

	require.Equal(t, "conn_response_1", state.PreferredConnID)
	require.Equal(t, codexTransportPreferredConnSourcePreviousResponse, state.PreferredConnSource)
	require.True(t, state.HasExplicitContinuation)
	require.False(t, state.ForceNewConn)
}

func TestResolveCodexTransportState_IngressFallbackSessionAndWarmup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	store := NewOpenAIWSStateStore(&stubGatewayCache{})
	svc := &OpenAIGatewayService{
		cfg:                cfg,
		cache:              &stubGatewayCache{},
		openaiWSStateStore: store,
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/openai/v1/responses", nil)
	SetOpenAIClientTransport(c, OpenAIClientTransportWS)
	groupID := int64(203)
	c.Set("api_key", &APIKey{ID: 21, UserID: 34, GroupID: &groupID})

	body := []byte(`{"type":"response.create","model":"gpt-5.1","stream":false,"generate":false}`)
	state := svc.resolveCodexTransportState(c, codexTransportStateInput{
		AccountID:            19,
		Body:                 body,
		PreferIngressSession: true,
	})

	require.Equal(t, svc.GenerateOpenAIWSIngressSessionHash(c, body), state.SessionHash)
	require.NotEmpty(t, state.SessionHash)
	require.True(t, state.Warmup)
	require.False(t, state.HasExplicitContinuation)
	require.False(t, state.TurnStateRestored)
	require.Equal(t, codexTransportPreferredConnSourceNone, state.PreferredConnSource)
}

func TestResolveCodexTransportState_ForceNewConnForFreshStoreDisabledTurn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.StoreDisabledConnMode = "adaptive"

	svc := &OpenAIGatewayService{
		cfg:                cfg,
		cache:              &stubGatewayCache{},
		openaiWSStateStore: NewOpenAIWSStateStore(&stubGatewayCache{}),
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)

	state := svc.resolveCodexTransportState(c, codexTransportStateInput{
		LastFailureReason: "policy_violation",
		PromptCacheKey:    "pcache_turn_3",
		StoreDisabled:     true,
	})

	require.False(t, state.HasExplicitContinuation)
	require.NotEmpty(t, state.SessionHash)
	require.Equal(t, openAIWSStoreDisabledConnModeAdaptive, state.StoreDisabledConnMode)
	require.True(t, state.ForceNewConn)
	require.Empty(t, state.PreferredConnID)
}

func TestBindCodexSessionTransport_IgnoresWarmup(t *testing.T) {
	store := NewOpenAIWSStateStore(nil)
	svc := &OpenAIGatewayService{cfg: &config.Config{}}

	svc.bindCodexSessionTransport(store, 91, "session_hash_warmup", OpenAIUpstreamTransportResponsesWebsocketV2, true)
	_, ok := store.GetSessionTransport(91, "session_hash_warmup")
	require.False(t, ok)

	svc.bindCodexSessionTransport(store, 91, "session_hash_warmup", OpenAIUpstreamTransportResponsesWebsocketV2, false)
	transport, ok := store.GetSessionTransport(91, "session_hash_warmup")
	require.True(t, ok)
	require.Equal(t, OpenAIUpstreamTransportResponsesWebsocketV2, transport)
}
