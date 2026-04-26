package service

import (
	"testing"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestBillingServiceOpenAIGPT55FallbackUsesOfficialLongContextPricing(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)

	pricing, err := svc.GetModelPricing("gpt-5.5-2026-04-23")
	require.NoError(t, err)
	require.NotNil(t, pricing)
	require.InDelta(t, 5e-6, pricing.InputPricePerToken, 1e-12)
	require.InDelta(t, 30e-6, pricing.OutputPricePerToken, 1e-12)
	require.InDelta(t, 0.5e-6, pricing.CacheReadPricePerToken, 1e-12)
	require.Equal(t, 272000, pricing.LongContextInputThreshold)
	require.InDelta(t, 2.0, pricing.LongContextInputMultiplier, 1e-12)
	require.InDelta(t, 1.5, pricing.LongContextOutputMultiplier, 1e-12)

	tokens := UsageTokens{
		InputTokens:  300000,
		OutputTokens: 4000,
	}
	cost, err := svc.CalculateCost("gpt-5.5-2026-04-23", tokens, 1)
	require.NoError(t, err)
	require.InDelta(t, float64(tokens.InputTokens)*5e-6*2.0, cost.InputCost, 1e-10)
	require.InDelta(t, float64(tokens.OutputTokens)*30e-6*1.5, cost.OutputCost, 1e-10)
}
