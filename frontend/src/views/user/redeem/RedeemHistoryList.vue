<template>
  <div class="card">
    <div class="redeem-history-list__header">
      <h2 class="redeem-history-list__heading">
        {{ title }}
      </h2>
    </div>
    <div class="redeem-history-list__content">
      <div v-if="loading" class="redeem-history-list__status-state">
        <svg class="theme-loading-spinner h-6 w-6 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      </div>

      <div v-else-if="history.length > 0" class="space-y-3">
        <div
          v-for="item in history"
          :key="item.id"
          class="redeem-history-list__item"
        >
          <div class="redeem-history-list__item-main">
            <div
              class="redeem-history-list__icon-shell"
              :class="getIconShellClasses(buildRedeemHistoryPresentation(item).tone)"
            >
              <Icon
                :name="buildRedeemHistoryPresentation(item).iconName"
                size="md"
                :class="getToneClasses(buildRedeemHistoryPresentation(item).tone)"
              />
            </div>
            <div>
              <p class="redeem-history-list__title text-sm font-medium">
                {{ resolveRedeemHistoryTitle(item, t) }}
              </p>
              <p class="redeem-history-list__meta text-xs">
                {{ formatDateTime(item.used_at) }}
              </p>
            </div>
          </div>
          <div class="redeem-history-list__item-side">
            <p
              class="redeem-history-list__value"
              :class="getToneClasses(buildRedeemHistoryPresentation(item).tone)"
            >
              {{ formatRedeemHistoryValue(item, t) }}
            </p>
            <p
              v-if="!isAdminAdjustmentRedeemType(item.type)"
              class="redeem-history-list__code font-mono text-xs"
            >
              {{ item.code.slice(0, 8) }}...
            </p>
            <p v-else class="redeem-history-list__code text-xs">
              {{ adminAdjustmentLabel }}
            </p>
            <p
              v-if="item.notes"
              class="redeem-history-list__note"
              :title="item.notes"
            >
              {{ item.notes }}
            </p>
          </div>
        </div>
      </div>

      <div v-else class="empty-state redeem-history-list__status-state">
        <div class="redeem-history-list__empty-icon">
          <Icon name="clock" size="xl" class="redeem-history-list__code" />
        </div>
        <p class="redeem-history-list__meta redeem-history-list__empty-text">
          {{ emptyLabel }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { RedeemHistoryItem } from '@/api'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import {
  buildRedeemHistoryPresentation,
  formatRedeemHistoryValue,
  isAdminAdjustmentRedeemType,
  resolveRedeemHistoryTitle
} from './redeemView'

defineProps<{
  adminAdjustmentLabel: string
  emptyLabel: string
  history: RedeemHistoryItem[]
  loading: boolean
  title: string
}>()

const { t } = useI18n()

function getToneClasses(tone: 'success' | 'danger' | 'brand' | 'info' | 'warning') {
  return [
    'redeem-history-list__tone',
    `redeem-history-list__tone--${tone}`
  ]
}

function getIconShellClasses(tone: 'success' | 'danger' | 'brand' | 'info' | 'warning') {
  return [
    'redeem-history-list__icon-shell-surface',
    `redeem-history-list__icon-shell-surface--${tone}`
  ]
}
</script>

<style scoped>
.redeem-history-list__header {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  padding: calc(var(--theme-table-mobile-card-padding) * 1.1) calc(var(--theme-table-mobile-card-padding) * 1.5);
}

.redeem-history-list__heading {
  color: var(--theme-page-text);
  font-size: 1.125rem;
  font-weight: 600;
}

.redeem-history-list__content {
  padding: calc(var(--theme-table-mobile-card-padding) * 1.5);
}

.redeem-history-list__status-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) 0;
}

.redeem-history-list__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: calc(var(--theme-table-mobile-card-padding) * 0.75);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  padding: var(--theme-table-mobile-card-padding);
  background: color-mix(in srgb, var(--theme-surface-soft) 84%, var(--theme-surface));
}

.redeem-history-list__item-main {
  display: flex;
  align-items: center;
  gap: var(--theme-table-mobile-card-padding);
}

.redeem-history-list__item-side {
  text-align: right;
}

.redeem-history-list__value {
  font-size: 0.875rem;
  font-weight: 600;
}

.redeem-history-list__icon-shell {
  display: flex;
  height: calc(var(--theme-header-avatar-size) + 8px);
  width: calc(var(--theme-header-avatar-size) + 8px);
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: calc(var(--theme-surface-radius) + 4px);
}

.redeem-history-list__title {
  color: var(--theme-page-text);
  font-size: 0.875rem;
  font-weight: 500;
}

.redeem-history-list__meta,
.redeem-history-list__note {
  color: var(--theme-page-muted);
}

.redeem-history-list__empty-text {
  font-size: 0.875rem;
}

.redeem-history-list__code {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.redeem-history-list__note {
  margin-top: 0.25rem;
  max-width: min(100%, 12.5rem);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.75rem;
  font-style: italic;
}

.redeem-history-list__empty-icon {
  margin-bottom: calc(var(--theme-table-mobile-card-padding) * 0.75);
  display: flex;
  height: var(--theme-auth-logo-size);
  width: var(--theme-auth-logo-size);
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-auth-logo-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.redeem-history-list__icon-shell-surface--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
}

.redeem-history-list__icon-shell-surface--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 12%, var(--theme-surface));
}

.redeem-history-list__icon-shell-surface--brand {
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 12%, var(--theme-surface));
}

.redeem-history-list__icon-shell-surface--info {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 12%, var(--theme-surface));
}

.redeem-history-list__icon-shell-surface--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 12%, var(--theme-surface));
}

.redeem-history-list__tone--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.redeem-history-list__tone--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.redeem-history-list__tone--brand {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.redeem-history-list__tone--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.redeem-history-list__tone--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}
</style>
