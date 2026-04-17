package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

type benchmarkIndexedSelectionCache struct {
	all        []*Account
	allValues  []Account
	byID       map[int64]*Account
	indexes    map[string][]*Account
	indexReady map[string]struct{}
	patterns   []string
}

func newBenchmarkIndexedSelectionCache(accounts []Account) *benchmarkIndexedSelectionCache {
	cache := &benchmarkIndexedSelectionCache{
		all:        make([]*Account, 0, len(accounts)),
		byID:       make(map[int64]*Account, len(accounts)),
		indexes:    make(map[string][]*Account),
		indexReady: make(map[string]struct{}),
	}
	for i := range accounts {
		account := accounts[i]
		cloned := account
		cache.all = append(cache.all, &cloned)
		cache.byID[cloned.ID] = &cloned
	}

	deref := derefAccounts(cache.all)
	cache.allValues = deref
	cache.indexes[benchmarkCapabilityIndexKey(SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexAll})] = cache.all
	cache.indexes[benchmarkCapabilityIndexKey(SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexModelAny})] = benchmarkCloneAccountPointers(filterAccountsByCapabilityIndex(deref, SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexModelAny}))
	cache.indexes[benchmarkCapabilityIndexKey(SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexPrivacySet})] = benchmarkCloneAccountPointers(filterAccountsByCapabilityIndex(deref, SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexPrivacySet}))
	cache.indexes[benchmarkCapabilityIndexKey(SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexOpenAIWS})] = benchmarkCloneAccountPointers(filterAccountsByCapabilityIndex(deref, SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexOpenAIWS}))
	cache.patterns = collectCapabilityIndexValues(deref, SchedulerCapabilityIndexModelPattern)
	for key := range cache.indexes {
		cache.indexReady[key] = struct{}{}
	}
	return cache
}

func benchmarkCloneAccountPointers(accounts []Account) []*Account {
	cloned := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		copied := account
		cloned = append(cloned, &copied)
	}
	return cloned
}

func benchmarkCapabilityIndexKey(index SchedulerCapabilityIndex) string {
	return string(index.Kind) + "\x00" + index.Value
}

func (c *benchmarkIndexedSelectionCache) getOrBuildIndexedAccounts(index SchedulerCapabilityIndex) []*Account {
	key := benchmarkCapabilityIndexKey(index)
	if accounts, ok := c.indexes[key]; ok {
		if _, ready := c.indexReady[key]; ready {
			return accounts
		}
	}
	accounts := benchmarkCloneAccountPointers(filterAccountsByCapabilityIndex(c.allValues, index))
	c.indexes[key] = accounts
	c.indexReady[key] = struct{}{}
	return accounts
}

func (c *benchmarkIndexedSelectionCache) GetSnapshot(ctx context.Context, bucket SchedulerBucket) ([]*Account, bool, error) {
	return c.all, true, nil
}

func (c *benchmarkIndexedSelectionCache) SetSnapshot(ctx context.Context, bucket SchedulerBucket, accounts []Account) error {
	return nil
}

func (c *benchmarkIndexedSelectionCache) GetAccount(ctx context.Context, accountID int64) (*Account, error) {
	account := c.byID[accountID]
	if account == nil {
		return nil, nil
	}
	cloned := *account
	return &cloned, nil
}

func (c *benchmarkIndexedSelectionCache) SetAccount(ctx context.Context, account *Account) error {
	return nil
}

func (c *benchmarkIndexedSelectionCache) DeleteAccount(ctx context.Context, accountID int64) error {
	return nil
}

func (c *benchmarkIndexedSelectionCache) UpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	return nil
}

func (c *benchmarkIndexedSelectionCache) TryLockBucket(ctx context.Context, bucket SchedulerBucket, ttl time.Duration) (bool, error) {
	return true, nil
}

func (c *benchmarkIndexedSelectionCache) ListBuckets(ctx context.Context) ([]SchedulerBucket, error) {
	return nil, nil
}

func (c *benchmarkIndexedSelectionCache) GetOutboxWatermark(ctx context.Context) (int64, error) {
	return 0, nil
}

func (c *benchmarkIndexedSelectionCache) SetOutboxWatermark(ctx context.Context, id int64) error {
	return nil
}

