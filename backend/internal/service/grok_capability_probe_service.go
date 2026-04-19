package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/pkg/tlsfingerprint"
)

const (
	defaultGrokCapabilityProbeInterval = 6 * time.Hour
	grokCapabilityProbeTimeout         = 20 * time.Second
)

type grokCapabilityProbeAccountRepo interface {
	ListByPlatform(ctx context.Context, platform string) ([]Account, error)
}

type GrokCapabilityProbeService struct {
	accountRepo         grokCapabilityProbeAccountRepo
	stateSvc            *GrokAccountStateService
	httpUpstream        HTTPUpstream
	cfg                 *config.Config
	tlsFPProfileService *TLSFingerprintProfileService
	interval            time.Duration
	now                 func() time.Time
	stopCh              chan struct{}
	stopOnce            sync.Once
	wg                  sync.WaitGroup
}

func NewGrokCapabilityProbeService(
	accountRepo grokCapabilityProbeAccountRepo,
	stateSvc *GrokAccountStateService,
	httpUpstream HTTPUpstream,
	cfg *config.Config,
	tlsFPProfileService *TLSFingerprintProfileService,
) *GrokCapabilityProbeService {
	return &GrokCapabilityProbeService{
		accountRepo:         accountRepo,
		stateSvc:            stateSvc,
		httpUpstream:        httpUpstream,
		cfg:                 cfg,
		tlsFPProfileService: tlsFPProfileService,
		interval:            defaultGrokCapabilityProbeInterval,
		now: func() time.Time {
			return time.Now().UTC()
		},
		stopCh: make(chan struct{}),
	}
}

func ProvideGrokCapabilityProbeService(
	accountRepo AccountRepository,
	stateSvc *GrokAccountStateService,
	httpUpstream HTTPUpstream,
	cfg *config.Config,
	tlsFPProfileService *TLSFingerprintProfileService,
	lifecycle *LifecycleRegistry,
) *GrokCapabilityProbeService {
	svc := NewGrokCapabilityProbeService(accountRepo, stateSvc, httpUpstream, cfg, tlsFPProfileService)
	if err := svc.ProbeNow(context.Background()); err != nil {
		logger.LegacyPrintf("service.grok_capability_probe", "Warning: startup probe failed: %v", err)
	}
	return manageStartStopLifecycle(lifecycle, "GrokCapabilityProbeService", svc)
}

func (s *GrokCapabilityProbeService) Start() {
	if s == nil || s.accountRepo == nil || s.stateSvc == nil || s.httpUpstream == nil || s.interval <= 0 {
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
				ctx, cancel := context.WithTimeout(context.Background(), grokCapabilityProbeTimeout)
				err := s.ProbeNow(ctx)
				cancel()
				if err != nil {
					logger.LegacyPrintf("service.grok_capability_probe", "Warning: periodic probe failed: %v", err)
				}
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *GrokCapabilityProbeService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *GrokCapabilityProbeService) ProbeNow(ctx context.Context) error {
	if s == nil || s.accountRepo == nil || s.stateSvc == nil || s.httpUpstream == nil {
		return nil
	}

	accounts, err := s.accountRepo.ListByPlatform(ctx, PlatformGrok)
	if err != nil {
		return err
	}

	now := s.now().UTC()
	var firstErr error
	for i := range accounts {
		account := &accounts[i]
		if !s.shouldProbeAccount(account, now) {
			continue
		}
		if err := s.probeAccount(ctx, account); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			logger.LegacyPrintf("service.grok_capability_probe", "Warning: account %d probe failed: %v", account.ID, err)
		}
	}
	return firstErr
}

func (s *GrokCapabilityProbeService) shouldProbeAccount(account *Account, now time.Time) bool {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return false
	}
	if account.Type != AccountTypeAPIKey && account.Type != AccountTypeUpstream {
		return false
	}
	if !account.IsSchedulable() {
		return false
	}
	if account.GrokTierState().Normalized != grok.TierUnknown {
		return false
	}

	capabilityState := account.grokCapabilities()
	syncState := account.grokSyncState()
	if !capabilityState.hasModelSignal && !capabilityState.hasOperationSignal {
		return true
	}
	if syncState.LastProbeAt == nil {
		return true
	}
	return now.Sub(*syncState.LastProbeAt) >= s.interval
}

