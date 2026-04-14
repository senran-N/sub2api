package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/senran-N/sub2api/internal/util/urlvalidator"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type openAIRequestMeta struct {
	Model              string
	ModelExists        bool
	ModelType          gjson.Type
	Stream             bool
	StreamExists       bool
	StreamType         gjson.Type
	PromptCacheKey     string
	PreviousResponseID string
	ReasoningPresent   bool
	ReasoningEffort    string
}

type openAIRequestMetaCache struct {
	BodyBound bool
	BodyHash  uint64
}

func (s *OpenAIGatewayService) validateUpstreamBaseURL(raw string) (string, error) {
	if s.cfg != nil && !s.cfg.Security.URLAllowlist.Enabled {
		normalized, err := urlvalidator.ValidateURLFormat(raw, s.cfg.Security.URLAllowlist.AllowInsecureHTTP)
		if err != nil {
			return "", fmt.Errorf("invalid base_url: %w", err)
		}
		return normalized, nil
	}
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	return normalized, nil
}

func trimOpenAIEncryptedReasoningItems(reqBody map[string]any) bool {
	if len(reqBody) == 0 {
		return false
	}

	inputValue, has := reqBody["input"]
	if !has {
		return false
	}

	switch input := inputValue.(type) {
	case []any:
		filtered := input[:0]
		changed := false
		for _, item := range input {
			nextItem, itemChanged, keep := sanitizeEncryptedReasoningInputItem(item)
			if itemChanged {
				changed = true
			}
			if !keep {
				continue
			}
			filtered = append(filtered, nextItem)
		}
		if !changed {
			return false
		}
		if len(filtered) == 0 {
			delete(reqBody, "input")
			return true
		}
		reqBody["input"] = filtered
		return true
	case []map[string]any:
		filtered := input[:0]
		changed := false
		for _, item := range input {
			nextItem, itemChanged, keep := sanitizeEncryptedReasoningInputItem(item)
			if itemChanged {
				changed = true
			}
			if !keep {
				continue
			}
			nextMap, ok := nextItem.(map[string]any)
			if !ok {
				filtered = append(filtered, item)
				continue
			}
			filtered = append(filtered, nextMap)
		}
		if !changed {
			return false
		}
		if len(filtered) == 0 {
			delete(reqBody, "input")
			return true
		}
		reqBody["input"] = filtered
		return true
	case map[string]any:
		nextItem, changed, keep := sanitizeEncryptedReasoningInputItem(input)
		if !changed {
			return false
		}
		if !keep {
			delete(reqBody, "input")
			return true
		}
		nextMap, ok := nextItem.(map[string]any)
		if !ok {
			return false
		}
		reqBody["input"] = nextMap
		return true
	default:
		return false
	}
}

func sanitizeEncryptedReasoningInputItem(item any) (next any, changed bool, keep bool) {
	inputItem, ok := item.(map[string]any)
	if !ok {
		return item, false, true
	}

	itemType, _ := inputItem["type"].(string)
	if strings.TrimSpace(itemType) != "reasoning" {
		return item, false, true
	}

	if _, hasEncryptedContent := inputItem["encrypted_content"]; !hasEncryptedContent {
		return item, false, true
	}
	delete(inputItem, "encrypted_content")
	if len(inputItem) == 1 {
		return nil, true, false
	}
	return inputItem, true, true
}

func isOpenAIResponsesCompactPath(c *gin.Context) bool {
	suffix := strings.TrimSpace(openAIResponsesRequestPathSuffix(c))
	return suffix == "/compact" || strings.HasPrefix(suffix, "/compact/")
}

func normalizeOpenAICompactRequestBody(body []byte) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}

	normalized := []byte(`{}`)
	for _, field := range []string{"model", "input", "instructions", "previous_response_id"} {
		value := gjson.GetBytes(body, field)
		if !value.Exists() {
			continue
		}
		next, err := sjson.SetRawBytes(normalized, field, []byte(value.Raw))
		if err != nil {
			return body, false, fmt.Errorf("normalize compact body %s: %w", field, err)
		}
		normalized = next
	}

	if bytes.Equal(bytes.TrimSpace(body), bytes.TrimSpace(normalized)) {
		return body, false, nil
	}
	return normalized, true, nil
}

