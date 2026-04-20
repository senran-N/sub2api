package routes

import (
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/handler"
	"github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterGatewayRoutes 注册 API 网关路由（Claude/OpenAI/Gemini 兼容）
func RegisterGatewayRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
) {
	bodyLimit := middleware.RequestBodyLimit(cfg.Gateway.MaxBodySize)
	clientRequestID := middleware.ClientRequestID()
	opsErrorLogger := handler.OpsErrorLoggerMiddleware(opsService)
	endpointNorm := handler.InboundEndpointMiddleware()
	dispatcher := newGatewayProtocolDispatcher(h)

	// 未分组 Key 拦截中间件（按协议格式区分错误响应）
	requireGroupAnthropic := middleware.RequireGroupAssignment(settingService, middleware.AnthropicErrorWriter)
	requireGroupGoogle := middleware.RequireGroupAssignment(settingService, middleware.GoogleErrorWriter)

	// API网关（Claude API兼容）
	gateway := r.Group("/v1")
	gateway.Use(bodyLimit)
	gateway.Use(clientRequestID)
	gateway.Use(opsErrorLogger)
	gateway.Use(endpointNorm)
	gateway.Use(gin.HandlerFunc(apiKeyAuth))
	gateway.Use(requireGroupAnthropic)
	{
		gateway.POST("/messages", dispatcher.Messages)
		gateway.POST("/messages/count_tokens", dispatcher.CountTokens)
		gateway.GET("/models", dispatcher.Models)
		gateway.GET("/models/:model", dispatcher.GetModel)
		gateway.GET("/usage", h.Gateway.Usage)
		gateway.POST("/responses", dispatcher.Responses)
		gateway.POST("/responses/*subpath", dispatcher.Responses)
		gateway.GET("/responses", h.CompatibleGateway.ResponsesWebSocket)
		gateway.POST("/chat/completions", dispatcher.ChatCompletions)
		gateway.Any("/images/*subpath", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/audio/*subpath", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/embeddings", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/moderations", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/tts", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/stt", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/videos", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/videos/*subpath", dispatcher.OpenAICompatiblePassthrough)
		gateway.Any("/realtime/client_secrets", dispatcher.OpenAICompatiblePassthrough)
	}

	// Gemini 原生 API 兼容层（Gemini SDK/CLI 直连）
	gemini := r.Group("/v1beta")
	gemini.Use(bodyLimit)
	gemini.Use(clientRequestID)
	gemini.Use(opsErrorLogger)
	gemini.Use(endpointNorm)
	gemini.Use(middleware.APIKeyAuthWithSubscriptionGoogle(apiKeyService, subscriptionService, cfg))
	gemini.Use(requireGroupGoogle)
	{
		gemini.GET("/models", h.Gateway.GeminiV1BetaListModels)
		gemini.GET("/models/:model", h.Gateway.GeminiV1BetaGetModel)
		// Gin treats ":" as a param marker, but Gemini uses "{model}:{action}" in the same segment.
		gemini.POST("/models/*modelAction", h.Gateway.GeminiV1BetaModels)
	}

	r.POST("/responses", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, dispatcher.Responses)
	r.POST("/responses/*subpath", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, dispatcher.Responses)
	r.GET("/responses", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, h.CompatibleGateway.ResponsesWebSocket)
	r.POST("/chat/completions", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, dispatcher.ChatCompletions)
	r.GET("/grok/media/assets/:assetID", clientRequestID, opsErrorLogger, endpointNorm, h.GrokGateway.MediaAssetContent)

	grokV1 := r.Group("/grok/v1")
	grokV1.Use(bodyLimit)
	grokV1.Use(clientRequestID)
	grokV1.Use(opsErrorLogger)
	grokV1.Use(endpointNorm)
	grokV1.Use(middleware.ForcePlatform(service.PlatformGrok))
	grokV1.Use(gin.HandlerFunc(apiKeyAuth))
	grokV1.Use(requireGroupAnthropic)
	{
		grokV1.GET("/models", h.GrokGateway.Models)
		grokV1.GET("/models/:model", h.GrokGateway.GetModel)
		grokV1.POST("/chat/completions", h.GrokGateway.ChatCompletions)
		grokV1.POST("/responses", h.GrokGateway.Responses)
		grokV1.POST("/responses/*subpath", h.GrokGateway.Responses)
		grokV1.POST("/messages", h.GrokGateway.Messages)
		grokV1.Any("/images/*subpath", h.GrokGateway.Images)
		grokV1.Any("/audio/*subpath", h.GrokGateway.Passthrough)
		grokV1.Any("/embeddings", h.GrokGateway.Passthrough)
		grokV1.Any("/moderations", h.GrokGateway.Passthrough)
		grokV1.Any("/tts", h.GrokGateway.Passthrough)
		grokV1.Any("/stt", h.GrokGateway.Passthrough)
		grokV1.Any("/videos", h.GrokGateway.Videos)
		grokV1.Any("/videos/*subpath", h.GrokGateway.Videos)
		grokV1.Any("/realtime/client_secrets", h.GrokGateway.Passthrough)
	}

	// Antigravity 模型列表
	r.GET("/antigravity/models", gin.HandlerFunc(apiKeyAuth), requireGroupAnthropic, h.Gateway.AntigravityModels)

	// Antigravity 专用路由（仅使用 antigravity 账户，不混合调度）
	antigravityV1 := r.Group("/antigravity/v1")
	antigravityV1.Use(bodyLimit)
	antigravityV1.Use(clientRequestID)
	antigravityV1.Use(opsErrorLogger)
	antigravityV1.Use(endpointNorm)
	antigravityV1.Use(middleware.ForcePlatform(service.PlatformAntigravity))
	antigravityV1.Use(gin.HandlerFunc(apiKeyAuth))
	antigravityV1.Use(requireGroupAnthropic)
	{
		antigravityV1.POST("/messages", h.Gateway.Messages)
		antigravityV1.POST("/messages/count_tokens", h.Gateway.CountTokens)
		antigravityV1.GET("/models", h.Gateway.AntigravityModels)
		antigravityV1.GET("/usage", h.Gateway.Usage)
	}

	antigravityV1Beta := r.Group("/antigravity/v1beta")
	antigravityV1Beta.Use(bodyLimit)
	antigravityV1Beta.Use(clientRequestID)
	antigravityV1Beta.Use(opsErrorLogger)
	antigravityV1Beta.Use(endpointNorm)
	antigravityV1Beta.Use(middleware.ForcePlatform(service.PlatformAntigravity))
	antigravityV1Beta.Use(middleware.APIKeyAuthWithSubscriptionGoogle(apiKeyService, subscriptionService, cfg))
	antigravityV1Beta.Use(requireGroupGoogle)
	{
		antigravityV1Beta.GET("/models", h.Gateway.GeminiV1BetaListModels)
		antigravityV1Beta.GET("/models/:model", h.Gateway.GeminiV1BetaGetModel)
		antigravityV1Beta.POST("/models/*modelAction", h.Gateway.GeminiV1BetaModels)
	}
}
