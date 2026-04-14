package service

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type OpenAICodexCompatibilitySummarySnapshot struct {
	OfficialRequestTotal               int64   `json:"official_request_total"`
	OfficialResponsesHTTPRequests      int64   `json:"official_responses_http_requests_total"`
	OfficialResponsesWebSocketRequests int64   `json:"official_responses_websocket_requests_total"`
	ChainSelectionTotal                int64   `json:"chain_selection_total"`
	ChainHitTotal                      int64   `json:"chain_hit_total"`
	ChainPreviousResponseHitTotal      int64   `json:"chain_previous_response_hit_total"`
	ChainSessionHitTotal               int64   `json:"chain_session_hit_total"`
	ChainHitRate                       float64 `json:"chain_hit_rate"`
	ChainPreviousResponseHitRate       float64 `json:"chain_previous_response_hit_rate"`
	ChainSessionHitRate                float64 `json:"chain_session_hit_rate"`
	SessionHTTPFallbackHitRate         float64 `json:"session_http_fallback_hit_rate"`
	SessionTransportHTTPDowngradeRate  float64 `json:"session_transport_http_downgrade_rate"`
	MinimalRewriteAppliedTotal         int64   `json:"minimal_rewrite_applied_total"`
	MinimalRewriteRate                 float64 `json:"minimal_rewrite_rate"`
}

type OpenAICodexCompatibilityVersionMetricsSnapshot struct {
	Version                         string  `json:"version"`
	OfficialRequestTotal            int64   `json:"official_request_total"`
	ResponsesHTTPRequestsTotal      int64   `json:"responses_http_requests_total"`
	ResponsesWebSocketRequestsTotal int64   `json:"responses_websocket_requests_total"`
	WarmupTotal                     int64   `json:"warmup_total"`
	ChainSelectionTotal             int64   `json:"chain_selection_total"`
	ChainHitTotal                   int64   `json:"chain_hit_total"`
	ChainHitRate                    float64 `json:"chain_hit_rate"`
}

type OpenAICodexCompatibilityTransportMetricsSnapshot struct {
	Transport                  string  `json:"transport"`
	OfficialRequestTotal       int64   `json:"official_request_total"`
	WarmupTotal                int64   `json:"warmup_total"`
	ChainSelectionTotal        int64   `json:"chain_selection_total"`
	ChainHitTotal              int64   `json:"chain_hit_total"`
	ChainHitRate               float64 `json:"chain_hit_rate"`
	MinimalRewriteAppliedTotal int64   `json:"minimal_rewrite_applied_total"`
}

type OpenAICodexCompatibilityMetricsSnapshot struct {
	TransportWarmupTotal                  int64                                              `json:"transport_warmup_total"`
	TransportFallbackCoolingHitTotal      int64                                              `json:"transport_fallback_cooling_hit_total"`
	SessionPreferredTransportHitTotal     int64                                              `json:"session_preferred_transport_hit_total"`
	SessionPreferredTransportWSHitTotal   int64                                              `json:"session_preferred_transport_ws_hit_total"`
	SessionPreferredTransportHTTPHitTotal int64                                              `json:"session_preferred_transport_http_hit_total"`
	SessionHTTPFallbackHitTotal           int64                                              `json:"session_http_fallback_hit_total"`
	SessionTransportBindWSTotal           int64                                              `json:"session_transport_bind_ws_total"`
	SessionTransportBindHTTPTotal         int64                                              `json:"session_transport_bind_http_total"`
	SessionTransportWarmupIgnoredTotal    int64                                              `json:"session_transport_warmup_ignored_total"`
	SessionTransportHTTPDowngradeTotal    int64                                              `json:"session_transport_http_downgrade_total"`
	RecoveryFailoverExhaustAppliedTotal   int64                                              `json:"recovery_failover_exhaust_applied_total"`
	RecoveryFailoverExhaustSkippedTotal   int64                                              `json:"recovery_failover_exhaust_skipped_total"`
	RecoveryTransportCooldownAppliedTotal int64                                              `json:"recovery_transport_cooldown_applied_total"`
	RecoveryTransportCooldownSkippedTotal int64                                              `json:"recovery_transport_cooldown_skipped_total"`
	RecoveryAccountSwitchAppliedTotal     int64                                              `json:"recovery_account_switch_applied_total"`
	RecoveryAccountSwitchSkippedTotal     int64                                              `json:"recovery_account_switch_skipped_total"`
	RecoveryWSRetryTotal                  int64                                              `json:"recovery_ws_retry_total"`
	RecoveryHTTPRetryTotal                int64                                              `json:"recovery_http_retry_total"`
	RecoveryPreviousResponseAppliedTotal  int64                                              `json:"recovery_previous_response_applied_total"`
	RecoveryPreviousResponseSkippedTotal  int64                                              `json:"recovery_previous_response_skipped_total"`
	RecoveryInvalidEncryptedAppliedTotal  int64                                              `json:"recovery_invalid_encrypted_applied_total"`
	RecoveryInvalidEncryptedSkippedTotal  int64                                              `json:"recovery_invalid_encrypted_skipped_total"`
	RecoveryDropPreviousResponseIDTotal   int64                                              `json:"recovery_drop_previous_response_id_total"`
	RecoveryTrimEncryptedReasoningTotal   int64                                              `json:"recovery_trim_encrypted_reasoning_total"`
	Summary                               OpenAICodexCompatibilitySummarySnapshot            `json:"summary"`
	ByVersion                             []OpenAICodexCompatibilityVersionMetricsSnapshot   `json:"by_version"`
	ByTransport                           []OpenAICodexCompatibilityTransportMetricsSnapshot `json:"by_transport"`
}

