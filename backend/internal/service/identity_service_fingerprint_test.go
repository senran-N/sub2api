package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIdentityService_GetOrCreateFingerprint_KeepsEstablishedPersona(t *testing.T) {
	cache := &identityCacheStub{
		fingerprint: &Fingerprint{
			ClientID:                "stable-client",
			UserAgent:               "claude-cli/2.1.78 (darwin; arm64)",
			StainlessLang:           "js",
			StainlessPackageVersion: "0.91.0",
			StainlessOS:             "Darwin",
			StainlessArch:           "arm64",
			StainlessRuntime:        "node",
			StainlessRuntimeVersion: "v22.14.0",
			UpdatedAt:               time.Now().Unix(),
		},
	}
	svc := NewIdentityService(cache)

	headers := http.Header{}
	headers.Set("User-Agent", "claude-cli/2.1.88 (darwin; arm64)")
	headers.Set("X-Stainless-Package-Version", "0.92.0")
	headers.Set("X-Stainless-Runtime-Version", "v24.15.0")

	fp, err := svc.GetOrCreateFingerprint(context.Background(), 123, headers)
	require.NoError(t, err)
	require.NotNil(t, fp)
	require.Equal(t, "stable-client", fp.ClientID)
	require.Equal(t, "claude-cli/2.1.78 (darwin; arm64)", fp.UserAgent)
	require.Equal(t, "0.91.0", fp.StainlessPackageVersion)
	require.Equal(t, "v22.14.0", fp.StainlessRuntimeVersion)
	require.Zero(t, cache.setFingerprintCount)
}

func TestIdentityService_GetOrCreateFingerprint_BootstrapsBlankPersonaFromObservedHeaders(t *testing.T) {
	cache := &identityCacheStub{
		fingerprint: &Fingerprint{
			ClientID:  "stable-client",
			UpdatedAt: time.Now().Unix(),
		},
	}
	svc := NewIdentityService(cache)

	headers := http.Header{}
	headers.Set("User-Agent", "claude-cli/2.1.88 (darwin; arm64)")
	headers.Set("X-Stainless-Lang", "js")
	headers.Set("X-Stainless-Package-Version", "0.92.0")
	headers.Set("X-Stainless-OS", "Darwin")
	headers.Set("X-Stainless-Arch", "arm64")
	headers.Set("X-Stainless-Runtime", "node")
	headers.Set("X-Stainless-Runtime-Version", "v24.15.0")

	fp, err := svc.GetOrCreateFingerprint(context.Background(), 123, headers)
	require.NoError(t, err)
	require.NotNil(t, fp)
	require.Equal(t, "stable-client", fp.ClientID)
	require.Equal(t, "claude-cli/2.1.88 (darwin; arm64)", fp.UserAgent)
	require.Equal(t, "0.92.0", fp.StainlessPackageVersion)
	require.Equal(t, "v24.15.0", fp.StainlessRuntimeVersion)
	require.Equal(t, 1, cache.setFingerprintCount)
}

func TestIdentityService_GetOrCreateFingerprint_RefreshesTTLWithoutChangingPersona(t *testing.T) {
	previousUpdatedAt := time.Now().Add(-25 * time.Hour).Unix()
	cache := &identityCacheStub{
		fingerprint: &Fingerprint{
			ClientID:  "stable-client",
			UserAgent: "claude-cli/2.1.78 (darwin; arm64)",
			UpdatedAt: previousUpdatedAt,
		},
	}
	svc := NewIdentityService(cache)

	headers := http.Header{}
	headers.Set("User-Agent", "claude-cli/2.1.88 (darwin; arm64)")

	fp, err := svc.GetOrCreateFingerprint(context.Background(), 123, headers)
	require.NoError(t, err)
	require.NotNil(t, fp)
	require.Equal(t, "claude-cli/2.1.78 (darwin; arm64)", fp.UserAgent)
	require.Equal(t, 1, cache.setFingerprintCount)
	require.Greater(t, fp.UpdatedAt, previousUpdatedAt)
}
