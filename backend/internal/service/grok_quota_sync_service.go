package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

const (
	grokQuotaSyncTimeout       = 30 * time.Second
	grokTierSourceQuotaWindows = "quota_windows"
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
	accountRepo         grokQuotaSyncAccountRepo
	stateSvc            *GrokAccountStateService
	tierSvc             *GrokTierService
	settingSvc          *SettingService
	httpUpstream        HTTPUpstream
	tlsFPProfileService *TLSFingerprintProfileService
	now                 func() time.Time

	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewGrokQuotaSyncService(
	accountRepo grokQuotaSyncAccountRepo,
	stateSvc *GrokAccountStateService,
	tierSvc *GrokTierService,
	settingSvc *SettingService,
) *GrokQuotaSyncService {
	if tierSvc == nil {
		tierSvc = NewGrokTierService()
	}
	return &GrokQuotaSyncService{
		accountRepo: accountRepo,
		stateSvc:    stateSvc,
		tierSvc:     tierSvc,
		settingSvc:  settingSvc,
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
	settingSvc *SettingService,
	httpUpstream HTTPUpstream,
	tlsFPProfileService *TLSFingerprintProfileService,
	lifecycle *LifecycleRegistry,
) *GrokQuotaSyncService {
	svc := NewGrokQuotaSyncService(accountRepo, stateSvc, tierSvc, settingSvc)
	svc.httpUpstream = httpUpstream
	svc.tlsFPProfileService = tlsFPProfileService
	if err := svc.SyncNow(context.Background()); err != nil {
		logger.LegacyPrintf("service.grok_quota_sync", "Warning: startup sync failed: %v", err)
	}
	return manageStartStopLifecycle(lifecycle, "GrokQuotaSyncService", svc)
}

func (s *GrokQuotaSyncService) Start() {
	if s == nil || s.accountRepo == nil || s.stateSvc == nil {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			timer := time.NewTimer(s.currentInterval(context.Background()))
			select {
			case <-timer.C:
				ctx, cancel := context.WithTimeout(context.Background(), grokQuotaSyncTimeout)
				err := s.SyncNow(ctx)
				cancel()
				if err != nil {
					logger.LegacyPrintf("service.grok_quota_sync", "Warning: periodic sync failed: %v", err)
				}
			case <-s.stopCh:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			}
		}
	}()
}

func (s *GrokQuotaSyncService) currentInterval(ctx context.Context) time.Duration {
	if s == nil || s.settingSvc == nil {
		return DefaultGrokRuntimeSettings().QuotaSyncInterval()
	}
	return s.settingSvc.GetGrokRuntimeSettings(ctx).QuotaSyncInterval()
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
	runtimeSettings := DefaultGrokRuntimeSettings()
	if s.settingSvc != nil {
		runtimeSettings = s.settingSvc.GetGrokRuntimeSettings(ctx)
	}

	return runGrokParallelAccounts(accounts, runtimeSettings.UsageSyncWorkers(), func(account *Account) error {
		snapshot, syncErr := s.buildSyncSnapshot(ctx, account, now)
		if len(snapshot.Tier) == 0 && len(snapshot.QuotaWindows) == 0 && len(snapshot.SyncState) == 0 {
			return syncErr
		}
		s.stateSvc.PersistSyncSnapshot(ctx, account, snapshot)
		return syncErr
	})
}

func (s *GrokQuotaSyncService) buildSyncSnapshot(
	ctx context.Context,
	account *Account,
	now time.Time,
) (grokStateSyncSnapshot, error) {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return grokStateSyncSnapshot{}, nil
	}

	tierSnapshot := s.tierSnapshot(account)
	normalizedTier := grokNormalizeTier(getStringFromMaps(tierSnapshot, nil, "normalized"))
	quotaWindows := buildGrokSyncedQuotaWindows(account, normalizedTier, now)
	syncState := cloneAnyMap(grokNestedMap(account.grokExtraMap()["sync_state"]))

	liveQuotaAttempted := s.shouldFetchLiveSessionQuota(account)
	var syncErr error
	if liveQuotaAttempted {
		liveResult, err := s.fetchLiveSessionQuota(ctx, account, now)
		if err != nil {
			syncErr = err
		} else if liveResult != nil {
			if liveResult.Tier != grok.TierUnknown {
				normalizedTier = liveResult.Tier
				tierSnapshot = applyLiveGrokTierSnapshot(tierSnapshot, liveResult.Tier)
			}
			quotaWindows = buildGrokSyncedQuotaWindowsFromRaw(liveResult.QuotaWindows, normalizedTier, now)
			syncState = buildGrokQuotaSyncState(syncState, now, true, liveResult.StatusCode, nil)
		}
		if syncErr != nil {
			syncState = buildGrokQuotaSyncState(syncState, now, true, grokSessionRateLimitStatusCode(syncErr), syncErr)
		}
	}
	if !liveQuotaAttempted {
		syncState = buildGrokQuotaSyncState(syncState, now, false, 0, nil)
	}

	return grokStateSyncSnapshot{
		AuthMode:     defaultGrokAuthMode(account.Type),
		Tier:         tierSnapshot,
		QuotaWindows: quotaWindows,
		Capabilities: buildGrokCapabilitySyncSnapshot(account, normalizedTier),
		SyncState:    syncState,
	}, syncErr
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
	return buildGrokSyncedQuotaWindowsFromRaw(account.grokExtraMap()["quota_windows"], tier, now)
}

func buildGrokSyncedQuotaWindowsFromRaw(raw any, tier grok.Tier, now time.Time) map[string]any {
	windows := normalizeGrokQuotaWindows(raw, tier)
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

func buildGrokQuotaSyncState(
	current map[string]any,
	now time.Time,
	attemptedLiveFetch bool,
	statusCode int,
	syncErr error,
) map[string]any {
	syncState := cloneAnyMap(current)
	if len(syncState) == 0 {
		syncState = make(map[string]any, 4)
	}

	timestamp := now.UTC().Format(time.RFC3339)
	syncState["last_sync_at"] = timestamp
	if !attemptedLiveFetch {
		return syncState
	}

	if statusCode > 0 {
		syncState["last_sync_status_code"] = statusCode
	} else {
		delete(syncState, "last_sync_status_code")
	}

	if syncErr == nil {
		syncState["last_sync_ok_at"] = timestamp
		delete(syncState, "last_sync_error")
		delete(syncState, "last_sync_error_at")
		return syncState
	}

	syncState["last_sync_error_at"] = timestamp
	syncState["last_sync_error"] = strings.TrimSpace(syncErr.Error())
	return syncState
}

func applyLiveGrokTierSnapshot(snapshot map[string]any, tier grok.Tier) map[string]any {
	if tier == grok.TierUnknown {
		return cloneAnyMap(snapshot)
	}

	merged := cloneAnyMap(snapshot)
	if len(merged) == 0 {
		merged = make(map[string]any, 3)
	}
	merged["normalized"] = string(tier)
	merged["source"] = grokTierSourceUsageAPI
	merged["confidence"] = 1.0
	return merged
}

func grokSessionRateLimitStatusCode(err error) int {
	var rateErr *grokSessionRateLimitsError
	if errors.As(err, &rateErr) && rateErr != nil {
		return rateErr.StatusCode
	}
	return 0
}
