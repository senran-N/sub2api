package ports

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
)

// APIKeyCache stores auth snapshots and API key related Redis counters.
type APIKeyCache interface {
	GetCreateAttemptCount(ctx context.Context, userID int64) (int, error)
	IncrementCreateAttemptCount(ctx context.Context, userID int64) error
	DeleteCreateAttemptCount(ctx context.Context, userID int64) error

	IncrementDailyUsage(ctx context.Context, apiKey string) error
	SetDailyUsageExpiry(ctx context.Context, apiKey string, ttl time.Duration) error

	GetAuthCache(ctx context.Context, key string) (*domain.APIKeyAuthCacheEntry, error)
	SetAuthCache(ctx context.Context, key string, entry *domain.APIKeyAuthCacheEntry, ttl time.Duration) error
	DeleteAuthCache(ctx context.Context, key string) error

	PublishAuthCacheInvalidation(ctx context.Context, cacheKey string) error
	SubscribeAuthCacheInvalidation(ctx context.Context, handler func(cacheKey string)) error
}
