package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCodexRecoveryPolicy_PreviousResponseNotFound(t *testing.T) {
	policy := CodexRecoveryPolicy{}

	t.Run("applies_by_dropping_previous_response_id", func(t *testing.T) {
		reqBody := map[string]any{
			"previous_response_id": "resp_prev_anchor",
			"input": []any{
				map[string]any{"type": "input_text", "text": "hello"},
			},
		}

		decision := policy.Apply(reqBody, CodexRecoveryPolicyInput{
			Reason:    codexRecoveryReasonPreviousResponseNotFound,
			Transport: OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.True(t, decision.Applied)
		require.Equal(t, codexRecoveryActionDropPreviousResponseID, decision.Action)
		require.True(t, decision.DroppedPreviousResponseID)
		require.Equal(t, "resp_prev_anchor", decision.PreviousResponseID)
		require.Equal(t, OpenAIPreviousResponseIDKindResponseID, decision.PreviousResponseIDKind)
		require.False(t, decision.HasFunctionCallOutput)
		require.Empty(t, openAIWSPayloadString(reqBody, "previous_response_id"))
	})

	t.Run("skips_function_call_output_chain", func(t *testing.T) {
		reqBody := map[string]any{
			"previous_response_id": "resp_prev_tool",
			"input": []any{
				map[string]any{"type": "function_call_output", "call_id": "call_1", "output": "ok"},
			},
		}

		decision := policy.Apply(reqBody, CodexRecoveryPolicyInput{
			Reason:    codexRecoveryReasonPreviousResponseNotFound,
			Transport: OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.False(t, decision.Applied)
		require.Equal(t, "has_function_call_output", decision.SkipReason)
		require.True(t, decision.HasFunctionCallOutput)
		require.Equal(t, "resp_prev_tool", openAIWSPayloadString(reqBody, "previous_response_id"))
	})
}

func TestCodexRecoveryPolicy_InvalidEncryptedContent(t *testing.T) {
	policy := CodexRecoveryPolicy{}

	t.Run("ws_trims_reasoning_and_drops_previous_response_id", func(t *testing.T) {
		reqBody := map[string]any{
			"previous_response_id": "resp_prev_encrypted",
			"input": []any{
				map[string]any{"type": "reasoning", "encrypted_content": "gAAA", "summary": []any{map[string]any{"type": "summary_text", "text": "keep me"}}},
				map[string]any{"type": "input_text", "text": "hello"},
			},
		}

		decision := policy.Apply(reqBody, CodexRecoveryPolicyInput{
			Reason:    codexRecoveryReasonInvalidEncryptedContent,
			Transport: OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.True(t, decision.Applied)
		require.Equal(t, codexRecoveryActionTrimEncryptedReasoning, decision.Action)
		require.True(t, decision.TrimmedEncryptedReasoning)
		require.True(t, decision.DroppedPreviousResponseID)
		require.Empty(t, openAIWSPayloadString(reqBody, "previous_response_id"))

		input, ok := reqBody["input"].([]any)
		require.True(t, ok)
		require.Len(t, input, 2)
		reasoning, ok := input[0].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "reasoning", reasoning["type"])
		require.NotContains(t, reasoning, "encrypted_content")
		require.Contains(t, reasoning, "summary")
	})

	t.Run("http_trims_reasoning_but_keeps_previous_response_id", func(t *testing.T) {
		reqBody := map[string]any{
			"previous_response_id": "resp_prev_http",
			"input": []any{
				map[string]any{"type": "reasoning", "encrypted_content": "gBBB"},
				map[string]any{"type": "input_text", "text": "hello"},
			},
		}

		decision := policy.Apply(reqBody, CodexRecoveryPolicyInput{
			Reason:    codexRecoveryReasonInvalidEncryptedContent,
			Transport: OpenAIUpstreamTransportHTTPSSE,
		})

		require.True(t, decision.Applied)
		require.True(t, decision.TrimmedEncryptedReasoning)
		require.False(t, decision.DroppedPreviousResponseID)
		require.Equal(t, "resp_prev_http", openAIWSPayloadString(reqBody, "previous_response_id"))
	})
}

func TestCodexRecoveryPolicy_TransportFailure(t *testing.T) {
	policy := CodexRecoveryPolicy{}

	t.Run("marks_ws_transport_cooldown_for_retryable_failure", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			AccountID:     41,
			FailureReason: "prewarm_read_event",
			Reason:        codexRecoveryReasonTransportFailure,
			Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.True(t, decision.Applied)
		require.Equal(t, codexRecoveryActionMarkTransportCooldown, decision.Action)
		require.True(t, decision.MarkedTransportCooldown)
		require.Equal(t, "read_event", decision.FailureReason)
	})

	t.Run("skips_non_transport_failure_reason", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			AccountID:     41,
			FailureReason: "upstream_rate_limited",
			Reason:        codexRecoveryReasonTransportFailure,
			Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.False(t, decision.Applied)
		require.Equal(t, "unsupported_failure_reason", decision.SkipReason)
		require.False(t, decision.MarkedTransportCooldown)
	})
}

func TestCodexRecoveryPolicy_Failover(t *testing.T) {
	policy := CodexRecoveryPolicy{}

	t.Run("exhausts_rate_limited_failover", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			FailureReason: "upstream_rate_limited",
			Reason:        codexRecoveryReasonFailover,
			StatusCode:    http.StatusTooManyRequests,
			Transport:     OpenAIUpstreamTransportHTTPSSE,
		})

		require.True(t, decision.Applied)
		require.True(t, decision.ExhaustFailover)
		require.Equal(t, codexRecoveryActionExhaustFailover, decision.Action)
		require.False(t, decision.SwitchAccount)
	})

	t.Run("switches_account_on_upstream_5xx", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			AccountID:     44,
			FailureReason: "upstream_5xx",
			Reason:        codexRecoveryReasonFailover,
			StatusCode:    http.StatusBadGateway,
			Transport:     OpenAIUpstreamTransportHTTPSSE,
		})

		require.True(t, decision.Applied)
		require.False(t, decision.ExhaustFailover)
		require.True(t, decision.SwitchAccount)
		require.Equal(t, codexRecoveryActionSwitchAccount, decision.Action)
	})
}

