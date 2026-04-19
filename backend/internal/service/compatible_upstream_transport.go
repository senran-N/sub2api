package service

import (
	"net/http"
	"net/url"
	"strings"
)

func applyCompatibleAuthHeaders(header http.Header, token, authMode string) {
	if header == nil {
		return
	}

	header.Del(compatibleUpstreamAuthHeaderAuthorization)
	header.Del(compatibleUpstreamAuthHeaderAPIKey)
	header.Del("x-api-key")

	switch normalizeCompatibleAuthMode(authMode) {
	case UpstreamAuthModeXAPIKey:
		header.Set("x-api-key", token)
	case UpstreamAuthModeAPIKey:
		header.Set(compatibleUpstreamAuthHeaderAPIKey, token)
	case UpstreamAuthModeDual:
		header.Set(compatibleUpstreamAuthHeaderAuthorization, "Bearer "+token)
		header.Set("x-api-key", token)
	default:
		header.Set(compatibleUpstreamAuthHeaderAuthorization, "Bearer "+token)
	}
}

func resolveCompatibleEndpointURL(base, defaultPath, endpointOverride string) string {
	normalizedBase := strings.TrimSpace(base)
	if normalizedBase == "" {
		if strings.TrimSpace(endpointOverride) != "" {
			return strings.TrimSpace(endpointOverride)
		}
		return strings.TrimSpace(defaultPath)
	}

	override := strings.TrimSpace(endpointOverride)
	if override == "" {
		parsed, err := url.Parse(normalizedBase)
		if err != nil {
			return strings.TrimRight(normalizedBase, "/") + defaultPath
		}
		parsed.Path = appendCompatibleDefaultPath(parsed.Path, defaultPath)
		parsed.RawPath = ""
		return parsed.String()
	}

	if parsedOverride, err := url.Parse(override); err == nil && parsedOverride.IsAbs() {
		return parsedOverride.String()
	}

	parsedBase, err := url.Parse(normalizedBase)
	if err != nil {
		if strings.HasPrefix(override, "/") {
			return strings.TrimRight(normalizedBase, "/") + override
		}
		return strings.TrimRight(normalizedBase, "/") + "/" + override
	}

	parsedOverride, err := url.Parse(override)
	if err != nil {
		if strings.HasPrefix(override, "/") {
			return strings.TrimRight(normalizedBase, "/") + override
		}
		return strings.TrimRight(normalizedBase, "/") + "/" + override
	}

	if strings.HasPrefix(override, "/") {
		parsedBase.Path = parsedOverride.Path
	} else {
		parsedBase.Path = appendCompatibleDefaultPath(parsedBase.Path, parsedOverride.Path)
	}
	parsedBase.RawPath = ""
	parsedBase.RawQuery = parsedOverride.RawQuery
	parsedBase.Fragment = parsedOverride.Fragment
	return parsedBase.String()
}

func appendCompatibleDefaultPath(basePath, defaultPath string) string {
	trimmedBase := strings.TrimRight(strings.TrimSpace(basePath), "/")
	trimmedDefault := strings.TrimSpace(defaultPath)
	if trimmedBase == "" {
		if strings.HasPrefix(trimmedDefault, "/") {
			return trimmedDefault
		}
		return "/" + trimmedDefault
	}
	if trimmedDefault == "" {
		return trimmedBase
	}

	trimmedDefault = strings.TrimLeft(trimmedDefault, "/")
	if trimmedDefault == "" {
		return trimmedBase
	}
	if trimmedBase == "/"+trimmedDefault || strings.HasSuffix(trimmedBase, "/"+trimmedDefault) {
		return trimmedBase
	}
	if strings.HasSuffix(trimmedBase, "/v1") && strings.HasPrefix(trimmedDefault, "v1/") {
		return trimmedBase + "/" + strings.TrimPrefix(trimmedDefault, "v1/")
	}
	return trimmedBase + "/" + trimmedDefault
}
