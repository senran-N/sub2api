package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGatewayService_BuildUpstreamRequest_OAuthMimicUsesMinimalStableDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	svc := &GatewayService{
		identityService: NewIdentityService(&identityCacheStub{}),
	}
	account := &Account{
		ID:       101,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
	}

	req, err := svc.buildUpstreamRequest(
		context.Background(),
		c,
		account,
		[]byte(`{"model":"claude-sonnet-4-5"}`),
		"oauth-token",
		"oauth",
		"claude-sonnet-4-5",
		true,
		true,
	)
	require.NoError(t, err)

	require.Equal(t, "Bearer oauth-token", getHeaderRaw(req.Header, "authorization"))
	require.Equal(t, "application/json", getHeaderRaw(req.Header, "accept"))
	require.Equal(t, "cli", getHeaderRaw(req.Header, "x-app"))
	require.Equal(t, "true", getHeaderRaw(req.Header, "anthropic-dangerous-direct-browser-access"))
	require.Equal(t, "0", getHeaderRaw(req.Header, "x-stainless-retry-count"))
	require.Equal(t, "600", getHeaderRaw(req.Header, "x-stainless-timeout"))
	require.Equal(t, "stream", getHeaderRaw(req.Header, "x-stainless-helper-method"))

	require.Empty(t, getHeaderRaw(req.Header, "user-agent"))
	require.Empty(t, getHeaderRaw(req.Header, "x-stainless-lang"))
	require.Empty(t, getHeaderRaw(req.Header, "x-stainless-package-version"))
	require.Empty(t, getHeaderRaw(req.Header, "x-stainless-os"))
	require.Empty(t, getHeaderRaw(req.Header, "x-stainless-arch"))
	require.Empty(t, getHeaderRaw(req.Header, "x-stainless-runtime"))
	require.Empty(t, getHeaderRaw(req.Header, "x-stainless-runtime-version"))
}

func TestGatewayService_BuildUpstreamRequest_OAuthMimicPreservesObservedFingerprintHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set("User-Agent", "claude-cli/2.1.88 (darwin; arm64)")
	c.Request.Header.Set("X-Stainless-Lang", "js")
	c.Request.Header.Set("X-Stainless-Package-Version", "0.92.0")
	c.Request.Header.Set("X-Stainless-OS", "Darwin")
	c.Request.Header.Set("X-Stainless-Arch", "arm64")
	c.Request.Header.Set("X-Stainless-Runtime", "node")
	c.Request.Header.Set("X-Stainless-Runtime-Version", "v24.15.0")

	svc := &GatewayService{
		identityService: NewIdentityService(&identityCacheStub{}),
	}
	account := &Account{
		ID:       102,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
	}

	req, err := svc.buildUpstreamRequest(
		context.Background(),
		c,
		account,
		[]byte(`{"model":"claude-sonnet-4-5"}`),
		"oauth-token",
		"oauth",
		"claude-sonnet-4-5",
		false,
		true,
	)
	require.NoError(t, err)

	require.Equal(t, "claude-cli/2.1.88 (darwin; arm64)", getHeaderRaw(req.Header, "user-agent"))
	require.Equal(t, "js", getHeaderRaw(req.Header, "x-stainless-lang"))
	require.Equal(t, "0.92.0", getHeaderRaw(req.Header, "x-stainless-package-version"))
	require.Equal(t, "Darwin", getHeaderRaw(req.Header, "x-stainless-os"))
	require.Equal(t, "arm64", getHeaderRaw(req.Header, "x-stainless-arch"))
	require.Equal(t, "node", getHeaderRaw(req.Header, "x-stainless-runtime"))
	require.Equal(t, "v24.15.0", getHeaderRaw(req.Header, "x-stainless-runtime-version"))
}

