//go:build unit

package service

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/model"
	"github.com/stretchr/testify/require"
)

func TestAccountTestService_GrokSessionChallengeReturnsSanitizedErrorWithoutAuthFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, recorder := newAccountTestContext()

	resp := newJSONResponse(http.StatusForbidden, `<!DOCTYPE html><html><head><title>Just a moment...</title></head><body>Enable JavaScript and cookies to continue<script>window._cf_chl_opt={};</script></body></html>`)
	resp.Header.Set("content-type", "text/html; charset=UTF-8")
	resp.Header.Set("cf-ray", "account-test-ray-1")

	repo := &openAIAccountTestRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}
	account := &Account{
		ID:          901,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "test-session-token",
		},
	}

	err := svc.testGrokAccountConnection(ctx, account, "grok-3-fast", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Cloudflare challenge encountered")
	require.NotContains(t, err.Error(), "<!DOCTYPE html>")
	require.Zero(t, repo.setErrorID)

	_, errMsg := parseTestSSEOutput(recorder.Body.String())
	require.Contains(t, errMsg, "Cloudflare challenge encountered")
	require.NotContains(t, errMsg, "<!DOCTYPE html>")

	grokExtra := grokExtraMap(repo.updatedExtra)
	require.Contains(t, getStringFromMaps(grokNestedMap(grokExtra["sync_state"]), nil, "last_probe_error"), "Cloudflare challenge encountered")
	require.NotContains(t, getStringFromMaps(grokNestedMap(grokExtra["sync_state"]), nil, "last_probe_error"), "<!DOCTYPE html>")
}

func TestAccountTestService_GrokSessionUsesTLSFingerprintWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newAccountTestContext()

	resp := newJSONResponse(http.StatusOK, "")
	resp.Body = io.NopCloser(strings.NewReader("data: {\"type\":\"response.output_text.delta\",\"delta\":\"ok\"}\n\ndata: [DONE]\n\n"))

	repo := &openAIAccountTestRepo{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	tlsSvc := &TLSFingerprintProfileService{
		localCache: map[int64]*model.TLSFingerprintProfile{
			13: {ID: 13, Name: "grok-test-profile"},
		},
	}
	svc := &AccountTestService{
		accountRepo:             repo,
		httpUpstream:            upstream,
		tlsFPProfileService:     tlsSvc,
		grokAccountStateService: NewGrokAccountStateService(repo),
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
	}
	account := &Account{
		ID:          902,
		Platform:    PlatformGrok,
		Type:        AccountTypeSession,
		Concurrency: 1,
		Credentials: map[string]any{
			"session_token": "test-session-token",
		},
		Extra: map[string]any{
			"enable_tls_fingerprint":     true,
			"tls_fingerprint_profile_id": 13,
		},
	}

	err := svc.testGrokSessionConnection(ctx, account, "grok-3-fast", "", mustResolveGrokSessionTarget(t, account))
	require.NoError(t, err)
	require.Equal(t, []bool{true}, upstream.tlsFlags)
}

func mustResolveGrokSessionTarget(t *testing.T, account *Account) grokTransportTarget {
	t.Helper()
	target, err := resolveGrokTransportTargetWithSettings(account, nil, DefaultGrokRuntimeSettings())
	require.NoError(t, err)
	return target
}
