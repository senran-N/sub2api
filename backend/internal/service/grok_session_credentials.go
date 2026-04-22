package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const grokSessionFingerprintPrefix = "sha256:"
const minGrokSessionTokenLength = 24

func NormalizeGrokSessionCookieHeader(raw string) (string, error) {
	return BuildGrokSessionCookieHeader(raw, "", "")
}

func BuildGrokSessionCookieHeader(rawSessionToken string, rawCFCookies string, rawCFClearance string) (string, error) {
	jar, err := parseGrokSessionCookieJar(rawSessionToken)
	if err != nil {
		return "", err
	}

	sessionToken := firstNonEmptyGrokSessionCookieValue(
		jar.Get(grokSessionCookieNameSSO),
		jar.Get(grokSessionCookieNameSSORW),
	)
	if sessionToken == "" {
		return "", errors.New("missing sso cookie")
	}
	jar.Set(grokSessionCookieNameSSO, sessionToken)
	jar.Set(grokSessionCookieNameSSORW, firstNonEmptyGrokSessionCookieValue(jar.Get(grokSessionCookieNameSSORW), sessionToken))

	if strings.TrimSpace(rawCFCookies) != "" {
		extraCookies, err := parseOptionalGrokSessionCookieHeader(rawCFCookies)
		if err != nil {
			return "", fmt.Errorf("invalid cf_cookies: %w", err)
		}
		for _, name := range extraCookies.Order() {
			if strings.EqualFold(name, grokSessionCookieNameSSO) || strings.EqualFold(name, grokSessionCookieNameSSORW) {
				continue
			}
			jar.Set(name, extraCookies.Get(name))
		}
	}

	clearance := strings.TrimSpace(rawCFClearance)
	if clearance == "" {
		clearance = jar.Get(grokSessionCookieNameCFClearance)
	}
	if clearance != "" {
		jar.Set(grokSessionCookieNameCFClearance, clearance)
	}

	normalized := jar.HeaderString()
	if normalized == "" {
		return "", errors.New("session_token does not contain a valid Grok cookie")
	}
	return normalized, nil
}

func ValidateGrokSessionImportToken(raw string) (string, error) {
	normalized, err := NormalizeGrokSessionCookieHeader(raw)
	if err != nil {
		return "", err
	}
	jar, err := parseOptionalGrokSessionCookieHeader(normalized)
	if err != nil {
		return "", err
	}

	sessionToken := firstNonEmptyGrokSessionCookieValue(
		jar.Get(grokSessionCookieNameSSO),
		jar.Get(grokSessionCookieNameSSORW),
	)
	if sessionToken == "" {
		return "", errors.New("missing sso cookie")
	}
	if err := validateGrokSessionPrimaryTokenValue(sessionToken); err != nil {
		return "", err
	}
	return normalized, nil
}

func FingerprintGrokSessionToken(normalizedCookieHeader string) string {
	normalizedCookieHeader = strings.TrimSpace(normalizedCookieHeader)
	if normalizedCookieHeader == "" {
		return ""
	}
	fingerprintSource := resolveGrokSessionFingerprintSource(normalizedCookieHeader)
	if fingerprintSource == "" {
		fingerprintSource = normalizedCookieHeader
	}
	sum := sha256.Sum256([]byte(fingerprintSource))
	return grokSessionFingerprintPrefix + hex.EncodeToString(sum[:])
}

func MaskGrokSessionFingerprint(fingerprint string) string {
	fingerprint = strings.TrimSpace(fingerprint)
	if fingerprint == "" {
		return ""
	}
	if !strings.HasPrefix(fingerprint, grokSessionFingerprintPrefix) {
		if len(fingerprint) <= 12 {
			return fingerprint
		}
		return fingerprint[:8] + "..." + fingerprint[len(fingerprint)-4:]
	}

	hexPart := strings.TrimPrefix(fingerprint, grokSessionFingerprintPrefix)
	if len(hexPart) <= 12 {
		return fingerprint
	}
	return grokSessionFingerprintPrefix + hexPart[:8] + "..." + hexPart[len(hexPart)-4:]
}

func grokSessionCookieHeaderHasSSO(cookieHeader string) bool {
	jar, err := parseOptionalGrokSessionCookieHeader(cookieHeader)
	if err != nil {
		return false
	}
	return firstNonEmptyGrokSessionCookieValue(
		jar.Get(grokSessionCookieNameSSO),
		jar.Get(grokSessionCookieNameSSORW),
	) != ""
}

func validateGrokSessionPrimaryTokenValue(token string) error {
	normalized := strings.TrimSpace(token)
	if normalized == "" {
		return errors.New("missing sso cookie")
	}
	if len(normalized) < minGrokSessionTokenLength {
		return errors.New("grok session token format is invalid")
	}
	if strings.ContainsAny(normalized, " \t\r\n;") {
		return errors.New("grok session token format is invalid")
	}
	return nil
}

