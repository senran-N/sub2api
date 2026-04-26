package service

import (
	"encoding/json"
	"hash/fnv"
	"reflect"
	"sort"
	"strings"

	"github.com/senran-N/sub2api/internal/domain"
	"github.com/senran-N/sub2api/internal/pkg/openai"
)

func (a *Account) GetModelMapping() map[string]string {
	credentialsPtr := mapPtr(a.Credentials)
	rawMapping, _ := a.Credentials["model_mapping"].(map[string]any)
	rawPtr := mapPtr(rawMapping)
	rawLen := len(rawMapping)
	rawSig := uint64(0)
	rawSigReady := false

	if a.modelMappingCacheReady &&
		a.modelMappingCacheCredentialsPtr == credentialsPtr &&
		a.modelMappingCacheRawPtr == rawPtr &&
		a.modelMappingCacheRawLen == rawLen {
		rawSig = modelMappingSignature(rawMapping)
		rawSigReady = true
		if a.modelMappingCacheRawSig == rawSig {
			return a.modelMappingCache
		}
	}

	mapping := a.resolveModelMapping(rawMapping)
	if !rawSigReady {
		rawSig = modelMappingSignature(rawMapping)
	}

	a.modelMappingCache = mapping
	a.modelMappingCacheReady = true
	a.modelMappingCacheCredentialsPtr = credentialsPtr
	a.modelMappingCacheRawPtr = rawPtr
	a.modelMappingCacheRawLen = rawLen
	a.modelMappingCacheRawSig = rawSig
	return mapping
}

func (a *Account) resolveModelMapping(rawMapping map[string]any) map[string]string {
	if a.Credentials == nil {
		if a.Platform == domain.PlatformAntigravity {
			return domain.DefaultAntigravityModelMapping
		}
		return nil
	}
	if len(rawMapping) == 0 {
		if a.Platform == domain.PlatformAntigravity {
			return domain.DefaultAntigravityModelMapping
		}
		return nil
	}

	result := make(map[string]string)
	for key, value := range rawMapping {
		if mapped, ok := value.(string); ok {
			result[key] = mapped
		}
	}
	if len(result) > 0 {
		switch a.Platform {
		case domain.PlatformAntigravity:
			ensureDefaultModelPassthroughs(result, []string{
				"gemini-3-flash",
				"gemini-3.1-pro-high",
				"gemini-3.1-pro-low",
				"gemini-3.1-pro-preview-customtools",
			})
		case domain.PlatformOpenAI:
			ensureOpenAIDefaultImagePassthroughs(result)
		}
		return result
	}

	if a.Platform == domain.PlatformAntigravity {
		return domain.DefaultAntigravityModelMapping
	}
	return nil
}

func mapPtr(mapping map[string]any) uintptr {
	if mapping == nil {
		return 0
	}
	return reflect.ValueOf(mapping).Pointer()
}

