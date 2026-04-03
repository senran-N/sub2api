//go:build unit

package service

import (
	"context"
	"testing"
)

func TestPrepareGatewayUsageForBilling(t *testing.T) {
	result := &ForwardResult{
		Usage: ClaudeUsage{
			InputTokens:              11,
			CacheCreationInputTokens: 5,
		},
	}
	account := &Account{
		ID:       1,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"cache_ttl_override_enabled": true,
			"cache_ttl_override_target":  "1h",
		},
	}

	cacheTTLOverridden := prepareGatewayUsageForBilling(result, account, true)

	if !cacheTTLOverridden {
		t.Fatalf("expected cache ttl override to be detected")
	}
	if result.Usage.InputTokens != 0 {
		t.Fatalf("input_tokens=%d want=0", result.Usage.InputTokens)
	}
	if result.Usage.CacheReadInputTokens != 11 {
		t.Fatalf("cache_read_input_tokens=%d want=11", result.Usage.CacheReadInputTokens)
	}
	if result.Usage.CacheCreation1hTokens != 5 {
		t.Fatalf("cache_creation_1h_tokens=%d want=5", result.Usage.CacheCreation1hTokens)
	}
}

func TestBuildGatewayUsageLog(t *testing.T) {
	groupID := int64(9)
	subscriptionID := int64(44)
	result := buildGatewayUsageLog(context.Background(), usageLogBuildInput{
		Result: &ForwardResult{
			RequestID:     "req-1",
			Model:         "claude-sonnet-4",
			UpstreamModel: "claude-sonnet-4-20250514",
			MediaType:     "image",
			ImageSize:     "1024x1024",
			ImageCount:    2,
		},
		APIKey: &APIKey{
			ID:      2,
			GroupID: &groupID,
			Group:   &Group{SubscriptionType: SubscriptionTypeSubscription},
		},
		User:               &User{ID: 3},
		Account:            &Account{ID: 4},
		Subscription:       &UserSubscription{ID: subscriptionID},
		Cost:               &CostBreakdown{ActualCost: 1.2, TotalCost: 2.3},
		Multiplier:         1.5,
		CacheTTLOverridden: true,
		InboundEndpoint:    "/v1/messages",
		UpstreamEndpoint:   "/v1/upstream",
		UserAgent:          "ua",
		IPAddress:          "127.0.0.1",
		IncludeMediaType:   true,
	})

	if !result.IsSubscriptionBilling {
		t.Fatalf("expected subscription billing")
	}
	if result.RequestID != "req-1" {
		t.Fatalf("requestID=%s want=req-1", result.RequestID)
	}
	if result.UsageLog == nil {
		t.Fatalf("usage log should not be nil")
	}
	if result.UsageLog.MediaType == nil || *result.UsageLog.MediaType != "image" {
		t.Fatalf("mediaType=%v want=image", result.UsageLog.MediaType)
	}
	if result.UsageLog.SubscriptionID == nil || *result.UsageLog.SubscriptionID != subscriptionID {
		t.Fatalf("subscriptionID=%v want=%d", result.UsageLog.SubscriptionID, subscriptionID)
	}
	if result.UsageLog.GroupID == nil || *result.UsageLog.GroupID != groupID {
		t.Fatalf("groupID=%v want=%d", result.UsageLog.GroupID, groupID)
	}
	if result.UsageLog.UpstreamModel == nil || *result.UsageLog.UpstreamModel != "claude-sonnet-4-20250514" {
		t.Fatalf("upstreamModel=%v", result.UsageLog.UpstreamModel)
	}
}
