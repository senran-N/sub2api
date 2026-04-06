<template>
  <span
    v-if="value"
    class="user-usage-user-agent-cell theme-text-muted block whitespace-normal break-all text-sm"
    :title="value"
  >
    {{ value }}
  </span>
  <span v-else class="theme-text-subtle text-sm">-</span>
</template>

<script setup lang="ts">
defineProps<{
  value?: string | null
}>()
</script>

<style scoped>
.user-usage-user-agent-cell {
  max-width: var(--theme-usage-table-user-agent-max-width);
}
</style>
