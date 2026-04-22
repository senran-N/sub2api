<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, ArcElement, Legend, Tooltip } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import type { OpsErrorDistributionResponse } from '@/api/admin/ops'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import type { ChartState } from '../types'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { readThemeCssVariable, readThemeRgb } from '@/utils/themeStyles'

ChartJS.register(ArcElement, Tooltip, Legend)

interface Props {
  data: OpsErrorDistributionResponse | null
  loading: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'openDetails'): void
}>()
const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()

const colors = computed(() => {
  void themeVersion.value

  return {
    info: readThemeRgb('--theme-info-rgb'),
    danger: readThemeRgb('--theme-danger-rgb'),
    warning: readThemeRgb('--theme-warning-rgb'),
    muted: readThemeCssVariable('--theme-page-muted'),
    tooltipBg: readThemeCssVariable('--theme-surface-contrast'),
    tooltipText: readThemeCssVariable('--theme-surface-contrast-text')
  }
})

const hasData = computed(() => (props.data?.total ?? 0) > 0)

const state = computed<ChartState>(() => {
  if (hasData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

interface ErrorCategory {
  label: string
  count: number
  color: string
}

const categories = computed<ErrorCategory[]>(() => {
  if (!props.data) return []

  let upstream = 0 // 502, 503, 504
  let client = 0 // 4xx
  let system = 0 // 500
  let other = 0

  for (const item of props.data.items || []) {
    const code = Number(item.status_code || 0)
    const count = Number(item.total || 0)
    if (!Number.isFinite(code) || !Number.isFinite(count)) continue

    if ([502, 503, 504].includes(code)) upstream += count
    else if (code >= 400 && code < 500) client += count
    else if (code === 500) system += count
    else other += count
  }

  const out: ErrorCategory[] = []
  if (upstream > 0) out.push({ label: t('admin.ops.upstream'), count: upstream, color: colors.value.warning })
  if (client > 0) out.push({ label: t('admin.ops.client'), count: client, color: colors.value.info })
  if (system > 0) out.push({ label: t('admin.ops.system'), count: system, color: colors.value.danger })
  if (other > 0) out.push({ label: t('admin.ops.other'), count: other, color: colors.value.muted })
  return out
})

const topReason = computed(() => {
  if (categories.value.length === 0) return null
  return categories.value.reduce((prev, cur) => (cur.count > prev.count ? cur : prev))
})

const chartData = computed(() => {
  if (!hasData.value || categories.value.length === 0) return null
  return {
    labels: categories.value.map((c) => c.label),
    datasets: [
      {
        data: categories.value.map((c) => c.count),
        backgroundColor: categories.value.map((c) => c.color),
        borderWidth: 0
      }
    ]
  }
})

const options = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: colors.value.tooltipBg,
      titleColor: colors.value.tooltipText,
      bodyColor: colors.value.tooltipText
    }
  }
}))
</script>

<template>
  <div class="ops-chart-card">
    <div class="ops-chart-card__header">
      <h3 class="ops-chart-card__title">
        <svg class="ops-chart-card__icon ops-chart-card__icon--danger" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        {{ t('admin.ops.errorDistribution') }}
        <HelpTooltip :content="t('admin.ops.tooltips.errorDistribution')" />
      </h3>
      <button
        type="button"
        class="ops-chart-card__action"
        :disabled="state !== 'ready'"
        :title="t('admin.ops.errorTrend')"
        @click="emit('openDetails')"
      >
        {{ t('admin.ops.requestDetails.details') }}
      </button>
    </div>

    <div class="ops-chart-card__content ops-chart-card__content--relative">
      <div v-if="state === 'ready' && chartData" class="ops-chart-card__summary-shell">
        <div class="ops-chart-card__summary-chart">
          <Doughnut :data="chartData" :options="{ ...options, cutout: '65%' }" />
        </div>
        <div class="ops-chart-card__summary-footer">
          <div v-if="topReason" class="ops-chart-card__summary-top">
            {{ t('admin.ops.top') }}: <span :style="{ color: topReason.color }">{{ topReason.label }}</span>
          </div>
          <div class="ops-chart-card__summary-list">
            <div v-for="item in categories" :key="item.label" class="ops-chart-card__summary-item">
              <span class="ops-chart-card__summary-dot" :style="{ backgroundColor: item.color }"></span>
              <span class="ops-chart-card__summary-count">{{ item.count }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="ops-chart-card__state">
        <div v-if="state === 'loading'" class="ops-chart-card__placeholder ops-chart-card__placeholder--loading">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyError')" />
      </div>
    </div>
  </div>
</template>
