<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  opsAPI,
  type OpsAccountAvailabilityStatsResponse,
  type OpsConcurrencyStatsResponse,
  type OpsUserConcurrencyStatsResponse,
  type RuntimeObservabilitySnapshot
} from '@/api/admin/ops'
import { resolveRequestErrorMessage } from '@/utils/requestError'

interface Props {
  platformFilter?: string
  groupIdFilter?: number | null
  refreshToken: number
}

const props = withDefaults(defineProps<Props>(), {
  platformFilter: '',
  groupIdFilter: null
})

const { t } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const concurrency = ref<OpsConcurrencyStatsResponse | null>(null)
const availability = ref<OpsAccountAvailabilityStatsResponse | null>(null)
const userConcurrency = ref<OpsUserConcurrencyStatsResponse | null>(null)
let loadSequence = 0

// 用户视图开关
const showByUser = ref(false)

const realtimeEnabled = computed(() => {
  return (concurrency.value?.enabled ?? true) && (availability.value?.enabled ?? true)
})

function safeNumber(n: unknown): number {
  return typeof n === 'number' && Number.isFinite(n) ? n : 0
}

type RuntimeTone = 'healthy' | 'warning' | 'critical'

interface RuntimeHealthItem {
  key: string
  label: string
  value: string
  tone: RuntimeTone
}

interface RuntimeHeadline {
  tone: RuntimeTone
  title: string
  detail: string
}

const runtimeObservability = computed<RuntimeObservabilitySnapshot | null>(() => {
  if (showByUser.value) {
    return userConcurrency.value?.runtime_observability ?? null
  }
  return concurrency.value?.runtime_observability ?? availability.value?.runtime_observability ?? null
})

function formatRuntimePercent(value: number | null | undefined): string {
  return typeof value === 'number' && Number.isFinite(value) ? `${(value * 100).toFixed(1)}%` : '-'
}

function formatRuntimeFixed(value: number | null | undefined, digits = 1): string {
  return typeof value === 'number' && Number.isFinite(value) ? value.toFixed(digits) : '-'
}

function runtimeToneByRate(value: number | null | undefined, warn: number, critical: number): RuntimeTone {
  if (typeof value !== 'number' || !Number.isFinite(value)) return 'healthy'
  if (value < critical) return 'critical'
  if (value < warn) return 'warning'
  return 'healthy'
}

function runtimeToneByReverseRate(value: number | null | undefined, warn: number, critical: number): RuntimeTone {
  if (typeof value !== 'number' || !Number.isFinite(value)) return 'healthy'
  if (value > critical) return 'critical'
  if (value > warn) return 'warning'
  return 'healthy'
}

const runtimeHealthItems = computed<RuntimeHealthItem[]>(() => {
  const summary = runtimeObservability.value?.summary
  const raw = runtimeObservability.value?.scheduling_runtime_kernel
  const idempotency = runtimeObservability.value?.summary?.idempotency
  const hasWaitPlan = (raw?.runtime_wait_plan_attempts ?? 0) > 0

  return [
    {
      key: 'page_density',
      label: t('admin.ops.runtimeObservability.pageDensity'),
      value: formatRuntimeFixed(summary?.scheduling_runtime_kernel?.avg_fetched_accounts_per_page),
      tone: runtimeToneByRate(
        summary?.scheduling_runtime_kernel?.avg_fetched_accounts_per_page,
        8,
        3
      )
    },
    {
      key: 'acquire_success',
      label: t('admin.ops.runtimeObservability.acquireSuccess'),
      value: formatRuntimePercent(summary?.scheduling_runtime_kernel?.acquire_success_rate),
      tone: runtimeToneByRate(summary?.scheduling_runtime_kernel?.acquire_success_rate, 0.75, 0.5)
    },
    {
      key: 'wait_plan_success',
      label: t('admin.ops.runtimeObservability.waitPlanSuccess'),
      value: hasWaitPlan
        ? formatRuntimePercent(summary?.scheduling_runtime_kernel?.wait_plan_success_rate)
        : t('admin.ops.runtimeObservability.notTriggered'),
      tone: hasWaitPlan
        ? runtimeToneByRate(summary?.scheduling_runtime_kernel?.wait_plan_success_rate, 0.6, 0.35)
        : 'healthy'
    },
    {
      key: 'idempotency_avg',
      label: t('admin.ops.runtimeObservability.idempotencyAvg'),
      value: `${formatRuntimeFixed(idempotency?.avg_processing_duration_ms)}ms`,
      tone: runtimeToneByReverseRate(idempotency?.avg_processing_duration_ms, 80, 250)
    }
  ]
})

