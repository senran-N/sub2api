package service

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

type grokRuntimeErrorClass string

const (
	grokRuntimeErrorClassAuth             grokRuntimeErrorClass = "auth"
	grokRuntimeErrorClassRateLimited      grokRuntimeErrorClass = "rate_limited"
	grokRuntimeErrorClassModelUnsupported grokRuntimeErrorClass = "model_unsupported"
	grokRuntimeErrorClassUpstream         grokRuntimeErrorClass = "upstream"
	grokRuntimeErrorClassTransport        grokRuntimeErrorClass = "transport"
	grokRuntimeErrorClassInvalidRequest   grokRuntimeErrorClass = "invalid_request"
	grokRuntimeErrorClassUnknown          grokRuntimeErrorClass = "unknown"
)

type grokRuntimePenaltyScope string

const (
	grokRuntimePenaltyScopeNone    grokRuntimePenaltyScope = ""
	grokRuntimePenaltyScopeAccount grokRuntimePenaltyScope = "account"
	grokRuntimePenaltyScopeModel   grokRuntimePenaltyScope = "model"
)

type grokRuntimeErrorClassification struct {
	StatusCode int
	Reason     string
	Class      grokRuntimeErrorClass
	Scope      grokRuntimePenaltyScope
	Retryable  bool
	Cooldown   time.Duration
}

type grokRuntimeErrorSignal struct {
	statusCode int
	code       string
	message    string
}

var grokInvalidCredentialMarkers = []string{
	"invalid-credentials",
	"invalid_credentials",
	"bad-credentials",
	"bad_credentials",
	"failed to look up session id",
	"blocked-user",
	"blocked_user",
	"email-domain-rejected",
	"email_domain_rejected",
	"session not found",
	"account suspended",
	"token revoked",
	"token expired",
	"invalid credentials",
}

func classifyGrokRuntimeError(input GrokRuntimeFeedbackInput) grokRuntimeErrorClassification {
	signal := extractGrokRuntimeErrorSignal(input)
	if signal.statusCode == 0 && signal.message == "" {
		return grokRuntimeErrorClassification{}
	}

	classification := grokRuntimeErrorClassification{
		StatusCode: signal.statusCode,
		Reason:     signal.message,
		Class:      grokRuntimeErrorClassUnknown,
		Scope:      grokRuntimePenaltyScopeAccount,
		Cooldown:   3 * time.Minute,
	}

	if grokSignalLooksLikeInvalidCredentials(signal) {
		classification.Class = grokRuntimeErrorClassAuth
		classification.Scope = grokRuntimePenaltyScopeAccount
		classification.Cooldown = 30 * time.Minute
		return classification
	}

	if isGrokRuntimeModelUnsupportedSignal(signal) {
		classification.Class = grokRuntimeErrorClassModelUnsupported
		classification.Scope = grokRuntimePenaltyScopeModel
		classification.Cooldown = 45 * time.Minute
		return classification
	}

	switch {
	case signal.statusCode == http.StatusUnauthorized:
		classification.Class = grokRuntimeErrorClassAuth
		classification.Scope = grokRuntimePenaltyScopeAccount
		classification.Cooldown = 30 * time.Minute
	case signal.statusCode == http.StatusForbidden:
		classification.Class = grokRuntimeErrorClassAuth
		classification.Scope = grokRuntimePenaltyScopeAccount
		classification.Cooldown = 30 * time.Minute
	case signal.statusCode == http.StatusTooManyRequests:
		classification.Class = grokRuntimeErrorClassRateLimited
		classification.Scope = grokRuntimePenaltyScopeAccount
		classification.Retryable = true
		classification.Cooldown = 10 * time.Minute
	case signal.statusCode >= 500:
		classification.Class = grokRuntimeErrorClassUpstream
		classification.Scope = grokRuntimePenaltyScopeAccount
		classification.Retryable = true
		classification.Cooldown = 5 * time.Minute
	case signal.statusCode == 0:
		classification.Class = grokRuntimeErrorClassTransport
		classification.Scope = grokRuntimePenaltyScopeAccount
		classification.Retryable = true
		classification.Cooldown = 2 * time.Minute
	case signal.statusCode == http.StatusBadRequest:
		classification.Class = grokRuntimeErrorClassInvalidRequest
		classification.Scope = grokRuntimePenaltyScopeNone
		classification.Cooldown = 0
	default:
		classification.Class = grokRuntimeErrorClassUnknown
	}

	return classification
}

func extractGrokRuntimeErrorSignal(input GrokRuntimeFeedbackInput) grokRuntimeErrorSignal {
	statusCode := input.StatusCode
	message := ""
	code := ""

	var failoverErr *UpstreamFailoverError
	if errors.As(input.Err, &failoverErr) && failoverErr != nil {
		if failoverErr.StatusCode > 0 {
			statusCode = failoverErr.StatusCode
		}
		code = strings.ToLower(strings.TrimSpace(extractUpstreamErrorCode(failoverErr.ResponseBody)))
		message = strings.TrimSpace(ExtractUpstreamErrorMessage(failoverErr.ResponseBody))
		if message == "" {
			message = strings.TrimSpace(failoverErr.FailureReason)
		}
	}

	if message == "" && input.Err != nil {
		message = strings.TrimSpace(input.Err.Error())
	}

	return grokRuntimeErrorSignal{
		statusCode: statusCode,
		code:       code,
		message:    strings.ToLower(strings.TrimSpace(message)),
	}
}

func isGrokRuntimeModelUnsupportedSignal(signal grokRuntimeErrorSignal) bool {
	switch signal.code {
	case "model_not_found", "invalid_model", "unsupported_model", "model_not_supported", "insufficient_tier", "tier_required":
		return true
	}

	if signal.message == "" {
		return false
	}

	for _, pattern := range []string{
		"model not found",
		"unknown model",
		"unsupported model",
		"model is not supported",
		"does not support model",
		"tier required",
		"requires super",
		"requires heavy",
		"requires basic",
		"available on super",
		"available on heavy",
		"insufficient tier",
		"model access denied",
	} {
		if strings.Contains(signal.message, pattern) {
			return true
		}
	}

	return signal.statusCode == http.StatusNotFound && strings.Contains(signal.message, "model")
}

func grokSignalLooksLikeInvalidCredentials(signal grokRuntimeErrorSignal) bool {
	if grokInvalidCredentialsCode(signal.code) {
		return true
	}
	return grokInvalidCredentialsBody(signal.message)
}

func grokInvalidCredentialsCode(code string) bool {
	normalized := strings.ToLower(strings.TrimSpace(code))
	if normalized == "" {
		return false
	}

	switch normalized {
	case "invalid-credentials",
		"invalid_credentials",
		"bad-credentials",
		"bad_credentials",
		"blocked-user",
		"blocked_user",
		"email-domain-rejected",
		"email_domain_rejected",
		"session_not_found",
		"session-not-found",
		"account_suspended",
		"account-suspended",
		"token_revoked",
		"token-revoked",
		"token_expired",
		"token-expired":
		return true
	default:
		return false
	}
}

func grokInvalidCredentialsBody(body string) bool {
	text := strings.ToLower(strings.TrimSpace(body))
	if text == "" {
		return false
	}
	for _, marker := range grokInvalidCredentialMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}
