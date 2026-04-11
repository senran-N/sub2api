import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import AnnouncementBell from '../AnnouncementBell.vue'

const showErrorMock = vi.fn()
const showSuccessMock = vi.fn()
const markAsReadMock = vi.fn()
const markAllAsReadMock = vi.fn()

const announcements = ref([
  {
    id: 1,
    title: 'Maintenance',
    content: 'Window notice',
    created_at: '2026-04-11T00:00:00.000Z',
    read_at: null,
    notify_mode: 'inbox'
  }
])
const loading = ref(false)
const currentPopup = ref(null)

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock
  })
}))

vi.mock('@/stores/announcements', () => ({
  useAnnouncementStore: () => ({
    announcements,
    loading,
    currentPopup,
    get unreadCount() {
      return announcements.value.filter((item) => !item.read_at).length
    },
    markAsRead: markAsReadMock,
    markAllAsRead: markAllAsReadMock,
    fetchAnnouncements: vi.fn(),
    dismissPopup: vi.fn(),
    reset: vi.fn()
  })
}))

vi.mock('pinia', async () => {
  const actual = await vi.importActual<typeof import('pinia')>('pinia')
  return {
    ...actual,
    storeToRefs: (store: any) => ({
      announcements: store.announcements,
      loading: store.loading
    })
  }
})

vi.mock('@/utils/bodyScrollLock', () => ({
  lockBodyScroll: vi.fn(),
  unlockBodyScroll: vi.fn()
}))

vi.mock('@/utils/format', () => ({
  formatRelativeTime: vi.fn(() => '1h ago'),
  formatRelativeWithDateTime: vi.fn(() => '2026-04-11 08:00')
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

describe('AnnouncementBell', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    announcements.value = [
      {
        id: 1,
        title: 'Maintenance',
        content: 'Window notice',
        created_at: '2026-04-11T00:00:00.000Z',
        read_at: null,
        notify_mode: 'inbox'
      }
    ]
    loading.value = false
    currentPopup.value = null
    markAsReadMock.mockResolvedValue(undefined)
    markAllAsReadMock.mockResolvedValue(undefined)
  })

  it('prefers backend detail when mark-all-as-read fails', async () => {
    markAllAsReadMock.mockRejectedValue({
      response: {
        data: {
          detail: 'announcement detail error'
        }
      },
      message: 'generic announcement error'
    })

    const wrapper = mount(AnnouncementBell, {
      global: {
        stubs: {
          Teleport: true,
          Transition: true,
          Icon: true
        }
      }
    })

    await wrapper.get('button[aria-label="announcements.title"]').trigger('click')
    await flushPromises()
    await wrapper.get('.announcement-bell__primary-action--compact').trigger('click')
    await flushPromises()

    expect(markAllAsReadMock).toHaveBeenCalledTimes(1)
    expect(showErrorMock).toHaveBeenCalledWith('announcement detail error')
  })
})
