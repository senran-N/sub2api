package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/senran-N/sub2api/internal/config"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
)

const (
	defaultWeChatConnectMode     = "open"
	defaultWeChatConnectScopes   = "snsapi_login"
	defaultWeChatConnectFrontend = "/auth/wechat/callback"
)

func normalizeWeChatConnectModeSetting(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mp":
		return "mp"
	case "mobile":
		return "mobile"
	default:
		return defaultWeChatConnectMode
	}
}

func defaultWeChatConnectScopeForMode(mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return "snsapi_userinfo"
	case "mobile":
		return ""
	default:
		return defaultWeChatConnectScopes
	}
}

func normalizeWeChatConnectScopeSetting(raw, mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		scope := strings.TrimSpace(raw)
		switch scope {
		case "snsapi_base", "snsapi_userinfo":
			return scope
		default:
			return defaultWeChatConnectScopeForMode(mode)
		}
	case "mobile":
		return ""
	default:
		scope := strings.TrimSpace(raw)
		if scope == "" {
			return defaultWeChatConnectScopes
		}
		return scope
	}
}

func normalizeWeChatConnectStoredMode(openEnabled, mpEnabled, mobileEnabled bool, mode string) string {
	mode = normalizeWeChatConnectModeSetting(mode)
	switch mode {
	case "open":
		if openEnabled {
			return mode
		}
	case "mp":
		if mpEnabled {
			return mode
		}
	case "mobile":
		if mobileEnabled {
			return mode
		}
	}
	if openEnabled {
		return "open"
	}
	if mpEnabled {
		return "mp"
	}
	if mobileEnabled {
		return "mobile"
	}
	return mode
}

func mergeWeChatConnectCapabilitySettings(settings map[string]string, base config.WeChatConnectConfig, mode string) (bool, bool, bool) {
	openEnabled := base.OpenEnabled
	mpEnabled := base.MPEnabled
	mobileEnabled := base.MobileEnabled

	if raw, ok := settings[SettingKeyWeChatConnectOpenEnabled]; ok && strings.TrimSpace(raw) != "" {
		openEnabled = strings.TrimSpace(raw) == "true"
	}
	if raw, ok := settings[SettingKeyWeChatConnectMPEnabled]; ok && strings.TrimSpace(raw) != "" {
		mpEnabled = strings.TrimSpace(raw) == "true"
	}
	if raw, ok := settings[SettingKeyWeChatConnectMobileEnabled]; ok && strings.TrimSpace(raw) != "" {
		mobileEnabled = strings.TrimSpace(raw) == "true"
	}

	if !openEnabled && !mpEnabled && !mobileEnabled {
		switch normalizeWeChatConnectModeSetting(mode) {
		case "mp":
			mpEnabled = true
		case "mobile":
			mobileEnabled = true
		default:
			openEnabled = true
		}
	}

	return openEnabled, mpEnabled, mobileEnabled
}

func (s *SettingService) effectiveWeChatConnectOAuthConfig(settings map[string]string) config.WeChatConnectConfig {
	base := config.WeChatConnectConfig{}
	if s != nil && s.cfg != nil {
		base = s.cfg.WeChat
	}

	effective := base
	if raw, ok := settings[SettingKeyWeChatConnectEnabled]; ok {
		effective.Enabled = strings.TrimSpace(raw) == "true"
	}

	legacyAppID := strings.TrimSpace(firstNonEmpty(
		settings[SettingKeyWeChatConnectAppID],
		base.AppID,
		base.OpenAppID,
		base.MPAppID,
		base.MobileAppID,
	))
	legacyAppSecret := strings.TrimSpace(firstNonEmpty(
		settings[SettingKeyWeChatConnectAppSecret],
		base.AppSecret,
		base.OpenAppSecret,
		base.MPAppSecret,
		base.MobileAppSecret,
	))

	effective.AppID = legacyAppID
	effective.AppSecret = legacyAppSecret
	effective.OpenAppID = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectOpenAppID], base.OpenAppID, legacyAppID))
	effective.OpenAppSecret = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectOpenAppSecret], base.OpenAppSecret, legacyAppSecret))
	effective.MPAppID = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMPAppID], base.MPAppID, legacyAppID))
	effective.MPAppSecret = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMPAppSecret], base.MPAppSecret, legacyAppSecret))
	effective.MobileAppID = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMobileAppID], base.MobileAppID, legacyAppID))
	effective.MobileAppSecret = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMobileAppSecret], base.MobileAppSecret, legacyAppSecret))

	modeRaw := firstNonEmpty(settings[SettingKeyWeChatConnectMode], base.Mode)
	openEnabled, mpEnabled, mobileEnabled := mergeWeChatConnectCapabilitySettings(settings, base, modeRaw)
	effective.OpenEnabled = openEnabled
	effective.MPEnabled = mpEnabled
	effective.MobileEnabled = mobileEnabled
	effective.Mode = normalizeWeChatConnectStoredMode(openEnabled, mpEnabled, mobileEnabled, modeRaw)
	effective.Scopes = normalizeWeChatConnectScopeSetting(firstNonEmpty(settings[SettingKeyWeChatConnectScopes], base.Scopes), effective.Mode)
	effective.RedirectURL = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectRedirectURL], base.RedirectURL))
	effective.FrontendRedirectURL = strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectFrontendRedirectURL], base.FrontendRedirectURL, defaultWeChatConnectFrontend))
	return effective
}

