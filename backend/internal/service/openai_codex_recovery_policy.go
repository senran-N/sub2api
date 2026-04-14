package service

import "strings"

const (
	codexRecoveryActionNone                     = ""
	codexRecoveryActionDropPreviousResponseID   = "drop_previous_response_id"
	codexRecoveryActionTrimEncryptedReasoning   = "drop_encrypted_reasoning_items"
	codexRecoveryReasonInvalidEncryptedContent  = "invalid_encrypted_content"
	codexRecoveryReasonPreviousResponseNotFound = "previous_response_not_found"
)

type CodexRecoveryPolicyInput struct {
	Reason    string
	Transport OpenAIUpstreamTransport
}

type CodexRecoveryDecision struct {
	Action                    string
	Applied                   bool
	DroppedPreviousResponseID bool
	HasFunctionCallOutput     bool
	PreviousResponseID        string
	PreviousResponseIDKind    string
	Reason                    string
	SkipReason                string
	Transport                 OpenAIUpstreamTransport
	TrimmedEncryptedReasoning bool
}

// CodexRecoveryPolicy centralizes Codex-specific minimal recovery rewrites so
// HTTP and WS forwarding can share the same decision matrix and metrics.
type CodexRecoveryPolicy struct{}

func (CodexRecoveryPolicy) Apply(reqBody map[string]any, input CodexRecoveryPolicyInput) CodexRecoveryDecision {
	decision := CodexRecoveryDecision{
		Action:                codexRecoveryActionNone,
		PreviousResponseID:    openAIWSPayloadString(reqBody, "previous_response_id"),
		Reason:                strings.TrimSpace(input.Reason),
		Transport:             normalizeOpenAIWSSessionTransport(input.Transport),
		HasFunctionCallOutput: HasFunctionCallOutput(reqBody),
	}
	decision.PreviousResponseIDKind = ClassifyOpenAIPreviousResponseIDKind(decision.PreviousResponseID)

	switch decision.Reason {
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
