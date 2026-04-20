package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/senran-N/sub2api/internal/pkg/grok"
	"github.com/tidwall/gjson"
)

const defaultGrokChatVideoPollInterval = 100 * time.Millisecond

func (s *GrokGatewayService) handleChatCompletionsMedia(c *gin.Context, groupID *int64, body []byte) bool {
	if s == nil || s.mediaService == nil || c == nil || c.Request == nil || !gjson.ValidBytes(body) {
		return false
	}

	var req apicompat.ChatCompletionsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}

	requestedModel := strings.TrimSpace(req.Model)
	if requestedModel == "" {
		return false
	}

	_, mappedModel, restricted := s.mediaService.resolveMappedModel(c.Request.Context(), groupID, requestedModel)
	spec, ok := grok.LookupModelSpec(firstNonEmpty(mappedModel, requestedModel))
	if !ok {
		spec, ok = grok.LookupModelSpec(requestedModel)
	}
	if !ok {
		return false
	}
	if spec.Capability != grok.CapabilityImage && spec.Capability != grok.CapabilityImageEdit && spec.Capability != grok.CapabilityVideo {
		return false
	}
	if restricted {
		writeGatewayCCError(c, http.StatusBadRequest, "invalid_request_error", "Requested model is not allowed for this channel")
		return true
	}

	account, err := s.mediaService.selectCompatibleAccount(c.Request.Context(), groupID, mappedModel)
	if err != nil {
		s.mediaService.writeSelectionError(c, requestedModel, mappedModel, err)
		return true
	}

	switch spec.Capability {
	case grok.CapabilityImage, grok.CapabilityImageEdit:
		requestBody, routePath, err := buildGrokChatImageRequest(&req, spec.Capability)
		if err != nil {
			writeGatewayCCError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
			return true
		}

		if account != nil && account.Type == AccountTypeSession && s.mediaService.sessionRuntime != nil {
			content, reasoning, err := s.mediaService.sessionRuntime.buildChatImageCompletion(c, account, requestedModel, firstNonEmpty(mappedModel, requestedModel), routePath, requestBody)
			if err != nil {
				writeGatewayCCError(c, firstNonZero(grokSessionMediaFeedbackStatusCode(err), http.StatusBadGateway), "api_error", sanitizeUpstreamErrorMessage(firstNonEmpty(strings.TrimSpace(err.Error()), "Grok image request failed")))
				return true
			}
			writeGrokChatMediaCompletion(c, requestedModel, req.Stream, content, reasoning)
			return true
		}
		resp, err := s.runChatMediaRequestWithAccount(c, account, firstNonEmpty(mappedModel, requestedModel), http.MethodPost, routePath, requestBody, "image")
		if err != nil {
			writeGatewayCCError(c, http.StatusBadGateway, "api_error", "Grok image request failed")
			return true
		}
		if !grokChatMediaSuccessStatus(resp.StatusCode) {
			writeGrokChatMediaError(c, resp.StatusCode, resp.Body)
			return true
		}
		content, err := grokChatImageContentFromResponse(resp.Body)
		if err != nil {
			writeGatewayCCError(c, http.StatusBadGateway, "api_error", err.Error())
			return true
		}
		writeGrokChatMediaCompletion(c, requestedModel, req.Stream, content, "")
		return true
	case grok.CapabilityVideo:
		requestBody, err := buildGrokChatVideoRequest(&req)
		if err != nil {
			writeGatewayCCError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
			return true
		}

		if account != nil && account.Type == AccountTypeSession && s.mediaService.sessionRuntime != nil {
			content, reasoning, err := s.mediaService.sessionRuntime.buildChatVideoCompletion(c, account, requestedModel, firstNonEmpty(mappedModel, requestedModel), requestBody)
			if err != nil {
				writeGatewayCCError(c, firstNonZero(grokSessionMediaFeedbackStatusCode(err), http.StatusBadGateway), "api_error", sanitizeUpstreamErrorMessage(firstNonEmpty(strings.TrimSpace(err.Error()), "Grok video request failed")))
				return true
			}
			writeGrokChatMediaCompletion(c, requestedModel, req.Stream, content, reasoning)
			return true
		}
		resp, err := s.runChatVideoCompletion(c, account, firstNonEmpty(mappedModel, requestedModel), requestBody)
		if err != nil {
			writeGatewayCCError(c, http.StatusBadGateway, "api_error", "Grok video request failed")
			return true
		}
		if !grokChatMediaSuccessStatus(resp.StatusCode) {
			writeGrokChatMediaError(c, resp.StatusCode, resp.Body)
			return true
		}
		content, err := grokChatVideoContentFromResponse(resp.Body)
		if err != nil {
			writeGatewayCCError(c, http.StatusBadGateway, "api_error", err.Error())
			return true
		}
		writeGrokChatMediaCompletion(c, requestedModel, req.Stream, content, "")
		return true
	default:
		return false
	}
}

