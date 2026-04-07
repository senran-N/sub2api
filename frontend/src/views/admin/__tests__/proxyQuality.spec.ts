import { describe, expect, it } from 'vitest'
import type { Proxy, ProxyQualityCheckResult } from '@/types'
import {
  applyProxyConnectivityFromQualityResult,
  applyProxyLatencyResult,
  applyProxyQualityResult,
  createProxyBatchQualitySummary,
  recordProxyBatchQualityResult,
  summarizeProxyQualityStatus
} from '../proxies/proxyQuality'

function createProxy(overrides: Partial<Proxy> = {}): Proxy {
  return {
    id: 1,
    name: 'Proxy',
    protocol: 'http',
    host: 'proxy.local',
    port: 8080,
    username: null,
    password: null,
    status: 'active',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

function createQualityResult(
  overrides: Partial<ProxyQualityCheckResult> = {}
): ProxyQualityCheckResult {
  return {
    proxy_id: 1,
    score: 92,
    grade: 'A',
    summary: 'Healthy',
    exit_ip: '1.1.1.1',
    country: 'United States',
    country_code: 'US',
    base_latency_ms: 180,
    passed_count: 2,
    warn_count: 0,
    failed_count: 0,
    challenge_count: 0,
    checked_at: 1234567890,
    items: [{ target: 'base_connectivity', status: 'pass' }],
    ...overrides
  }
}

describe('proxyQuality', () => {
  it('applies latency results and clears stale location data on failure', () => {
    const proxy = createProxy({
      latency_status: 'success',
      latency_ms: 120,
      ip_address: '1.1.1.1',
      country: 'United States',
      country_code: 'US',
      region: 'CA',
      city: 'San Francisco'
    })

    applyProxyLatencyResult(proxy, {
      success: false,
      message: 'timeout'
    })

    expect(proxy.latency_status).toBe('failed')
    expect(proxy.latency_ms).toBeUndefined()
    expect(proxy.ip_address).toBeUndefined()
    expect(proxy.country).toBeUndefined()
    expect(proxy.country_code).toBeUndefined()
    expect(proxy.region).toBeUndefined()
    expect(proxy.city).toBeUndefined()
    expect(proxy.latency_message).toBe('timeout')
  })

  it('summarizes and applies quality results', () => {
    const proxy = createProxy()
    const result = createQualityResult({
      score: 68,
      grade: 'C',
      summary: 'Warnings detected',
      warn_count: 1
    })

    expect(summarizeProxyQualityStatus(result)).toBe('warn')

    applyProxyQualityResult(proxy, result)

    expect(proxy.quality_status).toBe('warn')
    expect(proxy.quality_score).toBe(68)
    expect(proxy.quality_grade).toBe('C')
    expect(proxy.quality_summary).toBe('Warnings detected')
    expect(proxy.quality_checked).toBe(1234567890)
  })

  it('updates connectivity from quality results only when base connectivity passes', () => {
    const proxy = createProxy()
    const successResult = createQualityResult()
    const failedResult = createQualityResult({
      items: [{ target: 'base_connectivity', status: 'fail' }],
      base_latency_ms: 999
    })

    expect(applyProxyConnectivityFromQualityResult(proxy, successResult)).toBe(true)
    expect(proxy.latency_status).toBe('success')
    expect(proxy.latency_ms).toBe(180)
    expect(proxy.ip_address).toBe('1.1.1.1')

    proxy.latency_ms = 180
    expect(applyProxyConnectivityFromQualityResult(proxy, failedResult)).toBe(false)
    expect(proxy.latency_ms).toBe(180)
  })

  it('tracks batch quality counts by derived status', () => {
    const summary = createProxyBatchQualitySummary(4)

    recordProxyBatchQualityResult(summary, createQualityResult())
    recordProxyBatchQualityResult(summary, createQualityResult({ warn_count: 1 }))
    recordProxyBatchQualityResult(summary, createQualityResult({ failed_count: 1 }))
    recordProxyBatchQualityResult(summary, createQualityResult({ challenge_count: 1 }))

    expect(summary).toEqual({
      total: 4,
      healthy: 1,
      warn: 1,
      challenge: 1,
      failed: 1
    })
  })
})
