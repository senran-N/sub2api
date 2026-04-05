package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senran-N/sub2api/internal/pkg/pagination"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type adminUsageListRepoStub struct {
	service.UsageLogRepository
	logs []service.UsageLog
}

func (s *adminUsageListRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters usagestats.UsageLogFilters) ([]service.UsageLog, *pagination.PaginationResult, error) {
	return s.logs, &pagination.PaginationResult{
		Total:    int64(len(s.logs)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    1,
	}, nil
}

func (s *adminUsageListRepoStub) GetStatsWithFilters(context.Context, usagestats.UsageLogFilters) (*usagestats.UsageStats, error) {
	return &usagestats.UsageStats{}, nil
}

func newAdminUsageChannelFieldsRouter(repo *adminUsageListRepoStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	usageSvc := service.NewUsageService(repo, nil, nil, nil)
	handler := NewUsageHandler(usageSvc, nil, nil, nil)
	router := gin.New()
	router.GET("/admin/usage", handler.List)
	return router
}

func TestAdminUsageListIncludesChannelFieldsForAdmin(t *testing.T) {
	channelID := int64(77)
	upstreamModel := "gpt-5.4-20260101"
	mappingChain := "gpt-5→gpt-5.4→gpt-5.4-20260101"
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	repo := &adminUsageListRepoStub{
		logs: []service.UsageLog{{
			ID:                1,
			UserID:            10,
			APIKeyID:          20,
			AccountID:         30,
			RequestID:         "req_admin_usage_channel",
			Model:             "gpt-5.4",
			RequestedModel:    "gpt-5",
			UpstreamModel:     &upstreamModel,
			ChannelID:         &channelID,
			ModelMappingChain: &mappingChain,
			InputTokens:       100,
			OutputTokens:      25,
			TotalCost:         0.8,
			ActualCost:        0.8,
			RateMultiplier:    1,
			CreatedAt:         now,
		}},
	}
	router := newAdminUsageChannelFieldsRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/admin/usage", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				Model             string  `json:"model"`
				UpstreamModel     *string `json:"upstream_model"`
				ChannelID         *int64  `json:"channel_id"`
				ModelMappingChain *string `json:"model_mapping_chain"`
			} `json:"items"`
			Total int64 `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, int64(1), resp.Data.Total)
	require.Len(t, resp.Data.Items, 1)
	require.Equal(t, "gpt-5", resp.Data.Items[0].Model)
	require.NotNil(t, resp.Data.Items[0].UpstreamModel)
	require.Equal(t, upstreamModel, *resp.Data.Items[0].UpstreamModel)
	require.NotNil(t, resp.Data.Items[0].ChannelID)
	require.Equal(t, channelID, *resp.Data.Items[0].ChannelID)
	require.NotNil(t, resp.Data.Items[0].ModelMappingChain)
	require.Equal(t, mappingChain, *resp.Data.Items[0].ModelMappingChain)
}
