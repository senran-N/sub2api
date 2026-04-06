<template>
  <div
    v-if="rows.length > 0"
    class="key-usage-detail-card fade-up fade-up-delay-3 overflow-hidden"
  >
    <div class="key-usage-detail-card__header">
      <h3 class="key-usage-detail-card__title text-sm font-semibold uppercase tracking-wider">
        {{ title }}
      </h3>
    </div>
    <div class="key-usage-detail-card__list">
      <div
        v-for="(row, index) in rows"
        :key="index"
        class="key-usage-detail-card__row flex items-center justify-between"
      >
        <div class="flex items-center gap-3">
          <div class="key-usage-detail-card__icon-shell flex h-8 w-8 items-center justify-center" :class="row.iconBg">
            <svg
              class="key-usage-detail-card__icon h-4 w-4"
              :class="row.iconColor"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              v-html="row.iconSvg"
            ></svg>
          </div>
          <span class="key-usage-detail-card__label text-sm">{{ row.label }}</span>
        </div>
        <span
          class="text-sm font-semibold tabular-nums"
          :class="row.valueClass || 'theme-text-default'"
        >
          {{ row.value }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageDetailRow } from './keyUsageView'

defineProps<{
  rows: KeyUsageDetailRow[]
  title: string
}>()
</script>

<style scoped>
.key-usage-detail-card {
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  border-radius: var(--theme-key-usage-card-radius);
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
  backdrop-filter: blur(14px);
}

.key-usage-detail-card__header {
  padding: var(--theme-key-usage-card-header-padding-y) var(--theme-key-usage-card-header-padding-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
}

.key-usage-detail-card__title {
  color: var(--theme-page-muted);
}

.key-usage-detail-card__list {
  background: var(--theme-surface);
}

.key-usage-detail-card__row + .key-usage-detail-card__row {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.key-usage-detail-card__row {
  padding: var(--theme-key-usage-card-row-padding-y) var(--theme-key-usage-card-row-padding-x);
}

.key-usage-detail-card__label {
  color: color-mix(in srgb, var(--theme-page-text) 80%, transparent);
}

.key-usage-detail-card__icon-shell {
  border-radius: var(--theme-key-usage-card-icon-radius);
}

.key-usage-detail-card__icon-shell--accent {
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
}

.key-usage-detail-card__icon-shell--info {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.key-usage-detail-card__icon-shell--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.key-usage-detail-card__icon-shell--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
}

.key-usage-detail-card__icon--accent {
  color: var(--theme-accent);
}

.key-usage-detail-card__icon--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.key-usage-detail-card__icon--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.key-usage-detail-card__icon--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}
</style>
