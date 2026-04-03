package service

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/senran-N/sub2api/internal/pkg/claude"
)

func currentClaudeCodeSystemPrompt() string {
	prompt := strings.TrimSpace(claude.SystemPromptText())
	if prompt != "" {
		return prompt
	}
	return claudeCodeSystemPrompt
}

func currentClaudeCodePromptPrefixes() []string {
	templates := claude.SystemPromptTemplates()
	if len(templates) == 0 {
		return claudeCodePromptPrefixes
	}
	prefixes := make([]string, 0, len(templates)+2)
	prefixes = append(prefixes, templates...)
	prefixes = append(prefixes,
		"You are a file search specialist for Claude Code",
		"You are a helpful AI assistant tasked with summarizing conversations",
	)
	return prefixes
}

// applyClaudeCodeMimicHeaders normalizes a minimal set of stable Claude Code headers.
// 版本化/设备相关字段优先透传或复用缓存指纹，不再在这里硬编码。
func applyClaudeCodeMimicHeaders(req *http.Request, isStream bool) {
	if req == nil {
		return
	}

	applyClaudeOAuthHeaderDefaults(req)
	for key, value := range claude.StableHeaders() {
		if value == "" {
			continue
		}
		setHeaderRaw(req.Header, resolveWireCasing(key), value)
	}

	setHeaderRaw(req.Header, "Accept", "application/json")
	if isStream {
		setHeaderRaw(req.Header, "x-stainless-helper-method", "stream")
	}
	if getHeaderRaw(req.Header, "x-client-request-id") == "" {
		setHeaderRaw(req.Header, "x-client-request-id", uuid.New().String())
	}
}
