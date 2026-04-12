package service

import "strings"

func ResolveOpenAIForwardDefaultMappedModel(apiKey *APIKey, fallbackModel string) string {
	if fallbackModel = strings.TrimSpace(fallbackModel); fallbackModel != "" {
		return fallbackModel
	}
	if apiKey == nil || apiKey.Group == nil {
		return ""
	}
	return strings.TrimSpace(apiKey.Group.DefaultMappedModel)
}

func resolveMappedModelWithOpenAIReasoningFallback(account *Account, requestedModel string) (mappedModel string, matched bool) {
	if account == nil {
		return strings.TrimSpace(requestedModel), false
	}

	mappedModel, matched = account.ResolveMappedModel(requestedModel)
	if matched {
		return strings.TrimSpace(mappedModel), true
	}

	_, baseModel, _, ok := splitOpenAICompatReasoningModel(requestedModel)
	if !ok || baseModel == "" {
		return strings.TrimSpace(mappedModel), false
	}

	baseMappedModel, baseMatched := account.ResolveMappedModel(baseModel)
	if !baseMatched {
		return strings.TrimSpace(mappedModel), false
	}
	return strings.TrimSpace(baseMappedModel), true
}

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

	mappedModel, matched := resolveMappedModelWithOpenAIReasoningFallback(account, requestedModel)
	if !matched && defaultMappedModel != "" {
		return applyReasoningVariant(defaultMappedModel)
	}
	return applyReasoningVariant(mappedModel)
}
