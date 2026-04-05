package service

import (
	"context"
	"testing"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func newGatewayChannelUsageServiceForTest(usageRepo UsageLogRepository, userRepo UserRepository, subRepo UserSubscriptionRepository) *GatewayService {
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1.1
	return NewGatewayService(
		nil,
		nil,
		usageRepo,
		nil,
		userRepo,
		subRepo,
		nil,
		nil,
		cfg,
		nil,
		nil,
		NewBillingService(cfg, nil),
		nil,
		&BillingCacheService{},
		nil,
		nil,
		&DeferredService{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func attachGatewayChannelPricing(groupID int64, platform string, svc *GatewayService, ch Channel) {
	channelService := newTestChannelService(makeStandardChannelRepo(ch, map[int64]string{groupID: platform}))
	svc.channelService = channelService
	svc.resolver = NewModelPricingResolver(channelService, svc.billingService)
}

func attachOpenAIChannelPricing(groupID int64, platform string, svc *OpenAIGatewayService, ch Channel) {
	channelService := newTestChannelService(makeStandardChannelRepo(ch, map[int64]string{groupID: platform}))
	svc.channelService = channelService
	svc.resolver = NewModelPricingResolver(channelService, svc.billingService)
}

func TestResolveChannelBillingModel_UsesConfiguredSource(t *testing.T) {
	tests := []struct {
		name     string
		fields   ChannelUsageFields
		fallback string
		want     string
	}{
		{
			name: "requested source prefers original model",
			fields: ChannelUsageFields{
				OriginalModel:      "claude-sonnet-4",
				ChannelMappedModel: "mapped-claude",
				BillingModelSource: BillingModelSourceRequested,
			},
			fallback: "upstream-claude",
			want:     "claude-sonnet-4",
		},
		{
			name: "channel mapped source prefers mapped model",
			fields: ChannelUsageFields{
				OriginalModel:      "claude-sonnet-4",
				ChannelMappedModel: "mapped-claude",
				BillingModelSource: BillingModelSourceChannelMapped,
			},
			fallback: "upstream-claude",
			want:     "mapped-claude",
		},
		{
			name: "upstream source uses fallback billing model",
			fields: ChannelUsageFields{
				OriginalModel:      "gpt-5",
				ChannelMappedModel: "gpt-5.4",
				BillingModelSource: BillingModelSourceUpstream,
			},
			fallback: "gpt-5.4",
			want:     "gpt-5.4",
		},
		{
			name: "empty source falls back",
			fields: ChannelUsageFields{
				OriginalModel:      "gpt-5",
				ChannelMappedModel: "gpt-5.4",
			},
			fallback: "gpt-5.4",
			want:     "gpt-5.4",
		},
		{
			name: "missing preferred model falls back",
			fields: ChannelUsageFields{
				BillingModelSource: BillingModelSourceRequested,
			},
			fallback: "claude-sonnet-4-upstream",
			want:     "claude-sonnet-4-upstream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveChannelBillingModel(tt.fields, tt.fallback)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGatewayServiceRecordUsage_ChannelMappedPricingAndFields(t *testing.T) {
	const (
		groupID   int64   = 101
		channelID int64   = 501
		unitPrice float64 = 0.42
	)

	usageRepo := &usageLogWriterStub{}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newGatewayChannelUsageServiceForTest(usageRepo, userRepo, subRepo)
	attachGatewayChannelPricing(groupID, PlatformAnthropic, svc, Channel{
		ID:                 channelID,
		Name:               "anthropic-channel",
		Status:             StatusActive,
		GroupIDs:           []int64{groupID},
		BillingModelSource: BillingModelSourceChannelMapped,
		ModelPricing: []ChannelModelPricing{{
			Platform:        PlatformAnthropic,
			Models:          []string{"claude-sonnet-4-channel"},
			BillingMode:     BillingModePerRequest,
			PerRequestPrice: testChannelFloat64Ptr(unitPrice),
		}},
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-sonnet-4": "claude-sonnet-4-channel",
			},
		},
	})

	err := svc.RecordUsage(context.Background(), &RecordUsageInput{
		Result: &ForwardResult{
			RequestID:     "gateway-channel-usage",
			Model:         "claude-sonnet-4",
			UpstreamModel: "claude-sonnet-4-20250514",
			Usage: ClaudeUsage{
				InputTokens:  120,
				OutputTokens: 24,
			},
			Duration: time.Second,
		},
		APIKey: &APIKey{
			ID:      601,
			GroupID: testGatewayGroupIDPtr(groupID),
			Group:   &Group{ID: groupID, RateMultiplier: 1},
			Quota:   100,
		},
		User:    &User{ID: 701},
		Account: &Account{ID: 801},
	})

	require.NoError(t, err)
	require.Equal(t, 1, userRepo.deductCalls)
	require.InDelta(t, unitPrice, userRepo.lastAmount, 1e-12)
	require.NotNil(t, usageRepo.lastUsageLogRequest)
	require.Equal(t, "claude-sonnet-4", usageRepo.lastUsageLogRequest.Model)
	require.Equal(t, "claude-sonnet-4", usageRepo.lastUsageLogRequest.RequestedModel)
	require.NotNil(t, usageRepo.lastUsageLogRequest.UpstreamModel)
	require.Equal(t, "claude-sonnet-4-20250514", *usageRepo.lastUsageLogRequest.UpstreamModel)
	require.NotNil(t, usageRepo.lastUsageLogRequest.ChannelID)
	require.Equal(t, channelID, *usageRepo.lastUsageLogRequest.ChannelID)
	require.NotNil(t, usageRepo.lastUsageLogRequest.ModelMappingChain)
	require.Equal(t, "claude-sonnet-4→claude-sonnet-4-channel→claude-sonnet-4-20250514", *usageRepo.lastUsageLogRequest.ModelMappingChain)
	require.NotNil(t, usageRepo.lastUsageLogRequest.BillingMode)
	require.Equal(t, string(BillingModePerRequest), *usageRepo.lastUsageLogRequest.BillingMode)
	require.InDelta(t, unitPrice, usageRepo.lastUsageLogRequest.TotalCost, 1e-12)
	require.InDelta(t, unitPrice, usageRepo.lastUsageLogRequest.ActualCost, 1e-12)
}

func TestOpenAIGatewayServiceRecordUsage_UpstreamChannelPricingAndFields(t *testing.T) {
	const (
		groupID   int64   = 202
		channelID int64   = 502
		unitPrice float64 = 0.8
	)

	usageRepo := &usageLogWriterStub{}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newOpenAIRecordUsageServiceForTest(usageRepo, userRepo, subRepo, nil)
	attachOpenAIChannelPricing(groupID, PlatformOpenAI, svc, Channel{
		ID:                 channelID,
		Name:               "openai-channel",
		Status:             StatusActive,
		GroupIDs:           []int64{groupID},
		BillingModelSource: BillingModelSourceUpstream,
		ModelPricing: []ChannelModelPricing{{
			Platform:        PlatformOpenAI,
			Models:          []string{"gpt-5.4"},
			BillingMode:     BillingModePerRequest,
			PerRequestPrice: testChannelFloat64Ptr(unitPrice),
		}},
		ModelMapping: map[string]map[string]string{
			PlatformOpenAI: {
				"gpt-5": "gpt-5.4",
			},
		},
	})

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID:     "openai-channel-usage",
			Model:         "gpt-5",
			BillingModel:  "gpt-5.4",
			UpstreamModel: "gpt-5.4-20260101",
			Usage: OpenAIUsage{
				InputTokens:  80,
				OutputTokens: 16,
			},
			Duration: time.Second,
		},
		APIKey: &APIKey{
			ID:      602,
			GroupID: testGatewayGroupIDPtr(groupID),
			Group:   &Group{ID: groupID, RateMultiplier: 1},
			Quota:   100,
		},
		User:    &User{ID: 702},
		Account: &Account{ID: 802},
	})

	require.NoError(t, err)
	require.Equal(t, 1, userRepo.deductCalls)
	require.InDelta(t, unitPrice, userRepo.lastAmount, 1e-12)
	require.NotNil(t, usageRepo.lastUsageLogRequest)
	require.Equal(t, "gpt-5", usageRepo.lastUsageLogRequest.Model)
	require.Equal(t, "gpt-5", usageRepo.lastUsageLogRequest.RequestedModel)
	require.NotNil(t, usageRepo.lastUsageLogRequest.UpstreamModel)
	require.Equal(t, "gpt-5.4-20260101", *usageRepo.lastUsageLogRequest.UpstreamModel)
	require.NotNil(t, usageRepo.lastUsageLogRequest.ChannelID)
	require.Equal(t, channelID, *usageRepo.lastUsageLogRequest.ChannelID)
	require.NotNil(t, usageRepo.lastUsageLogRequest.ModelMappingChain)
	require.Equal(t, "gpt-5→gpt-5.4→gpt-5.4-20260101", *usageRepo.lastUsageLogRequest.ModelMappingChain)
	require.NotNil(t, usageRepo.lastUsageLogRequest.BillingMode)
	require.Equal(t, string(BillingModePerRequest), *usageRepo.lastUsageLogRequest.BillingMode)
	require.InDelta(t, unitPrice, usageRepo.lastUsageLogRequest.TotalCost, 1e-12)
	require.InDelta(t, unitPrice, usageRepo.lastUsageLogRequest.ActualCost, 1e-12)
}
