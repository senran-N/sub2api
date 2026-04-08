package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	coderws "github.com/coder/websocket"
)

func (s *OpenAIGatewayService) writeOpenAIWSIngressClientMessage(ctx context.Context, clientConn *coderws.Conn, message []byte) error {
	if clientConn == nil {
		return errors.New("client websocket is nil")
	}
	writeCtx, cancel := context.WithTimeout(ctx, s.openAIWSWriteTimeout())
	defer cancel()
	return clientConn.Write(writeCtx, coderws.MessageText, message)
}

func (s *OpenAIGatewayService) readOpenAIWSIngressClientMessage(ctx context.Context, clientConn *coderws.Conn) ([]byte, error) {
	if clientConn == nil {
		return nil, errors.New("client websocket is nil")
	}
	msgType, payload, readErr := clientConn.Read(ctx)
	if readErr != nil {
		return nil, readErr
	}
	if msgType != coderws.MessageText && msgType != coderws.MessageBinary {
		return nil, NewOpenAIWSClientCloseError(
			coderws.StatusPolicyViolation,
			fmt.Sprintf("unsupported websocket client message type: %s", msgType.String()),
			nil,
		)
	}
	return payload, nil
}

func (s *OpenAIGatewayService) relayOpenAIWSIngressTurn(
	ctx context.Context,
	clientConn *coderws.Conn,
	account *Account,
	lease *openAIWSConnLease,
	turn int,
	payload []byte,
	payloadBytes int,
	payloadMeta openAIWSIngressPayloadMeta,
	originalModel string,
	debugEnabled bool,
) (*OpenAIForwardResult, error) {
	if lease == nil {
		return nil, errors.New("upstream websocket lease is nil")
	}
	if account == nil {
		return nil, errors.New("account is nil")
	}

	turnStart := time.Now()
	wroteDownstream := false
	if err := lease.WriteJSONWithContextTimeout(ctx, json.RawMessage(payload), s.openAIWSWriteTimeout()); err != nil {
		return nil, wrapOpenAIWSIngressTurnError(
			"write_upstream",
			fmt.Errorf("write upstream websocket request: %w", err),
			false,
		)
	}
	if debugEnabled {
		logOpenAIWSModeDebug(
			"ingress_ws_turn_request_sent account_id=%d turn=%d conn_id=%s payload_bytes=%d",
			account.ID,
			turn,
			truncateOpenAIWSLogValue(lease.ConnID(), openAIWSIDValueMaxLen),
			payloadBytes,
		)
	}

	responseID := ""
	usage := OpenAIUsage{}
	var firstTokenMs *int
	reqStream := payloadMeta.stream
	turnPreviousResponseID := payloadMeta.previousResponseID
	turnPreviousResponseIDKind := payloadMeta.previousResponseIDKind
	turnStoreDisabled := payloadMeta.storeDisabled
	turnHasFunctionCallOutput := payloadMeta.hasFunctionCallOutput
	eventCount := 0
	tokenEventCount := 0
	terminalEventCount := 0
	firstEventType := ""
	lastEventType := ""
	needModelReplace := false
	clientDisconnected := false
	mappedModel := ""
	var mappedModelBytes []byte
	if originalModel != "" {
		mappedModel = account.GetMappedModel(originalModel)
		if normalizedModel := normalizeCodexModel(mappedModel); normalizedModel != "" {
			mappedModel = normalizedModel
		}
		needModelReplace = mappedModel != "" && mappedModel != originalModel
		if needModelReplace {
			mappedModelBytes = []byte(mappedModel)
		}
	}

	for {
		upstreamMessage, readErr := lease.ReadMessageWithContextTimeout(ctx, s.openAIWSReadTimeout())
		if readErr != nil {
			lease.MarkBroken()
			return nil, wrapOpenAIWSIngressTurnError(
				"read_upstream",
				fmt.Errorf("read upstream websocket event: %w", readErr),
				wroteDownstream,
			)
		}

		eventType, eventResponseID, _ := parseOpenAIWSEventEnvelope(upstreamMessage)
		if responseID == "" && eventResponseID != "" {
			responseID = eventResponseID
		}
		if eventType != "" {
			eventCount++
			if firstEventType == "" {
				firstEventType = eventType
			}
			lastEventType = eventType
		}
		if eventType == "error" {
			errCodeRaw, errTypeRaw, errMsgRaw := parseOpenAIWSErrorEventFields(upstreamMessage)
			s.persistOpenAIWSRateLimitSignal(ctx, account, lease.HandshakeHeaders(), upstreamMessage, errCodeRaw, errTypeRaw, errMsgRaw)
			fallbackReason, _ := classifyOpenAIWSErrorEventFromRaw(errCodeRaw, errTypeRaw, errMsgRaw)
			errCode, errType, errMessage := summarizeOpenAIWSErrorEventFieldsFromRaw(errCodeRaw, errTypeRaw, errMsgRaw)
			recoverablePrevNotFound := fallbackReason == openAIWSIngressStagePreviousResponseNotFound &&
				turnPreviousResponseID != "" &&
				!turnHasFunctionCallOutput &&
				s.openAIWSIngressPreviousResponseRecoveryEnabled() &&
				!wroteDownstream
			if recoverablePrevNotFound {
				logOpenAIWSModeInfo(
					"ingress_ws_prev_response_recoverable account_id=%d turn=%d conn_id=%s idx=%d reason=%s code=%s type=%s message=%s previous_response_id=%s previous_response_id_kind=%s response_id=%s store_disabled=%v has_prompt_cache_key=%v",
					account.ID,
					turn,
					truncateOpenAIWSLogValue(lease.ConnID(), openAIWSIDValueMaxLen),
					eventCount,
					truncateOpenAIWSLogValue(fallbackReason, openAIWSLogValueMaxLen),
					errCode,
					errType,
					errMessage,
					truncateOpenAIWSLogValue(turnPreviousResponseID, openAIWSIDValueMaxLen),
					normalizeOpenAIWSLogValue(turnPreviousResponseIDKind),
					truncateOpenAIWSLogValue(responseID, openAIWSIDValueMaxLen),
					turnStoreDisabled,
					payloadMeta.hasPromptCacheKey,
				)
			} else {
				logOpenAIWSModeInfo(
					"ingress_ws_error_event account_id=%d turn=%d conn_id=%s idx=%d fallback_reason=%s err_code=%s err_type=%s err_message=%s previous_response_id=%s previous_response_id_kind=%s response_id=%s store_disabled=%v has_prompt_cache_key=%v",
					account.ID,
					turn,
					truncateOpenAIWSLogValue(lease.ConnID(), openAIWSIDValueMaxLen),
					eventCount,
					truncateOpenAIWSLogValue(fallbackReason, openAIWSLogValueMaxLen),
					errCode,
					errType,
					errMessage,
					truncateOpenAIWSLogValue(turnPreviousResponseID, openAIWSIDValueMaxLen),
					normalizeOpenAIWSLogValue(turnPreviousResponseIDKind),
					truncateOpenAIWSLogValue(responseID, openAIWSIDValueMaxLen),
					turnStoreDisabled,
					payloadMeta.hasPromptCacheKey,
				)
			}
			if recoverablePrevNotFound {
				lease.MarkBroken()
				errMsg := strings.TrimSpace(errMsgRaw)
				if errMsg == "" {
					errMsg = "previous response not found"
				}
				return nil, wrapOpenAIWSIngressTurnError(
					openAIWSIngressStagePreviousResponseNotFound,
					errors.New(errMsg),
					false,
				)
			}
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
			ms := int(time.Since(turnStart).Milliseconds())
			firstTokenMs = &ms
		}
		if openAIWSEventShouldParseUsage(eventType) {
			parseOpenAIWSResponseUsageFromCompletedEvent(upstreamMessage, &usage)
		}

		if !clientDisconnected {
			if needModelReplace && len(mappedModelBytes) > 0 && openAIWSEventMayContainModel(eventType) && bytes.Contains(upstreamMessage, mappedModelBytes) {
				upstreamMessage = replaceOpenAIWSMessageModel(upstreamMessage, mappedModel, originalModel)
			}
			if openAIWSEventMayContainToolCalls(eventType) && openAIWSMessageLikelyContainsToolCalls(upstreamMessage) {
				if corrected, changed := s.toolCorrector.CorrectToolCallsInSSEBytes(upstreamMessage); changed {
					upstreamMessage = corrected
				}
			}
			if err := s.writeOpenAIWSIngressClientMessage(ctx, clientConn, upstreamMessage); err != nil {
				if isOpenAIWSClientDisconnectError(err) {
					clientDisconnected = true
					closeStatus, closeReason := summarizeOpenAIWSReadCloseError(err)
					logOpenAIWSModeInfo(
						"ingress_ws_client_disconnected_drain account_id=%d turn=%d conn_id=%s close_status=%s close_reason=%s",
						account.ID,
						turn,
						truncateOpenAIWSLogValue(lease.ConnID(), openAIWSIDValueMaxLen),
						closeStatus,
						truncateOpenAIWSLogValue(closeReason, openAIWSHeaderValueMaxLen),
					)
				} else {
					return nil, wrapOpenAIWSIngressTurnError(
						"write_client",
						fmt.Errorf("write client websocket event: %w", err),
						wroteDownstream,
					)
				}
			} else {
				wroteDownstream = true
			}
		}

		if isTerminalEvent {
			if clientDisconnected {
				lease.MarkBroken()
			}
			firstTokenMsValue := -1
			if firstTokenMs != nil {
				firstTokenMsValue = *firstTokenMs
			}
			if debugEnabled {
				logOpenAIWSModeDebug(
					"ingress_ws_turn_completed account_id=%d turn=%d conn_id=%s response_id=%s duration_ms=%d events=%d token_events=%d terminal_events=%d first_event=%s last_event=%s first_token_ms=%d client_disconnected=%v",
					account.ID,
					turn,
					truncateOpenAIWSLogValue(lease.ConnID(), openAIWSIDValueMaxLen),
					truncateOpenAIWSLogValue(responseID, openAIWSIDValueMaxLen),
					time.Since(turnStart).Milliseconds(),
					eventCount,
					tokenEventCount,
					terminalEventCount,
					truncateOpenAIWSLogValue(firstEventType, openAIWSLogValueMaxLen),
					truncateOpenAIWSLogValue(lastEventType, openAIWSLogValueMaxLen),
					firstTokenMsValue,
					clientDisconnected,
				)
			}
			return &OpenAIForwardResult{
				RequestID:       responseID,
				Usage:           usage,
				Model:           originalModel,
				UpstreamModel:   mappedModel,
				ServiceTier:     extractOpenAIServiceTierFromBody(payload),
				ReasoningEffort: extractOpenAIReasoningEffortFromBody(payload, originalModel),
				Stream:          reqStream,
				OpenAIWSMode:    true,
				ResponseHeaders: lease.HandshakeHeaders(),
				Duration:        time.Since(turnStart),
				FirstTokenMs:    firstTokenMs,
			}, nil
		}
	}
}
