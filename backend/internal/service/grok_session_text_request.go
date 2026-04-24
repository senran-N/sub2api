//nolint:unused
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/senran-N/sub2api/internal/pkg/grok"
)

const (
	grokSessionModeAuto   = "auto"
	grokSessionModeFast   = "fast"
	grokSessionModeExpert = "expert"
	grokSessionModeHeavy  = "heavy"

	grokSessionDeepsearchDefault = "default"
	grokSessionDeepsearchDeeper  = "deeper"
)

var (
	grokSessionThinkTagRe        = regexp.MustCompile(`(?is)<(?:think|thinking)>[\s\S]*?</(?:think|thinking)>`)
	grokSessionInlineBase64ImgRe = regexp.MustCompile(`!\[image\]\(data:[^)]*?base64,[^)]*?\)`)
	grokSessionSourcesStripRe    = regexp.MustCompile(`(?:^|\r?\n\r?\n)## Sources\r?\n\[grok2api-sources\]: #\r?\n[\s\S]*$`)
	grokSessionToolCallsStripRe  = regexp.MustCompile(`(?is)<tool_calls\s*>[\s\S]*?</tool_calls\s*>`)
	grokSessionFunctionStripRe   = regexp.MustCompile(`(?is)<function_call\s*>[\s\S]*?</function_call\s*>`)
	grokSessionInvokeStripRe     = regexp.MustCompile(`(?is)<invoke\b[\s\S]*?</invoke\s*>`)
)

type grokSessionTextRequest struct {
	ModelID          string
	ModeID           string
	DeepsearchPreset string
	SystemPrompt     string
	Message          string
	FileAttachments  []string
	AttachmentInputs []grokSessionUploadInput
	ToolNames        []string
	ToolPrompt       string
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
	if input.DeepsearchPreset != "" {
		payload["deepsearchPreset"] = input.DeepsearchPreset
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

	message, systemPrompt, attachmentInputs, err := extractGrokSessionPromptFromResponsesInput(req.Input)
	if err != nil {
		return grokSessionTextRequest{}, err
	}

	modeID, err := resolveGrokSessionModeID(req.Model)
	if err != nil {
		return grokSessionTextRequest{}, err
	}
	toolPrompt, toolNames := grokSessionToolConfigFromResponsesRequest(req)
	deepsearchPreset, err := normalizeGrokSessionDeepsearchPreset(req.Deepsearch)
	if err != nil {
		return grokSessionTextRequest{}, err
	}

	return grokSessionTextRequest{
		ModelID:          strings.TrimSpace(req.Model),
		ModeID:           modeID,
		DeepsearchPreset: deepsearchPreset,
		SystemPrompt:     systemPrompt,
		Message:          message,
		AttachmentInputs: append([]grokSessionUploadInput(nil), attachmentInputs...),
		ToolNames:        append([]string(nil), toolNames...),
		ToolPrompt:       toolPrompt,
	}, nil
}

func extractGrokSessionPromptFromResponsesInput(raw json.RawMessage) (message string, systemPrompt string, attachmentInputs []grokSessionUploadInput, err error) {
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

func grokSessionPromptFromResponsesItems(items []apicompat.ResponsesInputItem) (message string, systemPrompt string, attachmentInputs []grokSessionUploadInput, err error) {
	systemParts := make([]string, 0, 2)
	historyParts := make([]string, 0, len(items))
	uploads := make([]grokSessionUploadInput, 0, 2)
	lastUserOnly := ""
	lastUserOnlyCount := 0

	for _, item := range items {
		role := strings.ToLower(strings.TrimSpace(item.Role))
		switch role {
		case "system", "developer":
			content, err := grokSessionTextContentFromResponsesContent(item.Content, role)
			if err != nil {
				return "", "", nil, err
			}
			uploads = append(uploads, content.AttachmentInputs...)
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
					historyParts = append(historyParts, grokSessionFormatTranscriptBlock("assistant", xmlBlock))
				}
			}
			lastUserOnly = ""
			continue
		case "function_call_output":
			callID := strings.TrimSpace(item.CallID)
			output := strings.TrimSpace(item.Output)
			switch {
			case callID != "" && output != "":
				historyParts = append(historyParts, grokSessionFormatToolResultBlock(callID, output))
			case output != "":
				historyParts = append(historyParts, grokSessionFormatToolResultBlock("", output))
			}
			lastUserOnly = ""
			continue
		}

		if role == "" {
			continue
		}

		content, err := grokSessionTextContentFromResponsesContent(item.Content, role)
		if err != nil {
			return "", "", nil, err
		}
		uploads = append(uploads, content.AttachmentInputs...)
		if content.Text == "" {
			continue
		}

		historyParts = append(historyParts, grokSessionFormatTranscriptBlock(role, content.Text))
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

func normalizeGrokSessionDeepsearchPreset(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return "", nil
	case grokSessionDeepsearchDefault:
		return grokSessionDeepsearchDefault, nil
	case grokSessionDeepsearchDeeper:
		return grokSessionDeepsearchDeeper, nil
	default:
		return "", fmt.Errorf("unsupported deepsearch value %q", strings.TrimSpace(raw))
	}
}

func grokSessionFormatTranscriptBlock(role string, text string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	switch {
	case strings.Contains(text, "\n"):
		return fmt.Sprintf("[%s]:\n%s", firstNonEmpty(role, "user"), text)
	default:
		return fmt.Sprintf("[%s]: %s", firstNonEmpty(role, "user"), text)
	}
}

func grokSessionFormatToolResultBlock(callID string, output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}
	label := "[tool result]"
	if callID = strings.TrimSpace(callID); callID != "" {
		label = fmt.Sprintf("[tool result for %s]", callID)
	}
	return fmt.Sprintf("%s:\n%s", label, output)
}

