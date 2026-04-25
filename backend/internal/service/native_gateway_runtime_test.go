package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type nativeGatewayRuntimeCall struct {
	name    string
	account *Account
	model   string
	action  string
	stream  bool
	sticky  bool
	body    string
	parsed  *ParsedRequest
}

func TestNativeGatewayRuntimeForwardRoutesAnthropicMessages(t *testing.T) {
	var calls []nativeGatewayRuntimeCall
	runtime := newNativeGatewayRuntimeWithHooks(nativeGatewayRuntimeHooks{
		forwardAnthropicMessages: func(_ context.Context, _ *gin.Context, account *Account, parsed *ParsedRequest) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "anthropic_messages", account: account, parsed: parsed})
			return &ForwardResult{RequestID: "anthropic_messages"}, nil
		},
		forwardAntigravityMessages: func(_ context.Context, _ *gin.Context, account *Account, body []byte, sticky bool) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "antigravity_messages", account: account, body: string(body), sticky: sticky})
			return &ForwardResult{RequestID: "antigravity_messages"}, nil
		},
	})

	parsed := &ParsedRequest{Model: "claude-sonnet"}
	result, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:   PlatformAnthropic,
		Protocol:   GatewayProtocolMessages,
		Account:    &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeOAuth},
		GinContext: &gin.Context{},
		Parsed:     parsed,
		Body:       []byte(`{"model":"claude-sonnet"}`),
	})

	require.NoError(t, err)
	require.Equal(t, "anthropic_messages", result.RequestID)
	require.Len(t, calls, 1)
	require.Equal(t, "anthropic_messages", calls[0].name)
	require.Equal(t, parsed, calls[0].parsed)
}

func TestNativeGatewayRuntimeForwardRoutesAntigravityMessages(t *testing.T) {
	var calls []nativeGatewayRuntimeCall
	runtime := newNativeGatewayRuntimeWithHooks(nativeGatewayRuntimeHooks{
		forwardAnthropicMessages: func(context.Context, *gin.Context, *Account, *ParsedRequest) (*ForwardResult, error) {
			t.Fatal("anthropic messages forwarder must not be used for antigravity non-apikey accounts")
			return nil, nil
		},
		forwardAntigravityMessages: func(_ context.Context, _ *gin.Context, account *Account, body []byte, sticky bool) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "antigravity_messages", account: account, body: string(body), sticky: sticky})
			return &ForwardResult{RequestID: "antigravity_messages"}, nil
		},
	})

	result, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:        PlatformAntigravity,
		Protocol:        GatewayProtocolMessages,
		Account:         &Account{ID: 2, Platform: PlatformAntigravity, Type: AccountTypeOAuth},
		GinContext:      &gin.Context{},
		Body:            []byte(`{"model":"claude-sonnet"}`),
		HasBoundSession: true,
	})

	require.NoError(t, err)
	require.Equal(t, "antigravity_messages", result.RequestID)
	require.Len(t, calls, 1)
	require.True(t, calls[0].sticky)
	require.Equal(t, `{"model":"claude-sonnet"}`, calls[0].body)
}

