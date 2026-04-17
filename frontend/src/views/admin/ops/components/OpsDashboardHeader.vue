<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import {
  opsAPI,
  type OpsDashboardOverview,
  type OpsMetricThresholds,
  type OpsRealtimeTrafficSummary,
  type RuntimeObservabilitySnapshot
} from '@/api/admin/ops'
import type { OpsRequestDetailsPreset } from './OpsRequestDetailsModal.vue'
import {
  buildDiagnosisReport,
  buildGoroutineStatusDisplay,
  buildJobsStatusDisplay,
  buildPoolUsageDisplay,
  type DiagnosisItem,
  formatTimeShort,
  formatCustomTimeRangeLabel,
  getRequestErrorRateThresholdLevel,
  getSLAThresholdLevel,
  getThresholdColorClass,
  getTTFTThresholdLevel,
  getUpstreamErrorRateThresholdLevel
} from './opsDashboardHeaderHelpers'
import { useAdminSettingsStore } from '@/stores'
import { formatNumber } from '@/utils/format'

type RealtimeWindow = '1min' | '5min' | '30min' | '1h'

interface Props {
  overview?: OpsDashboardOverview | null
  platform: string
  groupId: number | null
  timeRange: string
  queryMode: string
  loading: boolean
  lastUpdated: Date | null
  thresholds?: OpsMetricThresholds | null // 阈值配置
  autoRefreshEnabled?: boolean
  autoRefreshCountdown?: number
  fullscreen?: boolean
  customStartTime?: string | null
  customEndTime?: string | null
}

interface Emits {
  (e: 'update:platform', value: string): void
  (e: 'update:group', value: number | null): void
  (e: 'update:timeRange', value: string): void
  (e: 'update:queryMode', value: string): void
  (e: 'update:customTimeRange', startTime: string, endTime: string): void
  (e: 'refresh'): void
  (e: 'openRequestDetails', preset?: OpsRequestDetailsPreset): void
  (e: 'openErrorDetails', kind: 'request' | 'upstream'): void
  (e: 'openSettings'): void
  (e: 'openAlertRules'): void
  (e: 'enterFullscreen'): void
  (e: 'exitFullscreen'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const adminSettingsStore = useAdminSettingsStore()

const realtimeWindow = ref<RealtimeWindow>('1min')

const overview = computed(() => props.overview ?? null)
const systemMetrics = computed(() => overview.value?.system_metrics ?? null)

const REALTIME_WINDOW_MINUTES: Record<RealtimeWindow, number> = {
  '1min': 1,
  '5min': 5,
  '30min': 30,
  '1h': 60
}

const TOOLBAR_RANGE_MINUTES: Record<string, number> = {
  '5m': 5,
  '30m': 30,
  '1h': 60,
  '6h': 6 * 60,
  '24h': 24 * 60
}

const availableRealtimeWindows = computed(() => {
  const toolbarMinutes = TOOLBAR_RANGE_MINUTES[props.timeRange] ?? 60
  return (['1min', '5min', '30min', '1h'] as const).filter((w) => REALTIME_WINDOW_MINUTES[w] <= toolbarMinutes)
})

watch(
  () => props.timeRange,
  () => {
    // The realtime window must be inside the toolbar window; reset to keep UX predictable.
    realtimeWindow.value = '1min'
    // Keep realtime traffic consistent with toolbar changes even when the window is already 1min.
    loadRealtimeTrafficSummary()
  }
)

// --- Filters ---

const showCustomTimeRangeDialog = ref(false)
const customStartTimeInput = ref('')
const customEndTimeInput = ref('')

const groups = ref<Array<{ id: number; name: string; platform: string }>>([])

const platformOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' }
])

const timeRangeOptions = computed(() => [
  { value: '5m', label: t('admin.ops.timeRange.5m') },
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '6h', label: t('admin.ops.timeRange.6h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  {
    value: 'custom',
    label: props.timeRange === 'custom' && props.customStartTime && props.customEndTime
      ? `${t('admin.ops.timeRange.custom')} (${formatCustomTimeRangeLabel(props.customStartTime, props.customEndTime)})`
      : t('admin.ops.timeRange.custom')
  }
])

const queryModeOptions = computed(() => [
  { value: 'auto', label: t('admin.ops.queryMode.auto') },
  { value: 'raw', label: t('admin.ops.queryMode.raw') },
  { value: 'preagg', label: t('admin.ops.queryMode.preagg') }
])

const groupOptions = computed(() => {
  const filtered = props.platform ? groups.value.filter((g) => g.platform === props.platform) : groups.value
  return [{ value: null, label: t('common.all') }, ...filtered.map((g) => ({ value: g.id, label: g.name }))]
})

const getCurrentSLAThresholdLevel = (slaPercent: number | null) =>
  getSLAThresholdLevel(slaPercent, props.thresholds)

const getCurrentTTFTThresholdLevel = (ttftMs: number | null) =>
  getTTFTThresholdLevel(ttftMs, props.thresholds)

const getCurrentRequestErrorRateThresholdLevel = (errorRatePercent: number | null) =>
  getRequestErrorRateThresholdLevel(errorRatePercent, props.thresholds)

const getCurrentUpstreamErrorRateThresholdLevel = (upstreamErrorRatePercent: number | null) =>
  getUpstreamErrorRateThresholdLevel(upstreamErrorRatePercent, props.thresholds)

watch(
  () => props.platform,
  (newPlatform) => {
    if (!newPlatform) return
    const currentGroup = groups.value.find((g) => g.id === props.groupId)
    if (currentGroup && currentGroup.platform !== newPlatform) {
      emit('update:group', null)
    }
  }
)

onMounted(async () => {
  try {
    const list = await adminAPI.groups.getAll()
    groups.value = list.map((g) => ({ id: g.id, name: g.name, platform: g.platform }))
  } catch (e) {
    console.error('[OpsDashboardHeader] Failed to load groups', e)
    groups.value = []
  }
})

function handlePlatformChange(val: string | number | boolean | null) {
  emit('update:platform', String(val || ''))
}

function handleGroupChange(val: string | number | boolean | null) {
  if (val === null || val === '' || typeof val === 'boolean') {
    emit('update:group', null)
    return
  }
  const id = typeof val === 'number' ? val : Number.parseInt(String(val), 10)
  emit('update:group', Number.isFinite(id) && id > 0 ? id : null)
}

function handleTimeRangeChange(val: string | number | boolean | null) {
  const newValue = String(val || '1h')
  if (newValue === 'custom') {
    // 初始化为最近1小时
    const now = new Date()
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000)
    customStartTimeInput.value = oneHourAgo.toISOString().slice(0, 16)
    customEndTimeInput.value = now.toISOString().slice(0, 16)
    showCustomTimeRangeDialog.value = true
  } else {
    emit('update:timeRange', newValue)
  }
}

function handleCustomTimeRangeConfirm() {
  if (!customStartTimeInput.value || !customEndTimeInput.value) return
  const startTime = new Date(customStartTimeInput.value).toISOString()
  const endTime = new Date(customEndTimeInput.value).toISOString()
  // Emit custom time range first so the parent can build correct API params
  // when it reacts to timeRange switching to "custom".
  emit('update:customTimeRange', startTime, endTime)
  emit('update:timeRange', 'custom')
  showCustomTimeRangeDialog.value = false
}

function handleCustomTimeRangeCancel() {
  showCustomTimeRangeDialog.value = false
  // 如果当前不是 custom，不需要做任何事
  // 如果当前是 custom，保持不变
}

function handleQueryModeChange(val: string | number | boolean | null) {
  emit('update:queryMode', String(val || 'auto'))
}

function openDetails(preset?: OpsRequestDetailsPreset) {
  emit('openRequestDetails', preset)
}

function openErrorDetails(kind: 'request' | 'upstream') {
  emit('openErrorDetails', kind)
}

// --- Realtime / Overview labels ---

const totalRequestsLabel = computed(() => formatNumber(overview.value?.request_count_total ?? 0))
const totalTokensLabel = computed(() => formatNumber(overview.value?.token_consumed ?? 0))

