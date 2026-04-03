package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// isClaudeCodeClient 判断请求是否来自 Claude Code 客户端
// 简化判断：User-Agent 匹配 + metadata.user_id 存在
func isClaudeCodeClient(userAgent string, metadataUserID string) bool {
	if metadataUserID == "" {
		return false
	}
	return claudeCliUserAgentRe.MatchString(userAgent)
}

func isClaudeCodeRequest(ctx context.Context, c *gin.Context, parsed *ParsedRequest) bool {
	if IsClaudeCodeClient(ctx) {
		return true
	}
	if parsed == nil || c == nil {
		return false
	}
	return isClaudeCodeClient(c.GetHeader("User-Agent"), parsed.MetadataUserID)
}

// normalizeSystemParam 将 json.RawMessage 类型的 system 参数转为标准 Go 类型（string / []any / nil），
// 避免 type switch 中 json.RawMessage（底层 []byte）无法匹配 case string / case []any / case nil 的问题。
// 这是 Go 的 typed nil 陷阱：(json.RawMessage, nil) ≠ (nil, nil)。
func normalizeSystemParam(system any) any {
	raw, ok := system.(json.RawMessage)
	if !ok {
		return system
	}
	if len(raw) == 0 {
		return nil
	}
	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil
	}
	return parsed
}

// systemIncludesClaudeCodePrompt 检查 system 中是否已包含 Claude Code 提示词
// 使用前缀匹配支持多种变体（标准版、Agent SDK 版等）
func systemIncludesClaudeCodePrompt(system any) bool {
	system = normalizeSystemParam(system)
	switch v := system.(type) {
	case string:
		return hasClaudeCodePrefix(v)
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				if text, ok := m["text"].(string); ok && hasClaudeCodePrefix(text) {
					return true
				}
			}
		}
	}
	return false
}

// hasClaudeCodePrefix 检查文本是否以 Claude Code 提示词的特征前缀开头
func hasClaudeCodePrefix(text string) bool {
	for _, prefix := range currentClaudeCodePromptPrefixes() {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

func inferStreamingErrorStatusCode(httpStatus int, responseBody []byte) int {
	if code := extractStreamingErrorStatusCode(responseBody); code != 0 {
		return code
	}
	if httpStatus >= 400 {
		return httpStatus
	}
	if errType := strings.TrimSpace(gjson.GetBytes(responseBody, "error.type").String()); errType != "" {
		switch errType {
		case "authentication_error":
			return http.StatusUnauthorized
		case "permission_error":
			return http.StatusForbidden
		case "rate_limit_error":
			return http.StatusTooManyRequests
		case "not_found_error":
			return http.StatusNotFound
		case "invalid_request_error":
			return http.StatusBadRequest
		case "overloaded_error":
			return 529
		case "api_error":
			return http.StatusBadGateway
		}
	}
	// Claude SSE error events are delivered after the HTTP stream is already 200.
	// Preserve the prior failover behavior instead of downgrading them to 200.
	return http.StatusForbidden
}

func extractStreamingErrorStatusCode(responseBody []byte) int {
	if len(responseBody) == 0 {
		return 0
	}
	for _, path := range []string{
		"error.status_code",
		"error.statusCode",
		"status_code",
		"statusCode",
		"error.error_code",
	} {
		if code := gjson.GetBytes(responseBody, path); code.Exists() {
			if parsed := parseStreamingErrorStatusValue(code); parsed != 0 {
				return parsed
			}
		}
	}

	inner := strings.TrimSpace(gjson.GetBytes(responseBody, "error.message").String())
	if strings.HasPrefix(inner, "{") {
		return extractStreamingErrorStatusCode([]byte(inner))
	}
	return 0
}

func parseStreamingErrorStatusValue(value gjson.Result) int {
	switch value.Type {
	case gjson.Number:
		code := int(value.Int())
		if code > 0 {
			return code
		}
	case gjson.String:
		code, err := strconv.Atoi(strings.TrimSpace(value.String()))
		if err == nil && code > 0 {
			return code
		}
	}
	return 0
}
