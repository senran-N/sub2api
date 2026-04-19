package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	grokWebBaseURL                  = "https://grok.com"
	grokSessionConversationEndpoint = "/rest/app-chat/conversations/new"
)

type grokTransportKind string

const (
	grokTransportKindCompatible grokTransportKind = "compatible"
	grokTransportKindSession    grokTransportKind = "session"
)

type grokTransportTarget struct {
	Kind         grokTransportKind
	URL          string
	AuthToken    string
	AuthHeader   string
	CookieHeader string
}

func (t grokTransportTarget) Apply(req *http.Request) {
	if req == nil {
		return
	}
	if t.CookieHeader != "" {
		req.Header.Set("Cookie", t.CookieHeader)
	}
	if t.AuthHeader != "" && t.AuthToken != "" {
		applyCompatibleAuthHeaders(req.Header, t.AuthToken, t.AuthHeader)
	}
}

func (a *Account) GetGrokSessionToken() string {
	if a == nil || NormalizeCompatibleGatewayPlatform(a.Platform) != PlatformGrok || a.Type != AccountTypeSession {
		return ""
	}
	return strings.TrimSpace(a.GetCredential("session_token"))
}

func resolveGrokTransportTarget(account *Account, validateBaseURL func(string) (string, error)) (grokTransportTarget, error) {
	if account == nil {
		return grokTransportTarget{}, errors.New("account is nil")
	}
	if NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return grokTransportTarget{}, fmt.Errorf("unsupported grok platform: %s", account.Platform)
	}

	switch account.Type {
	case AccountTypeAPIKey, AccountTypeUpstream:
		apiKey := strings.TrimSpace(account.GetOpenAIApiKey())
		if apiKey == "" {
			return grokTransportTarget{}, errors.New("api_key not found in credentials")
		}

		baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
		if baseURL == "" {
			baseURL = CompatibleGatewayDefaultBaseURL(account.Platform)
		}
		if validateBaseURL != nil {
			normalizedBaseURL, err := validateBaseURL(baseURL)
			if err != nil {
				return grokTransportTarget{}, err
			}
			baseURL = normalizedBaseURL
		}

		target := newCompatibleResponsesUpstreamTargetWithOptions(
			baseURL,
			account.GetCompatibleAuthMode(""),
			account.GetCompatibleEndpointOverride("responses"),
		)
		return grokTransportTarget{
			Kind:       grokTransportKindCompatible,
			URL:        target.URL,
			AuthToken:  apiKey,
			AuthHeader: target.AuthHeader,
		}, nil
	case AccountTypeSession:
		cookieHeader, err := normalizeGrokSessionCookieHeader(account.GetGrokSessionToken())
		if err != nil {
			return grokTransportTarget{}, err
		}
		targetURL, err := buildGrokSessionTransportURL(validateBaseURL)
		if err != nil {
			return grokTransportTarget{}, err
		}
		return grokTransportTarget{
			Kind:         grokTransportKindSession,
			URL:          targetURL,
			CookieHeader: cookieHeader,
		}, nil
	default:
		return grokTransportTarget{}, fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func buildGrokSessionTransportURL(validateBaseURL func(string) (string, error)) (string, error) {
	baseURL := grokWebBaseURL
	if validateBaseURL != nil {
		normalizedBaseURL, err := validateBaseURL(baseURL)
		if err != nil {
			return "", err
		}
		baseURL = normalizedBaseURL
	}
	return appendGrokSessionTransportPath(baseURL), nil
}

func appendGrokSessionTransportPath(baseURL string) string {
	trimmedBase := strings.TrimSpace(baseURL)
	parsed, err := url.Parse(trimmedBase)
	if err != nil {
		return strings.TrimRight(trimmedBase, "/") + grokSessionConversationEndpoint
	}

	parsed.Path = strings.TrimRight(strings.TrimSpace(parsed.Path), "/") + grokSessionConversationEndpoint
	parsed.RawPath = ""
	return parsed.String()
}

// Grok session accounts may paste either a full browser cookie string or just the
// bare session value. When only a single token is present, treat it as the `sso`
// cookie so the provider-owned transport can build a valid Cookie header.
func normalizeGrokSessionCookieHeader(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "Cookie:")
	trimmed = strings.TrimPrefix(trimmed, "cookie:")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "", errors.New("session_token not found in credentials")
	}
	if !strings.Contains(trimmed, "=") {
		return "sso=" + trimmed, nil
	}

	parts := strings.Split(trimmed, ";")
	normalizedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || !strings.Contains(part, "=") {
			continue
		}
		normalizedParts = append(normalizedParts, part)
	}
	if len(normalizedParts) == 0 {
		return "", errors.New("session_token does not contain a valid Grok cookie")
	}
	return strings.Join(normalizedParts, "; "), nil
}
