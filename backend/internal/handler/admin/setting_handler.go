package admin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/handler/dto"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/senran-N/sub2api/internal/pkg/response"
	"github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// semverPattern 预编译 semver 格式校验正则
var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// menuItemIDPattern validates custom menu item IDs: alphanumeric, hyphens, underscores only.
var menuItemIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// generateMenuItemID generates a short random hex ID for a custom menu item.
func generateMenuItemID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate menu item ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// SettingHandler 系统设置处理器
type SettingHandler struct {
	settingService   *service.SettingService
	emailService     *service.EmailService
	turnstileService *service.TurnstileService
	opsService       *service.OpsService
}

// NewSettingHandler 创建系统设置处理器
func NewSettingHandler(settingService *service.SettingService, emailService *service.EmailService, turnstileService *service.TurnstileService, opsService *service.OpsService) *SettingHandler {
	return &SettingHandler{
		settingService:   settingService,
		emailService:     emailService,
		turnstileService: turnstileService,
		opsService:       opsService,
	}
}

type ModelCatalogEntry struct {
	ID             string   `json:"id"`
	DisplayName    string   `json:"display_name"`
	Capability     string   `json:"capability"`
	ProtocolFamily string   `json:"protocol_family"`
	RequiredTier   string   `json:"required_tier"`
	Aliases        []string `json:"aliases,omitempty"`
	SupportsStream bool     `json:"supports_stream"`
	SupportsTools  bool     `json:"supports_tools"`
}

type ModelCatalogResponse struct {
	Platform string              `json:"platform"`
	Models   []ModelCatalogEntry `json:"models"`
}

// GetModelCatalog returns the backend-owned model catalog for a platform.
// GET /api/v1/admin/model-catalog?platform=grok
func (h *SettingHandler) GetModelCatalog(c *gin.Context) {
	platform := strings.ToLower(strings.TrimSpace(c.Query("platform")))
	switch platform {
	case service.PlatformGrok:
		specs := grok.Specs()
		models := make([]ModelCatalogEntry, 0, len(specs))
		for _, spec := range specs {
			models = append(models, ModelCatalogEntry{
				ID:             spec.ID,
				DisplayName:    spec.DisplayName,
				Capability:     string(spec.Capability),
				ProtocolFamily: string(spec.ProtocolFamily),
				RequiredTier:   string(spec.RequiredTier),
				Aliases:        append([]string(nil), spec.Aliases...),
				SupportsStream: spec.SupportsStream,
				SupportsTools:  spec.SupportsTools,
			})
		}
		response.Success(c, ModelCatalogResponse{
			Platform: platform,
			Models:   models,
		})
		return
	default:
		response.BadRequest(c, "unsupported platform")
	}
}

