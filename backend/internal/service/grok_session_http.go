package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const grokSessionProbeUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
const grokSessionDefaultAcceptLanguage = "zh-CN,zh;q=0.9,en;q=0.8"
const grokSessionDefaultStatsigID = "ZTpUeXBlRXJyb3I6IENhbm5vdCByZWFkIHByb3BlcnRpZXMgb2YgdW5kZWZpbmVkIChyZWFkaW5nICdjaGlsZE5vZGVzJyk="

var grokSessionBrowserVersionPattern = regexp.MustCompile(`(?:chrome|chromium|crios|edg|brave)/(\d{2,3})`)

func applyGrokSessionBrowserHeaders(header http.Header, target grokTransportTarget, accept string) {
	if header == nil {
		return
	}
	if value := strings.TrimSpace(accept); value != "" {
		header.Set("Accept", value)
	}

	userAgent := grokSessionResolvedUserAgent(target)
	acceptLanguage := grokSessionResolvedAcceptLanguage(target)
	sessionBaseURL := firstNonEmptyGrokSessionHeaderValue(target.SessionBaseURL, grokWebBaseURL)

	header.Set("Accept-Language", acceptLanguage)
	header.Set("Origin", sessionBaseURL)
	header.Set("Referer", strings.TrimRight(sessionBaseURL, "/")+"/")
	header.Set("Priority", "u=1, i")
	header.Set("Sec-Fetch-Dest", grokSessionResolvedFetchDest(firstNonEmptyGrokSessionHeaderValue(strings.TrimSpace(accept), header.Get("Accept"))))
	header.Set("Sec-Fetch-Mode", "cors")
	header.Set("Sec-Fetch-Site", "same-origin")
	header.Set("User-Agent", userAgent)
	header.Set("x-statsig-id", grokSessionDefaultStatsigID)
	header.Set("x-xai-request-id", uuid.NewString())
	applyGrokSessionClientHints(header, userAgent)
}

func newGrokSessionJSONRequest(
	ctx context.Context,
	method string,
	target grokTransportTarget,
	payload []byte,
	accept string,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, target.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	applyGrokSessionBrowserHeaders(req.Header, target, accept)
	target.Apply(req)
	return req, nil
}

func grokSessionResolvedUserAgent(target grokTransportTarget) string {
	return firstNonEmptyGrokSessionHeaderValue(target.UserAgent, grokSessionProbeUserAgent)
}

func grokSessionResolvedAcceptLanguage(target grokTransportTarget) string {
	return firstNonEmptyGrokSessionHeaderValue(target.AcceptLang, grokSessionDefaultAcceptLanguage)
}

func grokSessionResolvedFetchDest(accept string) string {
	normalizedAccept := strings.ToLower(strings.TrimSpace(accept))
	if normalizedAccept == "" {
		return "empty"
	}
	if strings.Contains(normalizedAccept, "image/") || strings.Contains(normalizedAccept, "video/") || strings.Contains(normalizedAccept, "text/html") {
		return "document"
	}
	return "empty"
}

func applyGrokSessionClientHints(header http.Header, userAgent string) {
	if header == nil {
		return
	}

	brand, version := grokSessionChromiumBrandAndVersion(userAgent)
	if brand == "" || version == "" {
		return
	}

	header.Set("Sec-Ch-Ua", fmt.Sprintf(`"%s";v="%s", "Chromium";v="%s", "Not(A:Brand";v="24"`, brand, version, version))
	header.Set("Sec-Ch-Ua-Mobile", grokSessionClientHintMobile(userAgent))
	header.Set("Sec-Ch-Ua-Model", "")
	if platform := grokSessionClientHintPlatform(userAgent); platform != "" {
		header.Set("Sec-Ch-Ua-Platform", fmt.Sprintf(`"%s"`, platform))
	}
	if arch := grokSessionClientHintArch(userAgent); arch != "" {
		header.Set("Sec-Ch-Ua-Arch", arch)
		header.Set("Sec-Ch-Ua-Bitness", "64")
	}
}

func grokSessionChromiumBrandAndVersion(userAgent string) (string, string) {
	normalized := strings.ToLower(strings.TrimSpace(userAgent))
	if normalized == "" {
		return "", ""
	}
	if strings.Contains(normalized, "firefox/") || (strings.Contains(normalized, "safari/") && !strings.Contains(normalized, "chrome/") && !strings.Contains(normalized, "chromium/") && !strings.Contains(normalized, "edg/")) {
		return "", ""
	}

	match := grokSessionBrowserVersionPattern.FindStringSubmatch(normalized)
	if len(match) < 2 {
		return "", ""
	}

	switch {
	case strings.Contains(normalized, "edg/"):
		return "Microsoft Edge", match[1]
	case strings.Contains(normalized, "brave"):
		return "Brave", match[1]
	case strings.Contains(normalized, "chromium/"):
		return "Chromium", match[1]
	default:
		return "Google Chrome", match[1]
	}
}

func grokSessionClientHintPlatform(userAgent string) string {
	normalized := strings.ToLower(userAgent)
	switch {
	case strings.Contains(normalized, "windows"):
		return "Windows"
	case strings.Contains(normalized, "mac os x"), strings.Contains(normalized, "macintosh"):
		return "macOS"
	case strings.Contains(normalized, "android"):
		return "Android"
	case strings.Contains(normalized, "iphone"), strings.Contains(normalized, "ipad"):
		return "iOS"
	case strings.Contains(normalized, "linux"):
		return "Linux"
	default:
		return ""
	}
}

func grokSessionClientHintArch(userAgent string) string {
	normalized := strings.ToLower(userAgent)
	switch {
	case strings.Contains(normalized, "aarch64"), strings.Contains(normalized, " arm"):
		return "arm"
	case strings.Contains(normalized, "x86_64"),
		strings.Contains(normalized, "x64"),
		strings.Contains(normalized, "win64"),
		strings.Contains(normalized, "intel"):
		return "x86"
	default:
		return ""
	}
}

func grokSessionClientHintMobile(userAgent string) string {
	normalized := strings.ToLower(userAgent)
	if strings.Contains(normalized, "mobile") || strings.Contains(normalized, "android") || strings.Contains(normalized, "iphone") || strings.Contains(normalized, "ipad") {
		return "?1"
	}
	return "?0"
}

func firstNonEmptyGrokSessionHeaderValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
