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
	"github.com/senran-N/sub2api/internal/pkg/logger"
)

func (s *AntigravityGatewayService) handleGeminiStreamingResponse(c *gin.Context, resp *http.Response, startTime time.Time) (*antigravityStreamResult, error) {
	c.Status(resp.StatusCode)
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/event-stream; charset=utf-8"
	}
	c.Header("Content-Type", contentType)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.settingService.cfg.Gateway.MaxLineSize
	}
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], maxLineSize)
	usage := &ClaudeUsage{}
	var firstTokenMs *int

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

	cw := newAntigravityClientWriter(c.Writer, flusher, "antigravity gemini")
	errorEventSent := false
	sendErrorEvent := func(reason string) {
		if errorEventSent || cw.Disconnected() {
			return
		}
		errorEventSent = true
		_, _ = fmt.Fprintf(c.Writer, "event: error\ndata: {\"error\":\"%s\"}\n\n", reason)
		flusher.Flush()
	}

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: cw.Disconnected()}, nil
			}
			if ev.err != nil {
				if disconnect, handled := handleStreamReadError(ev.err, cw.Disconnected(), "antigravity gemini"); handled {
					return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: disconnect}, nil
				}
				if errors.Is(ev.err, bufio.ErrTooLong) {
					logger.LegacyPrintf("service.antigravity_gateway", "SSE line too long (antigravity): max_size=%d error=%v", maxLineSize, ev.err)
					sendErrorEvent("response_too_large")
					return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs}, ev.err
				}
				sendErrorEvent("stream_read_error")
				return nil, ev.err
			}

			lastDataAt = time.Now()

			line := ev.line
			trimmed := strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(trimmed, "data:") {
				payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if payload == "" || payload == "[DONE]" {
					cw.Fprintf("%s\n", line)
					continue
				}

				inner, parseErr := s.unwrapV1InternalResponse([]byte(payload))
				if parseErr == nil && inner != nil {
					payload = string(inner)
				}

				if u := extractGeminiUsage(inner); u != nil {
					usage = u
				}

				var parsed map[string]any
				if json.Unmarshal(inner, &parsed) == nil {
					if candidates, ok := parsed["candidates"].([]any); ok && len(candidates) > 0 {
						if cand, ok := candidates[0].(map[string]any); ok {
							if fr, ok := cand["finishReason"].(string); ok && fr == "MALFORMED_FUNCTION_CALL" {
								logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] MALFORMED_FUNCTION_CALL detected in forward stream")
								if content, ok := cand["content"]; ok {
									if body, err := json.Marshal(content); err == nil {
										logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Malformed content: %s", string(body))
									}
								}
							}
						}
					}
				}

				if firstTokenMs == nil {
					ms := int(time.Since(startTime).Milliseconds())
					firstTokenMs = &ms
				}

				cw.Fprintf("data: %s\n\n", payload)
				continue
			}

			cw.Fprintf("%s\n", line)

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			if cw.Disconnected() {
				logger.LegacyPrintf("service.antigravity_gateway", "Upstream timeout after client disconnect (antigravity gemini), returning collected usage")
				return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: true}, nil
			}
			logger.LegacyPrintf("service.antigravity_gateway", "Stream data interval timeout (antigravity)")
			sendErrorEvent("stream_timeout")
			return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream data interval timeout")

		case <-keepaliveCh:
			if cw.Disconnected() {
				continue
			}
			if time.Since(lastDataAt) < keepaliveInterval {
				continue
			}
			if !cw.Fprintf(":\n\n") {
				logger.LegacyPrintf("service.antigravity_gateway", "Client disconnected during keepalive ping (antigravity gemini), continuing to drain upstream for billing")
			}
		}
	}
}

// handleGeminiStreamToNonStreaming 读取上游流式响应，合并为非流式响应返回给客户端。
// Gemini 流式响应是增量的，需要累积所有 chunk 的内容。
func (s *AntigravityGatewayService) handleGeminiStreamToNonStreaming(c *gin.Context, resp *http.Response, startTime time.Time) (*antigravityStreamResult, error) {
	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.settingService.cfg.Gateway.MaxLineSize
	}
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], maxLineSize)

	usage := &ClaudeUsage{}
	var firstTokenMs *int
	var last map[string]any
	var lastWithParts map[string]any
	var collectedImageParts []map[string]any
	var collectedTextParts []string

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
					logger.LegacyPrintf("service.antigravity_gateway", "SSE line too long (antigravity non-stream): max_size=%d error=%v", maxLineSize, ev.err)
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

			if u := extractGeminiUsage(inner); u != nil {
				usage = u
			}

			if candidates, ok := parsed["candidates"].([]any); ok && len(candidates) > 0 {
				if cand, ok := candidates[0].(map[string]any); ok {
					if fr, ok := cand["finishReason"].(string); ok && fr == "MALFORMED_FUNCTION_CALL" {
						logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] MALFORMED_FUNCTION_CALL detected in forward non-stream collect")
						if content, ok := cand["content"]; ok {
							if body, err := json.Marshal(content); err == nil {
								logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Malformed content: %s", string(body))
							}
						}
					}
				}
			}

			if parts := extractGeminiParts(parsed); len(parts) > 0 {
				lastWithParts = parsed
				for _, part := range parts {
					if inlineData, ok := part["inlineData"].(map[string]any); ok {
						collectedImageParts = append(collectedImageParts, part)
						_ = inlineData
					}
					if text, ok := part["text"].(string); ok && text != "" {
						collectedTextParts = append(collectedTextParts, text)
					}
				}
			}

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			logger.LegacyPrintf("service.antigravity_gateway", "Stream data interval timeout (antigravity non-stream)")
			return nil, fmt.Errorf("stream data interval timeout")
		}
	}

returnResponse:
	finalResponse := pickGeminiCollectResult(last, lastWithParts)
	if last == nil && lastWithParts == nil {
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] warning: empty stream response (gemini non-stream), triggering failover")
		return nil, &UpstreamFailoverError{
			StatusCode:             http.StatusBadGateway,
			ResponseBody:           []byte(`{"error":"empty stream response from upstream"}`),
			RetryableOnSameAccount: true,
		}
	}

	if len(collectedImageParts) > 0 {
		finalResponse = mergeImagePartsToResponse(finalResponse, collectedImageParts)
	}
	if len(collectedTextParts) > 0 {
		finalResponse = mergeTextPartsToResponse(finalResponse, collectedTextParts)
	}

	respBody, err := json.Marshal(finalResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	c.Data(http.StatusOK, "application/json", respBody)

	return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs}, nil
}
