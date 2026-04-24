import { ref, type Ref } from 'vue'
import {
  opsAPI,
  type OpsDashboardOverview,
  type OpsErrorDistributionResponse,
  type OpsErrorTrendResponse,
  type OpsLatencyHistogramResponse,
  type OpsMetricThresholds,
  type OpsThroughputTrendResponse
} from '@/api/admin/ops'
import { isAbortError, resolveRequestErrorMessage } from '@/utils/requestError'

export type OpsTimeRange = '5m' | '30m' | '1h' | '6h' | '24h' | 'custom'
export type OpsQueryMode = 'auto' | 'raw' | 'preagg'

type DashboardAPIParams = {
  time_range?: Exclude<OpsTimeRange, 'custom'>
  start_time?: string
  end_time?: string
  platform?: string
  group_id?: number
  mode: OpsQueryMode
}

interface UseOpsDashboardDataOptions {
  autoRefreshCountdown: Ref<number>
  autoRefreshEnabled: Ref<boolean>
  autoRefreshIntervalMs: Ref<number>
  customEndTime: Ref<string | null>
  customStartTime: Ref<string | null>
  groupId: Ref<number | null>
  opsEnabled: Ref<boolean>
  platform: Ref<string>
  queryMode: Ref<OpsQueryMode>
  showError: (message: string) => void
  t: (key: string) => string
  timeRange: Ref<OpsTimeRange>
}

const switchTrendWindowHours = 5
const switchTrendWindowMs = switchTrendWindowHours * 60 * 60 * 1000

function isCanceledRequest(err: unknown): boolean {
  return isAbortError(err)
}

function isOpsDisabledError(err: unknown): boolean {
  return (
    !!err &&
    typeof err === 'object' &&
    'code' in err &&
    typeof (err as Record<string, unknown>).code === 'string' &&
    (err as Record<string, unknown>).code === 'OPS_DISABLED'
  )
}

