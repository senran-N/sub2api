package service

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (s *OpenAIGatewayService) isOpenAIWSGeneratePrewarmEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.Gateway.OpenAIWS.PrewarmGenerateEnabled
}

// performOpenAIWSGeneratePrewarm 在 WSv2 下执行可选的 generate=false 预热。
// 预热默认关闭，仅在配置开启后生效；失败时按可恢复错误回退到 HTTP。
func (s *OpenAIGatewayService) performOpenAIWSGeneratePrewarm(
	ctx context.Context,
	lease *openAIWSConnLease,
	decision OpenAIWSProtocolDecision,
	payload map[string]any,
	previousResponseID string,
	reqBody map[string]any,
	account *Account,
	stateStore OpenAIWSStateStore,
	groupID int64,
) error {
	if s == nil {
		return nil
	}
	if lease == nil || account == nil {
		logOpenAIWSModeInfo("prewarm_skip reason=invalid_state has_lease=%v has_account=%v", lease != nil, account != nil)
		return nil
	}
	connID := strings.TrimSpace(lease.ConnID())
	if !s.isOpenAIWSGeneratePrewarmEnabled() {
		return nil
	}
	if decision.Transport != OpenAIUpstreamTransportResponsesWebsocketV2 {
		logOpenAIWSModeInfo(
			"prewarm_skip account_id=%d conn_id=%s reason=transport_not_v2 transport=%s",
			account.ID,
			connID,
			normalizeOpenAIWSLogValue(string(decision.Transport)),
		)
		return nil
	}
	if strings.TrimSpace(previousResponseID) != "" {
		logOpenAIWSModeInfo(
			"prewarm_skip account_id=%d conn_id=%s reason=has_previous_response_id previous_response_id=%s",
			account.ID,
			connID,
			truncateOpenAIWSLogValue(previousResponseID, openAIWSIDValueMaxLen),
		)
		return nil
	}
	if lease.IsPrewarmed() {
		logOpenAIWSModeInfo("prewarm_skip account_id=%d conn_id=%s reason=already_prewarmed", account.ID, connID)
		return nil
	}
	if NeedsToolContinuation(reqBody) {
		logOpenAIWSModeInfo("prewarm_skip account_id=%d conn_id=%s reason=tool_continuation", account.ID, connID)
		return nil
	}
	prewarmStart := time.Now()
	logOpenAIWSModeInfo("prewarm_start account_id=%d conn_id=%s", account.ID, connID)

	prewarmPayload := make(map[string]any, len(payload)+1)
	for k, v := range payload {
		prewarmPayload[k] = v
	}
	prewarmPayload["generate"] = false
	prewarmPayloadJSON := payloadAsJSONBytes(prewarmPayload)

	if err := lease.WriteJSONWithContextTimeout(ctx, prewarmPayload, s.openAIWSWriteTimeout()); err != nil {
		lease.MarkBroken()
		logOpenAIWSModeInfo(
			"prewarm_write_fail account_id=%d conn_id=%s cause=%s",
			account.ID,
			connID,
			truncateOpenAIWSLogValue(err.Error(), openAIWSLogValueMaxLen),
		)
		return wrapOpenAIWSFallback("prewarm_write", err)
	}
	logOpenAIWSModeInfo("prewarm_write_sent account_id=%d conn_id=%s payload_bytes=%d", account.ID, connID, len(prewarmPayloadJSON))

	prewarmResponseID := ""
	prewarmEventCount := 0
	prewarmTerminalCount := 0
	for {
		message, readErr := lease.ReadMessageWithContextTimeout(ctx, s.openAIWSReadTimeout())
		if readErr != nil {
			lease.MarkBroken()
			closeStatus, closeReason := summarizeOpenAIWSReadCloseError(readErr)
			logOpenAIWSModeInfo(
				"prewarm_read_fail account_id=%d conn_id=%s close_status=%s close_reason=%s cause=%s events=%d",
				account.ID,
				connID,
				closeStatus,
				closeReason,
				truncateOpenAIWSLogValue(readErr.Error(), openAIWSLogValueMaxLen),
				prewarmEventCount,
			)
			return wrapOpenAIWSFallback("prewarm_"+classifyOpenAIWSReadFallbackReason(readErr), readErr)
		}

		eventType, eventResponseID, _ := parseOpenAIWSEventEnvelope(message)
		if eventType == "" {
			continue
		}
		prewarmEventCount++
		if prewarmResponseID == "" && eventResponseID != "" {
			prewarmResponseID = eventResponseID
		}
		if prewarmEventCount <= openAIWSPrewarmEventLogHead || eventType == "error" || isOpenAIWSTerminalEvent(eventType) {
			logOpenAIWSModeInfo(
				"prewarm_event account_id=%d conn_id=%s idx=%d type=%s bytes=%d",
				account.ID,
				connID,
				prewarmEventCount,
				truncateOpenAIWSLogValue(eventType, openAIWSLogValueMaxLen),
				len(message),
			)
		}

		if eventType == "error" {
			errCodeRaw, errTypeRaw, errMsgRaw := parseOpenAIWSErrorEventFields(message)
			s.persistOpenAIWSRateLimitSignal(ctx, account, lease.HandshakeHeaders(), message, errCodeRaw, errTypeRaw, errMsgRaw)
			errMsg := strings.TrimSpace(errMsgRaw)
			if errMsg == "" {
				errMsg = "OpenAI websocket prewarm error"
			}
			fallbackReason, canFallback := classifyOpenAIWSErrorEventFromRaw(errCodeRaw, errTypeRaw, errMsgRaw)
			errCode, errType, errMessage := summarizeOpenAIWSErrorEventFieldsFromRaw(errCodeRaw, errTypeRaw, errMsgRaw)
			logOpenAIWSModeInfo(
				"prewarm_error_event account_id=%d conn_id=%s idx=%d fallback_reason=%s can_fallback=%v err_code=%s err_type=%s err_message=%s",
				account.ID,
				connID,
				prewarmEventCount,
				truncateOpenAIWSLogValue(fallbackReason, openAIWSLogValueMaxLen),
				canFallback,
				errCode,
				errType,
				errMessage,
			)
			lease.MarkBroken()
			if canFallback {
				return wrapOpenAIWSFallback("prewarm_"+fallbackReason, errors.New(errMsg))
			}
			return wrapOpenAIWSFallback("prewarm_error_event", errors.New(errMsg))
		}

		if isOpenAIWSTerminalEvent(eventType) {
			prewarmTerminalCount++
			break
		}
	}

	lease.MarkPrewarmed()
	if prewarmResponseID != "" && stateStore != nil {
		ttl := s.openAIWSResponseStickyTTL()
		logOpenAIWSBindResponseAccountWarn(groupID, account.ID, prewarmResponseID, stateStore.BindResponseAccount(ctx, groupID, prewarmResponseID, account.ID, ttl))
		stateStore.BindResponseConn(prewarmResponseID, lease.ConnID(), ttl)
	}
	logOpenAIWSModeInfo(
		"prewarm_done account_id=%d conn_id=%s response_id=%s events=%d terminal_events=%d duration_ms=%d",
		account.ID,
		connID,
		truncateOpenAIWSLogValue(prewarmResponseID, openAIWSIDValueMaxLen),
		prewarmEventCount,
		prewarmTerminalCount,
		time.Since(prewarmStart).Milliseconds(),
	)
	return nil
}
