<template>
  <div v-if="status" class="mb-2 flex items-center justify-center fade-up">
    <div class="key-usage-status-badge__pill inline-flex items-center gap-2">
      <span
        class="key-usage-status-badge__dot pulse-dot h-2.5 w-2.5 rounded-full"
        :class="
          status.isActive
            ? 'key-usage-status-badge__dot--active'
            : 'key-usage-status-badge__dot--inactive'
        "
      ></span>
      <span class="key-usage-status-badge__label text-sm font-medium">{{ status.label }}</span>
      <span class="key-usage-status-badge__separator text-xs">|</span>
      <span class="key-usage-status-badge__meta text-xs">{{ status.statusText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageStatusInfo } from './keyUsageView'

defineProps<{
  status: KeyUsageStatusInfo | null
}>()
</script>

<style scoped>
.key-usage-status-badge__pill {
  padding:
    var(--theme-key-usage-query-range-chip-padding-y)
    calc(var(--theme-key-usage-query-range-chip-padding-x) + 0.5rem);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
  box-shadow: var(--theme-card-shadow);
  backdrop-filter: blur(14px);
}

.key-usage-status-badge__label {
  color: var(--theme-page-text);
}

.key-usage-status-badge__separator {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.key-usage-status-badge__meta {
  color: var(--theme-page-muted);
}

.key-usage-status-badge__dot--active {
  background: rgb(var(--theme-success-rgb));
}

.key-usage-status-badge__dot--inactive {
  background: rgb(var(--theme-danger-rgb));
}
</style>