type codexCompatibilityDimensionTotals struct {
	OfficialRequestTotal            int64
	ResponsesHTTPRequestsTotal      int64
	ResponsesWebSocketRequestsTotal int64
	WarmupTotal                     int64
	ChainSelectionTotal             int64
	ChainHitTotal                   int64
	MinimalRewriteAppliedTotal      int64
}

var (
	openAICodexTransportWarmupTotal                    atomic.Int64
	openAICodexTransportFallbackCoolingHitTotal        atomic.Int64
	openAICodexSessionPreferredTransportHitTotal       atomic.Int64
	openAICodexSessionPreferredTransportWSHitTotal     atomic.Int64
	openAICodexSessionPreferredTransportHTTPHitTotal   atomic.Int64
	openAICodexSessionHTTPFallbackHitTotal             atomic.Int64
	openAICodexSessionTransportBindWSTotal             atomic.Int64
	openAICodexSessionTransportBindHTTPTotal           atomic.Int64
	openAICodexSessionTransportWarmupIgnoredTotal      atomic.Int64
	openAICodexSessionTransportHTTPDowngradeTotal      atomic.Int64
	openAICodexRecoveryFailoverExhaustAppliedTotal     atomic.Int64
	openAICodexRecoveryFailoverExhaustSkippedTotal     atomic.Int64
	openAICodexRecoveryTransportCooldownAppliedTotal   atomic.Int64
	openAICodexRecoveryTransportCooldownSkippedTotal   atomic.Int64
	openAICodexRecoveryAccountSwitchAppliedTotal       atomic.Int64
	openAICodexRecoveryAccountSwitchSkippedTotal       atomic.Int64
	openAICodexRecoveryWSRetryTotal                    atomic.Int64
	openAICodexRecoveryHTTPRetryTotal                  atomic.Int64
	openAICodexRecoveryPreviousResponseAppliedTotal    atomic.Int64
	openAICodexRecoveryPreviousResponseSkippedTotal    atomic.Int64
	openAICodexRecoveryInvalidEncryptedAppliedTotal    atomic.Int64
	openAICodexRecoveryInvalidEncryptedSkippedTotal    atomic.Int64
	openAICodexRecoveryDropPreviousResponseIDTotal     atomic.Int64
	openAICodexRecoveryTrimEncryptedReasoningTotal     atomic.Int64
	openAICodexOfficialRequestTotal                    atomic.Int64
	openAICodexOfficialResponsesHTTPRequestsTotal      atomic.Int64
	openAICodexOfficialResponsesWebSocketRequestsTotal atomic.Int64
	openAICodexChainSelectionTotal                     atomic.Int64
	openAICodexChainHitTotal                           atomic.Int64
	openAICodexChainPreviousResponseHitTotal           atomic.Int64
	openAICodexChainSessionHitTotal                    atomic.Int64
	openAICodexMinimalRewriteAppliedTotal              atomic.Int64

	openAICodexCompatibilityMetricsMu sync.Mutex
	openAICodexVersionMetrics         = map[string]codexCompatibilityDimensionTotals{}
	openAICodexTransportMetrics       = map[string]codexCompatibilityDimensionTotals{}
)