type grokSessionTextContent struct {
	Text             string
	AttachmentInputs []grokSessionUploadInput
}

func grokSessionTextContentFromResponsesContent(raw json.RawMessage, role string) (grokSessionTextContent, error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return grokSessionTextContent{}, nil
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return grokSessionTextContent{Text: grokSessionSanitizeTranscriptText(strings.TrimSpace(plain), role)}, nil
	}

	var parts []apicompat.ResponsesContentPart
	if err := json.Unmarshal(raw, &parts); err != nil {
		return grokSessionTextContent{}, fmt.Errorf("parse responses content: %w", err)
	}

	textParts := make([]string, 0, len(parts))
	attachmentInputs := make([]grokSessionUploadInput, 0, len(parts))
	for _, part := range parts {
		switch strings.TrimSpace(part.Type) {
		case "input_text", "output_text", "text":
			if text := grokSessionSanitizeTranscriptText(strings.TrimSpace(part.Text), role); text != "" {
				textParts = append(textParts, text)
			}
		case "input_image":
			if imageURL := strings.TrimSpace(part.ImageURL); imageURL != "" {
				attachmentInputs = append(attachmentInputs, grokSessionUploadInput{Source: imageURL})
			}
		case "input_audio":
			if part.InputAudio == nil {
				continue
			}
			if attachment, ok := grokSessionUploadInputFromInlineContent(
				part.InputAudio.Data,
				part.InputAudio.Filename,
				firstNonEmpty(strings.TrimSpace(part.InputAudio.MIMEType), grokSessionAudioMimeType(part.InputAudio.Format)),
			); ok {
				attachmentInputs = append(attachmentInputs, attachment)
			}
		case "file":
			if part.File == nil {
				continue
			}
			if attachment, ok := grokSessionUploadInputFromInlineContent(
				part.File.FileData,
				part.File.Filename,
				strings.TrimSpace(part.File.MIMEType),
			); ok {
				attachmentInputs = append(attachmentInputs, attachment)
			}
		}
	}

	return grokSessionTextContent{
		Text:             strings.Join(textParts, "\n"),
		AttachmentInputs: attachmentInputs,
	}, nil
}

func grokSessionSanitizeTranscriptText(text string, role string) string {
	if text == "" {
		return ""
	}
	if strings.EqualFold(strings.TrimSpace(role), "assistant") {
		text = grokSessionSourcesStripRe.ReplaceAllString(text, "")
		text = grokSessionToolCallsStripRe.ReplaceAllString(text, "")
		text = grokSessionFunctionStripRe.ReplaceAllString(text, "")
		text = grokSessionInvokeStripRe.ReplaceAllString(text, "")
	}
	text = grokSessionThinkTagRe.ReplaceAllString(text, "")
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	return grokSessionInlineBase64ImgRe.ReplaceAllString(text, "[图片]")
}

func grokSessionUploadInputFromInlineContent(raw string, fileName string, mimeType string) (grokSessionUploadInput, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return grokSessionUploadInput{}, false
	}

	input := grokSessionUploadInput{
		FileName: strings.TrimSpace(fileName),
		MimeType: strings.TrimSpace(mimeType),
	}
	switch {
	case strings.HasPrefix(raw, "data:"):
		input.Source = raw
	case grokSessionLooksLikeAbsoluteURL(raw):
		input.Source = raw
	default:
		input.Base64 = raw
		if input.FileName == "" && input.MimeType != "" {
			input.FileName = grokSessionDefaultFileName(input.MimeType)
		}
	}
	return input, true
}

func grokSessionLooksLikeAbsoluteURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func grokSessionAudioMimeType(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "flac":
		return "audio/flac"
	case "m4a":
		return "audio/mp4"
	case "mp3", "mpeg":
		return "audio/mpeg"
	case "ogg":
		return "audio/ogg"
	case "wav", "wave":
		return "audio/wav"
	case "webm":
		return "audio/webm"
	default:
		return ""
	}
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
		return grokSessionModeIDForQuotaWindow(grok.QuotaWindowHeavy), nil
	case grok.QuotaWindowExpert:
		return grokSessionModeIDForQuotaWindow(grok.QuotaWindowExpert), nil
	case grok.QuotaWindowFast:
		return grokSessionModeIDForQuotaWindow(grok.QuotaWindowFast), nil
	case grok.QuotaWindowAuto, "":
		return grokSessionModeIDForQuotaWindow(grok.QuotaWindowAuto), nil
	default:
		return grokSessionModeIDForQuotaWindow(grok.QuotaWindowAuto), nil
	}
}

func grokSessionModeIDForQuotaWindow(quotaWindow string) string {
	switch strings.TrimSpace(quotaWindow) {
	case grok.QuotaWindowHeavy:
		return grokSessionModeHeavy
	case grok.QuotaWindowExpert:
		return grokSessionModeExpert
	case grok.QuotaWindowFast:
		return grokSessionModeFast
	case grok.QuotaWindowAuto, "":
		return grokSessionModeAuto
	default:
		return grokSessionModeAuto
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
