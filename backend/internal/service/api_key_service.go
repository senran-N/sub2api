package service

import (
	"context"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/domain"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"golang.org/x/sync/singleflight"
)

var (
	ErrAPIKeyNotFound       = infraerrors.NotFound("API_KEY_NOT_FOUND", "api key not found")
	ErrGroupNotAllowed      = infraerrors.Forbidden("GROUP_NOT_ALLOWED", "user is not allowed to bind this group")
	ErrAPIKeyExists         = infraerrors.Conflict("API_KEY_EXISTS", "api key already exists")
	ErrAPIKeyTooShort       = infraerrors.BadRequest("API_KEY_TOO_SHORT", "api key must be at least 16 characters")
	ErrAPIKeyInvalidChars   = infraerrors.BadRequest("API_KEY_INVALID_CHARS", "api key can only contain letters, numbers, underscores, and hyphens")
	ErrAPIKeyRateLimited    = infraerrors.TooManyRequests("API_KEY_RATE_LIMITED", "too many failed attempts, please try again later")
	ErrInvalidIPPattern     = infraerrors.BadRequest("INVALID_IP_PATTERN", "invalid IP or CIDR pattern")
	ErrAPIKeyExpired        = infraerrors.Forbidden("API_KEY_EXPIRED", "api key 已过期")
	ErrAPIKeyQuotaExhausted = infraerrors.TooManyRequests(
		"API_KEY_QUOTA_EXHAUSTED",
		"api key 额度已用完",
	)

	ErrAPIKeyRateLimit5hExceeded = infraerrors.TooManyRequests("API_KEY_RATE_5H_EXCEEDED", "api key 5小时限额已用完")
	ErrAPIKeyRateLimit1dExceeded = infraerrors.TooManyRequests("API_KEY_RATE_1D_EXCEEDED", "api key 日限额已用完")
	ErrAPIKeyRateLimit7dExceeded = infraerrors.TooManyRequests("API_KEY_RATE_7D_EXCEEDED", "api key 7天限额已用完")
)

const (
	apiKeyMaxErrorsPerHour = 20
	apiKeyLastUsedMinTouch = 30 * time.Second
	// DB 写失败后的短退避，避免请求路径持续同步重试造成写风暴与高延迟。
	apiKeyLastUsedFailBackoff = 5 * time.Second
)

type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByID(ctx context.Context, id int64) (*APIKey, error)
	GetKeyAndOwnerID(ctx context.Context, id int64) (string, int64, error)
	GetByKey(ctx context.Context, key string) (*APIKey, error)
	GetByKeyForAuth(ctx context.Context, key string) (*APIKey, error)
	Update(ctx context.Context, key *APIKey) error
	Delete(ctx context.Context, id int64) error

	ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams, filters APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error)
	VerifyOwnership(ctx context.Context, userID int64, apiKeyIDs []int64) ([]int64, error)
	CountByUserID(ctx context.Context, userID int64) (int64, error)
	ExistsByKey(ctx context.Context, key string) (bool, error)
	ListByGroupID(ctx context.Context, groupID int64, params pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error)
	SearchAPIKeys(ctx context.Context, userID int64, keyword string, limit int) ([]APIKey, error)
	ClearGroupIDByGroupID(ctx context.Context, groupID int64) (int64, error)
	UpdateGroupIDByUserAndGroup(ctx context.Context, userID, oldGroupID, newGroupID int64) (int64, error)
	CountByGroupID(ctx context.Context, groupID int64) (int64, error)
	ListKeysByUserID(ctx context.Context, userID int64) ([]string, error)
	ListKeysByGroupID(ctx context.Context, groupID int64) ([]string, error)

	IncrementQuotaUsed(ctx context.Context, id int64, amount float64) (float64, error)
	UpdateLastUsed(ctx context.Context, id int64, usedAt time.Time) error

	IncrementRateLimitUsage(ctx context.Context, id int64, cost float64) error
	ResetRateLimitWindows(ctx context.Context, id int64) error
	GetRateLimitData(ctx context.Context, id int64) (*APIKeyRateLimitData, error)
}

type APIKeyRateLimitData = domain.APIKeyRateLimitData
type APIKeyQuotaUsageState = domain.APIKeyQuotaUsageState

type APIKeyAuthCacheInvalidator interface {
	InvalidateAuthCacheByKey(ctx context.Context, key string)
	InvalidateAuthCacheByUserID(ctx context.Context, userID int64)
	InvalidateAuthCacheByGroupID(ctx context.Context, groupID int64)
}

type CreateAPIKeyRequest = domain.CreateAPIKeyRequest
type UpdateAPIKeyRequest = domain.UpdateAPIKeyRequest

type RateLimitCacheInvalidator interface {
	InvalidateAPIKeyRateLimit(ctx context.Context, keyID int64) error
}

type APIKeyService struct {
	apiKeyRepo            APIKeyRepository
	userRepo              UserRepository
	groupRepo             GroupRepository
	userSubRepo           UserSubscriptionRepository
	userGroupRateRepo     UserGroupRateRepository
	cache                 APIKeyCache
	rateLimitCacheInvalid RateLimitCacheInvalidator
	cfg                   *config.Config
	authCacheL1           *ristretto.Cache
	authCfg               apiKeyAuthCacheConfig
	authGroup             singleflight.Group
	lastUsedTouchL1       sync.Map
	lastUsedTouchSF       singleflight.Group
}

func NewAPIKeyService(
	apiKeyRepo APIKeyRepository,
	userRepo UserRepository,
	groupRepo GroupRepository,
	userSubRepo UserSubscriptionRepository,
	userGroupRateRepo UserGroupRateRepository,
	cache APIKeyCache,
	cfg *config.Config,
) *APIKeyService {
	svc := &APIKeyService{
		apiKeyRepo:        apiKeyRepo,
		userRepo:          userRepo,
		groupRepo:         groupRepo,
		userSubRepo:       userSubRepo,
		userGroupRateRepo: userGroupRateRepo,
		cache:             cache,
		cfg:               cfg,
	}
	svc.initAuthCache(cfg)
	return svc
}

func (s *APIKeyService) SetRateLimitCacheInvalidator(inv RateLimitCacheInvalidator) {
	s.rateLimitCacheInvalid = inv
}
