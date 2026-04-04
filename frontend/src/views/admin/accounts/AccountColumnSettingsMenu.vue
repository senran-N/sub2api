<template>
  <div class="max-h-80 overflow-y-auto p-2">
    <button
      v-for="column in toggleableColumns"
      :key="column.key"
      class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
      @click="emit('toggle-column', column.key)"
    >
      <span>{{ column.label }}</span>
      <Icon
        v-if="isColumnVisible(column.key)"
        name="check"
        size="sm"
        class="text-primary-500"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'toggle-column': [key: string]
}>()
</script>
