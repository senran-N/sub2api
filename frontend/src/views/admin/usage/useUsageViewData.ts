import { reactive, ref, watch, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { adminUsageAPI, type AdminUsageQueryParams, type AdminUsageStatsResponse } from '@/api/admin/usage'
import { useTableLoader } from '@/composables/useTableLoader'
import { formatReasoningEffort } from '@/utils/format'
import { isAbortError } from '@/utils/requestError'
import { requestTypeToLegacyStream } from '@/utils/usageRequestType'
import type {
  AdminUsageLog,
  EndpointStat,
  GroupStat,
  ModelStat,
  TrendDataPoint
} from '@/types'
import {
  resetUsagePaginationPage,
  type UsageGranularity,
  type UsagePaginationState
} from './usageViewState'
import { getUsageRequestTypeLabel, type UsageLabelTranslator } from '@/utils/usagePresentation'

type SaveAsFunction = typeof import('file-saver').saveAs

export type ModelDistributionSource = 'requested' | 'upstream' | 'mapping'

export interface UsageViewPaginationState extends UsagePaginationState {
  total: number
}

export interface UsageViewDataOptions {
  filters: Ref<AdminUsageQueryParams>
  startDate: Ref<string>
  endDate: Ref<string>
  granularity: Ref<UsageGranularity>
  modelDistributionSource: Ref<ModelDistributionSource>
  pagination: UsageViewPaginationState
  t: UsageLabelTranslator
  showSuccess: (message: string) => void
  showError: (message: string) => void
}

interface ResolvedUsageDateRange {
  start_date: string
  end_date: string
}

type UsageScopeParams = Pick<
  AdminUsageQueryParams,
  | 'user_id'
  | 'api_key_id'
  | 'account_id'
  | 'group_id'
  | 'model'
  | 'request_type'
  | 'stream'
  | 'billing_type'
> & {
  start_date: string
  end_date: string
}

let saveAsPromise: Promise<SaveAsFunction> | null = null

function getSaveAs(): Promise<SaveAsFunction> {
  if (!saveAsPromise) {
    saveAsPromise = import('file-saver').then(({ saveAs }) => saveAs)
  }

  return saveAsPromise
}

function buildUsageExportHeaders(t: UsageLabelTranslator): string[] {
  return [
    t('usage.time'),
    t('admin.usage.user'),
    t('usage.apiKeyFilter'),
    t('admin.usage.account'),
    t('usage.model'),
    t('usage.upstreamModel'),
    t('usage.reasoningEffort'),
    t('admin.usage.group'),
    t('usage.inboundEndpoint'),
    t('usage.upstreamEndpoint'),
    t('usage.type'),
    t('admin.usage.inputTokens'),
    t('admin.usage.outputTokens'),
    t('admin.usage.cacheReadTokens'),
    t('admin.usage.cacheCreationTokens'),
    t('admin.usage.inputCost'),
    t('admin.usage.outputCost'),
    t('admin.usage.cacheReadCost'),
    t('admin.usage.cacheCreationCost'),
    t('usage.rate'),
    t('usage.accountMultiplier'),
    t('usage.original'),
    t('usage.userBilled'),
    t('usage.accountBilled'),
    t('usage.firstToken'),
    t('usage.duration'),
    t('admin.usage.requestId'),
    t('usage.userAgent'),
    t('admin.usage.ipAddress')
  ]
}

function buildUsageExportRows(logs: AdminUsageLog[], t: UsageLabelTranslator): Array<Array<string | number>> {
  return logs.map((log) => [
    log.created_at,
    log.user?.email || '',
    log.api_key?.name || '',
    log.account?.name || '',
    log.model,
    log.upstream_model || '',
    formatReasoningEffort(log.reasoning_effort),
    log.group?.name || '',
    log.inbound_endpoint || '',
    log.upstream_endpoint || '',
    getUsageRequestTypeLabel(log, t),
    log.input_tokens,
    log.output_tokens,
    log.cache_read_tokens,
    log.cache_creation_tokens,
    log.input_cost?.toFixed(6) || '0.000000',
    log.output_cost?.toFixed(6) || '0.000000',
    log.cache_read_cost?.toFixed(6) || '0.000000',
    log.cache_creation_cost?.toFixed(6) || '0.000000',
    log.rate_multiplier?.toFixed(2) || '1.00',
    (log.account_rate_multiplier ?? 1).toFixed(2),
    log.total_cost?.toFixed(6) || '0.000000',
    log.actual_cost?.toFixed(6) || '0.000000',
    (log.total_cost * (log.account_rate_multiplier ?? 1)).toFixed(6),
    log.first_token_ms ?? '',
    log.duration_ms,
    log.request_id || '',
    log.user_agent || '',
    log.ip_address || ''
  ])
}

export function useUsageViewData(options: UsageViewDataOptions) {
  const usageStats = ref<AdminUsageStatsResponse | null>(null)
  const usageLogs = ref<AdminUsageLog[]>([])
  const loading = ref(false)
  const exporting = ref(false)
  const trendData = ref<TrendDataPoint[]>([])
  const requestedModelStats = ref<ModelStat[]>([])
  const upstreamModelStats = ref<ModelStat[]>([])
  const mappingModelStats = ref<ModelStat[]>([])
  const groupStats = ref<GroupStat[]>([])
  const chartsLoading = ref(false)
  const modelStatsLoading = ref(false)
  const inboundEndpointStats = ref<EndpointStat[]>([])
  const upstreamEndpointStats = ref<EndpointStat[]>([])
  const endpointPathStats = ref<EndpointStat[]>([])
  const endpointStatsLoading = ref(false)
  const exportProgress = reactive({
    show: false,
    progress: 0,
    current: 0,
    total: 0,
    estimatedTime: ''
  })
  const loadedModelSources = reactive<Record<ModelDistributionSource, boolean>>({
    requested: false,
    upstream: false,
    mapping: false
  })

  let exportAbortController: AbortController | null = null
  let initialChartTimer: number | null = null
  let chartRequestSequence = 0
  let statsRequestSequence = 0
  let modelStatsRequestSequence = 0

  const resolveUsageDateRange = (): ResolvedUsageDateRange => ({
    start_date: options.filters.value.start_date || options.startDate.value,
    end_date: options.filters.value.end_date || options.endDate.value
  })

  const buildUsageFilters = (): AdminUsageQueryParams => {
    const requestType = options.filters.value.request_type
    const legacyStream = requestType
      ? requestTypeToLegacyStream(requestType)
      : options.filters.value.stream
    const range = resolveUsageDateRange()

    return {
      ...options.filters.value,
      start_date: range.start_date,
      end_date: range.end_date,
      stream: legacyStream === null ? undefined : legacyStream
    }
  }

  const buildUsageScopeParams = (): UsageScopeParams => {
    const usageFilters = buildUsageFilters()

    return {
      start_date: usageFilters.start_date!,
      end_date: usageFilters.end_date!,
      user_id: usageFilters.user_id,
      api_key_id: usageFilters.api_key_id,
      account_id: usageFilters.account_id,
      group_id: usageFilters.group_id,
      model: usageFilters.model,
      request_type: usageFilters.request_type,
      stream: usageFilters.stream,
      billing_type: usageFilters.billing_type
    }
  }

  const setModelStats = (source: ModelDistributionSource, models: ModelStat[]) => {
    if (source === 'requested') {
      requestedModelStats.value = models
      return
    }
    if (source === 'upstream') {
      upstreamModelStats.value = models
      return
    }

    mappingModelStats.value = models
  }

  const resetModelStatsCache = () => {
    requestedModelStats.value = []
    upstreamModelStats.value = []
    mappingModelStats.value = []
    loadedModelSources.requested = false
    loadedModelSources.upstream = false
    loadedModelSources.mapping = false
  }

  const {
    dispose: disposeUsageLogs,
    handlePageChange: handleUsageLogPageChange,
    handlePageSizeChange: handleUsageLogPageSizeChange,
    load: loadLogs,
    loading: logsLoading
  } = useTableLoader<AdminUsageLog, Record<string, never>>({
    pagination: options.pagination,
    clampPageChange: false,
    fetchFn: (page, pageSize, _params, fetchOptions) =>
      adminAPI.usage.list(
        {
          page,
          page_size: pageSize,
          exact_total: false,
          ...buildUsageFilters()
        },
        fetchOptions
      ),
    onLoaded: (response) => {
      usageLogs.value = response.items
      options.pagination.total = response.total
    },
    onError: (error) => {
      console.error('Failed to load usage logs:', error)
    }
  })

  watch(logsLoading, (value) => {
    loading.value = value
  }, { immediate: true })

  const loadStats = async () => {
    const requestSequence = ++statsRequestSequence
    endpointStatsLoading.value = true

    try {
      const stats = await adminAPI.usage.getStats(buildUsageFilters())

      if (requestSequence !== statsRequestSequence) {
        return
      }

      usageStats.value = stats
      inboundEndpointStats.value = stats.endpoints || []
      upstreamEndpointStats.value = stats.upstream_endpoints || []
      endpointPathStats.value = stats.endpoint_paths || []
    } catch (error) {
      if (requestSequence !== statsRequestSequence) {
        return
      }

      console.error('Failed to load usage stats:', error)
      inboundEndpointStats.value = []
      upstreamEndpointStats.value = []
      endpointPathStats.value = []
    } finally {
      if (requestSequence === statsRequestSequence) {
        endpointStatsLoading.value = false
      }
    }
  }

  const loadModelStats = async (source: ModelDistributionSource, force = false) => {
    if (!force && loadedModelSources[source]) {
      return
    }

    const requestSequence = ++modelStatsRequestSequence
    modelStatsLoading.value = true

    try {
      const response = await adminAPI.dashboard.getModelStats({
        ...buildUsageScopeParams(),
        model_source: source
      })

      if (requestSequence !== modelStatsRequestSequence) {
        return
      }

      setModelStats(source, response.models || [])
      loadedModelSources[source] = true
    } catch (error) {
      if (requestSequence !== modelStatsRequestSequence) {
        return
      }

      console.error('Failed to load model stats:', error)
      setModelStats(source, [])
      loadedModelSources[source] = false
    } finally {
      if (requestSequence === modelStatsRequestSequence) {
        modelStatsLoading.value = false
      }
    }
  }

  const loadChartData = async () => {
    const requestSequence = ++chartRequestSequence
    chartsLoading.value = true

    try {
      const snapshot = await adminAPI.dashboard.getSnapshotV2({
        ...buildUsageScopeParams(),
        granularity: options.granularity.value,
        include_stats: false,
        include_trend: true,
        include_model_stats: false,
        include_group_stats: true,
        include_users_trend: false
      })

      if (requestSequence !== chartRequestSequence) {
        return
      }

      trendData.value = snapshot.trend || []
      groupStats.value = snapshot.groups || []
    } catch (error) {
      console.error('Failed to load chart data:', error)
    } finally {
      if (requestSequence === chartRequestSequence) {
        chartsLoading.value = false
      }
    }
  }

  const refreshUsageData = (resetPage = false) => {
    if (resetPage) {
      resetUsagePaginationPage(options.pagination)
    }

    resetModelStatsCache()
    void loadLogs()
    void loadStats()
    void loadModelStats(options.modelDistributionSource.value, true)
    void loadChartData()
  }

  const applyFilters = () => {
    refreshUsageData(true)
  }

  const refreshData = () => {
    refreshUsageData(false)
  }

  const handlePageChange = (page: number) => {
    void handleUsageLogPageChange(page)
  }

  const handlePageSizeChange = (pageSize: number) => {
    void handleUsageLogPageSizeChange(pageSize)
  }

  const cancelExport = () => {
    exportAbortController?.abort()
  }

  const exportToExcel = async () => {
    if (exporting.value) {
      return
    }

    exporting.value = true
    exportProgress.show = true
    exportProgress.progress = 0
    exportProgress.current = 0
    exportProgress.total = 0
    exportProgress.estimatedTime = ''

    const controller = new AbortController()
    exportAbortController = controller

    try {
      const usageFilters = buildUsageFilters()
      const [XLSX, saveAs] = await Promise.all([import('xlsx'), getSaveAs()])
      const worksheet = XLSX.utils.aoa_to_sheet([buildUsageExportHeaders(options.t)])
      let page = 1
      let total = options.pagination.total
      let exportedCount = 0

      while (true) {
        const response = await adminUsageAPI.list(
          {
            page,
            page_size: 100,
            exact_total: true,
            ...usageFilters
          },
          { signal: controller.signal }
        )

        if (controller.signal.aborted) {
          return
        }

        if (page === 1) {
          total = response.total
          exportProgress.total = total
        }

        const rows = buildUsageExportRows(response.items || [], options.t)
        if (rows.length > 0) {
          XLSX.utils.sheet_add_aoa(worksheet, rows, { origin: -1 })
        }

        exportedCount += rows.length
        exportProgress.current = exportedCount
        exportProgress.progress =
          total > 0 ? Math.min(100, Math.round((exportedCount / total) * 100)) : 0

        if (exportedCount >= total || response.items.length < 100) {
          break
        }

        page += 1
      }

      const workbook = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(workbook, worksheet, 'Usage')
      saveAs(
        new Blob([XLSX.write(workbook, { bookType: 'xlsx', type: 'array' })], {
          type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        }),
        `usage_${usageFilters.start_date || options.startDate.value}_to_${usageFilters.end_date || options.endDate.value}.xlsx`
      )
      options.showSuccess(options.t('usage.exportSuccess'))
    } catch (error: unknown) {
      if (isAbortError(error)) {
        return
      }

      console.error('Failed to export:', error)
      options.showError(options.t('usage.exportFailed'))
    } finally {
      if (exportAbortController === controller) {
        exportAbortController = null
        exporting.value = false
        exportProgress.show = false
      }
    }
  }

  const loadInitialData = () => {
    void loadLogs()
    void loadStats()
    void loadModelStats(options.modelDistributionSource.value, true)

    if (initialChartTimer !== null) {
      window.clearTimeout(initialChartTimer)
    }

    initialChartTimer = window.setTimeout(() => {
      void loadChartData()
    }, 120)
  }

  const dispose = () => {
    disposeUsageLogs()
    exportAbortController?.abort()

    if (initialChartTimer !== null) {
      window.clearTimeout(initialChartTimer)
      initialChartTimer = null
    }
  }

  return {
    usageStats,
    usageLogs,
    loading,
    exporting,
    trendData,
    requestedModelStats,
    upstreamModelStats,
    mappingModelStats,
    groupStats,
    chartsLoading,
    modelStatsLoading,
    inboundEndpointStats,
    upstreamEndpointStats,
    endpointPathStats,
    endpointStatsLoading,
    exportProgress,
    applyFilters,
    refreshData,
    loadLogs,
    loadStats,
    loadModelStats,
    loadChartData,
    loadInitialData,
    handlePageChange,
    handlePageSizeChange,
    cancelExport,
    exportToExcel,
    dispose
  }
}
