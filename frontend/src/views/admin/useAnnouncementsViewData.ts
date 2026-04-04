import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import type { Announcement } from '@/types'
import { buildAnnouncementListFilters, createDefaultAnnouncementFilters } from './announcementsForm'

interface AnnouncementsViewDataOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
}

export function useAnnouncementsViewData(options: AnnouncementsViewDataOptions) {
  const announcements = ref<Announcement[]>([])
  const loading = ref(false)
  const filters = reactive(createDefaultAnnouncementFilters())
  const searchQuery = ref('')
  const pagination = reactive({
    page: 1,
    page_size: getPersistedPageSize(),
    total: 0,
    pages: 0
  })

  let currentController: AbortController | null = null
  let searchDebounceTimer: number | null = null

  const loadAnnouncements = async () => {
    currentController?.abort()

    const requestController = new AbortController()
    currentController = requestController
    loading.value = true

    try {
      const response = await adminAPI.announcements.list(
        pagination.page,
        pagination.page_size,
        buildAnnouncementListFilters(filters, searchQuery.value),
        {
          signal: requestController.signal
        }
      )

      if (requestController.signal.aborted || currentController !== requestController) {
        return
      }

      announcements.value = response.items
      pagination.total = response.total
      pagination.pages = response.pages
      pagination.page = response.page
      pagination.page_size = response.page_size
    } catch (error: any) {
      if (
        requestController.signal.aborted ||
        error?.name === 'AbortError' ||
        error?.code === 'ERR_CANCELED'
      ) {
        return
      }
      console.error('Error loading announcements:', error)
      options.showError(error.response?.data?.detail || options.t('admin.announcements.failedToLoad'))
    } finally {
      if (currentController === requestController) {
        loading.value = false
        currentController = null
      }
    }
  }

  const handlePageChange = (page: number) => {
    pagination.page = page
    void loadAnnouncements()
  }

  const handlePageSizeChange = (pageSize: number) => {
    pagination.page_size = pageSize
    pagination.page = 1
    void loadAnnouncements()
  }

  const handleStatusChange = () => {
    pagination.page = 1
    void loadAnnouncements()
  }

  const handleSearch = () => {
    if (searchDebounceTimer) {
      window.clearTimeout(searchDebounceTimer)
    }
    searchDebounceTimer = window.setTimeout(() => {
      pagination.page = 1
      void loadAnnouncements()
    }, 300)
  }

  const dispose = () => {
    currentController?.abort()
    if (searchDebounceTimer) {
      window.clearTimeout(searchDebounceTimer)
    }
  }

  return {
    announcements,
    loading,
    filters,
    searchQuery,
    pagination,
    loadAnnouncements,
    handlePageChange,
    handlePageSizeChange,
    handleStatusChange,
    handleSearch,
    dispose
  }
}
