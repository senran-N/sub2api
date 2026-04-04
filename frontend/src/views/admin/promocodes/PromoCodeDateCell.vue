<template>
  <span class="text-sm text-gray-500 dark:text-dark-400">
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
