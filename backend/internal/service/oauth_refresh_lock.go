package service

import (
	"context"
	"log/slog"
	"time"
)

const oauthRefreshLockReleaseTimeout = 5 * time.Second

func releaseOAuthRefreshLock(cache GeminiTokenCache, cacheKey string) {
	if cache == nil || cacheKey == "" {
		return
	}

	releaseCtx, cancel := context.WithTimeout(context.Background(), oauthRefreshLockReleaseTimeout)
	defer cancel()

	if err := cache.ReleaseRefreshLock(releaseCtx, cacheKey); err != nil {
		slog.Warn("oauth_refresh_lock_release_failed", "cache_key", cacheKey, "error", err)
	}
}
