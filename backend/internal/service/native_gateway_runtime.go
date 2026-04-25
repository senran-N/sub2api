package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type NativeGatewayForwardRequest struct {
	Provider        string
	Protocol        GatewayProtocol
	Account         *Account
	GinContext      *gin.Context
	Model           string
	Action          string
	Stream          bool
	Body            []byte
	Parsed          *ParsedRequest
	HasBoundSession bool
}

type NativeGatewayCountTokensForwardRequest struct {
	Account    *Account
	GinContext *gin.Context
	Parsed     *ParsedRequest
}

type NativeGatewayRuntime struct {
	forwardAnthropicMessages        func(context.Context, *gin.Context, *Account, *ParsedRequest) (*ForwardResult, error)
	forwardAnthropicChatCompletions func(context.Context, *gin.Context, *Account, []byte, *ParsedRequest) (*ForwardResult, error)
	forwardAnthropicResponses       func(context.Context, *gin.Context, *Account, []byte, *ParsedRequest) (*ForwardResult, error)
	forwardCountTokens              func(context.Context, *gin.Context, *Account, *ParsedRequest) error
	forwardGeminiMessages           func(context.Context, *gin.Context, *Account, []byte) (*ForwardResult, error)
	forwardGeminiNative             func(context.Context, *gin.Context, *Account, string, string, bool, []byte) (*ForwardResult, error)
	forwardAntigravityMessages      func(context.Context, *gin.Context, *Account, []byte, bool) (*ForwardResult, error)
	forwardAntigravityGemini        func(context.Context, *gin.Context, *Account, string, string, bool, []byte, bool) (*ForwardResult, error)
}

type nativeGatewayRuntimeHooks struct {
	forwardAnthropicMessages        func(context.Context, *gin.Context, *Account, *ParsedRequest) (*ForwardResult, error)
	forwardAnthropicChatCompletions func(context.Context, *gin.Context, *Account, []byte, *ParsedRequest) (*ForwardResult, error)
	forwardAnthropicResponses       func(context.Context, *gin.Context, *Account, []byte, *ParsedRequest) (*ForwardResult, error)
	forwardCountTokens              func(context.Context, *gin.Context, *Account, *ParsedRequest) error
	forwardGeminiMessages           func(context.Context, *gin.Context, *Account, []byte) (*ForwardResult, error)
	forwardGeminiNative             func(context.Context, *gin.Context, *Account, string, string, bool, []byte) (*ForwardResult, error)
	forwardAntigravityMessages      func(context.Context, *gin.Context, *Account, []byte, bool) (*ForwardResult, error)
	forwardAntigravityGemini        func(context.Context, *gin.Context, *Account, string, string, bool, []byte, bool) (*ForwardResult, error)
}

func NewNativeGatewayRuntime(
	gatewayService *GatewayService,
	geminiCompatService *GeminiMessagesCompatService,
	antigravityGatewayService *AntigravityGatewayService,
) *NativeGatewayRuntime {
	var hooks nativeGatewayRuntimeHooks
	if gatewayService != nil {
		hooks.forwardAnthropicMessages = gatewayService.Forward
		hooks.forwardAnthropicChatCompletions = gatewayService.ForwardAsChatCompletions
		hooks.forwardAnthropicResponses = gatewayService.ForwardAsResponses
		hooks.forwardCountTokens = gatewayService.ForwardCountTokens
	}
	if geminiCompatService != nil {
		hooks.forwardGeminiMessages = geminiCompatService.Forward
		hooks.forwardGeminiNative = geminiCompatService.ForwardNative
	}
	if antigravityGatewayService != nil {
		hooks.forwardAntigravityMessages = antigravityGatewayService.Forward
		hooks.forwardAntigravityGemini = antigravityGatewayService.ForwardGemini
	}
	return newNativeGatewayRuntimeWithHooks(hooks)
}

func newNativeGatewayRuntimeWithHooks(hooks nativeGatewayRuntimeHooks) *NativeGatewayRuntime {
	return &NativeGatewayRuntime{
		forwardAnthropicMessages:        hooks.forwardAnthropicMessages,
		forwardAnthropicChatCompletions: hooks.forwardAnthropicChatCompletions,
		forwardAnthropicResponses:       hooks.forwardAnthropicResponses,
		forwardCountTokens:              hooks.forwardCountTokens,
		forwardGeminiMessages:           hooks.forwardGeminiMessages,
		forwardGeminiNative:             hooks.forwardGeminiNative,
		forwardAntigravityMessages:      hooks.forwardAntigravityMessages,
		forwardAntigravityGemini:        hooks.forwardAntigravityGemini,
	}
}