func (s *SettingService) GetWeChatConnectOAuthConfig(ctx context.Context) (config.WeChatConnectConfig, error) {
	if s == nil || s.cfg == nil {
		return config.WeChatConnectConfig{}, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "config not loaded")
	}

	keys := []string{
		SettingKeyWeChatConnectEnabled,
		SettingKeyWeChatConnectAppID,
		SettingKeyWeChatConnectAppSecret,
		SettingKeyWeChatConnectOpenAppID,
		SettingKeyWeChatConnectOpenAppSecret,
		SettingKeyWeChatConnectMPAppID,
		SettingKeyWeChatConnectMPAppSecret,
		SettingKeyWeChatConnectMobileAppID,
		SettingKeyWeChatConnectMobileAppSecret,
		SettingKeyWeChatConnectOpenEnabled,
		SettingKeyWeChatConnectMPEnabled,
		SettingKeyWeChatConnectMobileEnabled,
		SettingKeyWeChatConnectMode,
		SettingKeyWeChatConnectScopes,
		SettingKeyWeChatConnectRedirectURL,
		SettingKeyWeChatConnectFrontendRedirectURL,
	}
	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return config.WeChatConnectConfig{}, fmt.Errorf("get wechat connect settings: %w", err)
	}

	effective := s.effectiveWeChatConnectOAuthConfig(settings)
	if !effective.Enabled {
		return config.WeChatConnectConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "oauth login is disabled")
	}
	if strings.TrimSpace(effective.RedirectURL) == "" {
		return config.WeChatConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat redirect url not configured")
	}
	if strings.TrimSpace(effective.FrontendRedirectURL) == "" {
		return config.WeChatConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat frontend redirect url not configured")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.RedirectURL); err != nil {
		return config.WeChatConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat redirect url invalid")
	}
	if err := config.ValidateFrontendRedirectURL(effective.FrontendRedirectURL); err != nil {
		return config.WeChatConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat frontend redirect url invalid")
	}

	mode := normalizeWeChatConnectModeSetting(effective.Mode)
	appID, appSecret := wechatCredentialsForMode(effective, mode)
	if strings.TrimSpace(appID) == "" {
		return config.WeChatConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat app id not configured")
	}
	if strings.TrimSpace(appSecret) == "" {
		return config.WeChatConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat app secret not configured")
	}

	return effective, nil
}

func wechatCredentialsForMode(cfg config.WeChatConnectConfig, mode string) (string, string) {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return strings.TrimSpace(firstNonEmpty(cfg.MPAppID, cfg.AppID)), strings.TrimSpace(firstNonEmpty(cfg.MPAppSecret, cfg.AppSecret))
	case "mobile":
		return strings.TrimSpace(firstNonEmpty(cfg.MobileAppID, cfg.AppID)), strings.TrimSpace(firstNonEmpty(cfg.MobileAppSecret, cfg.AppSecret))
	default:
		return strings.TrimSpace(firstNonEmpty(cfg.OpenAppID, cfg.AppID)), strings.TrimSpace(firstNonEmpty(cfg.OpenAppSecret, cfg.AppSecret))
	}
}