func (c *benchmarkIndexedSelectionCache) GetCapabilityIndexPage(ctx context.Context, bucket SchedulerBucket, index SchedulerCapabilityIndex, offset, limit int) ([]*Account, bool, bool, error) {
	accounts := c.indexes[benchmarkCapabilityIndexKey(index)]
	if index.Kind == SchedulerCapabilityIndexModelExact {
		accounts = c.getOrBuildIndexedAccounts(index)
	}
	if offset >= len(accounts) {
		return []*Account{}, true, false, nil
	}
	if limit <= 0 {
		limit = 1
	}
	end := offset + limit
	if end > len(accounts) {
		end = len(accounts)
	}
	return accounts[offset:end], true, end < len(accounts), nil
}

func (c *benchmarkIndexedSelectionCache) HasCapabilityIndexMembers(ctx context.Context, bucket SchedulerBucket, index SchedulerCapabilityIndex, accountIDs []int64) (map[int64]bool, bool, error) {
	accounts := c.indexes[benchmarkCapabilityIndexKey(index)]
	if index.Kind == SchedulerCapabilityIndexModelExact {
		accounts = c.getOrBuildIndexedAccounts(index)
	}
	allowed := make(map[int64]struct{}, len(accounts))
	for i := range accounts {
		allowed[accounts[i].ID] = struct{}{}
	}
	matches := make(map[int64]bool, len(accountIDs))
	for _, accountID := range accountIDs {
		if _, ok := allowed[accountID]; ok {
			matches[accountID] = true
		}
	}
	return matches, true, nil
}

func (c *benchmarkIndexedSelectionCache) ListCapabilityIndexValues(ctx context.Context, bucket SchedulerBucket, kind SchedulerCapabilityIndexKind) ([]string, bool, error) {
	if kind == SchedulerCapabilityIndexModelPattern {
		return append([]string(nil), c.patterns...), true, nil
	}
	return nil, true, nil
}

func buildBenchmarkSelectionAccounts(count int, platform, accountType string) []Account {
	accounts := make([]Account, 0, count)
	baseTime := time.Unix(1_700_000_000, 0)
	for i := 0; i < count; i++ {
		lastUsedAt := baseTime.Add(time.Duration(i) * time.Second)
		accounts = append(accounts, Account{
			ID:          int64(i + 1),
			Name:        "bench-account",
			Platform:    platform,
			Type:        accountType,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 1,
			Priority:    10,
			LastUsedAt:  &lastUsedAt,
		})
	}
	best := &accounts[len(accounts)-1]
	best.Priority = 1
	best.LastUsedAt = nil
	return accounts
}

func reportSchedulingKernelBenchmarkMetrics(b *testing.B, before SchedulingRuntimeKernelMetricsSnapshot) {
	after := SnapshotSchedulingRuntimeKernelMetrics()
	ops := float64(max(1, b.N))
	b.ReportMetric(float64(after.IndexPageFetches-before.IndexPageFetches)/ops, "index_pages/op")
	b.ReportMetric(float64(after.IndexReturnedAccounts-before.IndexReturnedAccounts)/ops, "index_accounts/op")
	b.ReportMetric(float64(after.OrderedRuntimeProbes-before.OrderedRuntimeProbes)/ops, "runtime_probes/op")
	b.ReportMetric(float64(after.RuntimeAcquireAttempts-before.RuntimeAcquireAttempts)/ops, "acquire_attempts/op")
	b.ReportMetric(float64(after.RuntimeAcquireSuccess-before.RuntimeAcquireSuccess)/ops, "acquire_success/op")
	b.ReportMetric(float64(after.RuntimeWaitPlanAttempts-before.RuntimeWaitPlanAttempts)/ops, "waitplan_attempts/op")
}

func benchmarkGatewayServiceWithIndexedSnapshot(accounts []Account, cfg *config.Config) *GatewayService {
	clonedCfg := &config.Config{}
	if cfg != nil {
		*clonedCfg = *cfg
	}
	clonedCfg.Gateway.Scheduling.SnapshotPageSize = 256
	cache := newBenchmarkIndexedSelectionCache(accounts)
	return &GatewayService{
		schedulerSnapshot: &SchedulerSnapshotService{cache: cache},
		cfg:               clonedCfg,
	}
}

