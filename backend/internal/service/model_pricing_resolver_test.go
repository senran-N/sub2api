package service

import (
	"context"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type mockChannelRepositoryForResolver struct {
	channels       []Channel
	groupPlatforms map[int64]string
}

func (m *mockChannelRepositoryForResolver) Create(context.Context, *Channel) error { return nil }
func (m *mockChannelRepositoryForResolver) GetByID(context.Context, int64) (*Channel, error) {
	return nil, ErrChannelNotFound
}
func (m *mockChannelRepositoryForResolver) Update(context.Context, *Channel) error { return nil }
func (m *mockChannelRepositoryForResolver) Delete(context.Context, int64) error    { return nil }
func (m *mockChannelRepositoryForResolver) List(context.Context, pagination.PaginationParams, string, string) ([]Channel, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (m *mockChannelRepositoryForResolver) ListAll(context.Context) ([]Channel, error) {
	return m.channels, nil
}
func (m *mockChannelRepositoryForResolver) ExistsByName(context.Context, string) (bool, error) {
	return false, nil
}
func (m *mockChannelRepositoryForResolver) ExistsByNameExcluding(context.Context, string, int64) (bool, error) {
	return false, nil
}
func (m *mockChannelRepositoryForResolver) GetGroupIDs(context.Context, int64) ([]int64, error) {
	return nil, nil
}
func (m *mockChannelRepositoryForResolver) SetGroupIDs(context.Context, int64, []int64) error {
	return nil
}
func (m *mockChannelRepositoryForResolver) GetChannelIDByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *mockChannelRepositoryForResolver) GetGroupsInOtherChannels(context.Context, int64, []int64) ([]int64, error) {
	return nil, nil
}
func (m *mockChannelRepositoryForResolver) GetGroupPlatforms(context.Context, []int64) (map[int64]string, error) {
	return m.groupPlatforms, nil
}
func (m *mockChannelRepositoryForResolver) ListModelPricing(context.Context, int64) ([]ChannelModelPricing, error) {
	return nil, nil
}
func (m *mockChannelRepositoryForResolver) CreateModelPricing(context.Context, *ChannelModelPricing) error {
	return nil
}
func (m *mockChannelRepositoryForResolver) UpdateModelPricing(context.Context, *ChannelModelPricing) error {
	return nil
}
func (m *mockChannelRepositoryForResolver) DeleteModelPricing(context.Context, int64) error {
	return nil
}
func (m *mockChannelRepositoryForResolver) ReplaceModelPricing(context.Context, int64, []ChannelModelPricing) error {
	return nil
}

func newTestBillingServiceForResolver() *BillingService {
	return NewBillingService(nil, nil)
}

func newResolverWithChannelPricing(pricing []ChannelModelPricing) *ModelPricingResolver {
	const groupID = 100
	repo := &mockChannelRepositoryForResolver{
		channels: []Channel{{
			ID:           1,
			Name:         "test-channel",
			Status:       StatusActive,
			GroupIDs:     []int64{groupID},
			ModelPricing: pricing,
		}},
		groupPlatforms: map[int64]string{groupID: PlatformAnthropic},
	}
	return NewModelPricingResolver(NewChannelService(repo, nil), newTestBillingServiceForResolver())
}

func testResolverGroupIDPtr() *int64 {
	v := int64(100)
	return &v
}

func TestModelPricingResolverResolveWithoutGroupUsesBasePricing(t *testing.T) {
	resolver := NewModelPricingResolver(&ChannelService{}, newTestBillingServiceForResolver())

	resolved := resolver.Resolve(context.Background(), PricingInput{
		Model:   "claude-sonnet-4",
		GroupID: nil,
	})

	require.NotNil(t, resolved)
	require.Equal(t, BillingModeToken, resolved.Mode)
	require.NotNil(t, resolved.BasePricing)
	require.Equal(t, PricingSourceLiteLLM, resolved.Source)
	require.InDelta(t, 3e-6, resolved.BasePricing.InputPricePerToken, 1e-12)
}

func TestModelPricingResolverResolveChannelFlatOverride(t *testing.T) {
	resolver := newResolverWithChannelPricing([]ChannelModelPricing{{
		Platform:    PlatformAnthropic,
		Models:      []string{"claude-sonnet-4"},
		BillingMode: BillingModeToken,
		InputPrice:  testChannelFloat64Ptr(10e-6),
		OutputPrice: testChannelFloat64Ptr(50e-6),
	}})

	resolved := resolver.Resolve(context.Background(), PricingInput{
		Model:   "claude-sonnet-4",
		GroupID: testResolverGroupIDPtr(),
	})

	require.NotNil(t, resolved)
	require.Equal(t, PricingSourceChannel, resolved.Source)
	require.Equal(t, BillingModeToken, resolved.Mode)
	require.NotNil(t, resolved.BasePricing)
	require.InDelta(t, 10e-6, resolved.BasePricing.InputPricePerToken, 1e-12)
	require.InDelta(t, 50e-6, resolved.BasePricing.OutputPricePerToken, 1e-12)
}

func TestModelPricingResolverResolveChannelIntervals(t *testing.T) {
	resolver := newResolverWithChannelPricing([]ChannelModelPricing{{
		Platform:    PlatformAnthropic,
		Models:      []string{"claude-sonnet-4"},
		BillingMode: BillingModeToken,
		Intervals: []PricingInterval{
			{MinTokens: 0, MaxTokens: testChannelIntPtr(128000), InputPrice: testChannelFloat64Ptr(2e-6), OutputPrice: testChannelFloat64Ptr(8e-6)},
			{MinTokens: 128000, MaxTokens: nil, InputPrice: testChannelFloat64Ptr(4e-6), OutputPrice: testChannelFloat64Ptr(16e-6)},
		},
	}})

	resolved := resolver.Resolve(context.Background(), PricingInput{
		Model:   "claude-sonnet-4",
		GroupID: testResolverGroupIDPtr(),
	})

	require.Equal(t, PricingSourceChannel, resolved.Source)
	require.Len(t, resolved.Intervals, 2)

	pricingA := resolver.GetIntervalPricing(resolved, 50000)
	require.NotNil(t, pricingA)
	require.InDelta(t, 2e-6, pricingA.InputPricePerToken, 1e-12)
	require.InDelta(t, 8e-6, pricingA.OutputPricePerToken, 1e-12)

	pricingB := resolver.GetIntervalPricing(resolved, 200000)
	require.NotNil(t, pricingB)
	require.InDelta(t, 4e-6, pricingB.InputPricePerToken, 1e-12)
	require.InDelta(t, 16e-6, pricingB.OutputPricePerToken, 1e-12)
}

func TestModelPricingResolverResolvePerRequestOverride(t *testing.T) {
	resolver := newResolverWithChannelPricing([]ChannelModelPricing{{
		Platform:        PlatformAnthropic,
		Models:          []string{"claude-sonnet-4"},
		BillingMode:     BillingModePerRequest,
		PerRequestPrice: testChannelFloat64Ptr(0.05),
		Intervals: []PricingInterval{
			{MinTokens: 0, MaxTokens: testChannelIntPtr(128000), PerRequestPrice: testChannelFloat64Ptr(0.03)},
			{MinTokens: 128000, MaxTokens: nil, PerRequestPrice: testChannelFloat64Ptr(0.10)},
		},
	}})

	resolved := resolver.Resolve(context.Background(), PricingInput{
		Model:   "claude-sonnet-4",
		GroupID: testResolverGroupIDPtr(),
	})

	require.Equal(t, PricingSourceChannel, resolved.Source)
	require.Equal(t, BillingModePerRequest, resolved.Mode)
	require.InDelta(t, 0.05, resolved.DefaultPerRequestPrice, 1e-12)
	require.InDelta(t, 0.03, resolver.GetRequestTierPriceByContext(resolved, 50000), 1e-12)
	require.InDelta(t, 0.10, resolver.GetRequestTierPriceByContext(resolved, 200000), 1e-12)
}

func TestFilterValidIntervals(t *testing.T) {
	intervals := []PricingInterval{
		{MinTokens: 0, MaxTokens: testChannelIntPtr(1000)},
		{MinTokens: 1000, MaxTokens: testChannelIntPtr(2000), InputPrice: testChannelFloat64Ptr(1e-6)},
		{MinTokens: 2000, MaxTokens: nil, PerRequestPrice: testChannelFloat64Ptr(0.1)},
	}

	filtered := filterValidIntervals(intervals)
	require.Len(t, filtered, 2)
	require.Equal(t, 1000, filtered[0].MinTokens)
	require.Equal(t, 2000, filtered[1].MinTokens)
}
