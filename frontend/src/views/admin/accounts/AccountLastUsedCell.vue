<template>
  <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatRelativeTime(value) }}</span>
</template>

<script setup lang="ts">
import { formatRelativeTime } from '@/utils/format'

defineProps<{
  value: string | null
}>()
</script>
