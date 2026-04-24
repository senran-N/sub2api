package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func testAdminFloat64Ptr(v float64) *float64 { return &v }
func testAdminIntPtr(v int) *int             { return &v }

type channelHandlerRepoStub struct {
	channels                map[int64]*service.Channel
	nextID                  int64
	lastListParams          pagination.PaginationParams
	lastListStatus          string
	lastListSearch          string
	deletedIDs              []int64
	listFn                  func(ctx context.Context, params pagination.PaginationParams, status, search string) ([]service.Channel, *pagination.PaginationResult, error)
	listAllFn               func(ctx context.Context) ([]service.Channel, error)
	getByIDFn               func(ctx context.Context, id int64) (*service.Channel, error)
	existsByNameFn          func(ctx context.Context, name string) (bool, error)
	existsByNameExcludingFn func(ctx context.Context, name string, excludeID int64) (bool, error)
	getGroupPlatformsFn     func(ctx context.Context, groupIDs []int64) (map[int64]string, error)
}

func newChannelHandlerRepoStub(channels ...service.Channel) *channelHandlerRepoStub {
	store := make(map[int64]*service.Channel, len(channels))
	var maxID int64
	for i := range channels {
		ch := channels[i].Clone()
		store[ch.ID] = ch
		if ch.ID > maxID {
			maxID = ch.ID
		}
	}
	return &channelHandlerRepoStub{
		channels: store,
		nextID:   maxID + 1,
	}
}

func (s *channelHandlerRepoStub) Create(_ context.Context, channel *service.Channel) error {
	if channel.ID == 0 {
		channel.ID = s.nextID
		s.nextID++
	}
	s.channels[channel.ID] = channel.Clone()
	return nil
}

func (s *channelHandlerRepoStub) GetByID(ctx context.Context, id int64) (*service.Channel, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	ch, ok := s.channels[id]
	if !ok {
		return nil, service.ErrChannelNotFound
	}
	return ch.Clone(), nil
}

func (s *channelHandlerRepoStub) Update(_ context.Context, channel *service.Channel) error {
	s.channels[channel.ID] = channel.Clone()
	return nil
}

func (s *channelHandlerRepoStub) Delete(_ context.Context, id int64) error {
	delete(s.channels, id)
	s.deletedIDs = append(s.deletedIDs, id)
	return nil
}

