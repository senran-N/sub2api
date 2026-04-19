package service

import "strings"

const (
	UpstreamAuthModeBearer  = "bearer"
	UpstreamAuthModeXAPIKey = "x-api-key"
	UpstreamAuthModeAPIKey  = "api-key"
	UpstreamAuthModeDual    = "dual"
)

func normalizeCompatibleAuthMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case UpstreamAuthModeBearer:
		return UpstreamAuthModeBearer
	case UpstreamAuthModeXAPIKey, "x_api_key":
		return UpstreamAuthModeXAPIKey
	case UpstreamAuthModeAPIKey, "apikey":
		return UpstreamAuthModeAPIKey
	case UpstreamAuthModeDual:
		return UpstreamAuthModeDual
	default:
		return ""
	}
}

func (a *Account) isCompatibleAPIKeyAccount() bool {
	if a == nil {
		return false
	}
	switch a.Platform {
	case PlatformOpenAI, PlatformGrok, PlatformAnthropic:
		return a.Type == AccountTypeAPIKey || a.Type == AccountTypeUpstream
	default:
		return false
	}
}

func (a *Account) DefaultCompatibleAuthMode() string {
	if a == nil {
		return ""
	}
	switch a.Platform {
	case PlatformOpenAI, PlatformGrok:
		baseURL := a.GetOpenAIBaseURL()
		if isAzureOpenAIResponsesBaseURL(baseURL) {
			return UpstreamAuthModeAPIKey
		}
		return UpstreamAuthModeBearer
	case PlatformAnthropic:
		if a.IsOAuth() {
			return UpstreamAuthModeBearer
		}
		return UpstreamAuthModeXAPIKey
	default:
		return ""
	}
}

func (a *Account) GetCompatibleAuthMode(defaultMode string) string {
	if a == nil {
		return normalizeCompatibleAuthMode(defaultMode)
	}
	if normalized := normalizeCompatibleAuthMode(a.GetCredential("auth_mode")); normalized != "" {
		return normalized
	}
	if normalized := normalizeCompatibleAuthMode(defaultMode); normalized != "" {
		return normalized
	}
	return a.DefaultCompatibleAuthMode()
}

func (a *Account) GetCompatibleBaseURL() string {
	if a == nil {
		return ""
	}
	switch a.Platform {
	case PlatformOpenAI, PlatformGrok:
		if a.Type == AccountTypeAPIKey || a.Type == AccountTypeUpstream {
			return a.GetOpenAIBaseURL()
		}
	case PlatformAnthropic:
		if a.Type == AccountTypeAPIKey || a.Type == AccountTypeUpstream {
			return a.GetBaseURL()
		}
	}
	return ""
}

func (a *Account) GetCompatibleAPIKey() string {
	if a == nil {
		return ""
	}
	switch a.Platform {
	case PlatformOpenAI, PlatformGrok:
		return a.GetOpenAIApiKey()
	case PlatformAnthropic:
		if a.Type == AccountTypeAPIKey || a.Type == AccountTypeUpstream {
			return a.GetCredential("api_key")
		}
	}
	return ""
}

func (a *Account) GetCompatibleEndpointOverride(name string) string {
	if a == nil || a.Credentials == nil {
		return ""
	}
	raw, ok := a.Credentials["endpoint_overrides"]
	if !ok || raw == nil {
		return ""
	}
	mapping, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	value, ok := mapping[name]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func (a *Account) SupportsCompatibleModelDiscovery() bool {
	if !a.isCompatibleAPIKeyAccount() {
		return false
	}
	return strings.TrimSpace(a.GetCompatibleBaseURL()) != "" && strings.TrimSpace(a.GetCompatibleAPIKey()) != ""
}
