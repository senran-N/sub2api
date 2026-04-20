package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/apicompat"
	"github.com/tidwall/gjson"
)

type grokSessionResponseDelta struct {
	token              string
	reasoning          bool
	completion         bool
	errorText          string
	cardAttachmentJSON string
	webSearchResults   string
	xSearchResults     string
}

type grokSessionPendingCitation struct {
	URL    string
	Title  string
	Needle string
}

type grokSessionLocalAnnotation struct {
	URL        string
	Title      string
	StartIndex int
	EndIndex   int
}

type grokSessionResponsesState struct {
	responseID           string
	model                string
	sequenceNumber       int
	reasoningOpened      bool
	reasoningOutputIndex int
	reasoningItemID      string
	messageOpened        bool
	messageOutputIndex   int
	messageItemID        string
	accumulator          *apicompat.BufferedResponseAccumulator
	cardAttachments      map[string]string
	citationIndexByURL   map[string]int
	nextCitationIndex    int
	lastCitationIndex    int
	pendingCitations     []grokSessionPendingCitation
	messageTextOffset    int
	messageAnnotations   []apicompat.ResponsesTextAnnotation
	searchSources        []apicompat.SearchSource
	searchSourceSeen     map[string]struct{}
	searchSourceTitle    map[string]string
}

var errGrokSessionTextStreamHandled = errors.New("grok session text stream already handled")
var grokSessionRenderRe = regexp.MustCompile(`<grok:render\s+card_id="([^"]+)"\s+card_type="([^"]+)"\s+type="([^"]+)"[^>]*>.*?</grok:render>`)

type grokSessionResponsesCallbacks struct {
	onEvent         func(apicompat.ResponsesStreamEvent) error
	onStreamFailure func(string, apicompat.ResponsesStreamEvent) error
	onSearchSources func([]apicompat.SearchSource)
}

type grokSessionResponseMetadata struct {
	SearchSources []apicompat.SearchSource
}

func relayGrokSessionResponses(c *gin.Context, upstream io.Reader, model string, stream bool, toolNames []string) error {
	if c == nil {
		return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "request context is nil")
	}
	if !stream {
		finalResponse, _, err := collectGrokSessionResponses(upstream, model, toolNames, nil)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, finalResponse)
		return nil
	}

	streamWriter, streamFlusher, err := openGrokSessionSSEWriter(c)
	if err != nil {
		return err
	}

	_, _, err = collectGrokSessionResponses(upstream, model, toolNames, &grokSessionResponsesCallbacks{
		onEvent: func(event apicompat.ResponsesStreamEvent) error {
			return writeGrokSessionResponsesEvent(streamWriter, streamFlusher, event)
		},
		onStreamFailure: func(message string, failedEvent apicompat.ResponsesStreamEvent) error {
			return writeGrokSessionResponsesEvent(streamWriter, streamFlusher, failedEvent)
		},
	})
	if errors.Is(err, errGrokSessionTextStreamHandled) {
		return nil
	}
	return err
}

func relayGrokSessionChatCompletions(c *gin.Context, upstream io.Reader, model string, stream bool, includeUsage bool, toolNames []string) error {
	if c == nil {
		return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "request context is nil")
	}
	if !stream {
		finalResponse, metadata, err := collectGrokSessionResponses(upstream, model, toolNames, nil)
		if err != nil {
			return err
		}
		chatResponse := apicompat.ResponsesToChatCompletions(finalResponse, strings.TrimSpace(model))
		if len(chatResponse.SearchSources) == 0 && len(metadata.SearchSources) > 0 {
			chatResponse.SearchSources = append([]apicompat.SearchSource(nil), metadata.SearchSources...)
		}
		c.JSON(http.StatusOK, chatResponse)
		return nil
	}

	streamWriter, streamFlusher, err := openGrokSessionSSEWriter(c)
	if err != nil {
		return err
	}

	state := apicompat.NewResponsesEventToChatState()
	state.Model = strings.TrimSpace(model)
	state.IncludeUsage = includeUsage

	_, _, err = collectGrokSessionResponses(upstream, model, toolNames, &grokSessionResponsesCallbacks{
		onEvent: func(event apicompat.ResponsesStreamEvent) error {
			return writeGrokSessionChatChunks(streamWriter, streamFlusher, state, event)
		},
		onStreamFailure: func(message string, _ apicompat.ResponsesStreamEvent) error {
			payload, marshalErr := json.Marshal(passthroughRuleResult{
				StatusCode: http.StatusBadGateway,
				ErrType:    "api_error",
				ErrMessage: message,
			}.openAIPayload())
			if marshalErr != nil {
				return marshalErr
			}
			if err := writeGrokSessionStreamPayload(streamWriter, streamFlusher, "data: "+string(payload)+"\n\n"); err != nil {
				return err
			}
			return writeGrokSessionStreamPayload(streamWriter, streamFlusher, "data: [DONE]\n\n")
		},
		onSearchSources: func(searchSources []apicompat.SearchSource) {
			state.PendingSearchSources = append([]apicompat.SearchSource(nil), searchSources...)
		},
	})
	if errors.Is(err, errGrokSessionTextStreamHandled) {
		return nil
	}
	if err != nil {
		return err
	}

	if finalChunks := apicompat.FinalizeResponsesChatStream(state); len(finalChunks) > 0 {
		if err := writeGrokSessionChatChunkBatch(streamWriter, streamFlusher, finalChunks); err != nil {
			return err
		}
	}
	return writeGrokSessionStreamPayload(streamWriter, streamFlusher, "data: [DONE]\n\n")
}

