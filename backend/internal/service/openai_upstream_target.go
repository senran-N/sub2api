package service

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	openAIUpstreamAuthHeaderAuthorization = "authorization"
	openAIUpstreamAuthHeaderAPIKey        = "api-key"
)

type openAIResponsesUpstreamTarget struct {
	URL        string
	AuthHeader string
}

func newOpenAIResponsesUpstreamTarget(base string) openAIResponsesUpstreamTarget {
	return newOpenAIResponsesUpstreamTargetWithOptions(base, "", "")
}

func newOpenAIResponsesUpstreamTargetWithOptions(base, authMode, endpointOverride string) openAIResponsesUpstreamTarget {
	normalizedBase := strings.TrimSpace(base)
	if normalizedBase == "" {
		normalizedBase = openaiPlatformAPIURL
	}
	resolvedAuthHeader := resolveOpenAIResponsesAuthHeader(normalizedBase)
	if normalized := normalizeCompatibleAuthMode(authMode); normalized != "" {
		switch normalized {
		case UpstreamAuthModeXAPIKey:
			resolvedAuthHeader = "x-api-key"
		case UpstreamAuthModeAPIKey:
			resolvedAuthHeader = openAIUpstreamAuthHeaderAPIKey
		case UpstreamAuthModeDual, UpstreamAuthModeBearer:
			resolvedAuthHeader = openAIUpstreamAuthHeaderAuthorization
		}
	}

	targetURL := buildOpenAIResponsesURL(normalizedBase)
	if strings.TrimSpace(endpointOverride) != "" {
		targetURL = resolveCompatibleEndpointURL(normalizedBase, "/v1/responses", endpointOverride)
	}
	return openAIResponsesUpstreamTarget{
		URL:        targetURL,
		AuthHeader: resolvedAuthHeader,
	}
}

func newOpenAIPassthroughUpstreamTarget(base, rawRequestPath string) openAIResponsesUpstreamTarget {
	return newOpenAIPassthroughUpstreamTargetWithOptions(base, rawRequestPath, "", "", "")
}

func newOpenAIPassthroughUpstreamTargetWithOptions(base, rawRequestPath, authMode, responsesOverride, chatOverride string) openAIResponsesUpstreamTarget {
	normalizedBase := strings.TrimSpace(base)
	if normalizedBase == "" {
		normalizedBase = openaiPlatformAPIURL
	}

	targetURL := buildOpenAIResponsesURL(normalizedBase)
	if strings.TrimSpace(responsesOverride) != "" {
		targetURL = resolveCompatibleEndpointURL(normalizedBase, "/v1/responses", responsesOverride)
	}
	switch {
	case hasOpenAIResponsesRequestPath(rawRequestPath):
		if suffix := openAIResponsesRequestPathSuffixFromPath(rawRequestPath); suffix != "" {
			targetURL = appendOpenAIResponsesRequestPathSuffix(targetURL, suffix)
		}
	case isOpenAIChatCompletionsPath(rawRequestPath):
		targetURL = buildOpenAIChatCompletionsURL(normalizedBase)
		if strings.TrimSpace(chatOverride) != "" {
			targetURL = resolveCompatibleEndpointURL(normalizedBase, "/v1/chat/completions", chatOverride)
		}
	default:
		if passthroughPath := normalizeOpenAICompatiblePassthroughRequestPath(rawRequestPath); passthroughPath != "" {
			targetURL = resolveCompatibleEndpointURL(normalizedBase, passthroughPath, "")
		}
	}

	resolvedAuthHeader := resolveOpenAIResponsesAuthHeader(normalizedBase)
	if normalized := normalizeCompatibleAuthMode(authMode); normalized != "" {
		switch normalized {
		case UpstreamAuthModeXAPIKey:
			resolvedAuthHeader = "x-api-key"
		case UpstreamAuthModeAPIKey:
			resolvedAuthHeader = openAIUpstreamAuthHeaderAPIKey
		case UpstreamAuthModeDual, UpstreamAuthModeBearer:
			resolvedAuthHeader = openAIUpstreamAuthHeaderAuthorization
		}
	}
	return openAIResponsesUpstreamTarget{
		URL:        targetURL,
		AuthHeader: resolvedAuthHeader,
	}
}

func (t openAIResponsesUpstreamTarget) ApplyAuthHeader(header http.Header, token string) {
	applyCompatibleAuthHeaders(header, token, t.AuthHeader)
}

func resolveOpenAIResponsesAuthHeader(base string) string {
	if isAzureOpenAIResponsesBaseURL(base) {
		return openAIUpstreamAuthHeaderAPIKey
	}
	return openAIUpstreamAuthHeaderAuthorization
}

// buildOpenAIResponsesURL 组装 OpenAI Responses 端点。
func buildOpenAIResponsesURL(base string) string {
	normalized := strings.TrimSpace(base)
	parsed, err := url.Parse(normalized)
	if err != nil {
		return buildOpenAIResponsesURLLegacy(normalized)
	}

	parsed.Path = normalizeOpenAIResponsesPath(parsed.Path, isAzureOpenAIParsedBaseURL(parsed))
	return parsed.String()
}

