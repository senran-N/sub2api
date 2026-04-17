import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import ErrorPassthroughRulesModal from '../ErrorPassthroughRulesModal.vue'
import type { ErrorPassthroughRule } from '@/api/admin/errorPassthrough'

const mockListRules = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    errorPassthrough: {
      list: (...args: any[]) => mockListRules(...args),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      toggleEnabled: vi.fn(),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false },
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const ConfirmDialogStub = defineComponent({
  name: 'ConfirmDialogStub',
  props: {
    show: { type: Boolean, default: false },
  },
  emits: ['confirm', 'cancel'],
  template: '<div v-if="show" />',
})

const IconStub = defineComponent({
  name: 'IconStub',
  template: '<span />',
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createRule(overrides: Partial<ErrorPassthroughRule> = {}): ErrorPassthroughRule {
  return {
    id: 1,
    name: 'Rate limit passthrough',
    enabled: true,
    priority: 10,
    error_codes: [429],
    keywords: [],
    match_mode: 'any',
    platforms: ['openai'],
    passthrough_code: true,
    response_code: null,
    passthrough_body: true,
    custom_message: null,
    skip_monitoring: false,
    description: 'default',
    created_at: '2026-04-17T00:00:00Z',
    updated_at: '2026-04-17T00:00:00Z',
    ...overrides
  }
}

function mountModal(props: { show?: boolean } = {}) {
  return mount(ErrorPassthroughRulesModal, {
    props: {
      show: props.show ?? true,
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        ConfirmDialog: ConfirmDialogStub,
        Icon: IconStub,
      },
    },
  })
}

describe('ErrorPassthroughRulesModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('prefers backend detail when loading rules fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockListRules.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'error passthrough detail error',
        },
      },
      message: 'generic error passthrough error',
    })

    mountModal()

    await flushPromises()

    expect(showError).toHaveBeenCalledWith('error passthrough detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('keeps the latest rule list when the modal is reopened before the previous load resolves', async () => {
    const firstLoad = createDeferred<ErrorPassthroughRule[]>()
    const secondLoad = createDeferred<ErrorPassthroughRule[]>()

    mockListRules
      .mockImplementationOnce(() => firstLoad.promise)
      .mockImplementationOnce(() => secondLoad.promise)

    const wrapper = mountModal()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })

    secondLoad.resolve([createRule({ id: 2, name: 'Quota passthrough' })])
    await flushPromises()

    firstLoad.resolve([createRule({ id: 1, name: 'Old stale rule' })])
    await flushPromises()

    expect((wrapper.vm as any).rules).toEqual([
      expect.objectContaining({ id: 2, name: 'Quota passthrough' })
    ])
    expect(showError).not.toHaveBeenCalled()
  })
})
