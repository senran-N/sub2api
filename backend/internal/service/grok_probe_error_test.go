//go:build unit

package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrokSummarizeProbeHTTPError_CloudflareChallengeSanitized(t *testing.T) {
	resp := newJSONResponse(http.StatusForbidden, `<!DOCTYPE html><html><head><title>Just a moment...</title></head><body>Enable JavaScript and cookies to continue<script>window._cf_chl_opt={};</script></body></html>`)
	resp.Header.Set("content-type", "text/html; charset=UTF-8")
	resp.Header.Set("cf-ray", "test-ray-123")

	summary := grokSummarizeProbeHTTPError(resp, grokReadProbeErrorBody(resp))

	require.True(t, summary.IsCloudflareChallenge)
	require.Equal(t, "Cloudflare challenge encountered (HTTP 403)", summary.Message)
	require.NotContains(t, summary.Message, "<!DOCTYPE html>")
	require.Empty(t, summary.AuthenticationMessage)
}

func TestGrokSummarizeProbeHTTPError_JSONBodyUsesCompactDetail(t *testing.T) {
	resp := newJSONResponse(http.StatusUnauthorized, `{"error":{"code":"invalid_credentials","message":"session expired"}}`)

	summary := grokSummarizeProbeHTTPError(resp, grokReadProbeErrorBody(resp))

	require.False(t, summary.IsCloudflareChallenge)
	require.Equal(t, "API returned 401: session expired (invalid_credentials)", summary.Message)
	require.Equal(t, "Authentication failed (401): session expired (invalid_credentials)", summary.AuthenticationMessage)
}
