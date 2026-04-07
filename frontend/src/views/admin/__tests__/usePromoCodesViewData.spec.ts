import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { usePromoCodesViewData } from '../promocodes/usePromoCodesViewData'

const { listPromoCodes } = vi.hoisted(() => ({
  listPromoCodes: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    promo: {
      list: listPromoCodes
    }
  }
}))

describe('usePromoCodesViewData', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    listPromoCodes.mockReset()
    listPromoCodes.mockResolvedValue({
      items: [{ id: 1, code: 'WELCOME' }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
  })

  it('loads promo codes and updates search/filter/pagination state', async () => {
    const showError = vi.fn()
    const copyToClipboard = vi.fn().mockResolvedValue(true)
    const state = usePromoCodesViewData({
      t: (key: string) => key,
      showError,
      copyToClipboard
    })

    await state.loadCodes()
    expect(listPromoCodes).toHaveBeenCalledWith(
      1,
      expect.any(Number),
      {
        status: undefined,
        search: undefined
      },
      expect.any(Object)
    )
    expect(state.codes.value).toEqual([{ id: 1, code: 'WELCOME' }])

    state.searchQuery.value = 'spring'
    state.handleSearch()
    await vi.advanceTimersByTimeAsync(300)
    await nextTick()
    expect(listPromoCodes).toHaveBeenLastCalledWith(
      1,
      expect.any(Number),
      {
        status: undefined,
        search: 'spring'
      },
      expect.any(Object)
    )

    state.filters.status = 'disabled'
    await state.loadCodes()
    expect(listPromoCodes).toHaveBeenLastCalledWith(
      1,
      expect.any(Number),
      {
        status: 'disabled',
        search: 'spring'
      },
      expect.any(Object)
    )

    state.handlePageChange(3)
    expect(listPromoCodes).toHaveBeenLastCalledWith(
      3,
      expect.any(Number),
      {
        status: 'disabled',
        search: 'spring'
      },
      expect.any(Object)
    )

    state.handlePageSizeChange(50)
    expect(listPromoCodes).toHaveBeenLastCalledWith(
      1,
      50,
      {
        status: 'disabled',
        search: 'spring'
      },
      expect.any(Object)
    )

    await state.handleCopyCode('WELCOME')
    expect(copyToClipboard).toHaveBeenCalledWith('WELCOME', 'admin.promo.copied')
    expect(state.copiedCode.value).toBe('WELCOME')
    await vi.advanceTimersByTimeAsync(2000)
    expect(state.copiedCode.value).toBeNull()
    expect(showError).not.toHaveBeenCalled()
  })

  it('reports non-abort failures and ignores cancellations', async () => {
    const showError = vi.fn()
    const state = usePromoCodesViewData({
      t: (key: string) => key,
      showError,
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    listPromoCodes.mockRejectedValueOnce({
      response: { data: { message: 'promo-load-failed' } }
    })
    await state.loadCodes()
    expect(showError).toHaveBeenCalledWith('promo-load-failed')

    listPromoCodes.mockRejectedValueOnce({ name: 'AbortError' })
    await state.loadCodes()
    expect(showError).toHaveBeenCalledTimes(1)
  })
})
