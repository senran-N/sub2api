package main

import (
	"sync/atomic"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/handler"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestProvideServiceBuildInfo(t *testing.T) {
	in := handler.BuildInfo{
		Version:   "v-test",
		BuildType: "release",
	}
	out := provideServiceBuildInfo(in)
	require.Equal(t, in.Version, out.Version)
	require.Equal(t, in.BuildType, out.BuildType)
}

func TestProvideCleanup_WithMinimalDependencies_NoPanic(t *testing.T) {
	cfg := &config.Config{}

	oauthSvc := service.NewOAuthService(nil, nil)
	openAIOAuthSvc := service.NewOpenAIOAuthService(nil, nil)
	geminiOAuthSvc := service.NewGeminiOAuthService(nil, nil, nil, nil, cfg)
	antigravityOAuthSvc := service.NewAntigravityOAuthService(nil)

	emailQueueSvc := service.NewEmailQueueService(nil, 1)
	billingCacheSvc := service.NewBillingCacheService(nil, nil, nil, nil, cfg)
	lifecycleRegistry := service.NewLifecycleRegistry()

	cleanup := provideCleanup(
		cfg,
		nil, // entClient
		nil, // redis
		lifecycleRegistry,
		nil, // claudeProfileSync
		emailQueueSvc,
		billingCacheSvc,
		&service.UsageRecordWorkerPool{},
		&service.SubscriptionService{},
		oauthSvc,
		openAIOAuthSvc,
		geminiOAuthSvc,
		antigravityOAuthSvc,
		nil, // openAIGateway
		nil, // grokQuotaSync
		nil, // grokCapabilityProbe
	)

	require.NotPanics(t, func() {
		cleanup()
	})
}

func TestProvideCleanup_StopsLifecycleRegistryEntries(t *testing.T) {
	cfg := &config.Config{}
	lifecycleRegistry := service.NewLifecycleRegistry()

	var stopped atomic.Int32
	lifecycleRegistry.Register("probe", func() {
		stopped.Add(1)
	})

	cleanup := provideCleanup(
		cfg,
		nil,
		nil,
		lifecycleRegistry,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	cleanup()

	require.EqualValues(t, 1, stopped.Load())
}
