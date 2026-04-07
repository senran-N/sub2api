import {
  computed,
  getCurrentInstance,
  onUnmounted,
  reactive,
  ref,
  type ComputedRef,
  type Ref
} from 'vue'
import type { UserAttributeDefinition } from '@/types'
import {
  DEFAULT_USER_HIDDEN_COLUMNS,
  applyUsersPageChange,
  applyUsersPageSizeChange,
  createDefaultUsersFilters,
  getUserColumnToggleEffects,
  toggleBuiltInUserFilter,
  toggleUserAttributeFilter,
  type BuiltInUserFilterKey,
  type UsersFilterState,
  type UsersPaginationState
} from './usersTable'

export const USER_HIDDEN_COLUMNS_KEY = 'user-hidden-columns'
export const USER_FILTER_VALUES_KEY = 'user-filter-values'
export const USER_VISIBLE_FILTERS_KEY = 'user-visible-filters'

interface UsersViewStateOptions {
  attributeDefinitions: Ref<UserAttributeDefinition[]>
  initialPageSize: number
  loadUsers: () => void | Promise<void>
  loadGroups: () => void | Promise<void>
  loadSecondaryData: (
    userIds: number[],
    signal?: AbortSignal,
    expectedSeq?: number
  ) => void | Promise<void>
  resetSecondaryData: () => void
}

interface PersistedFilterValues {
  role?: UsersFilterState['role']
  status?: UsersFilterState['status']
  group?: string
  attributes?: Record<number, string>
}

