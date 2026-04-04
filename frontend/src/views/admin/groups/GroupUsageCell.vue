<template>
  <div v-if="loading" class="text-xs text-gray-400">—</div>
  <div v-else class="space-y-0.5 text-xs">
    <div class="text-gray-500 dark:text-gray-400">
      <span class="text-gray-400 dark:text-gray-500">{{ t('admin.groups.usageToday') }}</span>
      <span class="ml-1 font-medium text-gray-700 dark:text-gray-300">${{ formatGroupCost(summary?.today_cost ?? 0) }}</span>
    </div>
    <div class="text-gray-500 dark:text-gray-400">
      <span class="text-gray-400 dark:text-gray-500">{{ t('admin.groups.usageTotal') }}</span>
      <span class="ml-1 font-medium text-gray-700 dark:text-gray-300">${{ formatGroupCost(summary?.total_cost ?? 0) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatGroupCost } from '../groupsTable'

defineProps<{
  loading: boolean
  summary?: {
    today_cost: number
    total_cost: number
  }
}>()

const { t } = useI18n()
</script>
