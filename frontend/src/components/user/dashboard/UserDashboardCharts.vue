<template>
  <div class="space-y-6">
    <!-- Date Range Filter -->
    <div class="user-dashboard-charts__filters card">
      <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:gap-4">
        <div class="flex items-center gap-2">
          <span class="user-dashboard-charts__label text-sm font-medium">{{ t('dashboard.timeRange') }}:</span>
          <DateRangePicker :start-date="startDate" :end-date="endDate" @update:startDate="$emit('update:startDate', $event)" @update:endDate="$emit('update:endDate', $event)" @change="$emit('dateRangeChange', $event)" />
        </div>
        <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary">
          {{ t('common.refresh') }}
        </button>
        <div class="flex items-center gap-2 sm:ml-auto">
          <span class="user-dashboard-charts__label text-sm font-medium">{{ t('dashboard.granularity') }}:</span>
          <div class="w-28">
            <Select :model-value="granularity" :options="[{value:'day', label:t('dashboard.day')}, {value:'hour', label:t('dashboard.hour')}]" @update:model-value="$emit('update:granularity', $event)" @change="$emit('granularityChange')" />
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Model Distribution Chart -->
      <div class="user-dashboard-charts__panel card relative overflow-hidden">
        <div v-if="loading" class="user-dashboard-charts__overlay absolute inset-0 z-10 flex items-center justify-center backdrop-blur-sm">
          <LoadingSpinner size="md" />
        </div>
        <h3 class="user-dashboard-charts__title mb-4 text-sm font-semibold">{{ t('dashboard.modelDistribution') }}</h3>
        <div class="flex flex-col items-center gap-4 sm:flex-row sm:items-start sm:gap-6">
          <div class="h-48 w-48 flex-shrink-0">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div v-else class="user-dashboard-charts__muted flex h-full items-center justify-center text-sm">{{ t('dashboard.noDataAvailable') }}</div>
          </div>
          <div class="user-dashboard-charts__legend min-w-0 flex-1 self-stretch overflow-y-auto">
            <table class="w-full text-xs">
              <thead>
                <tr class="user-dashboard-charts__table-head">
                  <th class="pb-2 text-left">{{ t('dashboard.model') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.requests') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.tokens') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.actual') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.standard') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="model in models" :key="model.model" class="user-dashboard-charts__table-row">
                  <td class="user-dashboard-charts__cell user-dashboard-charts__cell--compact user-dashboard-charts__cell--model truncate font-medium" :title="model.model">{{ model.model }}</td>
                  <td class="user-dashboard-charts__cell user-dashboard-charts__cell--compact text-right">{{ formatNumber(model.requests) }}</td>
                  <td class="user-dashboard-charts__cell user-dashboard-charts__cell--compact text-right">{{ formatTokens(model.total_tokens) }}</td>
                  <td class="user-dashboard-charts__cell user-dashboard-charts__cell--compact user-dashboard-charts__cell--success text-right">${{ formatCost(model.actual_cost) }}</td>
                  <td class="user-dashboard-charts__cell user-dashboard-charts__cell--compact user-dashboard-charts__cell--muted text-right">${{ formatCost(model.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Token Usage Trend Chart -->
      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler } from 'chart.js'
import { getThemeChartSequence, getThemeChartTooltipColors } from '@/utils/themeStyles'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ loading: boolean, startDate: string, endDate: string, granularity: string, trend: TrendDataPoint[], models: ModelStat[] }>()
defineEmits(['update:startDate', 'update:endDate', 'update:granularity', 'dateRangeChange', 'granularityChange', 'refresh'])
const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()

const chartPalette = computed(() => {
  void themeVersion.value
  return getThemeChartSequence()
})

const tooltipColors = computed(() => {
  void themeVersion.value
  return getThemeChartTooltipColors()
})

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: chartPalette.value
  }]
})

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: tooltipColors.value.background,
      titleColor: tooltipColors.value.text,
      bodyColor: tooltipColors.value.text,
      callbacks: {
        label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens`
      }
    }
  }
}))
</script>

<style scoped>
.user-dashboard-charts__label,
.user-dashboard-charts__title,
.user-dashboard-charts__cell--model {
  color: var(--theme-page-text);
}

.user-dashboard-charts__filters,
.user-dashboard-charts__panel {
  padding: var(--theme-user-dashboard-charts-card-padding);
}

.user-dashboard-charts__overlay {
  background: color-mix(in srgb, var(--theme-page-backdrop) 82%, transparent);
}

.user-dashboard-charts__legend {
  max-height: var(--theme-user-dashboard-charts-legend-max-height);
}

.user-dashboard-charts__muted,
.user-dashboard-charts__table-head,
.user-dashboard-charts__cell,
.user-dashboard-charts__cell--muted {
  color: var(--theme-page-muted);
}

.user-dashboard-charts__cell--compact {
  padding-block: var(--theme-user-dashboard-charts-table-cell-padding-y);
}

.user-dashboard-charts__cell--model {
  max-width: var(--theme-user-dashboard-charts-model-cell-max-width);
}

.user-dashboard-charts__table-row {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.user-dashboard-charts__cell--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}
</style>
