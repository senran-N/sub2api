import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import type { Group, UserSubscription } from '@/types'
import {
  buildSubscriptionListFilters,
  createDefaultSubscriptionFilters,
  type SubscriptionFiltersState,
  type SubscriptionSortState
} from './subscriptionForm'

interface SubscriptionsViewDataOptions {
  showLoadError: () => void
}

export function useSubscriptionsViewData(options: SubscriptionsViewDataOptions) {
  const subscriptions = ref<UserSubscription[]>([])
  const groups = ref<Group[]>([])
  const loading = ref(false)
  const filters = reactive<SubscriptionFiltersState>(createDefaultSubscriptionFilters())
  const sortState = reactive<SubscriptionSortState>({
    sort_by: 'created_at',
    sort_order: 'desc'
  })
  const pagination = reactive({
    page: 1,
    page_size: getPersistedPageSize(),
    total: 0,
    pages: 0
  })

  let abortController: AbortController | null = null

  const loadSubscriptions = async () => {
    abortController?.abort()

    const requestController = new AbortController()
    abortController = requestController
    loading.value = true

    try {
      const response = await adminAPI.subscriptions.list(
        pagination.page,
        pagination.page_size,
        buildSubscriptionListFilters(filters, sortState),
        {
          signal: requestController.signal
        }
      )

      if (requestController.signal.aborted || abortController !== requestController) {
        return
      }

      subscriptions.value = response.items
      pagination.total = response.total
      pagination.pages = response.pages
    } catch (error: any) {
      if (
        requestController.signal.aborted ||
        error?.name === 'AbortError' ||
        error?.code === 'ERR_CANCELED'
      ) {
        return
      }

      options.showLoadError()
      console.error('Error loading subscriptions:', error)
    } finally {
      if (abortController === requestController) {
        loading.value = false
        abortController = null
      }
    }
  }

  const loadGroups = async () => {
    try {
      groups.value = await adminAPI.groups.getAll()
    } catch (error) {
      console.error('Error loading groups:', error)
    }
  }

  const applyFilters = () => {
    pagination.page = 1
    void loadSubscriptions()
  }

  const handlePageChange = (page: number) => {
    pagination.page = page
    void loadSubscriptions()
  }

  const handlePageSizeChange = (pageSize: number) => {
    pagination.page_size = pageSize
    pagination.page = 1
    void loadSubscriptions()
  }

  const handleSort = (key: string, order: 'asc' | 'desc') => {
    sortState.sort_by = key
    sortState.sort_order = order
    pagination.page = 1
    void loadSubscriptions()
  }

  const loadInitialData = () => {
    void loadSubscriptions()
    void loadGroups()
  }

  const dispose = () => {
    abortController?.abort()
  }

  return {
    subscriptions,
    groups,
    loading,
    filters,
    sortState,
    pagination,
    loadSubscriptions,
    loadGroups,
    applyFilters,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
    loadInitialData,
    dispose
  }
}