func resolveOpenAICompactSessionID(c *gin.Context) string {
	if c != nil {
		if sessionID := strings.TrimSpace(c.GetHeader("session_id")); sessionID != "" {
			return sessionID
		}
		if conversationID := strings.TrimSpace(c.GetHeader("conversation_id")); conversationID != "" {
			return conversationID
		}
		if seed, ok := c.Get(openAICompactSessionSeedKey); ok {
			if seedStr, ok := seed.(string); ok && strings.TrimSpace(seedStr) != "" {
				return strings.TrimSpace(seedStr)
			}
		}
	}
	return uuid.NewString()
}

func openAIResponsesRequestPathSuffix(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return openAIResponsesRequestPathSuffixFromPath(c.Request.URL.Path)
}

func openAIResponsesRequestPathSuffixFromPath(rawPath string) string {
	normalizedPath := strings.TrimRight(strings.TrimSpace(rawPath), "/")
	if normalizedPath == "" {
		return ""
	}
	index := strings.LastIndex(normalizedPath, "/responses")
	if index < 0 {
		return ""
	}
	suffix := normalizedPath[index+len("/responses"):]
	if suffix == "" || suffix == "/" {
		return ""
	}
	if !strings.HasPrefix(suffix, "/") {
		return ""
	}
	return suffix
}

func appendOpenAIResponsesRequestPathSuffix(baseURL, suffix string) string {
	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	trimmedSuffix := strings.TrimSpace(suffix)
	if trimmedBase == "" || trimmedSuffix == "" {
		return trimmedBase
	}
	return trimmedBase + trimmedSuffix
}

func getOpenAIReasoningEffortFromReqBody(reqBody map[string]any) (value string, present bool) {
	if reqBody == nil {
		return "", false
	}
	if reasoning, ok := reqBody["reasoning"].(map[string]any); ok {
		if effort, ok := reasoning["effort"].(string); ok {
			return normalizeOpenAIReasoningEffort(effort), true
		}
	}
	if effort, ok := reqBody["reasoning_effort"].(string); ok {
		return normalizeOpenAIReasoningEffort(effort), true
	}
	return "", false
}

func deriveOpenAIReasoningEffortFromModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return ""
	}

	modelID := strings.TrimSpace(model)
	if strings.Contains(modelID, "/") {
		parts := strings.Split(modelID, "/")
		modelID = parts[len(parts)-1]
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
		return ""
	}
	return normalizeOpenAIReasoningEffort(parts[len(parts)-1])
}

func buildOpenAIRequestMeta(body []byte) openAIRequestMeta {
	if len(body) == 0 {
		return openAIRequestMeta{}
	}

	values := gjson.GetManyBytes(
		body,
		"model",
		"stream",
		"prompt_cache_key",
		"previous_response_id",
		"reasoning.effort",
		"reasoning_effort",
	)
	meta := openAIRequestMeta{
		ModelExists:        values[0].Exists(),
		ModelType:          values[0].Type,
		Model:              strings.TrimSpace(values[0].String()),
		StreamExists:       values[1].Exists(),
		StreamType:         values[1].Type,
		Stream:             values[1].Bool(),
		PromptCacheKey:     strings.TrimSpace(values[2].String()),
		PreviousResponseID: strings.TrimSpace(values[3].String()),
	}

	reasoningEffort := strings.TrimSpace(values[4].String())
	meta.ReasoningPresent = values[4].Exists()
	if reasoningEffort == "" {
		reasoningEffort = strings.TrimSpace(values[5].String())
		if reasoningEffort != "" || values[5].Exists() {
			meta.ReasoningPresent = true
		}
	}
	meta.ReasoningEffort = normalizeOpenAIReasoningEffort(reasoningEffort)
	return meta
}

