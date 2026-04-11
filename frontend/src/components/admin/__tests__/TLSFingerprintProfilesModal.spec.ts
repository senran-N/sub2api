import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import TLSFingerprintProfilesModal from '../TLSFingerprintProfilesModal.vue'

const mockListProfiles = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    tlsFingerprintProfiles: {
      list: (...args: any[]) => mockListProfiles(...args),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
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

describe('TLSFingerprintProfilesModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('prefers backend detail when loading profiles fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockListProfiles.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'tls profiles detail error',
        },
      },
      message: 'generic tls profiles error',
    })

    mount(TLSFingerprintProfilesModal, {
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

    expect(showError).toHaveBeenCalledWith('tls profiles detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })
})
