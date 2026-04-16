package service

import (
	"context"
	"strings"
	"time"
)

const (
	openAIAccountScheduleLayerPreviousResponse = "previous_response_id"
	openAIAccountScheduleLayerSessionSticky    = "session_hash"
	openAIAccountScheduleLayerLoadBalance      = "load_balance"

	openAIAccountScheduleLoadBalanceStrategyFullScan        = "full_scan"
	openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot = "indexed_snapshot"
)

type OpenAIAccountScheduleRequest struct {
	GroupID            *int64
	SessionHash        string
	StickyAccountID    int64
	PreviousResponseID string
	RequestedModel     string
	RequiredTransport  OpenAIUpstreamTransport
	ExcludedIDs        map[int64]struct{}
}

type OpenAIAccountScheduleDecision struct {
	Layer               string
	StickyPreviousHit   bool
	StickySessionHit    bool
	HadStickyIntent     bool
	StickyMissFallback  bool
	CandidateCount      int
	TopK                int
	LatencyMs           int64
	LoadSkew            float64
	LoadBalanceStrategy string
	SelectedAccountID   int64
	SelectedAccountType string
}

type OpenAIAccountSchedulerMetricsSnapshot struct {
	SelectTotal                    int64   `json:"select_total"`
	StickyPreviousHitTotal         int64   `json:"sticky_previous_hit_total"`
	StickySessionHitTotal          int64   `json:"sticky_session_hit_total"`
	StickyIntentTotal              int64   `json:"sticky_intent_total"`
	NonStickyIntentTotal           int64   `json:"non_sticky_intent_total"`
	StickyMissFallbackTotal        int64   `json:"sticky_miss_fallback_total"`
	LoadBalanceSelectTotal         int64   `json:"load_balance_select_total"`
	IndexedLoadBalanceSelectTotal  int64   `json:"indexed_load_balance_select_total"`
	FullScanLoadBalanceSelectTotal int64   `json:"full_scan_load_balance_select_total"`
	StickyMissIndexedSelectTotal   int64   `json:"sticky_miss_indexed_select_total"`
	AccountSwitchTotal             int64   `json:"account_switch_total"`
	SchedulerLatencyMsTotal        int64   `json:"scheduler_latency_ms_total"`
	SchedulerLatencyMsAvg          float64 `json:"scheduler_latency_ms_avg"`
	StickyHitRatio                 float64 `json:"sticky_hit_ratio"`
	StickyIntentHitRate            float64 `json:"sticky_intent_hit_rate"`
	StickyIntentMissRate           float64 `json:"sticky_intent_miss_rate"`
	NonStickyIntentShare           float64 `json:"non_sticky_intent_share"`
	AccountSwitchRate              float64 `json:"account_switch_rate"`
	LoadSkewAvg                    float64 `json:"load_skew_avg"`
	IndexedLoadBalanceShare        float64 `json:"indexed_load_balance_share"`
	StickyMissIndexedShare         float64 `json:"sticky_miss_indexed_share"`
	RuntimeStatsAccountCount       int     `json:"runtime_stats_account_count"`
}

type OpenAIAccountScheduler interface {
	Select(ctx context.Context, req OpenAIAccountScheduleRequest) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error)
	ReportResult(accountID int64, success bool, firstTokenMs *int)
	ReportSwitch()
	SnapshotMetrics() OpenAIAccountSchedulerMetricsSnapshot
}

type defaultOpenAIAccountScheduler struct {
	service *OpenAIGatewayService
	metrics openAIAccountSchedulerMetrics
	stats   *openAIAccountRuntimeStats
}

func newDefaultOpenAIAccountScheduler(service *OpenAIGatewayService, stats *openAIAccountRuntimeStats) OpenAIAccountScheduler {
	if stats == nil {
		stats = newOpenAIAccountRuntimeStats()
	}
	return &defaultOpenAIAccountScheduler{
		service: service,
		stats:   stats,
	}
}

func (s *defaultOpenAIAccountScheduler) Select(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
	decision := OpenAIAccountScheduleDecision{}
	start := time.Now()
	defer func() {
		decision.LatencyMs = time.Since(start).Milliseconds()
		s.metrics.recordSelect(decision)
	}()

	previousResponseID := strings.TrimSpace(req.PreviousResponseID)
	decision.HadStickyIntent = previousResponseID != ""
	if previousResponseID != "" {
		selection, err := s.service.selectAccountByPreviousResponseIDForScheduler(
			ctx,
			req.GroupID,
			previousResponseID,
			req.RequestedModel,
			req.ExcludedIDs,
			req.RequiredTransport,
		)
		if err != nil {
			return nil, decision, err
		}
		selection = s.filterSchedulerSelectionByTransport(selection, req.RequiredTransport)
		if s.applySchedulerSelectionDecision(ctx, req, selection, openAISelectionDecisionSpec{
			layer:             openAIAccountScheduleLayerPreviousResponse,
			stickyPreviousHit: true,
			bindSession:       true,
		}, &decision) {
			return selection, decision, nil
		}
	}

	selection, sessionStickyIntent, err := s.selectBySessionHash(ctx, req)
	if err != nil {
		return nil, decision, err
	}
	decision.HadStickyIntent = decision.HadStickyIntent || sessionStickyIntent
	if s.applySchedulerSelectionDecision(ctx, req, selection, openAISelectionDecisionSpec{
		layer:            openAIAccountScheduleLayerSessionSticky,
		stickySessionHit: true,
	}, &decision) {
		return selection, decision, nil
	}

	selection, candidateCount, topK, loadSkew, loadBalanceStrategy, err := s.selectByLoadBalance(ctx, req)
	decision.Layer = openAIAccountScheduleLayerLoadBalance
	decision.StickyMissFallback = decision.HadStickyIntent
	decision.CandidateCount = candidateCount
	decision.TopK = topK
	decision.LoadSkew = loadSkew
	decision.LoadBalanceStrategy = loadBalanceStrategy
	if err != nil {
		return nil, decision, err
	}
	s.applySchedulerSelectionDecision(ctx, req, selection, openAISelectionDecisionSpec{
		layer: openAIAccountScheduleLayerLoadBalance,
	}, &decision)
	return selection, decision, nil
}
