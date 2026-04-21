package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type grokSessionReasoningEmission struct {
	Text string
	Line bool
}

type grokSessionReasoningEmitter struct {
	summaryMode     bool
	formatter       *grokSessionReasoningFormatter
	emittedLineKeys map[string]struct{}
	lastRollout     string
}

type grokSessionReasoningFormatter struct {
	language       string
	sectionStarted map[string]struct{}
	emittedKeys    map[string]struct{}
}

type grokSessionToolDisplay struct {
	emoji   string
	argKeys []string
}

var (
	grokSessionProgressHintRe = regexp.MustCompile(`(?i)(正在|准备|计划|查找|搜索|浏览|确认|核对|整合|分析|search|browse|check|verify|plan|investigat|research|review|look)`)
	grokSessionFindingHintRe  = regexp.MustCompile(`(?i)(已经|已|确认|发现|显示|表明|结果|找到|confirmed|found|shows|indicates|result|conclusion|evidence)`)
	grokSessionURLStripRe     = regexp.MustCompile(`https?://\S+`)
	grokSessionKeyNormalizeRe = regexp.MustCompile(`[^\p{Han}\p{L}\p{N}]+`)
	grokSessionHanRe          = regexp.MustCompile(`[\p{Han}]`)
	grokSessionLatinRe        = regexp.MustCompile(`[A-Za-z]`)
	grokSessionToolDisplays   = map[string]grokSessionToolDisplay{
		"web_search":        {emoji: "🔍", argKeys: []string{"query", "q"}},
		"x_search":          {emoji: "🔍", argKeys: []string{"query"}},
		"x_keyword_search":  {emoji: "🔍", argKeys: []string{"query"}},
		"x_semantic_search": {emoji: "🔍", argKeys: []string{"query"}},
		"browse_page":       {emoji: "🌐", argKeys: []string{"url"}},
		"search_images":     {emoji: "🖼️", argKeys: []string{"image_description", "imageDescription"}},
		"image_search":      {emoji: "🖼️", argKeys: []string{"image_description", "imageDescription"}},
		"chatroom_send":     {emoji: "📋", argKeys: []string{"message"}},
		"code_execution":    {emoji: "💻", argKeys: nil},
	}
)

func newGrokSessionReasoningEmitter(settings GrokTextSettings) *grokSessionReasoningEmitter {
	emitter := &grokSessionReasoningEmitter{
		summaryMode:     settings.ThinkingSummary,
		emittedLineKeys: make(map[string]struct{}),
	}
	if settings.ThinkingSummary {
		emitter.formatter = newGrokSessionReasoningFormatter()
	}
	return emitter
}

func newGrokSessionReasoningFormatter() *grokSessionReasoningFormatter {
	return &grokSessionReasoningFormatter{
		sectionStarted: make(map[string]struct{}),
		emittedKeys:    make(map[string]struct{}),
	}
}

func (e *grokSessionReasoningEmitter) Handle(
	delta grokSessionResponseDelta,
	contentStarted bool,
	streamMode bool,
) []grokSessionReasoningEmission {
	if e == nil {
		return nil
	}

	switch delta.messageTag {
	case "raw_function_result":
		return nil
	case "tool_usage_card":
		if contentStarted {
			return nil
		}
		return e.handleToolUsage(delta.toolUsageCardJSON, delta.rolloutID, delta.messageStepID)
	}

	if !delta.reasoning || strings.TrimSpace(delta.token) == "" {
		return nil
	}
	if contentStarted && streamMode {
		return nil
	}
	if e.summaryMode && e.formatter != nil {
		return e.wrapLines(e.formatter.OnThinking(delta.token, delta.messageTag, delta.messageStepID))
	}
	return e.handleDetailedThinking(delta.token, delta.rolloutID)
}

func (e *grokSessionReasoningEmitter) Finalize() []grokSessionReasoningEmission {
	if e == nil || !e.summaryMode || e.formatter == nil {
		return nil
	}
	return e.wrapLines(e.formatter.Finalize())
}

func (e *grokSessionReasoningEmitter) handleDetailedThinking(token string, rollout string) []grokSessionReasoningEmission {
	raw := token
	if strings.HasPrefix(raw, "- ") {
		raw = raw[2:]
	}
	if raw == "" {
		return nil
	}

	emissions := make([]grokSessionReasoningEmission, 0, 2)
	if rollout = strings.TrimSpace(rollout); rollout != "" && rollout != e.lastRollout {
		e.lastRollout = rollout
		emissions = append(emissions, grokSessionReasoningEmission{
			Text: fmt.Sprintf("\n[%s]\n", rollout),
		})
	}
	emissions = append(emissions, grokSessionReasoningEmission{Text: raw})
	return emissions
}

