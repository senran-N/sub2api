package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// GetFrontendURL 获取前端基础URL（数据库优先，fallback 到配置文件）
func (s *SettingService) GetFrontendURL(ctx context.Context) string {
	val, err := s.settingRepo.GetValue(ctx, SettingKeyFrontendURL)
	if err == nil && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return s.cfg.Server.FrontendURL
}

// GetPublicSettings 获取公开设置（无需登录）
func (s *SettingService) GetPublicSettings(ctx context.Context) (*PublicSettings, error) {
	keys := []string{
		SettingKeyRegistrationEnabled,
		SettingKeyEmailVerifyEnabled,
		SettingKeyForceEmailOnThirdPartySignup,
		SettingKeyRegistrationEmailSuffixWhitelist,
		SettingKeyPromoCodeEnabled,
		SettingKeyPasswordResetEnabled,
		SettingKeyInvitationCodeEnabled,
		SettingKeyTotpEnabled,
		SettingKeyTurnstileEnabled,
		SettingKeyTurnstileSiteKey,
		SettingKeySiteName,
		SettingKeySiteLogo,
		SettingKeySiteSubtitle,
		SettingKeyFrontendTheme,
		SettingKeyAPIBaseURL,
		SettingKeyContactInfo,
		SettingKeyDocURL,
		SettingKeyHomeContent,
		SettingKeyHideCcsImportButton,
		SettingKeyPurchaseSubscriptionEnabled,
		SettingKeyPurchaseSubscriptionURL,
		SettingKeyCustomMenuItems,
		SettingKeyCustomEndpoints,
		SettingKeyLinuxDoConnectEnabled,
		SettingKeyWeChatConnectEnabled,
		SettingKeyWeChatConnectOpenEnabled,
		SettingKeyWeChatConnectMPEnabled,
		SettingKeyWeChatConnectMobileEnabled,
		SettingKeyOIDCConnectEnabled,
		SettingKeyOIDCConnectProviderName,
		SettingKeyBackendModeEnabled,
		SettingPaymentEnabled,
		SettingKeyBalanceLowNotifyEnabled,
		SettingKeyAccountQuotaNotifyEnabled,
		SettingKeyBalanceLowNotifyThreshold,
		SettingKeyBalanceLowNotifyRechargeURL,
	}

	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get public settings: %w", err)
	}

	linuxDoEnabled := false
	if raw, ok := settings[SettingKeyLinuxDoConnectEnabled]; ok {
		linuxDoEnabled = raw == "true"
	} else {
		linuxDoEnabled = s.cfg != nil && s.cfg.LinuxDo.Enabled
	}
	wechatEnabled := false
	wechatOpenEnabled := false
	wechatMPEnabled := false
	wechatMobileEnabled := false
	if s != nil {
		effective := s.effectiveWeChatConnectOAuthConfig(settings)
		wechatEnabled = effective.Enabled
		wechatOpenEnabled = effective.OpenEnabled
		wechatMPEnabled = effective.MPEnabled
		wechatMobileEnabled = effective.MobileEnabled
	}
	oidcEnabled := false
	oidcProviderName := "OIDC"
	if raw, ok := settings[SettingKeyOIDCConnectEnabled]; ok {
		oidcEnabled = raw == "true"
	} else {
		oidcEnabled = s.cfg != nil && s.cfg.OIDC.Enabled
	}
	if raw, ok := settings[SettingKeyOIDCConnectProviderName]; ok && strings.TrimSpace(raw) != "" {
		oidcProviderName = strings.TrimSpace(raw)
	} else if s.cfg != nil && strings.TrimSpace(s.cfg.OIDC.ProviderName) != "" {
		oidcProviderName = strings.TrimSpace(s.cfg.OIDC.ProviderName)
	}

	emailVerifyEnabled := settings[SettingKeyEmailVerifyEnabled] == "true"
	passwordResetEnabled := emailVerifyEnabled && settings[SettingKeyPasswordResetEnabled] == "true"
	registrationEmailSuffixWhitelist := ParseRegistrationEmailSuffixWhitelist(
		settings[SettingKeyRegistrationEmailSuffixWhitelist],
	)

	return &PublicSettings{
		RegistrationEnabled:              settings[SettingKeyRegistrationEnabled] == "true",
		EmailVerifyEnabled:               emailVerifyEnabled,
		ForceEmailOnThirdPartySignup:     settings[SettingKeyForceEmailOnThirdPartySignup] == "true",
		RegistrationEmailSuffixWhitelist: registrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings[SettingKeyPromoCodeEnabled] != "false",
		PasswordResetEnabled:             passwordResetEnabled,
		InvitationCodeEnabled:            settings[SettingKeyInvitationCodeEnabled] == "true",
		TotpEnabled:                      settings[SettingKeyTotpEnabled] == "true",
		TurnstileEnabled:                 settings[SettingKeyTurnstileEnabled] == "true",
		TurnstileSiteKey:                 settings[SettingKeyTurnstileSiteKey],
		SiteName:                         s.getStringOrDefault(settings, SettingKeySiteName, "Sub2API"),
		SiteLogo:                         settings[SettingKeySiteLogo],
		SiteSubtitle:                     s.getStringOrDefault(settings, SettingKeySiteSubtitle, "Subscription to API Conversion Platform"),
		FrontendTheme:                    s.getStringOrDefault(settings, SettingKeyFrontendTheme, "factory"),
		APIBaseURL:                       settings[SettingKeyAPIBaseURL],
		ContactInfo:                      settings[SettingKeyContactInfo],
		DocURL:                           settings[SettingKeyDocURL],
		HomeContent:                      settings[SettingKeyHomeContent],
		HideCcsImportButton:              settings[SettingKeyHideCcsImportButton] == "true",
		PurchaseSubscriptionEnabled:      settings[SettingKeyPurchaseSubscriptionEnabled] == "true",
		PurchaseSubscriptionURL:          strings.TrimSpace(settings[SettingKeyPurchaseSubscriptionURL]),
		CustomMenuItems:                  settings[SettingKeyCustomMenuItems],
		CustomEndpoints:                  settings[SettingKeyCustomEndpoints],
		LinuxDoOAuthEnabled:              linuxDoEnabled,
		WeChatOAuthEnabled:               wechatEnabled,
		WeChatOAuthOpenEnabled:           wechatOpenEnabled,
		WeChatOAuthMPEnabled:             wechatMPEnabled,
		WeChatOAuthMobileEnabled:         wechatMobileEnabled,
		OIDCOAuthEnabled:                 oidcEnabled,
		OIDCOAuthProviderName:            oidcProviderName,
		BackendModeEnabled:               settings[SettingKeyBackendModeEnabled] == "true",
		PaymentEnabled:                   settings[SettingPaymentEnabled] == "true",
		BalanceLowNotifyEnabled:          settings[SettingKeyBalanceLowNotifyEnabled] == "true",
		AccountQuotaNotifyEnabled:        settings[SettingKeyAccountQuotaNotifyEnabled] == "true",
		BalanceLowNotifyThreshold:        parsePublicFloatSetting(settings[SettingKeyBalanceLowNotifyThreshold]),
		BalanceLowNotifyRechargeURL:      strings.TrimSpace(settings[SettingKeyBalanceLowNotifyRechargeURL]),
	}, nil
}

