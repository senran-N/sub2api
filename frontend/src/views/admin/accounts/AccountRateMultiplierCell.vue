<template>
  <span class="text-sm font-mono text-gray-700 dark:text-gray-300">
    {{ formattedRateMultiplier }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  rateMultiplier?: number | null
}>()

const formattedRateMultiplier = computed(
  () => `${(props.rateMultiplier ?? 1).toFixed(2)}x`
)
</script>
