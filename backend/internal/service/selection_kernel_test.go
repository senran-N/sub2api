package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIUpstreamTransportGatewayMapping(t *testing.T) {
	require.Equal(t, GatewayTransportAny, OpenAIUpstreamTransportToGatewayTransport(OpenAIUpstreamTransportAny))
	require.Equal(t, GatewayTransportHTTPSSE, OpenAIUpstreamTransportToGatewayTransport(OpenAIUpstreamTransportHTTPSSE))
	require.Equal(t, GatewayTransportResponsesWebSocket, OpenAIUpstreamTransportToGatewayTransport(OpenAIUpstreamTransportResponsesWebsocket))
	require.Equal(t, GatewayTransportResponsesWebSocketV2, OpenAIUpstreamTransportToGatewayTransport(OpenAIUpstreamTransportResponsesWebsocketV2))

	require.Equal(t, OpenAIUpstreamTransportAny, gatewayTransportToOpenAIUpstreamTransport(GatewayTransportAny))
	require.Equal(t, OpenAIUpstreamTransportHTTPSSE, gatewayTransportToOpenAIUpstreamTransport(GatewayTransportHTTPSSE))
	require.Equal(t, OpenAIUpstreamTransportResponsesWebsocket, gatewayTransportToOpenAIUpstreamTransport(GatewayTransportResponsesWebSocket))
	require.Equal(t, OpenAIUpstreamTransportResponsesWebsocketV2, gatewayTransportToOpenAIUpstreamTransport(GatewayTransportResponsesWebSocketV2))
}

func TestOpenAISelectionKernelNilService(t *testing.T) {
	var kernel *OpenAISelectionKernel
	selection, decision, err := kernel.Select(context.Background(), SelectionRequest{})

	require.Nil(t, selection)
	require.Empty(t, decision)
	require.ErrorContains(t, err, "openai selection kernel is not configured")
}

func TestGatewaySelectionKernelNilService(t *testing.T) {
	var kernel *GatewaySelectionKernel
	selection, decision, err := kernel.Select(context.Background(), SelectionRequest{})

	require.Nil(t, selection)
	require.Empty(t, decision)
	require.ErrorContains(t, err, "gateway selection kernel is not configured")
}

func TestBuildGatewaySelectionDecision(t *testing.T) {
	selection := &AccountSelectionResult{Account: &Account{ID: 42}}

	decision := buildGatewaySelectionDecision(SelectionRequest{
		Provider: PlatformGemini,
		Protocol: GatewayProtocolMessages,
		Model:    "gemini-3-pro",
	}, selection)

	require.Equal(t, PlatformGemini, decision.Provider)
	require.Equal(t, GatewayProtocolMessages, decision.Protocol)
	require.Equal(t, "gemini-3-pro", decision.Model)
	require.Equal(t, int64(42), decision.AccountID)
	require.Equal(t, "gateway_load_aware", decision.Layer)
}
