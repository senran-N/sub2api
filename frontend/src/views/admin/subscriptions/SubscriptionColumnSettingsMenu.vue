<template>
  <div
    class="absolute right-0 z-50 mt-2 w-48 origin-top-right rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
  >
    <div class="p-2">
      <div class="mb-2 border-b border-gray-200 pb-2 dark:border-gray-700">
        <div class="px-3 py-1 text-xs font-medium text-gray-500 dark:text-gray-400">
          {{ t('admin.subscriptions.columns.user') }}
        </div>
        <button
          class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
          @click="emit('set-user-mode', 'email')"
        >
          <span>{{ t('admin.users.columns.email') }}</span>
          <Icon v-if="userColumnMode === 'email'" name="check" size="sm" class="text-primary-500" />
        </button>
        <button
          class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
          @click="emit('set-user-mode', 'username')"
        >
          <span>{{ t('admin.users.columns.username') }}</span>
          <Icon v-if="userColumnMode === 'username'" name="check" size="sm" class="text-primary-500" />
        </button>
      </div>
      <button
        v-for="column in toggleableColumns"
        :key="column.key"
        class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
        @click="emit('toggle-column', column.key)"
      >
        <span>{{ column.label }}</span>
        <Icon v-if="isColumnVisible(column.key)" name="check" size="sm" class="text-primary-500" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  userColumnMode: 'email' | 'username'
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'set-user-mode': [mode: 'email' | 'username']
  'toggle-column': [key: string]
}>()

const { t } = useI18n()
</script>
