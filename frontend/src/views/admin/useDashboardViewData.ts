import { computed, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  ModelStat,
  TrendDataPoint,
  UserSpendingRankingItem,
  UserUsageTrendPoint
} from '@/types'
import {
  getDashboardGranularityForRange,
  getDashboardLast24HoursRangeDates
} from './dashboardView'

const DASHBOARD_RANKING_LIMIT = 12

interface DashboardViewDataOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
}

export function useDashboardViewData(options: DashboardViewDataOptions) {
  const stats = ref<DashboardStats | null>(null)
  const loading = ref(false)
  const chartsLoading = ref(false)
  const userTrendLoading = ref(false)
  const rankingLoading = ref(false)
  const rankingError = ref(false)

  const trendData = ref<TrendDataPoint[]>([])
  const modelStats = ref<ModelStat[]>([])
  const userTrend = ref<UserUsageTrendPoint[]>([])
  const rankingItems = ref<UserSpendingRankingItem[]>([])
  const rankingTotalActualCost = ref(0)
  const rankingTotalRequests = ref(0)
  const rankingTotalTokens = ref(0)

  const defaultRange = getDashboardLast24HoursRangeDates()
  const startDate = ref(defaultRange.start)
  const endDate = ref(defaultRange.end)
  const granularity = ref<'day' | 'hour'>('hour')

  const granularityOptions = computed(() => [
    { value: 'day', label: options.t('admin.dashboard.day') },
    { value: 'hour', label: options.t('admin.dashboard.hour') }
  ])

  let chartLoadSeq = 0
  let usersTrendLoadSeq = 0
  let rankingLoadSeq = 0

  const loadDashboardSnapshot = async (includeStats: boolean) => {
    const currentSeq = ++chartLoadSeq
    if (includeStats && !stats.value) {
      loading.value = true
    }
    chartsLoading.value = true

    try {
      const response = await adminAPI.dashboard.getSnapshotV2({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value,
        include_stats: includeStats,
        include_trend: true,
        include_model_stats: true,
        include_group_stats: false,
        include_users_trend: false
      })
      if (currentSeq !== chartLoadSeq) {
        return
      }

      if (includeStats && response.stats) {
        stats.value = response.stats
      }
      trendData.value = response.trend || []
      modelStats.value = response.models || []
    } catch (error) {
      if (currentSeq !== chartLoadSeq) {
        return
      }
      options.showError(options.t('admin.dashboard.failedToLoad'))
      console.error('Error loading dashboard snapshot:', error)
    } finally {
      if (currentSeq === chartLoadSeq) {
        loading.value = false
        chartsLoading.value = false
      }
    }
  }

  const loadUsersTrend = async () => {
    const currentSeq = ++usersTrendLoadSeq
    userTrendLoading.value = true

    try {
      const response = await adminAPI.dashboard.getUserUsageTrend({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value,
        limit: 12
      })
      if (currentSeq !== usersTrendLoadSeq) {
        return
      }
      userTrend.value = response.trend || []
    } catch (error) {
      if (currentSeq !== usersTrendLoadSeq) {
        return
      }
      console.error('Error loading users trend:', error)
      userTrend.value = []
    } finally {
      if (currentSeq === usersTrendLoadSeq) {
        userTrendLoading.value = false
      }
    }
  }

  const loadUserSpendingRanking = async () => {
    const currentSeq = ++rankingLoadSeq
    rankingLoading.value = true
    rankingError.value = false

    try {
      const response = await adminAPI.dashboard.getUserSpendingRanking({
        start_date: startDate.value,
        end_date: endDate.value,
        limit: DASHBOARD_RANKING_LIMIT
      })
      if (currentSeq !== rankingLoadSeq) {
        return
      }
      rankingItems.value = response.ranking || []
      rankingTotalActualCost.value = response.total_actual_cost || 0
      rankingTotalRequests.value = response.total_requests || 0
      rankingTotalTokens.value = response.total_tokens || 0
    } catch (error) {
      if (currentSeq !== rankingLoadSeq) {
        return
      }
      console.error('Error loading user spending ranking:', error)
      rankingItems.value = []
      rankingTotalActualCost.value = 0
      rankingTotalRequests.value = 0
      rankingTotalTokens.value = 0
      rankingError.value = true
    } finally {
      if (currentSeq === rankingLoadSeq) {
        rankingLoading.value = false
      }
    }
  }

  const loadDashboardStats = async () => {
    await Promise.all([
      loadDashboardSnapshot(true),
      loadUsersTrend(),
      loadUserSpendingRanking()
    ])
  }

  const loadChartData = async () => {
    await Promise.all([
      loadDashboardSnapshot(false),
      loadUsersTrend(),
      loadUserSpendingRanking()
    ])
  }

  const onDateRangeChange = (range: {
    startDate: string
    endDate: string
    preset: string | null
  }) => {
    void range.preset
    startDate.value = range.startDate
    endDate.value = range.endDate
    granularity.value = getDashboardGranularityForRange(range.startDate, range.endDate)
    void loadChartData()
  }

  return {
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
  }
}
