import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import type { RedeemCode } from '@/types'
import { useRedeemViewData } from '../useRedeemViewData'

const { listCodes, exportCodes, deleteCode, batchDelete } = vi.hoisted(() => ({
  listCodes: vi.fn(),
  exportCodes: vi.fn(),
  deleteCode: vi.fn(),
  batchDelete: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    redeem: {
      list: listCodes,
      exportCodes,
      delete: deleteCode,
      batchDelete
    }
  }
}))

function createCode(overrides: Partial<RedeemCode> = {}): RedeemCode {
  return {
    id: 1,
    code: 'ABC',
    type: 'balance',
    value: 10,
    status: 'unused',
    used_by: null,
    used_at: null,
    created_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

function createComposable() {
  const showError = vi.fn()
  const showInfo = vi.fn()
  const showSuccess = vi.fn()
  const copyToClipboard = vi.fn().mockResolvedValue(true)
  const composable = useRedeemViewData({
    t: (key: string) => key,
    showError,
    showInfo,
    showSuccess,
    copyToClipboard
  })

  return {
    composable,
    showError,
    showInfo,
    showSuccess,
    copyToClipboard
  }
}

describe('useRedeemViewData', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-04-04T00:00:00Z'))

    listCodes.mockReset()
    exportCodes.mockReset()
    deleteCode.mockReset()
    batchDelete.mockReset()

    listCodes.mockResolvedValue({
      items: [createCode()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    exportCodes.mockResolvedValue(new Blob(['code'], { type: 'text/csv' }))
    deleteCode.mockResolvedValue({ message: 'ok' })
    batchDelete.mockResolvedValue({ deleted: 1, message: 'ok' })
  })

  it('loads codes, debounces search, and updates pagination state', async () => {
    const setup = createComposable()

    await setup.composable.loadCodes()
    expect(listCodes).toHaveBeenCalledWith(
      1,
      expect.any(Number),
      {
        type: undefined,
        status: undefined,
        search: undefined
      },
      expect.any(Object)
    )
    expect(setup.composable.codes.value).toHaveLength(1)

    setup.composable.searchQuery.value = 'vip'
    setup.composable.filters.type = 'subscription'
    setup.composable.handleSearch()
    await vi.advanceTimersByTimeAsync(300)
    await nextTick()

    expect(listCodes).toHaveBeenLastCalledWith(
      1,
      expect.any(Number),
      {
        type: 'subscription',
        status: undefined,
        search: 'vip'
      },
      expect.any(Object)
    )

    setup.composable.handlePageChange(3)
    expect(listCodes).toHaveBeenLastCalledWith(
      3,
      expect.any(Number),
      {
        type: 'subscription',
        status: undefined,
        search: 'vip'
      },
      expect.any(Object)
    )

    setup.composable.handlePageSizeChange(50)
    expect(listCodes).toHaveBeenLastCalledWith(
      1,
      50,
      {
        type: 'subscription',
        status: undefined,
        search: 'vip'
      },
      expect.any(Object)
    )
  })

  it('exports codes, marks copied rows, and confirms deletes', async () => {
    const setup = createComposable()
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    const originalCreateElement = document.createElement.bind(document)
    const clickSpy = vi.fn()

    window.URL.createObjectURL = vi.fn(() => 'blob:redeem-export') as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn() as typeof window.URL.revokeObjectURL
    document.createElement = vi.fn((tagName: string) => {
      if (tagName === 'a') {
        const link = originalCreateElement('a')
        link.click = clickSpy
        return link
      }
      return originalCreateElement(tagName)
    }) as typeof document.createElement

    await setup.composable.handleExportCodes()
    expect(exportCodes).toHaveBeenCalledWith({
      type: undefined,
      status: undefined
    })
    expect(clickSpy).toHaveBeenCalledTimes(1)
    expect(setup.showSuccess).toHaveBeenCalledWith('admin.redeem.codesExported')

    await setup.composable.copyCodeToClipboard('ABC')
    expect(setup.copyToClipboard).toHaveBeenCalledWith('ABC', 'admin.redeem.copied')
    expect(setup.composable.copiedCode.value).toBe('ABC')
    await vi.advanceTimersByTimeAsync(2000)
    expect(setup.composable.copiedCode.value).toBeNull()

    setup.composable.handleDelete(createCode())
    await setup.composable.confirmDelete()
    expect(deleteCode).toHaveBeenCalledWith(1)
    expect(setup.showSuccess).toHaveBeenCalledWith('admin.redeem.codeDeleted')

    document.createElement = originalCreateElement
    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
  })

  it('reports empty unused sets and batch deletes unused codes', async () => {
    const setup = createComposable()

    listCodes.mockResolvedValueOnce({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    })
    await setup.composable.confirmDeleteUnused()
    expect(setup.showInfo).toHaveBeenCalledWith('admin.redeem.noUnusedCodes')
    expect(batchDelete).not.toHaveBeenCalled()

    listCodes.mockResolvedValueOnce({
      items: [createCode({ id: 7 }), createCode({ id: 9, code: 'XYZ' })],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })
    listCodes.mockResolvedValueOnce({
      items: [createCode({ id: 7 })],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })

    await setup.composable.confirmDeleteUnused()
    expect(batchDelete).toHaveBeenCalledWith([7, 9])
    expect(setup.showSuccess).toHaveBeenCalledWith('admin.redeem.codesDeleted')
  })
})
