package service

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	codexRecoveryActionNone                     = ""
	codexRecoveryActionDropPreviousResponseID   = "drop_previous_response_id"
	codexRecoveryActionExhaustFailover          = "exhaust_failover"
	codexRecoveryActionMarkTransportCooldown    = "mark_transport_cooldown"
	codexRecoveryActionSwitchAccount            = "switch_account"
	codexRecoveryActionTrimEncryptedReasoning   = "drop_encrypted_reasoning_items"
	codexRecoveryReasonFailover                 = "failover"
	codexRecoveryReasonAccountSwitch            = "account_switch"
	codexRecoveryReasonInvalidEncryptedContent  = "invalid_encrypted_content"
	codexRecoveryReasonPreviousResponseNotFound = "previous_response_not_found"
	codexRecoveryReasonTransportFailure         = "transport_failure"
)

type CodexRecoveryPolicyInput struct {
	AccountID     int64
	FailureReason string
	Reason        string
	StatusCode    int
	Transport     OpenAIUpstreamTransport
}

type CodexRecoveryDecision struct {
	Action                    string
	AccountID                 int64
	Applied                   bool
	DroppedPreviousResponseID bool
	ExhaustFailover           bool
	FailureReason             string
	HasFunctionCallOutput     bool
	MarkedTransportCooldown   bool
	PreviousResponseID        string
	PreviousResponseIDKind    string
	Reason                    string
	SkipReason                string
	StatusCode                int
	SwitchAccount             bool
	Transport                 OpenAIUpstreamTransport
	TrimmedEncryptedReasoning bool
}

// CodexRecoveryPolicy centralizes Codex-specific minimal recovery rewrites so
// HTTP and WS forwarding can share the same decision matrix and metrics.
type CodexRecoveryPolicy struct{}