func relayGrokSessionAnthropic(c *gin.Context, upstream io.Reader, model string, stream bool, toolNames []string) error {
	if c == nil {
		return newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "request context is nil")
	}
	if !stream {
		finalResponse, metadata, err := collectGrokSessionResponses(upstream, model, toolNames, nil)
		if err != nil {
			return err
		}
		anthropicResponse := apicompat.ResponsesToAnthropic(finalResponse, strings.TrimSpace(model))
		if len(anthropicResponse.SearchSources) == 0 && len(metadata.SearchSources) > 0 {
			anthropicResponse.SearchSources = append([]apicompat.SearchSource(nil), metadata.SearchSources...)
		}
		c.JSON(http.StatusOK, anthropicResponse)
		return nil
	}

	streamWriter, streamFlusher, err := openGrokSessionSSEWriter(c)
	if err != nil {
		return err
	}

	state := apicompat.NewResponsesEventToAnthropicState()
	state.Model = strings.TrimSpace(model)

	_, _, err = collectGrokSessionResponses(upstream, model, toolNames, &grokSessionResponsesCallbacks{
		onEvent: func(event apicompat.ResponsesStreamEvent) error {
			return writeGrokSessionAnthropicEvents(streamWriter, streamFlusher, state, event)
		},
		onStreamFailure: func(message string, _ apicompat.ResponsesStreamEvent) error {
			payload, marshalErr := json.Marshal(passthroughRuleResult{
				StatusCode: http.StatusBadGateway,
				ErrType:    "api_error",
				ErrMessage: message,
			}.anthropicPayload())
			if marshalErr != nil {
				return marshalErr
			}
			return writeGrokSessionStreamPayload(streamWriter, streamFlusher, "event: error\ndata: "+string(payload)+"\n\n")
		},
		onSearchSources: func(searchSources []apicompat.SearchSource) {
			state.PendingSearchSources = append([]apicompat.SearchSource(nil), searchSources...)
		},
	})
	if errors.Is(err, errGrokSessionTextStreamHandled) {
		return nil
	}
	if err != nil {
		return err
	}

	if finalEvents := apicompat.FinalizeResponsesAnthropicStream(state); len(finalEvents) > 0 {
		if err := writeGrokSessionAnthropicEventBatch(streamWriter, streamFlusher, finalEvents); err != nil {
			return err
		}
	}
	return nil
}

