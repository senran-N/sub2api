//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/config"
)

func TestCalculateGatewayUsageCost_PromptMediaReturnsZeroCost(t *testing.T) {
	svc := &GatewayService{
		billingService: NewBillingService(&config.Config{}, nil),
	}

	cost := svc.calculateGatewayUsageCost(context.Background(), &ForwardResult{
		Model:     "sora-prompt",
		MediaType: "prompt",
	}, &APIKey{}, "", 1.2)

	if cost == nil {
		t.Fatalf("cost should not be nil")
	}
	if cost.TotalCost != 0 || cost.ActualCost != 0 {
		t.Fatalf("prompt cost should be zero, got total=%v actual=%v", cost.TotalCost, cost.ActualCost)
	}
}

func TestCalculateGatewayLongContextUsageCost_UsesLongContextPricing(t *testing.T) {
	svc := &GatewayService{
		billingService: NewBillingService(&config.Config{}, nil),
	}
	result := &ForwardResult{
		Model: "claude-sonnet-4",
		Usage: ClaudeUsage{
			InputTokens:  210000,
			OutputTokens: 1000,
		},
	}

	got := svc.calculateGatewayLongContextUsageCost(context.Background(), result, &APIKey{}, "", 1.0, 200000, 2.0)
	want, err := svc.billingService.CalculateCostWithLongContext("claude-sonnet-4", usageTokensFromClaudeUsage(result.Usage), 1.0, 200000, 2.0)
	if err != nil {
		t.Fatalf("unexpected pricing error: %v", err)
	}

	if got.TotalCost != want.TotalCost || got.ActualCost != want.ActualCost {
		t.Fatalf("got total=%v actual=%v want total=%v actual=%v", got.TotalCost, got.ActualCost, want.TotalCost, want.ActualCost)
	}
}
