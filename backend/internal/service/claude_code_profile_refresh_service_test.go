package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/stretchr/testify/require"
)

func TestClaudeCodeProfileSyncService_RefreshOnceUpdatesRuntimeProfile(t *testing.T) {
	previous := claude.CurrentMimicProfile()
	t.Cleanup(func() {
		claude.ApplyMimicProfile(previous)
	})

	tarball := buildClaudeCodeTarballForTest(t, map[string]string{
		"package/cli.js": `var oauth="oauth-2099-01-01",cc="claude-code-20990101",thinking="interleaved-thinking-2099-01-01",count="token-counting-2099-01-01",ctx="context-1m-2099-01-01",fast="fast-mode-2099-01-01";
var headers={"x-app":"cli-test"};
var base="You are Claude Code, Anthropic's official CLI for Claude.";
var sdk="You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK.";
var agent="You are a Claude agent, built on Anthropic's Claude Agent SDK.";`,
		"package/package.json": `{"name":"@anthropic-ai/claude-code"}`,
	})

	tarballServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(tarball)
	}))
	defer tarballServer.Close()

	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"name":"@anthropic-ai/claude-code","version":"9.9.9","dist":{"tarball":"` + tarballServer.URL + `/claude.tgz"}}`))
	}))
	defer registryServer.Close()

	cfg := &config.Config{}
	enabled := true
	cfg.Gateway.ClaudeCodeSync.Enabled = &enabled
	cfg.Gateway.ClaudeCodeSync.RegistryURL = registryServer.URL
	cfg.Gateway.ClaudeCodeSync.PackageName = "@anthropic-ai/claude-code"
	cfg.Gateway.ClaudeCodeSync.RequestTimeoutSeconds = 5

	svc := NewClaudeCodeProfileSyncService(cfg)
	err := svc.refreshOnce(context.Background())
	require.NoError(t, err)

	profile := claude.CurrentMimicProfile()
	require.Equal(t, "npm:9.9.9", profile.Source)
	require.Equal(t, "9.9.9", profile.PackageVersion)
	require.Equal(t, "claude-cli/9.9.9 (external, cli)", claude.DefaultUserAgent())
	require.Equal(t, "oauth-2099-01-01", claude.OAuthBetaToken())
	require.Equal(t, "claude-code-20990101", claude.ClaudeCodeBetaToken())
	require.Equal(t, "interleaved-thinking-2099-01-01", claude.InterleavedThinkingBetaToken())
	require.Equal(t, "", claude.FineGrainedToolStreamingBetaToken())
	require.NotContains(t, claude.DefaultAnthropicBetaHeader(), "fine-grained-tool-streaming")
	require.Equal(t, "cli-test", claude.StableHeaders()["X-App"])
	require.Equal(t, "You are Claude Code, Anthropic's official CLI for Claude.", claude.SystemPromptText())
}

func buildClaudeCodeTarballForTest(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var archive bytes.Buffer
	gzWriter := gzip.NewWriter(&archive)
	tarWriter := tar.NewWriter(gzWriter)

	for name, content := range files {
		require.NoError(t, tarWriter.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(content)),
		}))
		_, err := tarWriter.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzWriter.Close())
	return archive.Bytes()
}
