<template>
  <span
    v-if="value"
    class="block max-w-[320px] whitespace-normal break-all text-sm text-gray-600 dark:text-gray-400"
    :title="value"
  >
    {{ value }}
  </span>
  <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
</template>

<script setup lang="ts">
defineProps<{
  value?: string | null
}>()
</script>
