package service

import (
	"context"
	"testing"
	"time"
)

func TestAppendSelectionFailureSampleID_LimitsToFive(t *testing.T) {
	samples := []int64{1, 2, 3, 4, 5}

	got := appendSelectionFailureSampleID(samples, 6)

	if len(got) != 5 {
		t.Fatalf("len=%d want=5", len(got))
	}
	if got[4] != 5 {
		t.Fatalf("last=%d want=5", got[4])
	}
}

func TestAppendSelectionFailureRateSample_LimitsToFive(t *testing.T) {
	samples := []string{"1(1s)", "2(2s)", "3(3s)", "4(4s)", "5(5s)"}

	got := appendSelectionFailureRateSample(samples, 6, 6*time.Second)

	if len(got) != 5 {
		t.Fatalf("len=%d want=5", len(got))
	}
	if got[4] != "5(5s)" {
		t.Fatalf("last=%s want=5(5s)", got[4])
	}
}

func TestSummarizeSelectionFailureStats(t *testing.T) {
	stats := selectionFailureStats{
		Total:             7,
		Eligible:          1,
		Excluded:          2,
		Unschedulable:     3,
		PlatformFiltered:  4,
		ModelUnsupported:  5,
		ModelRateLimited:  6,
		QuotaLimited:      7,
		WindowCostLimited: 8,
		RPMLimited:        9,
	}

	got := summarizeSelectionFailureStats(stats)
	want := "total=7 eligible=1 excluded=2 unschedulable=3 platform_filtered=4 model_unsupported=5 model_rate_limited=6 quota_limited=7 window_cost_limited=8 rpm_limited=9"
	if got != want {
		t.Fatalf("summary=%q want=%q", got, want)
	}
}

func TestIsPlatformFilteredForSelection(t *testing.T) {
	tests := []struct {
		name                 string
		account              *Account
		platform             string
		allowMixedScheduling bool
		want                 bool
	}{
		{
			name:     "nil account is filtered",
			account:  nil,
			platform: PlatformAnthropic,
			want:     true,
		},
		{
			name: "blank platform keeps account when mixed scheduling disabled",
			account: &Account{
				Platform: PlatformAnthropic,
			},
			platform: " ",
			want:     false,
		},
		{
			name: "same platform is eligible",
			account: &Account{
				Platform: PlatformAnthropic,
			},
			platform: PlatformAnthropic,
			want:     false,
		},
		{
			name: "different platform is filtered",
			account: &Account{
				Platform: PlatformOpenAI,
			},
			platform: PlatformAnthropic,
			want:     true,
		},
		{
			name: "mixed scheduling antigravity enabled is not filtered",
			account: &Account{
				Platform: PlatformAntigravity,
				Extra: map[string]any{
					"mixed_scheduling": true,
				},
			},
			platform:             PlatformAnthropic,
			allowMixedScheduling: true,
			want:                 false,
		},
		{
			name: "mixed scheduling antigravity disabled is filtered",
			account: &Account{
				Platform: PlatformAntigravity,
				Extra: map[string]any{
					"mixed_scheduling": false,
				},
			},
			platform:             PlatformAnthropic,
			allowMixedScheduling: true,
			want:                 true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPlatformFilteredForSelection(tt.account, tt.platform, tt.allowMixedScheduling)
			if got != tt.want {
				t.Fatalf("filtered=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestDiagnoseSelectionFailure_OAuthCredentialIssue(t *testing.T) {
	svc := &GatewayService{}
	expiredAt := time.Now().Add(-time.Minute).Format(time.RFC3339)
	acc := &Account{
		ID:          11,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"access_token": "expired-token",
			"expires_at":   expiredAt,
		},
	}

	diagnosis := svc.diagnoseSelectionFailure(
		context.Background(),
		acc,
		"",
		PlatformAnthropic,
		nil,
		false,
	)

	if diagnosis.Category != "unschedulable" {
		t.Fatalf("category=%q want=unschedulable", diagnosis.Category)
	}
	if diagnosis.Detail != "oauth_access_token_expired" {
		t.Fatalf("detail=%q want=oauth_access_token_expired", diagnosis.Detail)
	}
}

func TestDiagnoseSelectionFailure_QuotaLimited(t *testing.T) {
	svc := &GatewayService{}
	acc := &Account{
		ID:          21,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			"quota_limit": 10,
			"quota_used":  10,
		},
	}

	diagnosis := svc.diagnoseSelectionFailure(
		context.Background(),
		acc,
		"",
		PlatformOpenAI,
		nil,
		false,
	)

	if diagnosis.Category != "quota_limited" {
		t.Fatalf("category=%q want=quota_limited", diagnosis.Category)
	}
	if diagnosis.Detail != "quota_exceeded" {
		t.Fatalf("detail=%q want=quota_exceeded", diagnosis.Detail)
	}
}

func TestDiagnoseSelectionFailure_WindowCostLimited(t *testing.T) {
	cache := &sessionLimitRuntimeStub{
		windowCost: 21,
		windowHit:  true,
	}
	svc := &GatewayService{sessionLimitCache: cache}
	acc := &Account{
		ID:          22,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			"window_cost_limit": 10.0,
		},
	}

	diagnosis := svc.diagnoseSelectionFailure(
		context.Background(),
		acc,
		"",
		PlatformAnthropic,
		nil,
		false,
	)

	if diagnosis.Category != "window_cost_limited" {
		t.Fatalf("category=%q want=window_cost_limited", diagnosis.Category)
	}
	if diagnosis.Detail != "window_cost_exceeded" {
		t.Fatalf("detail=%q want=window_cost_exceeded", diagnosis.Detail)
	}
}

func TestDiagnoseSelectionFailure_RPMLimited(t *testing.T) {
	svc := &GatewayService{
		rpmCache: &rpmCacheRuntimeLimitsStub{
			getCount: 12,
		},
	}
	acc := &Account{
		ID:          23,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			"base_rpm":          10,
			"rpm_sticky_buffer": 2,
		},
	}

	diagnosis := svc.diagnoseSelectionFailure(
		context.Background(),
		acc,
		"",
		PlatformAnthropic,
		nil,
		false,
	)

	if diagnosis.Category != "rpm_limited" {
		t.Fatalf("category=%q want=rpm_limited", diagnosis.Category)
	}
	if diagnosis.Detail != "rpm_exceeded" {
		t.Fatalf("detail=%q want=rpm_exceeded", diagnosis.Detail)
	}
}
