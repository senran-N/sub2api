package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeGrokSessionCookieHeader_BareTokenUsesSSOCookie(t *testing.T) {
	normalized, err := NormalizeGrokSessionCookieHeader("  raw-session-token  ")
	require.NoError(t, err)
	require.Equal(t, "sso=raw-session-token; sso-rw=raw-session-token", normalized)
}

func TestNormalizeGrokSessionCookieHeader_CompactsCookiePairs(t *testing.T) {
	normalized, err := NormalizeGrokSessionCookieHeader("Cookie: sso=session-a ; sso-rw=session-b; invalid ; x-anonuserid = anon-1 ")
	require.NoError(t, err)
	require.Equal(t, "sso=session-a; sso-rw=session-b; x-anonuserid=anon-1", normalized)
}

func TestValidateGrokSessionImportToken_RequiresSSOCookie(t *testing.T) {
	normalized, err := ValidateGrokSessionImportToken("x-anonuserid=anon-only; sso-rw=rw-only")
	require.NoError(t, err)
	require.Equal(t, "sso=rw-only; sso-rw=rw-only; x-anonuserid=anon-only", normalized)
}

func TestMaskGrokSessionFingerprint_RedactsFullDigest(t *testing.T) {
	fingerprint := FingerprintGrokSessionToken("sso=session-a; sso-rw=session-b")
	require.True(t, len(fingerprint) > len("sha256:")+12)
	require.NotEqual(t, fingerprint, MaskGrokSessionFingerprint(fingerprint))
	require.NotContains(t, MaskGrokSessionFingerprint(fingerprint), "session-a")
}
