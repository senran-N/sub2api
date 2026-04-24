<template>
  <div class="endpoint-distribution-chart__card card">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="endpoint-distribution-chart__title text-sm font-semibold">
        {{ title || t('usage.endpointDistribution') }}
      </h3>
      <div class="flex flex-wrap items-center justify-end gap-2">
        <div
          v-if="showSourceToggle"
          class="segmented-control"
        >
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': source === 'inbound' }"
            @click="emit('update:source', 'inbound')"
          >
            {{ t('usage.inbound') }}
          </button>
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': source === 'upstream' }"
            @click="emit('update:source', 'upstream')"
          >
            {{ t('usage.upstream') }}
          </button>
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': source === 'path' }"
            @click="emit('update:source', 'path')"
          >
            {{ t('usage.path') }}
          </button>
        </div>

        <div
          v-if="showMetricToggle"
          class="segmented-control"
        >
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': metric === 'tokens' }"
            @click="emit('update:metric', 'tokens')"
          >
            {{ t('admin.dashboard.metricTokens') }}
          </button>
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': metric === 'actual_cost' }"
            @click="emit('update:metric', 'actual_cost')"
          >
            {{ t('admin.dashboard.metricActualCost') }}
          </button>
        </div>
      </div>
    </div>
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="displayEndpointStats.length > 0 && chartData" class="flex items-center gap-6">
      <div class="h-48 w-48">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="max-h-48 flex-1 overflow-y-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="endpoint-distribution-chart__table-head">
              <th class="pb-2 text-left">{{ t('usage.endpoint') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.requests') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.tokens') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.actual') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="item in displayEndpointStats" :key="item.endpoint">
              <tr
                class="endpoint-distribution-chart__table-row endpoint-distribution-chart__table-row--interactive transition-colors"
                @click="toggleBreakdown(item.endpoint)"
              >
                <td class="endpoint-distribution-chart__name endpoint-distribution-chart__name--interactive endpoint-distribution-chart__cell endpoint-distribution-chart__cell--name truncate font-medium" :title="item.endpoint">
                  <span class="inline-flex items-center gap-1">
                    <svg v-if="expandedKey === item.endpoint" class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
                    <svg v-else class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                    {{ item.endpoint }}
                  </span>
                </td>
                <td class="endpoint-distribution-chart__muted endpoint-distribution-chart__cell text-right">
                  {{ formatNumber(item.requests) }}
                </td>
                <td class="endpoint-distribution-chart__muted endpoint-distribution-chart__cell text-right">
                  {{ formatTokens(item.total_tokens) }}
                </td>
                <td class="endpoint-distribution-chart__actual endpoint-distribution-chart__cell text-right">
                  ${{ formatCost(item.actual_cost) }}
                </td>
                <td class="endpoint-distribution-chart__standard endpoint-distribution-chart__cell text-right">
                  ${{ formatCost(item.cost) }}
                </td>
              </tr>
              <tr v-if="expandedKey === item.endpoint">
                <td colspan="5" class="endpoint-distribution-chart__breakdown-cell">
                  <UserBreakdownSubTable
                    :items="breakdownItems"
                    :loading="breakdownLoading"
                  />
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
    <div v-else class="endpoint-distribution-chart__muted flex h-48 items-center justify-center text-sm">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import UserBreakdownSubTable from './UserBreakdownSubTable.vue'
import type { EndpointStat, UserBreakdownItem } from '@/types'
import { getUserBreakdown } from '@/api/admin/dashboard'
import { getThemeChartSequence, getThemeChartTooltipColors } from '@/utils/themeStyles'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()

type DistributionMetric = 'tokens' | 'actual_cost'
type EndpointSource = 'inbound' | 'upstream' | 'path'

const props = withDefaults(
  defineProps<{
    endpointStats: EndpointStat[]
    upstreamEndpointStats?: EndpointStat[]
    endpointPathStats?: EndpointStat[]
    loading?: boolean
    title?: string
    metric?: DistributionMetric
    source?: EndpointSource
    showMetricToggle?: boolean
    showSourceToggle?: boolean
    startDate?: string
    endDate?: string
    filters?: Record<string, any>
  }>(),
  {
    upstreamEndpointStats: () => [],
    endpointPathStats: () => [],
    loading: false,
    title: '',
    metric: 'tokens',
    source: 'inbound',
    showMetricToggle: false,
    showSourceToggle: false
  }
)

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
  'update:source': [value: EndpointSource]
}>()

