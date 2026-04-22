<template>
  <span
    class="order-status-badge"
    :class="statusClass"
  >
    {{ statusLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OrderStatus } from '@/types/payment'

const props = defineProps<{
  status: OrderStatus
}>()

const { t } = useI18n()

const statusMap: Record<OrderStatus, { key: string; class: string }> = {
  PENDING: { key: 'payment.status.pending', class: 'order-status-badge--warning' },
  PAID: { key: 'payment.status.paid', class: 'order-status-badge--info' },
  RECHARGING: { key: 'payment.status.recharging', class: 'order-status-badge--info' },
  COMPLETED: { key: 'payment.status.completed', class: 'order-status-badge--success' },
  EXPIRED: { key: 'payment.status.expired', class: 'order-status-badge--muted' },
  CANCELLED: { key: 'payment.status.cancelled', class: 'order-status-badge--muted' },
  FAILED: { key: 'payment.status.failed', class: 'order-status-badge--danger' },
  REFUND_REQUESTED: { key: 'payment.status.refund_requested', class: 'order-status-badge--brand-orange' },
  REFUNDING: { key: 'payment.status.refunding', class: 'order-status-badge--brand-orange' },
  REFUNDED: { key: 'payment.status.refunded', class: 'order-status-badge--brand-purple' },
  PARTIALLY_REFUNDED: { key: 'payment.status.partially_refunded', class: 'order-status-badge--brand-purple' },
  REFUND_FAILED: { key: 'payment.status.refund_failed', class: 'order-status-badge--danger' },
}

const statusLabel = computed(() => {
  const entry = statusMap[props.status]
  return entry ? t(entry.key) : props.status
})

const statusClass = computed(() => {
  const entry = statusMap[props.status]
  return entry?.class ?? 'order-status-badge--muted'
})
</script>

<style scoped>
.order-status-badge {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--theme-chip-border, color-mix(in srgb, var(--theme-card-border) 72%, transparent));
  border-radius: 999px;
  padding: 0.125rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.2;
  background: var(--theme-chip-bg, color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface)));
  color: var(--theme-chip-fg, var(--theme-page-muted));
}

.order-status-badge--success {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-success-rgb)) 18%, var(--theme-card-border));
}

.order-status-badge--info {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-info-rgb)) 18%, var(--theme-card-border));
}

.order-status-badge--warning {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 18%, var(--theme-card-border));
}

.order-status-badge--danger {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 18%, var(--theme-card-border));
}

.order-status-badge--brand-orange {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 18%, var(--theme-card-border));
}

.order-status-badge--brand-purple {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 18%, var(--theme-card-border));
}
</style>
