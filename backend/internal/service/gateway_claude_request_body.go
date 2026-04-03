package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/senran-N/sub2api/internal/pkg/claude"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type claudeOAuthNormalizeOptions struct {
	injectMetadata          bool
	metadataUserID          string
	stripSystemCacheControl bool
}

type anthropicCacheControlPayload struct {
	Type string `json:"type"`
}

type anthropicSystemTextBlockPayload struct {
	Type         string                        `json:"type"`
	Text         string                        `json:"text"`
	CacheControl *anthropicCacheControlPayload `json:"cache_control,omitempty"`
}

type anthropicMetadataPayload struct {
	UserID string `json:"user_id"`
}

// sanitizeSystemText rewrites only the fixed OpenCode identity sentence (if present).
// We intentionally avoid broad keyword replacement in system prompts to prevent
// accidentally changing user-provided instructions.
func sanitizeSystemText(text string) string {
	if text == "" {
		return text
	}
	text = strings.ReplaceAll(
		text,
		"You are OpenCode, the best coding agent on the planet.",
		strings.TrimSpace(currentClaudeCodeSystemPrompt()),
	)
	return text
}

func marshalAnthropicSystemTextBlock(text string, includeCacheControl bool) ([]byte, error) {
	block := anthropicSystemTextBlockPayload{
		Type: "text",
		Text: text,
	}
	if includeCacheControl {
		block.CacheControl = &anthropicCacheControlPayload{Type: "ephemeral"}
	}
	return json.Marshal(block)
}

func marshalAnthropicMetadata(userID string) ([]byte, error) {
	return json.Marshal(anthropicMetadataPayload{UserID: userID})
}

func buildJSONArrayRaw(items [][]byte) []byte {
	if len(items) == 0 {
		return []byte("[]")
	}

	total := 2
	for _, item := range items {
		total += len(item)
	}
	total += len(items) - 1

	buf := make([]byte, 0, total)
	buf = append(buf, '[')
	for i, item := range items {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, item...)
	}
	buf = append(buf, ']')
	return buf
}

func setJSONValueBytes(body []byte, path string, value any) ([]byte, bool) {
	next, err := sjson.SetBytes(body, path, value)
	if err != nil {
		return body, false
	}
	return next, true
}

func setJSONRawBytes(body []byte, path string, raw []byte) ([]byte, bool) {
	next, err := sjson.SetRawBytes(body, path, raw)
	if err != nil {
		return body, false
	}
	return next, true
}

func deleteJSONPathBytes(body []byte, path string) ([]byte, bool) {
	next, err := sjson.DeleteBytes(body, path)
	if err != nil {
		return body, false
	}
	return next, true
}

func normalizeClaudeOAuthSystemBody(body []byte, opts claudeOAuthNormalizeOptions) ([]byte, bool) {
	sys := gjson.GetBytes(body, "system")
	if !sys.Exists() {
		return body, false
	}

	out := body
	modified := false

	switch {
	case sys.Type == gjson.String:
		sanitized := sanitizeSystemText(sys.String())
		if sanitized != sys.String() {
			if next, ok := setJSONValueBytes(out, "system", sanitized); ok {
				out = next
				modified = true
			}
		}
	case sys.IsArray():
		index := 0
		sys.ForEach(func(_, item gjson.Result) bool {
			if item.Get("type").String() == "text" {
				textResult := item.Get("text")
				if textResult.Exists() && textResult.Type == gjson.String {
					text := textResult.String()
					sanitized := sanitizeSystemText(text)
					if sanitized != text {
						if next, ok := setJSONValueBytes(out, fmt.Sprintf("system.%d.text", index), sanitized); ok {
							out = next
							modified = true
						}
					}
				}
			}

			if opts.stripSystemCacheControl && item.Get("cache_control").Exists() {
				if next, ok := deleteJSONPathBytes(out, fmt.Sprintf("system.%d.cache_control", index)); ok {
					out = next
					modified = true
				}
			}

			index++
			return true
		})
	}

	return out, modified
}

