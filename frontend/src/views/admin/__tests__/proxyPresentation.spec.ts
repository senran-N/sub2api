import { describe, expect, it } from 'vitest'
import type { Proxy, QualityCategoryScores } from '@/types'
import {
  buildProxyFlagUrl,
  buildProxyQualityCategoryEntries,
  formatProxyLocation,
  getDnsLeakBadgeClass,
  getDnsLeakLabel,
  getIpTypeBadgeClass,
  getIpTypeLabel,
  getProxyScoreBarColor,
  getQualityOverallClass,
  getQualityOverallLabel,
  getQualityStatusClass,
  getQualityStatusLabel,
  getQualityTargetLabel
} from '../proxies/proxyPresentation'

const t = (key: string) => key

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

describe('proxyPresentation', () => {
  it('formats proxy location and flag urls', () => {
    expect(formatProxyLocation(createProxy({ country: 'US', city: 'Seattle' }))).toBe('US · Seattle')
    expect(formatProxyLocation(createProxy({ country: 'US', city: undefined }))).toBe('US')
    expect(buildProxyFlagUrl('US')).toBe('https://unpkg.com/flag-icons/flags/4x3/us.svg')
  })

  it('maps quality statuses and overall statuses to classes and labels', () => {
    expect(getQualityStatusClass('pass')).toBe('badge-success')
    expect(getQualityStatusClass('fail')).toBe('badge-danger')
    expect(getQualityStatusLabel('warn', t)).toBe('admin.proxies.qualityStatusWarn')
    expect(getQualityOverallClass('healthy')).toBe('badge-success')
    expect(getQualityOverallClass(undefined)).toBe('badge-danger')
    expect(getQualityOverallLabel('challenge', t)).toBe(
      'admin.proxies.qualityStatusChallenge'
    )
  })

  it('maps quality targets, ip types, dns leak states, and score colors', () => {
    expect(getQualityTargetLabel('base_connectivity', t)).toBe(
      'admin.proxies.qualityTargetBase'
    )
    expect(getQualityTargetLabel('openai', t)).toBe('OpenAI')
    expect(getIpTypeBadgeClass('mobile')).toBe('badge-info')
    expect(getIpTypeLabel('vpn', t)).toBe('admin.proxies.qualityIPTypeVPN')
    expect(getDnsLeakBadgeClass('possible')).toBe('badge-warning')
    expect(getDnsLeakLabel('detected', t)).toBe('admin.proxies.qualityDNSLeakDetected')
    expect(getProxyScoreBarColor(85)).toBe('theme-progress-fill--success')
    expect(getProxyScoreBarColor(45)).toBe('theme-progress-fill--brand-orange')
  })

  it('builds localized category score entries in display order', () => {
    const scores: QualityCategoryScores = {
      reachability: 90,
      ip_risk: 70,
      ip_type: 80,
      abuse_history: 60,
      latency: 50
    }

    expect(buildProxyQualityCategoryEntries(scores, t)).toEqual([
      {
        key: 'reachability',
        label: 'admin.proxies.qualityCategoryReachability',
        weight: 30,
        score: 90
      },
      {
        key: 'ip_risk',
        label: 'admin.proxies.qualityCategoryIPRisk',
        weight: 25,
        score: 70
      },
      {
        key: 'ip_type',
        label: 'admin.proxies.qualityCategoryIPType',
        weight: 20,
        score: 80
      },
      {
        key: 'abuse_history',
        label: 'admin.proxies.qualityCategoryAbuse',
        weight: 15,
        score: 60
      },
      {
        key: 'latency',
        label: 'admin.proxies.qualityCategoryLatency',
        weight: 10,
        score: 50
      }
    ])

    expect(buildProxyQualityCategoryEntries(undefined, t)).toEqual([])
  })
})
