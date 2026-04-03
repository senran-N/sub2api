//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"net/http"
	"time"

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
	entClient *ent.Client,
	rdb *redis.Client,
	opsMetricsCollector *service.OpsMetricsCollector,
	opsAggregation *service.OpsAggregationService,
	opsAlertEvaluator *service.OpsAlertEvaluatorService,
	opsCleanup *service.OpsCleanupService,
	opsScheduledReport *service.OpsScheduledReportService,
	opsSystemLogSink *service.OpsSystemLogSink,
	soraMediaCleanup *service.SoraMediaCleanupService,
	schedulerSnapshot *service.SchedulerSnapshotService,
	tokenRefresh *service.TokenRefreshService,
	accountExpiry *service.AccountExpiryService,
	subscriptionExpiry *service.SubscriptionExpiryService,
	usageCleanup *service.UsageCleanupService,
	idempotencyCleanup *service.IdempotencyCleanupService,
	pricing *service.PricingService,
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
	scheduledTestRunner *service.ScheduledTestRunnerService,
	backupSvc *service.BackupService,
) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 应用层清理步骤可并行执行，基础设施资源（Redis/Ent）最后按顺序关闭。
		parallelSteps := []cleanupStep{
			stopStep("OpsScheduledReportService", opsScheduledReport),
			stopStep("OpsCleanupService", opsCleanup),
			stopStep("OpsSystemLogSink", opsSystemLogSink),
			stopStep("SoraMediaCleanupService", soraMediaCleanup),
			stopStep("OpsAlertEvaluatorService", opsAlertEvaluator),
			stopStep("OpsAggregationService", opsAggregation),
			stopStep("OpsMetricsCollector", opsMetricsCollector),
			stopStep("SchedulerSnapshotService", schedulerSnapshot),
			stopStep("UsageCleanupService", usageCleanup),
			stopStep("IdempotencyCleanupService", idempotencyCleanup),
			stopStep("ClaudeCodeProfileSyncService", claudeProfileSync),
			stopStep("TokenRefreshService", tokenRefresh),
			stopStep("AccountExpiryService", accountExpiry),
			stopStep("SubscriptionExpiryService", subscriptionExpiry),
			stopStep("SubscriptionService", subscriptionService),
			stopStep("PricingService", pricing),
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
			stopStep("ScheduledTestRunnerService", scheduledTestRunner),
			stopStep("BackupService", backupSvc),
		}

		infraSteps := []cleanupStep{
			closeStep("Redis", rdb),
			closeStep("Ent", entClient),
		}

		runCleanup(ctx, parallelSteps, infraSteps)
	}
}