func buildOpenAIRequestMetaFromMap(reqBody map[string]any) openAIRequestMeta {
	meta := openAIRequestMeta{}
	if reqBody == nil {
		return meta
	}

	if model, ok := reqBody["model"].(string); ok {
		meta.ModelExists = true
		meta.ModelType = gjson.String
		meta.Model = strings.TrimSpace(model)
	} else if _, exists := reqBody["model"]; exists {
		meta.ModelExists = true
	}
	if stream, ok := reqBody["stream"].(bool); ok {
		meta.StreamExists = true
		meta.StreamType = gjson.True
		if !stream {
			meta.StreamType = gjson.False
		}
		meta.Stream = stream
	} else if _, exists := reqBody["stream"]; exists {
		meta.StreamExists = true
	}
	if promptCacheKey, ok := reqBody["prompt_cache_key"].(string); ok {
		meta.PromptCacheKey = strings.TrimSpace(promptCacheKey)
	}
	if previousResponseID, ok := reqBody["previous_response_id"].(string); ok {
		meta.PreviousResponseID = strings.TrimSpace(previousResponseID)
	}
	if reasoningEffort, present := getOpenAIReasoningEffortFromReqBody(reqBody); present {
		meta.ReasoningPresent = true
		meta.ReasoningEffort = reasoningEffort
	}
	return meta
}

func cacheOpenAIRequestMeta(c *gin.Context, meta openAIRequestMeta) openAIRequestMeta {
	if c != nil {
		c.Set(OpenAIParsedRequestMetaKey, meta)
		c.Set(OpenAIParsedRequestMetaCacheKey, openAIRequestMetaCache{})
	}
	return meta
}

func cacheOpenAIRequestMetaFromMap(c *gin.Context, reqBody map[string]any) openAIRequestMeta {
	return cacheOpenAIRequestMeta(c, buildOpenAIRequestMetaFromMap(reqBody))
}

func cacheOpenAIRequestMetaFromBody(c *gin.Context, body []byte, meta openAIRequestMeta) openAIRequestMeta {
	if c != nil {
		c.Set(OpenAIParsedRequestMetaKey, meta)
		c.Set(OpenAIParsedRequestMetaCacheKey, openAIRequestMetaCache{
			BodyBound: true,
			BodyHash:  xxhash.Sum64(body),
		})
	}
	return meta
}

func getOpenAIRequestMeta(c *gin.Context, body []byte) openAIRequestMeta {
	if c != nil {
		if cached, ok := c.Get(OpenAIParsedRequestMetaKey); ok {
			if meta, ok := cached.(openAIRequestMeta); ok {
				cacheState, hasCacheState := c.Get(OpenAIParsedRequestMetaCacheKey)
				if !hasCacheState {
					return meta
				}
				if cacheTag, ok := cacheState.(openAIRequestMetaCache); ok {
					if !cacheTag.BodyBound || len(body) == 0 || cacheTag.BodyHash == xxhash.Sum64(body) {
						return meta
					}
				}
			}
		}
	}
	return cacheOpenAIRequestMetaFromBody(c, body, buildOpenAIRequestMeta(body))
}

// GetOpenAIRequestMeta returns lightweight request metadata and caches it on gin.Context.
func GetOpenAIRequestMeta(c *gin.Context, body []byte) openAIRequestMeta {
	return getOpenAIRequestMeta(c, body)
}

// CacheOpenAIRequestMetaFromBodyMap caches request metadata derived from a decoded request body map.
func CacheOpenAIRequestMetaFromBodyMap(c *gin.Context, reqBody map[string]any) {
	cacheOpenAIRequestMetaFromMap(c, reqBody)
}

func extractOpenAIRequestMetaFromBody(body []byte) (model string, stream bool, promptCacheKey string) {
	meta := buildOpenAIRequestMeta(body)
	return meta.Model, meta.Stream, meta.PromptCacheKey
}

