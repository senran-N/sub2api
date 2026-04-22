package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeGrokSessionCookieHeader_BareTokenUsesSSOCookie(t *testing.T) {
	normalized, err := NormalizeGrokSessionCookieHeader("  groksessiontoken1234567890abcd  ")
	require.NoError(t, err)
	require.Equal(t, "sso=groksessiontoken1234567890abcd; sso-rw=groksessiontoken1234567890abcd", normalized)
}

func TestNormalizeGrokSessionCookieHeader_CompactsCookiePairs(t *testing.T) {
	normalized, err := NormalizeGrokSessionCookieHeader("Cookie: sso=abcdefghijklmnopqrstuvwxyz123456 ; sso-rw=mnopqrstuvwxyzabcdef123456; invalid ; x-anonuserid = anon-1 ")
	require.NoError(t, err)
	require.Equal(t, "sso=abcdefghijklmnopqrstuvwxyz123456; sso-rw=mnopqrstuvwxyzabcdef123456; x-anonuserid=anon-1", normalized)
}

func TestValidateGrokSessionImportToken_RequiresSSOCookie(t *testing.T) {
	normalized, err := ValidateGrokSessionImportToken("x-anonuserid=anon-only; sso-rw=abcdefghijklmnopqrstuvwxyz123456")
	require.NoError(t, err)
	require.Equal(t, "sso=abcdefghijklmnopqrstuvwxyz123456; sso-rw=abcdefghijklmnopqrstuvwxyz123456; x-anonuserid=anon-only", normalized)
}

func TestValidateGrokSessionImportToken_RejectsShortBareToken(t *testing.T) {
	normalized, err := ValidateGrokSessionImportToken("short-token")
	require.ErrorContains(t, err, "grok session token format is invalid")
	require.Empty(t, normalized)
}

func TestMaskGrokSessionFingerprint_RedactsFullDigest(t *testing.T) {
	fingerprint := FingerprintGrokSessionToken("sso=abcdefghijklmnopqrstuvwxyz123456; sso-rw=mnopqrstuvwxyzabcdef123456")
	require.True(t, len(fingerprint) > len("sha256:")+12)
	require.NotEqual(t, fingerprint, MaskGrokSessionFingerprint(fingerprint))
	require.NotContains(t, MaskGrokSessionFingerprint(fingerprint), "abcdefghijklmnopqrstuvwxyz123456")
}
