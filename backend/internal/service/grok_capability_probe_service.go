package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	grokCapabilityProbeTimeout         = 20 * time.Second
	grokCapabilityTierBootstrapModelID = "grok-4.20-expert"
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
	settingSvc          *SettingService
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
	settingSvc *SettingService,
) *GrokCapabilityProbeService {
	return &GrokCapabilityProbeService{
		accountRepo:         accountRepo,
		stateSvc:            stateSvc,
		httpUpstream:        httpUpstream,
		cfg:                 cfg,
		tlsFPProfileService: tlsFPProfileService,
		settingSvc:          settingSvc,
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
	settingSvc *SettingService,
	lifecycle *LifecycleRegistry,
) *GrokCapabilityProbeService {
	svc := NewGrokCapabilityProbeService(accountRepo, stateSvc, httpUpstream, cfg, tlsFPProfileService, settingSvc)
	if err := svc.ProbeNow(context.Background()); err != nil {
		logger.LegacyPrintf("service.grok_capability_probe", "Warning: startup probe failed: %v", err)
	}
	return manageStartStopLifecycle(lifecycle, "GrokCapabilityProbeService", svc)
}

func (s *GrokCapabilityProbeService) Start() {
	if s == nil || s.accountRepo == nil || s.stateSvc == nil || s.httpUpstream == nil {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			timer := time.NewTimer(s.currentInterval(context.Background()))
			select {
			case <-timer.C:
				ctx, cancel := context.WithTimeout(context.Background(), grokCapabilityProbeTimeout)
				err := s.ProbeNow(ctx)
				cancel()
				if err != nil {
					logger.LegacyPrintf("service.grok_capability_probe", "Warning: periodic probe failed: %v", err)
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

	accounts, err := listGrokBackgroundAccounts(ctx, s.accountRepo)
	if err != nil {
		return err
	}

	now := s.now().UTC()
	runtimeSettings := s.currentRuntimeSettings(ctx)
	return runGrokParallelAccounts(accounts, runtimeSettings.CapabilityProbeWorkers(), func(account *Account) error {
		interval := s.currentIntervalForAccount(account, runtimeSettings)
		if !s.shouldProbeAccount(account, now, interval) {
			return nil
		}
		if err := s.probeAccount(ctx, account); err != nil {
			logger.LegacyPrintf("service.grok_capability_probe", "Warning: account %d probe failed: %v", account.ID, err)
			return err
		}
		return nil
	})
}

func (s *GrokCapabilityProbeService) ProbeAccount(ctx context.Context, account *Account) error {
	if s == nil || account == nil {
		return nil
	}
	return s.probeAccount(ctx, account)
}

func (s *GrokCapabilityProbeService) shouldProbeAccount(account *Account, now time.Time, interval time.Duration) bool {
	if account == nil || NormalizeCompatibleGatewayPlatform(account.Platform) != PlatformGrok {
		return false
	}
	if account.Type != AccountTypeAPIKey && account.Type != AccountTypeUpstream && account.Type != AccountTypeSession {
		return false
	}
	if !account.IsMaintenanceSchedulable() {
		return false
	}
	switch account.Type {
	case AccountTypeAPIKey, AccountTypeUpstream:
		if strings.TrimSpace(account.GetOpenAIApiKey()) == "" {
			return false
		}
	case AccountTypeSession:
		if strings.TrimSpace(account.GetGrokSessionToken()) == "" {
			return false
		}
	}

	capabilityState := account.grokCapabilities()
	syncState := account.grokSyncState()
	if !capabilityState.hasModelSignal && !capabilityState.hasOperationSignal {
		return true
	}
	if syncState.LastProbeAt == nil {
		return true
	}
	return now.Sub(*syncState.LastProbeAt) >= interval
}

func (s *GrokCapabilityProbeService) currentInterval(ctx context.Context) time.Duration {
	return s.loopInterval(s.currentRuntimeSettings(ctx))
}

func (s *GrokCapabilityProbeService) currentRuntimeSettings(ctx context.Context) GrokRuntimeSettings {
	if s == nil || s.settingSvc == nil {
		return DefaultGrokRuntimeSettings()
	}
	return s.settingSvc.GetGrokRuntimeSettings(ctx)
}

func (s *GrokCapabilityProbeService) currentIntervalForAccount(
	account *Account,
	settings GrokRuntimeSettings,
) time.Duration {
	if account != nil && account.Type == AccountTypeSession {
		return settings.SessionValidityCheckInterval()
	}
	return settings.CapabilityProbeInterval()
}

func (s *GrokCapabilityProbeService) loopInterval(settings GrokRuntimeSettings) time.Duration {
	probeInterval := settings.CapabilityProbeInterval()
	sessionInterval := settings.SessionValidityCheckInterval()
	if sessionInterval < probeInterval {
		return sessionInterval
	}
	return probeInterval
}

func (s *GrokCapabilityProbeService) probeAccount(ctx context.Context, account *Account) error {
	baseURLValidator := s.validateUpstreamBaseURL
	if account != nil && account.Type == AccountTypeSession {
		// Session transport uses the provider-owned Grok web origin, so probe behavior
		// should match runtime selection instead of compatible-upstream allowlists.
		baseURLValidator = nil
	}
	target, err := resolveGrokTransportTargetWithSettings(
		account,
		baseURLValidator,
		s.currentRuntimeSettings(ctx),
	)
	if err != nil {
		s.stateSvc.PersistProbeResult(ctx, account, grok.DefaultTestModel, nil, err)
		return err
	}

	reqCtx := ctx
	if reqCtx == nil {
		reqCtx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(reqCtx, grokCapabilityProbeTimeout)
	defer cancel()

	candidates := grokCapabilityProbeModelCandidates(account, "")
	if len(candidates) == 0 {
		candidates = []string{grok.DefaultTestModel}
	}

	var (
		lastModel string
		lastResp  *http.Response
		lastErr   error
	)
	for _, modelID := range candidates {
		lastModel = modelID
		resp, probeErr := s.executeProbeAttempt(reqCtx, account, target, modelID)
		if probeErr == nil {
			s.stateSvc.PersistProbeResult(ctx, account, modelID, resp, nil)
			return nil
		}
		lastResp = resp
		lastErr = probeErr
	}

	if lastModel == "" {
		lastModel = grok.DefaultTestModel
	}
	s.stateSvc.PersistProbeResult(ctx, account, lastModel, lastResp, lastErr)
	return lastErr
}

func (s *GrokCapabilityProbeService) executeProbeAttempt(
	ctx context.Context,
	account *Account,
	target grokTransportTarget,
	modelID string,
) (*http.Response, error) {
	req, err := s.buildProbeRequest(ctx, account, target, modelID)
	if err != nil {
		return nil, fmt.Errorf("create grok capability probe request: %w", err)
	}

	resp, err := s.httpUpstream.DoWithTLS(
		req,
		accountTestProxyURL(account),
		account.ID,
		account.Concurrency,
		s.tlsProfile(account),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	summary := &http.Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header.Clone(),
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body := grokReadProbeErrorBody(resp)
		errSummary := grokSummarizeProbeHTTPError(resp, body)
		return summary, fmt.Errorf("%s", errSummary.Message)
	}
	return summary, nil
}

func (s *GrokCapabilityProbeService) buildProbeRequest(
	ctx context.Context,
	account *Account,
	target grokTransportTarget,
	modelID string,
) (*http.Request, error) {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		modelID = grok.DefaultTestModel
	}
	switch target.Kind {
	case grokTransportKindCompatible:
		payload := createCompatibleGatewayTestPayload(modelID, false, "")
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal grok capability probe payload: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		target.Apply(req)
		return req, nil
	case grokTransportKindSession:
		payload, err := createGrokSessionTestPayload(modelID, "")
		if err != nil {
			return nil, fmt.Errorf("build grok session capability probe payload: %w", err)
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal grok session capability probe payload: %w", err)
		}
		return newGrokSessionJSONRequest(ctx, http.MethodPost, target, payloadBytes, grokSessionProbeAcceptHeader)
	default:
		if account == nil {
			return nil, fmt.Errorf("unsupported grok transport kind: %s", target.Kind)
		}
		return nil, fmt.Errorf("unsupported grok transport kind for account %d: %s", account.ID, target.Kind)
	}
}

func grokCapabilityProbeModelCandidates(account *Account, requestedModel string) []string {
	candidates := make([]string, 0, 2)
	appendCandidate := func(modelID string) {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			return
		}
		if canonical := grok.ResolveCanonicalModelID(modelID); canonical != "" {
			modelID = canonical
		}
		for _, existing := range candidates {
			if existing == modelID {
				return
			}
		}
		candidates = append(candidates, modelID)
	}

	if explicit := strings.TrimSpace(requestedModel); explicit != "" {
		appendCandidate(explicit)
		return candidates
	}

	if account != nil {
		switch account.GrokTierState().Normalized {
		case grok.TierUnknown, grok.TierSuper, grok.TierHeavy:
			appendCandidate(grokCapabilityTierBootstrapModelID)
		}
	}
	appendCandidate(grok.DefaultTestModel)
	return candidates
}

func (s *GrokCapabilityProbeService) validateUpstreamBaseURL(raw string) (string, error) {
	return validateCompatibleUpstreamBaseURL(s.cfg, raw)
}

func (s *GrokCapabilityProbeService) tlsProfile(account *Account) *tlsfingerprint.Profile {
	if s == nil {
		return nil
	}
	return resolveGrokTLSProfile(account, s.tlsFPProfileService)
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
	return buildGrokTierCapabilityBaselineFromState(account, account.grokCapabilities(), tier)
}

func buildGrokTierCapabilityBaselineFromState(account *Account, state grokCapabilityState, tier grok.Tier) map[string]any {
	if account == nil || tier == grok.TierUnknown {
		return nil
	}

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
