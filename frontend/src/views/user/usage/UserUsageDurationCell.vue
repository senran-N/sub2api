<template>
  <span v-if="value != null" class="text-sm text-gray-600 dark:text-gray-400">
    {{ formatUserUsageDuration(value) }}
  </span>
  <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
</template>

<script setup lang="ts">
import { formatUserUsageDuration } from '../userUsageView'

defineProps<{
  value: number | null
}>()
</script>