const runtimeHeadline = computed<RuntimeHeadline | null>(() => {
  const summary = runtimeObservability.value?.summary
  const raw = runtimeObservability.value?.scheduling_runtime_kernel
  if (!summary || !raw) {
    return null
  }

  const acquireTone = runtimeToneByRate(summary.scheduling_runtime_kernel.acquire_success_rate, 0.75, 0.5)
  if (acquireTone !== 'healthy') {
    return {
      tone: acquireTone,
      title: t('admin.ops.runtimeObservability.acquireRisk'),
      detail: t('admin.ops.runtimeObservability.acquireRiskHint')
    }
  }

  const densityTone = runtimeToneByRate(summary.scheduling_runtime_kernel.avg_fetched_accounts_per_page, 8, 3)
  if ((raw.index_page_fetches ?? 0) > 0 && densityTone !== 'healthy') {
    return {
      tone: densityTone,
      title: t('admin.ops.runtimeObservability.pageDensityRisk'),
      detail: t('admin.ops.runtimeObservability.pageDensityRiskHint')
    }
  }

  const hasWaitPlan = (raw.runtime_wait_plan_attempts ?? 0) > 0
  const waitPlanTone = hasWaitPlan
    ? runtimeToneByRate(summary.scheduling_runtime_kernel.wait_plan_success_rate, 0.6, 0.35)
    : 'healthy'
  if (waitPlanTone !== 'healthy') {
    return {
      tone: waitPlanTone,
      title: t('admin.ops.runtimeObservability.waitPlanRisk'),
      detail: t('admin.ops.runtimeObservability.waitPlanRiskHint')
    }
  }

  const idempotencyTone = runtimeToneByReverseRate(summary.idempotency.avg_processing_duration_ms, 80, 250)
  if (idempotencyTone !== 'healthy') {
    return {
      tone: idempotencyTone,
      title: t('admin.ops.runtimeObservability.idempotencyRisk'),
      detail: t('admin.ops.runtimeObservability.idempotencyRiskHint')
    }
  }

  return {
    tone: 'healthy',
    title: t('admin.ops.runtimeObservability.healthyTitle'),
    detail: t('admin.ops.runtimeObservability.healthyHint')
  }
})

// 计算显示维度
const displayDimension = computed<'platform' | 'group' | 'account' | 'user'>(() => {
  if (showByUser.value) {
    return 'user'
  }
  if (typeof props.groupIdFilter === 'number' && props.groupIdFilter > 0) {
    return 'account'
  }
  if (props.platformFilter) {
    return 'group'
  }
  return 'platform'
})

// 平台/分组汇总行数据
interface SummaryRow {
  key: string
  name: string
  platform?: string
  // 账号统计
  total_accounts: number
  available_accounts: number
  rate_limited_accounts: number
  error_accounts: number
  // 并发统计
  total_concurrency: number
  used_concurrency: number
  waiting_in_queue: number
  // 计算字段
  availability_percentage: number
  concurrency_percentage: number
}

// 账号详细行数据
interface AccountRow {
  key: string
  name: string
  platform: string
  group_name: string
  // 并发
  current_in_use: number
  max_capacity: number
  waiting_in_queue: number
  load_percentage: number
  // 状态
  is_available: boolean
  is_rate_limited: boolean
  rate_limit_remaining_sec?: number
  is_overloaded: boolean
  overload_remaining_sec?: number
  has_error: boolean
  error_message?: string
}

// 用户行数据
interface UserRow {
  key: string
  user_id: number
  user_email: string
  username: string
  current_in_use: number
  max_capacity: number
  waiting_in_queue: number
  load_percentage: number
}

// 平台维度汇总
const platformRows = computed((): SummaryRow[] => {
  const concStats = concurrency.value?.platform || {}
  const availStats = availability.value?.platform || {}

  const platforms = new Set([...Object.keys(concStats), ...Object.keys(availStats)])

  return Array.from(platforms).map(platform => {
    const conc = concStats[platform] || {}
    const avail = availStats[platform] || {}

    const totalAccounts = safeNumber(avail.total_accounts)
    const availableAccounts = safeNumber(avail.available_count)
    const totalConcurrency = safeNumber(conc.max_capacity)
    const usedConcurrency = safeNumber(conc.current_in_use)

    return {
      key: platform,
      name: platform.toUpperCase(),
      total_accounts: totalAccounts,
      available_accounts: availableAccounts,
      rate_limited_accounts: safeNumber(avail.rate_limit_count),

      error_accounts: safeNumber(avail.error_count),
      total_concurrency: totalConcurrency,
      used_concurrency: usedConcurrency,
      waiting_in_queue: safeNumber(conc.waiting_in_queue),
      availability_percentage: totalAccounts > 0 ? Math.round((availableAccounts / totalAccounts) * 100) : 0,
      concurrency_percentage: totalConcurrency > 0 ? Math.round((usedConcurrency / totalConcurrency) * 100) : 0
    }
  }).sort((a, b) => b.concurrency_percentage - a.concurrency_percentage)
})

