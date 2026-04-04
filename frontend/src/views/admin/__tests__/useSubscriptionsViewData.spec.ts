import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useSubscriptionsViewData } from '../useSubscriptionsViewData'

const { listSubscriptions, getAllGroups } = vi.hoisted(() => ({
  listSubscriptions: vi.fn(),
  getAllGroups: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    subscriptions: {
      list: listSubscriptions
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

describe('useSubscriptionsViewData', () => {
  beforeEach(() => {
    listSubscriptions.mockReset()
    getAllGroups.mockReset()

    listSubscriptions.mockResolvedValue({
      items: [{ id: 1 }],
      total: 1,
      pages: 1
    })
    getAllGroups.mockResolvedValue([{ id: 2, name: 'Group A' }])
  })

  it('loads subscriptions and groups and applies list mutations', async () => {
    const showLoadError = vi.fn()
    const state = useSubscriptionsViewData({ showLoadError })

    await state.loadInitialData()

    expect(listSubscriptions).toHaveBeenCalledWith(
      1,
      expect.any(Number),
      expect.objectContaining({
        status: 'active',
        sort_by: 'created_at',
        sort_order: 'desc'
      }),
      expect.any(Object)
    )
    expect(getAllGroups).toHaveBeenCalledTimes(1)
    expect(state.subscriptions.value).toEqual([{ id: 1 }])
    expect(state.groups.value).toEqual([{ id: 2, name: 'Group A' }])

    state.filters.platform = 'openai'
    state.applyFilters()
    expect(state.pagination.page).toBe(1)

    state.handlePageChange(3)
    expect(state.pagination.page).toBe(3)

    state.handlePageSizeChange(50)
    expect(state.pagination.page_size).toBe(50)
    expect(state.pagination.page).toBe(1)

    state.handleSort('expires_at', 'asc')
    expect(state.sortState).toEqual({
      sort_by: 'expires_at',
      sort_order: 'asc'
    })
  })

  it('reports non-abort subscription load failures', async () => {
    const showLoadError = vi.fn()
    const state = useSubscriptionsViewData({ showLoadError })
    listSubscriptions.mockRejectedValueOnce(new Error('boom'))

    await state.loadSubscriptions()

    expect(showLoadError).toHaveBeenCalledTimes(1)
  })
})
