package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/antigravity"
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

// handleClaudeStreamToNonStreaming 收集上游流式响应，转换为 Claude 非流式格式返回。
// 用于处理客户端非流式请求但上游只支持流式的情况。
func (s *AntigravityGatewayService) handleClaudeStreamToNonStreaming(c *gin.Context, resp *http.Response, startTime time.Time, originalModel string) (*antigravityStreamResult, error) {
	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.settingService.cfg.Gateway.MaxLineSize
	}
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], maxLineSize)

	var firstTokenMs *int
	var last map[string]any
	var lastWithParts map[string]any
	var collectedParts []map[string]any

	type scanEvent struct {
		line string
		err  error
	}

	events := make(chan scanEvent, 16)
	done := make(chan struct{})
	sendEvent := func(ev scanEvent) bool {
		select {
		case events <- ev:
			return true
		case <-done:
			return false
		}
	}

	var lastReadAt int64
	atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
	go func(scanBuf *sseScannerBuf64K) {
		defer putSSEScannerBuf64K(scanBuf)
		defer close(events)
		for scanner.Scan() {
			atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
			if !sendEvent(scanEvent{line: scanner.Text()}) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			_ = sendEvent(scanEvent{err: err})
		}
	}(scanBuf)
	defer close(done)

	streamInterval := time.Duration(0)
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.StreamDataIntervalTimeout > 0 {
		streamInterval = time.Duration(s.settingService.cfg.Gateway.StreamDataIntervalTimeout) * time.Second
	}
	var intervalTicker *time.Ticker
	if streamInterval > 0 {
		intervalTicker = time.NewTicker(streamInterval)
		defer intervalTicker.Stop()
	}
	var intervalCh <-chan time.Time
	if intervalTicker != nil {
		intervalCh = intervalTicker.C
	}

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				goto returnResponse
			}
			if ev.err != nil {
				if errors.Is(ev.err, bufio.ErrTooLong) {
					logger.LegacyPrintf("service.antigravity_gateway", "SSE line too long (antigravity claude non-stream): max_size=%d error=%v", maxLineSize, ev.err)
				}
				return nil, ev.err
			}

			line := ev.line
			trimmed := strings.TrimRight(line, "\r\n")
			if !strings.HasPrefix(trimmed, "data:") {
				continue
			}

			payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			if payload == "" || payload == "[DONE]" {
				continue
			}

			inner, parseErr := s.unwrapV1InternalResponse([]byte(payload))
			if parseErr != nil {
				continue
			}

			var parsed map[string]any
			if err := json.Unmarshal(inner, &parsed); err != nil {
				continue
			}

			if firstTokenMs == nil {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}

			last = parsed
			if parts := extractGeminiParts(parsed); len(parts) > 0 {
				lastWithParts = parsed
				collectedParts = append(collectedParts, parts...)
			}

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			logger.LegacyPrintf("service.antigravity_gateway", "Stream data interval timeout (antigravity claude non-stream)")
			return nil, fmt.Errorf("stream data interval timeout")
		}
	}

returnResponse:
	finalResponse := pickGeminiCollectResult(last, lastWithParts)
	if last == nil && lastWithParts == nil {
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] warning: empty stream response (claude non-stream), triggering failover")
		return nil, &UpstreamFailoverError{
			StatusCode:             http.StatusBadGateway,
			ResponseBody:           []byte(`{"error":"empty stream response from upstream"}`),
			RetryableOnSameAccount: true,
		}
	}

	if len(collectedParts) > 0 {
		finalResponse = mergeCollectedPartsToResponse(finalResponse, collectedParts)
	}

	geminiBody, err := json.Marshal(finalResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini response: %w", err)
	}

	claudeResp, agUsage, err := antigravity.TransformGeminiToClaude(geminiBody, originalModel)
	if err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] transform_error error=%v body=%s", err, string(geminiBody))
		return nil, s.writeClaudeError(c, http.StatusBadGateway, "upstream_error", "Failed to parse upstream response")
	}

	c.Data(http.StatusOK, "application/json", claudeResp)
	return &antigravityStreamResult{usage: convertAntigravityClaudeUsage(agUsage), firstTokenMs: firstTokenMs}, nil
}