// 分组维度汇总
const groupRows = computed((): SummaryRow[] => {
  const concStats = concurrency.value?.group || {}
  const availStats = availability.value?.group || {}

  const groupIds = new Set([...Object.keys(concStats), ...Object.keys(availStats)])

  const rows = Array.from(groupIds)
    .map(gid => {
      const conc = concStats[gid] || {}
      const avail = availStats[gid] || {}

      // 只显示匹配的平台
      if (props.platformFilter && conc.platform !== props.platformFilter && avail.platform !== props.platformFilter) {
        return null
      }

      const totalAccounts = safeNumber(avail.total_accounts)
      const availableAccounts = safeNumber(avail.available_count)
      const totalConcurrency = safeNumber(conc.max_capacity)
      const usedConcurrency = safeNumber(conc.current_in_use)

      return {
        key: gid,
        name: String(conc.group_name || avail.group_name || `Group ${gid}`),
        platform: String(conc.platform || avail.platform || ''),
        total_accounts: totalAccounts,
        available_accounts: availableAccounts,
        rate_limited_accounts: safeNumber(avail.rate_limit_count),
  
        error_accounts: safeNumber(avail.error_count),
        total_concurrency: totalConcurrency,
        used_concurrency: usedConcurrency,
        waiting_in_queue: safeNumber(conc.waiting_in_queue),
        availability_percentage: totalAccounts > 0 ? Math.round((availableAccounts / totalAccounts) * 100) : 0,
        concurrency_percentage: totalConcurrency > 0 ? Math.round((usedConcurrency / totalConcurrency) * 100) : 0
      }
    })
    .filter((row): row is NonNullable<typeof row> => row !== null)

  return rows.sort((a, b) => b.concurrency_percentage - a.concurrency_percentage)
})

// 账号维度详细
const accountRows = computed((): AccountRow[] => {
  const concStats = concurrency.value?.account || {}
  const availStats = availability.value?.account || {}

  const accountIds = new Set([...Object.keys(concStats), ...Object.keys(availStats)])

  const rows = Array.from(accountIds)
    .map(aid => {
      const conc = concStats[aid] || {}
      const avail = availStats[aid] || {}

      // 只显示匹配的分组
      if (typeof props.groupIdFilter === 'number' && props.groupIdFilter > 0) {
        if (conc.group_id !== props.groupIdFilter && avail.group_id !== props.groupIdFilter) {
          return null
        }
      }

      return {
        key: aid,
        name: String(conc.account_name || avail.account_name || `Account ${aid}`),
        platform: String(conc.platform || avail.platform || ''),
        group_name: String(conc.group_name || avail.group_name || ''),
        current_in_use: safeNumber(conc.current_in_use),
        max_capacity: safeNumber(conc.max_capacity),
        waiting_in_queue: safeNumber(conc.waiting_in_queue),
        load_percentage: safeNumber(conc.load_percentage),
        is_available: avail.is_available || false,
        is_rate_limited: avail.is_rate_limited || false,
        rate_limit_remaining_sec: avail.rate_limit_remaining_sec,
        is_overloaded: avail.is_overloaded || false,
        overload_remaining_sec: avail.overload_remaining_sec,
        has_error: avail.has_error || false,
        error_message: avail.error_message || ''
      }
    })
    .filter((row): row is NonNullable<typeof row> => row !== null)

  return rows.sort((a, b) => {
    // 优先显示异常账号
    if (a.has_error !== b.has_error) return a.has_error ? -1 : 1
    if (a.is_rate_limited !== b.is_rate_limited) return a.is_rate_limited ? -1 : 1
    // 然后按负载排序
    return b.load_percentage - a.load_percentage
  })
})

// 用户维度详细
const userRows = computed((): UserRow[] => {
  const userStats = userConcurrency.value?.user || {}

  return Object.keys(userStats)
    .map(uid => {
      const u = userStats[uid] || {}
      return {
        key: uid,
        user_id: safeNumber(u.user_id),
        user_email: u.user_email || `User ${uid}`,
        username: u.username || '',
        current_in_use: safeNumber(u.current_in_use),
        max_capacity: safeNumber(u.max_capacity),
        waiting_in_queue: safeNumber(u.waiting_in_queue),
        load_percentage: safeNumber(u.load_percentage)
      }
    })
    .sort((a, b) => b.current_in_use - a.current_in_use || b.load_percentage - a.load_percentage)
})

// 根据维度选择数据
const displayRows = computed(() => {
  if (displayDimension.value === 'user') return userRows.value
  if (displayDimension.value === 'account') return accountRows.value
  if (displayDimension.value === 'group') return groupRows.value
  return platformRows.value
})

const displayTitle = computed(() => {
  if (displayDimension.value === 'user') return t('admin.ops.concurrency.byUser')
  if (displayDimension.value === 'account') return t('admin.ops.concurrency.byAccount')
  if (displayDimension.value === 'group') return t('admin.ops.concurrency.byGroup')
  return t('admin.ops.concurrency.byPlatform')
})

type ConcurrencyTone = 'danger' | 'info' | 'neutral' | 'success' | 'warning'

