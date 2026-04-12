import { describe, expect, it } from 'vitest'
import {
  buildCcsImportDeeplink,
  formatApiKeyResetTime,
  getApiKeyQuotaBarTone,
  getApiKeyQuotaProgressWidth,
  getApiKeyQuotaTextTone,
  getApiKeyRateLimitBarTone,
  getApiKeyRateLimitProgressWidth,
  getApiKeyRateLimitTextTone,
  getApiKeyRateLimitWindows,
  getApiKeyExpirationTextClass,
  getApiKeyStatusBadgeClass,
  getApiKeyUsageSummary,
  hasApiKeyRateLimitUsage,
  hasApiKeyIpRestrictions,
  maskUserApiKey
} from '../keys/keysView'

describe('keysView helpers', () => {
  it('masks keys and summarizes usage', () => {
    expect(maskUserApiKey('sk-1234567890abcdef')).toBe('sk-12345...cdef')
    expect(maskUserApiKey('short-key')).toBe('short-key')
    expect(
      getApiKeyUsageSummary({
        today_actual_cost: 1.2345,
        total_actual_cost: 6.789
      } as any)
    ).toEqual({
      todayCost: '1.2345',
      totalCost: '6.7890'
    })
  })

  it('builds quota display tones and widths', () => {
    const row = { quota: 100, quota_used: 85 } as any
    expect(getApiKeyQuotaProgressWidth(row)).toBe('85%')
    expect(getApiKeyQuotaTextTone(row)).toBe('theme-text-warning')
    expect(getApiKeyQuotaBarTone(row)).toBe('theme-progress-fill--warning')
  })

  it('builds rate limit windows and tones', () => {
    const row = {
      rate_limit_5h: 10,
      rate_limit_1d: 0,
      rate_limit_7d: 30,
      usage_5h: 8,
      usage_1d: 0,
      usage_7d: 31,
      reset_5h_at: '2026-04-05T10:00:00Z',
      reset_1d_at: null,
      reset_7d_at: '2026-04-12T10:00:00Z'
    } as any

    expect(getApiKeyRateLimitWindows(row)).toEqual([
      {
        key: '5h',
        label: '5h',
        usage: 8,
        limit: 10,
        resetAt: '2026-04-05T10:00:00Z'
      },
      {
        key: '7d',
        label: '7d',
        usage: 31,
        limit: 30,
        resetAt: '2026-04-12T10:00:00Z'
      }
    ])
    expect(getApiKeyRateLimitTextTone(8, 10)).toBe('theme-text-warning')
    expect(getApiKeyRateLimitBarTone(31, 30)).toBe('theme-progress-fill--danger')
    expect(getApiKeyRateLimitProgressWidth(15, 10)).toBe('100%')
    expect(hasApiKeyRateLimitUsage(row)).toBe(true)
  })

  it('returns status badge classes', () => {
    expect(getApiKeyStatusBadgeClass('active')).toBe('badge-success')
    expect(getApiKeyStatusBadgeClass('quota_exhausted')).toBe('badge-warning')
    expect(getApiKeyStatusBadgeClass('expired')).toBe('badge-danger')
    expect(getApiKeyStatusBadgeClass('inactive')).toBe('badge-gray')
  })

  it('detects ip restrictions and expiration presentation', () => {
    expect(hasApiKeyIpRestrictions({ ip_whitelist: ['1.1.1.1'], ip_blacklist: [] } as any)).toBe(
      true
    )
    expect(getApiKeyExpirationTextClass(undefined)).toContain('theme-text-subtle')
    expect(
      getApiKeyExpirationTextClass('2026-04-04T00:00:00Z', new Date('2026-04-05T00:00:00Z'))
    ).toContain('theme-text-danger')
  })

  it('formats reset timers and builds ccs import deeplinks', () => {
    expect(
      formatApiKeyResetTime(
        '2026-04-05T12:30:00Z',
        new Date('2026-04-05T10:00:00Z'),
        (key) => key
      )
    ).toBe('2h 30m')

    const deeplink = buildCcsImportDeeplink(
      {
        key: 'sk-test',
        group: {
          platform: 'antigravity'
        }
      } as any,
      {
        api_base_url: 'https://api.example.com',
        site_name: 'Sub2API'
      },
      'gemini',
      'https://fallback.example.com'
    )

    expect(deeplink).toContain('ccswitch://v1/import?')
    expect(deeplink).toContain('app=gemini')
    expect(deeplink).toContain('endpoint=https%3A%2F%2Fapi.example.com%2Fantigravity')
    expect(deeplink).toContain('name=Sub2API')

    const usageScript = new URLSearchParams(deeplink.split('?')[1]).get('usageScript')
    expect(usageScript).not.toBeNull()
    expect(atob(usageScript!)).toContain('url: "https://api.example.com/v1/usage"')
  })

  it('normalizes usage endpoint when api_base_url already contains /api/v1', () => {
    const deeplink = buildCcsImportDeeplink(
      {
        key: 'sk-test',
        group: {
          platform: 'openai'
        }
      } as any,
      {
        api_base_url: 'https://api.example.com/api/v1',
        site_name: 'Sub2API'
      },
      'codex',
      'https://fallback.example.com'
    )

    const usageScript = new URLSearchParams(deeplink.split('?')[1]).get('usageScript')
    expect(usageScript).not.toBeNull()
    expect(atob(usageScript!)).toContain('url: "https://api.example.com/api/v1/usage"')
    expect(atob(usageScript!)).not.toContain('/api/v1/v1/usage')
  })
})
