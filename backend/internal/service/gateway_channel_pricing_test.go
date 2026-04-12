package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testGatewayGroupIDPtr(v int64) *int64 { return &v }

func newGatewayServiceWithChannelRepo(repo *channelServiceRepoStub) *GatewayService {
	return &GatewayService{channelService: newTestChannelService(repo)}
}

func TestBillingModelForRestriction(t *testing.T) {
	tests := []struct {
		name               string
		source             string
		requestedModel     string
		channelMappedModel string
		want               string
	}{
		{
			name:               "requested source uses requested model",
			source:             BillingModelSourceRequested,
			requestedModel:     "claude-sonnet-4",
			channelMappedModel: "mapped-claude",
			want:               "claude-sonnet-4",
		},
		{
			name:               "channel mapped source uses mapped model",
			source:             BillingModelSourceChannelMapped,
			requestedModel:     "claude-sonnet-4",
			channelMappedModel: "mapped-claude",
			want:               "mapped-claude",
		},
		{
			name:               "upstream source defers restriction",
			source:             BillingModelSourceUpstream,
			requestedModel:     "claude-sonnet-4",
			channelMappedModel: "mapped-claude",
			want:               "",
		},
		{
			name:               "empty source falls back to mapped model",
			source:             "",
			requestedModel:     "claude-sonnet-4",
			channelMappedModel: "mapped-claude",
			want:               "mapped-claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := billingModelForRestriction(tt.source, tt.requestedModel, tt.channelMappedModel)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestResolveAccountUpstreamModel(t *testing.T) {
	t.Run("nil account", func(t *testing.T) {
		require.Equal(t, "", resolveAccountUpstreamModel(nil, "claude-sonnet-4"))
	})

	t.Run("standard account uses direct mapping", func(t *testing.T) {
		account := &Account{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-4o": "gpt-4.1",
				},
			},
		}

		require.Equal(t, "gpt-4.1", resolveAccountUpstreamModel(account, "gpt-4o"))
		require.Equal(t, "gpt-5", resolveAccountUpstreamModel(account, "gpt-5"))
	})

	t.Run("openai reasoning variant keeps suffix after base fallback", func(t *testing.T) {
		account := &Account{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.4": "gpt-5.3-codex-spark",
				},
			},
		}

		require.Equal(t, "gpt-5.3-codex-spark-xhigh", resolveAccountUpstreamModel(account, "gpt-5.4-xhigh"))
	})

	t.Run("antigravity uses antigravity resolver", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAntigravity,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-sonnet-4-5": "claude-sonnet-4-5-thinking",
				},
			},
		}

		require.Equal(t, "claude-sonnet-4-5-thinking", resolveAccountUpstreamModel(account, "claude-sonnet-4-5"))
		require.Equal(t, "", resolveAccountUpstreamModel(account, "unsupported-model"))
	})
}

func TestGatewayServiceCheckChannelPricingRestriction(t *testing.T) {
	const groupID int64 = 10
	svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
		ID:                 1,
		Name:               "test-channel",
		Status:             StatusActive,
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceChannelMapped,
		GroupIDs:           []int64{groupID},
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"mapped-claude"}},
			{Platform: PlatformAnthropic, Models: []string{"claude-requested"}},
		},
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-sonnet-4": "mapped-claude",
				"claude-opus-4":   "mapped-missing",
			},
		},
	}, map[int64]string{groupID: PlatformAnthropic}))

	require.False(t, svc.checkChannelPricingRestriction(context.Background(), testGatewayGroupIDPtr(groupID), "claude-sonnet-4"))
	require.True(t, svc.checkChannelPricingRestriction(context.Background(), testGatewayGroupIDPtr(groupID), "claude-opus-4"))
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), nil, "claude-sonnet-4"))
}

func TestGatewayServiceCheckChannelPricingRestrictionRequestedSource(t *testing.T) {
	const groupID int64 = 11
	svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
		ID:                 2,
		Name:               "requested-channel",
		Status:             StatusActive,
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceRequested,
		GroupIDs:           []int64{groupID},
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4"}},
		},
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-sonnet-4": "mapped-claude",
			},
		},
	}, map[int64]string{groupID: PlatformAnthropic}))

	require.False(t, svc.checkChannelPricingRestriction(context.Background(), testGatewayGroupIDPtr(groupID), "claude-sonnet-4"))
	require.True(t, svc.checkChannelPricingRestriction(context.Background(), testGatewayGroupIDPtr(groupID), "claude-opus-4"))
}

func TestGatewayServiceCheckChannelPricingRestrictionUpstreamSourceDefers(t *testing.T) {
	const groupID int64 = 12
	svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
		ID:                 3,
		Name:               "upstream-channel",
		Status:             StatusActive,
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceUpstream,
		GroupIDs:           []int64{groupID},
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"claude-upstream-allowed"}},
		},
	}, map[int64]string{groupID: PlatformAnthropic}))

	require.False(t, svc.checkChannelPricingRestriction(context.Background(), testGatewayGroupIDPtr(groupID), "claude-sonnet-4"))
}