func TestGatewayService_BuildUpstreamRequest_OAuthInjectsClaudeCodeSessionHeaderFromMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	svc := &GatewayService{
		identityService: NewIdentityService(&identityCacheStub{}),
	}
	account := &Account{
		ID:       103,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"account_uuid": "acc-uuid",
		},
	}

	originalUserID := FormatMetadataUserID(
		"downstream-client",
		"",
		"7578cf37-aaca-46e4-a45c-71285d9dbb83",
		"2.1.78",
	)
	req, err := svc.buildUpstreamRequest(
		context.Background(),
		c,
		account,
		[]byte(`{"model":"claude-sonnet-4-5","metadata":{"user_id":`+strconv.Quote(originalUserID)+`}}`),
		"oauth-token",
		"oauth",
		"claude-sonnet-4-5",
		true,
		false,
	)
	require.NoError(t, err)

	bodyBytes := readRequestBodyForTest(t, req)
	rewritten := gjson.GetBytes(bodyBytes, "metadata.user_id").String()
	parsed := ParseMetadataUserID(rewritten)
	require.NotNil(t, parsed)
	require.Equal(t, parsed.SessionID, getHeaderRaw(req.Header, "X-Claude-Code-Session-Id"))
}

func TestGatewayService_BuildUpstreamRequest_OAuthOverridesExistingClaudeCodeSessionHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set("X-Claude-Code-Session-Id", "00000000-1111-4222-8333-444444444444")

	svc := &GatewayService{
		identityService: NewIdentityService(&identityCacheStub{}),
	}
	account := &Account{
		ID:       104,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"account_uuid": "acc-uuid",
		},
	}

	originalUserID := FormatMetadataUserID(
		"downstream-client",
		"",
		"7578cf37-aaca-46e4-a45c-71285d9dbb83",
		"2.1.78",
	)
	req, err := svc.buildUpstreamRequest(
		context.Background(),
		c,
		account,
		[]byte(`{"model":"claude-sonnet-4-5","metadata":{"user_id":`+strconv.Quote(originalUserID)+`}}`),
		"oauth-token",
		"oauth",
		"claude-sonnet-4-5",
		true,
		false,
	)
	require.NoError(t, err)

	bodyBytes := readRequestBodyForTest(t, req)
	rewritten := gjson.GetBytes(bodyBytes, "metadata.user_id").String()
	parsed := ParseMetadataUserID(rewritten)
	require.NotNil(t, parsed)
	require.NotEqual(t, "00000000-1111-4222-8333-444444444444", parsed.SessionID)
	require.Equal(t, parsed.SessionID, getHeaderRaw(req.Header, "X-Claude-Code-Session-Id"))
}

func TestIdentityService_RewriteUserID_EmptyFingerprintUAUsesNewMetadataFormat(t *testing.T) {
	cache := &identityCacheStub{}
	svc := NewIdentityService(cache)

	legacyUserID := FormatMetadataUserID(
		"d61f76d0730d2b920763648949bad5c79742155c27037fc77ac3f9805cb90169",
		"",
		"7578cf37-aaca-46e4-a45c-71285d9dbb83",
		"2.1.22",
	)
	body := []byte(`{"metadata":{"user_id":"` + legacyUserID + `"}}`)

	result, err := svc.RewriteUserID(body, 123, "acc-uuid", "client-xyz", "")
	require.NoError(t, err)

	rewritten := gjson.GetBytes(result, "metadata.user_id").String()
	parsed := ParseMetadataUserID(rewritten)
	require.NotNil(t, parsed)
	require.True(t, parsed.IsNewFormat)
	require.Equal(t, "client-xyz", parsed.DeviceID)
	require.Equal(t, "acc-uuid", parsed.AccountUUID)
}

func TestIsNewerVersion_TreatsObservedUAAsUpgradeFromEmptyCache(t *testing.T) {
	require.True(t, isNewerVersion("claude-cli/2.1.88 (darwin; arm64)", ""))
	require.False(t, isNewerVersion("Mozilla/5.0", ""))
}

func readRequestBodyForTest(t *testing.T, req *http.Request) []byte {
	t.Helper()
	require.NotNil(t, req)
	require.NotNil(t, req.Body)

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.NoError(t, req.Body.Close())
	req.Body = io.NopCloser(strings.NewReader(string(body)))
	return body
}
