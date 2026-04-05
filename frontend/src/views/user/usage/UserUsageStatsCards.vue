<template>
  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
          <Icon name="document" size="md" class="text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('usage.totalRequests') }}
          </p>
          <p class="text-xl font-bold text-gray-900 dark:text-white">
            {{ stats?.total_requests?.toLocaleString() || '0' }}
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('usage.inSelectedRange') }}
          </p>
        </div>
      </div>
    </div>

    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
          <Icon name="cube" size="md" class="text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('usage.totalTokens') }}
          </p>
          <p class="text-xl font-bold text-gray-900 dark:text-white">
            {{ formatUserUsageTokens(stats?.total_tokens || 0) }}
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('usage.in') }}: {{ formatUserUsageTokens(stats?.total_input_tokens || 0) }} /
            {{ t('usage.out') }}: {{ formatUserUsageTokens(stats?.total_output_tokens || 0) }}
          </p>
        </div>
      </div>
    </div>

    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
          <Icon name="dollar" size="md" class="text-green-600 dark:text-green-400" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('usage.totalCost') }}
          </p>
          <p class="text-xl font-bold text-green-600 dark:text-green-400">
            ${{ (stats?.total_actual_cost || 0).toFixed(4) }}
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('usage.actualCost') }} /
            <span class="line-through">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
            {{ t('usage.standardCost') }}
          </p>
        </div>
      </div>
    </div>

    <div class="card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
          <Icon name="clock" size="md" class="text-purple-600 dark:text-purple-400" />
        </div>
        <div>
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('usage.avgDuration') }}
          </p>
          <p class="text-xl font-bold text-gray-900 dark:text-white">
            {{ formatUserUsageDuration(stats?.average_duration_ms || 0) }}
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.perRequest') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UsageStatsResponse } from '@/types'
import {
  formatUserUsageDuration,
  formatUserUsageTokens
} from '../userUsageView'

defineProps<{
  stats: UsageStatsResponse | null
}>()

const { t } = useI18n()
</script>