// GetSettings 获取所有系统设置
// GET /api/v1/admin/settings
func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	authSourceDefaults, err := h.settingService.GetAuthSourceDefaultSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Check if ops monitoring is enabled (respects config.ops.enabled)
	opsEnabled := h.opsService != nil && h.opsService.IsMonitoringEnabled(c.Request.Context())
	defaultSubscriptions := make([]dto.DefaultSubscriptionSetting, 0, len(settings.DefaultSubscriptions))
	for _, sub := range settings.DefaultSubscriptions {
		defaultSubscriptions = append(defaultSubscriptions, dto.DefaultSubscriptionSetting{
			GroupID:      sub.GroupID,
			ValidityDays: sub.ValidityDays,
		})
	}

	response.Success(c, dto.SystemSettings{
		RegistrationEnabled:                      settings.RegistrationEnabled,
		EmailVerifyEnabled:                       settings.EmailVerifyEnabled,
		ForceEmailOnThirdPartySignup:             authSourceDefaults != nil && authSourceDefaults.ForceEmailOnThirdPartySignup,
		RegistrationEmailSuffixWhitelist:         settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                         settings.PromoCodeEnabled,
		PasswordResetEnabled:                     settings.PasswordResetEnabled,
		FrontendURL:                              settings.FrontendURL,
		InvitationCodeEnabled:                    settings.InvitationCodeEnabled,
		TotpEnabled:                              settings.TotpEnabled,
		TotpEncryptionKeyConfigured:              h.settingService.IsTotpEncryptionKeyConfigured(),
		AffiliateEnabled:                         settings.AffiliateEnabled,
		AffiliateRebateRatePercent:               settings.AffiliateRebateRatePercent,
		AffiliateRebateFreezeHours:               settings.AffiliateRebateFreezeHours,
		AffiliateRebateDurationDays:              settings.AffiliateRebateDurationDays,
		AffiliateRebatePerInviteeCap:             settings.AffiliateRebatePerInviteeCap,
		SMTPHost:                                 settings.SMTPHost,
		SMTPPort:                                 settings.SMTPPort,
		SMTPUsername:                             settings.SMTPUsername,
		SMTPPasswordConfigured:                   settings.SMTPPasswordConfigured,
		SMTPFrom:                                 settings.SMTPFrom,
		SMTPFromName:                             settings.SMTPFromName,
		SMTPUseTLS:                               settings.SMTPUseTLS,
		BalanceLowNotifyEnabled:                  settings.BalanceLowNotifyEnabled,
		BalanceLowNotifyThreshold:                settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:              settings.BalanceLowNotifyRechargeURL,
		AccountQuotaNotifyEnabled:                settings.AccountQuotaNotifyEnabled,
		AccountQuotaNotifyEmails:                 dto.NotifyEmailEntriesFromService(settings.AccountQuotaNotifyEmails),
		TurnstileEnabled:                         settings.TurnstileEnabled,
		TurnstileSiteKey:                         settings.TurnstileSiteKey,
		TurnstileSecretKeyConfigured:             settings.TurnstileSecretKeyConfigured,
		LinuxDoConnectEnabled:                    settings.LinuxDoConnectEnabled,
		LinuxDoConnectClientID:                   settings.LinuxDoConnectClientID,
		LinuxDoConnectClientSecretConfigured:     settings.LinuxDoConnectClientSecretConfigured,
		LinuxDoConnectRedirectURL:                settings.LinuxDoConnectRedirectURL,
		WeChatConnectEnabled:                     settings.WeChatConnectEnabled,
		WeChatConnectOpenAppID:                   settings.WeChatConnectOpenAppID,
		WeChatConnectOpenAppSecretConfigured:     settings.WeChatConnectOpenAppSecretConfigured,
		WeChatConnectMPAppID:                     settings.WeChatConnectMPAppID,
		WeChatConnectMPAppSecretConfigured:       settings.WeChatConnectMPAppSecretConfigured,
		WeChatConnectMobileAppID:                 settings.WeChatConnectMobileAppID,
		WeChatConnectMobileAppSecretConfigured:   settings.WeChatConnectMobileAppSecretConfigured,
		WeChatConnectOpenEnabled:                 settings.WeChatConnectOpenEnabled,
		WeChatConnectMPEnabled:                   settings.WeChatConnectMPEnabled,
		WeChatConnectMobileEnabled:               settings.WeChatConnectMobileEnabled,
		WeChatConnectMode:                        settings.WeChatConnectMode,
		WeChatConnectScopes:                      settings.WeChatConnectScopes,
		WeChatConnectRedirectURL:                 settings.WeChatConnectRedirectURL,
		WeChatConnectFrontendRedirectURL:         settings.WeChatConnectFrontendRedirectURL,
		OIDCConnectEnabled:                       settings.OIDCConnectEnabled,
		OIDCConnectProviderName:                  settings.OIDCConnectProviderName,
		OIDCConnectClientID:                      settings.OIDCConnectClientID,
		OIDCConnectClientSecretConfigured:        settings.OIDCConnectClientSecretConfigured,
		OIDCConnectIssuerURL:                     settings.OIDCConnectIssuerURL,
		OIDCConnectDiscoveryURL:                  settings.OIDCConnectDiscoveryURL,
		OIDCConnectAuthorizeURL:                  settings.OIDCConnectAuthorizeURL,
		OIDCConnectTokenURL:                      settings.OIDCConnectTokenURL,
		OIDCConnectUserInfoURL:                   settings.OIDCConnectUserInfoURL,
		OIDCConnectJWKSURL:                       settings.OIDCConnectJWKSURL,
		OIDCConnectScopes:                        settings.OIDCConnectScopes,
		OIDCConnectRedirectURL:                   settings.OIDCConnectRedirectURL,
		OIDCConnectFrontendRedirectURL:           settings.OIDCConnectFrontendRedirectURL,
		OIDCConnectTokenAuthMethod:               settings.OIDCConnectTokenAuthMethod,
		OIDCConnectUsePKCE:                       settings.OIDCConnectUsePKCE,
		OIDCConnectValidateIDToken:               settings.OIDCConnectValidateIDToken,
		OIDCConnectAllowedSigningAlgs:            settings.OIDCConnectAllowedSigningAlgs,
		OIDCConnectClockSkewSeconds:              settings.OIDCConnectClockSkewSeconds,
		OIDCConnectRequireEmailVerified:          settings.OIDCConnectRequireEmailVerified,
		OIDCConnectUserInfoEmailPath:             settings.OIDCConnectUserInfoEmailPath,
		OIDCConnectUserInfoIDPath:                settings.OIDCConnectUserInfoIDPath,
		OIDCConnectUserInfoUsernamePath:          settings.OIDCConnectUserInfoUsernamePath,
		SiteName:                                 settings.SiteName,
		SiteLogo:                                 settings.SiteLogo,
		SiteSubtitle:                             settings.SiteSubtitle,
		FrontendTheme:                            settings.FrontendTheme,
		APIBaseURL:                               settings.APIBaseURL,
		ContactInfo:                              settings.ContactInfo,
		DocURL:                                   settings.DocURL,
		HomeContent:                              settings.HomeContent,
		HideCcsImportButton:                      settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:              settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:                  settings.PurchaseSubscriptionURL,
		CustomMenuItems:                          dto.ParseCustomMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                          dto.ParseCustomEndpoints(settings.CustomEndpoints),
		DefaultConcurrency:                       settings.DefaultConcurrency,
		DefaultBalance:                           settings.DefaultBalance,
		DefaultSubscriptions:                     defaultSubscriptions,
		AuthSourceDefaultEmailBalance:            authSourceDefaults.Email.Balance,
		AuthSourceDefaultEmailConcurrency:        authSourceDefaults.Email.Concurrency,
		AuthSourceDefaultEmailSubscriptions:      dto.DefaultSubscriptionSettingsFromService(authSourceDefaults.Email.Subscriptions),
		AuthSourceDefaultEmailGrantOnSignup:      authSourceDefaults.Email.GrantOnSignup,
		AuthSourceDefaultEmailGrantOnFirstBind:   authSourceDefaults.Email.GrantOnFirstBind,
		AuthSourceDefaultLinuxDoBalance:          authSourceDefaults.LinuxDo.Balance,
		AuthSourceDefaultLinuxDoConcurrency:      authSourceDefaults.LinuxDo.Concurrency,
		AuthSourceDefaultLinuxDoSubscriptions:    dto.DefaultSubscriptionSettingsFromService(authSourceDefaults.LinuxDo.Subscriptions),
		AuthSourceDefaultLinuxDoGrantOnSignup:    authSourceDefaults.LinuxDo.GrantOnSignup,
		AuthSourceDefaultLinuxDoGrantOnFirstBind: authSourceDefaults.LinuxDo.GrantOnFirstBind,
		AuthSourceDefaultOIDCBalance:             authSourceDefaults.OIDC.Balance,
		AuthSourceDefaultOIDCConcurrency:         authSourceDefaults.OIDC.Concurrency,
		AuthSourceDefaultOIDCSubscriptions:       dto.DefaultSubscriptionSettingsFromService(authSourceDefaults.OIDC.Subscriptions),
		AuthSourceDefaultOIDCGrantOnSignup:       authSourceDefaults.OIDC.GrantOnSignup,
		AuthSourceDefaultOIDCGrantOnFirstBind:    authSourceDefaults.OIDC.GrantOnFirstBind,
		AuthSourceDefaultWeChatBalance:           authSourceDefaults.WeChat.Balance,
		AuthSourceDefaultWeChatConcurrency:       authSourceDefaults.WeChat.Concurrency,
		AuthSourceDefaultWeChatSubscriptions:     dto.DefaultSubscriptionSettingsFromService(authSourceDefaults.WeChat.Subscriptions),
		AuthSourceDefaultWeChatGrantOnSignup:     authSourceDefaults.WeChat.GrantOnSignup,
		AuthSourceDefaultWeChatGrantOnFirstBind:  authSourceDefaults.WeChat.GrantOnFirstBind,
		EnableModelFallback:                      settings.EnableModelFallback,
		FallbackModelAnthropic:                   settings.FallbackModelAnthropic,
		FallbackModelOpenAI:                      settings.FallbackModelOpenAI,
		FallbackModelGrok:                        settings.FallbackModelGrok,
		FallbackModelGemini:                      settings.FallbackModelGemini,
		FallbackModelAntigravity:                 settings.FallbackModelAntigravity,
		GrokOfficialBaseURL:                      settings.GrokOfficialBaseURL,
		GrokSessionBaseURL:                       settings.GrokSessionBaseURL,
		GrokThinkingSummary:                      settings.GrokThinkingSummary,
		GrokShowSearchSources:                    settings.GrokShowSearchSources,
		GrokImageOutputFormat:                    settings.GrokImageOutputFormat,
		GrokVideoOutputFormat:                    settings.GrokVideoOutputFormat,
		GrokMediaProxyEnabled:                    settings.GrokMediaProxyEnabled,
		GrokMediaCacheRetentionHours:             settings.GrokMediaCacheRetentionHours,
		GrokQuotaSyncIntervalSeconds:             settings.GrokQuotaSyncIntervalSeconds,
		GrokUsageSyncConcurrency:                 settings.GrokUsageSyncConcurrency,
		GrokCapabilityProbeIntervalSeconds:       settings.GrokCapabilityProbeIntervalSeconds,
		GrokCapabilityProbeConcurrency:           settings.GrokCapabilityProbeConcurrency,
		GrokSessionValidityCheckIntervalSeconds:  settings.GrokSessionValidityCheckIntervalSeconds,
		GrokVideoTimeoutSeconds:                  settings.GrokVideoTimeoutSeconds,
		EnableIdentityPatch:                      settings.EnableIdentityPatch,
		IdentityPatchPrompt:                      settings.IdentityPatchPrompt,
		OpsMonitoringEnabled:                     opsEnabled && settings.OpsMonitoringEnabled,
		OpsRealtimeMonitoringEnabled:             settings.OpsRealtimeMonitoringEnabled,
		OpsQueryModeDefault:                      settings.OpsQueryModeDefault,
		OpsMetricsIntervalSeconds:                settings.OpsMetricsIntervalSeconds,
		MinClaudeCodeVersion:                     settings.MinClaudeCodeVersion,
		MaxClaudeCodeVersion:                     settings.MaxClaudeCodeVersion,
		AllowUngroupedKeyScheduling:              settings.AllowUngroupedKeyScheduling,
		BackendModeEnabled:                       settings.BackendModeEnabled,
		EnableFingerprintUnification:             settings.EnableFingerprintUnification,
		EnableMetadataPassthrough:                settings.EnableMetadataPassthrough,
		EnableCCHSigning:                         settings.EnableCCHSigning,
	})
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	// 注册设置
	RegistrationEnabled              bool     `json:"registration_enabled"`
	EmailVerifyEnabled               bool     `json:"email_verify_enabled"`
	ForceEmailOnThirdPartySignup     *bool    `json:"force_email_on_third_party_signup"`
	RegistrationEmailSuffixWhitelist []string `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool     `json:"password_reset_enabled"`
	FrontendURL                      string   `json:"frontend_url"`
	InvitationCodeEnabled            bool     `json:"invitation_code_enabled"`
	TotpEnabled                      bool     `json:"totp_enabled"` // TOTP 双因素认证
	AffiliateEnabled                 *bool    `json:"affiliate_enabled"`
	AffiliateRebateRatePercent       *float64 `json:"affiliate_rebate_rate"`
	AffiliateRebateFreezeHours       *int     `json:"affiliate_rebate_freeze_hours"`
	AffiliateRebateDurationDays      *int     `json:"affiliate_rebate_duration_days"`
	AffiliateRebatePerInviteeCap     *float64 `json:"affiliate_rebate_per_invitee_cap"`

	// 邮件服务设置
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from_email"`
	SMTPFromName string `json:"smtp_from_name"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`

	BalanceLowNotifyEnabled     *bool                   `json:"balance_low_notify_enabled"`
	BalanceLowNotifyThreshold   *float64                `json:"balance_low_notify_threshold"`
	BalanceLowNotifyRechargeURL *string                 `json:"balance_low_notify_recharge_url"`
	AccountQuotaNotifyEnabled   *bool                   `json:"account_quota_notify_enabled"`
	AccountQuotaNotifyEmails    *[]dto.NotifyEmailEntry `json:"account_quota_notify_emails"`

	// Cloudflare Turnstile 设置
	TurnstileEnabled   bool   `json:"turnstile_enabled"`
	TurnstileSiteKey   string `json:"turnstile_site_key"`
	TurnstileSecretKey string `json:"turnstile_secret_key"`

	// LinuxDo Connect OAuth 登录
	LinuxDoConnectEnabled      bool   `json:"linuxdo_connect_enabled"`
	LinuxDoConnectClientID     string `json:"linuxdo_connect_client_id"`
	LinuxDoConnectClientSecret string `json:"linuxdo_connect_client_secret"`
	LinuxDoConnectRedirectURL  string `json:"linuxdo_connect_redirect_url"`

	// WeChat Connect OAuth 登录
	WeChatConnectEnabled             bool   `json:"wechat_connect_enabled"`
	WeChatConnectOpenAppID           string `json:"wechat_connect_open_app_id"`
	WeChatConnectOpenAppSecret       string `json:"wechat_connect_open_app_secret"`
	WeChatConnectMPAppID             string `json:"wechat_connect_mp_app_id"`
	WeChatConnectMPAppSecret         string `json:"wechat_connect_mp_app_secret"`
	WeChatConnectMobileAppID         string `json:"wechat_connect_mobile_app_id"`
	WeChatConnectMobileAppSecret     string `json:"wechat_connect_mobile_app_secret"`
	WeChatConnectOpenEnabled         bool   `json:"wechat_connect_open_enabled"`
	WeChatConnectMPEnabled           bool   `json:"wechat_connect_mp_enabled"`
	WeChatConnectMobileEnabled       bool   `json:"wechat_connect_mobile_enabled"`
	WeChatConnectMode                string `json:"wechat_connect_mode"`
	WeChatConnectScopes              string `json:"wechat_connect_scopes"`
	WeChatConnectRedirectURL         string `json:"wechat_connect_redirect_url"`
	WeChatConnectFrontendRedirectURL string `json:"wechat_connect_frontend_redirect_url"`

	// Generic OIDC OAuth 登录
	OIDCConnectEnabled              bool   `json:"oidc_connect_enabled"`
	OIDCConnectProviderName         string `json:"oidc_connect_provider_name"`
	OIDCConnectClientID             string `json:"oidc_connect_client_id"`
	OIDCConnectClientSecret         string `json:"oidc_connect_client_secret"`
	OIDCConnectIssuerURL            string `json:"oidc_connect_issuer_url"`
	OIDCConnectDiscoveryURL         string `json:"oidc_connect_discovery_url"`
	OIDCConnectAuthorizeURL         string `json:"oidc_connect_authorize_url"`
	OIDCConnectTokenURL             string `json:"oidc_connect_token_url"`
	OIDCConnectUserInfoURL          string `json:"oidc_connect_userinfo_url"`
	OIDCConnectJWKSURL              string `json:"oidc_connect_jwks_url"`
	OIDCConnectScopes               string `json:"oidc_connect_scopes"`
	OIDCConnectRedirectURL          string `json:"oidc_connect_redirect_url"`
	OIDCConnectFrontendRedirectURL  string `json:"oidc_connect_frontend_redirect_url"`
	OIDCConnectTokenAuthMethod      string `json:"oidc_connect_token_auth_method"`
	OIDCConnectUsePKCE              bool   `json:"oidc_connect_use_pkce"`
	OIDCConnectValidateIDToken      bool   `json:"oidc_connect_validate_id_token"`
	OIDCConnectAllowedSigningAlgs   string `json:"oidc_connect_allowed_signing_algs"`
	OIDCConnectClockSkewSeconds     int    `json:"oidc_connect_clock_skew_seconds"`
	OIDCConnectRequireEmailVerified bool   `json:"oidc_connect_require_email_verified"`
	OIDCConnectUserInfoEmailPath    string `json:"oidc_connect_userinfo_email_path"`
	OIDCConnectUserInfoIDPath       string `json:"oidc_connect_userinfo_id_path"`
	OIDCConnectUserInfoUsernamePath string `json:"oidc_connect_userinfo_username_path"`

	// OEM设置
	SiteName                    string                `json:"site_name"`
	SiteLogo                    string                `json:"site_logo"`
	SiteSubtitle                string                `json:"site_subtitle"`
	FrontendTheme               string                `json:"frontend_theme"`
	APIBaseURL                  string                `json:"api_base_url"`
	ContactInfo                 string                `json:"contact_info"`
	DocURL                      string                `json:"doc_url"`
	HomeContent                 string                `json:"home_content"`
	HideCcsImportButton         bool                  `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled *bool                 `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL     *string               `json:"purchase_subscription_url"`
	CustomMenuItems             *[]dto.CustomMenuItem `json:"custom_menu_items"`
	CustomEndpoints             *[]dto.CustomEndpoint `json:"custom_endpoints"`

	// 默认配置
	DefaultConcurrency                       int                               `json:"default_concurrency"`
	DefaultBalance                           float64                           `json:"default_balance"`
	DefaultSubscriptions                     []dto.DefaultSubscriptionSetting  `json:"default_subscriptions"`
	AuthSourceDefaultEmailBalance            *float64                          `json:"auth_source_default_email_balance"`
	AuthSourceDefaultEmailConcurrency        *int                              `json:"auth_source_default_email_concurrency"`
	AuthSourceDefaultEmailSubscriptions      *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_email_subscriptions"`
	AuthSourceDefaultEmailGrantOnSignup      *bool                             `json:"auth_source_default_email_grant_on_signup"`
	AuthSourceDefaultEmailGrantOnFirstBind   *bool                             `json:"auth_source_default_email_grant_on_first_bind"`
	AuthSourceDefaultLinuxDoBalance          *float64                          `json:"auth_source_default_linuxdo_balance"`
	AuthSourceDefaultLinuxDoConcurrency      *int                              `json:"auth_source_default_linuxdo_concurrency"`
	AuthSourceDefaultLinuxDoSubscriptions    *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_linuxdo_subscriptions"`
	AuthSourceDefaultLinuxDoGrantOnSignup    *bool                             `json:"auth_source_default_linuxdo_grant_on_signup"`
	AuthSourceDefaultLinuxDoGrantOnFirstBind *bool                             `json:"auth_source_default_linuxdo_grant_on_first_bind"`
	AuthSourceDefaultOIDCBalance             *float64                          `json:"auth_source_default_oidc_balance"`
	AuthSourceDefaultOIDCConcurrency         *int                              `json:"auth_source_default_oidc_concurrency"`
	AuthSourceDefaultOIDCSubscriptions       *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_oidc_subscriptions"`
	AuthSourceDefaultOIDCGrantOnSignup       *bool                             `json:"auth_source_default_oidc_grant_on_signup"`
	AuthSourceDefaultOIDCGrantOnFirstBind    *bool                             `json:"auth_source_default_oidc_grant_on_first_bind"`
	AuthSourceDefaultWeChatBalance           *float64                          `json:"auth_source_default_wechat_balance"`
	AuthSourceDefaultWeChatConcurrency       *int                              `json:"auth_source_default_wechat_concurrency"`
	AuthSourceDefaultWeChatSubscriptions     *[]dto.DefaultSubscriptionSetting `json:"auth_source_default_wechat_subscriptions"`
	AuthSourceDefaultWeChatGrantOnSignup     *bool                             `json:"auth_source_default_wechat_grant_on_signup"`
	AuthSourceDefaultWeChatGrantOnFirstBind  *bool                             `json:"auth_source_default_wechat_grant_on_first_bind"`

	// Model fallback configuration
	EnableModelFallback                bool   `json:"enable_model_fallback"`
	FallbackModelAnthropic             string `json:"fallback_model_anthropic"`
	FallbackModelOpenAI                string `json:"fallback_model_openai"`
	FallbackModelGrok                  string `json:"fallback_model_grok"`
	FallbackModelGemini                string `json:"fallback_model_gemini"`
	FallbackModelAntigravity           string `json:"fallback_model_antigravity"`
	GrokOfficialBaseURL                string `json:"grok_official_base_url"`
	GrokSessionBaseURL                 string `json:"grok_session_base_url"`
	GrokThinkingSummary                *bool  `json:"grok_thinking_summary"`
	GrokShowSearchSources              *bool  `json:"grok_show_search_sources"`
	GrokImageOutputFormat              string `json:"grok_image_output_format"`
	GrokVideoOutputFormat              string `json:"grok_video_output_format"`
	GrokMediaProxyEnabled              *bool  `json:"grok_media_proxy_enabled"`
	GrokMediaCacheRetentionHours       *int   `json:"grok_media_cache_retention_hours"`
	GrokQuotaSyncIntervalSeconds       *int   `json:"grok_quota_sync_interval_seconds"`
	GrokUsageSyncConcurrency           *int   `json:"grok_usage_sync_concurrency"`
	GrokCapabilityProbeIntervalSeconds *int   `json:"grok_capability_probe_interval_seconds"`
	GrokCapabilityProbeConcurrency     *int   `json:"grok_capability_probe_concurrency"`
	GrokSessionValidityCheckInterval   *int   `json:"grok_session_validity_check_interval"`
	GrokVideoTimeout                   *int   `json:"grok_video_timeout"`

	// Identity patch configuration (Claude -> Gemini)
	EnableIdentityPatch bool   `json:"enable_identity_patch"`
	IdentityPatchPrompt string `json:"identity_patch_prompt"`

	// Ops monitoring (vNext)
	OpsMonitoringEnabled         *bool   `json:"ops_monitoring_enabled"`
	OpsRealtimeMonitoringEnabled *bool   `json:"ops_realtime_monitoring_enabled"`
	OpsQueryModeDefault          *string `json:"ops_query_mode_default"`
	OpsMetricsIntervalSeconds    *int    `json:"ops_metrics_interval_seconds"`

	MinClaudeCodeVersion string `json:"min_claude_code_version"`
	MaxClaudeCodeVersion string `json:"max_claude_code_version"`

	// 分组隔离
	AllowUngroupedKeyScheduling bool `json:"allow_ungrouped_key_scheduling"`

	// Backend Mode
	BackendModeEnabled bool `json:"backend_mode_enabled"`

	// Gateway forwarding behavior
	EnableFingerprintUnification *bool `json:"enable_fingerprint_unification"`
	EnableMetadataPassthrough    *bool `json:"enable_metadata_passthrough"`
	EnableCCHSigning             *bool `json:"enable_cch_signing"`
}

