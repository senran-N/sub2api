package service

import (
	"time"

	"github.com/dgraph-io/ristretto"
	dbent "github.com/senran-N/sub2api/ent"
	"github.com/senran-N/sub2api/internal/config"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
	"golang.org/x/sync/singleflight"
)

var MaxExpiresAt = time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

const MaxValidityDays = 36500

var (
	ErrSubscriptionNotFound       = infraerrors.NotFound("SUBSCRIPTION_NOT_FOUND", "subscription not found")
	ErrSubscriptionExpired        = infraerrors.Forbidden("SUBSCRIPTION_EXPIRED", "subscription has expired")
	ErrSubscriptionSuspended      = infraerrors.Forbidden("SUBSCRIPTION_SUSPENDED", "subscription is suspended")
	ErrSubscriptionAlreadyExists  = infraerrors.Conflict("SUBSCRIPTION_ALREADY_EXISTS", "subscription already exists for this user and group")
	ErrSubscriptionAssignConflict = infraerrors.Conflict("SUBSCRIPTION_ASSIGN_CONFLICT", "subscription exists but request conflicts with existing assignment semantics")
	ErrGroupNotSubscriptionType   = infraerrors.BadRequest("GROUP_NOT_SUBSCRIPTION_TYPE", "group is not a subscription type")
	ErrInvalidInput               = infraerrors.BadRequest("INVALID_INPUT", "at least one of resetDaily, resetWeekly, or resetMonthly must be true")
	ErrDailyLimitExceeded         = infraerrors.TooManyRequests("DAILY_LIMIT_EXCEEDED", "daily usage limit exceeded")
	ErrWeeklyLimitExceeded        = infraerrors.TooManyRequests("WEEKLY_LIMIT_EXCEEDED", "weekly usage limit exceeded")
	ErrMonthlyLimitExceeded       = infraerrors.TooManyRequests("MONTHLY_LIMIT_EXCEEDED", "monthly usage limit exceeded")
	ErrSubscriptionNilInput       = infraerrors.BadRequest("SUBSCRIPTION_NIL_INPUT", "subscription input cannot be nil")
	ErrAdjustWouldExpire          = infraerrors.BadRequest("ADJUST_WOULD_EXPIRE", "adjustment would result in expired subscription (remaining days must be > 0)")
)

type SubscriptionService struct {
	groupRepo           GroupRepository
	userSubRepo         UserSubscriptionRepository
	billingCacheService *BillingCacheService
	entClient           *dbent.Client

	subCacheL1     *ristretto.Cache
	subCacheGroup  singleflight.Group
	subCacheTTL    time.Duration
	subCacheJitter int

	maintenanceQueue *SubscriptionMaintenanceQueue
}

func NewSubscriptionService(groupRepo GroupRepository, userSubRepo UserSubscriptionRepository, billingCacheService *BillingCacheService, entClient *dbent.Client, cfg *config.Config) *SubscriptionService {
	svc := &SubscriptionService{
		groupRepo:           groupRepo,
		userSubRepo:         userSubRepo,
		billingCacheService: billingCacheService,
		entClient:           entClient,
	}
	svc.initSubCache(cfg)
	svc.initMaintenanceQueue(cfg)
	return svc
}

func (s *SubscriptionService) initMaintenanceQueue(cfg *config.Config) {
	if cfg == nil {
		return
	}
	mc := cfg.SubscriptionMaintenance
	if mc.WorkerCount <= 0 || mc.QueueSize <= 0 {
		return
	}
	s.maintenanceQueue = NewSubscriptionMaintenanceQueue(mc.WorkerCount, mc.QueueSize)
}

func (s *SubscriptionService) Stop() {
	if s == nil {
		return
	}
	if s.maintenanceQueue != nil {
		s.maintenanceQueue.Stop()
	}
}