func (r *NativeGatewayRuntime) Forward(ctx context.Context, req NativeGatewayForwardRequest) (*ForwardResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if req.Account == nil {
		return nil, errors.New("native gateway runtime: account is required")
	}
	if req.GinContext == nil {
		return nil, errors.New("native gateway runtime: gin context is required")
	}
	if req.Protocol == "" {
		return nil, errors.New("native gateway runtime: protocol is required")
	}

	switch req.Protocol {
	case GatewayProtocolMessages:
		if req.Provider == PlatformGemini {
			return r.forwardGeminiMessagesProtocol(ctx, req)
		}
		return r.forwardAnthropicMessagesProtocol(ctx, req)
	case GatewayProtocolChatCompletions:
		return r.forwardAnthropicChatCompletionsProtocol(ctx, req)
	case GatewayProtocolResponses:
		return r.forwardAnthropicResponsesProtocol(ctx, req)
	case GatewayProtocolPassthrough:
		if req.Provider == PlatformGemini {
			return r.forwardGeminiNativeProtocol(ctx, req)
		}
	}

	return nil, fmt.Errorf("native gateway runtime: unsupported provider/protocol %s/%s", req.Provider, req.Protocol)
}

func (r *NativeGatewayRuntime) ForwardCountTokens(ctx context.Context, req NativeGatewayCountTokensForwardRequest) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if req.Account == nil {
		return errors.New("native gateway runtime: account is required")
	}
	if req.GinContext == nil {
		return errors.New("native gateway runtime: gin context is required")
	}
	if req.Parsed == nil {
		return errors.New("native gateway runtime: parsed request is required")
	}
	if r == nil || r.forwardCountTokens == nil {
		return errors.New("native gateway runtime: count_tokens forwarder is not configured")
	}
	return r.forwardCountTokens(ctx, req.GinContext, req.Account, req.Parsed)
}

func (r *NativeGatewayRuntime) forwardAnthropicMessagesProtocol(ctx context.Context, req NativeGatewayForwardRequest) (*ForwardResult, error) {
	if req.Account.Platform == PlatformAntigravity && req.Account.Type != AccountTypeAPIKey {
		if r == nil || r.forwardAntigravityMessages == nil {
			return nil, errors.New("native gateway runtime: antigravity messages forwarder is not configured")
		}
		return r.forwardAntigravityMessages(ctx, req.GinContext, req.Account, req.Body, req.HasBoundSession)
	}
	if req.Parsed == nil {
		return nil, errors.New("native gateway runtime: parsed request is required")
	}
	if r == nil || r.forwardAnthropicMessages == nil {
		return nil, errors.New("native gateway runtime: anthropic messages forwarder is not configured")
	}
	return r.forwardAnthropicMessages(ctx, req.GinContext, req.Account, req.Parsed)
}

func (r *NativeGatewayRuntime) forwardAnthropicChatCompletionsProtocol(ctx context.Context, req NativeGatewayForwardRequest) (*ForwardResult, error) {
	if req.Parsed == nil {
		return nil, errors.New("native gateway runtime: parsed request is required")
	}
	if r == nil || r.forwardAnthropicChatCompletions == nil {
		return nil, errors.New("native gateway runtime: anthropic chat completions forwarder is not configured")
	}
	return r.forwardAnthropicChatCompletions(ctx, req.GinContext, req.Account, req.Body, req.Parsed)
}

func (r *NativeGatewayRuntime) forwardAnthropicResponsesProtocol(ctx context.Context, req NativeGatewayForwardRequest) (*ForwardResult, error) {
	if req.Parsed == nil {
		return nil, errors.New("native gateway runtime: parsed request is required")
	}
	if r == nil || r.forwardAnthropicResponses == nil {
		return nil, errors.New("native gateway runtime: anthropic responses forwarder is not configured")
	}
	return r.forwardAnthropicResponses(ctx, req.GinContext, req.Account, req.Body, req.Parsed)
}

func (r *NativeGatewayRuntime) forwardGeminiMessagesProtocol(ctx context.Context, req NativeGatewayForwardRequest) (*ForwardResult, error) {
	if req.Account.Platform == PlatformAntigravity {
		return r.forwardAntigravityGeminiProtocol(ctx, req)
	}
	if r == nil || r.forwardGeminiMessages == nil {
		return nil, errors.New("native gateway runtime: gemini messages forwarder is not configured")
	}
	return r.forwardGeminiMessages(ctx, req.GinContext, req.Account, req.Body)
}

func (r *NativeGatewayRuntime) forwardGeminiNativeProtocol(ctx context.Context, req NativeGatewayForwardRequest) (*ForwardResult, error) {
	if req.Account.Platform == PlatformAntigravity && req.Account.Type != AccountTypeAPIKey {
		return r.forwardAntigravityGeminiProtocol(ctx, req)
	}
	if r == nil || r.forwardGeminiNative == nil {
		return nil, errors.New("native gateway runtime: gemini native forwarder is not configured")
	}
	return r.forwardGeminiNative(ctx, req.GinContext, req.Account, req.Model, req.Action, req.Stream, req.Body)
}

func (r *NativeGatewayRuntime) forwardAntigravityGeminiProtocol(ctx context.Context, req NativeGatewayForwardRequest) (*ForwardResult, error) {
	if r == nil || r.forwardAntigravityGemini == nil {
		return nil, errors.New("native gateway runtime: antigravity gemini forwarder is not configured")
	}
	return r.forwardAntigravityGemini(ctx, req.GinContext, req.Account, req.Model, req.Action, req.Stream, req.Body, req.HasBoundSession)
}
