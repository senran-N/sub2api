<template>
  <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDateTime(value) }}</span>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/format'

defineProps<{
  value: string
}>()
</script>
