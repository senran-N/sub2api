package service

import (
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/apicompat"
)

func NormalizeOpenAICompatRequestedModel(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return ""
	}

	normalized, _, _, ok := splitOpenAICompatReasoningModel(trimmed)
	if !ok || normalized == "" {
		return trimmed
	}
	return normalized
}

func applyOpenAICompatModelNormalization(req *apicompat.AnthropicRequest) {
	if req == nil {
		return
	}

	originalModel := strings.TrimSpace(req.Model)
	if originalModel == "" {
		return
	}

	normalizedModel, baseModel, derivedEffort, hasReasoningSuffix := splitOpenAICompatReasoningModel(originalModel)
	if hasReasoningSuffix && normalizedModel != "" {
		req.Model = normalizedModel
	}

	if req.OutputConfig != nil && strings.TrimSpace(req.OutputConfig.Effort) != "" {
		if hasReasoningSuffix && baseModel != "" {
			req.Model = baseModel
		}
		return
	}

	claudeEffort := openAIReasoningEffortToClaudeOutputEffort(derivedEffort)
	if claudeEffort == "" {
		return
	}

	if req.OutputConfig == nil {
		req.OutputConfig = &apicompat.AnthropicOutputConfig{}
	}
	req.OutputConfig.Effort = claudeEffort
}

func splitOpenAICompatReasoningModel(model string) (normalizedModel string, baseModel string, reasoningEffort string, ok bool) {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return "", "", "", false
	}

	modelID := trimmed
	if strings.Contains(modelID, "/") {
		parts := strings.Split(modelID, "/")
		modelID = parts[len(parts)-1]
	}
	modelID = strings.TrimSpace(modelID)
	lowerModelID := strings.ToLower(modelID)
	if !strings.HasPrefix(lowerModelID, "gpt-") && !strings.HasPrefix(lowerModelID, "gpt ") {
		return trimmed, "", "", false
	}

	parts := strings.FieldsFunc(strings.ToLower(modelID), func(r rune) bool {
		switch r {
		case '-', '_', ' ':
			return true
		default:
			return false
		}
	})
	if len(parts) == 0 {
		return trimmed, "", "", false
	}

	last := strings.NewReplacer("-", "", "_", "", " ", "").Replace(parts[len(parts)-1])
	switch last {
	case "none", "minimal":
	case "low", "medium", "high":
		reasoningEffort = last
	case "xhigh", "extrahigh":
		reasoningEffort = "xhigh"
	default:
		return trimmed, "", "", false
	}

	baseModel = normalizeCodexModel(modelID)
	normalizedModel = normalizeCodexUpstreamModel(modelID)
	if normalizedModel == "" {
		normalizedModel = baseModel
	}
	if normalizedModel == baseModel {
		normalizedModel = appendOpenAICompatReasoningSuffix(baseModel, reasoningEffort)
	}

	return normalizedModel, baseModel, reasoningEffort, true
}

func openAIReasoningEffortToClaudeOutputEffort(effort string) string {
	switch strings.TrimSpace(effort) {
	case "low", "medium", "high":
		return effort
	case "xhigh":
		return "max"
	default:
		return ""
	}
}

func appendOpenAICompatReasoningSuffix(model string, effort string) string {
	base := strings.TrimSpace(model)
	if base == "" {
		return ""
	}

	switch strings.TrimSpace(effort) {
	case "none":
		return base + "-none"
	case "low":
		return base + "-low"
	case "medium":
		return base + "-medium"
	case "high":
		return base + "-high"
	case "xhigh":
		return base + "-xhigh"
	default:
		return base
	}
}