async function loadData() {
  const requestSequence = ++loadSequence
  loading.value = true
  errorMessage.value = ''
  try {
    if (showByUser.value) {
      // 用户视图模式只加载用户并发数据
      const userData = await opsAPI.getUserConcurrencyStats()
      if (requestSequence !== loadSequence) return
      userConcurrency.value = userData
    } else {
      // 常规模式加载账号/平台/分组数据
      const [concData, availData] = await Promise.all([
        opsAPI.getConcurrencyStats(props.platformFilter, props.groupIdFilter),
        opsAPI.getAccountAvailabilityStats(props.platformFilter, props.groupIdFilter)
      ])
      if (requestSequence !== loadSequence) return
      concurrency.value = concData
      availability.value = availData
    }
  } catch (err: unknown) {
    if (requestSequence !== loadSequence) return
    console.error('[OpsConcurrencyCard] Failed to load data', err)
    errorMessage.value = resolveRequestErrorMessage(err, t('admin.ops.concurrency.loadFailed'))
  } finally {
    if (requestSequence === loadSequence) {
      loading.value = false
    }
  }
}

// 刷新节奏由父组件统一控制（OpsDashboard Header 的刷新状态/倒计时）
watch(
  () => props.refreshToken,
  () => {
    if (!realtimeEnabled.value) return
    loadData()
  }
)

// 切换用户视图时重新加载数据
watch(
  () => showByUser.value,
  () => {
    loadData()
  }
)

function joinClassNames(classNames: Array<string | false | null | undefined>): string {
  return classNames.filter(Boolean).join(' ')
}

function resolveLoadTone(loadPct: number): ConcurrencyTone {
  if (loadPct >= 90) return 'danger'
  if (loadPct >= 70) return 'warning'
  if (loadPct >= 50) return 'info'
  return 'success'
}

function getLoadBarClass(loadPct: number): string {
  const tone = resolveLoadTone(loadPct)
  return joinClassNames([
    'ops-concurrency-card__bar',
    tone === 'danger' && 'ops-concurrency-card__bar--danger',
    tone === 'info' && 'ops-concurrency-card__bar--info',
    tone === 'success' && 'ops-concurrency-card__bar--success',
    tone === 'warning' && 'ops-concurrency-card__bar--warning'
  ])
}

function getLoadBarStyle(loadPct: number): string {
  return `width: ${Math.min(100, Math.max(0, loadPct))}%`
}

function getLoadTextClass(loadPct: number): string {
  const tone = resolveLoadTone(loadPct)
  return joinClassNames([
    'ops-concurrency-card__load-text',
    tone === 'danger' && 'ops-concurrency-card__load-text--danger',
    tone === 'info' && 'ops-concurrency-card__load-text--info',
    tone === 'success' && 'ops-concurrency-card__load-text--success',
    tone === 'warning' && 'ops-concurrency-card__load-text--warning'
  ])
}

function getViewToggleClasses(isActive: boolean): string {
  return joinClassNames([
    'ops-concurrency-card__view-toggle',
    isActive
      ? 'ops-concurrency-card__view-toggle--active'
      : 'ops-concurrency-card__view-toggle--idle'
  ])
}

function getRuntimeHeadlineClasses(tone: RuntimeTone): string {
  return joinClassNames([
    'ops-concurrency-card__runtime-headline',
    tone === 'critical' && 'ops-concurrency-card__runtime-headline--critical',
    tone === 'warning' && 'ops-concurrency-card__runtime-headline--warning',
    tone === 'healthy' && 'ops-concurrency-card__runtime-headline--healthy'
  ])
}

function getRuntimeMetricClasses(tone: RuntimeTone): string {
  return joinClassNames([
    'ops-concurrency-card__runtime-metric',
    tone === 'critical' && 'ops-concurrency-card__runtime-metric--critical',
    tone === 'warning' && 'ops-concurrency-card__runtime-metric--warning',
    tone === 'healthy' && 'ops-concurrency-card__runtime-metric--healthy'
  ])
}

function getStatusChipClasses(tone: ConcurrencyTone): string {
  return joinClassNames([
    'ops-concurrency-card__status-chip theme-chip theme-chip--compact',
    tone === 'danger' && 'theme-chip--danger',
    tone === 'info' && 'theme-chip--info',
    tone === 'neutral' && 'theme-chip--neutral',
    tone === 'success' && 'theme-chip--success',
    tone === 'warning' && 'theme-chip--warning'
  ])
}

function getAccountStatusClasses(row: AccountRow): string {
  if (row.is_available) return getStatusChipClasses('success')
  if (row.is_rate_limited) return getStatusChipClasses('warning')
  if (row.is_overloaded) return getStatusChipClasses('danger')
  if (row.has_error) return getStatusChipClasses('danger')
  return getStatusChipClasses('neutral')
}

