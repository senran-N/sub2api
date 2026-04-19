package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

const (
	defaultGrokQuotaSyncInterval = 15 * time.Minute
	grokQuotaSyncTimeout         = 30 * time.Second
	grokTierSourceQuotaWindows   = "quota_windows"
)

type grokQuotaSyncAccountRepo interface {
	ListByPlatform(ctx context.Context, platform string) ([]Account, error)
}

type GrokTierService struct{}

func NewGrokTierService() *GrokTierService {
	return &GrokTierService{}
}

func (s *GrokTierService) BuildSnapshot(account *Account) map[string]any {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return nil
	}

	state := account.GrokTierState()
	snapshot := map[string]any{
		"normalized": string(state.Normalized),
	}
	if raw := strings.TrimSpace(state.Raw); raw != "" {
		snapshot["raw"] = raw
	}
	if source := strings.TrimSpace(state.Source); source != "" {
		snapshot["source"] = source
	} else if inferredSource := inferGrokTierSnapshotSource(account, state); inferredSource != "" {
		snapshot["source"] = inferredSource
	}
	if state.Confidence > 0 {
		snapshot["confidence"] = state.Confidence
	}
	return snapshot
}

func inferGrokTierSnapshotSource(account *Account, state GrokTierState) string {
	if account == nil {
		return ""
	}
	if strings.TrimSpace(state.Raw) != "" {
		return "raw"
	}
	if state.Normalized != grok.TierUnknown && account.grokQuotaWindow(grok.QuotaWindowAuto).Total > 0 {
		return grokTierSourceQuotaWindows
	}
	if state.Normalized != grok.TierUnknown {
		return "sync"
	}
	return ""
}

type GrokQuotaSyncService struct {
	accountRepo grokQuotaSyncAccountRepo
	stateSvc    *GrokAccountStateService
	tierSvc     *GrokTierService
	interval    time.Duration
	now         func() time.Time

	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewGrokQuotaSyncService(
	accountRepo grokQuotaSyncAccountRepo,
	stateSvc *GrokAccountStateService,
	tierSvc *GrokTierService,
) *GrokQuotaSyncService {
	if tierSvc == nil {
		tierSvc = NewGrokTierService()
	}
	return &GrokQuotaSyncService{
		accountRepo: accountRepo,
		stateSvc:    stateSvc,
		tierSvc:     tierSvc,
		interval:    defaultGrokQuotaSyncInterval,
		now: func() time.Time {
			return time.Now().UTC()
		},
		stopCh: make(chan struct{}),
	}
}

func ProvideGrokQuotaSyncService(
	accountRepo AccountRepository,
	stateSvc *GrokAccountStateService,
	tierSvc *GrokTierService,
	lifecycle *LifecycleRegistry,
) *GrokQuotaSyncService {
	svc := NewGrokQuotaSyncService(accountRepo, stateSvc, tierSvc)
	if err := svc.SyncNow(context.Background()); err != nil {
		logger.LegacyPrintf("service.grok_quota_sync", "Warning: startup sync failed: %v", err)
	}
	return manageStartStopLifecycle(lifecycle, "GrokQuotaSyncService", svc)
}

func (s *GrokQuotaSyncService) Start() {
	if s == nil || s.accountRepo == nil || s.stateSvc == nil || s.interval <= 0 {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), grokQuotaSyncTimeout)
				err := s.SyncNow(ctx)
				cancel()
				if err != nil {
					logger.LegacyPrintf("service.grok_quota_sync", "Warning: periodic sync failed: %v", err)
				}
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *GrokQuotaSyncService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *GrokQuotaSyncService) SyncNow(ctx context.Context) error {
	if s == nil || s.accountRepo == nil || s.stateSvc == nil {
		return nil
	}

	accounts, err := s.accountRepo.ListByPlatform(ctx, PlatformGrok)
	if err != nil {
		return err
	}

	now := s.now().UTC()
	for i := range accounts {
		account := &accounts[i]
		snapshot := s.buildSyncSnapshot(account, now)
		if len(snapshot.Tier) == 0 && len(snapshot.QuotaWindows) == 0 && len(snapshot.SyncState) == 0 {
			continue
		}
		s.stateSvc.PersistSyncSnapshot(ctx, account, snapshot)
	}
	return nil
}

func (s *GrokQuotaSyncService) buildSyncSnapshot(account *Account, now time.Time) grokStateSyncSnapshot {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return grokStateSyncSnapshot{}
	}

	tierSnapshot := s.tierSnapshot(account)
	normalizedTier := grokNormalizeTier(getStringFromMaps(tierSnapshot, nil, "normalized"))
	quotaWindows := buildGrokSyncedQuotaWindows(account, normalizedTier, now)

	syncState := cloneAnyMap(grokNestedMap(account.grokExtraMap()["sync_state"]))
	if len(syncState) == 0 {
		syncState = make(map[string]any, 1)
	}
	syncState["last_sync_at"] = now.Format(time.RFC3339)

	return grokStateSyncSnapshot{
		AuthMode:     defaultGrokAuthMode(account.Type),
		Tier:         tierSnapshot,
		QuotaWindows: quotaWindows,
		Capabilities: buildGrokCapabilitySyncSnapshot(account, normalizedTier),
		SyncState:    syncState,
	}
}

func (s *GrokQuotaSyncService) tierSnapshot(account *Account) map[string]any {
	if s == nil || s.tierSvc == nil {
		return nil
	}
	return s.tierSvc.BuildSnapshot(account)
}

func buildGrokSyncedQuotaWindows(account *Account, tier grok.Tier, now time.Time) map[string]any {
	if account == nil {
		return nil
	}

	grokExtra := account.grokExtraMap()
	windows := normalizeGrokQuotaWindows(grokExtra["quota_windows"], tier)
	if len(windows) == 0 {
		return nil
	}

	for windowName, rawWindow := range windows {
		window := grokNestedMap(rawWindow)
		if len(window) == 0 {
			continue
		}

		if _, ok := window["remaining"]; !ok {
			if total := grokParseInt(window["total"]); total > 0 {
				window["remaining"] = total
			}
		}
		if source := strings.TrimSpace(getStringFromMaps(window, nil, "source")); source == "" {
			window["source"] = grok.QuotaSourceDefault
		}
		if resetAt := grokParseTime(window["reset_at"]); resetAt != nil && !resetAt.After(now) {
			if total := grokParseInt(window["total"]); total > 0 {
				window["remaining"] = total
			}
			window["reset_at"] = ""
		}
		windows[windowName] = window
	}

	return windows
}
