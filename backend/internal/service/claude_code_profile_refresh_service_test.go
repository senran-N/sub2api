package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/claude"
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

func TestClaudeCodeProfileSyncService_RefreshOnceSupportsWrapperAndNativePackages(t *testing.T) {
	previous := claude.CurrentMimicProfile()
	t.Cleanup(func() {
		claude.ApplyMimicProfile(previous)
	})

	wrapperTarball := buildClaudeCodeTarballForTest(t, map[string]string{
		"package/cli-wrapper.cjs": `const WRAPPER_NAME=require('./package.json').name;`,
		"package/install.cjs":     `const PACKAGE_PREFIX='@anthropic-ai/claude-code';`,
		"package/package.json":    `{"name":"@anthropic-ai/claude-code","version":"9.9.9","optionalDependencies":{"@anthropic-ai/claude-code-linux-x64":"9.9.9"}}`,
	})
	nativeTarball := buildClaudeCodeTarballForTest(t, map[string]string{
		"package/claude": pseudoBinaryFixtureForTest(
			`var oauth="oauth-2099-01-01",cc="claude-code-20990101",thinking="interleaved-thinking-2099-01-01";`,
			`var headers={"x-app":"cli-native"};`,
			`var base="You are Claude Code, Anthropic's official CLI for Claude.";`,
			`var sdk="You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK.";`,
			`const VERSION="0.88.0";var stainlessHeader="X-Stainless-Package-Version";`,
			`x-anthropic-billing-header: cc_version=2.1.114.abc; cc_entrypoint=cli`,
			`["59cf53e54c78",[4,7,20]]`,
		),
		"package/package.json": `{"name":"@anthropic-ai/claude-code-linux-x64","version":"9.9.9"}`,
	})

	tarballServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		switch r.URL.Path {
		case "/wrapper.tgz":
			_, _ = w.Write(wrapperTarball)
		case "/native.tgz":
			_, _ = w.Write(nativeTarball)
		default:
			http.NotFound(w, r)
		}
	}))
	defer tarballServer.Close()

	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "claude-code/latest"):
			_, _ = w.Write([]byte(`{"name":"@anthropic-ai/claude-code","version":"9.9.9","dist":{"tarball":"` + tarballServer.URL + `/wrapper.tgz"}}`))
		case strings.Contains(r.URL.Path, "claude-code-linux-x64/9.9.9"):
			_, _ = w.Write([]byte(`{"name":"@anthropic-ai/claude-code-linux-x64","version":"9.9.9","dist":{"tarball":"` + tarballServer.URL + `/native.tgz"}}`))
		default:
			http.NotFound(w, r)
		}
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
	require.Equal(t, "9.9.9", profile.PackageVersion)
	require.Equal(t, "cli-native", profile.XApp)
	require.Equal(t, "oauth-2099-01-01", profile.OAuthBeta)
	require.Equal(t, "claude-code-20990101", profile.ClaudeCodeBeta)
	require.Equal(t, "interleaved-thinking-2099-01-01", profile.InterleavedThinkingBeta)
	require.Equal(t, "0.88.0", profile.SDKVersion)
	require.Equal(t, "59cf53e54c78", profile.AttributionSalt)
	require.Equal(t, []int{4, 7, 20}, profile.AttributionIndices)
}

