package service

import (
	"context"
	"errors"
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type channelServiceRepoStub struct {
	listAllFn                  func(ctx context.Context) ([]Channel, error)
	getGroupPlatformsFn        func(ctx context.Context, groupIDs []int64) (map[int64]string, error)
	createFn                   func(ctx context.Context, channel *Channel) error
	getByIDFn                  func(ctx context.Context, id int64) (*Channel, error)
	updateFn                   func(ctx context.Context, channel *Channel) error
	deleteFn                   func(ctx context.Context, id int64) error
	listFn                     func(ctx context.Context, params pagination.PaginationParams, status, search string) ([]Channel, *pagination.PaginationResult, error)
	existsByNameFn             func(ctx context.Context, name string) (bool, error)
	existsByNameExcludingFn    func(ctx context.Context, name string, excludeID int64) (bool, error)
	getGroupIDsFn              func(ctx context.Context, channelID int64) ([]int64, error)
	setGroupIDsFn              func(ctx context.Context, channelID int64, groupIDs []int64) error
	getChannelIDByGroupIDFn    func(ctx context.Context, groupID int64) (int64, error)
	getGroupsInOtherChannelsFn func(ctx context.Context, channelID int64, groupIDs []int64) ([]int64, error)
	listModelPricingFn         func(ctx context.Context, channelID int64) ([]ChannelModelPricing, error)
	createModelPricingFn       func(ctx context.Context, pricing *ChannelModelPricing) error
	updateModelPricingFn       func(ctx context.Context, pricing *ChannelModelPricing) error
	deleteModelPricingFn       func(ctx context.Context, id int64) error
	replaceModelPricingFn      func(ctx context.Context, channelID int64, pricingList []ChannelModelPricing) error
}

func (m *channelServiceRepoStub) Create(ctx context.Context, channel *Channel) error {
	if m.createFn != nil {
		return m.createFn(ctx, channel)
	}
	return nil
}

func (m *channelServiceRepoStub) GetByID(ctx context.Context, id int64) (*Channel, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, ErrChannelNotFound
}

func (m *channelServiceRepoStub) Update(ctx context.Context, channel *Channel) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, channel)
	}
	return nil
}

func (m *channelServiceRepoStub) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *channelServiceRepoStub) List(ctx context.Context, params pagination.PaginationParams, status, search string) ([]Channel, *pagination.PaginationResult, error) {
	if m.listFn != nil {
		return m.listFn(ctx, params, status, search)
	}
	return nil, nil, nil
}

func (m *channelServiceRepoStub) ListAll(ctx context.Context) ([]Channel, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return nil, nil
}

func (m *channelServiceRepoStub) ExistsByName(ctx context.Context, name string) (bool, error) {
	if m.existsByNameFn != nil {
		return m.existsByNameFn(ctx, name)
	}
	return false, nil
}

func (m *channelServiceRepoStub) ExistsByNameExcluding(ctx context.Context, name string, excludeID int64) (bool, error) {
	if m.existsByNameExcludingFn != nil {
		return m.existsByNameExcludingFn(ctx, name, excludeID)
	}
	return false, nil
}

func (m *channelServiceRepoStub) GetGroupIDs(ctx context.Context, channelID int64) ([]int64, error) {
	if m.getGroupIDsFn != nil {
		return m.getGroupIDsFn(ctx, channelID)
	}
	return nil, nil
}

func (m *channelServiceRepoStub) SetGroupIDs(ctx context.Context, channelID int64, groupIDs []int64) error {
	if m.setGroupIDsFn != nil {
		return m.setGroupIDsFn(ctx, channelID, groupIDs)
	}
	return nil
}

func (m *channelServiceRepoStub) GetChannelIDByGroupID(ctx context.Context, groupID int64) (int64, error) {
	if m.getChannelIDByGroupIDFn != nil {
		return m.getChannelIDByGroupIDFn(ctx, groupID)
	}
	return 0, nil
}

func (m *channelServiceRepoStub) GetGroupsInOtherChannels(ctx context.Context, channelID int64, groupIDs []int64) ([]int64, error) {
	if m.getGroupsInOtherChannelsFn != nil {
		return m.getGroupsInOtherChannelsFn(ctx, channelID, groupIDs)
	}
	return nil, nil
}

func (m *channelServiceRepoStub) GetGroupPlatforms(ctx context.Context, groupIDs []int64) (map[int64]string, error) {
	if m.getGroupPlatformsFn != nil {
		return m.getGroupPlatformsFn(ctx, groupIDs)
	}
	return nil, nil
}

