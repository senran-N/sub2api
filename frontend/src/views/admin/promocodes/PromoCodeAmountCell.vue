<template>
  <span class="theme-text-strong text-sm font-medium">
    ${{ amount.toFixed(2) }}
  </span>
</template>

<script setup lang="ts">
defineProps<{
  amount: number
}>()
</script>
