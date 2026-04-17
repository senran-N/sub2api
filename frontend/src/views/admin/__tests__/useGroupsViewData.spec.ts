import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useGroupsViewData } from '../groups/useGroupsViewData'
import type { AdminGroup } from '@/types'

const {
  listGroups,
  getUsageSummary,
  getCapacitySummary,
  getAllGroups,
  updateSortOrder
} = vi.hoisted(() => ({
  listGroups: vi.fn(),
  getUsageSummary: vi.fn(),
  getCapacitySummary: vi.fn(),
  getAllGroups: vi.fn(),
  updateSortOrder: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      list: listGroups,
      getUsageSummary,
      getCapacitySummary,
      getAll: getAllGroups,
      updateSortOrder
    }
  }
}))

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'Alpha',
    description: null,
    platform: 'anthropic',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    sora_image_price_360: null,
    sora_image_price_540: null,
    sora_video_price_per_request: null,
    sora_video_price_per_request_hd: null,
    sora_storage_quota_bytes: 0,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    allow_messages_dispatch: false,
    require_oauth_only: false,
    require_privacy_set: false,
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: true,
    simulate_claude_max_enabled: false,
    sort_order: 10,
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

describe('useGroupsViewData', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    listGroups.mockReset()
    getUsageSummary.mockReset()
    getCapacitySummary.mockReset()
    getAllGroups.mockReset()
    updateSortOrder.mockReset()

    listGroups.mockResolvedValue({
      items: [createGroup()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    getUsageSummary.mockResolvedValue([{ group_id: 1, today_cost: 1.2, total_cost: 9.8 }])
    getCapacitySummary.mockResolvedValue([
      {
        group_id: 1,
        concurrency_used: 2,
        concurrency_max: 10,
        sessions_used: 1,
        sessions_max: 4,
        rpm_used: 50,
        rpm_max: 100
      }
    ])
    getAllGroups.mockResolvedValue([
      createGroup({ id: 2, sort_order: 20 }),
      createGroup({ id: 1, sort_order: 10 })
    ])
    updateSortOrder.mockResolvedValue({ message: 'ok' })
  })

  it('loads groups and refreshes usage/capacity summaries', async () => {
    const showError = vi.fn()
    const showSuccess = vi.fn()
    const state = useGroupsViewData({
      t: (key: string) => key,
      showError,
      showSuccess
    })

    await state.loadGroups()
    await vi.runAllTimersAsync()

    expect(listGroups).toHaveBeenCalledWith(
      1,
      expect.any(Number),
      {
        platform: undefined,
        status: undefined,
        is_exclusive: undefined,
        search: undefined
      },
      { signal: expect.any(AbortSignal) }
    )
    expect(state.groups.value).toEqual([createGroup()])
    expect(state.pagination.total).toBe(1)
    expect(state.usageMap.value.get(1)).toEqual({ today_cost: 1.2, total_cost: 9.8 })
    expect(state.capacityMap.value.get(1)).toEqual({
      concurrencyUsed: 2,
      concurrencyMax: 10,
      sessionsUsed: 1,
      sessionsMax: 4,
      rpmUsed: 50,
      rpmMax: 100
    })
    expect(showError).not.toHaveBeenCalled()
    expect(showSuccess).not.toHaveBeenCalled()
  })

  it('debounces search and manages sort modal actions', async () => {
    const showSuccess = vi.fn()
    const state = useGroupsViewData({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess
    })

    state.pagination.page = 3
    state.searchQuery.value = 'anthropic'
    state.handleSearch()
    await vi.advanceTimersByTimeAsync(300)
    expect(state.pagination.page).toBe(1)
    expect(listGroups).toHaveBeenLastCalledWith(
      1,
      expect.any(Number),
      {
        platform: undefined,
        status: undefined,
        is_exclusive: undefined,
        search: 'anthropic'
      },
      { signal: expect.any(AbortSignal) }
    )

    await state.openSortModal()
    expect(state.showSortModal.value).toBe(true)
    expect(state.sortableGroups.value.map((group) => group.id)).toEqual([1, 2])

    await state.saveSortOrder()
    expect(updateSortOrder).toHaveBeenCalledWith([
      { id: 1, sort_order: 0 },
      { id: 2, sort_order: 10 }
    ])
    expect(showSuccess).toHaveBeenCalledWith('admin.groups.sortOrderUpdated')
    expect(state.showSortModal.value).toBe(false)
  })

  it('ignores aborted requests without surfacing errors', async () => {
    listGroups.mockRejectedValueOnce({ name: 'AbortError' })
    const showError = vi.fn()
    const state = useGroupsViewData({
      t: (key: string) => key,
      showError,
      showSuccess: vi.fn()
    })

    await state.loadGroups()
    expect(showError).not.toHaveBeenCalled()
  })

  it('keeps the latest usage and capacity summaries when overlapping summary requests resolve out of order', async () => {
    const firstUsage = createDeferred<Array<{ group_id: number; today_cost: number; total_cost: number }>>()
    const secondUsage = createDeferred<Array<{ group_id: number; today_cost: number; total_cost: number }>>()
    const firstCapacity = createDeferred<Array<{
      group_id: number
      concurrency_used: number
      concurrency_max: number
      sessions_used: number
      sessions_max: number
      rpm_used: number
      rpm_max: number
    }>>()
    const secondCapacity = createDeferred<Array<{
      group_id: number
      concurrency_used: number
      concurrency_max: number
      sessions_used: number
      sessions_max: number
      rpm_used: number
      rpm_max: number
    }>>()

    getUsageSummary
      .mockImplementationOnce(() => firstUsage.promise)
      .mockImplementationOnce(() => secondUsage.promise)
    getCapacitySummary
      .mockImplementationOnce(() => firstCapacity.promise)
      .mockImplementationOnce(() => secondCapacity.promise)

    const state = useGroupsViewData({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess: vi.fn()
    })

    const usageLoadOne = state.loadUsageSummary()
    const usageLoadTwo = state.loadUsageSummary()
    const capacityLoadOne = state.loadCapacitySummary()
    const capacityLoadTwo = state.loadCapacitySummary()

    secondUsage.resolve([{ group_id: 1, today_cost: 4.5, total_cost: 20 }])
    secondCapacity.resolve([
      {
        group_id: 1,
        concurrency_used: 4,
        concurrency_max: 10,
        sessions_used: 2,
        sessions_max: 4,
        rpm_used: 80,
        rpm_max: 100
      }
    ])
    await Promise.resolve()

    firstUsage.resolve([{ group_id: 1, today_cost: 1.2, total_cost: 9.8 }])
    firstCapacity.resolve([
      {
        group_id: 1,
        concurrency_used: 2,
        concurrency_max: 10,
        sessions_used: 1,
        sessions_max: 4,
        rpm_used: 50,
        rpm_max: 100
      }
    ])

    await Promise.all([usageLoadOne, usageLoadTwo, capacityLoadOne, capacityLoadTwo])

    expect(state.usageMap.value.get(1)).toEqual({ today_cost: 4.5, total_cost: 20 })
    expect(state.capacityMap.value.get(1)).toEqual({
      concurrencyUsed: 4,
      concurrencyMax: 10,
      sessionsUsed: 2,
      sessionsMax: 4,
      rpmUsed: 80,
      rpmMax: 100
    })
    expect(state.usageLoading.value).toBe(false)
  })

  it('surfaces shared request details for sort order failures', async () => {
    updateSortOrder.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'sort-update-failed'
        }
      }
    })
    const showError = vi.fn()
    const state = useGroupsViewData({
      t: (key: string) => key,
      showError,
      showSuccess: vi.fn()
    })
    state.sortableGroups.value = [createGroup({ id: 1, sort_order: 10 })]

    await state.saveSortOrder()

    expect(showError).toHaveBeenCalledWith('sort-update-failed')
  })
})