func ensureClaudeOAuthMetadataUserID(body []byte, userID string) ([]byte, bool) {
	if strings.TrimSpace(userID) == "" {
		return body, false
	}

	metadata := gjson.GetBytes(body, "metadata")
	if !metadata.Exists() || metadata.Type == gjson.Null {
		raw, err := marshalAnthropicMetadata(userID)
		if err != nil {
			return body, false
		}
		return setJSONRawBytes(body, "metadata", raw)
	}

	trimmedRaw := strings.TrimSpace(metadata.Raw)
	if strings.HasPrefix(trimmedRaw, "{") {
		existing := metadata.Get("user_id")
		if existing.Exists() && existing.Type == gjson.String && existing.String() != "" {
			return body, false
		}
		return setJSONValueBytes(body, "metadata.user_id", userID)
	}

	raw, err := marshalAnthropicMetadata(userID)
	if err != nil {
		return body, false
	}
	return setJSONRawBytes(body, "metadata", raw)
}

func normalizeClaudeOAuthRequestBody(body []byte, modelID string, opts claudeOAuthNormalizeOptions) ([]byte, string) {
	if len(body) == 0 {
		return body, modelID
	}

	out := body
	modified := false

	if next, changed := normalizeClaudeOAuthSystemBody(out, opts); changed {
		out = next
		modified = true
	}

	rawModel := gjson.GetBytes(out, "model")
	if rawModel.Exists() && rawModel.Type == gjson.String {
		normalized := claude.NormalizeModelID(rawModel.String())
		if normalized != rawModel.String() {
			if next, ok := setJSONValueBytes(out, "model", normalized); ok {
				out = next
				modified = true
			}
			modelID = normalized
		}
	}

	if !gjson.GetBytes(out, "tools").Exists() {
		if next, ok := setJSONRawBytes(out, "tools", []byte("[]")); ok {
			out = next
			modified = true
		}
	}

	if opts.injectMetadata && opts.metadataUserID != "" {
		if next, changed := ensureClaudeOAuthMetadataUserID(out, opts.metadataUserID); changed {
			out = next
			modified = true
		}
	}

	if gjson.GetBytes(out, "temperature").Exists() {
		if next, ok := deleteJSONPathBytes(out, "temperature"); ok {
			out = next
			modified = true
		}
	}
	if gjson.GetBytes(out, "tool_choice").Exists() {
		if next, ok := deleteJSONPathBytes(out, "tool_choice"); ok {
			out = next
			modified = true
		}
	}

	if !modified {
		return body, modelID
	}

	return out, modelID
}

func (s *GatewayService) buildOAuthMetadataUserID(parsed *ParsedRequest, account *Account, fp *Fingerprint) string {
	if parsed == nil || account == nil {
		return ""
	}
	if parsed.MetadataUserID != "" {
		return ""
	}

	userID := strings.TrimSpace(account.GetClaudeUserID())
	if userID == "" && fp != nil {
		userID = fp.ClientID
	}
	if userID == "" {
		userID = generateClientID()
	}

	sessionHash := s.GenerateSessionHash(parsed)
	sessionID := uuid.NewString()
	if sessionHash != "" {
		seed := fmt.Sprintf("%d::%s", account.ID, sessionHash)
		sessionID = generateSessionUUID(seed)
	}

	var uaVersion string
	if fp != nil {
		uaVersion = ExtractCLIVersion(fp.UserAgent)
	}
	accountUUID := strings.TrimSpace(account.GetExtraString("account_uuid"))
	return FormatMetadataUserID(userID, accountUUID, sessionID, uaVersion)
}

// GenerateSessionUUID creates a deterministic UUID4 from a seed string.
func GenerateSessionUUID(seed string) string {
	return generateSessionUUID(seed)
}

