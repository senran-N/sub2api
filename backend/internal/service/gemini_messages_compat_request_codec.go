package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

func ensureGeminiFunctionCallThoughtSignatures(body []byte) []byte {
	// Fast path: only run when functionCall is present.
	if !bytes.Contains(body, []byte(`"functionCall"`)) {
		return body
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body
	}

	contentsAny, ok := payload["contents"].([]any)
	if !ok || len(contentsAny) == 0 {
		return body
	}

	modified := false
	for _, c := range contentsAny {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		partsAny, ok := cm["parts"].([]any)
		if !ok || len(partsAny) == 0 {
			continue
		}
		for _, p := range partsAny {
			pm, ok := p.(map[string]any)
			if !ok || pm == nil {
				continue
			}
			if fc, ok := pm["functionCall"].(map[string]any); !ok || fc == nil {
				continue
			}
			ts, _ := pm["thoughtSignature"].(string)
			if strings.TrimSpace(ts) == "" {
				pm["thoughtSignature"] = geminiDummyThoughtSignature
				modified = true
			}
		}
	}

	if !modified {
		return body
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return body
	}
	return encoded
}

func extractGeminiFinishReason(geminiResp map[string]any) string {
	if candidates, ok := geminiResp["candidates"].([]any); ok && len(candidates) > 0 {
		if cand, ok := candidates[0].(map[string]any); ok {
			if finishReason, ok := cand["finishReason"].(string); ok {
				return finishReason
			}
		}
	}
	return ""
}

func extractGeminiParts(geminiResp map[string]any) []map[string]any {
	if candidates, ok := geminiResp["candidates"].([]any); ok && len(candidates) > 0 {
		if cand, ok := candidates[0].(map[string]any); ok {
			if content, ok := cand["content"].(map[string]any); ok {
				if partsAny, ok := content["parts"].([]any); ok && len(partsAny) > 0 {
					parts := make([]map[string]any, 0, len(partsAny))
					for _, part := range partsAny {
						partMap, ok := part.(map[string]any)
						if !ok {
							continue
						}
						parts = append(parts, partMap)
					}
					return parts
				}
			}
		}
	}
	return nil
}

func computeGeminiTextDelta(seen, incoming string) (delta, newSeen string) {
	incoming = strings.TrimSuffix(incoming, "\u0000")
	if incoming == "" {
		return "", seen
	}

	if strings.HasPrefix(incoming, seen) {
		return strings.TrimPrefix(incoming, seen), incoming
	}
	if strings.HasPrefix(seen, incoming) {
		return "", seen
	}
	return incoming, seen + incoming
}

func mapGeminiFinishReasonToClaudeStopReason(finishReason string) string {
	switch strings.ToUpper(strings.TrimSpace(finishReason)) {
	case "MAX_TOKENS":
		return "max_tokens"
	case "STOP":
		return "end_turn"
	default:
		return "end_turn"
	}
}

func convertClaudeMessagesToGeminiGenerateContent(body []byte) ([]byte, error) {
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, err
	}

	toolUseIDToName := make(map[string]string)
	systemText := extractClaudeSystemText(request["system"])
	contents, err := convertClaudeMessagesToGeminiContents(request["messages"], toolUseIDToName)
	if err != nil {
		return nil, err
	}

	output := make(map[string]any)
	if systemText != "" {
		output["systemInstruction"] = map[string]any{
			"parts": []any{map[string]any{"text": systemText}},
		}
	}
	output["contents"] = contents

	if tools := convertClaudeToolsToGeminiTools(request["tools"]); tools != nil {
		output["tools"] = tools
	}

	generationConfig := convertClaudeGenerationConfig(request)
	if generationConfig != nil {
		output["generationConfig"] = generationConfig
	}

	stripGeminiFunctionIDs(output)
	return json.Marshal(output)
}

func stripGeminiFunctionIDs(req map[string]any) {
	contents, ok := req["contents"].([]any)
	if !ok {
		return
	}
	for _, content := range contents {
		contentMap, ok := content.(map[string]any)
		if !ok {
			continue
		}
		contentParts, ok := contentMap["parts"].([]any)
		if !ok {
			continue
		}
		for _, part := range contentParts {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if functionCall, ok := partMap["functionCall"].(map[string]any); ok && functionCall != nil {
				delete(functionCall, "id")
			}
			if functionResponse, ok := partMap["functionResponse"].(map[string]any); ok && functionResponse != nil {
				delete(functionResponse, "id")
			}
		}
	}
}