// UpdateSettings 更新系统设置
// PUT /api/v1/admin/settings
func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	previousSettings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	previousAuthSourceDefaults, err := h.settingService.GetAuthSourceDefaultSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 验证参数
	if req.DefaultConcurrency < 1 {
		req.DefaultConcurrency = 1
	}
	if req.DefaultBalance < 0 {
		req.DefaultBalance = 0
	}
	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.SMTPUsername = strings.TrimSpace(req.SMTPUsername)
	req.SMTPPassword = strings.TrimSpace(req.SMTPPassword)
	req.SMTPFrom = strings.TrimSpace(req.SMTPFrom)
	req.SMTPFromName = strings.TrimSpace(req.SMTPFromName)
	if req.SMTPPort <= 0 {
		req.SMTPPort = 587
	}
	req.DefaultSubscriptions = normalizeDefaultSubscriptions(req.DefaultSubscriptions)
	req.AuthSourceDefaultEmailSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultEmailSubscriptions)
	req.AuthSourceDefaultLinuxDoSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultLinuxDoSubscriptions)
	req.AuthSourceDefaultOIDCSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultOIDCSubscriptions)
	req.AuthSourceDefaultWeChatSubscriptions = normalizeOptionalDefaultSubscriptions(req.AuthSourceDefaultWeChatSubscriptions)

	// SMTP 配置保护：如果请求中 smtp_host 为空但数据库中已有配置，则保留已有 SMTP 配置
	// 防止前端加载设置失败时空表单覆盖已保存的 SMTP 配置
	if req.SMTPHost == "" && previousSettings.SMTPHost != "" {
		req.SMTPHost = previousSettings.SMTPHost
		req.SMTPPort = previousSettings.SMTPPort
		req.SMTPUsername = previousSettings.SMTPUsername
		req.SMTPFrom = previousSettings.SMTPFrom
		req.SMTPFromName = previousSettings.SMTPFromName
		req.SMTPUseTLS = previousSettings.SMTPUseTLS
	}

	// Turnstile 参数验证
	if req.TurnstileEnabled {
		// 检查必填字段
		if req.TurnstileSiteKey == "" {
			response.BadRequest(c, "Turnstile Site Key is required when enabled")
			return
		}
		// 如果未提供 secret key，使用已保存的值（留空保留当前值）
		if req.TurnstileSecretKey == "" {
			if previousSettings.TurnstileSecretKey == "" {
				response.BadRequest(c, "Turnstile Secret Key is required when enabled")
				return
			}
			req.TurnstileSecretKey = previousSettings.TurnstileSecretKey
		}

		// 当 site_key 或 secret_key 任一变化时验证（避免配置错误导致无法登录）
		siteKeyChanged := previousSettings.TurnstileSiteKey != req.TurnstileSiteKey
		secretKeyChanged := previousSettings.TurnstileSecretKey != req.TurnstileSecretKey
		if siteKeyChanged || secretKeyChanged {
			if err := h.turnstileService.ValidateSecretKey(c.Request.Context(), req.TurnstileSecretKey); err != nil {
				response.ErrorFrom(c, err)
				return
			}
		}
	}

	// TOTP 双因素认证参数验证
	// 只有手动配置了加密密钥才允许启用 TOTP 功能
	if req.TotpEnabled && !previousSettings.TotpEnabled {
		// 尝试启用 TOTP，检查加密密钥是否已手动配置
		if !h.settingService.IsTotpEncryptionKeyConfigured() {
			response.BadRequest(c, "Cannot enable TOTP: TOTP_ENCRYPTION_KEY environment variable must be configured first. Generate a key with 'openssl rand -hex 32' and set it in your environment.")
			return
		}
	}

	// LinuxDo Connect 参数验证
	if req.LinuxDoConnectEnabled {
		req.LinuxDoConnectClientID = strings.TrimSpace(req.LinuxDoConnectClientID)
		req.LinuxDoConnectClientSecret = strings.TrimSpace(req.LinuxDoConnectClientSecret)
		req.LinuxDoConnectRedirectURL = strings.TrimSpace(req.LinuxDoConnectRedirectURL)

		if req.LinuxDoConnectClientID == "" {
			response.BadRequest(c, "LinuxDo Client ID is required when enabled")
			return
		}
		if req.LinuxDoConnectRedirectURL == "" {
			response.BadRequest(c, "LinuxDo Redirect URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.LinuxDoConnectRedirectURL); err != nil {
			response.BadRequest(c, "LinuxDo Redirect URL must be an absolute http(s) URL")
			return
		}

		// 如果未提供 client_secret，则保留现有值（如有）。
		if req.LinuxDoConnectClientSecret == "" {
			if previousSettings.LinuxDoConnectClientSecret == "" {
				response.BadRequest(c, "LinuxDo Client Secret is required when enabled")
				return
			}
			req.LinuxDoConnectClientSecret = previousSettings.LinuxDoConnectClientSecret
		}
	}

	// WeChat Connect 参数验证
	if req.WeChatConnectEnabled {
		req.WeChatConnectOpenAppID = strings.TrimSpace(req.WeChatConnectOpenAppID)
		req.WeChatConnectOpenAppSecret = strings.TrimSpace(req.WeChatConnectOpenAppSecret)
		req.WeChatConnectMPAppID = strings.TrimSpace(req.WeChatConnectMPAppID)
		req.WeChatConnectMPAppSecret = strings.TrimSpace(req.WeChatConnectMPAppSecret)
		req.WeChatConnectMobileAppID = strings.TrimSpace(req.WeChatConnectMobileAppID)
		req.WeChatConnectMobileAppSecret = strings.TrimSpace(req.WeChatConnectMobileAppSecret)
		req.WeChatConnectMode = strings.ToLower(strings.TrimSpace(req.WeChatConnectMode))
		req.WeChatConnectScopes = strings.TrimSpace(req.WeChatConnectScopes)
		req.WeChatConnectRedirectURL = strings.TrimSpace(req.WeChatConnectRedirectURL)
		req.WeChatConnectFrontendRedirectURL = strings.TrimSpace(req.WeChatConnectFrontendRedirectURL)
		if req.WeChatConnectMode == "" {
			req.WeChatConnectMode = "open"
		}
		if !req.WeChatConnectOpenEnabled && !req.WeChatConnectMPEnabled && !req.WeChatConnectMobileEnabled {
			switch req.WeChatConnectMode {
			case "mp":
				req.WeChatConnectMPEnabled = true
			case "mobile":
				req.WeChatConnectMobileEnabled = true
			default:
				req.WeChatConnectOpenEnabled = true
			}
		}
		if req.WeChatConnectRedirectURL == "" {
			response.BadRequest(c, "WeChat Redirect URL is required when enabled")
			return
		}
		if req.WeChatConnectFrontendRedirectURL == "" {
			response.BadRequest(c, "WeChat Frontend Redirect URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.WeChatConnectRedirectURL); err != nil {
			response.BadRequest(c, "WeChat Redirect URL must be an absolute http(s) URL")
			return
		}
		if err := config.ValidateFrontendRedirectURL(req.WeChatConnectFrontendRedirectURL); err != nil {
			response.BadRequest(c, "WeChat Frontend Redirect URL is invalid")
			return
		}
		switch req.WeChatConnectMode {
		case "mp":
			if req.WeChatConnectMPAppID == "" {
				response.BadRequest(c, "WeChat MP App ID is required for mp mode")
				return
			}
			if req.WeChatConnectMPAppSecret == "" && previousSettings.WeChatConnectMPAppSecret == "" {
				response.BadRequest(c, "WeChat MP App Secret is required for mp mode")
				return
			}
			if req.WeChatConnectMPAppSecret == "" {
				req.WeChatConnectMPAppSecret = previousSettings.WeChatConnectMPAppSecret
			}
		case "mobile":
			if req.WeChatConnectMobileAppID == "" {
				response.BadRequest(c, "WeChat Mobile App ID is required for mobile mode")
				return
			}
			if req.WeChatConnectMobileAppSecret == "" && previousSettings.WeChatConnectMobileAppSecret == "" {
				response.BadRequest(c, "WeChat Mobile App Secret is required for mobile mode")
				return
			}
			if req.WeChatConnectMobileAppSecret == "" {
				req.WeChatConnectMobileAppSecret = previousSettings.WeChatConnectMobileAppSecret
			}
		default:
			if req.WeChatConnectOpenAppID == "" {
				response.BadRequest(c, "WeChat Open App ID is required for open mode")
				return
			}
			if req.WeChatConnectOpenAppSecret == "" && previousSettings.WeChatConnectOpenAppSecret == "" {
				response.BadRequest(c, "WeChat Open App Secret is required for open mode")
				return
			}
			if req.WeChatConnectOpenAppSecret == "" {
				req.WeChatConnectOpenAppSecret = previousSettings.WeChatConnectOpenAppSecret
			}
		}
	}

	// OIDC Connect 参数验证
	if req.OIDCConnectEnabled {
		req.OIDCConnectProviderName = strings.TrimSpace(req.OIDCConnectProviderName)
		req.OIDCConnectClientID = strings.TrimSpace(req.OIDCConnectClientID)
		req.OIDCConnectClientSecret = strings.TrimSpace(req.OIDCConnectClientSecret)
		req.OIDCConnectIssuerURL = strings.TrimSpace(req.OIDCConnectIssuerURL)
		req.OIDCConnectDiscoveryURL = strings.TrimSpace(req.OIDCConnectDiscoveryURL)
		req.OIDCConnectAuthorizeURL = strings.TrimSpace(req.OIDCConnectAuthorizeURL)
		req.OIDCConnectTokenURL = strings.TrimSpace(req.OIDCConnectTokenURL)
		req.OIDCConnectUserInfoURL = strings.TrimSpace(req.OIDCConnectUserInfoURL)
		req.OIDCConnectJWKSURL = strings.TrimSpace(req.OIDCConnectJWKSURL)
		req.OIDCConnectScopes = strings.TrimSpace(req.OIDCConnectScopes)
		req.OIDCConnectRedirectURL = strings.TrimSpace(req.OIDCConnectRedirectURL)
		req.OIDCConnectFrontendRedirectURL = strings.TrimSpace(req.OIDCConnectFrontendRedirectURL)
		req.OIDCConnectTokenAuthMethod = strings.ToLower(strings.TrimSpace(req.OIDCConnectTokenAuthMethod))
		req.OIDCConnectAllowedSigningAlgs = strings.TrimSpace(req.OIDCConnectAllowedSigningAlgs)
		req.OIDCConnectUserInfoEmailPath = strings.TrimSpace(req.OIDCConnectUserInfoEmailPath)
		req.OIDCConnectUserInfoIDPath = strings.TrimSpace(req.OIDCConnectUserInfoIDPath)
		req.OIDCConnectUserInfoUsernamePath = strings.TrimSpace(req.OIDCConnectUserInfoUsernamePath)

		if req.OIDCConnectProviderName == "" {
			req.OIDCConnectProviderName = "OIDC"
		}
		if req.OIDCConnectClientID == "" {
			response.BadRequest(c, "OIDC Client ID is required when enabled")
			return
		}
		if req.OIDCConnectAuthorizeURL == "" || req.OIDCConnectTokenURL == "" {
			response.BadRequest(c, "OIDC Authorize URL and Token URL are required when enabled")
			return
		}
		if req.OIDCConnectRedirectURL == "" {
			response.BadRequest(c, "OIDC Redirect URL is required when enabled")
			return
		}
		if req.OIDCConnectFrontendRedirectURL == "" {
			response.BadRequest(c, "OIDC Frontend Redirect URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectAuthorizeURL); err != nil {
			response.BadRequest(c, "OIDC Authorize URL must be an absolute http(s) URL")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectTokenURL); err != nil {
			response.BadRequest(c, "OIDC Token URL must be an absolute http(s) URL")
			return
		}
		if req.OIDCConnectUserInfoURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectUserInfoURL); err != nil {
				response.BadRequest(c, "OIDC UserInfo URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectJWKSURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectJWKSURL); err != nil {
				response.BadRequest(c, "OIDC JWKS URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectIssuerURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectIssuerURL); err != nil {
				response.BadRequest(c, "OIDC Issuer URL must be an absolute http(s) URL")
				return
			}
		}
		if req.OIDCConnectDiscoveryURL != "" {
			if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectDiscoveryURL); err != nil {
				response.BadRequest(c, "OIDC Discovery URL must be an absolute http(s) URL")
				return
			}
		}
		if err := config.ValidateAbsoluteHTTPURL(req.OIDCConnectRedirectURL); err != nil {
			response.BadRequest(c, "OIDC Redirect URL must be an absolute http(s) URL")
			return
		}
		if err := config.ValidateFrontendRedirectURL(req.OIDCConnectFrontendRedirectURL); err != nil {
			response.BadRequest(c, "OIDC Frontend Redirect URL is invalid")
			return
		}
		switch req.OIDCConnectTokenAuthMethod {
		case "", "client_secret_post", "client_secret_basic":
			if req.OIDCConnectClientSecret == "" {
				if previousSettings.OIDCConnectClientSecret == "" {
					response.BadRequest(c, "OIDC Client Secret is required when enabled")
					return
				}
				req.OIDCConnectClientSecret = previousSettings.OIDCConnectClientSecret
			}
		case "none":
			if !req.OIDCConnectUsePKCE {
				response.BadRequest(c, "OIDC PKCE must be enabled when token auth method is none")
				return
			}
		default:
			response.BadRequest(c, "OIDC token auth method is invalid")
			return
		}
		if req.OIDCConnectValidateIDToken && req.OIDCConnectJWKSURL == "" {
			response.BadRequest(c, "OIDC JWKS URL is required when ID token validation is enabled")
			return
		}
	}

	// “购买订阅”页面配置验证
	purchaseEnabled := previousSettings.PurchaseSubscriptionEnabled
	if req.PurchaseSubscriptionEnabled != nil {
		purchaseEnabled = *req.PurchaseSubscriptionEnabled
	}
	purchaseURL := previousSettings.PurchaseSubscriptionURL
	if req.PurchaseSubscriptionURL != nil {
		purchaseURL = strings.TrimSpace(*req.PurchaseSubscriptionURL)
	}

	// - 启用时要求 URL 合法且非空
	// - 禁用时允许为空；若提供了 URL 也做基本校验，避免误配置
	if purchaseEnabled {
		if purchaseURL == "" {
			response.BadRequest(c, "Purchase Subscription URL is required when enabled")
			return
		}
		if err := config.ValidateAbsoluteHTTPURL(purchaseURL); err != nil {
			response.BadRequest(c, "Purchase Subscription URL must be an absolute http(s) URL")
			return
		}
	} else if purchaseURL != "" {
		if err := config.ValidateAbsoluteHTTPURL(purchaseURL); err != nil {
			response.BadRequest(c, "Purchase Subscription URL must be an absolute http(s) URL")
			return
		}
	}

	// Frontend URL 验证
	req.FrontendURL = strings.TrimSpace(req.FrontendURL)
	req.FrontendTheme = strings.TrimSpace(req.FrontendTheme)
	if req.FrontendTheme == "" {
		req.FrontendTheme = previousSettings.FrontendTheme
		if req.FrontendTheme == "" {
			req.FrontendTheme = "factory"
		}
	}
	if req.FrontendURL != "" {
		if err := config.ValidateAbsoluteHTTPURL(req.FrontendURL); err != nil {
			response.BadRequest(c, "Frontend URL must be an absolute http(s) URL")
			return
		}
	}
	if value := strings.TrimSpace(req.GrokOfficialBaseURL); value != "" {
		if err := config.ValidateAbsoluteHTTPURL(value); err != nil {
			response.BadRequest(c, "Grok Official Base URL must be an absolute http(s) URL")
			return
		}
	}
	if value := strings.TrimSpace(req.GrokSessionBaseURL); value != "" {
		if err := config.ValidateAbsoluteHTTPURL(value); err != nil {
			response.BadRequest(c, "Grok Session Base URL must be an absolute http(s) URL")
			return
		}
	}

	// 自定义菜单项验证
	const (
		maxCustomMenuItems    = 20
		maxMenuItemLabelLen   = 50
		maxMenuItemURLLen     = 2048
		maxMenuItemIconSVGLen = 10 * 1024 // 10KB
		maxMenuItemIDLen      = 32
	)

	customMenuJSON := previousSettings.CustomMenuItems
	if req.CustomMenuItems != nil {
		items := *req.CustomMenuItems
		if len(items) > maxCustomMenuItems {
			response.BadRequest(c, "Too many custom menu items (max 20)")
			return
		}
		for i, item := range items {
			if strings.TrimSpace(item.Label) == "" {
				response.BadRequest(c, "Custom menu item label is required")
				return
			}
			if len(item.Label) > maxMenuItemLabelLen {
				response.BadRequest(c, "Custom menu item label is too long (max 50 characters)")
				return
			}
			if strings.TrimSpace(item.URL) == "" {
				response.BadRequest(c, "Custom menu item URL is required")
				return
			}
			if len(item.URL) > maxMenuItemURLLen {
				response.BadRequest(c, "Custom menu item URL is too long (max 2048 characters)")
				return
			}
			if err := config.ValidateAbsoluteHTTPURL(strings.TrimSpace(item.URL)); err != nil {
				response.BadRequest(c, "Custom menu item URL must be an absolute http(s) URL")
				return
			}
			if item.Visibility != "user" && item.Visibility != "admin" {
				response.BadRequest(c, "Custom menu item visibility must be 'user' or 'admin'")
				return
			}
			if len(item.IconSVG) > maxMenuItemIconSVGLen {
				response.BadRequest(c, "Custom menu item icon SVG is too large (max 10KB)")
				return
			}
			// Auto-generate ID if missing
			if strings.TrimSpace(item.ID) == "" {
				id, err := generateMenuItemID()
				if err != nil {
					response.Error(c, http.StatusInternalServerError, "Failed to generate menu item ID")
					return
				}
				items[i].ID = id
			} else if len(item.ID) > maxMenuItemIDLen {
				response.BadRequest(c, "Custom menu item ID is too long (max 32 characters)")
				return
			} else if !menuItemIDPattern.MatchString(item.ID) {
				response.BadRequest(c, "Custom menu item ID contains invalid characters (only a-z, A-Z, 0-9, - and _ are allowed)")
				return
			}
		}
		// ID uniqueness check
		seen := make(map[string]struct{}, len(items))
		for _, item := range items {
			if _, exists := seen[item.ID]; exists {
				response.BadRequest(c, "Duplicate custom menu item ID: "+item.ID)
				return
			}
			seen[item.ID] = struct{}{}
		}
		menuBytes, err := json.Marshal(items)
		if err != nil {
			response.BadRequest(c, "Failed to serialize custom menu items")
			return
		}
		customMenuJSON = string(menuBytes)
	}

	// 自定义端点验证
	const (
		maxCustomEndpoints        = 10
		maxEndpointNameLen        = 50
		maxEndpointURLLen         = 2048
		maxEndpointDescriptionLen = 200
	)

	customEndpointsJSON := previousSettings.CustomEndpoints
	if req.CustomEndpoints != nil {
		endpoints := *req.CustomEndpoints
		if len(endpoints) > maxCustomEndpoints {
			response.BadRequest(c, "Too many custom endpoints (max 10)")
			return
		}
		for _, ep := range endpoints {
			if strings.TrimSpace(ep.Name) == "" {
				response.BadRequest(c, "Custom endpoint name is required")
				return
			}
			if len(ep.Name) > maxEndpointNameLen {
				response.BadRequest(c, "Custom endpoint name is too long (max 50 characters)")
				return
			}
			if strings.TrimSpace(ep.Endpoint) == "" {
				response.BadRequest(c, "Custom endpoint URL is required")
				return
			}
			if len(ep.Endpoint) > maxEndpointURLLen {
				response.BadRequest(c, "Custom endpoint URL is too long (max 2048 characters)")
				return
			}
			if err := config.ValidateAbsoluteHTTPURL(strings.TrimSpace(ep.Endpoint)); err != nil {
				response.BadRequest(c, "Custom endpoint URL must be an absolute http(s) URL")
				return
			}
			if len(ep.Description) > maxEndpointDescriptionLen {
				response.BadRequest(c, "Custom endpoint description is too long (max 200 characters)")
				return
			}
		}
		endpointBytes, err := json.Marshal(endpoints)
		if err != nil {
			response.BadRequest(c, "Failed to serialize custom endpoints")
			return
		}
		customEndpointsJSON = string(endpointBytes)
	}

	// Ops metrics collector interval validation (seconds).
	if req.OpsMetricsIntervalSeconds != nil {
		v := *req.OpsMetricsIntervalSeconds
		if v < 60 {
			v = 60
		}
		if v > 3600 {
			v = 3600
		}
		req.OpsMetricsIntervalSeconds = &v
	}
	defaultSubscriptions := make([]service.DefaultSubscriptionSetting, 0, len(req.DefaultSubscriptions))
	for _, sub := range req.DefaultSubscriptions {
		defaultSubscriptions = append(defaultSubscriptions, service.DefaultSubscriptionSetting{
			GroupID:      sub.GroupID,
			ValidityDays: sub.ValidityDays,
		})
	}

	if req.AccountQuotaNotifyEmails != nil {
		for _, entry := range *req.AccountQuotaNotifyEmails {
			email := strings.TrimSpace(entry.Email)
			if email == "" {
				response.BadRequest(c, "Account quota notify email is required")
				return
			}
			if _, err := mail.ParseAddress(email); err != nil {
				response.BadRequest(c, "Account quota notify email must be a valid email address")
				return
			}
		}
	}

	// 验证最低版本号格式（空字符串=禁用，或合法 semver）
	if req.MinClaudeCodeVersion != "" {
		if !semverPattern.MatchString(req.MinClaudeCodeVersion) {
			response.Error(c, http.StatusBadRequest, "min_claude_code_version must be empty or a valid semver (e.g. 2.1.63)")
			return
		}
	}

	// 验证最高版本号格式（空字符串=禁用，或合法 semver）
	if req.MaxClaudeCodeVersion != "" {
		if !semverPattern.MatchString(req.MaxClaudeCodeVersion) {
			response.Error(c, http.StatusBadRequest, "max_claude_code_version must be empty or a valid semver (e.g. 3.0.0)")
			return
		}
	}

	// 交叉验证：如果同时设置了最低和最高版本号，最高版本号必须 >= 最低版本号
	if req.MinClaudeCodeVersion != "" && req.MaxClaudeCodeVersion != "" {
		if service.CompareVersions(req.MaxClaudeCodeVersion, req.MinClaudeCodeVersion) < 0 {
			response.Error(c, http.StatusBadRequest, "max_claude_code_version must be greater than or equal to min_claude_code_version")
			return
		}
	}

	settings := &service.SystemSettings{
		RegistrationEnabled:              req.RegistrationEnabled,
		EmailVerifyEnabled:               req.EmailVerifyEnabled,
		RegistrationEmailSuffixWhitelist: req.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 req.PromoCodeEnabled,
		PasswordResetEnabled:             req.PasswordResetEnabled,
		FrontendURL:                      req.FrontendURL,
		InvitationCodeEnabled:            req.InvitationCodeEnabled,
		TotpEnabled:                      req.TotpEnabled,
		AffiliateEnabled: func() bool {
			if req.AffiliateEnabled != nil {
				return *req.AffiliateEnabled
			}
			return previousSettings.AffiliateEnabled
		}(),
		AffiliateRebateRatePercent: func() float64 {
			if req.AffiliateRebateRatePercent != nil {
				return *req.AffiliateRebateRatePercent
			}
			return previousSettings.AffiliateRebateRatePercent
		}(),
		AffiliateRebateFreezeHours: func() int {
			if req.AffiliateRebateFreezeHours != nil {
				return *req.AffiliateRebateFreezeHours
			}
			return previousSettings.AffiliateRebateFreezeHours
		}(),
		AffiliateRebateDurationDays: func() int {
			if req.AffiliateRebateDurationDays != nil {
				return *req.AffiliateRebateDurationDays
			}
			return previousSettings.AffiliateRebateDurationDays
		}(),
		AffiliateRebatePerInviteeCap: func() float64 {
			if req.AffiliateRebatePerInviteeCap != nil {
				return *req.AffiliateRebatePerInviteeCap
			}
			return previousSettings.AffiliateRebatePerInviteeCap
		}(),
		SMTPHost:     req.SMTPHost,
		SMTPPort:     req.SMTPPort,
		SMTPUsername: req.SMTPUsername,
		SMTPPassword: req.SMTPPassword,
		SMTPFrom:     req.SMTPFrom,
		SMTPFromName: req.SMTPFromName,
		SMTPUseTLS:   req.SMTPUseTLS,
		BalanceLowNotifyEnabled: func() bool {
			if req.BalanceLowNotifyEnabled != nil {
				return *req.BalanceLowNotifyEnabled
			}
			return previousSettings.BalanceLowNotifyEnabled
		}(),
		BalanceLowNotifyThreshold: func() float64 {
			if req.BalanceLowNotifyThreshold != nil {
				return *req.BalanceLowNotifyThreshold
			}
			return previousSettings.BalanceLowNotifyThreshold
		}(),
		BalanceLowNotifyRechargeURL: func() string {
			if req.BalanceLowNotifyRechargeURL != nil {
				return strings.TrimSpace(*req.BalanceLowNotifyRechargeURL)
			}
			return previousSettings.BalanceLowNotifyRechargeURL
		}(),
		AccountQuotaNotifyEnabled: func() bool {
			if req.AccountQuotaNotifyEnabled != nil {
				return *req.AccountQuotaNotifyEnabled
			}
			return previousSettings.AccountQuotaNotifyEnabled
		}(),
		AccountQuotaNotifyEmails: func() []service.NotifyEmailEntry {
			if req.AccountQuotaNotifyEmails != nil {
				return dto.NotifyEmailEntriesToService(*req.AccountQuotaNotifyEmails)
			}
			return previousSettings.AccountQuotaNotifyEmails
		}(),
		TurnstileEnabled:                 req.TurnstileEnabled,
		TurnstileSiteKey:                 req.TurnstileSiteKey,
		TurnstileSecretKey:               req.TurnstileSecretKey,
		LinuxDoConnectEnabled:            req.LinuxDoConnectEnabled,
		LinuxDoConnectClientID:           req.LinuxDoConnectClientID,
		LinuxDoConnectClientSecret:       req.LinuxDoConnectClientSecret,
		LinuxDoConnectRedirectURL:        req.LinuxDoConnectRedirectURL,
		WeChatConnectEnabled:             req.WeChatConnectEnabled,
		WeChatConnectOpenAppID:           req.WeChatConnectOpenAppID,
		WeChatConnectOpenAppSecret:       req.WeChatConnectOpenAppSecret,
		WeChatConnectMPAppID:             req.WeChatConnectMPAppID,
		WeChatConnectMPAppSecret:         req.WeChatConnectMPAppSecret,
		WeChatConnectMobileAppID:         req.WeChatConnectMobileAppID,
		WeChatConnectMobileAppSecret:     req.WeChatConnectMobileAppSecret,
		WeChatConnectOpenEnabled:         req.WeChatConnectOpenEnabled,
		WeChatConnectMPEnabled:           req.WeChatConnectMPEnabled,
		WeChatConnectMobileEnabled:       req.WeChatConnectMobileEnabled,
		WeChatConnectMode:                req.WeChatConnectMode,
		WeChatConnectScopes:              req.WeChatConnectScopes,
		WeChatConnectRedirectURL:         req.WeChatConnectRedirectURL,
		WeChatConnectFrontendRedirectURL: req.WeChatConnectFrontendRedirectURL,
		OIDCConnectEnabled:               req.OIDCConnectEnabled,
		OIDCConnectProviderName:          req.OIDCConnectProviderName,
		OIDCConnectClientID:              req.OIDCConnectClientID,
		OIDCConnectClientSecret:          req.OIDCConnectClientSecret,
		OIDCConnectIssuerURL:             req.OIDCConnectIssuerURL,
		OIDCConnectDiscoveryURL:          req.OIDCConnectDiscoveryURL,
		OIDCConnectAuthorizeURL:          req.OIDCConnectAuthorizeURL,
		OIDCConnectTokenURL:              req.OIDCConnectTokenURL,
		OIDCConnectUserInfoURL:           req.OIDCConnectUserInfoURL,
		OIDCConnectJWKSURL:               req.OIDCConnectJWKSURL,
		OIDCConnectScopes:                req.OIDCConnectScopes,
		OIDCConnectRedirectURL:           req.OIDCConnectRedirectURL,
		OIDCConnectFrontendRedirectURL:   req.OIDCConnectFrontendRedirectURL,
		OIDCConnectTokenAuthMethod:       req.OIDCConnectTokenAuthMethod,
		OIDCConnectUsePKCE:               req.OIDCConnectUsePKCE,
		OIDCConnectValidateIDToken:       req.OIDCConnectValidateIDToken,
		OIDCConnectAllowedSigningAlgs:    req.OIDCConnectAllowedSigningAlgs,
		OIDCConnectClockSkewSeconds:      req.OIDCConnectClockSkewSeconds,
		OIDCConnectRequireEmailVerified:  req.OIDCConnectRequireEmailVerified,
		OIDCConnectUserInfoEmailPath:     req.OIDCConnectUserInfoEmailPath,
		OIDCConnectUserInfoIDPath:        req.OIDCConnectUserInfoIDPath,
		OIDCConnectUserInfoUsernamePath:  req.OIDCConnectUserInfoUsernamePath,
		SiteName:                         req.SiteName,
		SiteLogo:                         req.SiteLogo,
		SiteSubtitle:                     req.SiteSubtitle,
		FrontendTheme:                    req.FrontendTheme,
		APIBaseURL:                       req.APIBaseURL,
		ContactInfo:                      req.ContactInfo,
		DocURL:                           req.DocURL,
		HomeContent:                      req.HomeContent,
		HideCcsImportButton:              req.HideCcsImportButton,
		PurchaseSubscriptionEnabled:      purchaseEnabled,
		PurchaseSubscriptionURL:          purchaseURL,
		CustomMenuItems:                  customMenuJSON,
		CustomEndpoints:                  customEndpointsJSON,
		DefaultConcurrency:               req.DefaultConcurrency,
		DefaultBalance:                   req.DefaultBalance,
		DefaultSubscriptions:             defaultSubscriptions,
		EnableModelFallback:              req.EnableModelFallback,
		FallbackModelAnthropic:           req.FallbackModelAnthropic,
		FallbackModelOpenAI:              req.FallbackModelOpenAI,
		FallbackModelGrok:                req.FallbackModelGrok,
		FallbackModelGemini:              req.FallbackModelGemini,
		FallbackModelAntigravity:         req.FallbackModelAntigravity,
		GrokOfficialBaseURL: func() string {
			if value := strings.TrimSpace(req.GrokOfficialBaseURL); value != "" {
				return value
			}
			return previousSettings.GrokOfficialBaseURL
		}(),
		GrokSessionBaseURL: func() string {
			if value := strings.TrimSpace(req.GrokSessionBaseURL); value != "" {
				return value
			}
			return previousSettings.GrokSessionBaseURL
		}(),
		GrokThinkingSummary: func() bool {
			if req.GrokThinkingSummary != nil {
				return *req.GrokThinkingSummary
			}
			return previousSettings.GrokThinkingSummary
		}(),
		GrokShowSearchSources: func() bool {
			if req.GrokShowSearchSources != nil {
				return *req.GrokShowSearchSources
			}
			return previousSettings.GrokShowSearchSources
		}(),
		GrokImageOutputFormat: req.GrokImageOutputFormat,
		GrokVideoOutputFormat: req.GrokVideoOutputFormat,
		GrokMediaProxyEnabled: func() bool {
			if req.GrokMediaProxyEnabled != nil {
				return *req.GrokMediaProxyEnabled
			}
			return previousSettings.GrokMediaProxyEnabled
		}(),
		GrokMediaCacheRetentionHours: func() int {
			if req.GrokMediaCacheRetentionHours != nil {
				return *req.GrokMediaCacheRetentionHours
			}
			return previousSettings.GrokMediaCacheRetentionHours
		}(),
		GrokQuotaSyncIntervalSeconds: func() int {
			if req.GrokQuotaSyncIntervalSeconds != nil {
				return *req.GrokQuotaSyncIntervalSeconds
			}
			return previousSettings.GrokQuotaSyncIntervalSeconds
		}(),
		GrokUsageSyncConcurrency: func() int {
			if req.GrokUsageSyncConcurrency != nil {
				return *req.GrokUsageSyncConcurrency
			}
			return previousSettings.GrokUsageSyncConcurrency
		}(),
		GrokCapabilityProbeIntervalSeconds: func() int {
			if req.GrokCapabilityProbeIntervalSeconds != nil {
				return *req.GrokCapabilityProbeIntervalSeconds
			}
			return previousSettings.GrokCapabilityProbeIntervalSeconds
		}(),
		GrokCapabilityProbeConcurrency: func() int {
			if req.GrokCapabilityProbeConcurrency != nil {
				return *req.GrokCapabilityProbeConcurrency
			}
			return previousSettings.GrokCapabilityProbeConcurrency
		}(),
		GrokSessionValidityCheckIntervalSeconds: func() int {
			if req.GrokSessionValidityCheckInterval != nil {
				return *req.GrokSessionValidityCheckInterval
			}
			return previousSettings.GrokSessionValidityCheckIntervalSeconds
		}(),
		GrokVideoTimeoutSeconds: func() int {
			if req.GrokVideoTimeout != nil {
				return *req.GrokVideoTimeout
			}
			return previousSettings.GrokVideoTimeoutSeconds
		}(),
		EnableIdentityPatch:         req.EnableIdentityPatch,
		IdentityPatchPrompt:         req.IdentityPatchPrompt,
		MinClaudeCodeVersion:        req.MinClaudeCodeVersion,
		MaxClaudeCodeVersion:        req.MaxClaudeCodeVersion,
		AllowUngroupedKeyScheduling: req.AllowUngroupedKeyScheduling,
		BackendModeEnabled:          req.BackendModeEnabled,
		OpsMonitoringEnabled: func() bool {
			if req.OpsMonitoringEnabled != nil {
				return *req.OpsMonitoringEnabled
			}
			return previousSettings.OpsMonitoringEnabled
		}(),
		OpsRealtimeMonitoringEnabled: func() bool {
			if req.OpsRealtimeMonitoringEnabled != nil {
				return *req.OpsRealtimeMonitoringEnabled
			}
			return previousSettings.OpsRealtimeMonitoringEnabled
		}(),
		OpsQueryModeDefault: func() string {
			if req.OpsQueryModeDefault != nil {
				return *req.OpsQueryModeDefault
			}
			return previousSettings.OpsQueryModeDefault
		}(),
		OpsMetricsIntervalSeconds: func() int {
			if req.OpsMetricsIntervalSeconds != nil {
				return *req.OpsMetricsIntervalSeconds
			}
			return previousSettings.OpsMetricsIntervalSeconds
		}(),
		EnableFingerprintUnification: func() bool {
			if req.EnableFingerprintUnification != nil {
				return *req.EnableFingerprintUnification
			}
			return previousSettings.EnableFingerprintUnification
		}(),
		EnableMetadataPassthrough: func() bool {
			if req.EnableMetadataPassthrough != nil {
				return *req.EnableMetadataPassthrough
			}
			return previousSettings.EnableMetadataPassthrough
		}(),
		EnableCCHSigning: func() bool {
			if req.EnableCCHSigning != nil {
				return *req.EnableCCHSigning
			}
			return previousSettings.EnableCCHSigning
		}(),
	}

	if err := h.settingService.UpdateSettings(c.Request.Context(), settings); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	updatedAuthSourceDefaults := previousAuthSourceDefaults
	if updatedAuthSourceDefaults == nil {
		updatedAuthSourceDefaults = &service.AuthSourceDefaultSettings{}
	}
	updatedAuthSourceDefaults = &service.AuthSourceDefaultSettings{
		Email: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultEmailBalance, updatedAuthSourceDefaults.Email.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultEmailConcurrency, updatedAuthSourceDefaults.Email.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultEmailSubscriptions, updatedAuthSourceDefaults.Email.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultEmailGrantOnSignup, updatedAuthSourceDefaults.Email.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultEmailGrantOnFirstBind, updatedAuthSourceDefaults.Email.GrantOnFirstBind),
		},
		LinuxDo: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultLinuxDoBalance, updatedAuthSourceDefaults.LinuxDo.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultLinuxDoConcurrency, updatedAuthSourceDefaults.LinuxDo.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultLinuxDoSubscriptions, updatedAuthSourceDefaults.LinuxDo.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultLinuxDoGrantOnSignup, updatedAuthSourceDefaults.LinuxDo.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultLinuxDoGrantOnFirstBind, updatedAuthSourceDefaults.LinuxDo.GrantOnFirstBind),
		},
		OIDC: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultOIDCBalance, updatedAuthSourceDefaults.OIDC.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultOIDCConcurrency, updatedAuthSourceDefaults.OIDC.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultOIDCSubscriptions, updatedAuthSourceDefaults.OIDC.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultOIDCGrantOnSignup, updatedAuthSourceDefaults.OIDC.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultOIDCGrantOnFirstBind, updatedAuthSourceDefaults.OIDC.GrantOnFirstBind),
		},
		WeChat: service.ProviderDefaultGrantSettings{
			Balance:          float64ValueOrDefault(req.AuthSourceDefaultWeChatBalance, updatedAuthSourceDefaults.WeChat.Balance),
			Concurrency:      intValueOrDefault(req.AuthSourceDefaultWeChatConcurrency, updatedAuthSourceDefaults.WeChat.Concurrency),
			Subscriptions:    defaultSubscriptionsValueOrDefault(req.AuthSourceDefaultWeChatSubscriptions, updatedAuthSourceDefaults.WeChat.Subscriptions),
			GrantOnSignup:    boolValueOrDefault(req.AuthSourceDefaultWeChatGrantOnSignup, updatedAuthSourceDefaults.WeChat.GrantOnSignup),
			GrantOnFirstBind: boolValueOrDefault(req.AuthSourceDefaultWeChatGrantOnFirstBind, updatedAuthSourceDefaults.WeChat.GrantOnFirstBind),
		},
		ForceEmailOnThirdPartySignup: boolValueOrDefault(req.ForceEmailOnThirdPartySignup, updatedAuthSourceDefaults.ForceEmailOnThirdPartySignup),
	}
	if err := h.settingService.UpdateAuthSourceDefaultSettings(c.Request.Context(), updatedAuthSourceDefaults); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.auditSettingsUpdate(c, previousSettings, settings, req)

	// 重新获取设置返回
	updatedSettings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	updatedAuthSourceDefaults, err = h.settingService.GetAuthSourceDefaultSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	updatedDefaultSubscriptions := make([]dto.DefaultSubscriptionSetting, 0, len(updatedSettings.DefaultSubscriptions))
	for _, sub := range updatedSettings.DefaultSubscriptions {
		updatedDefaultSubscriptions = append(updatedDefaultSubscriptions, dto.DefaultSubscriptionSetting{
			GroupID:      sub.GroupID,
			ValidityDays: sub.ValidityDays,
		})
	}

	response.Success(c, dto.SystemSettings{
		RegistrationEnabled:                      updatedSettings.RegistrationEnabled,
		EmailVerifyEnabled:                       updatedSettings.EmailVerifyEnabled,
		ForceEmailOnThirdPartySignup:             updatedAuthSourceDefaults != nil && updatedAuthSourceDefaults.ForceEmailOnThirdPartySignup,
		RegistrationEmailSuffixWhitelist:         updatedSettings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                         updatedSettings.PromoCodeEnabled,
		PasswordResetEnabled:                     updatedSettings.PasswordResetEnabled,
		FrontendURL:                              updatedSettings.FrontendURL,
		InvitationCodeEnabled:                    updatedSettings.InvitationCodeEnabled,
		TotpEnabled:                              updatedSettings.TotpEnabled,
		TotpEncryptionKeyConfigured:              h.settingService.IsTotpEncryptionKeyConfigured(),
		AffiliateEnabled:                         updatedSettings.AffiliateEnabled,
		AffiliateRebateRatePercent:               updatedSettings.AffiliateRebateRatePercent,
		AffiliateRebateFreezeHours:               updatedSettings.AffiliateRebateFreezeHours,
		AffiliateRebateDurationDays:              updatedSettings.AffiliateRebateDurationDays,
		AffiliateRebatePerInviteeCap:             updatedSettings.AffiliateRebatePerInviteeCap,
		SMTPHost:                                 updatedSettings.SMTPHost,
		SMTPPort:                                 updatedSettings.SMTPPort,
		SMTPUsername:                             updatedSettings.SMTPUsername,
		SMTPPasswordConfigured:                   updatedSettings.SMTPPasswordConfigured,
		SMTPFrom:                                 updatedSettings.SMTPFrom,
		SMTPFromName:                             updatedSettings.SMTPFromName,
		SMTPUseTLS:                               updatedSettings.SMTPUseTLS,
		BalanceLowNotifyEnabled:                  updatedSettings.BalanceLowNotifyEnabled,
		BalanceLowNotifyThreshold:                updatedSettings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:              updatedSettings.BalanceLowNotifyRechargeURL,
		AccountQuotaNotifyEnabled:                updatedSettings.AccountQuotaNotifyEnabled,
		AccountQuotaNotifyEmails:                 dto.NotifyEmailEntriesFromService(updatedSettings.AccountQuotaNotifyEmails),
		TurnstileEnabled:                         updatedSettings.TurnstileEnabled,
		TurnstileSiteKey:                         updatedSettings.TurnstileSiteKey,
		TurnstileSecretKeyConfigured:             updatedSettings.TurnstileSecretKeyConfigured,
		LinuxDoConnectEnabled:                    updatedSettings.LinuxDoConnectEnabled,
		LinuxDoConnectClientID:                   updatedSettings.LinuxDoConnectClientID,
		LinuxDoConnectClientSecretConfigured:     updatedSettings.LinuxDoConnectClientSecretConfigured,
		LinuxDoConnectRedirectURL:                updatedSettings.LinuxDoConnectRedirectURL,
		WeChatConnectEnabled:                     updatedSettings.WeChatConnectEnabled,
		WeChatConnectOpenAppID:                   updatedSettings.WeChatConnectOpenAppID,
		WeChatConnectOpenAppSecretConfigured:     updatedSettings.WeChatConnectOpenAppSecretConfigured,
		WeChatConnectMPAppID:                     updatedSettings.WeChatConnectMPAppID,
		WeChatConnectMPAppSecretConfigured:       updatedSettings.WeChatConnectMPAppSecretConfigured,
		WeChatConnectMobileAppID:                 updatedSettings.WeChatConnectMobileAppID,
		WeChatConnectMobileAppSecretConfigured:   updatedSettings.WeChatConnectMobileAppSecretConfigured,
		WeChatConnectOpenEnabled:                 updatedSettings.WeChatConnectOpenEnabled,
		WeChatConnectMPEnabled:                   updatedSettings.WeChatConnectMPEnabled,
		WeChatConnectMobileEnabled:               updatedSettings.WeChatConnectMobileEnabled,
		WeChatConnectMode:                        updatedSettings.WeChatConnectMode,
		WeChatConnectScopes:                      updatedSettings.WeChatConnectScopes,
		WeChatConnectRedirectURL:                 updatedSettings.WeChatConnectRedirectURL,
		WeChatConnectFrontendRedirectURL:         updatedSettings.WeChatConnectFrontendRedirectURL,
		OIDCConnectEnabled:                       updatedSettings.OIDCConnectEnabled,
		OIDCConnectProviderName:                  updatedSettings.OIDCConnectProviderName,
		OIDCConnectClientID:                      updatedSettings.OIDCConnectClientID,
		OIDCConnectClientSecretConfigured:        updatedSettings.OIDCConnectClientSecretConfigured,
		OIDCConnectIssuerURL:                     updatedSettings.OIDCConnectIssuerURL,
		OIDCConnectDiscoveryURL:                  updatedSettings.OIDCConnectDiscoveryURL,
		OIDCConnectAuthorizeURL:                  updatedSettings.OIDCConnectAuthorizeURL,
		OIDCConnectTokenURL:                      updatedSettings.OIDCConnectTokenURL,
		OIDCConnectUserInfoURL:                   updatedSettings.OIDCConnectUserInfoURL,
		OIDCConnectJWKSURL:                       updatedSettings.OIDCConnectJWKSURL,
		OIDCConnectScopes:                        updatedSettings.OIDCConnectScopes,
		OIDCConnectRedirectURL:                   updatedSettings.OIDCConnectRedirectURL,
		OIDCConnectFrontendRedirectURL:           updatedSettings.OIDCConnectFrontendRedirectURL,
		OIDCConnectTokenAuthMethod:               updatedSettings.OIDCConnectTokenAuthMethod,
		OIDCConnectUsePKCE:                       updatedSettings.OIDCConnectUsePKCE,
		OIDCConnectValidateIDToken:               updatedSettings.OIDCConnectValidateIDToken,
		OIDCConnectAllowedSigningAlgs:            updatedSettings.OIDCConnectAllowedSigningAlgs,
		OIDCConnectClockSkewSeconds:              updatedSettings.OIDCConnectClockSkewSeconds,
		OIDCConnectRequireEmailVerified:          updatedSettings.OIDCConnectRequireEmailVerified,
		OIDCConnectUserInfoEmailPath:             updatedSettings.OIDCConnectUserInfoEmailPath,
		OIDCConnectUserInfoIDPath:                updatedSettings.OIDCConnectUserInfoIDPath,
		OIDCConnectUserInfoUsernamePath:          updatedSettings.OIDCConnectUserInfoUsernamePath,
		SiteName:                                 updatedSettings.SiteName,
		SiteLogo:                                 updatedSettings.SiteLogo,
		SiteSubtitle:                             updatedSettings.SiteSubtitle,
		FrontendTheme:                            updatedSettings.FrontendTheme,
		APIBaseURL:                               updatedSettings.APIBaseURL,
		ContactInfo:                              updatedSettings.ContactInfo,
		DocURL:                                   updatedSettings.DocURL,
		HomeContent:                              updatedSettings.HomeContent,
		HideCcsImportButton:                      updatedSettings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:              updatedSettings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:                  updatedSettings.PurchaseSubscriptionURL,
		CustomMenuItems:                          dto.ParseCustomMenuItems(updatedSettings.CustomMenuItems),
		CustomEndpoints:                          dto.ParseCustomEndpoints(updatedSettings.CustomEndpoints),
		DefaultConcurrency:                       updatedSettings.DefaultConcurrency,
		DefaultBalance:                           updatedSettings.DefaultBalance,
		DefaultSubscriptions:                     updatedDefaultSubscriptions,
		AuthSourceDefaultEmailBalance:            updatedAuthSourceDefaults.Email.Balance,
		AuthSourceDefaultEmailConcurrency:        updatedAuthSourceDefaults.Email.Concurrency,
		AuthSourceDefaultEmailSubscriptions:      dto.DefaultSubscriptionSettingsFromService(updatedAuthSourceDefaults.Email.Subscriptions),
		AuthSourceDefaultEmailGrantOnSignup:      updatedAuthSourceDefaults.Email.GrantOnSignup,
		AuthSourceDefaultEmailGrantOnFirstBind:   updatedAuthSourceDefaults.Email.GrantOnFirstBind,
		AuthSourceDefaultLinuxDoBalance:          updatedAuthSourceDefaults.LinuxDo.Balance,
		AuthSourceDefaultLinuxDoConcurrency:      updatedAuthSourceDefaults.LinuxDo.Concurrency,
		AuthSourceDefaultLinuxDoSubscriptions:    dto.DefaultSubscriptionSettingsFromService(updatedAuthSourceDefaults.LinuxDo.Subscriptions),
		AuthSourceDefaultLinuxDoGrantOnSignup:    updatedAuthSourceDefaults.LinuxDo.GrantOnSignup,
		AuthSourceDefaultLinuxDoGrantOnFirstBind: updatedAuthSourceDefaults.LinuxDo.GrantOnFirstBind,
		AuthSourceDefaultOIDCBalance:             updatedAuthSourceDefaults.OIDC.Balance,
		AuthSourceDefaultOIDCConcurrency:         updatedAuthSourceDefaults.OIDC.Concurrency,
		AuthSourceDefaultOIDCSubscriptions:       dto.DefaultSubscriptionSettingsFromService(updatedAuthSourceDefaults.OIDC.Subscriptions),
		AuthSourceDefaultOIDCGrantOnSignup:       updatedAuthSourceDefaults.OIDC.GrantOnSignup,
		AuthSourceDefaultOIDCGrantOnFirstBind:    updatedAuthSourceDefaults.OIDC.GrantOnFirstBind,
		AuthSourceDefaultWeChatBalance:           updatedAuthSourceDefaults.WeChat.Balance,
		AuthSourceDefaultWeChatConcurrency:       updatedAuthSourceDefaults.WeChat.Concurrency,
		AuthSourceDefaultWeChatSubscriptions:     dto.DefaultSubscriptionSettingsFromService(updatedAuthSourceDefaults.WeChat.Subscriptions),
		AuthSourceDefaultWeChatGrantOnSignup:     updatedAuthSourceDefaults.WeChat.GrantOnSignup,
		AuthSourceDefaultWeChatGrantOnFirstBind:  updatedAuthSourceDefaults.WeChat.GrantOnFirstBind,
		EnableModelFallback:                      updatedSettings.EnableModelFallback,
		FallbackModelAnthropic:                   updatedSettings.FallbackModelAnthropic,
		FallbackModelOpenAI:                      updatedSettings.FallbackModelOpenAI,
		FallbackModelGrok:                        updatedSettings.FallbackModelGrok,
		FallbackModelGemini:                      updatedSettings.FallbackModelGemini,
		FallbackModelAntigravity:                 updatedSettings.FallbackModelAntigravity,
		GrokOfficialBaseURL:                      updatedSettings.GrokOfficialBaseURL,
		GrokSessionBaseURL:                       updatedSettings.GrokSessionBaseURL,
		GrokThinkingSummary:                      updatedSettings.GrokThinkingSummary,
		GrokShowSearchSources:                    updatedSettings.GrokShowSearchSources,
		GrokImageOutputFormat:                    updatedSettings.GrokImageOutputFormat,
		GrokVideoOutputFormat:                    updatedSettings.GrokVideoOutputFormat,
		GrokMediaProxyEnabled:                    updatedSettings.GrokMediaProxyEnabled,
		GrokMediaCacheRetentionHours:             updatedSettings.GrokMediaCacheRetentionHours,
		GrokQuotaSyncIntervalSeconds:             updatedSettings.GrokQuotaSyncIntervalSeconds,
		GrokUsageSyncConcurrency:                 updatedSettings.GrokUsageSyncConcurrency,
		GrokCapabilityProbeIntervalSeconds:       updatedSettings.GrokCapabilityProbeIntervalSeconds,
		GrokCapabilityProbeConcurrency:           updatedSettings.GrokCapabilityProbeConcurrency,
		GrokSessionValidityCheckIntervalSeconds:  updatedSettings.GrokSessionValidityCheckIntervalSeconds,
		GrokVideoTimeoutSeconds:                  updatedSettings.GrokVideoTimeoutSeconds,
		EnableIdentityPatch:                      updatedSettings.EnableIdentityPatch,
		IdentityPatchPrompt:                      updatedSettings.IdentityPatchPrompt,
		OpsMonitoringEnabled:                     updatedSettings.OpsMonitoringEnabled,
		OpsRealtimeMonitoringEnabled:             updatedSettings.OpsRealtimeMonitoringEnabled,
		OpsQueryModeDefault:                      updatedSettings.OpsQueryModeDefault,
		OpsMetricsIntervalSeconds:                updatedSettings.OpsMetricsIntervalSeconds,
		MinClaudeCodeVersion:                     updatedSettings.MinClaudeCodeVersion,
		MaxClaudeCodeVersion:                     updatedSettings.MaxClaudeCodeVersion,
		AllowUngroupedKeyScheduling:              updatedSettings.AllowUngroupedKeyScheduling,
		BackendModeEnabled:                       updatedSettings.BackendModeEnabled,
		EnableFingerprintUnification:             updatedSettings.EnableFingerprintUnification,
		EnableMetadataPassthrough:                updatedSettings.EnableMetadataPassthrough,
		EnableCCHSigning:                         updatedSettings.EnableCCHSigning,
	})
}

