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

function createDeferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void

  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })

  return { promise, resolve, reject }
}

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

  it('keeps the newest cleanup task page when earlier loads resolve late', async () => {
    const firstLoad = createDeferred<{
      items: Array<{ id: number; status: string; created_at: string; filters: Record<string, string>; deleted_rows: number }>
      total: number
      page: number
      page_size: number
    }>()
    const secondLoad = createDeferred<{
      items: Array<{ id: number; status: string; created_at: string; filters: Record<string, string>; deleted_rows: number }>
      total: number
      page: number
      page_size: number
    }>()

    listCleanupTasksMock
      .mockReturnValueOnce(firstLoad.promise)
      .mockReturnValueOnce(secondLoad.promise)

    const wrapper = mount(UsageCleanupDialog, {
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

    await wrapper.find('button.btn.btn-ghost.btn-sm').trigger('click')

    secondLoad.resolve({
      items: [
        {
          id: 22,
          status: 'running',
          created_at: '2026-04-17T08:00:00Z',
          filters: {
            start_time: '2026-04-01T00:00:00Z',
            end_time: '2026-04-11T00:00:00Z'
          },
          deleted_rows: 7
        }
      ],
      total: 1,
      page: 1,
      page_size: 5
    })
    await flushPromises()

    firstLoad.resolve({
      items: [
        {
          id: 11,
          status: 'failed',
          created_at: '2026-04-16T08:00:00Z',
          filters: {
            start_time: '2026-03-01T00:00:00Z',
            end_time: '2026-03-11T00:00:00Z'
          },
          deleted_rows: 99
        }
      ],
      total: 1,
      page: 1,
      page_size: 5
    })
    await flushPromises()

    expect(wrapper.text()).toContain('#22')
    expect(wrapper.text()).toContain('7')
    expect(wrapper.text()).not.toContain('#11')
    expect(wrapper.text()).not.toContain('99')
  })

  it('ignores stale cleanup task responses after close and reopen', async () => {
    const firstLoad = createDeferred<{
      items: Array<{ id: number; status: string; created_at: string; filters: Record<string, string>; deleted_rows: number }>
      total: number
      page: number
      page_size: number
    }>()
    const secondLoad = createDeferred<{
      items: Array<{ id: number; status: string; created_at: string; filters: Record<string, string>; deleted_rows: number }>
      total: number
      page: number
      page_size: number
    }>()

    listCleanupTasksMock
      .mockReturnValueOnce(firstLoad.promise)
      .mockReturnValueOnce(secondLoad.promise)

    const wrapper = mount(UsageCleanupDialog, {
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

    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })

    secondLoad.resolve({
      items: [
        {
          id: 33,
          status: 'succeeded',
          created_at: '2026-04-17T09:00:00Z',
          filters: {
            start_time: '2026-04-10T00:00:00Z',
            end_time: '2026-04-11T00:00:00Z'
          },
          deleted_rows: 3
        }
      ],
      total: 1,
      page: 1,
      page_size: 5
    })
    await flushPromises()

    firstLoad.resolve({
      items: [
        {
          id: 44,
          status: 'failed',
          created_at: '2026-04-15T09:00:00Z',
          filters: {
            start_time: '2026-04-01T00:00:00Z',
            end_time: '2026-04-02T00:00:00Z'
          },
          deleted_rows: 55
        }
      ],
      total: 1,
      page: 1,
      page_size: 5
    })
    await flushPromises()

    expect(wrapper.text()).toContain('#33')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).not.toContain('#44')
    expect(wrapper.text()).not.toContain('55')
  })
})
