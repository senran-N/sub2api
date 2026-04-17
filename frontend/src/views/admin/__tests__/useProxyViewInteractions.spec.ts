import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { Proxy, ProxyAccountSummary } from '@/types'
import { useProxyViewInteractions } from '../proxies/useProxyViewInteractions'

const { batchDelete, deleteProxy, exportData, getProxyAccounts } = vi.hoisted(() => ({
  batchDelete: vi.fn(),
  deleteProxy: vi.fn(),
  exportData: vi.fn(),
  getProxyAccounts: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    proxies: {
      batchDelete,
      delete: deleteProxy,
      exportData,
      getProxyAccounts
    }
  }
}))

function createProxy(overrides: Partial<Proxy> = {}): Proxy {
  return {
    id: 1,
    name: 'Proxy',
    protocol: 'http',
    host: 'proxy.local',
    port: 8080,
    username: null,
    password: null,
    status: 'active',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

function createAccountSummary(
  overrides: Partial<ProxyAccountSummary> = {}
): ProxyAccountSummary {
  return {
    id: 1,
    name: 'Account',
    platform: 'openai',
    type: 'oauth',
    ...overrides
  }
}

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
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

function createComposable(selectedIds: number[] = []) {
  const selectedProxyIds = ref(new Set(selectedIds))
  const selectedCount = ref(selectedIds.length)
  const copyToClipboard = vi.fn()
  const clearSelectedProxies = vi.fn()
  const removeSelectedProxies = vi.fn()
  const loadProxies = vi.fn(async () => {})
  const showSuccess = vi.fn()
  const showError = vi.fn()
  const showInfo = vi.fn()

  const composable = useProxyViewInteractions({
    selectedCount,
    selectedProxyIds,
    filters: {
      protocol: 'http',
      status: 'active'
    },
    searchQuery: ref('edge'),
    copyToClipboard,
    clearSelectedProxies,
    removeSelectedProxies,
    loadProxies,
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key,
    showSuccess,
    showError,
    showInfo
  })

  return {
    composable,
    clearSelectedProxies,
    copyToClipboard,
    loadProxies,
    removeSelectedProxies,
    showError,
    showInfo,
    showSuccess
  }
}

describe('useProxyViewInteractions', () => {
  beforeEach(() => {
    batchDelete.mockReset()
    deleteProxy.mockReset()
    exportData.mockReset()
    getProxyAccounts.mockReset()
    vi.restoreAllMocks()
    vi.stubGlobal(
      'URL',
      Object.assign(URL, {
        createObjectURL: vi.fn(() => 'blob:proxy-export'),
        revokeObjectURL: vi.fn()
      })
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('copies proxy urls and toggles the copy menu', () => {
    const setup = createComposable()
    const proxy = createProxy({ username: 'alice', password: 'secret' })

    setup.composable.toggleCopyMenu(1)
    expect(setup.composable.copyMenuProxyId.value).toBe(1)

    setup.composable.copyProxyUrl(proxy)

    expect(setup.copyToClipboard).toHaveBeenCalledWith(
      'http://alice:secret@proxy.local:8080',
      'admin.proxies.urlCopied'
    )
    expect(setup.composable.copyMenuProxyId.value).toBeNull()
  })

  it('blocks deleting proxies that are still in use', () => {
    const setup = createComposable()

    setup.composable.handleDelete(createProxy({ account_count: 2 }))

    expect(setup.showError).toHaveBeenCalledWith('admin.proxies.deleteBlockedInUse')
    expect(setup.composable.showDeleteDialog.value).toBe(false)
  })

  it('confirms single and batch deletes then reloads data', async () => {
    const single = createComposable()
    deleteProxy.mockResolvedValue({ message: 'ok' })

    single.composable.handleDelete(createProxy({ id: 7 }))
    await single.composable.confirmDelete()

    expect(deleteProxy).toHaveBeenCalledWith(7)
    expect(single.removeSelectedProxies).toHaveBeenCalledWith([7])
    expect(single.loadProxies).toHaveBeenCalledTimes(1)

    const batch = createComposable([3, 4])
    batchDelete.mockResolvedValue({
      deleted_ids: [3, 4],
      skipped: []
    })

    await batch.composable.confirmBatchDelete()

    expect(batchDelete).toHaveBeenCalledWith([3, 4])
    expect(batch.clearSelectedProxies).toHaveBeenCalledTimes(1)
    expect(batch.loadProxies).toHaveBeenCalledTimes(1)
  })

  it('exports by current filters and loads proxy accounts', async () => {
    const setup = createComposable()
    exportData.mockResolvedValue({
      exported_at: '2026-04-04T00:00:00Z',
      proxies: [],
      accounts: []
    })
    getProxyAccounts.mockResolvedValue([createAccountSummary({ id: 9 })])

    const click = vi.fn()
    const createElement = vi
      .spyOn(document, 'createElement')
      .mockReturnValue({ click } as unknown as HTMLAnchorElement)

    await setup.composable.handleExportData()
    await setup.composable.openAccountsModal(createProxy({ id: 9 }))

    expect(exportData).toHaveBeenCalledWith({
      filters: {
        protocol: 'http',
        status: 'active',
        search: 'edge'
      }
    })
    expect(click).toHaveBeenCalledTimes(1)
    expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:proxy-export')
    expect(getProxyAccounts).toHaveBeenCalledWith(9)
    expect(setup.composable.proxyAccounts.value).toEqual([createAccountSummary({ id: 9 })])

    createElement.mockRestore()
  })

  it('shows response detail when export fails', async () => {
    const setup = createComposable()
    exportData.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'proxy-export-blocked'
        }
      }
    })

    await setup.composable.handleExportData()

    expect(setup.showError).toHaveBeenCalledWith('proxy-export-blocked')
    expect(setup.showSuccess).not.toHaveBeenCalled()
    expect(setup.composable.exportingData.value).toBe(false)
    expect(setup.composable.showExportDataDialog.value).toBe(false)
  })

  it('uses resolved request messages for delete, batch delete, and account load failures', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const single = createComposable()
    deleteProxy.mockRejectedValueOnce(new Error('delete unavailable'))
    single.composable.handleDelete(createProxy({ id: 8 }))
    await single.composable.confirmDelete()

    const batch = createComposable([3, 4])
    batchDelete.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'batch delete blocked'
        }
      }
    })
    await batch.composable.confirmBatchDelete()

    const accounts = createComposable()
    getProxyAccounts.mockRejectedValueOnce(new Error('accounts unavailable'))
    await accounts.composable.openAccountsModal(createProxy({ id: 12 }))

    expect(single.showError).toHaveBeenCalledWith('delete unavailable')
    expect(batch.showError).toHaveBeenCalledWith('batch delete blocked')
    expect(accounts.showError).toHaveBeenCalledWith('accounts unavailable')
    expect(consoleSpy).toHaveBeenCalledTimes(3)
  })

  it('keeps proxy accounts modal bound to the latest opened proxy', async () => {
    const setup = createComposable()
    const firstRequest = createDeferred<ProxyAccountSummary[]>()
    const secondRequest = createDeferred<ProxyAccountSummary[]>()

    getProxyAccounts
      .mockImplementationOnce(() => firstRequest.promise)
      .mockImplementationOnce(() => secondRequest.promise)

    const firstOpen = setup.composable.openAccountsModal(createProxy({ id: 3, name: 'First' }))
    const secondOpen = setup.composable.openAccountsModal(createProxy({ id: 4, name: 'Second' }))

    secondRequest.resolve([createAccountSummary({ id: 40, name: 'Second Account' })])
    await secondOpen

    expect(setup.composable.accountsProxy.value?.id).toBe(4)
    expect(setup.composable.proxyAccounts.value).toEqual([
      createAccountSummary({ id: 40, name: 'Second Account' })
    ])
    expect(setup.composable.accountsLoading.value).toBe(false)

    firstRequest.resolve([createAccountSummary({ id: 30, name: 'First Account' })])
    await firstOpen

    expect(setup.composable.accountsProxy.value?.id).toBe(4)
    expect(setup.composable.proxyAccounts.value).toEqual([
      createAccountSummary({ id: 40, name: 'Second Account' })
    ])
    expect(setup.showError).not.toHaveBeenCalled()
  })

  it('invalidates proxy accounts requests when the modal closes', async () => {
    const setup = createComposable()
    const pendingRequest = createDeferred<ProxyAccountSummary[]>()

    getProxyAccounts.mockImplementationOnce(() => pendingRequest.promise)

    const openPromise = setup.composable.openAccountsModal(createProxy({ id: 6, name: 'Closable' }))
    expect(setup.composable.showAccountsModal.value).toBe(true)
    expect(setup.composable.accountsLoading.value).toBe(true)

    setup.composable.closeAccountsModal()
    expect(setup.composable.showAccountsModal.value).toBe(false)
    expect(setup.composable.accountsLoading.value).toBe(false)
    expect(setup.composable.accountsProxy.value).toBeNull()

    pendingRequest.resolve([createAccountSummary({ id: 60, name: 'Should Stay Hidden' })])
    await openPromise

    expect(setup.composable.showAccountsModal.value).toBe(false)
    expect(setup.composable.accountsProxy.value).toBeNull()
    expect(setup.composable.proxyAccounts.value).toEqual([])
    expect(setup.showError).not.toHaveBeenCalled()
  })
})
