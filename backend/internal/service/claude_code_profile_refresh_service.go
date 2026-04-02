package service

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

type ClaudeCodeProfileSyncService struct {
	cfg      *config.Config
	client   *http.Client
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

type npmLatestPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Dist    struct {
		Tarball string `json:"tarball"`
	} `json:"dist"`
}

var (
	reBetaToken   = regexp.MustCompile(`(?:oauth|claude-code|interleaved-thinking|fine-grained-tool-streaming|token-counting|context-1m|fast-mode|redact-thinking|token-efficient-tools|effort|task-budgets|prompt-caching-scope|structured-outputs|web-search|advanced-tool-use|tool-search-tool|summarize-connector-text|afk-mode|cli-internal|advisor-tool|mcp-servers|files-api|environments|ccr-byoc)-\d{4}-\d{2}-\d{2}|claude-code-\d{8}`)
	reXAppLiteral = regexp.MustCompile(`["']x-app["']\s*:\s*["']([^"']+)["']`)
	rePackageName = regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)

	// Dynamic system prompt extraction: match strings starting with known
	// prefixes that can evolve across versions, rather than hardcoding full text.
	// NOTE: Longest prefix MUST come first — Go regexp uses leftmost-first matching,
	// so "You are Claude Code...running within" must precede "You are Claude Code...".
	reSystemPrompt = regexp.MustCompile(`"(You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK[^"]{0,300})"|"(You are Claude Code, Anthropic's official CLI for Claude[^"]{0,300})"|"(You are a Claude agent, built on Anthropic's Claude Agent SDK[^"]{0,300})"`)

	// Attribution fingerprint extraction patterns.
	reAttributionAnchor  = regexp.MustCompile(`cc_entrypoint`)
	reAttributionHexSalt = regexp.MustCompile(`["']([0-9a-f]{8,24})["']`)
	reAttributionIntArr  = regexp.MustCompile(`\[\s*(\d{1,3}(?:\s*,\s*\d{1,3}){1,9})\s*\]`)
)

func NewClaudeCodeProfileSyncService(cfg *config.Config) *ClaudeCodeProfileSyncService {
	return &ClaudeCodeProfileSyncService{
		cfg:    cfg,
		client: buildClaudeCodeProfileHTTPClient(cfg),
		stopCh: make(chan struct{}),
	}
}

func (s *ClaudeCodeProfileSyncService) Initialize() error {
	conf := s.syncConfig()
	if !conf.Enabled {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.RequestTimeoutSeconds)*time.Second)
	defer cancel()
	return s.refreshOnce(ctx)
}

func (s *ClaudeCodeProfileSyncService) Start() {
	conf := s.syncConfig()
	if !conf.Enabled || conf.RefreshIntervalHours <= 0 {
		return
	}
	interval := time.Duration(conf.RefreshIntervalHours) * time.Hour
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.RequestTimeoutSeconds)*time.Second)
				err := s.refreshOnce(ctx)
				cancel()
				if err != nil {
					logger.LegacyPrintf("service.claude_profile", "Warning: periodic sync failed: %v", err)
				}
			case <-s.stopCh:
				return
			}
		}
	}()
	logger.LegacyPrintf("service.claude_profile", "Claude Code profile sync started (every %dh)", conf.RefreshIntervalHours)
}

func (s *ClaudeCodeProfileSyncService) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func ProvideClaudeCodeProfileSyncService(cfg *config.Config) *ClaudeCodeProfileSyncService {
	svc := NewClaudeCodeProfileSyncService(cfg)
	if err := svc.Initialize(); err != nil {
		logger.LegacyPrintf("service.claude_profile", "Warning: startup sync failed: %v", err)
	}
	svc.Start()
	return svc
}

func (s *ClaudeCodeProfileSyncService) refreshOnce(ctx context.Context) error {
	conf := s.syncConfig()
	latest, err := s.fetchLatestPackage(ctx, conf)
	if err != nil {
		return err
	}
	profile, err := s.fetchAndBuildProfile(ctx, latest)
	if err != nil {
		return err
	}
	current := claude.CurrentMimicProfile()
	if current.PackageVersion == profile.PackageVersion && current.Source == profile.Source {
		return nil
	}
	claude.ApplyMimicProfile(profile)
	logger.LegacyPrintf("service.claude_profile", "Applied Claude Code mimic profile from npm: version=%s", profile.PackageVersion)
	return nil
}

