package claude

import (
	"strings"
	"sync/atomic"
)

type MimicProfile struct {
	Source                       string
	PackageName                  string
	PackageVersion               string
	SDKVersion                   string // @anthropic-ai/sdk version for X-Stainless-Package-Version
	UserAgent                    string
	XApp                         string
	OAuthBeta                    string
	ClaudeCodeBeta               string
	InterleavedThinkingBeta      string
	FineGrainedToolStreamingBeta string
	TokenCountingBeta            string
	Context1MBeta                string
	FastModeBeta                 string
	SystemPrompt                 string
	SystemPromptPrefixes         []string
	DefaultHeaders               map[string]string
	StableDefaultHeaders         map[string]string
	AttributionSalt              string // Attribution fingerprint salt, e.g. "59cf53e54c78"
	AttributionIndices           []int  // Attribution fingerprint char indices, e.g. [4, 7, 20]
}

var currentMimicProfile atomic.Value

func init() {
	currentMimicProfile.Store(cloneMimicProfile(defaultMimicProfile()))
}

func defaultMimicProfile() MimicProfile {
	return MimicProfile{
		Source:                       "builtin",
		PackageName:                  "@anthropic-ai/claude-code",
		PackageVersion:               "2.1.22",
		SDKVersion:                   DefaultHeaders["X-Stainless-Package-Version"],
		UserAgent:                    DefaultHeaders["User-Agent"],
		XApp:                         StableDefaultHeaders["X-App"],
		OAuthBeta:                    BetaOAuth,
		ClaudeCodeBeta:               BetaClaudeCode,
		InterleavedThinkingBeta:      BetaInterleavedThinking,
		FineGrainedToolStreamingBeta: BetaFineGrainedToolStreaming,
		TokenCountingBeta:            BetaTokenCounting,
		Context1MBeta:                BetaContext1M,
		FastModeBeta:                 BetaFastMode,
		SystemPrompt:                 "You are Claude Code, Anthropic's official CLI for Claude.",
		SystemPromptPrefixes: []string{
			"You are Claude Code, Anthropic's official CLI for Claude.",
			"You are a Claude agent, built on Anthropic's Claude Agent SDK.",
			"You are Claude Code, Anthropic's official CLI for Claude, running within the Claude Agent SDK.",
		},
		DefaultHeaders:       cloneStringMap(DefaultHeaders),
		StableDefaultHeaders: cloneStringMap(StableDefaultHeaders),
		// Builtin fallback values extracted from Claude Code 2.1.88.
		// These are overwritten at runtime by npm sync; kept here as last-resort defaults.
		AttributionSalt:    "59cf53e54c78",
		AttributionIndices: []int{4, 7, 20},
	}
}

func CurrentMimicProfile() MimicProfile {
	profile, _ := currentMimicProfile.Load().(MimicProfile)
	return cloneMimicProfile(profile)
}

func BuiltinMimicProfile() MimicProfile {
	return cloneMimicProfile(defaultMimicProfile())
}

func ApplyMimicProfile(profile MimicProfile) {
	currentMimicProfile.Store(cloneMimicProfile(normalizeMimicProfile(profile)))
}

func DefaultUserAgent() string {
	return currentProfile().UserAgent
}

func DefaultHeaderSet() map[string]string {
	return cloneStringMap(currentProfile().DefaultHeaders)
}

func StableHeaders() map[string]string {
	return cloneStringMap(currentProfile().StableDefaultHeaders)
}

func SystemPromptText() string {
	return currentProfile().SystemPrompt
}

func SystemPromptTemplates() []string {
	return cloneStringSlice(currentProfile().SystemPromptPrefixes)
}

func OAuthBetaToken() string {
	return currentProfile().OAuthBeta
}

func ClaudeCodeBetaToken() string {
	return currentProfile().ClaudeCodeBeta
}

func InterleavedThinkingBetaToken() string {
	return currentProfile().InterleavedThinkingBeta
}

func FineGrainedToolStreamingBetaToken() string {
	return currentProfile().FineGrainedToolStreamingBeta
}

func TokenCountingBetaToken() string {
	return currentProfile().TokenCountingBeta
}

func Context1MBetaToken() string {
	return currentProfile().Context1MBeta
}

func FastModeBetaToken() string {
	return currentProfile().FastModeBeta
}

func SDKVersion() string {
	return currentProfile().SDKVersion
}

func AttributionSalt() string {
	return currentProfile().AttributionSalt
}

