package service

import (
	"context"
	"log/slog"
	"time"
)

// HandleStreamTimeout 处理流数据超时。
// 根据系统设置决定是否标记账户为临时不可调度或错误状态。
func (s *RateLimitService) HandleStreamTimeout(ctx context.Context, account *Account, model string) bool {
	if account == nil {
		return false
	}
	if s.settingService == nil {
		slog.Warn("stream_timeout_setting_service_missing", "account_id", account.ID)
		return false
	}

	settings, err := s.settingService.GetStreamTimeoutSettings(ctx)
	if err != nil {
		slog.Warn("stream_timeout_get_settings_failed", "account_id", account.ID, "error", err)
		return false
	}
	if !settings.Enabled || settings.Action == StreamTimeoutActionNone {
		return false
	}

	count := s.incrementStreamTimeoutCount(ctx, account.ID, settings)
	slog.Info("stream_timeout_count", "account_id", account.ID, "count", count, "threshold", settings.ThresholdCount, "window_minutes", settings.ThresholdWindowMinutes, "model", model)
	if count < int64(settings.ThresholdCount) {
		return false
	}

	switch settings.Action {
	case StreamTimeoutActionTempUnsched:
		return s.triggerStreamTimeoutTempUnsched(ctx, account, settings, model)
	case StreamTimeoutActionError:
		return s.triggerStreamTimeoutError(ctx, account, model)
	default:
		return false
	}
}

func (s *RateLimitService) incrementStreamTimeoutCount(ctx context.Context, accountID int64, settings *StreamTimeoutSettings) int64 {
	if s.timeoutCounterCache == nil {
		return 1
	}

	count, err := s.timeoutCounterCache.IncrementTimeoutCount(ctx, accountID, settings.ThresholdWindowMinutes)
	if err != nil {
		slog.Warn("stream_timeout_increment_count_failed", "account_id", accountID, "error", err)
		return 1
	}
	return count
}

// triggerStreamTimeoutTempUnsched 触发流超时临时不可调度。
func (s *RateLimitService) triggerStreamTimeoutTempUnsched(ctx context.Context, account *Account, settings *StreamTimeoutSettings, model string) bool {
	now := time.Now()
	until := now.Add(time.Duration(settings.TempUnschedMinutes) * time.Minute)
	state := &TempUnschedState{
		UntilUnix:       until.Unix(),
		TriggeredAtUnix: now.Unix(),
		StatusCode:      0,
		MatchedKeyword:  "stream_timeout",
		RuleIndex:       -1,
		ErrorMessage:    "Stream data interval timeout for model: " + model,
	}
	if !s.persistTempUnschedState(ctx, account.ID, until, state, "stream_timeout_set_temp_unsched_failed") {
		return false
	}

	s.resetStreamTimeoutCount(ctx, account.ID)
	slog.Info("stream_timeout_temp_unschedulable", "account_id", account.ID, "until", until, "model", model)
	return true
}

// triggerStreamTimeoutError 触发流超时错误状态。
func (s *RateLimitService) triggerStreamTimeoutError(ctx context.Context, account *Account, model string) bool {
	errorMessage := "Stream data interval timeout (repeated failures) for model: " + model
	if err := s.accountRepo.SetError(ctx, account.ID, errorMessage); err != nil {
		slog.Warn("stream_timeout_set_error_failed", "account_id", account.ID, "error", err)
		return false
	}

	s.resetStreamTimeoutCount(ctx, account.ID)
	slog.Warn("stream_timeout_account_error", "account_id", account.ID, "model", model)
	return true
}

func (s *RateLimitService) resetStreamTimeoutCount(ctx context.Context, accountID int64) {
	if s.timeoutCounterCache == nil {
		return
	}
	if err := s.timeoutCounterCache.ResetTimeoutCount(ctx, accountID); err != nil {
		slog.Warn("stream_timeout_reset_count_failed", "account_id", accountID, "error", err)
	}
}
