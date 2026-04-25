package service

import (
	"context"
	"net/http"
)

type GatewayProtocol string

const (
	GatewayProtocolResponses       GatewayProtocol = "responses"
	GatewayProtocolChatCompletions GatewayProtocol = "chat_completions"
	GatewayProtocolMessages        GatewayProtocol = "messages"
	GatewayProtocolCountTokens     GatewayProtocol = "count_tokens"
	GatewayProtocolPassthrough     GatewayProtocol = "passthrough"
)

type GatewayTransport string

const (
	GatewayTransportAny                  GatewayTransport = ""
	GatewayTransportHTTPSSE              GatewayTransport = "http_sse"
	GatewayTransportWebSocket            GatewayTransport = "websocket"
	GatewayTransportResponsesWebSocket   GatewayTransport = "responses_websockets"
	GatewayTransportResponsesWebSocketV2 GatewayTransport = "responses_websockets_v2"
	GatewayTransportSession              GatewayTransport = "session"
)

// GatewayRequest is the provider-neutral request envelope used by gateway
// runtime kernels before a provider facade applies protocol-specific details.
type GatewayRequest struct {
	Provider           string
	Protocol           GatewayProtocol
	Model              string
	GroupID            *int64
	SessionHash        string
	PreviousResponseID string
	Transport          GatewayTransport
	Stream             bool
	Body               []byte
	Header             http.Header
	Metadata           map[string]string
}

// GatewayResponse is the provider-neutral execution result emitted by gateway
// runtimes before a protocol renderer writes the final HTTP response.
type GatewayResponse struct {
	Result     *OpenAIForwardResult
	StatusCode int
	Header     http.Header
	Body       []byte
	Stream     bool
}

// GatewayRuntime is the common provider runtime contract for compatible
// gateway execution. Provider-specific runtimes can wrap this with narrower
// HTTP-facing adapters while sharing selection and feedback kernels.
type GatewayRuntime interface {
	ExecuteGatewayRequest(ctx context.Context, request GatewayRequest) (GatewayResponse, error)
}
