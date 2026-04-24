package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type openAIModelFallbackSettingRepo struct {
	values map[string]string
}

func (r openAIModelFallbackSettingRepo) Get(ctx context.Context, key string) (*Setting, error) {
	return nil, ErrSettingNotFound
}

func (r openAIModelFallbackSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (r openAIModelFallbackSettingRepo) Set(ctx context.Context, key, value string) error {
	return nil
}

func (r openAIModelFallbackSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (r openAIModelFallbackSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	return nil
}

func (r openAIModelFallbackSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string, len(r.values))
	for key, value := range r.values {
		result[key] = value
	}
	return result, nil
}

func (r openAIModelFallbackSettingRepo) Delete(ctx context.Context, key string) error {
	return nil
}

func TestResolveOpenAIForwardModel(t *testing.T) {
	tests := []struct {
		name               string
		account            *Account
		requestedModel     string
		defaultMappedModel string
		expectedModel      string
	}{
		{
			name: "uses explicit fallback when account has no mapping",
			account: &Account{
				Credentials: map[string]any{},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-4o-mini",
		},
		{
			name: "preserves exact passthrough mapping instead of explicit fallback",
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
			name: "preserves wildcard passthrough mapping instead of explicit fallback",
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
			name:               "inherits explicit reasoning suffix from selected fallback",
			account:            &Account{Credentials: map[string]any{}},
			requestedModel:     "gpt-5.4-xhigh",
			defaultMappedModel: "gpt-5.2",
			expectedModel:      "gpt-5.2-xhigh",
		},
		{
			name:               "does not duplicate reasoning suffix without mapping or fallback",
			account:            &Account{Credentials: map[string]any{}},
			requestedModel:     "gpt-5.4-xhigh",
			defaultMappedModel: "",
			expectedModel:      "gpt-5.4-xhigh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedModel, resolveOpenAIForwardModel(tt.account, tt.requestedModel, tt.defaultMappedModel))
		})
	}
}

func TestResolveOpenAIForwardModel_PreventsClaudeModelFromFallingBackToLegacyDefaults(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{},
	}

	withoutDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "")
	require.Equal(t, "gpt-5.4", normalizeCodexModel(withoutDefault))

	withDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "gpt-5.4")
	require.Equal(t, "gpt-5.4", normalizeCodexModel(withDefault))
}

func TestNormalizeCodexModel_UpstreamAlignment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normalizes bare codex spark distinctly", input: "gpt-5.3-codex-spark", want: "gpt-5.3-codex-spark"},
		{name: "normalizes trimmed bare codex spark distinctly", input: "  gpt-5.3-codex-spark  ", want: "gpt-5.3-codex-spark"},
		{name: "normalizes spaced bare codex spark distinctly", input: "gpt 5.3 codex spark", want: "gpt-5.3-codex-spark"},
		{name: "normalizes spark high distinctly", input: "gpt-5.3-codex-spark-high", want: "gpt-5.3-codex-spark"},
		{name: "normalizes gpt 5.5 alias", input: "gpt 5.5", want: "gpt-5.5"},
		{name: "normalizes gpt 5.5 variant", input: "gpt-5.5-high", want: "gpt-5.5"},
		{name: "normalizes gpt 5.4 alias", input: "gpt 5.4", want: "gpt-5.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, normalizeCodexModel(tt.input))
		})
	}
}

func TestNormalizeOpenAIModelForUpstream(t *testing.T) {
	t.Run("oauth accounts normalize codex aliases", func(t *testing.T) {
		account := &Account{Type: AccountTypeOAuth}
		require.Equal(t, "gpt-5.4", normalizeOpenAIModelForUpstream(account, "gpt 5.4"))
	})

	t.Run("oauth accounts preserve explicit upstream variants", func(t *testing.T) {
		account := &Account{Type: AccountTypeOAuth}
		require.Equal(t, "gpt-5.4-xhigh", normalizeOpenAIModelForUpstream(account, "gpt-5.4-xhigh"))
		require.Equal(t, "gpt-5.3-codex-spark", normalizeOpenAIModelForUpstream(account, "gpt-5.3-codex-spark"))
	})

	t.Run("api key accounts preserve mapped custom model", func(t *testing.T) {
		account := &Account{Type: AccountTypeAPIKey}
		require.Equal(t, "custom/upstream-model", normalizeOpenAIModelForUpstream(account, " custom/upstream-model "))
	})

	t.Run("api key accounts preserve official non-codex models", func(t *testing.T) {
		account := &Account{Type: AccountTypeAPIKey}
		require.Equal(t, "gpt-4.1", normalizeOpenAIModelForUpstream(account, "gpt-4.1"))
	})

	t.Run("api key accounts preserve arbitrary compatible upstream models", func(t *testing.T) {
		account := &Account{Type: AccountTypeAPIKey}
		require.Equal(t, "gemini-3-flash-preview", normalizeOpenAIModelForUpstream(account, "gemini-3-flash-preview"))
	})
}

func TestResolveOpenAISelectionFallbackModelRequiresSetting(t *testing.T) {
	apiKey := &APIKey{
		Group: &Group{DefaultMappedModel: "gpt-5.4"},
	}

	disabledSvc := &OpenAIGatewayService{
		settingService: NewSettingService(openAIModelFallbackSettingRepo{values: map[string]string{
			SettingKeyEnableModelFallback: "false",
		}}, &config.Config{}),
	}
	model, ok := disabledSvc.ResolveOpenAISelectionFallbackModel(context.Background(), apiKey, "kimi-2.6")
	require.False(t, ok)
	require.Empty(t, model)

	enabledSvc := &OpenAIGatewayService{
		settingService: NewSettingService(openAIModelFallbackSettingRepo{values: map[string]string{
			SettingKeyEnableModelFallback: "true",
		}}, &config.Config{}),
	}
	model, ok = enabledSvc.ResolveOpenAISelectionFallbackModel(context.Background(), apiKey, "kimi-2.6")
	require.True(t, ok)
	require.Equal(t, "gpt-5.4", model)
}

func TestResolveOpenAIForwardDefaultMappedModelOnlyUsesExplicitFallback(t *testing.T) {
	apiKey := &APIKey{Group: &Group{DefaultMappedModel: "gpt-5.4"}}

	require.Empty(t, ResolveOpenAIForwardDefaultMappedModel(apiKey, ""))
	require.Equal(t, "gpt-5.2", ResolveOpenAIForwardDefaultMappedModel(apiKey, " gpt-5.2 "))
}
