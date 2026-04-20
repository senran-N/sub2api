import { useIntervalFn } from '@vueuse/core'
import { ref, toRaw, watch, type ComputedRef, type Ref } from 'vue'
import type { Account, WindowStats } from '@/types'
import {
  ACCOUNT_AUTO_REFRESH_INTERVALS,
  ACCOUNT_AUTO_REFRESH_SILENT_WINDOW_MS,
  ACCOUNT_AUTO_REFRESH_STORAGE_KEY,
  buildAccountTodayStatsMap,
  mergeIncrementalAccountRows
} from './accountsList'

interface AccountsPaginationState {
  page: number
  page_size: number
  total: number
  pages: number
}

interface AccountsIncrementalResponse {
  etag?: string | null
  notModified?: boolean
  data?: {
    total?: number
    pages?: number
    items?: Account[]
  } | null
}

interface AccountsViewRefreshOptions {
  accounts: Ref<Account[]>
  loading: Ref<boolean>
  params: Record<string, unknown>
  pagination: AccountsPaginationState
  hiddenColumns: Set<string>
  isAnyModalOpen: ComputedRef<boolean>
  isActionMenuOpen: ComputedRef<boolean>
  loadBase: () => Promise<void>
  reloadBase: () => Promise<void>
  debouncedReloadBase: () => void
  handlePageChangeBase: (page: number) => void
  handlePageSizeChangeBase: (size: number) => void
  fetchTodayStats: (accountIds: number[]) => Promise<{ stats?: Record<string, WindowStats> }>
  fetchAccountsIncrementally: (
    page: number,
    pageSize: number,
    params: Record<string, unknown>,
    options: { etag: string | null }
  ) => Promise<AccountsIncrementalResponse>
  syncAccountRefs: (nextAccount: Account) => void
}

interface AutoRefreshStorageState {
  enabled?: boolean
  interval_seconds?: number
}

