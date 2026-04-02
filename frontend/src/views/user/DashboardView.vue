<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Skeleton loading state -->
      <template v-if="loading">
        <!-- Stats skeleton -->
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="i in 4" :key="i" class="card p-5">
            <div class="flex items-start gap-4">
              <div class="h-12 w-12 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700"></div>
              <div class="flex-1 space-y-2">
                <div class="h-4 w-20 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
                <div class="h-7 w-28 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
              </div>
            </div>
          </div>
        </div>
        <!-- Charts skeleton -->
        <div class="card p-6">
          <div class="mb-4 flex items-center justify-between">
            <div class="h-5 w-32 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="flex gap-2">
              <div class="h-8 w-24 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
              <div class="h-8 w-24 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
            </div>
          </div>
          <div class="h-64 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700"></div>
        </div>
        <!-- Bottom grid skeleton -->
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2 card p-6">
            <div class="mb-4 h-5 w-40 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="space-y-3">
              <div v-for="i in 4" :key="i" class="flex items-center gap-4">
                <div class="h-4 w-full animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
              </div>
            </div>
          </div>
          <div class="lg:col-span-1 card p-6">
            <div class="mb-4 h-5 w-32 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="space-y-3">
              <div v-for="i in 3" :key="i" class="h-10 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
            </div>
          </div>
        </div>
      </template>
      <template v-else-if="stats">
        <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" />
        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" @refresh="refreshAll" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'; import { useAuthStore } from '@/stores/auth'; import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'

const authStore = useAuthStore(); const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])

const formatLD = (d: Date) => d.toISOString().split('T')[0]
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatLD(new Date())); const granularity = ref('day')

const loadStats = async () => { loading.value = true; try { await authStore.refreshUser(); stats.value = await usageAPI.getDashboardStats() } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false } }
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }
const refreshAll = () => { loadStats(); loadCharts(); loadRecent() }

onMounted(() => { refreshAll() })
</script>
