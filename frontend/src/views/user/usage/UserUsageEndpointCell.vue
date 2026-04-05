<template>
  <span class="block max-w-[320px] whitespace-normal break-all text-sm text-gray-600 dark:text-gray-300">
    {{ formatUserUsageEndpoints(endpoint) }}
  </span>
</template>

<script setup lang="ts">
import { formatUserUsageEndpoints } from '../userUsageView'

defineProps<{
  endpoint?: string | null
}>()
</script>
