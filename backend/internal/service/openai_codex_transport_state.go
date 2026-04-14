package service

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	codexTransportPreferredConnSourceNone             = ""
	codexTransportPreferredConnSourcePreviousResponse = "previous_response_id"
	codexTransportPreferredConnSourceSession          = "session"
	codexTransportPreferredTransportSourceNone        = ""
	codexTransportPreferredTransportSourceSession     = "session"
)

// CodexTransportState captures the resolved transport/continuation state shared by
// HTTP-triggered WS forwarding and WS ingress turn bootstrap.
type CodexTransportState struct {
	FallbackCooling          bool
	ForceNewConn             bool
	GroupID                  int64
	HasExplicitContinuation  bool
	PreferredConnID          string
	PreferredConnSource      string
	PreferredHTTPFallback    bool
	PreferredTransport       OpenAIUpstreamTransport
	PreferredTransportSource string
	SessionHash              string
	StoreDisabled            bool
	StoreDisabledConnMode    string
	TurnState                string
	TurnStateRestored        bool
	Warmup                   bool
}

type codexTransportStateInput struct {
	AccountID             int64
	Body                  []byte
	HasFunctionCallOutput bool
	LastFailureReason     string
	PreferIngressSession  bool
	PreviousResponseID    string
	PromptCacheKey        string
	StoreDisabled         bool
	TurnState             string
}

func (s *OpenAIGatewayService) resolveCodexTransportState(c *gin.Context, input codexTransportStateInput) CodexTransportState {
	state := CodexTransportState{
		FallbackCooling:          s.isOpenAIWSFallbackCooling(input.AccountID),
		GroupID:                  getOpenAIGroupIDFromContext(c),
		PreferredConnSource:      codexTransportPreferredConnSourceNone,
		PreferredTransport:       OpenAIUpstreamTransportAny,
		PreferredTransportSource: codexTransportPreferredTransportSourceNone,
		StoreDisabled:            input.StoreDisabled,
		StoreDisabledConnMode:    s.openAIWSStoreDisabledConnMode(),
		TurnState:                strings.TrimSpace(input.TurnState),
	}

	forceCodexCLI := s != nil && s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
	if len(input.Body) > 0 {
		profile := GetCodexRequestProfile(c, input.Body, forceCodexCLI)
		state.Warmup = profile.Warmup
		if profile.OfficialClient && state.Warmup {
			recordOpenAICodexTransportWarmup()
		}
	}

	if input.PreferIngressSession {
		state.SessionHash = s.GenerateOpenAIWSIngressSessionHash(c, input.Body)
	} else {
		state.SessionHash = s.GenerateSessionHash(c, nil)
		if state.SessionHash == "" {
			var legacySessionHash string
			state.SessionHash, legacySessionHash = openAIWSSessionHashesFromID(input.PromptCacheKey)
			attachOpenAILegacySessionHashToGin(c, legacySessionHash)
		}
	}

	state.HasExplicitContinuation = strings.TrimSpace(input.PreviousResponseID) != "" || state.TurnState != "" || input.HasFunctionCallOutput
	stateStore := s.getOpenAIWSStateStore()
	shouldRestoreTurnState := state.HasExplicitContinuation || input.PreferIngressSession
	if stateStore != nil && state.TurnState == "" && shouldRestoreTurnState && state.SessionHash != "" {
		if savedTurnState, ok := stateStore.GetSessionTurnState(state.GroupID, state.SessionHash); ok {
			state.TurnState = savedTurnState
			state.TurnStateRestored = true
		}
	}
	if preferredTransport, ok := s.getCodexSessionPreferredTransport(stateStore, state.GroupID, state.SessionHash); ok {
		state.PreferredTransport = preferredTransport
		state.PreferredTransportSource = codexTransportPreferredTransportSourceSession
		state.PreferredHTTPFallback = preferredTransport == OpenAIUpstreamTransportHTTPSSE && stateStore.HasSessionTransportFallback(state.GroupID, state.SessionHash)
		recordOpenAICodexSessionPreferredTransportHit(preferredTransport, state.PreferredHTTPFallback)
	}
	if state.FallbackCooling {
		recordOpenAICodexTransportFallbackCoolingHit()
	}

	resolveCtx := context.Background()
	if c != nil && c.Request != nil {
		resolveCtx = c.Request.Context()
	}
	chainState := s.resolveCodexChainState(resolveCtx, codexChainStateInput{
		GroupID:                 state.GroupID,
		HasExplicitContinuation: state.HasExplicitContinuation,
		PreviousResponseID:      input.PreviousResponseID,
		SessionHash:             state.SessionHash,
		StoreDisabled:           state.StoreDisabled,
		Transport:               OpenAIUpstreamTransportResponsesWebsocketV2,
	})
	state.PreferredConnID = chainState.PreferredConnID
	state.PreferredConnSource = chainState.PreferredConnSource

	forceNewConnByPolicy := shouldForceNewConnOnStoreDisabled(state.StoreDisabledConnMode, input.LastFailureReason)
	state.ForceNewConn = forceNewConnByPolicy && state.StoreDisabled && !state.HasExplicitContinuation && state.SessionHash != "" && state.PreferredConnID == ""
	return state
}

