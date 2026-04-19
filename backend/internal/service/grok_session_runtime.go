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

func (r *GrokSessionRuntime) Execute(c *gin.Context, preparation *grokTextPreparation) {
	if c == nil {
		return
	}
	if r == nil || r.gatewayService == nil || r.gatewayService.httpUpstream == nil {
		writeResponsesError(c, http.StatusInternalServerError, "api_error", "Grok gateway service is not configured")
		return
	}
	if preparation == nil || preparation.account == nil {
		writeResponsesError(c, http.StatusServiceUnavailable, "api_error", "No available Grok session accounts")
		return
	}

	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodPost,
		preparation.target.URL,
		bytes.NewReader(preparation.payload),
	)
	if err != nil {
		writeResponsesError(c, http.StatusInternalServerError, "api_error", "Failed to create Grok upstream request")
		return
	}
	req.Header.Set("Accept", "application/json, text/event-stream, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", grokWebBaseURL)
	req.Header.Set("Referer", grokWebBaseURL+"/")
	req.Header.Set("User-Agent", grokSessionProbeUserAgent)
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
		upstreamMsg := sanitizeUpstreamErrorMessage(err.Error())
		if upstreamMsg == "" {
			upstreamMsg = "Upstream request failed"
		}
		writeGrokTextError(c, preparation.protocolFamily, http.StatusBadGateway, "api_error", upstreamMsg)
		return
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
		writeGrokTextError(c, preparation.protocolFamily, mapUpstreamStatusCode(resp.StatusCode), grokResponsesErrorCodeForStatus(resp.StatusCode), upstreamMsg)
		return
	}

	var relayErr error
	switch preparation.protocolFamily {
	case CompatibleGatewayProtocolFamilyChatCompletions:
		relayErr = relayGrokSessionChatCompletions(c, resp.Body, preparation.requestedModel, preparation.stream, preparation.includeUsage)
	case CompatibleGatewayProtocolFamilyMessages:
		relayErr = relayGrokSessionAnthropic(c, resp.Body, preparation.requestedModel, preparation.stream)
	default:
		relayErr = relayGrokSessionResponses(c, resp.Body, preparation.requestedModel, preparation.stream)
	}
	if relayErr != nil {
		if c.Writer.Written() {
			return
		}
		var httpErr *grokResponsesHTTPError
		if errors.As(relayErr, &httpErr) {
			writeGrokTextError(c, preparation.protocolFamily, httpErr.statusCode, httpErr.code, httpErr.message)
			return
		}
		writeGrokTextError(c, preparation.protocolFamily, http.StatusBadGateway, "api_error", "Upstream stream ended without a response")
	}
}

func resolveGrokGatewayTLSProfile(gatewayService *GatewayService, account *Account) *tlsfingerprint.Profile {
	if gatewayService == nil || gatewayService.tlsFPProfileService == nil {
		return nil
	}
	return gatewayService.tlsFPProfileService.ResolveTLSProfile(account)
}
