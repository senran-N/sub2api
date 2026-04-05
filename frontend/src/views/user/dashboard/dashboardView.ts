import { computed, ref } from 'vue'
import { usageAPI, type TrendParams, type UserDashboardStats } from '@/api/usage'
import { useAuthStore } from '@/stores/auth'
import type { ModelStat, TrendDataPoint, UsageLog } from '@/types'

type DashboardGranularity = NonNullable<TrendParams['granularity']>

export function formatDashboardDateValue(date: Date): string {
  return date.toISOString().split('T')[0]
}

export function createDashboardDateRange(now: Date = new Date()) {
  return {
    startDate: formatDashboardDateValue(new Date(now.getTime() - 6 * 86400000)),
    endDate: formatDashboardDateValue(now)
  }
}

export function useDashboardViewModel() {
  const authStore = useAuthStore()
  const user = computed(() => authStore.user)

  const stats = ref<UserDashboardStats | null>(null)
  const loading = ref(false)
  const loadingUsage = ref(false)
  const loadingCharts = ref(false)
  const trendData = ref<TrendDataPoint[]>([])
  const modelStats = ref<ModelStat[]>([])
  const recentUsage = ref<UsageLog[]>([])

  const initialRange = createDashboardDateRange()
  const startDate = ref(initialRange.startDate)
  const endDate = ref(initialRange.endDate)
  const granularity = ref<DashboardGranularity>('day')

  const chartParams = computed<TrendParams>(() => ({
    start_date: startDate.value,
    end_date: endDate.value,
    granularity: granularity.value
  }))

  async function loadStats() {
    loading.value = true
    try {
      await authStore.refreshUser()
      stats.value = await usageAPI.getDashboardStats()
    } catch (error) {
      console.error('Failed to load dashboard stats:', error)
    } finally {
      loading.value = false
    }
  }

  async function loadCharts() {
    loadingCharts.value = true
    try {
      const [trendResponse, modelsResponse] = await Promise.all([
        usageAPI.getDashboardTrend(chartParams.value),
        usageAPI.getDashboardModels(chartParams.value)
      ])

      trendData.value = trendResponse.trend || []
      modelStats.value = modelsResponse.models || []
    } catch (error) {
      console.error('Failed to load charts:', error)
    } finally {
      loadingCharts.value = false
    }
  }

  async function loadRecent() {
    loadingUsage.value = true
    try {
      const response = await usageAPI.getByDateRange(startDate.value, endDate.value)
      recentUsage.value = response.items.slice(0, 5)
    } catch (error) {
      console.error('Failed to load recent usage:', error)
    } finally {
      loadingUsage.value = false
    }
  }

  async function refreshAll() {
    await Promise.allSettled([loadStats(), loadCharts(), loadRecent()])
  }

  return {
    authStore,
    user,
    stats,
    loading,
    loadingUsage,
    loadingCharts,
    trendData,
    modelStats,
    recentUsage,
    startDate,
    endDate,
    granularity,
    loadCharts,
    refreshAll
  }
}