const (
	grokSessionCookieNameSSO         = "sso"
	grokSessionCookieNameSSORW       = "sso-rw"
	grokSessionCookieNameCFClearance = "cf_clearance"
)

type grokSessionCookieJar struct {
	values map[string]string
	order  []string
}

func parseGrokSessionCookieJar(raw string) (*grokSessionCookieJar, error) {
	trimmed := trimGrokSessionCookieHeaderPrefix(raw)
	if trimmed == "" {
		return nil, errors.New("session_token not found in credentials")
	}
	if !strings.Contains(trimmed, "=") {
		jar := newGrokSessionCookieJar()
		jar.Set(grokSessionCookieNameSSO, trimmed)
		return jar, nil
	}
	return parseOptionalGrokSessionCookieHeader(trimmed)
}

func parseOptionalGrokSessionCookieHeader(raw string) (*grokSessionCookieJar, error) {
	trimmed := trimGrokSessionCookieHeaderPrefix(raw)
	if trimmed == "" {
		return newGrokSessionCookieJar(), nil
	}

	jar := newGrokSessionCookieJar()
	for _, part := range strings.Split(trimmed, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		separator := strings.Index(part, "=")
		if separator <= 0 {
			continue
		}
		name := strings.TrimSpace(part[:separator])
		value := strings.TrimSpace(part[separator+1:])
		if name == "" || value == "" {
			continue
		}
		jar.Set(name, value)
	}
	if len(jar.values) == 0 {
		return nil, errors.New("session_token does not contain a valid Grok cookie")
	}
	return jar, nil
}

func newGrokSessionCookieJar() *grokSessionCookieJar {
	return &grokSessionCookieJar{
		values: make(map[string]string),
		order:  make([]string, 0, 4),
	}
}

func (j *grokSessionCookieJar) Set(name string, value string) {
	if j == nil {
		return
	}
	normalizedName := strings.TrimSpace(name)
	normalizedValue := strings.TrimSpace(value)
	if normalizedName == "" || normalizedValue == "" {
		return
	}
	key := strings.ToLower(normalizedName)
	if j.values == nil {
		j.values = make(map[string]string)
	}
	if _, exists := j.values[key]; !exists {
		j.order = append(j.order, key)
	}
	j.values[key] = normalizedValue
}

func (j *grokSessionCookieJar) Get(name string) string {
	if j == nil || len(j.values) == 0 {
		return ""
	}
	return strings.TrimSpace(j.values[strings.ToLower(strings.TrimSpace(name))])
}

func (j *grokSessionCookieJar) Order() []string {
	if j == nil || len(j.order) == 0 {
		return nil
	}
	out := make([]string, 0, len(j.order))
	out = append(out, j.order...)
	return out
}

func (j *grokSessionCookieJar) HeaderString() string {
	if j == nil || len(j.values) == 0 {
		return ""
	}

	ordered := make([]string, 0, len(j.values))
	appendCookieName := func(name string) {
		value := j.Get(name)
		if value == "" {
			return
		}
		ordered = append(ordered, strings.ToLower(strings.TrimSpace(name)))
	}

	appendCookieName(grokSessionCookieNameSSO)
	appendCookieName(grokSessionCookieNameSSORW)
	for _, name := range j.order {
		lowerName := strings.ToLower(strings.TrimSpace(name))
		if lowerName == grokSessionCookieNameSSO || lowerName == grokSessionCookieNameSSORW {
			continue
		}
		appendCookieName(lowerName)
	}

	parts := make([]string, 0, len(ordered))
	seen := make(map[string]struct{}, len(ordered))
	for _, name := range ordered {
		if _, ok := seen[name]; ok {
			continue
		}
		value := j.Get(name)
		if value == "" {
			continue
		}
		parts = append(parts, name+"="+value)
		seen[name] = struct{}{}
	}
	return strings.Join(parts, "; ")
}

func trimGrokSessionCookieHeaderPrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "Cookie:")
	trimmed = strings.TrimPrefix(trimmed, "cookie:")
	return strings.TrimSpace(trimmed)
}

func firstNonEmptyGrokSessionCookieValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func resolveGrokSessionFingerprintSource(cookieHeader string) string {
	trimmed := strings.TrimSpace(cookieHeader)
	if trimmed == "" {
		return ""
	}

	jar, err := parseGrokSessionCookieJar(trimmed)
	if err != nil {
		return trimmed
	}
	return firstNonEmptyGrokSessionCookieValue(
		jar.Get(grokSessionCookieNameSSO),
		jar.Get(grokSessionCookieNameSSORW),
	)
}
