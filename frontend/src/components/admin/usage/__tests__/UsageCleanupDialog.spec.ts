import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UsageCleanupDialog from '../UsageCleanupDialog.vue'

const listCleanupTasksMock = vi.fn()
const createCleanupTaskMock = vi.fn()
const cancelCleanupTaskMock = vi.fn()
const showErrorMock = vi.fn()
const showSuccessMock = vi.fn()

vi.mock('@/api/admin/usage', () => ({
  default: {
    listCleanupTasks: (...args: any[]) => listCleanupTasksMock(...args),
    createCleanupTask: (...args: any[]) => createCleanupTaskMock(...args),
    cancelCleanupTask: (...args: any[]) => cancelCleanupTaskMock(...args)
  },
  adminUsageAPI: {
    listCleanupTasks: (...args: any[]) => listCleanupTasksMock(...args),
    createCleanupTask: (...args: any[]) => createCleanupTaskMock(...args),
    cancelCleanupTask: (...args: any[]) => cancelCleanupTaskMock(...args)
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
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

describe('UsageCleanupDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listCleanupTasksMock.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 5
    })
    createCleanupTaskMock.mockResolvedValue({})
    cancelCleanupTaskMock.mockResolvedValue({})
  })

  it('loads tasks immediately when mounted open and surfaces request details', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    listCleanupTasksMock.mockRejectedValue({
      response: {
        data: {
          detail: 'cleanup task detail error'
        }
      },
      message: 'generic cleanup error'
    })

    mount(UsageCleanupDialog, {
      props: {
        show: true,
        filters: {},
        startDate: '2026-04-01',
        endDate: '2026-04-11'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          ConfirmDialog: true,
          Pagination: true,
          UsageFilters: true
        }
      }
    })

    await flushPromises()

    expect(listCleanupTasksMock).toHaveBeenCalledWith({
      page: 1,
      page_size: 5
    })
    expect(showErrorMock).toHaveBeenCalledWith('cleanup task detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })
})
