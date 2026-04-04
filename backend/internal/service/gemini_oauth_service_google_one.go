package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/geminicli"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func buildGeminiDriveExtra(storageInfo *geminicli.DriveStorageInfo) map[string]any {
	if storageInfo == nil {
		return nil
	}

	return map[string]any{
		"drive_storage_limit":   storageInfo.Limit,
		"drive_storage_usage":   storageInfo.Usage,
		"drive_tier_updated_at": time.Now().Format(time.RFC3339),
	}
}

func (s *GeminiOAuthService) FetchGoogleOneTier(ctx context.Context, accessToken, proxyURL string) (string, *geminicli.DriveStorageInfo, error) {
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Starting FetchGoogleOneTier (Google One personal account)")
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Calling Drive API for storage quota...")

	storageInfo, err := s.driveClient.GetStorageQuota(ctx, accessToken, proxyURL)
	if err != nil {
		if strings.Contains(err.Error(), "status 403") {
			logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Drive API scope not available (403): %v", err)
			return GeminiTierGoogleOneUnknown, nil, err
		}

		logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Failed to fetch Drive storage: %v", err)
		return GeminiTierGoogleOneUnknown, nil, err
	}

	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Drive API response - Limit: %d bytes (%.2f TB), Usage: %d bytes (%.2f GB)",
		storageInfo.Limit, float64(storageInfo.Limit)/float64(TB),
		storageInfo.Usage, float64(storageInfo.Usage)/float64(GB))

	tierID := inferGoogleOneTier(storageInfo.Limit)
	logger.LegacyPrintf("service.gemini_oauth", "[GeminiOAuth] Inferred tier from storage: %s", tierID)

	return tierID, storageInfo, nil
}

func (s *GeminiOAuthService) RefreshAccountGoogleOneTier(ctx context.Context, account *Account) (tierID string, extra map[string]any, credentials map[string]any, err error) {
	if account == nil {
		return "", nil, nil, fmt.Errorf("account is nil")
	}

	oauthType, ok := account.Credentials["oauth_type"].(string)
	if !ok || oauthType != "google_one" {
		return "", nil, nil, fmt.Errorf("not a google_one OAuth account")
	}

	accessToken, ok := account.Credentials["access_token"].(string)
	if !ok || accessToken == "" {
		return "", nil, nil, fmt.Errorf("missing access_token")
	}

	tierID, storageInfo, err := s.FetchGoogleOneTier(ctx, accessToken, s.resolveAccountProxyURL(ctx, account))
	if err != nil {
		return "", nil, nil, err
	}

	extra = copyMap(account.Extra)
	for key, value := range buildGeminiDriveExtra(storageInfo) {
		extra[key] = value
	}

	credentials = copyMap(account.Credentials)
	credentials["tier_id"] = tierID

	return tierID, extra, credentials, nil
}
