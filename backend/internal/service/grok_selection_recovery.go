package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	grokOnDemandRecoveryMinInterval = 30 * time.Second
	grokOnDemandRecoveryTimeout     = 20 * time.Second
)

type grokOnDemandSelectionRecovery interface {
	RecoverOnDemand(ctx context.Context, accounts []Account, requestedModel string) bool
}

type grokSelectionQuotaRecoverer interface {
	SyncAccount(ctx context.Context, account *Account) error
}

type grokSelectionCapabilityRecoverer interface {
	ProbeAccount(ctx context.Context, account *Account) error
}

type GrokOnDemandRecoveryService struct {
	quotaSync       grokSelectionQuotaRecoverer
	capabilityProbe grokSelectionCapabilityRecoverer
	settingSvc      *SettingService
	now             func() time.Time
	lastRunAt       atomic.Int64
	sf              singleflight.Group
}

func NewGrokOnDemandRecoveryService(
	quotaSync grokSelectionQuotaRecoverer,
	capabilityProbe grokSelectionCapabilityRecoverer,
	settingSvc *SettingService,
) *GrokOnDemandRecoveryService {
	return &GrokOnDemandRecoveryService{
		quotaSync:       quotaSync,
		capabilityProbe: capabilityProbe,
		settingSvc:      settingSvc,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *GrokOnDemandRecoveryService) RecoverOnDemand(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
) bool {
	if s == nil || len(accounts) == 0 {
		return false
	}

	result, _, _ := s.sf.Do("grok_on_demand_selection_recovery", func() (any, error) {
		now := s.now().UTC()
		if last := s.lastAttemptTime(); !last.IsZero() && now.Sub(last) < grokOnDemandRecoveryMinInterval {
			return false, nil
		}
		s.lastRunAt.Store(now.UnixNano())

		recoveryCtx, cancel := context.WithTimeout(withoutCancelContext(ctx), grokOnDemandRecoveryTimeout)
		defer cancel()

		_ = s.recoverAccounts(recoveryCtx, accounts, requestedModel)
		return true, nil
	})

	attempted, _ := result.(bool)
	return attempted
}

func (s *GrokOnDemandRecoveryService) recoverAccounts(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
) error {
	if s == nil || len(accounts) == 0 {
		return nil
	}
	_ = requestedModel

	sessionAccounts := make([]*Account, 0, len(accounts))
	probeAccounts := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok || !account.IsSchedulable() {
			continue
		}
		switch account.Type {
		case AccountTypeSession:
			if s.quotaSync != nil {
				sessionAccounts = append(sessionAccounts, account)
				continue
			}
			if s.capabilityProbe != nil {
				probeAccounts = append(probeAccounts, account)
			}
		case AccountTypeAPIKey, AccountTypeUpstream:
			if s.capabilityProbe != nil {
				probeAccounts = append(probeAccounts, account)
			}
		}
	}

	settings := DefaultGrokRuntimeSettings()
	if s.settingSvc != nil {
		settings = s.settingSvc.GetGrokRuntimeSettings(ctx)
	}

	var errs []error
	if s.quotaSync != nil && len(sessionAccounts) > 0 {
		if err := runGrokParallelAccountPointers(sessionAccounts, settings.UsageSyncWorkers(), func(account *Account) error {
			return s.quotaSync.SyncAccount(ctx, account)
		}); err != nil {
			errs = append(errs, err)
		}
	}
	if s.capabilityProbe != nil && len(probeAccounts) > 0 {
		if err := runGrokParallelAccountPointers(probeAccounts, settings.CapabilityProbeWorkers(), func(account *Account) error {
			return s.capabilityProbe.ProbeAccount(ctx, account)
		}); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (s *GrokOnDemandRecoveryService) lastAttemptTime() time.Time {
	if s == nil {
		return time.Time{}
	}
	raw := s.lastRunAt.Load()
	if raw <= 0 {
		return time.Time{}
	}
	return time.Unix(0, raw).UTC()
}

func (s *GatewayService) SetGrokSelectionRecovery(recovery grokOnDemandSelectionRecovery) {
	if s == nil {
		return
	}
	s.grokSelectionRecovery = recovery
}

func (s *GatewayService) tryRecoverGrokSelection(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
) bool {
	if s == nil || s.grokSelectionRecovery == nil || len(accounts) == 0 {
		return false
	}
	return s.grokSelectionRecovery.RecoverOnDemand(ctx, accounts, requestedModel)
}

func selectSchedulableGrokAccount(
	ctx context.Context,
	gatewayService *GatewayService,
	groupID *int64,
	requestedModel string,
	excludedIDs map[int64]struct{},
	filter grokAccountSelectionFilter,
	noAccountsErr string,
) (*Account, error) {
	if gatewayService == nil {
		return nil, errors.New("grok gateway service is not configured")
	}

	accounts, _, err := gatewayService.listSchedulableAccounts(ctx, groupID, PlatformGrok, true)
	if err != nil {
		return nil, err
	}

	candidates, modelAvailable := resolveSchedulableGrokCandidates(ctx, accounts, requestedModel, excludedIDs, filter)
	if len(candidates) == 0 && gatewayService.tryRecoverGrokSelection(ctx, accounts, requestedModel) {
		candidates, modelAvailable = resolveSchedulableGrokCandidates(ctx, accounts, requestedModel, excludedIDs, filter)
	}

	if len(candidates) == 0 {
		if !modelAvailable {
			return nil, fmt.Errorf("requested model unavailable:%s", requestedModel)
		}
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok accounts"))
	}

	var loadMap map[int64]*AccountLoadInfo
	if gatewayService.concurrencyService != nil {
		if snapshot, loadErr := gatewayService.concurrencyService.GetAccountsLoadBatch(ctx, buildAccountLoadRequests(candidates)); loadErr == nil {
			loadMap = snapshot
		}
	}

	selected := defaultGrokAccountSelector.SelectBestCandidateWithContext(ctx, candidates, requestedModel, loadMap)
	if selected == nil {
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok accounts"))
	}

	hydrated, err := gatewayService.hydrateSelectedAccount(ctx, selected)
	if err != nil {
		return nil, err
	}
	if hydrated == nil {
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok accounts"))
	}
	if filter != nil && !filter(hydrated) {
		return nil, errors.New(firstNonEmpty(noAccountsErr, "no compatible grok accounts"))
	}
	return hydrated, nil
}

func resolveSchedulableGrokCandidates(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
	filter grokAccountSelectionFilter,
) ([]*Account, bool) {
	candidates := defaultGrokAccountSelector.FilterSchedulableCandidatesWithContext(ctx, accounts, requestedModel, excludedIDs)
	if len(candidates) > 0 && filter != nil {
		filtered := make([]*Account, 0, len(candidates))
		for i := range candidates {
			if candidates[i] == nil || !filter(candidates[i]) {
				continue
			}
			filtered = append(filtered, candidates[i])
		}
		candidates = filtered
	}

	if len(candidates) > 0 {
		return candidates, true
	}
	return candidates, defaultGrokAccountSelector.RequestedModelAvailableWithContext(ctx, accounts, requestedModel)
}

func withoutCancelContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}
