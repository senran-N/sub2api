<template>
  <div class="space-y-3 sm:space-y-4">
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
            <Icon name="key" size="md" class="text-blue-600 dark:text-blue-400" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.apiKeys') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ stats.total_api_keys }}
            </p>
            <p class="text-xs text-green-600 dark:text-green-400">
              {{ stats.active_api_keys }} {{ t('common.active') }}
            </p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
            <Icon name="server" size="md" class="text-purple-600 dark:text-purple-400" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.accounts') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ stats.total_accounts }}
            </p>
            <p class="text-xs">
              <span class="text-green-600 dark:text-green-400">
                {{ stats.normal_accounts }} {{ t('common.active') }}
              </span>
              <span v-if="stats.error_accounts > 0" class="ml-1 text-red-500">
                {{ stats.error_accounts }} {{ t('common.error') }}
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
            <Icon name="chart" size="md" class="text-green-600 dark:text-green-400" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.todayRequests') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ stats.today_requests }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}
            </p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-emerald-100 p-2 dark:bg-emerald-900/30">
            <Icon name="userPlus" size="md" class="text-emerald-600 dark:text-emerald-400" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.users') }}
            </p>
            <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">
              +{{ stats.today_new_users }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('common.total') }}: {{ formatNumber(stats.total_users) }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
            <Icon name="cube" size="md" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.todayTokens') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ formatTokens(stats.today_tokens) }}
            </p>
            <p class="text-xs">
              <span class="text-amber-600 dark:text-amber-400" :title="t('admin.dashboard.actual')">
                ${{ formatCost(stats.today_actual_cost) }}
              </span>
              <span class="text-gray-400 dark:text-gray-500" :title="t('admin.dashboard.standard')">
                / ${{ formatCost(stats.today_cost) }}
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-indigo-100 p-2 dark:bg-indigo-900/30">
            <Icon name="database" size="md" class="text-indigo-600 dark:text-indigo-400" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.totalTokens') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ formatTokens(stats.total_tokens) }}
            </p>
            <p class="text-xs">
              <span class="text-indigo-600 dark:text-indigo-400" :title="t('admin.dashboard.actual')">
                ${{ formatCost(stats.total_actual_cost) }}
              </span>
              <span class="text-gray-400 dark:text-gray-500" :title="t('admin.dashboard.standard')">
                / ${{ formatCost(stats.total_cost) }}
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-violet-100 p-2 dark:bg-violet-900/30">
            <Icon name="bolt" size="md" class="text-violet-600 dark:text-violet-400" :stroke-width="2" />
          </div>
          <div class="flex-1">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.performance') }}
            </p>
            <div class="flex items-baseline gap-2">
              <p class="text-xl font-bold text-gray-900 dark:text-white">
                {{ formatTokens(stats.rpm) }}
              </p>
              <span class="text-xs text-gray-500 dark:text-gray-400">RPM</span>
            </div>
            <div class="flex items-baseline gap-2">
              <p class="text-sm font-semibold text-violet-600 dark:text-violet-400">
                {{ formatTokens(stats.tpm) }}
              </p>
              <span class="text-xs text-gray-500 dark:text-gray-400">TPM</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-rose-100 p-2 dark:bg-rose-900/30">
            <Icon name="clock" size="md" class="text-rose-600 dark:text-rose-400" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.dashboard.avgResponse') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ formatDuration(stats.average_duration_ms) }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ stats.active_users }} {{ t('admin.dashboard.activeUsers') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { DashboardStats } from '@/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  stats: DashboardStats
  formatTokens: (value: number | undefined | null) => string
  formatNumber: (value: number) => string
  formatCost: (value: number) => string
  formatDuration: (value: number) => string
}>()

const { t } = useI18n()
</script>
