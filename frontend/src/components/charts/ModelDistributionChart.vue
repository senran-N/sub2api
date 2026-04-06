<template>
  <div class="model-distribution-chart__card card">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="model-distribution-chart__title">
        {{ !enableRankingView || activeView === 'model_distribution'
          ? t('admin.dashboard.modelDistribution')
          : t('admin.dashboard.spendingRankingTitle') }}
      </h3>
      <div class="flex flex-wrap items-center justify-end gap-2">
        <div
          v-if="showSourceToggle"
          class="segmented-control"
        >
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': source === 'requested' }"
            @click="emit('update:source', 'requested')"
          >
            {{ t('usage.requestedModel') }}
          </button>
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': source === 'upstream' }"
            @click="emit('update:source', 'upstream')"
          >
            {{ t('usage.upstreamModel') }}
          </button>
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': source === 'mapping' }"
            @click="emit('update:source', 'mapping')"
          >
            {{ t('usage.mapping') }}
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
        <div v-if="enableRankingView" class="model-distribution-chart__view-toggle segmented-control border-0">
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': activeView === 'model_distribution' }"
            @click="activeView = 'model_distribution'"
          >
            {{ t('admin.dashboard.viewModelDistribution') }}
          </button>
          <button
            type="button"
            class="segmented-option"
            :class="{ 'segmented-option-active': activeView === 'spending_ranking' }"
            @click="activeView = 'spending_ranking'"
          >
            {{ t('admin.dashboard.viewSpendingRanking') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="activeView === 'model_distribution' && loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="activeView === 'model_distribution' && displayModelStats.length > 0 && chartData"
      class="flex items-center gap-6"
    >
      <div class="h-48 w-48">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="model-distribution-chart__legend flex-1 overflow-y-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="model-distribution-chart__table-head">
              <th class="model-distribution-chart__table-head-cell model-distribution-chart__table-head-cell--left">{{ t('admin.dashboard.model') }}</th>
              <th class="model-distribution-chart__table-head-cell">{{ t('admin.dashboard.requests') }}</th>
              <th class="model-distribution-chart__table-head-cell">{{ t('admin.dashboard.tokens') }}</th>
              <th class="model-distribution-chart__table-head-cell">{{ t('admin.dashboard.actual') }}</th>
              <th class="model-distribution-chart__table-head-cell">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="model in displayModelStats" :key="model.model">
              <tr
                class="model-distribution-chart__row model-distribution-chart__row--interactive"
                @click="toggleBreakdown('model', model.model)"
              >
                <td
                  class="model-distribution-chart__cell model-distribution-chart__cell--link"
                  :title="model.model"
                >
                  <span class="inline-flex items-center gap-1">
                    <svg v-if="expandedKey === `model-${model.model}`" class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
                    <svg v-else class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                    {{ model.model }}
                  </span>
                </td>
                <td class="model-distribution-chart__cell">
                  {{ formatNumber(model.requests) }}
                </td>
                <td class="model-distribution-chart__cell">
                  {{ formatTokens(model.total_tokens) }}
                </td>
                <td class="model-distribution-chart__cell model-distribution-chart__cell--success">
                  ${{ formatCost(model.actual_cost) }}
                </td>
                <td class="model-distribution-chart__cell model-distribution-chart__cell--muted">
                  ${{ formatCost(model.cost) }}
                </td>
              </tr>
              <tr v-if="expandedKey === `model-${model.model}`">
                <td colspan="5" class="model-distribution-chart__subtable-cell">
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
      v-else-if="activeView === 'model_distribution'"
      class="model-distribution-chart__empty-state"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>

    <div v-else-if="rankingLoading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="rankingError"
      class="model-distribution-chart__empty-state"
    >
      {{ t('admin.dashboard.failedToLoad') }}
    </div>
    <div v-else-if="rankingDisplayItems.length > 0 && rankingChartData" class="flex items-center gap-6">
      <div class="h-48 w-48">
        <Doughnut :data="rankingChartData" :options="rankingDoughnutOptions" />
      </div>
      <div class="model-distribution-chart__legend flex-1 overflow-y-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="model-distribution-chart__table-head">
              <th class="model-distribution-chart__table-head-cell model-distribution-chart__table-head-cell--left">{{ t('admin.dashboard.spendingRankingUser') }}</th>
              <th class="model-distribution-chart__table-head-cell">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
              <th class="model-distribution-chart__table-head-cell">{{ t('admin.dashboard.spendingRankingTokens') }}</th>
              <th class="model-distribution-chart__table-head-cell">{{ t('admin.dashboard.spendingRankingSpend') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in rankingDisplayItems"
              :key="item.isOther ? 'others' : `${item.user_id}-${index}`"
              :class="getRankingRowClasses(item.isOther)"
              @click="item.isOther ? undefined : emit('ranking-click', item)"
            >
              <td class="model-distribution-chart__cell model-distribution-chart__cell--left">
                <div class="flex min-w-0 items-center gap-2">
                  <span class="model-distribution-chart__ranking-index">
                    {{ item.isOther ? 'Σ' : `#${index + 1}` }}
                  </span>
                  <span
                    class="model-distribution-chart__ranking-label"
                    :title="getRankingRowLabel(item)"
                  >
                    {{ getRankingRowLabel(item) }}
                  </span>
                </div>
              </td>
              <td class="model-distribution-chart__cell">
                {{ formatNumber(item.requests) }}
              </td>
              <td class="model-distribution-chart__cell">
                {{ formatTokens(item.tokens) }}
              </td>
              <td class="model-distribution-chart__cell model-distribution-chart__cell--success">
                ${{ formatCost(item.actual_cost) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else
      class="model-distribution-chart__empty-state"
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
import type { ModelStat, UserSpendingRankingItem, UserBreakdownItem } from '@/types'
import { getUserBreakdown } from '@/api/admin/dashboard'
import {
  getThemeChartSequence,
  getThemeChartTooltipColors,
  readThemeCssVariable
} from '@/utils/themeStyles'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()

type DistributionMetric = 'tokens' | 'actual_cost'
type ModelSource = 'requested' | 'upstream' | 'mapping'
type RankingDisplayItem = UserSpendingRankingItem & { isOther?: boolean }
const props = withDefaults(defineProps<{
  modelStats: ModelStat[]
  upstreamModelStats?: ModelStat[]
  mappingModelStats?: ModelStat[]
  source?: ModelSource
  enableRankingView?: boolean
  rankingItems?: UserSpendingRankingItem[]
  rankingTotalActualCost?: number
  rankingTotalRequests?: number
  rankingTotalTokens?: number
  loading?: boolean
  metric?: DistributionMetric
  showSourceToggle?: boolean
  showMetricToggle?: boolean
  rankingLoading?: boolean
  rankingError?: boolean
  startDate?: string
  endDate?: string
}>(), {
  upstreamModelStats: () => [],
  mappingModelStats: () => [],
  source: 'requested',
  enableRankingView: false,
  rankingItems: () => [],
  rankingTotalActualCost: 0,
  rankingTotalRequests: 0,
  rankingTotalTokens: 0,
  loading: false,
  metric: 'tokens',
  showSourceToggle: false,
  showMetricToggle: false,
  rankingLoading: false,
  rankingError: false
})

const expandedKey = ref<string | null>(null)
const breakdownItems = ref<UserBreakdownItem[]>([])
const breakdownLoading = ref(false)
const themeVersion = useDocumentThemeVersion()

const toggleBreakdown = async (type: string, id: string) => {
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
      model: id,
      model_source: props.source,
    })
    breakdownItems.value = res.users || []
  } catch {
    breakdownItems.value = []
  } finally {
    breakdownLoading.value = false
  }
}

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
  'update:source': [value: ModelSource]
  'ranking-click': [item: UserSpendingRankingItem]
}>()

const enableRankingView = computed(() => props.enableRankingView)
const activeView = ref<'model_distribution' | 'spending_ranking'>('model_distribution')

const chartPalette = computed(() => {
  void themeVersion.value

  return {
    muted: readThemeCssVariable('--theme-page-muted')
  }
})

const chartColors = computed(() => {
  void themeVersion.value
  return getThemeChartSequence()
})

const displayModelStats = computed(() => {
  const sourceStats = props.source === 'upstream'
    ? props.upstreamModelStats
    : props.source === 'mapping'
      ? props.mappingModelStats
      : props.modelStats
  if (!sourceStats?.length) return []

  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...sourceStats].sort((a, b) => b[metricKey] - a[metricKey])
})

const chartData = computed(() => {
  void themeVersion.value
  if (!displayModelStats.value.length) return null

  return {
    labels: displayModelStats.value.map((m) => m.model),
    datasets: [
      {
        data: displayModelStats.value.map((m) => props.metric === 'actual_cost' ? m.actual_cost : m.total_tokens),
        backgroundColor: chartColors.value.slice(0, displayModelStats.value.length),
        borderWidth: 0
      }
    ]
  }
})

const rankingChartData = computed(() => {
  void themeVersion.value
  if (!props.rankingItems?.length) return null

  const labels = props.rankingItems.map((item, index) => `#${index + 1} ${getRankingUserLabel(item)}`)
  const data = props.rankingItems.map((item) => item.actual_cost)
  const backgroundColor = chartColors.value.slice(0, props.rankingItems.length)

  if (otherRankingItem.value) {
    labels.push(t('admin.dashboard.spendingRankingOther'))
    data.push(otherRankingItem.value.actual_cost)
    backgroundColor.push(chartPalette.value.muted)
  }

  return {
    labels,
    datasets: [
      {
        data,
        backgroundColor,
        borderWidth: 0
      }
    ]
  }
})

const otherRankingItem = computed<RankingDisplayItem | null>(() => {
  if (!props.rankingItems?.length) return null

  const rankedActualCost = props.rankingItems.reduce((sum, item) => sum + item.actual_cost, 0)
  const rankedRequests = props.rankingItems.reduce((sum, item) => sum + item.requests, 0)
  const rankedTokens = props.rankingItems.reduce((sum, item) => sum + item.tokens, 0)

  const otherActualCost = Math.max((props.rankingTotalActualCost || 0) - rankedActualCost, 0)
  const otherRequests = Math.max((props.rankingTotalRequests || 0) - rankedRequests, 0)
  const otherTokens = Math.max((props.rankingTotalTokens || 0) - rankedTokens, 0)

  if (otherActualCost <= 0.000001 && otherRequests <= 0 && otherTokens <= 0) return null

  return {
    user_id: 0,
    email: '',
    actual_cost: otherActualCost,
    requests: otherRequests,
    tokens: otherTokens,
    isOther: true
  }
})

const rankingDisplayItems = computed<RankingDisplayItem[]>(() => {
  if (!props.rankingItems?.length) return []
  return otherRankingItem.value
    ? [...props.rankingItems, otherRankingItem.value]
    : [...props.rankingItems]
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

const rankingDoughnutOptions = computed(() => {
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
            return `${context.label}: $${formatCost(value)} (${percentage}%)`
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

const getRankingUserLabel = (item: UserSpendingRankingItem): string => {
  if (item.email) return item.email
  return t('admin.redeem.userPrefix', { id: item.user_id })
}

const getRankingRowLabel = (item: RankingDisplayItem): string => {
  if (item.isOther) return t('admin.dashboard.spendingRankingOther')
  return getRankingUserLabel(item)
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

const getRankingRowClasses = (isOther?: boolean) => {
  return [
    'model-distribution-chart__row',
    isOther
      ? 'model-distribution-chart__row--other'
      : 'model-distribution-chart__row--interactive'
  ]
}

</script>

<style scoped>
.model-distribution-chart__card {
  padding: var(--theme-user-dashboard-charts-card-padding);
}

.model-distribution-chart__title {
  color: var(--theme-page-text);
  font-size: 0.875rem;
  font-weight: 600;
}

.model-distribution-chart__view-toggle {
  padding: var(--theme-settings-tabs-nav-padding);
}

.model-distribution-chart__table-head {
  color: var(--theme-page-muted);
}

.model-distribution-chart__table-head-cell {
  padding-bottom: 0.5rem;
  text-align: right;
}

.model-distribution-chart__table-head-cell--left {
  text-align: left;
}

.model-distribution-chart__row {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  transition: background-color 0.18s ease;
}

.model-distribution-chart__row--interactive {
  cursor: pointer;
}

.model-distribution-chart__row--interactive:hover {
  background: color-mix(in srgb, var(--theme-button-ghost-hover-bg) 90%, transparent);
}

.model-distribution-chart__row--other {
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}

.model-distribution-chart__cell {
  padding: 0.375rem 0;
  color: var(--theme-page-muted);
  text-align: right;
}

.model-distribution-chart__cell--left {
  text-align: left;
}

.model-distribution-chart__cell--link {
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: rgb(var(--theme-info-rgb));
  font-weight: 600;
}

.model-distribution-chart__cell--link:hover {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 82%, var(--theme-page-text));
}

.model-distribution-chart__legend {
  max-height: var(--theme-user-dashboard-charts-legend-max-height);
}

.model-distribution-chart__subtable-cell {
  padding: 0;
}

.model-distribution-chart__cell--success {
  color: rgb(var(--theme-success-rgb));
}

.model-distribution-chart__cell--muted,
.model-distribution-chart__ranking-index {
  color: var(--theme-page-muted);
}

.model-distribution-chart__ranking-index {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 600;
}

.model-distribution-chart__ranking-label {
  display: block;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--theme-page-text);
  font-weight: 600;
}

.model-distribution-chart__empty-state {
  display: flex;
  height: 12rem;
  align-items: center;
  justify-content: center;
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}
</style>