func (m *channelServiceRepoStub) ListModelPricing(ctx context.Context, channelID int64) ([]ChannelModelPricing, error) {
	if m.listModelPricingFn != nil {
		return m.listModelPricingFn(ctx, channelID)
	}
	return nil, nil
}

func (m *channelServiceRepoStub) CreateModelPricing(ctx context.Context, pricing *ChannelModelPricing) error {
	if m.createModelPricingFn != nil {
		return m.createModelPricingFn(ctx, pricing)
	}
	return nil
}

func (m *channelServiceRepoStub) UpdateModelPricing(ctx context.Context, pricing *ChannelModelPricing) error {
	if m.updateModelPricingFn != nil {
		return m.updateModelPricingFn(ctx, pricing)
	}
	return nil
}

func (m *channelServiceRepoStub) DeleteModelPricing(ctx context.Context, id int64) error {
	if m.deleteModelPricingFn != nil {
		return m.deleteModelPricingFn(ctx, id)
	}
	return nil
}

func (m *channelServiceRepoStub) ReplaceModelPricing(ctx context.Context, channelID int64, pricingList []ChannelModelPricing) error {
	if m.replaceModelPricingFn != nil {
		return m.replaceModelPricingFn(ctx, channelID, pricingList)
	}
	return nil
}

type channelAuthCacheInvalidatorStub struct {
	groupIDs []int64
}

func (m *channelAuthCacheInvalidatorStub) InvalidateAuthCacheByKey(context.Context, string) {}

func (m *channelAuthCacheInvalidatorStub) InvalidateAuthCacheByUserID(context.Context, int64) {}

func (m *channelAuthCacheInvalidatorStub) InvalidateAuthCacheByGroupID(_ context.Context, groupID int64) {
	m.groupIDs = append(m.groupIDs, groupID)
}

func newTestChannelService(repo *channelServiceRepoStub) *ChannelService {
	return NewChannelService(repo, nil)
}

func newTestChannelServiceWithAuth(repo *channelServiceRepoStub, auth *channelAuthCacheInvalidatorStub) *ChannelService {
	return NewChannelService(repo, auth)
}

func makeStandardChannelRepo(ch Channel, groupPlatforms map[int64]string) *channelServiceRepoStub {
	return &channelServiceRepoStub{
		listAllFn: func(context.Context) ([]Channel, error) {
			return []Channel{ch}, nil
		},
		getGroupPlatformsFn: func(context.Context, []int64) (map[int64]string, error) {
			return groupPlatforms, nil
		},
	}
}

func testChannelStringPtr(v string) *string { return &v }

func TestChannelServiceGetChannelForGroupReturnsClone(t *testing.T) {
	ch := Channel{
		ID:       1,
		Name:     "test-channel",
		Status:   StatusActive,
		GroupIDs: []int64{10},
	}
	svc := newTestChannelService(makeStandardChannelRepo(ch, map[int64]string{10: PlatformAnthropic}))

	result, err := svc.GetChannelForGroup(context.Background(), 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(1), result.ID)

	result.Name = "mutated"

	result2, err := svc.GetChannelForGroup(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, "test-channel", result2.Name)
}

func TestChannelServiceBuildCacheUsesShortErrorCacheOnFailure(t *testing.T) {
	callCount := 0
	repo := &channelServiceRepoStub{
		listAllFn: func(context.Context) ([]Channel, error) {
			callCount++
			return nil, errors.New("database down")
		},
	}
	svc := newTestChannelService(repo)

	result, err := svc.GetChannelForGroup(context.Background(), 10)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, 1, callCount)

	result, err = svc.GetChannelForGroup(context.Background(), 10)
	require.NoError(t, err)
	require.Nil(t, result)
	require.Equal(t, 1, callCount)
}

func TestChannelServiceInvalidateCachePreservesPreviousSnapshotOnRebuildFailure(t *testing.T) {
	callCount := 0
	ch := Channel{
		ID:       1,
		Name:     "stable-channel",
		Status:   StatusActive,
		GroupIDs: []int64{10},
	}
	repo := &channelServiceRepoStub{
		listAllFn: func(context.Context) ([]Channel, error) {
			callCount++
			if callCount == 1 {
				return []Channel{ch}, nil
			}
			return nil, errors.New("database down")
		},
		getGroupPlatformsFn: func(context.Context, []int64) (map[int64]string, error) {
			return map[int64]string{10: PlatformAnthropic}, nil
		},
	}
	svc := newTestChannelService(repo)

	result, err := svc.GetChannelForGroup(context.Background(), 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "stable-channel", result.Name)

	svc.invalidateCache()

	result, err = svc.GetChannelForGroup(context.Background(), 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "stable-channel", result.Name)
	require.Equal(t, 2, callCount)
}

