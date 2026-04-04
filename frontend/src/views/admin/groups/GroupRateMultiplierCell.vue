<template>
  <span class="text-sm text-gray-700 dark:text-gray-300">{{ rateMultiplier }}x</span>
</template>

<script setup lang="ts">
defineProps<{
  rateMultiplier: number
}>()
</script>
