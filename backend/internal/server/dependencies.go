package server

import (
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/handler"
	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/redis/go-redis/v9"
)

// GatewayRouteDependencies groups route-only gateway collaborators so router
// construction does not need to carry a long, unstable parameter list.
type GatewayRouteDependencies struct {
	APIKeyAuth          middleware2.APIKeyAuthMiddleware
	APIKeyService       *service.APIKeyService
	SubscriptionService *service.SubscriptionService
	OpsService          *service.OpsService
	SettingService      *service.SettingService
	Config              *config.Config
}

// RouteDependencies centralizes router construction inputs and keeps the
// composition root explicit without spreading service references across
// multiple function signatures.
type RouteDependencies struct {
	Config      *config.Config
	Handlers    *handler.Handlers
	JWTAuth     middleware2.JWTAuthMiddleware
	AdminAuth   middleware2.AdminAuthMiddleware
	RedisClient *redis.Client
	Gateway     GatewayRouteDependencies
}

func ProvideGatewayRouteDependencies(
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
) GatewayRouteDependencies {
	return GatewayRouteDependencies{
		APIKeyAuth:          apiKeyAuth,
		APIKeyService:       apiKeyService,
		SubscriptionService: subscriptionService,
		OpsService:          opsService,
		SettingService:      settingService,
		Config:              cfg,
	}
}

func ProvideRouteDependencies(
	cfg *config.Config,
	handlers *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	redisClient *redis.Client,
	gateway GatewayRouteDependencies,
) *RouteDependencies {
	return &RouteDependencies{
		Config:      cfg,
		Handlers:    handlers,
		JWTAuth:     jwtAuth,
		AdminAuth:   adminAuth,
		RedisClient: redisClient,
		Gateway:     gateway,
	}
}