func TestGatewayServiceNeedsUpstreamChannelRestrictionCheck(t *testing.T) {
	const groupID int64 = 20

	t.Run("enabled only for restricted upstream channel", func(t *testing.T) {
		svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
			ID:                 1,
			Status:             StatusActive,
			GroupIDs:           []int64{groupID},
			RestrictModels:     true,
			BillingModelSource: BillingModelSourceUpstream,
		}, map[int64]string{groupID: PlatformAnthropic}))
		require.True(t, svc.needsUpstreamChannelRestrictionCheck(context.Background(), testGatewayGroupIDPtr(groupID)))
	})

	t.Run("disabled when source is not upstream", func(t *testing.T) {
		svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
			ID:                 2,
			Status:             StatusActive,
			GroupIDs:           []int64{groupID},
			RestrictModels:     true,
			BillingModelSource: BillingModelSourceRequested,
		}, map[int64]string{groupID: PlatformAnthropic}))
		require.False(t, svc.needsUpstreamChannelRestrictionCheck(context.Background(), testGatewayGroupIDPtr(groupID)))
	})

	t.Run("disabled when channel is not restricted", func(t *testing.T) {
		svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
			ID:                 3,
			Status:             StatusActive,
			GroupIDs:           []int64{groupID},
			RestrictModels:     false,
			BillingModelSource: BillingModelSourceUpstream,
		}, map[int64]string{groupID: PlatformAnthropic}))
		require.False(t, svc.needsUpstreamChannelRestrictionCheck(context.Background(), testGatewayGroupIDPtr(groupID)))
	})

	t.Run("cache load failure fails open", func(t *testing.T) {
		repo := &channelServiceRepoStub{
			listAllFn: func(context.Context) ([]Channel, error) {
				return nil, errors.New("db down")
			},
		}
		svc := newGatewayServiceWithChannelRepo(repo)
		require.False(t, svc.needsUpstreamChannelRestrictionCheck(context.Background(), testGatewayGroupIDPtr(groupID)))
	})
}

func TestGatewayServiceIsChannelModelRestrictedForSelectionWithGroup(t *testing.T) {
	const groupID int64 = 30

	t.Run("requested source restricts before account selection", func(t *testing.T) {
		svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
			ID:                 1,
			Status:             StatusActive,
			GroupIDs:           []int64{groupID},
			RestrictModels:     true,
			BillingModelSource: BillingModelSourceRequested,
			ModelPricing: []ChannelModelPricing{
				{Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4"}},
			},
		}, map[int64]string{groupID: PlatformAnthropic}))

		require.True(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), &Account{}, "claude-opus-4"))
		require.False(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), &Account{}, "claude-sonnet-4"))
	})

	t.Run("upstream source checks mapped upstream model per account", func(t *testing.T) {
		svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
			ID:                 2,
			Status:             StatusActive,
			GroupIDs:           []int64{groupID},
			RestrictModels:     true,
			BillingModelSource: BillingModelSourceUpstream,
			ModelPricing: []ChannelModelPricing{
				{Platform: PlatformAnthropic, Models: []string{"claude-upstream-allowed"}},
			},
		}, map[int64]string{groupID: PlatformAnthropic}))

		allowedAccount := &Account{
			Platform: PlatformAnthropic,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-sonnet-4": "claude-upstream-allowed",
				},
			},
		}
		restrictedAccount := &Account{
			Platform: PlatformAnthropic,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-sonnet-4": "claude-upstream-blocked",
				},
			},
		}

		require.False(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), allowedAccount, "claude-sonnet-4"))
		require.True(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), restrictedAccount, "claude-sonnet-4"))
	})

	t.Run("upstream source uses reasoning variant fallback before restriction", func(t *testing.T) {
		svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
			ID:                 4,
			Status:             StatusActive,
			GroupIDs:           []int64{groupID},
			RestrictModels:     true,
			BillingModelSource: BillingModelSourceUpstream,
			ModelPricing: []ChannelModelPricing{
				{Platform: PlatformAnthropic, Models: []string{"claude-upstream-allowed"}},
			},
		}, map[int64]string{groupID: PlatformAnthropic}))

		allowedAccount := &Account{
			Platform: PlatformAnthropic,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.4": "claude-upstream-allowed",
				},
			},
		}
		restrictedAccount := &Account{
			Platform: PlatformAnthropic,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.4": "claude-upstream-blocked",
				},
			},
		}

		require.False(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), allowedAccount, "gpt-5.4-xhigh"))
		require.True(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), restrictedAccount, "gpt-5.4-xhigh"))
	})

	t.Run("upstream source without account cannot prove restriction", func(t *testing.T) {
		svc := newGatewayServiceWithChannelRepo(makeStandardChannelRepo(Channel{
			ID:                 3,
			Status:             StatusActive,
			GroupIDs:           []int64{groupID},
			RestrictModels:     true,
			BillingModelSource: BillingModelSourceUpstream,
		}, map[int64]string{groupID: PlatformAnthropic}))

		require.False(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), nil, "claude-sonnet-4"))
		require.False(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), nil, nil, "claude-sonnet-4"))
		require.False(t, svc.isChannelModelRestrictedForSelectionWithGroup(context.Background(), testGatewayGroupIDPtr(groupID), nil, ""))
	})
}