func stringMappingFromRaw(raw any) map[string]string {
	switch mapping := raw.(type) {
	case map[string]any:
		if len(mapping) == 0 {
			return nil
		}
		result := make(map[string]string, len(mapping))
		for key, value := range mapping {
			if str, ok := value.(string); ok {
				result[key] = str
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case map[string]string:
		if len(mapping) == 0 {
			return nil
		}
		result := make(map[string]string, len(mapping))
		for key, value := range mapping {
			result[key] = value
		}
		return result
	default:
		return nil
	}
}

func modelMappingSignature(rawMapping map[string]any) uint64 {
	if len(rawMapping) == 0 {
		return 0
	}

	keys := make([]string, 0, len(rawMapping))
	for key := range rawMapping {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	hasher := fnv.New64a()
	for _, key := range keys {
		_, _ = hasher.Write([]byte(key))
		_, _ = hasher.Write([]byte{0})
		if value, ok := rawMapping[key].(string); ok {
			_, _ = hasher.Write([]byte(value))
		} else {
			_, _ = hasher.Write([]byte{1})
		}
		_, _ = hasher.Write([]byte{0xff})
	}
	return hasher.Sum64()
}

func ensureDefaultModelPassthrough(mapping map[string]string, model string) {
	if mapping == nil || model == "" {
		return
	}
	if _, exists := mapping[model]; exists {
		return
	}
	for pattern := range mapping {
		if matchWildcard(pattern, model) {
			return
		}
	}
	mapping[model] = model
}

func ensureDefaultModelPassthroughs(mapping map[string]string, models []string) {
	for _, model := range models {
		ensureDefaultModelPassthrough(mapping, model)
	}
}

func ensureOpenAIDefaultImagePassthroughs(mapping map[string]string) {
	for _, model := range openai.DefaultModelIDs() {
		if isOpenAIImageGenerationModel(model) {
			ensureOpenAIImagePassthrough(mapping, model)
		}
	}
}

func ensureOpenAIImagePassthrough(mapping map[string]string, model string) {
	if mapping == nil || model == "" {
		return
	}
	if _, exists := mapping[model]; exists {
		return
	}
	for pattern := range mapping {
		if isOpenAIImageGenerationModel(pattern) && matchWildcard(pattern, model) {
			return
		}
	}
	mapping[model] = model
}

func (a *Account) IsModelSupported(requestedModel string) bool {
	mapping := a.GetModelMapping()
	if len(mapping) == 0 {
		return true
	}
	if _, exists := mapping[requestedModel]; exists {
		return true
	}
	for pattern := range mapping {
		if matchWildcard(pattern, requestedModel) {
			return true
		}
	}
	return false
}

func (a *Account) GetExplicitSupportedModels() []string {
	if a == nil || len(a.Extra) == 0 {
		return nil
	}
	return normalizeSupportedModelList(a.Extra["supported_models"])
}

func (a *Account) IsExplicitSupportedModel(requestedModel string) bool {
	return supportedModelListMatches(a.GetExplicitSupportedModels(), requestedModel)
}

// SchedulerModelCapabilityValues returns the model capability values used by
// the scheduler index. OpenAI accounts only opt into model-specific indexing
// when they explicitly declare supported_models; an empty declaration means the
// capability is unknown and must not restrict routing.
func (a *Account) SchedulerModelCapabilityValues() (values []string, unrestricted bool) {
	if a == nil {
		return nil, false
	}
	if a.Platform == PlatformOpenAI {
		supportedModels := a.GetExplicitSupportedModels()
		if len(supportedModels) == 0 {
			return nil, true
		}
		return supportedModels, false
	}

	mapping := a.GetModelMapping()
	if len(mapping) == 0 {
		return nil, true
	}
	values = make([]string, 0, len(mapping))
	for model := range mapping {
		model = strings.TrimSpace(model)
		if model != "" {
			values = append(values, model)
		}
	}
	sort.Strings(values)
	return values, false
}

func normalizeSupportedModelList(raw any) []string {
	switch typed := raw.(type) {
	case nil:
		return nil
	case []string:
		return normalizeSupportedModelStrings(typed)
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if value, ok := item.(string); ok {
				values = append(values, value)
			}
		}
		return normalizeSupportedModelStrings(values)
	case string:
		return normalizeSupportedModelString(typed)
	default:
		return nil
	}
}

func normalizeSupportedModelString(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return normalizeSupportedModelStrings(values)
	}

	var anyValues []any
	if err := json.Unmarshal([]byte(raw), &anyValues); err == nil {
		values = values[:0]
		for _, item := range anyValues {
			if value, ok := item.(string); ok {
				values = append(values, value)
			}
		}
		return normalizeSupportedModelStrings(values)
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	return normalizeSupportedModelStrings(parts)
}

func normalizeSupportedModelStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func supportedModelListMatches(values []string, requestedModel string) bool {
	requestedModel = strings.TrimSpace(requestedModel)
	if requestedModel == "" {
		return true
	}
	if supportedModelListContains(values, requestedModel) {
		return true
	}

	_, baseModel, _, ok := splitOpenAICompatReasoningModel(requestedModel)
	if !ok || baseModel == "" || baseModel == requestedModel {
		return false
	}
	return supportedModelListContains(values, baseModel)
}

func supportedModelListContains(values []string, requestedModel string) bool {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if value == requestedModel || matchWildcard(value, requestedModel) {
			return true
		}
	}
	return false
}

func (a *Account) GetMappedModel(requestedModel string) string {
	mappedModel, _ := a.ResolveMappedModel(requestedModel)
	return mappedModel
}

func (a *Account) ResolveMappedModel(requestedModel string) (mappedModel string, matched bool) {
	mapping := a.GetModelMapping()
	if len(mapping) == 0 {
		return requestedModel, false
	}
	if mappedModel, exists := mapping[requestedModel]; exists {
		return mappedModel, true
	}
	return matchWildcardMappingResult(mapping, requestedModel)
}

// GetCompactModelMapping returns compact-only model remapping configuration.
// This mapping is intended for /responses/compact only and does not affect normal /responses traffic.
func (a *Account) GetCompactModelMapping() map[string]string {
	if a == nil || a.Credentials == nil {
		return nil
	}
	return stringMappingFromRaw(a.Credentials["compact_model_mapping"])
}

// ResolveCompactMappedModel resolves compact-only model remapping and reports whether a rule matched.
func (a *Account) ResolveCompactMappedModel(requestedModel string) (mappedModel string, matched bool) {
	mapping := a.GetCompactModelMapping()
	if len(mapping) == 0 {
		return requestedModel, false
	}
	if mappedModel, exists := mapping[requestedModel]; exists {
		return mappedModel, true
	}
	return matchWildcardMappingResult(mapping, requestedModel)
}

func matchAntigravityWildcard(pattern, candidate string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(candidate, pattern[:len(pattern)-1])
	}
	return pattern == candidate
}

func matchWildcard(pattern, candidate string) bool {
	return matchAntigravityWildcard(pattern, candidate)
}

func matchWildcardMappingResult(mapping map[string]string, requestedModel string) (string, bool) {
	type patternMatch struct {
		pattern string
		target  string
	}

	var matches []patternMatch
	for pattern, target := range mapping {
		if matchWildcard(pattern, requestedModel) {
			matches = append(matches, patternMatch{pattern: pattern, target: target})
		}
	}
	if len(matches) == 0 {
		return requestedModel, false
	}

	sort.Slice(matches, func(i, j int) bool {
		if len(matches[i].pattern) != len(matches[j].pattern) {
			return len(matches[i].pattern) > len(matches[j].pattern)
		}
		return matches[i].pattern < matches[j].pattern
	})

	return matches[0].target, true
}