func TestClaudeCodeProfileSyncService_RefreshOnceUsesNativeOptionalDependencyVersion(t *testing.T) {
	previous := claude.CurrentMimicProfile()
	t.Cleanup(func() {
		claude.ApplyMimicProfile(previous)
	})

	wrapperTarball := buildClaudeCodeTarballForTest(t, map[string]string{
		"package/cli-wrapper.cjs": `const WRAPPER_NAME=require('./package.json').name;`,
		"package/package.json":    `{"name":"@anthropic-ai/claude-code","version":"9.9.9","optionalDependencies":{"@anthropic-ai/claude-code-linux-x64":"9.9.8"}}`,
	})
	nativeTarball := buildClaudeCodeTarballForTest(t, map[string]string{
		"package/claude": pseudoBinaryFixtureForTest(
			`var oauth="oauth-2099-01-01",cc="claude-code-20990101",thinking="interleaved-thinking-2099-01-01";`,
			`var headers={"x-app":"cli-native-versioned"};`,
			`var base="You are Claude Code, Anthropic's official CLI for Claude.";`,
		),
		"package/package.json": `{"name":"@anthropic-ai/claude-code-linux-x64","version":"9.9.8"}`,
	})

	tarballServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		switch r.URL.Path {
		case "/wrapper.tgz":
			_, _ = w.Write(wrapperTarball)
		case "/native-9.9.8.tgz":
			_, _ = w.Write(nativeTarball)
		default:
			http.NotFound(w, r)
		}
	}))
	defer tarballServer.Close()

	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "claude-code/latest"):
			_, _ = w.Write([]byte(`{"name":"@anthropic-ai/claude-code","version":"9.9.9","dist":{"tarball":"` + tarballServer.URL + `/wrapper.tgz"}}`))
		case strings.Contains(r.URL.Path, "claude-code-linux-x64/9.9.8"):
			_, _ = w.Write([]byte(`{"name":"@anthropic-ai/claude-code-linux-x64","version":"9.9.8","dist":{"tarball":"` + tarballServer.URL + `/native-9.9.8.tgz"}}`))
		default:
			http.NotFound(w, r)
		}
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
	require.Equal(t, "9.9.9", profile.PackageVersion)
	require.Equal(t, "cli-native-versioned", profile.XApp)
	require.Equal(t, "oauth-2099-01-01", profile.OAuthBeta)
	require.Equal(t, "claude-code-20990101", profile.ClaudeCodeBeta)
}

func TestClaudeCodeProfileSyncService_RefreshOnceInfersNativePackagesFromWrapperSource(t *testing.T) {
	previous := claude.CurrentMimicProfile()
	t.Cleanup(func() {
		claude.ApplyMimicProfile(previous)
	})

	wrapperTarball := buildClaudeCodeTarballForTest(t, map[string]string{
		"package/install.cjs": `const PACKAGE_PREFIX='@anthropic-ai/claude-code';
const PLATFORMS={'linux-x64':{pkg: PACKAGE_PREFIX + '-linux-x64'}};`,
		"package/package.json": `{"name":"@anthropic-ai/claude-code","version":"9.9.9"}`,
	})
	nativeTarball := buildClaudeCodeTarballForTest(t, map[string]string{
		"package/claude": pseudoBinaryFixtureForTest(
			`var oauth="oauth-2099-01-01",cc="claude-code-20990101",thinking="interleaved-thinking-2099-01-01";`,
			`var headers={"x-app":"cli-inferred-native"};`,
			`var base="You are Claude Code, Anthropic's official CLI for Claude.";`,
		),
		"package/package.json": `{"name":"@anthropic-ai/claude-code-linux-x64","version":"9.9.9"}`,
	})

	tarballServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		switch r.URL.Path {
		case "/wrapper.tgz":
			_, _ = w.Write(wrapperTarball)
		case "/native.tgz":
			_, _ = w.Write(nativeTarball)
		default:
			http.NotFound(w, r)
		}
	}))
	defer tarballServer.Close()

	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "claude-code/latest"):
			_, _ = w.Write([]byte(`{"name":"@anthropic-ai/claude-code","version":"9.9.9","dist":{"tarball":"` + tarballServer.URL + `/wrapper.tgz"}}`))
		case strings.Contains(r.URL.Path, "claude-code-linux-x64/9.9.9"):
			_, _ = w.Write([]byte(`{"name":"@anthropic-ai/claude-code-linux-x64","version":"9.9.9","dist":{"tarball":"` + tarballServer.URL + `/native.tgz"}}`))
		default:
			http.NotFound(w, r)
		}
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
	require.Equal(t, "9.9.9", profile.PackageVersion)
	require.Equal(t, "cli-inferred-native", profile.XApp)
	require.Equal(t, "oauth-2099-01-01", profile.OAuthBeta)
	require.Equal(t, "claude-code-20990101", profile.ClaudeCodeBeta)
}