func TestChannelServiceGetChannelModelPricingAntigravityMatchesAnthropicAndGemini(t *testing.T) {
	ch := Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{10},
		ModelPricing: []ChannelModelPricing{
			{ID: 101, Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4"}, InputPrice: testChannelFloat64Ptr(3e-6)},
			{ID: 202, Platform: PlatformGemini, Models: []string{"gemini-2.5-pro"}, InputPrice: testChannelFloat64Ptr(2e-6)},
		},
	}
	svc := newTestChannelService(makeStandardChannelRepo(ch, map[int64]string{10: PlatformAntigravity}))

	anthropicPricing := svc.GetChannelModelPricing(context.Background(), 10, "claude-sonnet-4")
	require.NotNil(t, anthropicPricing)
	require.Equal(t, int64(101), anthropicPricing.ID)

	geminiPricing := svc.GetChannelModelPricing(context.Background(), 10, "gemini-2.5-pro")
	require.NotNil(t, geminiPricing)
	require.Equal(t, int64(202), geminiPricing.ID)
}

func TestChannelServiceResolveChannelMappingAntigravityMatchesPlatformSpecificRules(t *testing.T) {
	ch := Channel{
		ID:                 9,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		BillingModelSource: BillingModelSourceUpstream,
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-sonnet-4": "claude-sonnet-4-20250514",
			},
			PlatformGemini: {
				"gemini-2.5-pro": "gemini-2.5-pro-preview",
			},
		},
	}
	svc := newTestChannelService(makeStandardChannelRepo(ch, map[int64]string{10: PlatformAntigravity}))

	anthropicResult := svc.ResolveChannelMapping(context.Background(), 10, "claude-sonnet-4")
	require.True(t, anthropicResult.Mapped)
	require.Equal(t, "claude-sonnet-4-20250514", anthropicResult.MappedModel)
	require.Equal(t, int64(9), anthropicResult.ChannelID)
	require.Equal(t, BillingModelSourceUpstream, anthropicResult.BillingModelSource)

	geminiResult := svc.ResolveChannelMapping(context.Background(), 10, "gemini-2.5-pro")
	require.True(t, geminiResult.Mapped)
	require.Equal(t, "gemini-2.5-pro-preview", geminiResult.MappedModel)
}

func TestChannelServiceIsModelRestrictedUsesExpandedPlatformsForAntigravity(t *testing.T) {
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"claude-*"}},
			{Platform: PlatformGemini, Models: []string{"gemini-2.5-pro"}},
		},
	}
	svc := newTestChannelService(makeStandardChannelRepo(ch, map[int64]string{10: PlatformAntigravity}))

	require.False(t, svc.IsModelRestricted(context.Background(), 10, "claude-sonnet-4"))
	require.False(t, svc.IsModelRestricted(context.Background(), 10, "gemini-2.5-pro"))
	require.True(t, svc.IsModelRestricted(context.Background(), 10, "gpt-5.1"))
}

func TestChannelServiceResolveChannelMappingAndRestrictPreservesCompatibilityContract(t *testing.T) {
	ch := Channel{
		ID:       1,
		Status:   StatusActive,
		GroupIDs: []int64{10},
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-sonnet-4": "claude-sonnet-4-20250514",
			},
		},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4"}},
		},
	}
	svc := newTestChannelService(makeStandardChannelRepo(ch, map[int64]string{10: PlatformAnthropic}))

	groupID := int64(10)
	mapping, restricted := svc.ResolveChannelMappingAndRestrict(context.Background(), &groupID, "claude-sonnet-4")
	require.True(t, mapping.Mapped)
	require.Equal(t, "claude-sonnet-4-20250514", mapping.MappedModel)
	require.False(t, restricted)
}

