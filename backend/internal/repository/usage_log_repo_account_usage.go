package repository

import (
	"context"
	"time"

	"github.com/senran-N/sub2api/internal/pkg/timezone"
	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"golang.org/x/sync/errgroup"
)

// AccountUsageHistory represents daily usage history for an account
type AccountUsageHistory = usagestats.AccountUsageHistory

// AccountUsageSummary represents summary statistics for an account
type AccountUsageSummary = usagestats.AccountUsageSummary

// AccountUsageStatsResponse represents the full usage statistics response for an account
type AccountUsageStatsResponse = usagestats.AccountUsageStatsResponse

// GetAccountUsageStats returns comprehensive usage statistics for an account over a time range
func (r *usageLogRepository) GetAccountUsageStats(ctx context.Context, accountID int64, startTime, endTime time.Time) (resp *AccountUsageStatsResponse, err error) {
	daysCount := int(endTime.Sub(startTime).Hours()/24) + 1
	if daysCount <= 0 {
		daysCount = 30
	}

	var (
		history              []AccountUsageHistory
		totalDurationMs      int64
		totalDurationSamples int64
		models               []ModelStat
		endpointStats        = usageLogEndpointStatsResult{
			Endpoints:         []EndpointStat{},
			UpstreamEndpoints: []EndpointStat{},
		}
	)
	loadHistory := func(queryCtx context.Context) error {
		var historyErr error
		history, totalDurationMs, totalDurationSamples, historyErr = r.loadAccountUsageHistory(queryCtx, accountID, startTime, endTime, daysCount)
		return historyErr
	}
	loadModels := func(queryCtx context.Context) {
		result, queryErr := r.GetModelStatsWithFilters(queryCtx, startTime, endTime, 0, 0, accountID, 0, nil, nil, nil)
		if queryErr != nil {
			return
		}
		models = result
	}
	loadEndpointStats := func(queryCtx context.Context) {
		endpointStats = r.collectEndpointStatsWithFilters(queryCtx, startTime, endTime, 0, 0, accountID, 0, "", nil, nil, nil, false)
	}
	if !r.shouldRunConcurrentReadQueries() {
		if err := loadHistory(ctx); err != nil {
			return nil, err
		}
		loadModels(ctx)
		loadEndpointStats(ctx)
	} else {
		queryGroup, queryCtx := errgroup.WithContext(ctx)
		queryGroup.Go(func() error {
			return loadHistory(queryCtx)
		})
		queryGroup.Go(func() error {
			loadModels(queryCtx)
			return nil
		})
		queryGroup.Go(func() error {
			loadEndpointStats(queryCtx)
			return nil
		})
		if err := queryGroup.Wait(); err != nil {
			return nil, err
		}
	}

	var totalAccountCost, totalUserCost, totalStandardCost float64
	var totalRequests, totalTokens int64
	var highestCostDay, highestRequestDay *AccountUsageHistory

	for i := range history {
		h := &history[i]
		totalAccountCost += h.ActualCost
		totalUserCost += h.UserCost
		totalStandardCost += h.Cost
		totalRequests += h.Requests
		totalTokens += h.Tokens

		if highestCostDay == nil || h.ActualCost > highestCostDay.ActualCost {
			highestCostDay = h
		}
		if highestRequestDay == nil || h.Requests > highestRequestDay.Requests {
			highestRequestDay = h
		}
	}

	actualDaysUsed := len(history)
	if actualDaysUsed == 0 {
		actualDaysUsed = 1
	}

	var avgDuration float64
	if totalDurationSamples > 0 {
		avgDuration = float64(totalDurationMs) / float64(totalDurationSamples)
	}

	summary := AccountUsageSummary{
		Days:              daysCount,
		ActualDaysUsed:    actualDaysUsed,
		TotalCost:         totalAccountCost,
		TotalUserCost:     totalUserCost,
		TotalStandardCost: totalStandardCost,
		TotalRequests:     totalRequests,
		TotalTokens:       totalTokens,
		AvgDailyCost:      totalAccountCost / float64(actualDaysUsed),
		AvgDailyUserCost:  totalUserCost / float64(actualDaysUsed),
		AvgDailyRequests:  float64(totalRequests) / float64(actualDaysUsed),
		AvgDailyTokens:    float64(totalTokens) / float64(actualDaysUsed),
		AvgDurationMs:     avgDuration,
	}

	todayStr := timezone.Now().Format("2006-01-02")
	for i := range history {
		if history[i].Date == todayStr {
			summary.Today = &struct {
				Date     string  `json:"date"`
				Cost     float64 `json:"cost"`
				UserCost float64 `json:"user_cost"`
				Requests int64   `json:"requests"`
				Tokens   int64   `json:"tokens"`
			}{
				Date:     history[i].Date,
				Cost:     history[i].ActualCost,
				UserCost: history[i].UserCost,
				Requests: history[i].Requests,
				Tokens:   history[i].Tokens,
			}
			break
		}
	}

	if highestCostDay != nil {
		summary.HighestCostDay = &struct {
			Date     string  `json:"date"`
			Label    string  `json:"label"`
			Cost     float64 `json:"cost"`
			UserCost float64 `json:"user_cost"`
			Requests int64   `json:"requests"`
		}{
			Date:     highestCostDay.Date,
			Label:    highestCostDay.Label,
			Cost:     highestCostDay.ActualCost,
			UserCost: highestCostDay.UserCost,
			Requests: highestCostDay.Requests,
		}
	}

	if highestRequestDay != nil {
		summary.HighestRequestDay = &struct {
			Date     string  `json:"date"`
			Label    string  `json:"label"`
			Requests int64   `json:"requests"`
			Cost     float64 `json:"cost"`
			UserCost float64 `json:"user_cost"`
		}{
			Date:     highestRequestDay.Date,
			Label:    highestRequestDay.Label,
			Requests: highestRequestDay.Requests,
			Cost:     highestRequestDay.ActualCost,
			UserCost: highestRequestDay.UserCost,
		}
	}

	resp = &AccountUsageStatsResponse{
		History:           history,
		Summary:           summary,
		Models:            models,
		Endpoints:         endpointStats.Endpoints,
		UpstreamEndpoints: endpointStats.UpstreamEndpoints,
	}
	return resp, nil
}

