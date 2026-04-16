package service

import (
	"context"
	"fmt"
	"strings"
)

// GetOverloadCooldownSettings 获取529过载冷却配置
func (s *SettingService) GetOverloadCooldownSettings(ctx context.Context) (*OverloadCooldownSettings, error) {
	return readJSONSetting(
		ctx,
		s.settingRepo,
		SettingKeyOverloadCooldownSettings,
		"overload cooldown settings",
		DefaultOverloadCooldownSettings,
		normalizeOverloadCooldownSettings,
	)
}

func normalizeOverloadCooldownSettings(settings *OverloadCooldownSettings) {
	if settings.CooldownMinutes < 1 {
		settings.CooldownMinutes = 1
	}
	if settings.CooldownMinutes > 120 {
		settings.CooldownMinutes = 120
	}
}

// SetOverloadCooldownSettings 设置529过载冷却配置
func (s *SettingService) SetOverloadCooldownSettings(ctx context.Context, settings *OverloadCooldownSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	normalized := *settings
	if normalized.CooldownMinutes < 1 || normalized.CooldownMinutes > 120 {
		if normalized.Enabled {
			return fmt.Errorf("cooldown_minutes must be between 1-120")
		}
		normalized.CooldownMinutes = 10
	}

	return writeJSONSetting(ctx, s.settingRepo, SettingKeyOverloadCooldownSettings, "overload cooldown settings", &normalized)
}

// GetStreamTimeoutSettings 获取流超时处理配置
func (s *SettingService) GetStreamTimeoutSettings(ctx context.Context) (*StreamTimeoutSettings, error) {
	return readJSONSetting(
		ctx,
		s.settingRepo,
		SettingKeyStreamTimeoutSettings,
		"stream timeout settings",
		DefaultStreamTimeoutSettings,
		normalizeStreamTimeoutSettings,
	)
}

func normalizeStreamTimeoutSettings(settings *StreamTimeoutSettings) {
	if settings.TempUnschedMinutes < 1 {
		settings.TempUnschedMinutes = 1
	}
	if settings.TempUnschedMinutes > 60 {
		settings.TempUnschedMinutes = 60
	}
	if settings.ThresholdCount < 1 {
		settings.ThresholdCount = 1
	}
	if settings.ThresholdCount > 10 {
		settings.ThresholdCount = 10
	}
	if settings.ThresholdWindowMinutes < 1 {
		settings.ThresholdWindowMinutes = 1
	}
	if settings.ThresholdWindowMinutes > 60 {
		settings.ThresholdWindowMinutes = 60
	}

	switch settings.Action {
	case StreamTimeoutActionTempUnsched, StreamTimeoutActionError, StreamTimeoutActionNone:
	default:
		settings.Action = StreamTimeoutActionTempUnsched
	}
}

// SetStreamTimeoutSettings 设置流超时处理配置
func (s *SettingService) SetStreamTimeoutSettings(ctx context.Context, settings *StreamTimeoutSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	if settings.TempUnschedMinutes < 1 || settings.TempUnschedMinutes > 60 {
		return fmt.Errorf("temp_unsched_minutes must be between 1-60")
	}
	if settings.ThresholdCount < 1 || settings.ThresholdCount > 10 {
		return fmt.Errorf("threshold_count must be between 1-10")
	}
	if settings.ThresholdWindowMinutes < 1 || settings.ThresholdWindowMinutes > 60 {
		return fmt.Errorf("threshold_window_minutes must be between 1-60")
	}

	switch settings.Action {
	case StreamTimeoutActionTempUnsched, StreamTimeoutActionError, StreamTimeoutActionNone:
	default:
		return fmt.Errorf("invalid action: %s", settings.Action)
	}

	return writeJSONSetting(ctx, s.settingRepo, SettingKeyStreamTimeoutSettings, "stream timeout settings", settings)
}

// GetRectifierSettings 获取请求整流器配置
func (s *SettingService) GetRectifierSettings(ctx context.Context) (*RectifierSettings, error) {
	return readJSONSetting(
		ctx,
		s.settingRepo,
		SettingKeyRectifierSettings,
		"rectifier settings",
		DefaultRectifierSettings,
		nil,
	)
}

