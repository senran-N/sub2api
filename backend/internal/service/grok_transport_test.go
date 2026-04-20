package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveGrokTransportTarget_SessionBareTokenBuildsCookieHeader(t *testing.T) {
	target, err := resolveGrokTransportTarget(&Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Credentials: map[string]any{
			"session_token": "sess-raw-token",
		},
	}, nil)

	require.NoError(t, err)
	require.Equal(t, grokTransportKindSession, target.Kind)
	require.Equal(t, "https://grok.com/rest/app-chat/conversations/new", target.URL)
	require.Equal(t, "sso=sess-raw-token; sso-rw=sess-raw-token", target.CookieHeader)
}

func TestResolveGrokTransportTargetWithSettings_SessionUsesConfiguredBaseURL(t *testing.T) {
	target, err := resolveGrokTransportTargetWithSettings(&Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Credentials: map[string]any{
			"session_token": "sess-raw-token",
		},
	}, nil, GrokRuntimeSettings{
		SessionBaseURL: "https://session.grok.example/root",
	})

	require.NoError(t, err)
	require.Equal(t, grokTransportKindSession, target.Kind)
	require.Equal(t, "https://session.grok.example/root/rest/app-chat/conversations/new", target.URL)
	require.Equal(t, "https://session.grok.example/root", target.SessionBaseURL)
	require.Equal(t, "sso=sess-raw-token; sso-rw=sess-raw-token", target.CookieHeader)
}

func TestResolveGrokTransportTarget_SessionCookieStringPreservesPairs(t *testing.T) {
	target, err := resolveGrokTransportTarget(&Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Credentials: map[string]any{
			"session_token": "Cookie: sso=session-a; sso-rw=session-b; x-anonuserid=anon-1",
		},
	}, nil)

	require.NoError(t, err)
	require.Equal(t, grokTransportKindSession, target.Kind)
	require.Equal(t, "sso=session-a; sso-rw=session-b; x-anonuserid=anon-1", target.CookieHeader)
}

func TestResolveGrokTransportTarget_SessionRequiresToken(t *testing.T) {
	_, err := resolveGrokTransportTarget(&Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
	}, nil)

	require.EqualError(t, err, "session_token not found in credentials")
}

func TestResolveGrokTransportTargetWithSettings_CompatibleUsesConfiguredOfficialBaseURL(t *testing.T) {
	target, err := resolveGrokTransportTargetWithSettings(&Account{
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "xai-test-key",
		},
	}, nil, GrokRuntimeSettings{
		OfficialBaseURL: "https://official.grok.example/api",
	})

	require.NoError(t, err)
	require.Equal(t, grokTransportKindCompatible, target.Kind)
	require.Equal(t, "https://official.grok.example/api/v1/responses", target.URL)
	require.Equal(t, "authorization", target.AuthHeader)
	require.Equal(t, "xai-test-key", target.AuthToken)
}

func TestResolveGrokMediaTransportTarget_SessionUsesMediaPath(t *testing.T) {
	target, err := resolveGrokMediaTransportTarget(&Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Credentials: map[string]any{
			"session_token": "sess-media-token",
		},
	}, nil, "/grok/v1/videos/job_123/content")

	require.NoError(t, err)
	require.Equal(t, grokTransportKindSession, target.Kind)
	require.Equal(t, "https://grok.com/v1/videos/job_123/content", target.URL)
	require.Equal(t, "sso=sess-media-token; sso-rw=sess-media-token", target.CookieHeader)
}

func TestResolveGrokMediaTransportTargetWithSettings_SessionUsesConfiguredBaseURL(t *testing.T) {
	target, err := resolveGrokMediaTransportTargetWithSettings(&Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Credentials: map[string]any{
			"session_token": "sess-media-token",
		},
	}, nil, GrokRuntimeSettings{
		SessionBaseURL: "https://session.grok.example/root",
	}, "/grok/v1/videos/job_123/content")

	require.NoError(t, err)
	require.Equal(t, grokTransportKindSession, target.Kind)
	require.Equal(t, "https://session.grok.example/root/v1/videos/job_123/content", target.URL)
	require.Equal(t, "https://session.grok.example/root", target.SessionBaseURL)
	require.Equal(t, "sso=sess-media-token; sso-rw=sess-media-token", target.CookieHeader)
}
