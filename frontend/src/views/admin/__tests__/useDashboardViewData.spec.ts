import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import type {
  DashboardStats,
  ModelStat,
  TrendDataPoint,
  UserSpendingRankingItem,
  UserUsageTrendPoint
} from '@/types'
import { useDashboardViewData } from '../dashboard/useDashboardViewData'

const { getSnapshotV2, getUserUsageTrend, getUserSpendingRanking } = vi.hoisted(() => ({
  getSnapshotV2: vi.fn(),
  getUserUsageTrend: vi.fn(),
  getUserSpendingRanking: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    dashboard: {
      getSnapshotV2,
      getUserUsageTrend,
      getUserSpendingRanking
    }
  }
}))

function createDashboardStats(overrides: Partial<DashboardStats> = {}): DashboardStats {
  return {
    total_users: 100,
    today_new_users: 2,
    active_users: 20,
    hourly_active_users: 8,
    stats_updated_at: '2026-04-04T00:00:00Z',
    stats_stale: false,
    total_api_keys: 50,
    active_api_keys: 48,
    total_accounts: 12,
    normal_accounts: 10,
    error_accounts: 2,
    ratelimit_accounts: 0,
    overload_accounts: 0,
    total_requests: 3000,
    total_input_tokens: 100000,
    total_output_tokens: 50000,
    total_cache_creation_tokens: 1000,
    total_cache_read_tokens: 2000,
    total_tokens: 153000,
    total_cost: 12.3,
    total_actual_cost: 11.1,
    today_requests: 120,
    today_input_tokens: 5000,
    today_output_tokens: 2200,
    today_cache_creation_tokens: 10,
    today_cache_read_tokens: 20,
    today_tokens: 7230,
    today_cost: 0.88,
    today_actual_cost: 0.8,
    average_duration_ms: 230,
    uptime: 3600,
    rpm: 80,
    tpm: 9000,
    ...overrides
  }
}

function createTrendPoint(overrides: Partial<TrendDataPoint> = {}): TrendDataPoint {
  return {
    date: '2026-04-04',
    requests: 100,
    input_tokens: 1000,
    output_tokens: 800,
    cache_creation_tokens: 20,
    cache_read_tokens: 50,
    total_tokens: 1870,
    cost: 1.2,
    actual_cost: 1.1,
    ...overrides
  }
}

function createModelStat(overrides: Partial<ModelStat> = {}): ModelStat {
  return {
    model: 'gpt-test',
    requests: 100,
    input_tokens: 1000,
    output_tokens: 800,
    cache_creation_tokens: 20,
    cache_read_tokens: 50,
    total_tokens: 1870,
    cost: 1.2,
    actual_cost: 1.1,
    ...overrides
  }
}

function createUserTrendPoint(overrides: Partial<UserUsageTrendPoint> = {}): UserUsageTrendPoint {
  return {
    date: '2026-04-04',
    user_id: 1,
    email: 'user@example.com',
    username: 'user',
    requests: 10,
    tokens: 500,
    cost: 0.3,
    actual_cost: 0.28,
    ...overrides
  }
}

function createRankingItem(
  overrides: Partial<UserSpendingRankingItem> = {}
): UserSpendingRankingItem {
  return {
    user_id: 1,
    email: 'user@example.com',
    actual_cost: 4.2,
    requests: 88,
    tokens: 9999,
    ...overrides
  }
}

function createDeferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return {
    promise,
    resolve,
    reject
  }
}

