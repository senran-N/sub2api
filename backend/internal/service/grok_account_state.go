package service

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/grok"
)

type GrokTierState struct {
	Normalized grok.Tier
	Raw        string
	Source     string
	Confidence float64
}

type grokCapabilityState struct {
	operations         map[grok.Capability]bool
	models             map[string]struct{}
	hasOperationSignal bool
	hasModelSignal     bool
}

type grokQuotaWindowState struct {
	Name          string
	Remaining     int
	Total         int
	WindowSeconds int
	Source        string
	ResetAt       *time.Time
	HasSignal     bool
}

type grokSyncState struct {
	LastSyncAt    *time.Time
	LastProbeAt   *time.Time
	LastProbeOKAt *time.Time
}

type grokRuntimeSelectionState struct {
	LastUseAt             *time.Time
	LastFailAt            *time.Time
	LastFailoverAt        *time.Time
	LastFailStatusCode    int
	LastFailClass         grokRuntimeErrorClass
	LastFailScope         grokRuntimePenaltyScope
	LastFailRetryable     bool
	CooldownUntil         *time.Time
	CooldownScope         grokRuntimePenaltyScope
	CooldownModel         string
	HasFailRetryableState bool
}

func (a *Account) grokExtraMap() map[string]any {
	if a == nil || a.Extra == nil {
		return nil
	}
	raw, ok := a.Extra["grok"]
	if !ok || raw == nil {
		return nil
	}
	value, _ := raw.(map[string]any)
	return value
}

func grokNestedMap(raw any) map[string]any {
	value, _ := raw.(map[string]any)
	return value
}

func grokNormalizeTier(raw string) grok.Tier {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(grok.TierBasic):
		return grok.TierBasic
	case string(grok.TierSuper):
		return grok.TierSuper
	case string(grok.TierHeavy):
		return grok.TierHeavy
	default:
		return grok.TierUnknown
	}
}

func grokParseStringSlice(raw any) []string {
	switch typed := raw.(type) {
	case []string:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(item)
			if text != "" {
				result = append(result, text)
			}
		}
		return result
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				text = strings.TrimSpace(text)
				if text != "" {
					result = append(result, text)
				}
			}
		}
		return result
	default:
		return nil
	}
}

func grokParseBool(raw any) (bool, bool) {
	switch typed := raw.(type) {
	case bool:
		return typed, true
	case string:
		if typed == "" {
			return false, false
		}
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		if err != nil {
			return false, false
		}
		return parsed, true
	default:
		return false, false
	}
}

func grokParseFloat(raw any) float64 {
	switch typed := raw.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return parsed
		}
	}
	return 0
}

func grokParseInt(raw any) int {
	switch typed := raw.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed)
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed
		}
	}
	return 0
}

func grokParseTime(raw any) *time.Time {
	switch typed := raw.(type) {
	case time.Time:
		value := typed.UTC()
		return &value
	case *time.Time:
		if typed == nil {
			return nil
		}
		value := typed.UTC()
		return &value
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return nil
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			parsed, err := time.Parse(layout, text)
			if err == nil {
				value := parsed.UTC()
				return &value
			}
		}
	}
	return nil
}

func grokQuotaWindowsMap(raw any) map[string]any {
	value, _ := raw.(map[string]any)
	return value
}

func grokInferTierFromQuotaWindows(raw any) grok.Tier {
	windows := grokQuotaWindowsMap(raw)
	if len(windows) == 0 {
		return grok.TierUnknown
	}
	auto := grokNestedMap(windows[grok.QuotaWindowAuto])
	if len(auto) == 0 {
		return grok.TierUnknown
	}
	return grok.InferTierFromAutoTotal(grokParseInt(auto["total"]))
}

func normalizeGrokCapability(raw string) grok.Capability {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(grok.CapabilityChat):
		return grok.CapabilityChat
	case string(grok.CapabilityImage):
		return grok.CapabilityImage
	case string(grok.CapabilityImageEdit):
		return grok.CapabilityImageEdit
	case string(grok.CapabilityVideo):
		return grok.CapabilityVideo
	case string(grok.CapabilityVoice):
		return grok.CapabilityVoice
	default:
		return ""
	}
}

func (a *Account) GrokTierState() GrokTierState {
	state := GrokTierState{Normalized: grok.TierUnknown}
	extra := a.grokExtraMap()
	if extra == nil {
		return state
	}

	tier := grokNestedMap(extra["tier"])
	if normalized := grokNormalizeTier(getStringFromMaps(tier, extra, "normalized", "tier_normalized")); normalized != grok.TierUnknown {
		state.Normalized = normalized
	}
	if state.Normalized == grok.TierUnknown {
		state.Normalized = grokNormalizeTier(getStringFromMaps(extra, nil, "pool"))
	}
	if state.Normalized == grok.TierUnknown {
		state.Normalized = grokInferTierFromQuotaWindows(extra["quota_windows"])
	}
	state.Raw = getStringFromMaps(tier, extra, "raw", "raw_tier", "tier_raw")
	state.Source = getStringFromMaps(tier, extra, "source", "tier_source")
	if tier != nil {
		state.Confidence = grokParseFloat(tier["confidence"])
	}
	if state.Confidence == 0 {
		state.Confidence = grokParseFloat(extra["tier_confidence"])
	}
	return state
}

