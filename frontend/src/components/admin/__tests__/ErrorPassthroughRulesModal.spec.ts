import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import ErrorPassthroughRulesModal from '../ErrorPassthroughRulesModal.vue'

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

    mount(ErrorPassthroughRulesModal, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          ConfirmDialog: ConfirmDialogStub,
          Icon: IconStub,
        },
      },
    })

    await flushPromises()

    expect(showError).toHaveBeenCalledWith('error passthrough detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })
})
