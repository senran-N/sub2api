package service

import (
	"context"
	"testing"
	"time"
)

func TestCollectSelectionFailureStats(t *testing.T) {
	svc := &GatewayService{}
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
	}

	excluded := map[int64]struct{}{1: {}}
	stats := svc.collectSelectionFailureStats(context.Background(), accounts, model, PlatformOpenAI, excluded, false)
	if stats.Total != 6 || stats.Excluded != 1 || stats.Unschedulable != 1 || stats.PlatformFiltered != 1 || stats.ModelUnsupported != 1 || stats.ModelRateLimited != 1 || stats.Eligible != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
