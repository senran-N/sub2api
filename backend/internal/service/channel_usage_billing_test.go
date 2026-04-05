package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnrichChannelUsageFieldsResolvesMapping(t *testing.T) {
	const groupID int64 = 10
	channelService := newTestChannelService(makeStandardChannelRepo(Channel{
		ID:                 9,
		Name:               "mapped-channel",
		Status:             StatusActive,
		GroupIDs:           []int64{groupID},
		BillingModelSource: BillingModelSourceChannelMapped,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4-upstream"}},
		},
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-sonnet-4": "claude-sonnet-4-upstream",
			},
		},
	}, map[int64]string{groupID: PlatformAnthropic}))

	fields := enrichChannelUsageFields(
		context.Background(),
		channelService,
		testGatewayGroupIDPtr(groupID),
		"claude-sonnet-4",
		"claude-sonnet-4-20250514",
		ChannelUsageFields{},
	)

	require.Equal(t, int64(9), fields.ChannelID)
	require.Equal(t, "claude-sonnet-4", fields.OriginalModel)
	require.Equal(t, "claude-sonnet-4-upstream", fields.ChannelMappedModel)
	require.Equal(t, BillingModelSourceChannelMapped, fields.BillingModelSource)
	require.Equal(t, "claude-sonnet-4→claude-sonnet-4-upstream→claude-sonnet-4-20250514", fields.ModelMappingChain)
}

func TestEnrichChannelUsageFieldsPreservesExplicitValues(t *testing.T) {
	const groupID int64 = 11
	channelService := newTestChannelService(makeStandardChannelRepo(Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{groupID},
		BillingModelSource: BillingModelSourceRequested,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4"}},
		},
	}, map[int64]string{groupID: PlatformAnthropic}))

	original := ChannelUsageFields{
		ChannelID:          99,
		OriginalModel:      "custom-request",
		ChannelMappedModel: "custom-mapped",
		BillingModelSource: BillingModelSourceUpstream,
		ModelMappingChain:  "a→b→c",
	}

	fields := enrichChannelUsageFields(
		context.Background(),
		channelService,
		testGatewayGroupIDPtr(groupID),
		"claude-sonnet-4",
		"claude-upstream",
		original,
	)

	require.Equal(t, original, fields)
}

func TestEnrichChannelUsageFieldsDefaultsWithoutChannelLookup(t *testing.T) {
	fields := enrichChannelUsageFields(
		context.Background(),
		nil,
		nil,
		"claude-sonnet-4",
		"claude-upstream",
		ChannelUsageFields{},
	)

	require.Zero(t, fields.ChannelID)
	require.Equal(t, "claude-sonnet-4", fields.OriginalModel)
	require.Equal(t, "claude-sonnet-4", fields.ChannelMappedModel)
	require.Empty(t, fields.BillingModelSource)
	require.Empty(t, fields.ModelMappingChain)
}

func TestResolveChannelPricingReturnsChannelResolvedPricing(t *testing.T) {
	const groupID int64 = 100
	resolver := newResolverWithChannelPricing([]ChannelModelPricing{{
		Platform:        PlatformAnthropic,
		Models:          []string{"claude-sonnet-4"},
		BillingMode:     BillingModePerRequest,
		PerRequestPrice: testChannelFloat64Ptr(0.05),
	}})

	apiKey := &APIKey{
		Group: &Group{ID: groupID},
	}

	resolved := resolveChannelPricing(context.Background(), resolver, apiKey, "claude-sonnet-4")
	require.NotNil(t, resolved)
	require.Equal(t, PricingSourceChannel, resolved.Source)
	require.Equal(t, BillingModePerRequest, resolved.Mode)
	require.InDelta(t, 0.05, resolved.DefaultPerRequestPrice, 1e-12)
}

func TestResolveChannelPricingReturnsNilWhenChannelNotMatched(t *testing.T) {
	resolver := NewModelPricingResolver(newTestChannelService(&channelServiceRepoStub{
		listAllFn: func(context.Context) ([]Channel, error) {
			return []Channel{}, nil
		},
		getGroupPlatformsFn: func(context.Context, []int64) (map[int64]string, error) {
			return map[int64]string{}, nil
		},
	}), newTestBillingServiceForResolver())

	t.Run("nil resolver or api key", func(t *testing.T) {
		require.Nil(t, resolveChannelPricing(context.Background(), nil, &APIKey{Group: &Group{ID: 1}}, "claude-sonnet-4"))
		require.Nil(t, resolveChannelPricing(context.Background(), resolver, nil, "claude-sonnet-4"))
	})

	t.Run("missing group or model", func(t *testing.T) {
		require.Nil(t, resolveChannelPricing(context.Background(), resolver, &APIKey{}, "claude-sonnet-4"))
		require.Nil(t, resolveChannelPricing(context.Background(), resolver, &APIKey{Group: &Group{ID: 1}}, ""))
	})

	t.Run("base pricing only is not treated as channel pricing", func(t *testing.T) {
		resolved := resolveChannelPricing(context.Background(), resolver, &APIKey{Group: &Group{ID: 1}}, "claude-sonnet-4")
		require.Nil(t, resolved)
	})
}