func collectGrokSessionResponses(
	upstream io.Reader,
	model string,
	toolNames []string,
	callbacks *grokSessionResponsesCallbacks,
) (*apicompat.ResponsesResponse, grokSessionResponseMetadata, error) {
	if upstream == nil {
		return nil, grokSessionResponseMetadata{}, newGrokResponsesHTTPError(http.StatusBadGateway, "api_error", "Upstream stream ended without a response")
	}

	state := grokSessionResponsesState{
		responseID:         newGrokSessionResponseID("resp"),
		model:              strings.TrimSpace(model),
		accumulator:        apicompat.NewBufferedResponseAccumulator(),
		cardAttachments:    make(map[string]string),
		citationIndexByURL: make(map[string]int),
		lastCitationIndex:  -1,
		searchSourceSeen:   make(map[string]struct{}),
		searchSourceTitle:  make(map[string]string),
	}
	if state.model == "" {
		state.model = "grok"
	}

	scanner := bufio.NewScanner(upstream)
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], defaultMaxLineSize)
	defer putSSEScannerBuf64K(scanBuf)

	var (
		sawDelta       bool
		sawCompletion  bool
		createdEmitted bool
		reasoningDone  bool
		messageDone    bool
		toolSieve      *grokToolSieve
		toolCallsSeen  bool
	)
	if len(toolNames) > 0 {
		toolSieve = newGrokToolSieve(toolNames)
	}

	emitCreated := func() error {
		if createdEmitted || callbacks == nil || callbacks.onEvent == nil {
			return nil
		}
		createdEmitted = true
		return callbacks.onEvent(state.nextCreatedEvent())
	}
	emitEvent := func(event apicompat.ResponsesStreamEvent) error {
		if callbacks == nil || callbacks.onEvent == nil {
			return nil
		}
		return callbacks.onEvent(event)
	}
	emitSearchSources := func() {
		if callbacks == nil || callbacks.onSearchSources == nil || len(state.searchSources) == 0 {
			return
		}
		callbacks.onSearchSources(append([]apicompat.SearchSource(nil), state.searchSources...))
	}
	emitReasoningDone := func() error {
		if reasoningDone || !state.reasoningOpened {
			return nil
		}
		if err := emitEvent(state.nextReasoningDoneEvent()); err != nil {
			return err
		}
		if err := emitEvent(state.nextReasoningSummaryPartDoneEvent()); err != nil {
			return err
		}
		if err := emitEvent(state.nextReasoningItemDoneEvent()); err != nil {
			return err
		}
		reasoningDone = true
		return nil
	}
	emitMessageDone := func() error {
		if messageDone || !state.messageOpened {
			return nil
		}
		if err := emitEvent(state.nextOutputTextDoneEvent()); err != nil {
			return err
		}
		if err := emitEvent(state.nextContentPartDoneEvent()); err != nil {
			return err
		}
		if err := emitEvent(state.nextMessageItemDoneEvent()); err != nil {
			return err
		}
		messageDone = true
		return nil
	}
	emitToolCalls := func(calls []grokParsedToolCall) error {
		if len(calls) == 0 {
			return nil
		}
		if err := emitCreated(); err != nil {
			return err
		}
		if err := emitReasoningDone(); err != nil {
			return err
		}
		if err := emitMessageDone(); err != nil {
			return err
		}
		baseIndex := state.nextFunctionCallOutputIndex()
		for idx, call := range calls {
			added := state.nextFunctionCallAddedEvent(baseIndex+idx, call)
			state.accumulator.ProcessEvent(&added)
			if err := emitEvent(added); err != nil {
				return err
			}
			if call.Arguments != "" {
				deltaEvent := state.nextFunctionCallArgumentsDeltaEvent(baseIndex+idx, added.Item.ID, call.Arguments)
				state.accumulator.ProcessEvent(&deltaEvent)
				if err := emitEvent(deltaEvent); err != nil {
					return err
				}
			}
			done := state.nextFunctionCallArgumentsDoneEvent(baseIndex+idx, added.Item.ID, call.Arguments)
			if err := emitEvent(done); err != nil {
				return err
			}
			if err := emitEvent(state.nextFunctionCallDoneEvent(baseIndex+idx, call, added.Item.ID)); err != nil {
				return err
			}
		}
		toolCallsSeen = true
		return nil
	}
	emitReasoning := func(token string) error {
		if token == "" {
			return nil
		}
		if err := emitCreated(); err != nil {
			return err
		}
		if !state.reasoningOpened {
			if err := emitEvent(state.openReasoningItemEvent()); err != nil {
				return err
			}
			if err := emitEvent(state.nextReasoningSummaryPartAddedEvent()); err != nil {
				return err
			}
		}
		event := state.nextReasoningDeltaEvent(token)
		state.accumulator.ProcessEvent(&event)
		if err := emitEvent(event); err != nil {
			return err
		}
		return nil
	}
	appendReferencesSuffix := func() error {
		suffix := state.referencesSuffix()
		if suffix == "" {
			return nil
		}
		if state.messageOpened {
			event := state.nextOutputTextDeltaEvent(suffix)
			state.accumulator.ProcessEvent(&event)
			return emitEvent(event)
		}
		state.accumulator.ProcessEvent(&apicompat.ResponsesStreamEvent{
			Type:  "response.output_text.delta",
			Delta: suffix,
		})
		return nil
	}
	handleStreamFailure := func(message string) error {
		if callbacks == nil || callbacks.onStreamFailure == nil || !createdEmitted {
			return newGrokResponsesHTTPError(http.StatusBadGateway, "api_error", message)
		}
		if strings.TrimSpace(message) == "" {
			message = "Upstream stream ended without a response"
		}
		if err := callbacks.onStreamFailure(message, state.nextFailedEvent(message)); err != nil {
			return err
		}
		return errGrokSessionTextStreamHandled
	}

	for scanner.Scan() {
		delta, ok := parseGrokSessionResponseDelta(scanner.Text())
		if !ok {
			continue
		}
		if delta.errorText != "" {
			return nil, grokSessionResponseMetadata{}, handleStreamFailure(delta.errorText)
		}
		applyGrokSessionDeltaMetadata(&state, delta)
		emitSearchSources()

		if delta.token != "" {
			sawDelta = true
			if delta.reasoning {
				if err := emitReasoning(delta.token); err != nil {
					return nil, grokSessionResponseMetadata{}, err
				}
			} else {
				if toolCallsSeen {
					continue
				}
				if err := emitReasoningDone(); err != nil {
					return nil, grokSessionResponseMetadata{}, err
				}
				textToken := delta.token
				if toolSieve != nil {
					safeText, calls, ok := toolSieve.Feed(delta.token)
					textToken = safeText
					if ok {
						if err := emitToolCalls(calls); err != nil {
							return nil, grokSessionResponseMetadata{}, err
						}
						continue
					}
				}
				if textToken == "" {
					continue
				}
				if err := emitCreated(); err != nil {
					return nil, grokSessionResponseMetadata{}, err
				}
				if !state.messageOpened {
					if err := emitEvent(state.openMessageItemEvent()); err != nil {
						return nil, grokSessionResponseMetadata{}, err
					}
					if err := emitEvent(state.nextContentPartAddedEvent()); err != nil {
						return nil, grokSessionResponseMetadata{}, err
					}
				}
				cleanedToken, localAnnotations := cleanGrokSessionTextToken(&state, textToken)
				if cleanedToken == "" {
					continue
				}
				annotationStart := len(state.messageAnnotations)
				state.recordMessageText(cleanedToken, localAnnotations)
				event := state.nextOutputTextDeltaEvent(cleanedToken)
				state.accumulator.ProcessEvent(&event)
				if err := emitEvent(event); err != nil {
					return nil, grokSessionResponseMetadata{}, err
				}
				for idx := annotationStart; idx < len(state.messageAnnotations); idx++ {
					if err := emitEvent(state.nextOutputTextAnnotationAddedEvent(idx, state.messageAnnotations[idx])); err != nil {
						return nil, grokSessionResponseMetadata{}, err
					}
				}
			}
		}

		if delta.completion {
			sawCompletion = true
		}
	}

	if err := scanner.Err(); err != nil {
		if callbacks != nil && callbacks.onStreamFailure != nil && createdEmitted {
			return nil, grokSessionResponseMetadata{}, handleStreamFailure("stream read error")
		}
		return nil, grokSessionResponseMetadata{}, newGrokResponsesHTTPError(http.StatusBadGateway, "api_error", "stream read error")
	}

	if !sawDelta && !sawCompletion {
		return nil, grokSessionResponseMetadata{}, newGrokResponsesHTTPError(http.StatusBadGateway, "api_error", "Upstream stream ended without a response")
	}
	if toolSieve != nil && !toolCallsSeen {
		if calls, ok := toolSieve.Flush(); ok && len(calls) > 0 {
			if err := emitToolCalls(calls); err != nil {
				return nil, grokSessionResponseMetadata{}, err
			}
		}
	}
	if !toolCallsSeen {
		if err := appendReferencesSuffix(); err != nil {
			return nil, grokSessionResponseMetadata{}, err
		}
	}
	if err := emitReasoningDone(); err != nil {
		return nil, grokSessionResponseMetadata{}, err
	}
	if err := emitMessageDone(); err != nil {
		return nil, grokSessionResponseMetadata{}, err
	}

	finalResponse := state.finalResponse()
	if len(toolNames) > 0 && !toolCallsSeen {
		finalResponse = rewriteGrokSessionFinalResponseForToolCalls(finalResponse, toolNames)
	}
	if err := emitCreated(); err != nil {
		return nil, grokSessionResponseMetadata{}, err
	}
	if err := emitEvent(state.nextCompletedEvent(finalResponse)); err != nil {
		return nil, grokSessionResponseMetadata{}, err
	}
	return finalResponse, state.metadata(), nil
}

