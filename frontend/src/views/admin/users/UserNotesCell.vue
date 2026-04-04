<template>
  <div class="max-w-xs">
    <span
      v-if="notes"
      :title="notes.length > 30 ? notes : undefined"
      class="block truncate text-sm text-gray-600 dark:text-gray-400"
    >
      {{ displayNotes }}
    </span>
    <span v-else class="text-sm text-gray-400">-</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  notes?: string | null
}>()

const displayNotes = computed(() => {
  if (!props.notes) {
    return '-'
  }

  return props.notes.length > 30 ? `${props.notes.substring(0, 25)}...` : props.notes
})
</script>
