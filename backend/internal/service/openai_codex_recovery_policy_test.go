package service

import (
	"errors"
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
		require.True(t, decision.Warmup)
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
		require.False(t, decision.Warmup)
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
		require.False(t, decision.Warmup)
	})

	t.Run("marks_prewarm_failover_as_warmup", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			AccountID:     48,
			FailureReason: "prewarm_upstream_5xx",
			Reason:        codexRecoveryReasonFailover,
			StatusCode:    http.StatusBadGateway,
			Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.True(t, decision.Applied)
		require.True(t, decision.SwitchAccount)
		require.Equal(t, "upstream_5xx", decision.FailureReason)
		require.True(t, decision.Warmup)
	})

	t.Run("switches_account_on_ws_rate_limit_failover", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			AccountID:     45,
			FailureReason: "upstream_rate_limited",
			Reason:        codexRecoveryReasonFailover,
			StatusCode:    http.StatusTooManyRequests,
			Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.True(t, decision.Applied)
		require.False(t, decision.ExhaustFailover)
		require.True(t, decision.SwitchAccount)
		require.Equal(t, codexRecoveryActionSwitchAccount, decision.Action)
	})

	t.Run("exhausts_ws_auth_failed_forbidden", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			FailureReason: "auth_failed",
			Reason:        codexRecoveryReasonFailover,
			StatusCode:    http.StatusForbidden,
			Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
		})

		require.True(t, decision.Applied)
		require.True(t, decision.ExhaustFailover)
		require.False(t, decision.SwitchAccount)
		require.Equal(t, codexRecoveryActionExhaustFailover, decision.Action)
	})

	t.Run("switches_account_on_http_request_timeout", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			AccountID:     46,
			FailureReason: "request_timeout",
			Reason:        codexRecoveryReasonFailover,
			StatusCode:    http.StatusGatewayTimeout,
			Transport:     OpenAIUpstreamTransportHTTPSSE,
		})

		require.True(t, decision.Applied)
		require.True(t, decision.SwitchAccount)
		require.Equal(t, codexRecoveryActionSwitchAccount, decision.Action)
	})

	t.Run("switches_account_on_http_request_error", func(t *testing.T) {
		decision := policy.Apply(nil, CodexRecoveryPolicyInput{
			AccountID:     47,
			FailureReason: "request_error",
			Reason:        codexRecoveryReasonFailover,
			StatusCode:    http.StatusBadGateway,
			Transport:     OpenAIUpstreamTransportHTTPSSE,
		})

		require.True(t, decision.Applied)
		require.True(t, decision.SwitchAccount)
		require.Equal(t, codexRecoveryActionSwitchAccount, decision.Action)
	})
}

func TestOpenAIGatewayService_ApplyCodexTransportCooldownRecovery(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.FallbackCooldownSeconds = 30
	svc := &OpenAIGatewayService{cfg: cfg}

	decision := svc.applyCodexTransportCooldownRecovery(52, "read_event", OpenAIUpstreamTransportResponsesWebsocketV2, true)

	require.True(t, decision.Applied)
	require.True(t, decision.MarkedTransportCooldown)
	require.True(t, svc.isOpenAIWSFallbackCooling(52))
	require.True(t, decision.TrackCompatibilityMetrics)
}

func TestOpenAIGatewayService_ApplyCodexTransportCooldownRecovery_NonNativeMetricsSkipped(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.FallbackCooldownSeconds = 30
	svc := &OpenAIGatewayService{cfg: cfg}

	before := SnapshotOpenAICodexCompatibilityMetrics()
	decision := svc.applyCodexTransportCooldownRecovery(53, "read_event", OpenAIUpstreamTransportResponsesWebsocketV2, false)
	after := SnapshotOpenAICodexCompatibilityMetrics()

	require.True(t, decision.Applied)
	require.True(t, decision.MarkedTransportCooldown)
	require.True(t, svc.isOpenAIWSFallbackCooling(53))
	require.False(t, decision.TrackCompatibilityMetrics)
	require.Equal(t, before.RecoveryTransportCooldownAppliedTotal, after.RecoveryTransportCooldownAppliedTotal)
}

