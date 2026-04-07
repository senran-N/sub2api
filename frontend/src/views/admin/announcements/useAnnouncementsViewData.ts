import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { useTableLoader } from '@/composables/useTableLoader'
import type { Announcement } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import { buildAnnouncementListFilters, createDefaultAnnouncementFilters } from './announcementsForm'

interface AnnouncementsViewDataOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
}

export function useAnnouncementsViewData(options: AnnouncementsViewDataOptions) {
  const filters = reactive(createDefaultAnnouncementFilters())
  const searchQuery = ref('')
  const {
    items: announcements,
    loading,
    pagination,
    load: loadAnnouncements,
    debouncedReload,
    handlePageChange,
    handlePageSizeChange,
    dispose
  } = useTableLoader<Announcement, Record<string, never>>({
    fetchFn: (page, pageSize, _params, requestOptions) =>
      adminAPI.announcements.list(
        page,
        pageSize,
        buildAnnouncementListFilters(filters, searchQuery.value),
        requestOptions
      ),
    onError: (error) => {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.announcements.failedToLoad'))
      )
    },
    syncPaginationFromResponse: true,
    clampPageChange: false
  })

  const handleStatusChange = () => {
    void loadAnnouncements()
  }

  const handleSearch = () => {
    void debouncedReload()
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