func TestNativeGatewayRuntimeForwardRoutesOpenAICompatibleProtocols(t *testing.T) {
	var calls []nativeGatewayRuntimeCall
	runtime := newNativeGatewayRuntimeWithHooks(nativeGatewayRuntimeHooks{
		forwardAnthropicChatCompletions: func(_ context.Context, _ *gin.Context, account *Account, body []byte, parsed *ParsedRequest) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "chat_completions", account: account, body: string(body), parsed: parsed})
			return &ForwardResult{RequestID: "chat_completions"}, nil
		},
		forwardAnthropicResponses: func(_ context.Context, _ *gin.Context, account *Account, body []byte, parsed *ParsedRequest) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "responses", account: account, body: string(body), parsed: parsed})
			return &ForwardResult{RequestID: "responses"}, nil
		},
	})
	account := &Account{ID: 3, Platform: PlatformAnthropic, Type: AccountTypeOAuth}
	parsed := &ParsedRequest{Model: "claude-sonnet"}

	chatResult, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:   PlatformAnthropic,
		Protocol:   GatewayProtocolChatCompletions,
		Account:    account,
		GinContext: &gin.Context{},
		Body:       []byte(`{"model":"chat"}`),
		Parsed:     parsed,
	})
	require.NoError(t, err)
	require.Equal(t, "chat_completions", chatResult.RequestID)

	responsesResult, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:   PlatformAnthropic,
		Protocol:   GatewayProtocolResponses,
		Account:    account,
		GinContext: &gin.Context{},
		Body:       []byte(`{"model":"responses"}`),
		Parsed:     parsed,
	})
	require.NoError(t, err)
	require.Equal(t, "responses", responsesResult.RequestID)

	require.Len(t, calls, 2)
	require.Equal(t, "chat_completions", calls[0].name)
	require.Equal(t, `{"model":"chat"}`, calls[0].body)
	require.Equal(t, "responses", calls[1].name)
	require.Equal(t, `{"model":"responses"}`, calls[1].body)
}

func TestNativeGatewayRuntimeForwardRoutesCountTokens(t *testing.T) {
	var calls []nativeGatewayRuntimeCall
	expectedErr := errors.New("upstream count_tokens failed")
	runtime := newNativeGatewayRuntimeWithHooks(nativeGatewayRuntimeHooks{
		forwardCountTokens: func(_ context.Context, _ *gin.Context, account *Account, parsed *ParsedRequest) error {
			calls = append(calls, nativeGatewayRuntimeCall{name: "count_tokens", account: account, parsed: parsed})
			return expectedErr
		},
	})
	account := &Account{ID: 4, Platform: PlatformAnthropic, Type: AccountTypeOAuth}
	parsed := &ParsedRequest{Model: "claude-sonnet"}

	err := runtime.ForwardCountTokens(context.Background(), NativeGatewayCountTokensForwardRequest{
		Account:    account,
		GinContext: &gin.Context{},
		Parsed:     parsed,
	})

	require.ErrorIs(t, err, expectedErr)
	require.Len(t, calls, 1)
	require.Equal(t, "count_tokens", calls[0].name)
	require.Equal(t, account, calls[0].account)
	require.Equal(t, parsed, calls[0].parsed)
}

func TestNativeGatewayRuntimeForwardRoutesGeminiProtocols(t *testing.T) {
	var calls []nativeGatewayRuntimeCall
	runtime := newNativeGatewayRuntimeWithHooks(nativeGatewayRuntimeHooks{
		forwardGeminiMessages: func(_ context.Context, _ *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "gemini_messages", account: account, body: string(body)})
			return &ForwardResult{RequestID: "gemini_messages"}, nil
		},
		forwardGeminiNative: func(_ context.Context, _ *gin.Context, account *Account, model string, action string, stream bool, body []byte) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "gemini_native", account: account, model: model, action: action, stream: stream, body: string(body)})
			return &ForwardResult{RequestID: "gemini_native"}, nil
		},
		forwardAntigravityGemini: func(_ context.Context, _ *gin.Context, account *Account, model string, action string, stream bool, body []byte, sticky bool) (*ForwardResult, error) {
			calls = append(calls, nativeGatewayRuntimeCall{name: "antigravity_gemini", account: account, model: model, action: action, stream: stream, body: string(body), sticky: sticky})
			return &ForwardResult{RequestID: "antigravity_gemini"}, nil
		},
	})

	messagesResult, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:   PlatformGemini,
		Protocol:   GatewayProtocolMessages,
		Account:    &Account{ID: 4, Platform: PlatformGemini, Type: AccountTypeOAuth},
		GinContext: &gin.Context{},
		Body:       []byte(`{"model":"claude-as-gemini"}`),
	})
	require.NoError(t, err)
	require.Equal(t, "gemini_messages", messagesResult.RequestID)

	nativeResult, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:   PlatformGemini,
		Protocol:   GatewayProtocolPassthrough,
		Account:    &Account{ID: 5, Platform: PlatformGemini, Type: AccountTypeOAuth},
		GinContext: &gin.Context{},
		Model:      "gemini-2.5-pro",
		Action:     "streamGenerateContent",
		Stream:     true,
		Body:       []byte(`{"contents":[]}`),
	})
	require.NoError(t, err)
	require.Equal(t, "gemini_native", nativeResult.RequestID)

	antigravityResult, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:        PlatformGemini,
		Protocol:        GatewayProtocolPassthrough,
		Account:         &Account{ID: 6, Platform: PlatformAntigravity, Type: AccountTypeOAuth},
		GinContext:      &gin.Context{},
		Model:           "gemini-2.5-pro",
		Action:          "generateContent",
		Body:            []byte(`{"contents":[]}`),
		HasBoundSession: true,
	})
	require.NoError(t, err)
	require.Equal(t, "antigravity_gemini", antigravityResult.RequestID)

	require.Len(t, calls, 3)
	require.Equal(t, "gemini_messages", calls[0].name)
	require.Equal(t, "gemini_native", calls[1].name)
	require.True(t, calls[1].stream)
	require.Equal(t, "streamGenerateContent", calls[1].action)
	require.Equal(t, "antigravity_gemini", calls[2].name)
	require.True(t, calls[2].sticky)
}