// GetPublicSettingsForInjection returns public settings in a format suitable for HTML injection.
func (s *SettingService) GetPublicSettingsForInjection(ctx context.Context) (any, error) {
	settings, err := s.GetPublicSettings(ctx)
	if err != nil {
		return nil, err
	}

	return &struct {
		RegistrationEnabled              bool            `json:"registration_enabled"`
		EmailVerifyEnabled               bool            `json:"email_verify_enabled"`
		ForceEmailOnThirdPartySignup     bool            `json:"force_email_on_third_party_signup"`
		RegistrationEmailSuffixWhitelist []string        `json:"registration_email_suffix_whitelist"`
		PromoCodeEnabled                 bool            `json:"promo_code_enabled"`
		PasswordResetEnabled             bool            `json:"password_reset_enabled"`
		InvitationCodeEnabled            bool            `json:"invitation_code_enabled"`
		TotpEnabled                      bool            `json:"totp_enabled"`
		TurnstileEnabled                 bool            `json:"turnstile_enabled"`
		TurnstileSiteKey                 string          `json:"turnstile_site_key,omitempty"`
		SiteName                         string          `json:"site_name"`
		SiteLogo                         string          `json:"site_logo,omitempty"`
		SiteSubtitle                     string          `json:"site_subtitle,omitempty"`
		FrontendTheme                    string          `json:"frontend_theme,omitempty"`
		APIBaseURL                       string          `json:"api_base_url,omitempty"`
		ContactInfo                      string          `json:"contact_info,omitempty"`
		DocURL                           string          `json:"doc_url,omitempty"`
		HomeContent                      string          `json:"home_content,omitempty"`
		HideCcsImportButton              bool            `json:"hide_ccs_import_button"`
		PurchaseSubscriptionEnabled      bool            `json:"purchase_subscription_enabled"`
		PurchaseSubscriptionURL          string          `json:"purchase_subscription_url,omitempty"`
		CustomMenuItems                  json.RawMessage `json:"custom_menu_items"`
		CustomEndpoints                  json.RawMessage `json:"custom_endpoints"`
		LinuxDoOAuthEnabled              bool            `json:"linuxdo_oauth_enabled"`
		WeChatOAuthEnabled               bool            `json:"wechat_oauth_enabled"`
		WeChatOAuthOpenEnabled           bool            `json:"wechat_oauth_open_enabled"`
		WeChatOAuthMPEnabled             bool            `json:"wechat_oauth_mp_enabled"`
		WeChatOAuthMobileEnabled         bool            `json:"wechat_oauth_mobile_enabled"`
		OIDCOAuthEnabled                 bool            `json:"oidc_oauth_enabled"`
		OIDCOAuthProviderName            string          `json:"oidc_oauth_provider_name"`
		BackendModeEnabled               bool            `json:"backend_mode_enabled"`
		PaymentEnabled                   bool            `json:"payment_enabled"`
		BalanceLowNotifyEnabled          bool            `json:"balance_low_notify_enabled"`
		AccountQuotaNotifyEnabled        bool            `json:"account_quota_notify_enabled"`
		BalanceLowNotifyThreshold        float64         `json:"balance_low_notify_threshold"`
		BalanceLowNotifyRechargeURL      string          `json:"balance_low_notify_recharge_url"`
		Version                          string          `json:"version,omitempty"`
	}{
		RegistrationEnabled:              settings.RegistrationEnabled,
		EmailVerifyEnabled:               settings.EmailVerifyEnabled,
		ForceEmailOnThirdPartySignup:     settings.ForceEmailOnThirdPartySignup,
		RegistrationEmailSuffixWhitelist: settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings.PromoCodeEnabled,
		PasswordResetEnabled:             settings.PasswordResetEnabled,
		InvitationCodeEnabled:            settings.InvitationCodeEnabled,
		TotpEnabled:                      settings.TotpEnabled,
		TurnstileEnabled:                 settings.TurnstileEnabled,
		TurnstileSiteKey:                 settings.TurnstileSiteKey,
		SiteName:                         settings.SiteName,
		SiteLogo:                         settings.SiteLogo,
		SiteSubtitle:                     settings.SiteSubtitle,
		FrontendTheme:                    settings.FrontendTheme,
		APIBaseURL:                       settings.APIBaseURL,
		ContactInfo:                      settings.ContactInfo,
		DocURL:                           settings.DocURL,
		HomeContent:                      settings.HomeContent,
		HideCcsImportButton:              settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:      settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:          settings.PurchaseSubscriptionURL,
		CustomMenuItems:                  filterUserVisibleMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                  safeRawJSONArray(settings.CustomEndpoints),
		LinuxDoOAuthEnabled:              settings.LinuxDoOAuthEnabled,
		WeChatOAuthEnabled:               settings.WeChatOAuthEnabled,
		WeChatOAuthOpenEnabled:           settings.WeChatOAuthOpenEnabled,
		WeChatOAuthMPEnabled:             settings.WeChatOAuthMPEnabled,
		WeChatOAuthMobileEnabled:         settings.WeChatOAuthMobileEnabled,
		OIDCOAuthEnabled:                 settings.OIDCOAuthEnabled,
		OIDCOAuthProviderName:            settings.OIDCOAuthProviderName,
		BackendModeEnabled:               settings.BackendModeEnabled,
		PaymentEnabled:                   settings.PaymentEnabled,
		BalanceLowNotifyEnabled:          settings.BalanceLowNotifyEnabled,
		AccountQuotaNotifyEnabled:        settings.AccountQuotaNotifyEnabled,
		BalanceLowNotifyThreshold:        settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings.BalanceLowNotifyRechargeURL,
		Version:                          s.version,
	}, nil
}

