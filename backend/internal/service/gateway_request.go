package service

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unsafe"

	"github.com/senran-N/sub2api/internal/domain"
	"github.com/tidwall/gjson"
)

var (
	patternTypeThinking         = []byte(`"type":"thinking"`)
	patternTypeThinkingSpaced   = []byte(`"type": "thinking"`)
	patternTypeRedactedThinking = []byte(`"type":"redacted_thinking"`)
	patternTypeRedactedSpaced   = []byte(`"type": "redacted_thinking"`)

	patternThinkingField       = []byte(`"thinking":`)
	patternThinkingFieldSpaced = []byte(`"thinking" :`)

	patternEmptyContent       = []byte(`"content":[]`)
	patternEmptyContentSpaced = []byte(`"content": []`)
	patternEmptyContentSp1    = []byte(`"content" : []`)
	patternEmptyContentSp2    = []byte(`"content" :[]`)

	patternEmptyText       = []byte(`"text":""`)
	patternEmptyTextSpaced = []byte(`"text": ""`)
	patternEmptyTextSp1    = []byte(`"text" : ""`)
	patternEmptyTextSp2    = []byte(`"text" :""`)

	sessionUserAgentProductPattern = regexp.MustCompile(`([A-Za-z0-9._-]+)/[A-Za-z0-9._-]+`)
	sessionUserAgentVersionPattern = regexp.MustCompile(`\bv?\d+(?:\.\d+){1,3}\b`)
)

type SessionContext struct {
	ClientIP  string
	UserAgent string
	APIKeyID  int64
}

type ParsedRequest struct {
	Body            []byte
	Model           string
	Stream          bool
	MetadataUserID  string
	System          any
	Messages        []any
	HasSystem       bool
	ThinkingEnabled bool
	OutputEffort    string
	MaxTokens       int
	SessionContext  *SessionContext

	OnUpstreamAccepted func()
}

func NormalizeSessionUserAgent(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	matches := sessionUserAgentProductPattern.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return normalizeSessionUserAgentFallback(raw)
	}

	products := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		product := strings.ToLower(strings.TrimSpace(match[1]))
		if product == "" {
			continue
		}
		if _, exists := seen[product]; exists {
			continue
		}
		seen[product] = struct{}{}
		products = append(products, product)
	}
	if len(products) == 0 {
		return normalizeSessionUserAgentFallback(raw)
	}
	sort.Strings(products)
	return strings.Join(products, "+")
}

func normalizeSessionUserAgentFallback(raw string) string {
	normalized := strings.ToLower(strings.Join(strings.Fields(raw), " "))
	normalized = sessionUserAgentVersionPattern.ReplaceAllString(normalized, "")
	return strings.Join(strings.Fields(normalized), " ")
}

func ParseGatewayRequest(body []byte, protocol string) (*ParsedRequest, error) {
	if !gjson.ValidBytes(body) {
		return nil, fmt.Errorf("invalid json")
	}

	jsonStr := *(*string)(unsafe.Pointer(&body))
	parsed := &ParsedRequest{Body: body}

	modelResult := gjson.Get(jsonStr, "model")
	if modelResult.Exists() {
		if modelResult.Type != gjson.String {
			return nil, fmt.Errorf("invalid model field type")
		}
		parsed.Model = modelResult.String()
	}

	streamResult := gjson.Get(jsonStr, "stream")
	if streamResult.Exists() {
		if streamResult.Type != gjson.True && streamResult.Type != gjson.False {
			return nil, fmt.Errorf("invalid stream field type")
		}
		parsed.Stream = streamResult.Bool()
	}

	parsed.MetadataUserID = gjson.Get(jsonStr, "metadata.user_id").String()

	thinkingType := gjson.Get(jsonStr, "thinking.type").String()
	if thinkingType == "enabled" || thinkingType == "adaptive" {
		parsed.ThinkingEnabled = true
	}

	parsed.OutputEffort = strings.TrimSpace(gjson.Get(jsonStr, "output_config.effort").String())

	maxTokensResult := gjson.Get(jsonStr, "max_tokens")
	if maxTokensResult.Exists() && maxTokensResult.Type == gjson.Number {
		f := maxTokensResult.Float()
		if !math.IsNaN(f) && !math.IsInf(f, 0) && f == math.Trunc(f) &&
			f <= float64(math.MaxInt) && f >= float64(math.MinInt) {
			parsed.MaxTokens = int(f)
		}
	}

	switch protocol {
	case domain.PlatformGemini:
		if sysParts := gjson.Get(jsonStr, "systemInstruction.parts"); sysParts.Exists() && sysParts.IsArray() {
			var parts []any
			if err := json.Unmarshal(sliceRawFromBody(body, sysParts), &parts); err != nil {
				return nil, err
			}
			parsed.System = parts
		}
		if contents := gjson.Get(jsonStr, "contents"); contents.Exists() && contents.IsArray() {
			var msgs []any
			if err := json.Unmarshal(sliceRawFromBody(body, contents), &msgs); err != nil {
				return nil, err
			}
			parsed.Messages = msgs
		}
	default:
		if sys := gjson.Get(jsonStr, "system"); sys.Exists() {
			parsed.HasSystem = true
			switch sys.Type {
			case gjson.Null:
				parsed.System = nil
			case gjson.String:
				parsed.System = sys.String()
			default:
				var system any
				if err := json.Unmarshal(sliceRawFromBody(body, sys), &system); err != nil {
					return nil, err
				}
				parsed.System = system
			}
		}

		if msgs := gjson.Get(jsonStr, "messages"); msgs.Exists() && msgs.IsArray() {
			var messages []any
			if err := json.Unmarshal(sliceRawFromBody(body, msgs), &messages); err != nil {
				return nil, err
			}
			parsed.Messages = messages
		}
	}

	return parsed, nil
}

func sliceRawFromBody(body []byte, r gjson.Result) []byte {
	if r.Index > 0 {
		end := r.Index + len(r.Raw)
		if end <= len(body) {
			return body[r.Index:end]
		}
	}
	return []byte(r.Raw)
}
