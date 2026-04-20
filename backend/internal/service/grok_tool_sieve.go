package service

import "regexp"

var (
	grokToolCapturePatterns = []grokToolCapturePattern{
		{
			prefix:  "<tool_calls",
			openRe:  regexp.MustCompile(`(?i)<tool_calls[\s>]?`),
			closeRe: regexp.MustCompile(`(?i)</tool_calls\s*>`),
		},
		{
			prefix:  "<function_call",
			openRe:  regexp.MustCompile(`(?i)<function_call[\s>]?`),
			closeRe: regexp.MustCompile(`(?i)</function_call\s*>`),
		},
		{
			prefix:  "<invoke",
			openRe:  regexp.MustCompile(`(?i)<invoke[\s>]`),
			closeRe: regexp.MustCompile(`(?i)</invoke\s*>`),
		},
	}
)

type grokToolCapturePattern struct {
	prefix  string
	openRe  *regexp.Regexp
	closeRe *regexp.Regexp
}

type grokToolSieve struct {
	toolNames      []string
	buffer         string
	capturing      bool
	done           bool
	capturePattern *grokToolCapturePattern
}

func newGrokToolSieve(toolNames []string) *grokToolSieve {
	return &grokToolSieve{toolNames: append([]string(nil), toolNames...)}
}

func (s *grokToolSieve) Feed(chunk string) (string, []grokParsedToolCall, bool) {
	if s == nil {
		return chunk, nil, false
	}
	if s.done || chunk == "" {
		if s.capturing {
			return "", nil, false
		}
		return chunk, nil, false
	}
	if s.capturing {
		return s.feedCapturing(chunk)
	}
	return s.feedScanning(chunk)
}

func (s *grokToolSieve) Flush() ([]grokParsedToolCall, bool) {
	if s == nil || s.done || s.buffer == "" {
		return nil, false
	}
	s.done = true
	result := parseGrokToolCalls(s.buffer, s.toolNames)
	s.buffer = ""
	if result.SawToolSyntax {
		return result.Calls, true
	}
	return nil, false
}

func (s *grokToolSieve) feedScanning(chunk string) (string, []grokParsedToolCall, bool) {
	combined := s.buffer + chunk
	s.buffer = ""

	pattern, match := findGrokToolCaptureStart(combined)
	if pattern == nil || match == nil {
		safe, leftover := splitGrokToolBoundary(combined, grokToolCapturePrefixes())
		s.buffer = leftover
		return safe, nil, false
	}

	safePart := combined[:match[0]]
	s.buffer = combined[match[0]:]
	s.capturing = true
	s.capturePattern = pattern
	captureSafe, calls, ok := s.feedCapturing("")
	return safePart + captureSafe, calls, ok
}

func (s *grokToolSieve) feedCapturing(chunk string) (string, []grokParsedToolCall, bool) {
	s.buffer += chunk
	if s.capturePattern == nil || s.capturePattern.closeRe == nil {
		return "", nil, false
	}
	match := s.capturePattern.closeRe.FindStringIndex(s.buffer)
	if match == nil {
		return "", nil, false
	}
	xmlBlock := s.buffer[:match[1]]
	s.buffer = ""
	s.capturing = false
	s.done = true
	s.capturePattern = nil
	result := parseGrokToolCalls(xmlBlock, s.toolNames)
	if result.SawToolSyntax {
		return "", result.Calls, true
	}
	return "", nil, false
}

func findGrokToolCaptureStart(text string) (*grokToolCapturePattern, []int) {
	bestIndex := -1
	var bestPattern *grokToolCapturePattern
	var bestMatch []int
	for idx := range grokToolCapturePatterns {
		pattern := &grokToolCapturePatterns[idx]
		match := pattern.openRe.FindStringIndex(text)
		if match == nil {
			continue
		}
		if bestIndex >= 0 && match[0] >= bestIndex {
			continue
		}
		bestIndex = match[0]
		bestPattern = pattern
		bestMatch = match
	}
	return bestPattern, bestMatch
}

func grokToolCapturePrefixes() []string {
	prefixes := make([]string, 0, len(grokToolCapturePatterns))
	for _, pattern := range grokToolCapturePatterns {
		if pattern.prefix != "" {
			prefixes = append(prefixes, pattern.prefix)
		}
	}
	return prefixes
}

func splitGrokToolBoundary(text string, prefixes []string) (string, string) {
	bestLen := 0
	for _, prefix := range prefixes {
		maxLen := len(prefix) - 1
		if len(text) < maxLen {
			maxLen = len(text)
		}
		for i := maxLen; i > 0; i-- {
			if len(text) >= i && text[len(text)-i:] == prefix[:i] {
				if i > bestLen {
					bestLen = i
				}
				break
			}
		}
	}
	if bestLen > 0 {
		return text[:len(text)-bestLen], text[len(text)-bestLen:]
	}
	return text, ""
}