func generateSessionUUID(seed string) string {
	if seed == "" {
		return uuid.NewString()
	}
	hash := sha256.Sum256([]byte(seed))
	bytes := hash[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

func syncClaudeCodeSessionHeader(req *http.Request, body []byte, injectMissing bool) {
	if req == nil {
		return
	}

	hasExistingHeader := getHeaderRaw(req.Header, "X-Claude-Code-Session-Id") != ""
	if !injectMissing && !hasExistingHeader {
		return
	}

	uid := gjson.GetBytes(body, "metadata.user_id").String()
	if uid == "" {
		return
	}
	parsed := ParseMetadataUserID(uid)
	if parsed == nil || parsed.SessionID == "" {
		return
	}
	setHeaderRaw(req.Header, "X-Claude-Code-Session-Id", parsed.SessionID)
}

// injectClaudeCodePrompt 在 system 开头注入 Claude Code 提示词
// 处理 null、字符串、数组三种格式
func injectClaudeCodePrompt(body []byte, system any) []byte {
	system = normalizeSystemParam(system)
	claudeCodePrompt := currentClaudeCodeSystemPrompt()
	claudeCodeBlock, err := marshalAnthropicSystemTextBlock(claudeCodePrompt, true)
	if err != nil {
		logger.LegacyPrintf("service.gateway", "Warning: failed to build Claude Code prompt block: %v", err)
		return body
	}
	claudeCodePrefix := strings.TrimSpace(claudeCodePrompt)

	var items [][]byte

	switch v := system.(type) {
	case nil:
		items = [][]byte{claudeCodeBlock}
	case string:
		if strings.TrimSpace(v) == "" || strings.TrimSpace(v) == strings.TrimSpace(claudeCodePrompt) {
			items = [][]byte{claudeCodeBlock}
		} else {
			merged := v
			if !strings.HasPrefix(v, claudeCodePrefix) {
				merged = claudeCodePrefix + "\n\n" + v
			}
			nextBlock, buildErr := marshalAnthropicSystemTextBlock(merged, false)
			if buildErr != nil {
				logger.LegacyPrintf("service.gateway", "Warning: failed to build prefixed Claude Code system block: %v", buildErr)
				return body
			}
			items = [][]byte{claudeCodeBlock, nextBlock}
		}
	case []any:
		items = make([][]byte, 0, len(v)+1)
		items = append(items, claudeCodeBlock)
		prefixedNext := false
		systemResult := gjson.GetBytes(body, "system")
		if systemResult.IsArray() {
			systemResult.ForEach(func(_, item gjson.Result) bool {
				textResult := item.Get("text")
				if textResult.Exists() && textResult.Type == gjson.String &&
					strings.TrimSpace(textResult.String()) == strings.TrimSpace(claudeCodePrompt) {
					return true
				}

				raw := []byte(item.Raw)
				if !prefixedNext && item.Get("type").String() == "text" && textResult.Exists() && textResult.Type == gjson.String {
					text := textResult.String()
					if strings.TrimSpace(text) != "" && !strings.HasPrefix(text, claudeCodePrefix) {
						next, setErr := sjson.SetBytes(raw, "text", claudeCodePrefix+"\n\n"+text)
						if setErr == nil {
							raw = next
							prefixedNext = true
						}
					}
				}
				items = append(items, raw)
				return true
			})
		} else {
			for _, item := range v {
				m, ok := item.(map[string]any)
				if !ok {
					raw, marshalErr := json.Marshal(item)
					if marshalErr == nil {
						items = append(items, raw)
					}
					continue
				}
				if text, ok := m["text"].(string); ok && strings.TrimSpace(text) == strings.TrimSpace(claudeCodePrompt) {
					continue
				}
				if !prefixedNext {
					if blockType, _ := m["type"].(string); blockType == "text" {
						if text, ok := m["text"].(string); ok && strings.TrimSpace(text) != "" && !strings.HasPrefix(text, claudeCodePrefix) {
							m["text"] = claudeCodePrefix + "\n\n" + text
							prefixedNext = true
						}
					}
				}
				raw, marshalErr := json.Marshal(m)
				if marshalErr == nil {
					items = append(items, raw)
				}
			}
		}
	default:
		items = [][]byte{claudeCodeBlock}
	}

	result, ok := setJSONRawBytes(body, "system", buildJSONArrayRaw(items))
	if !ok {
		logger.LegacyPrintf("service.gateway", "Warning: failed to inject Claude Code prompt")
		return body
	}
	return result
}

// computeAttributionFingerprint 计算 Claude Code 归因指纹。
// Algorithm: SHA256(SALT + msg[4] + msg[7] + msg[20] + version)[:3]
func computeAttributionFingerprint(firstMessageText, version string) string {
	if strings.TrimSpace(version) == "" {
		return ""
	}
	indices := claude.AttributionIndices()
	if len(indices) == 0 {
		return ""
	}
	chars := make([]byte, len(indices))
	for i, idx := range indices {
		if idx < len(firstMessageText) {
			chars[i] = firstMessageText[idx]
		} else {
			chars[i] = '0'
		}
	}
	salt := claude.AttributionSalt()
	if salt == "" {
		return ""
	}
	input := salt + string(chars) + version
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash[:2])[:3]
}

// extractFirstUserMessageText 从 Anthropic messages body 中提取首条用户消息文本。
func extractFirstUserMessageText(body []byte) string {
	messages := gjson.GetBytes(body, "messages")
	if !messages.IsArray() {
		return ""
	}
	var text string
	messages.ForEach(func(_, msg gjson.Result) bool {
		if msg.Get("role").String() != "user" {
			return true
		}
		content := msg.Get("content")
		if content.Type == gjson.String {
			text = content.String()
			return false
		}
		if content.IsArray() {
			content.ForEach(func(_, block gjson.Result) bool {
				if block.Get("type").String() == "text" {
					text = block.Get("text").String()
					return false
				}
				return true
			})
			return false
		}
		return true
	})
	return text
}

// buildAttributionHeaderText 构造归因 header 文本（注入到 system prompt 首个 text block）。
func buildAttributionHeaderText(body []byte, fingerprintUA string) string {
	version := ExtractCLIVersion(fingerprintUA)
	if version == "" {
		return ""
	}
	msgText := extractFirstUserMessageText(body)
	fp := computeAttributionFingerprint(msgText, version)
	if fp == "" {
		return ""
	}
	return fmt.Sprintf("x-anthropic-billing-header: cc_version=%s.%s; cc_entrypoint=cli;", version, fp)
}

// injectAttributionHeaderBlock 将归因 header 作为 system array 的第一个 text block 插入。
func injectAttributionHeaderBlock(body []byte, attrHeader string) []byte {
	if attrHeader == "" {
		return body
	}
	attrBlock, err := marshalAnthropicSystemTextBlock(attrHeader, false)
	if err != nil {
		return body
	}

	systemResult := gjson.GetBytes(body, "system")
	if !systemResult.Exists() || systemResult.Type == gjson.Null {
		result, ok := setJSONRawBytes(body, "system", buildJSONArrayRaw([][]byte{attrBlock}))
		if ok {
			return result
		}
		return body
	}

	if systemResult.IsArray() {
		var items [][]byte
		items = append(items, attrBlock)
		systemResult.ForEach(func(_, item gjson.Result) bool {
			if item.Get("type").String() == "text" {
				text := item.Get("text").String()
				if strings.HasPrefix(text, "x-anthropic-billing-header:") {
					return true
				}
			}
			items = append(items, []byte(item.Raw))
			return true
		})
		result, ok := setJSONRawBytes(body, "system", buildJSONArrayRaw(items))
		if ok {
			return result
		}
	}
	return body
}
