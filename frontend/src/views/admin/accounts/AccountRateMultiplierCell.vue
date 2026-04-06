<template>
  <span class="theme-text-default text-sm font-mono">
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