func TestBenchmarkIndexedSelectionCache_CachesEmptyExactIndex(t *testing.T) {
	cache := newBenchmarkIndexedSelectionCache(buildBenchmarkSelectionAccounts(2, PlatformAnthropic, AccountTypeAPIKey))
	index := SchedulerCapabilityIndex{Kind: SchedulerCapabilityIndexModelExact, Value: "claude-sonnet-4-5"}
	key := benchmarkCapabilityIndexKey(index)

	first := cache.getOrBuildIndexedAccounts(index)
	second := cache.getOrBuildIndexedAccounts(index)

	if first == nil || second == nil {
		t.Fatalf("expected empty exact index slices to be cached")
	}
	if len(first) != 0 || len(second) != 0 {
		t.Fatalf("expected empty exact index, got %d and %d entries", len(first), len(second))
	}
	if _, ok := cache.indexes[key]; !ok {
		t.Fatalf("expected exact index entry to be stored")
	}
	if _, ok := cache.indexReady[key]; !ok {
		t.Fatalf("expected exact index entry to be marked ready")
	}
}

func BenchmarkGatewayService_SelectAccountForModelWithExclusions_IndexedSnapshot100K(b *testing.B) {
	accounts := buildBenchmarkSelectionAccounts(100000, PlatformAnthropic, AccountTypeAPIKey)
	svc := benchmarkGatewayServiceWithIndexedSnapshot(accounts, &config.Config{})

	ctx := context.Background()
	b.ReportAllocs()
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account, err := svc.SelectAccountForModelWithExclusions(ctx, nil, "", "claude-sonnet-4-5", nil)
		if err != nil {
			b.Fatalf("select gateway account: %v", err)
		}
		if account == nil || account.ID != int64(len(accounts)) {
			b.Fatalf("unexpected gateway account: %+v", account)
		}
	}
	reportSchedulingKernelBenchmarkMetrics(b, before)
}

func BenchmarkGatewayService_SelectAccountForModelWithExclusions_StickyHit100K(b *testing.B) {
	accounts := buildBenchmarkSelectionAccounts(100000, PlatformAnthropic, AccountTypeAPIKey)
	cache := &stubGatewayCache{sessionBindings: map[string]int64{"sticky-session": int64(len(accounts))}}
	svc := benchmarkGatewayServiceWithIndexedSnapshot(accounts, &config.Config{})
	svc.cache = cache

	ctx := context.Background()
	b.ReportAllocs()
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account, err := svc.SelectAccountForModelWithExclusions(ctx, nil, "sticky-session", "claude-sonnet-4-5", nil)
		if err != nil {
			b.Fatalf("select sticky gateway account: %v", err)
		}
		if account == nil || account.ID != int64(len(accounts)) {
			b.Fatalf("unexpected sticky gateway account: %+v", account)
		}
	}
	reportSchedulingKernelBenchmarkMetrics(b, before)
}

func BenchmarkGatewayService_SelectAccountForModelWithExclusions_RoutedHit100K(b *testing.B) {
	accounts := buildBenchmarkSelectionAccounts(100000, PlatformAnthropic, AccountTypeAPIKey)
	groupID := int64(7001)
	targetID := int64(len(accounts))
	svc := benchmarkGatewayServiceWithIndexedSnapshot(accounts, &config.Config{})
	group := &Group{
		ID:                  groupID,
		Platform:            PlatformAnthropic,
		Status:              StatusActive,
		Hydrated:            true,
		ModelRoutingEnabled: true,
		ModelRouting: map[string][]int64{
			"claude-sonnet-*": {targetID},
		},
	}
	ctx := svc.withGroupContext(context.Background(), group)
	b.ReportAllocs()
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account, err := svc.SelectAccountForModelWithExclusions(ctx, &groupID, "", "claude-sonnet-4-5", nil)
		if err != nil {
			b.Fatalf("select routed gateway account: %v", err)
		}
		if account == nil || account.ID != targetID {
			b.Fatalf("unexpected routed gateway account: %+v", account)
		}
	}
	reportSchedulingKernelBenchmarkMetrics(b, before)
}

