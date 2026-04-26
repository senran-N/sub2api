package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestIsClaudeCodeClient(t *testing.T) {
	tests := []struct {
		name           string
		userAgent      string
		metadataUserID string
		want           bool
	}{
		{
			name:           "Claude Code client",
			userAgent:      "claude-cli/1.0.62 (darwin; arm64)",
			metadataUserID: "user_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2_account__session_123e4567-e89b-12d3-a456-426614174000",
			want:           true,
		},
		{
			name:           "Claude Code without version suffix",
			userAgent:      "claude-cli/2.0.0",
			metadataUserID: `{"device_id":"d61f76d0aabbccdd00112233445566778899aabbccddeeff0011223344556677","account_uuid":"","session_id":"123e4567-e89b-12d3-a456-426614174001"}`,
			want:           true,
		},
		{
			name:           "Claude CLI with invalid metadata user_id",
			userAgent:      "claude-cli/2.0.0",
			metadataUserID: "session_abc",
			want:           false,
		},
		{
			name:           "Missing metadata user_id",
			userAgent:      "claude-cli/1.0.0",
			metadataUserID: "",
			want:           false,
		},
		{
			name:           "Different user agent",
			userAgent:      "curl/7.68.0",
			metadataUserID: "user123",
			want:           false,
		},
		{
			name:           "Empty user agent",
			userAgent:      "",
			metadataUserID: "user123",
			want:           false,
		},
		{
			name:           "Similar but not Claude CLI",
			userAgent:      "claude-api/1.0.0",
			metadataUserID: "user123",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaudeCodeClient(tt.userAgent, tt.metadataUserID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSystemIncludesClaudeCodePrompt(t *testing.T) {
	tests := []struct {
		name   string
		system any
		want   bool
	}{
		{
			name:   "nil system",
			system: nil,
			want:   false,
		},
		{
			name:   "empty string",
			system: "",
			want:   false,
		},
		{
			name:   "string with Claude Code prompt",
			system: claudeCodeSystemPrompt,
			want:   true,
		},
		{
			name:   "string with different content",
			system: "You are a helpful assistant.",
			want:   false,
		},
		{
			name:   "empty array",
			system: []any{},
			want:   false,
		},
		{
			name: "array with Claude Code prompt",
			system: []any{
				map[string]any{
					"type": "text",
					"text": claudeCodeSystemPrompt,
				},
			},
			want: true,
		},
		{
			name: "array with Claude Code prompt in second position",
			system: []any{
				map[string]any{"type": "text", "text": "First prompt"},
				map[string]any{"type": "text", "text": claudeCodeSystemPrompt},
			},
			want: true,
		},
		{
			name: "array without Claude Code prompt",
			system: []any{
				map[string]any{"type": "text", "text": "Custom prompt"},
			},
			want: false,
		},
		{
			name: "array with partial match (should not match)",
			system: []any{
				map[string]any{"type": "text", "text": "You are Claude"},
			},
			want: false,
		},
		// json.RawMessage cases (conversion path: ForwardAsResponses / ForwardAsChatCompletions)
		{
			name:   "json.RawMessage string with Claude Code prompt",
			system: json.RawMessage(`"` + claudeCodeSystemPrompt + `"`),
			want:   true,
		},
		{
			name:   "json.RawMessage string without Claude Code prompt",
			system: json.RawMessage(`"You are a helpful assistant"`),
			want:   false,
		},
		{
			name:   "json.RawMessage nil (empty)",
			system: json.RawMessage(nil),
			want:   false,
		},
		{
			name:   "json.RawMessage empty string",
			system: json.RawMessage(`""`),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := systemIncludesClaudeCodePrompt(tt.system)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestInjectClaudeCodePrompt(t *testing.T) {
	claudePrefix := strings.TrimSpace(claudeCodeSystemPrompt)

	tests := []struct {
		name           string
		body           string
		system         any
		wantSystemLen  int
		wantFirstText  string
		wantSecondText string
	}{
		{
			name:          "nil system",
			body:          `{"model":"claude-3"}`,
			system:        nil,
			wantSystemLen: 1,
			wantFirstText: claudeCodeSystemPrompt,
		},
		{
			name:          "empty string system",
			body:          `{"model":"claude-3"}`,
			system:        "",
			wantSystemLen: 1,
			wantFirstText: claudeCodeSystemPrompt,
		},
		{
			name:           "string system",
			body:           `{"model":"claude-3"}`,
			system:         "Custom prompt",
			wantSystemLen:  2,
			wantFirstText:  claudeCodeSystemPrompt,
			wantSecondText: claudePrefix + "\n\nCustom prompt",
		},
		{
			name:          "string system equals Claude Code prompt",
			body:          `{"model":"claude-3"}`,
			system:        claudeCodeSystemPrompt,
			wantSystemLen: 1,
			wantFirstText: claudeCodeSystemPrompt,
		},
		{
			name:   "array system",
			body:   `{"model":"claude-3"}`,
			system: []any{map[string]any{"type": "text", "text": "Custom"}},
			// Claude Code + Custom = 2
			wantSystemLen:  2,
			wantFirstText:  claudeCodeSystemPrompt,
			wantSecondText: claudePrefix + "\n\nCustom",
		},
		{
			name: "array system with existing Claude Code prompt (should dedupe)",
			body: `{"model":"claude-3"}`,
			system: []any{
				map[string]any{"type": "text", "text": claudeCodeSystemPrompt},
				map[string]any{"type": "text", "text": "Other"},
			},
			// Claude Code at start + Other = 2 (deduped)
			wantSystemLen:  2,
			wantFirstText:  claudeCodeSystemPrompt,
			wantSecondText: claudePrefix + "\n\nOther",
		},
		{
			name:          "empty array",
			body:          `{"model":"claude-3"}`,
			system:        []any{},
			wantSystemLen: 1,
			wantFirstText: claudeCodeSystemPrompt,
		},
		// json.RawMessage cases (conversion path: ForwardAsResponses / ForwardAsChatCompletions)
		{
			name:           "json.RawMessage string system",
			body:           `{"model":"claude-3","system":"Custom prompt"}`,
			system:         json.RawMessage(`"Custom prompt"`),
			wantSystemLen:  2,
			wantFirstText:  claudeCodeSystemPrompt,
			wantSecondText: claudePrefix + "\n\nCustom prompt",
		},
		{
			name:          "json.RawMessage nil system",
			body:          `{"model":"claude-3"}`,
			system:        json.RawMessage(nil),
			wantSystemLen: 1,
			wantFirstText: claudeCodeSystemPrompt,
		},
		{
			name:          "json.RawMessage Claude Code prompt (should not duplicate)",
			body:          `{"model":"claude-3","system":"` + claudeCodeSystemPrompt + `"}`,
			system:        json.RawMessage(`"` + claudeCodeSystemPrompt + `"`),
			wantSystemLen: 1,
			wantFirstText: claudeCodeSystemPrompt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectClaudeCodePrompt([]byte(tt.body), tt.system)

			var parsed map[string]any
			err := json.Unmarshal(result, &parsed)
			require.NoError(t, err)

			system, ok := parsed["system"].([]any)
			require.True(t, ok, "system should be an array")
			require.Len(t, system, tt.wantSystemLen)

			first, ok := system[0].(map[string]any)
			require.True(t, ok)
			require.Equal(t, tt.wantFirstText, first["text"])
			require.Equal(t, "text", first["type"])

			// Check cache_control
			cc, ok := first["cache_control"].(map[string]any)
			require.True(t, ok)
			require.Equal(t, "ephemeral", cc["type"])

			if tt.wantSecondText != "" && len(system) > 1 {
				second, ok := system[1].(map[string]any)
				require.True(t, ok)
				require.Equal(t, tt.wantSecondText, second["text"])
			}
		})
	}
}

func TestRewriteSystemForNonClaudeCode(t *testing.T) {
	tests := []struct {
		name                   string
		body                   string
		system                 any
		wantSystemLen          int
		wantPreservedSystemRaw string
		wantMessagesLen        int
		wantFirstMessageText   string
	}{
		{
			name:            "nil system keeps only Claude Code block",
			body:            `{"model":"claude-3","messages":[{"role":"user","content":"hello"}]}`,
			system:          nil,
			wantSystemLen:   1,
			wantMessagesLen: 1,
		},
		{
			name:                 "custom string system migrates into messages",
			body:                 `{"model":"claude-3","messages":[{"role":"user","content":"hello"}]}`,
			system:               "You are a personal assistant running inside OpenClaw.",
			wantSystemLen:        1,
			wantMessagesLen:      3,
			wantFirstMessageText: "[System Instructions]\nYou are a personal assistant running inside OpenClaw.",
		},
		{
			name: "custom array system migrates but billing header stays in system",
			body: `{"model":"claude-3","messages":[{"role":"user","content":"hello"}]}`,
			system: []any{
				map[string]any{"type": "text", "text": "First instruction"},
				map[string]any{"type": "text", "text": "x-anthropic-billing-header: cc_version=2.1.87.449; cc_entrypoint=cli;"},
				map[string]any{"type": "text", "text": "Second instruction"},
			},
			wantSystemLen:          2,
			wantPreservedSystemRaw: "x-anthropic-billing-header: cc_version=2.1.87.449; cc_entrypoint=cli;",
			wantMessagesLen:        3,
			wantFirstMessageText:   "[System Instructions]\nFirst instruction\n\nSecond instruction",
		},
		{
			name:                 "existing Claude Code prompt is deduplicated while custom text migrates",
			body:                 `{"model":"claude-3","messages":[{"role":"user","content":"hello"}]}`,
			system:               []any{map[string]any{"type": "text", "text": claudeCodeSystemPrompt}, map[string]any{"type": "text", "text": "Custom prompt"}},
			wantSystemLen:        1,
			wantMessagesLen:      3,
			wantFirstMessageText: "[System Instructions]\nCustom prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriteSystemForNonClaudeCode([]byte(tt.body), tt.system)

			var parsed map[string]any
			err := json.Unmarshal(result, &parsed)
			require.NoError(t, err)

			system, ok := parsed["system"].([]any)
			require.True(t, ok, "system should be an array")
			require.Len(t, system, tt.wantSystemLen)

			first, ok := system[0].(map[string]any)
			require.True(t, ok)
			require.Equal(t, claudeCodeSystemPrompt, first["text"])

			if tt.wantPreservedSystemRaw != "" {
				require.Contains(t, gjson.GetBytes(result, "system").Raw, tt.wantPreservedSystemRaw)
			}

			messages, ok := parsed["messages"].([]any)
			require.True(t, ok, "messages should be an array")
			require.Len(t, messages, tt.wantMessagesLen)

			if tt.wantFirstMessageText == "" {
				return
			}

			firstMessage, ok := messages[0].(map[string]any)
			require.True(t, ok)
			require.Equal(t, "user", firstMessage["role"])

			firstContent, ok := firstMessage["content"].([]any)
			require.True(t, ok)
			require.Len(t, firstContent, 1)

			firstBlock, ok := firstContent[0].(map[string]any)
			require.True(t, ok)
			require.Equal(t, tt.wantFirstMessageText, firstBlock["text"])

			ackMessage, ok := messages[1].(map[string]any)
			require.True(t, ok)
			require.Equal(t, "assistant", ackMessage["role"])
			require.Contains(t, gjson.GetBytes(result, "messages.1.content.0.text").String(), "Understood. I will follow these instructions.")
		})
	}
}
