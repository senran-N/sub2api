package service

import (
	"encoding/json"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/antigravity"
)

// extractImageSize 从 Gemini 请求中提取 image_size 参数
func (s *AntigravityGatewayService) extractImageSize(body []byte) string {
	var req antigravity.GeminiRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return "2K"
	}

	if req.GenerationConfig != nil && req.GenerationConfig.ImageConfig != nil {
		size := strings.ToUpper(strings.TrimSpace(req.GenerationConfig.ImageConfig.ImageSize))
		if size == "1K" || size == "2K" || size == "4K" {
			return size
		}
	}

	return "2K"
}

// isImageGenerationModel 判断模型是否为图片生成模型
// 支持的模型：gemini-3.1-flash-image, gemini-3-pro-image, gemini-2.5-flash-image 等
func isImageGenerationModel(model string) bool {
	modelLower := strings.ToLower(model)
	modelLower = strings.TrimPrefix(modelLower, "models/")

	return modelLower == "gemini-3.1-flash-image" ||
		modelLower == "gemini-3.1-flash-image-preview" ||
		strings.HasPrefix(modelLower, "gemini-3.1-flash-image-") ||
		modelLower == "gemini-3-pro-image" ||
		modelLower == "gemini-3-pro-image-preview" ||
		strings.HasPrefix(modelLower, "gemini-3-pro-image-") ||
		modelLower == "gemini-2.5-flash-image" ||
		modelLower == "gemini-2.5-flash-image-preview" ||
		strings.HasPrefix(modelLower, "gemini-2.5-flash-image-")
}

// cleanGeminiRequest 清理 Gemini 请求体中的 Schema
func cleanGeminiRequest(body []byte) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	modified := false
	if tools, ok := payload["tools"].([]any); ok && len(tools) > 0 {
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]any)
			if !ok {
				continue
			}

			var funcs []any
			if value, ok := toolMap["functionDeclarations"].([]any); ok {
				funcs = value
			} else if value, ok := toolMap["function_declarations"].([]any); ok {
				funcs = value
			}
			if len(funcs) == 0 {
				continue
			}

			for _, function := range funcs {
				funcMap, ok := function.(map[string]any)
				if !ok {
					continue
				}
				if params, ok := funcMap["parameters"].(map[string]any); ok {
					antigravity.DeepCleanUndefined(params)
					funcMap["parameters"] = antigravity.CleanJSONSchema(params)
					modified = true
				}
			}
		}
	}

	if !modified {
		return body, nil
	}

	return json.Marshal(payload)
}

// filterEmptyPartsFromGeminiRequest 过滤掉 parts 为空的消息
// Gemini API 不接受空 parts，需要在请求前过滤
func filterEmptyPartsFromGeminiRequest(body []byte) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	contents, ok := payload["contents"].([]any)
	if !ok || len(contents) == 0 {
		return body, nil
	}

	filtered := make([]any, 0, len(contents))
	modified := false
	for _, content := range contents {
		contentMap, ok := content.(map[string]any)
		if !ok {
			filtered = append(filtered, content)
			continue
		}

		parts, hasParts := contentMap["parts"]
		if !hasParts {
			filtered = append(filtered, content)
			continue
		}

		partSlice, ok := parts.([]any)
		if !ok {
			filtered = append(filtered, content)
			continue
		}

		if len(partSlice) == 0 {
			modified = true
			continue
		}

		filtered = append(filtered, content)
	}

	if !modified {
		return body, nil
	}

	payload["contents"] = filtered
	return json.Marshal(payload)
}
