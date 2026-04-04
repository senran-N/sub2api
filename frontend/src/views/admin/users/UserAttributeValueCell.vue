<template>
  <div class="max-w-xs">
    <span
      class="block truncate text-sm text-gray-700 dark:text-gray-300"
      :title="value"
    >
      {{ value }}
    </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  value: string
}>()
</script>