func recordOpenAICodexTransportWarmup() {
	openAICodexTransportWarmupTotal.Add(1)
}

func recordOpenAICodexTransportFallbackCoolingHit() {
	openAICodexTransportFallbackCoolingHitTotal.Add(1)
}

func recordOpenAICodexSessionPreferredTransportHit(transport OpenAIUpstreamTransport, httpFallback bool) {
	openAICodexSessionPreferredTransportHitTotal.Add(1)
	switch normalizeOpenAIWSSessionTransport(transport) {
	case OpenAIUpstreamTransportHTTPSSE:
		openAICodexSessionPreferredTransportHTTPHitTotal.Add(1)
		if httpFallback {
			openAICodexSessionHTTPFallbackHitTotal.Add(1)
		}
	case OpenAIUpstreamTransportResponsesWebsocket, OpenAIUpstreamTransportResponsesWebsocketV2:
		openAICodexSessionPreferredTransportWSHitTotal.Add(1)
	}
}

func recordOpenAICodexSessionTransportBind(transport OpenAIUpstreamTransport, warmupIgnored bool, httpDowngrade bool) {
	if warmupIgnored {
		openAICodexSessionTransportWarmupIgnoredTotal.Add(1)
		return
	}
	switch normalizeOpenAIWSSessionTransport(transport) {
	case OpenAIUpstreamTransportHTTPSSE:
		openAICodexSessionTransportBindHTTPTotal.Add(1)
		if httpDowngrade {
			openAICodexSessionTransportHTTPDowngradeTotal.Add(1)
		}
	case OpenAIUpstreamTransportResponsesWebsocket, OpenAIUpstreamTransportResponsesWebsocketV2:
		openAICodexSessionTransportBindWSTotal.Add(1)
	}
}

func ObserveOpenAICodexRequestProfile(profile CodexRequestProfile) {
	if !profile.OfficialClient {
		return
	}

	openAICodexOfficialRequestTotal.Add(1)
	dimensionKey := normalizeCodexCompatibilityTransportKey(profile.WireAPI)
	switch dimensionKey {
	case string(CodexWireAPIResponsesHTTP):
		openAICodexOfficialResponsesHTTPRequestsTotal.Add(1)
	case string(CodexWireAPIResponsesWebSocket):
		openAICodexOfficialResponsesWebSocketRequestsTotal.Add(1)
	}

	versionKey := normalizeCodexCompatibilityVersionKey(profile.CodexVersion)
	openAICodexCompatibilityMetricsMu.Lock()
	defer openAICodexCompatibilityMetricsMu.Unlock()

	versionTotals := openAICodexVersionMetrics[versionKey]
	versionTotals.OfficialRequestTotal++
	switch dimensionKey {
	case string(CodexWireAPIResponsesHTTP):
		versionTotals.ResponsesHTTPRequestsTotal++
	case string(CodexWireAPIResponsesWebSocket):
		versionTotals.ResponsesWebSocketRequestsTotal++
	}
	if profile.Warmup {
		versionTotals.WarmupTotal++
	}
	openAICodexVersionMetrics[versionKey] = versionTotals

	transportTotals := openAICodexTransportMetrics[dimensionKey]
	transportTotals.OfficialRequestTotal++
	if profile.Warmup {
		transportTotals.WarmupTotal++
	}
	openAICodexTransportMetrics[dimensionKey] = transportTotals
}

func ObserveOpenAICodexSchedulingDecision(profile CodexRequestProfile, decision OpenAIAccountScheduleDecision) {
	if !profile.OfficialClient {
		return
	}

	openAICodexChainSelectionTotal.Add(1)
	chainHit := decision.StickyPreviousHit || decision.StickySessionHit
	if chainHit {
		openAICodexChainHitTotal.Add(1)
	}
	if decision.StickyPreviousHit {
		openAICodexChainPreviousResponseHitTotal.Add(1)
	}
	if decision.StickySessionHit {
		openAICodexChainSessionHitTotal.Add(1)
	}

	versionKey := normalizeCodexCompatibilityVersionKey(profile.CodexVersion)
	transportKey := normalizeCodexCompatibilityTransportKey(profile.WireAPI)
	openAICodexCompatibilityMetricsMu.Lock()
	defer openAICodexCompatibilityMetricsMu.Unlock()

	versionTotals := openAICodexVersionMetrics[versionKey]
	versionTotals.ChainSelectionTotal++
	if chainHit {
		versionTotals.ChainHitTotal++
	}
	openAICodexVersionMetrics[versionKey] = versionTotals

	transportTotals := openAICodexTransportMetrics[transportKey]
	transportTotals.ChainSelectionTotal++
	if chainHit {
		transportTotals.ChainHitTotal++
	}
	openAICodexTransportMetrics[transportKey] = transportTotals
}