func (s *ClaudeCodeProfileSyncService) fetchLatestPackage(ctx context.Context, conf resolvedClaudeCodeSyncConfig) (*npmLatestPackage, error) {
	base := strings.TrimRight(conf.RegistryURL, "/")
	endpoint := base + "/" + url.PathEscape(conf.PackageName) + "/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("registry returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var latest npmLatestPackage
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return nil, err
	}
	if latest.Version == "" || latest.Dist.Tarball == "" {
		return nil, fmt.Errorf("registry payload missing version or tarball")
	}
	return &latest, nil
}

// maxTarballSize limits the decompressed tarball to 64 MiB to prevent OOM from
// malicious gzip bombs or oversized packages.
const maxTarballSize = 64 << 20

func (s *ClaudeCodeProfileSyncService) fetchAndBuildProfile(ctx context.Context, latest *npmLatestPackage) (claude.MimicProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latest.Dist.Tarball, nil)
	if err != nil {
		return claude.MimicProfile{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return claude.MimicProfile{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return claude.MimicProfile{}, fmt.Errorf("tarball returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	// Wrap with a size limit to guard against gzip bombs.
	limitedBody := io.LimitReader(resp.Body, maxTarballSize)
	files, err := extractClaudeCodePackageFiles(limitedBody)
	if err != nil {
		return claude.MimicProfile{}, err
	}
	// Build profile, then clear extracted content so large strings are GC-eligible.
	profile, buildErr := buildClaudeCodeProfile(latest, files)
	for k := range files {
		delete(files, k)
	}
	return profile, buildErr
}

type resolvedClaudeCodeSyncConfig struct {
	Enabled               bool
	RegistryURL           string
	PackageName           string
	RequestTimeoutSeconds int
	RefreshIntervalHours  int
}

func (s *ClaudeCodeProfileSyncService) syncConfig() resolvedClaudeCodeSyncConfig {
	conf := resolvedClaudeCodeSyncConfig{
		Enabled:               true,
		RegistryURL:           "https://registry.npmjs.org",
		PackageName:           "@anthropic-ai/claude-code",
		RequestTimeoutSeconds: 20,
		RefreshIntervalHours:  12,
	}
	if s == nil || s.cfg == nil {
		return conf
	}
	cfg := s.cfg.Gateway.ClaudeCodeSync
	if cfg.RegistryURL != "" {
		conf.RegistryURL = cfg.RegistryURL
	}
	if cfg.PackageName != "" {
		conf.PackageName = cfg.PackageName
	}
	if cfg.RequestTimeoutSeconds > 0 {
		conf.RequestTimeoutSeconds = cfg.RequestTimeoutSeconds
	}
	if cfg.RefreshIntervalHours >= 0 {
		conf.RefreshIntervalHours = cfg.RefreshIntervalHours
	}
	if cfg.Enabled != nil {
		conf.Enabled = *cfg.Enabled
	}
	return conf
}

func buildClaudeCodeProfileHTTPClient(cfg *config.Config) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	timeout := 20 * time.Second
	if cfg != nil && strings.TrimSpace(cfg.Update.ProxyURL) != "" {
		if proxyURL, err := url.Parse(strings.TrimSpace(cfg.Update.ProxyURL)); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	if cfg != nil && cfg.Gateway.ClaudeCodeSync.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(cfg.Gateway.ClaudeCodeSync.RequestTimeoutSeconds) * time.Second
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// maxExtractedFileSize limits any single file extracted from the tarball to 32 MiB.
const maxExtractedFileSize = 32 << 20

func extractClaudeCodePackageFiles(r io.Reader) (map[string]string, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)
	targets := map[string]struct{}{
		"package/cli.js":       {},
		"package/package.json": {},
	}
	files := make(map[string]string, len(targets))
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, ok := targets[header.Name]; !ok {
			continue
		}
		if header.Size > maxExtractedFileSize {
			return nil, fmt.Errorf("file %s too large: %d bytes (limit %d)", header.Name, header.Size, maxExtractedFileSize)
		}
		content, readErr := io.ReadAll(io.LimitReader(tarReader, maxExtractedFileSize))
		if readErr != nil {
			return nil, readErr
		}
		files[header.Name] = string(content)
	}
	if files["package/cli.js"] == "" {
		return nil, fmt.Errorf("required file missing from tarball: package/cli.js")
	}
	return files, nil
}

func buildClaudeCodeProfile(latest *npmLatestPackage, files map[string]string) (claude.MimicProfile, error) {
	cliBundle := files["package/cli.js"]
	if strings.TrimSpace(cliBundle) == "" {
		return claude.MimicProfile{}, fmt.Errorf("cli bundle is empty")
	}

	profile := claude.BuiltinMimicProfile()
	profile.Source = "npm:" + latest.Version
	profile.PackageName = resolveClaudeCodePackageName(latest.Name, files["package/package.json"])
	profile.PackageVersion = latest.Version
	profile.UserAgent = "claude-cli/" + latest.Version + " (external, cli)"
	profile.OAuthBeta = findBetaToken(cliBundle, "oauth-")
	profile.ClaudeCodeBeta = findBetaToken(cliBundle, "claude-code-")
	profile.InterleavedThinkingBeta = findBetaToken(cliBundle, "interleaved-thinking-")
	profile.FineGrainedToolStreamingBeta = findBetaToken(cliBundle, "fine-grained-tool-streaming-")
	profile.TokenCountingBeta = findBetaToken(cliBundle, "token-counting-")
	profile.Context1MBeta = findBetaToken(cliBundle, "context-1m-")
	profile.FastModeBeta = findBetaToken(cliBundle, "fast-mode-")
	profile.SystemPromptPrefixes = extractSystemPromptPrefixes(cliBundle)
	if len(profile.SystemPromptPrefixes) == 0 {
		return claude.MimicProfile{}, fmt.Errorf("failed to extract Claude Code system prompts from cli bundle")
	}
	profile.SystemPrompt = profile.SystemPromptPrefixes[0]
	profile.XApp = parseXApp(cliBundle)
	if profile.OAuthBeta == "" || profile.ClaudeCodeBeta == "" || profile.InterleavedThinkingBeta == "" || profile.XApp == "" {
		return claude.MimicProfile{}, fmt.Errorf("failed to extract required Claude Code runtime traits from cli bundle")
	}

	// Extract attribution fingerprint salt and indices from the CLI bundle.
	if salt, indices, ok := extractAttributionParams(cliBundle); ok {
		profile.AttributionSalt = salt
		profile.AttributionIndices = indices
	}

	// Extract SDK package version (X-Stainless-Package-Version) from the bundled SDK.
	if sdkVersion := extractSDKVersion(cliBundle); sdkVersion != "" {
		profile.SDKVersion = sdkVersion
	}

	defaultHeaders := claude.DefaultHeaderSet()
	defaultHeaders["User-Agent"] = profile.UserAgent
	if profile.SDKVersion != "" {
		defaultHeaders["X-Stainless-Package-Version"] = profile.SDKVersion
	}
	profile.DefaultHeaders = defaultHeaders

	stableHeaders := claude.StableHeaders()
	if profile.XApp != "" {
		stableHeaders["X-App"] = profile.XApp
	}
	profile.StableDefaultHeaders = stableHeaders
	return profile, nil
}

func parseXApp(source string) string {
	match := reXAppLiteral.FindStringSubmatch(source)
	if len(match) >= 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func findBetaToken(source, prefix string) string {
	matches := reBetaToken.FindAllString(source, -1)
	var token string
	for _, match := range matches {
		candidate := strings.TrimSpace(match)
		if !strings.HasPrefix(candidate, prefix) {
			continue
		}
		if token == "" {
			token = candidate
			continue
		}
		if token != candidate {
			return ""
		}
	}
	return token
}

func extractSystemPromptPrefixes(source string) []string {
	matches := reSystemPrompt.FindAllStringSubmatch(source, -1)
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		// Each submatch group corresponds to one alternative in the regex.
		for _, group := range m[1:] {
			text := strings.TrimSpace(group)
			if text == "" {
				continue
			}
			if _, exists := seen[text]; exists {
				continue
			}
			seen[text] = struct{}{}
			out = append(out, text)
		}
	}
	return orderSystemPromptPrefixes(out)
}

func orderSystemPromptPrefixes(prompts []string) []string {
	if len(prompts) == 0 {
		return nil
	}

	preferredPrefixes := []string{
		"You are Claude Code, Anthropic's official CLI for Claude.",
		"You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK.",
		"You are a Claude agent, built on Anthropic's Claude Agent SDK.",
	}

	ordered := make([]string, 0, len(prompts))
	used := make([]bool, len(prompts))
	for _, preferred := range preferredPrefixes {
		for i, prompt := range prompts {
			if used[i] || !strings.HasPrefix(prompt, preferred) {
				continue
			}
			ordered = append(ordered, prompt)
			used[i] = true
		}
	}
	for i, prompt := range prompts {
		if used[i] {
			continue
		}
		ordered = append(ordered, prompt)
	}
	return ordered
}

func resolveClaudeCodePackageName(registryName, packageJSON string) string {
	if strings.TrimSpace(registryName) != "" {
		return strings.TrimSpace(registryName)
	}
	match := rePackageName.FindStringSubmatch(packageJSON)
	if len(match) >= 2 {
		return strings.TrimSpace(match[1])
	}
	return "@anthropic-ai/claude-code"
}

// extractAttributionParams extracts the attribution fingerprint salt and character
// indices from the CLI bundle by locating the "cc_entrypoint" anchor and then
// scanning backward for the hex salt and integer array.
func extractAttributionParams(source string) (salt string, indices []int, ok bool) {
	anchorLoc := reAttributionAnchor.FindStringIndex(source)
	if anchorLoc == nil {
		return "", nil, false
	}
	// Search in a window before the anchor (up to 4KB).
	windowStart := anchorLoc[0] - 4096
	if windowStart < 0 {
		windowStart = 0
	}
	window := source[windowStart:anchorLoc[1]]

	saltMatch := reAttributionHexSalt.FindAllStringSubmatch(window, -1)
	if len(saltMatch) == 0 {
		return "", nil, false
	}
	// Use the last match closest to the anchor.
	salt = saltMatch[len(saltMatch)-1][1]

	arrMatch := reAttributionIntArr.FindAllStringSubmatch(window, -1)
	if len(arrMatch) == 0 {
		return "", nil, false
	}
	arrStr := arrMatch[len(arrMatch)-1][1]
	parts := strings.Split(arrStr, ",")
	indices = make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return "", nil, false
		}
		indices = append(indices, v)
	}
	if len(indices) == 0 {
		return "", nil, false
	}
	return salt, indices, true
}

// extractSDKVersion extracts the @anthropic-ai/sdk package version from the CLI
// bundle. The bundled SDK exports a VERSION constant near stainless platform
// detection code. We anchor on "X-Stainless-Package-Version" to avoid matching
// unrelated VERSION constants (Node.js, other libraries).
var reSDKVersion = regexp.MustCompile(`VERSION\s*=\s*["'](\d+\.\d+\.\d+)["']`)

func extractSDKVersion(source string) string {
	const anchor = "X-Stainless-Package-Version"
	idx := strings.Index(source, anchor)
	if idx < 0 {
		return ""
	}
	// Search within 8KB before the anchor for the VERSION constant.
	windowStart := idx - 8192
	if windowStart < 0 {
		windowStart = 0
	}
	window := source[windowStart : idx+len(anchor)]
	matches := reSDKVersion.FindAllStringSubmatch(window, -1)
	if len(matches) == 0 {
		return ""
	}
	// Use the last match closest to the anchor.
	return matches[len(matches)-1][1]
}
