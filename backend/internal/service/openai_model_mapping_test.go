package service

import "testing"

func TestResolveOpenAIForwardModel(t *testing.T) {
	tests := []struct {
		name               string
		account            *Account
		requestedModel     string
		defaultMappedModel string
		expectedModel      string
	}{
		{
			name: "falls back to group default when account has no mapping",
			account: &Account{
				Credentials: map[string]any{},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-4o-mini",
		},
		{
			name: "preserves exact passthrough mapping instead of group default",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5.4": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "preserves wildcard passthrough mapping instead of group default",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-*": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "uses account remap when explicit target differs",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "inherits explicit reasoning suffix from base passthrough mapping",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5.4": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5.4-xhigh",
			defaultMappedModel: "gpt-5.1",
			expectedModel:      "gpt-5.4-xhigh",
		},
		{
			name: "inherits explicit reasoning suffix from base remap target",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5.4": "gpt-5.3-codex-spark",
					},
				},
			},
			requestedModel:     "gpt-5.4-xhigh",
			defaultMappedModel: "gpt-5.1",
			expectedModel:      "gpt-5.3-codex-spark-xhigh",
		},
		{
			name:               "inherits explicit reasoning suffix from group default fallback",
			account:            &Account{Credentials: map[string]any{}},
			requestedModel:     "gpt-5.4-xhigh",
			defaultMappedModel: "gpt-5.2",
			expectedModel:      "gpt-5.2-xhigh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveOpenAIForwardModel(tt.account, tt.requestedModel, tt.defaultMappedModel); got != tt.expectedModel {
				t.Fatalf("resolveOpenAIForwardModel(...) = %q, want %q", got, tt.expectedModel)
			}
		})
	}
}

func TestResolveOpenAIForwardModel_PreventsClaudeModelFromFallingBackToGpt51(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{},
	}

	withoutDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "")
	if got := normalizeCodexModel(withoutDefault); got != "gpt-5.1" {
		t.Fatalf("normalizeCodexModel(%q) = %q, want %q", withoutDefault, got, "gpt-5.1")
	}

	withDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "gpt-5.4")
	if got := normalizeCodexModel(withDefault); got != "gpt-5.4" {
		t.Fatalf("normalizeCodexModel(%q) = %q, want %q", withDefault, got, "gpt-5.4")
	}
}

func TestNormalizeCodexModel_UpstreamAlignment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normalizes bare codex spark to codex", input: "gpt-5.3-codex-spark", want: "gpt-5.3-codex"},
		{name: "normalizes trimmed bare codex spark to codex", input: "  gpt-5.3-codex-spark  ", want: "gpt-5.3-codex"},
		{name: "normalizes spaced bare codex spark to codex", input: "gpt 5.3 codex spark", want: "gpt-5.3-codex"},
		{name: "normalizes spark high to codex", input: "gpt-5.3-codex-spark-high", want: "gpt-5.3-codex"},
		{name: "normalizes gpt 5.4 alias", input: "gpt 5.4", want: "gpt-5.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeCodexModel(tt.input); got != tt.want {
				t.Fatalf("normalizeCodexModel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeOpenAIModelForUpstream(t *testing.T) {
	t.Run("oauth accounts normalize codex aliases", func(t *testing.T) {
		account := &Account{Type: AccountTypeOAuth}
		if got := normalizeOpenAIModelForUpstream(account, "gpt 5.4"); got != "gpt-5.4" {
			t.Fatalf("normalizeOpenAIModelForUpstream(...) = %q, want %q", got, "gpt-5.4")
		}
	})

	t.Run("oauth accounts preserve explicit upstream variants", func(t *testing.T) {
		account := &Account{Type: AccountTypeOAuth}
		if got := normalizeOpenAIModelForUpstream(account, "gpt-5.4-xhigh"); got != "gpt-5.4-xhigh" {
			t.Fatalf("normalizeOpenAIModelForUpstream(...) = %q, want %q", got, "gpt-5.4-xhigh")
		}
		if got := normalizeOpenAIModelForUpstream(account, "gpt-5.3-codex-spark"); got != "gpt-5.3-codex-spark" {
			t.Fatalf("normalizeOpenAIModelForUpstream(...) = %q, want %q", got, "gpt-5.3-codex-spark")
		}
	})

	t.Run("api key accounts preserve mapped custom model", func(t *testing.T) {
		account := &Account{Type: AccountTypeAPIKey}
		if got := normalizeOpenAIModelForUpstream(account, " custom/upstream-model "); got != "custom/upstream-model" {
			t.Fatalf("normalizeOpenAIModelForUpstream(...) = %q, want %q", got, "custom/upstream-model")
		}
	})
}
