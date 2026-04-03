package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestEnforceCacheControlLimit_RemovesIllegalThinkingCacheControl(t *testing.T) {
	body := []byte(`{
		"system":[
			{"type":"thinking","text":"reason","cache_control":{"type":"ephemeral"}},
			{"type":"text","text":"safe","cache_control":{"type":"ephemeral"}}
		],
		"messages":[
			{"role":"user","content":[
				{"type":"thinking","text":"step","cache_control":{"type":"ephemeral"}},
				{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}
			]}
		]
	}`)

	result := enforceCacheControlLimit(body)

	require.False(t, gjson.GetBytes(result, "system.0.cache_control").Exists())
	require.True(t, gjson.GetBytes(result, "system.1.cache_control").Exists())
	require.False(t, gjson.GetBytes(result, "messages.0.content.0.cache_control").Exists())
	require.True(t, gjson.GetBytes(result, "messages.0.content.1.cache_control").Exists())
}
