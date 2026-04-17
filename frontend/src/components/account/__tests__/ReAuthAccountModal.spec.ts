import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import ReAuthAccountModal from '../ReAuthAccountModal.vue'
import type { Account } from '@/types'

const {
  generateAuthUrlMock,
  exchangeCodeMock,
  refreshOpenAITokenMock,
  updateAccountMock,
  clearErrorMock,
  geminiGenerateAuthUrlMock,
  geminiExchangeCodeMock,
  geminiCapabilitiesMock,
  antigravityGenerateAuthUrlMock,
  antigravityExchangeCodeMock,
  antigravityRefreshTokenMock,
  showSuccessMock,
  showErrorMock
} = vi.hoisted(() => ({
  generateAuthUrlMock: vi.fn(),
  exchangeCodeMock: vi.fn(),
  refreshOpenAITokenMock: vi.fn(),
  updateAccountMock: vi.fn(),
  clearErrorMock: vi.fn(),
  geminiGenerateAuthUrlMock: vi.fn(),
  geminiExchangeCodeMock: vi.fn(),
  geminiCapabilitiesMock: vi.fn(),
  antigravityGenerateAuthUrlMock: vi.fn(),
  antigravityExchangeCodeMock: vi.fn(),
  antigravityRefreshTokenMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      generateAuthUrl: generateAuthUrlMock,
      exchangeCode: exchangeCodeMock,
      refreshOpenAIToken: refreshOpenAITokenMock,
      update: updateAccountMock,
      clearError: clearErrorMock
    },
    gemini: {
      generateAuthUrl: geminiGenerateAuthUrlMock,
      exchangeCode: geminiExchangeCodeMock,
      getCapabilities: geminiCapabilitiesMock
    },
    antigravity: {
      generateAuthUrl: antigravityGenerateAuthUrlMock,
      exchangeCode: antigravityExchangeCodeMock,
      refreshAntigravityToken: antigravityRefreshTokenMock
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock
  })
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

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false }
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const OAuthAuthorizationFlowStub = defineComponent({
  name: 'OAuthAuthorizationFlowStub',
  emits: ['generate-url', 'cookie-auth'],
  setup(_, { emit, expose }) {
    const exposed = {
      authCode: 'auth-code',
      oauthState: 'oauth-state',
      projectId: '',
      sessionKey: 'session-key',
      inputMethod: 'manual',
      reset: vi.fn()
    }

    expose(exposed)

    return {
      emitGenerateUrl: () => emit('generate-url'),
      emitCookieAuth: () => emit('cookie-auth', exposed.sessionKey)
    }
  },
  template: `
    <div>
      <button type="button" data-testid="reauth-generate-url" @click="emitGenerateUrl">generate</button>
      <button type="button" data-testid="reauth-cookie-auth" @click="emitCookieAuth">cookie</button>
    </div>
  `
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function makeAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'OpenAI Account',
    platform: 'openai',
    type: 'oauth',
    credentials: {},
    extra: {},
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    group_ids: [],
    notes: '',
    rate_multiplier: 1,
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

function mountModal(account = makeAccount()) {
  return mount(ReAuthAccountModal, {
    props: {
      show: true,
      account
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        OAuthAuthorizationFlow: OAuthAuthorizationFlowStub,
        Icon: true
      }
    }
  })
}

describe('ReAuthAccountModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    generateAuthUrlMock.mockResolvedValue({
      auth_url: 'https://auth.example/callback?state=oauth-state',
      session_id: 'session-1'
    })
    refreshOpenAITokenMock.mockResolvedValue(null)
    updateAccountMock.mockResolvedValue(undefined)
    clearErrorMock.mockResolvedValue(makeAccount({ id: 1, name: 'Updated Account' }))
    geminiCapabilitiesMock.mockResolvedValue(null)
  })

  it('ignores a stale OpenAI exchange after account switch', async () => {
    const exchangeRequest = createDeferred<{ access_token: string; expires_at: number }>()
    exchangeCodeMock.mockReturnValueOnce(exchangeRequest.promise)

    const wrapper = mountModal(makeAccount({ id: 1, name: 'Old Account' }))

    await wrapper.get('[data-testid="reauth-generate-url"]').trigger('click')
    await flushPromises()
    await wrapper.get('.btn-primary').trigger('click')

    expect(exchangeCodeMock).toHaveBeenCalledWith('/admin/openai/exchange-code', {
      session_id: 'session-1',
      code: 'auth-code',
      state: 'oauth-state'
    })

    await wrapper.setProps({
      account: makeAccount({ id: 2, name: 'New Account' })
    })
    await flushPromises()

    exchangeRequest.resolve({
      access_token: 'stale-token',
      expires_at: 1710000000
    })
    await flushPromises()

    expect(updateAccountMock).not.toHaveBeenCalled()
    expect(clearErrorMock).not.toHaveBeenCalled()
    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('reauthorized')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
  })

  it('ignores a stale reauthorize success after close and reopen', async () => {
    const updateRequest = createDeferred<void>()
    exchangeCodeMock.mockResolvedValueOnce({
      access_token: 'fresh-token',
      expires_at: 1710000000
    })
    updateAccountMock.mockReturnValueOnce(updateRequest.promise)

    const wrapper = mountModal(makeAccount({ id: 1, name: 'Closing Account' }))

    await wrapper.get('[data-testid="reauth-generate-url"]').trigger('click')
    await flushPromises()
    await wrapper.get('.btn-primary').trigger('click')
    await flushPromises()

    expect(updateAccountMock).toHaveBeenCalledWith(1, {
      type: 'oauth',
      credentials: {
        access_token: 'fresh-token',
        expires_at: 1710000000
      },
      extra: undefined
    })

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({
      show: true,
      account: makeAccount({ id: 2, name: 'Fresh Account' })
    })
    await flushPromises()

    updateRequest.resolve()
    await flushPromises()

    expect(clearErrorMock).not.toHaveBeenCalled()
    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('reauthorized')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
  })
})
