<template>
  <span class="theme-text-default text-sm">{{ value }}</span>
</template>

<script setup lang="ts">
defineProps<{
  value: number
}>()
</script>
