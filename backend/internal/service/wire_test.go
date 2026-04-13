package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/zeromicro/go-zero/core/collection"
)

func TestProvideTimingWheelService_ReturnsError(t *testing.T) {
	original := newTimingWheel
	t.Cleanup(func() { newTimingWheel = original })

	newTimingWheel = func(_ time.Duration, _ int, _ collection.Execute) (*collection.TimingWheel, error) {
		return nil, errors.New("boom")
	}

	svc, err := ProvideTimingWheelService(NewLifecycleRegistry())
	if err == nil {
		t.Fatalf("期望返回 error，但得到 nil")
	}
	if svc != nil {
		t.Fatalf("期望返回 nil svc，但得到非空")
	}
}

func TestProvideTimingWheelService_Success(t *testing.T) {
	registry := NewLifecycleRegistry()
	svc, err := ProvideTimingWheelService(registry)
	if err != nil {
		t.Fatalf("期望 err 为 nil，但得到: %v", err)
	}
	if svc == nil {
		t.Fatalf("期望 svc 非空，但得到 nil")
	}
	entries := registry.Entries()
	if len(entries) != 1 || entries[0].Name != "TimingWheelService" {
		t.Fatalf("unexpected lifecycle entries: %+v", entries)
	}
	entries[0].Stop()
}

func TestProvideDashboardAggregationService_RegistersLifecycleStop(t *testing.T) {
	registry := NewLifecycleRegistry()
	timingWheel, err := NewTimingWheelService()
	if err != nil {
		t.Fatalf("expected timing wheel, got error: %v", err)
	}
	defer timingWheel.Stop()

	cfg := &config.Config{
		DashboardAgg: config.DashboardAggregationConfig{
			Enabled:         true,
			IntervalSeconds: 60,
		},
	}

	svc := ProvideDashboardAggregationService(&dashboardAggregationRepoTestStub{}, timingWheel, cfg, registry)
	if svc == nil {
		t.Fatalf("expected dashboard aggregation service")
	}

	entries := registry.Entries()
	if len(entries) != 1 || entries[0].Name != "DashboardAggregationService" {
		t.Fatalf("unexpected lifecycle entries: %+v", entries)
	}

	entries[0].Stop()
}

type apiKeyAuthLifecycleCacheStub struct {
	subscribeCtx context.Context
}

func (s *apiKeyAuthLifecycleCacheStub) GetCreateAttemptCount(context.Context, int64) (int, error) {
	return 0, nil
}

func (s *apiKeyAuthLifecycleCacheStub) IncrementCreateAttemptCount(context.Context, int64) error {
	return nil
}

func (s *apiKeyAuthLifecycleCacheStub) DeleteCreateAttemptCount(context.Context, int64) error {
	return nil
}

func (s *apiKeyAuthLifecycleCacheStub) IncrementDailyUsage(context.Context, string) error {
	return nil
}

func (s *apiKeyAuthLifecycleCacheStub) SetDailyUsageExpiry(context.Context, string, time.Duration) error {
	return nil
}

func (s *apiKeyAuthLifecycleCacheStub) GetAuthCache(context.Context, string) (*APIKeyAuthCacheEntry, error) {
	return nil, nil
}

func (s *apiKeyAuthLifecycleCacheStub) SetAuthCache(context.Context, string, *APIKeyAuthCacheEntry, time.Duration) error {
	return nil
}

func (s *apiKeyAuthLifecycleCacheStub) DeleteAuthCache(context.Context, string) error {
	return nil
}

func (s *apiKeyAuthLifecycleCacheStub) PublishAuthCacheInvalidation(context.Context, string) error {
	return nil
}

func (s *apiKeyAuthLifecycleCacheStub) SubscribeAuthCacheInvalidation(ctx context.Context, handler func(string)) error {
	s.subscribeCtx = ctx
	return nil
}

func TestProvideAPIKeyAuthCacheInvalidator_RegistersLifecycleCancellation(t *testing.T) {
	registry := NewLifecycleRegistry()
	cache := &apiKeyAuthLifecycleCacheStub{}
	cfg := &config.Config{
		APIKeyAuth: config.APIKeyAuthCacheConfig{
			L1Size:       4,
			L1TTLSeconds: 60,
		},
	}
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, cfg)

	got := ProvideAPIKeyAuthCacheInvalidator(svc, registry)
	if got != svc {
		t.Fatalf("expected provider to return original service instance")
	}
	if cache.subscribeCtx == nil {
		t.Fatalf("expected auth cache subscription to start")
	}

	entries := registry.Entries()
	if len(entries) != 1 || entries[0].Name != "APIKeyAuthCacheInvalidationSubscriber" {
		t.Fatalf("unexpected lifecycle entries: %+v", entries)
	}

	entries[0].Stop()

	deadline := time.Now().Add(time.Second)
	for cache.subscribeCtx.Err() == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if cache.subscribeCtx.Err() != context.Canceled {
		t.Fatalf("expected subscription context canceled, got %v", cache.subscribeCtx.Err())
	}
}
