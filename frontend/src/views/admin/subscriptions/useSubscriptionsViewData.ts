import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { useTableLoader } from '@/composables/useTableLoader'
import type { Group, UserSubscription } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  buildSubscriptionListFilters,
  createDefaultSubscriptionFilters,
  type SubscriptionFiltersState,
  type SubscriptionSortState
} from './subscriptionForm'

interface SubscriptionsViewDataOptions {
  showLoadError: (message: string) => void
  t: (key: string, params?: Record<string, unknown>) => string
}

export function useSubscriptionsViewData(options: SubscriptionsViewDataOptions) {
  const groups = ref<Group[]>([])
  const filters = reactive<SubscriptionFiltersState>(createDefaultSubscriptionFilters())
  const sortState = reactive<SubscriptionSortState>({
    sort_by: 'created_at',
    sort_order: 'desc'
  })
  const {
    items: subscriptions,
    loading,
    pagination,
    load: loadSubscriptions,
    reload,
    handlePageChange,
    handlePageSizeChange,
    dispose
  } = useTableLoader<UserSubscription, Record<string, never>>({
    fetchFn: (page, pageSize, _params, requestOptions) =>
      adminAPI.subscriptions.list(
        page,
        pageSize,
        buildSubscriptionListFilters(filters, sortState),
        requestOptions
      ),
    onError: (error) => {
      options.showLoadError(
        resolveRequestErrorMessage(error, options.t('admin.subscriptions.failedToLoad'))
      )
    },
    clampPageChange: false
  })

  const loadGroups = async () => {
    try {
      groups.value = await adminAPI.groups.getAll()
    } catch (error) {
      console.error('Error loading groups:', error)
    }
  }

  const applyFilters = () => {
    void reload()
  }

  const handleSort = (key: string, order: 'asc' | 'desc') => {
    sortState.sort_by = key
    sortState.sort_order = order
    void reload()
  }

  const loadInitialData = () => {
    void loadSubscriptions()
    void loadGroups()
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