func TestChannelServiceCreateDefaultsBillingModelSourceAndInvalidatesCache(t *testing.T) {
	loadCount := 0
	var createdChannel *Channel
	repo := &channelServiceRepoStub{
		listAllFn: func(context.Context) ([]Channel, error) {
			loadCount++
			return []Channel{{
				ID:       1,
				Status:   StatusActive,
				GroupIDs: []int64{10},
				ModelPricing: []ChannelModelPricing{
					{ID: 100, Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4"}},
				},
			}}, nil
		},
		getGroupPlatformsFn: func(context.Context, []int64) (map[int64]string, error) {
			return map[int64]string{10: PlatformAnthropic}, nil
		},
		existsByNameFn: func(context.Context, string) (bool, error) {
			return false, nil
		},
		getGroupsInOtherChannelsFn: func(context.Context, int64, []int64) ([]int64, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, ch *Channel) error {
			createdChannel = ch.Clone()
			ch.ID = 2
			return nil
		},
		getByIDFn: func(_ context.Context, id int64) (*Channel, error) {
			return &Channel{
				ID:                 id,
				Name:               "new-channel",
				Status:             StatusActive,
				BillingModelSource: createdChannel.BillingModelSource,
			}, nil
		},
	}
	svc := newTestChannelService(repo)

	require.NotNil(t, svc.GetChannelModelPricing(context.Background(), 10, "claude-sonnet-4"))
	require.Equal(t, 1, loadCount)

	created, err := svc.Create(context.Background(), &CreateChannelInput{Name: "new-channel"})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, BillingModelSourceChannelMapped, created.BillingModelSource)
	require.NotNil(t, createdChannel)
	require.Equal(t, BillingModelSourceChannelMapped, createdChannel.BillingModelSource)

	require.NotNil(t, svc.GetChannelModelPricing(context.Background(), 10, "claude-sonnet-4"))
	require.Equal(t, 2, loadCount)
}

func TestChannelServiceCreateRejectsMissingGroups(t *testing.T) {
	svc := newTestChannelService(&channelServiceRepoStub{
		existsByNameFn: func(context.Context, string) (bool, error) {
			return false, nil
		},
		getGroupPlatformsFn: func(context.Context, []int64) (map[int64]string, error) {
			return map[int64]string{}, nil
		},
	})

	_, err := svc.Create(context.Background(), &CreateChannelInput{
		Name:     "missing-group-channel",
		GroupIDs: []int64{999999},
	})

	require.ErrorIs(t, err, ErrGroupNotFound)
}

func TestChannelServiceUpdateInvalidatesOldAndNewGroupAuthCache(t *testing.T) {
	auth := &channelAuthCacheInvalidatorStub{}
	getByIDCalls := 0
	repo := &channelServiceRepoStub{
		getByIDFn: func(context.Context, int64) (*Channel, error) {
			getByIDCalls++
			if getByIDCalls == 1 {
				return (&Channel{
					ID:       1,
					Name:     "original",
					Status:   StatusActive,
					GroupIDs: []int64{10, 20},
				}).Clone(), nil
			}
			return &Channel{
				ID:       1,
				Name:     "updated",
				Status:   StatusActive,
				GroupIDs: []int64{20, 30},
			}, nil
		},
		getGroupsInOtherChannelsFn: func(context.Context, int64, []int64) ([]int64, error) {
			return nil, nil
		},
		getGroupPlatformsFn: func(context.Context, []int64) (map[int64]string, error) {
			return map[int64]string{
				20: PlatformAnthropic,
				30: PlatformAnthropic,
			}, nil
		},
		getGroupIDsFn: func(context.Context, int64) ([]int64, error) {
			return []int64{10, 20}, nil
		},
		updateFn: func(_ context.Context, channel *Channel) error {
			require.Equal(t, []int64{20, 30}, channel.GroupIDs)
			return nil
		},
		listAllFn: func(context.Context) ([]Channel, error) {
			return nil, nil
		},
	}
	svc := newTestChannelServiceWithAuth(repo, auth)

	groupIDs := []int64{20, 30}
	updated, err := svc.Update(context.Background(), 1, &UpdateChannelInput{
		Name:        "updated",
		Description: testChannelStringPtr("new description"),
		GroupIDs:    &groupIDs,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, []int64{10, 20, 30}, auth.groupIDs)
}

func TestChannelServiceUpdateRejectsMissingGroups(t *testing.T) {
	svc := newTestChannelService(&channelServiceRepoStub{
		getByIDFn: func(context.Context, int64) (*Channel, error) {
			return &Channel{
				ID:       1,
				Name:     "primary",
				Status:   StatusActive,
				GroupIDs: []int64{10},
			}, nil
		},
		getGroupPlatformsFn: func(context.Context, []int64) (map[int64]string, error) {
			return map[int64]string{}, nil
		},
	})

	groupIDs := []int64{999999}
	_, err := svc.Update(context.Background(), 1, &UpdateChannelInput{
		GroupIDs: &groupIDs,
	})

	require.ErrorIs(t, err, ErrGroupNotFound)
}

func TestChannelServiceDeleteInvalidatesGroupAuthCache(t *testing.T) {
	auth := &channelAuthCacheInvalidatorStub{}
	repo := &channelServiceRepoStub{
		getGroupIDsFn: func(context.Context, int64) ([]int64, error) {
			return []int64{10, 20}, nil
		},
		deleteFn: func(context.Context, int64) error {
			return nil
		},
		listAllFn: func(context.Context) ([]Channel, error) {
			return nil, nil
		},
	}
	svc := newTestChannelServiceWithAuth(repo, auth)

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, []int64{10, 20}, auth.groupIDs)
}