func AttributionIndices() []int {
	src := currentProfile().AttributionIndices
	if len(src) == 0 {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

func DefaultAnthropicBetaHeader() string {
	return joinBetaTokens(
		ClaudeCodeBetaToken(),
		OAuthBetaToken(),
		InterleavedThinkingBetaToken(),
		FineGrainedToolStreamingBetaToken(),
	)
}

func CountTokensAnthropicBetaHeader() string {
	return joinBetaTokens(
		ClaudeCodeBetaToken(),
		OAuthBetaToken(),
		InterleavedThinkingBetaToken(),
		TokenCountingBetaToken(),
	)
}

func HaikuAnthropicBetaHeader() string {
	return joinBetaTokens(
		OAuthBetaToken(),
		InterleavedThinkingBetaToken(),
	)
}

func APIKeyAnthropicBetaHeader() string {
	return joinBetaTokens(
		ClaudeCodeBetaToken(),
		InterleavedThinkingBetaToken(),
		FineGrainedToolStreamingBetaToken(),
	)
}

func APIKeyHaikuAnthropicBetaHeader() string {
	return joinBetaTokens(InterleavedThinkingBetaToken())
}

func currentProfile() MimicProfile {
	profile, _ := currentMimicProfile.Load().(MimicProfile)
	return profile
}

func normalizeMimicProfile(profile MimicProfile) MimicProfile {
	fallback := defaultMimicProfile()
	if strings.TrimSpace(profile.Source) == "" {
		profile.Source = fallback.Source
	}
	if strings.TrimSpace(profile.PackageName) == "" {
		profile.PackageName = fallback.PackageName
	}
	if strings.TrimSpace(profile.PackageVersion) == "" {
		profile.PackageVersion = fallback.PackageVersion
	}
	if strings.TrimSpace(profile.UserAgent) == "" {
		profile.UserAgent = "claude-cli/" + profile.PackageVersion + " (external, cli)"
	}
	if strings.TrimSpace(profile.XApp) == "" {
		profile.XApp = fallback.XApp
	}
	if strings.TrimSpace(profile.OAuthBeta) == "" {
		profile.OAuthBeta = fallback.OAuthBeta
	}
	if strings.TrimSpace(profile.ClaudeCodeBeta) == "" {
		profile.ClaudeCodeBeta = fallback.ClaudeCodeBeta
	}
	if strings.TrimSpace(profile.InterleavedThinkingBeta) == "" {
		profile.InterleavedThinkingBeta = fallback.InterleavedThinkingBeta
	}
	if strings.TrimSpace(profile.TokenCountingBeta) == "" {
		profile.TokenCountingBeta = fallback.TokenCountingBeta
	}
	if strings.TrimSpace(profile.Context1MBeta) == "" {
		profile.Context1MBeta = fallback.Context1MBeta
	}
	if strings.TrimSpace(profile.FastModeBeta) == "" {
		profile.FastModeBeta = fallback.FastModeBeta
	}
	if strings.TrimSpace(profile.SystemPrompt) == "" {
		profile.SystemPrompt = fallback.SystemPrompt
	}
	if len(profile.SystemPromptPrefixes) == 0 {
		profile.SystemPromptPrefixes = cloneStringSlice(fallback.SystemPromptPrefixes)
	}
	if len(profile.DefaultHeaders) == 0 {
		profile.DefaultHeaders = cloneStringMap(fallback.DefaultHeaders)
	}
	if len(profile.StableDefaultHeaders) == 0 {
		profile.StableDefaultHeaders = cloneStringMap(fallback.StableDefaultHeaders)
	}
	if strings.TrimSpace(profile.AttributionSalt) == "" {
		profile.AttributionSalt = fallback.AttributionSalt
	}
	if len(profile.AttributionIndices) == 0 {
		profile.AttributionIndices = cloneIntSlice(fallback.AttributionIndices)
	}
	if strings.TrimSpace(profile.SDKVersion) == "" {
		profile.SDKVersion = fallback.SDKVersion
	}

	profile.DefaultHeaders["User-Agent"] = profile.UserAgent
	if profile.SDKVersion != "" {
		profile.DefaultHeaders["X-Stainless-Package-Version"] = profile.SDKVersion
	}
	profile.StableDefaultHeaders["X-App"] = profile.XApp
	return profile
}

func cloneMimicProfile(profile MimicProfile) MimicProfile {
	profile.SystemPromptPrefixes = cloneStringSlice(profile.SystemPromptPrefixes)
	profile.DefaultHeaders = cloneStringMap(profile.DefaultHeaders)
	profile.StableDefaultHeaders = cloneStringMap(profile.StableDefaultHeaders)
	profile.AttributionIndices = cloneIntSlice(profile.AttributionIndices)
	return profile
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneIntSlice(src []int) []int {
	if len(src) == 0 {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

func cloneStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func joinBetaTokens(tokens ...string) string {
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return strings.Join(out, ",")
}
