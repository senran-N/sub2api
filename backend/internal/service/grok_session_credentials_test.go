package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeGrokSessionCookieHeader_BareTokenUsesSSOCookie(t *testing.T) {
	normalized, err := NormalizeGrokSessionCookieHeader("  raw-session-token  ")
	require.NoError(t, err)
	require.Equal(t, "sso=raw-session-token", normalized)
}

func TestNormalizeGrokSessionCookieHeader_CompactsCookiePairs(t *testing.T) {
	normalized, err := NormalizeGrokSessionCookieHeader("Cookie: sso=session-a ; sso-rw=session-b; invalid ; x-anonuserid = anon-1 ")
	require.NoError(t, err)
	require.Equal(t, "sso=session-a; sso-rw=session-b; x-anonuserid=anon-1", normalized)
}

func TestValidateGrokSessionImportToken_RequiresSSOCookie(t *testing.T) {
	_, err := ValidateGrokSessionImportToken("x-anonuserid=anon-only; sso-rw=rw-only")
	require.EqualError(t, err, "missing sso cookie")
}

func TestMaskGrokSessionFingerprint_RedactsFullDigest(t *testing.T) {
	fingerprint := FingerprintGrokSessionToken("sso=session-a; sso-rw=session-b")
	require.True(t, len(fingerprint) > len("sha256:")+12)
	require.NotEqual(t, fingerprint, MaskGrokSessionFingerprint(fingerprint))
	require.NotContains(t, MaskGrokSessionFingerprint(fingerprint), "session-a")
}
