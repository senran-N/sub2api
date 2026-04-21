package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

type GrokAccountSelector struct{}

var defaultGrokAccountSelector = GrokAccountSelector{}

type grokAccountScoreContext struct {
	minPriority int
	maxPriority int
	maxWaiting  int
}

type grokAccountCandidateScore struct {
	account  *Account
	loadInfo *AccountLoadInfo
	score    float64
}

func newGrokAccountScoreContext(firstPriority int) grokAccountScoreContext {
	return grokAccountScoreContext{
		minPriority: firstPriority,
		maxPriority: firstPriority,
		maxWaiting:  1,
	}
}

func (ctx *grokAccountScoreContext) observe(account *Account, loadInfo *AccountLoadInfo) {
	if account == nil {
		return
	}
	if account.Priority < ctx.minPriority {
		ctx.minPriority = account.Priority
	}
	if account.Priority > ctx.maxPriority {
		ctx.maxPriority = account.Priority
	}
	if loadInfo != nil && loadInfo.WaitingCount > ctx.maxWaiting {
		ctx.maxWaiting = loadInfo.WaitingCount
	}
}

func (GrokAccountSelector) IsSelectionPlatformAccount(account *Account) bool {
	if account == nil {
		return false
	}
	return NormalizeCompatibleGatewayPlatform(account.Platform) == PlatformGrok
}

func (selector GrokAccountSelector) IsBaseEligible(account *Account) bool {
	return selector.IsBaseEligibleWithContext(context.TODO(), account)
}

func (selector GrokAccountSelector) IsBaseEligibleWithContext(ctx context.Context, account *Account) bool {
	if !selector.IsSelectionPlatformAccount(account) || !account.IsSchedulable() {
		return false
	}
	if account.SupportsCompatibleGatewaySharedRuntime() {
		return true
	}
	return account.Type == AccountTypeSession && AllowsGrokSessionProviderRuntime(ctx)
}

func (selector GrokAccountSelector) IsModelEligible(account *Account, requestedModel string) bool {
	return isGrokAccountModelEligible(account, requestedModel)
}

func (selector GrokAccountSelector) IsModelAvailableWithContext(
	ctx context.Context,
	account *Account,
	requestedModel string,
) bool {
	if !selector.IsBaseEligibleWithContext(ctx, account) {
		return false
	}
	if requestedModel != "" && !selector.IsModelEligible(account, requestedModel) {
		return false
	}

	runtimeState := account.grokRuntimeSelectionState()
	return !grokRuntimeModelSelectionBlocked(runtimeState, requestedModel, time.Now().UTC())
}

func (selector GrokAccountSelector) IsRuntimeEligible(account *Account, requestedModel string) bool {
	return selector.IsRuntimeEligibleWithContext(context.TODO(), account, requestedModel)
}

func (selector GrokAccountSelector) IsRuntimeEligibleWithContext(ctx context.Context, account *Account, requestedModel string) bool {
	if !selector.IsModelAvailableWithContext(ctx, account, requestedModel) {
		return false
	}
	if oauthSelectionCredentialIssue(account) != "" {
		return false
	}
	if grokRuntimeSelectionBlocked(account.grokRuntimeSelectionState(), requestedModel, time.Now().UTC()) {
		return false
	}
	return true
}

func (selector GrokAccountSelector) RequestedModelAvailable(accounts []Account, requestedModel string) bool {
	return selector.RequestedModelAvailableWithContext(context.TODO(), accounts, requestedModel)
}

func (selector GrokAccountSelector) RequestedModelAvailableWithContext(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
) bool {
	if requestedModel == "" {
		return true
	}
	for i := range accounts {
		if selector.IsModelAvailableWithContext(ctx, &accounts[i], requestedModel) {
			return true
		}
	}
	return false
}

func (selector GrokAccountSelector) FilterSchedulableCandidates(
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
) []*Account {
	return selector.FilterSchedulableCandidatesWithContext(context.TODO(), accounts, requestedModel, excludedIDs)
}

