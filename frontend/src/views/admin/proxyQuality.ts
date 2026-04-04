import type { Proxy, ProxyQualityCheckResult } from '@/types'

export interface ProxyLatencyResult {
  success: boolean
  latency_ms?: number
  message?: string
  ip_address?: string
  country?: string
  country_code?: string
  region?: string
  city?: string
}

export interface ProxyBatchQualitySummary {
  total: number
  healthy: number
  warn: number
  challenge: number
  failed: number
}

export function applyProxyLatencyResult(proxy: Proxy, result: ProxyLatencyResult) {
  if (result.success) {
    proxy.latency_status = 'success'
    proxy.latency_ms = result.latency_ms
    proxy.ip_address = result.ip_address
    proxy.country = result.country
    proxy.country_code = result.country_code
    proxy.region = result.region
    proxy.city = result.city
  } else {
    proxy.latency_status = 'failed'
    proxy.latency_ms = undefined
    proxy.ip_address = undefined
    proxy.country = undefined
    proxy.country_code = undefined
    proxy.region = undefined
    proxy.city = undefined
  }

  proxy.latency_message = result.message
}

export function summarizeProxyQualityStatus(
  result: ProxyQualityCheckResult
): Proxy['quality_status'] {
  if (result.challenge_count > 0) {
    return 'challenge'
  }
  if (result.failed_count > 0) {
    return 'failed'
  }
  if (result.warn_count > 0) {
    return 'warn'
  }
  return 'healthy'
}

export function applyProxyQualityResult(proxy: Proxy, result: ProxyQualityCheckResult) {
  proxy.quality_status = summarizeProxyQualityStatus(result)
  proxy.quality_score = result.score
  proxy.quality_grade = result.grade
  proxy.quality_summary = result.summary
  proxy.quality_checked = result.checked_at
}

export function applyProxyConnectivityFromQualityResult(
  proxy: Proxy,
  result: ProxyQualityCheckResult
) {
  const baseConnectivity = result.items.find((item) => item.target === 'base_connectivity')
  if (baseConnectivity?.status !== 'pass') {
    return false
  }

  applyProxyLatencyResult(proxy, {
    success: true,
    latency_ms: result.base_latency_ms,
    message: result.summary,
    ip_address: result.exit_ip,
    country: result.country,
    country_code: result.country_code
  })
  return true
}

export function createProxyBatchQualitySummary(total: number): ProxyBatchQualitySummary {
  return {
    total,
    healthy: 0,
    warn: 0,
    challenge: 0,
    failed: 0
  }
}

export function recordProxyBatchQualityResult(
  summary: ProxyBatchQualitySummary,
  result: ProxyQualityCheckResult
) {
  const status = summarizeProxyQualityStatus(result)

  if (status === 'challenge') {
    summary.challenge += 1
    return
  }
  if (status === 'failed') {
    summary.failed += 1
    return
  }
  if (status === 'warn') {
    summary.warn += 1
    return
  }

  summary.healthy += 1
}
