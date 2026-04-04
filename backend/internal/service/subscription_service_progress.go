package service

import (
	"context"
	"time"
)

type SubscriptionProgress struct {
	ID            int64                `json:"id"`
	GroupName     string               `json:"group_name"`
	ExpiresAt     time.Time            `json:"expires_at"`
	ExpiresInDays int                  `json:"expires_in_days"`
	Daily         *UsageWindowProgress `json:"daily,omitempty"`
	Weekly        *UsageWindowProgress `json:"weekly,omitempty"`
	Monthly       *UsageWindowProgress `json:"monthly,omitempty"`
}

type UsageWindowProgress struct {
	LimitUSD        float64   `json:"limit_usd"`
	UsedUSD         float64   `json:"used_usd"`
	RemainingUSD    float64   `json:"remaining_usd"`
	Percentage      float64   `json:"percentage"`
	WindowStart     time.Time `json:"window_start"`
	ResetsAt        time.Time `json:"resets_at"`
	ResetsInSeconds int64     `json:"resets_in_seconds"`
}

func (s *SubscriptionService) GetSubscriptionProgress(ctx context.Context, subscriptionID int64) (*SubscriptionProgress, error) {
	sub, err := s.userSubRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	group := sub.Group
	if group == nil {
		group, err = s.groupRepo.GetByID(ctx, sub.GroupID)
		if err != nil {
			return nil, err
		}
	}

	return s.calculateProgress(sub, group), nil
}

func (s *SubscriptionService) calculateProgress(sub *UserSubscription, group *Group) *SubscriptionProgress {
	progress := &SubscriptionProgress{
		ID:            sub.ID,
		GroupName:     group.Name,
		ExpiresAt:     sub.ExpiresAt,
		ExpiresInDays: sub.DaysRemaining(),
	}

	if group.HasDailyLimit() && sub.DailyWindowStart != nil {
		progress.Daily = buildUsageWindowProgress(*group.DailyLimitUSD, sub.DailyUsageUSD, *sub.DailyWindowStart, 24*time.Hour)
	}
	if group.HasWeeklyLimit() && sub.WeeklyWindowStart != nil {
		progress.Weekly = buildUsageWindowProgress(*group.WeeklyLimitUSD, sub.WeeklyUsageUSD, *sub.WeeklyWindowStart, 7*24*time.Hour)
	}
	if group.HasMonthlyLimit() && sub.MonthlyWindowStart != nil {
		progress.Monthly = buildUsageWindowProgress(*group.MonthlyLimitUSD, sub.MonthlyUsageUSD, *sub.MonthlyWindowStart, 30*24*time.Hour)
	}

	return progress
}

func buildUsageWindowProgress(limit, used float64, windowStart time.Time, duration time.Duration) *UsageWindowProgress {
	resetsAt := windowStart.Add(duration)
	progress := &UsageWindowProgress{
		LimitUSD:        limit,
		UsedUSD:         used,
		RemainingUSD:    limit - used,
		Percentage:      (used / limit) * 100,
		WindowStart:     windowStart,
		ResetsAt:        resetsAt,
		ResetsInSeconds: int64(time.Until(resetsAt).Seconds()),
	}
	if progress.RemainingUSD < 0 {
		progress.RemainingUSD = 0
	}
	if progress.Percentage > 100 {
		progress.Percentage = 100
	}
	if progress.ResetsInSeconds < 0 {
		progress.ResetsInSeconds = 0
	}
	return progress
}

func (s *SubscriptionService) GetUserSubscriptionsWithProgress(ctx context.Context, userID int64) ([]SubscriptionProgress, error) {
	subs, err := s.userSubRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	progresses := make([]SubscriptionProgress, 0, len(subs))
	for i := range subs {
		sub := &subs[i]
		if sub.Group == nil {
			continue
		}
		progresses = append(progresses, *s.calculateProgress(sub, sub.Group))
	}

	return progresses, nil
}
