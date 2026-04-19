package admin

import "github.com/senran-N/sub2api/internal/service"

func ProvideUserHandler(adminService service.AdminService, concurrencyService *service.ConcurrencyService) *UserHandler {
	return NewUserHandler(adminService, concurrencyService)
}

func ProvideGroupHandler(
	adminService service.AdminService,
	dashboardService *service.DashboardService,
	groupCapacityService *service.GroupCapacityService,
) *GroupHandler {
	return NewGroupHandler(adminService, dashboardService, groupCapacityService)
}

func ProvideAccountHandler(
	adminService service.AdminService,
	oauthService *service.OAuthService,
	openaiOAuthService *service.OpenAIOAuthService,
	geminiOAuthService *service.GeminiOAuthService,
	antigravityOAuthService *service.AntigravityOAuthService,
	compatibleUpstreamModelsService *service.CompatibleUpstreamModelsService,
	rateLimitService *service.RateLimitService,
	accountUsageService *service.AccountUsageService,
	accountTestService *service.AccountTestService,
	concurrencyService *service.ConcurrencyService,
	crsSyncService *service.CRSSyncService,
	sessionLimitCache service.SessionLimitCache,
	rpmCache service.RPMCache,
	tokenCacheInvalidator service.TokenCacheInvalidator,
) *AccountHandler {
	return NewAccountHandler(
		adminService,
		oauthService,
		openaiOAuthService,
		geminiOAuthService,
		antigravityOAuthService,
		compatibleUpstreamModelsService,
		rateLimitService,
		accountUsageService,
		accountTestService,
		concurrencyService,
		crsSyncService,
		sessionLimitCache,
		rpmCache,
		tokenCacheInvalidator,
	)
}

func ProvideOpenAIOAuthHandler(openaiOAuthService *service.OpenAIOAuthService, adminService service.AdminService) *OpenAIOAuthHandler {
	return NewOpenAIOAuthHandler(openaiOAuthService, adminService)
}

func ProvideProxyHandler(adminService service.AdminService) *ProxyHandler {
	return NewProxyHandler(adminService)
}

func ProvideRedeemHandler(adminService service.AdminService, redeemService *service.RedeemService) *RedeemHandler {
	return NewRedeemHandler(adminService, redeemService)
}

func ProvideUsageHandler(
	usageService *service.UsageService,
	apiKeyService *service.APIKeyService,
	adminService service.AdminService,
	cleanupService *service.UsageCleanupService,
) *UsageHandler {
	return NewUsageHandler(usageService, apiKeyService, adminService, cleanupService)
}

func ProvideAdminAPIKeyHandler(adminService service.AdminService) *AdminAPIKeyHandler {
	return NewAdminAPIKeyHandler(adminService)
}
