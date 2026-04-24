import { ref, type ComputedRef, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { useTableLoader } from '@/composables/useTableLoader'
import type { AdminGroup, AdminUser, UserAttributeDefinition } from '@/types'
import type { BatchUserUsageStats } from '@/api/admin/dashboard'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import { buildUserListFilters, type UsersFilterState, type UsersPaginationState } from './usersTable'

interface UsersViewDataOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  filters: UsersFilterState
  activeAttributeFilters: Record<number, string>
  searchQuery: Ref<string>
  pagination: UsersPaginationState & { total: number }
  hasVisibleUsageColumn: ComputedRef<boolean>
  hasVisibleSubscriptionsColumn: ComputedRef<boolean>
  hasVisibleAttributeColumns: ComputedRef<boolean>
  isSecondaryDataRequestCurrent: (expectedSeq?: number) => boolean
  setCurrentUserIds: (userIds: number[]) => void
  resetSecondaryDataState: () => void
  scheduleUsersSecondaryDataLoad: (signal?: AbortSignal) => void
}

export function useUsersViewData(options: UsersViewDataOptions) {
  const usageStats = ref<Record<string, BatchUserUsageStats>>({})
  const attributeDefinitions = ref<UserAttributeDefinition[]>([])
  const userAttributeValues = ref<Record<number, Record<number, string>>>({})
  const allGroups = ref<AdminGroup[]>([])
  let activeLoadSignal: AbortSignal | undefined

  const resetSecondaryData = () => {
    usageStats.value = {}
    userAttributeValues.value = {}
  }

  async function loadAllGroups() {
    if (allGroups.value.length > 0) {
      return
    }
    try {
      allGroups.value = await adminAPI.groups.getAll()
    } catch (error) {
      console.error('Failed to load groups:', error)
    }
  }

  async function loadAttributeDefinitions() {
    try {
      attributeDefinitions.value = await adminAPI.userAttributes.listEnabledDefinitions()
    } catch (error) {
      console.error('Failed to load attribute definitions:', error)
    }
  }

  async function loadUsersSecondaryData(
    userIds: number[],
    signal?: AbortSignal,
    expectedSeq?: number
  ) {
    if (userIds.length === 0) {
      return
    }

    const tasks: Promise<void>[] = []

    if (options.hasVisibleUsageColumn.value) {
      tasks.push(
        (async () => {
          try {
            const usageResponse = await adminAPI.dashboard.getBatchUsersUsage(userIds)
            if (signal?.aborted) return
            if (!options.isSecondaryDataRequestCurrent(expectedSeq)) return
            usageStats.value = usageResponse.stats
          } catch (error) {
            if (signal?.aborted) return
            console.error('Failed to load usage stats:', error)
          }
        })()
      )
    }

    if (attributeDefinitions.value.length > 0 && options.hasVisibleAttributeColumns.value) {
      tasks.push(
        (async () => {
          try {
            const attrResponse = await adminAPI.userAttributes.getBatchUserAttributes(userIds)
            if (signal?.aborted) return
            if (!options.isSecondaryDataRequestCurrent(expectedSeq)) return
            userAttributeValues.value = attrResponse.attributes
          } catch (error) {
            if (signal?.aborted) return
            console.error('Failed to load user attribute values:', error)
          }
        })()
      )
    }

    if (tasks.length > 0) {
      await Promise.allSettled(tasks)
    }
  }
  const {
    items: users,
    loading,
    load: loadUsers,
    dispose
  } = useTableLoader<AdminUser, Record<string, never>>({
    fetchFn: (page, pageSize, _params, requestOptions) => {
      activeLoadSignal = requestOptions?.signal
      return adminAPI.users.list(
        page,
        pageSize,
        buildUserListFilters(
          options.filters,
          options.searchQuery.value,
          options.activeAttributeFilters,
          options.hasVisibleSubscriptionsColumn.value
        ),
        requestOptions
      )
    },
    pagination: options.pagination,
    onError: (error) => {
      const message = resolveRequestErrorMessage(error, options.t('admin.users.failedToLoad'))
      options.showError(message)
    },
    onLoaded: (response) => {
      options.setCurrentUserIds(response.items.map((user) => user.id))
      options.resetSecondaryDataState()
      resetSecondaryData()
      if (response.items.length > 0) {
        options.scheduleUsersSecondaryDataLoad(activeLoadSignal)
      }
    }
  })


  return {
    usageStats,
    attributeDefinitions,
    userAttributeValues,
    users,
    loading,
    allGroups,
    loadAllGroups,
    loadAttributeDefinitions,
    loadUsersSecondaryData,
    loadUsers,
    resetSecondaryData,
    dispose
  }
}