func parseGrokSessionResponseDelta(line string) (grokSessionResponseDelta, bool) {
	payload, ok := normalizeGrokSessionResponseLine(line)
	if !ok {
		return grokSessionResponseDelta{}, false
	}

	if errText := strings.TrimSpace(gjson.Get(payload, "error.message").String()); errText != "" {
		return grokSessionResponseDelta{errorText: sanitizeUpstreamErrorMessage(errText)}, true
	}
	if errText := strings.TrimSpace(gjson.Get(payload, "message").String()); errText != "" && !gjson.Get(payload, "result.response").Exists() {
		return grokSessionResponseDelta{errorText: sanitizeUpstreamErrorMessage(errText)}, true
	}

	responsePath := gjson.Get(payload, "result.response")
	if !responsePath.Exists() {
		return grokSessionResponseDelta{}, false
	}

	finalMetadata := responsePath.Get("finalMetadata")
	completion := responsePath.Get("isSoftStop").Bool() || (finalMetadata.Exists() && finalMetadata.Type != gjson.Null)
	token := responsePath.Get("token").String()
	reasoning := responsePath.Get("isThinking").Bool()
	cardAttachmentJSON := strings.TrimSpace(responsePath.Get("cardAttachment.jsonData").String())
	webSearchResults := strings.TrimSpace(responsePath.Get("webSearchResults").Raw)
	xSearchResults := strings.TrimSpace(responsePath.Get("xSearchResults").Raw)
	if token == "" && !completion && cardAttachmentJSON == "" && webSearchResults == "" && xSearchResults == "" {
		return grokSessionResponseDelta{}, false
	}

	return grokSessionResponseDelta{
		token:              token,
		reasoning:          reasoning,
		completion:         completion,
		cardAttachmentJSON: cardAttachmentJSON,
		webSearchResults:   webSearchResults,
		xSearchResults:     xSearchResults,
	}, true
}

func normalizeGrokSessionResponseLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, ":") || strings.HasPrefix(trimmed, "event:") {
		return "", false
	}
	if strings.HasPrefix(trimmed, "data:") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
	}
	if trimmed == "" || trimmed == "[DONE]" {
		return "", false
	}
	if !gjson.Valid(trimmed) {
		return "", false
	}
	return trimmed, true
}

func openGrokSessionSSEWriter(c *gin.Context) (*bufio.Writer, http.Flusher, error) {
	if c == nil {
		return nil, nil, newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "request context is nil")
	}
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil, nil, newGrokResponsesHTTPError(http.StatusInternalServerError, "api_error", "streaming not supported")
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	return bufio.NewWriterSize(c.Writer, 4*1024), flusher, nil
}

func writeGrokSessionStreamPayload(writer *bufio.Writer, flusher http.Flusher, payload string) error {
	if writer == nil || flusher == nil {
		return fmt.Errorf("stream writer is not configured")
	}
	if _, err := writer.WriteString(payload); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func writeGrokSessionResponsesEvent(writer *bufio.Writer, flusher http.Flusher, event apicompat.ResponsesStreamEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return writeGrokSessionStreamPayload(writer, flusher, "data: "+string(payload)+"\n\n")
}

func writeGrokSessionChatChunks(
	writer *bufio.Writer,
	flusher http.Flusher,
	state *apicompat.ResponsesEventToChatState,
	event apicompat.ResponsesStreamEvent,
) error {
	return writeGrokSessionChatChunkBatch(writer, flusher, apicompat.ResponsesEventToChatChunks(&event, state))
}

func writeGrokSessionChatChunkBatch(writer *bufio.Writer, flusher http.Flusher, chunks []apicompat.ChatCompletionsChunk) error {
	for _, chunk := range chunks {
		sse, err := apicompat.ChatChunkToSSE(chunk)
		if err != nil {
			return err
		}
		if err := writeGrokSessionStreamPayload(writer, flusher, sse); err != nil {
			return err
		}
	}
	return nil
}

func writeGrokSessionAnthropicEvents(
	writer *bufio.Writer,
	flusher http.Flusher,
	state *apicompat.ResponsesEventToAnthropicState,
	event apicompat.ResponsesStreamEvent,
) error {
	return writeGrokSessionAnthropicEventBatch(writer, flusher, apicompat.ResponsesEventToAnthropicEvents(&event, state))
}

func writeGrokSessionAnthropicEventBatch(writer *bufio.Writer, flusher http.Flusher, events []apicompat.AnthropicStreamEvent) error {
	for _, event := range events {
		sse, err := apicompat.ResponsesAnthropicEventToSSE(event)
		if err != nil {
			return err
		}
		if err := writeGrokSessionStreamPayload(writer, flusher, sse); err != nil {
			return err
		}
	}
	return nil
}

func (s *grokSessionResponsesState) nextCreatedEvent() apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.created", &apicompat.ResponsesStreamEvent{
		Response: &apicompat.ResponsesResponse{
			ID:     s.responseID,
			Object: "response",
			Model:  s.model,
			Status: "in_progress",
			Output: []apicompat.ResponsesOutput{},
		},
	})
}

