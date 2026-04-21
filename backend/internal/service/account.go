// Package service provides business logic and domain services for the application.
package service

import (
	"errors"
	"time"

	"github.com/senran-N/sub2api/internal/domain"
)

type Account struct {
	ID          int64
	Name        string
	Notes       *string
	Platform    string
	Type        string
	Credentials map[string]any
	Extra       map[string]any
	ProxyID     *int64
	Concurrency int
	Priority    int
	// RateMultiplier 账号计费倍率（>=0，允许 0 表示该账号计费为 0）。
	// 使用指针用于兼容旧版本调度缓存（Redis）中缺字段的情况：nil 表示按 1.0 处理。
	RateMultiplier     *float64
	LoadFactor         *int // 调度负载因子；nil 表示使用 Concurrency
	Status             string
	ErrorMessage       string
	LastUsedAt         *time.Time
	ExpiresAt          *time.Time
	AutoPauseOnExpired bool
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Schedulable bool

	RateLimitedAt    *time.Time
	RateLimitResetAt *time.Time
	OverloadUntil    *time.Time

	TempUnschedulableUntil  *time.Time
	TempUnschedulableReason string

	SessionWindowStart  *time.Time
	SessionWindowEnd    *time.Time
	SessionWindowStatus string

	Proxy         *Proxy
	AccountGroups []AccountGroup
	GroupIDs      []int64
	Groups        []*Group

	modelMappingCache               map[string]string
	modelMappingCacheReady          bool
	modelMappingCacheCredentialsPtr uintptr
	modelMappingCacheRawPtr         uintptr
	modelMappingCacheRawLen         int
	modelMappingCacheRawSig         uint64
}

type TempUnschedulableRule = domain.TempUnschedulableRule

func (a *Account) IsActive() bool {
	return a.Status == StatusActive
}

func (a *Account) BillingRateMultiplier() float64 {
	if a == nil || a.RateMultiplier == nil {
		return 1.0
	}
	if *a.RateMultiplier < 0 {
		return 1.0
	}
	return *a.RateMultiplier
}

func (a *Account) EffectiveLoadFactor() int {
	if a == nil {
		return 1
	}
	if a.LoadFactor != nil && *a.LoadFactor > 0 {
		return *a.LoadFactor
	}
	if a.Concurrency > 0 {
		return a.Concurrency
	}
	return 1
}

func (a *Account) IsSchedulable() bool {
	if !a.IsActive() {
		return false
	}
	if a.IsAPIKeyOrBedrock() && a.IsQuotaExceeded() {
		return false
	}
	return a.isSchedulableAt(time.Now())
}

func (a *Account) IsMaintenanceSchedulable() bool {
	if a == nil {
		return false
	}
	switch a.Status {
	case StatusActive, StatusError:
	default:
		return false
	}
	return a.isSchedulableAt(time.Now())
}

func (a *Account) isSchedulableAt(now time.Time) bool {
	if !a.Schedulable {
		return false
	}
	if a.AutoPauseOnExpired && a.ExpiresAt != nil && !now.Before(*a.ExpiresAt) {
		return false
	}
	if a.OverloadUntil != nil && now.Before(*a.OverloadUntil) {
		return false
	}
	if a.RateLimitResetAt != nil && now.Before(*a.RateLimitResetAt) {
		return false
	}
	if a.TempUnschedulableUntil != nil && now.Before(*a.TempUnschedulableUntil) {
		return false
	}
	return true
}

func (a *Account) IsRateLimited() bool {
	if a.RateLimitResetAt == nil {
		return false
	}
	return time.Now().Before(*a.RateLimitResetAt)
}

func (a *Account) IsOverloaded() bool {
	if a.OverloadUntil == nil {
		return false
	}
	return time.Now().Before(*a.OverloadUntil)
}

func (a *Account) IsOAuth() bool {
	return a.Type == AccountTypeOAuth || a.Type == AccountTypeSetupToken
}

func (a *Account) GetQuotaLimit() float64 {
	return a.getExtraFloat64("quota_limit")
}

func (a *Account) GetQuotaUsed() float64 {
	return a.getExtraFloat64("quota_used")
}

func (a *Account) GetQuotaDailyLimit() float64 {
	return a.getExtraFloat64("quota_daily_limit")
}

func (a *Account) GetQuotaDailyUsed() float64 {
	return a.getExtraFloat64("quota_daily_used")
}

