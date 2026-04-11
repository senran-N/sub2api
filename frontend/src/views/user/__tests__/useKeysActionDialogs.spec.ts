import { afterEach, describe, expect, it, vi } from 'vitest'
import { createApp, defineComponent, h, ref } from 'vue'
import type { ApiKey } from '@/types'
import { useKeysActionDialogs } from '../keys/useKeysActionDialogs'

function createApiKey(overrides: Partial<ApiKey> = {}): ApiKey {
  return {
    id: 1,
    user_id: 7,
    key: 'sk-test-key',
    name: 'Demo Key',
    group_id: 2,
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    reset_5h_at: null,
    reset_1d_at: null,
    reset_7d_at: null,
    ...overrides
  }
}

function mountComposable() {
  const apiKeys = ref<ApiKey[]>([])
  const showError = vi.fn()
  const showSuccess = vi.fn()
  const loadApiKeys = vi.fn(async () => {})
  const create = vi.fn()
  const update = vi.fn()
  const deleteKey = vi.fn()
  const toggleStatus = vi.fn()

  let composable!: ReturnType<typeof useKeysActionDialogs>

  const app = createApp(
    defineComponent({
      setup() {
        composable = useKeysActionDialogs({
          t: (key: string) => key,
          showError,
          showSuccess,
          apiKeys,
          publicSettings: ref(null),
          keysAPI: {
            create,
            update,
            delete: deleteKey,
            toggleStatus
          },
          loadApiKeys,
          isOnboardingSubmitStep: () => false,
          advanceOnboardingStep: vi.fn()
        })

        return () => h('div')
      }
    })
  )

  const container = document.createElement('div')
  app.mount(container)

  return {
    app,
    composable,
    deleteKey,
    loadApiKeys,
    showError,
    showSuccess
  }
}

describe('useKeysActionDialogs', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('uses response detail when deleting a key fails', async () => {
    const setup = mountComposable()
    const key = createApiKey({ id: 9 })
    setup.deleteKey.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'delete-blocked-by-server'
        }
      }
    })

    setup.composable.confirmDelete(key)
    await setup.composable.handleDelete()

    expect(setup.showError).toHaveBeenCalledWith('delete-blocked-by-server')
    expect(setup.showSuccess).not.toHaveBeenCalled()
    setup.app.unmount()
  })

  it('falls back to plain error messages when request detail is missing', async () => {
    const setup = mountComposable()
    const key = createApiKey({ id: 11 })
    setup.deleteKey.mockRejectedValueOnce(new Error('network down'))

    setup.composable.confirmDelete(key)
    await setup.composable.handleDelete()

    expect(setup.showError).toHaveBeenCalledWith('network down')
    expect(setup.loadApiKeys).not.toHaveBeenCalled()
    setup.app.unmount()
  })
})