func (s *grokSessionResponsesState) openReasoningItemEvent() apicompat.ResponsesStreamEvent {
	if !s.reasoningOpened {
		s.reasoningOpened = true
		s.reasoningOutputIndex = 0
		s.reasoningItemID = newGrokSessionResponseID("item_reasoning")
	}
	return s.nextEvent("response.output_item.added", &apicompat.ResponsesStreamEvent{
		OutputIndex: s.reasoningOutputIndex,
		Item: &apicompat.ResponsesOutput{
			Type:   "reasoning",
			ID:     s.reasoningItemID,
			Status: "in_progress",
		},
	})
}

func (s *grokSessionResponsesState) openMessageItemEvent() apicompat.ResponsesStreamEvent {
	if !s.messageOpened {
		s.messageOpened = true
		if s.reasoningOpened {
			s.messageOutputIndex = 1
		}
		s.messageItemID = newGrokSessionResponseID("item_message")
	}
	return s.nextEvent("response.output_item.added", &apicompat.ResponsesStreamEvent{
		OutputIndex: s.messageOutputIndex,
		Item: &apicompat.ResponsesOutput{
			Type:   "message",
			ID:     s.messageItemID,
			Role:   "assistant",
			Status: "in_progress",
		},
	})
}

func (s *grokSessionResponsesState) nextReasoningDeltaEvent(token string) apicompat.ResponsesStreamEvent {
	if !s.reasoningOpened {
		s.reasoningOpened = true
		s.reasoningOutputIndex = 0
		s.reasoningItemID = newGrokSessionResponseID("item_reasoning")
	}
	return s.nextEvent("response.reasoning_summary_text.delta", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.reasoningOutputIndex,
		SummaryIndex: 0,
		Delta:        token,
		ItemID:       s.reasoningItemID,
	})
}

func (s *grokSessionResponsesState) nextReasoningDoneEvent() apicompat.ResponsesStreamEvent {
	text := ""
	if item := s.currentReasoningOutput(); item != nil && len(item.Summary) > 0 {
		text = item.Summary[0].Text
	}
	return s.nextEvent("response.reasoning_summary_text.done", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.reasoningOutputIndex,
		SummaryIndex: 0,
		Text:         text,
		ItemID:       s.reasoningItemID,
	})
}

func (s *grokSessionResponsesState) nextReasoningSummaryPartAddedEvent() apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.reasoning_summary_part.added", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.reasoningOutputIndex,
		SummaryIndex: 0,
		ItemID:       s.reasoningItemID,
		Part: &apicompat.ResponsesContentPart{
			Type: "summary_text",
			Text: "",
		},
	})
}

func (s *grokSessionResponsesState) nextReasoningSummaryPartDoneEvent() apicompat.ResponsesStreamEvent {
	text := ""
	if item := s.currentReasoningOutput(); item != nil && len(item.Summary) > 0 {
		text = item.Summary[0].Text
	}
	return s.nextEvent("response.reasoning_summary_part.done", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.reasoningOutputIndex,
		SummaryIndex: 0,
		ItemID:       s.reasoningItemID,
		Part: &apicompat.ResponsesContentPart{
			Type: "summary_text",
			Text: text,
		},
	})
}

func (s *grokSessionResponsesState) nextReasoningItemDoneEvent() apicompat.ResponsesStreamEvent {
	item := &apicompat.ResponsesOutput{
		Type:   "reasoning",
		ID:     s.reasoningItemID,
		Status: "completed",
	}
	if current := s.currentReasoningOutput(); current != nil && len(current.Summary) > 0 {
		item.Summary = append([]apicompat.ResponsesSummary(nil), current.Summary...)
	}
	return s.nextEvent("response.output_item.done", &apicompat.ResponsesStreamEvent{
		OutputIndex: s.reasoningOutputIndex,
		Item:        item,
	})
}

func (s *grokSessionResponsesState) nextOutputTextDeltaEvent(token string) apicompat.ResponsesStreamEvent {
	if !s.messageOpened {
		s.messageOpened = true
		if s.reasoningOpened {
			s.messageOutputIndex = 1
		}
		s.messageItemID = newGrokSessionResponseID("item_message")
	}
	return s.nextEvent("response.output_text.delta", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.messageOutputIndex,
		ContentIndex: 0,
		Delta:        token,
		ItemID:       s.messageItemID,
	})
}

func (s *grokSessionResponsesState) nextContentPartAddedEvent() apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.content_part.added", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.messageOutputIndex,
		ContentIndex: 0,
		ItemID:       s.messageItemID,
		Part: &apicompat.ResponsesContentPart{
			Type: "output_text",
			Text: "",
		},
	})
}

func (s *grokSessionResponsesState) nextOutputTextDoneEvent() apicompat.ResponsesStreamEvent {
	text := ""
	if item := s.currentMessageOutput(); item != nil {
		for _, part := range item.Content {
			if part.Type == "output_text" {
				text = part.Text
				break
			}
		}
	}
	return s.nextEvent("response.output_text.done", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.messageOutputIndex,
		ContentIndex: 0,
		Text:         text,
		ItemID:       s.messageItemID,
	})
}