func (a *Account) GetQuotaWeeklyLimit() float64 {
	return a.getExtraFloat64("quota_weekly_limit")
}

func (a *Account) GetQuotaWeeklyUsed() float64 {
	return a.getExtraFloat64("quota_weekly_used")
}

func (a *Account) GetQuotaDailyResetMode() string {
	if mode := a.getExtraString("quota_daily_reset_mode"); mode == "fixed" {
		return "fixed"
	}
	return "rolling"
}

func (a *Account) GetQuotaDailyResetHour() int {
	return a.getExtraInt("quota_daily_reset_hour")
}

func (a *Account) GetQuotaWeeklyResetMode() string {
	if mode := a.getExtraString("quota_weekly_reset_mode"); mode == "fixed" {
		return "fixed"
	}
	return "rolling"
}

func (a *Account) GetQuotaWeeklyResetDay() int {
	if a.Extra == nil {
		return 1
	}
	if _, ok := a.Extra["quota_weekly_reset_day"]; !ok {
		return 1
	}
	return a.getExtraInt("quota_weekly_reset_day")
}

func (a *Account) GetQuotaWeeklyResetHour() int {
	return a.getExtraInt("quota_weekly_reset_hour")
}

func (a *Account) GetQuotaResetTimezone() string {
	if timezone := a.getExtraString("quota_reset_timezone"); timezone != "" {
		return timezone
	}
	return "UTC"
}

func nextFixedDailyReset(hour int, timezone *time.Location, after time.Time) time.Time {
	current := after.In(timezone)
	todayReset := time.Date(current.Year(), current.Month(), current.Day(), hour, 0, 0, 0, timezone)
	if !after.Before(todayReset) {
		return todayReset.AddDate(0, 0, 1)
	}
	return todayReset
}

func lastFixedDailyReset(hour int, timezone *time.Location, now time.Time) time.Time {
	current := now.In(timezone)
	todayReset := time.Date(current.Year(), current.Month(), current.Day(), hour, 0, 0, 0, timezone)
	if now.Before(todayReset) {
		return todayReset.AddDate(0, 0, -1)
	}
	return todayReset
}

func nextFixedWeeklyReset(day, hour int, timezone *time.Location, after time.Time) time.Time {
	current := after.In(timezone)
	todayReset := time.Date(current.Year(), current.Month(), current.Day(), hour, 0, 0, 0, timezone)
	currentDay := int(todayReset.Weekday())

	daysForward := (day - currentDay + 7) % 7
	if daysForward == 0 && !after.Before(todayReset) {
		daysForward = 7
	}
	return todayReset.AddDate(0, 0, daysForward)
}

func lastFixedWeeklyReset(day, hour int, timezone *time.Location, now time.Time) time.Time {
	current := now.In(timezone)
	todayReset := time.Date(current.Year(), current.Month(), current.Day(), hour, 0, 0, 0, timezone)
	currentDay := int(todayReset.Weekday())

	daysBack := (currentDay - day + 7) % 7
	if daysBack == 0 && now.Before(todayReset) {
		daysBack = 7
	}
	return todayReset.AddDate(0, 0, -daysBack)
}

func (a *Account) isFixedDailyPeriodExpired(periodStart time.Time) bool {
	if periodStart.IsZero() {
		return true
	}
	timezone, err := time.LoadLocation(a.GetQuotaResetTimezone())
	if err != nil {
		timezone = time.UTC
	}
	lastReset := lastFixedDailyReset(a.GetQuotaDailyResetHour(), timezone, time.Now())
	return periodStart.Before(lastReset)
}

func (a *Account) isFixedWeeklyPeriodExpired(periodStart time.Time) bool {
	if periodStart.IsZero() {
		return true
	}
	timezone, err := time.LoadLocation(a.GetQuotaResetTimezone())
	if err != nil {
		timezone = time.UTC
	}
	lastReset := lastFixedWeeklyReset(a.GetQuotaWeeklyResetDay(), a.GetQuotaWeeklyResetHour(), timezone, time.Now())
	return periodStart.Before(lastReset)
}