func recordOpenAICodexRecoveryDecision(decision CodexRecoveryDecision) {
	if !decision.TrackCompatibilityMetrics {
		return
	}

	transportKey := normalizeCodexCompatibilityRecoveryTransportKey(decision.Transport)
	minimalRewriteApplied := false

	switch decision.Reason {
	case codexRecoveryReasonFailover:
		if decision.ExhaustFailover {
			openAICodexRecoveryFailoverExhaustAppliedTotal.Add(1)
		} else {
			openAICodexRecoveryFailoverExhaustSkippedTotal.Add(1)
		}
		return
	case codexRecoveryReasonTransportFailure:
		if decision.Applied {
			openAICodexRecoveryTransportCooldownAppliedTotal.Add(1)
		} else {
			openAICodexRecoveryTransportCooldownSkippedTotal.Add(1)
		}
		return
	case codexRecoveryReasonAccountSwitch:
		if decision.Applied {
			openAICodexRecoveryAccountSwitchAppliedTotal.Add(1)
		} else {
			openAICodexRecoveryAccountSwitchSkippedTotal.Add(1)
		}
		return
	case codexRecoveryReasonPreviousResponseNotFound:
		if decision.Applied {
			openAICodexRecoveryPreviousResponseAppliedTotal.Add(1)
			minimalRewriteApplied = true
		} else {
			openAICodexRecoveryPreviousResponseSkippedTotal.Add(1)
		}
	case codexRecoveryReasonInvalidEncryptedContent:
		if decision.Applied {
			openAICodexRecoveryInvalidEncryptedAppliedTotal.Add(1)
			minimalRewriteApplied = true
		} else {
			openAICodexRecoveryInvalidEncryptedSkippedTotal.Add(1)
		}
	default:
		return
	}

	if decision.DroppedPreviousResponseID {
		openAICodexRecoveryDropPreviousResponseIDTotal.Add(1)
	}
	if decision.TrimmedEncryptedReasoning {
		openAICodexRecoveryTrimEncryptedReasoningTotal.Add(1)
	}
	if minimalRewriteApplied {
		openAICodexMinimalRewriteAppliedTotal.Add(1)
		openAICodexCompatibilityMetricsMu.Lock()
		transportTotals := openAICodexTransportMetrics[transportKey]
		transportTotals.MinimalRewriteAppliedTotal++
		openAICodexTransportMetrics[transportKey] = transportTotals
		openAICodexCompatibilityMetricsMu.Unlock()
	}
	if !decision.Applied {
		return
	}
	switch decision.Transport {
	case OpenAIUpstreamTransportHTTPSSE:
		openAICodexRecoveryHTTPRetryTotal.Add(1)
	case OpenAIUpstreamTransportResponsesWebsocket, OpenAIUpstreamTransportResponsesWebsocketV2:
		openAICodexRecoveryWSRetryTotal.Add(1)
	}
}

