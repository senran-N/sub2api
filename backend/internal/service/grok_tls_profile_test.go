//go:build unit

package service

import (
	"net/http"
	"testing"

	"github.com/senran-N/sub2api/internal/model"
	"github.com/stretchr/testify/require"
)

func TestAccountSupportsTLSFingerprint_GrokSession(t *testing.T) {
	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"enable_tls_fingerprint": true,
		},
	}

	require.True(t, account.SupportsTLSFingerprint())
	require.True(t, account.IsTLSFingerprintEnabled())
}

func TestTLSFingerprintProfileService_ResolveTLSProfile_GrokSession(t *testing.T) {
	svc := &TLSFingerprintProfileService{
		localCache: map[int64]*model.TLSFingerprintProfile{
			7: {ID: 7, Name: "grok-browser-profile"},
		},
	}
	account := &Account{
		ID:       601,
		Platform: PlatformGrok,
		Type:     AccountTypeSession,
		Extra: map[string]any{
			"enable_tls_fingerprint":     true,
			"tls_fingerprint_profile_id": 7,
		},
	}

	profile := svc.ResolveTLSProfile(account)
	require.NotNil(t, profile)
	require.Equal(t, "grok-browser-profile", profile.Name)
}

func TestProbeGrokSessionConnectionWithSettings_UsesTLSProfileWhenEnabled(t *testing.T) {
	upstream := &queuedHTTPUpstream{
		responses: []*http.Response{
			newJSONResponse(http.StatusOK, `{"conversationId":"conv_123"}`),
		},
	}
	account := &Account{
		ID:          602,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "grok-session-token",
		},
		Extra: map[string]any{
			"enable_tls_fingerprint": true,
		},
	}

	_, err := ProbeGrokSessionConnectionWithSettings(
		nil,
		upstream,
		account,
		"",
		"",
		DefaultGrokRuntimeSettings(),
	)
	require.NoError(t, err)
	require.Equal(t, []bool{true}, upstream.tlsFlags)
}