func getStringFromMaps(primary map[string]any, secondary map[string]any, keys ...string) string {
	for _, key := range keys {
		if primary != nil {
			if value, ok := primary[key].(string); ok {
				if trimmed := strings.TrimSpace(value); trimmed != "" {
					return trimmed
				}
			}
		}
		if secondary != nil {
			if value, ok := secondary[key].(string); ok {
				if trimmed := strings.TrimSpace(value); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return ""
}

func (a *Account) grokCapabilities() grokCapabilityState {
	state := grokCapabilityState{
		operations: make(map[grok.Capability]bool),
		models:     make(map[string]struct{}),
	}
	extra := a.grokExtraMap()
	if extra == nil {
		return state
	}
	rawCaps := extra["capabilities"]
	if rawCaps == nil {
		return state
	}

	switch typed := rawCaps.(type) {
	case []string, []any:
		for _, item := range grokParseStringSlice(typed) {
			capability := normalizeGrokCapability(item)
			if capability == "" {
				continue
			}
			state.operations[capability] = true
			state.hasOperationSignal = true
		}
	case map[string]any:
		if _, ok := typed["models"]; ok {
			state.hasModelSignal = true
		}
		if models := grokParseStringSlice(typed["models"]); len(models) > 0 {
			for _, model := range models {
				state.models[grok.ResolveCanonicalModelID(model)] = struct{}{}
			}
		}
		if operations := grokParseStringSlice(typed["operations"]); len(operations) > 0 {
			state.hasOperationSignal = true
			for _, item := range operations {
				capability := normalizeGrokCapability(item)
				if capability != "" {
					state.operations[capability] = true
				}
			}
		}
		for _, capability := range []grok.Capability{
			grok.CapabilityChat,
			grok.CapabilityImage,
			grok.CapabilityImageEdit,
			grok.CapabilityVideo,
			grok.CapabilityVoice,
		} {
			if allowed, ok := grokParseBool(typed[string(capability)]); ok {
				state.operations[capability] = allowed
				state.hasOperationSignal = true
			}
		}
	}

	return state
}

func (a *Account) GrokCapabilityModelIDs() []string {
	state := a.grokCapabilities()
	if len(state.models) == 0 {
		return nil
	}
	ids := make([]string, 0, len(state.models))
	for id := range state.models {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (a *Account) grokQuotaWindow(name string) grokQuotaWindowState {
	state := grokQuotaWindowState{Name: strings.TrimSpace(name)}
	if state.Name == "" {
		return state
	}

	extra := a.grokExtraMap()
	if extra == nil {
		return state
	}

	window := grokNestedMap(grokQuotaWindowsMap(extra["quota_windows"])[state.Name])
	if len(window) == 0 {
		return state
	}

	state.HasSignal = true
	state.Remaining = grokParseInt(window["remaining"])
	state.Total = grokParseInt(window["total"])
	state.WindowSeconds = grokParseInt(window["window_seconds"])
	state.Source = getStringFromMaps(window, nil, "source")
	state.ResetAt = grokParseTime(window["reset_at"])
	return state
}

func (a *Account) grokSyncState() grokSyncState {
	extra := a.grokExtraMap()
	if extra == nil {
		return grokSyncState{}
	}

	syncState := grokNestedMap(extra["sync_state"])
	return grokSyncState{
		LastSyncAt:    grokParseTime(syncState["last_sync_at"]),
		LastProbeAt:   grokParseTime(syncState["last_probe_at"]),
		LastProbeOKAt: grokParseTime(syncState["last_probe_ok_at"]),
	}
}

func (a *Account) grokRuntimeSelectionState() grokRuntimeSelectionState {
	extra := a.grokExtraMap()
	if extra == nil {
		return grokRuntimeSelectionState{}
	}

	runtimeState := grokNestedMap(extra["runtime_state"])
	lastFailRetryable, hasLastFailRetryable := grokParseBool(runtimeState["last_fail_retryable"])
	return grokRuntimeSelectionState{
		LastUseAt:             grokParseTime(runtimeState["last_use_at"]),
		LastFailAt:            grokParseTime(runtimeState["last_fail_at"]),
		LastFailoverAt:        grokParseTime(runtimeState["last_failover_at"]),
		LastFailStatusCode:    grokParseInt(runtimeState["last_fail_status_code"]),
		LastFailClass:         grokRuntimeErrorClass(getStringFromMaps(runtimeState, nil, "last_fail_class")),
		LastFailScope:         grokRuntimePenaltyScope(getStringFromMaps(runtimeState, nil, "last_fail_scope")),
		LastFailRetryable:     lastFailRetryable,
		CooldownUntil:         grokParseTime(runtimeState["selection_cooldown_until"]),
		CooldownScope:         grokRuntimePenaltyScope(getStringFromMaps(runtimeState, nil, "selection_cooldown_scope")),
		CooldownModel:         strings.TrimSpace(getStringFromMaps(runtimeState, nil, "selection_cooldown_model")),
		HasFailRetryableState: hasLastFailRetryable,
	}
}
