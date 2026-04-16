package service

import (
	"math"
	"sync/atomic"
)

type openAIAccountSchedulerMetrics struct {
	selectTotal                    atomic.Int64
	stickyPreviousHitTotal         atomic.Int64
	stickySessionHitTotal          atomic.Int64
	stickyIntentTotal              atomic.Int64
	nonStickyIntentTotal           atomic.Int64
	stickyMissFallbackTotal        atomic.Int64
	loadBalanceSelectTotal         atomic.Int64
	indexedLoadBalanceSelectTotal  atomic.Int64
	fullScanLoadBalanceSelectTotal atomic.Int64
	stickyMissIndexedSelectTotal   atomic.Int64
	accountSwitchTotal             atomic.Int64
	latencyMsTotal                 atomic.Int64
	loadSkewMilliTotal             atomic.Int64
}

var defaultOpenAIAccountSchedulerMetrics openAIAccountSchedulerMetrics

func recordOpenAIAccountSchedulerSelect(target *openAIAccountSchedulerMetrics, decision OpenAIAccountScheduleDecision) {
	if target == nil {
		return
	}
	target.selectTotal.Add(1)
	target.latencyMsTotal.Add(decision.LatencyMs)
	target.loadSkewMilliTotal.Add(int64(math.Round(decision.LoadSkew * 1000)))
	if decision.StickyPreviousHit {
		target.stickyPreviousHitTotal.Add(1)
	}
	if decision.StickySessionHit {
		target.stickySessionHitTotal.Add(1)
	}
	if decision.HadStickyIntent {
		target.stickyIntentTotal.Add(1)
	} else {
		target.nonStickyIntentTotal.Add(1)
	}
	if decision.Layer == openAIAccountScheduleLayerLoadBalance {
		target.loadBalanceSelectTotal.Add(1)
		if decision.StickyMissFallback {
			target.stickyMissFallbackTotal.Add(1)
		}
		switch decision.LoadBalanceStrategy {
		case openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot:
			target.indexedLoadBalanceSelectTotal.Add(1)
			if decision.StickyMissFallback {
				target.stickyMissIndexedSelectTotal.Add(1)
			}
		case openAIAccountScheduleLoadBalanceStrategyFullScan:
			target.fullScanLoadBalanceSelectTotal.Add(1)
		}
	}
}

func (m *openAIAccountSchedulerMetrics) recordSelect(decision OpenAIAccountScheduleDecision) {
	recordOpenAIAccountSchedulerSelect(m, decision)
	recordOpenAIAccountSchedulerSelect(&defaultOpenAIAccountSchedulerMetrics, decision)
}

func recordOpenAIAccountSchedulerSwitch(target *openAIAccountSchedulerMetrics) {
	if target == nil {
		return
	}
	target.accountSwitchTotal.Add(1)
}

func (m *openAIAccountSchedulerMetrics) recordSwitch() {
	recordOpenAIAccountSchedulerSwitch(m)
	recordOpenAIAccountSchedulerSwitch(&defaultOpenAIAccountSchedulerMetrics)
}

func (s *defaultOpenAIAccountScheduler) ReportSwitch() {
	if s == nil {
		return
	}
	s.metrics.recordSwitch()
}

func (s *defaultOpenAIAccountScheduler) SnapshotMetrics() OpenAIAccountSchedulerMetricsSnapshot {
	if s == nil {
		return snapshotOpenAIAccountSchedulerMetrics(nil, 0)
	}
	return snapshotOpenAIAccountSchedulerMetrics(&s.metrics, s.stats.size())
}

func snapshotOpenAIAccountSchedulerMetrics(metrics *openAIAccountSchedulerMetrics, runtimeStatsCount int) OpenAIAccountSchedulerMetricsSnapshot {
	if metrics == nil {
		return OpenAIAccountSchedulerMetricsSnapshot{}
	}

	selectTotal := metrics.selectTotal.Load()
	prevHit := metrics.stickyPreviousHitTotal.Load()
	sessionHit := metrics.stickySessionHitTotal.Load()
	stickyIntentTotal := metrics.stickyIntentTotal.Load()
	stickyMissFallbackTotal := metrics.stickyMissFallbackTotal.Load()
	nonStickyIntentTotal := metrics.nonStickyIntentTotal.Load()
	switchTotal := metrics.accountSwitchTotal.Load()
	latencyTotal := metrics.latencyMsTotal.Load()
	loadSkewTotal := metrics.loadSkewMilliTotal.Load()
	indexedLoadBalanceTotal := metrics.indexedLoadBalanceSelectTotal.Load()
	stickyMissIndexedTotal := metrics.stickyMissIndexedSelectTotal.Load()

	snapshot := OpenAIAccountSchedulerMetricsSnapshot{
		SelectTotal:                    selectTotal,
		StickyPreviousHitTotal:         prevHit,
		StickySessionHitTotal:          sessionHit,
		StickyIntentTotal:              stickyIntentTotal,
		NonStickyIntentTotal:           nonStickyIntentTotal,
		StickyMissFallbackTotal:        stickyMissFallbackTotal,
		LoadBalanceSelectTotal:         metrics.loadBalanceSelectTotal.Load(),
		IndexedLoadBalanceSelectTotal:  indexedLoadBalanceTotal,
		FullScanLoadBalanceSelectTotal: metrics.fullScanLoadBalanceSelectTotal.Load(),
		StickyMissIndexedSelectTotal:   stickyMissIndexedTotal,
		AccountSwitchTotal:             switchTotal,
		SchedulerLatencyMsTotal:        latencyTotal,
		RuntimeStatsAccountCount:       runtimeStatsCount,
	}
	if selectTotal > 0 {
		snapshot.SchedulerLatencyMsAvg = float64(latencyTotal) / float64(selectTotal)
		snapshot.StickyHitRatio = float64(prevHit+sessionHit) / float64(selectTotal)
		snapshot.NonStickyIntentShare = float64(nonStickyIntentTotal) / float64(selectTotal)
		snapshot.AccountSwitchRate = float64(switchTotal) / float64(selectTotal)
		snapshot.LoadSkewAvg = float64(loadSkewTotal) / 1000 / float64(selectTotal)
		snapshot.IndexedLoadBalanceShare = float64(indexedLoadBalanceTotal) / float64(selectTotal)
		snapshot.StickyMissIndexedShare = float64(stickyMissIndexedTotal) / float64(selectTotal)
	}
	if stickyIntentTotal > 0 {
		snapshot.StickyIntentHitRate = float64(prevHit+sessionHit) / float64(stickyIntentTotal)
		snapshot.StickyIntentMissRate = float64(stickyMissFallbackTotal) / float64(stickyIntentTotal)
	}
	return snapshot
}

func snapshotDefaultOpenAIAccountSchedulerMetrics() OpenAIAccountSchedulerMetricsSnapshot {
	return snapshotOpenAIAccountSchedulerMetrics(&defaultOpenAIAccountSchedulerMetrics, 0)
}

func resetDefaultOpenAIAccountSchedulerMetrics() {
	defaultOpenAIAccountSchedulerMetrics = openAIAccountSchedulerMetrics{}
}
