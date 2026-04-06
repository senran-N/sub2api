<template>
  <AppLayout>
    <div class="space-y-6">
      <template v-if="loading">
        <DashboardLoadingSkeleton />
      </template>

      <template v-else-if="stats">
        <DashboardStatsCards
          :stats="stats"
          :format-tokens="formatTokens"
          :format-number="formatNumber"
          :format-cost="formatCost"
          :format-duration="formatDuration"
        />

        <div class="space-y-6">
          <DashboardChartControls
            :start-date="startDate"
            :end-date="endDate"
            :granularity="granularity"
            :granularity-options="granularityOptions"
            :loading="chartsLoading"
            @update:start-date="startDate = $event"
            @update:end-date="endDate = $event"
            @update:granularity="granularity = $event"
            @date-range-change="onDateRangeChange"
            @refresh="loadDashboardStats"
            @granularity-change="loadChartData"
          />

          <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <ModelDistributionChart
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-actual-cost="rankingTotalActualCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToUserUsage"
            />
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>

          <DashboardUserTrendCard
            :loading="userTrendLoading"
            :chart-data="userTrendChartData"
            :chart-options="lineOptions"
          />
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import { useAppStore } from '@/stores/app'
import type { UserSpendingRankingItem } from '@/types'
import { readThemeCssVariable } from '@/utils/themeStyles'
import AppLayout from '@/components/layout/AppLayout.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import DashboardChartControls from './dashboard/DashboardChartControls.vue'
import DashboardLoadingSkeleton from './dashboard/DashboardLoadingSkeleton.vue'
import DashboardStatsCards from './dashboard/DashboardStatsCards.vue'
import DashboardUserTrendCard from './dashboard/DashboardUserTrendCard.vue'
import {
  buildDashboardUserTrendChartData,
  formatDashboardCost,
  formatDashboardDuration,
  formatDashboardNumber,
  formatDashboardTokens
} from './dashboardView'
import { useDashboardViewData } from './useDashboardViewData'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'

const { t } = useI18n()
const themeVersion = useDocumentThemeVersion()

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

const appStore = useAppStore()
const router = useRouter()
const {
  stats,
  loading,
  chartsLoading,
  userTrendLoading,
  rankingLoading,
  rankingError,
  trendData,
  modelStats,
  userTrend,
  rankingItems,
  rankingTotalActualCost,
  rankingTotalRequests,
  rankingTotalTokens,
  startDate,
  endDate,
  granularity,
  granularityOptions,
  loadDashboardStats,
  loadChartData,
  onDateRangeChange
} = useDashboardViewData({
  t,
  showError: appStore.showError
})

// Chart colors
const chartColors = computed(() => {
  void themeVersion.value

  return {
    text: readThemeCssVariable('--theme-page-text'),
    grid: readThemeCssVariable('--theme-page-border')
  }
})

// Line chart options (for user trend chart)
const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      itemSort: (a: any, b: any) => {
        const aValue = typeof a?.raw === 'number' ? a.raw : Number(a?.parsed?.y ?? 0)
        const bValue = typeof b?.raw === 'number' ? b.raw : Number(b?.parsed?.y ?? 0)
        return bValue - aValue
      },
      callbacks: {
        label: (context: any) => {
          return `${context.dataset.label}: ${formatDashboardTokens(context.raw)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        }
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatDashboardTokens(Number(value))
      }
    }
  }
}))

const userTrendChartData = computed(() =>
  buildDashboardUserTrendChartData(userTrend.value, t)
)

const formatTokens = formatDashboardTokens
const formatNumber = formatDashboardNumber
const formatCost = formatDashboardCost
const formatDuration = formatDashboardDuration

const goToUserUsage = (item: UserSpendingRankingItem) => {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(item.user_id),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

onMounted(() => {
  void loadDashboardStats()
})
</script>

<style scoped>
</style>