func (h *SettingHandler) auditSettingsUpdate(c *gin.Context, before *service.SystemSettings, after *service.SystemSettings, req UpdateSettingsRequest) {
	if before == nil || after == nil {
		return
	}

	changed := diffSettings(before, after, req)
	if len(changed) == 0 {
		return
	}

	subject, _ := middleware.GetAuthSubjectFromContext(c)
	role, _ := middleware.GetUserRoleFromContext(c)
	log.Printf("AUDIT: settings updated at=%s user_id=%d role=%s changed=%v",
		time.Now().UTC().Format(time.RFC3339),
		subject.UserID,
		role,
		changed,
	)
}

func diffSettings(before *service.SystemSettings, after *service.SystemSettings, req UpdateSettingsRequest) []string {
	changed := make([]string, 0, 20)
	if before.RegistrationEnabled != after.RegistrationEnabled {
		changed = append(changed, "registration_enabled")
	}
	if before.EmailVerifyEnabled != after.EmailVerifyEnabled {
		changed = append(changed, "email_verify_enabled")
	}
	if !equalStringSlice(before.RegistrationEmailSuffixWhitelist, after.RegistrationEmailSuffixWhitelist) {
		changed = append(changed, "registration_email_suffix_whitelist")
	}
	if before.PasswordResetEnabled != after.PasswordResetEnabled {
		changed = append(changed, "password_reset_enabled")
	}
	if before.FrontendURL != after.FrontendURL {
		changed = append(changed, "frontend_url")
	}
	if before.FallbackModelGrok != after.FallbackModelGrok {
		changed = append(changed, "fallback_model_grok")
	}
	if before.GrokOfficialBaseURL != after.GrokOfficialBaseURL {
		changed = append(changed, "grok_official_base_url")
	}
	if before.GrokSessionBaseURL != after.GrokSessionBaseURL {
		changed = append(changed, "grok_session_base_url")
	}
	if before.GrokThinkingSummary != after.GrokThinkingSummary {
		changed = append(changed, "grok_thinking_summary")
	}
	if before.GrokShowSearchSources != after.GrokShowSearchSources {
		changed = append(changed, "grok_show_search_sources")
	}
	if before.GrokImageOutputFormat != after.GrokImageOutputFormat {
		changed = append(changed, "grok_image_output_format")
	}
	if before.GrokVideoOutputFormat != after.GrokVideoOutputFormat {
		changed = append(changed, "grok_video_output_format")
	}
	if before.GrokMediaProxyEnabled != after.GrokMediaProxyEnabled {
		changed = append(changed, "grok_media_proxy_enabled")
	}
	if before.GrokMediaCacheRetentionHours != after.GrokMediaCacheRetentionHours {
		changed = append(changed, "grok_media_cache_retention_hours")
	}
	if before.GrokQuotaSyncIntervalSeconds != after.GrokQuotaSyncIntervalSeconds {
		changed = append(changed, "grok_quota_sync_interval_seconds")
	}
	if before.GrokUsageSyncConcurrency != after.GrokUsageSyncConcurrency {
		changed = append(changed, "grok_usage_sync_concurrency")
	}
	if before.GrokCapabilityProbeIntervalSeconds != after.GrokCapabilityProbeIntervalSeconds {
		changed = append(changed, "grok_capability_probe_interval_seconds")
	}
	if before.GrokCapabilityProbeConcurrency != after.GrokCapabilityProbeConcurrency {
		changed = append(changed, "grok_capability_probe_concurrency")
	}
	if before.GrokSessionValidityCheckIntervalSeconds != after.GrokSessionValidityCheckIntervalSeconds {
		changed = append(changed, "grok_session_validity_check_interval")
	}
	if before.GrokVideoTimeoutSeconds != after.GrokVideoTimeoutSeconds {
		changed = append(changed, "grok_video_timeout")
	}
	if before.TotpEnabled != after.TotpEnabled {
		changed = append(changed, "totp_enabled")
	}
	if before.AffiliateEnabled != after.AffiliateEnabled {
		changed = append(changed, "affiliate_enabled")
	}
	if before.AffiliateRebateRatePercent != after.AffiliateRebateRatePercent {
		changed = append(changed, "affiliate_rebate_rate")
	}
	if before.AffiliateRebateFreezeHours != after.AffiliateRebateFreezeHours {
		changed = append(changed, "affiliate_rebate_freeze_hours")
	}
	if before.AffiliateRebateDurationDays != after.AffiliateRebateDurationDays {
		changed = append(changed, "affiliate_rebate_duration_days")
	}
	if before.AffiliateRebatePerInviteeCap != after.AffiliateRebatePerInviteeCap {
		changed = append(changed, "affiliate_rebate_per_invitee_cap")
	}
	if before.SMTPHost != after.SMTPHost {
		changed = append(changed, "smtp_host")
	}
	if before.SMTPPort != after.SMTPPort {
		changed = append(changed, "smtp_port")
	}
	if before.SMTPUsername != after.SMTPUsername {
		changed = append(changed, "smtp_username")
	}
	if req.SMTPPassword != "" {
		changed = append(changed, "smtp_password")
	}
	if before.SMTPFrom != after.SMTPFrom {
		changed = append(changed, "smtp_from_email")
	}
	if before.SMTPFromName != after.SMTPFromName {
		changed = append(changed, "smtp_from_name")
	}
	if before.SMTPUseTLS != after.SMTPUseTLS {
		changed = append(changed, "smtp_use_tls")
	}
	if before.TurnstileEnabled != after.TurnstileEnabled {
		changed = append(changed, "turnstile_enabled")
	}
	if before.TurnstileSiteKey != after.TurnstileSiteKey {
		changed = append(changed, "turnstile_site_key")
	}
	if req.TurnstileSecretKey != "" {
		changed = append(changed, "turnstile_secret_key")
	}
	if before.LinuxDoConnectEnabled != after.LinuxDoConnectEnabled {
		changed = append(changed, "linuxdo_connect_enabled")
	}
	if before.LinuxDoConnectClientID != after.LinuxDoConnectClientID {
		changed = append(changed, "linuxdo_connect_client_id")
	}
	if req.LinuxDoConnectClientSecret != "" {
		changed = append(changed, "linuxdo_connect_client_secret")
	}
	if before.LinuxDoConnectRedirectURL != after.LinuxDoConnectRedirectURL {
		changed = append(changed, "linuxdo_connect_redirect_url")
	}
	if before.WeChatConnectEnabled != after.WeChatConnectEnabled {
		changed = append(changed, "wechat_connect_enabled")
	}
	if before.WeChatConnectOpenAppID != after.WeChatConnectOpenAppID {
		changed = append(changed, "wechat_connect_open_app_id")
	}
	if req.WeChatConnectOpenAppSecret != "" {
		changed = append(changed, "wechat_connect_open_app_secret")
	}
	if before.WeChatConnectMPAppID != after.WeChatConnectMPAppID {
		changed = append(changed, "wechat_connect_mp_app_id")
	}
	if req.WeChatConnectMPAppSecret != "" {
		changed = append(changed, "wechat_connect_mp_app_secret")
	}
	if before.WeChatConnectMobileAppID != after.WeChatConnectMobileAppID {
		changed = append(changed, "wechat_connect_mobile_app_id")
	}
	if req.WeChatConnectMobileAppSecret != "" {
		changed = append(changed, "wechat_connect_mobile_app_secret")
	}
	if before.WeChatConnectOpenEnabled != after.WeChatConnectOpenEnabled {
		changed = append(changed, "wechat_connect_open_enabled")
	}
	if before.WeChatConnectMPEnabled != after.WeChatConnectMPEnabled {
		changed = append(changed, "wechat_connect_mp_enabled")
	}
	if before.WeChatConnectMobileEnabled != after.WeChatConnectMobileEnabled {
		changed = append(changed, "wechat_connect_mobile_enabled")
	}
	if before.WeChatConnectMode != after.WeChatConnectMode {
		changed = append(changed, "wechat_connect_mode")
	}
	if before.WeChatConnectScopes != after.WeChatConnectScopes {
		changed = append(changed, "wechat_connect_scopes")
	}
	if before.WeChatConnectRedirectURL != after.WeChatConnectRedirectURL {
		changed = append(changed, "wechat_connect_redirect_url")
	}
	if before.WeChatConnectFrontendRedirectURL != after.WeChatConnectFrontendRedirectURL {
		changed = append(changed, "wechat_connect_frontend_redirect_url")
	}
	if before.OIDCConnectEnabled != after.OIDCConnectEnabled {
		changed = append(changed, "oidc_connect_enabled")
	}
	if before.OIDCConnectProviderName != after.OIDCConnectProviderName {
		changed = append(changed, "oidc_connect_provider_name")
	}
	if before.OIDCConnectClientID != after.OIDCConnectClientID {
		changed = append(changed, "oidc_connect_client_id")
	}
	if req.OIDCConnectClientSecret != "" {
		changed = append(changed, "oidc_connect_client_secret")
	}
	if before.OIDCConnectIssuerURL != after.OIDCConnectIssuerURL {
		changed = append(changed, "oidc_connect_issuer_url")
	}
	if before.OIDCConnectDiscoveryURL != after.OIDCConnectDiscoveryURL {
		changed = append(changed, "oidc_connect_discovery_url")
	}
	if before.OIDCConnectAuthorizeURL != after.OIDCConnectAuthorizeURL {
		changed = append(changed, "oidc_connect_authorize_url")
	}
	if before.OIDCConnectTokenURL != after.OIDCConnectTokenURL {
		changed = append(changed, "oidc_connect_token_url")
	}
	if before.OIDCConnectUserInfoURL != after.OIDCConnectUserInfoURL {
		changed = append(changed, "oidc_connect_userinfo_url")
	}
	if before.OIDCConnectJWKSURL != after.OIDCConnectJWKSURL {
		changed = append(changed, "oidc_connect_jwks_url")
	}
	if before.OIDCConnectScopes != after.OIDCConnectScopes {
		changed = append(changed, "oidc_connect_scopes")
	}
	if before.OIDCConnectRedirectURL != after.OIDCConnectRedirectURL {
		changed = append(changed, "oidc_connect_redirect_url")
	}
	if before.OIDCConnectFrontendRedirectURL != after.OIDCConnectFrontendRedirectURL {
		changed = append(changed, "oidc_connect_frontend_redirect_url")
	}
	if before.OIDCConnectTokenAuthMethod != after.OIDCConnectTokenAuthMethod {
		changed = append(changed, "oidc_connect_token_auth_method")
	}
	if before.OIDCConnectUsePKCE != after.OIDCConnectUsePKCE {
		changed = append(changed, "oidc_connect_use_pkce")
	}
	if before.OIDCConnectValidateIDToken != after.OIDCConnectValidateIDToken {
		changed = append(changed, "oidc_connect_validate_id_token")
	}
	if before.OIDCConnectAllowedSigningAlgs != after.OIDCConnectAllowedSigningAlgs {
		changed = append(changed, "oidc_connect_allowed_signing_algs")
	}
	if before.OIDCConnectClockSkewSeconds != after.OIDCConnectClockSkewSeconds {
		changed = append(changed, "oidc_connect_clock_skew_seconds")
	}
	if before.OIDCConnectRequireEmailVerified != after.OIDCConnectRequireEmailVerified {
		changed = append(changed, "oidc_connect_require_email_verified")
	}
	if before.OIDCConnectUserInfoEmailPath != after.OIDCConnectUserInfoEmailPath {
		changed = append(changed, "oidc_connect_userinfo_email_path")
	}
	if before.OIDCConnectUserInfoIDPath != after.OIDCConnectUserInfoIDPath {
		changed = append(changed, "oidc_connect_userinfo_id_path")
	}
	if before.OIDCConnectUserInfoUsernamePath != after.OIDCConnectUserInfoUsernamePath {
		changed = append(changed, "oidc_connect_userinfo_username_path")
	}
	if before.SiteName != after.SiteName {
		changed = append(changed, "site_name")
	}
	if before.SiteLogo != after.SiteLogo {
		changed = append(changed, "site_logo")
	}
	if before.SiteSubtitle != after.SiteSubtitle {
		changed = append(changed, "site_subtitle")
	}
	if before.FrontendTheme != after.FrontendTheme {
		changed = append(changed, "frontend_theme")
	}
	if before.APIBaseURL != after.APIBaseURL {
		changed = append(changed, "api_base_url")
	}
	if before.ContactInfo != after.ContactInfo {
		changed = append(changed, "contact_info")
	}
	if before.DocURL != after.DocURL {
		changed = append(changed, "doc_url")
	}
	if before.HomeContent != after.HomeContent {
		changed = append(changed, "home_content")
	}
	if before.HideCcsImportButton != after.HideCcsImportButton {
		changed = append(changed, "hide_ccs_import_button")
	}
	if before.DefaultConcurrency != after.DefaultConcurrency {
		changed = append(changed, "default_concurrency")
	}
	if before.DefaultBalance != after.DefaultBalance {
		changed = append(changed, "default_balance")
	}
	if !equalDefaultSubscriptions(before.DefaultSubscriptions, after.DefaultSubscriptions) {
		changed = append(changed, "default_subscriptions")
	}
	if before.EnableModelFallback != after.EnableModelFallback {
		changed = append(changed, "enable_model_fallback")
	}
	if before.FallbackModelAnthropic != after.FallbackModelAnthropic {
		changed = append(changed, "fallback_model_anthropic")
	}
	if before.FallbackModelOpenAI != after.FallbackModelOpenAI {
		changed = append(changed, "fallback_model_openai")
	}
	if before.FallbackModelGemini != after.FallbackModelGemini {
		changed = append(changed, "fallback_model_gemini")
	}
	if before.FallbackModelAntigravity != after.FallbackModelAntigravity {
		changed = append(changed, "fallback_model_antigravity")
	}
	if before.EnableIdentityPatch != after.EnableIdentityPatch {
		changed = append(changed, "enable_identity_patch")
	}
	if before.IdentityPatchPrompt != after.IdentityPatchPrompt {
		changed = append(changed, "identity_patch_prompt")
	}
	if before.OpsMonitoringEnabled != after.OpsMonitoringEnabled {
		changed = append(changed, "ops_monitoring_enabled")
	}
	if before.OpsRealtimeMonitoringEnabled != after.OpsRealtimeMonitoringEnabled {
		changed = append(changed, "ops_realtime_monitoring_enabled")
	}
	if before.OpsQueryModeDefault != after.OpsQueryModeDefault {
		changed = append(changed, "ops_query_mode_default")
	}
	if before.OpsMetricsIntervalSeconds != after.OpsMetricsIntervalSeconds {
		changed = append(changed, "ops_metrics_interval_seconds")
	}
	if before.MinClaudeCodeVersion != after.MinClaudeCodeVersion {
		changed = append(changed, "min_claude_code_version")
	}
	if before.MaxClaudeCodeVersion != after.MaxClaudeCodeVersion {
		changed = append(changed, "max_claude_code_version")
	}
	if before.AllowUngroupedKeyScheduling != after.AllowUngroupedKeyScheduling {
		changed = append(changed, "allow_ungrouped_key_scheduling")
	}
	if before.BackendModeEnabled != after.BackendModeEnabled {
		changed = append(changed, "backend_mode_enabled")
	}
	if before.PurchaseSubscriptionEnabled != after.PurchaseSubscriptionEnabled {
		changed = append(changed, "purchase_subscription_enabled")
	}
	if before.PurchaseSubscriptionURL != after.PurchaseSubscriptionURL {
		changed = append(changed, "purchase_subscription_url")
	}
	if before.CustomMenuItems != after.CustomMenuItems {
		changed = append(changed, "custom_menu_items")
	}
	if before.EnableFingerprintUnification != after.EnableFingerprintUnification {
		changed = append(changed, "enable_fingerprint_unification")
	}
	if before.EnableMetadataPassthrough != after.EnableMetadataPassthrough {
		changed = append(changed, "enable_metadata_passthrough")
	}
	if before.EnableCCHSigning != after.EnableCCHSigning {
		changed = append(changed, "enable_cch_signing")
	}
	return changed
}