func extractClaudeSystemText(system any) string {
	switch value := system.(type) {
	case string:
		return strings.TrimSpace(value)
	case []any:
		var parts []string
		for _, part := range value {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if partType, _ := partMap["type"].(string); partType != "text" {
				continue
			}
			if text, ok := partMap["text"].(string); ok && strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		return ""
	}
}

func convertClaudeMessagesToGeminiContents(messages any, toolUseIDToName map[string]string) ([]any, error) {
	messageList, ok := messages.([]any)
	if !ok {
		return nil, errors.New("messages must be an array")
	}

	contents := make([]any, 0, len(messageList))
	for _, message := range messageList {
		messageMap, ok := message.(map[string]any)
		if !ok {
			continue
		}

		role, _ := messageMap["role"].(string)
		role = strings.ToLower(strings.TrimSpace(role))
		geminiRole := "user"
		if role == "assistant" {
			geminiRole = "model"
		}

		parts := make([]any, 0)
		switch content := messageMap["content"].(type) {
		case string:
			parts = append(parts, map[string]any{"text": content})
		case []any:
			singleBlock := len(content) == 1

			for _, block := range content {
				blockMap, ok := block.(map[string]any)
				if !ok {
					continue
				}
				blockType, _ := blockMap["type"].(string)
				switch blockType {
				case "text":
					if text, ok := blockMap["text"].(string); ok {
						if singleBlock || strings.TrimSpace(text) != "" {
							parts = append(parts, map[string]any{"text": text})
						}
					}
				case "tool_use":
					id, _ := blockMap["id"].(string)
					name, _ := blockMap["name"].(string)
					if strings.TrimSpace(id) != "" && strings.TrimSpace(name) != "" {
						toolUseIDToName[id] = name
					}
					signature, _ := blockMap["signature"].(string)
					signature = strings.TrimSpace(signature)
					if signature == "" {
						signature = geminiDummyThoughtSignature
					}
					parts = append(parts, map[string]any{
						"thoughtSignature": signature,
						"functionCall": map[string]any{
							"name": name,
							"args": blockMap["input"],
						},
					})
				case "tool_result":
					toolUseID, _ := blockMap["tool_use_id"].(string)
					name := toolUseIDToName[toolUseID]
					if name == "" {
						name = "tool"
					}
					parts = append(parts, map[string]any{
						"functionResponse": map[string]any{
							"name": name,
							"response": map[string]any{
								"content": extractClaudeContentText(blockMap["content"]),
							},
						},
					})
				case "image":
					if source, ok := blockMap["source"].(map[string]any); ok {
						if sourceType, _ := source["type"].(string); sourceType == "base64" {
							mediaType, _ := source["media_type"].(string)
							data, _ := source["data"].(string)
							if mediaType != "" && data != "" {
								parts = append(parts, map[string]any{
									"inlineData": map[string]any{
										"mimeType": mediaType,
										"data":     data,
									},
								})
							}
						}
					}
				default:
					if encoded, err := json.Marshal(blockMap); err == nil {
						parts = append(parts, map[string]any{"text": string(encoded)})
					}
				}
			}
		}

		contents = append(contents, map[string]any{
			"role":  geminiRole,
			"parts": parts,
		})
	}
	return contents, nil
}

func extractClaudeContentText(value any) string {
	switch content := value.(type) {
	case string:
		return content
	case []any:
		var builder strings.Builder
		for _, part := range content {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if partMap["type"] == "text" {
				if text, ok := partMap["text"].(string); ok {
					_, _ = builder.WriteString(text)
				}
			}
		}
		return builder.String()
	default:
		encoded, _ := json.Marshal(content)
		return string(encoded)
	}
}

func convertClaudeToolsToGeminiTools(tools any) []any {
	toolList, ok := tools.([]any)
	if !ok || len(toolList) == 0 {
		return nil
	}

	functionDeclarations := make([]any, 0, len(toolList))
	for _, tool := range toolList {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}

		var name string
		var description string
		var parameters any

		toolType, _ := toolMap["type"].(string)
		if toolType == "custom" {
			customConfig, ok := toolMap["custom"].(map[string]any)
			if !ok {
				continue
			}
			name, _ = toolMap["name"].(string)
			description, _ = customConfig["description"].(string)
			parameters = customConfig["input_schema"]
		} else {
			name, _ = toolMap["name"].(string)
			description, _ = toolMap["description"].(string)
			parameters = toolMap["input_schema"]
		}

		if name == "" {
			continue
		}
		if parameters == nil {
			parameters = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}
		}

		functionDeclarations = append(functionDeclarations, map[string]any{
			"name":        name,
			"description": description,
			"parameters":  cleanToolSchema(parameters),
		})
	}

	if len(functionDeclarations) == 0 {
		return nil
	}
	return []any{
		map[string]any{
			"functionDeclarations": functionDeclarations,
		},
	}
}

func cleanToolSchema(schema any) any {
	if schema == nil {
		return nil
	}

	switch value := schema.(type) {
	case map[string]any:
		cleaned := make(map[string]any)
		for key, item := range value {
			if key == "$schema" || key == "$id" || key == "$ref" ||
				key == "additionalProperties" || key == "patternProperties" || key == "minLength" ||
				key == "maxLength" || key == "minItems" || key == "maxItems" {
				continue
			}
			cleaned[key] = cleanToolSchema(item)
		}
		if typeValue, ok := cleaned["type"].(string); ok {
			cleaned["type"] = strings.ToUpper(typeValue)
		}
		return cleaned
	case []any:
		cleaned := make([]any, len(value))
		for index, item := range value {
			cleaned[index] = cleanToolSchema(item)
		}
		return cleaned
	default:
		return value
	}
}

func convertClaudeGenerationConfig(req map[string]any) map[string]any {
	config := make(map[string]any)
	if maxTokens, ok := asInt(req["max_tokens"]); ok && maxTokens > 0 {
		config["maxOutputTokens"] = maxTokens
	}
	if temperature, ok := req["temperature"].(float64); ok {
		config["temperature"] = temperature
	}
	if topP, ok := req["top_p"].(float64); ok {
		config["topP"] = topP
	}
	if stopSequences, ok := req["stop_sequences"].([]any); ok && len(stopSequences) > 0 {
		config["stopSequences"] = stopSequences
	}
	if len(config) == 0 {
		return nil
	}
	return config
}

func (s *GeminiMessagesCompatService) extractImageSize(body []byte) string {
	var req struct {
		GenerationConfig *struct {
			ImageConfig *struct {
				ImageSize string `json:"imageSize"`
			} `json:"imageConfig"`
		} `json:"generationConfig"`
	}
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