func TestValidateNoConflictingModelsRejectsWildcardOverlap(t *testing.T) {
	err := validateNoConflictingModels([]ChannelModelPricing{
		{Platform: PlatformAnthropic, Models: []string{"claude-*"}},
		{Platform: PlatformAnthropic, Models: []string{"claude-sonnet-*"}},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "MODEL_PATTERN_CONFLICT")
}

func TestValidateNoConflictingModelsAllowsCrossPlatformReuse(t *testing.T) {
	err := validateNoConflictingModels([]ChannelModelPricing{
		{Platform: PlatformAnthropic, Models: []string{"shared-model"}},
		{Platform: PlatformOpenAI, Models: []string{"shared-model"}},
	})

	require.NoError(t, err)
}

func TestValidateNoConflictingMappingsRejectsWildcardOverlap(t *testing.T) {
	err := validateNoConflictingMappings(map[string]map[string]string{
		PlatformAnthropic: {
			"claude-*":        "claude-sonnet-4",
			"claude-sonnet-*": "claude-sonnet-4-20250514",
		},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "MAPPING_PATTERN_CONFLICT")
}

func TestChannelServiceCreateRejectsInvalidIntervals(t *testing.T) {
	svc := newTestChannelService(&channelServiceRepoStub{
		existsByNameFn: func(context.Context, string) (bool, error) {
			return false, nil
		},
	})

	_, err := svc.Create(context.Background(), &CreateChannelInput{
		Name: "bad-interval-channel",
		ModelPricing: []ChannelModelPricing{
			{
				Platform: PlatformAnthropic,
				Models:   []string{"claude-sonnet-4"},
				Intervals: []PricingInterval{
					{MinTokens: 0, MaxTokens: testChannelIntPtr(1000), InputPrice: testChannelFloat64Ptr(1e-6)},
					{MinTokens: 500, MaxTokens: testChannelIntPtr(2000), InputPrice: testChannelFloat64Ptr(2e-6)},
				},
			},
		},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_PRICING_INTERVALS")
}

func TestChannelServiceUpdateRejectsNameConflict(t *testing.T) {
	svc := newTestChannelService(&channelServiceRepoStub{
		getByIDFn: func(context.Context, int64) (*Channel, error) {
			return &Channel{ID: 1, Name: "original", Status: StatusActive}, nil
		},
		existsByNameExcludingFn: func(context.Context, string, int64) (bool, error) {
			return true, nil
		},
	})

	_, err := svc.Update(context.Background(), 1, &UpdateChannelInput{Name: "conflict"})
	require.ErrorIs(t, err, ErrChannelExists)
}

func TestChannelServiceUpdateRejectsConflictingMappings(t *testing.T) {
	svc := newTestChannelService(&channelServiceRepoStub{
		getByIDFn: func(context.Context, int64) (*Channel, error) {
			return &Channel{
				ID:     1,
				Name:   "original",
				Status: StatusActive,
			}, nil
		},
	})

	_, err := svc.Update(context.Background(), 1, &UpdateChannelInput{
		ModelMapping: map[string]map[string]string{
			PlatformAnthropic: {
				"claude-*":        "claude-sonnet-4",
				"claude-sonnet-*": "claude-sonnet-4-20250514",
			},
		},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "MAPPING_PATTERN_CONFLICT")
}

func TestChannelServiceDeletePropagatesRepositoryError(t *testing.T) {
	svc := newTestChannelService(&channelServiceRepoStub{
		deleteFn: func(context.Context, int64) error {
			return ErrChannelNotFound
		},
	})

	err := svc.Delete(context.Background(), 999)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChannelNotFound)
}
