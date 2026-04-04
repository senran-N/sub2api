package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func buildSoraNonStreamResponse(content, model string) map[string]any {
	return map[string]any{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []any{
			map[string]any{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
	}
}

func buildSoraPromptForwardResult(startTime time.Time, originalModel, upstreamModel string, stream bool, firstTokenMs *int) *ForwardResult {
	return &ForwardResult{
		RequestID:     "",
		Model:         originalModel,
		UpstreamModel: upstreamModel,
		Stream:        stream,
		Duration:      time.Since(startTime),
		FirstTokenMs:  firstTokenMs,
		Usage:         ClaudeUsage{},
		MediaType:     "prompt",
	}
}

func soraCharacterResponseFields(result *soraCharacterFlowResult) map[string]any {
	if result == nil {
		return nil
	}
	return map[string]any{
		"character_id":           result.CharacterID,
		"cameo_id":               result.CameoID,
		"character_username":     result.Username,
		"character_display_name": result.DisplayName,
	}
}

func soraMediaResponseFields(urls []string) map[string]any {
	if len(urls) == 0 {
		return nil
	}
	fields := map[string]any{
		"media_url": urls[0],
	}
	if len(urls) > 1 {
		fields["media_urls"] = urls
	}
	return fields
}

func soraImageSizeFromModel(model string) string {
	modelLower := strings.ToLower(model)
	if size, ok := soraImageSizeMap[modelLower]; ok {
		return size
	}
	if strings.Contains(modelLower, "landscape") || strings.Contains(modelLower, "portrait") {
		return "540"
	}
	return "360"
}

func firstMediaURL(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

func (s *SoraGatewayService) buildSoraMediaURL(path string, rawQuery string) string {
	if path == "" {
		return path
	}
	prefix := "/sora/media"
	values := url.Values{}
	if rawQuery != "" {
		if parsed, err := url.ParseQuery(rawQuery); err == nil {
			values = parsed
		}
	}

	signKey := ""
	ttlSeconds := 0
	if s != nil && s.cfg != nil {
		signKey = strings.TrimSpace(s.cfg.Gateway.SoraMediaSigningKey)
		ttlSeconds = s.cfg.Gateway.SoraMediaSignedURLTTLSeconds
	}
	values.Del("sig")
	values.Del("expires")
	signingQuery := values.Encode()
	if signKey != "" && ttlSeconds > 0 {
		expires := time.Now().Add(time.Duration(ttlSeconds) * time.Second).Unix()
		signature := SignSoraMediaURL(path, signingQuery, expires, signKey)
		if signature != "" {
			values.Set("expires", strconv.FormatInt(expires, 10))
			values.Set("sig", signature)
			prefix = "/sora/media-signed"
		}
	}

	encoded := values.Encode()
	if encoded == "" {
		return prefix + path
	}
	return prefix + path + "?" + encoded
}

func (s *SoraGatewayService) normalizeSoraMediaURLs(urls []string) []string {
	if len(urls) == 0 {
		return urls
	}
	output := make([]string, 0, len(urls))
	for _, raw := range urls {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
			output = append(output, raw)
			continue
		}
		pathVal := raw
		if !strings.HasPrefix(pathVal, "/") {
			pathVal = "/" + pathVal
		}
		output = append(output, s.buildSoraMediaURL(pathVal, ""))
	}
	return output
}

func jsonMarshalRaw(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return b, nil
}

func (s *SoraGatewayService) writeSoraCompletionResponse(c *gin.Context, model, content string, startTime time.Time, stream bool, fields map[string]any) (*int, error) {
	if stream {
		return s.writeSoraStream(c, model, content, startTime)
	}
	if c == nil {
		return nil, nil
	}
	response := buildSoraNonStreamResponse(content, model)
	for key, value := range fields {
		response[key] = value
	}
	c.JSON(http.StatusOK, response)
	return nil, nil
}

func buildSoraContent(mediaType string, urls []string) string {
	switch mediaType {
	case "image":
		parts := make([]string, 0, len(urls))
		for _, item := range urls {
			parts = append(parts, fmt.Sprintf("![image](%s)", item))
		}
		return strings.Join(parts, "\n")
	case "video":
		if len(urls) == 0 {
			return ""
		}
		return fmt.Sprintf("```html\n<video src='%s' controls></video>\n```", urls[0])
	default:
		return ""
	}
}
