package service

import (
	"sort"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

func grokTierRank(tier grok.Tier) int {
	switch tier {
	case grok.TierBasic:
		return 0
	case grok.TierSuper:
		return 1
	case grok.TierHeavy:
		return 2
	default:
		return -1
	}
}

func grokAccountTypeAllowed(account *Account, spec grok.ModelSpec) bool {
	if account == nil {
		return false
	}
	if len(spec.AllowedAccountTypes) == 0 {
		return true
	}
	accountType := strings.TrimSpace(account.Type)
	for _, allowed := range spec.AllowedAccountTypes {
		if accountType == allowed {
			return true
		}
	}
	return false
}

func grokCapabilityAllowsSpec(account *Account, spec grok.ModelSpec, requestedModel string) bool {
	state := account.grokCapabilities()
	if state.hasModelSignal {
		if _, ok := state.models[grok.ResolveCanonicalModelID(requestedModel)]; ok {
			return true
		}
		if _, ok := state.models[spec.ID]; ok {
			return true
		}
		return false
	}
	if state.hasOperationSignal {
		allowed, ok := state.operations[spec.Capability]
		if !ok {
			return false
		}
		return allowed
	}
	return true
}

func grokTierAllowsSpec(account *Account, spec grok.ModelSpec, requestedModel string) bool {
	required := spec.RequiredTier
	if required == "" || required == grok.TierUnknown {
		return true
	}

	state := account.GrokTierState()
	if grokTierRank(state.Normalized) >= grokTierRank(required) {
		return true
	}

	if state.Normalized != grok.TierUnknown {
		return false
	}

	if capabilityModels := account.GrokCapabilityModelIDs(); len(capabilityModels) > 0 {
		canonicalRequested := grok.ResolveCanonicalModelID(requestedModel)
		for _, supportedModel := range capabilityModels {
			if supportedModel == canonicalRequested || supportedModel == spec.ID {
				return true
			}
		}
		return false
	}

	if account.Type == AccountTypeSession {
		return required == grok.TierBasic
	}

	return true
}

func grokAccountMatchesModelMapping(account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	mapping := account.GetModelMapping()
	if len(mapping) == 0 {
		return true
	}
	if account.IsModelSupported(requestedModel) {
		return true
	}
	canonicalRequested := grok.ResolveCanonicalModelID(requestedModel)
	return canonicalRequested != requestedModel && account.IsModelSupported(canonicalRequested)
}

func grokQuotaWindowForModel(modelID string) string {
	if spec, ok := grok.LookupModelSpec(modelID); ok && strings.TrimSpace(spec.QuotaWindow) != "" {
		return spec.QuotaWindow
	}

	normalized := strings.ToLower(strings.TrimSpace(modelID))
	switch {
	case strings.Contains(normalized, "4.20-heavy"), strings.Contains(normalized, "-heavy"):
		return grok.QuotaWindowHeavy
	case strings.Contains(normalized, "4.20-expert"), strings.Contains(normalized, "-expert"):
		return grok.QuotaWindowExpert
	case strings.Contains(normalized, "4.20-fast"), strings.HasSuffix(normalized, "-fast"):
		return grok.QuotaWindowFast
	}

	return ""
}

func isGrokAccountModelEligible(account *Account, requestedModel string) bool {
	model := strings.TrimSpace(requestedModel)
	if model == "" {
		return true
	}
	if account == nil || account.Platform != PlatformGrok {
		return false
	}

	spec, ok := grok.LookupModelSpec(model)
	if !ok {
		return account.IsModelSupported(model)
	}

	if !grokAccountTypeAllowed(account, spec) {
		return false
	}
	if !grokAccountMatchesModelMapping(account, model) {
		return false
	}
	if !grokCapabilityAllowsSpec(account, spec, model) {
		return false
	}
	return grokTierAllowsSpec(account, spec, model)
}

func GrokAvailableModelIDsForAccount(account *Account) []string {
	if account == nil || account.Platform != PlatformGrok {
		return nil
	}

	mapping := account.GetModelMapping()
	if len(mapping) > 0 {
		models := make([]string, 0, len(mapping))
		for requestedModel := range mapping {
			if isGrokAccountModelEligible(account, requestedModel) {
				models = append(models, requestedModel)
			}
		}
		sort.Strings(models)
		return models
	}

	specs := grok.Specs()
	models := make([]string, 0, len(specs))
	for _, spec := range specs {
		if isGrokAccountModelEligible(account, spec.ID) {
			models = append(models, spec.ID)
		}
	}
	return models
}