func (s *grokSessionResponsesState) nextContentPartDoneEvent() apicompat.ResponsesStreamEvent {
	part := apicompat.ResponsesContentPart{Type: "output_text"}
	if item := s.currentMessageOutput(); item != nil {
		for _, candidate := range item.Content {
			if candidate.Type != "output_text" {
				continue
			}
			part = candidate
			break
		}
	}
	return s.nextEvent("response.content_part.done", &apicompat.ResponsesStreamEvent{
		OutputIndex:  s.messageOutputIndex,
		ContentIndex: 0,
		ItemID:       s.messageItemID,
		Part:         &part,
	})
}

func (s *grokSessionResponsesState) nextOutputTextAnnotationAddedEvent(
	annotationIndex int,
	annotation apicompat.ResponsesTextAnnotation,
) apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.output_text.annotation.added", &apicompat.ResponsesStreamEvent{
		OutputIndex:     s.messageOutputIndex,
		ContentIndex:    0,
		ItemID:          s.messageItemID,
		AnnotationIndex: annotationIndex,
		Annotation:      &annotation,
	})
}

func (s *grokSessionResponsesState) nextMessageItemDoneEvent() apicompat.ResponsesStreamEvent {
	item := &apicompat.ResponsesOutput{
		Type:   "message",
		ID:     s.messageItemID,
		Role:   "assistant",
		Status: "completed",
	}
	if current := s.currentMessageOutput(); current != nil {
		if len(current.Content) > 0 {
			item.Content = append([]apicompat.ResponsesContentPart(nil), current.Content...)
		}
		if len(current.SearchSources) > 0 {
			item.SearchSources = append([]apicompat.SearchSource(nil), current.SearchSources...)
		}
	}
	return s.nextEvent("response.output_item.done", &apicompat.ResponsesStreamEvent{
		OutputIndex: s.messageOutputIndex,
		Item:        item,
	})
}

func (s *grokSessionResponsesState) nextFunctionCallOutputIndex() int {
	index := 0
	if s.reasoningOpened {
		index++
	}
	if s.messageOpened {
		index++
	}
	return index
}

func (s *grokSessionResponsesState) nextFunctionCallAddedEvent(outputIndex int, call grokParsedToolCall) apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.output_item.added", &apicompat.ResponsesStreamEvent{
		OutputIndex: outputIndex,
		Item: &apicompat.ResponsesOutput{
			ID:     newGrokSessionResponseID("item_function"),
			Type:   "function_call",
			CallID: call.CallID,
			Name:   call.Name,
			Status: "in_progress",
		},
	})
}

func (s *grokSessionResponsesState) nextFunctionCallArgumentsDeltaEvent(outputIndex int, itemID, arguments string) apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.function_call_arguments.delta", &apicompat.ResponsesStreamEvent{
		OutputIndex: outputIndex,
		ItemID:      itemID,
		Delta:       arguments,
	})
}

func (s *grokSessionResponsesState) nextFunctionCallArgumentsDoneEvent(outputIndex int, itemID, arguments string) apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.function_call_arguments.done", &apicompat.ResponsesStreamEvent{
		OutputIndex: outputIndex,
		ItemID:      itemID,
		Arguments:   arguments,
	})
}

func (s *grokSessionResponsesState) nextFunctionCallDoneEvent(outputIndex int, call grokParsedToolCall, itemID string) apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.output_item.done", &apicompat.ResponsesStreamEvent{
		OutputIndex: outputIndex,
		Item: &apicompat.ResponsesOutput{
			ID:        itemID,
			Type:      "function_call",
			CallID:    call.CallID,
			Name:      call.Name,
			Arguments: call.Arguments,
			Status:    "completed",
		},
	})
}

func (s *grokSessionResponsesState) nextCompletedEvent(resp *apicompat.ResponsesResponse) apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.completed", &apicompat.ResponsesStreamEvent{
		Response: resp,
	})
}

func (s *grokSessionResponsesState) nextFailedEvent(message string) apicompat.ResponsesStreamEvent {
	return s.nextEvent("response.failed", &apicompat.ResponsesStreamEvent{
		Response: &apicompat.ResponsesResponse{
			ID:     s.responseID,
			Object: "response",
			Model:  s.model,
			Status: "failed",
			Output: []apicompat.ResponsesOutput{},
			Error: &apicompat.ResponsesError{
				Code:    "api_error",
				Message: message,
			},
		},
	})
}

func (s *grokSessionResponsesState) nextEvent(eventType string, template *apicompat.ResponsesStreamEvent) apicompat.ResponsesStreamEvent {
	event := apicompat.ResponsesStreamEvent{Type: eventType}
	if template != nil {
		event = *template
		event.Type = eventType
	}
	event.SequenceNumber = s.sequenceNumber
	s.sequenceNumber++
	return event
}

func (s *grokSessionResponsesState) finalResponse() *apicompat.ResponsesResponse {
	response := &apicompat.ResponsesResponse{
		ID:     s.responseID,
		Object: "response",
		Model:  s.model,
		Status: "completed",
		Output: s.accumulator.BuildOutput(),
	}
	s.accumulator.SupplementResponseOutput(response)
	s.applyMessageMetadata(response)
	return response
}

func (s *grokSessionResponsesState) metadata() grokSessionResponseMetadata {
	if s == nil || len(s.searchSources) == 0 {
		return grokSessionResponseMetadata{}
	}
	return grokSessionResponseMetadata{
		SearchSources: append([]apicompat.SearchSource(nil), s.searchSources...),
	}
}

func (s *grokSessionResponsesState) currentReasoningOutput() *apicompat.ResponsesOutput {
	return s.currentOutputByType("reasoning")
}

