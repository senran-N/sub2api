import { ref, type Ref } from 'vue'
import { keysAPI, usageAPI } from '@/api'
import { formatReasoningEffort } from '@/utils/format'
import { isAbortError, resolveRequestErrorMessage } from '@/utils/requestError'
import { getUsageRequestTypeExportText } from '@/utils/usagePresentation'
import type {
  ApiKey,
  UsageLog,
  UsageQueryParams,
  UsageStatsResponse
} from '@/types'
import {
  applyUserUsagePageChange,
  applyUserUsagePageSizeChange,
  resetUserUsagePaginationPage,
  type UserUsagePaginationState
} from './userUsageViewState'

interface UserUsageViewDataOptions {
  filters: Ref<UsageQueryParams>
  startDate: Ref<string>
  endDate: Ref<string>
  pagination: UserUsagePaginationState & { total: number; pages: number }
  showError: (message: string) => void
  showWarning: (message: string) => void
  showSuccess: (message: string) => void
  showInfo: (message: string) => void
  t: (key: string) => string
}

interface ResolvedUsageDateRange {
  startDate: string
  endDate: string
}

function escapeCSVValue(value: unknown): string {
  if (value == null) {
    return ''
  }

  const stringValue = String(value)
  const escaped = stringValue.replace(/"/g, '""')

  if (/^[=+\-@\t\r]/.test(stringValue)) {
    return `"\'${escaped}"`
  }

  if (/[,"\n\r]/.test(stringValue)) {
    return `"${escaped}"`
  }

  return stringValue
}

function buildUserUsageCsvContent(logs: UsageLog[]): string {
  const headers = [
    'Time',
    'API Key Name',
    'Model',
    'Reasoning Effort',
    'Inbound Endpoint',
    'Type',
    'Input Tokens',
    'Output Tokens',
    'Cache Read Tokens',
    'Cache Creation Tokens',
    'Rate Multiplier',
    'Billed Cost',
    'Original Cost',
    'First Token (ms)',
    'Duration (ms)'
  ]

  const rows = logs.map((log) =>
    [
      log.created_at,
      log.api_key?.name || '',
      log.model,
      formatReasoningEffort(log.reasoning_effort),
      log.inbound_endpoint || '',
      getUsageRequestTypeExportText(log),
      log.input_tokens,
      log.output_tokens,
      log.cache_read_tokens,
      log.cache_creation_tokens,
      log.rate_multiplier,
      log.actual_cost.toFixed(8),
      log.total_cost.toFixed(8),
      log.first_token_ms ?? '',
      log.duration_ms
    ].map(escapeCSVValue)
  )

  return [headers.map(escapeCSVValue).join(','), ...rows.map((row) => row.join(','))].join('\n')
}

function resolveUsageDateRange(
  options: UserUsageViewDataOptions
): ResolvedUsageDateRange {
  return {
    startDate: options.filters.value.start_date || options.startDate.value,
    endDate: options.filters.value.end_date || options.endDate.value
  }
}

function buildUsageQueryParams(
  options: UserUsageViewDataOptions,
  page: number,
  pageSize: number
): UsageQueryParams & { page: number; page_size: number } {
  const range = resolveUsageDateRange(options)

  return {
    page,
    page_size: pageSize,
    ...options.filters.value,
    start_date: range.startDate,
    end_date: range.endDate
  }
}

export function useUserUsageViewData(options: UserUsageViewDataOptions) {
  const usageStats = ref<UsageStatsResponse | null>(null)
  const usageLogs = ref<UsageLog[]>([])
  const apiKeys = ref<ApiKey[]>([])
  const loading = ref(false)
  const exporting = ref(false)

  let usageAbortController: AbortController | null = null

  const loadUsageLogs = async () => {
    usageAbortController?.abort()

    const controller = new AbortController()
    usageAbortController = controller
    loading.value = true

    try {
      const response = await usageAPI.query(
        buildUsageQueryParams(
          options,
          options.pagination.page,
          options.pagination.page_size
        ),
        { signal: controller.signal }
      )

      if (controller.signal.aborted) {
        return
      }

      usageLogs.value = response.items
      options.pagination.total = response.total
      options.pagination.pages = response.pages
    } catch (error) {
      if (controller.signal.aborted || isAbortError(error)) {
        return
      }

      options.showError(resolveRequestErrorMessage(error, options.t('usage.failedToLoad')))
    } finally {
      if (usageAbortController === controller) {
        loading.value = false
      }
    }
  }

  const loadApiKeys = async () => {
    try {
      const response = await keysAPI.list(1, 100)
      apiKeys.value = response.items
    } catch (error) {
      console.error('Failed to load API keys:', error)
    }
  }

  const loadUsageStats = async () => {
    try {
      const range = resolveUsageDateRange(options)
      const apiKeyId = options.filters.value.api_key_id
        ? Number(options.filters.value.api_key_id)
        : undefined

      usageStats.value = await usageAPI.getStatsByDateRange(
        range.startDate,
        range.endDate,
        apiKeyId
      )
    } catch (error) {
      console.error('Failed to load usage stats:', error)
    }
  }

  const refreshData = (resetPage = false) => {
    if (resetPage) {
      resetUserUsagePaginationPage(options.pagination)
    }

    void loadUsageLogs()
    void loadUsageStats()
  }

  const applyFilters = () => {
    refreshData(true)
  }

  const handlePageChange = (page: number) => {
    applyUserUsagePageChange(options.pagination, page)
    void loadUsageLogs()
  }

  const handlePageSizeChange = (pageSize: number) => {
    applyUserUsagePageSizeChange(options.pagination, pageSize)
    void loadUsageLogs()
  }

  const exportToCSV = async () => {
    if (options.pagination.total === 0) {
      options.showWarning(options.t('usage.noDataToExport'))
      return
    }

    exporting.value = true
    options.showInfo(options.t('usage.preparingExport'))

    try {
      const range = resolveUsageDateRange(options)
      const allLogs: UsageLog[] = []
      const pageSize = 100
      const totalRequests = Math.ceil(options.pagination.total / pageSize)

      for (let page = 1; page <= totalRequests; page += 1) {
        const response = await usageAPI.query(
          buildUsageQueryParams(options, page, pageSize)
        )
        allLogs.push(...response.items)
      }

      if (allLogs.length === 0) {
        options.showWarning(options.t('usage.noDataToExport'))
        return
      }

      const csvContent = buildUserUsageCsvContent(allLogs)
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `usage_${range.startDate}_to_${range.endDate}.csv`
      link.click()
      window.URL.revokeObjectURL(url)

      options.showSuccess(options.t('usage.exportSuccess'))
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('usage.exportFailed')))
      console.error('CSV Export failed:', error)
    } finally {
      exporting.value = false
    }
  }

  const loadInitialData = () => {
    void loadApiKeys()
    void loadUsageLogs()
    void loadUsageStats()
  }

  const dispose = () => {
    usageAbortController?.abort()
  }

  return {
    usageStats,
    usageLogs,
    apiKeys,
    loading,
    exporting,
    loadUsageLogs,
    loadApiKeys,
    loadUsageStats,
    applyFilters,
    refreshData,
    handlePageChange,
    handlePageSizeChange,
    exportToCSV,
    loadInitialData,
    dispose
  }
}
