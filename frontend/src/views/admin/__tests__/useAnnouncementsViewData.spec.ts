import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { useAnnouncementsViewData } from '../useAnnouncementsViewData'

const { listAnnouncements } = vi.hoisted(() => ({
  listAnnouncements: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    announcements: {
      list: listAnnouncements
    }
  }
}))

describe('useAnnouncementsViewData', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    listAnnouncements.mockReset()
    listAnnouncements.mockResolvedValue({
      items: [{ id: 1, title: 'Maintenance' }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
  })

  it('loads announcements and updates search/filter/pagination state', async () => {
    const showError = vi.fn()
    const state = useAnnouncementsViewData({
      t: (key: string) => key,
      showError
    })

    await state.loadAnnouncements()
    expect(listAnnouncements).toHaveBeenCalledWith(
      1,
      expect.any(Number),
      {
        status: undefined,
        search: undefined
      },
      expect.any(Object)
    )
    expect(state.announcements.value).toEqual([{ id: 1, title: 'Maintenance' }])

    state.searchQuery.value = 'window'
    state.handleSearch()
    await vi.advanceTimersByTimeAsync(300)
    await nextTick()
    expect(listAnnouncements).toHaveBeenLastCalledWith(
      1,
      expect.any(Number),
      {
        status: undefined,
        search: 'window'
      },
      expect.any(Object)
    )

    state.filters.status = 'active'
    state.handleStatusChange()
    expect(listAnnouncements).toHaveBeenLastCalledWith(
      1,
      expect.any(Number),
      {
        status: 'active',
        search: 'window'
      },
      expect.any(Object)
    )

    state.handlePageChange(3)
    expect(listAnnouncements).toHaveBeenLastCalledWith(
      3,
      expect.any(Number),
      {
        status: 'active',
        search: 'window'
      },
      expect.any(Object)
    )

    state.handlePageSizeChange(50)
    expect(listAnnouncements).toHaveBeenLastCalledWith(
      1,
      50,
      {
        status: 'active',
        search: 'window'
      },
      expect.any(Object)
    )
    expect(showError).not.toHaveBeenCalled()
  })

  it('reports non-abort failures and ignores cancellations', async () => {
    const showError = vi.fn()
    const state = useAnnouncementsViewData({
      t: (key: string) => key,
      showError
    })

    listAnnouncements.mockRejectedValueOnce(new Error('boom'))
    await state.loadAnnouncements()
    expect(showError).toHaveBeenCalledWith('admin.announcements.failedToLoad')

    listAnnouncements.mockRejectedValueOnce({ name: 'AbortError' })
    await state.loadAnnouncements()
    expect(showError).toHaveBeenCalledTimes(1)
  })
})
