package service

import (
	"context"
	"errors"
	"strings"
)

type SelectionRequest struct {
	Provider           string
	Protocol           GatewayProtocol
	Model              string
	GroupID            *int64
	SessionHash        string
	PreviousResponseID string
	Transport          GatewayTransport
	ExcludedIDs        map[int64]struct{}
}

type SelectionDecision struct {
	Provider  string
	Protocol  GatewayProtocol
	Model     string
	AccountID int64
	Layer     string
	OpenAI    OpenAIAccountScheduleDecision
}

type SelectionKernel interface {
	Select(ctx context.Context, request SelectionRequest) (*AccountSelectionResult, SelectionDecision, error)
}

type OpenAISelectionKernel struct {
	service *OpenAIGatewayService
}

type GatewaySelectionKernel struct {
	service *GatewayService
}

func NewOpenAISelectionKernel(service *OpenAIGatewayService) *OpenAISelectionKernel {
	if service == nil {
		return nil
	}
	return &OpenAISelectionKernel{service: service}
}

func NewGatewaySelectionKernel(service *GatewayService) *GatewaySelectionKernel {
	if service == nil {
		return nil
	}
	return &GatewaySelectionKernel{service: service}
}

func (k *OpenAISelectionKernel) Select(ctx context.Context, request SelectionRequest) (*AccountSelectionResult, SelectionDecision, error) {
	if k == nil || k.service == nil {
		return nil, SelectionDecision{}, errors.New("openai selection kernel is not configured")
	}

	s := k.service
	decision := OpenAIAccountScheduleDecision{}
	requiredTransport := gatewayTransportToOpenAIUpstreamTransport(request.Transport)
	requiredTransport = s.resolveCodexPreferredTransport(ctx, request.GroupID, request.SessionHash, requiredTransport)
	scheduler := s.getOpenAIAccountScheduler()
	if scheduler == nil {
		selection, err := s.SelectAccountWithLoadAwareness(ctx, request.GroupID, request.SessionHash, request.Model, request.ExcludedIDs)
		decision.Layer = openAIAccountScheduleLayerLoadBalance
		return selection, buildOpenAISelectionDecision(request, selection, decision), err
	}

	stickyAccountID := s.resolveCodexChainState(ctx, codexChainStateInput{
		GroupID:     derefGroupID(request.GroupID),
		SessionHash: request.SessionHash,
		Transport:   requiredTransport,
	}).SessionStickyAccount

	selection, decision, err := scheduler.Select(ctx, OpenAIAccountScheduleRequest{
		GroupID:            request.GroupID,
		SessionHash:        request.SessionHash,
		StickyAccountID:    stickyAccountID,
		PreviousResponseID: request.PreviousResponseID,
		RequestedModel:     request.Model,
		RequiredTransport:  requiredTransport,
		ExcludedIDs:        request.ExcludedIDs,
	})
	return selection, buildOpenAISelectionDecision(request, selection, decision), err
}

func (k *GatewaySelectionKernel) Select(ctx context.Context, request SelectionRequest) (*AccountSelectionResult, SelectionDecision, error) {
	if k == nil || k.service == nil {
		return nil, SelectionDecision{}, errors.New("gateway selection kernel is not configured")
	}

	selection, err := k.service.SelectAccountWithLoadAwareness(
		ctx,
		request.GroupID,
		request.SessionHash,
		request.Model,
		request.ExcludedIDs,
		"",
	)
	return selection, buildGatewaySelectionDecision(request, selection), err
}

func buildOpenAISelectionDecision(
	request SelectionRequest,
	selection *AccountSelectionResult,
	openAIDecision OpenAIAccountScheduleDecision,
) SelectionDecision {
	decision := SelectionDecision{
		Provider: firstNonEmpty(strings.TrimSpace(request.Provider), PlatformOpenAI),
		Protocol: request.Protocol,
		Model:    strings.TrimSpace(request.Model),
		Layer:    openAIDecision.Layer,
		OpenAI:   openAIDecision,
	}
	if selection != nil && selection.Account != nil {
		decision.AccountID = selection.Account.ID
	}
	return decision
}

func buildGatewaySelectionDecision(request SelectionRequest, selection *AccountSelectionResult) SelectionDecision {
	decision := SelectionDecision{
		Provider: firstNonEmpty(strings.TrimSpace(request.Provider), "gateway"),
		Protocol: request.Protocol,
		Model:    strings.TrimSpace(request.Model),
		Layer:    "gateway_load_aware",
	}
	if selection != nil && selection.Account != nil {
		decision.AccountID = selection.Account.ID
	}
	return decision
}

func OpenAIUpstreamTransportToGatewayTransport(transport OpenAIUpstreamTransport) GatewayTransport {
	switch transport {
	case OpenAIUpstreamTransportHTTPSSE:
		return GatewayTransportHTTPSSE
	case OpenAIUpstreamTransportResponsesWebsocket:
		return GatewayTransportResponsesWebSocket
	case OpenAIUpstreamTransportResponsesWebsocketV2:
		return GatewayTransportResponsesWebSocketV2
	default:
		return GatewayTransportAny
	}
}

func gatewayTransportToOpenAIUpstreamTransport(transport GatewayTransport) OpenAIUpstreamTransport {
	switch transport {
	case GatewayTransportHTTPSSE:
		return OpenAIUpstreamTransportHTTPSSE
	case GatewayTransportResponsesWebSocket:
		return OpenAIUpstreamTransportResponsesWebsocket
	case GatewayTransportResponsesWebSocketV2:
		return OpenAIUpstreamTransportResponsesWebsocketV2
	default:
		return OpenAIUpstreamTransportAny
	}
}