func (selector GrokAccountSelector) FilterSchedulableCandidatesWithContext(
	ctx context.Context,
	accounts []Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
) []*Account {
	candidates := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		if isOpenAIAccountExcluded(excludedIDs, account.ID) {
			continue
		}
		if !selector.IsRuntimeEligibleWithContext(ctx, account, requestedModel) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

func (selector GrokAccountSelector) FilterSchedulableAccountPointers(
	accounts []*Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
) []*Account {
	return selector.FilterSchedulableAccountPointersWithContext(context.TODO(), accounts, requestedModel, excludedIDs)
}

func (selector GrokAccountSelector) FilterSchedulableAccountPointersWithContext(
	ctx context.Context,
	accounts []*Account,
	requestedModel string,
	excludedIDs map[int64]struct{},
) []*Account {
	candidates := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		if isOpenAIAccountExcluded(excludedIDs, account.ID) {
			continue
		}
		if !selector.IsRuntimeEligibleWithContext(ctx, account, requestedModel) {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates
}

func (selector GrokAccountSelector) SelectBestCandidateWithContext(
	ctx context.Context,
	candidates []*Account,
	requestedModel string,
	loadMap map[int64]*AccountLoadInfo,
) *Account {
	if len(candidates) == 0 {
		return nil
	}

	scoreCtx := newGrokAccountScoreContext(candidates[0].Priority)
	scored := make([]grokAccountCandidateScore, 0, len(candidates))
	for _, account := range candidates {
		if account == nil {
			continue
		}
		loadInfo := loadMap[account.ID]
		if loadInfo == nil {
			loadInfo = &AccountLoadInfo{AccountID: account.ID}
		}
		scoreCtx.observe(account, loadInfo)
		scored = append(scored, grokAccountCandidateScore{
			account:  account,
			loadInfo: loadInfo,
		})
	}
	if len(scored) == 0 {
		return nil
	}

	selector.scoreCandidates(scored, scoreCtx, requestedModel)

	best := &scored[0]
	for i := 1; i < len(scored); i++ {
		if grokCandidateBetter(scored[i], *best) {
			best = &scored[i]
		}
	}
	return best.account
}

func (selector GrokAccountSelector) scoreCandidates(
	candidates []grokAccountCandidateScore,
	scoreCtx grokAccountScoreContext,
	requestedModel string,
) {
	spec, hasSpec := grok.LookupModelSpec(requestedModel)
	quotaWindowName := grokQuotaWindowForModel(requestedModel)
	now := time.Now().UTC()

	for i := range candidates {
		item := &candidates[i]
		account := item.account
		if account == nil {
			continue
		}

		priorityFactor := 1.0
		if scoreCtx.maxPriority > scoreCtx.minPriority {
			priorityFactor = 1 - float64(account.Priority-scoreCtx.minPriority)/float64(scoreCtx.maxPriority-scoreCtx.minPriority)
		}

		loadRate := 0.0
		waitingCount := 0
		if item.loadInfo != nil {
			loadRate = float64(item.loadInfo.LoadRate)
			waitingCount = item.loadInfo.WaitingCount
		}
		loadFactor := 1 - clamp01(loadRate/100.0)
		occupancyFactor := 1 - clamp01(float64(waitingCount)/float64(scoreCtx.maxWaiting))

		tierState := account.GrokTierState()
		capabilityState := account.grokCapabilities()
		quotaState := account.grokQuotaWindow(quotaWindowName)
		syncState := account.grokSyncState()
		runtimeState := account.grokRuntimeSelectionState()

		tierFactor := grokTierMatchQualityFactor(tierState, capabilityState, requestedModel, spec, hasSpec)
		capabilityFactor := grokCapabilitySignalFactor(capabilityState, requestedModel, spec, hasSpec)
		quotaFactor := grokQuotaHeadroomFactor(quotaState)
		freshnessFactor := grokQuotaFreshnessFactor(syncState, now)
		failureFactor := grokRecentFailureFactor(runtimeState, requestedModel, now)

		baseScore := 2.4*priorityFactor +
			1.1*loadFactor +
			0.7*occupancyFactor +
			1.3*tierFactor +
			1.0*capabilityFactor +
			1.0*quotaFactor +
			0.6*freshnessFactor
		item.score = baseScore * failureFactor
	}
}

func grokTierMatchQualityFactor(
	tierState GrokTierState,
	capabilityState grokCapabilityState,
	requestedModel string,
	spec grok.ModelSpec,
	hasSpec bool,
) float64 {
	if !hasSpec || spec.RequiredTier == "" || spec.RequiredTier == grok.TierUnknown {
		if tierState.Normalized != grok.TierUnknown {
			return 0.8
		}
		if capabilityState.hasModelSignal {
			return 0.7
		}
		if capabilityState.hasOperationSignal {
			return 0.55
		}
		return 0.4
	}

	if tierState.Normalized == grok.TierUnknown {
		canonicalRequested := grok.ResolveCanonicalModelID(requestedModel)
		if capabilityState.hasModelSignal {
			if _, ok := capabilityState.models[canonicalRequested]; ok {
				return 0.78
			}
			if _, ok := capabilityState.models[spec.ID]; ok {
				return 0.78
			}
		}
		if capabilityState.hasOperationSignal && capabilityState.operations[spec.Capability] {
			return 0.58
		}
		return 0.35
	}

	diff := grokTierRank(tierState.Normalized) - grokTierRank(spec.RequiredTier)
	switch {
	case diff <= 0:
		return 1.0
	case diff == 1:
		return 0.82
	default:
		return 0.68
	}
}

func grokCapabilitySignalFactor(
	capabilityState grokCapabilityState,
	requestedModel string,
	spec grok.ModelSpec,
	hasSpec bool,
) float64 {
	if !hasSpec {
		if capabilityState.hasModelSignal || capabilityState.hasOperationSignal {
			return 0.75
		}
		return 0.45
	}

	canonicalRequested := grok.ResolveCanonicalModelID(requestedModel)
	if capabilityState.hasModelSignal {
		if _, ok := capabilityState.models[canonicalRequested]; ok {
			return 1.0
		}
		if _, ok := capabilityState.models[spec.ID]; ok {
			return 0.95
		}
		return 0.2
	}

	if capabilityState.hasOperationSignal {
		if allowed, ok := capabilityState.operations[spec.Capability]; ok && allowed {
			return 0.78
		}
		return 0.2
	}

	return 0.45
}

func grokQuotaHeadroomFactor(quotaState grokQuotaWindowState) float64 {
	if strings.TrimSpace(quotaState.Name) == "" {
		return 0.45
	}
	if !quotaState.HasSignal {
		return 0.4
	}
	if quotaState.Remaining <= 0 {
		return 0.05
	}
	if quotaState.Total <= 0 {
		return 0.6
	}
	return 0.2 + 0.8*clamp01(float64(quotaState.Remaining)/float64(quotaState.Total))
}

func grokQuotaFreshnessFactor(syncState grokSyncState, now time.Time) float64 {
	freshest := syncState.LastSyncAt
	if freshest == nil {
		freshest = syncState.LastProbeOKAt
	}
	if freshest == nil {
		freshest = syncState.LastProbeAt
	}
	if freshest == nil {
		return 0.35
	}

	age := now.Sub(*freshest)
	switch {
	case age <= 15*time.Minute:
		return 1.0
	case age <= time.Hour:
		return 0.9
	case age <= 6*time.Hour:
		return 0.75
	case age <= 24*time.Hour:
		return 0.55
	case age <= 72*time.Hour:
		return 0.4
	default:
		return 0.25
	}
}

func grokRecentFailureFactor(runtimeState grokRuntimeSelectionState, requestedModel string, now time.Time) float64 {
	latestFailure := runtimeState.LastFailAt
	if latestFailure == nil || (runtimeState.LastFailoverAt != nil && runtimeState.LastFailoverAt.After(*latestFailure)) {
		latestFailure = runtimeState.LastFailoverAt
	}
	if latestFailure == nil {
		return 1.0
	}
	if runtimeState.LastUseAt != nil && !runtimeState.LastUseAt.Before(*latestFailure) {
		return 1.0
	}

	age := now.Sub(*latestFailure)
	factor := 0.95
	switch {
	case age <= 5*time.Minute:
		factor = 0.15
	case age <= 30*time.Minute:
		factor = 0.4
	case age <= 2*time.Hour:
		factor = 0.65
	case age <= 12*time.Hour:
		factor = 0.82
	}

	switch runtimeState.LastFailClass {
	case grokRuntimeErrorClassAuth:
		factor *= 0.55
	case grokRuntimeErrorClassRateLimited:
		factor *= 0.72
	case grokRuntimeErrorClassModelUnsupported:
		if grokRuntimePenaltyAppliesToModel(runtimeState.CooldownModel, requestedModel) {
			factor *= 0.65
		}
	case grokRuntimeErrorClassUpstream:
		factor *= 0.88
	case grokRuntimeErrorClassTransport:
		factor *= 0.92
	case grokRuntimeErrorClassInvalidRequest:
		factor *= 0.97
	default:
		if runtimeState.LastFailStatusCode == 429 {
			factor *= 0.82
		} else if runtimeState.LastFailStatusCode >= 500 {
			factor *= 0.9
		}
	}
	return clamp01(factor)
}

func grokRuntimeSelectionBlocked(runtimeState grokRuntimeSelectionState, requestedModel string, now time.Time) bool {
	if runtimeState.CooldownUntil == nil || !now.Before(*runtimeState.CooldownUntil) {
		return false
	}

	switch runtimeState.CooldownScope {
	case grokRuntimePenaltyScopeAccount:
		return true
	case grokRuntimePenaltyScopeModel:
		return grokRuntimePenaltyAppliesToModel(runtimeState.CooldownModel, requestedModel)
	default:
		return false
	}
}

func grokRuntimeModelSelectionBlocked(runtimeState grokRuntimeSelectionState, requestedModel string, now time.Time) bool {
	if runtimeState.CooldownUntil == nil || !now.Before(*runtimeState.CooldownUntil) {
		return false
	}
	return runtimeState.CooldownScope == grokRuntimePenaltyScopeModel &&
		grokRuntimePenaltyAppliesToModel(runtimeState.CooldownModel, requestedModel)
}

func grokRuntimePenaltyAppliesToModel(cooldownModel string, requestedModel string) bool {
	if strings.TrimSpace(requestedModel) == "" {
		return false
	}
	if strings.TrimSpace(cooldownModel) == "" {
		return true
	}
	return grok.ResolveCanonicalModelID(cooldownModel) == grok.ResolveCanonicalModelID(requestedModel)
}

func grokCandidateBetter(left grokAccountCandidateScore, right grokAccountCandidateScore) bool {
	if diff := left.score - right.score; math.Abs(diff) > 1e-9 {
		return diff > 0
	}
	if left.account.Priority != right.account.Priority {
		return left.account.Priority < right.account.Priority
	}
	if left.loadInfo.LoadRate != right.loadInfo.LoadRate {
		return left.loadInfo.LoadRate < right.loadInfo.LoadRate
	}
	if left.loadInfo.WaitingCount != right.loadInfo.WaitingCount {
		return left.loadInfo.WaitingCount < right.loadInfo.WaitingCount
	}
	switch {
	case left.account.LastUsedAt == nil && right.account.LastUsedAt != nil:
		return true
	case left.account.LastUsedAt != nil && right.account.LastUsedAt == nil:
		return false
	case left.account.LastUsedAt != nil && right.account.LastUsedAt != nil:
		if !left.account.LastUsedAt.Equal(*right.account.LastUsedAt) {
			return left.account.LastUsedAt.Before(*right.account.LastUsedAt)
		}
	}
	return left.account.ID < right.account.ID
}
