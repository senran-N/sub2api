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
	"path"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

const attributionSearchWindowBytes = 4096
const nativeClaudeCodePackagePrefix = "@anthropic-ai/claude-code-"

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

type claudeCodePackageManifest struct {
	Name                 string            `json:"name"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

type claudeCodeSourceAsset struct {
	Path    string
	Content string
}

type claudeCodeNativePackageCandidate struct {
	Name    string
	Version string
}

type claudeCodePackageAssets struct {
	PackageJSON             string
	Manifest                claudeCodePackageManifest
	Sources                 []claudeCodeSourceAsset
	NativePackageCandidates []claudeCodeNativePackageCandidate
}

var (
	reBetaToken            = regexp.MustCompile(`(?:oauth|claude-code|interleaved-thinking|fine-grained-tool-streaming|token-counting|context-1m|fast-mode|redact-thinking|token-efficient-tools|effort|task-budgets|prompt-caching-scope|structured-outputs|web-search|advanced-tool-use|tool-search-tool|summarize-connector-text|afk-mode|cli-internal|advisor-tool|mcp-servers|files-api|environments|ccr-byoc|compact|managed-agents|skills)-\d{4}-\d{2}-\d{2}|claude-code-\d{8}`)
	reXAppLiteral          = regexp.MustCompile(`["']x-app["']\s*:\s*["']([^"']+)["']`)
	rePackageName          = regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
	reNativePackageLiteral = regexp.MustCompile(`@anthropic-ai/claude-code-(?:darwin|linux|win32)-(?:x64|arm64)(?:-musl)?`)
	reNativePackagePrefix  = regexp.MustCompile(`PACKAGE_PREFIX\s*=\s*["'](@anthropic-ai/claude-code)["']`)
	reNativePackageSuffix  = regexp.MustCompile(`PACKAGE_PREFIX\s*\+\s*["'](-(?:darwin|linux|win32)-(?:x64|arm64)(?:-musl)?)["']`)

	// Dynamic system prompt extraction: match strings starting with known
	// prefixes that can evolve across versions, rather than hardcoding full text.
	// NOTE: Longest prefix MUST come first — Go regexp uses leftmost-first matching,
	// so "You are Claude Code...running within" must precede "You are Claude Code...".
	reSystemPrompt = regexp.MustCompile(`"(You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK[^"]{0,300})"|"(You are Claude Code, Anthropic's official CLI for Claude[^"]{0,300})"|"(You are a Claude agent, built on Anthropic's Claude Agent SDK[^"]{0,300})"`)

	// Attribution fingerprint extraction patterns.
	reAttributionAnchor   = regexp.MustCompile(`cc_entrypoint`)
	reAttributionHexSalt  = regexp.MustCompile(`["']([0-9a-f]{8,24})["']`)
	reAttributionIntArr   = regexp.MustCompile(`\[\s*(\d{1,3}(?:\s*,\s*\d{1,3}){1,9})\s*\]`)
	reBinarySnippetAnchor = regexp.MustCompile(`cc_entrypoint|X-Stainless-Package-Version|You are Claude Code|You are a Claude agent|["']x-app["']`)
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
	profile, err := s.fetchAndBuildProfile(ctx, conf, latest)
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
	return s.fetchPackageRelease(ctx, conf, conf.PackageName, "")
}

func (s *ClaudeCodeProfileSyncService) fetchPackageRelease(ctx context.Context, conf resolvedClaudeCodeSyncConfig, packageName, version string) (*npmLatestPackage, error) {
	base := strings.TrimRight(conf.RegistryURL, "/")
	endpoint := base + "/" + url.PathEscape(packageName)
	if strings.TrimSpace(version) == "" {
		endpoint += "/latest"
	} else {
		endpoint += "/" + url.PathEscape(version)
	}
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

// maxTarballSize limits the compressed tarball body to 128 MiB.
// Claude Code native packages are already around 70 MiB compressed in 2.1.114,
// so the older 64 MiB limit would reject legitimate releases.
const maxTarballSize = 128 << 20

func (s *ClaudeCodeProfileSyncService) fetchAndBuildProfile(ctx context.Context, conf resolvedClaudeCodeSyncConfig, latest *npmLatestPackage) (claude.MimicProfile, error) {
	assets, err := s.fetchPackageAssets(ctx, latest)
	if err != nil {
		return claude.MimicProfile{}, err
	}
	if profile, buildErr := buildClaudeCodeProfile(latest, assets); buildErr == nil {
		return profile, nil
	} else if len(assets.NativePackageCandidates) == 0 {
		return claude.MimicProfile{}, buildErr
	}

	var lastErr error
	for _, candidate := range preferredNativePackageCandidates(assets.NativePackageCandidates) {
		nativeVersion := strings.TrimSpace(candidate.Version)
		if nativeVersion == "" {
			nativeVersion = latest.Version
		}
		nativePkg, err := s.fetchPackageRelease(ctx, conf, candidate.Name, nativeVersion)
		if err != nil {
			lastErr = fmt.Errorf("fetch native package metadata %s@%s: %w", candidate.Name, nativeVersion, err)
			continue
		}
		nativeAssets, err := s.fetchPackageAssets(ctx, nativePkg)
		if err != nil {
			lastErr = fmt.Errorf("extract native package %s@%s: %w", candidate.Name, nativePkg.Version, err)
			continue
		}
		profile, err := buildClaudeCodeProfile(latest, mergeClaudeCodePackageAssets(assets, nativeAssets))
		if err == nil {
			return profile, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return claude.MimicProfile{}, lastErr
	}
	return claude.MimicProfile{}, fmt.Errorf("failed to extract Claude Code runtime traits from npm package")
}

func (s *ClaudeCodeProfileSyncService) fetchPackageAssets(ctx context.Context, pkg *npmLatestPackage) (claudeCodePackageAssets, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pkg.Dist.Tarball, nil)
	if err != nil {
		return claudeCodePackageAssets{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return claudeCodePackageAssets{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return claudeCodePackageAssets{}, fmt.Errorf("tarball returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	// Wrap with a compressed-size limit so we can still process native packages
	// without holding their full binaries in memory.
	limitedBody := io.LimitReader(resp.Body, maxTarballSize)
	assets, err := extractClaudeCodePackageAssets(limitedBody)
	if err != nil {
		return claudeCodePackageAssets{}, err
	}
	return assets, nil
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
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok || baseTransport == nil {
		baseTransport = (&http.Transport{}).Clone()
	}
	transport := baseTransport.Clone()
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

// Per-file extraction limits keep wrapper sources small while still allowing one
// native Claude binary to be scanned in a streaming fashion.
const (
	maxPackageJSONSize         = 1 << 20
	maxTextSourceSize          = 8 << 20
	maxNativeBinarySize        = 256 << 20
	maxBinaryRunBufferBytes    = 128 << 10
	maxBinarySnippetBytes      = 2 << 20
	binarySnippetContextBytes  = 8 << 10
	binarySnippetAnchorContext = 8 << 10
	binarySnippetTokenContext  = 256
)

func extractClaudeCodePackageAssets(r io.Reader) (claudeCodePackageAssets, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return claudeCodePackageAssets{}, err
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)
	var assets claudeCodePackageAssets
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return claudeCodePackageAssets{}, err
		}
		switch {
		case header.Name == "package/package.json":
			content, readErr := readTarEntryString(tarReader, header.Size, maxPackageJSONSize)
			if readErr != nil {
				return claudeCodePackageAssets{}, readErr
			}
			assets.PackageJSON = content
			assets.Manifest = parseClaudeCodePackageManifest(content)
			assets.NativePackageCandidates = manifestNativePackageCandidates(assets.Manifest.OptionalDependencies)
		case isClaudeCodeTextSourcePath(header.Name):
			content, readErr := readTarEntryString(tarReader, header.Size, maxTextSourceSize)
			if readErr != nil {
				return claudeCodePackageAssets{}, readErr
			}
			if strings.TrimSpace(content) != "" {
				assets.Sources = append(assets.Sources, claudeCodeSourceAsset{Path: header.Name, Content: content})
			}
		case isClaudeCodeNativeBinaryPath(header.Name):
			content, readErr := extractRelevantTextFromBinary(tarReader, header.Size)
			if readErr != nil {
				return claudeCodePackageAssets{}, readErr
			}
			if strings.TrimSpace(content) != "" {
				assets.Sources = append(assets.Sources, claudeCodeSourceAsset{Path: header.Name, Content: content})
			}
		}
	}
	if strings.TrimSpace(assets.PackageJSON) == "" {
		return claudeCodePackageAssets{}, fmt.Errorf("required file missing from tarball: package/package.json")
	}
	assets.NativePackageCandidates = mergeManifestAndSourceNativePackageCandidates(
		assets.NativePackageCandidates,
		sourceNativePackageCandidates(assets.Sources),
	)
	if len(assets.Sources) == 0 && len(assets.NativePackageCandidates) == 0 {
		return claudeCodePackageAssets{}, fmt.Errorf("no Claude Code source assets found in tarball")
	}
	return assets, nil
}

func buildClaudeCodeProfile(latest *npmLatestPackage, assets claudeCodePackageAssets) (claude.MimicProfile, error) {
	combinedSource := combineClaudeCodeSources(assets.Sources)
	if strings.TrimSpace(combinedSource) == "" {
		return claude.MimicProfile{}, fmt.Errorf("claude code source corpus is empty")
	}

	profile := claude.BuiltinMimicProfile()
	profile.Source = "npm:" + latest.Version
	profile.PackageName = resolveClaudeCodePackageName(latest.Name, assets.PackageJSON)
	profile.PackageVersion = latest.Version
	profile.UserAgent = "claude-cli/" + latest.Version + " (external, cli)"
	profile.OAuthBeta = findBetaToken(combinedSource, "oauth-")
	profile.ClaudeCodeBeta = findBetaToken(combinedSource, "claude-code-")
	profile.InterleavedThinkingBeta = findBetaToken(combinedSource, "interleaved-thinking-")
	profile.FineGrainedToolStreamingBeta = findBetaToken(combinedSource, "fine-grained-tool-streaming-")
	profile.TokenCountingBeta = findBetaToken(combinedSource, "token-counting-")
	profile.Context1MBeta = findBetaToken(combinedSource, "context-1m-")
	profile.FastModeBeta = findBetaToken(combinedSource, "fast-mode-")
	profile.SystemPromptPrefixes = extractSystemPromptPrefixes(combinedSource)
	if len(profile.SystemPromptPrefixes) == 0 {
		return claude.MimicProfile{}, fmt.Errorf("failed to extract Claude Code system prompts from package assets")
	}
	profile.SystemPrompt = profile.SystemPromptPrefixes[0]
	profile.XApp = parseXApp(combinedSource)
	if profile.OAuthBeta == "" || profile.ClaudeCodeBeta == "" || profile.InterleavedThinkingBeta == "" || profile.XApp == "" {
		return claude.MimicProfile{}, fmt.Errorf("failed to extract required Claude Code runtime traits from package assets")
	}

	// Extract attribution fingerprint salt and indices from the extracted source corpus.
	if salt, indices, ok := extractAttributionParams(combinedSource); ok {
		profile.AttributionSalt = salt
		profile.AttributionIndices = indices
	}

	// Extract SDK package version (X-Stainless-Package-Version) from the extracted source corpus.
	if sdkVersion := extractSDKVersion(combinedSource); sdkVersion != "" {
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

func readTarEntryString(r io.Reader, size int64, limit int64) (string, error) {
	if size > limit {
		return "", fmt.Errorf("file too large: %d bytes (limit %d)", size, limit)
	}
	content, err := io.ReadAll(io.LimitReader(r, limit))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func parseClaudeCodePackageManifest(packageJSON string) claudeCodePackageManifest {
	var manifest claudeCodePackageManifest
	_ = json.Unmarshal([]byte(packageJSON), &manifest)
	return manifest
}

func manifestNativePackageCandidates(optionalDependencies map[string]string) []claudeCodeNativePackageCandidate {
	if len(optionalDependencies) == 0 {
		return nil
	}
	out := make([]claudeCodeNativePackageCandidate, 0, len(optionalDependencies))
	for packageName, version := range optionalDependencies {
		if strings.HasPrefix(packageName, nativeClaudeCodePackagePrefix) {
			out = append(out, claudeCodeNativePackageCandidate{
				Name:    packageName,
				Version: strings.TrimSpace(version),
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Version < out[j].Version
	})
	return out
}

func sourceNativePackageCandidates(sources []claudeCodeSourceAsset) []claudeCodeNativePackageCandidate {
	if len(sources) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]claudeCodeNativePackageCandidate, 0)
	for _, source := range sources {
		content := source.Content
		if strings.TrimSpace(content) == "" {
			continue
		}
		for _, match := range reNativePackageLiteral.FindAllString(content, -1) {
			name := strings.TrimSpace(match)
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, claudeCodeNativePackageCandidate{Name: name})
		}

		prefix := strings.TrimSuffix(nativeClaudeCodePackagePrefix, "-")
		if match := reNativePackagePrefix.FindStringSubmatch(content); len(match) >= 2 {
			if candidatePrefix := strings.TrimSpace(match[1]); candidatePrefix != "" {
				prefix = candidatePrefix
			}
		}
		for _, match := range reNativePackageSuffix.FindAllStringSubmatch(content, -1) {
			if len(match) < 2 {
				continue
			}
			suffix := strings.TrimSpace(match[1])
			if suffix == "" {
				continue
			}
			name := prefix + suffix
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, claudeCodeNativePackageCandidate{Name: name})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Version < out[j].Version
	})
	return out
}

func mergeManifestAndSourceNativePackageCandidates(existing, inferred []claudeCodeNativePackageCandidate) []claudeCodeNativePackageCandidate {
	if len(existing) == 0 {
		return append([]claudeCodeNativePackageCandidate(nil), inferred...)
	}
	if len(inferred) == 0 {
		return existing
	}
	out := append([]claudeCodeNativePackageCandidate(nil), existing...)
	seen := make(map[string]struct{}, len(existing))
	for _, candidate := range existing {
		seen[candidate.Name] = struct{}{}
	}
	for _, candidate := range inferred {
		if _, exists := seen[candidate.Name]; exists {
			continue
		}
		seen[candidate.Name] = struct{}{}
		out = append(out, candidate)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Version < out[j].Version
	})
	return out
}

func preferredNativePackageCandidates(candidates []claudeCodeNativePackageCandidate) []claudeCodeNativePackageCandidate {
	if len(candidates) <= 1 {
		return append([]claudeCodeNativePackageCandidate(nil), candidates...)
	}
	exactHost := hostClaudeCodeNativePackageName()
	sameOSPrefix := hostClaudeCodeNativeOSPrefix()
	ordered := append([]claudeCodeNativePackageCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftScore := nativePackageCandidateScore(ordered[i].Name, exactHost, sameOSPrefix)
		rightScore := nativePackageCandidateScore(ordered[j].Name, exactHost, sameOSPrefix)
		if leftScore != rightScore {
			return leftScore < rightScore
		}
		if ordered[i].Name != ordered[j].Name {
			return ordered[i].Name < ordered[j].Name
		}
		return ordered[i].Version < ordered[j].Version
	})
	return ordered
}

func nativePackageCandidateScore(packageName, exactHost, sameOSPrefix string) int {
	switch {
	case exactHost != "" && packageName == exactHost:
		return 0
	case exactHost != "" && strings.HasPrefix(packageName, exactHost+"-"):
		return 1
	case sameOSPrefix != "" && strings.HasPrefix(packageName, sameOSPrefix):
		return 2
	default:
		return 3
	}
}

func hostClaudeCodeNativePackageName() string {
	hostOS := runtime.GOOS
	switch hostOS {
	case "windows":
		hostOS = "win32"
	case "darwin", "linux":
	default:
		return ""
	}

	hostArch := runtime.GOARCH
	switch hostArch {
	case "amd64":
		hostArch = "x64"
	case "arm64":
	default:
		return ""
	}
	return nativeClaudeCodePackagePrefix + hostOS + "-" + hostArch
}

func hostClaudeCodeNativeOSPrefix() string {
	hostOS := runtime.GOOS
	switch hostOS {
	case "windows":
		hostOS = "win32"
	case "darwin", "linux":
	default:
		return ""
	}
	return nativeClaudeCodePackagePrefix + hostOS + "-"
}

func mergeClaudeCodePackageAssets(base, extra claudeCodePackageAssets) claudeCodePackageAssets {
	merged := base
	if strings.TrimSpace(merged.PackageJSON) == "" {
		merged.PackageJSON = extra.PackageJSON
		merged.Manifest = extra.Manifest
	}
	sourceSeen := make(map[string]struct{}, len(base.Sources)+len(extra.Sources))
	merged.Sources = merged.Sources[:0]
	for _, source := range append(append([]claudeCodeSourceAsset(nil), base.Sources...), extra.Sources...) {
		key := source.Path + "\x00" + source.Content
		if _, exists := sourceSeen[key]; exists {
			continue
		}
		sourceSeen[key] = struct{}{}
		merged.Sources = append(merged.Sources, source)
	}
	candidateSeen := make(map[string]struct{}, len(base.NativePackageCandidates)+len(extra.NativePackageCandidates))
	merged.NativePackageCandidates = merged.NativePackageCandidates[:0]
	for _, candidate := range append(append([]claudeCodeNativePackageCandidate(nil), base.NativePackageCandidates...), extra.NativePackageCandidates...) {
		key := candidate.Name + "\x00" + candidate.Version
		if _, exists := candidateSeen[key]; exists {
			continue
		}
		candidateSeen[key] = struct{}{}
		merged.NativePackageCandidates = append(merged.NativePackageCandidates, candidate)
	}
	sort.Slice(merged.NativePackageCandidates, func(i, j int) bool {
		if merged.NativePackageCandidates[i].Name != merged.NativePackageCandidates[j].Name {
			return merged.NativePackageCandidates[i].Name < merged.NativePackageCandidates[j].Name
		}
		return merged.NativePackageCandidates[i].Version < merged.NativePackageCandidates[j].Version
	})
	return merged
}

func combineClaudeCodeSources(sources []claudeCodeSourceAsset) string {
	if len(sources) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, source := range sources {
		if strings.TrimSpace(source.Content) == "" {
			continue
		}
		if builder.Len() > 0 {
			_ = builder.WriteByte('\n')
		}
		_, _ = builder.WriteString(source.Content)
	}
	return builder.String()
}

func isClaudeCodeTextSourcePath(name string) bool {
	if !strings.HasPrefix(name, "package/") {
		return false
	}
	ext := strings.ToLower(path.Ext(name))
	switch ext {
	case ".js", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func isClaudeCodeNativeBinaryPath(name string) bool {
	if !strings.HasPrefix(name, "package/") {
		return false
	}
	base := strings.ToLower(path.Base(name))
	return base == "claude" || base == "claude.exe"
}

func isPrintableBinaryTextByte(b byte) bool {
	return b >= 0x20 && b <= 0x7e
}

func extractRelevantTextFromBinary(r io.Reader, size int64) (string, error) {
	if size > maxNativeBinarySize {
		return "", fmt.Errorf("native binary too large: %d bytes (limit %d)", size, maxNativeBinarySize)
	}
	collector := newBinarySnippetCollector(maxBinarySnippetBytes)
	buf := make([]byte, 32<<10)
	run := make([]byte, 0, maxBinaryRunBufferBytes)
	flushRun := func(content []byte) {
		if len(content) == 0 {
			return
		}
		collector.CollectString(string(content))
	}
	for {
		n, err := r.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				if isPrintableBinaryTextByte(b) {
					run = append(run, b)
					if len(run) > maxBinaryRunBufferBytes {
						processLen := len(run) - binarySnippetContextBytes
						if processLen > 0 {
							flushRun(run[:processLen])
							tailStart := processLen - binarySnippetContextBytes
							if tailStart < 0 {
								tailStart = 0
							}
							run = append([]byte(nil), run[tailStart:]...)
						}
					}
					continue
				}
				flushRun(run)
				run = run[:0]
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	flushRun(run)
	return collector.String(), nil
}

type binarySnippetCollector struct {
	maxBytes int
	size     int
	seen     map[string]struct{}
	snippets []string
}

func newBinarySnippetCollector(maxBytes int) *binarySnippetCollector {
	return &binarySnippetCollector{
		maxBytes: maxBytes,
		seen:     make(map[string]struct{}),
	}
}

func (c *binarySnippetCollector) CollectString(source string) {
	if c == nil || c.size >= c.maxBytes || strings.TrimSpace(source) == "" {
		return
	}
	c.collectMatches(source, reBetaToken, binarySnippetTokenContext, binarySnippetTokenContext)
	c.collectMatches(source, reBinarySnippetAnchor, binarySnippetAnchorContext, binarySnippetAnchorContext)
}

func (c *binarySnippetCollector) collectMatches(source string, re *regexp.Regexp, before, after int) {
	if c == nil || c.size >= c.maxBytes {
		return
	}
	for _, match := range re.FindAllStringIndex(source, -1) {
		start := match[0] - before
		if start < 0 {
			start = 0
		}
		end := match[1] + after
		if end > len(source) {
			end = len(source)
		}
		c.addSnippet(source[start:end])
		if c.size >= c.maxBytes {
			return
		}
	}
}

func (c *binarySnippetCollector) addSnippet(snippet string) {
	if c == nil || c.size >= c.maxBytes {
		return
	}
	snippet = strings.TrimSpace(snippet)
	if snippet == "" {
		return
	}
	if _, exists := c.seen[snippet]; exists {
		return
	}
	remaining := c.maxBytes - c.size
	if remaining <= 0 {
		return
	}
	if len(snippet) > remaining {
		snippet = snippet[:remaining]
	}
	c.seen[snippet] = struct{}{}
	c.snippets = append(c.snippets, snippet)
	c.size += len(snippet)
}

func (c *binarySnippetCollector) String() string {
	if c == nil || len(c.snippets) == 0 {
		return ""
	}
	return strings.Join(c.snippets, "\n")
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
	var (
		token        string
		tokenVersion string
	)
	for _, match := range matches {
		candidate := strings.TrimSpace(match)
		if !strings.HasPrefix(candidate, prefix) {
			continue
		}
		version, ok := betaTokenVersionKey(candidate, prefix)
		if !ok {
			if token == "" {
				token = candidate
			}
			continue
		}
		if token == "" || version > tokenVersion {
			token = candidate
			tokenVersion = version
		}
	}
	return token
}

// betaTokenVersionKey normalizes supported beta token date suffixes into
// YYYYMMDD so lexical comparison selects the newest bundle token.
func betaTokenVersionKey(token, prefix string) (string, bool) {
	if !strings.HasPrefix(token, prefix) {
		return "", false
	}
	version := strings.TrimSpace(strings.TrimPrefix(token, prefix))
	version = strings.ReplaceAll(version, "-", "")
	if len(version) != 8 {
		return "", false
	}
	for _, ch := range version {
		if ch < '0' || ch > '9' {
			return "", false
		}
	}
	return version, true
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
	windowStart := anchorLoc[0] - attributionSearchWindowBytes
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
