import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import AnnouncementReadStatusDialog from '../AnnouncementReadStatusDialog.vue'

const getReadStatusMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    announcements: {
      getReadStatus: (...args: any[]) => getReadStatusMock(...args)
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock
  })
}))

vi.mock('@/composables/usePersistedPageSize', () => ({
  getPersistedPageSize: () => 20
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

describe('AnnouncementReadStatusDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getReadStatusMock.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    })
  })

  it('loads immediately when mounted open and prefers backend detail failures', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    getReadStatusMock.mockRejectedValue({
      response: {
        data: {
          detail: 'read status detail error'
        }
      },
      message: 'generic read status error'
    })

    mount(AnnouncementReadStatusDialog, {
      props: {
        show: true,
        announcementId: 12
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          DataTable: true,
          Pagination: true,
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(getReadStatusMock).toHaveBeenCalledWith(12, 1, 20, '')
    expect(showErrorMock).toHaveBeenCalledWith('read status detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })
})
