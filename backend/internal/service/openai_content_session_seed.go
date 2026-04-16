package service

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

// contentSessionSeedPrefix prevents collisions between content-derived seeds
// and explicit session identifiers such as session_id or prompt_cache_key.
const contentSessionSeedPrefix = "compat_cs_"

// deriveOpenAIContentSessionSeed builds a stable fallback seed from fields that
// should remain constant for the same logical conversation bootstrap.
func deriveOpenAIContentSessionSeed(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var builder strings.Builder

	if model := gjson.GetBytes(body, "model").String(); model != "" {
		_, _ = builder.WriteString("model=")
		_, _ = builder.WriteString(model)
	}

	if tools := gjson.GetBytes(body, "tools"); tools.Exists() && tools.IsArray() && tools.Raw != "[]" {
		_, _ = builder.WriteString("|tools=")
		_, _ = builder.WriteString(normalizeCompatSeedJSON(json.RawMessage(tools.Raw)))
	}

	if functions := gjson.GetBytes(body, "functions"); functions.Exists() && functions.IsArray() && functions.Raw != "[]" {
		_, _ = builder.WriteString("|functions=")
		_, _ = builder.WriteString(normalizeCompatSeedJSON(json.RawMessage(functions.Raw)))
	}

	if instructions := gjson.GetBytes(body, "instructions").String(); instructions != "" {
		_, _ = builder.WriteString("|instructions=")
		_, _ = builder.WriteString(instructions)
	}

	firstUserCaptured := false

	messages := gjson.GetBytes(body, "messages")
	if messages.Exists() && messages.IsArray() {
		messages.ForEach(func(_, message gjson.Result) bool {
			role := strings.TrimSpace(message.Get("role").String())
			switch role {
			case "system", "developer":
				_, _ = builder.WriteString("|system=")
				if content := message.Get("content"); content.Exists() {
					_, _ = builder.WriteString(normalizeCompatSeedJSON(json.RawMessage(content.Raw)))
				}
			case "user":
				if firstUserCaptured {
					return true
				}
				_, _ = builder.WriteString("|first_user=")
				if content := message.Get("content"); content.Exists() {
					_, _ = builder.WriteString(normalizeCompatSeedJSON(json.RawMessage(content.Raw)))
				}
				firstUserCaptured = true
			}
			return true
		})
	} else if input := gjson.GetBytes(body, "input"); input.Exists() {
		if input.Type == gjson.String {
			_, _ = builder.WriteString("|input=")
			_, _ = builder.WriteString(input.String())
		} else if input.IsArray() {
			input.ForEach(func(_, item gjson.Result) bool {
				role := strings.TrimSpace(item.Get("role").String())
				switch role {
				case "system", "developer":
					_, _ = builder.WriteString("|system=")
					if content := item.Get("content"); content.Exists() {
						_, _ = builder.WriteString(normalizeCompatSeedJSON(json.RawMessage(content.Raw)))
					}
				case "user":
					if firstUserCaptured {
						return true
					}
					_, _ = builder.WriteString("|first_user=")
					if content := item.Get("content"); content.Exists() {
						_, _ = builder.WriteString(normalizeCompatSeedJSON(json.RawMessage(content.Raw)))
					}
					firstUserCaptured = true
				}

				if !firstUserCaptured && item.Get("type").String() == "input_text" {
					_, _ = builder.WriteString("|first_user=")
					_, _ = builder.WriteString(item.Get("text").String())
					firstUserCaptured = true
				}
				return true
			})
		}
	}

	if builder.Len() == 0 {
		return ""
	}

	return contentSessionSeedPrefix + builder.String()
}