// GetFrameSrcOrigins returns deduplicated http(s) origins from home_content URL,
// purchase_subscription_url, and all custom_menu_items URLs. Used by the router layer for CSP frame-src injection.
func (s *SettingService) GetFrameSrcOrigins(ctx context.Context) ([]string, error) {
	settings, err := s.GetPublicSettings(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var origins []string

	addOrigin := func(rawURL string) {
		if origin := extractOriginFromURL(rawURL); origin != "" {
			if _, ok := seen[origin]; !ok {
				seen[origin] = struct{}{}
				origins = append(origins, origin)
			}
		}
	}

	if settings.PurchaseSubscriptionEnabled {
		addOrigin(settings.PurchaseSubscriptionURL)
	}
	addOrigin(settings.HomeContent)
	for _, item := range parseCustomMenuItemURLs(settings.CustomMenuItems) {
		addOrigin(item)
	}

	return origins, nil
}

// filterUserVisibleMenuItems filters out admin-only menu items from a raw JSON array string.
func filterUserVisibleMenuItems(raw string) json.RawMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return json.RawMessage("[]")
	}

	var items []struct {
		Visibility string `json:"visibility"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return json.RawMessage("[]")
	}

	var fullItems []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fullItems); err != nil {
		return json.RawMessage("[]")
	}

	var filtered []json.RawMessage
	for i, item := range items {
		if item.Visibility != "admin" {
			filtered = append(filtered, fullItems[i])
		}
	}
	if len(filtered) == 0 {
		return json.RawMessage("[]")
	}

	result, err := json.Marshal(filtered)
	if err != nil {
		return json.RawMessage("[]")
	}
	return result
}

// safeRawJSONArray returns raw as json.RawMessage if it's valid JSON, otherwise "[]".
func safeRawJSONArray(raw string) json.RawMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return json.RawMessage("[]")
	}
	if json.Valid([]byte(raw)) {
		return json.RawMessage(raw)
	}
	return json.RawMessage("[]")
}

// extractOriginFromURL returns the scheme+host origin from rawURL.
func extractOriginFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Host == "" {
		return ""
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ""
	}
	return parsedURL.Scheme + "://" + parsedURL.Host
}

// parseCustomMenuItemURLs extracts URLs from a raw JSON array of custom menu items.
func parseCustomMenuItemURLs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}

	var items []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}

	urls := make([]string, 0, len(items))
	for _, item := range items {
		if item.URL != "" {
			urls = append(urls, item.URL)
		}
	}
	return urls
}

func parsePublicFloatSetting(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return value
}
