package service

import "strings"

func wechatCredentialsForMode(cfg WeChatConnectOAuthConfig, mode string) (string, string) {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return strings.TrimSpace(firstNonEmptyString(cfg.MPAppID, cfg.LegacyAppID)), strings.TrimSpace(firstNonEmptyString(cfg.MPAppSecret, cfg.LegacyAppSecret))
	case "mobile":
		return strings.TrimSpace(firstNonEmptyString(cfg.MobileAppID, cfg.LegacyAppID)), strings.TrimSpace(firstNonEmptyString(cfg.MobileAppSecret, cfg.LegacyAppSecret))
	default:
		return strings.TrimSpace(firstNonEmptyString(cfg.OpenAppID, cfg.LegacyAppID)), strings.TrimSpace(firstNonEmptyString(cfg.OpenAppSecret, cfg.LegacyAppSecret))
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
