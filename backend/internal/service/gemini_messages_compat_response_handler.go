package service

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/googleapi"
	"github.com/senran-N/sub2api/internal/pkg/logger"
	"github.com/senran-N/sub2api/internal/util/responseheaders"

	"github.com/gin-gonic/gin"
)

func (s *GeminiMessagesCompatService) handleNonStreamingResponse(c *gin.Context, resp *http.Response, originalModel string) (*ClaudeUsage, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, s.writeClaudeError(c, http.StatusBadGateway, "upstream_error", "Failed to read upstream response")
	}

	unwrappedBody, err := unwrapGeminiResponse(body)
	if err != nil {
		return nil, s.writeClaudeError(c, http.StatusBadGateway, "upstream_error", "Failed to parse upstream response")
	}

	var geminiResp map[string]any
	if err := json.Unmarshal(unwrappedBody, &geminiResp); err != nil {
		return nil, s.writeClaudeError(c, http.StatusBadGateway, "upstream_error", "Failed to parse upstream response")
	}

	claudeResp, usage := convertGeminiToClaudeMessage(geminiResp, originalModel, unwrappedBody)
	c.JSON(http.StatusOK, claudeResp)
	return usage, nil
}

func (s *GeminiMessagesCompatService) handleStreamingResponse(c *gin.Context, resp *http.Response, startTime time.Time, originalModel string) (*geminiStreamResult, error) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	messageID := "msg_" + randomHex(12)
	messageStart := map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id":            messageID,
			"type":          "message",
			"role":          "assistant",
			"model":         originalModel,
			"content":       []any{},
			"stop_reason":   nil,
			"stop_sequence": nil,
			"usage": map[string]any{
				"input_tokens":  0,
				"output_tokens": 0,
			},
		},
	}
	writeSSE(c.Writer, "message_start", messageStart)
	flusher.Flush()

	var firstTokenMs *int
	var usage ClaudeUsage
	finishReason := ""
	sawToolUse := false

	nextBlockIndex := 0
	openBlockIndex := -1
	openBlockType := ""
	seenText := ""
	openToolIndex := -1
	openToolID := ""
	openToolName := ""
	seenToolJSON := ""

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("stream read error: %w", err)
		}

		if !strings.HasPrefix(line, "data:") {
			if errors.Is(err, io.EOF) {
				break
			}
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			if errors.Is(err, io.EOF) {
				break
			}
			continue
		}

		unwrappedBytes, err := unwrapGeminiResponse([]byte(payload))
		if err != nil {
			continue
		}

		var geminiResp map[string]any
		if err := json.Unmarshal(unwrappedBytes, &geminiResp); err != nil {
			continue
		}

		if finish := extractGeminiFinishReason(geminiResp); finish != "" {
			finishReason = finish
		}

		parts := extractGeminiParts(geminiResp)
		for _, part := range parts {
			if text, ok := part["text"].(string); ok && text != "" {
				delta, newSeen := computeGeminiTextDelta(seenText, text)
				seenText = newSeen
				if delta == "" {
					continue
				}

				if openBlockType != "text" {
					if openBlockIndex >= 0 {
						writeSSE(c.Writer, "content_block_stop", map[string]any{
							"type":  "content_block_stop",
							"index": openBlockIndex,
						})
					}
					openBlockType = "text"
					openBlockIndex = nextBlockIndex
					nextBlockIndex++
					writeSSE(c.Writer, "content_block_start", map[string]any{
						"type":  "content_block_start",
						"index": openBlockIndex,
						"content_block": map[string]any{
							"type": "text",
							"text": "",
						},
					})
				}

				if firstTokenMs == nil {
					ms := int(time.Since(startTime).Milliseconds())
					firstTokenMs = &ms
				}
				writeSSE(c.Writer, "content_block_delta", map[string]any{
					"type":  "content_block_delta",
					"index": openBlockIndex,
					"delta": map[string]any{
						"type": "text_delta",
						"text": delta,
					},
				})
				flusher.Flush()
				continue
			}

			if functionCall, ok := part["functionCall"].(map[string]any); ok && functionCall != nil {
				name, _ := functionCall["name"].(string)
				args := functionCall["args"]
				if strings.TrimSpace(name) == "" {
					name = "tool"
				}

				if openBlockIndex >= 0 {
					writeSSE(c.Writer, "content_block_stop", map[string]any{
						"type":  "content_block_stop",
						"index": openBlockIndex,
					})
					openBlockIndex = -1
					openBlockType = ""
				}

				if openToolIndex >= 0 && openToolName != name {
					writeSSE(c.Writer, "content_block_stop", map[string]any{
						"type":  "content_block_stop",
						"index": openToolIndex,
					})
					openToolIndex = -1
					openToolName = ""
					seenToolJSON = ""
				}

				if openToolIndex < 0 {
					openToolID = "toolu_" + randomHex(8)
					openToolIndex = nextBlockIndex
					openToolName = name
					nextBlockIndex++
					sawToolUse = true

					writeSSE(c.Writer, "content_block_start", map[string]any{
						"type":  "content_block_start",
						"index": openToolIndex,
						"content_block": map[string]any{
							"type":  "tool_use",
							"id":    openToolID,
							"name":  name,
							"input": map[string]any{},
						},
					})
				}

				argsJSONText := "{}"
				switch value := args.(type) {
				case nil:
				case string:
					if strings.TrimSpace(value) != "" {
						argsJSONText = value
					}
				default:
					if encoded, err := json.Marshal(args); err == nil && len(encoded) > 0 {
						argsJSONText = string(encoded)
					}
				}

				delta, newSeen := computeGeminiTextDelta(seenToolJSON, argsJSONText)
				seenToolJSON = newSeen
				if delta != "" {
					writeSSE(c.Writer, "content_block_delta", map[string]any{
						"type":  "content_block_delta",
						"index": openToolIndex,
						"delta": map[string]any{
							"type":         "input_json_delta",
							"partial_json": delta,
						},
					})
				}
				flusher.Flush()
			}
		}

		if parsedUsage := extractGeminiUsage(unwrappedBytes); parsedUsage != nil {
			usage = *parsedUsage
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	if openBlockIndex >= 0 {
		writeSSE(c.Writer, "content_block_stop", map[string]any{
			"type":  "content_block_stop",
			"index": openBlockIndex,
		})
	}
	if openToolIndex >= 0 {
		writeSSE(c.Writer, "content_block_stop", map[string]any{
			"type":  "content_block_stop",
			"index": openToolIndex,
		})
	}

	stopReason := mapGeminiFinishReasonToClaudeStopReason(finishReason)
	if sawToolUse {
		stopReason = "tool_use"
	}

	usageObject := map[string]any{
		"output_tokens": usage.OutputTokens,
	}
	if usage.InputTokens > 0 {
		usageObject["input_tokens"] = usage.InputTokens
	}
	writeSSE(c.Writer, "message_delta", map[string]any{
		"type": "message_delta",
		"delta": map[string]any{
			"stop_reason":   stopReason,
			"stop_sequence": nil,
		},
		"usage": usageObject,
	})
	writeSSE(c.Writer, "message_stop", map[string]any{
		"type": "message_stop",
	})
	flusher.Flush()

	return &geminiStreamResult{usage: &usage, firstTokenMs: firstTokenMs}, nil
}

