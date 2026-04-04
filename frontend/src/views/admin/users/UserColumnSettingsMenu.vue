<template>
  <div
    class="absolute right-0 top-full z-50 mt-1 max-h-80 w-48 overflow-y-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
  >
    <button
      v-for="column in toggleableColumns"
      :key="column.key"
      class="flex w-full items-center justify-between px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
      @click="emit('toggle-column', column.key)"
    >
      <span>{{ column.label }}</span>
      <Icon
        v-if="isColumnVisible(column.key)"
        name="check"
        size="sm"
        class="text-primary-500"
        :stroke-width="2"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'

defineProps<{
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'toggle-column': [key: string]
}>()
</script>
