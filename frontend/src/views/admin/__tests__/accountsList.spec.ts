import { describe, expect, it, vi } from 'vitest'
import type { Account } from '@/types'
import {
  accountMatchesCurrentFilters,
  buildAccountTodayStatsMap,
  buildDefaultTodayStats,
  mergeIncrementalAccountRows,
  mergeAccountRuntimeFields,
  normalizeBulkSchedulableResult,
  patchAccountList,
  shouldReplaceAutoRefreshAccountRow
} from '../accounts/accountsList'

function createAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'Account',
    platform: 'openai',
    type: 'oauth',
    credentials: {},
    extra: {},
    proxy_id: null,
    concurrency: 1,
    current_concurrency: 0,
    priority: 0,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    current_window_cost: 0,
    active_sessions: 0,
    ...overrides
  }
}

describe('accountsList helpers', () => {
  it('builds today stats map and fills missing accounts with zero stats', () => {
    expect(
      buildAccountTodayStatsMap([1, 2], {
        2: {
          requests: 3,
          tokens: 4,
          cost: 5,
          standard_cost: 6,
          user_cost: 7
        }
      })
    ).toEqual({
      1: buildDefaultTodayStats(),
      2: {
        requests: 3,
        tokens: 4,
        cost: 5,
        standard_cost: 6,
        user_cost: 7
      }
    })
  })

  it('detects when auto refresh rows need replacement', () => {
    const current = createAccount()
    const same = createAccount()
    const changedUsage = createAccount({
      extra: {
        codex_usage_updated_at: '2026-01-02T00:00:00Z'
      }
    })

    expect(shouldReplaceAutoRefreshAccountRow(current, same)).toBe(false)
    expect(
      shouldReplaceAutoRefreshAccountRow(
        current,
        createAccount({ updated_at: '2026-01-02T00:00:00Z' })
      )
    ).toBe(true)
    expect(shouldReplaceAutoRefreshAccountRow(current, changedUsage)).toBe(true)
  })

  it('merges incremental account rows and preserves unchanged row references', () => {
    const currentA = createAccount({ id: 1, name: 'A' })
    const currentB = createAccount({ id: 2, name: 'B' })
    const replaced = createAccount({ id: 2, name: 'B2', updated_at: '2026-01-02T00:00:00Z' })
    const onReplaced = vi.fn()
    const currentRows = [currentA, currentB]

    const result = mergeIncrementalAccountRows(currentRows, [currentA, replaced], onReplaced)

    expect(result.changed).toBe(true)
    expect(result.rows[0]).toBe(currentA)
    expect(result.rows[1]).toBe(replaced)
    expect(onReplaced).toHaveBeenCalledWith(replaced)

    const unchangedRows = [currentA]
    const unchanged = mergeIncrementalAccountRows(
      unchangedRows,
      [createAccount({ id: 1, name: 'A' })]
    )
    expect(unchanged.changed).toBe(false)
    expect(unchanged.rows).toBe(unchangedRows)
    expect(unchanged.rows[0]).toBe(currentA)
  })

  it('normalizes bulk schedulable responses across shapes', () => {
    expect(
      normalizeBulkSchedulableResult(
        {
          success_ids: [1, 2],
          failed_ids: [3]
        },
        [1, 2, 3]
      )
    ).toEqual({
      successIds: [1, 2],
      failedIds: [3],
      successCount: 2,
      failedCount: 1,
      hasIds: true,
      hasCounts: true
    })

    expect(
      normalizeBulkSchedulableResult(
        {
          success: 2,
          failed: 0
        },
        [7, 8]
      )
    ).toEqual({
      successIds: [7, 8],
      failedIds: [],
      successCount: 2,
      failedCount: 0,
      hasIds: true,
      hasCounts: true
    })

    expect(
      normalizeBulkSchedulableResult(
        {
          success: 1,
          failed: 1
        },
        [4, 5]
      )
    ).toEqual({
      successIds: [],
      failedIds: [],
      successCount: 1,
      failedCount: 1,
      hasIds: false,
      hasCounts: true
    })
  })

  it('matches account filters and preserves runtime fields on patch', () => {
    const rateLimited = createAccount({
      id: 9,
      name: 'Primary OpenAI',
      status: 'error',
      rate_limit_reset_at: '2026-01-03T00:00:00Z',
      extra: { privacy_mode: 'training_off' },
      group_ids: [12],
      current_concurrency: 4,
      current_window_cost: 8,
      active_sessions: 2
    })

    expect(
      accountMatchesCurrentFilters(
        rateLimited,
        {
          platform: 'openai',
          type: 'oauth',
          status: 'rate_limited',
          privacy_mode: 'training_off',
          group: '12',
          search: 'openai'
        },
        new Date('2026-01-02T00:00:00Z').getTime()
      )
    ).toBe(true)

    expect(
      accountMatchesCurrentFilters(
        rateLimited,
        {
          status: 'active'
        },
        new Date('2026-01-02T00:00:00Z').getTime()
      )
    ).toBe(false)

    expect(
      accountMatchesCurrentFilters(rateLimited, {
        search: '9'
      })
    ).toBe(true)

    expect(
      accountMatchesCurrentFilters(rateLimited, {
        privacy_mode: '__unset__'
      })
    ).toBe(false)

    expect(
      accountMatchesCurrentFilters(rateLimited, {
        group: 'ungrouped'
      })
    ).toBe(false)

    expect(
      mergeAccountRuntimeFields(
        rateLimited,
        createAccount({
          id: 9,
          current_concurrency: undefined,
          current_window_cost: undefined,
          active_sessions: undefined
        })
      )
    ).toEqual(
      expect.objectContaining({
        current_concurrency: 4,
        current_window_cost: 8,
        active_sessions: 2
      })
    )
  })

  it('patches account rows locally and removes rows that no longer match filters', () => {
    const current = createAccount({
      id: 3,
      name: 'OpenAI Main',
      current_concurrency: 1,
      current_window_cost: 2,
      active_sessions: 3
    })

    const updated = createAccount({
      id: 3,
      name: 'OpenAI Main Updated',
      current_concurrency: undefined,
      current_window_cost: undefined,
      active_sessions: undefined
    })

    const patched = patchAccountList(
      [current],
      updated,
      { platform: 'openai', search: 'main' },
      { page: 1, page_size: 20, total: 1, pages: 1 },
      false,
      null
    )

    expect(patched.removedAccountId).toBeNull()
    expect(patched.patchedAccount).toEqual(
      expect.objectContaining({
        id: 3,
        name: 'OpenAI Main Updated',
        current_concurrency: 1,
        current_window_cost: 2,
        active_sessions: 3
      })
    )

    const removed = patchAccountList(
      [current],
      createAccount({ id: 3, platform: 'anthropic', name: 'Claude' }),
      { platform: 'openai' },
      { page: 2, page_size: 20, total: 4, pages: 2 },
      false,
      3
    )

    expect(removed.accounts).toEqual([])
    expect(removed.pagination).toEqual({
      page: 1,
      total: 3,
      pages: 1
    })
    expect(removed.hasPendingListSync).toBe(true)
    expect(removed.removedAccountId).toBe(3)
    expect(removed.shouldCloseMenu).toBe(true)
  })
})
