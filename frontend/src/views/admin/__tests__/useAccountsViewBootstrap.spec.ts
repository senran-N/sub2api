import { flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { useAccountsViewBootstrap } from '../useAccountsViewBootstrap'

describe('useAccountsViewBootstrap', () => {
  it('loads reference data, registers listeners, and disposes cleanly', async () => {
    const load = vi.fn().mockResolvedValue(undefined)
    const fetchProxies = vi.fn().mockResolvedValue([{ id: 1, name: 'Proxy A' }])
    const fetchGroups = vi.fn().mockResolvedValue([{ id: 2, name: 'Group A' }])
    const closeActionMenu = vi.fn()
    const initializeAutoRefresh = vi.fn()
    const disposeAutoRefresh = vi.fn()
    const windowTarget = {
      addEventListener: vi.fn(),
      removeEventListener: vi.fn()
    } as any

    const state = useAccountsViewBootstrap({
      t: (key: string) => key,
      showError: vi.fn(),
      load,
      fetchProxies,
      fetchGroups,
      closeActionMenu,
      initializeAutoRefresh,
      disposeAutoRefresh,
      windowTarget
    })

    state.initialize()
    await flushPromises()

    expect(load).toHaveBeenCalledTimes(1)
    expect(fetchProxies).toHaveBeenCalledTimes(1)
    expect(fetchGroups).toHaveBeenCalledTimes(1)
    expect(initializeAutoRefresh).toHaveBeenCalledTimes(1)
    expect(state.proxies.value).toEqual([{ id: 1, name: 'Proxy A' }])
    expect(state.groups.value).toEqual([{ id: 2, name: 'Group A' }])

    const scrollHandler = windowTarget.addEventListener.mock.calls[0][1]
    scrollHandler()

    expect(closeActionMenu).toHaveBeenCalledTimes(1)

    state.dispose()

    expect(windowTarget.removeEventListener).toHaveBeenCalledWith('scroll', scrollHandler, true)
    expect(disposeAutoRefresh).toHaveBeenCalledTimes(1)
  })

  it('surfaces reference data errors and clears stale options', async () => {
    const showError = vi.fn()
    const state = useAccountsViewBootstrap({
      t: (key: string) => key,
      showError,
      load: vi.fn().mockResolvedValue(undefined),
      fetchProxies: vi.fn().mockRejectedValue(new Error('boom')),
      fetchGroups: vi.fn().mockResolvedValue([{ id: 2, name: 'Group A' }]),
      closeActionMenu: vi.fn(),
      initializeAutoRefresh: vi.fn(),
      disposeAutoRefresh: vi.fn(),
      windowTarget: {
        addEventListener: vi.fn(),
        removeEventListener: vi.fn()
      } as any
    })

    state.proxies.value = [{ id: 99, name: 'Old Proxy' }] as any
    state.groups.value = [{ id: 88, name: 'Old Group' }] as any

    state.initialize()
    await flushPromises()

    expect(state.proxies.value).toEqual([])
    expect(state.groups.value).toEqual([])
    expect(showError).toHaveBeenCalledWith('boom')
  })
})
