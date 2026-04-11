import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsErrorDetailModal from '../OpsErrorDetailModal.vue'

const getRequestErrorDetailMock = vi.fn()
const getUpstreamErrorDetailMock = vi.fn()
const listRequestErrorUpstreamErrorsMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getRequestErrorDetail: (...args: any[]) => getRequestErrorDetailMock(...args),
    getUpstreamErrorDetail: (...args: any[]) => getUpstreamErrorDetailMock(...args),
    listRequestErrorUpstreamErrors: (...args: any[]) => listRequestErrorUpstreamErrorsMock(...args)
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: showErrorMock
  })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
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
    show: { type: Boolean, default: false },
    title: { type: String, default: '' }
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /></div>'
})

describe('OpsErrorDetailModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getRequestErrorDetailMock.mockResolvedValue({})
    getUpstreamErrorDetailMock.mockResolvedValue({})
    listRequestErrorUpstreamErrorsMock.mockResolvedValue({ items: [] })
  })

  it('prefers backend detail when upstream error detail loading fails', async () => {
    getUpstreamErrorDetailMock.mockRejectedValue({
      response: {
        data: {
          detail: 'ops detail error'
        }
      },
      message: 'generic ops error'
    })

    mount(OpsErrorDetailModal, {
      props: {
        show: true,
        errorId: 9,
        errorType: 'upstream'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(getUpstreamErrorDetailMock).toHaveBeenCalledWith(9)
    expect(showErrorMock).toHaveBeenCalledWith('ops detail error')
  })
})
