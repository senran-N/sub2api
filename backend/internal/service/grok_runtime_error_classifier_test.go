//go:build unit

package service

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClassifyGrokRuntimeError_ModelTierFailureUsesModelScope(t *testing.T) {
	classification := classifyGrokRuntimeError(GrokRuntimeFeedbackInput{
		RequestedModel: "grok-imagine-video",
		StatusCode:     http.StatusForbidden,
		Err: &UpstreamFailoverError{
			StatusCode:   http.StatusForbidden,
			ResponseBody: []byte(`{"error":{"code":"tier_required","message":"video tier required"}}`),
		},
	})

	require.Equal(t, http.StatusForbidden, classification.StatusCode)
	require.Equal(t, grokRuntimeErrorClassModelUnsupported, classification.Class)
	require.Equal(t, grokRuntimePenaltyScopeModel, classification.Scope)
	require.False(t, classification.Retryable)
	require.Equal(t, 45*time.Minute, classification.Cooldown)
}

func TestClassifyGrokRuntimeError_RateLimitUsesAccountCooldown(t *testing.T) {
	classification := classifyGrokRuntimeError(GrokRuntimeFeedbackInput{
		StatusCode: http.StatusTooManyRequests,
		Err:        errors.New("rate limited"),
	})

	require.Equal(t, grokRuntimeErrorClassRateLimited, classification.Class)
	require.Equal(t, grokRuntimePenaltyScopeAccount, classification.Scope)
	require.True(t, classification.Retryable)
	require.Equal(t, 10*time.Minute, classification.Cooldown)
}

func TestClassifyGrokRuntimeError_TransportFailureDoesNotPretendToBeAuth(t *testing.T) {
	classification := classifyGrokRuntimeError(GrokRuntimeFeedbackInput{
		Err: errors.New("dial tcp timeout"),
	})

	require.Equal(t, grokRuntimeErrorClassTransport, classification.Class)
	require.Equal(t, grokRuntimePenaltyScopeAccount, classification.Scope)
	require.True(t, classification.Retryable)
	require.Equal(t, 2*time.Minute, classification.Cooldown)
}
