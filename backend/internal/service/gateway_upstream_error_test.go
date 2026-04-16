package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractUpstreamErrorMessage_UnwrapsNestedJSONString(t *testing.T) {
	body := []byte(`{"error":{"message":"{\"error\":{\"code\":\"invalid_request\",\"message\":\"nested failure\"}}"}}`)

	message := ExtractUpstreamErrorMessage(body)
	require.Equal(t, "nested failure", message)
}

func TestExtractUpstreamErrorCode_UnwrapsNestedJSONString(t *testing.T) {
	body := []byte(`{"error":{"message":"{\"error\":{\"code\":\"quota_exceeded\",\"message\":\"nested failure\"}}"}}`)

	code := extractUpstreamErrorCode(body)
	require.Equal(t, "quota_exceeded", code)
}

func TestExtractUpstreamErrorInfo_UnwrapsNestedStructuredFields(t *testing.T) {
	body := []byte(`{"error":{"message":"{\"error\":{\"type\":\"invalid_request_error\",\"code\":\"missing_required_parameter\",\"param\":\"instructions\",\"message\":\"nested failure\"}}（traceid: abc123）"}}`)

	info := extractUpstreamErrorInfo(body)
	require.Equal(t, "nested failure", info.Message)
	require.Equal(t, "invalid_request_error", info.Type)
	require.True(t, info.HasCode)
	require.Equal(t, "missing_required_parameter", info.Code)
	require.True(t, info.HasParam)
	require.Equal(t, "instructions", info.Param)
}

func TestExtractUpstreamErrorInfo_UsesDetailObjectWhenPresent(t *testing.T) {
	body := []byte(`{"detail":{"code":"deactivated_workspace","message":"workspace disabled"}}`)

	info := extractUpstreamErrorInfo(body)
	require.Equal(t, "workspace disabled", info.Message)
	require.True(t, info.HasCode)
	require.Equal(t, "deactivated_workspace", info.Code)
}

func TestGatewayServiceShouldFailoverOn400_RecognizesCompatibilityErrors(t *testing.T) {
	service := &GatewayService{}

	require.True(t, service.shouldFailoverOn400([]byte(`{"error":{"message":"requires beta feature anthropic-beta"}}`)))
	require.True(t, service.shouldFailoverOn400([]byte(`{"error":{"message":"Invalid signature in thinking block"}}`)))
	require.True(t, service.shouldFailoverOn400([]byte(`{"error":{"message":"tools payload invalid"}}`)))
	require.False(t, service.shouldFailoverOn400([]byte(`{"error":{"message":"ordinary validation error"}}`)))
}

func TestGatewayServiceIsThinkingBlockSignatureError_RecognizesEmptyContentVariant(t *testing.T) {
	service := &GatewayService{}

	require.True(t, service.isThinkingBlockSignatureError([]byte(`{"error":{"message":"messages: text content blocks must be non-empty"}}`)))
	require.False(t, service.isThinkingBlockSignatureError([]byte(`{"error":{"message":"validation failed"}}`)))
}
