package service

import (
	"context"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/geminicli"
)

func (s *GeminiOAuthService) configuredOAuthConfig() geminicli.OAuthConfig {
	return geminicli.OAuthConfig{
		ClientID:     s.cfg.Gemini.OAuth.ClientID,
		ClientSecret: s.cfg.Gemini.OAuth.ClientSecret,
		Scopes:       s.cfg.Gemini.OAuth.Scopes,
	}
}

func (s *GeminiOAuthService) effectiveOAuthConfig(oauthType string) (geminicli.OAuthConfig, error) {
	return geminicli.EffectiveOAuthConfig(s.configuredOAuthConfig(), oauthType)
}

func (s *GeminiOAuthService) resolveProxyURL(ctx context.Context, proxyID *int64, fallback string) string {
	if proxyID == nil {
		return fallback
	}
	if s.proxyRepo == nil {
		return fallback
	}

	proxy, err := s.proxyRepo.GetByID(ctx, *proxyID)
	if err != nil || proxy == nil {
		return fallback
	}

	return proxy.URL()
}

func (s *GeminiOAuthService) resolveAccountProxyURL(ctx context.Context, account *Account) string {
	if account == nil {
		return ""
	}

	fallback := ""
	if account.Proxy != nil {
		fallback = account.Proxy.URL()
	}
	if account.ProxyID == nil {
		return fallback
	}

	return s.resolveProxyURL(ctx, account.ProxyID, fallback)
}

func geminiOAuthTypeOrDefault(oauthType string) string {
	oauthType = strings.TrimSpace(oauthType)
	if oauthType == "" {
		return "code_assist"
	}
	return oauthType
}

func geminiTokenExpiresAt(expiresIn int64) int64 {
	const safetyWindow = 300
	const minTTL = 30

	expiresAt := time.Now().Unix() + expiresIn - safetyWindow
	minExpiresAt := time.Now().Unix() + minTTL
	if expiresAt < minExpiresAt {
		return minExpiresAt
	}

	return expiresAt
}