func BenchmarkGatewayService_SelectAccountWithLoadAwareness_IndexedSnapshot100K(b *testing.B) {
	accounts := buildBenchmarkSelectionAccounts(100000, PlatformAnthropic, AccountTypeAPIKey)
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.SnapshotPageSize = 256
	cfg.Gateway.Scheduling.LoadBatchEnabled = true
	svc := benchmarkGatewayServiceWithIndexedSnapshot(accounts, cfg)
	svc.concurrencyService = NewConcurrencyService(stubConcurrencyCache{})

	ctx := context.Background()
	b.ReportAllocs()
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := svc.SelectAccountWithLoadAwareness(ctx, nil, "", "claude-sonnet-4-5", nil, "")
		if err != nil {
			b.Fatalf("select load-aware gateway account: %v", err)
		}
		if result == nil || result.Account == nil || !result.Acquired || result.WaitPlan != nil {
			b.Fatalf("unexpected load-aware gateway result: %+v", result)
		}
		if result.ReleaseFunc != nil {
			result.ReleaseFunc()
		}
	}
	reportSchedulingKernelBenchmarkMetrics(b, before)
}

func BenchmarkGeminiMessagesCompatService_SelectAccountForModelWithExclusions_IndexedSnapshot100K(b *testing.B) {
	accounts := buildBenchmarkSelectionAccounts(100000, PlatformGemini, AccountTypeOAuth)
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.SnapshotPageSize = 256
	cache := newBenchmarkIndexedSelectionCache(accounts)
	svc := &GeminiMessagesCompatService{
		schedulerSnapshot: &SchedulerSnapshotService{cache: cache},
		cfg:               cfg,
		cache:             &stubGatewayCache{},
	}

	ctx := context.Background()
	b.ReportAllocs()
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account, err := svc.SelectAccountForModelWithExclusions(ctx, nil, "", "gemini-2.5-flash", nil)
		if err != nil {
			b.Fatalf("select gemini account: %v", err)
		}
		if account == nil || account.ID != int64(len(accounts)) {
			b.Fatalf("unexpected gemini account: %+v", account)
		}
	}
	reportSchedulingKernelBenchmarkMetrics(b, before)
}

func BenchmarkOpenAIGatewayService_SelectAccountWithLoadAwareness_IndexedSnapshot100K(b *testing.B) {
	accounts := buildBenchmarkSelectionAccounts(100000, PlatformOpenAI, AccountTypeAPIKey)
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.SnapshotPageSize = 256
	cfg.Gateway.Scheduling.LoadBatchEnabled = true
	cache := newBenchmarkIndexedSelectionCache(accounts)
	svc := &OpenAIGatewayService{
		schedulerSnapshot:  &SchedulerSnapshotService{cache: cache},
		cfg:                cfg,
		cache:              &stubGatewayCache{},
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
	}

	ctx := context.Background()
	b.ReportAllocs()
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := svc.SelectAccountWithLoadAwareness(ctx, nil, "", "gpt-5.2", nil)
		if err != nil {
			b.Fatalf("select openai load-aware account: %v", err)
		}
		if result == nil || result.Account == nil || !result.Acquired || result.WaitPlan != nil {
			b.Fatalf("unexpected openai load-aware result: %+v", result)
		}
		if result.ReleaseFunc != nil {
			result.ReleaseFunc()
		}
	}
	reportSchedulingKernelBenchmarkMetrics(b, before)
}

func BenchmarkOpenAIGatewayService_SelectAccountWithLoadAwareness_StickyHit100K(b *testing.B) {
	accounts := buildBenchmarkSelectionAccounts(100000, PlatformOpenAI, AccountTypeAPIKey)
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.SnapshotPageSize = 256
	cfg.Gateway.Scheduling.LoadBatchEnabled = true
	cache := newBenchmarkIndexedSelectionCache(accounts)
	svc := &OpenAIGatewayService{
		schedulerSnapshot:  &SchedulerSnapshotService{cache: cache},
		cfg:                cfg,
		cache:              &stubGatewayCache{sessionBindings: map[string]int64{"openai:sticky-session": int64(len(accounts))}},
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
	}

	ctx := context.Background()
	b.ReportAllocs()
	resetSchedulingRuntimeKernelStats()
	before := SnapshotSchedulingRuntimeKernelMetrics()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := svc.SelectAccountWithLoadAwareness(ctx, nil, "sticky-session", "gpt-5.2", nil)
		if err != nil {
			b.Fatalf("select openai sticky load-aware account: %v", err)
		}
		if result == nil || result.Account == nil || !result.Acquired || result.WaitPlan != nil {
			b.Fatalf("unexpected openai sticky load-aware result: %+v", result)
		}
		if result.ReleaseFunc != nil {
			result.ReleaseFunc()
		}
	}
	reportSchedulingKernelBenchmarkMetrics(b, before)
}
