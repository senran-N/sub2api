import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, reactive, ref } from 'vue'
import type { Account } from '@/types'
import { useAccountsViewRefresh } from '../useAccountsViewRefresh'

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

async function flushMicrotasks() {
  await Promise.resolve()
  await Promise.resolve()
}

describe('useAccountsViewRefresh', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    localStorage.clear()
    Object.defineProperty(document, 'hidden', {
      configurable: true,
      value: false
    })
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  function createComposable() {
    const accounts = ref<Account[]>([])
    const loading = ref(false)
    const params = reactive<Record<string, unknown>>({
      platform: '',
      type: '',
      status: '',
      privacy_mode: '',
      group: '',
      search: ''
    })
    const pagination = reactive({
      page: 1,
      page_size: 20,
      total: 0,
      pages: 0
    })
    const hiddenColumns = reactive(new Set<string>())
    const loadBase = vi.fn(async () => {
      accounts.value = [createAccount({ id: 1 })]
    })
    const reloadBase = vi.fn(async () => {
      accounts.value = [createAccount({ id: 2 })]
    })
    const debouncedReloadBase = vi.fn()
    const handlePageChangeBase = vi.fn()
    const handlePageSizeChangeBase = vi.fn()
    const fetchTodayStats = vi.fn(async (accountIds: number[]) => ({
      stats: {
        [String(accountIds[0])]: {
          requests: 1,
          tokens: 2,
          cost: 3,
          standard_cost: 4,
          user_cost: 5
        }
      }
    }))
    const fetchAccountsIncrementally = vi.fn(async () => ({
      etag: 'etag-1',
      notModified: false,
      data: {
        total: 1,
        pages: 1,
        items: [createAccount({ id: 1, updated_at: '2026-01-02T00:00:00Z' })]
      }
    }))
    const syncAccountRefs = vi.fn()

    const composable = useAccountsViewRefresh({
      accounts,
      loading,
      params,
      pagination,
      hiddenColumns,
      isAnyModalOpen: computed(() => false),
      isActionMenuOpen: computed(() => false),
      loadBase,
      reloadBase,
      debouncedReloadBase,
      handlePageChangeBase,
      handlePageSizeChangeBase,
      fetchTodayStats,
      fetchAccountsIncrementally,
      syncAccountRefs
    })

    return {
      accounts,
      params,
      pagination,
      hiddenColumns,
      loadBase,
      reloadBase,
      debouncedReloadBase,
      handlePageChangeBase,
      handlePageSizeChangeBase,
      fetchTodayStats,
      fetchAccountsIncrementally,
      syncAccountRefs,
      composable
    }
  }

  it('loads first page with lite mode and refreshes today stats', async () => {
    const setup = createComposable()

    await setup.composable.load()

    expect(setup.loadBase).toHaveBeenCalledTimes(1)
    expect(setup.params.lite).toBeUndefined()
    expect(setup.fetchTodayStats).toHaveBeenCalledWith([1])
    expect(setup.composable.todayStatsByAccountId.value['1']).toEqual({
      requests: 1,
      tokens: 2,
      cost: 3,
      standard_cost: 4,
      user_cost: 5
    })
  })

  it('skips today stats requests when both dependent columns are hidden', async () => {
    const setup = createComposable()
    setup.accounts.value = [createAccount({ id: 9 })]
    setup.hiddenColumns.add('today_stats')
    setup.hiddenColumns.add('usage')

    await setup.composable.refreshTodayStatsBatch()

    expect(setup.fetchTodayStats).not.toHaveBeenCalled()
    expect(setup.composable.todayStatsLoading.value).toBe(false)
    expect(setup.composable.todayStatsError.value).toBeNull()
  })

  it('auto refreshes incrementally and syncs updated account refs', async () => {
    const setup = createComposable()
    setup.accounts.value = [createAccount({ id: 1 })]

    setup.composable.setAutoRefreshInterval(5)
    setup.composable.setAutoRefreshEnabled(true)

    await vi.advanceTimersByTimeAsync(6000)
    await flushMicrotasks()

    expect(setup.fetchAccountsIncrementally).toHaveBeenCalledTimes(1)
    expect(setup.syncAccountRefs).toHaveBeenCalledWith(
      expect.objectContaining({
        id: 1,
        updated_at: '2026-01-02T00:00:00Z'
      })
    )
    expect(setup.pagination.total).toBe(1)
    expect(setup.composable.todayStatsByAccountId.value['1']).toEqual({
      requests: 1,
      tokens: 2,
      cost: 3,
      standard_cost: 4,
      user_cost: 5
    })
    setup.composable.dispose()
  })

  it('manual refresh and pending list sync both bump usage refresh token', async () => {
    const setup = createComposable()

    await setup.composable.handleManualRefresh()
    await setup.composable.syncPendingListChanges()

    expect(setup.composable.usageManualRefreshToken.value).toBe(2)
    expect(setup.loadBase).toHaveBeenCalledTimes(2)
  })
})
