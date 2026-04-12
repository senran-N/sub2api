package service

import "strings"

// resolveOpenAIForwardModel determines the upstream model for OpenAI-compatible
// forwarding. Group-level default mapping only applies when the account itself
// did not match any explicit model_mapping rule.
func resolveOpenAIForwardModel(account *Account, requestedModel, defaultMappedModel string) string {
	applyReasoningVariant := func(target string) string {
		normalizedTarget := strings.TrimSpace(target)
		if normalizedTarget == "" {
			return normalizedTarget
		}

		_, _, effort, ok := splitOpenAICompatReasoningModel(requestedModel)
		if !ok || effort == "" {
			return normalizedTarget
		}
		return appendOpenAICompatReasoningSuffix(normalizeCodexUpstreamModel(normalizedTarget), effort)
	}

	if account == nil {
		if defaultMappedModel != "" {
			return applyReasoningVariant(defaultMappedModel)
		}
		return requestedModel
	}

	mappedModel, matched := account.ResolveMappedModel(requestedModel)
	if !matched {
		_, baseModel, _, ok := splitOpenAICompatReasoningModel(requestedModel)
		if ok && baseModel != "" {
			if baseMappedModel, baseMatched := account.ResolveMappedModel(baseModel); baseMatched {
				return applyReasoningVariant(baseMappedModel)
			}
		}
	}
	if !matched && defaultMappedModel != "" {
		return applyReasoningVariant(defaultMappedModel)
	}
	return applyReasoningVariant(mappedModel)
}