describe('useDashboardViewData', () => {
  beforeEach(() => {
    getSnapshotV2.mockReset()
    getUserUsageTrend.mockReset()
    getUserSpendingRanking.mockReset()

    getSnapshotV2.mockResolvedValue({
      stats: createDashboardStats(),
      trend: [createTrendPoint()],
      models: [createModelStat()]
    })
    getUserUsageTrend.mockResolvedValue({
      trend: [createUserTrendPoint()],
      start_date: '2026-04-03',
      end_date: '2026-04-04',
      granularity: 'hour'
    })
    getUserSpendingRanking.mockResolvedValue({
      ranking: [createRankingItem()],
      total_actual_cost: 4.2,
      total_requests: 88,
      total_tokens: 9999,
      start_date: '2026-04-03',
      end_date: '2026-04-04'
    })
  })

  it('loads dashboard stats, trends, and ranking data together', async () => {
    const showError = vi.fn()
    const state = useDashboardViewData({
      t: (key: string) => key,
      showError
    })

    await state.loadDashboardStats()

    expect(getSnapshotV2).toHaveBeenCalledWith({
      start_date: state.startDate.value,
      end_date: state.endDate.value,
      granularity: 'hour',
      include_stats: true,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    expect(getUserUsageTrend).toHaveBeenCalledWith({
      start_date: state.startDate.value,
      end_date: state.endDate.value,
      granularity: 'hour',
      limit: 12
    })
    expect(getUserSpendingRanking).toHaveBeenCalledWith({
      start_date: state.startDate.value,
      end_date: state.endDate.value,
      limit: 12
    })

    expect(state.stats.value?.total_users).toBe(100)
    expect(state.trendData.value).toEqual([createTrendPoint()])
    expect(state.modelStats.value).toEqual([createModelStat()])
    expect(state.userTrend.value).toEqual([createUserTrendPoint()])
    expect(state.rankingItems.value).toEqual([createRankingItem()])
    expect(state.rankingTotalActualCost.value).toBe(4.2)
    expect(state.rankingTotalRequests.value).toBe(88)
    expect(state.rankingTotalTokens.value).toBe(9999)
    expect(showError).not.toHaveBeenCalled()
  })

  it('updates range and granularity when the date picker changes', async () => {
    const state = useDashboardViewData({
      t: (key: string) => key,
      showError: vi.fn()
    })

    state.onDateRangeChange({
      startDate: '2026-04-01',
      endDate: '2026-04-04',
      preset: null
    })
    await flushPromises()

    expect(state.startDate.value).toBe('2026-04-01')
    expect(state.endDate.value).toBe('2026-04-04')
    expect(state.granularity.value).toBe('day')
    expect(getSnapshotV2).toHaveBeenLastCalledWith({
      start_date: '2026-04-01',
      end_date: '2026-04-04',
      granularity: 'day',
      include_stats: false,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })

    state.onDateRangeChange({
      startDate: '2026-04-04',
      endDate: '2026-04-04',
      preset: 'today'
    })
    await flushPromises()

    expect(state.granularity.value).toBe('hour')
    expect(getSnapshotV2).toHaveBeenLastCalledWith({
      start_date: '2026-04-04',
      end_date: '2026-04-04',
      granularity: 'hour',
      include_stats: false,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
  })

  it('keeps newer concurrent responses and ignores stale results', async () => {
    const firstSnapshot = createDeferred<{
      stats?: DashboardStats
      trend?: TrendDataPoint[]
      models?: ModelStat[]
    }>()
    const secondSnapshot = createDeferred<{
      stats?: DashboardStats
      trend?: TrendDataPoint[]
      models?: ModelStat[]
    }>()
    const firstUserTrend = createDeferred<{
      trend?: UserUsageTrendPoint[]
      start_date: string
      end_date: string
      granularity: string
    }>()
    const secondUserTrend = createDeferred<{
      trend?: UserUsageTrendPoint[]
      start_date: string
      end_date: string
      granularity: string
    }>()
    const firstRanking = createDeferred<{
      ranking?: UserSpendingRankingItem[]
      total_actual_cost?: number
      total_requests?: number
      total_tokens?: number
      start_date: string
      end_date: string
    }>()
    const secondRanking = createDeferred<{
      ranking?: UserSpendingRankingItem[]
      total_actual_cost?: number
      total_requests?: number
      total_tokens?: number
      start_date: string
      end_date: string
    }>()

    getSnapshotV2
      .mockReturnValueOnce(firstSnapshot.promise)
      .mockReturnValueOnce(secondSnapshot.promise)
    getUserUsageTrend
      .mockReturnValueOnce(firstUserTrend.promise)
      .mockReturnValueOnce(secondUserTrend.promise)
    getUserSpendingRanking
      .mockReturnValueOnce(firstRanking.promise)
      .mockReturnValueOnce(secondRanking.promise)

    const state = useDashboardViewData({
      t: (key: string) => key,
      showError: vi.fn()
    })

    state.startDate.value = '2026-04-01'
    state.endDate.value = '2026-04-02'
    const firstLoad = state.loadChartData()

    state.startDate.value = '2026-04-03'
    state.endDate.value = '2026-04-04'
    const secondLoad = state.loadChartData()

    secondSnapshot.resolve({
      trend: [createTrendPoint({ date: '2026-04-04', requests: 200 })],
      models: [createModelStat({ model: 'new-model' })]
    })
    secondUserTrend.resolve({
      trend: [createUserTrendPoint({ user_id: 2, tokens: 900 })],
      start_date: '2026-04-03',
      end_date: '2026-04-04',
      granularity: 'hour'
    })
    secondRanking.resolve({
      ranking: [createRankingItem({ user_id: 2, email: 'new@example.com' })],
      total_actual_cost: 9.9,
      total_requests: 120,
      total_tokens: 22222,
      start_date: '2026-04-03',
      end_date: '2026-04-04'
    })
    await flushPromises()

    firstSnapshot.resolve({
      trend: [createTrendPoint({ date: '2026-04-02', requests: 10 })],
      models: [createModelStat({ model: 'old-model' })]
    })
    firstUserTrend.resolve({
      trend: [createUserTrendPoint({ user_id: 1, tokens: 100 })],
      start_date: '2026-04-01',
      end_date: '2026-04-02',
      granularity: 'hour'
    })
    firstRanking.resolve({
      ranking: [createRankingItem({ user_id: 1, email: 'old@example.com' })],
      total_actual_cost: 1.1,
      total_requests: 10,
      total_tokens: 1000,
      start_date: '2026-04-01',
      end_date: '2026-04-02'
    })

    await Promise.all([firstLoad, secondLoad])

    expect(state.trendData.value).toEqual([createTrendPoint({ date: '2026-04-04', requests: 200 })])
    expect(state.modelStats.value).toEqual([createModelStat({ model: 'new-model' })])
    expect(state.userTrend.value).toEqual([createUserTrendPoint({ user_id: 2, tokens: 900 })])
    expect(state.rankingItems.value).toEqual([createRankingItem({ user_id: 2, email: 'new@example.com' })])
    expect(state.rankingTotalActualCost.value).toBe(9.9)
    expect(state.rankingTotalRequests.value).toBe(120)
    expect(state.rankingTotalTokens.value).toBe(22222)
  })
})