func (r *usageLogRepository) loadAccountUsageHistory(ctx context.Context, accountID int64, startTime, endTime time.Time, capacityHint int) (history []AccountUsageHistory, totalDurationMs int64, totalDurationSamples int64, err error) {
	query := `
		SELECT
			TO_CHAR(created_at, 'YYYY-MM-DD') as date,
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as tokens,
			COALESCE(SUM(total_cost), 0) as cost,
			COALESCE(SUM(total_cost * COALESCE(account_rate_multiplier, 1)), 0) as actual_cost,
			COALESCE(SUM(actual_cost), 0) as user_cost,
			COALESCE(SUM(duration_ms), 0) as total_duration_ms,
			COUNT(duration_ms) as duration_samples
		FROM usage_logs
		WHERE account_id = $1 AND created_at >= $2 AND created_at < $3
		GROUP BY date
		ORDER BY date ASC
	`

	rows, err := r.sql.QueryContext(ctx, query, accountID, startTime, endTime)
	if err != nil {
		return nil, 0, 0, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			history = nil
		}
	}()

	history = make([]AccountUsageHistory, 0, capacityHint)
	for rows.Next() {
		var date string
		var requests int64
		var tokens int64
		var cost float64
		var actualCost float64
		var userCost float64
		var durationMs int64
		var durationSamples int64
		if err := rows.Scan(&date, &requests, &tokens, &cost, &actualCost, &userCost, &durationMs, &durationSamples); err != nil {
			return nil, 0, 0, err
		}
		t, _ := time.Parse("2006-01-02", date)
		totalDurationMs += durationMs
		totalDurationSamples += durationSamples
		history = append(history, AccountUsageHistory{
			Date:       date,
			Label:      t.Format("01/02"),
			Requests:   requests,
			Tokens:     tokens,
			Cost:       cost,
			ActualCost: actualCost,
			UserCost:   userCost,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, err
	}
	return history, totalDurationMs, totalDurationSamples, nil
}
