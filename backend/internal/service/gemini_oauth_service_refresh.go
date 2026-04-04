package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (s *GeminiOAuthService) RefreshToken(ctx context.Context, oauthType, refreshToken, proxyURL string) (*GeminiTokenInfo, error) {
	var lastErr error

	for attempt := 0; attempt <= 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			time.Sleep(backoff)
		}

		tokenResp, err := s.oauthClient.RefreshToken(ctx, oauthType, refreshToken, proxyURL)
		if err == nil {
			return &GeminiTokenInfo{
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: tokenResp.RefreshToken,
				TokenType:    tokenResp.TokenType,
				ExpiresIn:    tokenResp.ExpiresIn,
				ExpiresAt:    geminiTokenExpiresAt(tokenResp.ExpiresIn),
				Scope:        tokenResp.Scope,
			}, nil
		}

		if isNonRetryableGeminiOAuthError(err) {
			return nil, err
		}
		lastErr = err
	}

	return nil, fmt.Errorf("token refresh failed after retries: %w", lastErr)
}

func isNonRetryableGeminiOAuthError(err error) bool {
	msg := err.Error()
	for _, needle := range []string{"invalid_grant", "invalid_client", "unauthorized_client", "access_denied"} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func (s *GeminiOAuthService) RefreshAccountToken(ctx context.Context, account *Account) (*GeminiTokenInfo, error) {
	if account.Platform != PlatformGemini || account.Type != AccountTypeOAuth {
		return nil, fmt.Errorf("account is not a Gemini OAuth account")
	}

	refreshToken := account.GetCredential("refresh_token")
	if strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	oauthType := geminiOAuthTypeOrDefault(account.GetCredential("oauth_type"))
	proxyURL := s.resolveAccountProxyURL(ctx, account)

	tokenInfo, err := s.RefreshToken(ctx, oauthType, refreshToken, proxyURL)
	if err != nil && oauthType == "code_assist" && strings.Contains(err.Error(), "unauthorized_client") && s.GetOAuthConfig().AIStudioOAuthEnabled {
		if alt, altErr := s.RefreshToken(ctx, "ai_studio", refreshToken, proxyURL); altErr == nil {
			tokenInfo = alt
			err = nil
		}
	}
	if err != nil && oauthType == "google_one" && strings.Contains(err.Error(), "unauthorized_client") && s.GetOAuthConfig().AIStudioOAuthEnabled {
		if alt, altErr := s.RefreshToken(ctx, "code_assist", refreshToken, proxyURL); altErr == nil {
			tokenInfo = alt
			err = nil
		}
	}
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized_client") {
			return nil, fmt.Errorf("%w (OAuth client mismatch: the refresh_token is bound to the OAuth client used during authorization; please re-authorize this account or restore the original GEMINI_OAUTH_CLIENT_ID/SECRET)", err)
		}
		return nil, err
	}

	tokenInfo.OAuthType = oauthType

	existingProjectID := strings.TrimSpace(account.GetCredential("project_id"))
	if existingProjectID != "" {
		tokenInfo.ProjectID = existingProjectID
	}

	existingTierID := strings.TrimSpace(account.GetCredential("tier_id"))

	switch oauthType {
	case "code_assist":
		if existingTierID != "" {
			tokenInfo.TierID = canonicalGeminiTierIDForOAuthType(oauthType, existingTierID)
		}
		if tokenInfo.TierID == "" {
			tokenInfo.TierID = GeminiTierGCPStandard
		}

		if strings.TrimSpace(tokenInfo.ProjectID) == "" || tokenInfo.TierID == "" {
			projectID, tierID, detectErr := s.fetchProjectID(ctx, tokenInfo.AccessToken, proxyURL)
			if detectErr != nil {
				fmt.Printf("[GeminiOAuth] Warning: failed to auto-detect project/tier: %v\n", detectErr)
			} else {
				if strings.TrimSpace(tokenInfo.ProjectID) == "" && projectID != "" {
					tokenInfo.ProjectID = projectID
				}
				if canonical := canonicalGeminiTierIDForOAuthType(oauthType, tierID); canonical != "" {
					tokenInfo.TierID = canonical
				}
			}
		}

		if strings.TrimSpace(tokenInfo.ProjectID) == "" {
			return nil, fmt.Errorf("failed to auto-detect project_id: empty result")
		}

	case "google_one":
		canonicalExistingTier := canonicalGeminiTierIDForOAuthType(oauthType, existingTierID)
		needsRefresh := true
		if account.Extra != nil {
			if updatedAtStr, ok := account.Extra["drive_tier_updated_at"].(string); ok {
				if updatedAt, parseErr := time.Parse(time.RFC3339, updatedAtStr); parseErr == nil && time.Since(updatedAt) <= 24*time.Hour {
					needsRefresh = false
					tokenInfo.TierID = canonicalExistingTier
				}
			}
		}

		if tokenInfo.TierID == "" {
			tokenInfo.TierID = canonicalExistingTier
		}

		if needsRefresh {
			tierID, storageInfo, fetchErr := s.FetchGoogleOneTier(ctx, tokenInfo.AccessToken, proxyURL)
			if fetchErr == nil {
				if canonical := canonicalGeminiTierIDForOAuthType(oauthType, tierID); canonical != "" && canonical != GeminiTierGoogleOneUnknown {
					tokenInfo.TierID = canonical
				}
				if extra := buildGeminiDriveExtra(storageInfo); len(extra) > 0 {
					tokenInfo.Extra = extra
				}
			}
		}

		if tokenInfo.TierID == "" || tokenInfo.TierID == GeminiTierGoogleOneUnknown {
			if canonicalExistingTier != "" {
				tokenInfo.TierID = canonicalExistingTier
			} else {
				tokenInfo.TierID = GeminiTierGoogleOneFree
			}
		}
	}

	return tokenInfo, nil
}

func (s *GeminiOAuthService) BuildAccountCredentials(tokenInfo *GeminiTokenInfo) map[string]any {
	creds := map[string]any{
		"access_token": tokenInfo.AccessToken,
		"expires_at":   strconv.FormatInt(tokenInfo.ExpiresAt, 10),
	}
	if tokenInfo.RefreshToken != "" {
		creds["refresh_token"] = tokenInfo.RefreshToken
	}
	if tokenInfo.TokenType != "" {
		creds["token_type"] = tokenInfo.TokenType
	}
	if tokenInfo.Scope != "" {
		creds["scope"] = tokenInfo.Scope
	}
	if tokenInfo.ProjectID != "" {
		creds["project_id"] = tokenInfo.ProjectID
	}
	if tokenInfo.TierID != "" {
		if err := validateTierID(tokenInfo.TierID); err == nil {
			creds["tier_id"] = tokenInfo.TierID
			fmt.Printf("[GeminiOAuth] Storing tier_id: %s\n", tokenInfo.TierID)
		} else {
			fmt.Printf("[GeminiOAuth] Invalid tier_id %s: %v\n", tokenInfo.TierID, err)
		}
	}
	if tokenInfo.OAuthType != "" {
		creds["oauth_type"] = tokenInfo.OAuthType
	}
	for key, value := range tokenInfo.Extra {
		creds[key] = value
	}
	return creds
}