func TestOpenAIGatewayService_ApplyCodexTransportCooldownRecovery(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.FallbackCooldownSeconds = 30
	svc := &OpenAIGatewayService{cfg: cfg}

	decision := svc.applyCodexTransportCooldownRecovery(52, "read_event", OpenAIUpstreamTransportResponsesWebsocketV2)

	require.True(t, decision.Applied)
	require.True(t, decision.MarkedTransportCooldown)
	require.True(t, svc.isOpenAIWSFallbackCooling(52))
}

func TestOpenAIGatewayService_RecordCodexRecoveryAccountSwitch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	c.Set("openai_ws_transport_decision", string(OpenAIUpstreamTransportResponsesWebsocketV2))

	svc := &OpenAIGatewayService{}
	account := &Account{ID: 61}
	failoverErr := &UpstreamFailoverError{StatusCode: http.StatusBadGateway}

	decision := svc.RecordCodexRecoveryAccountSwitch(c, account, failoverErr)

	require.True(t, decision.Applied)
	require.Equal(t, codexRecoveryActionSwitchAccount, decision.Action)
	require.True(t, decision.SwitchAccount)
	require.Equal(t, OpenAIUpstreamTransportResponsesWebsocketV2, decision.Transport)
	require.Equal(t, http.StatusBadGateway, decision.StatusCode)
}

func TestSnapshotOpenAICodexCompatibilityMetrics_RecoveryCounters(t *testing.T) {
	before := SnapshotOpenAICodexCompatibilityMetrics()
	policy := CodexRecoveryPolicy{}

	policy.Apply(nil, CodexRecoveryPolicyInput{
		FailureReason: "upstream_rate_limited",
		Reason:        codexRecoveryReasonFailover,
		StatusCode:    http.StatusTooManyRequests,
		Transport:     OpenAIUpstreamTransportHTTPSSE,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		FailureReason: "upstream_5xx",
		Reason:        codexRecoveryReasonFailover,
		StatusCode:    http.StatusBadGateway,
		Transport:     OpenAIUpstreamTransportHTTPSSE,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:     71,
		FailureReason: "read_event",
		Reason:        codexRecoveryReasonTransportFailure,
		Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		Reason:    codexRecoveryReasonTransportFailure,
		Transport: OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:  72,
		Reason:     codexRecoveryReasonAccountSwitch,
		StatusCode: http.StatusBadGateway,
		Transport:  OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		Reason:     codexRecoveryReasonAccountSwitch,
		StatusCode: http.StatusBadGateway,
		Transport:  OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_prev_metrics",
		"input": []any{
			map[string]any{"type": "input_text", "text": "hello"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:    codexRecoveryReasonPreviousResponseNotFound,
		Transport: OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_invalid_metrics",
		"input": []any{
			map[string]any{"type": "reasoning", "encrypted_content": "gCCC"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:    codexRecoveryReasonInvalidEncryptedContent,
		Transport: OpenAIUpstreamTransportHTTPSSE,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_prev_skip_metrics",
		"input": []any{
			map[string]any{"type": "function_call_output", "call_id": "call_1", "output": "ok"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:    codexRecoveryReasonPreviousResponseNotFound,
		Transport: OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_invalid_skip_metrics",
		"input": []any{
			map[string]any{"type": "input_text", "text": "no encrypted reasoning"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:    codexRecoveryReasonInvalidEncryptedContent,
		Transport: OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	after := SnapshotOpenAICodexCompatibilityMetrics()
	require.GreaterOrEqual(t, after.RecoveryFailoverExhaustAppliedTotal, before.RecoveryFailoverExhaustAppliedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryFailoverExhaustSkippedTotal, before.RecoveryFailoverExhaustSkippedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryTransportCooldownAppliedTotal, before.RecoveryTransportCooldownAppliedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryTransportCooldownSkippedTotal, before.RecoveryTransportCooldownSkippedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryAccountSwitchAppliedTotal, before.RecoveryAccountSwitchAppliedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryAccountSwitchSkippedTotal, before.RecoveryAccountSwitchSkippedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryWSRetryTotal, before.RecoveryWSRetryTotal+1)
	require.GreaterOrEqual(t, after.RecoveryHTTPRetryTotal, before.RecoveryHTTPRetryTotal+1)
	require.GreaterOrEqual(t, after.RecoveryPreviousResponseAppliedTotal, before.RecoveryPreviousResponseAppliedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryPreviousResponseSkippedTotal, before.RecoveryPreviousResponseSkippedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryInvalidEncryptedAppliedTotal, before.RecoveryInvalidEncryptedAppliedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryInvalidEncryptedSkippedTotal, before.RecoveryInvalidEncryptedSkippedTotal+1)
	require.GreaterOrEqual(t, after.RecoveryDropPreviousResponseIDTotal, before.RecoveryDropPreviousResponseIDTotal+1)
	require.GreaterOrEqual(t, after.RecoveryTrimEncryptedReasoningTotal, before.RecoveryTrimEncryptedReasoningTotal+1)
}
