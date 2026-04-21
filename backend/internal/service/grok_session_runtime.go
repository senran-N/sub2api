package service

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"
)

type GrokSessionRuntime struct {
	gatewayService *GatewayService
}

func NewGrokSessionRuntime(gatewayService *GatewayService) *GrokSessionRuntime {
	return &GrokSessionRuntime{gatewayService: gatewayService}
}

func (r *GrokSessionRuntime) Execute(c *gin.Context, preparation *grokTextPreparation) error {
	if c == nil {
		return nil
	}
	if r == nil || r.gatewayService == nil || r.gatewayService.httpUpstream == nil {
		writeResponsesError(c, http.StatusInternalServerError, "api_error", "Grok gateway service is not configured")
		return nil
	}
	if preparation == nil || preparation.account == nil {
		writeResponsesError(c, http.StatusServiceUnavailable, "api_error", "No available Grok session accounts")
		return nil
	}

	persistFeedback := func(statusCode int, runtimeErr error) {
		if r == nil || r.gatewayService == nil {
			return
		}
		persistGrokRuntimeFeedbackToRepo(c.Request.Context(), r.gatewayService.accountRepo, GrokRuntimeFeedbackInput{
			Account:        preparation.account,
			RequestedModel: preparation.requestedModel,
			QuotaWindow:    preparation.quotaWindow,
			StatusCode:     statusCode,
			ProtocolFamily: preparation.protocolFamily,
			Err:            runtimeErr,
		})
	}

	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodPost,
		preparation.target.URL,
		bytes.NewReader(preparation.payload),
	)
	if err != nil {
		writeResponsesError(c, http.StatusInternalServerError, "api_error", "Failed to create Grok upstream request")
		return err
	}
	applyGrokSessionBrowserHeaders(req.Header, preparation.target, grokSessionTextAcceptHeader)
	req.Header.Set("Content-Type", "application/json")
	preparation.target.Apply(req)

	proxyURL := ""
	if preparation.account.ProxyID != nil && preparation.account.Proxy != nil {
		proxyURL = preparation.account.Proxy.URL()
	}

	resp, err := r.gatewayService.httpUpstream.DoWithTLS(
		req,
		proxyURL,
		preparation.account.ID,
		preparation.account.Concurrency,
		resolveGrokGatewayTLSProfile(r.gatewayService, preparation.account),
	)
	if err != nil {
		failoverErr := newGrokSessionFailoverError(0, nil, err)
		persistFeedback(0, firstNonNilError(failoverErr, err))
		if failoverErr != nil {
			return failoverErr
		}
		upstreamMsg := sanitizeUpstreamErrorMessage(err.Error())
		if upstreamMsg == "" {
			upstreamMsg = "Upstream request failed"
		}
		writeGrokTextError(c, preparation.protocolFamily, http.StatusBadGateway, "api_error", upstreamMsg)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if requestID := strings.TrimSpace(resp.Header.Get("x-request-id")); requestID != "" {
		c.Header("x-request-id", requestID)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(resp.Body)
		upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		if upstreamMsg == "" {
			upstreamMsg = http.StatusText(resp.StatusCode)
		}
		runtimeErr := errors.New(upstreamMsg)
		failoverErr := newGrokSessionFailoverError(resp.StatusCode, respBody, runtimeErr)
		persistFeedback(resp.StatusCode, firstNonNilError(failoverErr, runtimeErr))
		if failoverErr != nil {
			return failoverErr
		}
		writeGrokTextError(c, preparation.protocolFamily, mapUpstreamStatusCode(resp.StatusCode), grokResponsesErrorCodeForStatus(resp.StatusCode), upstreamMsg)
		return runtimeErr
	}

	var relayErr error
	switch preparation.protocolFamily {
	case CompatibleGatewayProtocolFamilyChatCompletions:
		relayErr = relayGrokSessionChatCompletionsWithSettings(
			c,
			resp.Body,
			preparation.requestedModel,
			preparation.stream,
			preparation.includeUsage,
			preparation.toolNames,
			preparation.textSettings,
		)
	case CompatibleGatewayProtocolFamilyMessages:
		relayErr = relayGrokSessionAnthropicWithSettings(
			c,
			resp.Body,
			preparation.requestedModel,
			preparation.stream,
			preparation.toolNames,
			preparation.textSettings,
		)
	default:
		relayErr = relayGrokSessionResponsesWithSettings(
			c,
			resp.Body,
			preparation.requestedModel,
			preparation.stream,
			preparation.toolNames,
			preparation.textSettings,
		)
	}
	statusCode := resp.StatusCode
	if relayErr != nil {
		var httpErr *grokResponsesHTTPError
		if errors.As(relayErr, &httpErr) && httpErr != nil && httpErr.statusCode > 0 {
			statusCode = httpErr.statusCode
		}
	}
	if relayErr == nil {
		persistFeedback(statusCode, nil)
	}
	if relayErr != nil {
		if c.Writer.Written() {
			persistFeedback(statusCode, relayErr)
			return relayErr
		}
		failoverErr := newGrokSessionFailoverError(statusCode, nil, relayErr)
		persistFeedback(statusCode, firstNonNilError(failoverErr, relayErr))
		if failoverErr != nil {
			return failoverErr
		}
		var httpErr *grokResponsesHTTPError
		if errors.As(relayErr, &httpErr) {
			writeGrokTextError(c, preparation.protocolFamily, httpErr.statusCode, httpErr.code, httpErr.message)
			return relayErr
		}
		writeGrokTextError(c, preparation.protocolFamily, http.StatusBadGateway, "api_error", "Upstream stream ended without a response")
	}
	return relayErr
}

func resolveGrokGatewayTLSProfile(gatewayService *GatewayService, account *Account) *tlsfingerprint.Profile {
	if gatewayService == nil {
		return nil
	}
	return resolveGrokTLSProfile(account, gatewayService.tlsFPProfileService)
}

func newGrokSessionFailoverError(statusCode int, responseBody []byte, runtimeErr error) *UpstreamFailoverError {
	input := GrokRuntimeFeedbackInput{
		StatusCode: statusCode,
		Err:        runtimeErr,
	}
	if len(responseBody) > 0 {
		input.Err = &UpstreamFailoverError{
			StatusCode:   statusCode,
			ResponseBody: append([]byte(nil), responseBody...),
		}
	}

	classification := classifyGrokRuntimeError(input)
	if classification.Scope == grokRuntimePenaltyScopeNone {
		return nil
	}

	return &UpstreamFailoverError{
		StatusCode:   statusCode,
		ResponseBody: append([]byte(nil), responseBody...),
		FailureReason: firstNonEmpty(
			strings.TrimSpace(classification.Reason),
			strings.TrimSpace(extractUpstreamErrorMessage(responseBody)),
			strings.TrimSpace(string(classification.Class)),
		),
	}
}

func firstNonNilError(preferred error, fallback error) error {
	if preferred != nil {
		return preferred
	}
	return fallback
}
