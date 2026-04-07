package service

import (
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

// backgroundStartable captures services that need an explicit Start phase.
// We keep startup orchestration centralized here so providers can compose
// "construct then start" without duplicating lifecycle glue.
type backgroundStartable interface {
	Start()
}

func startBackgroundService[T backgroundStartable](svc T) T {
	svc.Start()
	return svc
}

func newTimingWheelServiceForLifecycle() (*TimingWheelService, error) {
	return NewTimingWheelService()
}

func newDashboardAggregationServiceForLifecycle(
	repo DashboardAggregationRepository,
	timingWheel *TimingWheelService,
	cfg *config.Config,
) *DashboardAggregationService {
	return NewDashboardAggregationService(repo, timingWheel, cfg)
}

func newUsageCleanupServiceForLifecycle(
	repo UsageCleanupRepository,
	timingWheel *TimingWheelService,
	dashboardAgg *DashboardAggregationService,
	cfg *config.Config,
) *UsageCleanupService {
	return NewUsageCleanupService(repo, timingWheel, dashboardAgg, cfg)
}

func newAccountExpiryServiceForLifecycle(accountRepo AccountRepository) *AccountExpiryService {
	return NewAccountExpiryService(accountRepo, time.Minute)
}

func newSubscriptionExpiryServiceForLifecycle(
	userSubRepo UserSubscriptionRepository,
) *SubscriptionExpiryService {
	return NewSubscriptionExpiryService(userSubRepo, time.Minute)
}

func newDeferredServiceForLifecycle(
	accountRepo AccountRepository,
	timingWheel *TimingWheelService,
) *DeferredService {
	return NewDeferredService(accountRepo, timingWheel, 10*time.Second)
}
