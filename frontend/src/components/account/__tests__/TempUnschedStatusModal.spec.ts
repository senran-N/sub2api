import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import TempUnschedStatusModal from '../TempUnschedStatusModal.vue'
import type { Account, TempUnschedulableStatus } from '@/types'

const {
  getTempUnschedulableStatusMock,
  recoverStateMock,
  showSuccessMock,
  showErrorMock
} = vi.hoisted(() => ({
  getTempUnschedulableStatusMock: vi.fn(),
  recoverStateMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getTempUnschedulableStatus: getTempUnschedulableStatusMock,
      recoverState: recoverStateMock
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
    name: 'Account Alpha',
    platform: 'openai',
    type: 'oauth',
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: true,
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

function makeStatusResponse(overrides: Partial<TempUnschedulableStatus> = {}): TempUnschedulableStatus {
  return {
    active: true,
    state: {
      until_unix: Math.floor(Date.now() / 1000) + 3600,
      triggered_at_unix: Math.floor(Date.now() / 1000) - 600,
      status_code: 429,
      matched_keyword: 'rate limit',
      rule_index: 0,
      error_message: 'too many requests'
    },
    ...overrides
  }
}

describe('TempUnschedStatusModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getTempUnschedulableStatusMock.mockResolvedValue(makeStatusResponse())
    recoverStateMock.mockResolvedValue(makeAccount({ id: 1, name: 'Recovered Account' }))
  })

  it('loads status immediately when mounted open', async () => {
    const wrapper = mount(TempUnschedStatusModal, {
      props: {
        show: true,
        account: makeAccount({ id: 42, name: 'Status Account' })
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub
        }
      }
    })

    await flushPromises()

    expect(getTempUnschedulableStatusMock).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('Status Account')
    expect(wrapper.text()).toContain('too many requests')
  })

  it('ignores a stale status load after account switch', async () => {
    const loadRequest = createDeferred<TempUnschedulableStatus>()
    getTempUnschedulableStatusMock.mockReturnValueOnce(loadRequest.promise)
    getTempUnschedulableStatusMock.mockResolvedValueOnce(
      makeStatusResponse({
        state: {
          until_unix: Math.floor(Date.now() / 1000) + 7200,
          triggered_at_unix: Math.floor(Date.now() / 1000) - 300,
          status_code: 503,
          matched_keyword: 'overload',
          rule_index: 1,
          error_message: 'new context overload'
        }
      })
    )

    const wrapper = mount(TempUnschedStatusModal, {
      props: {
        show: true,
        account: makeAccount({ id: 1, name: 'Old Account' })
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub
        }
      }
    })

    await flushPromises()

    await wrapper.setProps({
      account: makeAccount({ id: 2, name: 'New Account' })
    })
    await flushPromises()

    loadRequest.resolve(
      makeStatusResponse({
        state: {
          until_unix: Math.floor(Date.now() / 1000) + 3600,
          triggered_at_unix: Math.floor(Date.now() / 1000) - 900,
          status_code: 429,
          matched_keyword: 'stale keyword',
          rule_index: 0,
          error_message: 'stale status payload'
        }
      })
    )
    await flushPromises()

    expect(wrapper.text()).toContain('New Account')
    expect(wrapper.text()).toContain('new context overload')
    expect(wrapper.text()).not.toContain('stale status payload')
    expect(showErrorMock).not.toHaveBeenCalled()
  })

  it('ignores a stale reset success after close and reopen', async () => {
    const resetRequest = createDeferred<Account>()
    recoverStateMock.mockReturnValueOnce(resetRequest.promise)

    const wrapper = mount(TempUnschedStatusModal, {
      props: {
        show: true,
        account: makeAccount({ id: 1, name: 'Reset Account' })
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub
        }
      }
    })

    await flushPromises()

    await wrapper.get('.btn-primary').trigger('click')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({
      show: true,
      account: makeAccount({ id: 2, name: 'Fresh Account' })
    })
    await flushPromises()

    resetRequest.resolve(makeAccount({ id: 1, name: 'Recovered Old Account' }))
    await flushPromises()

    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('reset')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
    expect(wrapper.text()).toContain('Fresh Account')
  })
})
