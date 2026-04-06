<template>
  <div class="min-w-0">
    <div class="flex items-center gap-2">
      <span class="theme-text-strong truncate font-medium">{{ title }}</span>
    </div>
    <div class="theme-text-muted mt-1 flex items-center gap-2 text-xs">
      <span>#{{ id }}</span>
      <span class="theme-text-subtle">·</span>
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
