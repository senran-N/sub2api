package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/util/responseheaders"
	"github.com/tidwall/gjson"
)

type openAIWSForwardRelayRequest struct {
	ctx                       context.Context
	c                         *gin.Context
	account                   *Account
	lease                     *openAIWSConnLease
	payload                   map[string]any
	payloadBytes              int
	reqBody                   map[string]any
	originalModel             string
	mappedModel               string
	reqStream                 bool
	debugEnabled              bool
	previousResponseID        string
	previousResponseKind      string
	sessionResolution         openAIWSSessionHeaderResolution
	wsHeaders                 http.Header
	storeDisabled             bool
	stateStore                OpenAIWSStateStore
	groupID                   int64
	sessionHash               string
	turnState                 string
	promptCacheKey            string
	startTime                 time.Time
	trackCompatibilityMetrics bool
	warmup                    bool
}

type openAIWSForwardRelayResult struct {
	result    *OpenAIForwardResult
	cleanExit bool
}

func (s *OpenAIGatewayService) relayOpenAIWSForwardV2(req openAIWSForwardRelayRequest) (*openAIWSForwardRelayResult, error) {
	if s == nil {
		return nil, errors.New("service is nil")
	}
	if req.account == nil {
		return nil, errors.New("account is nil")
	}
	if req.lease == nil {
		return nil, errors.New("openai ws lease is nil")
	}

	connID := strings.TrimSpace(req.lease.ConnID())
	if err := req.lease.WriteJSONWithContextTimeout(req.ctx, req.payload, s.openAIWSWriteTimeout()); err != nil {
		req.lease.MarkBroken()
		logOpenAIWSModeInfo(
			"write_request_fail account_id=%d conn_id=%s cause=%s payload_bytes=%d",
			req.account.ID,
			connID,
			truncateOpenAIWSLogValue(err.Error(), openAIWSLogValueMaxLen),
			req.payloadBytes,
		)
		return nil, wrapOpenAIWSFallback("write_request", err)
	}
	if req.debugEnabled {
		logOpenAIWSModeDebug(
			"write_request_sent account_id=%d conn_id=%s stream=%v payload_bytes=%d previous_response_id=%s",
			req.account.ID,
			connID,
			req.reqStream,
			req.payloadBytes,
			truncateOpenAIWSLogValue(req.previousResponseID, openAIWSIDValueMaxLen),
		)
	}

	usage := &OpenAIUsage{}
	var firstTokenMs *int
	responseID := ""
	var finalResponse []byte
	wroteDownstream := false
	needModelReplace := req.originalModel != req.mappedModel
	var mappedModelBytes []byte
	if needModelReplace && req.mappedModel != "" {
		mappedModelBytes = []byte(req.mappedModel)
	}
	bufferedStreamEvents := make([][]byte, 0, 4)
	eventCount := 0
	tokenEventCount := 0
	terminalEventCount := 0
	bufferedEventCount := 0
	flushedBufferedEventCount := 0
	firstEventType := ""
	lastEventType := ""

	var flusher http.Flusher
	if req.reqStream {
		if s.responseHeaderFilter != nil {
			responseheaders.WriteFilteredHeaders(req.c.Writer.Header(), http.Header{}, s.responseHeaderFilter)
		}
		req.c.Header("Content-Type", "text/event-stream")
		req.c.Header("Cache-Control", "no-cache")
		req.c.Header("Connection", "keep-alive")
		req.c.Header("X-Accel-Buffering", "no")
		f, ok := req.c.Writer.(http.Flusher)
		if !ok {
			req.lease.MarkBroken()
			return nil, wrapOpenAIWSFallback("streaming_not_supported", errors.New("streaming not supported"))
		}
		flusher = f
	}

	clientDisconnected := false
	flushBatchSize := s.openAIWSEventFlushBatchSize()
	flushInterval := s.openAIWSEventFlushInterval()
	pendingFlushEvents := 0
	lastFlushAt := time.Now()
	flushStreamWriter := func(force bool) {
		if clientDisconnected || flusher == nil || pendingFlushEvents <= 0 {
			return
		}
		if !force && flushBatchSize > 1 && pendingFlushEvents < flushBatchSize {
			if flushInterval <= 0 || time.Since(lastFlushAt) < flushInterval {
				return
			}
		}
		flusher.Flush()
		pendingFlushEvents = 0
		lastFlushAt = time.Now()
	}
	emitStreamMessage := func(message []byte, forceFlush bool) {
		if clientDisconnected {
			return
		}
		frame := make([]byte, 0, len(message)+8)
		frame = append(frame, "data: "...)
		frame = append(frame, message...)
		frame = append(frame, '\n', '\n')
		_, wErr := req.c.Writer.Write(frame)
		if wErr == nil {
			wroteDownstream = true
			pendingFlushEvents++
			flushStreamWriter(forceFlush)
			return
		}
		clientDisconnected = true
		logger.LegacyPrintf("service.openai_gateway", "[OpenAI WS Mode] client disconnected, continue draining upstream: account=%d", req.account.ID)
	}
	flushBufferedStreamEvents := func(reason string) {
		if len(bufferedStreamEvents) == 0 {
			return
		}
		flushed := len(bufferedStreamEvents)
		for _, buffered := range bufferedStreamEvents {
			emitStreamMessage(buffered, false)
		}
		bufferedStreamEvents = bufferedStreamEvents[:0]
		flushStreamWriter(true)
		flushedBufferedEventCount += flushed
		if req.debugEnabled {
			logOpenAIWSModeDebug(
				"buffer_flush account_id=%d conn_id=%s reason=%s flushed=%d total_flushed=%d client_disconnected=%v",
				req.account.ID,
				connID,
				truncateOpenAIWSLogValue(reason, openAIWSLogValueMaxLen),
				flushed,
				flushedBufferedEventCount,
				clientDisconnected,
			)
		}
	}

	readTimeout := s.openAIWSReadTimeout()
	cleanExit := false

	for {
		message, readErr := req.lease.ReadMessageWithContextTimeout(req.ctx, readTimeout)
		if readErr != nil {
			req.lease.MarkBroken()
			closeStatus, closeReason := summarizeOpenAIWSReadCloseError(readErr)
			logOpenAIWSModeInfo(
				"read_fail account_id=%d conn_id=%s wrote_downstream=%v close_status=%s close_reason=%s cause=%s events=%d token_events=%d terminal_events=%d buffered_pending=%d buffered_flushed=%d first_event=%s last_event=%s",
				req.account.ID,
				connID,
				wroteDownstream,
				closeStatus,
				closeReason,
				truncateOpenAIWSLogValue(readErr.Error(), openAIWSLogValueMaxLen),
				eventCount,
				tokenEventCount,
				terminalEventCount,
				len(bufferedStreamEvents),
				flushedBufferedEventCount,
				truncateOpenAIWSLogValue(firstEventType, openAIWSLogValueMaxLen),
				truncateOpenAIWSLogValue(lastEventType, openAIWSLogValueMaxLen),
			)
			if !wroteDownstream {
				return nil, wrapOpenAIWSFallback(classifyOpenAIWSReadFallbackReason(readErr), readErr)
			}
			if clientDisconnected {
				break
			}
			setOpsUpstreamError(req.c, 0, sanitizeUpstreamErrorMessage(readErr.Error()), "")
			return nil, fmt.Errorf("openai ws read event: %w", readErr)
		}

		eventType, eventResponseID, responseField := parseOpenAIWSEventEnvelope(message)
		if eventType == "" {
			continue
		}
		eventCount++
		if firstEventType == "" {
			firstEventType = eventType
		}
		lastEventType = eventType

		if responseID == "" && eventResponseID != "" {
			responseID = eventResponseID
		}

		isTokenEvent := isOpenAIWSTokenEvent(eventType)
		if isTokenEvent {
			tokenEventCount++
		}
		isTerminalEvent := isOpenAIWSTerminalEvent(eventType)
		if isTerminalEvent {
			terminalEventCount++
		}
		if firstTokenMs == nil && isTokenEvent {
			ms := int(time.Since(req.startTime).Milliseconds())
			firstTokenMs = &ms
		}
		if req.debugEnabled && shouldLogOpenAIWSEvent(eventCount, eventType) {
			logOpenAIWSModeDebug(
				"event_received account_id=%d conn_id=%s idx=%d type=%s bytes=%d token=%v terminal=%v buffered_pending=%d",
				req.account.ID,
				connID,
				eventCount,
				truncateOpenAIWSLogValue(eventType, openAIWSLogValueMaxLen),
				len(message),
				isTokenEvent,
				isTerminalEvent,
				len(bufferedStreamEvents),
			)
		}

		if !clientDisconnected {
			if needModelReplace && len(mappedModelBytes) > 0 && openAIWSEventMayContainModel(eventType) && bytes.Contains(message, mappedModelBytes) {
				message = replaceOpenAIWSMessageModel(message, req.mappedModel, req.originalModel)
			}
			if openAIWSEventMayContainToolCalls(eventType) && openAIWSMessageLikelyContainsToolCalls(message) {
				if corrected, changed := s.toolCorrector.CorrectToolCallsInSSEBytes(message); changed {
					message = corrected
				}
			}
		}
		if openAIWSEventShouldParseUsage(eventType) {
			parseOpenAIWSResponseUsageFromCompletedEvent(message, usage)
		}

		if eventType == "error" {
			errCodeRaw, errTypeRaw, errMsgRaw := parseOpenAIWSErrorEventFields(message)
			s.persistOpenAIWSRateLimitSignal(req.ctx, req.account, req.lease.HandshakeHeaders(), message, errCodeRaw, errTypeRaw, errMsgRaw)
			errMsg := strings.TrimSpace(errMsgRaw)
			if errMsg == "" {
				errMsg = "Upstream websocket error"
			}
			fallbackReason, canFallback := classifyOpenAIWSErrorEventFromRaw(errCodeRaw, errTypeRaw, errMsgRaw)
			errCode, errType, errMessage := summarizeOpenAIWSErrorEventFieldsFromRaw(errCodeRaw, errTypeRaw, errMsgRaw)
			logOpenAIWSModeInfo(
				"error_event account_id=%d conn_id=%s idx=%d fallback_reason=%s can_fallback=%v err_code=%s err_type=%s err_message=%s",
				req.account.ID,
				connID,
				eventCount,
				truncateOpenAIWSLogValue(fallbackReason, openAIWSLogValueMaxLen),
				canFallback,
				errCode,
				errType,
				errMessage,
			)
			if fallbackReason == "previous_response_not_found" {
				logOpenAIWSModeInfo(
					"previous_response_not_found_diag account_id=%d account_type=%s conn_id=%s previous_response_id=%s previous_response_id_kind=%s response_id=%s event_idx=%d req_stream=%v store_disabled=%v conn_reused=%v session_hash=%s header_session_id=%s header_conversation_id=%s session_id_source=%s conversation_id_source=%s has_turn_state=%v turn_state_len=%d has_prompt_cache_key=%v err_code=%s err_type=%s err_message=%s",
					req.account.ID,
					req.account.Type,
					connID,
					truncateOpenAIWSLogValue(req.previousResponseID, openAIWSIDValueMaxLen),
					normalizeOpenAIWSLogValue(req.previousResponseKind),
					truncateOpenAIWSLogValue(responseID, openAIWSIDValueMaxLen),
					eventCount,
					req.reqStream,
					req.storeDisabled,
					req.lease.Reused(),
					truncateOpenAIWSLogValue(req.sessionHash, 12),
					openAIWSHeaderValueForLog(req.wsHeaders, "session_id"),
					openAIWSHeaderValueForLog(req.wsHeaders, "conversation_id"),
					normalizeOpenAIWSLogValue(req.sessionResolution.SessionSource),
					normalizeOpenAIWSLogValue(req.sessionResolution.ConversationSource),
					req.turnState != "",
					len(req.turnState),
					req.promptCacheKey != "",
					errCode,
					errType,
					errMessage,
				)
			}
			req.lease.MarkBroken()
			if !wroteDownstream && canFallback {
				return nil, wrapOpenAIWSFallback(fallbackReason, errors.New(errMsg))
			}
			statusCode := openAIWSErrorHTTPStatusFromRaw(errCodeRaw, errTypeRaw)
			setOpsUpstreamError(req.c, statusCode, errMsg, "")
			if req.reqStream && !clientDisconnected {
				flushBufferedStreamEvents("error_event")
				emitStreamMessage(message, true)
			}
			if !req.reqStream {
				req.c.JSON(statusCode, gin.H{
					"error": gin.H{
						"type":    "upstream_error",
						"message": errMsg,
					},
				})
			}
			return nil, fmt.Errorf("openai ws error event: %s", errMsg)
		}

		if req.reqStream {
			shouldBuffer := firstTokenMs == nil && !isTokenEvent && !isTerminalEvent
			if shouldBuffer {
				buffered := make([]byte, len(message))
				copy(buffered, message)
				bufferedStreamEvents = append(bufferedStreamEvents, buffered)
				bufferedEventCount++
				if req.debugEnabled && shouldLogOpenAIWSBufferedEvent(bufferedEventCount) {
					logOpenAIWSModeDebug(
						"buffer_enqueue account_id=%d conn_id=%s idx=%d event_idx=%d event_type=%s buffer_size=%d",
						req.account.ID,
						connID,
						bufferedEventCount,
						eventCount,
						truncateOpenAIWSLogValue(eventType, openAIWSLogValueMaxLen),
						len(bufferedStreamEvents),
					)
				}
			} else {
				flushBufferedStreamEvents(eventType)
				emitStreamMessage(message, isTerminalEvent)
			}
		} else if responseField.Exists() && responseField.Type == gjson.JSON {
			finalResponse = []byte(responseField.Raw)
		}

		if isTerminalEvent {
			cleanExit = true
			break
		}
	}

	if !req.reqStream {
		if len(finalResponse) == 0 {
			logOpenAIWSModeInfo(
				"missing_final_response account_id=%d conn_id=%s events=%d token_events=%d terminal_events=%d wrote_downstream=%v",
				req.account.ID,
				connID,
				eventCount,
				tokenEventCount,
				terminalEventCount,
				wroteDownstream,
			)
			if !wroteDownstream {
				return nil, wrapOpenAIWSFallback("missing_final_response", errors.New("no terminal response payload"))
			}
			return nil, errors.New("ws finished without final response")
		}

		if needModelReplace {
			finalResponse = s.replaceModelInResponseBody(finalResponse, req.mappedModel, req.originalModel)
		}
		finalResponse = s.correctToolCallsInResponseBody(finalResponse)
		populateOpenAIUsageFromResponseJSON(finalResponse, usage)
		if responseID == "" {
			responseID = strings.TrimSpace(gjson.GetBytes(finalResponse, "id").String())
		}

		req.c.Data(http.StatusOK, "application/json", finalResponse)
	} else {
		flushStreamWriter(true)
	}

	s.bindCodexChainSuccess(req.ctx, req.stateStore, codexChainBinding{
		AccountID:     req.account.ID,
		ConnID:        req.lease.ConnID(),
		GroupID:       req.groupID,
		ResponseID:    responseID,
		ResponseTTL:   s.openAIWSResponseStickyTTL(),
		SessionHash:   req.sessionHash,
		SessionTTL:    s.openAIWSSessionStickyTTL(),
		StoreDisabled: req.storeDisabled,
		Transport:     OpenAIUpstreamTransportResponsesWebsocketV2,
	})
	s.bindCodexSessionTransport(req.stateStore, req.groupID, req.sessionHash, OpenAIUpstreamTransportResponsesWebsocketV2, req.warmup, req.trackCompatibilityMetrics)
	firstTokenMsValue := -1
	if firstTokenMs != nil {
		firstTokenMsValue = *firstTokenMs
	}
	logOpenAIWSModeDebug(
		"completed account_id=%d conn_id=%s response_id=%s stream=%v duration_ms=%d events=%d token_events=%d terminal_events=%d buffered_events=%d buffered_flushed=%d first_event=%s last_event=%s first_token_ms=%d wrote_downstream=%v client_disconnected=%v",
		req.account.ID,
		connID,
		truncateOpenAIWSLogValue(strings.TrimSpace(responseID), openAIWSIDValueMaxLen),
		req.reqStream,
		time.Since(req.startTime).Milliseconds(),
		eventCount,
		tokenEventCount,
		terminalEventCount,
		bufferedEventCount,
		flushedBufferedEventCount,
		truncateOpenAIWSLogValue(firstEventType, openAIWSLogValueMaxLen),
		truncateOpenAIWSLogValue(lastEventType, openAIWSLogValueMaxLen),
		firstTokenMsValue,
		wroteDownstream,
		clientDisconnected,
	)

	return &openAIWSForwardRelayResult{
		result: &OpenAIForwardResult{
			RequestID:       responseID,
			Usage:           *usage,
			Model:           req.originalModel,
			UpstreamModel:   req.mappedModel,
			ServiceTier:     extractOpenAIServiceTier(req.reqBody),
			ReasoningEffort: extractOpenAIReasoningEffort(req.reqBody, req.originalModel),
			Stream:          req.reqStream,
			OpenAIWSMode:    true,
			ResponseHeaders: req.lease.HandshakeHeaders(),
			Duration:        time.Since(req.startTime),
			FirstTokenMs:    firstTokenMs,
		},
		cleanExit: cleanExit,
	}, nil
}