func (s *GrokGatewayService) runChatVideoCompletion(c *gin.Context, account *Account, defaultMappedModel string, body []byte) (*grokMediaForwardResponse, error) {
	createResp, err := s.runChatMediaRequestWithAccount(c, account, defaultMappedModel, http.MethodPost, "/grok/v1/videos", body, "video")
	if err != nil {
		return nil, err
	}
	if !grokChatMediaSuccessStatus(createResp.StatusCode) {
		return createResp, nil
	}

	jobID := strings.TrimSpace(extractGrokVideoJobID(createResp.Body))
	if jobID == "" {
		return nil, errors.New("grok video create response did not include a job id")
	}

	deadline := time.Now().UTC().Add(s.resolveChatVideoTimeout(c))
	currentResp := createResp
	for {
		status := normalizeGrokVideoStatus(extractGrokVideoStatus(currentResp.Body))
		switch status {
		case "completed":
			return s.runChatMediaRequestWithAccount(c, account, defaultMappedModel, http.MethodGet, "/grok/v1/videos/"+jobID+"/content", nil, "video")
		case "failed":
			return currentResp, nil
		}

		if time.Now().UTC().After(deadline) {
			return nil, errors.New("grok video generation timed out")
		}

		time.Sleep(resolveGrokChatVideoPollInterval(currentResp.Body, deadline))
		currentResp, err = s.runChatMediaRequestWithAccount(c, account, defaultMappedModel, http.MethodGet, "/grok/v1/videos/"+jobID, nil, "video")
		if err != nil {
			return nil, err
		}
		if !grokChatMediaSuccessStatus(currentResp.StatusCode) {
			return currentResp, nil
		}
	}
}

func (s *GrokGatewayService) resolveChatVideoTimeout(c *gin.Context) time.Duration {
	timeout := DefaultGrokRuntimeSettings().VideoTimeout()
	if s == nil || s.mediaService == nil || s.mediaService.gatewayService == nil || s.mediaService.gatewayService.settingService == nil || c == nil || c.Request == nil {
		return timeout
	}
	return s.mediaService.gatewayService.settingService.GetGrokRuntimeSettings(c.Request.Context()).VideoTimeout()
}

func resolveGrokChatVideoPollInterval(body []byte, deadline time.Time) time.Duration {
	delay := defaultGrokChatVideoPollInterval
	if pollAfter := extractGrokVideoPollAfter(body); pollAfter != nil {
		wait := time.Until(*pollAfter)
		if wait > 0 {
			delay = wait
		}
	}
	if delay <= 0 {
		delay = defaultGrokChatVideoPollInterval
	}
	remaining := time.Until(deadline)
	if remaining > 0 && delay > remaining {
		return remaining
	}
	return delay
}

