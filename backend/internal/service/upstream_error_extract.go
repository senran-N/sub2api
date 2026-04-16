package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/tidwall/gjson"
)

type upstreamErrorInfo struct {
	Message  string
	Type     string
	Status   string
	Code     any
	HasCode  bool
	Param    any
	HasParam bool
}

func extractUpstreamErrorInfo(body []byte) upstreamErrorInfo {
	root := gjson.ParseBytes(body)
	info := upstreamErrorInfo{
		Message: firstNonEmptyJSONString(root,
			"error.message",
			"detail.message",
			"detail",
			"message",
		),
		Type: firstNonEmptyJSONString(root,
			"error.type",
			"type",
		),
		Status: firstNonEmptyJSONString(root,
			"error.status",
			"status",
		),
	}

	if code, ok := firstExistingJSONValue(root,
		"error.code",
		"detail.code",
		"code",
	); ok {
		info.Code = code
		info.HasCode = true
	}
	if param, ok := firstExistingJSONValue(root,
		"error.param",
		"param",
	); ok {
		info.Param = param
		info.HasParam = true
	}

	nested := firstEmbeddedJSON(root,
		"error.message",
		"detail",
		"message",
	)
	if !nested.Exists() {
		return info
	}

	if nestedMessage := firstNonEmptyJSONString(nested,
		"error.message",
		"detail.message",
		"detail",
		"message",
	); nestedMessage != "" {
		info.Message = nestedMessage
	}
	if info.Type == "" {
		info.Type = firstNonEmptyJSONString(nested,
			"error.type",
			"type",
		)
	}
	if info.Status == "" {
		info.Status = firstNonEmptyJSONString(nested,
			"error.status",
			"status",
		)
	}
	if !info.HasCode {
		if code, ok := firstExistingJSONValue(nested,
			"error.code",
			"detail.code",
			"code",
		); ok {
			info.Code = code
			info.HasCode = true
		}
	}
	if !info.HasParam {
		if param, ok := firstExistingJSONValue(nested,
			"error.param",
			"param",
		); ok {
			info.Param = param
			info.HasParam = true
		}
	}

	return info
}

func firstEmbeddedJSON(root gjson.Result, paths ...string) gjson.Result {
	for _, path := range paths {
		raw := strings.TrimSpace(root.Get(path).String())
		if raw == "" {
			continue
		}
		if nested := parseEmbeddedJSONObject(raw); nested.Exists() {
			return nested
		}
	}
	return gjson.Result{}
}

func parseEmbeddedJSONObject(raw string) gjson.Result {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return gjson.Result{}
	}

	start := strings.IndexByte(raw, '{')
	end := strings.LastIndexByte(raw, '}')
	if start < 0 || end < start {
		return gjson.Result{}
	}

	candidate := raw[start : end+1]
	if !gjson.Valid(candidate) {
		return gjson.Result{}
	}

	return gjson.Parse(candidate)
}

func firstNonEmptyJSONString(root gjson.Result, paths ...string) string {
	for _, path := range paths {
		value := strings.TrimSpace(root.Get(path).String())
		if value != "" {
			return value
		}
	}
	return ""
}

func firstExistingJSONValue(root gjson.Result, paths ...string) (any, bool) {
	for _, path := range paths {
		result := root.Get(path)
		if !result.Exists() || result.Type == gjson.Null {
			continue
		}
		value := normalizeJSONValue(result.Value())
		if value == nil {
			continue
		}
		if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
			continue
		}
		return value, true
	}
	return nil, false
}

func normalizeJSONValue(value any) any {
	number, ok := value.(float64)
	if !ok {
		return value
	}
	if math.Trunc(number) == number {
		return int(number)
	}
	return number
}

func stringifyUpstreamErrorCode(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}
