import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import {
  applyGroupPageChange,
  applyGroupPageReset,
  applyGroupPageSizeChange,
  buildGroupListFilters,
  buildGroupSortOrderUpdates,
  mapGroupCapacitySummary,
  mapGroupUsageSummary,
  sortGroupsBySortOrder,
  type GroupCapacitySnapshot,
  type GroupFiltersState,
  type GroupPaginationState
} from './groupsTable'
import type { AdminGroup } from '@/types'

interface GroupsViewDataOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showSuccess: (message: string) => void
}

export function useGroupsViewData(options: GroupsViewDataOptions) {
  const groups = ref<AdminGroup[]>([])
  const loading = ref(false)
  const usageLoading = ref(false)
  const usageMap = ref<Map<number, { today_cost: number; total_cost: number }>>(new Map())
  const capacityMap = ref<Map<number, GroupCapacitySnapshot>>(new Map())
  const searchQuery = ref('')
  const filters = reactive<GroupFiltersState>({
    platform: '',
    status: '',
    is_exclusive: ''
  })
  const pagination = reactive<GroupPaginationState & { total: number; pages: number }>({
    page: 1,
    page_size: getPersistedPageSize(),
    total: 0,
    pages: 0
  })
  const showSortModal = ref(false)
  const sortSubmitting = ref(false)
  const sortableGroups = ref<AdminGroup[]>([])

  let abortController: AbortController | null = null
  let searchTimeout: ReturnType<typeof setTimeout> | null = null

  const loadUsageSummary = async () => {
    usageLoading.value = true
    try {
      const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
      usageMap.value = mapGroupUsageSummary(await adminAPI.groups.getUsageSummary(timezone))
    } catch (error) {
      console.error('Error loading group usage summary:', error)
    } finally {
      usageLoading.value = false
    }
  }

  const loadCapacitySummary = async () => {
    try {
      capacityMap.value = mapGroupCapacitySummary(await adminAPI.groups.getCapacitySummary())
    } catch (error) {
      console.error('Error loading group capacity summary:', error)
    }
  }

  const loadGroups = async () => {
    if (abortController) {
      abortController.abort()
    }

    const currentController = new AbortController()
    abortController = currentController
    loading.value = true

    try {
      const response = await adminAPI.groups.list(
        pagination.page,
        pagination.page_size,
        buildGroupListFilters(filters, searchQuery.value),
        { signal: currentController.signal }
      )
      if (currentController.signal.aborted) {
        return
      }

      groups.value = response.items
      pagination.total = response.total
      pagination.pages = response.pages
      void loadUsageSummary()
      void loadCapacitySummary()
    } catch (error) {
      if (
        currentController.signal.aborted ||
        (typeof error === 'object' &&
          error !== null &&
          ('name' in error || 'code' in error) &&
          ((error as { name?: string }).name === 'AbortError' ||
            (error as { code?: string }).code === 'ERR_CANCELED'))
      ) {
        return
      }

      options.showError(options.t('admin.groups.failedToLoad'))
      console.error('Error loading groups:', error)
    } finally {
      if (abortController === currentController && !currentController.signal.aborted) {
        loading.value = false
      }
    }
  }

  const handleSearch = () => {
    if (searchTimeout) {
      clearTimeout(searchTimeout)
    }
    searchTimeout = setTimeout(() => {
      applyGroupPageReset(pagination)
      void loadGroups()
    }, 300)
  }

  const handlePageChange = (page: number) => {
    applyGroupPageChange(pagination, page)
    void loadGroups()
  }

  const handlePageSizeChange = (pageSize: number) => {
    applyGroupPageSizeChange(pagination, pageSize)
    void loadGroups()
  }

  const openSortModal = async () => {
    try {
      sortableGroups.value = sortGroupsBySortOrder(await adminAPI.groups.getAll())
      showSortModal.value = true
    } catch (error) {
      options.showError(options.t('admin.groups.failedToLoad'))
      console.error('Error loading groups for sorting:', error)
    }
  }

  const closeSortModal = () => {
    showSortModal.value = false
    sortableGroups.value = []
  }

  const saveSortOrder = async () => {
    sortSubmitting.value = true
    try {
      await adminAPI.groups.updateSortOrder(buildGroupSortOrderUpdates(sortableGroups.value))
      options.showSuccess(options.t('admin.groups.sortOrderUpdated'))
      closeSortModal()
      await loadGroups()
    } catch (error) {
      const detail: string =
        typeof error === 'object' &&
        error !== null &&
        'response' in error &&
        typeof (error as { response?: { data?: { detail?: string } } }).response?.data?.detail ===
          'string'
          ? (error as { response?: { data?: { detail?: string } } }).response!.data!.detail ??
            options.t('admin.groups.failedToUpdateSortOrder')
          : options.t('admin.groups.failedToUpdateSortOrder')
      options.showError(detail)
      console.error('Error updating sort order:', error)
    } finally {
      sortSubmitting.value = false
    }
  }

  const dispose = () => {
    if (abortController) {
      abortController.abort()
      abortController = null
    }
    if (searchTimeout) {
      clearTimeout(searchTimeout)
      searchTimeout = null
    }
  }

  return {
    groups,
    loading,
    usageMap,
    usageLoading,
    capacityMap,
    searchQuery,
    filters,
    pagination,
    showSortModal,
    sortSubmitting,
    sortableGroups,
    loadGroups,
    loadUsageSummary,
    loadCapacitySummary,
    handleSearch,
    handlePageChange,
    handlePageSizeChange,
    openSortModal,
    closeSortModal,
    saveSortOrder,
    dispose
  }
}