// handleClaudeStreamingResponse 处理 Claude 流式响应（Gemini SSE → Claude SSE 转换）。
func (s *AntigravityGatewayService) handleClaudeStreamingResponse(c *gin.Context, resp *http.Response, startTime time.Time, originalModel string) (*antigravityStreamResult, error) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	processor := antigravity.NewStreamingProcessor(originalModel)
	var firstTokenMs *int

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.settingService.cfg.Gateway.MaxLineSize
	}
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], maxLineSize)

	type scanEvent struct {
		line string
		err  error
	}

	events := make(chan scanEvent, 16)
	done := make(chan struct{})
	sendEvent := func(ev scanEvent) bool {
		select {
		case events <- ev:
			return true
		case <-done:
			return false
		}
	}

	var lastReadAt int64
	atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
	go func(scanBuf *sseScannerBuf64K) {
		defer putSSEScannerBuf64K(scanBuf)
		defer close(events)
		for scanner.Scan() {
			atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
			if !sendEvent(scanEvent{line: scanner.Text()}) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			_ = sendEvent(scanEvent{err: err})
		}
	}(scanBuf)
	defer close(done)

	streamInterval := time.Duration(0)
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.StreamDataIntervalTimeout > 0 {
		streamInterval = time.Duration(s.settingService.cfg.Gateway.StreamDataIntervalTimeout) * time.Second
	}
	var intervalTicker *time.Ticker
	if streamInterval > 0 {
		intervalTicker = time.NewTicker(streamInterval)
		defer intervalTicker.Stop()
	}
	var intervalCh <-chan time.Time
	if intervalTicker != nil {
		intervalCh = intervalTicker.C
	}

	keepaliveInterval := time.Duration(0)
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.StreamKeepaliveInterval > 0 {
		keepaliveInterval = time.Duration(s.settingService.cfg.Gateway.StreamKeepaliveInterval) * time.Second
	}
	var keepaliveTicker *time.Ticker
	if keepaliveInterval > 0 {
		keepaliveTicker = time.NewTicker(keepaliveInterval)
		defer keepaliveTicker.Stop()
	}
	var keepaliveCh <-chan time.Time
	if keepaliveTicker != nil {
		keepaliveCh = keepaliveTicker.C
	}
	lastDataAt := time.Now()

	cw := newAntigravityClientWriter(c.Writer, flusher, "antigravity claude")
	errorEventSent := false
	sendErrorEvent := func(reason string) {
		if errorEventSent || cw.Disconnected() {
			return
		}
		errorEventSent = true
		_, _ = fmt.Fprintf(c.Writer, "event: error\ndata: {\"error\":\"%s\"}\n\n", reason)
		flusher.Flush()
	}

	finishUsage := func() *ClaudeUsage {
		_, agUsage := processor.Finish()
		return convertAntigravityClaudeUsage(agUsage)
	}

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				finalEvents, agUsage := processor.Finish()
				if len(finalEvents) > 0 {
					cw.Write(finalEvents)
				} else if !processor.MessageStartSent() && !cw.Disconnected() {
					logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Claude-Stream] empty stream response (no valid events parsed), triggering failover")
					return nil, &UpstreamFailoverError{
						StatusCode:             http.StatusBadGateway,
						ResponseBody:           []byte(`{"error":"empty stream response from upstream"}`),
						RetryableOnSameAccount: true,
					}
				}
				return &antigravityStreamResult{usage: convertAntigravityClaudeUsage(agUsage), firstTokenMs: firstTokenMs, clientDisconnect: cw.Disconnected()}, nil
			}
			if ev.err != nil {
				if disconnect, handled := handleStreamReadError(ev.err, cw.Disconnected(), "antigravity claude"); handled {
					return &antigravityStreamResult{usage: finishUsage(), firstTokenMs: firstTokenMs, clientDisconnect: disconnect}, nil
				}
				if errors.Is(ev.err, bufio.ErrTooLong) {
					logger.LegacyPrintf("service.antigravity_gateway", "SSE line too long (antigravity): max_size=%d error=%v", maxLineSize, ev.err)
					sendErrorEvent("response_too_large")
					return &antigravityStreamResult{usage: convertAntigravityClaudeUsage(nil), firstTokenMs: firstTokenMs}, ev.err
				}
				sendErrorEvent("stream_read_error")
				return nil, fmt.Errorf("stream read error: %w", ev.err)
			}

			lastDataAt = time.Now()
			claudeEvents := processor.ProcessLine(strings.TrimRight(ev.line, "\r\n"))
			if len(claudeEvents) > 0 {
				if firstTokenMs == nil {
					ms := int(time.Since(startTime).Milliseconds())
					firstTokenMs = &ms
				}
				cw.Write(claudeEvents)
			}

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			if cw.Disconnected() {
				logger.LegacyPrintf("service.antigravity_gateway", "Upstream timeout after client disconnect (antigravity claude), returning collected usage")
				return &antigravityStreamResult{usage: finishUsage(), firstTokenMs: firstTokenMs, clientDisconnect: true}, nil
			}
			logger.LegacyPrintf("service.antigravity_gateway", "Stream data interval timeout (antigravity)")
			sendErrorEvent("stream_timeout")
			return &antigravityStreamResult{usage: convertAntigravityClaudeUsage(nil), firstTokenMs: firstTokenMs}, fmt.Errorf("stream data interval timeout")

		case <-keepaliveCh:
			if cw.Disconnected() {
				continue
			}
			if time.Since(lastDataAt) < keepaliveInterval {
				continue
			}
			if !cw.Fprintf("event: ping\ndata: {\"type\": \"ping\"}\n\n") {
				logger.LegacyPrintf("service.antigravity_gateway", "Client disconnected during keepalive ping (antigravity claude), continuing to drain upstream for billing")
			}
		}
	}
}