func TestBuildClaudeCodeProfile_UsesBuiltinFallbackForOptionalFields(t *testing.T) {
	previous := claude.CurrentMimicProfile()
	t.Cleanup(func() {
		claude.ApplyMimicProfile(previous)
	})

	claude.ApplyMimicProfile(claude.MimicProfile{
		Source:                  "test-current",
		PackageName:             "@anthropic-ai/claude-code",
		PackageVersion:          "99.99.99",
		UserAgent:               "claude-cli/99.99.99 (external, cli)",
		XApp:                    "current-x-app",
		OAuthBeta:               "oauth-2099-12-31",
		ClaudeCodeBeta:          "claude-code-20991231",
		InterleavedThinkingBeta: "interleaved-thinking-2099-12-31",
		SystemPrompt:            "You are Claude Code, Anthropic's official CLI for Claude.",
		SystemPromptPrefixes: []string{
			"You are Claude Code, Anthropic's official CLI for Claude.",
		},
		SDKVersion:         "9.9.9",
		AttributionSalt:    "deadbeef1234",
		AttributionIndices: []int{1, 2, 3},
	})

	profile, err := buildClaudeCodeProfile(&npmLatestPackage{
		Name:    "@anthropic-ai/claude-code",
		Version: "3.2.1",
	}, mustExtractClaudeCodePackageAssetsForTest(t, map[string]string{
		"package/cli.js": `var oauth="oauth-2030-01-01",cc="claude-code-20300101",thinking="interleaved-thinking-2030-01-01";
var headers={"x-app":"cli-fresh"};
var base="You are Claude Code, Anthropic's official CLI for Claude.";`,
		"package/package.json": `{"name":"@anthropic-ai/claude-code"}`,
	}))
	require.NoError(t, err)

	builtin := claude.BuiltinMimicProfile()
	require.Equal(t, builtin.SDKVersion, profile.SDKVersion)
	require.Equal(t, builtin.AttributionSalt, profile.AttributionSalt)
	require.Equal(t, builtin.AttributionIndices, profile.AttributionIndices)
	require.Equal(t, builtin.SDKVersion, profile.DefaultHeaders["X-Stainless-Package-Version"])
	require.Equal(t, "claude-cli/3.2.1 (external, cli)", profile.UserAgent)
}

func TestBuildClaudeCodeProfile_PrefersCanonicalSystemPrompt(t *testing.T) {
	profile, err := buildClaudeCodeProfile(&npmLatestPackage{
		Name:    "@anthropic-ai/claude-code",
		Version: "3.2.1",
	}, mustExtractClaudeCodePackageAssetsForTest(t, map[string]string{
		"package/cli.js": `var oauth="oauth-2030-01-01",cc="claude-code-20300101",thinking="interleaved-thinking-2030-01-01";
var headers={"x-app":"cli-fresh"};
var sdk="You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK.";
var agent="You are a Claude agent, built on Anthropic's Claude Agent SDK.";
var base="You are Claude Code, Anthropic's official CLI for Claude.";`,
		"package/package.json": `{"name":"@anthropic-ai/claude-code"}`,
	}))
	require.NoError(t, err)

	require.Equal(t, "You are Claude Code, Anthropic's official CLI for Claude.", profile.SystemPrompt)
	require.Equal(t, []string{
		"You are Claude Code, Anthropic's official CLI for Claude.",
		"You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK.",
		"You are a Claude agent, built on Anthropic's Claude Agent SDK.",
	}, profile.SystemPromptPrefixes)
}

