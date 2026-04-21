//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProbeGrokSessionConnectionWithSettings_UsesConfiguredSessionBaseURL(t *testing.T) {
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"conversationId":"conv_123"}`),
		},
	}
	account := &Account{
		ID:          1,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
	}

	result, err := ProbeGrokSessionConnectionWithSettings(
		context.Background(),
		upstream,
		account,
		"",
		"",
		GrokRuntimeSettings{SessionBaseURL: "https://session.grok.example/root"},
	)
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://session.grok.example/root/rest/app-chat/conversations/new", upstream.requests[0].URL.String())
	require.Equal(t, "https://session.grok.example/root", upstream.requests[0].Header.Get("Origin"))
	require.Equal(t, "https://session.grok.example/root/", upstream.requests[0].Header.Get("Referer"))
	require.Equal(t, grokSessionProbeAcceptHeader, upstream.requests[0].Header.Get("Accept"))
	require.Equal(t, "gzip, deflate, br, zstd", upstream.requests[0].Header.Get("Accept-Encoding"))
	require.Equal(t, grokSessionDefaultBaggage, upstream.requests[0].Header.Get("Baggage"))
	require.Equal(t, "https://session.grok.example/root/rest/app-chat/conversations/new", result.TargetURL)
	require.True(t, result.HasSSO)
	require.True(t, result.HasSSORW)
	require.False(t, requestBodyBool(t, upstream.requests[0], "disableSearch"))
	require.True(t, requestBodyBool(t, upstream.requests[0], "sendFinalMetadata"))
}

func requestBodyBool(t *testing.T, req *http.Request, key string) bool {
	t.Helper()
	require.NotNil(t, req)
	require.NotNil(t, req.Body)

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	value, ok := payload[key].(bool)
	require.True(t, ok)
	return value
}