func normalizeDefaultSubscriptions(input []dto.DefaultSubscriptionSetting) []dto.DefaultSubscriptionSetting {
	if len(input) == 0 {
		return nil
	}
	normalized := make([]dto.DefaultSubscriptionSetting, 0, len(input))
	for _, item := range input {
		if item.GroupID <= 0 || item.ValidityDays <= 0 {
			continue
		}
		if item.ValidityDays > service.MaxValidityDays {
			item.ValidityDays = service.MaxValidityDays
		}
		normalized = append(normalized, item)
	}
	return normalized
}

func normalizeOptionalDefaultSubscriptions(input *[]dto.DefaultSubscriptionSetting) *[]dto.DefaultSubscriptionSetting {
	if input == nil {
		return nil
	}
	normalized := normalizeDefaultSubscriptions(*input)
	return &normalized
}

func defaultSubscriptionsValueOrDefault(input *[]dto.DefaultSubscriptionSetting, fallback []service.DefaultSubscriptionSetting) []service.DefaultSubscriptionSetting {
	if input == nil {
		return fallback
	}
	result := make([]service.DefaultSubscriptionSetting, 0, len(*input))
	for _, item := range *input {
		result = append(result, service.DefaultSubscriptionSetting{
			GroupID:      item.GroupID,
			ValidityDays: item.ValidityDays,
		})
	}
	return result
}