func ComputeQuotaResetAt(extra map[string]any) {
	now := time.Now()
	timezoneName, _ := extra["quota_reset_timezone"].(string)
	if timezoneName == "" {
		timezoneName = "UTC"
	}
	timezone, err := time.LoadLocation(timezoneName)
	if err != nil {
		timezone = time.UTC
	}

	if mode, _ := extra["quota_daily_reset_mode"].(string); mode == "fixed" {
		hour := int(parseExtraFloat64(extra["quota_daily_reset_hour"]))
		if hour < 0 || hour > 23 {
			hour = 0
		}
		resetAt := nextFixedDailyReset(hour, timezone, now)
		extra["quota_daily_reset_at"] = resetAt.UTC().Format(time.RFC3339)
	} else {
		delete(extra, "quota_daily_reset_at")
	}

	if mode, _ := extra["quota_weekly_reset_mode"].(string); mode == "fixed" {
		day := 1
		if rawDay, ok := extra["quota_weekly_reset_day"]; ok {
			day = int(parseExtraFloat64(rawDay))
		}
		if day < 0 || day > 6 {
			day = 1
		}
		hour := int(parseExtraFloat64(extra["quota_weekly_reset_hour"]))
		if hour < 0 || hour > 23 {
			hour = 0
		}
		resetAt := nextFixedWeeklyReset(day, hour, timezone, now)
		extra["quota_weekly_reset_at"] = resetAt.UTC().Format(time.RFC3339)
	} else {
		delete(extra, "quota_weekly_reset_at")
	}
}

func ValidateQuotaResetConfig(extra map[string]any) error {
	if extra == nil {
		return nil
	}
	if timezone, ok := extra["quota_reset_timezone"].(string); ok && timezone != "" {
		if _, err := time.LoadLocation(timezone); err != nil {
			return errors.New("invalid quota_reset_timezone: must be a valid IANA timezone name")
		}
	}
	if mode, ok := extra["quota_daily_reset_mode"].(string); ok {
		if mode != "rolling" && mode != "fixed" {
			return errors.New("quota_daily_reset_mode must be 'rolling' or 'fixed'")
		}
	}
	if rawHour, ok := extra["quota_daily_reset_hour"]; ok {
		hour := int(parseExtraFloat64(rawHour))
		if hour < 0 || hour > 23 {
			return errors.New("quota_daily_reset_hour must be between 0 and 23")
		}
	}
	if mode, ok := extra["quota_weekly_reset_mode"].(string); ok {
		if mode != "rolling" && mode != "fixed" {
			return errors.New("quota_weekly_reset_mode must be 'rolling' or 'fixed'")
		}
	}
	if rawDay, ok := extra["quota_weekly_reset_day"]; ok {
		day := int(parseExtraFloat64(rawDay))
		if day < 0 || day > 6 {
			return errors.New("quota_weekly_reset_day must be between 0 (Sunday) and 6 (Saturday)")
		}
	}
	if rawHour, ok := extra["quota_weekly_reset_hour"]; ok {
		hour := int(parseExtraFloat64(rawHour))
		if hour < 0 || hour > 23 {
			return errors.New("quota_weekly_reset_hour must be between 0 and 23")
		}
	}
	return nil
}

func (a *Account) HasAnyQuotaLimit() bool {
	return a.GetQuotaLimit() > 0 || a.GetQuotaDailyLimit() > 0 || a.GetQuotaWeeklyLimit() > 0
}

func isPeriodExpired(periodStart time.Time, duration time.Duration) bool {
	if periodStart.IsZero() {
		return true
	}
	return time.Since(periodStart) >= duration
}

func (a *Account) IsDailyQuotaPeriodExpired() bool {
	start := a.getExtraTime("quota_daily_start")
	if a.GetQuotaDailyResetMode() == "fixed" {
		return a.isFixedDailyPeriodExpired(start)
	}
	return isPeriodExpired(start, 24*time.Hour)
}

func (a *Account) IsWeeklyQuotaPeriodExpired() bool {
	start := a.getExtraTime("quota_weekly_start")
	if a.GetQuotaWeeklyResetMode() == "fixed" {
		return a.isFixedWeeklyPeriodExpired(start)
	}
	return isPeriodExpired(start, 7*24*time.Hour)
}

