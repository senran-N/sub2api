package service

import "strings"

// getOrCreateGeminiParts 获取 Gemini 响应的 parts 结构，返回浅拷贝和更新回调。
func getOrCreateGeminiParts(response map[string]any) (result map[string]any, existingParts []any, setParts func([]any)) {
	result = make(map[string]any, len(response))
	for key, value := range response {
		result[key] = value
	}

	candidates, ok := result["candidates"].([]any)
	if !ok || len(candidates) == 0 {
		candidates = []any{map[string]any{}}
	}

	candidate, ok := candidates[0].(map[string]any)
	if !ok {
		candidate = make(map[string]any)
		candidates[0] = candidate
	}

	content, ok := candidate["content"].(map[string]any)
	if !ok {
		content = map[string]any{"role": "model"}
		candidate["content"] = content
	}

	existingParts, ok = content["parts"].([]any)
	if !ok {
		existingParts = []any{}
	}

	setParts = func(parts []any) {
		content["parts"] = parts
		result["candidates"] = candidates
	}

	return result, existingParts, setParts
}

// mergeCollectedPartsToResponse 将收集的所有 parts 合并到 Gemini 响应中。
// 保持原始顺序，仅合并连续的普通 text parts。
func mergeCollectedPartsToResponse(response map[string]any, collectedParts []map[string]any) map[string]any {
	if len(collectedParts) == 0 {
		return response
	}

	result, _, setParts := getOrCreateGeminiParts(response)

	var mergedParts []any
	var textBuffer strings.Builder
	flushTextBuffer := func() {
		if textBuffer.Len() == 0 {
			return
		}
		mergedParts = append(mergedParts, map[string]any{
			"text": textBuffer.String(),
		})
		textBuffer.Reset()
	}

	for _, part := range collectedParts {
		text, hasText := part["text"].(string)
		if hasText {
			if thought, _ := part["thought"].(bool); thought {
				flushTextBuffer()
				mergedParts = append(mergedParts, part)
				continue
			}

			_, _ = textBuffer.WriteString(text)
			continue
		}

		flushTextBuffer()
		mergedParts = append(mergedParts, part)
	}

	flushTextBuffer()
	setParts(mergedParts)
	return result
}

// mergeImagePartsToResponse 将收集到的图片 parts 合并到 Gemini 响应中。
func mergeImagePartsToResponse(response map[string]any, imageParts []map[string]any) map[string]any {
	if len(imageParts) == 0 {
		return response
	}

	result, existingParts, setParts := getOrCreateGeminiParts(response)
	for _, part := range existingParts {
		partMap, ok := part.(map[string]any)
		if !ok {
			continue
		}
		if _, hasInlineData := partMap["inlineData"]; hasInlineData {
			return result
		}
	}

	for _, imagePart := range imageParts {
		existingParts = append(existingParts, imagePart)
	}
	setParts(existingParts)
	return result
}

// mergeTextPartsToResponse 将收集到的文本合并到 Gemini 响应中。
func mergeTextPartsToResponse(response map[string]any, textParts []string) map[string]any {
	if len(textParts) == 0 {
		return response
	}

	mergedText := strings.Join(textParts, "")
	result, existingParts, setParts := getOrCreateGeminiParts(response)

	newParts := make([]any, 0, len(existingParts)+1)
	textUpdated := false
	for _, part := range existingParts {
		partMap, ok := part.(map[string]any)
		if !ok {
			newParts = append(newParts, part)
			continue
		}

		if _, hasText := partMap["text"]; hasText && !textUpdated {
			updatedPart := make(map[string]any, len(partMap))
			for key, value := range partMap {
				updatedPart[key] = value
			}
			updatedPart["text"] = mergedText
			newParts = append(newParts, updatedPart)
			textUpdated = true
			continue
		}

		newParts = append(newParts, partMap)
	}

	if !textUpdated {
		newParts = append([]any{map[string]any{"text": mergedText}}, newParts...)
	}

	setParts(newParts)
	return result
}