func TestClassifyCodexWSFailoverError(t *testing.T) {
	t.Run("ws_rate_limit", func(t *testing.T) {
		failoverErr := classifyCodexWSFailoverError(wrapOpenAIWSFallback("upstream_rate_limited", &openAIWSDialError{
			StatusCode: http.StatusTooManyRequests,
			Err:        errors.New("rate limited"),
		}))

		require.NotNil(t, failoverErr)
		require.Equal(t, http.StatusTooManyRequests, failoverErr.StatusCode)
		require.Equal(t, "upstream_rate_limited", failoverErr.FailureReason)
	})

	t.Run("ws_upstream_5xx", func(t *testing.T) {
		failoverErr := classifyCodexWSFailoverError(wrapOpenAIWSFallback("upstream_5xx", &openAIWSDialError{
			StatusCode: http.StatusBadGateway,
			Err:        errors.New("bad gateway"),
		}))

		require.NotNil(t, failoverErr)
		require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
		require.Equal(t, "upstream_5xx", failoverErr.FailureReason)
		require.False(t, failoverErr.Warmup)
	})

	t.Run("ws_auth_failed_forbidden", func(t *testing.T) {
		failoverErr := classifyCodexWSFailoverError(wrapOpenAIWSFallback("auth_failed", &openAIWSDialError{
			StatusCode: http.StatusForbidden,
			Err:        errors.New("forbidden"),
		}))

		require.NotNil(t, failoverErr)
		require.Equal(t, http.StatusForbidden, failoverErr.StatusCode)
		require.Equal(t, "auth_failed", failoverErr.FailureReason)
		require.False(t, failoverErr.Warmup)
	})

	t.Run("ws_prewarm_upstream_5xx", func(t *testing.T) {
		failoverErr := classifyCodexWSFailoverError(wrapOpenAIWSFallback("prewarm_upstream_5xx", &openAIWSDialError{
			StatusCode: http.StatusBadGateway,
			Err:        errors.New("bad gateway"),
		}))

		require.NotNil(t, failoverErr)
		require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
		require.Equal(t, "upstream_5xx", failoverErr.FailureReason)
		require.True(t, failoverErr.Warmup)
	})

	t.Run("transport_failure_stays_local", func(t *testing.T) {
		failoverErr := classifyCodexWSFailoverError(wrapOpenAIWSFallback("read_event", errors.New("read failed")))
		require.Nil(t, failoverErr)
	})
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

	decision := svc.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, true)

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
		FailureReason:             "upstream_rate_limited",
		Reason:                    codexRecoveryReasonFailover,
		StatusCode:                http.StatusTooManyRequests,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportHTTPSSE,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		FailureReason:             "upstream_5xx",
		Reason:                    codexRecoveryReasonFailover,
		StatusCode:                http.StatusBadGateway,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportHTTPSSE,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:                 71,
		FailureReason:             "read_event",
		Reason:                    codexRecoveryReasonTransportFailure,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		Reason:                    codexRecoveryReasonTransportFailure,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:                 72,
		Reason:                    codexRecoveryReasonAccountSwitch,
		StatusCode:                http.StatusBadGateway,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		Reason:                    codexRecoveryReasonAccountSwitch,
		StatusCode:                http.StatusBadGateway,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_prev_metrics",
		"input": []any{
			map[string]any{"type": "input_text", "text": "hello"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:                    codexRecoveryReasonPreviousResponseNotFound,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_invalid_metrics",
		"input": []any{
			map[string]any{"type": "reasoning", "encrypted_content": "gCCC"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:                    codexRecoveryReasonInvalidEncryptedContent,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportHTTPSSE,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_prev_skip_metrics",
		"input": []any{
			map[string]any{"type": "function_call_output", "call_id": "call_1", "output": "ok"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:                    codexRecoveryReasonPreviousResponseNotFound,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(map[string]any{
		"previous_response_id": "resp_invalid_skip_metrics",
		"input": []any{
			map[string]any{"type": "input_text", "text": "no encrypted reasoning"},
		},
	}, CodexRecoveryPolicyInput{
		Reason:                    codexRecoveryReasonInvalidEncryptedContent,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
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

func TestOpenAIGatewayService_RecordCodexRecoveryAccountSwitch_NonOfficialMetricsSkipped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/chat/completions", nil)
	c.Set("openai_ws_transport_decision", string(OpenAIUpstreamTransportResponsesWebsocketV2))

	before := SnapshotOpenAICodexCompatibilityMetrics()

	svc := &OpenAIGatewayService{}
	account := &Account{ID: 88}
	failoverErr := &UpstreamFailoverError{StatusCode: http.StatusBadGateway}

	decision := svc.RecordCodexRecoveryAccountSwitch(c, account, failoverErr, false)

	require.True(t, decision.Applied)
	require.True(t, decision.SwitchAccount)
	require.False(t, decision.TrackCompatibilityMetrics)

	after := SnapshotOpenAICodexCompatibilityMetrics()
	require.Equal(t, before.RecoveryAccountSwitchAppliedTotal, after.RecoveryAccountSwitchAppliedTotal)
	require.Equal(t, before.RecoveryWSRetryTotal, after.RecoveryWSRetryTotal)
}

func TestSnapshotOpenAICodexCompatibilityMetrics_RecoveryWarmupSkipped(t *testing.T) {
	before := SnapshotOpenAICodexCompatibilityMetrics()
	policy := CodexRecoveryPolicy{}

	policy.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:                 91,
		FailureReason:             "prewarm_read_event",
		Reason:                    codexRecoveryReasonTransportFailure,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:                 92,
		FailureReason:             "prewarm_upstream_5xx",
		Reason:                    codexRecoveryReasonFailover,
		StatusCode:                http.StatusBadGateway,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
	})

	policy.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:                 93,
		Reason:                    codexRecoveryReasonAccountSwitch,
		StatusCode:                http.StatusBadGateway,
		TrackCompatibilityMetrics: true,
		Transport:                 OpenAIUpstreamTransportResponsesWebsocketV2,
		Warmup:                    true,
	})

	after := SnapshotOpenAICodexCompatibilityMetrics()
	require.Equal(t, before.RecoveryTransportCooldownAppliedTotal, after.RecoveryTransportCooldownAppliedTotal)
	require.Equal(t, before.RecoveryFailoverExhaustAppliedTotal, after.RecoveryFailoverExhaustAppliedTotal)
	require.Equal(t, before.RecoveryFailoverExhaustSkippedTotal, after.RecoveryFailoverExhaustSkippedTotal)
	require.Equal(t, before.RecoveryAccountSwitchAppliedTotal, after.RecoveryAccountSwitchAppliedTotal)
	require.Equal(t, before.RecoveryWSRetryTotal, after.RecoveryWSRetryTotal)
}
