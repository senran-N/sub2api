package service

import (
	"context"
	"log"
	"time"
)

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (s *SubscriptionService) CheckAndActivateWindow(ctx context.Context, sub *UserSubscription) error {
	if sub.IsWindowActivated() {
		return nil
	}
	return s.userSubRepo.ActivateWindows(ctx, sub.ID, startOfDay(time.Now()))
}

func (s *SubscriptionService) AdminResetQuota(ctx context.Context, subscriptionID int64, resetDaily, resetWeekly, resetMonthly bool) (*UserSubscription, error) {
	if !resetDaily && !resetWeekly && !resetMonthly {
		return nil, ErrInvalidInput
	}

	sub, err := s.userSubRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	windowStart := startOfDay(time.Now())
	if resetDaily {
		if err := s.userSubRepo.ResetDailyUsage(ctx, sub.ID, windowStart); err != nil {
			return nil, err
		}
	}
	if resetWeekly {
		if err := s.userSubRepo.ResetWeeklyUsage(ctx, sub.ID, windowStart); err != nil {
			return nil, err
		}
	}
	if resetMonthly {
		if err := s.userSubRepo.ResetMonthlyUsage(ctx, sub.ID, windowStart); err != nil {
			return nil, err
		}
	}

	s.InvalidateSubCache(sub.UserID, sub.GroupID)
	if s.subCacheL1 != nil {
		s.subCacheL1.Wait()
	}
	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateSubscription(ctx, sub.UserID, sub.GroupID)
	}
	return s.userSubRepo.GetByID(ctx, subscriptionID)
}

func (s *SubscriptionService) CheckAndResetWindows(ctx context.Context, sub *UserSubscription) error {
	windowStart := startOfDay(time.Now())
	needsInvalidateCache := false

	if sub.NeedsDailyReset() {
		if err := s.userSubRepo.ResetDailyUsage(ctx, sub.ID, windowStart); err != nil {
			return err
		}
		sub.DailyWindowStart = &windowStart
		sub.DailyUsageUSD = 0
		needsInvalidateCache = true
	}
	if sub.NeedsWeeklyReset() {
		if err := s.userSubRepo.ResetWeeklyUsage(ctx, sub.ID, windowStart); err != nil {
			return err
		}
		sub.WeeklyWindowStart = &windowStart
		sub.WeeklyUsageUSD = 0
		needsInvalidateCache = true
	}
	if sub.NeedsMonthlyReset() {
		if err := s.userSubRepo.ResetMonthlyUsage(ctx, sub.ID, windowStart); err != nil {
			return err
		}
		sub.MonthlyWindowStart = &windowStart
		sub.MonthlyUsageUSD = 0
		needsInvalidateCache = true
	}

	if needsInvalidateCache {
		s.InvalidateSubCache(sub.UserID, sub.GroupID)
		if s.billingCacheService != nil {
			_ = s.billingCacheService.InvalidateSubscription(ctx, sub.UserID, sub.GroupID)
		}
	}

	return nil
}

func (s *SubscriptionService) CheckUsageLimits(ctx context.Context, sub *UserSubscription, group *Group, additionalCost float64) error {
	if !sub.CheckDailyLimit(group, additionalCost) {
		return ErrDailyLimitExceeded
	}
	if !sub.CheckWeeklyLimit(group, additionalCost) {
		return ErrWeeklyLimitExceeded
	}
	if !sub.CheckMonthlyLimit(group, additionalCost) {
		return ErrMonthlyLimitExceeded
	}
	return nil
}

func (s *SubscriptionService) ValidateAndCheckLimits(sub *UserSubscription, group *Group) (bool, error) {
	if sub.Status == SubscriptionStatusExpired {
		return false, ErrSubscriptionExpired
	}
	if sub.Status == SubscriptionStatusSuspended {
		return false, ErrSubscriptionSuspended
	}
	if sub.IsExpired() {
		return false, ErrSubscriptionExpired
	}

	needsMaintenance := false
	if sub.NeedsDailyReset() {
		sub.DailyUsageUSD = 0
		needsMaintenance = true
	}
	if sub.NeedsWeeklyReset() {
		sub.WeeklyUsageUSD = 0
		needsMaintenance = true
	}
	if sub.NeedsMonthlyReset() {
		sub.MonthlyUsageUSD = 0
		needsMaintenance = true
	}
	if !sub.IsWindowActivated() {
		needsMaintenance = true
	}

	if !sub.CheckDailyLimit(group, 0) {
		return needsMaintenance, ErrDailyLimitExceeded
	}
	if !sub.CheckWeeklyLimit(group, 0) {
		return needsMaintenance, ErrWeeklyLimitExceeded
	}
	if !sub.CheckMonthlyLimit(group, 0) {
		return needsMaintenance, ErrMonthlyLimitExceeded
	}

	return needsMaintenance, nil
}

func (s *SubscriptionService) DoWindowMaintenance(sub *UserSubscription) {
	if s == nil {
		return
	}
	if s.maintenanceQueue != nil {
		err := s.maintenanceQueue.TryEnqueue(func() {
			s.doWindowMaintenance(sub)
		})
		if err != nil {
			log.Printf("Subscription maintenance enqueue failed: %v", err)
		}
		return
	}
	s.doWindowMaintenance(sub)
}

func (s *SubscriptionService) doWindowMaintenance(sub *UserSubscription) {
	ctx, cancel := context.WithTimeout(context.Background(), subscriptionCacheInvalidateTimeout)
	defer cancel()

	if !sub.IsWindowActivated() {
		if err := s.CheckAndActivateWindow(ctx, sub); err != nil {
			log.Printf("Failed to activate subscription windows: %v", err)
		}
	}
	if err := s.CheckAndResetWindows(ctx, sub); err != nil {
		log.Printf("Failed to reset subscription windows: %v", err)
	}

	s.InvalidateSubCache(sub.UserID, sub.GroupID)
}

func (s *SubscriptionService) RecordUsage(ctx context.Context, subscriptionID int64, costUSD float64) error {
	return s.userSubRepo.IncrementUsage(ctx, subscriptionID, costUSD)
}

func (s *SubscriptionService) ValidateSubscription(ctx context.Context, sub *UserSubscription) error {
	if sub.Status == SubscriptionStatusExpired {
		return ErrSubscriptionExpired
	}
	if sub.Status == SubscriptionStatusSuspended {
		return ErrSubscriptionSuspended
	}
	if sub.IsExpired() {
		_ = s.userSubRepo.UpdateStatus(ctx, sub.ID, SubscriptionStatusExpired)
		return ErrSubscriptionExpired
	}
	return nil
}