func (s *grokSessionResponsesState) currentMessageOutput() *apicompat.ResponsesOutput {
	return s.currentOutputByType("message")
}

func (s *grokSessionResponsesState) currentOutputByType(outputType string) *apicompat.ResponsesOutput {
	if s == nil {
		return nil
	}
	response := &apicompat.ResponsesResponse{Output: s.accumulator.BuildOutput()}
	s.applyMessageMetadata(response)
	for i := range response.Output {
		if response.Output[i].Type != outputType {
			continue
		}
		item := response.Output[i]
		return &item
	}
	return nil
}

func rewriteGrokSessionFinalResponseForToolCalls(
	response *apicompat.ResponsesResponse,
	toolNames []string,
) *apicompat.ResponsesResponse {
	if response == nil || len(response.Output) == 0 || len(toolNames) == 0 {
		return response
	}

	messageIndex := -1
	messageText := ""
	for idx, item := range response.Output {
		if item.Type != "message" || len(item.Content) == 0 {
			continue
		}
		for _, part := range item.Content {
			if part.Type == "output_text" && strings.TrimSpace(part.Text) != "" {
				messageIndex = idx
				messageText = part.Text
				break
			}
		}
		if messageIndex >= 0 {
			break
		}
	}
	if messageIndex < 0 || strings.TrimSpace(messageText) == "" {
		return response
	}

	parsed := parseGrokToolCalls(messageText, toolNames)
	if len(parsed.Calls) == 0 {
		return response
	}

	rewritten := make([]apicompat.ResponsesOutput, 0, len(response.Output)-1+len(parsed.Calls))
	rewritten = append(rewritten, response.Output[:messageIndex]...)
	for _, call := range parsed.Calls {
		rewritten = append(rewritten, apicompat.ResponsesOutput{
			Type:      "function_call",
			ID:        newGrokSessionResponseID("fc"),
			CallID:    call.CallID,
			Name:      call.Name,
			Arguments: call.Arguments,
			Status:    "completed",
		})
	}
	rewritten = append(rewritten, response.Output[messageIndex+1:]...)
	response.Output = rewritten
	return response
}

func newGrokSessionResponseID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func applyGrokSessionDeltaMetadata(state *grokSessionResponsesState, delta grokSessionResponseDelta) {
	if state == nil {
		return
	}
	if cardID := strings.TrimSpace(gjson.Get(delta.cardAttachmentJSON, "id").String()); cardID != "" {
		state.cardAttachments[cardID] = delta.cardAttachmentJSON
	}
	cacheGrokSessionWebSearchResults(state, delta.webSearchResults)
	cacheGrokSessionXSearchResults(state, delta.xSearchResults)
}

func cleanGrokSessionTextToken(state *grokSessionResponsesState, token string) (string, []grokSessionLocalAnnotation) {
	if state == nil || !strings.Contains(token, "<grok:render") {
		return token, nil
	}
	cleaned := grokSessionRenderRe.ReplaceAllStringFunc(token, func(match string) string {
		return replaceGrokSessionRender(state, match)
	})
	if strings.HasPrefix(cleaned, "\n") && strings.Contains(cleaned, "[[") {
		cleaned = strings.TrimLeft(cleaned, "\n")
	}
	if len(state.pendingCitations) == 0 {
		return cleaned, nil
	}

	localAnnotations := make([]grokSessionLocalAnnotation, 0, len(state.pendingCitations))
	searchStart := 0
	for _, pending := range state.pendingCitations {
		relative := strings.Index(cleaned[searchStart:], pending.Needle)
		if relative < 0 {
			continue
		}
		startByte := searchStart + relative
		startIndex := utf8.RuneCountInString(cleaned[:startByte])
		localAnnotations = append(localAnnotations, grokSessionLocalAnnotation{
			URL:        pending.URL,
			Title:      pending.Title,
			StartIndex: startIndex,
			EndIndex:   startIndex + utf8.RuneCountInString(pending.Needle),
		})
		searchStart = startByte + len(pending.Needle)
	}
	state.pendingCitations = nil
	return cleaned, localAnnotations
}

func replaceGrokSessionRender(state *grokSessionResponsesState, match string) string {
	if state == nil {
		return ""
	}
	parts := grokSessionRenderRe.FindStringSubmatch(match)
	if len(parts) < 4 {
		return ""
	}
	cardJSON := state.cardAttachments[parts[1]]
	if cardJSON == "" {
		return ""
	}

	card := gjson.Parse(cardJSON)
	switch parts[3] {
	case "render_searched_image":
		title := strings.TrimSpace(card.Get("image.title").String())
		if title == "" {
			title = "image"
		}
		thumbnail := strings.TrimSpace(card.Get("image.thumbnail").String())
		if thumbnail == "" {
			thumbnail = strings.TrimSpace(card.Get("image.original").String())
		}
		link := strings.TrimSpace(card.Get("image.link").String())
		if thumbnail == "" && link == "" {
			return ""
		}
		if thumbnail == "" {
			return fmt.Sprintf("[%s](%s)", title, link)
		}
		if link != "" {
			return fmt.Sprintf("[![%s](%s)](%s)", title, thumbnail, link)
		}
		return fmt.Sprintf("![%s](%s)", title, thumbnail)
	case "render_generated_image":
		return ""
	case "render_inline_citation":
		url := strings.TrimSpace(card.Get("url").String())
		if url == "" {
			return ""
		}
		index, ok := state.citationIndexByURL[url]
		if !ok {
			state.nextCitationIndex++
			index = state.nextCitationIndex
			state.citationIndexByURL[url] = index
		}
		if index == state.lastCitationIndex {
			return ""
		}
		state.lastCitationIndex = index
		title := state.searchSourceTitle[url]
		if title == "" {
			title = strings.TrimSpace(card.Get("title").String())
		}
		if title == "" {
			title = url
		}
		citationText := fmt.Sprintf(" [[%d]](%s)", index, url)
		state.pendingCitations = append(state.pendingCitations, grokSessionPendingCitation{
			URL:    url,
			Title:  title,
			Needle: citationText,
		})
		return citationText
	default:
		return ""
	}
}

