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
)

// CodexTransportState captures the resolved transport/continuation state shared by
// HTTP-triggered WS forwarding and WS ingress turn bootstrap.
type CodexTransportState struct {
	FallbackCooling         bool
	ForceNewConn            bool
	GroupID                 int64
	HasExplicitContinuation bool
	PreferredConnID         string
	PreferredConnSource     string
	SessionHash             string
	StoreDisabled           bool
	StoreDisabledConnMode   string
	TurnState               string
	TurnStateRestored       bool
	Warmup                  bool
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
		FallbackCooling:       s.isOpenAIWSFallbackCooling(input.AccountID),
		GroupID:               getOpenAIGroupIDFromContext(c),
		PreferredConnSource:   codexTransportPreferredConnSourceNone,
		StoreDisabled:         input.StoreDisabled,
		StoreDisabledConnMode: s.openAIWSStoreDisabledConnMode(),
		TurnState:             strings.TrimSpace(input.TurnState),
	}

	forceCodexCLI := s != nil && s.cfg != nil && s.cfg.Gateway.ForceCodexCLI
	if len(input.Body) > 0 {
		state.Warmup = GetCodexRequestProfile(c, input.Body, forceCodexCLI).Warmup
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
