package service

import (
	"bytes"
	"encoding/json"
	"unsafe"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func stripEmptyTextBlocksFromSlice(blocks []any) ([]any, bool) {
	var result []any
	changed := false
	for i, block := range blocks {
		blockMap, ok := block.(map[string]any)
		if !ok {
			if result != nil {
				result = append(result, block)
			}
			continue
		}
		blockType, _ := blockMap["type"].(string)

		if blockType == "text" {
			if txt, _ := blockMap["text"].(string); txt == "" {
				if result == nil {
					result = make([]any, 0, len(blocks))
					result = append(result, blocks[:i]...)
				}
				changed = true
				continue
			}
		}

		if blockType == "tool_result" {
			if nestedContent, ok := blockMap["content"].([]any); ok {
				if cleaned, nestedChanged := stripEmptyTextBlocksFromSlice(nestedContent); nestedChanged {
					if result == nil {
						result = make([]any, 0, len(blocks))
						result = append(result, blocks[:i]...)
					}
					changed = true
					blockCopy := make(map[string]any, len(blockMap))
					for key, value := range blockMap {
						blockCopy[key] = value
					}
					blockCopy["content"] = cleaned
					result = append(result, blockCopy)
					continue
				}
			}
		}

		if result != nil {
			result = append(result, block)
		}
	}
	if !changed {
		return blocks, false
	}
	return result, true
}

func StripEmptyTextBlocks(body []byte) []byte {
	hasEmptyTextBlock := bytes.Contains(body, patternEmptyText) ||
		bytes.Contains(body, patternEmptyTextSpaced) ||
		bytes.Contains(body, patternEmptyTextSp1) ||
		bytes.Contains(body, patternEmptyTextSp2)
	if !hasEmptyTextBlock {
		return body
	}

	jsonStr := *(*string)(unsafe.Pointer(&body))
	msgsRes := gjson.Get(jsonStr, "messages")
	if !msgsRes.Exists() || !msgsRes.IsArray() {
		return body
	}

	var messages []any
	if err := json.Unmarshal(sliceRawFromBody(body, msgsRes), &messages); err != nil {
		return body
	}

	modified := false
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]any)
		if !ok {
			continue
		}
		content, ok := msgMap["content"].([]any)
		if !ok {
			continue
		}
		if cleaned, changed := stripEmptyTextBlocksFromSlice(content); changed {
			modified = true
			msgMap["content"] = cleaned
		}
	}
	if !modified {
		return body
	}

	msgsBytes, err := json.Marshal(messages)
	if err != nil {
		return body
	}
	out, err := sjson.SetRawBytes(body, "messages", msgsBytes)
	if err != nil {
		return body
	}
	return out
}
