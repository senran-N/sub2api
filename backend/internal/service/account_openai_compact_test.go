package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountOpenAICompactSupportModes(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	supported, known := account.OpenAICompactSupportKnown()
	require.False(t, supported)
	require.False(t, known)
	require.True(t, account.AllowsOpenAICompact())

	account.Extra = map[string]any{"openai_compact_supported": false}
	supported, known = account.OpenAICompactSupportKnown()
	require.False(t, supported)
	require.True(t, known)
	require.False(t, account.AllowsOpenAICompact())

	account.Extra["openai_compact_mode"] = OpenAICompactModeForceOn
	supported, known = account.OpenAICompactSupportKnown()
	require.True(t, supported)
	require.True(t, known)
	require.True(t, account.AllowsOpenAICompact())

	account.Extra["openai_compact_mode"] = OpenAICompactModeForceOff
	supported, known = account.OpenAICompactSupportKnown()
	require.False(t, supported)
	require.True(t, known)
	require.False(t, account.AllowsOpenAICompact())
}

func TestResolveOpenAICompactForwardModel(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"compact_model_mapping": map[string]any{
				"gpt-5.3-codex": "gpt-5.3-codex-compact",
				"gpt-5.4-*":     "gpt-5.4-compact",
			},
		},
	}

	require.Equal(t, "gpt-5.3-codex-compact", resolveOpenAICompactForwardModel(account, "gpt-5.3-codex"))
	require.Equal(t, "gpt-5.4-compact", resolveOpenAICompactForwardModel(account, "gpt-5.4-xhigh"))
	require.Equal(t, "gpt-5.2", resolveOpenAICompactForwardModel(account, "gpt-5.2"))
	require.Equal(t, "gpt-5.2", resolveOpenAICompactForwardModel(nil, "gpt-5.2"))
}
