package service

import (
	"net/http"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/stretchr/testify/require"
)

func withClaudeMimicProfile(t *testing.T, profile claude.MimicProfile) {
	t.Helper()

	previous := claude.CurrentMimicProfile()
	claude.ApplyMimicProfile(profile)
	t.Cleanup(func() {
		claude.ApplyMimicProfile(previous)
	})
}

func TestCurrentClaudeCodeSystemPrompt_UsesActiveProfilePrompt(t *testing.T) {
	profile := claude.CurrentMimicProfile()
	profile.SystemPrompt = "You are a custom Claude Code prompt."
	withClaudeMimicProfile(t, profile)

	require.Equal(t, "You are a custom Claude Code prompt.", currentClaudeCodeSystemPrompt())
}

func TestCurrentClaudeCodePromptPrefixes_AppendsFallbackPrefixes(t *testing.T) {
	profile := claude.CurrentMimicProfile()
	profile.SystemPromptPrefixes = []string{"Custom Prefix"}
	withClaudeMimicProfile(t, profile)

	got := currentClaudeCodePromptPrefixes()

	require.Equal(t, []string{
		"Custom Prefix",
		"You are a file search specialist for Claude Code",
		"You are a helpful AI assistant tasked with summarizing conversations",
	}, got)
}

func TestApplyClaudeCodeMimicHeaders_SetsStableDefaultsAndStreamingMarker(t *testing.T) {
	profile := claude.CurrentMimicProfile()
	profile.XApp = "cli-test"
	profile.StableDefaultHeaders = map[string]string{
		"X-App":                   "cli-test",
		"X-Stainless-Timeout":     "900",
		"X-Stainless-Retry-Count": "2",
	}
	withClaudeMimicProfile(t, profile)

	req, err := http.NewRequest(http.MethodPost, "https://example.com/v1/messages", nil)
	require.NoError(t, err)

	applyClaudeCodeMimicHeaders(req, true)

	require.Equal(t, "application/json", getHeaderRaw(req.Header, "accept"))
	require.Equal(t, "cli-test", getHeaderRaw(req.Header, "x-app"))
	require.Equal(t, "900", getHeaderRaw(req.Header, "x-stainless-timeout"))
	require.Equal(t, "2", getHeaderRaw(req.Header, "x-stainless-retry-count"))
	require.Equal(t, "stream", getHeaderRaw(req.Header, "x-stainless-helper-method"))
	require.NotEmpty(t, getHeaderRaw(req.Header, "x-client-request-id"))
}

func TestApplyClaudeCodeMimicHeaders_PreservesExistingClientRequestID(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://example.com/v1/messages", nil)
	require.NoError(t, err)
	setHeaderRaw(req.Header, "x-client-request-id", "existing-id")

	applyClaudeCodeMimicHeaders(req, false)

	require.Equal(t, "existing-id", getHeaderRaw(req.Header, "x-client-request-id"))
	require.Empty(t, getHeaderRaw(req.Header, "x-stainless-helper-method"))
}
