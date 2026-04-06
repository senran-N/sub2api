<template>
  <span class="theme-text-muted text-sm">
    {{ value ? formatDateTime(value) : fallbackText ?? '-' }}
  </span>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/format'

defineProps<{
  value: string | null
  fallbackText?: string
}>()
</script>
