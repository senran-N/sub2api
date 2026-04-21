package service

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/senran-N/sub2api/internal/util/cloudflareutil"
)

const (
	grokProbeErrorBodyLimit      = 16 << 10
	grokProbeErrorPreviewLimit   = 512
	grokSessionProbeAcceptHeader = "application/json, text/event-stream, text/plain, */*"
)

type grokProbeHTTPErrorSummary struct {
	StatusCode            int
	Message               string
	AuthenticationMessage string
	IsCloudflareChallenge bool
}

func grokReadProbeErrorBody(resp *http.Response) []byte {
	if resp == nil || resp.Body == nil {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, grokProbeErrorBodyLimit))
	return body
}

func grokSummarizeProbeHTTPError(resp *http.Response, body []byte) grokProbeHTTPErrorSummary {
	statusCode := 0
	var headers http.Header
	if resp != nil {
		statusCode = resp.StatusCode
		headers = resp.Header
	}

	if cloudflareutil.IsCloudflareChallengeResponse(statusCode, headers, body) {
		message := cloudflareutil.FormatCloudflareChallengeMessage(
			fmt.Sprintf(
				"Cloudflare challenge encountered (HTTP %d); upstream requires browser verification. Refresh Grok cookies if applicable and retry with a browser-matched TLS fingerprint",
				statusCode,
			),
			headers,
			body,
		)
		return grokProbeHTTPErrorSummary{
			StatusCode:            statusCode,
			Message:               normalizeGrokProbeErrorText(message),
			IsCloudflareChallenge: true,
		}
	}

	detail := grokProbeErrorDetail(body)
	message := fmt.Sprintf("API returned %d", statusCode)
	if detail != "" {
		message += ": " + detail
	}

	summary := grokProbeHTTPErrorSummary{
		StatusCode: statusCode,
		Message:    normalizeGrokProbeErrorText(message),
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		summary.AuthenticationMessage = fmt.Sprintf("Authentication failed (%d)", statusCode)
		if detail != "" {
			summary.AuthenticationMessage += ": " + detail
		}
		summary.AuthenticationMessage = normalizeGrokProbeErrorText(summary.AuthenticationMessage)
	}
	return summary
}

func grokProbeErrorDetail(body []byte) string {
	code, message := cloudflareutil.ExtractUpstreamErrorCodeAndMessage(body)
	code = normalizeGrokProbeErrorText(code)
	message = normalizeGrokProbeErrorText(message)

	switch {
	case code != "" && message != "" && !strings.Contains(strings.ToLower(message), strings.ToLower(code)):
		return fmt.Sprintf("%s (%s)", message, code)
	case message != "":
		return message
	case code != "":
		return code
	default:
		return normalizeGrokProbeErrorText(cloudflareutil.TruncateBody(body, grokProbeErrorPreviewLimit))
	}
}

func normalizeGrokProbeErrorText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return strings.Join(strings.Fields(raw), " ")
}
