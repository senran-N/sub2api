package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	openaipkg "github.com/senran-N/sub2api/internal/pkg/openai"
	"github.com/tidwall/gjson"
)

const (
	openAICodexMetadataInstallationIDKey = "x-codex-installation-id"
	openAICodexMetadataWindowIDKey       = "x-codex-window-id"
	openAICodexMetadataParentThreadIDKey = "x-codex-parent-thread-id"
	openAICodexMetadataSubagentKey       = "x-openai-subagent"
)

var openAICodexAllowedSubagents = map[string]struct{}{
	"review":               {},
	"compact":              {},
	"memory_consolidation": {},
	"collab_spawn":         {},
}

const openAICodexHTTPBetaValue = "responses=experimental"

var openAICodexVersionPattern = regexp.MustCompile(`^v?\d+(?:\.\d+){1,3}(?:[-+._A-Za-z0-9]+)?$`)

func applyOpenAICodexOAuthResponsesHeaders(headers http.Header, account *Account) {
	if headers == nil {
		return
	}

	headers.Set("OpenAI-Beta", openAICodexHTTPBetaValue)
	headers.Set("originator", resolveOpenAICodexUpstreamOriginator(account))
	if userAgent := resolveOpenAICodexUpstreamUserAgent(account); userAgent != "" {
		headers.Set("user-agent", userAgent)
	} else {
		headers.Del("user-agent")
	}
	if betaFeatures := resolveOpenAICodexUpstreamBetaFeatures(account); betaFeatures != "" {
		headers.Set("x-codex-beta-features", betaFeatures)
	} else {
		headers.Del("x-codex-beta-features")
	}
	if acceptLanguage := resolveOpenAICodexUpstreamAcceptLanguage(account); acceptLanguage != "" {
		headers.Set("accept-language", acceptLanguage)
	} else {
		headers.Del("accept-language")
	}
	if account != nil {
		if chatgptAccountID := account.GetChatGPTAccountID(); chatgptAccountID != "" {
			headers.Set("chatgpt-account-id", chatgptAccountID)
		} else {
			headers.Del("chatgpt-account-id")
		}
	}
}

func newOpenAICodexOAuthResponsesRequest(
	ctx context.Context,
	targetURL string,
	accessToken string,
	payload []byte,
	accept string,
	account *Account,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Host = "chatgpt.com"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", strings.TrimSpace(accept))
	applyOpenAICodexOAuthResponsesHeaders(req.Header, account)
	if strings.HasSuffix(strings.TrimRight(targetURL, "/"), "/compact") {
		req.Header.Set("version", resolveOpenAICodexUpstreamVersion(account))
	}
	return req, nil
}

func resolveOpenAICodexUpstreamUserAgent(account *Account) string {
	if version := resolveOpenAICodexOverrideVersion(account); version != "" {
		return "codex_cli_rs/" + version
	}
	return codexCLIUserAgent
}

func resolveOpenAICodexUpstreamOriginator(account *Account) string {
	if account != nil {
		if originator := strings.TrimSpace(account.GetOpenAICodexOriginator()); originator != "" {
			if openaipkg.IsCodexOfficialClientOriginator(originator) {
				return "codex_cli_rs"
			}
		}
		if userAgent := strings.TrimSpace(account.GetOpenAIUserAgent()); userAgent != "" {
			if openaipkg.IsCodexOfficialClientRequest(userAgent) {
				return "codex_cli_rs"
			}
		}
	}
	return "codex_cli_rs"
}

func resolveOpenAICodexUpstreamVersion(account *Account) string {
	if version := resolveOpenAICodexOverrideVersion(account); version != "" {
		return version
	}
	return codexCLIVersion
}

func resolveOpenAICodexOverrideVersion(account *Account) string {
	if account == nil {
		return ""
	}
	for _, candidate := range []string{
		strings.TrimSpace(account.GetOpenAICodexVersion()),
		resolveCodexVersion(strings.TrimSpace(account.GetOpenAIUserAgent()), ""),
	} {
		if version := normalizeOpenAICodexVersion(candidate); version != "" {
			return version
		}
	}
	return ""
}

func normalizeOpenAICodexVersion(raw string) string {
	value := trimCodexVersionToken(strings.TrimSpace(raw))
	if value == "" || !openAICodexVersionPattern.MatchString(value) {
		return ""
	}
	return strings.TrimPrefix(value, "v")
}

func resolveOpenAICodexUpstreamAcceptLanguage(account *Account) string {
	if account == nil {
		return ""
	}
	return strings.TrimSpace(account.GetOpenAICodexAcceptLanguage())
}

func resolveOpenAICodexUpstreamInstallationID(accountID int64) string {
	if accountID <= 0 {
		return ""
	}
	return generateSessionUUID(fmt.Sprintf("openai_codex_installation:%d", accountID))
}

func resolveOpenAICodexUpstreamWindowID(accountID int64) string {
	if accountID <= 0 {
		return ""
	}
	return generateSessionUUID(fmt.Sprintf("openai_codex_window:%d", accountID))
}

func resolveOpenAICodexUpstreamBetaFeatures(account *Account) string {
	if account == nil {
		return ""
	}
	return strings.TrimSpace(account.GetOpenAICodexBetaFeatures())
}

func resolveOpenAICodexUpstreamClientRequestID(accountID int64, rawClientRequestID, isolatedSessionID string) string {
	if value := strings.TrimSpace(isolatedSessionID); value != "" {
		return value
	}
	return isolateOpenAISessionID(accountID, rawClientRequestID)
}

