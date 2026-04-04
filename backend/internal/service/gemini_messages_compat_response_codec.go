package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/tidwall/gjson"
)

func unwrapIfNeeded(isOAuth bool, raw []byte) []byte {
	if !isOAuth {
		return raw
	}
	inner, err := unwrapGeminiResponse(raw)
	if err != nil {
		return raw
	}
	return inner
}

func collectGeminiSSE(body io.Reader, isOAuth bool) (map[string]any, *ClaudeUsage, error) {
	reader := bufio.NewReader(body)

	var last map[string]any
	var lastWithParts map[string]any
	var collectedTextParts []string
	usage := &ClaudeUsage{}

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(trimmed, "data:") {
				payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				switch payload {
				case "", "[DONE]":
					if payload == "[DONE]" {
						return mergeCollectedTextParts(pickGeminiCollectResult(last, lastWithParts), collectedTextParts), usage, nil
					}
				default:
					var parsed map[string]any
					var rawBytes []byte
					if isOAuth {
						innerBytes, err := unwrapGeminiResponse([]byte(payload))
						if err == nil {
							rawBytes = innerBytes
							_ = json.Unmarshal(innerBytes, &parsed)
						}
					} else {
						rawBytes = []byte(payload)
						_ = json.Unmarshal(rawBytes, &parsed)
					}
					if parsed != nil {
						last = parsed
						if parsedUsage := extractGeminiUsage(rawBytes); parsedUsage != nil {
							usage = parsedUsage
						}
						if parts := extractGeminiParts(parsed); len(parts) > 0 {
							lastWithParts = parsed
							for _, part := range parts {
								if text, ok := part["text"].(string); ok && text != "" {
									collectedTextParts = append(collectedTextParts, text)
								}
							}
						}
					}
				}
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, err
		}
	}

	return mergeCollectedTextParts(pickGeminiCollectResult(last, lastWithParts), collectedTextParts), usage, nil
}

func pickGeminiCollectResult(last map[string]any, lastWithParts map[string]any) map[string]any {
	if lastWithParts != nil {
		return lastWithParts
	}
	if last != nil {
		return last
	}
	return map[string]any{}
}

func mergeCollectedTextParts(response map[string]any, textParts []string) map[string]any {
	if len(textParts) == 0 {
		return response
	}

	mergedText := strings.Join(textParts, "")
	result := make(map[string]any, len(response))
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

	existingParts, ok := content["parts"].([]any)
	if !ok {
		existingParts = []any{}
	}

	newParts := make([]any, 0, len(existingParts)+1)
	textUpdated := false
	for _, part := range existingParts {
		partMap, ok := part.(map[string]any)
		if !ok {
			newParts = append(newParts, part)
			continue
		}
		if _, hasText := partMap["text"]; hasText && !textUpdated {
			newPart := make(map[string]any, len(partMap))
			for key, value := range partMap {
				newPart[key] = value
			}
			newPart["text"] = mergedText
			newParts = append(newParts, newPart)
			textUpdated = true
			continue
		}
		newParts = append(newParts, partMap)
	}

	if !textUpdated {
		newParts = append([]any{map[string]any{"text": mergedText}}, newParts...)
	}

	content["parts"] = newParts
	result["candidates"] = candidates
	return result
}

func estimateGeminiCountTokens(reqBody []byte) int {
	total := 0

	gjson.GetBytes(reqBody, "systemInstruction.parts").ForEach(func(_, part gjson.Result) bool {
		if text := strings.TrimSpace(part.Get("text").String()); text != "" {
			total += estimateTokensForText(text)
		}
		return true
	})

	gjson.GetBytes(reqBody, "contents").ForEach(func(_, content gjson.Result) bool {
		content.Get("parts").ForEach(func(_, part gjson.Result) bool {
			if text := strings.TrimSpace(part.Get("text").String()); text != "" {
				total += estimateTokensForText(text)
			}
			return true
		})
		return true
	})

	if total < 0 {
		return 0
	}
	return total
}

func estimateTokensForText(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return 0
	}
	asciiCount := 0
	for _, r := range runes {
		if r <= 0x7f {
			asciiCount++
		}
	}
	asciiRatio := float64(asciiCount) / float64(len(runes))
	if asciiRatio >= 0.8 {
		return (len(runes) + 3) / 4
	}
	return len(runes)
}

func unwrapGeminiResponse(raw []byte) ([]byte, error) {
	result := gjson.GetBytes(raw, "response")
	if result.Exists() && result.Type == gjson.JSON {
		return []byte(result.Raw), nil
	}
	return raw, nil
}

func convertGeminiToClaudeMessage(geminiResp map[string]any, originalModel string, rawData []byte) (map[string]any, *ClaudeUsage) {
	usage := extractGeminiUsage(rawData)
	if usage == nil {
		usage = &ClaudeUsage{}
	}

	contentBlocks := make([]any, 0)
	sawToolUse := false
	if candidates, ok := geminiResp["candidates"].([]any); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]any); ok {
			if content, ok := candidate["content"].(map[string]any); ok {
				if parts, ok := content["parts"].([]any); ok {
					for _, part := range parts {
						partMap, ok := part.(map[string]any)
						if !ok {
							continue
						}
						if text, ok := partMap["text"].(string); ok && text != "" {
							contentBlocks = append(contentBlocks, map[string]any{
								"type": "text",
								"text": text,
							})
						}
						if functionCall, ok := partMap["functionCall"].(map[string]any); ok {
							name, _ := functionCall["name"].(string)
							if strings.TrimSpace(name) == "" {
								name = "tool"
							}
							sawToolUse = true
							contentBlocks = append(contentBlocks, map[string]any{
								"type":  "tool_use",
								"id":    "toolu_" + randomHex(8),
								"name":  name,
								"input": functionCall["args"],
							})
						}
					}
				}
			}
		}
	}

	stopReason := mapGeminiFinishReasonToClaudeStopReason(extractGeminiFinishReason(geminiResp))
	if sawToolUse {
		stopReason = "tool_use"
	}

	response := map[string]any{
		"id":            "msg_" + randomHex(12),
		"type":          "message",
		"role":          "assistant",
		"model":         originalModel,
		"content":       contentBlocks,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]any{
			"input_tokens":  usage.InputTokens,
			"output_tokens": usage.OutputTokens,
		},
	}

	return response, usage
}

func extractGeminiUsage(data []byte) *ClaudeUsage {
	usage := gjson.GetBytes(data, "usageMetadata")
	if !usage.Exists() {
		return nil
	}
	promptTokens := int(usage.Get("promptTokenCount").Int())
	candidateTokens := int(usage.Get("candidatesTokenCount").Int())
	cachedTokens := int(usage.Get("cachedContentTokenCount").Int())
	thoughtTokens := int(usage.Get("thoughtsTokenCount").Int())

	return &ClaudeUsage{
		InputTokens:          promptTokens - cachedTokens,
		OutputTokens:         candidateTokens + thoughtTokens,
		CacheReadInputTokens: cachedTokens,
	}
}
