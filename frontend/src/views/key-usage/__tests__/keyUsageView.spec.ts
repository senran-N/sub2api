import { describe, expect, it } from 'vitest'
import {
  buildKeyUsageDateParams,
  buildKeyUsageDetailRows,
  buildKeyUsageRequestUrl,
  buildKeyUsageRingGridClass,
  buildKeyUsageRingItems,
  buildKeyUsageStatusInfo,
  buildKeyUsageUsageStatCells,
  formatKeyUsageDate,
  formatKeyUsageNumber,
  formatKeyUsageResetTime,
  formatKeyUsageUsd,
  resolveKeyUsageQueryErrorMessage
} from '../keyUsageView'

const t = (key: string, values?: Record<string, unknown>) => {
  if (key === 'keyUsage.daysLeft') {
    return `${String(values?.days)}d left`
  }

  return key
}

describe('keyUsageView helpers', () => {
  it('builds date params and formats values', () => {
    expect(
      buildKeyUsageDateParams({
        range: '7d',
        customStartDate: '',
        customEndDate: '',
        now: new Date('2026-04-05T08:00:00Z')
      })
    ).toBe('start_date=2026-03-29&end_date=2026-04-05')
    expect(
      buildKeyUsageDateParams({
        range: 'custom',
        customStartDate: '2026-04-01',
        customEndDate: '2026-04-05'
      })
    ).toBe('start_date=2026-04-01&end_date=2026-04-05')
    expect(buildKeyUsageRequestUrl('start_date=2026-04-01&end_date=2026-04-05')).toBe(
      '/v1/usage?start_date=2026-04-01&end_date=2026-04-05'
    )
    expect(buildKeyUsageRequestUrl('')).toBe('/v1/usage')
    expect(formatKeyUsageUsd(12.345)).toBe('$12.35')
    expect(formatKeyUsageUsd(-1)).toBe('-')
    expect(formatKeyUsageNumber(12345)).toBe('12,345')
    expect(formatKeyUsageDate('2026-04-05T00:00:00Z', 'en')).toContain('2026')
    expect(
      resolveKeyUsageQueryErrorMessage(
        { error: { message: 'bad request' } },
        400,
        'fallback'
      )
    ).toBe('bad request')
    expect(resolveKeyUsageQueryErrorMessage(null, 500, 'fallback')).toBe('fallback (500)')
  })

  it('builds status, ring items and grid classes', () => {
    const quotaResult = {
      mode: 'quota_limited',
      status: 'active',
      isValid: true,
      quota: { used: 10, limit: 20, remaining: 10 },
      rate_limits: [{ window: '1d', used: 5, limit: 10, reset_at: '2026-04-06T00:00:00Z' }]
    }

    expect(buildKeyUsageStatusInfo(quotaResult, t)).toEqual({
      label: 'keyUsage.quotaMode',
      statusText: 'Active',
      isActive: true
    })
    expect(
      buildKeyUsageRingItems(quotaResult, {
        t,
        usd: formatKeyUsageUsd
      })
    ).toHaveLength(2)
    expect(buildKeyUsageRingGridClass(1)).toContain('max-w-md')
    expect(buildKeyUsageRingGridClass(3)).toContain('lg:grid-cols-3')
  })

  it('builds detail rows, usage stats and reset time labels', () => {
    const result = {
      mode: 'quota_limited',
      quota: { used: 10, limit: 20, remaining: 2 },
      expires_at: '2026-04-10T00:00:00Z',
      days_until_expiry: 5,
      rate_limits: [{ window: '5h', used: 1, limit: 2, reset_at: '2026-04-05T10:30:00Z' }],
      usage: {
        today: { requests: 2, total_tokens: 123, actual_cost: 1.2 },
        total: { requests: 3, total_tokens: 456, actual_cost: 2.3 },
        rpm: 10,
        tpm: 20,
        average_duration_ms: 1234
      }
    }

    const detailRows = buildKeyUsageDetailRows(result, {
      t,
      locale: 'en',
      usd: formatKeyUsageUsd,
      formatDate: (value) => formatKeyUsageDate(value, 'en'),
      formatResetTime: (value) =>
        formatKeyUsageResetTime(value, new Date('2026-04-05T10:00:00Z'), 'now')
    })

    expect(detailRows[0]?.value).toBe('$2.00')
    expect(detailRows[1]?.value).toContain('5d left')
    expect(detailRows[2]?.value).toContain('30m')

    const statCells = buildKeyUsageUsageStatCells(result, {
      t,
      fmtNum: formatKeyUsageNumber,
      usd: formatKeyUsageUsd
    })
    expect(statCells).toHaveLength(16)
    expect(statCells[0]).toEqual({
      label: 'keyUsage.todayRequests',
      value: '2'
    })
    expect(
      formatKeyUsageResetTime(
        '2026-04-05T10:30:00Z',
        new Date('2026-04-05T10:00:00Z'),
        'now'
      )
    ).toBe('30m')
  })
})
