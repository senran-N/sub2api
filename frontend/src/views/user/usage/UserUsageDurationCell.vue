<template>
  <span v-if="value != null" class="theme-text-muted text-sm">
    {{ formatUserUsageDuration(value) }}
  </span>
  <span v-else class="theme-text-subtle text-sm">-</span>
</template>

<script setup lang="ts">
import { formatUserUsageDuration } from '../userUsageView'

defineProps<{
  value: number | null
}>()
</script>
