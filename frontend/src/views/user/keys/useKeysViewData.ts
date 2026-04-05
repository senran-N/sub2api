import { computed, ref, type ComputedRef } from 'vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { keysAPI, usageAPI, userGroupsAPI } from '@/api'
import type { BatchApiKeyUsageStats } from '@/api/usage'
import type { Column } from '@/components/common/types'
import type { ApiKey, Group, PublicSettings } from '@/types'
import {
  buildUserKeyGroupOptions,
  type UserKeyGroupOption
} from './keysForm'

interface KeysViewDataOptions {
  t: (key: string) => string
  showError: (message: string) => void
  fetchPublicSettings: () => Promise<PublicSettings | null>
  publicSettings: ComputedRef<PublicSettings | null>
}

function isAbortError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const { name, code } = error as { name?: string; code?: string }
  return name === 'AbortError' || code === 'ERR_CANCELED'
}

export function useKeysViewData(options: KeysViewDataOptions) {
  const columns = computed<Column[]>(() => [
    { key: 'name', label: options.t('common.name'), sortable: true },
    { key: 'key', label: options.t('keys.apiKey'), sortable: false },
    { key: 'group', label: options.t('keys.group'), sortable: false },
    { key: 'usage', label: options.t('keys.usage'), sortable: false },
    { key: 'rate_limit', label: options.t('keys.rateLimitColumn'), sortable: false },
    { key: 'expires_at', label: options.t('keys.expiresAt'), sortable: true },
    { key: 'status', label: options.t('common.status'), sortable: true },
    { key: 'last_used_at', label: options.t('keys.lastUsedAt'), sortable: true },
    { key: 'created_at', label: options.t('keys.created'), sortable: true },
    { key: 'actions', label: options.t('common.actions'), sortable: false }
  ])

  const apiKeys = ref<ApiKey[]>([])
  const groups = ref<Group[]>([])
  const loading = ref(false)
  const usageStats = ref<Record<string, BatchApiKeyUsageStats>>({})
  const userGroupRates = ref<Record<number, number>>({})

  const pagination = ref({
    page: 1,
    page_size: getPersistedPageSize(),
    total: 0,
    pages: 0
  })

  const filterSearch = ref('')
  const filterStatus = ref('')
  const filterGroupId = ref<string | number>('')

  const groupFilterOptions = computed(() => [
    { value: '', label: options.t('keys.allGroups') },
    { value: 0, label: options.t('keys.noGroup') },
    ...groups.value.map((group) => ({ value: group.id, label: group.name }))
  ])

  const statusFilterOptions = computed(() => [
    { value: '', label: options.t('keys.allStatus') },
    { value: 'active', label: options.t('keys.status.active') },
    { value: 'inactive', label: options.t('keys.status.inactive') },
    { value: 'quota_exhausted', label: options.t('keys.status.quota_exhausted') },
    { value: 'expired', label: options.t('keys.status.expired') }
  ])

  const groupOptions = computed<UserKeyGroupOption[]>(() =>
    buildUserKeyGroupOptions(groups.value, userGroupRates.value)
  )

  let abortController: AbortController | null = null

  async function loadApiKeys() {
    abortController?.abort()
    const controller = new AbortController()
    abortController = controller
    const { signal } = controller
    loading.value = true

    try {
      const filters: { search?: string; status?: string; group_id?: number | string } = {}
      if (filterSearch.value) filters.search = filterSearch.value
      if (filterStatus.value) filters.status = filterStatus.value
      if (filterGroupId.value !== '') filters.group_id = filterGroupId.value

      const response = await keysAPI.list(pagination.value.page, pagination.value.page_size, filters, {
        signal
      })
      if (signal.aborted) return

      apiKeys.value = response.items
      pagination.value.total = response.total
      pagination.value.pages = response.pages

      if (response.items.length === 0) {
        usageStats.value = {}
        return
      }

      try {
        const usageResponse = await usageAPI.getDashboardApiKeysUsage(
          response.items.map((key) => key.id),
          { signal }
        )
        if (signal.aborted) return
        usageStats.value = usageResponse.stats
      } catch (error) {
        if (!isAbortError(error)) {
          console.error('Failed to load usage stats:', error)
        }
      }
    } catch (error) {
      if (isAbortError(error)) {
        return
      }

      options.showError(options.t('keys.failedToLoad'))
    } finally {
      if (abortController === controller) {
        loading.value = false
      }
    }
  }

  async function loadGroups() {
    try {
      groups.value = await userGroupsAPI.getAvailable()
    } catch (error) {
      console.error('Failed to load groups:', error)
    }
  }

  async function loadUserGroupRates() {
    try {
      userGroupRates.value = await userGroupsAPI.getUserGroupRates()
    } catch (error) {
      console.error('Failed to load user group rates:', error)
    }
  }

  async function loadPublicSettings() {
    try {
      await options.fetchPublicSettings()
    } catch (error) {
      console.error('Failed to load public settings:', error)
    }
  }

  function onFilterChange() {
    pagination.value.page = 1
    void loadApiKeys()
  }

  function onGroupFilterChange(value: string | number | boolean | null) {
    filterGroupId.value = value as string | number
    onFilterChange()
  }

  function onStatusFilterChange(value: string | number | boolean | null) {
    filterStatus.value = value as string
    onFilterChange()
  }

  function handlePageChange(page: number) {
    pagination.value.page = page
    void loadApiKeys()
  }

  function handlePageSizeChange(pageSize: number) {
    pagination.value.page_size = pageSize
    pagination.value.page = 1
    void loadApiKeys()
  }

  function dispose() {
    abortController?.abort()
  }

  return {
    columns,
    apiKeys,
    groups,
    loading,
    usageStats,
    userGroupRates,
    pagination,
    filterSearch,
    filterStatus,
    filterGroupId,
    publicSettings: options.publicSettings,
    groupFilterOptions,
    statusFilterOptions,
    groupOptions,
    loadApiKeys,
    loadGroups,
    loadUserGroupRates,
    loadPublicSettings,
    onFilterChange,
    onGroupFilterChange,
    onStatusFilterChange,
    handlePageChange,
    handlePageSizeChange,
    dispose
  }
}
