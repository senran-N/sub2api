//go:build unit

package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessAnthropicStreamingEvent_ErrorEventReturnsUpstreamStreamError(t *testing.T) {
	svc := newMinimalGatewayService()

	_, err := svc.processAnthropicStreamingEvent(
		http.StatusOK,
		&Account{},
		"model-a",
		"model-a",
		[]string{
			"event: error",
			`data: {"type":"error","error":{"type":"rate_limit_error","status_code":429,"message":"rate limited"}}`,
		},
	)

	require.Error(t, err)
	var streamErr *upstreamStreamEventError
	require.True(t, errors.As(err, &streamErr))
	require.Equal(t, http.StatusTooManyRequests, streamErr.StatusCode())
	require.JSONEq(t, `{"type":"error","error":{"type":"rate_limit_error","status_code":429,"message":"rate limited"}}`, string(streamErr.ResponseBody()))
}

func TestProcessAnthropicStreamingEvent_DoneMarksTerminal(t *testing.T) {
	svc := newMinimalGatewayService()

	result, err := svc.processAnthropicStreamingEvent(
		http.StatusOK,
		&Account{},
		"model-a",
		"model-a",
		[]string{"data: [DONE]"},
	)

	require.NoError(t, err)
	require.True(t, result.terminal)
	require.Equal(t, "[DONE]", result.data)
	require.Equal(t, []string{"data: [DONE]\n\n"}, result.outputBlocks)
}

func TestProcessAnthropicStreamingEvent_RewritesMappedModelAndExtractsUsagePatch(t *testing.T) {
	svc := newMinimalGatewayService()

	result, err := svc.processAnthropicStreamingEvent(
		http.StatusOK,
		&Account{},
		"original-model",
		"mapped-model",
		[]string{`data: {"type":"message_start","message":{"model":"mapped-model","usage":{"input_tokens":11}}}`},
	)

	require.NoError(t, err)
	require.False(t, result.terminal)
	require.Contains(t, result.outputBlocks[0], `"model":"original-model"`)
	require.NotNil(t, result.usagePatch)
	require.True(t, result.usagePatch.hasInputTokens)
	require.Equal(t, 11, result.usagePatch.inputTokens)
}

func TestProcessAnthropicStreamingEvent_InvalidJSONPassesThroughRawBlock(t *testing.T) {
	svc := newMinimalGatewayService()

	result, err := svc.processAnthropicStreamingEvent(
		http.StatusOK,
		&Account{},
		"model-a",
		"model-b",
		[]string{"event: message_start", `data: {"broken-json"`},
	)

	require.NoError(t, err)
	require.False(t, result.terminal)
	require.Nil(t, result.usagePatch)
	require.Equal(t, `{"broken-json"`, result.data)
	require.Equal(t, []string{"event: message_start\ndata: {\"broken-json\"\n\n"}, result.outputBlocks)
}
