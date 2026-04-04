<template>
  <div class="flex items-center gap-1">
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
      @click="$emit('edit', group)"
    >
      <Icon name="edit" size="sm" />
      <span class="text-xs">{{ t('common.edit') }}</span>
    </button>
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-purple-600 dark:hover:bg-dark-700 dark:hover:text-purple-400"
      @click="$emit('rate-multipliers', group)"
    >
      <Icon name="dollar" size="sm" />
      <span class="text-xs">{{ t('admin.groups.rateMultipliers') }}</span>
    </button>
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
      @click="$emit('delete', group)"
    >
      <Icon name="trash" size="sm" />
      <span class="text-xs">{{ t('common.delete') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminGroup } from '@/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  group: AdminGroup
}>()

defineEmits<{
  edit: [group: AdminGroup]
  'rate-multipliers': [group: AdminGroup]
  delete: [group: AdminGroup]
}>()

const { t } = useI18n()
</script>