func buildOpenAIResponsesURLLegacy(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/responses") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/responses"
	}
	return normalized + "/v1/responses"
}

func buildOpenAIChatCompletionsURL(base string) string {
	normalized := strings.TrimSpace(base)
	parsed, err := url.Parse(normalized)
	if err != nil {
		return buildOpenAIChatCompletionsURLLegacy(normalized)
	}

	parsed.Path = normalizeOpenAIChatCompletionsPath(parsed.Path)
	return parsed.String()
}

func buildOpenAIChatCompletionsURLLegacy(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/chat/completions") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/chat/completions"
	}
	return normalized + "/v1/chat/completions"
}

func buildOpenAIModelsURL(base string) string {
	normalized := strings.TrimSpace(base)
	parsed, err := url.Parse(normalized)
	if err != nil {
		return buildOpenAIModelsURLLegacy(normalized)
	}

	parsed.Path = normalizeOpenAIModelsPath(parsed.Path, isAzureOpenAIParsedBaseURL(parsed))
	return parsed.String()
}

func buildOpenAIModelsURLLegacy(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/models") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/models"
	}
	return normalized + "/v1/models"
}

func normalizeOpenAIChatCompletionsPath(path string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(path), "/")
	switch {
	case trimmed == "":
		return "/v1/chat/completions"
	case strings.HasSuffix(trimmed, "/chat/completions"):
		return trimmed
	case strings.HasSuffix(trimmed, "/v1"):
		return trimmed + "/chat/completions"
	default:
		return trimmed + "/v1/chat/completions"
	}
}

func normalizeOpenAIModelsPath(path string, azure bool) string {
	trimmed := strings.TrimRight(strings.TrimSpace(path), "/")
	switch {
	case trimmed == "":
		if azure {
			return "/openai/models"
		}
		return "/v1/models"
	case strings.HasSuffix(trimmed, "/models"):
		return trimmed
	case azure && strings.HasSuffix(trimmed, "/openai/v1"):
		return strings.TrimSuffix(trimmed, "/v1") + "/models"
	case azure && trimmed == "/openai":
		return trimmed + "/models"
	case !azure && strings.HasSuffix(trimmed, "/v1"):
		return trimmed + "/models"
	case azure:
		return "/openai/models"
	default:
		return trimmed + "/v1/models"
	}
}

func isOpenAIChatCompletionsPath(rawPath string) bool {
	trimmed := strings.TrimRight(strings.TrimSpace(rawPath), "/")
	return strings.HasSuffix(trimmed, "/chat/completions")
}

func hasOpenAIResponsesRequestPath(rawPath string) bool {
	trimmed := strings.TrimRight(strings.TrimSpace(rawPath), "/")
	return trimmed == "/responses" || strings.Contains(trimmed, "/v1/responses")
}

func normalizeOpenAICompatiblePassthroughRequestPath(rawPath string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(rawPath), "/")
	switch {
	case trimmed == "":
		return ""
	case trimmed == "/responses" || strings.HasPrefix(trimmed, "/responses/"):
		return "/v1" + trimmed
	case trimmed == "/chat/completions":
		return "/v1/chat/completions"
	case strings.HasPrefix(trimmed, "/v1/"):
		return trimmed
	default:
		return trimmed
	}
}

func normalizeOpenAIResponsesPath(path string, azure bool) string {
	trimmed := strings.TrimRight(strings.TrimSpace(path), "/")
	switch {
	case trimmed == "":
		if azure {
			return "/openai/responses"
		}
		return "/v1/responses"
	case strings.HasSuffix(trimmed, "/responses"):
		return trimmed
	case azure && strings.HasSuffix(trimmed, "/openai/v1"):
		return strings.TrimSuffix(trimmed, "/v1") + "/responses"
	case azure && trimmed == "/openai":
		return trimmed + "/responses"
	case !azure && strings.HasSuffix(trimmed, "/v1"):
		return trimmed + "/responses"
	case azure:
		return "/openai/responses"
	default:
		return trimmed + "/v1/responses"
	}
}

func isAzureOpenAIResponsesBaseURL(base string) bool {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return false
	}
	return isAzureOpenAIParsedBaseURL(parsed)
}

func isAzureOpenAIParsedBaseURL(parsed *url.URL) bool {
	if parsed == nil {
		return false
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	switch {
	case strings.HasSuffix(host, ".openai.azure.com"):
		return true
	case strings.HasSuffix(host, ".cognitiveservices.azure.com"):
		return true
	}

	path := strings.TrimRight(strings.TrimSpace(parsed.Path), "/")
	if path == "/openai" || strings.HasPrefix(path, "/openai/") {
		return true
	}
	return parsed.Query().Has("api-version")
}
