<template>
  <span
    v-if="notes"
    class="theme-text-muted block max-w-xs truncate text-sm"
    :title="notes"
  >
    {{ notes }}
  </span>
  <span v-else class="theme-text-subtle text-sm">-</span>
</template>

<script setup lang="ts">
defineProps<{
  notes?: string | null
}>()
</script>
