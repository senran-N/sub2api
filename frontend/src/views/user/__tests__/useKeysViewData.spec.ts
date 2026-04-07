import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed } from 'vue'
import { useKeysViewData } from '../keys/useKeysViewData'

const { listKeys, getDashboardApiKeysUsage, getAvailableGroups, getUserGroupRates } = vi.hoisted(
  () => ({
    listKeys: vi.fn(),
    getDashboardApiKeysUsage: vi.fn(),
    getAvailableGroups: vi.fn(),
    getUserGroupRates: vi.fn()
  })
)

vi.mock('@/api', () => ({
  keysAPI: {
    list: listKeys
  },
  usageAPI: {
    getDashboardApiKeysUsage
  },
  userGroupsAPI: {
    getAvailable: getAvailableGroups,
    getUserGroupRates: getUserGroupRates
  }
}))

describe('useKeysViewData', () => {
  beforeEach(() => {
    listKeys.mockReset()
    getDashboardApiKeysUsage.mockReset()
    getAvailableGroups.mockReset()
    getUserGroupRates.mockReset()

    listKeys.mockResolvedValue({
      items: [{ id: 7, name: 'demo-key' }],
      total: 1,
      pages: 1
    })
    getDashboardApiKeysUsage.mockResolvedValue({
      stats: {
        7: {
          today_actual_cost: 1,
          total_actual_cost: 2
        }
      }
    })
    getAvailableGroups.mockResolvedValue([{ id: 3, name: 'Starter' }])
    getUserGroupRates.mockResolvedValue({ 3: 1 })
  })

  function createComposable() {
    const showError = vi.fn()
    const composable = useKeysViewData({
      t: (key: string) => key,
      showError,
      fetchPublicSettings: vi.fn().mockResolvedValue(null),
      publicSettings: computed(() => null)
    })

    return {
      composable,
      showError
    }
  }

  it('loads keys and usage stats', async () => {
    const setup = createComposable()

    await setup.composable.loadApiKeys()

    expect(listKeys).toHaveBeenCalledWith(
      1,
      expect.any(Number),
      {},
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
    expect(getDashboardApiKeysUsage).toHaveBeenCalledWith([7], expect.any(Object))
    expect(setup.composable.apiKeys.value).toEqual([{ id: 7, name: 'demo-key' }])
  })

  it('surfaces request details and ignores abort failures', async () => {
    const setup = createComposable()

    listKeys.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'keys-load-failed'
        }
      }
    })
    await setup.composable.loadApiKeys()
    expect(setup.showError).toHaveBeenCalledWith('keys-load-failed')

    listKeys.mockRejectedValueOnce({ name: 'AbortError' })
    await setup.composable.loadApiKeys()
    expect(setup.showError).toHaveBeenCalledTimes(1)
  })
})
