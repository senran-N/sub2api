package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testChannelFloat64Ptr(v float64) *float64 { return &v }
func testChannelIntPtr(v int) *int             { return &v }

func TestChannelGetModelPricing(t *testing.T) {
	ch := &Channel{
		ModelPricing: []ChannelModelPricing{
			{ID: 1, Models: []string{"claude-sonnet-4"}, BillingMode: BillingModeToken, InputPrice: testChannelFloat64Ptr(3e-6)},
			{ID: 2, Models: []string{"gpt-5.1"}, BillingMode: BillingModePerRequest},
		},
	}

	require.Equal(t, int64(1), ch.GetModelPricing("claude-sonnet-4").ID)
	require.Equal(t, int64(1), ch.GetModelPricing("Claude-Sonnet-4").ID)
	require.Equal(t, int64(2), ch.GetModelPricing("gpt-5.1").ID)
	require.Nil(t, ch.GetModelPricing("claude-opus-4"))
}

func TestChannelGetModelPricingReturnsCopy(t *testing.T) {
	ch := &Channel{
		ModelPricing: []ChannelModelPricing{
			{ID: 1, Models: []string{"claude-sonnet-4"}},
		},
	}

	result := ch.GetModelPricing("claude-sonnet-4")
	require.NotNil(t, result)

	result.Models[0] = "hacked"
	require.Equal(t, "claude-sonnet-4", ch.ModelPricing[0].Models[0])
}

func TestFindMatchingIntervalBoundaries(t *testing.T) {
	intervals := []PricingInterval{
		{MinTokens: 0, MaxTokens: testChannelIntPtr(128000), InputPrice: testChannelFloat64Ptr(1e-6)},
		{MinTokens: 128000, MaxTokens: nil, InputPrice: testChannelFloat64Ptr(2e-6)},
	}

	require.Nil(t, FindMatchingInterval(intervals, 0))
	require.NotNil(t, FindMatchingInterval(intervals, 128000))
	require.InDelta(t, 1e-6, *FindMatchingInterval(intervals, 128000).InputPrice, 1e-12)
	require.NotNil(t, FindMatchingInterval(intervals, 128001))
	require.InDelta(t, 2e-6, *FindMatchingInterval(intervals, 128001).InputPrice, 1e-12)
}

func TestChannelMappingResultBuildModelMappingChain(t *testing.T) {
	tests := []struct {
		name          string
		result        ChannelMappingResult
		requestModel  string
		upstreamModel string
		want          string
	}{
		{
			name:          "no mapping same upstream",
			result:        ChannelMappingResult{Mapped: false, MappedModel: "claude-sonnet-4"},
			requestModel:  "claude-sonnet-4",
			upstreamModel: "claude-sonnet-4",
			want:          "",
		},
		{
			name:          "no mapping upstream differs",
			result:        ChannelMappingResult{Mapped: false, MappedModel: "claude-sonnet-4"},
			requestModel:  "claude-sonnet-4",
			upstreamModel: "claude-sonnet-4-20250514",
			want:          "claude-sonnet-4→claude-sonnet-4-20250514",
		},
		{
			name:          "mapped and upstream differs",
			result:        ChannelMappingResult{Mapped: true, MappedModel: "mapped-model"},
			requestModel:  "req-model",
			upstreamModel: "upstream-model",
			want:          "req-model→mapped-model→upstream-model",
		},
		{
			name:          "mapped and upstream same",
			result:        ChannelMappingResult{Mapped: true, MappedModel: "mapped-model"},
			requestModel:  "req-model",
			upstreamModel: "mapped-model",
			want:          "req-model→mapped-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.result.BuildModelMappingChain(tt.requestModel, tt.upstreamModel))
		})
	}
}

func TestReplaceModelInBody(t *testing.T) {
	unchanged := []byte(`{"model":"claude-sonnet-4","temperature":0.7}`)
	require.Equal(t, unchanged, ReplaceModelInBody(unchanged, "claude-sonnet-4"))

	replaced := ReplaceModelInBody(unchanged, "claude-opus-4")
	require.Contains(t, string(replaced), `"model":"claude-opus-4"`)
	require.Contains(t, string(replaced), `"temperature":0.7`)

	added := ReplaceModelInBody([]byte(`{"temperature":0.7}`), "claude-opus-4")
	require.Contains(t, string(added), `"model":"claude-opus-4"`)
	require.Contains(t, string(added), `"temperature":0.7`)
}