func (s *GrokCapabilityProbeService) probeAccount(ctx context.Context, account *Account) error {
	target, err := resolveGrokTransportTarget(account, s.validateUpstreamBaseURL)
	if err != nil {
		s.stateSvc.PersistProbeResult(ctx, account, grok.DefaultTestModel, nil, err)
		return err
	}
	if target.Kind != grokTransportKindCompatible {
		return nil
	}

	payload := createCompatibleGatewayTestPayload(grok.DefaultTestModel, false, "")
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal grok capability probe payload: %w", err)
	}

	reqCtx := ctx
	if reqCtx == nil {
		reqCtx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(reqCtx, grokCapabilityProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, target.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("create grok capability probe request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	target.Apply(req)

	resp, err := s.httpUpstream.DoWithTLS(
		req,
		accountTestProxyURL(account),
		account.ID,
		account.Concurrency,
		s.tlsProfile(account),
	)
	if err != nil {
		probeErr := fmt.Errorf("request failed: %w", err)
		s.stateSvc.PersistProbeResult(ctx, account, grok.DefaultTestModel, nil, probeErr)
		return probeErr
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		probeErr := fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
		s.stateSvc.PersistProbeResult(ctx, account, grok.DefaultTestModel, resp, probeErr)
		return probeErr
	}

	s.stateSvc.PersistProbeResult(ctx, account, grok.DefaultTestModel, resp, nil)
	return nil
}

func (s *GrokCapabilityProbeService) validateUpstreamBaseURL(raw string) (string, error) {
	return validateCompatibleUpstreamBaseURL(s.cfg, raw)
}

func (s *GrokCapabilityProbeService) tlsProfile(account *Account) *tlsfingerprint.Profile {
	if s == nil || s.tlsFPProfileService == nil {
		return nil
	}
	return s.tlsFPProfileService.ResolveTLSProfile(account)
}

func buildGrokCapabilitySyncSnapshot(account *Account, tier grok.Tier) map[string]any {
	if account == nil {
		return nil
	}

	current := cloneAnyMap(grokNestedMap(account.grokExtraMap()["capabilities"]))
	if !shouldSeedGrokTierCapabilities(account, tier) {
		return current
	}

	baseline := buildGrokTierCapabilityBaseline(account, tier)
	if len(baseline) == 0 {
		return current
	}
	return baseline
}

func shouldSeedGrokTierCapabilities(account *Account, tier grok.Tier) bool {
	if account == nil || tier == grok.TierUnknown {
		return false
	}

	state := account.grokCapabilities()
	if !state.hasModelSignal && !state.hasOperationSignal {
		return true
	}
	if grokHasExplicitNegativeCapabilitySignal(state) {
		return false
	}

	enabled := grokEnabledCapabilities(state)
	if len(enabled) != 1 || enabled[0] != grok.CapabilityChat {
		return false
	}
	if !state.hasModelSignal {
		return true
	}
	if len(state.models) != 1 {
		return false
	}

	for modelID := range state.models {
		spec, ok := grok.LookupModelSpec(modelID)
		return ok && spec.Capability == grok.CapabilityChat
	}
	return false
}

func buildGrokTierCapabilityBaseline(account *Account, tier grok.Tier) map[string]any {
	if account == nil || tier == grok.TierUnknown {
		return nil
	}

	state := account.grokCapabilities()
	modelIDs := make([]string, 0, len(grok.Specs()))
	enabledSet := make(map[grok.Capability]struct{})
	for _, spec := range grok.Specs() {
		if !grokAccountTypeAllowed(account, spec) {
			continue
		}
		if grokTierRank(tier) < grokTierRank(spec.RequiredTier) {
			continue
		}
		if !grokAccountMatchesModelMapping(account, spec.ID) {
			continue
		}
		if allowed, ok := state.operations[spec.Capability]; ok && !allowed {
			continue
		}
		modelIDs = append(modelIDs, spec.ID)
		if spec.Capability != "" {
			enabledSet[spec.Capability] = struct{}{}
		}
	}

	if len(modelIDs) == 0 && len(enabledSet) == 0 {
		return nil
	}

	sort.Strings(modelIDs)
	enabled := make([]string, 0, len(enabledSet))
	for capability := range enabledSet {
		enabled = append(enabled, string(capability))
	}
	sort.Strings(enabled)

	capabilities := make(map[string]any, 2+len(state.operations))
	capabilities["models"] = modelIDs
	if len(enabled) > 0 {
		capabilities["operations"] = enabled
	}
	for capability, allowed := range state.operations {
		if !allowed {
			capabilities[string(capability)] = false
		}
	}
	return capabilities
}

func grokHasExplicitNegativeCapabilitySignal(state grokCapabilityState) bool {
	for _, allowed := range state.operations {
		if !allowed {
			return true
		}
	}
	return false
}

func grokEnabledCapabilities(state grokCapabilityState) []grok.Capability {
	enabled := make([]grok.Capability, 0, len(state.operations))
	for capability, allowed := range state.operations {
		if allowed {
			enabled = append(enabled, capability)
		}
	}
	sort.Slice(enabled, func(i, j int) bool {
		return strings.Compare(string(enabled[i]), string(enabled[j])) < 0
	})
	return enabled
}
