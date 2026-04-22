<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, BarElement, CategoryScale, Legend, LinearScale, Tooltip } from 'chart.js'
import { Bar } from 'vue-chartjs'
import type { OpsLatencyHistogramResponse } from '@/api/admin/ops'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import type { ChartState } from '../types'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { readThemeCssVariable, readThemeRgb } from '@/utils/themeStyles'

ChartJS.register(BarElement, CategoryScale, LinearScale, Tooltip, Legend)

interface Props {
  latencyData: OpsLatencyHistogramResponse | null
  loading: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()

const colors = computed(() => {
  void themeVersion.value

  return {
    info: readThemeRgb('--theme-info-rgb'),
    grid: readThemeCssVariable('--theme-card-border'),
    text: readThemeCssVariable('--theme-page-muted')
  }
})

const hasData = computed(() => (props.latencyData?.total_requests ?? 0) > 0)

const state = computed<ChartState>(() => {
  if (hasData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

const chartData = computed(() => {
  if (!props.latencyData || !hasData.value) return null
  const c = colors.value
  return {
    labels: props.latencyData.buckets.map((b) => b.range),
    datasets: [
      {
        label: t('admin.ops.requests'),
        data: props.latencyData.buckets.map((b) => b.count),
        backgroundColor: c.info,
        borderRadius: 4,
        barPercentage: 0.6
      }
    ]
  }
})

const options = computed(() => {
  const c = colors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false }
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: { color: c.text, font: { size: 10 } }
      },
      y: {
        beginAtZero: true,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: { color: c.text, font: { size: 10 } }
      }
    }
  }
})
</script>

<template>
  <div class="ops-chart-card">
    <div class="ops-chart-card__header">
      <h3 class="ops-chart-card__title">
        <svg class="ops-chart-card__icon ops-chart-card__icon--brand-purple" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        {{ t('admin.ops.latencyHistogram') }}
        <HelpTooltip :content="t('admin.ops.tooltips.latencyHistogram')" />
      </h3>
    </div>

    <div class="ops-chart-card__content">
      <Bar v-if="state === 'ready' && chartData" :data="chartData" :options="options" />
      <div v-else class="ops-chart-card__state">
        <div v-if="state === 'loading'" class="ops-chart-card__placeholder ops-chart-card__placeholder--loading">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyRequest')" />
      </div>
    </div>
  </div>
</template>