func SnapshotOpenAICodexCompatibilityMetrics() OpenAICodexCompatibilityMetricsSnapshot {
	versionSnapshots, transportSnapshots := snapshotOpenAICodexCompatibilityDimensions()
	summary := buildOpenAICodexCompatibilitySummary(OpenAICodexCompatibilityMetricsSnapshot{
		SessionPreferredTransportHTTPHitTotal: openAICodexSessionPreferredTransportHTTPHitTotal.Load(),
		SessionHTTPFallbackHitTotal:           openAICodexSessionHTTPFallbackHitTotal.Load(),
		SessionTransportBindHTTPTotal:         openAICodexSessionTransportBindHTTPTotal.Load(),
		SessionTransportHTTPDowngradeTotal:    openAICodexSessionTransportHTTPDowngradeTotal.Load(),
	})

	return OpenAICodexCompatibilityMetricsSnapshot{
		TransportWarmupTotal:                  openAICodexTransportWarmupTotal.Load(),
		TransportFallbackCoolingHitTotal:      openAICodexTransportFallbackCoolingHitTotal.Load(),
		SessionPreferredTransportHitTotal:     openAICodexSessionPreferredTransportHitTotal.Load(),
		SessionPreferredTransportWSHitTotal:   openAICodexSessionPreferredTransportWSHitTotal.Load(),
		SessionPreferredTransportHTTPHitTotal: openAICodexSessionPreferredTransportHTTPHitTotal.Load(),
		SessionHTTPFallbackHitTotal:           openAICodexSessionHTTPFallbackHitTotal.Load(),
		SessionTransportBindWSTotal:           openAICodexSessionTransportBindWSTotal.Load(),
		SessionTransportBindHTTPTotal:         openAICodexSessionTransportBindHTTPTotal.Load(),
		SessionTransportWarmupIgnoredTotal:    openAICodexSessionTransportWarmupIgnoredTotal.Load(),
		SessionTransportHTTPDowngradeTotal:    openAICodexSessionTransportHTTPDowngradeTotal.Load(),
		RecoveryFailoverExhaustAppliedTotal:   openAICodexRecoveryFailoverExhaustAppliedTotal.Load(),
		RecoveryFailoverExhaustSkippedTotal:   openAICodexRecoveryFailoverExhaustSkippedTotal.Load(),
		RecoveryTransportCooldownAppliedTotal: openAICodexRecoveryTransportCooldownAppliedTotal.Load(),
		RecoveryTransportCooldownSkippedTotal: openAICodexRecoveryTransportCooldownSkippedTotal.Load(),
		RecoveryAccountSwitchAppliedTotal:     openAICodexRecoveryAccountSwitchAppliedTotal.Load(),
		RecoveryAccountSwitchSkippedTotal:     openAICodexRecoveryAccountSwitchSkippedTotal.Load(),
		RecoveryWSRetryTotal:                  openAICodexRecoveryWSRetryTotal.Load(),
		RecoveryHTTPRetryTotal:                openAICodexRecoveryHTTPRetryTotal.Load(),
		RecoveryPreviousResponseAppliedTotal:  openAICodexRecoveryPreviousResponseAppliedTotal.Load(),
		RecoveryPreviousResponseSkippedTotal:  openAICodexRecoveryPreviousResponseSkippedTotal.Load(),
		RecoveryInvalidEncryptedAppliedTotal:  openAICodexRecoveryInvalidEncryptedAppliedTotal.Load(),
		RecoveryInvalidEncryptedSkippedTotal:  openAICodexRecoveryInvalidEncryptedSkippedTotal.Load(),
		RecoveryDropPreviousResponseIDTotal:   openAICodexRecoveryDropPreviousResponseIDTotal.Load(),
		RecoveryTrimEncryptedReasoningTotal:   openAICodexRecoveryTrimEncryptedReasoningTotal.Load(),
		Summary:                               summary,
		ByVersion:                             versionSnapshots,
		ByTransport:                           transportSnapshots,
	}
}

func buildOpenAICodexCompatibilitySummary(snapshot OpenAICodexCompatibilityMetricsSnapshot) OpenAICodexCompatibilitySummarySnapshot {
	officialRequestTotal := openAICodexOfficialRequestTotal.Load()
	chainSelectionTotal := openAICodexChainSelectionTotal.Load()
	chainHitTotal := openAICodexChainHitTotal.Load()
	chainPreviousHitTotal := openAICodexChainPreviousResponseHitTotal.Load()
	chainSessionHitTotal := openAICodexChainSessionHitTotal.Load()
	minimalRewriteAppliedTotal := openAICodexMinimalRewriteAppliedTotal.Load()

	return OpenAICodexCompatibilitySummarySnapshot{
		OfficialRequestTotal:               officialRequestTotal,
		OfficialResponsesHTTPRequests:      openAICodexOfficialResponsesHTTPRequestsTotal.Load(),
		OfficialResponsesWebSocketRequests: openAICodexOfficialResponsesWebSocketRequestsTotal.Load(),
		ChainSelectionTotal:                chainSelectionTotal,
		ChainHitTotal:                      chainHitTotal,
		ChainPreviousResponseHitTotal:      chainPreviousHitTotal,
		ChainSessionHitTotal:               chainSessionHitTotal,
		ChainHitRate:                       ratioOfInt64(chainHitTotal, chainSelectionTotal),
		ChainPreviousResponseHitRate:       ratioOfInt64(chainPreviousHitTotal, chainSelectionTotal),
		ChainSessionHitRate:                ratioOfInt64(chainSessionHitTotal, chainSelectionTotal),
		SessionHTTPFallbackHitRate:         ratioOfInt64(snapshot.SessionHTTPFallbackHitTotal, snapshot.SessionPreferredTransportHTTPHitTotal),
		SessionTransportHTTPDowngradeRate:  ratioOfInt64(snapshot.SessionTransportHTTPDowngradeTotal, snapshot.SessionTransportBindHTTPTotal),
		MinimalRewriteAppliedTotal:         minimalRewriteAppliedTotal,
		MinimalRewriteRate:                 ratioOfInt64(minimalRewriteAppliedTotal, officialRequestTotal),
	}
}

