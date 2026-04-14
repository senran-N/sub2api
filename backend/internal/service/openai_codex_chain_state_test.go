package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolveCodexChainState_PrefersPreviousResponseAffinity(t *testing.T) {
	ctx := context.Background()
	groupID := int64(301)
	sessionHash := "session-hash-301"
	cache := &stubGatewayCache{
		sessionBindings: map[string]int64{
			"openai:" + sessionHash: 19,
		},
	}
	store := NewOpenAIWSStateStore(cache)
	svc := &OpenAIGatewayService{
		cache:              cache,
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_chain", 11, time.Minute))
	store.BindResponseConn("resp_prev_chain", "conn_prev_chain", time.Minute)
	store.BindSessionConn(groupID, sessionHash, "conn_session_chain", time.Minute)

	state := svc.resolveCodexChainState(ctx, codexChainStateInput{
		GroupID:                 groupID,
		HasExplicitContinuation: true,
		PreviousResponseID:      "resp_prev_chain",
		SessionHash:             sessionHash,
		StoreDisabled:           true,
		Transport:               OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	require.Equal(t, groupID, state.GroupID)
	require.Equal(t, "resp_prev_chain", state.PreviousResponseID)
	require.Equal(t, int64(11), state.ResponseAccountID)
	require.Equal(t, int64(19), state.SessionStickyAccount)
	require.Equal(t, "conn_prev_chain", state.PreferredConnID)
	require.Equal(t, codexTransportPreferredConnSourcePreviousResponse, state.PreferredConnSource)
}

func TestResolveCodexChainState_UsesSessionConnForStoreDisabledContinuation(t *testing.T) {
	ctx := context.Background()
	groupID := int64(302)
	sessionHash := "session-hash-302"
	cache := &stubGatewayCache{
		sessionBindings: map[string]int64{
			"openai:" + sessionHash: 23,
		},
	}
	store := NewOpenAIWSStateStore(cache)
	svc := &OpenAIGatewayService{
		cache:              cache,
		openaiWSStateStore: store,
	}

	store.BindSessionConn(groupID, sessionHash, "conn_session_only", time.Minute)

	state := svc.resolveCodexChainState(ctx, codexChainStateInput{
		GroupID:                 groupID,
		HasExplicitContinuation: true,
		SessionHash:             sessionHash,
		StoreDisabled:           true,
		Transport:               OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	require.Equal(t, int64(23), state.SessionStickyAccount)
	require.Equal(t, "conn_session_only", state.PreferredConnID)
	require.Equal(t, codexTransportPreferredConnSourceSession, state.PreferredConnSource)
}

func TestBindCodexChainSuccess_BindsResponseAndSessionConn(t *testing.T) {
	ctx := context.Background()
	groupID := int64(303)
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	svc := &OpenAIGatewayService{
		cache:              cache,
		openaiWSStateStore: store,
	}

	svc.bindCodexChainSuccess(ctx, store, codexChainBinding{
		AccountID:     29,
		ConnID:        "conn_bind_303",
		GroupID:       groupID,
		ResponseID:    "resp_bind_303",
		ResponseTTL:   time.Minute,
		SessionHash:   "session-bind-303",
		SessionTTL:    time.Minute,
		StoreDisabled: true,
		Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	accountID, err := store.GetResponseAccount(ctx, groupID, "resp_bind_303")
	require.NoError(t, err)
	require.Equal(t, int64(29), accountID)

	connID, ok := store.GetResponseConn("resp_bind_303")
	require.True(t, ok)
	require.Equal(t, "conn_bind_303", connID)

	sessionConnID, ok := store.GetSessionConn(groupID, "session-bind-303")
	require.True(t, ok)
	require.Equal(t, "conn_bind_303", sessionConnID)
}
