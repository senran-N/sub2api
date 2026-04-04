<template>
  <span class="text-sm text-gray-700 dark:text-gray-300">{{ value }}</span>
</template>

<script setup lang="ts">
defineProps<{
  value: number
}>()
</script>