func (e *grokSessionReasoningEmitter) handleToolUsage(
	rawCard string,
	rollout string,
	stepID int,
) []grokSessionReasoningEmission {
	toolName, args := extractGrokSessionToolUsageInfo(rawCard)
	if toolName == "" {
		return nil
	}

	if e.summaryMode && e.formatter != nil {
		return e.wrapLines(e.formatter.OnToolUsage(toolName, args, rollout, stepID))
	}

	line := formatGrokSessionToolUsageLine(toolName, args, rollout)
	if line == "" {
		return nil
	}
	key := "detail:" + grokSessionNormalizeReasoningKey(strings.ToLower(strings.TrimSpace(line)))
	if _, exists := e.emittedLineKeys[key]; exists {
		return nil
	}
	e.emittedLineKeys[key] = struct{}{}
	return []grokSessionReasoningEmission{{Text: line, Line: true}}
}

func (e *grokSessionReasoningEmitter) wrapLines(lines []string) []grokSessionReasoningEmission {
	if len(lines) == 0 {
		return nil
	}
	emissions := make([]grokSessionReasoningEmission, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		emissions = append(emissions, grokSessionReasoningEmission{
			Text: line,
			Line: true,
		})
	}
	return emissions
}

func (f *grokSessionReasoningFormatter) OnThinking(token string, tag string, stepID int) []string {
	if f == nil {
		return nil
	}
	f.observeLanguage(token)

	text := strings.TrimSpace(token)
	if strings.HasPrefix(text, "- ") {
		text = strings.TrimSpace(text[2:])
	}
	if text == "" {
		return nil
	}

	section := "scope"
	switch strings.TrimSpace(tag) {
	case "header":
		if strings.EqualFold(text, "thinking about your request") {
			return nil
		}
		if stepID <= 1 {
			section = "understanding"
		} else {
			section = "evidence"
		}
	case "summary":
		section = f.sectionForSummary(text, stepID)
	default:
		section = f.sectionForSummary(text, stepID)
	}
	return f.emit(section, grokSessionBulletize(text))
}

func (f *grokSessionReasoningFormatter) OnToolUsage(
	toolName string,
	args map[string]any,
	_ string,
	_ int,
) []string {
	if f == nil {
		return nil
	}

	for _, key := range []string{"query", "q", "message", "instructions", "url", "image_description", "imageDescription"} {
		if value := strings.TrimSpace(toString(args[key])); value != "" {
			f.observeLanguage(value)
			break
		}
	}

	switch toolName {
	case "web_search":
		query := strings.TrimSpace(firstNonEmpty(toString(args["query"]), toString(args["q"])))
		if query == "" {
			return nil
		}
		return f.emit("scope", f.localizedToolLine("web_search", query))
	case "x_search", "x_keyword_search", "x_semantic_search":
		query := strings.TrimSpace(toString(args["query"]))
		if query == "" {
			return nil
		}
		return f.emit("evidence", f.localizedToolLine("social_search", query))
	case "browse_page":
		url := strings.TrimSpace(toString(args["url"]))
		if url == "" {
			return nil
		}
		return f.emit("evidence", f.localizedToolLine("browse_page", url))
	case "search_images", "image_search":
		desc := strings.TrimSpace(firstNonEmpty(toString(args["image_description"]), toString(args["imageDescription"])))
		if desc == "" {
			return nil
		}
		return f.emit("scope", f.localizedToolLine("image_search", desc))
	case "code_execution":
		return f.emit("evidence", f.localizedToolLine("code_execution", ""))
	case "chatroom_send":
		message := strings.TrimSpace(toString(args["message"]))
		if message == "" {
			return nil
		}
		return f.emit("finding", f.localizedToolLine("chatroom_send", message))
	default:
		return nil
	}
}

func (f *grokSessionReasoningFormatter) Finalize() []string {
	return nil
}

func (f *grokSessionReasoningFormatter) sectionForSummary(text string, stepID int) string {
	switch {
	case grokSessionFindingHintRe.MatchString(text):
		if stepID <= 1 {
			return "understanding"
		}
		return "finding"
	case grokSessionProgressHintRe.MatchString(text):
		if strings.Contains(strings.ToLower(text), "confirm") || strings.Contains(text, "核对") || strings.Contains(text, "确认") {
			return "evidence"
		}
		return "scope"
	case stepID <= 1:
		return "understanding"
	default:
		return "scope"
	}
}

