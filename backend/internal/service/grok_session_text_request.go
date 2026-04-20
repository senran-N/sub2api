package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/senran-N/sub2api/internal/pkg/grok"
)

const (
	grokSessionModeAuto   = "auto"
	grokSessionModeFast   = "fast"
	grokSessionModeExpert = "expert"
	grokSessionModeHeavy  = "heavy"
)

type grokSessionTextRequest struct {
	ModelID         string
	ModeID          string
	SystemPrompt    string
	Message         string
	FileAttachments []string
	ImageInputs     []grokSessionUploadInput
	ToolNames       []string
	ToolPrompt      string
}

func buildGrokSessionTextPayload(input grokSessionTextRequest) (map[string]any, error) {
	message := strings.TrimSpace(input.Message)
	if message == "" {
		return nil, errors.New("grok session payload requires a message")
	}

	modeID := strings.TrimSpace(input.ModeID)
	if modeID == "" {
		var err error
		modeID, err = resolveGrokSessionModeID(input.ModelID)
		if err != nil {
			return nil, err
		}
	}

	payload := map[string]any{
		"collectionIds":               []any{},
		"connectors":                  []any{},
		"deviceEnvInfo":               grokSessionDefaultDeviceEnvInfo(),
		"disableMemory":               true,
		"disableSearch":               false,
		"disableSelfHarmShortCircuit": false,
		"disableTextFollowUps":        false,
		"enableImageGeneration":       false,
		"enableImageStreaming":        false,
		"enableSideBySide":            true,
		"fileAttachments":             []any{},
		"forceConcise":                false,
		"forceSideBySide":             false,
		"imageAttachments":            []any{},
		"imageGenerationCount":        1,
		"isAsyncChat":                 false,
		"message":                     message,
		"modeId":                      modeID,
		"responseMetadata":            map[string]any{},
		"returnImageBytes":            false,
		"returnRawGrokInXaiRequest":   false,
		"searchAllConnectors":         false,
		"sendFinalMetadata":           true,
		"temporary":                   true,
		"toolOverrides":               grokSessionDefaultToolOverrides(),
	}
	if systemPrompt := strings.TrimSpace(mergeGrokSessionSystemPrompt(input.SystemPrompt, input.ToolPrompt)); systemPrompt != "" {
		payload["customPersonality"] = systemPrompt
	}
	if len(input.FileAttachments) > 0 {
		attachments := make([]any, 0, len(input.FileAttachments))
		for _, attachment := range input.FileAttachments {
			if trimmed := strings.TrimSpace(attachment); trimmed != "" {
				attachments = append(attachments, trimmed)
			}
		}
		payload["fileAttachments"] = attachments
	}

	return payload, nil
}

func buildGrokSessionTextPayloadFromResponsesRequest(req *apicompat.ResponsesRequest) (map[string]any, error) {
	request, err := grokSessionTextRequestFromResponsesRequest(req)
	if err != nil {
		return nil, err
	}
	return buildGrokSessionTextPayload(request)
}

func buildGrokSessionTextPayloadFromChatCompletionsRequest(req *apicompat.ChatCompletionsRequest) (map[string]any, error) {
	if req == nil {
		return nil, errors.New("chat completions request is nil")
	}

	responsesReq, err := apicompat.ChatCompletionsToResponses(req)
	if err != nil {
		return nil, fmt.Errorf("convert chat completions to responses: %w", err)
	}
	return buildGrokSessionTextPayloadFromResponsesRequest(responsesReq)
}

func buildGrokSessionTextPayloadFromAnthropicRequest(req *apicompat.AnthropicRequest) (map[string]any, error) {
	if req == nil {
		return nil, errors.New("anthropic request is nil")
	}

	responsesReq, err := apicompat.AnthropicToResponses(req)
	if err != nil {
		return nil, fmt.Errorf("convert anthropic messages to responses: %w", err)
	}
	return buildGrokSessionTextPayloadFromResponsesRequest(responsesReq)
}

func grokSessionTextRequestFromResponsesRequest(req *apicompat.ResponsesRequest) (grokSessionTextRequest, error) {
	if req == nil {
		return grokSessionTextRequest{}, errors.New("responses request is nil")
	}

	message, systemPrompt, imageInputs, err := extractGrokSessionPromptFromResponsesInput(req.Input)
	if err != nil {
		return grokSessionTextRequest{}, err
	}

	modeID, err := resolveGrokSessionModeID(req.Model)
	if err != nil {
		return grokSessionTextRequest{}, err
	}
	toolPrompt, toolNames := grokSessionToolConfigFromResponsesRequest(req)

	return grokSessionTextRequest{
		ModelID:      strings.TrimSpace(req.Model),
		ModeID:       modeID,
		SystemPrompt: systemPrompt,
		Message:      message,
		ImageInputs:  imageInputs,
		ToolNames:    append([]string(nil), toolNames...),
		ToolPrompt:   toolPrompt,
	}, nil
}

func extractGrokSessionPromptFromResponsesInput(raw json.RawMessage) (message string, systemPrompt string, imageInputs []grokSessionUploadInput, err error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return "", "", nil, errors.New("grok session text request is missing input")
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		plain = strings.TrimSpace(plain)
		if plain == "" {
			return "", "", nil, errors.New("grok session text request is missing input text")
		}
		return plain, "", nil, nil
	}

	var items []apicompat.ResponsesInputItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return "", "", nil, fmt.Errorf("parse responses input: %w", err)
	}
	return grokSessionPromptFromResponsesItems(items)
}