func (s *GrokGatewayService) runChatMediaRequestWithAccount(
	parent *gin.Context,
	account *Account,
	defaultMappedModel string,
	method string,
	path string,
	body []byte,
	mediaType string,
) (*grokMediaForwardResponse, error) {
	if s == nil || s.mediaService == nil || parent == nil || parent.Request == nil {
		return nil, errors.New("grok media service is not configured")
	}

	req := parent.Request.Clone(parent.Request.Context())
	if req.URL == nil {
		req.URL = parent.Request.URL
	}
	if req.URL != nil {
		clonedURL := *req.URL
		clonedURL.Path = path
		clonedURL.RawPath = path
		clonedURL.RawQuery = ""
		req.URL = &clonedURL
	}
	req.Method = method
	req.Header = parent.Request.Header.Clone()
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Del("Content-Type")
	}

	child := &gin.Context{Request: req}
	switch mediaType {
	case "video":
		return s.runChatVideoMediaRequest(child, account, defaultMappedModel, body)
	default:
		return s.runChatImageMediaRequest(child, account, defaultMappedModel, body)
	}
}

func (s *GrokGatewayService) runChatImageMediaRequest(child *gin.Context, account *Account, defaultMappedModel string, body []byte) (*grokMediaForwardResponse, error) {
	if s == nil || s.mediaService == nil || child == nil || child.Request == nil {
		return nil, errors.New("grok media service is not configured")
	}
	if account == nil {
		return nil, errors.New("no compatible grok media accounts")
	}

	requestedFormat, err := resolveGrokImageResponseFormatRequest(child, body)
	if err != nil {
		return buildChatMediaErrorResponse(http.StatusBadRequest, "invalid_request_error", err.Error()), nil
	}
	reqMeta := GetOpenAICompatiblePassthroughRequestMeta(child, body)
	requestedModel := strings.TrimSpace(reqMeta.Model)
	resp, err := s.mediaService.forwardCompatibleRequest(child, grokMediaForwardRequest{
		account:            account,
		body:               body,
		requestMeta:        reqMeta,
		defaultMappedModel: firstNonEmpty(defaultMappedModel, requestedModel),
		applyVideoTimeout:  false,
	})
	if err != nil {
		return nil, err
	}
	if grokChatMediaSuccessStatus(resp.StatusCode) {
		resp.Body, _, _ = s.mediaService.rewriteMediaResponse(child, account, requestedModel, firstNonEmpty(resp.UpstreamModel, requestedModel), "", "image", requestedFormat, resp.Body)
	}
	return resp, nil
}

func (s *GrokGatewayService) runChatVideoMediaRequest(child *gin.Context, account *Account, defaultMappedModel string, body []byte) (*grokMediaForwardResponse, error) {
	if s == nil || s.mediaService == nil || child == nil || child.Request == nil {
		return nil, errors.New("grok media service is not configured")
	}
	if account == nil {
		return nil, errors.New("no compatible grok media accounts")
	}

	reqMeta := GetOpenAICompatiblePassthroughRequestMeta(child, body)
	requestedModel := strings.TrimSpace(reqMeta.Model)
	resp, err := s.mediaService.forwardCompatibleRequest(child, grokMediaForwardRequest{
		account:            account,
		body:               body,
		requestMeta:        reqMeta,
		defaultMappedModel: firstNonEmpty(defaultMappedModel, requestedModel),
		applyVideoTimeout:  true,
	})
	if err != nil {
		return nil, err
	}
	if grokChatMediaSuccessStatus(resp.StatusCode) {
		resp.Body, _, _ = s.mediaService.rewriteMediaResponse(child, account, requestedModel, firstNonEmpty(resp.UpstreamModel, requestedModel), "", "video", "", resp.Body)
	}
	return resp, nil
}

