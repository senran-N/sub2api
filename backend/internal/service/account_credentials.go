package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

func (a *Account) GetCredential(key string) string {
	if a.Credentials == nil {
		return ""
	}
	value, ok := a.Credentials[key]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}

func (a *Account) GetCredentialAsTime(key string) *time.Time {
	value := a.GetCredential(key)
	if value == "" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed
	}
	if unixSeconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		parsed := time.Unix(unixSeconds, 0)
		return &parsed
	}
	return nil
}

func (a *Account) GetCredentialAsInt64(key string) int64 {
	if a == nil || a.Credentials == nil {
		return 0
	}
	value, ok := a.Credentials[key]
	if !ok || value == nil {
		return 0
	}

	switch typed := value.(type) {
	case int64:
		return typed
	case float64:
		return int64(typed)
	case int:
		return int64(typed)
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}

func (a *Account) IsTempUnschedulableEnabled() bool {
	if a.Credentials == nil {
		return false
	}
	raw, ok := a.Credentials["temp_unschedulable_enabled"]
	if !ok || raw == nil {
		return false
	}
	enabled, ok := raw.(bool)
	return ok && enabled
}

func (a *Account) GetTempUnschedulableRules() []TempUnschedulableRule {
	if a.Credentials == nil {
		return nil
	}
	raw, ok := a.Credentials["temp_unschedulable_rules"]
	if !ok || raw == nil {
		return nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	rules := make([]TempUnschedulableRule, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok || entry == nil {
			continue
		}

		rule := TempUnschedulableRule{
			ErrorCode:       parseTempUnschedInt(entry["error_code"]),
			Keywords:        parseTempUnschedStrings(entry["keywords"]),
			DurationMinutes: parseTempUnschedInt(entry["duration_minutes"]),
			Description:     parseTempUnschedString(entry["description"]),
		}
		if rule.ErrorCode <= 0 || rule.DurationMinutes <= 0 || len(rule.Keywords) == 0 {
			continue
		}
		rules = append(rules, rule)
	}

	return rules
}

func parseTempUnschedString(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func parseTempUnschedStrings(value any) []string {
	if value == nil {
		return nil
	}

	var raw []string
	switch typed := value.(type) {
	case []string:
		raw = typed
	case []any:
		raw = make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				raw = append(raw, text)
			}
		}
	default:
		return nil
	}

	result := make([]string, 0, len(raw))
	for _, item := range raw {
		text := strings.TrimSpace(item)
		if text != "" {
			result = append(result, text)
		}
	}
	return result
}

func normalizeAccountNotes(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func parseTempUnschedInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed)
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed
		}
	}
	return 0
}

func (a *Account) GetExtraString(key string) string {
	if a.Extra == nil {
		return ""
	}
	if value, ok := a.Extra[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return ""
}

func (a *Account) getExtraFloat64(key string) float64 {
	if a.Extra == nil {
		return 0
	}
	if value, ok := a.Extra[key]; ok {
		return parseExtraFloat64(value)
	}
	return 0
}

func (a *Account) getExtraTime(key string) time.Time {
	if a.Extra == nil {
		return time.Time{}
	}
	if value, ok := a.Extra[key]; ok {
		if text, ok := value.(string); ok {
			if parsed, err := time.Parse(time.RFC3339Nano, text); err == nil {
				return parsed
			}
			if parsed, err := time.Parse(time.RFC3339, text); err == nil {
				return parsed
			}
		}
	}
	return time.Time{}
}

func (a *Account) getExtraString(key string) string {
	if a.Extra == nil {
		return ""
	}
	if value, ok := a.Extra[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return ""
}

func (a *Account) getExtraInt(key string) int {
	if a.Extra == nil {
		return 0
	}
	if value, ok := a.Extra[key]; ok {
		return int(parseExtraFloat64(value))
	}
	return 0
}

func parseExtraFloat64(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return parsed
		}
	}
	return 0
}

func ParseExtraInt(value any) int {
	return parseExtraInt(value)
}

func parseExtraInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed)
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed
		}
	}
	return 0
}
