package service

import (
	"hash/fnv"
	"reflect"
	"sort"
	"strings"

	"github.com/senran-N/sub2api/internal/domain"
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
		if a.Platform == domain.PlatformAntigravity {
			ensureAntigravityDefaultPassthroughs(result, []string{
				"gemini-3-flash",
				"gemini-3.1-pro-high",
				"gemini-3.1-pro-low",
				"gemini-3.1-pro-preview-customtools",
			})
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

func ensureAntigravityDefaultPassthrough(mapping map[string]string, model string) {
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

func ensureAntigravityDefaultPassthroughs(mapping map[string]string, models []string) {
	for _, model := range models {
		ensureAntigravityDefaultPassthrough(mapping, model)
	}
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