func buildChatMediaErrorResponse(statusCode int, code string, message string) *grokMediaForwardResponse {
	body, _ := json.Marshal(gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
	return &grokMediaForwardResponse{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       body,
	}
}

func buildGrokChatImageRequest(req *apicompat.ChatCompletionsRequest, capability grok.Capability) ([]byte, string, error) {
	if req == nil {
		return nil, "", errors.New("chat completions request is nil")
	}

	imageConfig := apicompat.ChatImageConfig{}
	if req.ImageConfig != nil {
		imageConfig = *req.ImageConfig
	}

	switch capability {
	case grok.CapabilityImageEdit:
		prompt, inputs, err := extractGrokChatImageEditPromptAndInputs(req.Messages)
		if err != nil {
			return nil, "", err
		}
		payload := map[string]any{
			"model":  req.Model,
			"prompt": prompt,
			"n":      maxInt(imageConfig.N, 1),
			"size":   firstNonEmpty(imageConfig.Size, "1024x1024"),
		}
		if format := strings.TrimSpace(imageConfig.ResponseFormat); format != "" {
			payload["response_format"] = format
		}
		if len(inputs) == 1 {
			payload["image"] = inputs[0]
		} else {
			values := make([]string, 0, len(inputs))
			for _, input := range inputs {
				if trimmed := strings.TrimSpace(input); trimmed != "" {
					values = append(values, trimmed)
				}
			}
			payload["image"] = values
		}
		body, err := json.Marshal(payload)
		return body, "/grok/v1/images/edits", err
	default:
		prompt, err := extractGrokChatLatestUserText(req.Messages)
		if err != nil {
			return nil, "", err
		}
		payload := map[string]any{
			"model":  req.Model,
			"prompt": prompt,
			"n":      maxInt(imageConfig.N, 1),
			"size":   firstNonEmpty(imageConfig.Size, "1024x1024"),
		}
		if format := strings.TrimSpace(imageConfig.ResponseFormat); format != "" {
			payload["response_format"] = format
		}
		body, err := json.Marshal(payload)
		return body, "/grok/v1/images/generations", err
	}
}

func buildGrokChatVideoRequest(req *apicompat.ChatCompletionsRequest) ([]byte, error) {
	if req == nil {
		return nil, errors.New("chat completions request is nil")
	}

	prompt, referenceURL, err := extractGrokChatVideoPromptAndReference(req.Messages)
	if err != nil {
		return nil, err
	}

	videoConfig := apicompat.ChatVideoConfig{}
	if req.VideoConfig != nil {
		videoConfig = *req.VideoConfig
	}

	payload := map[string]any{
		"model":  req.Model,
		"prompt": prompt,
		"seconds": func() int {
			if videoConfig.Seconds > 0 {
				return videoConfig.Seconds
			}
			return 6
		}(),
		"size": firstNonEmpty(videoConfig.Size, "720x1280"),
	}
	if resolution := strings.TrimSpace(videoConfig.ResolutionName); resolution != "" {
		payload["resolution_name"] = resolution
	}
	if preset := strings.TrimSpace(videoConfig.Preset); preset != "" {
		payload["preset"] = preset
	}
	if referenceURL != "" {
		payload["input_reference"] = map[string]any{
			"image_url": referenceURL,
		}
	}
	return json.Marshal(payload)
}

func extractGrokChatLatestUserText(messages []apicompat.ChatMessage) (string, error) {
	for i := len(messages) - 1; i >= 0; i-- {
		if !strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			continue
		}
		text := strings.TrimSpace(extractGrokChatTextContent(messages[i].Content))
		if text != "" {
			return text, nil
		}
	}
	return "", errors.New("chat completions media request requires a non-empty user prompt")
}

func extractGrokChatImageEditPromptAndInputs(messages []apicompat.ChatMessage) (string, []string, error) {
	prompt := ""
	inputs := make([]string, 0, 4)

	for _, message := range messages {
		content := strings.TrimSpace(string(message.Content))
		if content == "" || content == "null" {
			continue
		}

		var plain string
		if err := json.Unmarshal(message.Content, &plain); err == nil {
			if trimmed := strings.TrimSpace(plain); trimmed != "" {
				prompt = trimmed
			}
			continue
		}

		var parts []apicompat.ChatContentPart
		if err := json.Unmarshal(message.Content, &parts); err != nil {
			continue
		}
		for _, part := range parts {
			switch strings.ToLower(strings.TrimSpace(part.Type)) {
			case "text":
				if trimmed := strings.TrimSpace(part.Text); trimmed != "" {
					prompt = trimmed
				}
			case "image_url":
				if part.ImageURL != nil {
					if trimmed := strings.TrimSpace(part.ImageURL.URL); trimmed != "" {
						inputs = append(inputs, trimmed)
					}
				}
			}
		}
	}

	if strings.TrimSpace(prompt) == "" {
		return "", nil, errors.New("chat completions image edit requires a non-empty text prompt")
	}
	if len(inputs) == 0 {
		return "", nil, errors.New("chat completions image edit requires at least one image_url content block")
	}
	if len(inputs) > 5 {
		inputs = inputs[len(inputs)-5:]
	}
	return prompt, inputs, nil
}