func (s *grokSessionResponsesState) recordMessageText(
	cleaned string,
	localAnnotations []grokSessionLocalAnnotation,
) {
	if s == nil || cleaned == "" {
		return
	}
	for _, ann := range localAnnotations {
		s.messageAnnotations = append(s.messageAnnotations, apicompat.ResponsesTextAnnotation{
			Type:       "url_citation",
			URL:        ann.URL,
			Title:      ann.Title,
			StartIndex: s.messageTextOffset + ann.StartIndex,
			EndIndex:   s.messageTextOffset + ann.EndIndex,
		})
	}
	s.messageTextOffset += utf8.RuneCountInString(cleaned)
}

func (s *grokSessionResponsesState) applyMessageMetadata(response *apicompat.ResponsesResponse) {
	if s == nil || response == nil {
		return
	}
	for i := range response.Output {
		if response.Output[i].Type != "message" {
			continue
		}
		if len(s.searchSources) > 0 && len(response.Output[i].SearchSources) == 0 {
			response.Output[i].SearchSources = append([]apicompat.SearchSource(nil), s.searchSources...)
		}
		if len(s.messageAnnotations) == 0 {
			return
		}
		for j := range response.Output[i].Content {
			if response.Output[i].Content[j].Type != "output_text" {
				continue
			}
			if len(response.Output[i].Content[j].Annotations) == 0 {
				response.Output[i].Content[j].Annotations = append(
					[]apicompat.ResponsesTextAnnotation(nil),
					s.messageAnnotations...,
				)
			}
			return
		}
		return
	}
}

func (s *grokSessionResponsesState) referencesSuffix() string {
	if s == nil || len(s.searchSources) == 0 {
		return ""
	}
	lines := []string{"\n\n## Sources", "[grok2api-sources]: #"}
	for _, source := range s.searchSources {
		title := firstNonEmpty(strings.TrimSpace(source.Title), strings.TrimSpace(source.URL))
		if title == "" || strings.TrimSpace(source.URL) == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- [%s](%s)", escapeGrokSessionMarkdownLinkText(title), source.URL))
	}
	if len(lines) == 2 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func cacheGrokSessionWebSearchResults(state *grokSessionResponsesState, raw string) {
	if state == nil || raw == "" {
		return
	}
	for _, item := range gjson.Parse(raw).Get("results").Array() {
		url := strings.TrimSpace(item.Get("url").String())
		if url == "" {
			continue
		}
		state.addSearchSource(url, strings.TrimSpace(item.Get("title").String()), "web")
	}
}

func cacheGrokSessionXSearchResults(state *grokSessionResponsesState, raw string) {
	if state == nil || raw == "" {
		return
	}
	for _, item := range gjson.Parse(raw).Get("results").Array() {
		username := strings.TrimSpace(item.Get("username").String())
		postID := strings.TrimSpace(item.Get("postId").String())
		if username == "" || postID == "" {
			continue
		}
		url := fmt.Sprintf("https://x.com/%s/status/%s", username, postID)
		title := fmt.Sprintf("X/@%s", username)
		if text := truncateGrokSessionSearchText(normalizeGrokSessionSearchText(item.Get("text").String()), 50); text != "" {
			title = fmt.Sprintf("X/@%s: %s", username, text)
		}
		state.addSearchSource(url, title, "x_post")
	}
}

func (s *grokSessionResponsesState) addSearchSource(url, title, sourceType string) {
	if s == nil || url == "" {
		return
	}
	if title == "" {
		title = url
	}
	if current := s.searchSourceTitle[url]; current == "" || current == url {
		s.searchSourceTitle[url] = title
	}
	if _, seen := s.searchSourceSeen[url]; seen {
		for i := range s.searchSources {
			if s.searchSources[i].URL != url {
				continue
			}
			if s.searchSources[i].Title == "" || s.searchSources[i].Title == url {
				s.searchSources[i].Title = title
			}
			if s.searchSources[i].Type == "" {
				s.searchSources[i].Type = sourceType
			}
		}
		for i := range s.messageAnnotations {
			if s.messageAnnotations[i].URL == url && (s.messageAnnotations[i].Title == "" || s.messageAnnotations[i].Title == url) {
				s.messageAnnotations[i].Title = title
			}
		}
		for i := range s.pendingCitations {
			if s.pendingCitations[i].URL == url && (s.pendingCitations[i].Title == "" || s.pendingCitations[i].Title == url) {
				s.pendingCitations[i].Title = title
			}
		}
		return
	}
	s.searchSourceSeen[url] = struct{}{}
	s.searchSources = append(s.searchSources, apicompat.SearchSource{
		URL:   url,
		Title: title,
		Type:  sourceType,
	})
}

func normalizeGrokSessionSearchText(text string) string {
	if text == "" {
		return ""
	}
	return strings.Join(strings.Fields(text), " ")
}

func truncateGrokSessionSearchText(text string, limit int) string {
	if limit <= 0 || text == "" {
		return text
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}

func escapeGrokSessionMarkdownLinkText(text string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`[`, `\[`,
		`]`, `\]`,
	)
	return replacer.Replace(text)
}
