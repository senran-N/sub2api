<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { opsAPI, type AlertEventsQuery } from '@/api/admin/ops'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import type { AlertEvent } from '../types'
import { formatDateTime } from '../utils/opsFormatters'

const { t } = useI18n()
const appStore = useAppStore()

const props = withDefaults(defineProps<{
  refreshToken?: number
}>(), {
  refreshToken: 0
})

const PAGE_SIZE = 10

const loading = ref(false)
const loadingMore = ref(false)
const events = ref<AlertEvent[]>([])
const hasMore = ref(true)
let listSequence = 0

// Detail modal
const showDetail = ref(false)
const selected = ref<AlertEvent | null>(null)
const detailLoading = ref(false)
const detailActionLoading = ref(false)
const historyLoading = ref(false)
const history = ref<AlertEvent[]>([])
let detailSequence = 0
let historySequence = 0
let detailActionSequence = 0
const historyRange = ref('7d')
const historyRangeOptions = computed(() => [
  { value: '7d', label: t('admin.ops.timeRange.7d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const silenceDuration = ref('1h')
const silenceDurationOptions = computed(() => [
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  { value: '7d', label: t('admin.ops.timeRange.7d') }
])

// Filters
const timeRange = ref('24h')
const timeRangeOptions = computed(() => [
  { value: '5m', label: t('admin.ops.timeRange.5m') },
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '6h', label: t('admin.ops.timeRange.6h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  { value: '7d', label: t('admin.ops.timeRange.7d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const severity = ref<string>('')
const severityOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'P0', label: 'P0' },
  { value: 'P1', label: 'P1' },
  { value: 'P2', label: 'P2' },
  { value: 'P3', label: 'P3' }
])

const status = ref<string>('')
const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'firing', label: t('admin.ops.alertEvents.status.firing') },
  { value: 'resolved', label: t('admin.ops.alertEvents.status.resolved') },
  { value: 'manual_resolved', label: t('admin.ops.alertEvents.status.manualResolved') }
])

const emailSent = ref<string>('')
const emailSentOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'true', label: t('admin.ops.alertEvents.table.emailSent') },
  { value: 'false', label: t('admin.ops.alertEvents.table.emailIgnored') }
])

function buildQuery(overrides: Partial<AlertEventsQuery> = {}): AlertEventsQuery {
  const q: AlertEventsQuery = {
    limit: PAGE_SIZE,
    time_range: timeRange.value
  }
  if (severity.value) q.severity = severity.value
  if (status.value) q.status = status.value
  if (emailSent.value === 'true') q.email_sent = true
  if (emailSent.value === 'false') q.email_sent = false
  return { ...q, ...overrides }
}