func extractGrokChatVideoPromptAndReference(messages []apicompat.ChatMessage) (string, string, error) {
	for i := len(messages) - 1; i >= 0; i-- {
		if !strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			continue
		}

		content := strings.TrimSpace(string(messages[i].Content))
		if content == "" || content == "null" {
			continue
		}

		var plain string
		if err := json.Unmarshal(messages[i].Content, &plain); err == nil {
			if trimmed := strings.TrimSpace(plain); trimmed != "" {
				return trimmed, "", nil
			}
			continue
		}

		var parts []apicompat.ChatContentPart
		if err := json.Unmarshal(messages[i].Content, &parts); err != nil {
			continue
		}

		textParts := make([]string, 0, len(parts))
		referenceURL := ""
		for _, part := range parts {
			switch strings.ToLower(strings.TrimSpace(part.Type)) {
			case "text":
				if trimmed := strings.TrimSpace(part.Text); trimmed != "" {
					textParts = append(textParts, trimmed)
				}
			case "image_url":
				if referenceURL == "" && part.ImageURL != nil {
					referenceURL = strings.TrimSpace(part.ImageURL.URL)
				}
			}
		}
		if len(textParts) > 0 {
			return strings.Join(textParts, " "), referenceURL, nil
		}
	}
	return "", "", errors.New("chat completions video request requires a non-empty user prompt")
}

func extractGrokChatTextContent(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return strings.TrimSpace(plain)
	}

	var parts []apicompat.ChatContentPart
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ""
	}

	textParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.EqualFold(strings.TrimSpace(part.Type), "text") {
			if text := strings.TrimSpace(part.Text); text != "" {
				textParts = append(textParts, text)
			}
		}
	}
	return strings.Join(textParts, " ")
}

func grokChatImageContentFromResponse(body []byte) (string, error) {
	if !gjson.ValidBytes(body) {
		return "", errors.New("grok image response was not valid JSON")
	}

	parts := make([]string, 0, 2)
	for _, item := range gjson.GetBytes(body, "data").Array() {
		value := firstNonEmpty(
			strings.TrimSpace(item.Get("url").String()),
			strings.TrimSpace(item.Get("b64_json").String()),
		)
		if value == "" {
			continue
		}
		parts = append(parts, normalizeGrokChatImageContentValue(value))
	}
	if len(parts) == 0 {
		return "", errors.New("grok image response did not include any generated images")
	}
	return strings.Join(parts, "\n\n"), nil
}

func normalizeGrokChatImageContentValue(value string) string {
	trimmed := strings.TrimSpace(value)
	switch {
	case trimmed == "":
		return ""
	case strings.HasPrefix(trimmed, "!["):
		return trimmed
	case strings.HasPrefix(trimmed, "http://"), strings.HasPrefix(trimmed, "https://"), strings.HasPrefix(trimmed, "data:"), strings.HasPrefix(trimmed, "/"):
		return fmt.Sprintf("![image](%s)", trimmed)
	default:
		return trimmed
	}
}

func grokChatVideoContentFromResponse(body []byte) (string, error) {
	value := firstNonEmpty(
		strings.TrimSpace(gjson.GetBytes(body, "content_url").String()),
		strings.TrimSpace(gjson.GetBytes(body, "url").String()),
		strings.TrimSpace(gjson.GetBytes(body, "video_url").String()),
		strings.TrimSpace(gjson.GetBytes(body, "data.content_url").String()),
	)
	if value == "" {
		return "", errors.New("grok video response did not include any playable content")
	}
	return value, nil
}