func float64ValueOrDefault(input *float64, fallback float64) float64 {
	if input == nil {
		return fallback
	}
	if *input < 0 {
		return 0
	}
	return *input
}

func intValueOrDefault(input *int, fallback int) int {
	if input == nil {
		return fallback
	}
	if *input < 1 {
		return 1
	}
	return *input
}

func boolValueOrDefault(input *bool, fallback bool) bool {
	if input == nil {
		return fallback
	}
	return *input
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalDefaultSubscriptions(a, b []service.DefaultSubscriptionSetting) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].GroupID != b[i].GroupID || a[i].ValidityDays != b[i].ValidityDays {
			return false
		}
	}
	return true
}

// TestSMTPRequest 测试SMTP连接请求
type TestSMTPRequest struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`
}

// TestSMTPConnection 测试SMTP连接
// POST /api/v1/admin/settings/test-smtp
func (h *SettingHandler) TestSMTPConnection(c *gin.Context) {
	var req TestSMTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.SMTPUsername = strings.TrimSpace(req.SMTPUsername)

	var savedConfig *service.SMTPConfig
	if cfg, err := h.emailService.GetSMTPConfig(c.Request.Context()); err == nil && cfg != nil {
		savedConfig = cfg
	}

	if req.SMTPHost == "" && savedConfig != nil {
		req.SMTPHost = savedConfig.Host
	}
	if req.SMTPPort <= 0 {
		if savedConfig != nil && savedConfig.Port > 0 {
			req.SMTPPort = savedConfig.Port
		} else {
			req.SMTPPort = 587
		}
	}
	if req.SMTPUsername == "" && savedConfig != nil {
		req.SMTPUsername = savedConfig.Username
	}
	password := strings.TrimSpace(req.SMTPPassword)
	if password == "" && savedConfig != nil {
		password = savedConfig.Password
	}
	if req.SMTPHost == "" {
		response.BadRequest(c, "SMTP host is required")
		return
	}

	config := &service.SMTPConfig{
		Host:     req.SMTPHost,
		Port:     req.SMTPPort,
		Username: req.SMTPUsername,
		Password: password,
		UseTLS:   req.SMTPUseTLS,
	}

	err := h.emailService.TestSMTPConnectionWithConfig(config)
	if err != nil {
		response.BadRequest(c, "SMTP connection test failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "SMTP connection successful"})
}

// SendTestEmailRequest 发送测试邮件请求
type SendTestEmailRequest struct {
	Email        string `json:"email" binding:"required,email"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from_email"`
	SMTPFromName string `json:"smtp_from_name"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`
}

