package service

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	compatibleUpstreamAuthHeaderAuthorization = "authorization"
	compatibleUpstreamAuthHeaderAPIKey        = "api-key"
)

type compatibleResponsesUpstreamTarget struct {
	URL        string
	AuthHeader string
}

func newCompatibleResponsesUpstreamTarget(base string) compatibleResponsesUpstreamTarget {
	return newCompatibleResponsesUpstreamTargetWithOptions(base, "", "")
}

func newCompatibleResponsesUpstreamTargetWithOptions(base, authMode, endpointOverride string) compatibleResponsesUpstreamTarget {
	normalizedBase := strings.TrimSpace(base)
	if normalizedBase == "" {
		normalizedBase = openaiPlatformAPIURL
	}
	resolvedAuthHeader := resolveCompatibleResponsesAuthHeader(normalizedBase)
	if normalized := normalizeCompatibleAuthMode(authMode); normalized != "" {
		switch normalized {
		case UpstreamAuthModeXAPIKey:
			resolvedAuthHeader = "x-api-key"
		case UpstreamAuthModeAPIKey:
			resolvedAuthHeader = compatibleUpstreamAuthHeaderAPIKey
		case UpstreamAuthModeDual, UpstreamAuthModeBearer:
			resolvedAuthHeader = compatibleUpstreamAuthHeaderAuthorization
		}
	}

	targetURL := buildCompatibleResponsesURL(normalizedBase)
	if strings.TrimSpace(endpointOverride) != "" {
		targetURL = resolveCompatibleEndpointURL(normalizedBase, "/v1/responses", endpointOverride)
	}
	return compatibleResponsesUpstreamTarget{
		URL:        targetURL,
		AuthHeader: resolvedAuthHeader,
	}
}

func newCompatiblePassthroughUpstreamTargetWithOptions(base, rawRequestPath, authMode, responsesOverride, chatOverride string) compatibleResponsesUpstreamTarget {
	normalizedBase := strings.TrimSpace(base)
	if normalizedBase == "" {
		normalizedBase = openaiPlatformAPIURL
	}

	targetURL := buildCompatibleResponsesURL(normalizedBase)
	if strings.TrimSpace(responsesOverride) != "" {
		targetURL = resolveCompatibleEndpointURL(normalizedBase, "/v1/responses", responsesOverride)
	}
	switch {
	case hasOpenAIResponsesRequestPath(rawRequestPath):
		if suffix := openAIResponsesRequestPathSuffixFromPath(rawRequestPath); suffix != "" {
			targetURL = appendOpenAIResponsesRequestPathSuffix(targetURL, suffix)
		}
	case isOpenAIChatCompletionsPath(rawRequestPath):
		targetURL = buildCompatibleChatCompletionsURL(normalizedBase)
		if strings.TrimSpace(chatOverride) != "" {
			targetURL = resolveCompatibleEndpointURL(normalizedBase, "/v1/chat/completions", chatOverride)
		}
	default:
		if passthroughPath := normalizeCompatiblePassthroughRequestPath(rawRequestPath); passthroughPath != "" {
			targetURL = resolveCompatibleEndpointURL(normalizedBase, passthroughPath, "")
		}
	}

	resolvedAuthHeader := resolveCompatibleResponsesAuthHeader(normalizedBase)
	if normalized := normalizeCompatibleAuthMode(authMode); normalized != "" {
		switch normalized {
		case UpstreamAuthModeXAPIKey:
			resolvedAuthHeader = "x-api-key"
		case UpstreamAuthModeAPIKey:
			resolvedAuthHeader = compatibleUpstreamAuthHeaderAPIKey
		case UpstreamAuthModeDual, UpstreamAuthModeBearer:
			resolvedAuthHeader = compatibleUpstreamAuthHeaderAuthorization
		}
	}
	return compatibleResponsesUpstreamTarget{
		URL:        targetURL,
		AuthHeader: resolvedAuthHeader,
	}
}

func (t compatibleResponsesUpstreamTarget) ApplyAuthHeader(header http.Header, token string) {
	applyCompatibleAuthHeaders(header, token, t.AuthHeader)
}

func resolveCompatibleResponsesAuthHeader(base string) string {
	if isAzureOpenAIResponsesBaseURL(base) {
		return compatibleUpstreamAuthHeaderAPIKey
	}
	return compatibleUpstreamAuthHeaderAuthorization
}

// buildCompatibleResponsesURL resolves the shared Responses endpoint for compatible gateways.
func buildCompatibleResponsesURL(base string) string {
	normalized := strings.TrimSpace(base)
	parsed, err := url.Parse(normalized)
	if err != nil {
		return buildCompatibleResponsesURLLegacy(normalized)
	}

	parsed.Path = normalizeCompatibleResponsesPath(parsed.Path, isAzureOpenAIParsedBaseURL(parsed))
	return parsed.String()
}

func buildCompatibleResponsesURLLegacy(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/responses") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/responses"
	}
	return normalized + "/v1/responses"
}

func buildCompatibleChatCompletionsURL(base string) string {
	normalized := strings.TrimSpace(base)
	parsed, err := url.Parse(normalized)
	if err != nil {
		return buildCompatibleChatCompletionsURLLegacy(normalized)
	}

	parsed.Path = normalizeCompatibleChatCompletionsPath(parsed.Path)
	return parsed.String()
}

func buildCompatibleChatCompletionsURLLegacy(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/chat/completions") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/chat/completions"
	}
	return normalized + "/v1/chat/completions"
}

func buildCompatibleModelsURL(base string) string {
	normalized := strings.TrimSpace(base)
	parsed, err := url.Parse(normalized)
	if err != nil {
		return buildCompatibleModelsURLLegacy(normalized)
	}

	parsed.Path = normalizeCompatibleModelsPath(parsed.Path, isAzureOpenAIParsedBaseURL(parsed))
	return parsed.String()
}

func buildCompatibleModelsURLLegacy(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/models") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/models"
	}
	return normalized + "/v1/models"
}

func normalizeCompatibleChatCompletionsPath(path string) string {
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

func normalizeCompatibleModelsPath(path string, azure bool) string {
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

func normalizeCompatiblePassthroughRequestPath(rawPath string) string {
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

func normalizeCompatibleResponsesPath(path string, azure bool) string {
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
