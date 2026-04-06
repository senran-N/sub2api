<template>
  <span class="theme-text-strong font-medium">{{ value }}</span>
</template>

<script setup lang="ts">
defineProps<{
  value: string
}>()
</script>