func (CodexRecoveryPolicy) Apply(reqBody map[string]any, input CodexRecoveryPolicyInput) CodexRecoveryDecision {
	decision := CodexRecoveryDecision{
		Action:                codexRecoveryActionNone,
		AccountID:             input.AccountID,
		FailureReason:         normalizeCodexRecoveryFailureReason(input.FailureReason),
		PreviousResponseID:    openAIWSPayloadString(reqBody, "previous_response_id"),
		Reason:                strings.TrimSpace(input.Reason),
		StatusCode:            input.StatusCode,
		Transport:             normalizeOpenAIWSSessionTransport(input.Transport),
		HasFunctionCallOutput: HasFunctionCallOutput(reqBody),
	}
	decision.PreviousResponseIDKind = ClassifyOpenAIPreviousResponseIDKind(decision.PreviousResponseID)

	switch decision.Reason {
	case codexRecoveryReasonFailover:
		switch {
		case isCodexImmediateFailoverExhaust(decision.StatusCode, decision.FailureReason):
			decision.Action = codexRecoveryActionExhaustFailover
			decision.Applied = true
			decision.ExhaustFailover = true
		case isCodexFailoverSwitchAccountStatus(decision.StatusCode):
			if decision.AccountID <= 0 {
				decision.SkipReason = "missing_account_id"
				recordOpenAICodexRecoveryDecision(decision)
				return decision
			}
			decision.Action = codexRecoveryActionSwitchAccount
			decision.Applied = true
			decision.SwitchAccount = true
		default:
			decision.SkipReason = "unsupported_failover_status"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
	case codexRecoveryReasonTransportFailure:
		if decision.AccountID <= 0 {
			decision.SkipReason = "missing_account_id"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		if !isOpenAIWSSessionWebsocketTransport(decision.Transport) {
			decision.SkipReason = "unsupported_transport"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		if decision.FailureReason == "" {
			decision.SkipReason = "missing_failure_reason"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		if !isCodexTransportCooldownFailureReason(decision.FailureReason) {
			decision.SkipReason = "unsupported_failure_reason"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		decision.Action = codexRecoveryActionMarkTransportCooldown
		decision.Applied = true
		decision.MarkedTransportCooldown = true
	case codexRecoveryReasonAccountSwitch:
		if decision.AccountID <= 0 {
			decision.SkipReason = "missing_account_id"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		if decision.StatusCode <= 0 {
			decision.SkipReason = "missing_status_code"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		decision.Action = codexRecoveryActionSwitchAccount
		decision.Applied = true
		decision.SwitchAccount = true
	case codexRecoveryReasonPreviousResponseNotFound:
		if decision.PreviousResponseID == "" {
			decision.SkipReason = "missing_previous_response_id"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		if decision.HasFunctionCallOutput {
			decision.SkipReason = "has_function_call_output"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		delete(reqBody, "previous_response_id")
		decision.Action = codexRecoveryActionDropPreviousResponseID
		decision.Applied = true
		decision.DroppedPreviousResponseID = true
	case codexRecoveryReasonInvalidEncryptedContent:
		decision.TrimmedEncryptedReasoning = trimOpenAIEncryptedReasoningItems(reqBody)
		if !decision.TrimmedEncryptedReasoning {
			decision.SkipReason = "missing_encrypted_reasoning_items"
			recordOpenAICodexRecoveryDecision(decision)
			return decision
		}
		decision.Action = codexRecoveryActionTrimEncryptedReasoning
		decision.Applied = true
		if decision.Transport != OpenAIUpstreamTransportHTTPSSE && decision.PreviousResponseID != "" && !decision.HasFunctionCallOutput {
			delete(reqBody, "previous_response_id")
			decision.DroppedPreviousResponseID = true
		}
	default:
		decision.SkipReason = "unsupported_reason"
		recordOpenAICodexRecoveryDecision(decision)
		return decision
	}

	recordOpenAICodexRecoveryDecision(decision)
	return decision
}

func normalizeCodexRecoveryFailureReason(reason string) string {
	reason = strings.TrimSpace(reason)
	reason = strings.TrimPrefix(reason, "prewarm_")
	return reason
}

func isCodexTransportCooldownFailureReason(reason string) bool {
	switch normalizeCodexRecoveryFailureReason(reason) {
	case "read_event",
		"write_request",
		"write",
		"acquire_timeout",
		"acquire_conn",
		"conn_queue_full",
		"dial_failed",
		"upstream_5xx",
		"event_error",
		"error_event",
		"upstream_error_event",
		"ws_connection_limit_reached",
		"missing_final_response":
		return true
	default:
		return false
	}
}

func isCodexImmediateFailoverExhaust(statusCode int, failureReason string) bool {
	switch normalizeCodexRecoveryFailureReason(failureReason) {
	case "upstream_rate_limited":
		return true
	}
	switch statusCode {
	case 401, 403, 429:
		return true
	default:
		return false
	}
}

func isCodexFailoverSwitchAccountStatus(statusCode int) bool {
	return statusCode >= 500
}

func resolveCodexRecoveryTransport(c *gin.Context) OpenAIUpstreamTransport {
	if c != nil {
		if raw := strings.TrimSpace(c.GetString("openai_ws_transport_decision")); raw != "" {
			return normalizeOpenAIWSSessionTransport(OpenAIUpstreamTransport(raw))
		}
	}
	switch GetOpenAIClientTransport(c) {
	case OpenAIClientTransportHTTP:
		return OpenAIUpstreamTransportHTTPSSE
	case OpenAIClientTransportWS:
		return OpenAIUpstreamTransportResponsesWebsocketV2
	default:
		return OpenAIUpstreamTransportAny
	}
}

func (s *OpenAIGatewayService) applyCodexTransportCooldownRecovery(
	accountID int64,
	reason string,
	transport OpenAIUpstreamTransport,
) CodexRecoveryDecision {
	decision := CodexRecoveryPolicy{}.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:     accountID,
		FailureReason: reason,
		Reason:        codexRecoveryReasonTransportFailure,
		Transport:     transport,
	})
	if decision.Applied && decision.MarkedTransportCooldown {
		s.markOpenAIWSFallbackCooling(accountID, decision.FailureReason)
	}
	return decision
}

func (s *OpenAIGatewayService) RecordCodexRecoveryAccountSwitch(
	c *gin.Context,
	account *Account,
	failoverErr *UpstreamFailoverError,
) CodexRecoveryDecision {
	if account == nil || failoverErr == nil {
		return CodexRecoveryDecision{}
	}
	return CodexRecoveryPolicy{}.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:     account.ID,
		Reason:        codexRecoveryReasonAccountSwitch,
		FailureReason: failoverErr.FailureReason,
		StatusCode:    failoverErr.StatusCode,
		Transport:     resolveCodexRecoveryTransport(c),
	})
}

func (s *OpenAIGatewayService) ResolveCodexFailoverRecovery(
	c *gin.Context,
	account *Account,
	failoverErr *UpstreamFailoverError,
) CodexRecoveryDecision {
	if failoverErr == nil {
		return CodexRecoveryDecision{}
	}
	accountID := int64(0)
	if account != nil {
		accountID = account.ID
	}
	return CodexRecoveryPolicy{}.Apply(nil, CodexRecoveryPolicyInput{
		AccountID:     accountID,
		FailureReason: failoverErr.FailureReason,
		Reason:        codexRecoveryReasonFailover,
		StatusCode:    failoverErr.StatusCode,
		Transport:     resolveCodexRecoveryTransport(c),
	})
}