func resolveOpenAICodexUpstreamSessionHeaderValue(accountID int64, raw, source string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if source == "prompt_cache_key" && isOpenAICodexVirtualizedID(value) {
		return value
	}
	return isolateOpenAISessionID(accountID, value)
}

func resolveOpenAICodexUpstreamTurnMetadata(accountID int64, raw string) string {
	return isolateOpenAISessionID(accountID, raw)
}

func rewriteOpenAICodexBodyIdentityMap(accountID int64, reqBody map[string]any) bool {
	if accountID <= 0 || len(reqBody) == 0 {
		return false
	}

	changed := false
	if promptCacheKey, ok := reqBody["prompt_cache_key"].(string); ok {
		virtualized := isolateOpenAISessionID(accountID, promptCacheKey)
		if virtualized != "" && virtualized != strings.TrimSpace(promptCacheKey) {
			reqBody["prompt_cache_key"] = virtualized
			changed = true
		}
	}
	if sanitizeOpenAICodexClientMetadata(accountID, reqBody) {
		changed = true
	}
	return changed
}

func sanitizeOpenAICodexClientMetadata(accountID int64, reqBody map[string]any) bool {
	if accountID <= 0 || len(reqBody) == 0 {
		return false
	}

	rawMetadata, exists := reqBody["client_metadata"]
	metadata := make(map[string]any)
	if exists {
		switch typed := rawMetadata.(type) {
		case map[string]any:
			for key, value := range typed {
				metadata[key] = value
			}
		case map[string]string:
			for key, value := range typed {
				metadata[key] = value
			}
		default:
			delete(reqBody, "client_metadata")
			return true
		}
	}

	sanitized := make(map[string]any)
	for key, rawValue := range metadata {
		value := strings.TrimSpace(fmt.Sprintf("%v", rawValue))
		if value == "" {
			continue
		}

		switch key {
		case openAIWSTurnMetadataHeader:
			sanitizedValue := resolveOpenAICodexUpstreamTurnMetadata(accountID, value)
			if sanitizedValue != "" {
				sanitized[key] = sanitizedValue
			}
		case openAICodexMetadataInstallationIDKey:
			if installationID := resolveOpenAICodexUpstreamInstallationID(accountID); installationID != "" {
				sanitized[key] = installationID
			}
		case openAICodexMetadataWindowIDKey:
			if windowID := resolveOpenAICodexUpstreamWindowID(accountID); windowID != "" {
				sanitized[key] = windowID
			}
		case openAICodexMetadataParentThreadIDKey:
			sanitized[key] = generateSessionUUID(fmt.Sprintf("openai_codex_parent_thread:%d:%s", accountID, value))
		case openAICodexMetadataSubagentKey:
			if normalized := normalizeOpenAICodexSubagent(value); normalized != "" {
				sanitized[key] = normalized
			}
		}
	}

	requestType := strings.TrimSpace(fmt.Sprintf("%v", reqBody["type"]))
	if installationID := resolveOpenAICodexUpstreamInstallationID(accountID); installationID != "" {
		sanitized[openAICodexMetadataInstallationIDKey] = installationID
	}
	if strings.EqualFold(requestType, "response.create") {
		if windowID := resolveOpenAICodexUpstreamWindowID(accountID); windowID != "" {
			sanitized[openAICodexMetadataWindowIDKey] = windowID
		}
	}

	if len(sanitized) == 0 && !exists {
		return false
	}
	if len(sanitized) == 0 {
		delete(reqBody, "client_metadata")
		return true
	}
	reqBody["client_metadata"] = sanitized
	return true
}

func normalizeOpenAICodexSubagent(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if _, ok := openAICodexAllowedSubagents[value]; ok {
		return value
	}
	return ""
}

func isOpenAICodexVirtualizedID(value string) bool {
	if len(value) != 16 {
		return false
	}
	for _, ch := range value {
		if !unicode.IsDigit(ch) && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}

func looksLikeOpenAICodexUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, ch := range value {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !unicode.IsDigit(ch) && (ch < 'a' || ch > 'f') {
				return false
			}
		}
	}
	return true
}

func openAICodexClientMetadataHeaderValue(body []byte, key string) string {
	if len(body) == 0 || strings.TrimSpace(key) == "" {
		return ""
	}
	return strings.TrimSpace(gjson.GetBytes(body, "client_metadata."+key).String())
}

func resolveOpenAICodexUpstreamSubagent(profile CodexRequestProfile, body []byte) string {
	if value := normalizeOpenAICodexSubagent(profile.Headers.Subagent); value != "" {
		return value
	}
	return normalizeOpenAICodexSubagent(openAICodexClientMetadataHeaderValue(body, openAICodexMetadataSubagentKey))
}

func resolveOpenAICodexUpstreamParentThreadID(accountID int64, profile CodexRequestProfile, body []byte) string {
	raw := strings.TrimSpace(profile.Headers.ParentThreadID)
	if raw == "" {
		raw = openAICodexClientMetadataHeaderValue(body, openAICodexMetadataParentThreadIDKey)
	}
	if raw == "" {
		return ""
	}
	if looksLikeOpenAICodexUUID(raw) {
		return raw
	}
	return generateSessionUUID(fmt.Sprintf("openai_codex_parent_thread:%d:%s", accountID, raw))
}