function formatDuration(seconds: number): string {
  if (seconds <= 0) return '0s'
  if (seconds < 60) return `${Math.round(seconds)}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  return `${hours}h`
}


watch(
  () => realtimeEnabled.value,
  async (enabled) => {
    if (enabled) {
      await loadData()
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="ops-concurrency-card">
    <!-- 头部 -->
    <div class="ops-concurrency-card__header">
      <h3 class="ops-concurrency-card__title">
        <svg class="ops-concurrency-card__title-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        {{ t('admin.ops.concurrency.title') }}
      </h3>
      <div class="ops-concurrency-card__header-actions">
        <!-- 用户视图切换按钮 -->
        <button
          :class="getViewToggleClasses(showByUser)"
          :title="showByUser ? t('admin.ops.concurrency.switchToPlatform') : t('admin.ops.concurrency.switchToUser')"
          @click="showByUser = !showByUser"
        >
          <svg class="ops-concurrency-card__icon-button-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        </button>
        <!-- 刷新按钮 -->
        <button
          class="ops-concurrency-card__refresh"
          :disabled="loading"
          :title="t('common.refresh')"
          @click="loadData"
        >
          <svg class="ops-concurrency-card__refresh-icon" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMessage" class="ops-concurrency-card__error">
      {{ errorMessage }}
    </div>

    <!-- 禁用状态 -->
    <div
      v-if="!realtimeEnabled"
      class="ops-concurrency-card__empty ops-concurrency-card__empty--disabled"
    >
      {{ t('admin.ops.concurrency.disabledHint') }}
    </div>

    <!-- 数据展示区域 -->
    <div v-else class="ops-concurrency-card__shell">
      <!-- 维度标题栏 -->
      <div class="ops-concurrency-card__shell-head">
        <span class="ops-concurrency-card__subtitle ops-concurrency-card__subtitle--label">
          {{ displayTitle }}
        </span>
        <span class="ops-concurrency-card__subtitle ops-concurrency-card__subtitle--count">
          {{ t('admin.ops.concurrency.totalRows', { count: displayRows.length }) }}
        </span>
      </div>

      <div v-if="runtimeHeadline" class="ops-concurrency-card__runtime">
        <div :class="getRuntimeHeadlineClasses(runtimeHeadline.tone)">
          <div class="ops-concurrency-card__runtime-copy">
            <div class="ops-concurrency-card__runtime-title">
              {{ runtimeHeadline.title }}
            </div>
            <div class="ops-concurrency-card__runtime-detail">
              {{ runtimeHeadline.detail }}
            </div>
          </div>
          <div class="ops-concurrency-card__runtime-probes">
            <div class="ops-concurrency-card__subtitle ops-concurrency-card__subtitle--micro-label">
              {{ t('admin.ops.runtimeObservability.runtimeProbes') }}
            </div>
            <div class="ops-concurrency-card__primary ops-concurrency-card__primary--hero">
              {{ runtimeObservability?.summary?.scheduling_runtime_kernel?.total_runtime_probes ?? 0 }}
            </div>
          </div>
        </div>

        <div class="ops-concurrency-card__runtime-grid">
          <div
            v-for="item in runtimeHealthItems"
            :key="item.key"
            :class="getRuntimeMetricClasses(item.tone)"
          >
            <div class="ops-concurrency-card__subtitle ops-concurrency-card__subtitle--micro-label">
              {{ item.label }}
            </div>
            <div class="ops-concurrency-card__primary ops-concurrency-card__primary--metric">
              {{ item.value }}
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="displayRows.length === 0" class="ops-concurrency-card__empty">
        {{ t('admin.ops.concurrency.empty') }}
      </div>

      <!-- 用户视图 -->
      <div v-else-if="displayDimension === 'user'" class="ops-concurrency-card__list custom-scrollbar">
        <div v-for="row in (displayRows as UserRow[])" :key="row.key" class="ops-concurrency-card__row ops-concurrency-card__row--compact">
          <!-- 用户信息和并发 -->
          <div class="ops-concurrency-card__row-head">
            <div class="ops-concurrency-card__row-identity">
              <span class="ops-concurrency-card__primary ops-concurrency-card__primary--compact ops-concurrency-card__truncate" :title="row.username || row.user_email">
                {{ row.username || row.user_email }}
              </span>
              <span v-if="row.username" class="ops-concurrency-card__subtitle ops-concurrency-card__subtitle--compact ops-concurrency-card__truncate" :title="row.user_email">
                {{ row.user_email }}
              </span>
            </div>
            <div class="ops-concurrency-card__row-metrics">
              <span class="ops-concurrency-card__primary ops-concurrency-card__primary--mono"> {{ row.current_in_use }}/{{ row.max_capacity }} </span>
              <span :class="getLoadTextClass(row.load_percentage)"> {{ Math.round(row.load_percentage) }}% </span>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="ops-concurrency-card__track">
            <div :class="getLoadBarClass(row.load_percentage)" :style="getLoadBarStyle(row.load_percentage)"></div>
          </div>

          <!-- 等待队列 -->
          <div v-if="row.waiting_in_queue > 0" class="ops-concurrency-card__status-row ops-concurrency-card__status-row--end">
            <span class="ops-concurrency-card__status-chip theme-chip theme-chip--brand-purple theme-chip--compact">
              {{ t('admin.ops.concurrency.queued', { count: row.waiting_in_queue }) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 汇总视图（平台/分组） -->
      <div v-else-if="displayDimension === 'platform' || displayDimension === 'group'" class="ops-concurrency-card__list custom-scrollbar">
        <div v-for="row in (displayRows as SummaryRow[])" :key="row.key" class="ops-concurrency-card__row ops-concurrency-card__row--regular">
          <!-- 标题行 -->
          <div class="ops-concurrency-card__row-head ops-concurrency-card__row-head--regular">
            <div class="ops-concurrency-card__row-title-group">
              <div class="ops-concurrency-card__primary ops-concurrency-card__primary--compact ops-concurrency-card__truncate" :title="row.name">
                {{ row.name }}
              </div>
              <span v-if="displayDimension === 'group' && row.platform" class="ops-concurrency-card__subtitle ops-concurrency-card__subtitle--compact">
                {{ row.platform.toUpperCase() }}
              </span>
            </div>
            <div class="ops-concurrency-card__row-metrics">
              <span class="ops-concurrency-card__primary ops-concurrency-card__primary--mono"> {{ row.used_concurrency }}/{{ row.total_concurrency }} </span>
              <span :class="getLoadTextClass(row.concurrency_percentage)"> {{ row.concurrency_percentage }}% </span>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="ops-concurrency-card__track ops-concurrency-card__track--spaced">
            <div
              :class="getLoadBarClass(row.concurrency_percentage)"
              :style="getLoadBarStyle(row.concurrency_percentage)"
            ></div>
          </div>

          <!-- 统计信息 -->
          <div class="ops-concurrency-card__status-row">
            <!-- 账号统计 -->
            <div class="ops-concurrency-card__status-inline">
              <svg class="ops-concurrency-card__subtitle ops-concurrency-card__status-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                />
              </svg>
              <span class="ops-concurrency-card__secondary">
                <span class="ops-concurrency-card__success ops-concurrency-card__text-strong">{{ row.available_accounts }}</span
                >/{{ row.total_accounts }}
              </span>
              <span class="ops-concurrency-card__subtitle">{{ row.availability_percentage }}%</span>
            </div>

            <!-- 限流账号 -->
            <span
              v-if="row.rate_limited_accounts > 0"
              class="ops-concurrency-card__status-chip theme-chip theme-chip--warning theme-chip--compact"
            >
              {{ t('admin.ops.concurrency.rateLimited', { count: row.rate_limited_accounts }) }}
            </span>

            <!-- 异常账号 -->
            <span
              v-if="row.error_accounts > 0"
              class="ops-concurrency-card__status-chip theme-chip theme-chip--danger theme-chip--compact"
            >
              {{ t('admin.ops.concurrency.errorAccounts', { count: row.error_accounts }) }}
            </span>

            <!-- 等待队列 -->
            <span
              v-if="row.waiting_in_queue > 0"
              class="ops-concurrency-card__status-chip theme-chip theme-chip--brand-purple theme-chip--compact"
            >
              {{ t('admin.ops.concurrency.queued', { count: row.waiting_in_queue }) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 账号详细视图 -->
      <div v-else class="ops-concurrency-card__list custom-scrollbar">
        <div v-for="row in (displayRows as AccountRow[])" :key="row.key" class="ops-concurrency-card__row ops-concurrency-card__row--compact">
          <!-- 账号名称和并发 -->
          <div class="ops-concurrency-card__row-head">
            <div class="ops-concurrency-card__row-copy">
              <div class="ops-concurrency-card__primary ops-concurrency-card__primary--compact ops-concurrency-card__truncate" :title="row.name">
                {{ row.name }}
              </div>
              <div class="ops-concurrency-card__subtitle ops-concurrency-card__subtitle--micro">
                {{ row.group_name }}
              </div>
            </div>
            <div class="ops-concurrency-card__row-metrics ops-concurrency-card__row-metrics--status">
              <!-- 并发使用 -->
              <span class="ops-concurrency-card__primary ops-concurrency-card__primary--mono"> {{ row.current_in_use }}/{{ row.max_capacity }} </span>
              <!-- 状态徽章 -->
              <span
                v-if="row.is_available"
                :class="getAccountStatusClasses(row)"
              >
                <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                {{ t('admin.ops.accountAvailability.available') }}
              </span>
              <span
                v-else-if="row.is_rate_limited"
                :class="getAccountStatusClasses(row)"
              >
                <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {{ formatDuration(row.rate_limit_remaining_sec || 0) }}
              </span>
              <span
                v-else-if="row.is_overloaded"
                :class="getAccountStatusClasses(row)"
              >
                <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
                {{ formatDuration(row.overload_remaining_sec || 0) }}
              </span>
              <span
                v-else-if="row.has_error"
                :class="getAccountStatusClasses(row)"
              >
                <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
                {{ t('admin.ops.accountAvailability.accountError') }}
              </span>
              <span
                v-else
                :class="getAccountStatusClasses(row)"
              >
                {{ t('admin.ops.accountAvailability.unavailable') }}
              </span>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="ops-concurrency-card__track">
            <div :class="getLoadBarClass(row.load_percentage)" :style="getLoadBarStyle(row.load_percentage)"></div>
          </div>

          <!-- 等待队列 -->
          <div v-if="row.waiting_in_queue > 0" class="ops-concurrency-card__status-row ops-concurrency-card__status-row--end">
            <span class="ops-concurrency-card__status-chip theme-chip theme-chip--brand-purple theme-chip--compact">
              {{ t('admin.ops.concurrency.queued', { count: row.waiting_in_queue }) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ops-concurrency-card {
  display: flex;
  height: 100%;
  flex-direction: column;
  padding: var(--theme-ops-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-surface-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.ops-concurrency-card__title,
.ops-concurrency-card__primary,
.ops-concurrency-card__secondary {
  color: var(--theme-page-text);
}

.ops-concurrency-card__header,
.ops-concurrency-card__shell-head,
.ops-concurrency-card__row-head,
.ops-concurrency-card__runtime-headline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--theme-ops-concurrency-header-gap);
}

.ops-concurrency-card__header {
  margin-bottom: var(--theme-ops-dashboard-banner-gap);
  flex-shrink: 0;
}

.ops-concurrency-card__header-actions,
.ops-concurrency-card__row-identity,
.ops-concurrency-card__row-title-group,
.ops-concurrency-card__row-metrics,
.ops-concurrency-card__status-row,
.ops-concurrency-card__status-inline {
  display: flex;
  align-items: center;
}

.ops-concurrency-card__header-actions,
.ops-concurrency-card__row-metrics,
.ops-concurrency-card__status-inline {
  gap: var(--theme-ops-concurrency-control-gap);
}

.ops-concurrency-card__status-row {
  flex-wrap: wrap;
  gap: var(--theme-ops-concurrency-status-gap);
}

.ops-concurrency-card__status-row--end {
  margin-top: calc(var(--theme-ops-concurrency-control-gap) * 3);
  justify-content: flex-end;
}

.ops-concurrency-card__row-identity,
.ops-concurrency-card__row-copy,
.ops-concurrency-card__runtime-copy {
  min-width: 0;
  flex: 1;
}

.ops-concurrency-card__row-copy,
.ops-concurrency-card__runtime-copy {
  display: flex;
  flex-direction: column;
}

.ops-concurrency-card__row-identity {
  flex: 1;
  gap: calc(var(--theme-ops-concurrency-control-gap) * 0.75);
}

.ops-concurrency-card__subtitle {
  color: var(--theme-page-muted);
}

.ops-concurrency-card__title {
  display: inline-flex;
  align-items: center;
  gap: var(--theme-ops-concurrency-control-gap);
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-concurrency-card__title-icon {
  width: 1rem;
  height: 1rem;
  color: rgb(var(--theme-info-rgb));
}

.ops-concurrency-card__view-toggle,
.ops-concurrency-card__refresh {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.ops-concurrency-card__view-toggle--active {
  padding: calc(var(--theme-ops-table-cell-padding-compact-y) * 0.9) var(--theme-ops-table-cell-padding-compact-x);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 12%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.ops-concurrency-card__view-toggle--idle,
.ops-concurrency-card__refresh {
  padding: calc(var(--theme-ops-table-cell-padding-compact-y) * 0.9) var(--theme-ops-table-cell-padding-compact-x);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.ops-concurrency-card__icon-button-icon {
  width: 0.875rem;
  height: 0.875rem;
}

.ops-concurrency-card__refresh {
  gap: calc(var(--theme-ops-concurrency-control-gap) * 0.5);
  font-size: 0.6875rem;
  font-weight: 600;
}

.ops-concurrency-card__refresh-icon {
  width: 0.75rem;
  height: 0.75rem;
}

.ops-concurrency-card__view-toggle--idle:hover,
.ops-concurrency-card__refresh:hover {
  background: color-mix(in srgb, var(--theme-page-border) 66%, var(--theme-surface));
  color: var(--theme-page-text);
}

.ops-concurrency-card__error {
  margin-bottom: var(--theme-ops-concurrency-control-gap);
  flex-shrink: 0;
  padding: var(--theme-ops-row-padding-compact);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  font-size: 0.75rem;
}

.ops-concurrency-card__empty,
.ops-concurrency-card__shell {
  border-radius: var(--theme-select-panel-radius);
  border-color: color-mix(in srgb, var(--theme-page-border) 74%, transparent);
}

.ops-concurrency-card__empty {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
}

.ops-concurrency-card__empty--disabled {
  border: 1px dashed color-mix(in srgb, var(--theme-page-border) 74%, transparent);
}

.ops-concurrency-card__shell {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--theme-page-border) 74%, transparent);
}

.ops-concurrency-card__shell-head {
  flex-shrink: 0;
  padding:
    var(--theme-ops-table-cell-padding-compact-y)
    var(--theme-ops-table-cell-padding-compact-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 70%, transparent);
  border-color: color-mix(in srgb, var(--theme-page-border) 70%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface));
}

.ops-concurrency-card__runtime {
  padding: 0.75rem;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 68%, transparent);
  border-color: color-mix(in srgb, var(--theme-page-border) 68%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}

.ops-concurrency-card__runtime-headline {
  align-items: flex-start;
  padding: 0.625rem 0.75rem;
  border: 1px solid transparent;
  border-radius: 0.75rem;
}

.ops-concurrency-card__runtime-title {
  font-size: 0.6875rem;
  font-weight: 700;
  color: var(--theme-page-text);
}

.ops-concurrency-card__runtime-detail {
  margin-top: calc(var(--theme-ops-concurrency-control-gap) * 0.5);
  font-size: 0.625rem;
  color: var(--theme-page-muted);
}

.ops-concurrency-card__runtime-probes {
  flex-shrink: 0;
  text-align: right;
}

.ops-concurrency-card__runtime-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--theme-ops-concurrency-metric-gap);
  margin-top: var(--theme-ops-concurrency-metric-gap);
}

.ops-concurrency-card__runtime-metric {
  padding: 0.5rem 0.625rem;
  border: 1px solid transparent;
  border-radius: 0.5rem;
}

.ops-concurrency-card__runtime-headline--healthy,
.ops-concurrency-card__runtime-metric--healthy {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 28%, transparent);
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 8%, var(--theme-surface));
}

.ops-concurrency-card__runtime-headline--warning,
.ops-concurrency-card__runtime-metric--warning {
  border-color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 32%, transparent);
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
}

.ops-concurrency-card__runtime-headline--critical,
.ops-concurrency-card__runtime-metric--critical {
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 34%, transparent);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.ops-concurrency-card__list {
  flex: 1;
  max-height: var(--theme-ops-list-max-height);
  overflow-y: auto;
  padding: var(--theme-ops-table-cell-padding-compact-x);
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-concurrency-list-gap);
}

.ops-concurrency-card__row {
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.ops-concurrency-card__row--compact {
  padding: var(--theme-ops-row-padding-compact);
}

.ops-concurrency-card__row--regular {
  padding: var(--theme-ops-row-padding-regular);
}

.ops-concurrency-card__row-head--regular {
  margin-bottom: var(--theme-ops-concurrency-control-gap);
}

.ops-concurrency-card__status-chip {
  display: inline-flex;
  align-items: center;
  gap: calc(var(--theme-ops-concurrency-control-gap) * 0.5);
  padding:
    calc(var(--theme-button-padding-y) * 0.2)
    calc(var(--theme-button-padding-x) * 0.32);
  border-radius: calc(var(--theme-button-radius) * 0.75);
  font-size: 0.625rem;
  font-weight: 600;
}

.ops-concurrency-card__track {
  width: 100%;
  height: 0.375rem;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--theme-page-border) 62%, var(--theme-surface));
}

.ops-concurrency-card__track--spaced {
  margin-bottom: var(--theme-ops-concurrency-control-gap);
}

.ops-concurrency-card__bar {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.ops-concurrency-card__load-text {
  font-size: 0.625rem;
  font-weight: 700;
}

.ops-concurrency-card__primary--compact {
  font-size: 0.6875rem;
  font-weight: 700;
}

.ops-concurrency-card__primary--mono {
  font-family: var(--theme-font-mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, Courier New, monospace);
  font-size: 0.6875rem;
  font-weight: 700;
}

.ops-concurrency-card__primary--hero {
  margin-top: calc(var(--theme-ops-concurrency-control-gap) * 0.5);
  font-size: 0.875rem;
  font-weight: 900;
}

.ops-concurrency-card__primary--metric {
  margin-top: calc(var(--theme-ops-concurrency-control-gap) * 0.5);
  font-size: 0.8125rem;
  font-weight: 900;
}

.ops-concurrency-card__subtitle--label,
.ops-concurrency-card__subtitle--micro-label {
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.ops-concurrency-card__subtitle--label,
.ops-concurrency-card__subtitle--count,
.ops-concurrency-card__subtitle--compact {
  font-size: 0.625rem;
}

.ops-concurrency-card__subtitle--micro,
.ops-concurrency-card__subtitle--micro-label {
  font-size: 0.5625rem;
}

.ops-concurrency-card__subtitle--micro {
  margin-top: calc(var(--theme-ops-concurrency-control-gap) * 0.25);
}

.ops-concurrency-card__status-icon {
  width: 0.75rem;
  height: 0.75rem;
}

.ops-concurrency-card__text-strong {
  font-weight: 700;
}

.ops-concurrency-card__truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ops-concurrency-card__row-metrics--status {
  flex-shrink: 0;
}

.ops-concurrency-card__bar--danger {
  background: rgb(var(--theme-danger-rgb));
}

.ops-concurrency-card__bar--warning {
  background: rgb(var(--theme-warning-rgb));
}

.ops-concurrency-card__bar--info {
  background: rgb(var(--theme-info-rgb));
}

.ops-concurrency-card__bar--success {
  background: rgb(var(--theme-success-rgb));
}

.ops-concurrency-card__load-text--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.ops-concurrency-card__load-text--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.ops-concurrency-card__load-text--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.ops-concurrency-card__load-text--success,
.ops-concurrency-card__success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: color-mix(in srgb, var(--theme-page-muted) 34%, transparent) transparent;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: color-mix(in srgb, var(--theme-page-muted) 34%, transparent);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: color-mix(in srgb, var(--theme-page-muted) 50%, transparent);
}
</style>