async function loadFirstPage() {
  const requestSequence = ++listSequence
  loading.value = true
  try {
    const data = await opsAPI.listAlertEvents(buildQuery())
    if (requestSequence !== listSequence) return
    events.value = data
    hasMore.value = data.length === PAGE_SIZE
  } catch (err: unknown) {
    if (requestSequence !== listSequence) return
    console.error('[OpsAlertEventsCard] Failed to load alert events', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertEvents.loadFailed')))
    events.value = []
    hasMore.value = false
  } finally {
    if (requestSequence === listSequence) {
      loading.value = false
    }
  }
}

async function loadMore() {
  if (loadingMore.value || loading.value) return
  if (!hasMore.value) return
  const last = events.value[events.value.length - 1]
  if (!last) return

  const requestSequence = listSequence
  loadingMore.value = true
  try {
    const data = await opsAPI.listAlertEvents(
      buildQuery({ before_fired_at: last.fired_at || last.created_at, before_id: last.id })
    )
    if (requestSequence !== listSequence) return
    if (!data.length) {
      hasMore.value = false
      return
    }
    events.value = [...events.value, ...data]
    if (data.length < PAGE_SIZE) hasMore.value = false
  } catch (err: unknown) {
    if (requestSequence !== listSequence) return
    console.error('[OpsAlertEventsCard] Failed to load more alert events', err)
    hasMore.value = false
  } finally {
    if (requestSequence === listSequence) {
      loadingMore.value = false
    }
  }
}

function onScroll(e: Event) {
  const el = e.target as HTMLElement | null
  if (!el) return
  const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 120
  if (nearBottom) loadMore()
}

function getDimensionString(event: AlertEvent | null | undefined, key: string): string {
  const v = event?.dimensions?.[key]
  if (v == null) return ''
  if (typeof v === 'string') return v
  if (typeof v === 'number' || typeof v === 'boolean') return String(v)
  return ''
}

function formatDurationMs(ms: number): string {
  const safe = Math.max(0, Math.floor(ms))
  const sec = Math.floor(safe / 1000)
  if (sec < 60) return `${sec}s`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h`
  const day = Math.floor(hr / 24)
  return `${day}d`
}

function formatDurationLabel(event: AlertEvent): string {
  const firedAt = new Date(event.fired_at || event.created_at)
  if (Number.isNaN(firedAt.getTime())) return '-'
  const resolvedAtStr = event.resolved_at || null
  const status = String(event.status || '').trim().toLowerCase()

  if (resolvedAtStr) {
    const resolvedAt = new Date(resolvedAtStr)
    if (!Number.isNaN(resolvedAt.getTime())) {
      const ms = resolvedAt.getTime() - firedAt.getTime()
      const prefix = status === 'manual_resolved'
        ? t('admin.ops.alertEvents.status.manualResolved')
        : t('admin.ops.alertEvents.status.resolved')
      return `${prefix} ${formatDurationMs(ms)}`
    }
  }

  const now = Date.now()
  const ms = now - firedAt.getTime()
  return `${t('admin.ops.alertEvents.status.firing')} ${formatDurationMs(ms)}`
}

function formatDimensionsSummary(event: AlertEvent): string {
  const parts: string[] = []
  const platform = getDimensionString(event, 'platform')
  if (platform) parts.push(`platform=${platform}`)
  const groupId = event.dimensions?.group_id
  if (groupId != null && groupId !== '') parts.push(`group_id=${String(groupId)}`)
  const region = getDimensionString(event, 'region')
  if (region) parts.push(`region=${region}`)
  return parts.length ? parts.join(' ') : '-'
}

function closeDetail() {
  detailSequence += 1
  historySequence += 1
  detailActionSequence += 1
  showDetail.value = false
  selected.value = null
  history.value = []
  detailLoading.value = false
  historyLoading.value = false
  detailActionLoading.value = false
}

async function openDetail(row: AlertEvent) {
  const requestSequence = ++detailSequence
  historySequence += 1
  detailActionSequence += 1
  showDetail.value = true
  selected.value = row
  detailLoading.value = true
  detailActionLoading.value = false
  history.value = []
  historyLoading.value = true

  try {
    const detail = await opsAPI.getAlertEvent(row.id)
    if (requestSequence !== detailSequence) return
    selected.value = detail
  } catch (err: unknown) {
    if (requestSequence !== detailSequence) return
    console.error('[OpsAlertEventsCard] Failed to load alert detail', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertEvents.detail.loadFailed')))
  } finally {
    if (requestSequence === detailSequence) {
      detailLoading.value = false
    }
  }

  if (requestSequence === detailSequence) {
    await loadHistory(requestSequence)
  }
}

async function loadHistory(expectedDetailSequence = detailSequence) {
  if (expectedDetailSequence !== detailSequence) return
  const ev = selected.value
  if (!ev) {
    historySequence += 1
    history.value = []
    historyLoading.value = false
    return
  }

  const requestSequence = ++historySequence
  historyLoading.value = true
  try {
    const platform = getDimensionString(ev, 'platform')
    const groupIdRaw = ev.dimensions?.group_id
    const groupId = typeof groupIdRaw === 'number' ? groupIdRaw : undefined

    const items = await opsAPI.listAlertEvents({
      limit: 20,
      time_range: historyRange.value,
      platform: platform || undefined,
      group_id: groupId,
      status: ''
    })
    if (requestSequence !== historySequence || expectedDetailSequence !== detailSequence) return

    // Best-effort: narrow to same rule_id + dimensions
    history.value = items.filter((it) => {
      if (it.rule_id !== ev.rule_id) return false
      const p1 = getDimensionString(it, 'platform')
      const p2 = getDimensionString(ev, 'platform')
      if ((p1 || '') !== (p2 || '')) return false
      const g1 = it.dimensions?.group_id
      const g2 = ev.dimensions?.group_id
      return (g1 ?? null) === (g2 ?? null)
    })
  } catch (err: unknown) {
    if (requestSequence !== historySequence || expectedDetailSequence !== detailSequence) return
    console.error('[OpsAlertEventsCard] Failed to load alert history', err)
    history.value = []
  } finally {
    if (requestSequence === historySequence && expectedDetailSequence === detailSequence) {
      historyLoading.value = false
    }
  }
}

function durationToUntilRFC3339(duration: string): string {
  const now = Date.now()
  if (duration === '1h') return new Date(now + 60 * 60 * 1000).toISOString()
  if (duration === '24h') return new Date(now + 24 * 60 * 60 * 1000).toISOString()
  if (duration === '7d') return new Date(now + 7 * 24 * 60 * 60 * 1000).toISOString()
  return new Date(now + 60 * 60 * 1000).toISOString()
}

async function silenceAlert() {
  const ev = selected.value
  if (!ev) return
  if (detailActionLoading.value) return
  const requestSequence = ++detailActionSequence
  detailActionLoading.value = true
  try {
    const platform = getDimensionString(ev, 'platform')
    const groupIdRaw = ev.dimensions?.group_id
    const groupId = typeof groupIdRaw === 'number' ? groupIdRaw : null
    const region = getDimensionString(ev, 'region') || null

    await opsAPI.createAlertSilence({
      rule_id: ev.rule_id,
      platform: platform || '',
      group_id: groupId ?? undefined,
      region: region ?? undefined,
      until: durationToUntilRFC3339(silenceDuration.value),
      reason: `silence from UI (${silenceDuration.value})`
    })

    appStore.showSuccess(t('admin.ops.alertEvents.detail.silenceSuccess'))
  } catch (err: unknown) {
    console.error('[OpsAlertEventsCard] Failed to silence alert', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertEvents.detail.silenceFailed')))
  } finally {
    if (requestSequence === detailActionSequence) {
      detailActionLoading.value = false
    }
  }
}

async function manualResolve() {
  const ev = selected.value
  if (!ev) return
  if (detailActionLoading.value) return
  const selectionSequence = detailSequence
  const requestSequence = ++detailActionSequence
  detailActionLoading.value = true
  try {
    await opsAPI.updateAlertEventStatus(ev.id, 'manual_resolved')
    appStore.showSuccess(t('admin.ops.alertEvents.detail.manualResolvedSuccess'))

    // Refresh detail + first page to reflect new status
    const detail = await opsAPI.getAlertEvent(ev.id)
    if (selectionSequence !== detailSequence) return
    selected.value = detail
    await loadFirstPage()
    if (selectionSequence !== detailSequence) return
    await loadHistory(selectionSequence)
  } catch (err: unknown) {
    console.error('[OpsAlertEventsCard] Failed to resolve alert', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertEvents.detail.manualResolvedFailed')))
  } finally {
    if (requestSequence === detailActionSequence) {
      detailActionLoading.value = false
    }
  }
}

async function refreshSelectedDetail() {
  const current = selected.value
  if (!current) return

  const requestSequence = ++detailSequence
  historySequence += 1
  detailActionSequence += 1
  detailLoading.value = true
  history.value = []
  historyLoading.value = true

  try {
    const detail = await opsAPI.getAlertEvent(current.id)
    if (requestSequence !== detailSequence) return
    selected.value = detail
  } catch (err: unknown) {
    if (requestSequence !== detailSequence) return
    console.error('[OpsAlertEventsCard] Failed to refresh alert detail', err)
    appStore.showError(resolveRequestErrorMessage(err, t('admin.ops.alertEvents.detail.loadFailed')))
  } finally {
    if (requestSequence === detailSequence) {
      detailLoading.value = false
    }
  }

  if (requestSequence === detailSequence) {
    await loadHistory(requestSequence)
  }
}

onMounted(() => {
  loadFirstPage()
})

watch([timeRange, severity, status, emailSent], () => {
  events.value = []
  hasMore.value = true
  loadFirstPage()
})

watch(historyRange, () => {
  if (showDetail.value) loadHistory(detailSequence)
})

watch(() => props.refreshToken, () => {
  void loadFirstPage()
  if (showDetail.value && selected.value) {
    void refreshSelectedDetail()
  }
})

function severityBadgeClass(severity: string | undefined): string {
  const s = String(severity || '').trim().toLowerCase()
  if (s === 'p0' || s === 'critical') return getBadgeClasses('danger')
  if (s === 'p1' || s === 'warning') return getBadgeClasses('warning')
  if (s === 'p2' || s === 'info') return getBadgeClasses('info')
  if (s === 'p3') return getBadgeClasses('neutral')
  return getBadgeClasses('neutral')
}

function statusBadgeClass(status: string | undefined): string {
  const s = String(status || '').trim().toLowerCase()
  if (s === 'firing') return getBadgeClasses('danger')
  if (s === 'resolved') return getBadgeClasses('success')
  if (s === 'manual_resolved') return getBadgeClasses('neutral')
  return getBadgeClasses('neutral')
}

function formatStatusLabel(status: string | undefined): string {
  const s = String(status || '').trim().toLowerCase()
  if (!s) return '-'
  if (s === 'firing') return t('admin.ops.alertEvents.status.firing')
  if (s === 'resolved') return t('admin.ops.alertEvents.status.resolved')
  if (s === 'manual_resolved') return t('admin.ops.alertEvents.status.manualResolved')
  return s.toUpperCase()
}

const empty = computed(() => events.value.length === 0 && !loading.value)

type AlertEventsTone = 'danger' | 'info' | 'neutral' | 'success' | 'warning'

function joinClassNames(classNames: Array<string | false | null | undefined>) {
  return classNames.filter(Boolean).join(' ')
}

function getBadgeClasses(tone: AlertEventsTone) {
  return joinClassNames([
    'theme-chip theme-chip--compact ops-alert-events-card__badge',
    tone === 'danger' && 'theme-chip--danger',
    tone === 'info' && 'theme-chip--info',
    tone === 'neutral' && 'theme-chip--neutral',
    tone === 'success' && 'theme-chip--success',
    tone === 'warning' && 'theme-chip--warning'
  ])
}

function getEmailStateClasses(emailSent: boolean) {
  return joinClassNames([
    'ops-alert-events-card__email-state',
    emailSent
      ? 'ops-alert-events-card__email-state--sent'
      : 'ops-alert-events-card__email-state--idle'
  ])
}
</script>

<template>
  <div class="ops-alert-events-card">
    <div class="ops-alert-events-card__header">
      <div class="ops-alert-events-card__header-copy">
        <h3 class="ops-alert-events-card__title">{{ t('admin.ops.alertEvents.title') }}</h3>
        <p class="ops-alert-events-card__subtitle ops-alert-events-card__subtitle--header">{{ t('admin.ops.alertEvents.description') }}</p>
      </div>

      <div class="ops-alert-events-card__filters">
        <div class="ops-alert-events-card__filter">
          <Select :model-value="timeRange" :options="timeRangeOptions" class="ops-alert-events-card__filter-select" @change="timeRange = String($event || '24h')" />
        </div>
        <div class="ops-alert-events-card__filter">
          <Select :model-value="severity" :options="severityOptions" class="ops-alert-events-card__filter-select" @change="severity = String($event || '')" />
        </div>
        <div class="ops-alert-events-card__filter">
          <Select :model-value="status" :options="statusOptions" class="ops-alert-events-card__filter-select" @change="status = String($event || '')" />
        </div>
        <div class="ops-alert-events-card__filter">
          <Select :model-value="emailSent" :options="emailSentOptions" class="ops-alert-events-card__filter-select" @change="emailSent = String($event || '')" />
        </div>
        <button
          class="ops-alert-events-card__refresh"
          :disabled="loading"
          @click="loadFirstPage"
        >
          <svg class="ops-alert-events-card__refresh-icon" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="ops-alert-events-card__muted ops-alert-events-card__loading">
      <svg class="ops-alert-events-card__loading-icon animate-spin" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      {{ t('admin.ops.alertEvents.loading') }}
    </div>

    <div v-else-if="empty" class="ops-alert-events-card__empty">
      {{ t('admin.ops.alertEvents.empty') }}
    </div>

    <div v-else class="ops-alert-events-card__table-shell">
      <div class="ops-alert-events-card__table-scroll" @scroll="onScroll">
        <table class="ops-alert-events-card__table">
          <thead class="ops-alert-events-card__table-head">
            <tr>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular">
                {{ t('admin.ops.alertEvents.table.time') }}
              </th>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular">
                {{ t('admin.ops.alertEvents.table.severity') }}
              </th>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular">
                {{ t('admin.ops.alertEvents.table.platform') }}
              </th>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular">
                {{ t('admin.ops.alertEvents.table.ruleId') }}
              </th>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular">
                {{ t('admin.ops.alertEvents.table.title') }}
              </th>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular">
                {{ t('admin.ops.alertEvents.table.duration') }}
              </th>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular">
                {{ t('admin.ops.alertEvents.table.dimensions') }}
              </th>
              <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--regular ops-alert-events-card__table-header--end">
                {{ t('admin.ops.alertEvents.table.email') }}
              </th>
            </tr>
          </thead>
          <tbody class="ops-alert-events-card__table-body">
            <tr
              v-for="row in events"
              :key="row.id"
              class="ops-alert-events-card__row"
              @click="openDetail(row)"
              :title="row.title || ''"
            >
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--nowrap ops-alert-events-card__table-cell--xs ops-alert-events-card__cell-secondary">
                {{ formatDateTime(row.fired_at || row.created_at) }}
              </td>
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--nowrap">
                <div class="ops-alert-events-card__badge-group">
                  <span :class="severityBadgeClass(String(row.severity || ''))">
                    {{ row.severity || '-' }}
                  </span>
                  <span :class="statusBadgeClass(row.status)">
                    {{ formatStatusLabel(row.status) }}
                  </span>
                </div>
              </td>
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--nowrap ops-alert-events-card__table-cell--xs ops-alert-events-card__cell-secondary">
                {{ getDimensionString(row, 'platform') || '-' }}
              </td>
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--nowrap ops-alert-events-card__table-cell--xs ops-alert-events-card__cell-secondary">
                <span class="ops-alert-events-card__mono">#{{ row.rule_id }}</span>
              </td>
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--xs ops-alert-events-card__cell-primary ops-alert-events-card__title-cell">
                <div class="ops-alert-events-card__title-text ops-alert-events-card__title-line ops-alert-events-card__title-line--strong">
                  {{ row.title || '-' }}
                </div>
                <div v-if="row.description" class="ops-alert-events-card__subtitle ops-alert-events-card__subtitle--clamped">
                  {{ row.description }}
                </div>
              </td>
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--nowrap ops-alert-events-card__table-cell--xs ops-alert-events-card__cell-secondary">
                {{ formatDurationLabel(row) }}
              </td>
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--nowrap ops-alert-events-card__table-cell--xxs ops-alert-events-card__subtitle">
                {{ formatDimensionsSummary(row) }}
              </td>
              <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--regular ops-alert-events-card__table-cell--nowrap ops-alert-events-card__table-cell--end ops-alert-events-card__table-cell--xs">
                <span
                  :class="getEmailStateClasses(!!row.email_sent)"
                  :title="row.email_sent ? t('admin.ops.alertEvents.table.emailSent') : t('admin.ops.alertEvents.table.emailIgnored')"
                >
                  <Icon
                    v-if="row.email_sent"
                    name="checkCircle"
                    size="sm"
                    class="ops-alert-events-card__email-icon"
                  />
                  <Icon
                    v-else
                    name="ban"
                    size="sm"
                    class="ops-alert-events-card__email-icon"
                  />
                  <span class="ops-alert-events-card__email-label">
                    {{ row.email_sent ? t('admin.ops.alertEvents.table.emailSent') : t('admin.ops.alertEvents.table.emailIgnored') }}
                  </span>
                </span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="loadingMore" class="ops-alert-events-card__muted ops-alert-events-card__state ops-alert-events-card__state--compact ops-alert-events-card__state--loading">
          <svg class="ops-alert-events-card__loading-icon animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ t('admin.ops.alertEvents.loading') }}
        </div>
        <div v-else-if="!hasMore && events.length > 0" class="ops-alert-events-card__subtitle ops-alert-events-card__state ops-alert-events-card__state--compact ops-alert-events-card__state--center ops-alert-events-card__table-cell--xs">
          -
        </div>
      </div>
    </div>

    <BaseDialog
      :show="showDetail"
      :title="t('admin.ops.alertEvents.detail.title')"
      width="wide"
      :close-on-click-outside="true"
      @close="closeDetail"
    >
      <div v-if="detailLoading" class="ops-alert-events-card__muted ops-alert-events-card__state ops-alert-events-card__state--regular ops-alert-events-card__state--center">
        {{ t('admin.ops.alertEvents.detail.loading') }}
      </div>

      <div v-else-if="!selected" class="ops-alert-events-card__muted ops-alert-events-card__state ops-alert-events-card__state--regular ops-alert-events-card__state--center">
        {{ t('admin.ops.alertEvents.detail.empty') }}
      </div>

      <div v-else class="ops-alert-events-card__detail-content">
        <div class="ops-alert-events-card__panel">
          <div class="ops-alert-events-card__detail-header">
            <div>
              <div class="ops-alert-events-card__badge-group">
                <span :class="severityBadgeClass(String(selected.severity || ''))">
                  {{ selected.severity || '-' }}
                </span>
                <span :class="statusBadgeClass(selected.status)">
                  {{ formatStatusLabel(selected.status) }}
                </span>
              </div>
              <div class="ops-alert-events-card__title-text ops-alert-events-card__detail-title">
                {{ selected.title || '-' }}
              </div>
              <div v-if="selected.description" class="ops-alert-events-card__cell-secondary ops-alert-events-card__detail-description">
                {{ selected.description }}
              </div>
            </div>

            <div class="ops-alert-events-card__detail-actions">
              <div class="ops-alert-events-card__action-group">
                <span class="ops-alert-events-card__cell-secondary ops-alert-events-card__detail-label">{{ t('admin.ops.alertEvents.detail.silence') }}</span>
                <Select
                  :model-value="silenceDuration"
                  :options="silenceDurationOptions"
                  class="ops-alert-events-card__select-inline"
                  @change="silenceDuration = String($event || '1h')"
                />
                <button type="button" class="btn btn-secondary btn-sm" :disabled="detailActionLoading" @click="silenceAlert">
                  <Icon name="ban" size="sm" />
                  {{ t('common.apply') }}
                </button>
              </div>

              <button type="button" class="btn btn-secondary btn-sm" :disabled="detailActionLoading" @click="manualResolve">
                <Icon name="checkCircle" size="sm" />
                {{ t('admin.ops.alertEvents.detail.manualResolve') }}
              </button>
            </div>
          </div>
        </div>

        <div class="ops-alert-events-card__detail-grid">
          <div class="ops-alert-events-card__panel">
            <div class="ops-alert-events-card__detail-kicker">{{ t('admin.ops.alertEvents.detail.firedAt') }}</div>
            <div class="ops-alert-events-card__title-text ops-alert-events-card__detail-value">{{ formatDateTime(selected.fired_at || selected.created_at) }}</div>
          </div>
          <div class="ops-alert-events-card__panel">
            <div class="ops-alert-events-card__detail-kicker">{{ t('admin.ops.alertEvents.detail.resolvedAt') }}</div>
            <div class="ops-alert-events-card__title-text ops-alert-events-card__detail-value">{{ selected.resolved_at ? formatDateTime(selected.resolved_at) : '-' }}</div>
          </div>
          <div class="ops-alert-events-card__panel">
            <div class="ops-alert-events-card__detail-kicker">{{ t('admin.ops.alertEvents.detail.ruleId') }}</div>
            <div class="ops-alert-events-card__detail-links">
              <div class="ops-alert-events-card__title-text ops-alert-events-card__mono ops-alert-events-card__detail-rule-id">#{{ selected.rule_id }}</div>
                <a
                  class="ops-alert-events-card__link"
                  :href="`/admin/ops?open_alert_rules=1&alert_rule_id=${selected.rule_id}`"
                >
                  <Icon name="externalLink" size="xs" />
                  {{ t('admin.ops.alertEvents.detail.viewRule') }}
                </a>
                <a
                  class="ops-alert-events-card__link"
                  :href="`/admin/ops?platform=${encodeURIComponent(getDimensionString(selected,'platform')||'')}&group_id=${selected.dimensions?.group_id || ''}&error_type=request&open_error_details=1`"
                >
                  <Icon name="externalLink" size="xs" />
                  {{ t('admin.ops.alertEvents.detail.viewLogs') }}
                </a>
              </div>
          </div>
          <div class="ops-alert-events-card__panel">
            <div class="ops-alert-events-card__detail-kicker">{{ t('admin.ops.alertEvents.detail.dimensions') }}</div>
            <div class="ops-alert-events-card__title-text ops-alert-events-card__detail-value">
              <div v-if="getDimensionString(selected, 'platform')">platform={{ getDimensionString(selected, 'platform') }}</div>
              <div v-if="selected.dimensions?.group_id">group_id={{ selected.dimensions.group_id }}</div>
              <div v-if="getDimensionString(selected, 'region')">region={{ getDimensionString(selected, 'region') }}</div>
            </div>
          </div>
        </div>

        <div class="ops-alert-events-card__history-shell">
          <div class="ops-alert-events-card__history-header">
            <div>
              <div class="ops-alert-events-card__title-text ops-alert-events-card__history-title">{{ t('admin.ops.alertEvents.detail.historyTitle') }}</div>
              <div class="ops-alert-events-card__subtitle ops-alert-events-card__subtitle--history">{{ t('admin.ops.alertEvents.detail.historyHint') }}</div>
            </div>
            <Select :model-value="historyRange" :options="historyRangeOptions" class="ops-alert-events-card__filter-select" @change="historyRange = String($event || '7d')" />
          </div>

          <div v-if="historyLoading" class="ops-alert-events-card__muted ops-alert-events-card__state ops-alert-events-card__state--history ops-alert-events-card__state--center ops-alert-events-card__table-cell--xs">
            {{ t('admin.ops.alertEvents.detail.historyLoading') }}
          </div>
          <div v-else-if="history.length === 0" class="ops-alert-events-card__muted ops-alert-events-card__state ops-alert-events-card__state--history ops-alert-events-card__state--center ops-alert-events-card__table-cell--xs">
            {{ t('admin.ops.alertEvents.detail.historyEmpty') }}
          </div>
          <div v-else class="ops-alert-events-card__history-table">
            <table class="ops-alert-events-card__table">
              <thead class="ops-alert-events-card__table-head">
                <tr>
                  <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--compact">{{ t('admin.ops.alertEvents.table.time') }}</th>
                  <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--compact">{{ t('admin.ops.alertEvents.table.status') }}</th>
                  <th class="ops-alert-events-card__table-header ops-alert-events-card__table-header--compact">{{ t('admin.ops.alertEvents.table.metric') }}</th>
                </tr>
              </thead>
              <tbody class="ops-alert-events-card__table-body">
                <tr v-for="it in history" :key="it.id" class="ops-alert-events-card__row">
                  <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--compact ops-alert-events-card__table-cell--xs ops-alert-events-card__cell-secondary">{{ formatDateTime(it.fired_at || it.created_at) }}</td>
                  <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--compact ops-alert-events-card__table-cell--xs">
                    <span :class="statusBadgeClass(it.status)">
                      {{ formatStatusLabel(it.status) }}
                    </span>
                  </td>
                  <td class="ops-alert-events-card__table-cell ops-alert-events-card__table-cell--compact ops-alert-events-card__table-cell--xs ops-alert-events-card__cell-secondary">
                    <span v-if="typeof it.metric_value === 'number' && typeof it.threshold_value === 'number'">
                      {{ it.metric_value.toFixed(2) }} / {{ it.threshold_value.toFixed(2) }}
                    </span>
                    <span v-else>-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </BaseDialog>
  </div>
</template>

<style scoped>
.ops-alert-events-card {
  padding: var(--theme-ops-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-surface-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.ops-alert-events-card__title,
.ops-alert-events-card__title-text {
  color: var(--theme-page-text);
}

.ops-alert-events-card__header {
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-alert-events-header-gap);
  margin-bottom: var(--theme-ops-panel-padding);
}

.ops-alert-events-card__header-copy {
  min-width: 0;
}

.ops-alert-events-card__title {
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-alert-events-card__subtitle--header {
  margin-top: 0.25rem;
  font-size: 0.75rem;
}

.ops-alert-events-card__filters {
  display: flex;
  flex-wrap: wrap;
  gap: var(--theme-ops-alert-events-filter-gap);
}

.ops-alert-events-card__filter {
  flex: 0 1 var(--theme-ops-alert-events-filter-basis);
  min-width: var(--theme-ops-alert-events-filter-min-width);
}

.ops-alert-events-card__filter-select {
  width: 100%;
}

.ops-alert-events-card__select-inline {
  width: calc(var(--theme-ops-table-min-width) * 0.1375);
}

.ops-alert-events-card__subtitle,
.ops-alert-events-card__muted,
.ops-alert-events-card__table-header,
.ops-alert-events-card__detail-kicker {
  color: var(--theme-page-muted);
}

.ops-alert-events-card__refresh {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: calc(var(--theme-button-padding-y) * 0.6) calc(var(--theme-button-padding-x) * 0.75);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
  font-size: 0.75rem;
  font-weight: 700;
  transition: background-color 0.2s ease, color 0.2s ease, opacity 0.2s ease;
}

.ops-alert-events-card__refresh:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.ops-alert-events-card__refresh-icon {
  width: 0.875rem;
  height: 0.875rem;
}

.ops-alert-events-card__refresh:hover {
  background: color-mix(in srgb, var(--theme-page-border) 68%, var(--theme-surface));
}

.ops-alert-events-card__loading {
  display: flex;
  align-items: center;
  gap: var(--theme-ops-alert-events-loading-gap);
  font-size: 0.875rem;
}

.ops-alert-events-card__loading-icon {
  width: 1rem;
  height: 1rem;
}

.ops-alert-events-card__empty,
.ops-alert-events-card__table-shell,
.ops-alert-events-card__history-shell,
.ops-alert-events-card__history-table {
  border-color: color-mix(in srgb, var(--theme-page-border) 74%, transparent);
}

.ops-alert-events-card__empty,
.ops-alert-events-card__table-shell,
.ops-alert-events-card__history-shell,
.ops-alert-events-card__history-table,
.ops-alert-events-card__panel {
  border-radius: var(--theme-select-panel-radius);
}

.ops-alert-events-card__empty {
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.67);
  border: 1px dashed color-mix(in srgb, var(--theme-page-border) 74%, transparent);
  text-align: center;
  font-size: 0.875rem;
}

.ops-alert-events-card__history-shell {
  padding: var(--theme-ops-panel-padding);
  border: 1px solid color-mix(in srgb, var(--theme-page-border) 74%, transparent);
}

.ops-alert-events-card__state--compact {
  padding-block: var(--theme-ops-alert-events-state-padding-compact);
}

.ops-alert-events-card__state--regular {
  padding-block: var(--theme-ops-alert-events-state-padding-regular);
}

.ops-alert-events-card__state--history {
  padding-block: var(--theme-ops-alert-events-state-padding-history);
}

.ops-alert-events-card__state--center {
  text-align: center;
  font-size: 0.875rem;
}

.ops-alert-events-card__state--loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--theme-ops-alert-events-loading-gap);
}

.ops-alert-events-card__table-scroll {
  max-height: var(--theme-ops-table-max-height);
  overflow: auto;
}

.ops-alert-events-card__table {
  min-width: var(--theme-ops-table-min-width);
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.ops-alert-events-card__table-header--regular,
.ops-alert-events-card__table-cell--regular {
  padding:
    var(--theme-ops-table-cell-padding-y)
    var(--theme-ops-table-cell-padding-x);
}

.ops-alert-events-card__table-header--compact,
.ops-alert-events-card__table-cell--compact {
  padding:
    var(--theme-ops-table-cell-padding-compact-y)
    var(--theme-ops-table-cell-padding-compact-x);
}

.ops-alert-events-card__table-head {
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface));
  position: sticky;
  top: 0;
  z-index: 10;
}

.ops-alert-events-card__table-body {
  background: var(--theme-surface);
}

.ops-alert-events-card__table-body :deep(tr + tr) td {
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 70%, transparent);
}

.ops-alert-events-card__row:hover {
  background: color-mix(in srgb, var(--theme-table-row-hover) 64%, var(--theme-surface));
}

.ops-alert-events-card__row {
  cursor: pointer;
}

.ops-alert-events-card__badge-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.ops-alert-events-card__cell-primary,
.ops-alert-events-card__cell-secondary,
.ops-alert-events-card__email-state {
  color: var(--theme-page-text);
}

.ops-alert-events-card__email-state {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.375rem;
}

.ops-alert-events-card__cell-secondary {
  color: color-mix(in srgb, var(--theme-page-text) 72%, var(--theme-page-muted));
}

.ops-alert-events-card__title-cell {
  min-width: var(--theme-ops-alert-events-title-min-width);
}

.ops-alert-events-card__title-line {
  max-width: var(--theme-ops-alert-events-title-max-width);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ops-alert-events-card__title-line--strong {
  font-size: 0.75rem;
  font-weight: 600;
}

.ops-alert-events-card__badge {
  justify-content: center;
  min-width: var(--theme-ops-alert-events-badge-min-width);
}

.ops-alert-events-card__mono {
  font-family: var(--theme-font-mono);
}

.ops-alert-events-card__table-cell--nowrap {
  white-space: nowrap;
}

.ops-alert-events-card__table-cell--end,
.ops-alert-events-card__table-header--end {
  text-align: right;
}

.ops-alert-events-card__table-cell--xs,
.ops-alert-events-card__table-header {
  font-size: 0.75rem;
}

.ops-alert-events-card__table-cell--xxs {
  font-size: 0.6875rem;
}

.ops-alert-events-card__table-header {
  text-align: left;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.ops-alert-events-card__email-state--sent {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.ops-alert-events-card__email-state--idle {
  color: var(--theme-page-muted);
}

.ops-alert-events-card__email-icon {
  color: currentColor;
}

.ops-alert-events-card__email-label {
  font-size: 0.6875rem;
  font-weight: 700;
}

.ops-alert-events-card__subtitle--clamped {
  margin-top: 0.125rem;
  font-size: 0.6875rem;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.ops-alert-events-card__panel {
  padding: var(--theme-ops-panel-padding);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.ops-alert-events-card__action-group,
.ops-alert-events-card__link {
  border: 1px solid color-mix(in srgb, var(--theme-page-border) 72%, transparent);
  border-radius: var(--theme-button-radius);
  background: var(--theme-surface);
}

.ops-alert-events-card__detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-alert-events-detail-gap);
}

.ops-alert-events-card__detail-header {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ops-alert-events-card__detail-title {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.ops-alert-events-card__detail-description {
  margin-top: 0.25rem;
  white-space: pre-wrap;
  font-size: 0.75rem;
}

.ops-alert-events-card__detail-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.ops-alert-events-card__action-group {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: var(--theme-ops-alert-events-link-padding-y) var(--theme-ops-alert-events-link-padding-x);
}

.ops-alert-events-card__detail-label {
  font-size: 0.6875rem;
  font-weight: 700;
}

.ops-alert-events-card__detail-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--theme-ops-alert-events-detail-grid-gap);
}

.ops-alert-events-card__detail-kicker {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.ops-alert-events-card__detail-value {
  margin-top: 0.25rem;
  font-size: 0.875rem;
}

.ops-alert-events-card__detail-links {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.ops-alert-events-card__detail-rule-id {
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-alert-events-card__link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: var(--theme-ops-alert-events-link-padding-y) var(--theme-ops-alert-events-link-padding-x);
  color: var(--theme-page-text);
  font-size: 0.6875rem;
  font-weight: 700;
}

.ops-alert-events-card__link:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 84%, var(--theme-surface));
}

.ops-alert-events-card__history-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.ops-alert-events-card__history-title {
  font-size: 0.875rem;
  font-weight: 700;
}

.ops-alert-events-card__subtitle--history {
  margin-top: 0.125rem;
  font-size: 0.75rem;
}

@media (min-width: 640px) {
  .ops-alert-events-card__header,
  .ops-alert-events-card__detail-header {
    flex-direction: row;
    align-items: flex-start;
    justify-content: space-between;
  }

  .ops-alert-events-card__detail-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