func (a *Account) IsQuotaExceeded() bool {
	if limit := a.GetQuotaLimit(); limit > 0 && a.GetQuotaUsed() >= limit {
		return true
	}
	if limit := a.GetQuotaDailyLimit(); limit > 0 {
		start := a.getExtraTime("quota_daily_start")
		expired := isPeriodExpired(start, 24*time.Hour)
		if a.GetQuotaDailyResetMode() == "fixed" {
			expired = a.isFixedDailyPeriodExpired(start)
		}
		if !expired && a.GetQuotaDailyUsed() >= limit {
			return true
		}
	}
	if limit := a.GetQuotaWeeklyLimit(); limit > 0 {
		start := a.getExtraTime("quota_weekly_start")
		expired := isPeriodExpired(start, 7*24*time.Hour)
		if a.GetQuotaWeeklyResetMode() == "fixed" {
			expired = a.isFixedWeeklyPeriodExpired(start)
		}
		if !expired && a.GetQuotaWeeklyUsed() >= limit {
			return true
		}
	}
	return false
}

func (a *Account) GetWindowCostLimit() float64 {
	if a.Extra == nil {
		return 0
	}
	if value, ok := a.Extra["window_cost_limit"]; ok {
		return parseExtraFloat64(value)
	}
	return 0
}

func (a *Account) GetWindowCostStickyReserve() float64 {
	if a.Extra == nil {
		return 10.0
	}
	if value, ok := a.Extra["window_cost_sticky_reserve"]; ok {
		parsed := parseExtraFloat64(value)
		if parsed > 0 {
			return parsed
		}
	}
	return 10.0
}

func (a *Account) GetMaxSessions() int {
	if a.Extra == nil {
		return 0
	}
	if value, ok := a.Extra["max_sessions"]; ok {
		return parseExtraInt(value)
	}
	return 0
}

func (a *Account) GetSessionIdleTimeoutMinutes() int {
	if a.Extra == nil {
		return 5
	}
	if value, ok := a.Extra["session_idle_timeout_minutes"]; ok {
		parsed := parseExtraInt(value)
		if parsed > 0 {
			return parsed
		}
	}
	return 5
}

func (a *Account) GetBaseRPM() int {
	if a.Extra == nil {
		return 0
	}
	if value, ok := a.Extra["base_rpm"]; ok {
		parsed := parseExtraInt(value)
		if parsed > 0 {
			return parsed
		}
	}
	return 0
}

func (a *Account) GetRPMStrategy() string {
	if a.Extra == nil {
		return "tiered"
	}
	if value, ok := a.Extra["rpm_strategy"]; ok {
		if strategy, ok := value.(string); ok && strategy == "sticky_exempt" {
			return "sticky_exempt"
		}
	}
	return "tiered"
}

func (a *Account) GetRPMStickyBuffer() int {
	if a.Extra == nil {
		return 0
	}
	if value, ok := a.Extra["rpm_sticky_buffer"]; ok {
		parsed := parseExtraInt(value)
		if parsed > 0 {
			return parsed
		}
	}

	baseRPM := a.GetBaseRPM()
	if baseRPM <= 0 {
		return 0
	}

	concurrency := a.Concurrency
	if concurrency < 0 {
		concurrency = 0
	}
	maxSessions := a.GetMaxSessions()
	if maxSessions < 0 {
		maxSessions = 0
	}

	buffer := concurrency + maxSessions
	floor := baseRPM / 5
	if floor < 1 {
		floor = 1
	}
	if buffer < floor {
		buffer = floor
	}
	return buffer
}

func (a *Account) CheckRPMSchedulability(currentRPM int) WindowCostSchedulability {
	baseRPM := a.GetBaseRPM()
	if baseRPM <= 0 {
		return WindowCostSchedulable
	}
	if currentRPM < baseRPM {
		return WindowCostSchedulable
	}
	if a.GetRPMStrategy() == "sticky_exempt" {
		return WindowCostStickyOnly
	}
	if currentRPM < baseRPM+a.GetRPMStickyBuffer() {
		return WindowCostStickyOnly
	}
	return WindowCostNotSchedulable
}

func (a *Account) CheckWindowCostSchedulability(currentWindowCost float64) WindowCostSchedulability {
	limit := a.GetWindowCostLimit()
	if limit <= 0 {
		return WindowCostSchedulable
	}
	if currentWindowCost < limit {
		return WindowCostSchedulable
	}
	if currentWindowCost < limit+a.GetWindowCostStickyReserve() {
		return WindowCostStickyOnly
	}
	return WindowCostNotSchedulable
}

func (a *Account) GetCurrentWindowStartTime() time.Time {
	now := time.Now()
	if a.SessionWindowStart != nil && a.SessionWindowEnd != nil && now.Before(*a.SessionWindowEnd) {
		return *a.SessionWindowStart
	}
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
}
