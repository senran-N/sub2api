package service

import "regexp"

var (
	grokToolOpenTagRe  = regexp.MustCompile(`(?i)<tool_calls[\s>]?`)
	grokToolCloseTagRe = regexp.MustCompile(`(?i)</tool_calls\s*>`)
)

type grokToolSieve struct {
	toolNames []string
	buffer    string
	capturing bool
	done      bool
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

	match := grokToolOpenTagRe.FindStringIndex(combined)
	if match == nil {
		safe, leftover := splitGrokToolBoundary(combined, "<tool_calls")
		s.buffer = leftover
		return safe, nil, false
	}

	safePart := combined[:match[0]]
	s.buffer = combined[match[0]:]
	s.capturing = true
	captureSafe, calls, ok := s.feedCapturing("")
	return safePart + captureSafe, calls, ok
}

func (s *grokToolSieve) feedCapturing(chunk string) (string, []grokParsedToolCall, bool) {
	s.buffer += chunk
	match := grokToolCloseTagRe.FindStringIndex(s.buffer)
	if match == nil {
		return "", nil, false
	}
	xmlBlock := s.buffer[:match[1]]
	s.buffer = ""
	s.capturing = false
	s.done = true
	result := parseGrokToolCalls(xmlBlock, s.toolNames)
	if result.SawToolSyntax {
		return "", result.Calls, true
	}
	return "", nil, false
}

func splitGrokToolBoundary(text, prefix string) (string, string) {
	maxLen := len(prefix) - 1
	if len(text) < maxLen {
		maxLen = len(text)
	}
	for i := maxLen; i > 0; i-- {
		if len(text) >= i && text[len(text)-i:] == prefix[:i] {
			return text[:len(text)-i], text[len(text)-i:]
		}
	}
	return text, ""
}
