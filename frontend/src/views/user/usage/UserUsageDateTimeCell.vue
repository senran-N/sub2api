<template>
  <span class="theme-text-muted text-sm">{{ formatDateTime(value) }}</span>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/format'

defineProps<{
  value: string
}>()
</script>