func (s *channelHandlerRepoStub) List(ctx context.Context, params pagination.PaginationParams, status, search string) ([]service.Channel, *pagination.PaginationResult, error) {
	s.lastListParams = params
	s.lastListStatus = status
	s.lastListSearch = search
	if s.listFn != nil {
		return s.listFn(ctx, params, status, search)
	}
	items := make([]service.Channel, 0, len(s.channels))
	for _, ch := range s.channels {
		items = append(items, *ch.Clone())
	}
	return items, &pagination.PaginationResult{
		Total:    int64(len(items)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    1,
	}, nil
}

func (s *channelHandlerRepoStub) ListAll(ctx context.Context) ([]service.Channel, error) {
	if s.listAllFn != nil {
		return s.listAllFn(ctx)
	}
	items := make([]service.Channel, 0, len(s.channels))
	for _, ch := range s.channels {
		items = append(items, *ch.Clone())
	}
	return items, nil
}

func (s *channelHandlerRepoStub) ExistsByName(ctx context.Context, name string) (bool, error) {
	if s.existsByNameFn != nil {
		return s.existsByNameFn(ctx, name)
	}
	for _, ch := range s.channels {
		if ch.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (s *channelHandlerRepoStub) ExistsByNameExcluding(ctx context.Context, name string, excludeID int64) (bool, error) {
	if s.existsByNameExcludingFn != nil {
		return s.existsByNameExcludingFn(ctx, name, excludeID)
	}
	for id, ch := range s.channels {
		if id != excludeID && ch.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (*channelHandlerRepoStub) GetGroupIDs(context.Context, int64) ([]int64, error) {
	return nil, nil
}

func (*channelHandlerRepoStub) SetGroupIDs(context.Context, int64, []int64) error {
	return nil
}

func (*channelHandlerRepoStub) GetChannelIDByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}

func (*channelHandlerRepoStub) GetGroupsInOtherChannels(context.Context, int64, []int64) ([]int64, error) {
	return nil, nil
}

func (s *channelHandlerRepoStub) GetGroupPlatforms(ctx context.Context, groupIDs []int64) (map[int64]string, error) {
	if s.getGroupPlatformsFn != nil {
		return s.getGroupPlatformsFn(ctx, groupIDs)
	}
	result := make(map[int64]string, len(groupIDs))
	for _, groupID := range groupIDs {
		result[groupID] = service.PlatformAnthropic
	}
	return result, nil
}

func (*channelHandlerRepoStub) ListModelPricing(context.Context, int64) ([]service.ChannelModelPricing, error) {
	return nil, nil
}

func (*channelHandlerRepoStub) CreateModelPricing(context.Context, *service.ChannelModelPricing) error {
	return nil
}

func (*channelHandlerRepoStub) UpdateModelPricing(context.Context, *service.ChannelModelPricing) error {
	return nil
}

func (*channelHandlerRepoStub) DeleteModelPricing(context.Context, int64) error {
	return nil
}

func (*channelHandlerRepoStub) ReplaceModelPricing(context.Context, int64, []service.ChannelModelPricing) error {
	return nil
}

func newChannelHandlerTestRouter(repo *channelHandlerRepoStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewChannelHandler(service.NewChannelService(repo, nil, nil, nil), nil)
	router.GET("/admin/channels", handler.List)
	router.GET("/admin/channels/:id", handler.GetByID)
	router.POST("/admin/channels", handler.Create)
	router.PUT("/admin/channels/:id", handler.Update)
	router.DELETE("/admin/channels/:id", handler.Delete)
	return router
}

func newChannelHandlerWithBillingTestRouter(repo *channelHandlerRepoStub, billing *service.BillingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewChannelHandler(service.NewChannelService(repo, nil, nil, nil), billing)
	router.GET("/admin/channels/model-pricing", handler.GetModelDefaultPricing)
	return router
}

func TestChannelToResponseNilInput(t *testing.T) {
	require.Nil(t, channelToResponse(nil))
}

func TestChannelToResponseFullChannel(t *testing.T) {
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	ch := &service.Channel{
		ID:                 42,
		Name:               "test-channel",
		Description:        "desc",
		Status:             "active",
		BillingModelSource: service.BillingModelSourceUpstream,
		RestrictModels:     true,
		CreatedAt:          now,
		UpdatedAt:          now.Add(time.Hour),
		GroupIDs:           []int64{1, 2, 3},
		ModelPricing: []service.ChannelModelPricing{
			{
				ID:              10,
				Platform:        service.PlatformOpenAI,
				Models:          []string{"gpt-4"},
				BillingMode:     service.BillingModeToken,
				InputPrice:      testAdminFloat64Ptr(0.01),
				OutputPrice:     testAdminFloat64Ptr(0.03),
				CacheWritePrice: testAdminFloat64Ptr(0.005),
				CacheReadPrice:  testAdminFloat64Ptr(0.002),
				PerRequestPrice: testAdminFloat64Ptr(0.5),
			},
		},
		ModelMapping: map[string]map[string]string{
			service.PlatformAnthropic: {"claude-3-haiku": "claude-haiku-3"},
		},
	}

	resp := channelToResponse(ch)
	require.NotNil(t, resp)
	require.Equal(t, int64(42), resp.ID)
	require.Equal(t, "test-channel", resp.Name)
	require.Equal(t, service.BillingModelSourceUpstream, resp.BillingModelSource)
	require.True(t, resp.RestrictModels)
	require.Equal(t, []int64{1, 2, 3}, resp.GroupIDs)
	require.Equal(t, "2025-06-01T12:00:00Z", resp.CreatedAt)
	require.Equal(t, "2025-06-01T13:00:00Z", resp.UpdatedAt)
	require.Equal(t, "claude-haiku-3", resp.ModelMapping[service.PlatformAnthropic]["claude-3-haiku"])
	require.Len(t, resp.ModelPricing, 1)
	require.Equal(t, "token", resp.ModelPricing[0].BillingMode)
}

func TestChannelToResponseDefaults(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ch := &service.Channel{
		ID:           1,
		Name:         "ch",
		CreatedAt:    now,
		UpdatedAt:    now,
		ModelMapping: nil,
		ModelPricing: []service.ChannelModelPricing{{
			Models: nil,
		}},
	}

	resp := channelToResponse(ch)
	require.Equal(t, service.BillingModelSourceChannelMapped, resp.BillingModelSource)
	require.NotNil(t, resp.GroupIDs)
	require.Empty(t, resp.GroupIDs)
	require.NotNil(t, resp.ModelMapping)
	require.Empty(t, resp.ModelMapping)
	require.Len(t, resp.ModelPricing, 1)
	require.Equal(t, service.PlatformAnthropic, resp.ModelPricing[0].Platform)
	require.Equal(t, string(service.BillingModeToken), resp.ModelPricing[0].BillingMode)
	require.NotNil(t, resp.ModelPricing[0].Models)
	require.Empty(t, resp.ModelPricing[0].Models)
}

func TestPricingRequestToServiceDefaultsAndIntervals(t *testing.T) {
	result := pricingRequestToService([]channelModelPricingRequest{
		{
			Models: []string{"m1"},
			Intervals: []pricingIntervalRequest{
				{
					MinTokens:       0,
					MaxTokens:       testAdminIntPtr(1000),
					TierLabel:       "1K",
					InputPrice:      testAdminFloat64Ptr(0.01),
					OutputPrice:     testAdminFloat64Ptr(0.02),
					CacheWritePrice: testAdminFloat64Ptr(0.003),
					CacheReadPrice:  testAdminFloat64Ptr(0.001),
					PerRequestPrice: testAdminFloat64Ptr(0.1),
					SortOrder:       1,
				},
			},
		},
	})

	require.Len(t, result, 1)
	require.Equal(t, service.PlatformAnthropic, result[0].Platform)
	require.Equal(t, service.BillingModeToken, result[0].BillingMode)
	require.Len(t, result[0].Intervals, 1)
	require.Equal(t, "1K", result[0].Intervals[0].TierLabel)
}

func TestValidatePricingBillingMode(t *testing.T) {
	err := validatePricingBillingMode([]service.ChannelModelPricing{{
		BillingMode: service.BillingModePerRequest,
		Models:      []string{"m1"},
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "per-request price or intervals required")

	err = validatePricingBillingMode([]service.ChannelModelPricing{{
		BillingMode: service.BillingModeToken,
		Models:      []string{"m1"},
		InputPrice:  testAdminFloat64Ptr(-1),
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "input_price must be >= 0")

	err = validatePricingBillingMode([]service.ChannelModelPricing{{
		BillingMode: service.BillingModeToken,
		Models:      []string{"m1"},
		Intervals: []service.PricingInterval{
			{MinTokens: 0, MaxTokens: testAdminIntPtr(1000)},
		},
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "has no price fields set")

	err = validatePricingBillingMode([]service.ChannelModelPricing{{
		BillingMode:     service.BillingModePerRequest,
		Models:          []string{"m1"},
		PerRequestPrice: testAdminFloat64Ptr(0.1),
	}})
	require.NoError(t, err)
}

func TestChannelHandlerListTrimsAndClampsSearch(t *testing.T) {
	now := time.Date(2025, 7, 1, 8, 0, 0, 0, time.UTC)
	repo := newChannelHandlerRepoStub()
	repo.listFn = func(_ context.Context, params pagination.PaginationParams, status, search string) ([]service.Channel, *pagination.PaginationResult, error) {
		return []service.Channel{{
				ID:        7,
				Name:      "primary",
				Status:    service.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}}, &pagination.PaginationResult{
				Total:    5,
				Page:     params.Page,
				PageSize: params.PageSize,
				Pages:    1,
			}, nil
	}
	router := newChannelHandlerTestRouter(repo)

	search := "  " + strings.Repeat("x", 120) + "  "
	req := httptest.NewRequest(http.MethodGet, "/admin/channels?page=2&page_size=25&status=active&search="+url.QueryEscape(search), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 2, repo.lastListParams.Page)
	require.Equal(t, 25, repo.lastListParams.PageSize)
	require.Equal(t, service.StatusActive, repo.lastListStatus)
	require.Len(t, repo.lastListSearch, 100)
	require.Equal(t, strings.Repeat("x", 100), repo.lastListSearch)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items    []channelResponse `json:"items"`
			Total    int64             `json:"total"`
			Page     int               `json:"page"`
			PageSize int               `json:"page_size"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, int64(5), resp.Data.Total)
	require.Equal(t, 2, resp.Data.Page)
	require.Equal(t, 25, resp.Data.PageSize)
	require.Len(t, resp.Data.Items, 1)
	require.Equal(t, int64(7), resp.Data.Items[0].ID)
}

func TestChannelHandlerGetByIDInvalidID(t *testing.T) {
	router := newChannelHandlerTestRouter(newChannelHandlerRepoStub())

	req := httptest.NewRequest(http.MethodGet, "/admin/channels/not-a-number", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "INVALID_CHANNEL_ID", resp.Reason)
}

func TestChannelHandlerGetByIDNotFound(t *testing.T) {
	router := newChannelHandlerTestRouter(newChannelHandlerRepoStub())

	req := httptest.NewRequest(http.MethodGet, "/admin/channels/99", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	var resp struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "CHANNEL_NOT_FOUND", resp.Reason)
}

func TestChannelHandlerCreateValidationFailure(t *testing.T) {
	router := newChannelHandlerTestRouter(newChannelHandlerRepoStub())

	req := httptest.NewRequest(http.MethodPost, "/admin/channels", bytes.NewBufferString(`{"description":"missing name"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "VALIDATION_ERROR", resp.Reason)
}

func TestChannelHandlerCreatePricingValidationFailure(t *testing.T) {
	router := newChannelHandlerTestRouter(newChannelHandlerRepoStub())

	req := httptest.NewRequest(http.MethodPost, "/admin/channels", bytes.NewBufferString(`{
		"name":"primary",
		"model_pricing":[{"models":["claude-sonnet-4"],"billing_mode":"per_request"}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "VALIDATION_ERROR", resp.Reason)
	require.Contains(t, resp.Message, "per-request price or intervals required")
}

func TestChannelHandlerCreateSuccessAppliesServiceDefaults(t *testing.T) {
	router := newChannelHandlerTestRouter(newChannelHandlerRepoStub())

	req := httptest.NewRequest(http.MethodPost, "/admin/channels", bytes.NewBufferString(`{
		"name":"primary",
		"group_ids":[11],
		"model_mapping":{"anthropic":{"claude-sonnet-4":"claude-4-upstream"}},
		"model_pricing":[{"models":["claude-sonnet-4"],"input_price":0.01,"output_price":0.02}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int             `json:"code"`
		Data channelResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.NotZero(t, resp.Data.ID)
	require.Equal(t, "primary", resp.Data.Name)
	require.Equal(t, service.BillingModelSourceChannelMapped, resp.Data.BillingModelSource)
	require.Len(t, resp.Data.GroupIDs, 1)
	require.Equal(t, int64(11), resp.Data.GroupIDs[0])
	require.Len(t, resp.Data.ModelPricing, 1)
	require.Equal(t, service.PlatformAnthropic, resp.Data.ModelPricing[0].Platform)
	require.Equal(t, string(service.BillingModeToken), resp.Data.ModelPricing[0].BillingMode)
	require.Equal(t, "claude-4-upstream", resp.Data.ModelMapping[service.PlatformAnthropic]["claude-sonnet-4"])
}

func TestChannelHandlerCreateRejectsMissingGroups(t *testing.T) {
	repo := newChannelHandlerRepoStub()
	repo.getGroupPlatformsFn = func(context.Context, []int64) (map[int64]string, error) {
		return map[int64]string{}, nil
	}
	router := newChannelHandlerTestRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/admin/channels", bytes.NewBufferString(`{
		"name":"primary",
		"group_ids":[999999]
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	var resp struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "GROUP_NOT_FOUND", resp.Reason)
}

func TestChannelHandlerUpdateInvalidID(t *testing.T) {
	router := newChannelHandlerTestRouter(newChannelHandlerRepoStub())

	req := httptest.NewRequest(http.MethodPut, "/admin/channels/nope", bytes.NewBufferString(`{"name":"updated"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "INVALID_CHANNEL_ID", resp.Reason)
}

func TestChannelHandlerUpdateSuccess(t *testing.T) {
	now := time.Date(2025, 7, 1, 8, 0, 0, 0, time.UTC)
	repo := newChannelHandlerRepoStub(service.Channel{
		ID:        3,
		Name:      "primary",
		Status:    service.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		ModelPricing: []service.ChannelModelPricing{{
			Platform:    service.PlatformAnthropic,
			Models:      []string{"claude-sonnet-4"},
			BillingMode: service.BillingModeToken,
		}},
	})
	router := newChannelHandlerTestRouter(repo)

	req := httptest.NewRequest(http.MethodPut, "/admin/channels/3", bytes.NewBufferString(`{
		"name":"renamed",
		"restrict_models":true,
		"model_pricing":[{"platform":"openai","models":["gpt-5"],"billing_mode":"per_request","per_request_price":1.2}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int             `json:"code"`
		Data channelResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, "renamed", resp.Data.Name)
	require.True(t, resp.Data.RestrictModels)
	require.Len(t, resp.Data.ModelPricing, 1)
	require.Equal(t, service.PlatformOpenAI, resp.Data.ModelPricing[0].Platform)
	require.Equal(t, string(service.BillingModePerRequest), resp.Data.ModelPricing[0].BillingMode)
	require.Equal(t, 1.2, *resp.Data.ModelPricing[0].PerRequestPrice)
}

func TestChannelHandlerDeleteInvalidID(t *testing.T) {
	router := newChannelHandlerTestRouter(newChannelHandlerRepoStub())

	req := httptest.NewRequest(http.MethodDelete, "/admin/channels/bad-id", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "INVALID_CHANNEL_ID", resp.Reason)
}

func TestChannelHandlerDeleteSuccess(t *testing.T) {
	repo := newChannelHandlerRepoStub(service.Channel{
		ID:     8,
		Name:   "primary",
		Status: service.StatusActive,
	})
	router := newChannelHandlerTestRouter(repo)

	req := httptest.NewRequest(http.MethodDelete, "/admin/channels/8", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []int64{8}, repo.deletedIDs)
	_, exists := repo.channels[8]
	require.False(t, exists)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, "Channel deleted successfully", resp.Data.Message)
}

func TestChannelHandlerGetModelDefaultPricingMissingModel(t *testing.T) {
	router := newChannelHandlerWithBillingTestRouter(newChannelHandlerRepoStub(), service.NewBillingService(nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/admin/channels/model-pricing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Reason   string            `json:"reason"`
		Metadata map[string]string `json:"metadata"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "MISSING_PARAMETER", resp.Reason)
	require.Equal(t, "model", resp.Metadata["param"])
}

func TestChannelHandlerGetModelDefaultPricingFound(t *testing.T) {
	router := newChannelHandlerWithBillingTestRouter(newChannelHandlerRepoStub(), service.NewBillingService(nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/admin/channels/model-pricing?model=claude-sonnet-4", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Found            bool    `json:"found"`
			InputPrice       float64 `json:"input_price"`
			OutputPrice      float64 `json:"output_price"`
			CacheWritePrice  float64 `json:"cache_write_price"`
			CacheReadPrice   float64 `json:"cache_read_price"`
			ImageOutputPrice float64 `json:"image_output_price"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.True(t, resp.Data.Found)
	require.InDelta(t, 3e-6, resp.Data.InputPrice, 1e-12)
	require.InDelta(t, 15e-6, resp.Data.OutputPrice, 1e-12)
	require.InDelta(t, 3.75e-6, resp.Data.CacheWritePrice, 1e-12)
	require.InDelta(t, 0.3e-6, resp.Data.CacheReadPrice, 1e-12)
	require.Zero(t, resp.Data.ImageOutputPrice)
}

func TestChannelHandlerGetModelDefaultPricingNotFound(t *testing.T) {
	router := newChannelHandlerWithBillingTestRouter(newChannelHandlerRepoStub(), service.NewBillingService(nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/admin/channels/model-pricing?model=totally-unknown-model", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Found bool `json:"found"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.False(t, resp.Data.Found)
}