func writeSSE(w io.Writer, event string, data any) {
	if event != "" {
		_, _ = fmt.Fprintf(w, "event: %s\n", event)
	}
	encoded, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", string(encoded))
}

func randomHex(nBytes int) string {
	bytes := make([]byte, nBytes)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *GeminiMessagesCompatService) writeClaudeError(c *gin.Context, status int, errType, message string) error {
	c.JSON(status, gin.H{
		"type":  "error",
		"error": gin.H{"type": errType, "message": message},
	})
	return fmt.Errorf("%s", message)
}

func (s *GeminiMessagesCompatService) writeGoogleError(c *gin.Context, status int, message string) error {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    status,
			"message": message,
			"status":  googleapi.HTTPStatusToGoogleStatus(status),
		},
	})
	return fmt.Errorf("%s", message)
}

type geminiNativeStreamResult struct {
	usage        *ClaudeUsage
	firstTokenMs *int
}

func isGeminiInsufficientScope(headers http.Header, body []byte) bool {
	if strings.Contains(strings.ToLower(headers.Get("Www-Authenticate")), "insufficient_scope") {
		return true
	}
	lowerBody := strings.ToLower(string(body))
	return strings.Contains(lowerBody, "insufficient authentication scopes") || strings.Contains(lowerBody, "access_token_scope_insufficient")
}

