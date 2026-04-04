<template>
  <span
    v-if="notes"
    class="block max-w-xs truncate text-sm text-gray-600 dark:text-gray-300"
    :title="notes"
  >
    {{ notes }}
  </span>
  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
defineProps<{
  notes?: string | null
}>()
</script>
