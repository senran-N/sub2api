import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import CreateAccountModal from '../CreateAccountModal.vue'

const {
  createAccountMock,
  checkMixedChannelRiskMock,
  generateAuthUrlMock,
  exchangeCodeMock,
  refreshOpenAITokenMock,
  listTlsFingerprintProfilesMock,
  geminiGenerateAuthUrlMock,
  geminiExchangeCodeMock,
  geminiCapabilitiesMock,
  antigravityGenerateAuthUrlMock,
  antigravityExchangeCodeMock,
  antigravityRefreshTokenMock,
  fetchAntigravityDefaultMappingsMock,
  showSuccessMock,
  showErrorMock,
  showWarningMock,
  showInfoMock
} = vi.hoisted(() => ({
  createAccountMock: vi.fn(),
  checkMixedChannelRiskMock: vi.fn(),
  generateAuthUrlMock: vi.fn(),
  exchangeCodeMock: vi.fn(),
  refreshOpenAITokenMock: vi.fn(),
  listTlsFingerprintProfilesMock: vi.fn(),
  geminiGenerateAuthUrlMock: vi.fn(),
  geminiExchangeCodeMock: vi.fn(),
  geminiCapabilitiesMock: vi.fn(),
  antigravityGenerateAuthUrlMock: vi.fn(),
  antigravityExchangeCodeMock: vi.fn(),
  antigravityRefreshTokenMock: vi.fn(),
  fetchAntigravityDefaultMappingsMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  showWarningMock: vi.fn(),
  showInfoMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      create: createAccountMock,
      checkMixedChannelRisk: checkMixedChannelRiskMock,
      generateAuthUrl: generateAuthUrlMock,
      exchangeCode: exchangeCodeMock,
      refreshOpenAIToken: refreshOpenAITokenMock
    },
    tlsFingerprintProfiles: {
      list: listTlsFingerprintProfilesMock
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
    showError: showErrorMock,
    showWarning: showWarningMock,
    showInfo: showInfoMock
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isSimpleMode: false
  })
}))

vi.mock('@/composables/useModelWhitelist', async () => {
  const actual = await vi.importActual<typeof import('@/composables/useModelWhitelist')>(
    '@/composables/useModelWhitelist'
  )
  return {
    ...actual,
    fetchAntigravityDefaultMappings: fetchAntigravityDefaultMappingsMock
  }
})

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

const ConfirmDialogStub = defineComponent({
  name: 'ConfirmDialogStub',
  props: {
    show: { type: Boolean, default: false },
    message: { type: String, default: '' }
  },
  emits: ['confirm', 'cancel'],
  template: `
    <div v-if="show" data-testid="mixed-channel-warning">
      <span>{{ message }}</span>
      <button type="button" data-testid="confirm-dialog-confirm" @click="$emit('confirm')">confirm</button>
      <button type="button" data-testid="confirm-dialog-cancel" @click="$emit('cancel')">cancel</button>
    </div>
  `
})

const OAuthAuthorizationFlowStub = defineComponent({
  name: 'OAuthAuthorizationFlowStub',
  emits: [
    'generate-url',
    'cookie-auth',
    'validate-refresh-token',
    'validate-mobile-refresh-token'
  ],
  setup(_, { emit, expose }) {
    const exposed = {
      authCode: 'auth-code',
      oauthState: 'oauth-state',
      projectId: '',
      sessionKey: 'session-key',
      refreshToken: 'refresh-token',
      sessionToken: 'session-token',
      inputMethod: 'manual',
      reset: vi.fn()
    }

    expose(exposed)

    return {
      emitValidateRefreshToken: () => emit('validate-refresh-token', 'refresh-token')
    }
  },
  template: `
    <div>
      <button
        type="button"
        data-testid="oauth-validate-refresh-token"
        @click="emitValidateRefreshToken"
      >
        validate-refresh-token
      </button>
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

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((button) => button.text().includes(text))
}

function mountModal() {
  return mount(CreateAccountModal, {
    props: {
      show: true,
      proxies: [],
      groups: []
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        ConfirmDialog: ConfirmDialogStub,
        OAuthAuthorizationFlow: OAuthAuthorizationFlowStub,
        ProxySelector: true,
        GroupSelector: true,
        ModelWhitelistSelector: true,
        QuotaLimitCard: true,
        Select: true,
        Icon: true
      }
    }
  })
}

describe('CreateAccountModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    createAccountMock.mockResolvedValue(undefined)
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    generateAuthUrlMock.mockResolvedValue({
      auth_url: 'https://auth.example/authorize?state=oauth-state',
      session_id: 'session-1'
    })
    exchangeCodeMock.mockResolvedValue({})
    refreshOpenAITokenMock.mockResolvedValue({
      access_token: 'access-token',
      refresh_token: 'refresh-token',
      expires_at: 1710000000
    })
    listTlsFingerprintProfilesMock.mockResolvedValue([])
    geminiGenerateAuthUrlMock.mockResolvedValue({
      auth_url: 'https://gemini.example/auth',
      session_id: 'gemini-session',
      state: 'gemini-state'
    })
    geminiExchangeCodeMock.mockResolvedValue({})
    geminiCapabilitiesMock.mockResolvedValue(null)
    antigravityGenerateAuthUrlMock.mockResolvedValue({
      auth_url: 'https://antigravity.example/auth',
      session_id: 'antigravity-session',
      state: 'antigravity-state'
    })
    antigravityExchangeCodeMock.mockResolvedValue({})
    antigravityRefreshTokenMock.mockResolvedValue({})
    fetchAntigravityDefaultMappingsMock.mockResolvedValue([])
  })

  it('ignores a stale direct create success after close and reopen', async () => {
    const createRequest = createDeferred<void>()
    createAccountMock.mockReturnValueOnce(createRequest.promise)

    const wrapper = mountModal()

    await wrapper.get('[data-tour="account-form-name"]').setValue('Claude Console Account')
    const consoleButton = findButtonByText(wrapper, 'admin.accounts.claudeConsole')
    expect(consoleButton).toBeTruthy()
    await consoleButton!.trigger('click')
    await wrapper.get('input[type="password"]').setValue('sk-ant-123')

    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(createAccountMock).toHaveBeenCalledTimes(1)

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({ show: true })
    await flushPromises()

    createRequest.resolve()
    await flushPromises()

    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('created')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
  })

  it('ignores a stale batch completion after returning to step 1 and switching platform', async () => {
    const createRequest = createDeferred<void>()
    createAccountMock.mockReturnValueOnce(createRequest.promise)

    const wrapper = mountModal()

    await wrapper.get('[data-tour="account-form-name"]').setValue('OpenAI OAuth Account')
    const openAIButton = findButtonByText(wrapper, 'OpenAI')
    expect(openAIButton).toBeTruthy()
    await openAIButton!.trigger('click')
    await flushPromises()

    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()

    await wrapper.get('[data-testid="oauth-validate-refresh-token"]').trigger('click')
    await flushPromises()

    expect(refreshOpenAITokenMock).toHaveBeenCalledWith(
      'refresh-token',
      null,
      '/admin/openai/refresh-token',
      undefined
    )
    expect(createAccountMock).toHaveBeenCalledTimes(1)

    const backButton = findButtonByText(wrapper, 'common.back')
    expect(backButton).toBeTruthy()
    await backButton!.trigger('click')
    await flushPromises()

    const anthropicButton = findButtonByText(wrapper, 'Anthropic')
    expect(anthropicButton).toBeTruthy()
    await anthropicButton!.trigger('click')
    await flushPromises()

    createRequest.resolve()
    await flushPromises()

    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showWarningMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('created')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
  })
})
