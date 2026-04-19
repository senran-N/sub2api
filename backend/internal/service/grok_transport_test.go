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
	require.Equal(t, "sso=sess-raw-token", target.CookieHeader)
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
