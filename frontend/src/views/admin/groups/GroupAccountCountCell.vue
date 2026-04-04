<template>
  <div class="space-y-0.5 text-xs">
    <div>
      <span class="text-gray-500 dark:text-gray-400">{{ t('admin.groups.accountsAvailable') }}</span>
      <span class="ml-1 font-medium text-emerald-600 dark:text-emerald-400">{{ availableCount }}</span>
      <span class="ml-1 inline-flex items-center rounded bg-gray-100 px-1.5 py-0.5 font-medium text-gray-800 dark:bg-dark-600 dark:text-gray-300">{{ t('admin.groups.accountsUnit') }}</span>
    </div>
    <div v-if="group.rate_limited_account_count">
      <span class="text-gray-500 dark:text-gray-400">{{ t('admin.groups.accountsRateLimited') }}</span>
      <span class="ml-1 font-medium text-amber-600 dark:text-amber-400">{{ group.rate_limited_account_count }}</span>
      <span class="ml-1 inline-flex items-center rounded bg-gray-100 px-1.5 py-0.5 font-medium text-gray-800 dark:bg-dark-600 dark:text-gray-300">{{ t('admin.groups.accountsUnit') }}</span>
    </div>
    <div>
      <span class="text-gray-500 dark:text-gray-400">{{ t('admin.groups.accountsTotal') }}</span>
      <span class="ml-1 font-medium text-gray-700 dark:text-gray-300">{{ group.account_count || 0 }}</span>
      <span class="ml-1 inline-flex items-center rounded bg-gray-100 px-1.5 py-0.5 font-medium text-gray-800 dark:bg-dark-600 dark:text-gray-300">{{ t('admin.groups.accountsUnit') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminGroup } from '@/types'

const props = defineProps<{
  group: AdminGroup
}>()

const { t } = useI18n()

const availableCount = computed(() => {
  return (props.group.active_account_count || 0) - (props.group.rate_limited_account_count || 0)
})
</script>