const realtimeTrafficSummary = ref<OpsRealtimeTrafficSummary | null>(null)
const realtimeRuntimeObservability = ref<RuntimeObservabilitySnapshot | null>(null)
const realtimeTrafficLoading = ref(false)
let realtimeTrafficRequestSequence = 0

function makeZeroRealtimeTrafficSummary(): OpsRealtimeTrafficSummary {
  const now = new Date().toISOString()
  return {
    window: realtimeWindow.value,
    start_time: now,
    end_time: now,
    platform: props.platform,
    group_id: props.groupId,
    qps: { current: 0, peak: 0, avg: 0 },
    tps: { current: 0, peak: 0, avg: 0 }
  }
}

async function loadRealtimeTrafficSummary() {
  const requestSequence = ++realtimeTrafficRequestSequence
  if (!adminSettingsStore.opsRealtimeMonitoringEnabled) {
    realtimeTrafficSummary.value = makeZeroRealtimeTrafficSummary()
    realtimeRuntimeObservability.value = null
    realtimeTrafficLoading.value = false
    return
  }
  realtimeTrafficLoading.value = true
  try {
    const res = await opsAPI.getRealtimeTrafficSummary(realtimeWindow.value, props.platform, props.groupId)
    if (requestSequence !== realtimeTrafficRequestSequence) return
    if (res && res.enabled === false) {
      adminSettingsStore.setOpsRealtimeMonitoringEnabledLocal(false)
    }
    realtimeTrafficSummary.value = res?.summary ?? null
    realtimeRuntimeObservability.value = res?.runtime_observability ?? null
  } catch (err) {
    if (requestSequence !== realtimeTrafficRequestSequence) return
    console.error('[OpsDashboardHeader] Failed to load realtime traffic summary', err)
    realtimeTrafficSummary.value = null
    realtimeRuntimeObservability.value = null
  } finally {
    if (requestSequence === realtimeTrafficRequestSequence) {
      realtimeTrafficLoading.value = false
    }
  }
}

watch(
  () => [realtimeWindow.value, props.platform, props.groupId] as const,
  () => {
    loadRealtimeTrafficSummary()
  },
  { immediate: true }
)

watch(
  () => adminSettingsStore.opsRealtimeMonitoringEnabled,
  (enabled) => {
    if (!enabled) {
      // Keep UI stable when realtime monitoring is turned off.
      realtimeTrafficSummary.value = makeZeroRealtimeTrafficSummary()
      realtimeRuntimeObservability.value = null
    } else {
      loadRealtimeTrafficSummary()
    }
  },
  { immediate: true }
)

// Realtime traffic refresh follows the parent (OpsDashboard) refresh cadence.
watch(
  () => [props.autoRefreshEnabled, props.autoRefreshCountdown, props.loading] as const,
  ([enabled, countdown, loading]) => {
    if (!enabled) return
    if (loading) return
    // Treat countdown reset (or reaching 0) as a refresh boundary.
    if (countdown === 0) {
      loadRealtimeTrafficSummary()
    }
  }
)

// no-op: parent controls refresh cadence

function formatFixedLabel(value: number | null | undefined, digits = 1) {
  return typeof value === 'number' && Number.isFinite(value) ? value.toFixed(digits) : '-'
}

function toPercent(value: number | null | undefined) {
  return typeof value === 'number' && Number.isFinite(value) ? value * 100 : null
}

function getThresholdIndicatorClass(level: 'normal' | 'warning' | 'critical') {
  if (level === 'critical') return 'ops-dashboard-header__indicator ops-dashboard-header__indicator--critical'
  if (level === 'warning') return 'ops-dashboard-header__indicator ops-dashboard-header__indicator--warning'
  return 'ops-dashboard-header__indicator ops-dashboard-header__indicator--healthy'
}

const realtimeTrafficDisplay = computed(() => {
  const qpsCurrent = realtimeTrafficSummary.value?.qps?.current
  const tpsCurrent = realtimeTrafficSummary.value?.tps?.current
  return {
    qpsCurrent: typeof qpsCurrent === 'number' && Number.isFinite(qpsCurrent) ? qpsCurrent : 0,
    tpsCurrent: typeof tpsCurrent === 'number' && Number.isFinite(tpsCurrent) ? tpsCurrent : 0,
    qpsPeakLabel: formatFixedLabel(realtimeTrafficSummary.value?.qps?.peak),
    tpsPeakLabel: formatFixedLabel(realtimeTrafficSummary.value?.tps?.peak),
    qpsAvgLabel: formatFixedLabel(realtimeTrafficSummary.value?.qps?.avg),
    tpsAvgLabel: formatFixedLabel(realtimeTrafficSummary.value?.tps?.avg)
  }
})

function formatRuntimePercent(value: number | null | undefined) {
  return typeof value === 'number' && Number.isFinite(value) ? `${(value * 100).toFixed(1)}%` : '-'
}

const runtimeObservabilityItems = computed(() => {
  const summary = realtimeRuntimeObservability.value?.summary
  return [
    {
      key: 'page_density',
      label: t('admin.ops.runtimeObservability.pageDensity'),
      value: formatFixedLabel(summary?.scheduling_runtime_kernel?.avg_fetched_accounts_per_page),
      hint: t('admin.ops.runtimeObservability.pageDensityHint')
    },
    {
      key: 'acquire_success',
      label: t('admin.ops.runtimeObservability.acquireSuccess'),
      value: formatRuntimePercent(summary?.scheduling_runtime_kernel?.acquire_success_rate),
      hint: t('admin.ops.runtimeObservability.acquireSuccessHint')
    },
    {
      key: 'wait_plan_success',
      label: t('admin.ops.runtimeObservability.waitPlanSuccess'),
      value: formatRuntimePercent(summary?.scheduling_runtime_kernel?.wait_plan_success_rate),
      hint: t('admin.ops.runtimeObservability.waitPlanSuccessHint')
    },
    {
      key: 'idempotency_duration',
      label: t('admin.ops.runtimeObservability.idempotencyAvg'),
      value: formatFixedLabel(summary?.idempotency?.avg_processing_duration_ms),
      suffix: 'ms',
      hint: t('admin.ops.runtimeObservability.idempotencyAvgHint')
    }
  ]
})

const overviewMetricDisplay = computed(() => ({
  qpsAvgLabel: formatFixedLabel(overview.value?.qps?.avg),
  tpsAvgLabel: formatFixedLabel(overview.value?.tps?.avg),
  slaPercent: toPercent(overview.value?.sla),
  errorRatePercent: toPercent(overview.value?.error_rate),
  upstreamErrorRatePercent: toPercent(overview.value?.upstream_error_rate)
}))

const durationMetrics = computed(() => overview.value?.duration ?? {})
const ttftMetrics = computed(() => overview.value?.ttft ?? {})

const slaDisplay = computed(() => {
  const percent = overviewMetricDisplay.value.slaPercent
  const level = getCurrentSLAThresholdLevel(percent)
  return {
    percent,
    level,
    colorClass: getThresholdColorClass(level),
    indicatorClass: getThresholdIndicatorClass(level),
    progressWidth: `${Math.max((percent ?? 0) - 90, 0) * 10}%`
  }
})

const requestErrorDisplay = computed(() => {
  const percent = overviewMetricDisplay.value.errorRatePercent
  const level = getCurrentRequestErrorRateThresholdLevel(percent)
  return {
    percent,
    colorClass: getThresholdColorClass(level)
  }
})

const upstreamErrorDisplay = computed(() => {
  const percent = overviewMetricDisplay.value.upstreamErrorRatePercent
  const level = getCurrentUpstreamErrorRateThresholdLevel(percent)
  return {
    percent,
    colorClass: getThresholdColorClass(level)
  }
})

// --- Health Score & Diagnosis (primary) ---

const isSystemIdle = computed(() => {
  const ov = overview.value
  if (!ov) return true
  const qps = ov.qps?.current
  const errorRate = ov.error_rate ?? 0
  return (qps ?? 0) === 0 && errorRate === 0
})

