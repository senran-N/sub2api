import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { Proxy } from '@/types'
import { useProxyListData } from '../useProxyListData'

const { listProxies } = vi.hoisted(() => ({
  listProxies: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    proxies: {
      list: listProxies
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

function createComposable() {
  const proxies = ref<Proxy[]>([])
  const loading = ref(false)
  const searchQuery = ref(' edge ')
  const filters = {
    protocol: 'socks5',
    status: 'active'
  }
  const pagination = {
    page: 2,
    page_size: 20,
    total: 0,
    pages: 0
  }
  const showError = vi.fn()

  const composable = useProxyListData({
    proxies,
    loading,
    searchQuery,
    filters,
    pagination,
    t: (key: string) => key,
    showError
  })

  return {
    composable,
    loading,
    pagination,
    proxies,
    searchQuery,
    showError
  }
}

describe('useProxyListData', () => {
  beforeEach(() => {
    listProxies.mockReset()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads proxies and updates pagination', async () => {
    const setup = createComposable()
    listProxies.mockResolvedValue({
      items: [createProxy({ id: 7 })],
      total: 1,
      pages: 1
    })

    await setup.composable.loadProxies()

    expect(listProxies).toHaveBeenCalledWith(
      2,
      20,
      {
        protocol: 'socks5',
        status: 'active',
        search: 'edge'
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal)
      })
    )
    expect(setup.proxies.value[0].id).toBe(7)
    expect(setup.pagination.total).toBe(1)
    expect(setup.pagination.pages).toBe(1)
    expect(setup.loading.value).toBe(false)
  })

  it('debounces search and resets page before loading', async () => {
    const setup = createComposable()
    listProxies.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0
    })

    setup.pagination.page = 9
    setup.composable.handleSearch()
    vi.advanceTimersByTime(299)
    expect(listProxies).not.toHaveBeenCalled()

    vi.advanceTimersByTime(1)
    await vi.runAllTimersAsync()

    expect(setup.pagination.page).toBe(1)
    expect(listProxies).toHaveBeenCalledTimes(1)
  })

  it('updates page and page size then reloads', async () => {
    const setup = createComposable()
    listProxies.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0
    })

    await setup.composable.handlePageChange(4)
    expect(setup.pagination.page).toBe(4)

    await setup.composable.handlePageSizeChange(50)
    expect(setup.pagination.page).toBe(1)
    expect(setup.pagination.page_size).toBe(50)
    expect(listProxies).toHaveBeenCalledTimes(2)
  })

  it('surfaces request details and ignores abort failures', async () => {
    const setup = createComposable()

    listProxies.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'proxy-load-failed'
        }
      }
    })
    await setup.composable.loadProxies()
    expect(setup.showError).toHaveBeenCalledWith('proxy-load-failed')

    listProxies.mockRejectedValueOnce({ name: 'AbortError' })
    await setup.composable.loadProxies()
    expect(setup.showError).toHaveBeenCalledTimes(1)
  })
})
