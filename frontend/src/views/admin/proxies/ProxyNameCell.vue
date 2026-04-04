<template>
  <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
</template>

<script setup lang="ts">
defineProps<{
  value: string
}>()
</script>
