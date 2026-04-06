<template>
  <div
    v-if="cells.length > 0"
    class="key-usage-usage-stats-card fade-up fade-up-delay-3 overflow-hidden border backdrop-blur-sm"
  >
    <div class="key-usage-usage-stats-card__header border-b">
      <h3 class="key-usage-usage-stats-card__label text-sm font-semibold uppercase tracking-wider">
        {{ title }}
      </h3>
    </div>
    <div class="key-usage-usage-stats-card__grid grid grid-cols-2 gap-px md:grid-cols-4">
      <div
        v-for="(cell, index) in cells"
        :key="index"
        class="key-usage-usage-stats-card__cell"
      >
        <div class="key-usage-usage-stats-card__label mb-1 text-xs">{{ cell.label }}</div>
        <div class="key-usage-usage-stats-card__value text-sm font-semibold tabular-nums">
          {{ cell.value }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageStatCell } from './keyUsageView'

defineProps<{
  cells: KeyUsageStatCell[]
  title: string
}>()
</script>

<style scoped>
.key-usage-usage-stats-card {
  border-radius: var(--theme-key-usage-card-radius);
  border-color: color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
}

.key-usage-usage-stats-card__header {
  padding: var(--theme-key-usage-card-header-padding-y) var(--theme-key-usage-card-header-padding-x);
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.key-usage-usage-stats-card__grid {
  background: color-mix(in srgb, var(--theme-card-border) 48%, transparent);
}

.key-usage-usage-stats-card__cell {
  padding: var(--theme-key-usage-card-cell-padding-y) var(--theme-key-usage-card-cell-padding-x);
  background: var(--theme-surface);
}

.key-usage-usage-stats-card__label {
  color: var(--theme-page-muted);
}

.key-usage-usage-stats-card__value {
  color: var(--theme-page-text);
}
</style>
