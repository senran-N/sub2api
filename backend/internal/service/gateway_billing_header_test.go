package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type gatewayForwardingRepoStub struct{}

func (s *gatewayForwardingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *gatewayForwardingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	return "", ErrSettingNotFound
}

func (s *gatewayForwardingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *gatewayForwardingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	return map[string]string{
		SettingKeyEnableFingerprintUnification: "true",
		SettingKeyEnableMetadataPassthrough:    "false",
		SettingKeyEnableCCHSigning:             "true",
	}, nil
}

func (s *gatewayForwardingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *gatewayForwardingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *gatewayForwardingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSyncBillingHeaderVersion(t *testing.T) {
	result := syncBillingHeaderVersion(
		[]byte(`{"system":[{"type":"text","text":"x-anthropic-billing-header: cc_version=2.1.81.df2; cc_entrypoint=cli; cch=00000;"}],"messages":[]}`),
		"claude-cli/2.1.22 (external, cli)",
	)
	require.Contains(t, string(result), "cc_version=2.1.22.df2")
	require.NotContains(t, string(result), "cc_version=2.1.81.df2")
}

func TestSignBillingHeaderCCH(t *testing.T) {
	body := []byte(`{"system":[{"type":"text","text":"x-anthropic-billing-header: cc_version=2.1.63.a43; cc_entrypoint=cli; cch=00000;"}],"messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]}`)
	expectedCCH := fmt.Sprintf("%05x", xxHash64Seeded(body, cchSeed)&0xFFFFF)

	result := signBillingHeaderCCH(body)
	billingText := gjson.GetBytes(result, "system.0.text").String()

	require.Contains(t, billingText, "cch="+expectedCCH+";")
	require.Contains(t, gjson.GetBytes(result, "messages.0.content.0.text").String(), "hello")
}

func TestGatewayService_BuildUpstreamRequest_OAuthSyncsBillingHeaderAndSignsCCH(t *testing.T) {
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
		settingService:  NewSettingService(&gatewayForwardingRepoStub{}, &config.Config{}),
	}
	account := &Account{
		ID:       106,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
	}

	req, err := svc.buildUpstreamRequest(
		context.Background(),
		c,
		account,
		[]byte(`{"model":"claude-sonnet-4-5","system":[{"type":"text","text":"x-anthropic-billing-header: cc_version=2.1.63.a43; cc_entrypoint=cli; cch=00000;"}],"messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]}`),
		"oauth-token",
		"oauth",
		"claude-sonnet-4-5",
		false,
		true,
	)
	require.NoError(t, err)

	bodyBytes := readRequestBodyForTest(t, req)
	billingText := gjson.GetBytes(bodyBytes, "system.0.text").String()
	require.Contains(t, billingText, "cc_version=2.1.88.a43")
	require.NotContains(t, billingText, "cch=00000")
}
