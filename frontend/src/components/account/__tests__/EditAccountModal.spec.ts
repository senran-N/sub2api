import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

const { showErrorMock, showSuccessMock, showInfoMock, updateAccountMock, checkMixedChannelRiskMock } = vi.hoisted(() => ({
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showInfoMock: vi.fn(),
  updateAccountMock: vi.fn(),
  checkMixedChannelRiskMock: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock,
    showInfo: showInfoMock
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isSimpleMode: true
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      update: updateAccountMock,
      checkMixedChannelRisk: checkMixedChannelRiskMock
    }
  }
}))

vi.mock('@/api/admin/accounts', () => ({
  getAntigravityDefaultModelMapping: vi.fn()
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

import EditAccountModal from '../EditAccountModal.vue'

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: {
      type: Boolean,
      default: false
    }
  },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const ModelWhitelistSelectorStub = defineComponent({
  name: 'ModelWhitelistSelector',
  props: {
    modelValue: {
      type: Array,
      default: () => []
    }
  },
  emits: ['update:modelValue'],
  template: `
    <div>
      <button
        type="button"
        data-testid="rewrite-to-snapshot"
        @click="$emit('update:modelValue', ['gpt-5.2-2025-12-11'])"
      >
        rewrite
      </button>
      <span data-testid="model-whitelist-value">
        {{ Array.isArray(modelValue) ? modelValue.join(',') : '' }}
      </span>
    </div>
  `
})

function buildAccount() {
  return {
    id: 1,
    name: 'OpenAI Key',
    notes: '',
    platform: 'openai',
    type: 'apikey',
    credentials: {
      api_key: 'sk-test',
      base_url: 'https://api.openai.com',
      model_mapping: {
        'gpt-5.2': 'gpt-5.2'
      }
    },
    extra: {},
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    rate_multiplier: 1,
    status: 'active',
    group_ids: [],
    expires_at: null,
    auto_pause_on_expired: false
  } as any
}

function buildGrokUpstreamAccount() {
  return {
    ...buildAccount(),
    name: 'Grok Upstream',
    platform: 'grok',
    type: 'upstream',
    credentials: {
      api_key: 'xai-upstream-key',
      base_url: 'https://grok-proxy.example',
      model_mapping: {
        'grok-4': 'grok-4'
      },
      pool_mode: true,
      pool_mode_retry_count: 5
    },
    extra: {
      quota_limit: 42
    }
  } as any
}

function buildGrokSessionAccount() {
  return {
    ...buildAccount(),
    name: 'Grok Session',
    platform: 'grok',
    type: 'session',
    credentials: {
      session_token: 'grok-session-existing'
    },
    extra: {
      grok: {
        auth_mode: 'session',
        auth_fingerprint: 'sha256:ab12...cd34',
        tier: {
          normalized: 'heavy',
          source: 'quota_sync',
          confidence: 0.92
        },
        capabilities: {
          operations: ['chat', 'video'],
          models: ['grok-4', 'grok-4-video']
        },
        quota_windows: {
          auto: {
            remaining: 17,
            total: 150,
            source: 'sync',
            reset_at: '2026-04-20T02:00:00Z'
          }
        },
        sync_state: {
          last_sync_at: '2026-04-20T00:00:00Z',
          last_probe_at: '2026-04-20T01:00:00Z',
          last_probe_ok_at: '2026-04-20T00:45:00Z',
          last_probe_error_at: '2026-04-20T01:00:00Z',
          last_probe_error: 'API returned 401 Unauthorized',
          last_probe_status_code: 401
        },
        runtime_state: {
          last_fail_at: '2026-04-20T01:05:00Z',
          last_fail_reason: 'video tier required'
        }
      }
    }
  } as any
}

function mountModal(account = buildAccount()) {
  return mount(EditAccountModal, {
    props: {
      show: true,
      account,
      proxies: [],
      groups: []
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Select: true,
        Icon: true,
        ProxySelector: true,
        GroupSelector: true,
        ModelWhitelistSelector: ModelWhitelistSelectorStub
      }
    }
  })
}

describe('EditAccountModal', () => {
  beforeEach(() => {
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    showInfoMock.mockReset()
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
  })

  it('reopening the same account rehydrates the OpenAI whitelist from props', async () => {
    const account = buildAccount()
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    expect(wrapper.get('[data-testid="model-whitelist-value"]').text()).toBe('gpt-5.2')

    await wrapper.get('[data-testid="rewrite-to-snapshot"]').trigger('click')
    expect(wrapper.get('[data-testid="model-whitelist-value"]').text()).toBe('gpt-5.2-2025-12-11')

    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })

    expect(wrapper.get('[data-testid="model-whitelist-value"]').text()).toBe('gpt-5.2')

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.model_mapping).toEqual({
      'gpt-5.2': 'gpt-5.2'
    })
  })

  it('update failure prefers backend detail message', async () => {
    updateAccountMock.mockRejectedValue({
      response: {
        data: {
          detail: 'edit detail error'
        }
      },
      message: 'generic error'
    })

    const wrapper = mountModal()

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(showErrorMock).toHaveBeenCalledWith('edit detail error')
  })

  it('applies the Grok official preset onto the Grok upstream base URL field', async () => {
    const wrapper = mountModal(buildGrokUpstreamAccount())
    const baseUrlInput = wrapper.get('input[placeholder="https://api.x.ai"]')
    const grokPreset = wrapper
      .findAll('button')
      .find((button) => button.text() === 'admin.accounts.grok.baseUrlPresets.official')

    expect(grokPreset).toBeTruthy()

    await grokPreset!.trigger('click')

    expect((baseUrlInput.element as HTMLInputElement).value).toBe('https://api.x.ai')
  })

  it('preserves the existing Grok session token when the edit field is left blank', async () => {
    const wrapper = mountModal(buildGrokSessionAccount())

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.session_token).toBe(
      'grok-session-existing'
    )
  })

  it('shows the Grok runtime summary for session accounts', () => {
    const wrapper = mountModal(buildGrokSessionAccount())

    expect(wrapper.text()).toContain('admin.accounts.grok.runtime.title')
    expect(wrapper.text()).toContain('admin.accounts.grok.runtime.tiers.heavy')
    expect(wrapper.text()).toContain('sha256:ab12...cd34')
    expect(wrapper.text()).toContain('admin.accounts.grok.runtime.capabilities.video')
    expect(wrapper.text()).toContain('admin.accounts.grok.runtime.probeFailed')
    expect(wrapper.text()).toContain('admin.accounts.grok.runtime.httpStatus')
    expect(wrapper.text()).toContain('API returned 401 Unauthorized')
    expect(wrapper.text()).toContain('video tier required')
  })

  it('hydrates Grok upstream settings and keeps the existing API key on save', async () => {
    const account = buildGrokUpstreamAccount()
    const wrapper = mountModal(account)

    const baseUrlInput = wrapper.get('input[placeholder="https://api.x.ai"]')
    expect((baseUrlInput.element as HTMLInputElement).value).toBe('https://grok-proxy.example')

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials).toMatchObject({
      api_key: 'xai-upstream-key',
      base_url: 'https://grok-proxy.example',
      model_mapping: {
        'grok-4': 'grok-4'
      },
      pool_mode: true,
      pool_mode_retry_count: 5
    })
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.quota_limit).toBe(42)
  })
})
