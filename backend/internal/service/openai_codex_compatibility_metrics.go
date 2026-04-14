package service

import "sync/atomic"

type OpenAICodexCompatibilityMetricsSnapshot struct {
	TransportWarmupTotal                  int64 `json:"transport_warmup_total"`
	TransportFallbackCoolingHitTotal      int64 `json:"transport_fallback_cooling_hit_total"`
	SessionPreferredTransportHitTotal     int64 `json:"session_preferred_transport_hit_total"`
	SessionPreferredTransportWSHitTotal   int64 `json:"session_preferred_transport_ws_hit_total"`
	SessionPreferredTransportHTTPHitTotal int64 `json:"session_preferred_transport_http_hit_total"`
	SessionHTTPFallbackHitTotal           int64 `json:"session_http_fallback_hit_total"`
	SessionTransportBindWSTotal           int64 `json:"session_transport_bind_ws_total"`
	SessionTransportBindHTTPTotal         int64 `json:"session_transport_bind_http_total"`
	SessionTransportWarmupIgnoredTotal    int64 `json:"session_transport_warmup_ignored_total"`
	SessionTransportHTTPDowngradeTotal    int64 `json:"session_transport_http_downgrade_total"`
}

var (
	openAICodexTransportWarmupTotal                  atomic.Int64
	openAICodexTransportFallbackCoolingHitTotal      atomic.Int64
	openAICodexSessionPreferredTransportHitTotal     atomic.Int64
	openAICodexSessionPreferredTransportWSHitTotal   atomic.Int64
	openAICodexSessionPreferredTransportHTTPHitTotal atomic.Int64
	openAICodexSessionHTTPFallbackHitTotal           atomic.Int64
	openAICodexSessionTransportBindWSTotal           atomic.Int64
	openAICodexSessionTransportBindHTTPTotal         atomic.Int64
	openAICodexSessionTransportWarmupIgnoredTotal    atomic.Int64
	openAICodexSessionTransportHTTPDowngradeTotal    atomic.Int64
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

func SnapshotOpenAICodexCompatibilityMetrics() OpenAICodexCompatibilityMetricsSnapshot {
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
	}
}