func (f *grokSessionReasoningFormatter) emit(section string, text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	key := section + ":" + grokSessionNormalizeReasoningKey(text)
	if _, exists := f.emittedKeys[key]; exists {
		return nil
	}
	f.emittedKeys[key] = struct{}{}

	lines := make([]string, 0, 2)
	if _, started := f.sectionStarted[section]; !started {
		f.sectionStarted[section] = struct{}{}
		lines = append(lines, f.sectionTitle(section))
	}
	lines = append(lines, text)
	return lines
}

func (f *grokSessionReasoningFormatter) sectionTitle(section string) string {
	english := f.language == "en"
	switch section {
	case "understanding":
		if english {
			return "Understanding"
		}
		return "理解问题"
	case "evidence":
		if english {
			return "Verification"
		}
		return "核验与证据"
	case "finding":
		if english {
			return "Key Findings"
		}
		return "关键发现"
	default:
		if english {
			return "Research Scope"
		}
		return "检索范围"
	}
}

func (f *grokSessionReasoningFormatter) localizedToolLine(kind string, value string) string {
	english := f.language == "en"
	switch kind {
	case "web_search":
		if english {
			return fmt.Sprintf("- Parallel research: %s.", value)
		}
		return fmt.Sprintf("- 并行检索：%s。", value)
	case "social_search":
		if english {
			return fmt.Sprintf("- Social cross-check: %s.", value)
		}
		return fmt.Sprintf("- 社媒交叉核验：%s。", value)
	case "browse_page":
		if english {
			return fmt.Sprintf("- Page verification: %s.", value)
		}
		return fmt.Sprintf("- 页面核对：%s。", value)
	case "image_search":
		if english {
			return fmt.Sprintf("- Visual asset search: %s.", value)
		}
		return fmt.Sprintf("- 视觉素材检索：%s。", value)
	case "code_execution":
		if english {
			return "- Executing code or generating runnable output."
		}
		return "- 正在执行代码或生成可运行内容。"
	case "chatroom_send":
		if english {
			return "- Consolidating findings from parallel agents."
		}
		return "- 正在整合并行代理返回的材料。"
	default:
		return ""
	}
}

func (f *grokSessionReasoningFormatter) observeLanguage(text string) {
	if f == nil || f.language != "" {
		return
	}
	if grokSessionHanRe.MatchString(text) {
		f.language = "zh"
		return
	}
	if grokSessionLatinRe.MatchString(text) {
		f.language = "en"
	}
}

func extractGrokSessionToolUsageInfo(raw string) (string, map[string]any) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	var card map[string]any
	if err := json.Unmarshal([]byte(raw), &card); err != nil {
		return "", nil
	}
	for key, value := range card {
		if key == "toolUsageCardId" {
			continue
		}
		payload, ok := value.(map[string]any)
		if !ok {
			continue
		}
		toolName := grokSessionCamelToSnake(strings.TrimSpace(key))
		args, _ := payload["args"].(map[string]any)
		if args == nil {
			args = map[string]any{}
		}
		return toolName, args
	}
	return "", nil
}

func formatGrokSessionToolUsageLine(toolName string, args map[string]any, rollout string) string {
	display, ok := grokSessionToolDisplays[toolName]
	if !ok {
		display = grokSessionToolDisplay{emoji: "🔧"}
	}

	argText := ""
	for _, key := range display.argKeys {
		value := strings.TrimSpace(toString(args[key]))
		if value != "" {
			argText = value
			break
		}
	}

	prefix := ""
	if rollout = strings.TrimSpace(rollout); rollout != "" {
		prefix = "[" + rollout + "] "
	}
	if argText != "" {
		return fmt.Sprintf("%s%s %s: %s", prefix, display.emoji, toolName, argText)
	}
	return fmt.Sprintf("%s%s %s", prefix, display.emoji, toolName)
}

func grokSessionBulletize(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if strings.HasPrefix(text, "- ") {
		return text
	}
	return "- " + text
}

func grokSessionNormalizeReasoningKey(text string) string {
	lowered := strings.ToLower(strings.TrimSpace(text))
	lowered = grokSessionURLStripRe.ReplaceAllString(lowered, "")
	return grokSessionKeyNormalizeRe.ReplaceAllString(lowered, "")
}

func grokSessionCamelToSnake(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value) + 4)
	for idx, r := range value {
		if unicode.IsUpper(r) {
			if idx > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(unicode.ToLower(r))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}
