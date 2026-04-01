package service

import "strings"

// resolveOpenAIForwardModel determines the upstream model for OpenAI-compatible
// forwarding. Group-level default mapping only applies when the account itself
// did not match any explicit model_mapping rule.
func resolveOpenAIForwardModel(account *Account, requestedModel, defaultMappedModel string) string {
	if account == nil {
		if defaultMappedModel != "" {
			return defaultMappedModel
		}
		return requestedModel
	}

	mappedModel, matched := account.ResolveMappedModel(requestedModel)
	if !matched && defaultMappedModel != "" {
		return defaultMappedModel
	}
	return mappedModel
}

// resolveOpenAIUpstreamModel normalizes models for the OpenAI upstream while
// preserving bare spark passthroughs that have distinct routing semantics.
func resolveOpenAIUpstreamModel(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return normalizeCodexModel(trimmed)
	}

	modelID := trimmed
	if strings.Contains(modelID, "/") {
		parts := strings.Split(modelID, "/")
		modelID = parts[len(parts)-1]
	}
	modelID = strings.TrimSpace(modelID)
	if isBareCodexSparkModel(modelID) {
		return "gpt-5.3-codex-spark"
	}

	return normalizeCodexModel(modelID)
}

func isBareCodexSparkModel(model string) bool {
	parts := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(model)), func(r rune) bool {
		switch r {
		case '-', '_', ' ':
			return true
		default:
			return false
		}
	})
	if len(parts) != 4 {
		return false
	}
	return parts[0] == "gpt" && parts[1] == "5.3" && parts[2] == "codex" && parts[3] == "spark"
}
