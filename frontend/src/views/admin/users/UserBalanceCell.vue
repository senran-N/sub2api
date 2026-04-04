<template>
  <div class="flex items-center gap-2">
    <div class="group relative">
      <button
        class="font-medium text-gray-900 underline decoration-dashed decoration-gray-300 underline-offset-4 transition-colors hover:text-primary-600 dark:text-white dark:decoration-dark-500 dark:hover:text-primary-400"
        @click="emit('history', user)"
      >
        ${{ user.balance.toFixed(2) }}
      </button>
      <div class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-1.5 -translate-x-1/2 whitespace-nowrap rounded bg-gray-900 px-2 py-1 text-xs text-white opacity-0 shadow-lg transition-opacity duration-75 group-hover:opacity-100 dark:bg-dark-600">
        {{ t('admin.users.balanceHistoryTip') }}
        <div class="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-gray-900 dark:border-t-dark-600"></div>
      </div>
    </div>
    <button
      class="rounded px-2 py-0.5 text-xs font-medium text-emerald-600 transition-colors hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-900/20"
      :title="t('admin.users.deposit')"
      @click.stop="emit('deposit', user)"
    >
      {{ t('admin.users.deposit') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminUser } from '@/types'

defineProps<{
  user: AdminUser
}>()

const emit = defineEmits<{
  deposit: [user: AdminUser]
  history: [user: AdminUser]
}>()

const { t } = useI18n()
</script>
