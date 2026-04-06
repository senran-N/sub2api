<template>
  <span class="theme-text-default text-sm">{{ rateMultiplier }}x</span>
</template>

<script setup lang="ts">
defineProps<{
  rateMultiplier: number
}>()
</script>
