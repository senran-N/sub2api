package ports

import (
	"context"
	"time"
)

// GeminiTokenCache stores short-lived Gemini access tokens and refresh locks.
type GeminiTokenCache interface {
	GetAccessToken(ctx context.Context, cacheKey string) (string, error)
	SetAccessToken(ctx context.Context, cacheKey string, token string, ttl time.Duration) error
	DeleteAccessToken(ctx context.Context, cacheKey string) error

	AcquireRefreshLock(ctx context.Context, cacheKey string, ttl time.Duration) (bool, error)
	ReleaseRefreshLock(ctx context.Context, cacheKey string) error
}
