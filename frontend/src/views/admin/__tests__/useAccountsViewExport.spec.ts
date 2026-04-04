import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import {
  downloadAccountsExportJson,
  useAccountsViewExport
} from '../useAccountsViewExport'

describe('useAccountsViewExport', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('exports selected or filtered data and closes the dialog', async () => {
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const exportData = vi.fn().mockResolvedValue({ items: [1] })
    const downloadJson = vi.fn()
    const showExportDataDialog = ref(true)
    const state = useAccountsViewExport({
      t: (key: string) => key,
      showSuccess,
      showError,
      selectedIds: ref([9]),
      includeProxyOnExport: ref(true),
      params: {
        platform: 'openai',
        type: 'oauth',
        status: 'active',
        search: 'main'
      },
      showExportDataDialog,
      exportData,
      downloadJson
    })

    await state.handleExportData()
    expect(exportData).toHaveBeenCalledWith({
      ids: [9],
      includeProxies: true
    })
    expect(downloadJson).toHaveBeenCalledWith(
      { items: [1] },
      expect.stringMatching(/^sub2api-account-\d{14}\.json$/)
    )
    expect(showSuccess).toHaveBeenCalledWith('admin.accounts.dataExported')
    expect(showError).not.toHaveBeenCalled()
    expect(showExportDataDialog.value).toBe(false)
  })

  it('surfaces export failures and provides browser download helper', async () => {
    const showError = vi.fn()
    const state = useAccountsViewExport({
      t: (key: string) => key,
      showSuccess: vi.fn(),
      showError,
      selectedIds: ref([]),
      includeProxyOnExport: ref(false),
      params: {
        platform: 'openai',
        type: 'oauth',
        status: 'active',
        search: 'main'
      },
      showExportDataDialog: ref(true),
      exportData: vi.fn().mockRejectedValue({ message: 'boom' }),
      downloadJson: vi.fn()
    })

    await state.handleExportData()
    expect(showError).toHaveBeenCalledWith('boom')

    const originalCreateObjectURL = URL.createObjectURL
    const originalRevokeObjectURL = URL.revokeObjectURL
    const anchor = document.createElement('a')
    const click = vi.spyOn(anchor, 'click').mockImplementation(() => {})
    const createElement = vi
      .spyOn(document, 'createElement')
      .mockReturnValue(anchor)
    const createObjectURL = vi.fn(() => 'blob:test')
    const revokeObjectURL = vi.fn()
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      writable: true,
      value: createObjectURL
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      writable: true,
      value: revokeObjectURL
    })

    downloadAccountsExportJson({ hello: 'world' }, 'accounts.json')
    expect(createElement).toHaveBeenCalledWith('a')
    expect(createObjectURL).toHaveBeenCalledTimes(1)
    expect(click).toHaveBeenCalledTimes(1)
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:test')

    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      writable: true,
      value: originalCreateObjectURL
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      writable: true,
      value: originalRevokeObjectURL
    })
  })
})
