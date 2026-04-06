<template>
  <span class="user-usage-endpoint-cell theme-text-default block whitespace-normal break-all text-sm">
    {{ formatUserUsageEndpoints(endpoint) }}
  </span>
</template>

<script setup lang="ts">
import { formatUserUsageEndpoints } from '../userUsageView'

defineProps<{
  endpoint?: string | null
}>()
</script>

<style scoped>
.user-usage-endpoint-cell {
  max-width: var(--theme-usage-table-endpoint-max-width);
}
</style>