func (s *OpenAIGatewayService) getCodexSessionPreferredTransport(
	store OpenAIWSStateStore,
	groupID int64,
	sessionHash string,
) (OpenAIUpstreamTransport, bool) {
	if s == nil || store == nil || strings.TrimSpace(sessionHash) == "" {
		return OpenAIUpstreamTransportAny, false
	}
	return store.GetSessionTransport(groupID, sessionHash)
}

func (s *OpenAIGatewayService) bindCodexSessionTransport(
	store OpenAIWSStateStore,
	groupID int64,
	sessionHash string,
	transport OpenAIUpstreamTransport,
	warmup bool,
) {
	if s == nil || store == nil {
		return
	}
	if warmup {
		recordOpenAICodexSessionTransportBind(transport, true, false)
		return
	}
	normalizedSessionHash := strings.TrimSpace(sessionHash)
	normalizedTransport := normalizeOpenAIWSSessionTransport(transport)
	if normalizedSessionHash == "" || normalizedTransport == OpenAIUpstreamTransportAny {
		return
	}
	previousTransport, hasPreviousTransport := store.GetSessionTransport(groupID, normalizedSessionHash)
	httpDowngrade := normalizedTransport == OpenAIUpstreamTransportHTTPSSE &&
		hasPreviousTransport &&
		isOpenAIWSSessionWebsocketTransport(previousTransport)
	if normalizedTransport == OpenAIUpstreamTransportHTTPSSE {
		if httpDowngrade || store.HasSessionTransportFallback(groupID, normalizedSessionHash) {
			store.MarkSessionTransportFallback(groupID, normalizedSessionHash, s.openAIWSSessionStickyTTL())
		} else {
			store.ClearSessionTransportFallback(groupID, normalizedSessionHash)
		}
	} else {
		store.ClearSessionTransportFallback(groupID, normalizedSessionHash)
	}
	store.BindSessionTransport(groupID, normalizedSessionHash, normalizedTransport, s.openAIWSSessionStickyTTL())
	recordOpenAICodexSessionTransportBind(normalizedTransport, false, httpDowngrade)
}

func (s *OpenAIGatewayService) bindCodexSessionTransportFromBody(
	c *gin.Context,
	body []byte,
	transport OpenAIUpstreamTransport,
) {
	if s == nil || c == nil || len(body) == 0 {
		return
	}
	store := s.getOpenAIWSStateStore()
	if store == nil {
		return
	}
	forceCodexCLI := s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
	profile := GetCodexRequestProfile(c, body, forceCodexCLI)
	if !profile.OfficialClient || profile.Warmup {
		return
	}
	sessionHash := s.GenerateSessionHash(c, body)
	s.bindCodexSessionTransport(store, getOpenAIGroupIDFromContext(c), sessionHash, transport, false)
}

func (s *OpenAIGatewayService) resolveCodexPreferredTransport(
	ctx context.Context,
	groupID *int64,
	sessionHash string,
	requiredTransport OpenAIUpstreamTransport,
) OpenAIUpstreamTransport {
	if requiredTransport != OpenAIUpstreamTransportAny {
		return requiredTransport
	}
	store := s.getOpenAIWSStateStore()
	preferredTransport, ok := s.getCodexSessionPreferredTransport(store, derefGroupID(groupID), sessionHash)
	if !ok {
		return requiredTransport
	}
	return preferredTransport
}

func isOpenAIWSSessionWebsocketTransport(transport OpenAIUpstreamTransport) bool {
	switch normalizeOpenAIWSSessionTransport(transport) {
	case OpenAIUpstreamTransportResponsesWebsocket, OpenAIUpstreamTransportResponsesWebsocketV2:
		return true
	default:
		return false
	}
}
