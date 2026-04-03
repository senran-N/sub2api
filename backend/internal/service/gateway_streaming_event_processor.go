package service

import (
	"encoding/json"
	"strings"
)

type processedStreamingEvent struct {
	outputBlocks []string
	data         string
	usagePatch   *sseUsagePatch
	terminal     bool
}

func (s *GatewayService) processAnthropicStreamingEvent(
	respStatus int,
	account *Account,
	originalModel string,
	mappedModel string,
	lines []string,
) (*processedStreamingEvent, error) {
	if len(lines) == 0 {
		return &processedStreamingEvent{}, nil
	}

	eventName := ""
	dataLine := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
			continue
		}
		if dataLine == "" && strings.HasPrefix(trimmed, "data:") {
			dataLine = strings.TrimLeft(trimmed[5:], " \t")
		}
	}

	if eventName == "error" {
		responseBody := []byte(strings.TrimSpace(dataLine))
		return nil, &upstreamStreamEventError{
			statusCode:   inferStreamingErrorStatusCode(respStatus, responseBody),
			responseBody: responseBody,
		}
	}

	if dataLine == "" {
		return &processedStreamingEvent{
			outputBlocks: []string{strings.Join(lines, "\n") + "\n\n"},
		}, nil
	}

	if dataLine == "[DONE]" {
		return &processedStreamingEvent{
			outputBlocks: []string{renderStreamingEventBlock(eventName, dataLine)},
			data:         dataLine,
			terminal:     true,
		}, nil
	}

	var event map[string]any
	if err := json.Unmarshal([]byte(dataLine), &event); err != nil {
		return &processedStreamingEvent{
			outputBlocks: []string{renderStreamingEventBlock(eventName, dataLine)},
			data:         dataLine,
		}, nil
	}

	eventType, _ := event["type"].(string)
	if eventName == "" {
		eventName = eventType
	}

	eventChanged := s.rewriteAnthropicStreamingEventUsage(eventType, event, account)
	if originalModel != mappedModel {
		if s.rewriteAnthropicStreamingEventModel(event, mappedModel, originalModel) {
			eventChanged = true
		}
	}

	usagePatch := s.extractSSEUsagePatch(event)
	terminal := anthropicStreamEventIsTerminal(eventName, dataLine)
	if !eventChanged {
		return &processedStreamingEvent{
			outputBlocks: []string{renderStreamingEventBlock(eventName, dataLine)},
			data:         dataLine,
			usagePatch:   usagePatch,
			terminal:     terminal,
		}, nil
	}

	newData, err := json.Marshal(event)
	if err != nil {
		return &processedStreamingEvent{
			outputBlocks: []string{renderStreamingEventBlock(eventName, dataLine)},
			data:         dataLine,
			usagePatch:   usagePatch,
			terminal:     terminal,
		}, nil
	}

	return &processedStreamingEvent{
		outputBlocks: []string{renderStreamingEventBlock(eventName, string(newData))},
		data:         string(newData),
		usagePatch:   usagePatch,
		terminal:     terminal,
	}, nil
}

func (s *GatewayService) rewriteAnthropicStreamingEventUsage(eventType string, event map[string]any, account *Account) bool {
	eventChanged := false

	if eventType == "message_start" {
		if message, ok := event["message"].(map[string]any); ok {
			if usage, ok := message["usage"].(map[string]any); ok {
				eventChanged = reconcileCachedTokens(usage) || eventChanged
				if account != nil && account.IsCacheTTLOverrideEnabled() {
					eventChanged = rewriteCacheCreationJSON(usage, account.GetCacheTTLOverrideTarget()) || eventChanged
				}
			}
		}
	}

	if eventType == "message_delta" {
		if usage, ok := event["usage"].(map[string]any); ok {
			eventChanged = reconcileCachedTokens(usage) || eventChanged
			if account != nil && account.IsCacheTTLOverrideEnabled() {
				eventChanged = rewriteCacheCreationJSON(usage, account.GetCacheTTLOverrideTarget()) || eventChanged
			}
		}
	}

	return eventChanged
}

func (s *GatewayService) rewriteAnthropicStreamingEventModel(event map[string]any, mappedModel string, originalModel string) bool {
	message, ok := event["message"].(map[string]any)
	if !ok {
		return false
	}
	model, ok := message["model"].(string)
	if !ok || model != mappedModel {
		return false
	}
	message["model"] = originalModel
	return true
}

func renderStreamingEventBlock(eventName string, data string) string {
	block := ""
	if eventName != "" {
		block = "event: " + eventName + "\n"
	}
	block += "data: " + data + "\n\n"
	return block
}
