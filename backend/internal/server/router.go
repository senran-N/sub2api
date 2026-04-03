package server

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	middleware2 "github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/server/routes"
	"github.com/senran-N/sub2api/internal/web"

	"github.com/gin-gonic/gin"
)

const frameSrcRefreshTimeout = 5 * time.Second

// SetupRouter 配置路由器中间件和路由
func SetupRouter(r *gin.Engine, deps *RouteDependencies) *gin.Engine {
	settingService := deps.Gateway.SettingService
	cfg := deps.Config

	// 缓存 iframe 页面的 origin 列表，用于动态注入 CSP frame-src
	var cachedFrameOrigins atomic.Pointer[[]string]
	emptyOrigins := []string{}
	cachedFrameOrigins.Store(&emptyOrigins)

	refreshFrameOrigins := func() {
		ctx, cancel := context.WithTimeout(context.Background(), frameSrcRefreshTimeout)
		defer cancel()
		origins, err := settingService.GetFrameSrcOrigins(ctx)
		if err != nil {
			// 获取失败时保留已有缓存，避免 frame-src 被意外清空
			return
		}
		cachedFrameOrigins.Store(&origins)
	}
	refreshFrameOrigins() // 启动时初始化

	// 应用中间件
	r.Use(middleware2.RequestLogger())
	r.Use(middleware2.Logger())
	r.Use(middleware2.CORS(cfg.CORS))
	r.Use(middleware2.SecurityHeaders(cfg.Security.CSP, func() []string {
		if p := cachedFrameOrigins.Load(); p != nil {
			return *p
		}
		return nil
	}))

	// Serve embedded frontend with settings injection if available
	if web.HasEmbeddedFrontend() {
		frontendServer, err := web.NewFrontendServer(settingService)
		if err != nil {
			log.Printf("Warning: Failed to create frontend server with settings injection: %v, using legacy mode", err)
			r.Use(web.ServeEmbeddedFrontend())
			settingService.SetOnUpdateCallback(refreshFrameOrigins)
		} else {
			// Register combined callback: invalidate HTML cache + refresh frame origins
			settingService.SetOnUpdateCallback(func() {
				frontendServer.InvalidateCache()
				refreshFrameOrigins()
			})
			r.Use(frontendServer.Middleware())
		}
	} else {
		settingService.SetOnUpdateCallback(refreshFrameOrigins)
	}

	// 注册路由
	registerRoutes(r, deps)

	return r
}

// registerRoutes 注册所有 HTTP 路由
func registerRoutes(r *gin.Engine, deps *RouteDependencies) {
	// 通用路由（健康检查、状态等）
	routes.RegisterCommonRoutes(r)

	// API v1
	v1 := r.Group("/api/v1")

	// 注册各模块路由
	routes.RegisterAuthRoutes(v1, deps.Handlers, deps.JWTAuth, deps.RedisClient, deps.Gateway.SettingService)
	routes.RegisterUserRoutes(v1, deps.Handlers, deps.JWTAuth, deps.Gateway.SettingService)
	routes.RegisterSoraClientRoutes(v1, deps.Handlers, deps.JWTAuth, deps.Gateway.SettingService)
	routes.RegisterAdminRoutes(v1, deps.Handlers, deps.AdminAuth)
	routes.RegisterGatewayRoutes(
		r,
		deps.Handlers,
		deps.Gateway.APIKeyAuth,
		deps.Gateway.APIKeyService,
		deps.Gateway.SubscriptionService,
		deps.Gateway.OpsService,
		deps.Gateway.SettingService,
		deps.Gateway.Config,
	)
}
