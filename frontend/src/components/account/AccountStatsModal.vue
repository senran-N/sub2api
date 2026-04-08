<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.usageStatistics')"
    width="extra-wide"
    @close="handleClose"
  >
    <div class="account-stats-modal__layout">
      <div v-if="account" class="account-stats-modal__header">
        <div class="account-stats-modal__header-main">
          <div class="account-stats-modal__hero-icon">
            <Icon name="chartBar" size="md" class="account-stats-modal__hero-icon-symbol" :stroke-width="2" />
          </div>
          <div>
            <div class="account-stats-modal__account-name font-semibold">{{ account.name }}</div>
            <div class="account-stats-modal__account-meta text-xs">
              {{ t('admin.accounts.last30DaysUsage') }}
            </div>
          </div>
        </div>
        <span :class="getAccountStatusClasses(account.status)">
          {{ account.status }}
        </span>
      </div>

      <div v-if="loading" class="account-stats-modal__loading-state">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
          <div :class="getSummaryCardClasses('success')">
            <div class="mb-2 flex items-center justify-between">
              <span class="account-stats-modal__label text-xs font-medium">
                {{ t('admin.accounts.stats.totalCost') }}
              </span>
              <div :class="getToneIconClasses('success')">
                <svg
                  class="account-stats-modal__tone-icon-symbol h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
            </div>
            <p class="account-stats-modal__value text-2xl font-bold">
              ${{ formatCost(stats.summary.total_cost) }}
            </p>
            <p class="account-stats-modal__meta mt-1 text-xs">
              {{ t('admin.accounts.stats.accumulatedCost') }}
              <span class="account-stats-modal__meta-detail">
                ({{ t('usage.userBilled') }}: ${{ formatCost(stats.summary.total_user_cost) }} ·
                {{ t('admin.accounts.stats.standardCost') }}:
                ${{ formatCost(stats.summary.total_standard_cost) }})
              </span>
            </p>
          </div>

          <div :class="getSummaryCardClasses('info')">
            <div class="mb-2 flex items-center justify-between">
              <span class="account-stats-modal__label text-xs font-medium">
                {{ t('admin.accounts.stats.totalRequests') }}
              </span>
              <div :class="getToneIconClasses('info')">
                <Icon
                  name="bolt"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
            </div>
            <p class="account-stats-modal__value text-2xl font-bold">
              {{ formatNumber(stats.summary.total_requests) }}
            </p>
            <p class="account-stats-modal__meta mt-1 text-xs">
              {{ t('admin.accounts.stats.totalCalls') }}
            </p>
          </div>

          <div :class="getSummaryCardClasses('warning')">
            <div class="mb-2 flex items-center justify-between">
              <span class="account-stats-modal__label text-xs font-medium">
                {{ t('admin.accounts.stats.avgDailyCost') }}
              </span>
              <div :class="getToneIconClasses('warning')">
                <Icon
                  name="calculator"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
            </div>
            <p class="account-stats-modal__value text-2xl font-bold">
              ${{ formatCost(stats.summary.avg_daily_cost) }}
            </p>
            <p class="account-stats-modal__meta mt-1 text-xs">
              {{
                t('admin.accounts.stats.basedOnActualDays', {
                  days: stats.summary.actual_days_used
                })
              }}
              <span class="account-stats-modal__meta-detail">
                ({{ t('usage.userBilled') }}: ${{ formatCost(stats.summary.avg_daily_user_cost) }})
              </span>
            </p>
          </div>

          <div :class="getSummaryCardClasses('purple')">
            <div class="mb-2 flex items-center justify-between">
              <span class="account-stats-modal__label text-xs font-medium">
                {{ t('admin.accounts.stats.avgDailyRequests') }}
              </span>
              <div :class="getToneIconClasses('purple')">
                <svg
                  class="account-stats-modal__tone-icon-symbol h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"
                  />
                </svg>
              </div>
            </div>
            <p class="account-stats-modal__value text-2xl font-bold">
              {{ formatNumber(Math.round(stats.summary.avg_daily_requests)) }}
            </p>
            <p class="account-stats-modal__meta mt-1 text-xs">
              {{ t('admin.accounts.stats.avgDailyUsage') }}
            </p>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
          <div class="account-stats-modal__panel card">
            <div class="mb-3 flex items-center gap-2">
              <div :class="getToneIconClasses('info')">
                <Icon
                  name="clock"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
              <span class="account-stats-modal__section-title text-sm font-semibold">
                {{ t('admin.accounts.stats.todayOverview') }}
              </span>
            </div>
            <div class="space-y-2">
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.accountBilled') }}</span>
                <span :class="getRowValueClasses()">${{ formatCost(stats.summary.today?.cost || 0) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.userBilled') }}</span>
                <span :class="getRowValueClasses()">${{ formatCost(stats.summary.today?.user_cost || 0) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.requests') }}
                </span>
                <span :class="getRowValueClasses()">{{ formatNumber(stats.summary.today?.requests || 0) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.tokens') }}
                </span>
                <span :class="getRowValueClasses()">{{ formatTokens(stats.summary.today?.tokens || 0) }}</span>
              </div>
            </div>
          </div>

          <div class="account-stats-modal__panel card">
            <div class="mb-3 flex items-center gap-2">
              <div :class="getToneIconClasses('orange')">
                <Icon
                  name="fire"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
              <span class="account-stats-modal__section-title text-sm font-semibold">
                {{ t('admin.accounts.stats.highestCostDay') }}
              </span>
            </div>
            <div class="space-y-2">
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.date') }}
                </span>
                <span :class="getRowValueClasses()">{{ stats.summary.highest_cost_day?.label || '-' }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.accountBilled') }}</span>
                <span :class="getRowValueClasses('orange')">
                  ${{ formatCost(stats.summary.highest_cost_day?.cost || 0) }}
                </span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.userBilled') }}</span>
                <span :class="getRowValueClasses()">
                  ${{ formatCost(stats.summary.highest_cost_day?.user_cost || 0) }}
                </span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.requests') }}
                </span>
                <span :class="getRowValueClasses()">
                  {{ formatNumber(stats.summary.highest_cost_day?.requests || 0) }}
                </span>
              </div>
            </div>
          </div>

          <div class="account-stats-modal__panel card">
            <div class="mb-3 flex items-center gap-2">
              <div :class="getToneIconClasses('purple')">
                <Icon
                  name="trendingUp"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
              <span class="account-stats-modal__section-title text-sm font-semibold">
                {{ t('admin.accounts.stats.highestRequestDay') }}
              </span>
            </div>
            <div class="space-y-2">
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.date') }}
                </span>
                <span :class="getRowValueClasses()">{{ stats.summary.highest_request_day?.label || '-' }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.requests') }}
                </span>
                <span :class="getRowValueClasses('purple')">
                  {{ formatNumber(stats.summary.highest_request_day?.requests || 0) }}
                </span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.accountBilled') }}</span>
                <span :class="getRowValueClasses()">
                  ${{ formatCost(stats.summary.highest_request_day?.cost || 0) }}
                </span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.userBilled') }}</span>
                <span :class="getRowValueClasses()">
                  ${{ formatCost(stats.summary.highest_request_day?.user_cost || 0) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
          <div class="account-stats-modal__panel card">
            <div class="mb-3 flex items-center gap-2">
              <div :class="getToneIconClasses('success')">
                <Icon
                  name="cube"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
              <span class="account-stats-modal__section-title text-sm font-semibold">
                {{ t('admin.accounts.stats.accumulatedTokens') }}
              </span>
            </div>
            <div class="space-y-2">
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.totalTokens') }}
                </span>
                <span :class="getRowValueClasses()">{{ formatTokens(stats.summary.total_tokens) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.dailyAvgTokens') }}
                </span>
                <span :class="getRowValueClasses()">
                  {{ formatTokens(Math.round(stats.summary.avg_daily_tokens)) }}
                </span>
              </div>
            </div>
          </div>

          <div class="account-stats-modal__panel card">
            <div class="mb-3 flex items-center gap-2">
              <div :class="getToneIconClasses('rose')">
                <Icon
                  name="bolt"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
              <span class="account-stats-modal__section-title text-sm font-semibold">
                {{ t('admin.accounts.stats.performance') }}
              </span>
            </div>
            <div class="space-y-2">
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.avgResponseTime') }}
                </span>
                <span :class="getRowValueClasses()">{{ formatDuration(stats.summary.avg_duration_ms) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.daysActive') }}
                </span>
                <span :class="getRowValueClasses()">
                  {{ stats.summary.actual_days_used }} / {{ stats.summary.days }}
                </span>
              </div>
            </div>
          </div>

          <div class="account-stats-modal__panel card">
            <div class="mb-3 flex items-center gap-2">
              <div :class="getToneIconClasses('success')">
                <Icon
                  name="clipboard"
                  size="sm"
                  class="account-stats-modal__tone-icon-symbol"
                  :stroke-width="2"
                />
              </div>
              <span class="account-stats-modal__section-title text-sm font-semibold">
                {{ t('admin.accounts.stats.recentActivity') }}
              </span>
            </div>
            <div class="space-y-2">
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.todayRequests') }}
                </span>
                <span :class="getRowValueClasses()">{{ formatNumber(stats.summary.today?.requests || 0) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">
                  {{ t('admin.accounts.stats.todayTokens') }}
                </span>
                <span :class="getRowValueClasses()">{{ formatTokens(stats.summary.today?.tokens || 0) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.accountBilled') }}</span>
                <span :class="getRowValueClasses()">${{ formatCost(stats.summary.today?.cost || 0) }}</span>
              </div>
              <div class="account-stats-modal__row flex items-center justify-between">
                <span class="account-stats-modal__row-label text-xs">{{ t('usage.userBilled') }}</span>
                <span :class="getRowValueClasses()">
                  ${{ formatCost(stats.summary.today?.user_cost || 0) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="account-stats-modal__panel card">
          <h3 class="account-stats-modal__section-title mb-4 text-sm font-semibold">
            {{ t('admin.accounts.stats.usageTrend') }}
          </h3>
          <div class="h-64">
            <Line v-if="trendChartData" :data="trendChartData" :options="lineChartOptions" />
            <div v-else class="account-stats-modal__chart-empty flex h-full items-center justify-center text-sm">
              {{ t('admin.dashboard.noDataAvailable') }}
            </div>
          </div>
        </div>

        <ModelDistributionChart :model-stats="stats.models" :loading="false" />

        <EndpointDistributionChart
          :endpoint-stats="stats.endpoints || []"
          :loading="false"
          :title="t('usage.inboundEndpoint')"
        />

        <EndpointDistributionChart
          :endpoint-stats="stats.upstream_endpoints || []"
          :loading="false"
          :title="t('usage.upstreamEndpoint')"
        />
      </template>

      <div
        v-else-if="!loading"
        class="account-stats-modal__empty-state"
      >
        <Icon
          name="chartBar"
          size="xl"
          class="account-stats-modal__empty-icon mb-4 h-12 w-12"
          :stroke-width="1.5"
        />
        <p class="text-sm">{{ t('admin.accounts.stats.noData') }}</p>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button class="btn btn-secondary" @click="handleClose">
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip,
  type ChartData,
  type ChartOptions,
  type TooltipItem
} from 'chart.js'
import { Line } from 'vue-chartjs'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import type { Account, AccountUsageStatsResponse } from '@/types'
import {
  getThemeChartTooltipColors,
  getThemeLineChartConfig,
  readThemeCssVariable,
  readThemeRgb,
  readThemeRgbAlpha
} from '@/utils/themeStyles'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

type StatsTone = 'success' | 'info' | 'warning' | 'purple' | 'orange' | 'rose'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(false)
const stats = ref<AccountUsageStatsResponse | null>(null)
const themeVersion = useDocumentThemeVersion()
let activeLoadToken = 0

const lineConfig = computed(() => {
  void themeVersion.value
  return getThemeLineChartConfig()
})

const chartPalette = computed(() => {
  void themeVersion.value

  return {
    text: readThemeCssVariable('--theme-page-text'),
    muted: readThemeCssVariable('--theme-page-muted'),
    grid: readThemeCssVariable('--theme-page-border'),
    accountCost: readThemeRgb('--theme-info-rgb'),
    accountCostFill: readThemeRgbAlpha('--theme-info-rgb', 0.12),
    userCost: readThemeRgb('--theme-success-rgb'),
    userCostFill: readThemeRgbAlpha('--theme-success-rgb', 0.08),
    requests: readThemeRgb('--theme-brand-orange-rgb'),
    requestsFill: readThemeRgbAlpha('--theme-brand-orange-rgb', 0.12)
  }
})

const getSummaryCardClasses = (tone: StatsTone) => [
  'account-stats-modal__summary-card card',
  `account-stats-modal__tone-surface--${tone}`
]

const getToneIconClasses = (tone: StatsTone) => [
  'account-stats-modal__tone-icon',
  `account-stats-modal__tone-icon--${tone}`
]

const getRowValueClasses = (tone?: StatsTone) => [
  'account-stats-modal__row-value text-sm font-semibold',
  tone ? `account-stats-modal__row-value--${tone}` : ''
]

const getAccountStatusClasses = (status: string) => [
  'account-stats-modal__status text-xs font-semibold',
  status === 'active'
    ? 'account-stats-modal__status--active'
    : 'account-stats-modal__status--inactive'
]

const trendChartData = computed<ChartData<'line'> | null>(() => {
  if (!stats.value?.history?.length) {
    return null
  }

  return {
    labels: stats.value.history.map((historyItem) => historyItem.label),
    datasets: [
      {
        label: `${t('usage.accountBilled')} (USD)`,
        data: stats.value.history.map((historyItem) => historyItem.actual_cost),
        borderColor: chartPalette.value.accountCost,
        backgroundColor: chartPalette.value.accountCostFill,
        fill: true,
        tension: 0.3,
        pointRadius: lineConfig.value.pointRadius,
        pointHoverRadius: lineConfig.value.pointHoverRadius,
        yAxisID: 'y'
      },
      {
        label: `${t('usage.userBilled')} (USD)`,
        data: stats.value.history.map((historyItem) => historyItem.user_cost),
        borderColor: chartPalette.value.userCost,
        backgroundColor: chartPalette.value.userCostFill,
        fill: false,
        tension: 0.3,
        borderDash: [5, 5],
        pointRadius: lineConfig.value.pointRadius,
        pointHoverRadius: lineConfig.value.pointHoverRadius,
        yAxisID: 'y'
      },
      {
        label: t('admin.accounts.stats.requests'),
        data: stats.value.history.map((historyItem) => historyItem.requests),
        borderColor: chartPalette.value.requests,
        backgroundColor: chartPalette.value.requestsFill,
        fill: false,
        tension: 0.3,
        pointRadius: lineConfig.value.pointRadius,
        pointHoverRadius: lineConfig.value.pointHoverRadius,
        yAxisID: 'y1'
      }
    ]
  }
})

const lineChartOptions = computed<ChartOptions<'line'>>(() => {
  const tooltipColors = getThemeChartTooltipColors()

  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      intersect: false,
      mode: 'index'
    },
    plugins: {
      legend: {
        position: 'top',
        labels: {
          color: chartPalette.value.text,
          usePointStyle: true,
          pointStyle: 'circle',
          padding: 15,
          font: {
            size: 11
          }
        }
      },
      tooltip: {
        backgroundColor: tooltipColors.background,
        titleColor: tooltipColors.text,
        bodyColor: tooltipColors.text,
        borderColor: chartPalette.value.grid,
        borderWidth: 1,
        callbacks: {
          label: (context: TooltipItem<'line'>) => {
            const label = context.dataset.label || ''
            const value = Number(context.raw ?? 0)

            if (label.includes('USD')) {
              return `${label}: $${formatCost(value)}`
            }

            return `${label}: ${formatNumber(value)}`
          }
        }
      }
    },
    scales: {
      x: {
        grid: {
          color: chartPalette.value.grid
        },
        ticks: {
          color: chartPalette.value.muted,
          font: {
            size: 10
          },
          maxRotation: 45,
          minRotation: 0
        }
      },
      y: {
        type: 'linear',
        display: true,
        position: 'left',
        grid: {
          color: chartPalette.value.grid
        },
        ticks: {
          color: chartPalette.value.accountCost,
          font: {
            size: 10
          },
          callback: (value) => `$${formatCost(Number(value))}`
        },
        title: {
          display: true,
          text: `${t('usage.accountBilled')} (USD)`,
          color: chartPalette.value.accountCost,
          font: {
            size: 11
          }
        }
      },
      y1: {
        type: 'linear',
        display: true,
        position: 'right',
        grid: {
          drawOnChartArea: false
        },
        ticks: {
          color: chartPalette.value.requests,
          font: {
            size: 10
          },
          callback: (value) => formatNumber(Number(value))
        },
        title: {
          display: true,
          text: t('admin.accounts.stats.requests'),
          color: chartPalette.value.requests,
          font: {
            size: 11
          }
        }
      }
    }
  }
})

const loadStats = async (accountId: Account['id']) => {
  const requestToken = ++activeLoadToken
  loading.value = true

  try {
    const response = await adminAPI.accounts.getStats(accountId, 30)

    if (requestToken !== activeLoadToken) {
      return
    }

    stats.value = response
  } catch (error) {
    if (requestToken !== activeLoadToken) {
      return
    }

    console.error('Failed to load account stats:', error)
    stats.value = null
  } finally {
    if (requestToken === activeLoadToken) {
      loading.value = false
    }
  }
}

watch(
  () => [props.show, props.account?.id] as const,
  async ([isVisible, accountId]) => {
    if (!isVisible || accountId == null) {
      activeLoadToken += 1
      loading.value = false
      stats.value = null
      return
    }

    await loadStats(accountId)
  },
  { immediate: true }
)

const handleClose = () => {
  emit('close')
}

const formatCost = (value: number): string => {
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)}K`
  }

  if (value >= 1) {
    return value.toFixed(2)
  }

  if (value >= 0.01) {
    return value.toFixed(3)
  }

  return value.toFixed(4)
}

const formatNumber = (value: number): string => {
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  }

  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }

  return value.toLocaleString()
}

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  }

  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  }

  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }

  return value.toLocaleString()
}

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }

  return `${Math.round(ms)}ms`
}
</script>

<style scoped>
.account-stats-modal__layout {
  display: flex;
  flex-direction: column;
  gap: var(--theme-table-layout-gap-lg);
}

.account-stats-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: calc(var(--theme-surface-radius) + 4px);
  padding: var(--theme-table-mobile-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-accent) 24%, var(--theme-card-border));
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 84%, var(--theme-surface)),
      color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface))
    );
}

.account-stats-modal__header-main {
  display: flex;
  align-items: center;
  gap: var(--theme-table-layout-gap);
}

.account-stats-modal__hero-icon {
  display: flex;
  width: var(--theme-stat-icon-size);
  height: var(--theme-stat-icon-size);
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-stat-icon-radius);
  color: var(--theme-filled-text);
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%)
  );
  box-shadow: 0 12px 28px color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

.account-stats-modal__loading-state,
.account-stats-modal__empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--theme-table-mobile-empty-padding) 0;
}

.account-stats-modal__account-name,
.account-stats-modal__value,
.account-stats-modal__section-title,
.account-stats-modal__row-value {
  color: var(--theme-page-text);
}

.account-stats-modal__account-meta,
.account-stats-modal__label,
.account-stats-modal__meta,
.account-stats-modal__row-label,
.account-stats-modal__chart-empty,
.account-stats-modal__empty-state {
  color: var(--theme-page-muted);
}

.account-stats-modal__meta-detail {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.account-stats-modal__status {
  border-radius: 999px;
  padding: 0.25rem 0.625rem;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.account-stats-modal__status--active {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.account-stats-modal__status--inactive {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-page-muted) 84%, var(--theme-page-text));
}

.account-stats-modal__summary-card,
.account-stats-modal__tone-icon {
  --account-stats-tone-rgb: var(--theme-info-rgb);
}

.account-stats-modal__summary-card {
  padding: var(--theme-stat-card-padding);
  border-color: color-mix(in srgb, rgb(var(--account-stats-tone-rgb)) 24%, var(--theme-card-border));
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, rgb(var(--account-stats-tone-rgb)) 10%, var(--theme-surface)),
      color-mix(in srgb, var(--theme-surface) 92%, var(--theme-surface-soft))
    );
}

.account-stats-modal__panel {
  padding: var(--theme-stat-card-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 84%, var(--theme-surface));
}

.account-stats-modal__tone-icon {
  border-radius: var(--theme-stat-icon-radius);
  padding: 0.375rem;
  background: color-mix(in srgb, rgb(var(--account-stats-tone-rgb)) 12%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--account-stats-tone-rgb)) 88%, var(--theme-page-text));
}

.account-stats-modal__tone-surface--success,
.account-stats-modal__tone-icon--success,
.account-stats-modal__row-value--success {
  --account-stats-tone-rgb: var(--theme-success-rgb);
}

.account-stats-modal__tone-surface--info,
.account-stats-modal__tone-icon--info,
.account-stats-modal__row-value--info {
  --account-stats-tone-rgb: var(--theme-info-rgb);
}

.account-stats-modal__tone-surface--warning,
.account-stats-modal__tone-icon--warning,
.account-stats-modal__row-value--warning {
  --account-stats-tone-rgb: var(--theme-warning-rgb);
}

.account-stats-modal__tone-surface--purple,
.account-stats-modal__tone-icon--purple,
.account-stats-modal__row-value--purple {
  --account-stats-tone-rgb: var(--theme-brand-purple-rgb);
}

.account-stats-modal__tone-surface--orange,
.account-stats-modal__tone-icon--orange,
.account-stats-modal__row-value--orange {
  --account-stats-tone-rgb: var(--theme-brand-orange-rgb);
}

.account-stats-modal__tone-surface--rose,
.account-stats-modal__tone-icon--rose,
.account-stats-modal__row-value--rose {
  --account-stats-tone-rgb: var(--theme-brand-rose-rgb);
}

.account-stats-modal__row-value--success,
.account-stats-modal__row-value--info,
.account-stats-modal__row-value--warning,
.account-stats-modal__row-value--purple,
.account-stats-modal__row-value--orange,
.account-stats-modal__row-value--rose {
  color: color-mix(in srgb, rgb(var(--account-stats-tone-rgb)) 84%, var(--theme-page-text));
}

.account-stats-modal__empty-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}
</style>
