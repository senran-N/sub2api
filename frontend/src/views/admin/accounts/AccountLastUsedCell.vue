<template>
  <span class="theme-text-muted text-sm">{{ formatRelativeTime(value) }}</span>
</template>

<script setup lang="ts">
import { formatRelativeTime } from '@/utils/format'

defineProps<{
  value: string | null
}>()
</script>
