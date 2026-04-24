package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"golang.org/x/sync/singleflight"

	"github.com/senran-N/sub2api/internal/config"
)

var (
	ErrSubscriptionInvalid       = infraerrors.Forbidden("SUBSCRIPTION_INVALID", "subscription is invalid or expired")
	ErrBillingServiceUnavailable = infraerrors.ServiceUnavailable("BILLING_SERVICE_ERROR", "Billing service temporarily unavailable. Please retry later.")
)

// subscriptionCacheData 订阅缓存数据结构（内部使用）
type subscriptionCacheData struct {
	Status       string
	ExpiresAt    time.Time
	DailyUsage   float64
	WeeklyUsage  float64
	MonthlyUsage float64
	Version      int64
}

type cacheWriteKind int

const (
	cacheWriteSetBalance cacheWriteKind = iota
	cacheWriteSetSubscription
	cacheWriteUpdateSubscriptionUsage
	cacheWriteDeductBalance
	cacheWriteUpdateRateLimitUsage
)

const (
	cacheWriteWorkerCount     = 10
	cacheWriteBufferSize      = 1000
	cacheWriteTimeout         = 2 * time.Second
	cacheWriteDropLogInterval = 5 * time.Second
	balanceLoadTimeout        = 3 * time.Second
)

// cacheWriteTask 缓存写入任务
type cacheWriteTask struct {
	kind             cacheWriteKind
	userID           int64
	groupID          int64
	apiKeyID         int64
	balance          float64
	amount           float64
	subscriptionData *subscriptionCacheData
}

// apiKeyRateLimitLoader defines the interface for loading rate limit data from DB.
type apiKeyRateLimitLoader interface {
	GetRateLimitData(ctx context.Context, keyID int64) (*APIKeyRateLimitData, error)
}

// BillingCacheService 计费缓存服务
type BillingCacheService struct {
	cache                 BillingCache
	userRepo              UserRepository
	subRepo               UserSubscriptionRepository
	apiKeyRateLimitLoader apiKeyRateLimitLoader
	cfg                   *config.Config
	circuitBreaker        *billingCircuitBreaker

	cacheWriteChan     chan cacheWriteTask
	cacheWriteWg       sync.WaitGroup
	cacheWriteStopOnce sync.Once
	cacheWriteMu       sync.RWMutex
	stopped            atomic.Bool
	balanceLoadSF      singleflight.Group

	cacheWriteDropFullCount     uint64
	cacheWriteDropFullLastLog   int64
	cacheWriteDropClosedCount   uint64
	cacheWriteDropClosedLastLog int64
}

// NewBillingCacheService 创建计费缓存服务
func NewBillingCacheService(cache BillingCache, userRepo UserRepository, subRepo UserSubscriptionRepository, apiKeyRepo APIKeyRepository, extra ...any) *BillingCacheService {
	cfg := (*config.Config)(nil)
	if len(extra) == 1 {
		cfg, _ = extra[0].(*config.Config)
	} else if len(extra) >= 3 {
		cfg, _ = extra[2].(*config.Config)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	svc := &BillingCacheService{
		cache:                 cache,
		userRepo:              userRepo,
		subRepo:               subRepo,
		apiKeyRateLimitLoader: apiKeyRepo,
		cfg:                   cfg,
	}
	svc.circuitBreaker = newBillingCircuitBreaker(cfg.Billing.CircuitBreaker)
	svc.startCacheWriteWorkers()
	return svc
}
