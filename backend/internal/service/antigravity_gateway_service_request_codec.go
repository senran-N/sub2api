package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/senran-N/sub2api/internal/pkg/antigravity"
	"github.com/tidwall/gjson"
)

func (s *AntigravityGatewayService) getClaudeTransformOptions(ctx context.Context) antigravity.TransformOptions {
	opts := antigravity.DefaultTransformOptions()
	if s.settingService == nil {
		return opts
	}
	opts.EnableIdentityPatch = s.settingService.IsIdentityPatchEnabled(ctx)
	opts.IdentityPatch = s.settingService.GetIdentityPatchPrompt(ctx)
	return opts
}

// extractTextFromSSEResponse 从 SSE 流式响应中提取文本
func extractTextFromSSEResponse(respBody []byte) string {
	var texts []string
	lines := bytes.Split(respBody, []byte("\n"))

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if bytes.HasPrefix(line, []byte("data:")) {
			line = bytes.TrimPrefix(line, []byte("data:"))
			line = bytes.TrimSpace(line)
		}
		if len(line) == 0 || line[0] != '{' {
			continue
		}

		var data map[string]any
		if err := json.Unmarshal(line, &data); err != nil {
			continue
		}

		response, ok := data["response"].(map[string]any)
		if !ok {
			response = data
		}

		candidates, ok := response["candidates"].([]any)
		if !ok || len(candidates) == 0 {
			continue
		}
		candidate, ok := candidates[0].(map[string]any)
		if !ok {
			continue
		}
		content, ok := candidate["content"].(map[string]any)
		if !ok {
			continue
		}
		parts, ok := content["parts"].([]any)
		if !ok {
			continue
		}

		for _, part := range parts {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := partMap["text"].(string); ok && text != "" {
				texts = append(texts, text)
			}
		}
	}

	return strings.Join(texts, "")
}

// injectIdentityPatchToGeminiRequest 为 Gemini 格式请求注入身份提示词
// 如果请求中已包含 "You are Antigravity" 则不重复注入
func injectIdentityPatchToGeminiRequest(body []byte) ([]byte, error) {
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, fmt.Errorf("解析 Gemini 请求失败: %w", err)
	}

	if sysInst, ok := request["systemInstruction"].(map[string]any); ok {
		if parts, ok := sysInst["parts"].([]any); ok {
			for _, part := range parts {
				partMap, ok := part.(map[string]any)
				if !ok {
					continue
				}
				if text, ok := partMap["text"].(string); ok && strings.Contains(text, "You are Antigravity") {
					return body, nil
				}
			}
		}
	}

	identityPatch := antigravity.GetDefaultIdentityPatch()
	newPart := map[string]any{"text": identityPatch}

	if existing, ok := request["systemInstruction"].(map[string]any); ok {
		if parts, ok := existing["parts"].([]any); ok {
			existing["parts"] = append([]any{newPart}, parts...)
		} else {
			existing["parts"] = []any{newPart}
		}
	} else {
		request["systemInstruction"] = map[string]any{
			"parts": []any{newPart},
		}
	}

	return json.Marshal(request)
}

// wrapV1InternalRequest 包装请求为 v1internal 格式
func (s *AntigravityGatewayService) wrapV1InternalRequest(projectID, model string, originalBody []byte) ([]byte, error) {
	var request any
	if err := json.Unmarshal(originalBody, &request); err != nil {
		return nil, fmt.Errorf("解析请求体失败: %w", err)
	}

	wrapped := map[string]any{
		"project":     projectID,
		"requestId":   "agent-" + uuid.New().String(),
		"userAgent":   "antigravity",
		"requestType": "agent",
		"model":       model,
		"request":     request,
	}

	return json.Marshal(wrapped)
}

// unwrapV1InternalResponse 解包 v1internal 响应
// 使用 gjson 零拷贝提取 response 字段，避免 Unmarshal+Marshal 双重开销
func (s *AntigravityGatewayService) unwrapV1InternalResponse(body []byte) ([]byte, error) {
	result := gjson.GetBytes(body, "response")
	if result.Exists() {
		return []byte(result.Raw), nil
	}
	return body, nil
}

// isModelNotFoundError 检测是否为模型不存在的 404 错误
func isModelNotFoundError(statusCode int, body []byte) bool {
	if statusCode != 404 {
		return false
	}

	bodyStr := strings.ToLower(string(body))
	keywords := []string{"model not found", "unknown model", "not found"}
	for _, keyword := range keywords {
		if strings.Contains(bodyStr, keyword) {
			return true
		}
	}
	return true
}
