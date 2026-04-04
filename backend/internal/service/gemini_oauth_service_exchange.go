package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/geminicli"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *GeminiOAuthService) ExchangeCode(ctx context.Context, input *GeminiExchangeCodeInput) (*GeminiTokenInfo, error) {
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ========== ExchangeCode START ==========")
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] SessionID: %s", input.SessionID)

	session, ok := s.sessionStore.Get(input.SessionID)
	if !ok {
		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ERROR: Session not found or expired")
		return nil, fmt.Errorf("session not found or expired")
	}
	if strings.TrimSpace(input.State) == "" || input.State != session.State {
		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ERROR: Invalid state")
		return nil, fmt.Errorf("invalid state")
	}

	proxyURL := s.resolveProxyURL(ctx, input.ProxyID, session.ProxyURL)
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ProxyURL: %s", proxyURL)

	redirectURI := session.RedirectURI
	oauthType := geminiOAuthTypeOrDefault(session.OAuthType)
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] OAuth Type: %s", oauthType)
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Project ID from session: %s", session.ProjectID)

	if oauthType == "ai_studio" {
		effectiveCfg, err := s.effectiveOAuthConfig("ai_studio")
		if err != nil {
			return nil, err
		}
		if effectiveCfg.ClientID == geminicli.GeminiCLIOAuthClientID {
			return nil, fmt.Errorf("AI Studio OAuth requires a custom OAuth Client. Please use an AI Studio API Key account, or configure GEMINI_OAUTH_CLIENT_ID / GEMINI_OAUTH_CLIENT_SECRET and re-authorize")
		}
	}

	if oauthType == "code_assist" || oauthType == "google_one" {
		redirectURI = geminicli.GeminiCLIRedirectURI
	}

	tokenResp, err := s.oauthClient.ExchangeCode(ctx, oauthType, input.Code, session.CodeVerifier, redirectURI, proxyURL)
	if err != nil {
		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ERROR: Failed to exchange code: %v", err)
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Token exchange successful")
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Token scope: %s", tokenResp.Scope)
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Token expires_in: %d seconds", tokenResp.ExpiresIn)

	projectID := strings.TrimSpace(session.ProjectID)
	s.sessionStore.Delete(input.SessionID)

	tierID := ""
	fallbackTierID := canonicalGeminiTierIDForOAuthType(oauthType, input.TierID)
	if fallbackTierID == "" {
		fallbackTierID = canonicalGeminiTierIDForOAuthType(oauthType, session.TierID)
	}

	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ========== Account Type Detection START ==========")
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] OAuth Type: %s", oauthType)

	switch oauthType {
	case "code_assist":
		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Processing code_assist OAuth type")
		if projectID == "" {
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] No project_id provided, attempting to fetch from LoadCodeAssist API...")
			projectID, tierID, err = s.fetchProjectID(ctx, tokenResp.AccessToken, proxyURL)
			if err != nil {
				fmt.Printf("[GeminiOAuth] Warning: Failed to fetch project_id during token exchange: %v\n", err)
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] WARNING: Failed to fetch project_id: %v", err)
			} else {
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Successfully fetched project_id: %s, tier_id: %s", projectID, tierID)
			}
		} else {
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] User provided project_id: %s, fetching tier_id...", projectID)
			_, fetchedTierID, fetchErr := s.fetchProjectID(ctx, tokenResp.AccessToken, proxyURL)
			if fetchErr != nil {
				fmt.Printf("[GeminiOAuth] Warning: Failed to fetch tierID: %v\n", fetchErr)
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] WARNING: Failed to fetch tier_id: %v", fetchErr)
			} else {
				tierID = fetchedTierID
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Successfully fetched tier_id: %s", tierID)
			}
		}
		if strings.TrimSpace(projectID) == "" {
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ERROR: Missing project_id for Code Assist OAuth")
			return nil, fmt.Errorf("missing project_id for Code Assist OAuth: please fill Project ID (optional field) and regenerate the auth URL, or ensure your Google account has an ACTIVE GCP project")
		}

		tierID = canonicalGeminiTierIDForOAuthType(oauthType, tierID)
		if tierID == "" {
			if fallbackTierID != "" {
				tierID = fallbackTierID
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Using fallback tier_id from user/session: %s", tierID)
			} else {
				tierID = GeminiTierGCPStandard
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Using default tier_id: %s", tierID)
			}
		}
		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Final code_assist result - project_id: %s, tier_id: %s", projectID, tierID)

	case "google_one":
		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Processing google_one OAuth type")
		if projectID == "" {
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] No project_id provided, attempting to fetch from LoadCodeAssist API...")
			projectID, _, err = s.fetchProjectID(ctx, tokenResp.AccessToken, proxyURL)
			if err != nil {
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ERROR: Failed to fetch project_id: %v", err)
				return nil, fmt.Errorf("google One accounts require a project_id, failed to auto-detect: %w", err)
			}
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Successfully fetched project_id: %s", projectID)
		}

		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Attempting to fetch Google One tier from Drive API...")
		var storageInfoExtra map[string]any
		detectedTierID, storageInfo, fetchErr := s.FetchGoogleOneTier(ctx, tokenResp.AccessToken, proxyURL)
		if fetchErr != nil {
			fmt.Printf("[GeminiOAuth] Warning: Failed to fetch Drive tier: %v\n", fetchErr)
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] WARNING: Failed to fetch Drive tier: %v", fetchErr)
		} else {
			tierID = detectedTierID
			storageInfoExtra = buildGeminiDriveExtra(storageInfo)
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Successfully fetched Drive tier: %s", tierID)
			if storageInfo != nil {
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Drive storage - Limit: %d bytes (%.2f TB), Usage: %d bytes (%.2f GB)",
					storageInfo.Limit, float64(storageInfo.Limit)/float64(TB),
					storageInfo.Usage, float64(storageInfo.Usage)/float64(GB))
			}
		}

		tierID = canonicalGeminiTierIDForOAuthType(oauthType, tierID)
		if tierID == "" || tierID == GeminiTierGoogleOneUnknown {
			if fallbackTierID != "" {
				tierID = fallbackTierID
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Using fallback tier_id from user/session: %s", tierID)
			} else {
				tierID = GeminiTierGoogleOneFree
				logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Using default tier_id: %s", tierID)
			}
		}
		fmt.Printf("[GeminiOAuth] Google One tierID after normalization: %s\n", tierID)

		if len(storageInfoExtra) > 0 {
			return &GeminiTokenInfo{
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: tokenResp.RefreshToken,
				TokenType:    tokenResp.TokenType,
				ExpiresIn:    tokenResp.ExpiresIn,
				ExpiresAt:    geminiTokenExpiresAt(tokenResp.ExpiresIn),
				Scope:        tokenResp.Scope,
				ProjectID:    projectID,
				TierID:       tierID,
				OAuthType:    oauthType,
				Extra:        storageInfoExtra,
			}, nil
		}

	case "ai_studio":
		if fallbackTierID != "" {
			tierID = fallbackTierID
		} else {
			tierID = GeminiTierAIStudioFree
		}

	default:
		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Processing %s OAuth type (no tier detection)", oauthType)
	}

	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ========== Account Type Detection END ==========")

	result := &GeminiTokenInfo{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    geminiTokenExpiresAt(tokenResp.ExpiresIn),
		Scope:        tokenResp.Scope,
		ProjectID:    projectID,
		TierID:       tierID,
		OAuthType:    oauthType,
	}
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Final result - OAuth Type: %s, Project ID: %s, Tier ID: %s", result.OAuthType, result.ProjectID, result.TierID)
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] ========== ExchangeCode END ==========")
	return result, nil
}
