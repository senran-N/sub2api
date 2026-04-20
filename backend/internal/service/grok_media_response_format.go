package service

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	grokOpenAIImageResponseFormatURL     = "url"
	grokOpenAIImageResponseFormatB64JSON = "b64_json"
)

func normalizeGrokOpenAIImageResponseFormat(raw string) (string, error) {
	format := strings.ToLower(strings.TrimSpace(raw))
	switch format {
	case "":
		return "", nil
	case grokOpenAIImageResponseFormatURL, grokOpenAIImageResponseFormatB64JSON:
		return format, nil
	default:
		return "", errors.New("response_format must be one of ['url', 'b64_json']")
	}
}

func resolveGrokImageResponseFormatRequest(c *gin.Context, body []byte) (string, error) {
	if len(body) == 0 {
		return "", nil
	}

	contentType := ""
	if c != nil && c.Request != nil {
		contentType = c.Request.Header.Get("Content-Type")
	}
	mediaType, params, _ := mime.ParseMediaType(contentType)
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))

	if isJSONLikeContentType(mediaType) || gjson.ValidBytes(body) {
		return normalizeGrokOpenAIImageResponseFormat(gjson.GetBytes(body, "response_format").String())
	}

	if mediaType != "multipart/form-data" {
		return "", nil
	}

	boundary := strings.TrimSpace(params["boundary"])
	if boundary == "" {
		return "", nil
	}
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", nil
			}
			return "", errors.New("failed to parse image request body")
		}

		if strings.TrimSpace(part.FormName()) != "response_format" {
			_ = part.Close()
			continue
		}

		value, readErr := readMultipartTextField(part, 64)
		_ = part.Close()
		if readErr != nil {
			return "", errors.New("failed to parse image request body")
		}
		return normalizeGrokOpenAIImageResponseFormat(value)
	}
}
