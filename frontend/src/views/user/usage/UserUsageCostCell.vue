<template>
  <div class="flex items-center gap-1.5 text-sm">
    <span class="font-medium text-green-600 dark:text-green-400">
      ${{ row.actual_cost.toFixed(6) }}
    </span>
    <button
      type="button"
      class="group relative"
      :aria-label="t('usage.costDetails')"
      @mouseenter="emit('show-details', $event, row)"
      @mouseleave="emit('hide-details')"
    >
      <div
        class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-blue-100 dark:bg-gray-700 dark:group-hover:bg-blue-900/50"
      >
        <Icon
          name="infoCircle"
          size="xs"
          class="text-gray-400 group-hover:text-blue-500 dark:text-gray-500 dark:group-hover:text-blue-400"
        />
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UsageLog } from '@/types'

defineProps<{
  row: UsageLog
}>()

const emit = defineEmits<{
  'show-details': [event: MouseEvent, row: UsageLog]
  'hide-details': []
}>()

const { t } = useI18n()
</script>