const healthScoreValue = computed<number | null>(() => {
  const v = overview.value?.health_score
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const healthScoreDisplay = computed(() => {
  const circleSize = props.fullscreen ? 140 : 100
  const strokeWidth = props.fullscreen ? 10 : 8
  const radius = (circleSize - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const score = healthScoreValue.value
  const tone =
    isSystemIdle.value || score == null
      ? { color: 'color-mix(in srgb, var(--theme-page-muted) 70%, transparent)', className: 'ops-dashboard-header__tone ops-dashboard-header__tone--muted' }
      : score >= 90
        ? { color: 'rgb(var(--theme-success-rgb))', className: 'ops-dashboard-header__tone ops-dashboard-header__tone--healthy' }
        : score >= 60
          ? { color: 'rgb(var(--theme-warning-rgb))', className: 'ops-dashboard-header__tone ops-dashboard-header__tone--warning' }
          : { color: 'rgb(var(--theme-danger-rgb))', className: 'ops-dashboard-header__tone ops-dashboard-header__tone--critical' }

  const clampedScore = score == null ? 0 : Math.max(0, Math.min(100, score))
  return {
    color: tone.color,
    className: tone.className,
    circleSize,
    strokeWidth,
    radius,
    circumference,
    dashOffset: isSystemIdle.value || score == null
      ? 0
      : circumference - (clampedScore / 100) * circumference
  }
})

const diagnosisReport = computed<DiagnosisItem[]>(() => {
  return buildDiagnosisReport({
    overview: overview.value,
    isSystemIdle: isSystemIdle.value,
    healthScore: healthScoreValue.value,
    t
  })
})

// --- System health (secondary) ---

const cpuPercentValue = computed<number | null>(() => {
  const v = systemMetrics.value?.cpu_usage_percent
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const cpuPercentClass = computed(() => {
  const v = cpuPercentValue.value
  if (v == null) return 'ops-dashboard-header__tone ops-dashboard-header__tone--default'
  if (v >= 95) return 'ops-dashboard-header__tone ops-dashboard-header__tone--critical'
  if (v >= 80) return 'ops-dashboard-header__tone ops-dashboard-header__tone--warning'
  return 'ops-dashboard-header__tone ops-dashboard-header__tone--healthy'
})

const memPercentValue = computed<number | null>(() => {
  const v = systemMetrics.value?.memory_usage_percent
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const memPercentClass = computed(() => {
  const v = memPercentValue.value
  if (v == null) return 'ops-dashboard-header__tone ops-dashboard-header__tone--default'
  if (v >= 95) return 'ops-dashboard-header__tone ops-dashboard-header__tone--critical'
  if (v >= 85) return 'ops-dashboard-header__tone ops-dashboard-header__tone--warning'
  return 'ops-dashboard-header__tone ops-dashboard-header__tone--healthy'
})

const dbConnActiveValue = computed<number | null>(() => {
  const v = systemMetrics.value?.db_conn_active
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const dbConnIdleValue = computed<number | null>(() => {
  const v = systemMetrics.value?.db_conn_idle
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const dbConnWaitingValue = computed<number | null>(() => {
  const v = systemMetrics.value?.db_conn_waiting
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const dbConnOpenValue = computed<number | null>(() => {
  if (dbConnActiveValue.value == null || dbConnIdleValue.value == null) return null
  return dbConnActiveValue.value + dbConnIdleValue.value
})

const dbMaxOpenConnsValue = computed<number | null>(() => {
  const v = systemMetrics.value?.db_max_open_conns
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const dbUsagePercent = computed<number | null>(() => {
  if (dbConnOpenValue.value == null || dbMaxOpenConnsValue.value == null || dbMaxOpenConnsValue.value <= 0) return null
  return Math.min(100, Math.max(0, (dbConnOpenValue.value / dbMaxOpenConnsValue.value) * 100))
})

const dbMiddleDisplay = computed(() => {
  return buildPoolUsageDisplay(systemMetrics.value?.db_ok, dbUsagePercent.value, t)
})

const redisConnTotalValue = computed<number | null>(() => {
  const v = systemMetrics.value?.redis_conn_total
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const redisConnIdleValue = computed<number | null>(() => {
  const v = systemMetrics.value?.redis_conn_idle
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const redisConnActiveValue = computed<number | null>(() => {
  if (redisConnTotalValue.value == null || redisConnIdleValue.value == null) return null
  return Math.max(redisConnTotalValue.value - redisConnIdleValue.value, 0)
})

const redisPoolSizeValue = computed<number | null>(() => {
  const v = systemMetrics.value?.redis_pool_size
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const redisUsagePercent = computed<number | null>(() => {
  if (redisConnTotalValue.value == null || redisPoolSizeValue.value == null || redisPoolSizeValue.value <= 0) return null
  return Math.min(100, Math.max(0, (redisConnTotalValue.value / redisPoolSizeValue.value) * 100))
})

const redisMiddleDisplay = computed(() => {
  return buildPoolUsageDisplay(systemMetrics.value?.redis_ok, redisUsagePercent.value, t)
})

const goroutineCountValue = computed<number | null>(() => {
  const v = systemMetrics.value?.goroutine_count
  return typeof v === 'number' && Number.isFinite(v) ? v : null
})

const goroutinesWarnThreshold = 8_000
const goroutinesCriticalThreshold = 15_000

const goroutineDisplay = computed(() => {
  return buildGoroutineStatusDisplay(
    goroutineCountValue.value,
    t,
    goroutinesWarnThreshold,
    goroutinesCriticalThreshold
  )
})

const jobHeartbeats = computed(() => overview.value?.job_heartbeats ?? [])

const jobsDisplay = computed(() => {
  return buildJobsStatusDisplay(jobHeartbeats.value, t)
})

const showJobsDetails = ref(false)

function openJobsDetails() {
  showJobsDetails.value = true
}

function handleToolbarRefresh() {
  loadRealtimeTrafficSummary()
  emit('refresh')
}
</script>

<template>
  <div :class="['ops-dashboard-header__surface flex flex-col gap-4', props.fullscreen ? 'ops-dashboard-header__surface--fullscreen' : 'ops-dashboard-header__surface--default']">
    <!-- Top Toolbar -->
    <div class="ops-dashboard-header__divider flex flex-wrap items-center justify-between gap-4 pb-4">
      <div>
        <h1 class="ops-dashboard-header__title flex items-center gap-2 text-xl font-black">
          <svg class="ops-dashboard-header__title-icon h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            />
          </svg>
          {{ t('admin.ops.title') }}
        </h1>

        <div v-if="!props.fullscreen" class="ops-dashboard-header__meta mt-1 flex items-center gap-3 text-xs">
          <span class="flex items-center gap-1.5" :title="props.loading ? t('admin.ops.loadingText') : t('admin.ops.ready')">
            <span class="relative flex h-2 w-2">
              <span
                class="ops-dashboard-header__status-dot relative inline-flex h-2 w-2 rounded-full"
                :class="props.loading ? 'ops-dashboard-header__status-dot--loading' : 'ops-dashboard-header__status-dot--ready'"
              ></span>
            </span>
            {{ props.loading ? t('admin.ops.loadingText') : t('admin.ops.ready') }}
          </span>

          <span>·</span>
          <span>{{ t('common.refresh') }}: {{ props.lastUpdated ? props.lastUpdated.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' }).replace(/\//g, '-') : t('common.unknown') }}</span>

          <template v-if="props.autoRefreshEnabled && props.autoRefreshCountdown !== undefined">
            <span>·</span>
            <span>剩余 {{ props.autoRefreshCountdown }}s</span>
          </template>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <template v-if="!props.fullscreen">
          <Select
            :model-value="platform"
            :options="platformOptions"
            class="ops-dashboard-header__select ops-dashboard-header__select--platform"
            @update:model-value="handlePlatformChange"
          />

          <Select
            :model-value="groupId"
            :options="groupOptions"
            class="ops-dashboard-header__select ops-dashboard-header__select--group"
            @update:model-value="handleGroupChange"
          />

          <div class="ops-dashboard-header__toolbar-divider mx-1 hidden sm:block"></div>

          <Select
            :model-value="timeRange"
            :options="timeRangeOptions"
            class="ops-dashboard-header__select ops-dashboard-header__select--time-range relative"
            @update:model-value="handleTimeRangeChange"
          />
        </template>

        <Select
          v-if="false"
          :model-value="queryMode"
          :options="queryModeOptions"
          class="ops-dashboard-header__select ops-dashboard-header__select--query-mode relative"
          @update:model-value="handleQueryModeChange"
        />

        <button
          v-if="!props.fullscreen"
          type="button"
          class="ops-dashboard-header__icon-button flex h-8 w-8 items-center justify-center transition-colors"
          :disabled="loading"
          :title="t('common.refresh')"
          @click="handleToolbarRefresh"
        >
          <svg class="h-4 w-4" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
        </button>

        <div v-if="!props.fullscreen" class="ops-dashboard-header__toolbar-divider mx-1 hidden sm:block"></div>

        <!-- Alert Rules Button (hidden in fullscreen) -->
        <button
          v-if="!props.fullscreen"
          type="button"
          class="ops-dashboard-header__action-button ops-dashboard-header__action-button--info flex h-8 items-center gap-1.5 text-xs font-bold transition-colors"
          :title="t('admin.ops.alertRules.title')"
          @click="emit('openAlertRules')"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          <span class="hidden sm:inline">{{ t('admin.ops.alertRules.manage') }}</span>
        </button>

        <!-- Settings Button (hidden in fullscreen) -->
        <button
          v-if="!props.fullscreen"
          type="button"
          class="ops-dashboard-header__action-button ops-dashboard-header__action-button--neutral flex h-8 items-center gap-1.5 text-xs font-bold transition-colors"
          :title="t('admin.ops.settings.title')"
          @click="emit('openSettings')"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span class="hidden sm:inline">{{ t('common.settings') }}</span>
        </button>

        <!-- Enter Fullscreen Button (hidden in fullscreen mode) -->
        <button
          v-if="!props.fullscreen"
          type="button"
          class="ops-dashboard-header__icon-button ops-dashboard-header__action-button--neutral flex h-8 w-8 items-center justify-center transition-colors"
          :title="t('admin.ops.fullscreen.enter')"
          @click="emit('enterFullscreen')"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
          </svg>
        </button>
      </div>
    </div>

    <div v-if="overview" class="grid grid-cols-1 gap-6 lg:grid-cols-12">
      <!-- Left: Health + Realtime -->
      <div :class="['ops-dashboard-header__panel lg:col-span-5', props.fullscreen ? 'ops-dashboard-header__panel--fullscreen' : 'ops-dashboard-header__panel--default']">
        <div class="grid h-full grid-cols-1 gap-6 md:grid-cols-[200px_1fr] md:items-center">
          <!-- 1) Health Score -->
          <div
            class="ops-dashboard-header__health group relative flex cursor-pointer flex-col items-center justify-center transition-all md:pr-6"
          >
            <!-- Diagnosis Popover (hover) -->
            <div
              class="pointer-events-none absolute left-1/2 top-full z-50 mt-2 w-72 -translate-x-1/2 opacity-0 transition-opacity duration-200 group-hover:pointer-events-auto group-hover:opacity-100 md:left-full md:top-0 md:ml-2 md:mt-0 md:translate-x-0"
            >
              <div class="ops-dashboard-header__popover">
                <h4 class="ops-dashboard-header__popover-title mb-3 flex items-center gap-2 pb-2 text-sm font-bold">
                  <Icon name="brain" size="sm" class="ops-dashboard-header__title-icon" />
                  {{ t('admin.ops.diagnosis.title') }}
                </h4>

                <div class="space-y-3">
                  <div v-for="(item, idx) in diagnosisReport" :key="idx" class="flex gap-3">
                    <div class="mt-0.5 shrink-0">
                      <svg v-if="item.type === 'critical'" class="ops-dashboard-header__tone ops-dashboard-header__tone--critical h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          fill-rule="evenodd"
                          d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                          clip-rule="evenodd"
                        />
                      </svg>
                      <svg v-else-if="item.type === 'warning'" class="ops-dashboard-header__tone ops-dashboard-header__tone--warning h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          fill-rule="evenodd"
                          d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                          clip-rule="evenodd"
                        />
                      </svg>
                      <svg v-else class="ops-dashboard-header__tone ops-dashboard-header__tone--info h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          fill-rule="evenodd"
                          d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 100 2 1 1 0 000-2zm-1 3a1 1 0 012 0v4a1 1 0 11-2 0v-4z"
                          clip-rule="evenodd"
                        />
                      </svg>
                    </div>
                    <div class="flex-1">
                      <div class="ops-dashboard-header__text text-xs font-semibold">{{ item.message }}</div>
                      <div class="ops-dashboard-header__meta mt-0.5 text-[11px]">{{ item.impact }}</div>
                      <div v-if="item.action" class="ops-dashboard-header__tone ops-dashboard-header__tone--info mt-1 flex items-center gap-1 text-[11px]">
                        <Icon name="lightbulb" size="xs" />
                        {{ item.action }}
                      </div>
                    </div>
                  </div>
                </div>

                <div class="ops-dashboard-header__popover-footer mt-3 pt-2 text-[10px]">
                  {{ t('admin.ops.diagnosis.footer') }}
                </div>
              </div>
            </div>

            <div class="relative flex items-center justify-center">
              <svg :width="healthScoreDisplay.circleSize" :height="healthScoreDisplay.circleSize" class="-rotate-90 transform">
                <circle
                  :cx="healthScoreDisplay.circleSize / 2"
                  :cy="healthScoreDisplay.circleSize / 2"
                  :r="healthScoreDisplay.radius"
                  :stroke-width="healthScoreDisplay.strokeWidth"
                  fill="transparent"
                  class="ops-dashboard-header__ring-track"
                  stroke="currentColor"
                />
                <circle
                  :cx="healthScoreDisplay.circleSize / 2"
                  :cy="healthScoreDisplay.circleSize / 2"
                  :r="healthScoreDisplay.radius"
                  :stroke-width="healthScoreDisplay.strokeWidth"
                  fill="transparent"
                  :stroke="healthScoreDisplay.color"
                  stroke-linecap="round"
                  :stroke-dasharray="healthScoreDisplay.circumference"
                  :stroke-dashoffset="healthScoreDisplay.dashOffset"
                  class="transition-all duration-1000 ease-out"
                />
              </svg>

              <div class="absolute flex flex-col items-center">
                <span :class="[props.fullscreen ? 'text-5xl' : 'text-3xl', 'font-black', healthScoreDisplay.className]">
                  {{ isSystemIdle ? t('admin.ops.idleStatus') : (overview.health_score ?? '--') }}
                </span>
                <span :class="[props.fullscreen ? 'text-xs' : 'text-[10px]', 'ops-dashboard-header__muted font-bold uppercase tracking-wider']">{{ t('admin.ops.health') }}</span>
              </div>
            </div>

            <div class="mt-4 text-center" v-if="!props.fullscreen">
              <div class="ops-dashboard-header__meta flex items-center justify-center gap-1 text-xs font-medium">
                {{ t('admin.ops.healthCondition') }}
                <HelpTooltip :content="t('admin.ops.healthHelp')" />
              </div>
              <div class="mt-1 text-xs font-bold" :class="healthScoreDisplay.className">
                {{
                  isSystemIdle
                    ? t('admin.ops.idleStatus')
                    : typeof overview.health_score === 'number' && overview.health_score >= 90
                      ? t('admin.ops.healthyStatus')
                      : t('admin.ops.riskyStatus')
                }}
              </div>
            </div>
          </div>

          <!-- 2) Realtime Traffic -->
          <div class="ops-dashboard-header__realtime flex h-full flex-col justify-center">
            <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
              <div class="flex items-center gap-2">
                <div class="relative flex h-3 w-3 shrink-0">
                  <span class="ops-dashboard-header__realtime-ping absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"></span>
                  <span class="ops-dashboard-header__realtime-dot relative inline-flex h-3 w-3 rounded-full"></span>
                </div>
                <h3 class="ops-dashboard-header__muted text-xs font-bold uppercase tracking-wider">{{ t('admin.ops.realtime.title') }}</h3>
                <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.qps')" />
              </div>

              <!-- Time Window Selector -->
              <div class="flex flex-wrap gap-1">
                <button
                  v-for="window in availableRealtimeWindows"
                  :key="window"
                  type="button"
                  class="ops-dashboard-header__window text-[9px] font-bold transition-colors sm:text-[10px]"
                  :class="realtimeWindow === window
                    ? 'ops-dashboard-header__window--active'
                    : 'ops-dashboard-header__window--idle'"
                  @click="realtimeWindow = window"
                >
                  {{ window }}
                </button>
              </div>
            </div>

            <div :class="props.fullscreen ? 'space-y-4' : 'space-y-3'">
              <!-- Row 1: Current -->
              <div>
                <div :class="[props.fullscreen ? 'text-xs' : 'text-[10px]', 'ops-dashboard-header__eyebrow font-bold uppercase']">{{ t('admin.ops.current') }}</div>
                <div class="mt-1 flex flex-wrap items-baseline gap-x-4 gap-y-2">
                  <div class="flex items-baseline gap-1.5">
                    <span :class="[props.fullscreen ? 'text-4xl' : 'text-xl sm:text-2xl', 'ops-dashboard-header__value font-black']">{{ realtimeTrafficDisplay.qpsCurrent.toFixed(1) }}</span>
                    <span :class="[props.fullscreen ? 'text-sm' : 'text-xs', 'ops-dashboard-header__label font-bold']">QPS</span>
                  </div>
                  <div class="flex items-baseline gap-1.5">
                    <span :class="[props.fullscreen ? 'text-4xl' : 'text-xl sm:text-2xl', 'ops-dashboard-header__value font-black']">{{ realtimeTrafficDisplay.tpsCurrent.toFixed(1) }}</span>
                    <span :class="[props.fullscreen ? 'text-sm' : 'text-xs', 'ops-dashboard-header__label font-bold']">{{ t('admin.ops.tps') }}</span>
                  </div>
                </div>
              </div>

              <!-- Row 2: Peak + Average -->
              <div class="grid grid-cols-2 gap-3">
                <!-- Peak -->
                <div>
                  <div :class="[props.fullscreen ? 'text-xs' : 'text-[10px]', 'ops-dashboard-header__eyebrow font-bold uppercase']">{{ t('admin.ops.peak') }}</div>
                  <div :class="[props.fullscreen ? 'text-base' : 'text-sm', 'ops-dashboard-header__metric-list mt-1 space-y-0.5 font-medium']">
                    <div class="flex items-baseline gap-1.5">
                      <span class="ops-dashboard-header__value font-black">{{ realtimeTrafficDisplay.qpsPeakLabel }}</span>
                      <span class="text-xs">QPS</span>
                    </div>
                    <div class="flex items-baseline gap-1.5">
                      <span class="ops-dashboard-header__value font-black">{{ realtimeTrafficDisplay.tpsPeakLabel }}</span>
                      <span class="text-xs">{{ t('admin.ops.tps') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Average -->
                <div>
                  <div :class="[props.fullscreen ? 'text-xs' : 'text-[10px]', 'ops-dashboard-header__eyebrow font-bold uppercase']">{{ t('admin.ops.average') }}</div>
                  <div :class="[props.fullscreen ? 'text-base' : 'text-sm', 'ops-dashboard-header__metric-list mt-1 space-y-0.5 font-medium']">
                    <div class="flex items-baseline gap-1.5">
                      <span class="ops-dashboard-header__value font-black">{{ realtimeTrafficDisplay.qpsAvgLabel }}</span>
                      <span class="text-xs">QPS</span>
                    </div>
                    <div class="flex items-baseline gap-1.5">
                      <span class="ops-dashboard-header__value font-black">{{ realtimeTrafficDisplay.tpsAvgLabel }}</span>
                      <span class="text-xs">{{ t('admin.ops.tps') }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Animated Pulse Line (Heart Beat Animation) -->
              <div class="h-8 w-full overflow-hidden opacity-50">
                <svg class="h-full w-full" viewBox="0 0 280 32" preserveAspectRatio="none">
                  <path
                    class="ops-dashboard-header__pulse-line"
                    d="M0 16 Q 20 16, 40 16 T 80 16 T 120 10 T 160 22 T 200 16 T 240 16 T 280 16"
                    fill="none"
                    stroke-width="2"
                    vector-effect="non-scaling-stroke"
                  >
                    <animate
                      attributeName="d"
                      dur="2s"
                      repeatCount="indefinite"
                      values="M0 16 Q 20 16, 40 16 T 80 16 T 120 10 T 160 22 T 200 16 T 240 16 T 280 16;
                              M0 16 Q 20 16, 40 16 T 80 16 T 120 16 T 160 16 T 200 10 T 240 22 T 280 16;
                              M0 16 Q 20 16, 40 16 T 80 16 T 120 16 T 160 16 T 200 16 T 240 16 T 280 16"
                      keyTimes="0;0.5;1"
                    />
                  </path>
                </svg>
              </div>

              <div class="ops-dashboard-header__runtime-strip">
                <div class="mb-2 flex items-center gap-2">
                  <div :class="[props.fullscreen ? 'text-xs' : 'text-[10px]', 'ops-dashboard-header__eyebrow font-bold uppercase']">
                    {{ t('admin.ops.runtimeObservability.title') }}
                  </div>
                  <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.runtimeObservability.help')" />
                </div>
                <div class="grid grid-cols-2 gap-2">
                  <div
                    v-for="item in runtimeObservabilityItems"
                    :key="item.key"
                    class="ops-dashboard-header__runtime-chip"
                    :title="item.hint"
                  >
                    <div class="ops-dashboard-header__muted text-[10px] font-bold uppercase tracking-wider">
                      {{ item.label }}
                    </div>
                    <div class="mt-1 flex items-baseline gap-1">
                      <span class="ops-dashboard-header__value text-sm font-black">{{ item.value }}</span>
                      <span v-if="item.suffix" class="ops-dashboard-header__label text-[10px] font-bold">{{ item.suffix }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: 6 cards (3 cols x 2 rows) -->
      <div class="grid h-full grid-cols-1 content-center gap-4 sm:grid-cols-2 lg:col-span-7 lg:grid-cols-3">
        <!-- Card 1: Requests -->
        <div class="ops-dashboard-header__metric-card" style="order: 1;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1">
              <span class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase">{{ t('admin.ops.requestsTitle') }}</span>
              <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.totalRequests')" />
            </div>
            <button
              v-if="!props.fullscreen"
              class="ops-dashboard-header__link text-[10px] font-bold"
              type="button"
              @click="openDetails({ title: t('admin.ops.requestDetails.title') })"
            >
              {{ t('admin.ops.requestDetails.details') }}
            </button>
          </div>
          <div class="mt-2 space-y-2 text-xs">
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.requests') }}:</span>
              <span class="ops-dashboard-header__value font-bold">{{ totalRequestsLabel }}</span>
            </div>
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.tokens') }}:</span>
              <span class="ops-dashboard-header__value font-bold">{{ totalTokensLabel }}</span>
            </div>
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.avgQps') }}:</span>
              <span class="ops-dashboard-header__value font-bold">{{ overviewMetricDisplay.qpsAvgLabel }}</span>
            </div>
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.avgTps') }}:</span>
              <span class="ops-dashboard-header__value font-bold">{{ overviewMetricDisplay.tpsAvgLabel }}</span>
            </div>
          </div>
        </div>

        <!-- Card 2: SLA -->
        <div class="ops-dashboard-header__metric-card" style="order: 2;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase">{{ t('admin.ops.sla') }}</span>
              <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.sla')" />
              <span class="h-1.5 w-1.5 rounded-full" :class="slaDisplay.indicatorClass"></span>
            </div>
            <button
              v-if="!props.fullscreen"
              class="ops-dashboard-header__link text-[10px] font-bold"
              type="button"
              @click="openDetails({ title: t('admin.ops.requestDetails.title'), kind: 'error' })"
            >
              {{ t('admin.ops.requestDetails.details') }}
            </button>
          </div>
          <div class="mt-2 text-3xl font-black" :class="slaDisplay.colorClass">
            {{ slaDisplay.percent == null ? '-' : `${slaDisplay.percent.toFixed(3)}%` }}
          </div>
          <div class="ops-dashboard-header__progress-track mt-3 h-2 w-full overflow-hidden rounded-full">
            <div class="h-full transition-all" :class="slaDisplay.indicatorClass" :style="{ width: slaDisplay.progressWidth }"></div>
          </div>
          <div class="mt-3 text-xs">
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.exceptions') }}:</span>
              <span class="ops-dashboard-header__tone ops-dashboard-header__tone--critical font-bold">{{ formatNumber((overview.request_count_sla ?? 0) - (overview.success_count ?? 0)) }}</span>
            </div>
          </div>
        </div>

        <!-- Card 4: Request Duration -->
        <div class="ops-dashboard-header__metric-card" style="order: 4;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1">
              <span class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase">{{ t('admin.ops.latencyDuration') }}</span>
              <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.latency')" />
            </div>
            <button
              v-if="!props.fullscreen"
              class="ops-dashboard-header__link text-[10px] font-bold"
              type="button"
              @click="openDetails({ title: t('admin.ops.latencyDuration'), sort: 'duration_desc' })"
            >
              {{ t('admin.ops.requestDetails.details') }}
            </button>
          </div>
          <div class="mt-2 flex items-baseline gap-2">
            <div class="ops-dashboard-header__value text-3xl font-black">
              {{ durationMetrics.p99_ms ?? '-' }}
            </div>
            <span class="ops-dashboard-header__eyebrow text-xs font-bold">ms (P99)</span>
          </div>
          <div class="mt-3 grid grid-cols-1 gap-x-3 gap-y-1 text-xs 2xl:grid-cols-2">
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">P95:</span>
              <span class="ops-dashboard-header__value font-bold">{{ durationMetrics.p95_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">P90:</span>
              <span class="ops-dashboard-header__value font-bold">{{ durationMetrics.p90_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">P50:</span>
              <span class="ops-dashboard-header__value font-bold">{{ durationMetrics.p50_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">Avg:</span>
              <span class="ops-dashboard-header__value font-bold">{{ durationMetrics.avg_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">Max:</span>
              <span class="ops-dashboard-header__value font-bold">{{ durationMetrics.max_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
          </div>
        </div>

        <!-- Card 5: TTFT -->
        <div class="ops-dashboard-header__metric-card" style="order: 5;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1">
              <span class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase">TTFT</span>
              <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.ttft')" />
            </div>
            <button
              v-if="!props.fullscreen"
              class="ops-dashboard-header__link text-[10px] font-bold"
              type="button"
              @click="openDetails({ title: t('admin.ops.ttftLabel'), sort: 'duration_desc' })"
            >
              {{ t('admin.ops.requestDetails.details') }}
            </button>
          </div>
          <div class="mt-2 flex items-baseline gap-2">
            <div class="text-3xl font-black" :class="getThresholdColorClass(getCurrentTTFTThresholdLevel(ttftMetrics.p99_ms ?? null))">
              {{ ttftMetrics.p99_ms ?? '-' }}
            </div>
            <span class="ops-dashboard-header__eyebrow text-xs font-bold">ms (P99)</span>
          </div>
          <div class="mt-3 grid grid-cols-1 gap-x-3 gap-y-1 text-xs 2xl:grid-cols-2">
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">P95:</span>
              <span class="font-bold" :class="getThresholdColorClass(getCurrentTTFTThresholdLevel(ttftMetrics.p95_ms ?? null))">{{ ttftMetrics.p95_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">P90:</span>
              <span class="font-bold" :class="getThresholdColorClass(getCurrentTTFTThresholdLevel(ttftMetrics.p90_ms ?? null))">{{ ttftMetrics.p90_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">P50:</span>
              <span class="font-bold" :class="getThresholdColorClass(getCurrentTTFTThresholdLevel(ttftMetrics.p50_ms ?? null))">{{ ttftMetrics.p50_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">Avg:</span>
              <span class="font-bold" :class="getThresholdColorClass(getCurrentTTFTThresholdLevel(ttftMetrics.avg_ms ?? null))">{{ ttftMetrics.avg_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
            <div class="flex items-baseline gap-1 whitespace-nowrap">
              <span class="ops-dashboard-header__label">Max:</span>
              <span class="font-bold" :class="getThresholdColorClass(getCurrentTTFTThresholdLevel(ttftMetrics.max_ms ?? null))">{{ ttftMetrics.max_ms ?? '-' }}</span>
              <span class="ops-dashboard-header__eyebrow">ms</span>
            </div>
          </div>
        </div>

        <!-- Card 3: Request Errors -->
        <div class="ops-dashboard-header__metric-card" style="order: 3;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1">
              <span class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase">{{ t('admin.ops.requestErrors') }}</span>
              <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.errors')" />
            </div>
            <button v-if="!props.fullscreen" class="ops-dashboard-header__link text-[10px] font-bold" type="button" @click="openErrorDetails('request')">
              {{ t('admin.ops.requestDetails.details') }}
            </button>
          </div>
          <div class="mt-2 text-3xl font-black" :class="requestErrorDisplay.colorClass">
            {{ requestErrorDisplay.percent == null ? '-' : `${requestErrorDisplay.percent.toFixed(2)}%` }}
          </div>
          <div class="mt-3 space-y-1 text-xs">
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.errorCount') }}:</span>
              <span class="ops-dashboard-header__value font-bold">{{ formatNumber(overview.error_count_sla ?? 0) }}</span>
            </div>
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.businessLimited') }}:</span>
              <span class="ops-dashboard-header__value font-bold">{{ formatNumber(overview.business_limited_count ?? 0) }}</span>
            </div>
          </div>
        </div>

        <!-- Card 6: Upstream Errors -->
        <div class="ops-dashboard-header__metric-card" style="order: 6;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1">
              <span class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase">{{ t('admin.ops.upstreamErrors') }}</span>
              <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.upstreamErrors')" />
            </div>
            <button v-if="!props.fullscreen" class="ops-dashboard-header__link text-[10px] font-bold" type="button" @click="openErrorDetails('upstream')">
              {{ t('admin.ops.requestDetails.details') }}
            </button>
          </div>
          <div class="mt-2 text-3xl font-black" :class="upstreamErrorDisplay.colorClass">
            {{ upstreamErrorDisplay.percent == null ? '-' : `${upstreamErrorDisplay.percent.toFixed(2)}%` }}
          </div>
          <div class="mt-3 space-y-1 text-xs">
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">{{ t('admin.ops.errorCountExcl429529') }}:</span>
              <span class="ops-dashboard-header__value font-bold">{{ formatNumber(overview.upstream_error_count_excl_429_529 ?? 0) }}</span>
            </div>
            <div class="ops-dashboard-header__stat-row flex justify-between">
              <span class="ops-dashboard-header__label">429/529:</span>
              <span class="ops-dashboard-header__value font-bold">{{ formatNumber((overview.upstream_429_count ?? 0) + (overview.upstream_529_count ?? 0)) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Integrated: System health (cards) -->
    <div v-if="overview" class="ops-dashboard-header__system-section mt-2 pt-4">
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
        <!-- CPU -->
        <div class="ops-dashboard-header__system-card">
          <div class="flex items-center gap-1">
            <div class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase tracking-wider">CPU</div>
            <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.cpu')" />
          </div>
          <div class="mt-1 text-lg font-black" :class="cpuPercentClass">
            {{ cpuPercentValue == null ? '-' : `${cpuPercentValue.toFixed(1)}%` }}
          </div>
          <div v-if="!props.fullscreen" class="ops-dashboard-header__label mt-1 text-[10px]">
            {{ t('common.warning') }} 80% · {{ t('common.critical') }} 95%
          </div>
        </div>

        <!-- MEM -->
        <div class="ops-dashboard-header__system-card">
          <div class="flex items-center gap-1">
            <div class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase tracking-wider">{{ t('admin.ops.memory') }}</div>
            <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.memory')" />
          </div>
          <div class="mt-1 text-lg font-black" :class="memPercentClass">
            {{ memPercentValue == null ? '-' : `${memPercentValue.toFixed(1)}%` }}
          </div>
          <div v-if="!props.fullscreen" class="ops-dashboard-header__label mt-1 text-[10px]">
            {{
              systemMetrics?.memory_used_mb == null || systemMetrics?.memory_total_mb == null
                ? '-'
                : `${formatNumber(systemMetrics.memory_used_mb)} / ${formatNumber(systemMetrics.memory_total_mb)} MB`
            }}
          </div>
        </div>

        <!-- DB -->
        <div class="ops-dashboard-header__system-card">
          <div class="flex items-center gap-1">
            <div class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase tracking-wider">{{ t('admin.ops.db') }}</div>
            <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.db')" />
          </div>
          <div class="mt-1 text-lg font-black" :class="dbMiddleDisplay.className">
            {{ dbMiddleDisplay.label }}
          </div>
          <div v-if="!props.fullscreen" class="ops-dashboard-header__label mt-1 text-[10px]">
            {{ t('admin.ops.conns') }} {{ dbConnOpenValue ?? '-' }} / {{ dbMaxOpenConnsValue ?? '-' }}
            · {{ t('admin.ops.active') }} {{ dbConnActiveValue ?? '-' }}
            · {{ t('admin.ops.idle') }} {{ dbConnIdleValue ?? '-' }}
            <span v-if="dbConnWaitingValue != null"> · {{ t('admin.ops.waiting') }} {{ dbConnWaitingValue }} </span>
          </div>
        </div>

        <!-- Redis -->
        <div class="ops-dashboard-header__system-card">
          <div class="flex items-center gap-1">
            <div class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase tracking-wider">Redis</div>
            <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.redis')" />
          </div>
          <div class="mt-1 text-lg font-black" :class="redisMiddleDisplay.className">
            {{ redisMiddleDisplay.label }}
          </div>
          <div v-if="!props.fullscreen" class="ops-dashboard-header__label mt-1 text-[10px]">
            {{ t('admin.ops.conns') }} {{ redisConnTotalValue ?? '-' }} / {{ redisPoolSizeValue ?? '-' }}
            <span v-if="redisConnActiveValue != null"> · {{ t('admin.ops.active') }} {{ redisConnActiveValue }} </span>
            <span v-if="redisConnIdleValue != null"> · {{ t('admin.ops.idle') }} {{ redisConnIdleValue }} </span>
          </div>
        </div>

        <!-- Goroutines -->
        <div class="ops-dashboard-header__system-card">
          <div class="flex items-center gap-1">
            <div class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase tracking-wider">{{ t('admin.ops.goroutines') }}</div>
            <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.goroutines')" />
          </div>
          <div class="mt-1 text-lg font-black" :class="goroutineDisplay.className">
            {{ goroutineDisplay.label }}
          </div>
          <div v-if="!props.fullscreen" class="ops-dashboard-header__label mt-1 text-[10px]">
            {{ t('admin.ops.current') }} <span class="font-mono">{{ goroutineCountValue ?? '-' }}</span>
            · {{ t('common.warning') }} <span class="font-mono">{{ goroutinesWarnThreshold }}</span>
            · {{ t('common.critical') }} <span class="font-mono">{{ goroutinesCriticalThreshold }}</span>
            <span v-if="systemMetrics?.concurrency_queue_depth != null">
              · {{ t('admin.ops.queue') }} <span class="font-mono">{{ systemMetrics.concurrency_queue_depth }}</span>
            </span>
          </div>
        </div>

        <!-- Jobs -->
        <div class="ops-dashboard-header__system-card">
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-1">
              <div class="ops-dashboard-header__eyebrow text-[10px] font-bold uppercase tracking-wider">{{ t('admin.ops.jobs') }}</div>
              <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.jobs')" />
            </div>
            <button v-if="!props.fullscreen" class="ops-dashboard-header__link text-[10px] font-bold" type="button" @click="openJobsDetails">
              {{ t('admin.ops.requestDetails.details') }}
            </button>
          </div>

          <div class="mt-1 text-lg font-black" :class="jobsDisplay.className">
            {{ jobsDisplay.label }}
          </div>

          <div v-if="!props.fullscreen" class="ops-dashboard-header__label mt-1 text-[10px]">
            {{ t('common.total') }} <span class="font-mono">{{ jobHeartbeats.length }}</span>
            · {{ t('common.warning') }} <span class="font-mono">{{ jobsDisplay.warnCount }}</span>
          </div>
        </div>
      </div>
    </div>

    <BaseDialog :show="showJobsDetails" :title="t('admin.ops.jobs')" width="wide" @close="showJobsDetails = false">
      <div v-if="!jobHeartbeats.length" class="ops-dashboard-header__label text-sm">
        {{ t('admin.ops.noData') }}
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="hb in jobHeartbeats"
          :key="hb.job_name"
          class="ops-dashboard-header__job-card"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="ops-dashboard-header__value truncate text-sm font-semibold">{{ hb.job_name }}</div>
            <div class="ops-dashboard-header__label flex items-center gap-3 text-xs">
              <span v-if="hb.last_duration_ms != null" class="font-mono">{{ hb.last_duration_ms }}ms</span>
              <span>{{ formatTimeShort(hb.updated_at) }}</span>
            </div>
          </div>

          <div class="ops-dashboard-header__metric-list mt-2 grid grid-cols-1 gap-2 text-xs sm:grid-cols-2">
            <div>
              {{ t('admin.ops.lastSuccess') }} <span class="font-mono">{{ formatTimeShort(hb.last_success_at) }}</span>
            </div>
            <div>
              {{ t('admin.ops.lastError') }} <span class="font-mono">{{ formatTimeShort(hb.last_error_at) }}</span>
            </div>
            <div>
              {{ t('admin.ops.result') }} <span class="font-mono">{{ hb.last_result || '-' }}</span>
            </div>
          </div>

          <div
            v-if="hb.last_error"
            class="ops-dashboard-header__job-error mt-3 text-xs"
          >
            {{ hb.last_error }}
          </div>
        </div>
      </div>
    </BaseDialog>

    <!-- Custom Time Range Dialog -->
    <BaseDialog :show="showCustomTimeRangeDialog" :title="t('admin.ops.timeRange.custom')" width="narrow" @close="handleCustomTimeRangeCancel">
      <div class="space-y-4">
        <div>
          <label class="ops-dashboard-header__dialog-label mb-1 block text-sm font-medium">
            {{ t('admin.ops.customTimeRange.startTime') }}
          </label>
          <input
            v-model="customStartTimeInput"
            type="datetime-local"
            class="ops-dashboard-header__dialog-input w-full text-sm"
          />
        </div>
        <div>
          <label class="ops-dashboard-header__dialog-label mb-1 block text-sm font-medium">
            {{ t('admin.ops.customTimeRange.endTime') }}
          </label>
          <input
            v-model="customEndTimeInput"
            type="datetime-local"
            class="ops-dashboard-header__dialog-input w-full text-sm"
          />
        </div>
        <div class="flex justify-end gap-3 pt-2">
          <button
            type="button"
            class="ops-dashboard-header__dialog-button ops-dashboard-header__dialog-button--secondary text-sm font-medium"
            @click="handleCustomTimeRangeCancel"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            class="ops-dashboard-header__dialog-button ops-dashboard-header__dialog-button--primary text-sm font-medium"
            @click="handleCustomTimeRangeConfirm"
          >
            {{ t('common.confirm') }}
          </button>
        </div>
      </div>
    </BaseDialog>
  </div>
</template>

<style scoped>
.ops-dashboard-header__surface {
  border-radius: calc(var(--theme-surface-radius) * 2);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.ops-dashboard-header__surface--default {
  padding: calc(var(--theme-ops-card-padding) * 1.33);
}

.ops-dashboard-header__surface--fullscreen {
  padding: calc(var(--theme-ops-card-padding) * 1.78);
}

.ops-dashboard-header__divider,
.ops-dashboard-header__popover-title,
.ops-dashboard-header__popover-footer {
  border-color: color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.ops-dashboard-header__divider {
  border-bottom-width: 1px;
}

.ops-dashboard-header__title,
.ops-dashboard-header__text {
  color: var(--theme-page-text);
}

.ops-dashboard-header__title-icon,
.ops-dashboard-header__realtime-dot,
.ops-dashboard-header__realtime-ping,
.ops-dashboard-header__tone--info {
  color: rgb(var(--theme-info-rgb));
}

.ops-dashboard-header__meta,
.ops-dashboard-header__muted,
.ops-dashboard-header__popover-footer {
  color: var(--theme-page-muted);
}

.ops-dashboard-header__eyebrow {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
}

.ops-dashboard-header__label {
  color: color-mix(in srgb, var(--theme-page-muted) 64%, transparent);
  font-weight: 500;
}

.ops-dashboard-header__value {
  color: var(--theme-page-text);
}

.ops-dashboard-header__metric-list {
  color: color-mix(in srgb, var(--theme-page-text) 68%, var(--theme-page-muted));
}

.ops-dashboard-header__status-dot--loading {
  background: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.ops-dashboard-header__status-dot--ready {
  background: rgb(var(--theme-success-rgb));
}

.ops-dashboard-header__toolbar-divider {
  width: max(var(--theme-card-border-width), 1px);
  height: calc(var(--theme-button-padding-y) * 0.4 + 0.75rem);
  background: color-mix(in srgb, var(--theme-page-border) 82%, transparent);
}

.ops-dashboard-header__select {
  width: 100%;
}

.ops-dashboard-header__realtime {
  padding-block: calc(var(--theme-ops-panel-padding) * 0.5);
}

@media (min-width: 640px) {
  .ops-dashboard-header__select--platform {
    width: calc(var(--theme-ops-table-min-width) * 0.175);
  }

  .ops-dashboard-header__select--group {
    width: calc(var(--theme-ops-table-min-width) * 0.2);
  }

  .ops-dashboard-header__select--time-range {
    width: calc(var(--theme-ops-table-min-width) * 0.1875);
  }

  .ops-dashboard-header__select--query-mode {
    width: calc(var(--theme-ops-table-min-width) * 0.2125);
  }
}

.ops-dashboard-header__icon-button,
.ops-dashboard-header__action-button--neutral {
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.ops-dashboard-header__action-button {
  padding-inline: calc(var(--theme-button-padding-x) * 0.75);
}

.ops-dashboard-header__icon-button:hover,
.ops-dashboard-header__action-button--neutral:hover {
  background: color-mix(in srgb, var(--theme-page-border) 68%, var(--theme-surface));
  color: var(--theme-page-text);
}

.ops-dashboard-header__action-button--info {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.ops-dashboard-header__action-button--info:hover {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 16%, var(--theme-surface));
}

.ops-dashboard-header__panel {
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.ops-dashboard-header__panel--default {
  padding: var(--theme-ops-panel-padding);
}

.ops-dashboard-header__panel--fullscreen {
  padding: calc(var(--theme-ops-panel-padding) * 1.5);
}

.ops-dashboard-header__metric-card,
.ops-dashboard-header__system-card {
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface));
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 64%, transparent);
}

.ops-dashboard-header__metric-card {
  padding: var(--theme-ops-panel-padding);
}

.ops-dashboard-header__metric-card {
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--theme-surface-contrast) 32%, transparent);
}

.ops-dashboard-header__system-card {
  padding: calc(var(--theme-ops-panel-padding) * 0.75);
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.ops-dashboard-header__system-section {
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.ops-dashboard-header__link {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.ops-dashboard-header__link:hover {
  color: rgb(var(--theme-info-rgb));
}

.ops-dashboard-header__stat-row {
  gap: 0.75rem;
}

.ops-dashboard-header__progress-track {
  display: none;
}

.ops-dashboard-header__pulse-line {
  stroke: rgb(var(--theme-info-rgb));
}

.ops-dashboard-header__runtime-strip {
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 72%, transparent);
  padding-top: 0.75rem;
}

.ops-dashboard-header__runtime-chip {
  border-radius: calc(var(--theme-button-radius) * 0.85);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 60%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  padding: 0.6rem 0.7rem;
}

.ops-dashboard-header__health:hover {
  background: color-mix(in srgb, var(--theme-page-backdrop) 86%, transparent);
}

.ops-dashboard-header__health {
  border-radius: var(--theme-select-panel-radius);
  padding-block: calc(var(--theme-ops-panel-padding) * 0.5);
}

@media (min-width: 768px) {
  .ops-dashboard-header__health {
    border-right: 1px solid color-mix(in srgb, var(--theme-page-border) 78%, transparent);
  }
}

.ops-dashboard-header__popover {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.ops-dashboard-header__ring-track {
  color: color-mix(in srgb, var(--theme-page-border) 82%, transparent);
}

.ops-dashboard-header__realtime-ping,
.ops-dashboard-header__realtime-dot {
  background: rgb(var(--theme-info-rgb));
}

.ops-dashboard-header__window--active {
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.ops-dashboard-header__window {
  padding:
    calc(var(--theme-button-padding-y) * 0.25)
    calc(var(--theme-button-padding-x) * 0.32);
  border-radius: calc(var(--theme-button-radius) * 0.75);
}

.ops-dashboard-header__window--idle {
  background: color-mix(in srgb, var(--theme-page-border) 84%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.ops-dashboard-header__window--idle:hover {
  background: color-mix(in srgb, var(--theme-page-border) 94%, var(--theme-surface));
}

.ops-dashboard-header__job-card {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: var(--theme-surface);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.ops-dashboard-header__job-error {
  padding: calc(var(--theme-ops-panel-padding) * 0.5);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, rgb(var(--theme-brand-rose-rgb)) 11%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-brand-rose-rgb)) 86%, var(--theme-page-text));
}

.ops-dashboard-header__dialog-label {
  color: var(--theme-page-text);
}

.ops-dashboard-header__dialog-input {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.75);
  border-radius: var(--theme-button-radius);
  border: 1px solid color-mix(in srgb, var(--theme-input-border) 78%, transparent);
  background: var(--theme-input-bg);
  color: var(--theme-page-text);
}

.ops-dashboard-header__dialog-input:focus {
  outline: none;
  border-color: color-mix(in srgb, var(--theme-accent) 72%, transparent);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--theme-accent) 36%, transparent);
}

.ops-dashboard-header__dialog-button {
  padding: calc(var(--theme-button-padding-y) * 0.8) var(--theme-button-padding-x);
  border-radius: var(--theme-button-radius);
  transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
}

.ops-dashboard-header__dialog-button--secondary {
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
  color: var(--theme-page-text);
}

.ops-dashboard-header__dialog-button--secondary:hover {
  background: color-mix(in srgb, var(--theme-page-border) 66%, var(--theme-surface));
}

.ops-dashboard-header__dialog-button--primary {
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.ops-dashboard-header__dialog-button--primary:hover {
  background: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-accent-strong));
}

.ops-dashboard-header__indicator--healthy,
.ops-dashboard-header__tone--healthy {
  background: transparent;
  color: rgb(var(--theme-success-rgb));
  border: none;
}

.ops-dashboard-header__indicator--warning,
.ops-dashboard-header__tone--warning {
  background: transparent;
  color: rgb(var(--theme-warning-rgb));
  border: none;
}

.ops-dashboard-header__indicator--critical,
.ops-dashboard-header__tone--critical {
  background: transparent;
  color: rgb(var(--theme-danger-rgb));
  border: none;
}

.ops-dashboard-header__tone--default {
  color: var(--theme-page-text);
}

.ops-dashboard-header__tone--muted {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}
</style>
