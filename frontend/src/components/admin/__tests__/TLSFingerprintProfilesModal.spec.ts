import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import TLSFingerprintProfilesModal from '../TLSFingerprintProfilesModal.vue'
import type { TLSFingerprintProfile } from '@/api/admin/tlsFingerprintProfile'

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

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createProfile(overrides: Partial<TLSFingerprintProfile> = {}): TLSFingerprintProfile {
  return {
    id: 1,
    name: 'Chrome Stable',
    description: 'stable client',
    enable_grease: true,
    cipher_suites: [4865, 4866],
    curves: [29],
    point_formats: [],
    signature_algorithms: [1027],
    alpn_protocols: ['h2'],
    supported_versions: [772],
    key_share_groups: [29],
    psk_modes: [1],
    extensions: [0, 10, 11],
    created_at: '2026-04-17T00:00:00Z',
    updated_at: '2026-04-17T00:00:00Z',
    ...overrides
  }
}

function mountModal(props: { show?: boolean } = {}) {
  return mount(TLSFingerprintProfilesModal, {
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

    mountModal()

    await flushPromises()

    expect(showError).toHaveBeenCalledWith('tls profiles detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('keeps the latest profile list when the modal is reopened before the previous load resolves', async () => {
    const firstLoad = createDeferred<TLSFingerprintProfile[]>()
    const secondLoad = createDeferred<TLSFingerprintProfile[]>()

    mockListProfiles
      .mockImplementationOnce(() => firstLoad.promise)
      .mockImplementationOnce(() => secondLoad.promise)

    const wrapper = mountModal()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })

    secondLoad.resolve([createProfile({ id: 2, name: 'Edge Stable' })])
    await flushPromises()

    firstLoad.resolve([createProfile({ id: 1, name: 'Old Chrome' })])
    await flushPromises()

    expect((wrapper.vm as any).profiles).toEqual([
      expect.objectContaining({ id: 2, name: 'Edge Stable' })
    ])
    expect(showError).not.toHaveBeenCalled()
  })
})
