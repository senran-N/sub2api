import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import ScheduledTestsPanel from '../ScheduledTestsPanel.vue'

const {
  showErrorMock,
  showSuccessMock,
  listByAccountMock,
  createMock,
  updateMock,
  deleteMock,
  listResultsMock
} = vi.hoisted(() => ({
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  listByAccountMock: vi.fn(),
  createMock: vi.fn(),
  updateMock: vi.fn(),
  deleteMock: vi.fn(),
  listResultsMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    scheduledTests: {
      listByAccount: listByAccountMock,
      create: createMock,
      update: updateMock,
      delete: deleteMock,
      listResults: listResultsMock
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock
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
    show: { type: Boolean, default: false },
    title: { type: String, default: '' }
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /></div>'
})

describe('ScheduledTestsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listByAccountMock.mockResolvedValue([])
    createMock.mockResolvedValue({})
    updateMock.mockResolvedValue({})
    deleteMock.mockResolvedValue(undefined)
    listResultsMock.mockResolvedValue([])
  })

  it('prefers backend detail when loading plans fails', async () => {
    listByAccountMock.mockRejectedValue({
      response: {
        data: {
          detail: 'scheduled tests detail'
        }
      },
      message: 'generic scheduled tests error'
    })

    mount(ScheduledTestsPanel, {
      props: {
        show: true,
        accountId: 7,
        modelOptions: []
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          ConfirmDialog: true,
          HelpTooltip: true,
          Select: true,
          Input: true,
          Toggle: true,
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(listByAccountMock).toHaveBeenCalledWith(7)
    expect(showErrorMock).toHaveBeenCalledWith('scheduled tests detail')
  })
})