export function useAccountsViewRefresh(options: AccountsViewRefreshOptions) {
  const autoRefreshIntervals = ACCOUNT_AUTO_REFRESH_INTERVALS
  const autoRefreshEnabled = ref(false)
  const autoRefreshIntervalSeconds = ref<(typeof autoRefreshIntervals)[number]>(30)
  const autoRefreshCountdown = ref(0)
  const autoRefreshETag = ref<string | null>(null)
  const autoRefreshFetching = ref(false)
  const autoRefreshSilentUntil = ref(0)
  const hasPendingListSync = ref(false)
  const todayStatsByAccountId = ref<Record<string, WindowStats>>({})
  const todayStatsLoading = ref(false)
  const todayStatsError = ref<string | null>(null)
  const pendingTodayStatsRefresh = ref(false)
  const todayStatsReqSeq = ref(0)
  const listRefreshReqSeq = ref(0)
  const usageManualRefreshToken = ref(0)
  const isFirstLoad = ref(true)

  const bumpUsageManualRefreshToken = () => {
    usageManualRefreshToken.value += 1
  }

  const createListRefreshRequest = () => {
    listRefreshReqSeq.value += 1
    return listRefreshReqSeq.value
  }

  const isLatestListRefreshRequest = (requestSeq: number) => requestSeq === listRefreshReqSeq.value

  const loadSavedAutoRefresh = () => {
    try {
      const saved = localStorage.getItem(ACCOUNT_AUTO_REFRESH_STORAGE_KEY)
      if (!saved) {
        return
      }

      const parsed = JSON.parse(saved) as AutoRefreshStorageState
      autoRefreshEnabled.value = parsed.enabled === true

      const interval = Number(parsed.interval_seconds)
      if (autoRefreshIntervals.includes(interval as (typeof autoRefreshIntervals)[number])) {
        autoRefreshIntervalSeconds.value = interval as (typeof autoRefreshIntervals)[number]
      }
    } catch (error) {
      console.error('Failed to load saved auto refresh settings:', error)
    }
  }

  const saveAutoRefreshToStorage = () => {
    try {
      localStorage.setItem(
        ACCOUNT_AUTO_REFRESH_STORAGE_KEY,
        JSON.stringify({
          enabled: autoRefreshEnabled.value,
          interval_seconds: autoRefreshIntervalSeconds.value
        })
      )
    } catch (error) {
      console.error('Failed to save auto refresh settings:', error)
    }
  }

  const refreshTodayStatsBatch = async (ownerListRequestSeq = listRefreshReqSeq.value) => {
    if (!isLatestListRefreshRequest(ownerListRequestSeq)) {
      return
    }

    // Skip when both render targets are hidden: today's metrics card and usage column.
    if (options.hiddenColumns.has('today_stats') && options.hiddenColumns.has('usage')) {
      todayStatsLoading.value = false
      todayStatsError.value = null
      return
    }

    const accountIds = options.accounts.value.map((account) => account.id)
    const requestSeq = ++todayStatsReqSeq.value
    if (accountIds.length === 0) {
      todayStatsByAccountId.value = {}
      todayStatsError.value = null
      todayStatsLoading.value = false
      return
    }

    todayStatsLoading.value = true
    todayStatsError.value = null

    try {
      const result = await options.fetchTodayStats(accountIds)
      if (
        requestSeq !== todayStatsReqSeq.value ||
        !isLatestListRefreshRequest(ownerListRequestSeq)
      ) {
        return
      }

      todayStatsByAccountId.value = buildAccountTodayStatsMap(accountIds, result.stats ?? {})
    } catch (error) {
      if (
        requestSeq !== todayStatsReqSeq.value ||
        !isLatestListRefreshRequest(ownerListRequestSeq)
      ) {
        return
      }

      todayStatsError.value = 'Failed'
      console.error('Failed to load account today stats:', error)
    } finally {
      if (
        requestSeq === todayStatsReqSeq.value &&
        isLatestListRefreshRequest(ownerListRequestSeq)
      ) {
        todayStatsLoading.value = false
      }
    }
  }

  const resetAutoRefreshCache = () => {
    autoRefreshETag.value = null
  }

  const load = async () => {
    const requestParams = options.params as Record<string, unknown>
    const requestSeq = createListRefreshRequest()
    hasPendingListSync.value = false
    resetAutoRefreshCache()
    pendingTodayStatsRefresh.value = false

    const shouldUseLite = isFirstLoad.value
    if (shouldUseLite) {
      requestParams.lite = '1'
    }

    await options.loadBase()

    if (!isLatestListRefreshRequest(requestSeq)) {
      return
    }

    if (shouldUseLite && isFirstLoad.value) {
      isFirstLoad.value = false
      delete requestParams.lite
    }

    await refreshTodayStatsBatch(requestSeq)
  }

  const reload = async () => {
    const requestSeq = createListRefreshRequest()
    hasPendingListSync.value = false
    resetAutoRefreshCache()
    pendingTodayStatsRefresh.value = false
    await options.reloadBase()

    if (!isLatestListRefreshRequest(requestSeq)) {
      return
    }

    await refreshTodayStatsBatch(requestSeq)
    bumpUsageManualRefreshToken()
  }

  const debouncedReload = () => {
    createListRefreshRequest()
    hasPendingListSync.value = false
    resetAutoRefreshCache()
    pendingTodayStatsRefresh.value = true
    options.debouncedReloadBase()
  }

  const handlePageChange = (page: number) => {
    createListRefreshRequest()
    hasPendingListSync.value = false
    resetAutoRefreshCache()
    pendingTodayStatsRefresh.value = true
    options.handlePageChangeBase(page)
  }

  const handlePageSizeChange = (size: number) => {
    createListRefreshRequest()
    hasPendingListSync.value = false
    resetAutoRefreshCache()
    pendingTodayStatsRefresh.value = true
    options.handlePageSizeChangeBase(size)
  }

  watch(options.loading, (isLoading, wasLoading) => {
    if (wasLoading && !isLoading && pendingTodayStatsRefresh.value) {
      pendingTodayStatsRefresh.value = false
      void refreshTodayStatsBatch()
    }
  })

  const enterAutoRefreshSilentWindow = () => {
    autoRefreshSilentUntil.value = Date.now() + ACCOUNT_AUTO_REFRESH_SILENT_WINDOW_MS
    autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
  }

  const inAutoRefreshSilentWindow = () => Date.now() < autoRefreshSilentUntil.value

  const mergeAccountsIncrementally = (nextRows: Account[]) => {
    const result = mergeIncrementalAccountRows(options.accounts.value, nextRows, options.syncAccountRefs)
    if (result.changed) {
      options.accounts.value = result.rows
    }
  }

  const refreshAccountsIncrementally = async () => {
    if (autoRefreshFetching.value) {
      return
    }

    const requestSeq = listRefreshReqSeq.value
    autoRefreshFetching.value = true
    try {
      const result = await options.fetchAccountsIncrementally(
        options.pagination.page,
        options.pagination.page_size,
        toRaw(options.params) as Record<string, unknown>,
        { etag: autoRefreshETag.value }
      )

      if (!isLatestListRefreshRequest(requestSeq)) {
        return
      }

      if (result.etag) {
        autoRefreshETag.value = result.etag
      }

      if (!result.notModified && result.data) {
        options.pagination.total = result.data.total || 0
        options.pagination.pages = result.data.pages || 0
        mergeAccountsIncrementally(result.data.items || [])
        hasPendingListSync.value = false
      }

      await refreshTodayStatsBatch(requestSeq)
    } catch (error) {
      console.error('Auto refresh failed:', error)
    } finally {
      autoRefreshFetching.value = false
    }
  }

  const handleManualRefresh = async () => {
    await load()
    bumpUsageManualRefreshToken()
  }

  const syncPendingListChanges = async () => {
    hasPendingListSync.value = false
    await load()
    bumpUsageManualRefreshToken()
  }

  const { pause: pauseAutoRefresh, resume: resumeAutoRefresh } = useIntervalFn(
    async () => {
      if (!autoRefreshEnabled.value) return
      if (document.hidden) return
      if (options.loading.value || autoRefreshFetching.value) return
      if (options.isAnyModalOpen.value) return
      if (options.isActionMenuOpen.value) return

      if (inAutoRefreshSilentWindow()) {
        autoRefreshCountdown.value = Math.max(
          0,
          Math.ceil((autoRefreshSilentUntil.value - Date.now()) / 1000)
        )
        return
      }

      if (autoRefreshCountdown.value <= 0) {
        autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
        await refreshAccountsIncrementally()
        return
      }

      autoRefreshCountdown.value -= 1
    },
    1000,
    { immediate: false }
  )

  const setAutoRefreshEnabled = (enabled: boolean) => {
    autoRefreshEnabled.value = enabled
    saveAutoRefreshToStorage()

    if (enabled) {
      autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
      resumeAutoRefresh()
      return
    }

    pauseAutoRefresh()
    autoRefreshCountdown.value = 0
  }

  const setAutoRefreshInterval = (seconds: (typeof autoRefreshIntervals)[number]) => {
    autoRefreshIntervalSeconds.value = seconds
    saveAutoRefreshToStorage()

    if (autoRefreshEnabled.value) {
      autoRefreshCountdown.value = seconds
    }
  }

  const initializeAutoRefresh = () => {
    if (typeof window !== 'undefined') {
      loadSavedAutoRefresh()
    }

    if (autoRefreshEnabled.value) {
      autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
      resumeAutoRefresh()
      return
    }

    pauseAutoRefresh()
  }

  const dispose = () => {
    pauseAutoRefresh()
  }

  return {
    autoRefreshIntervals,
    autoRefreshEnabled,
    autoRefreshIntervalSeconds,
    autoRefreshCountdown,
    hasPendingListSync,
    todayStatsByAccountId,
    todayStatsLoading,
    todayStatsError,
    usageManualRefreshToken,
    bumpUsageManualRefreshToken,
    refreshTodayStatsBatch,
    load,
    reload,
    debouncedReload,
    handlePageChange,
    handlePageSizeChange,
    handleManualRefresh,
    syncPendingListChanges,
    setAutoRefreshEnabled,
    setAutoRefreshInterval,
    enterAutoRefreshSilentWindow,
    initializeAutoRefresh,
    dispose
  }
}
