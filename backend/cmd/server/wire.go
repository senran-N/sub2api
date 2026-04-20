//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"net/http"

	"github.com/senran-N/sub2api/ent"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/handler"
	"github.com/senran-N/sub2api/internal/repository"
	"github.com/senran-N/sub2api/internal/server"
	"github.com/senran-N/sub2api/internal/server/middleware"
	"github.com/senran-N/sub2api/internal/service"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	Server  *http.Server
	Cleanup func()
}

func initializeApplication(buildInfo handler.BuildInfo) (*Application, error) {
	wire.Build(
		// Infrastructure layer ProviderSets
		config.ProviderSet,

		// Business layer ProviderSets
		repository.ProviderSet,
		service.ProviderSet,
		middleware.ProviderSet,
		handler.ProviderSet,

		// Server layer ProviderSet
		server.ProviderSet,

		// Privacy client factory for OpenAI training opt-out
		providePrivacyClientFactory,

		// BuildInfo provider
		provideServiceBuildInfo,

		// Cleanup function provider
		provideCleanup,

		// Application struct
		wire.Struct(new(Application), "Server", "Cleanup"),
	)
	return nil, nil
}

func providePrivacyClientFactory() service.PrivacyClientFactory {
	return repository.CreatePrivacyReqClient
}

func provideServiceBuildInfo(buildInfo handler.BuildInfo) service.BuildInfo {
	return service.BuildInfo{
		Version:   buildInfo.Version,
		BuildType: buildInfo.BuildType,
	}
}

func provideCleanup(
	cfg *config.Config,
	entClient *ent.Client,
	rdb *redis.Client,
	lifecycleRegistry *service.LifecycleRegistry,
	claudeProfileSync *service.ClaudeCodeProfileSyncService,
	emailQueue *service.EmailQueueService,
	billingCache *service.BillingCacheService,
	usageRecordWorkerPool *service.UsageRecordWorkerPool,
	subscriptionService *service.SubscriptionService,
	oauth *service.OAuthService,
	openaiOAuth *service.OpenAIOAuthService,
	geminiOAuth *service.GeminiOAuthService,
	antigravityOAuth *service.AntigravityOAuthService,
	openAIGateway *service.OpenAIGatewayService,
	_ *service.GrokQuotaSyncService,
	_ *service.GrokCapabilityProbeService,
) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), resolveShutdownTimeout(cfg))
		defer cancel()

		// 应用层清理步骤可并行执行，基础设施资源（Redis/Ent）最后按顺序关闭。
		parallelSteps := lifecycleRegistrySteps(lifecycleRegistry)
		parallelSteps = append(parallelSteps, []cleanupStep{
			stopStep("ClaudeCodeProfileSyncService", claudeProfileSync),
			stopStep("SubscriptionService", subscriptionService),
			stopStep("EmailQueueService", emailQueue),
			stopStep("BillingCacheService", billingCache),
			stopStep("UsageRecordWorkerPool", usageRecordWorkerPool),
			stopStep("OAuthService", oauth),
			stopStep("OpenAIOAuthService", openaiOAuth),
			stopStep("GeminiOAuthService", geminiOAuth),
			stopStep("AntigravityOAuthService", antigravityOAuth),
			callbackStep("OpenAIWSPool", func() {
				if openAIGateway != nil {
					openAIGateway.CloseOpenAIWSPool()
				}
			}),
		}...)

		infraSteps := []cleanupStep{
			closeStep("Redis", rdb),
			closeStep("Ent", entClient),
		}

		runCleanup(ctx, parallelSteps, infraSteps)
	}
}

func lifecycleRegistrySteps(registry *service.LifecycleRegistry) []cleanupStep {
	entries := registry.Entries()
	steps := make([]cleanupStep, 0, len(entries))
	for _, entry := range entries {
		entry := entry
		steps = append(steps, callbackStep(entry.Name, entry.Stop))
	}
	return steps
}