const expandedKey = ref<string | null>(null)
const breakdownItems = ref<UserBreakdownItem[]>([])
const breakdownLoading = ref(false)

const toggleBreakdown = async (endpoint: string) => {
  if (expandedKey.value === endpoint) {
    expandedKey.value = null
    return
  }
  expandedKey.value = endpoint
  breakdownLoading.value = true
  breakdownItems.value = []
  try {
    const res = await getUserBreakdown({
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      endpoint,
      endpoint_type: props.source,
    })
    breakdownItems.value = res.users || []
  } catch {
    breakdownItems.value = []
  } finally {
    breakdownLoading.value = false
  }
}

const chartColors = computed(() => {
  void themeVersion.value
  return getThemeChartSequence()
})

const displayEndpointStats = computed(() => {
  const sourceStats = props.source === 'upstream'
    ? props.upstreamEndpointStats
    : props.source === 'path'
      ? props.endpointPathStats
      : props.endpointStats
  if (!sourceStats?.length) return []

  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...sourceStats].sort((a, b) => b[metricKey] - a[metricKey])
})

const chartData = computed(() => {
  if (!displayEndpointStats.value?.length) return null

  return {
    labels: displayEndpointStats.value.map((item) => item.endpoint),
    datasets: [
      {
        data: displayEndpointStats.value.map((item) =>
          props.metric === 'actual_cost' ? item.actual_cost : item.total_tokens
        ),
        backgroundColor: chartColors.value.slice(0, displayEndpointStats.value.length),
        borderWidth: 0
      }
    ]
  }
})

const doughnutOptions = computed(() => {
  void themeVersion.value
  const tooltipColors = getThemeChartTooltipColors()

  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false
      },
      tooltip: {
        backgroundColor: tooltipColors.background,
        titleColor: tooltipColors.text,
        bodyColor: tooltipColors.text,
        callbacks: {
          label: (context: any) => {
            const value = context.raw as number
            const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
            const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
            const formattedValue = props.metric === 'actual_cost'
              ? `$${formatCost(value)}`
              : formatTokens(value)
            return `${context.label}: ${formattedValue} (${percentage}%)`
          }
        }
      }
    }
  }
})

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatNumber = (value: number): string => {
  return value.toLocaleString()
}

const formatCost = (value: number): string => {
  if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  } else if (value >= 1) {
    return value.toFixed(2)
  } else if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}
</script>

<style scoped>
.endpoint-distribution-chart__card {
  padding: var(--theme-endpoint-distribution-card-padding);
}

.endpoint-distribution-chart__title,
.endpoint-distribution-chart__name {
  color: var(--theme-page-text);
}

.endpoint-distribution-chart__muted,
.endpoint-distribution-chart__table-head,
.endpoint-distribution-chart__standard {
  color: var(--theme-page-muted);
}

.endpoint-distribution-chart__table-row {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.endpoint-distribution-chart__cell {
  padding-block: var(--theme-endpoint-distribution-cell-padding-y);
}

.endpoint-distribution-chart__cell--name {
  max-width: var(--theme-endpoint-distribution-name-max-width);
}

.endpoint-distribution-chart__breakdown-cell {
  padding: var(--theme-endpoint-distribution-breakdown-cell-padding);
}

.endpoint-distribution-chart__table-row--interactive {
  cursor: pointer;
}

.endpoint-distribution-chart__table-row--interactive:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 64%, var(--theme-surface));
}

.endpoint-distribution-chart__name--interactive {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.endpoint-distribution-chart__name--interactive:hover {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 94%, var(--theme-page-text));
}

.endpoint-distribution-chart__actual {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}
</style>
