package service

import (
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

func normalizePlatformAccountExtra(existing map[string]any, incoming map[string]any, platform string, accountType string) map[string]any {
	switch platform {
	case PlatformGrok:
		return normalizeGrokAccountExtra(existing, incoming, accountType)
	default:
		return incoming
	}
}

func normalizeGrokAccountExtra(existing map[string]any, incoming map[string]any, accountType string) map[string]any {
	mergedExtra := mergeAnyMaps(existing, incoming)
	grokExtra := mergeAnyMaps(grokExtraMap(existing), grokExtraMap(incoming))

	if authMode := defaultGrokAuthMode(accountType); authMode != "" {
		if strings.TrimSpace(getStringFromMaps(grokExtra, nil, "auth_mode")) == "" {
			grokExtra["auth_mode"] = authMode
		}
	}

	normalizedTier := grokNormalizeTier(getStringFromMaps(grokNestedMap(grokExtra["tier"]), grokExtra, "normalized", "tier_normalized"))
	if normalizedTier == grok.TierUnknown {
		normalizedTier = grokNormalizeTier(getStringFromMaps(grokExtra, nil, "pool"))
	}
	if normalizedTier == grok.TierUnknown {
		normalizedTier = grokInferTierFromQuotaWindows(grokExtra["quota_windows"])
	}

	tier := map[string]any{
		"normalized": string(normalizedTier),
	}
	if raw := getStringFromMaps(grokNestedMap(grokExtra["tier"]), grokExtra, "raw", "raw_tier", "tier_raw"); raw != "" {
		tier["raw"] = raw
	}
	if source := getStringFromMaps(grokNestedMap(grokExtra["tier"]), grokExtra, "source", "tier_source"); source != "" {
		tier["source"] = source
	}
	if confidence := grokParseFloat(getNestedGrokValue(grokExtra, "tier", "confidence")); confidence > 0 {
		tier["confidence"] = confidence
	} else if confidence := grokParseFloat(grokExtra["tier_confidence"]); confidence > 0 {
		tier["confidence"] = confidence
	}
	grokExtra["tier"] = tier
	delete(grokExtra, "pool")
	delete(grokExtra, "tier_normalized")
	delete(grokExtra, "raw_tier")
	delete(grokExtra, "tier_raw")
	delete(grokExtra, "tier_source")
	delete(grokExtra, "tier_confidence")

	if quotaWindows := normalizeGrokQuotaWindows(grokExtra["quota_windows"], normalizedTier); len(quotaWindows) > 0 {
		grokExtra["quota_windows"] = quotaWindows
	}

	mergedExtra["grok"] = grokExtra
	return mergedExtra
}

func normalizeGrokQuotaWindows(raw any, tier grok.Tier) map[string]any {
	current := grokQuotaWindowsMap(raw)
	switch tier {
	case grok.TierBasic, grok.TierSuper, grok.TierHeavy:
		defaults := grok.DefaultQuotaSet(tier).ToMap()
		windows := make(map[string]any, len(defaults))
		for windowName, defaultWindow := range defaults {
			windows[windowName] = canonicalizeGrokQuotaWindow(defaultWindow, current[windowName])
		}
		return windows
	default:
		return cloneAnyMap(current)
	}
}

func canonicalizeGrokQuotaWindow(defaultWindow any, currentWindow any) map[string]any {
	canonical := cloneAnyMap(grokNestedMap(defaultWindow))
	current := grokNestedMap(currentWindow)
	if len(canonical) == 0 {
		return cloneAnyMap(current)
	}
	if len(current) == 0 {
		return canonical
	}

	for key, value := range current {
		switch key {
		case "total", "window_seconds":
			continue
		case "reset_at":
			if value == nil {
				delete(canonical, key)
				continue
			}
			if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
				delete(canonical, key)
				continue
			}
		case "source":
			if strings.TrimSpace(getStringFromMaps(current, nil, key)) == "" {
				continue
			}
		}
		canonical[key] = cloneAnyValue(value)
	}
	return canonical
}

func defaultGrokAuthMode(accountType string) string {
	switch accountType {
	case AccountTypeSession:
		return AccountTypeSession
	case AccountTypeAPIKey:
		return AccountTypeAPIKey
	case AccountTypeUpstream:
		return AccountTypeUpstream
	default:
		return ""
	}
}

func grokExtraMap(extra map[string]any) map[string]any {
	if len(extra) == 0 {
		return nil
	}
	raw, ok := extra["grok"]
	if !ok || raw == nil {
		return nil
	}
	value, _ := raw.(map[string]any)
	return value
}

func getNestedGrokValue(extra map[string]any, key string, nested string) any {
	value := grokNestedMap(extra[key])
	if len(value) == 0 {
		return nil
	}
	return value[nested]
}

func mergeAnyMaps(base map[string]any, override map[string]any) map[string]any {
	if len(base) == 0 && len(override) == 0 {
		return map[string]any{}
	}
	result := cloneAnyMap(base)
	for key, value := range override {
		overrideMap, overrideIsMap := value.(map[string]any)
		baseMap, baseIsMap := result[key].(map[string]any)
		if overrideIsMap && baseIsMap {
			result[key] = mergeAnyMaps(baseMap, overrideMap)
			continue
		}
		result[key] = cloneAnyValue(value)
	}
	return result
}

func cloneAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneAnyValue(value)
	}
	return out
}

func cloneAnyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneAnyMap(typed)
	case []any:
		result := make([]any, len(typed))
		for idx := range typed {
			result[idx] = cloneAnyValue(typed[idx])
		}
		return result
	default:
		return value
	}
}