func TestBuildClaudeCodeProfile_PrefersNewestBetaTokenVersions(t *testing.T) {
	profile, err := buildClaudeCodeProfile(&npmLatestPackage{
		Name:    "@anthropic-ai/claude-code",
		Version: "3.2.1",
	}, mustExtractClaudeCodePackageAssetsForTest(t, map[string]string{
		"package/cli.js": `var oldOauth="oauth-2025-04-20",newOauth="oauth-2026-03-13";
var oldCC="claude-code-20250219",newCC="claude-code-20260112";
var oldThinking="interleaved-thinking-2025-05-14",newThinking="interleaved-thinking-2026-02-01";
var oldFast="fast-mode-2026-01-01",newFast="fast-mode-2026-02-01";
var headers={"x-app":"cli-fresh"};
var base="You are Claude Code, Anthropic's official CLI for Claude.";`,
		"package/package.json": `{"name":"@anthropic-ai/claude-code"}`,
	}))
	require.NoError(t, err)

	require.Equal(t, "oauth-2026-03-13", profile.OAuthBeta)
	require.Equal(t, "claude-code-20260112", profile.ClaudeCodeBeta)
	require.Equal(t, "interleaved-thinking-2026-02-01", profile.InterleavedThinkingBeta)
	require.Equal(t, "fast-mode-2026-02-01", profile.FastModeBeta)
}

func TestFindBetaToken_PrefersNewestObservedTokenFormats(t *testing.T) {
	source := strings.Join([]string{
		`oauth-2025-04-20`,
		`oauth-2026-03-13`,
		`claude-code-20250219`,
		`claude-code-20260112`,
		`compact-2025-12-31`,
		`compact-2026-01-12`,
	}, " ")

	require.Equal(t, "oauth-2026-03-13", findBetaToken(source, "oauth-"))
	require.Equal(t, "claude-code-20260112", findBetaToken(source, "claude-code-"))
	require.Equal(t, "compact-2026-01-12", findBetaToken(source, "compact-"))
}

func TestExtractClaudeCodePackageAssets_CollectsWrapperCandidatesAndNativeSnippets(t *testing.T) {
	assets, err := extractClaudeCodePackageAssets(bytes.NewReader(buildClaudeCodeTarballForTest(t, map[string]string{
		"package/cli-wrapper.cjs": `const WRAPPER_NAME=require('./package.json').name;`,
		"package/package.json":    `{"name":"@anthropic-ai/claude-code","optionalDependencies":{"@anthropic-ai/claude-code-linux-x64":"9.9.9"}}`,
		"package/claude": pseudoBinaryFixtureForTest(
			`var oauth="oauth-2099-01-01";`,
			`var headers={"x-app":"cli-native"};`,
			`var base="You are Claude Code, Anthropic's official CLI for Claude.";`,
		),
	})))
	require.NoError(t, err)

	require.Equal(t, []claudeCodeNativePackageCandidate{{
		Name:    "@anthropic-ai/claude-code-linux-x64",
		Version: "9.9.9",
	}}, assets.NativePackageCandidates)
	require.NotEmpty(t, assets.Sources)
	require.Contains(t, combineClaudeCodeSources(assets.Sources), `oauth-2099-01-01`)
	require.Contains(t, combineClaudeCodeSources(assets.Sources), `"x-app":"cli-native"`)
}

func TestExtractClaudeCodePackageAssets_InfersNativeCandidatesFromWrapperSource(t *testing.T) {
	assets, err := extractClaudeCodePackageAssets(bytes.NewReader(buildClaudeCodeTarballForTest(t, map[string]string{
		"package/install.cjs": `const PACKAGE_PREFIX='@anthropic-ai/claude-code';
const PLATFORMS={
  'linux-arm64-musl':{pkg: PACKAGE_PREFIX + '-linux-arm64-musl'},
  'linux-x64':{pkg: PACKAGE_PREFIX + '-linux-x64'}
};`,
		"package/package.json": `{"name":"@anthropic-ai/claude-code"}`,
	})))
	require.NoError(t, err)

	require.Equal(t, []claudeCodeNativePackageCandidate{
		{Name: "@anthropic-ai/claude-code-linux-arm64-musl"},
		{Name: "@anthropic-ai/claude-code-linux-x64"},
	}, assets.NativePackageCandidates)
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

func mustExtractClaudeCodePackageAssetsForTest(t *testing.T, files map[string]string) claudeCodePackageAssets {
	t.Helper()

	assets, err := extractClaudeCodePackageAssets(bytes.NewReader(buildClaudeCodeTarballForTest(t, files)))
	require.NoError(t, err)
	return assets
}

func pseudoBinaryFixtureForTest(chunks ...string) string {
	return "\x00" + strings.Join(chunks, "\x00") + "\x00"
}