func snapshotOpenAICodexCompatibilityDimensions() ([]OpenAICodexCompatibilityVersionMetricsSnapshot, []OpenAICodexCompatibilityTransportMetricsSnapshot) {
	openAICodexCompatibilityMetricsMu.Lock()
	defer openAICodexCompatibilityMetricsMu.Unlock()

	versionSnapshots := make([]OpenAICodexCompatibilityVersionMetricsSnapshot, 0, len(openAICodexVersionMetrics))
	for version, totals := range openAICodexVersionMetrics {
		versionSnapshots = append(versionSnapshots, OpenAICodexCompatibilityVersionMetricsSnapshot{
			Version:                         version,
			OfficialRequestTotal:            totals.OfficialRequestTotal,
			ResponsesHTTPRequestsTotal:      totals.ResponsesHTTPRequestsTotal,
			ResponsesWebSocketRequestsTotal: totals.ResponsesWebSocketRequestsTotal,
			WarmupTotal:                     totals.WarmupTotal,
			ChainSelectionTotal:             totals.ChainSelectionTotal,
			ChainHitTotal:                   totals.ChainHitTotal,
			ChainHitRate:                    ratioOfInt64(totals.ChainHitTotal, totals.ChainSelectionTotal),
		})
	}
	sort.Slice(versionSnapshots, func(i, j int) bool {
		if versionSnapshots[i].OfficialRequestTotal == versionSnapshots[j].OfficialRequestTotal {
			return versionSnapshots[i].Version < versionSnapshots[j].Version
		}
		return versionSnapshots[i].OfficialRequestTotal > versionSnapshots[j].OfficialRequestTotal
	})

	transportSnapshots := make([]OpenAICodexCompatibilityTransportMetricsSnapshot, 0, len(openAICodexTransportMetrics))
	for transport, totals := range openAICodexTransportMetrics {
		transportSnapshots = append(transportSnapshots, OpenAICodexCompatibilityTransportMetricsSnapshot{
			Transport:                  transport,
			OfficialRequestTotal:       totals.OfficialRequestTotal,
			WarmupTotal:                totals.WarmupTotal,
			ChainSelectionTotal:        totals.ChainSelectionTotal,
			ChainHitTotal:              totals.ChainHitTotal,
			ChainHitRate:               ratioOfInt64(totals.ChainHitTotal, totals.ChainSelectionTotal),
			MinimalRewriteAppliedTotal: totals.MinimalRewriteAppliedTotal,
		})
	}
	sort.Slice(transportSnapshots, func(i, j int) bool {
		if transportSnapshots[i].OfficialRequestTotal == transportSnapshots[j].OfficialRequestTotal {
			return transportSnapshots[i].Transport < transportSnapshots[j].Transport
		}
		return transportSnapshots[i].OfficialRequestTotal > transportSnapshots[j].OfficialRequestTotal
	})

	return versionSnapshots, transportSnapshots
}

func normalizeCodexCompatibilityVersionKey(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "unknown"
	}
	return version
}

func normalizeCodexCompatibilityTransportKey(wireAPI CodexWireAPI) string {
	switch strings.TrimSpace(string(wireAPI)) {
	case string(CodexWireAPIResponsesHTTP):
		return string(CodexWireAPIResponsesHTTP)
	case string(CodexWireAPIResponsesWebSocket):
		return string(CodexWireAPIResponsesWebSocket)
	default:
		return "unknown"
	}
}

func normalizeCodexCompatibilityRecoveryTransportKey(transport OpenAIUpstreamTransport) string {
	switch normalizeOpenAIWSSessionTransport(transport) {
	case OpenAIUpstreamTransportHTTPSSE:
		return string(CodexWireAPIResponsesHTTP)
	case OpenAIUpstreamTransportResponsesWebsocket, OpenAIUpstreamTransportResponsesWebsocketV2:
		return string(CodexWireAPIResponsesWebSocket)
	default:
		return "unknown"
	}
}
