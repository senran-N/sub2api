package service

import (
	"context"
	"testing"
	"time"
)

func TestCollectSelectionFailureStats(t *testing.T) {
	svc := &GatewayService{
		sessionLimitCache: &sessionLimitRuntimeStub{
			windowCost: 21,
			windowHit:  true,
		},
		rpmCache: &rpmCacheRuntimeLimitsStub{
			getCount: 12,
		},
	}
	model := "gpt-5.4"
	resetAt := time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339)

	accounts := []Account{
		{ID: 1, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true},
		{ID: 2, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: false},
		{ID: 3, Platform: PlatformGemini, Status: StatusActive, Schedulable: true},
		{ID: 4, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-4.1": "gpt-4.1"},
		}},
		{ID: 5, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Extra: map[string]any{
			"model_rate_limits": map[string]any{
				model: map[string]any{"rate_limit_reset_at": resetAt},
			},
		}},
		{ID: 6, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true},
		{ID: 7, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Extra: map[string]any{
			"quota_limit": 10,
			"quota_used":  10,
		}},
		{ID: 8, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Extra: map[string]any{
			"window_cost_limit": 10.0,
		}},
		{ID: 9, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Extra: map[string]any{
			"base_rpm":          10,
			"rpm_sticky_buffer": 2,
		}},
	}

	excluded := map[int64]struct{}{1: {}}
	stats := svc.collectSelectionFailureStats(context.Background(), accounts, model, PlatformOpenAI, excluded, false)
	if stats.Total != 9 ||
		stats.Excluded != 1 ||
		stats.Unschedulable != 1 ||
		stats.PlatformFiltered != 3 ||
		stats.ModelUnsupported != 1 ||
		stats.ModelRateLimited != 1 ||
		stats.QuotaLimited != 1 ||
		stats.WindowCostLimited != 0 ||
		stats.RPMLimited != 0 ||
		stats.Eligible != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestCollectSelectionFailureStats_TracksWindowCostAndRPM(t *testing.T) {
	svc := &GatewayService{
		sessionLimitCache: &sessionLimitRuntimeStub{
			windowCost: 21,
			windowHit:  true,
		},
		rpmCache: &rpmCacheRuntimeLimitsStub{
			getCount: 12,
		},
	}

	accounts := []Account{
		{ID: 8, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Extra: map[string]any{
			"window_cost_limit": 10.0,
		}},
		{ID: 9, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Extra: map[string]any{
			"base_rpm":          10,
			"rpm_sticky_buffer": 2,
		}},
	}

	stats := svc.collectSelectionFailureStats(context.Background(), accounts, "", PlatformAnthropic, nil, false)
	if stats.WindowCostLimited != 1 || stats.RPMLimited != 1 || stats.Eligible != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(stats.SampleWindowIDs) != 1 || stats.SampleWindowIDs[0] != 8 {
		t.Fatalf("unexpected window samples: %+v", stats.SampleWindowIDs)
	}
	if len(stats.SampleRPMIDs) != 1 || stats.SampleRPMIDs[0] != 9 {
		t.Fatalf("unexpected rpm samples: %+v", stats.SampleRPMIDs)
	}
}