export function useOpsDashboardData(options: UseOpsDashboardDataOptions) {
  const loading = ref(true)
  const hasLoadedOnce = ref(false)
  const errorMessage = ref('')
  const lastUpdated = ref<Date | null>(new Date())

  const overview = ref<OpsDashboardOverview | null>(null)
  const metricThresholds = ref<OpsMetricThresholds | null>(null)

  const throughputTrend = ref<OpsThroughputTrendResponse | null>(null)
  const loadingTrend = ref(false)

  const switchTrend = ref<OpsThroughputTrendResponse | null>(null)
  const loadingSwitchTrend = ref(false)

  const latencyHistogram = ref<OpsLatencyHistogramResponse | null>(null)
  const loadingLatency = ref(false)

  const errorTrend = ref<OpsErrorTrendResponse | null>(null)
  const loadingErrorTrend = ref(false)

  const errorDistribution = ref<OpsErrorDistributionResponse | null>(null)
  const loadingErrorDistribution = ref(false)

  const dashboardRefreshToken = ref(0)

  let dashboardFetchController: AbortController | null = null
  let dashboardFetchSeq = 0

  function abortDashboardFetch() {
    if (dashboardFetchController) {
      dashboardFetchController.abort()
      dashboardFetchController = null
    }
  }

  function buildApiParams(): DashboardAPIParams {
    const params: DashboardAPIParams = {
      platform: options.platform.value || undefined,
      group_id: options.groupId.value ?? undefined,
      mode: options.queryMode.value
    }

    if (options.timeRange.value === 'custom') {
      if (options.customStartTime.value && options.customEndTime.value) {
        params.start_time = options.customStartTime.value
        params.end_time = options.customEndTime.value
      } else {
        params.time_range = '1h'
      }
    } else {
      params.time_range = options.timeRange.value
    }

    return params
  }

  function buildSwitchTrendParams(): DashboardAPIParams {
    const params: DashboardAPIParams = {
      platform: options.platform.value || undefined,
      group_id: options.groupId.value ?? undefined,
      mode: options.queryMode.value
    }
    const endTime = new Date()
    const startTime = new Date(endTime.getTime() - switchTrendWindowMs)
    params.start_time = startTime.toISOString()
    params.end_time = endTime.toISOString()
    return params
  }

  async function refreshOverviewWithCancel(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    try {
      const data = await opsAPI.getDashboardOverview(buildApiParams(), { signal })
      if (fetchSeq !== dashboardFetchSeq) return
      overview.value = data
    } catch (err: unknown) {
      if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
      overview.value = null
      options.showError(resolveRequestErrorMessage(err, options.t('admin.ops.failedToLoadOverview')))
    }
  }

  async function refreshSwitchTrendWithCancel(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    loadingSwitchTrend.value = true
    try {
      const data = await opsAPI.getThroughputTrend(buildSwitchTrendParams(), { signal })
      if (fetchSeq !== dashboardFetchSeq) return
      switchTrend.value = data
    } catch (err: unknown) {
      if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
      switchTrend.value = null
      options.showError(resolveRequestErrorMessage(err, options.t('admin.ops.failedToLoadSwitchTrend')))
    } finally {
      if (fetchSeq === dashboardFetchSeq) {
        loadingSwitchTrend.value = false
      }
    }
  }

  async function refreshThroughputTrendWithCancel(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    loadingTrend.value = true
    try {
      const data = await opsAPI.getThroughputTrend(buildApiParams(), { signal })
      if (fetchSeq !== dashboardFetchSeq) return
      throughputTrend.value = data
    } catch (err: unknown) {
      if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
      throughputTrend.value = null
      options.showError(resolveRequestErrorMessage(err, options.t('admin.ops.failedToLoadThroughputTrend')))
    } finally {
      if (fetchSeq === dashboardFetchSeq) {
        loadingTrend.value = false
      }
    }
  }

  async function refreshErrorTrendWithCancel(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    loadingErrorTrend.value = true
    try {
      const data = await opsAPI.getErrorTrend(buildApiParams(), { signal })
      if (fetchSeq !== dashboardFetchSeq) return
      errorTrend.value = data
    } catch (err: unknown) {
      if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
      errorTrend.value = null
      options.showError(resolveRequestErrorMessage(err, options.t('admin.ops.failedToLoadErrorTrend')))
    } finally {
      if (fetchSeq === dashboardFetchSeq) {
        loadingErrorTrend.value = false
      }
    }
  }

  async function refreshCoreSnapshotWithCancel(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    loadingTrend.value = true
    loadingErrorTrend.value = true
    try {
      const data = await opsAPI.getDashboardSnapshotV2(buildApiParams(), { signal })
      if (fetchSeq !== dashboardFetchSeq) return
      overview.value = data.overview
      throughputTrend.value = data.throughput_trend
      errorTrend.value = data.error_trend
    } catch (err: unknown) {
      if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
      await Promise.all([
        refreshOverviewWithCancel(fetchSeq, signal),
        refreshThroughputTrendWithCancel(fetchSeq, signal),
        refreshErrorTrendWithCancel(fetchSeq, signal)
      ])
    } finally {
      if (fetchSeq === dashboardFetchSeq) {
        loadingTrend.value = false
        loadingErrorTrend.value = false
      }
    }
  }

  async function refreshLatencyHistogramWithCancel(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    loadingLatency.value = true
    try {
      const data = await opsAPI.getLatencyHistogram(buildApiParams(), { signal })
      if (fetchSeq !== dashboardFetchSeq) return
      latencyHistogram.value = data
    } catch (err: unknown) {
      if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
      latencyHistogram.value = null
      options.showError(resolveRequestErrorMessage(err, options.t('admin.ops.failedToLoadLatencyHistogram')))
    } finally {
      if (fetchSeq === dashboardFetchSeq) {
        loadingLatency.value = false
      }
    }
  }

  async function refreshErrorDistributionWithCancel(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    loadingErrorDistribution.value = true
    try {
      const data = await opsAPI.getErrorDistribution(buildApiParams(), { signal })
      if (fetchSeq !== dashboardFetchSeq) return
      errorDistribution.value = data
    } catch (err: unknown) {
      if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
      errorDistribution.value = null
      options.showError(resolveRequestErrorMessage(err, options.t('admin.ops.failedToLoadErrorDistribution')))
    } finally {
      if (fetchSeq === dashboardFetchSeq) {
        loadingErrorDistribution.value = false
      }
    }
  }

  async function refreshDeferredPanels(fetchSeq: number, signal: AbortSignal) {
    if (!options.opsEnabled.value) return
    await Promise.all([
      refreshLatencyHistogramWithCancel(fetchSeq, signal),
      refreshErrorDistributionWithCancel(fetchSeq, signal)
    ])
  }

  async function fetchData() {
    if (!options.opsEnabled.value) return

    abortDashboardFetch()
    dashboardFetchSeq += 1
    const fetchSeq = dashboardFetchSeq
    dashboardFetchController = new AbortController()

    loading.value = true
    errorMessage.value = ''
    try {
      await Promise.all([
        refreshCoreSnapshotWithCancel(fetchSeq, dashboardFetchController.signal),
        refreshSwitchTrendWithCancel(fetchSeq, dashboardFetchController.signal)
      ])
      if (fetchSeq !== dashboardFetchSeq) return

      lastUpdated.value = new Date()
      dashboardRefreshToken.value += 1

      if (options.autoRefreshEnabled.value) {
        options.autoRefreshCountdown.value = Math.floor(options.autoRefreshIntervalMs.value / 1000)
      }

      void refreshDeferredPanels(fetchSeq, dashboardFetchController.signal)
    } catch (err) {
      if (!isOpsDisabledError(err)) {
        console.error('[ops] failed to fetch dashboard data', err)
        errorMessage.value = options.t('admin.ops.failedToLoadData')
      }
    } finally {
      if (fetchSeq === dashboardFetchSeq) {
        loading.value = false
        hasLoadedOnce.value = true
      }
    }
  }

  async function loadThresholds() {
    try {
      const thresholds = await opsAPI.getMetricThresholds()
      metricThresholds.value = thresholds || null
    } catch (err) {
      console.warn('[OpsDashboard] Failed to load thresholds', err)
      metricThresholds.value = null
    }
  }

  return {
    dashboardRefreshToken,
    errorDistribution,
    errorMessage,
    errorTrend,
    fetchData,
    hasLoadedOnce,
    lastUpdated,
    latencyHistogram,
    loadThresholds,
    loading,
    loadingErrorDistribution,
    loadingErrorTrend,
    loadingLatency,
    loadingSwitchTrend,
    loadingTrend,
    metricThresholds,
    overview,
    switchTrend,
    switchTrendTimeRange: `${switchTrendWindowHours}h`,
    throughputTrend,
    abortDashboardFetch
  }
}
