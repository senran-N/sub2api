package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeOpenAIMessagesDispatchModelConfig(t *testing.T) {
	t.Parallel()

	cfg := normalizeOpenAIMessagesDispatchModelConfig(OpenAIMessagesDispatchModelConfig{
		OpusMappedModel:   " gpt 5.4 extrahigh ",
		SonnetMappedModel: " gpt-5.3-codex ",
		HaikuMappedModel:  " gpt-5.4-mini ",
		ExactModelMappings: map[string]string{
			" claude-sonnet-4-5-20250929 ": " gpt-5.2 ",
			"":                             "gpt-5.4",
			"claude-opus-4-6":              " ",
		},
	})

	require.Equal(t, OpenAIMessagesDispatchModelConfig{
		OpusMappedModel:   "gpt-5.4-xhigh",
		SonnetMappedModel: "gpt-5.3-codex",
		HaikuMappedModel:  "gpt-5.4-mini",
		ExactModelMappings: map[string]string{
			"claude-sonnet-4-5-20250929": "gpt-5.2",
		},
	}, cfg)
}

func TestGroupResolveMessagesDispatchModel(t *testing.T) {
	t.Parallel()

	t.Run("openai groups keep dispatch mapping", func(t *testing.T) {
		group := &Group{
			Platform: PlatformOpenAI,
			MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
				OpusMappedModel:   "gpt-5.4",
				SonnetMappedModel: "gpt-5.2",
				ExactModelMappings: map[string]string{
					"claude-opus-4-6": "gpt-5.1-codex-max",
				},
			},
		}

		require.Equal(t, "gpt-5.1-codex-max", group.ResolveMessagesDispatchModel("claude-opus-4-6"))
		require.Equal(t, "gpt-5.2", group.ResolveMessagesDispatchModel("claude-sonnet-4-5-20250929"))
		require.Equal(t, defaultOpenAIMessagesDispatchHaikuMappedModel, group.ResolveMessagesDispatchModel("claude-haiku-4-5-20251001"))
		require.Empty(t, group.ResolveMessagesDispatchModel("gpt-5.4"))
	})

	t.Run("grok groups do not inherit openai dispatch mapping", func(t *testing.T) {
		group := &Group{
			Platform: PlatformGrok,
			MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
				ExactModelMappings: map[string]string{
					"claude-opus-4-6": "gpt-5.1-codex-max",
				},
			},
		}

		require.Empty(t, group.ResolveMessagesDispatchModel("claude-opus-4-6"))
	})

	require.Empty(t, (*Group)(nil).ResolveMessagesDispatchModel("claude-opus-4-6"))
}

func TestSanitizeGroupMessagesDispatchFields(t *testing.T) {
	t.Parallel()

	group := &Group{
		Platform:              PlatformAnthropic,
		AllowMessagesDispatch: true,
		DefaultMappedModel:    "gpt-5.4",
		MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
			OpusMappedModel: "gpt-5.2",
		},
	}

	sanitizeGroupMessagesDispatchFields(group)

	require.False(t, group.AllowMessagesDispatch)
	require.Empty(t, group.DefaultMappedModel)
	require.Equal(t, OpenAIMessagesDispatchModelConfig{}, group.MessagesDispatchModelConfig)
}
