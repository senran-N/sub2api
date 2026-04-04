<template>
  <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/format'

defineProps<{
  value: string
}>()
</script>
