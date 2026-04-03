package service

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func (s *GatewayService) parseSSEUsage(data string, usage *ClaudeUsage) {
	if usage == nil {
		return
	}

	var event map[string]any
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return
	}

	if patch := s.extractSSEUsagePatch(event); patch != nil {
		mergeSSEUsagePatch(usage, patch)
	}
}

type sseUsagePatch struct {
	inputTokens              int
	hasInputTokens           bool
	outputTokens             int
	hasOutputTokens          bool
	cacheCreationInputTokens int
	hasCacheCreationInput    bool
	cacheReadInputTokens     int
	hasCacheReadInput        bool
	cacheCreation5mTokens    int
	hasCacheCreation5m       bool
	cacheCreation1hTokens    int
	hasCacheCreation1h       bool
}

func (s *GatewayService) extractSSEUsagePatch(event map[string]any) *sseUsagePatch {
	if len(event) == 0 {
		return nil
	}

	eventType, _ := event["type"].(string)
	switch eventType {
	case "message_start":
		message, _ := event["message"].(map[string]any)
		usageObject, _ := message["usage"].(map[string]any)
		if len(usageObject) == 0 {
			return nil
		}
		return buildMessageStartUsagePatch(usageObject)

	case "message_delta":
		usageObject, _ := event["usage"].(map[string]any)
		if len(usageObject) == 0 {
			return nil
		}
		return buildMessageDeltaUsagePatch(usageObject)
	}

	return nil
}

func buildMessageStartUsagePatch(usageObject map[string]any) *sseUsagePatch {
	patch := &sseUsagePatch{hasInputTokens: true, hasCacheCreationInput: true, hasCacheReadInput: true}
	if value, ok := parseSSEUsageInt(usageObject["input_tokens"]); ok {
		patch.inputTokens = value
	}
	if value, ok := parseSSEUsageInt(usageObject["cache_creation_input_tokens"]); ok {
		patch.cacheCreationInputTokens = value
	}
	if value, ok := parseSSEUsageInt(usageObject["cache_read_input_tokens"]); ok {
		patch.cacheReadInputTokens = value
	}
	applyNestedCacheCreationPatch(patch, usageObject, false)
	return patch
}

func buildMessageDeltaUsagePatch(usageObject map[string]any) *sseUsagePatch {
	patch := &sseUsagePatch{}
	if value, ok := parseSSEUsageInt(usageObject["input_tokens"]); ok && value > 0 {
		patch.inputTokens = value
		patch.hasInputTokens = true
	}
	if value, ok := parseSSEUsageInt(usageObject["output_tokens"]); ok && value > 0 {
		patch.outputTokens = value
		patch.hasOutputTokens = true
	}
	if value, ok := parseSSEUsageInt(usageObject["cache_creation_input_tokens"]); ok && value > 0 {
		patch.cacheCreationInputTokens = value
		patch.hasCacheCreationInput = true
	}
	if value, ok := parseSSEUsageInt(usageObject["cache_read_input_tokens"]); ok && value > 0 {
		patch.cacheReadInputTokens = value
		patch.hasCacheReadInput = true
	}
	applyNestedCacheCreationPatch(patch, usageObject, true)
	return patch
}

func applyNestedCacheCreationPatch(patch *sseUsagePatch, usageObject map[string]any, requirePositive bool) {
	cacheCreation, _ := usageObject["cache_creation"].(map[string]any)
	if len(cacheCreation) == 0 {
		return
	}

	if value, exists := parseSSEUsageInt(cacheCreation["ephemeral_5m_input_tokens"]); exists && (!requirePositive || value > 0) {
		patch.cacheCreation5mTokens = value
		patch.hasCacheCreation5m = true
	}
	if value, exists := parseSSEUsageInt(cacheCreation["ephemeral_1h_input_tokens"]); exists && (!requirePositive || value > 0) {
		patch.cacheCreation1hTokens = value
		patch.hasCacheCreation1h = true
	}
}

func mergeSSEUsagePatch(usage *ClaudeUsage, patch *sseUsagePatch) {
	if usage == nil || patch == nil {
		return
	}

	if patch.hasInputTokens {
		usage.InputTokens = patch.inputTokens
	}
	if patch.hasCacheCreationInput {
		usage.CacheCreationInputTokens = patch.cacheCreationInputTokens
	}
	if patch.hasCacheReadInput {
		usage.CacheReadInputTokens = patch.cacheReadInputTokens
	}
	if patch.hasOutputTokens {
		usage.OutputTokens = patch.outputTokens
	}
	if patch.hasCacheCreation5m {
		usage.CacheCreation5mTokens = patch.cacheCreation5mTokens
	}
	if patch.hasCacheCreation1h {
		usage.CacheCreation1hTokens = patch.cacheCreation1hTokens
	}
}

func parseSSEUsageInt(value any) (int, bool) {
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case float32:
		return int(typed), true
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case int32:
		return int(typed), true
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed), true
		}
		if parsed, err := typed.Float64(); err == nil {
			return int(parsed), true
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func applyCacheTTLOverride(usage *ClaudeUsage, target string) bool {
	if usage.CacheCreation5mTokens == 0 && usage.CacheCreation1hTokens == 0 && usage.CacheCreationInputTokens > 0 {
		usage.CacheCreation5mTokens = usage.CacheCreationInputTokens
	}

	total := usage.CacheCreation5mTokens + usage.CacheCreation1hTokens
	if total == 0 {
		return false
	}

	switch target {
	case "1h":
		if usage.CacheCreation1hTokens == total {
			return false
		}
		usage.CacheCreation1hTokens = total
		usage.CacheCreation5mTokens = 0
	default:
		if usage.CacheCreation5mTokens == total {
			return false
		}
		usage.CacheCreation5mTokens = total
		usage.CacheCreation1hTokens = 0
	}
	return true
}

func rewriteCacheCreationJSON(usageObject map[string]any, target string) bool {
	cacheCreation, ok := usageObject["cache_creation"].(map[string]any)
	if !ok {
		return false
	}

	value5m, _ := parseSSEUsageInt(cacheCreation["ephemeral_5m_input_tokens"])
	value1h, _ := parseSSEUsageInt(cacheCreation["ephemeral_1h_input_tokens"])
	total := value5m + value1h
	if total == 0 {
		return false
	}

	switch target {
	case "1h":
		if value1h == total {
			return false
		}
		cacheCreation["ephemeral_1h_input_tokens"] = float64(total)
		cacheCreation["ephemeral_5m_input_tokens"] = float64(0)
	default:
		if value5m == total {
			return false
		}
		cacheCreation["ephemeral_5m_input_tokens"] = float64(total)
		cacheCreation["ephemeral_1h_input_tokens"] = float64(0)
	}
	return true
}

func (s *GatewayService) replaceModelInResponseBody(body []byte, fromModel, toModel string) []byte {
	if model := gjson.GetBytes(body, "model"); model.Exists() && model.Str == fromModel {
		newBody, err := sjson.SetBytes(body, "model", toModel)
		if err != nil {
			return body
		}
		return newBody
	}
	return body
}