func grokSessionPromptFromResponsesItems(items []apicompat.ResponsesInputItem) (message string, systemPrompt string, imageInputs []grokSessionUploadInput, err error) {
	systemParts := make([]string, 0, 2)
	historyParts := make([]string, 0, len(items))
	uploads := make([]grokSessionUploadInput, 0, 2)
	lastUserOnly := ""
	lastUserOnlyCount := 0

	for _, item := range items {
		role := strings.ToLower(strings.TrimSpace(item.Role))
		switch role {
		case "system", "developer":
			content, err := grokSessionTextContentFromResponsesContent(item.Content)
			if err != nil {
				return "", "", nil, err
			}
			uploads = append(uploads, content.ImageInputs...)
			if content.Text != "" {
				systemParts = append(systemParts, content.Text)
			}
			continue
		}

		switch strings.TrimSpace(item.Type) {
		case "function_call":
			name := strings.TrimSpace(item.Name)
			arguments := strings.TrimSpace(item.Arguments)
			if name != "" {
				xmlBlock := grokToolCallsToXML([]grokParsedToolCall{{
					Name:      name,
					Arguments: arguments,
				}})
				if xmlBlock != "" {
					historyParts = append(historyParts, fmt.Sprintf("Assistant: %s", xmlBlock))
				}
			}
			lastUserOnly = ""
			continue
		case "function_call_output":
			callID := strings.TrimSpace(item.CallID)
			output := strings.TrimSpace(item.Output)
			switch {
			case callID != "" && output != "":
				historyParts = append(historyParts, fmt.Sprintf("Tool result (%s): %s", callID, output))
			case output != "":
				historyParts = append(historyParts, fmt.Sprintf("Tool result: %s", output))
			}
			lastUserOnly = ""
			continue
		}

		if role == "" {
			continue
		}

		content, err := grokSessionTextContentFromResponsesContent(item.Content)
		if err != nil {
			return "", "", nil, err
		}
		uploads = append(uploads, content.ImageInputs...)
		if content.Text == "" {
			continue
		}

		label := grokSessionTranscriptRoleLabel(role)
		historyParts = append(historyParts, fmt.Sprintf("%s: %s", label, content.Text))
		if role == "user" {
			lastUserOnly = content.Text
			lastUserOnlyCount++
		} else {
			lastUserOnly = ""
		}
	}

	if len(historyParts) == 0 {
		return "", "", nil, errors.New("grok session text request does not contain a supported text prompt")
	}

	systemPrompt = strings.Join(systemParts, "\n\n")
	if len(historyParts) == 1 && lastUserOnlyCount == 1 && lastUserOnly != "" {
		return lastUserOnly, systemPrompt, uploads, nil
	}
	return strings.Join(historyParts, "\n\n"), systemPrompt, uploads, nil
}

func grokSessionToolConfigFromResponsesRequest(req *apicompat.ResponsesRequest) (string, []string) {
	if req == nil {
		return "", nil
	}
	return grokSessionToolPromptFromResponsesTools(req.Tools, req.ToolChoice)
}

func grokSessionTranscriptRoleLabel(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "assistant":
		return "Assistant"
	case "system", "developer":
		return "System"
	default:
		return "User"
	}
}

type grokSessionTextContent struct {
	Text        string
	ImageInputs []grokSessionUploadInput
}

func grokSessionTextContentFromResponsesContent(raw json.RawMessage) (grokSessionTextContent, error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return grokSessionTextContent{}, nil
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return grokSessionTextContent{Text: strings.TrimSpace(plain)}, nil
	}

	var parts []apicompat.ResponsesContentPart
	if err := json.Unmarshal(raw, &parts); err != nil {
		return grokSessionTextContent{}, fmt.Errorf("parse responses content: %w", err)
	}

	textParts := make([]string, 0, len(parts))
	imageInputs := make([]grokSessionUploadInput, 0, len(parts))
	for _, part := range parts {
		switch strings.TrimSpace(part.Type) {
		case "input_text", "output_text", "text":
			if text := strings.TrimSpace(part.Text); text != "" {
				textParts = append(textParts, text)
			}
		case "input_image":
			if imageURL := strings.TrimSpace(part.ImageURL); imageURL != "" {
				imageInputs = append(imageInputs, grokSessionUploadInput{Source: imageURL})
			}
		}
	}

	return grokSessionTextContent{
		Text:        strings.Join(textParts, "\n"),
		ImageInputs: imageInputs,
	}, nil
}

func resolveGrokSessionModeID(modelID string) (string, error) {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return "", errors.New("grok session payload requires a model")
	}

	spec, ok := grok.LookupModelSpec(modelID)
	if !ok {
		return "", fmt.Errorf("unknown Grok model: %s", modelID)
	}
	if spec.Capability != grok.CapabilityChat {
		return "", fmt.Errorf("grok session text transport only supports chat models: %s", modelID)
	}

	switch grokQuotaWindowForModel(modelID) {
	case grok.QuotaWindowHeavy:
		return grokSessionModeHeavy, nil
	case grok.QuotaWindowExpert:
		return grokSessionModeExpert, nil
	case grok.QuotaWindowFast:
		return grokSessionModeFast, nil
	case grok.QuotaWindowAuto, "":
		return grokSessionModeAuto, nil
	default:
		return grokSessionModeAuto, nil
	}
}

func grokSessionDefaultDeviceEnvInfo() map[string]any {
	return map[string]any{
		"darkModeEnabled":  false,
		"devicePixelRatio": 2,
		"screenHeight":     1329,
		"screenWidth":      2056,
		"viewportHeight":   1083,
		"viewportWidth":    2056,
	}
}

func grokSessionDefaultToolOverrides() map[string]any {
	return map[string]any{
		"gmailSearch":           false,
		"googleCalendarSearch":  false,
		"outlookSearch":         false,
		"outlookCalendarSearch": false,
		"googleDriveSearch":     false,
	}
}