// SetRectifierSettings 设置请求整流器配置
func (s *SettingService) SetRectifierSettings(ctx context.Context, settings *RectifierSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	return writeJSONSetting(ctx, s.settingRepo, SettingKeyRectifierSettings, "rectifier settings", settings)
}

// IsSignatureRectifierEnabled 判断签名整流是否启用（总开关 && 签名子开关）
func (s *SettingService) IsSignatureRectifierEnabled(ctx context.Context) bool {
	settings, err := s.GetRectifierSettings(ctx)
	if err != nil {
		return true
	}
	return settings.Enabled && settings.ThinkingSignatureEnabled
}

// IsBudgetRectifierEnabled 判断 Budget 整流是否启用（总开关 && Budget 子开关）
func (s *SettingService) IsBudgetRectifierEnabled(ctx context.Context) bool {
	settings, err := s.GetRectifierSettings(ctx)
	if err != nil {
		return true
	}
	return settings.Enabled && settings.ThinkingBudgetEnabled
}

// GetBetaPolicySettings 获取 Beta 策略配置
func (s *SettingService) GetBetaPolicySettings(ctx context.Context) (*BetaPolicySettings, error) {
	return readJSONSetting(
		ctx,
		s.settingRepo,
		SettingKeyBetaPolicySettings,
		"beta policy settings",
		DefaultBetaPolicySettings,
		nil,
	)
}

// SetBetaPolicySettings 设置 Beta 策略配置
func (s *SettingService) SetBetaPolicySettings(ctx context.Context, settings *BetaPolicySettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	validActions := map[string]bool{
		BetaPolicyActionPass: true, BetaPolicyActionFilter: true, BetaPolicyActionBlock: true,
	}
	validScopes := map[string]bool{
		BetaPolicyScopeAll: true, BetaPolicyScopeOAuth: true, BetaPolicyScopeAPIKey: true, BetaPolicyScopeBedrock: true,
	}

	for i, rule := range settings.Rules {
		settings.Rules[i].BetaToken = strings.TrimSpace(rule.BetaToken)
		settings.Rules[i].ErrorMessage = strings.TrimSpace(rule.ErrorMessage)
		settings.Rules[i].FallbackErrorMessage = strings.TrimSpace(rule.FallbackErrorMessage)
		if settings.Rules[i].BetaToken == "" {
			return fmt.Errorf("rule[%d]: beta_token cannot be empty", i)
		}
		if !validActions[rule.Action] {
			return fmt.Errorf("rule[%d]: invalid action %q", i, rule.Action)
		}
		if !validScopes[rule.Scope] {
			return fmt.Errorf("rule[%d]: invalid scope %q", i, rule.Scope)
		}
		normalizedWhitelist := make([]string, 0, len(rule.ModelWhitelist))
		for _, pattern := range rule.ModelWhitelist {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}
			normalizedWhitelist = append(normalizedWhitelist, pattern)
		}
		settings.Rules[i].ModelWhitelist = normalizedWhitelist
		if len(normalizedWhitelist) == 0 {
			settings.Rules[i].FallbackAction = ""
			settings.Rules[i].FallbackErrorMessage = ""
			continue
		}
		if settings.Rules[i].FallbackAction == "" {
			settings.Rules[i].FallbackAction = BetaPolicyActionPass
			settings.Rules[i].FallbackErrorMessage = ""
			continue
		}
		if !validActions[settings.Rules[i].FallbackAction] {
			return fmt.Errorf("rule[%d]: invalid fallback_action %q", i, settings.Rules[i].FallbackAction)
		}
		if settings.Rules[i].FallbackAction != BetaPolicyActionBlock {
			settings.Rules[i].FallbackErrorMessage = ""
		}
	}

	return writeJSONSetting(ctx, s.settingRepo, SettingKeyBetaPolicySettings, "beta policy settings", settings)
}
