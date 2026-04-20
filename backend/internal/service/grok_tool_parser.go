package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	grokToolSyntaxRe   = regexp.MustCompile(`(?is)<tool_calls|<tool_call|<function_call|<invoke\s|"tool_calls"\s*:|\btool_calls\b`)
	grokXMLRootRe      = regexp.MustCompile(`(?is)<tool_calls\s*>(.*?)</tool_calls\s*>`)
	grokXMLCallRe      = regexp.MustCompile(`(?is)<tool_call\s*>(.*?)</tool_call\s*>`)
	grokXMLNameRe      = regexp.MustCompile(`(?is)<tool_name\s*>(.*?)</tool_name\s*>`)
	grokXMLParamsRe    = regexp.MustCompile(`(?is)<parameters\s*>(.*?)</parameters\s*>`)
	grokJSONArrRe      = regexp.MustCompile(`(?s)\[[\s\S]+\]`)
	grokFunctionCallRe = regexp.MustCompile(`(?is)<function_call\s*>(.*?)</function_call\s*>`)
	grokInvokeRe       = regexp.MustCompile(`(?is)<invoke\s+name=["']?(\w+)["']?\s*>(.*?)</invoke\s*>`)
	grokFunctionNameRe = regexp.MustCompile(`(?is)<name\s*>(.*?)</name\s*>`)
	grokFunctionArgsRe = regexp.MustCompile(`(?is)<arguments\s*>(.*?)</arguments\s*>`)
)

func parseGrokToolCalls(text string, availableTools []string) grokToolParseResult {
	result := grokToolParseResult{}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return result
	}
	if !grokToolSyntaxRe.MatchString(trimmed) {
		return result
	}
	result.SawToolSyntax = true

	calls := parseGrokXMLToolCalls(trimmed)
	if len(calls) == 0 {
		calls = parseGrokJSONToolEnvelope(trimmed)
	}
	if len(calls) == 0 {
		calls = parseGrokJSONArrayToolCalls(trimmed)
	}
	if len(calls) == 0 {
		calls = parseGrokAltXMLToolCalls(trimmed)
	}

	if len(availableTools) > 0 && len(calls) > 0 {
		allowed := make(map[string]struct{}, len(availableTools))
		for _, tool := range availableTools {
			name := strings.TrimSpace(tool)
			if name != "" {
				allowed[name] = struct{}{}
			}
		}
		filtered := make([]grokParsedToolCall, 0, len(calls))
		for _, call := range calls {
			if _, ok := allowed[call.Name]; ok {
				filtered = append(filtered, call)
			}
		}
		calls = filtered
	}

	result.Calls = calls
	return result
}

func parseGrokXMLToolCalls(text string) []grokParsedToolCall {
	root := grokXMLRootRe.FindStringSubmatch(text)
	if len(root) < 2 {
		return nil
	}
	matches := grokXMLCallRe.FindAllStringSubmatch(root[1], -1)
	if len(matches) == 0 {
		return nil
	}
	calls := make([]grokParsedToolCall, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		inner := match[1]
		nameMatch := grokXMLNameRe.FindStringSubmatch(inner)
		if len(nameMatch) < 2 {
			continue
		}
		name := strings.TrimSpace(nameMatch[1])
		if name == "" {
			continue
		}
		params := "{}"
		if paramsMatch := grokXMLParamsRe.FindStringSubmatch(inner); len(paramsMatch) >= 2 {
			params = strings.TrimSpace(paramsMatch[1])
		}
		parsedArgs, ok := parseGrokToolJSONTolerant(params)
		if !ok {
			continue
		}
		calls = append(calls, newGrokParsedToolCall(name, parsedArgs))
	}
	return calls
}

func parseGrokJSONToolEnvelope(text string) []grokParsedToolCall {
	if !strings.Contains(text, `"tool_calls"`) {
		return nil
	}
	start := strings.Index(text, "{")
	if start < 0 {
		return nil
	}

	var object map[string]any
	if err := decodeGrokOuterJSONObject(text[start:], &object); err != nil {
		return nil
	}
	items, _ := object["tool_calls"].([]any)
	return extractGrokToolCallsFromList(items)
}

func parseGrokJSONArrayToolCalls(text string) []grokParsedToolCall {
	match := grokJSONArrRe.FindString(text)
	if match == "" {
		return nil
	}
	var items []any
	if err := json.Unmarshal([]byte(match), &items); err != nil {
		return nil
	}
	return extractGrokToolCallsFromList(items)
}

func parseGrokAltXMLToolCalls(text string) []grokParsedToolCall {
	calls := make([]grokParsedToolCall, 0, 2)
	for _, match := range grokFunctionCallRe.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		inner := match[1]
		nameMatch := grokFunctionNameRe.FindStringSubmatch(inner)
		if len(nameMatch) < 2 {
			continue
		}
		name := strings.TrimSpace(nameMatch[1])
		if name == "" {
			continue
		}
		args := "{}"
		if argsMatch := grokFunctionArgsRe.FindStringSubmatch(inner); len(argsMatch) >= 2 {
			args = strings.TrimSpace(argsMatch[1])
		}
		parsedArgs, ok := parseGrokToolJSONTolerant(args)
		if !ok {
			continue
		}
		calls = append(calls, newGrokParsedToolCall(name, parsedArgs))
	}
	for _, match := range grokInvokeRe.FindAllStringSubmatch(text, -1) {
		if len(match) < 3 {
			continue
		}
		name := strings.TrimSpace(match[1])
		if name == "" {
			continue
		}
		parsedArgs, ok := parseGrokToolJSONTolerant(strings.TrimSpace(match[2]))
		if !ok {
			parsedArgs = map[string]any{}
		}
		calls = append(calls, newGrokParsedToolCall(name, parsedArgs))
	}
	return calls
}

func decodeGrokOuterJSONObject(text string, target any) error {
	end := strings.LastIndex(text, "}")
	if end < 0 {
		return json.Unmarshal([]byte(text), target)
	}
	return json.Unmarshal([]byte(text[:end+1]), target)
}

func extractGrokToolCallsFromList(items []any) []grokParsedToolCall {
	if len(items) == 0 {
		return nil
	}
	calls := make([]grokParsedToolCall, 0, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(grokToolAsString(firstNonNil(object["name"], object["tool_name"])))
		if name == "" {
			continue
		}
		args := firstNonNil(object["input"], object["arguments"], object["parameters"])
		if args == nil {
			args = map[string]any{}
		}
		calls = append(calls, newGrokParsedToolCall(name, args))
	}
	return calls
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func parseGrokToolJSONTolerant(raw string) (any, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{}, true
	}
	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return decoded, true
	}
	repaired := strings.ReplaceAll(trimmed, "\n", `\n`)
	if err := json.Unmarshal([]byte(repaired), &decoded); err == nil {
		return decoded, true
	}
	return nil, false
}

func newGrokParsedToolCall(name string, arguments any) grokParsedToolCall {
	return grokParsedToolCall{
		CallID:    newGrokToolCallID(),
		Name:      name,
		Arguments: compactGrokParsedArguments(arguments),
	}
}

func compactGrokParsedArguments(arguments any) string {
	if text, ok := arguments.(string); ok {
		return compactGrokToolArguments(text)
	}
	encoded, err := json.Marshal(arguments)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func newGrokToolCallID() string {
	var random [3]byte
	_, _ = rand.Read(random[:])
	return "call_" + strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + hex.EncodeToString(random[:])
}
