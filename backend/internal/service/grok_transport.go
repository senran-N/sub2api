//nolint:unused
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
	Kind           grokTransportKind
	URL            string
	SessionBaseURL string
	AuthToken      string
	AuthHeader     string
	CookieHeader   string
	UserAgent      string
	AcceptLang     string
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

func resolveGrokTransportTarget(account *Account, validateBaseURL func(string) (string, error)) (grokTransportTarget, error) {
	return resolveGrokTransportTargetWithSettings(account, validateBaseURL, DefaultGrokRuntimeSettings())
}

func resolveGrokTransportTargetWithSettings(
	account *Account,
	validateBaseURL func(string) (string, error),
	settings GrokRuntimeSettings,
) (grokTransportTarget, error) {
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

		baseURL := strings.TrimSpace(account.GetCredential("base_url"))
		if baseURL == "" {
			baseURL = normalizeGrokRuntimeBaseURL(settings.OfficialBaseURL, CompatibleGatewayDefaultBaseURL(account.Platform))
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
		cookieHeader, err := BuildGrokSessionCookieHeader(
			account.GetGrokSessionToken(),
			account.GetGrokSessionCFCookies(),
			account.GetGrokSessionCFClearance(),
		)
		if err != nil {
			return grokTransportTarget{}, err
		}
		targetURL, sessionBaseURL, err := buildGrokSessionTransportURLWithSettings(validateBaseURL, settings)
		if err != nil {
			return grokTransportTarget{}, err
		}
		return grokTransportTarget{
			Kind:           grokTransportKindSession,
			URL:            targetURL,
			SessionBaseURL: sessionBaseURL,
			CookieHeader:   cookieHeader,
			UserAgent:      account.GetGrokSessionUserAgent(),
			AcceptLang:     account.GetGrokSessionAcceptLanguage(),
		}, nil
	default:
		return grokTransportTarget{}, fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func resolveGrokMediaTransportTarget(
	account *Account,
	validateBaseURL func(string) (string, error),
	requestPath string,
) (grokTransportTarget, error) {
	return resolveGrokMediaTransportTargetWithSettings(account, validateBaseURL, DefaultGrokRuntimeSettings(), requestPath)
}

func resolveGrokMediaTransportTargetWithSettings(
	account *Account,
	validateBaseURL func(string) (string, error),
	settings GrokRuntimeSettings,
	requestPath string,
) (grokTransportTarget, error) {
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

		baseURL := strings.TrimSpace(account.GetCredential("base_url"))
		if baseURL == "" {
			baseURL = normalizeGrokRuntimeBaseURL(settings.OfficialBaseURL, CompatibleGatewayDefaultBaseURL(account.Platform))
		}
		if validateBaseURL != nil {
			normalizedBaseURL, err := validateBaseURL(baseURL)
			if err != nil {
				return grokTransportTarget{}, err
			}
			baseURL = normalizedBaseURL
		}

		target := newCompatiblePassthroughUpstreamTargetWithOptions(
			baseURL,
			normalizeGrokMediaUpstreamPath(requestPath),
			account.GetCompatibleAuthMode(""),
			account.GetCompatibleEndpointOverride("responses"),
			account.GetCompatibleEndpointOverride("chat_completions"),
		)
		return grokTransportTarget{
			Kind:       grokTransportKindCompatible,
			URL:        target.URL,
			AuthToken:  apiKey,
			AuthHeader: target.AuthHeader,
		}, nil
	case AccountTypeSession:
		cookieHeader, err := BuildGrokSessionCookieHeader(
			account.GetGrokSessionToken(),
			account.GetGrokSessionCFCookies(),
			account.GetGrokSessionCFClearance(),
		)
		if err != nil {
			return grokTransportTarget{}, err
		}
		targetURL, sessionBaseURL, err := buildGrokSessionTransportURLForPathWithSettings(
			validateBaseURL,
			settings,
			normalizeGrokMediaUpstreamPath(requestPath),
		)
		if err != nil {
			return grokTransportTarget{}, err
		}
		return grokTransportTarget{
			Kind:           grokTransportKindSession,
			URL:            targetURL,
			SessionBaseURL: sessionBaseURL,
			CookieHeader:   cookieHeader,
			UserAgent:      account.GetGrokSessionUserAgent(),
			AcceptLang:     account.GetGrokSessionAcceptLanguage(),
		}, nil
	default:
		return grokTransportTarget{}, fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func buildGrokSessionTransportURL(validateBaseURL func(string) (string, error)) (string, error) {
	targetURL, _, err := buildGrokSessionTransportURLWithSettings(validateBaseURL, DefaultGrokRuntimeSettings())
	return targetURL, err
}

func buildGrokSessionTransportURLForPath(validateBaseURL func(string) (string, error), requestPath string) (string, error) {
	targetURL, _, err := buildGrokSessionTransportURLForPathWithSettings(
		validateBaseURL,
		DefaultGrokRuntimeSettings(),
		requestPath,
	)
	return targetURL, err
}

func buildGrokSessionTransportURLWithSettings(
	validateBaseURL func(string) (string, error),
	settings GrokRuntimeSettings,
) (string, string, error) {
	return buildGrokSessionTransportURLForPathWithSettings(validateBaseURL, settings, grokSessionConversationEndpoint)
}

func buildGrokSessionTransportURLForPathWithSettings(
	validateBaseURL func(string) (string, error),
	settings GrokRuntimeSettings,
	requestPath string,
) (string, string, error) {
	baseURL := normalizeGrokRuntimeBaseURL(settings.SessionBaseURL, grokWebBaseURL)
	if validateBaseURL != nil {
		normalizedBaseURL, err := validateBaseURL(baseURL)
		if err != nil {
			return "", "", err
		}
		baseURL = normalizedBaseURL
	}
	return appendGrokSessionTransportPath(baseURL, requestPath), baseURL, nil
}

func appendGrokSessionTransportPath(baseURL string, requestPath string) string {
	trimmedBase := strings.TrimSpace(baseURL)
	parsed, err := url.Parse(trimmedBase)
	if err != nil {
		return strings.TrimRight(trimmedBase, "/") + normalizeGrokSessionTransportPath(requestPath)
	}

	parsed.Path = strings.TrimRight(strings.TrimSpace(parsed.Path), "/") + normalizeGrokSessionTransportPath(requestPath)
	parsed.RawPath = ""
	return parsed.String()
}

func normalizeGrokSessionTransportPath(requestPath string) string {
	trimmed := strings.TrimSpace(requestPath)
	if trimmed == "" {
		return grokSessionConversationEndpoint
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}
