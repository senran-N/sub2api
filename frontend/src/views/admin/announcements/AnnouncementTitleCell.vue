<template>
  <div class="min-w-0">
    <div class="flex items-center gap-2">
      <span class="truncate font-medium text-gray-900 dark:text-white">{{ title }}</span>
    </div>
    <div class="mt-1 flex items-center gap-2 text-xs text-gray-500 dark:text-dark-400">
      <span>#{{ id }}</span>
      <span class="text-gray-300 dark:text-dark-700">·</span>
      <span>{{ formatDateTime(createdAt) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/format'

defineProps<{
  id: number
  title: string
  createdAt: string
}>()
</script>