func writeGrokChatMediaCompletion(c *gin.Context, model string, stream bool, content string, reasoning string) {
	if c == nil {
		return
	}
	if stream {
		writeGrokChatMediaStream(c, model, content, reasoning)
		return
	}

	encodedContent, _ := json.Marshal(content)
	message := apicompat.ChatMessage{
		Role:    "assistant",
		Content: encodedContent,
	}
	if trimmed := strings.TrimSpace(reasoning); trimmed != "" {
		message.ReasoningContent = trimmed
	}
	c.JSON(http.StatusOK, apicompat.ChatCompletionsResponse{
		ID:      fmt.Sprintf("chatcmpl_%d", time.Now().UTC().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().UTC().Unix(),
		Model:   strings.TrimSpace(model),
		Choices: []apicompat.ChatChoice{{
			Index:        0,
			Message:      message,
			FinishReason: "stop",
		}},
	})
}

func writeGrokChatMediaStream(c *gin.Context, model string, content string, reasoning string) {
	responseID := fmt.Sprintf("chatcmpl_%d", time.Now().UTC().UnixNano())
	createdAt := time.Now().UTC().Unix()
	flusher, _ := c.Writer.(http.Flusher)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	writeChunk := func(chunk apicompat.ChatCompletionsChunk) {
		frame, err := apicompat.ChatChunkToSSE(chunk)
		if err != nil {
			return
		}
		_, _ = c.Writer.WriteString(frame)
		if flusher != nil {
			flusher.Flush()
		}
	}

	writeChunk(apicompat.ChatCompletionsChunk{
		ID:      responseID,
		Object:  "chat.completion.chunk",
		Created: createdAt,
		Model:   strings.TrimSpace(model),
		Choices: []apicompat.ChatChunkChoice{{
			Index: 0,
			Delta: apicompat.ChatDelta{Role: "assistant"},
		}},
	})

	if trimmed := strings.TrimSpace(reasoning); trimmed != "" {
		for _, part := range strings.Split(trimmed, "\n") {
			reasoningPart := strings.TrimSpace(part)
			if reasoningPart == "" {
				continue
			}
			reasoningCopy := reasoningPart
			writeChunk(apicompat.ChatCompletionsChunk{
				ID:      responseID,
				Object:  "chat.completion.chunk",
				Created: createdAt,
				Model:   strings.TrimSpace(model),
				Choices: []apicompat.ChatChunkChoice{{
					Index: 0,
					Delta: apicompat.ChatDelta{ReasoningContent: &reasoningCopy},
				}},
			})
		}
	}

	contentCopy := content
	writeChunk(apicompat.ChatCompletionsChunk{
		ID:      responseID,
		Object:  "chat.completion.chunk",
		Created: createdAt,
		Model:   strings.TrimSpace(model),
		Choices: []apicompat.ChatChunkChoice{{
			Index: 0,
			Delta: apicompat.ChatDelta{Content: &contentCopy},
		}},
	})

	finishReason := "stop"
	writeChunk(apicompat.ChatCompletionsChunk{
		ID:      responseID,
		Object:  "chat.completion.chunk",
		Created: createdAt,
		Model:   strings.TrimSpace(model),
		Choices: []apicompat.ChatChunkChoice{{
			Index:        0,
			Delta:        apicompat.ChatDelta{},
			FinishReason: &finishReason,
		}},
	})

	_, _ = c.Writer.WriteString("data: [DONE]\n\n")
	if flusher != nil {
		flusher.Flush()
	}
}

func writeGrokChatMediaError(c *gin.Context, statusCode int, body []byte) {
	message := firstNonEmpty(
		strings.TrimSpace(gjson.GetBytes(body, "error.message").String()),
		http.StatusText(statusCode),
	)
	errType := firstNonEmpty(
		strings.TrimSpace(gjson.GetBytes(body, "error.type").String()),
		strings.TrimSpace(gjson.GetBytes(body, "error.code").String()),
		"api_error",
	)
	writeGatewayCCError(c, statusCode, errType, message)
}

func grokChatMediaSuccessStatus(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
