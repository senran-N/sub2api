import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import AnnouncementReadStatusDialog from '../AnnouncementReadStatusDialog.vue'
import type { AnnouncementUserReadStatus } from '@/types'

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

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createReadStatusItem(
  overrides: Partial<AnnouncementUserReadStatus> = {}
): AnnouncementUserReadStatus {
  return {
    user_id: 1,
    email: 'user@example.com',
    username: 'user',
    balance: 0,
    eligible: true,
    read_at: '2026-04-17T00:00:00Z',
    ...overrides
  }
}

function mountDialog(props: { show?: boolean; announcementId?: number | null } = {}) {
  return mount(AnnouncementReadStatusDialog, {
    props: {
      show: props.show ?? true,
      announcementId: props.announcementId ?? 12
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
}

describe('AnnouncementReadStatusDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    getReadStatusMock.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    })
  })

  afterEach(() => {
    vi.useRealTimers()
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

    mountDialog()

    await flushPromises()

    expect(getReadStatusMock).toHaveBeenCalledWith(12, 1, 20, '', {
      signal: expect.any(AbortSignal)
    })
    expect(showErrorMock).toHaveBeenCalledWith('read status detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('keeps the latest announcement results when announcement changes before the previous request resolves', async () => {
    const firstResponse = createDeferred<{
      items: AnnouncementUserReadStatus[]
      total: number
      page: number
      page_size: number
      pages: number
    }>()
    const secondResponse = createDeferred<{
      items: AnnouncementUserReadStatus[]
      total: number
      page: number
      page_size: number
      pages: number
    }>()

    getReadStatusMock
      .mockImplementationOnce(() => firstResponse.promise)
      .mockImplementationOnce(() => secondResponse.promise)

    const wrapper = mountDialog({ announcementId: 12 })
    await wrapper.setProps({ announcementId: 13 })

    secondResponse.resolve({
      items: [createReadStatusItem({ user_id: 2, email: 'new@example.com' })],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    await flushPromises()

    firstResponse.resolve({
      items: [createReadStatusItem({ user_id: 1, email: 'old@example.com' })],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    await flushPromises()

    expect((wrapper.vm as any).items).toEqual([
      expect.objectContaining({ email: 'new@example.com' })
    ])
    expect(getReadStatusMock).toHaveBeenNthCalledWith(1, 12, 1, 20, '', {
      signal: expect.any(AbortSignal)
    })
    expect(getReadStatusMock).toHaveBeenNthCalledWith(2, 13, 1, 20, '', {
      signal: expect.any(AbortSignal)
    })
  })

  it('keeps the latest search results when overlapping searches resolve out of order', async () => {
    const wrapper = mountDialog()
    await flushPromises()

    const firstSearch = createDeferred<{
      items: AnnouncementUserReadStatus[]
      total: number
      page: number
      page_size: number
      pages: number
    }>()
    const secondSearch = createDeferred<{
      items: AnnouncementUserReadStatus[]
      total: number
      page: number
      page_size: number
      pages: number
    }>()

    getReadStatusMock
      .mockImplementationOnce(() => firstSearch.promise)
      .mockImplementationOnce(() => secondSearch.promise)

    await wrapper.get('input').setValue('alice')
    vi.advanceTimersByTime(300)
    await flushPromises()

    await wrapper.get('input').setValue('bob')
    vi.advanceTimersByTime(300)
    await flushPromises()

    secondSearch.resolve({
      items: [createReadStatusItem({ user_id: 3, email: 'bob@example.com' })],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    await flushPromises()

    firstSearch.resolve({
      items: [createReadStatusItem({ user_id: 4, email: 'alice@example.com' })],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    await flushPromises()

    expect((wrapper.vm as any).items).toEqual([
      expect.objectContaining({ email: 'bob@example.com' })
    ])
    expect((wrapper.vm as any).loading).toBe(false)
  })
})
