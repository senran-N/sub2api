package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/senran-N/sub2api/internal/pkg/apicompat"
)

const (
	grokToolChoiceAuto     = "WHEN TO CALL: Call a tool when it is clearly needed. Otherwise respond in plain text."
	grokToolChoiceNone     = "WHEN TO CALL: Do NOT call any tools. Respond in plain text only."
	grokToolChoiceRequired = "WHEN TO CALL: You MUST output a <tool_calls> XML block. Do NOT write any plain-text reply. If you are uncertain, still call the most relevant tool with your best guess at the parameters."
)

const grokToolSystemHeader = `You have access to the following tools.

AVAILABLE TOOLS:
%s

TOOL CALL FORMAT - follow these rules exactly:
- When calling a tool, output ONLY the XML block below. No text before or after it.
- <parameters> must be a single-line valid JSON object (no line breaks inside).
- Place multiple tool calls inside ONE <tool_calls> element.
- Do NOT use markdown code fences around the XML.
- Do NOT output any inner monologue or explanation alongside the XML.

<tool_calls>
  <tool_call>
    <tool_name>TOOL_NAME</tool_name>
    <parameters>{"key":"value"}</parameters>
  </tool_call>
</tool_calls>

WRONG (never do this):
<tool_calls>...</tool_calls> inside code fences
I'll call the search tool now. <tool_calls>...</tool_calls>

%s
NOTE: Even if you believe you cannot fulfill the request, you must still follow the WHEN TO CALL rule above.`

type grokParsedToolCall struct {
	CallID    string
	Name      string
	Arguments string
}

type grokToolParseResult struct {
	Calls         []grokParsedToolCall
	SawToolSyntax bool
}

func grokSessionToolPromptFromResponsesTools(tools []apicompat.ResponsesTool, toolChoice json.RawMessage) (string, []string) {
	functionTools := grokFunctionTools(tools)
	if len(functionTools) == 0 {
		return "", nil
	}
	return buildGrokToolSystemPrompt(functionTools, toolChoice), extractGrokToolNames(functionTools)
}

func grokFunctionTools(tools []apicompat.ResponsesTool) []apicompat.ResponsesTool {
	result := make([]apicompat.ResponsesTool, 0, len(tools))
	for _, tool := range tools {
		if strings.TrimSpace(tool.Type) != "function" {
			continue
		}
		if strings.TrimSpace(tool.Name) == "" {
			continue
		}
		result = append(result, tool)
	}
	return result
}

func extractGrokToolNames(tools []apicompat.ResponsesTool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names
}

func buildGrokToolSystemPrompt(tools []apicompat.ResponsesTool, toolChoice json.RawMessage) string {
	return fmt.Sprintf(
		grokToolSystemHeader,
		formatGrokToolDefinitions(tools),
		buildGrokToolChoiceInstruction(toolChoice),
	)
}

func formatGrokToolDefinitions(tools []apicompat.ResponsesTool) string {
	parts := make([]string, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			continue
		}
		lines := []string{fmt.Sprintf("Tool: %s", name)}
		if desc := strings.TrimSpace(tool.Description); desc != "" {
			lines = append(lines, fmt.Sprintf("Description: %s", desc))
		}
		if len(tool.Parameters) > 0 && string(tool.Parameters) != "null" {
			lines = append(lines, fmt.Sprintf("Parameters: %s", strings.TrimSpace(string(tool.Parameters))))
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}
	return strings.Join(parts, "\n\n")
}

func buildGrokToolChoiceInstruction(toolChoice json.RawMessage) string {
	trimmed := strings.TrimSpace(string(toolChoice))
	if trimmed == "" || trimmed == "null" {
		return grokToolChoiceAuto
	}

	var stringChoice string
	if err := json.Unmarshal(toolChoice, &stringChoice); err == nil {
		switch strings.TrimSpace(stringChoice) {
		case "", "auto":
			return grokToolChoiceAuto
		case "none":
			return grokToolChoiceNone
		case "required":
			return grokToolChoiceRequired
		}
	}

	var objectChoice map[string]any
	if err := json.Unmarshal(toolChoice, &objectChoice); err != nil {
		return grokToolChoiceAuto
	}

	switch strings.TrimSpace(grokToolAsString(objectChoice["type"])) {
	case "none":
		return grokToolChoiceNone
	case "required":
		return grokToolChoiceRequired
	case "function":
		if function, ok := objectChoice["function"].(map[string]any); ok {
			if forcedName := strings.TrimSpace(grokToolAsString(function["name"])); forcedName != "" {
				return fmt.Sprintf(
					`WHEN TO CALL: You MUST output a <tool_calls> XML block calling the tool named "%s". Do NOT write any plain-text reply under any circumstances.`,
					forcedName,
				)
			}
		}
	}
	return grokToolChoiceAuto
}

func grokToolCallsToXML(calls []grokParsedToolCall) string {
	if len(calls) == 0 {
		return ""
	}
	lines := make([]string, 0, len(calls)*4+2)
	lines = append(lines, "<tool_calls>")
	for _, call := range calls {
		name := strings.TrimSpace(call.Name)
		if name == "" {
			continue
		}
		args := compactGrokToolArguments(call.Arguments)
		lines = append(lines,
			"  <tool_call>",
			fmt.Sprintf("    <tool_name>%s</tool_name>", name),
			fmt.Sprintf("    <parameters>%s</parameters>", args),
			"  </tool_call>",
		)
	}
	lines = append(lines, "</tool_calls>")
	return strings.Join(lines, "\n")
}

func compactGrokToolArguments(arguments string) string {
	trimmed := strings.TrimSpace(arguments)
	if trimmed == "" {
		return "{}"
	}
	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return trimmed
	}
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return trimmed
	}
	return string(encoded)
}

func mergeGrokSessionSystemPrompt(base, injected string) string {
	base = strings.TrimSpace(base)
	injected = strings.TrimSpace(injected)
	switch {
	case base == "":
		return injected
	case injected == "":
		return base
	default:
		return base + "\n\n" + injected
	}
}

func grokToolAsString(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprint(value)
}
