package service

import (
	"context"
	"strings"
	"time"
)

// CodexChainState captures the continuation/account affinity resolved from
// previous_response_id, session hash, and transport class.
type CodexChainState struct {
	GroupID              int64
	PreferredConnID      string
	PreferredConnSource  string
	PreviousResponseID   string
	ResponseAccountID    int64
	SessionHash          string
	SessionStickyAccount int64
	Transport            OpenAIUpstreamTransport
}

type codexChainStateInput struct {
	GroupID                 int64
	HasExplicitContinuation bool
	PreviousResponseID      string
	SessionHash             string
	StoreDisabled           bool
	Transport               OpenAIUpstreamTransport
}

type codexChainBinding struct {
	AccountID     int64
	ConnID        string
	GroupID       int64
	ResponseID    string
	ResponseTTL   time.Duration
	SessionHash   string
	SessionTTL    time.Duration
	StoreDisabled bool
	Transport     OpenAIUpstreamTransport
}

func (s *OpenAIGatewayService) resolveCodexChainState(ctx context.Context, input codexChainStateInput) CodexChainState {
	state := CodexChainState{
		GroupID:             input.GroupID,
		PreferredConnSource: codexTransportPreferredConnSourceNone,
		PreviousResponseID:  strings.TrimSpace(input.PreviousResponseID),
		SessionHash:         strings.TrimSpace(input.SessionHash),
		Transport:           input.Transport,
	}

	if state.SessionHash != "" {
		if stickyAccountID, err := s.getStickySessionAccountID(ctx, optionalOpenAIGroupID(state.GroupID), state.SessionHash); err == nil && stickyAccountID > 0 {
			state.SessionStickyAccount = stickyAccountID
		}
	}

	if state.Transport != OpenAIUpstreamTransportResponsesWebsocketV2 {
		return state
	}

	store := s.getOpenAIWSStateStore()
	if store == nil {
		return state
	}

	if state.PreviousResponseID != "" {
		if accountID, err := store.GetResponseAccount(ctx, state.GroupID, state.PreviousResponseID); err == nil && accountID > 0 {
			state.ResponseAccountID = accountID
		}
		if connID, ok := store.GetResponseConn(state.PreviousResponseID); ok {
			state.PreferredConnID = connID
			state.PreferredConnSource = codexTransportPreferredConnSourcePreviousResponse
		}
	}

	if state.PreferredConnID == "" &&
		input.StoreDisabled &&
		state.PreviousResponseID == "" &&
		input.HasExplicitContinuation &&
		state.SessionHash != "" {
		if connID, ok := store.GetSessionConn(state.GroupID, state.SessionHash); ok {
			state.PreferredConnID = connID
			state.PreferredConnSource = codexTransportPreferredConnSourceSession
		}
	}

	return state
}

func (s *OpenAIGatewayService) bindCodexChainSuccess(ctx context.Context, store OpenAIWSStateStore, binding codexChainBinding) {
	if s == nil || store == nil || binding.Transport != OpenAIUpstreamTransportResponsesWebsocketV2 {
		return
	}

	responseID := strings.TrimSpace(binding.ResponseID)
	connID := strings.TrimSpace(binding.ConnID)
	if responseID != "" {
		responseTTL := binding.ResponseTTL
		if responseTTL <= 0 {
			responseTTL = s.openAIWSResponseStickyTTL()
		}
		logOpenAIWSBindResponseAccountWarn(
			binding.GroupID,
			binding.AccountID,
			responseID,
			store.BindResponseAccount(ctx, binding.GroupID, responseID, binding.AccountID, responseTTL),
		)
		if connID != "" {
			store.BindResponseConn(responseID, connID, responseTTL)
		}
	}

	if binding.StoreDisabled && strings.TrimSpace(binding.SessionHash) != "" && connID != "" {
		sessionTTL := binding.SessionTTL
		if sessionTTL <= 0 {
			sessionTTL = s.openAIWSSessionStickyTTL()
		}
		store.BindSessionConn(binding.GroupID, binding.SessionHash, connID, sessionTTL)
	}
}

func optionalOpenAIGroupID(groupID int64) *int64 {
	if groupID <= 0 {
		return nil
	}
	return &groupID
}
