<template>
  <span class="text-sm font-medium text-gray-900 dark:text-white">
    ${{ amount.toFixed(2) }}
  </span>
</template>

<script setup lang="ts">
defineProps<{
  amount: number
}>()
</script>