func TestNativeGatewayRuntimeForwardValidationAndMissingDependencies(t *testing.T) {
	var runtime *NativeGatewayRuntime

	_, err := runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Protocol:   GatewayProtocolMessages,
		GinContext: &gin.Context{},
	})
	require.ErrorContains(t, err, "account is required")

	_, err = runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Account:    &Account{ID: 7, Platform: PlatformAnthropic},
		GinContext: &gin.Context{},
	})
	require.ErrorContains(t, err, "protocol is required")

	_, err = runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Protocol:   GatewayProtocolMessages,
		Account:    &Account{ID: 8, Platform: PlatformAnthropic},
		GinContext: &gin.Context{},
		Parsed:     &ParsedRequest{Model: "claude-sonnet"},
	})
	require.ErrorContains(t, err, "anthropic messages forwarder is not configured")

	_, err = runtime.Forward(context.Background(), NativeGatewayForwardRequest{
		Provider:   PlatformOpenAI,
		Protocol:   GatewayProtocolPassthrough,
		Account:    &Account{ID: 9, Platform: PlatformOpenAI},
		GinContext: &gin.Context{},
	})
	require.ErrorContains(t, err, "unsupported provider/protocol")

	err = runtime.ForwardCountTokens(context.Background(), NativeGatewayCountTokensForwardRequest{
		GinContext: &gin.Context{},
		Parsed:     &ParsedRequest{Model: "claude-sonnet"},
	})
	require.ErrorContains(t, err, "account is required")

	err = runtime.ForwardCountTokens(context.Background(), NativeGatewayCountTokensForwardRequest{
		Account: &Account{ID: 10, Platform: PlatformAnthropic},
		Parsed:  &ParsedRequest{Model: "claude-sonnet"},
	})
	require.ErrorContains(t, err, "gin context is required")

	err = runtime.ForwardCountTokens(context.Background(), NativeGatewayCountTokensForwardRequest{
		Account:    &Account{ID: 11, Platform: PlatformAnthropic},
		GinContext: &gin.Context{},
	})
	require.ErrorContains(t, err, "parsed request is required")

	err = runtime.ForwardCountTokens(context.Background(), NativeGatewayCountTokensForwardRequest{
		Account:    &Account{ID: 12, Platform: PlatformAnthropic},
		GinContext: &gin.Context{},
		Parsed:     &ParsedRequest{Model: "claude-sonnet"},
	})
	require.ErrorContains(t, err, "count_tokens forwarder is not configured")
}
