//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newBalanceNotifyServiceForUnitTest() (*BalanceNotifyService, *mockSettingRepo) {
	repo := newMockSettingRepo()
	emailService := NewEmailService(repo, nil)
	return NewBalanceNotifyService(emailService, repo, nil), repo
}

func TestResolveBalanceThreshold(t *testing.T) {
	require.Equal(t, 10.0, resolveBalanceThreshold(10, thresholdTypeFixed, 1000))
	require.Equal(t, 100.0, resolveBalanceThreshold(10, thresholdTypePercentage, 1000))
	require.Equal(t, 10.0, resolveBalanceThreshold(10, thresholdTypePercentage, 0))
}

func TestQuotaDimResolvedThreshold(t *testing.T) {
	require.Equal(t, 900.0, (quotaDim{threshold: 100, thresholdType: thresholdTypeFixed, limit: 1000}).resolvedThreshold())
	require.Equal(t, 700.0, (quotaDim{threshold: 30, thresholdType: thresholdTypePercentage, limit: 1000}).resolvedThreshold())
	require.Equal(t, 0.0, (quotaDim{threshold: 30, thresholdType: thresholdTypeFixed, limit: 0}).resolvedThreshold())
}

func TestBuildQuotaDims(t *testing.T) {
	a := &Account{
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			"quota_notify_daily_enabled":         true,
			"quota_notify_daily_threshold":       100.0,
			"quota_notify_daily_threshold_type":  thresholdTypeFixed,
			"quota_notify_weekly_enabled":        true,
			"quota_notify_weekly_threshold":      20.0,
			"quota_notify_weekly_threshold_type": thresholdTypePercentage,
			"quota_daily_limit":                  500.0,
			"quota_weekly_limit":                 2000.0,
			"quota_limit":                        10000.0,
			"quota_daily_used":                   50.0,
			"quota_weekly_used":                  300.0,
			"quota_used":                         1000.0,
		},
	}

	dims := buildQuotaDims(a)
	require.Len(t, dims, 3)
	require.Equal(t, quotaDimDaily, dims[0].name)
	require.True(t, dims[0].enabled)
	require.Equal(t, 100.0, dims[0].threshold)
	require.Equal(t, thresholdTypeFixed, dims[0].thresholdType)
	require.Equal(t, 500.0, dims[0].limit)
	require.Equal(t, 50.0, dims[0].currentUsed)
	require.Equal(t, quotaDimWeekly, dims[1].name)
	require.True(t, dims[1].enabled)
	require.Equal(t, thresholdTypePercentage, dims[1].thresholdType)
	require.Equal(t, quotaDimTotal, dims[2].name)
	require.Equal(t, 10000.0, dims[2].limit)
}

func TestFilterVerifiedEmails_DeduplicatesAndSkipsInvalid(t *testing.T) {
	got := filterVerifiedEmails([]NotifyEmailEntry{
		{Email: "User@Example.com", Verified: true},
		{Email: "user@example.com", Verified: true},
		{Email: "disabled@example.com", Verified: true, Disabled: true},
		{Email: "unverified@example.com", Verified: false},
		{Email: " ", Verified: true},
	})
	require.Equal(t, []string{"User@Example.com"}, got)
}

func TestGetBalanceNotifyConfig(t *testing.T) {
	svc, repo := newBalanceNotifyServiceForUnitTest()
	repo.data[SettingKeyBalanceLowNotifyEnabled] = "true"
	repo.data[SettingKeyBalanceLowNotifyThreshold] = "12.5"
	repo.data[SettingKeyBalanceLowNotifyRechargeURL] = "https://example.com/pay"

	enabled, threshold, rechargeURL := svc.getBalanceNotifyConfig(context.Background())
	require.True(t, enabled)
	require.Equal(t, 12.5, threshold)
	require.Equal(t, "https://example.com/pay", rechargeURL)
}

func TestBuildBalanceLowEmailBody(t *testing.T) {
	svc := &BalanceNotifyService{}
	body := svc.buildBalanceLowEmailBody("Alice", 3.14, 10, "MySite", "https://example.com/pay?a=1&b=<script>")
	require.Contains(t, body, "MySite")
	require.Contains(t, body, "Alice")
	require.Contains(t, body, "$3.14")
	require.Contains(t, body, "$10.00")
	require.Contains(t, body, `href="https://example.com/pay?a=1&amp;b=&lt;script&gt;"`)
	require.NotContains(t, body, "%!")
}

func TestBuildQuotaAlertEmailBody(t *testing.T) {
	svc := &BalanceNotifyService{}
	body := svc.buildQuotaAlertEmailBody(42, "acc-foo", "anthropic", "日限额 / Daily", 750.5, 1000, 249.5, "$249.50", "MySite")
	require.Contains(t, body, "MySite")
	require.Contains(t, body, "#42")
	require.Contains(t, body, "acc-foo")
	require.Contains(t, body, "anthropic")
	require.Contains(t, body, "$750.50")
	require.Contains(t, body, "$1000.00")
	require.Contains(t, body, "$249.50")
	require.NotContains(t, body, "%!")
}

func TestQuotaAlertEmailBodyUnlimited(t *testing.T) {
	svc := &BalanceNotifyService{}
	body := svc.buildQuotaAlertEmailBody(1, "n", "p", "dim", 100, 0, 0, "30%", "Site")
	require.True(t, strings.Contains(body, "Unlimited") || strings.Contains(body, "无限制"))
}

func TestCheckBalanceAfterDeductionGuards(t *testing.T) {
	svc, repo := newBalanceNotifyServiceForUnitTest()
	repo.data[SettingKeyBalanceLowNotifyEnabled] = "true"
	repo.data[SettingKeyBalanceLowNotifyThreshold] = "10"
	svc.CheckBalanceAfterDeduction(context.Background(), nil, 20, 15)
	svc.CheckBalanceAfterDeduction(context.Background(), &User{BalanceNotifyEnabled: false}, 20, 15)
	svc.CheckBalanceAfterDeduction(context.Background(), &User{BalanceNotifyEnabled: true}, 20, 0)
}