type UpstreamHTTPResult struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (s *GeminiMessagesCompatService) handleNativeNonStreamingResponse(c *gin.Context, resp *http.Response, isOAuth bool) (*ClaudeUsage, error) {
	if s.cfg != nil && s.cfg.Gateway.GeminiDebugResponseHeaders {
		logger.LegacyPrintf("service.gemini_messages_compat", "[GeminiAPI] ========== Response Headers ==========")
		for key, values := range resp.Header {
			if strings.HasPrefix(strings.ToLower(key), "x-ratelimit") {
				logger.LegacyPrintf("service.gemini_messages_compat", "[GeminiAPI] %s: %v", key, values)
			}
		}
		logger.LegacyPrintf("service.gemini_messages_compat", "[GeminiAPI] ========================================")
	}

	maxBytes := resolveUpstreamResponseReadLimit(s.cfg)
	respBody, err := readUpstreamResponseBodyLimited(resp.Body, maxBytes)
	if err != nil {
		if errors.Is(err, ErrUpstreamResponseBodyTooLarge) {
			setOpsUpstreamError(c, http.StatusBadGateway, "upstream response too large", "")
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"type":    "upstream_error",
					"message": "Upstream response too large",
				},
			})
		}
		return nil, err
	}

	if isOAuth {
		unwrappedBody, unwrapErr := unwrapGeminiResponse(respBody)
		if unwrapErr == nil {
			respBody = unwrappedBody
		}
	}

	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(resp.StatusCode, contentType, respBody)

	if usage := extractGeminiUsage(respBody); usage != nil {
		return usage, nil
	}
	return &ClaudeUsage{}, nil
}

func (s *GeminiMessagesCompatService) handleNativeStreamingResponse(c *gin.Context, resp *http.Response, startTime time.Time, isOAuth bool) (*geminiNativeStreamResult, error) {
	if s.cfg != nil && s.cfg.Gateway.GeminiDebugResponseHeaders {
		logger.LegacyPrintf("service.gemini_messages_compat", "[GeminiAPI] ========== Streaming Response Headers ==========")
		for key, values := range resp.Header {
			if strings.HasPrefix(strings.ToLower(key), "x-ratelimit") {
				logger.LegacyPrintf("service.gemini_messages_compat", "[GeminiAPI] %s: %v", key, values)
			}
		}
		logger.LegacyPrintf("service.gemini_messages_compat", "[GeminiAPI] ====================================================")
	}

	if s.responseHeaderFilter != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
	}

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

	reader := bufio.NewReader(resp.Body)
	usage := &ClaudeUsage{}
	var firstTokenMs *int

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(trimmed, "data:") {
				payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if payload == "" || payload == "[DONE]" {
					_, _ = io.WriteString(c.Writer, line)
					flusher.Flush()
				} else {
					rawToWrite := payload
					var rawBytes []byte
					if isOAuth {
						innerBytes, unwrapErr := unwrapGeminiResponse([]byte(payload))
						if unwrapErr == nil {
							rawToWrite = string(innerBytes)
							rawBytes = innerBytes
						}
					} else {
						rawBytes = []byte(payload)
					}

					if parsedUsage := extractGeminiUsage(rawBytes); parsedUsage != nil {
						usage = parsedUsage
					}

					if firstTokenMs == nil {
						ms := int(time.Since(startTime).Milliseconds())
						firstTokenMs = &ms
					}

					if isOAuth {
						_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", rawToWrite)
					} else {
						_, _ = io.WriteString(c.Writer, line)
					}
					flusher.Flush()
				}
			} else {
				_, _ = io.WriteString(c.Writer, line)
				flusher.Flush()
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return &geminiNativeStreamResult{usage: usage, firstTokenMs: firstTokenMs}, nil
}
