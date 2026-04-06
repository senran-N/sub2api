import { describe, expect, it } from 'vitest'
import {
  buildAccountAutoRefreshIntervalLabel,
  buildAccountExportFilename,
  buildAccountExportRequest,
  buildAccountTableColumns,
  formatAccountExportTimestamp,
  formatAccountExpiresAt,
  getAccountAntigravityTier,
  getAccountAntigravityTierClass,
  getAccountAntigravityTierLabel,
  isAccountExpired
} from '../accountsView'

describe('accountsView helpers', () => {
  it('builds localized auto refresh labels and account table columns', () => {
    expect(buildAccountAutoRefreshIntervalLabel(5, (key: string) => key)).toBe(
      'admin.accounts.refreshInterval5s'
    )
    expect(buildAccountAutoRefreshIntervalLabel(42, (key: string) => key)).toBe('42s')

    const fullColumns = buildAccountTableColumns(false, (key: string) => key)
    const simpleColumns = buildAccountTableColumns(true, (key: string) => key)
    expect(fullColumns.some((column) => column.key === 'groups')).toBe(true)
    expect(simpleColumns.some((column) => column.key === 'groups')).toBe(false)
    expect(fullColumns.at(-1)?.key).toBe('actions')
  })

  it('formats export filenames and request payloads', () => {
    const date = new Date(2026, 3, 4, 12, 34, 56)
    expect(formatAccountExportTimestamp(date)).toBe('20260404123456')
    expect(buildAccountExportFilename(date)).toBe('sub2api-account-20260404123456.json')

    expect(
      buildAccountExportRequest([1, 2], true, {
        platform: 'openai',
        type: 'oauth',
        status: 'active',
        search: 'main'
      })
    ).toEqual({
      ids: [1, 2],
      includeProxies: true
    })

    expect(
      buildAccountExportRequest([], false, {
        platform: 'openai',
        type: 'oauth',
        status: 'active',
        search: 'main'
      })
    ).toEqual({
      includeProxies: false,
      filters: {
        platform: 'openai',
        type: 'oauth',
        status: 'active',
        search: 'main'
      }
    })
  })

  it('derives antigravity tier labels and expiration state', () => {
    const account = {
      platform: 'antigravity',
      extra: {
        load_code_assist: {
          currentTier: { id: 'free-tier' },
          paidTier: { id: 'g1-pro-tier' }
        }
      }
    } as const

    expect(getAccountAntigravityTier(account as any)).toBe('g1-pro-tier')
    expect(getAccountAntigravityTierLabel(account as any, (key: string) => key)).toBe(
      'admin.accounts.tier.pro'
    )
    expect(getAccountAntigravityTierClass(account as any)).toBe('theme-chip--info')

    expect(getAccountAntigravityTier({ platform: 'openai', extra: {} } as any)).toBeNull()
    expect(formatAccountExpiresAt(null)).toBe('-')
    expect(formatAccountExpiresAt(0)).toBe('-')
    expect(formatAccountExpiresAt(1767225600)).toContain('2026-01-01')

    expect(isAccountExpired(null, 1000)).toBe(false)
    expect(isAccountExpired(1, 999)).toBe(false)
    expect(isAccountExpired(1, 1000)).toBe(true)
  })
})