export function useUsersViewState(options: UsersViewStateOptions) {
  const hiddenColumns = reactive<Set<string>>(new Set())
  const filters = reactive(createDefaultUsersFilters())
  const activeAttributeFilters = reactive<Record<number, string>>({})
  const visibleFilters = reactive<Set<string>>(new Set())
  const searchQuery = ref('')
  const pagination = reactive<UsersPaginationState & { total: number }>({
    page: 1,
    page_size: options.initialPageSize,
    total: 0,
    pages: 0
  })

  const hasVisibleUsageColumn = computed(() => !hiddenColumns.has('usage'))
  const hasVisibleSubscriptionsColumn = computed(() => !hiddenColumns.has('subscriptions'))
  const hasVisibleGroupsColumn = computed(() => !hiddenColumns.has('groups'))
  const hasVisibleAttributeColumns = computed(() =>
    options.attributeDefinitions.value.some(
      (definition) => definition.enabled && !hiddenColumns.has(`attr_${definition.id}`)
    )
  )

  const currentUserIds = ref<number[]>([])
  let searchTimeout: ReturnType<typeof setTimeout> | null = null
  let secondaryDataTimeout: ReturnType<typeof setTimeout> | null = null
  let secondaryDataSeq = 0

  const clearSet = (target: Set<string>) => {
    Array.from(target).forEach((value) => target.delete(value))
  }

  const loadSavedColumns = () => {
    clearSet(hiddenColumns)

    try {
      const saved = localStorage.getItem(USER_HIDDEN_COLUMNS_KEY)
      if (saved) {
        const parsed = JSON.parse(saved) as string[]
        parsed.forEach((key) => hiddenColumns.add(key))
        return
      }
    } catch (error) {
      console.error('Failed to load saved columns:', error)
    }

    DEFAULT_USER_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
  }

  const saveColumnsToStorage = () => {
    try {
      localStorage.setItem(USER_HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
    } catch (error) {
      console.error('Failed to save columns:', error)
    }
  }

  const resetFiltersState = () => {
    Object.assign(filters, createDefaultUsersFilters())
    Object.keys(activeAttributeFilters).forEach((key) => {
      delete activeAttributeFilters[Number(key)]
    })
    clearSet(visibleFilters)
  }

  const loadSavedFilters = () => {
    resetFiltersState()

    try {
      const savedVisible = localStorage.getItem(USER_VISIBLE_FILTERS_KEY)
      if (savedVisible) {
        const parsed = JSON.parse(savedVisible) as string[]
        parsed.forEach((key) => visibleFilters.add(key))
      }

      const savedValues = localStorage.getItem(USER_FILTER_VALUES_KEY)
      if (!savedValues) {
        return
      }

      const parsed = JSON.parse(savedValues) as PersistedFilterValues
      if (parsed.role) {
        filters.role = parsed.role
      }
      if (parsed.status) {
        filters.status = parsed.status
      }
      if (parsed.group) {
        filters.group = parsed.group
      }
      if (parsed.attributes) {
        Object.assign(activeAttributeFilters, parsed.attributes)
      }
    } catch (error) {
      console.error('Failed to load saved filters:', error)
      resetFiltersState()
    }
  }

  const saveFiltersToStorage = () => {
    try {
      localStorage.setItem(USER_VISIBLE_FILTERS_KEY, JSON.stringify([...visibleFilters]))
      localStorage.setItem(
        USER_FILTER_VALUES_KEY,
        JSON.stringify({
          role: filters.role,
          status: filters.status,
          group: filters.group,
          attributes: activeAttributeFilters
        })
      )
    } catch (error) {
      console.error('Failed to save filters:', error)
    }
  }

  const initializePersistedState = () => {
    loadSavedFilters()
    loadSavedColumns()
    if (hasVisibleGroupsColumn.value || visibleFilters.has('group')) {
      void options.loadGroups()
    }
  }

  const isColumnVisible = (key: string) => !hiddenColumns.has(key)

  const toggleColumn = (key: string) => {
    const wasHidden = hiddenColumns.has(key)
    if (wasHidden) {
      hiddenColumns.delete(key)
    } else {
      hiddenColumns.add(key)
    }

    saveColumnsToStorage()

    const effects = getUserColumnToggleEffects(key, wasHidden)
    if (effects.refreshSecondaryData) {
      refreshCurrentPageSecondaryData()
    }
    if (effects.reloadUsers) {
      void options.loadUsers()
    }
    if (effects.loadGroups) {
      void options.loadGroups()
    }
  }

  const handleSearch = () => {
    if (searchTimeout) {
      clearTimeout(searchTimeout)
    }

    searchTimeout = setTimeout(() => {
      pagination.page = 1
      void options.loadUsers()
    }, 300)
  }

  const handlePageChange = (page: number) => {
    applyUsersPageChange(pagination, page)
    void options.loadUsers()
  }

  const handlePageSizeChange = (pageSize: number) => {
    applyUsersPageSizeChange(pagination, pageSize)
    void options.loadUsers()
  }

  const toggleBuiltInFilter = (key: BuiltInUserFilterKey) => {
    const { shouldLoadGroups } = toggleBuiltInUserFilter(visibleFilters, filters, key)
    if (shouldLoadGroups) {
      void options.loadGroups()
    }
    saveFiltersToStorage()
    pagination.page = 1
    void options.loadUsers()
  }

  const toggleAttributeFilter = (attributeId: number) => {
    toggleUserAttributeFilter(visibleFilters, activeAttributeFilters, attributeId)
    saveFiltersToStorage()
    pagination.page = 1
    void options.loadUsers()
  }

  const updateAttributeFilter = (attributeId: number, value: string) => {
    activeAttributeFilters[attributeId] = value
  }

  const applyFilter = () => {
    saveFiltersToStorage()
    void options.loadUsers()
  }

  const setCurrentUserIds = (userIds: number[]) => {
    currentUserIds.value = [...userIds]
  }

  const resetSecondaryDataState = () => {
    if (secondaryDataTimeout) {
      clearTimeout(secondaryDataTimeout)
      secondaryDataTimeout = null
    }

    secondaryDataSeq += 1
    options.resetSecondaryData()
  }

  const scheduleUsersSecondaryDataLoad = (signal?: AbortSignal) => {
    if (secondaryDataTimeout) {
      clearTimeout(secondaryDataTimeout)
      secondaryDataTimeout = null
    }
    if (currentUserIds.value.length === 0) {
      return
    }

    const requestSeq = secondaryDataSeq
    secondaryDataTimeout = setTimeout(() => {
      secondaryDataTimeout = null
      if (signal?.aborted || requestSeq !== secondaryDataSeq) {
        return
      }
      void options.loadSecondaryData(currentUserIds.value, signal, requestSeq)
    }, 50)
  }

  const refreshCurrentPageSecondaryData = () => {
    if (currentUserIds.value.length === 0) {
      return
    }

    secondaryDataSeq += 1
    const requestSeq = secondaryDataSeq
    void options.loadSecondaryData(currentUserIds.value, undefined, requestSeq)
  }

  const isSecondaryDataRequestCurrent = (expectedSeq?: number) =>
    typeof expectedSeq !== 'number' || expectedSeq === secondaryDataSeq

  const dispose = () => {
    if (searchTimeout) {
      clearTimeout(searchTimeout)
      searchTimeout = null
    }
    if (secondaryDataTimeout) {
      clearTimeout(secondaryDataTimeout)
      secondaryDataTimeout = null
    }
  }

  if (getCurrentInstance()) {
    onUnmounted(() => {
      dispose()
    })
  }

  return {
    hiddenColumns,
    filters,
    activeAttributeFilters,
    visibleFilters,
    searchQuery,
    pagination,
    hasVisibleUsageColumn,
    hasVisibleSubscriptionsColumn,
    hasVisibleGroupsColumn,
    hasVisibleAttributeColumns,
    initializePersistedState,
    isColumnVisible,
    toggleColumn,
    handleSearch,
    handlePageChange,
    handlePageSizeChange,
    toggleBuiltInFilter,
    toggleAttributeFilter,
    updateAttributeFilter,
    applyFilter,
    setCurrentUserIds,
    resetSecondaryDataState,
    scheduleUsersSecondaryDataLoad,
    refreshCurrentPageSecondaryData,
    isSecondaryDataRequestCurrent,
    dispose
  }
}

export type UsersViewState = ReturnType<typeof useUsersViewState>
export type UsersViewColumnVisibility = Pick<
  UsersViewState,
  | 'hiddenColumns'
  | 'hasVisibleUsageColumn'
  | 'hasVisibleSubscriptionsColumn'
  | 'hasVisibleGroupsColumn'
  | 'hasVisibleAttributeColumns'
>
export type UsersViewComputedFlag = ComputedRef<boolean>