// SendTestEmail 发送测试邮件
// POST /api/v1/admin/settings/send-test-email
func (h *SettingHandler) SendTestEmail(c *gin.Context) {
	var req SendTestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.SMTPUsername = strings.TrimSpace(req.SMTPUsername)
	req.SMTPFrom = strings.TrimSpace(req.SMTPFrom)
	req.SMTPFromName = strings.TrimSpace(req.SMTPFromName)

	var savedConfig *service.SMTPConfig
	if cfg, err := h.emailService.GetSMTPConfig(c.Request.Context()); err == nil && cfg != nil {
		savedConfig = cfg
	}

	if req.SMTPHost == "" && savedConfig != nil {
		req.SMTPHost = savedConfig.Host
	}
	if req.SMTPPort <= 0 {
		if savedConfig != nil && savedConfig.Port > 0 {
			req.SMTPPort = savedConfig.Port
		} else {
			req.SMTPPort = 587
		}
	}
	if req.SMTPUsername == "" && savedConfig != nil {
		req.SMTPUsername = savedConfig.Username
	}
	password := strings.TrimSpace(req.SMTPPassword)
	if password == "" && savedConfig != nil {
		password = savedConfig.Password
	}
	if req.SMTPFrom == "" && savedConfig != nil {
		req.SMTPFrom = savedConfig.From
	}
	if req.SMTPFromName == "" && savedConfig != nil {
		req.SMTPFromName = savedConfig.FromName
	}
	if req.SMTPHost == "" {
		response.BadRequest(c, "SMTP host is required")
		return
	}

	config := &service.SMTPConfig{
		Host:     req.SMTPHost,
		Port:     req.SMTPPort,
		Username: req.SMTPUsername,
		Password: password,
		From:     req.SMTPFrom,
		FromName: req.SMTPFromName,
		UseTLS:   req.SMTPUseTLS,
	}

	siteName := h.settingService.GetSiteName(c.Request.Context())
	subject := "[" + siteName + "] Test Email"
	body := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .content { padding: 40px 30px; text-align: center; }
        .success { color: #10b981; font-size: 48px; margin-bottom: 20px; }
        .footer { background-color: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>` + siteName + `</h1>
        </div>
        <div class="content">
            <div class="success">✓</div>
            <h2>Email Configuration Successful!</h2>
            <p>This is a test email to verify your SMTP settings are working correctly.</p>
        </div>
        <div class="footer">
            <p>This is an automated test message.</p>
        </div>
    </div>
</body>
</html>
`

	if err := h.emailService.SendEmailWithConfig(config, req.Email, subject, body); err != nil {
		response.BadRequest(c, "Failed to send test email: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Test email sent successfully"})
}

// GetAdminAPIKey 获取管理员 API Key 状态
// GET /api/v1/admin/settings/admin-api-key
func (h *SettingHandler) GetAdminAPIKey(c *gin.Context) {
	maskedKey, exists, err := h.settingService.GetAdminAPIKeyStatus(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"exists":     exists,
		"masked_key": maskedKey,
	})
}

// RegenerateAdminAPIKey 生成/重新生成管理员 API Key
// POST /api/v1/admin/settings/admin-api-key/regenerate
func (h *SettingHandler) RegenerateAdminAPIKey(c *gin.Context) {
	key, err := h.settingService.GenerateAdminAPIKey(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"key": key, // 完整 key 只在生成时返回一次
	})
}

// DeleteAdminAPIKey 删除管理员 API Key
// DELETE /api/v1/admin/settings/admin-api-key
func (h *SettingHandler) DeleteAdminAPIKey(c *gin.Context) {
	if err := h.settingService.DeleteAdminAPIKey(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Admin API key deleted"})
}

// GetOverloadCooldownSettings 获取529过载冷却配置
// GET /api/v1/admin/settings/overload-cooldown
func (h *SettingHandler) GetOverloadCooldownSettings(c *gin.Context) {
	settings, err := h.settingService.GetOverloadCooldownSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OverloadCooldownSettings{
		Enabled:         settings.Enabled,
		CooldownMinutes: settings.CooldownMinutes,
	})
}

// UpdateOverloadCooldownSettingsRequest 更新529过载冷却配置请求
type UpdateOverloadCooldownSettingsRequest struct {
	Enabled         bool `json:"enabled"`
	CooldownMinutes int  `json:"cooldown_minutes"`
}

// UpdateOverloadCooldownSettings 更新529过载冷却配置
// PUT /api/v1/admin/settings/overload-cooldown
func (h *SettingHandler) UpdateOverloadCooldownSettings(c *gin.Context) {
	var req UpdateOverloadCooldownSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings := &service.OverloadCooldownSettings{
		Enabled:         req.Enabled,
		CooldownMinutes: req.CooldownMinutes,
	}

	if err := h.settingService.SetOverloadCooldownSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updatedSettings, err := h.settingService.GetOverloadCooldownSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.OverloadCooldownSettings{
		Enabled:         updatedSettings.Enabled,
		CooldownMinutes: updatedSettings.CooldownMinutes,
	})
}

// GetStreamTimeoutSettings 获取流超时处理配置
// GET /api/v1/admin/settings/stream-timeout
func (h *SettingHandler) GetStreamTimeoutSettings(c *gin.Context) {
	settings, err := h.settingService.GetStreamTimeoutSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.StreamTimeoutSettings{
		Enabled:                settings.Enabled,
		Action:                 settings.Action,
		TempUnschedMinutes:     settings.TempUnschedMinutes,
		ThresholdCount:         settings.ThresholdCount,
		ThresholdWindowMinutes: settings.ThresholdWindowMinutes,
	})
}

// GetRectifierSettings 获取请求整流器配置
// GET /api/v1/admin/settings/rectifier
func (h *SettingHandler) GetRectifierSettings(c *gin.Context) {
	settings, err := h.settingService.GetRectifierSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	patterns := settings.APIKeySignaturePatterns
	if patterns == nil {
		patterns = []string{}
	}
	response.Success(c, dto.RectifierSettings{
		Enabled:                  settings.Enabled,
		ThinkingSignatureEnabled: settings.ThinkingSignatureEnabled,
		ThinkingBudgetEnabled:    settings.ThinkingBudgetEnabled,
		APIKeySignatureEnabled:   settings.APIKeySignatureEnabled,
		APIKeySignaturePatterns:  patterns,
	})
}

// UpdateRectifierSettingsRequest 更新整流器配置请求
type UpdateRectifierSettingsRequest struct {
	Enabled                  bool     `json:"enabled"`
	ThinkingSignatureEnabled bool     `json:"thinking_signature_enabled"`
	ThinkingBudgetEnabled    bool     `json:"thinking_budget_enabled"`
	APIKeySignatureEnabled   bool     `json:"apikey_signature_enabled"`
	APIKeySignaturePatterns  []string `json:"apikey_signature_patterns"`
}

// UpdateRectifierSettings 更新请求整流器配置
// PUT /api/v1/admin/settings/rectifier
func (h *SettingHandler) UpdateRectifierSettings(c *gin.Context) {
	var req UpdateRectifierSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 校验并清理自定义匹配关键词
	const maxPatterns = 50
	const maxPatternLen = 500
	if len(req.APIKeySignaturePatterns) > maxPatterns {
		response.BadRequest(c, "Too many signature patterns (max 50)")
		return
	}
	var cleanedPatterns []string
	for _, p := range req.APIKeySignaturePatterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len(p) > maxPatternLen {
			response.BadRequest(c, "Signature pattern too long (max 500 characters)")
			return
		}
		cleanedPatterns = append(cleanedPatterns, p)
	}

	settings := &service.RectifierSettings{
		Enabled:                  req.Enabled,
		ThinkingSignatureEnabled: req.ThinkingSignatureEnabled,
		ThinkingBudgetEnabled:    req.ThinkingBudgetEnabled,
		APIKeySignatureEnabled:   req.APIKeySignatureEnabled,
		APIKeySignaturePatterns:  cleanedPatterns,
	}

	if err := h.settingService.SetRectifierSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 重新获取设置返回
	updatedSettings, err := h.settingService.GetRectifierSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	updatedPatterns := updatedSettings.APIKeySignaturePatterns
	if updatedPatterns == nil {
		updatedPatterns = []string{}
	}
	response.Success(c, dto.RectifierSettings{
		Enabled:                  updatedSettings.Enabled,
		ThinkingSignatureEnabled: updatedSettings.ThinkingSignatureEnabled,
		ThinkingBudgetEnabled:    updatedSettings.ThinkingBudgetEnabled,
		APIKeySignatureEnabled:   updatedSettings.APIKeySignatureEnabled,
		APIKeySignaturePatterns:  updatedPatterns,
	})
}

// GetBetaPolicySettings 获取 Beta 策略配置
// GET /api/v1/admin/settings/beta-policy
func (h *SettingHandler) GetBetaPolicySettings(c *gin.Context) {
	settings, err := h.settingService.GetBetaPolicySettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	rules := make([]dto.BetaPolicyRule, len(settings.Rules))
	for i, r := range settings.Rules {
		rules[i] = dto.BetaPolicyRule(r)
	}
	response.Success(c, dto.BetaPolicySettings{Rules: rules})
}

// UpdateBetaPolicySettingsRequest 更新 Beta 策略配置请求
type UpdateBetaPolicySettingsRequest struct {
	Rules []dto.BetaPolicyRule `json:"rules"`
}

// UpdateBetaPolicySettings 更新 Beta 策略配置
// PUT /api/v1/admin/settings/beta-policy
func (h *SettingHandler) UpdateBetaPolicySettings(c *gin.Context) {
	var req UpdateBetaPolicySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	rules := make([]service.BetaPolicyRule, len(req.Rules))
	for i, r := range req.Rules {
		rules[i] = service.BetaPolicyRule(r)
	}

	settings := &service.BetaPolicySettings{Rules: rules}
	if err := h.settingService.SetBetaPolicySettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Re-fetch to return updated settings
	updated, err := h.settingService.GetBetaPolicySettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	outRules := make([]dto.BetaPolicyRule, len(updated.Rules))
	for i, r := range updated.Rules {
		outRules[i] = dto.BetaPolicyRule(r)
	}
	response.Success(c, dto.BetaPolicySettings{Rules: outRules})
}

// UpdateStreamTimeoutSettingsRequest 更新流超时配置请求
type UpdateStreamTimeoutSettingsRequest struct {
	Enabled                bool   `json:"enabled"`
	Action                 string `json:"action"`
	TempUnschedMinutes     int    `json:"temp_unsched_minutes"`
	ThresholdCount         int    `json:"threshold_count"`
	ThresholdWindowMinutes int    `json:"threshold_window_minutes"`
}

// UpdateStreamTimeoutSettings 更新流超时处理配置
// PUT /api/v1/admin/settings/stream-timeout
func (h *SettingHandler) UpdateStreamTimeoutSettings(c *gin.Context) {
	var req UpdateStreamTimeoutSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings := &service.StreamTimeoutSettings{
		Enabled:                req.Enabled,
		Action:                 req.Action,
		TempUnschedMinutes:     req.TempUnschedMinutes,
		ThresholdCount:         req.ThresholdCount,
		ThresholdWindowMinutes: req.ThresholdWindowMinutes,
	}

	if err := h.settingService.SetStreamTimeoutSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 重新获取设置返回
	updatedSettings, err := h.settingService.GetStreamTimeoutSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.StreamTimeoutSettings{
		Enabled:                updatedSettings.Enabled,
		Action:                 updatedSettings.Action,
		TempUnschedMinutes:     updatedSettings.TempUnschedMinutes,
		ThresholdCount:         updatedSettings.ThresholdCount,
		ThresholdWindowMinutes: updatedSettings.ThresholdWindowMinutes,
	})
}
