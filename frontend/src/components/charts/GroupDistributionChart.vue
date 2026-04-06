<template>
  <div class="card group-distribution-chart__panel">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="group-distribution-chart__title text-sm font-semibold">
        {{ t('admin.dashboard.groupDistribution') }}
      </h3>
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
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="displayGroupStats.length > 0 && chartData" class="flex items-center gap-6">
      <div class="h-48 w-48">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="group-distribution-chart__legend flex-1 overflow-y-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="group-distribution-chart__table-head">
              <th class="pb-2 text-left">{{ t('admin.dashboard.group') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.requests') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.tokens') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.actual') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="group in displayGroupStats" :key="group.group_id">
              <tr
                class="group-distribution-chart__table-row transition-colors"
                :class="{ 'group-distribution-chart__table-row--interactive': group.group_id > 0 }"
                @click="group.group_id > 0 && toggleBreakdown('group', group.group_id)"
              >
                <td
                  class="group-distribution-chart__name group-distribution-chart__cell group-distribution-chart__cell--name truncate font-medium"
                  :class="{ 'group-distribution-chart__name--interactive': group.group_id > 0 }"
                  :title="group.group_name || String(group.group_id)"
                >
                  <span class="inline-flex items-center gap-1">
                    <svg v-if="group.group_id > 0 && expandedKey === `group-${group.group_id}`" class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
                    <svg v-else-if="group.group_id > 0" class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                    {{ group.group_name || t('admin.dashboard.noGroup') }}
                  </span>
                </td>
                <td class="group-distribution-chart__muted group-distribution-chart__cell text-right">
                  {{ formatNumber(group.requests) }}
                </td>
                <td class="group-distribution-chart__muted group-distribution-chart__cell text-right">
                  {{ formatTokens(group.total_tokens) }}
                </td>
                <td class="group-distribution-chart__actual group-distribution-chart__cell text-right">
                  ${{ formatCost(group.actual_cost) }}
                </td>
                <td class="group-distribution-chart__standard group-distribution-chart__cell text-right">
                  ${{ formatCost(group.cost) }}
                </td>
              </tr>
              <!-- User breakdown sub-rows -->
              <tr v-if="expandedKey === `group-${group.group_id}`">
                <td colspan="5" class="group-distribution-chart__breakdown-cell">
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
    <div
      v-else
      class="group-distribution-chart__muted flex h-48 items-center justify-center text-sm"
    >
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
import type { GroupStat, UserBreakdownItem } from '@/types'
import { getUserBreakdown } from '@/api/admin/dashboard'
import { getThemeChartSequence, getThemeChartTooltipColors } from '@/utils/themeStyles'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()

type DistributionMetric = 'tokens' | 'actual_cost'

const props = withDefaults(defineProps<{
  groupStats: GroupStat[]
  loading?: boolean
  metric?: DistributionMetric
  showMetricToggle?: boolean
  startDate?: string
  endDate?: string
}>(), {
  loading: false,
  metric: 'tokens',
  showMetricToggle: false,
})

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
}>()

const expandedKey = ref<string | null>(null)
const breakdownItems = ref<UserBreakdownItem[]>([])
const breakdownLoading = ref(false)

const toggleBreakdown = async (type: string, id: number | string) => {
  const key = `${type}-${id}`
  if (expandedKey.value === key) {
    expandedKey.value = null
    return
  }
  expandedKey.value = key
  breakdownLoading.value = true
  breakdownItems.value = []
  try {
    const res = await getUserBreakdown({
      start_date: props.startDate,
      end_date: props.endDate,
      group_id: Number(id),
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

const displayGroupStats = computed(() => {
  if (!props.groupStats?.length) return []

  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...props.groupStats].sort((a, b) => b[metricKey] - a[metricKey])
})

const chartData = computed(() => {
  if (!props.groupStats?.length) return null

  return {
    labels: displayGroupStats.value.map((g) => g.group_name || String(g.group_id)),
    datasets: [
      {
        data: displayGroupStats.value.map((g) => props.metric === 'actual_cost' ? g.actual_cost : g.total_tokens),
        backgroundColor: chartColors.value.slice(0, displayGroupStats.value.length),
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
.group-distribution-chart__title,
.group-distribution-chart__name {
  color: var(--theme-page-text);
}

.group-distribution-chart__panel {
  padding: var(--theme-group-distribution-card-padding);
}

.group-distribution-chart__legend {
  max-height: var(--theme-group-distribution-legend-max-height);
}

.group-distribution-chart__muted,
.group-distribution-chart__table-head,
.group-distribution-chart__standard {
  color: var(--theme-page-muted);
}

.group-distribution-chart__table-row {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.group-distribution-chart__cell {
  padding-block: var(--theme-group-distribution-cell-padding-y);
}

.group-distribution-chart__cell--name {
  max-width: var(--theme-group-distribution-name-max-width);
}

.group-distribution-chart__breakdown-cell {
  padding: 0;
}

.group-distribution-chart__table-row--interactive {
  cursor: pointer;
}

.group-distribution-chart__table-row--interactive:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 64%, var(--theme-surface));
}

.group-distribution-chart__name--interactive {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.group-distribution-chart__name--interactive:hover {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 94%, var(--theme-page-text));
}

.group-distribution-chart__actual {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}
</style>