// normalizeOpenAIPassthroughOAuthBody 将透传 OAuth 请求体收敛为旧链路关键行为。
func normalizeOpenAIPassthroughOAuthBody(body []byte, policy CodexNativeMutationPolicy) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}

	normalized := body
	changed := false
	if policy.Profile.CompactPath {
		if store := gjson.GetBytes(normalized, "store"); store.Exists() {
			next, err := sjson.DeleteBytes(normalized, "store")
			if err != nil {
				return body, false, fmt.Errorf("normalize passthrough body delete store: %w", err)
			}
			normalized = next
			changed = true
		}
		if stream := gjson.GetBytes(normalized, "stream"); stream.Exists() {
			next, err := sjson.DeleteBytes(normalized, "stream")
			if err != nil {
				return body, false, fmt.Errorf("normalize passthrough body delete stream: %w", err)
			}
			normalized = next
			changed = true
		}
	} else {
		if store := gjson.GetBytes(normalized, "store"); !store.Exists() || store.Type != gjson.False {
			next, err := sjson.SetBytes(normalized, "store", false)
			if err != nil {
				return body, false, fmt.Errorf("normalize passthrough body store=false: %w", err)
			}
			normalized = next
			changed = true
		}
		if stream := gjson.GetBytes(normalized, "stream"); !stream.Exists() || stream.Type != gjson.True {
			next, err := sjson.SetBytes(normalized, "stream", true)
			if err != nil {
				return body, false, fmt.Errorf("normalize passthrough body stream=true: %w", err)
			}
			normalized = next
			changed = true
		}
	}

	model := strings.ToLower(strings.TrimSpace(gjson.GetBytes(normalized, "model").String()))
	if strings.Contains(model, "codex") && !policy.Profile.NativeClient {
		instructions := gjson.GetBytes(normalized, "instructions")
		if !instructions.Exists() || instructions.Type != gjson.String || strings.TrimSpace(instructions.String()) == "" {
			next, err := sjson.SetBytes(normalized, "instructions", defaultOpenAICodexInstructions)
			if err != nil {
				return body, false, fmt.Errorf("normalize passthrough body default instructions: %w", err)
			}
			normalized = next
			changed = true
		}
	}

	return normalized, changed, nil
}

func extractOpenAIReasoningEffortFromBody(body []byte, requestedModel string) *string {
	meta := buildOpenAIRequestMeta(body)
	if meta.ReasoningPresent {
		if meta.ReasoningEffort == "" {
			return nil
		}
		value := meta.ReasoningEffort
		return &value
	}
	if meta.ReasoningEffort != "" {
		value := meta.ReasoningEffort
		return &value
	}

	value := deriveOpenAIReasoningEffortFromModel(requestedModel)
	if value == "" {
		return nil
	}
	return &value
}

func extractOpenAIServiceTier(reqBody map[string]any) *string {
	if reqBody == nil {
		return nil
	}
	raw, ok := reqBody["service_tier"].(string)
	if !ok {
		return nil
	}
	return normalizeOpenAIServiceTier(raw)
}

func extractOpenAIServiceTierFromBody(body []byte) *string {
	if len(body) == 0 {
		return nil
	}
	return normalizeOpenAIServiceTier(gjson.GetBytes(body, "service_tier").String())
}

func normalizeOpenAIServiceTier(raw string) *string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return nil
	}
	if value == "fast" {
		value = "priority"
	}
	switch value {
	case "priority", "flex":
		return &value
	default:
		return nil
	}
}

func getOpenAIRequestBodyMap(c *gin.Context, body []byte) (map[string]any, error) {
	if c != nil {
		if cached, ok := c.Get(OpenAIParsedRequestBodyKey); ok {
			if reqBody, ok := cached.(map[string]any); ok && reqBody != nil {
				cacheOpenAIRequestMetaFromMap(c, reqBody)
				return reqBody, nil
			}
		}
	}

	var reqBody map[string]any
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}
	if c != nil {
		c.Set(OpenAIParsedRequestBodyKey, reqBody)
		cacheOpenAIRequestMetaFromMap(c, reqBody)
	}
	return reqBody, nil
}

func extractOpenAIReasoningEffort(reqBody map[string]any, requestedModel string) *string {
	if value, present := getOpenAIReasoningEffortFromReqBody(reqBody); present {
		if value == "" {
			return nil
		}
		return &value
	}
	value := deriveOpenAIReasoningEffortFromModel(requestedModel)
	if value == "" {
		return nil
	}
	return &value
}

func normalizeOpenAIReasoningEffort(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}
	value = strings.NewReplacer("-", "", "_", "", " ", "").Replace(value)
	switch value {
	case "none", "minimal":
		return ""
	case "low", "medium", "high":
		return value
	case "xhigh", "extrahigh":
		return "xhigh"
	default:
		return ""
	}
}
