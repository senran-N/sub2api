<template>
  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
    <div class="card usage-stats-cards__card flex items-center">
      <div class="usage-stats-cards__icon-shell usage-stats-cards__icon-shell--info">
        <Icon name="document" size="md" />
      </div>
      <div>
        <p class="usage-stats-cards__label text-xs font-medium">{{ t('usage.totalRequests') }}</p>
        <p class="usage-stats-cards__value text-xl font-bold">{{ stats?.total_requests?.toLocaleString() || '0' }}</p>
        <p class="usage-stats-cards__subtle text-xs">{{ t('usage.inSelectedRange') }}</p>
      </div>
    </div>
    <div class="card usage-stats-cards__card flex items-center">
      <div class="usage-stats-cards__icon-shell usage-stats-cards__icon-shell--warning"><svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /></svg></div>
      <div>
        <p class="usage-stats-cards__label text-xs font-medium">{{ t('usage.totalTokens') }}</p>
        <p class="usage-stats-cards__value text-xl font-bold">{{ formatTokens(stats?.total_tokens || 0) }}</p>
        <p class="usage-stats-cards__label text-xs">
          {{ t('usage.in') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} /
          {{ t('usage.out') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}
        </p>
      </div>
    </div>
    <div class="card usage-stats-cards__card flex items-center">
      <div class="usage-stats-cards__icon-shell usage-stats-cards__icon-shell--success">
        <Icon name="dollar" size="md" />
      </div>
      <div class="min-w-0 flex-1">
        <p class="usage-stats-cards__label text-xs font-medium">{{ t('usage.totalCost') }}</p>
        <p class="usage-stats-cards__value usage-stats-cards__value--success text-xl font-bold">
          ${{ ((stats?.total_account_cost ?? stats?.total_actual_cost) || 0).toFixed(4) }}
        </p>
        <p class="usage-stats-cards__subtle text-xs" v-if="stats?.total_account_cost != null">
          {{ t('usage.userBilled') }}:
          <span class="usage-stats-cards__muted-value">${{ (stats?.total_actual_cost || 0).toFixed(4) }}</span>
          · {{ t('usage.standardCost') }}:
          <span class="usage-stats-cards__muted-value">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
        </p>
        <p class="usage-stats-cards__subtle text-xs" v-else>
          {{ t('usage.standardCost') }}:
          <span class="usage-stats-cards__muted-value line-through">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
        </p>
      </div>
    </div>
    <div class="card usage-stats-cards__card flex items-center">
      <div class="usage-stats-cards__icon-shell usage-stats-cards__icon-shell--brand">
        <Icon name="clock" size="md" />
      </div>
      <div><p class="usage-stats-cards__label text-xs font-medium">{{ t('usage.avgDuration') }}</p><p class="usage-stats-cards__value text-xl font-bold">{{ formatDuration(stats?.average_duration_ms || 0) }}</p></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import Icon from '@/components/icons/Icon.vue'

defineProps<{ stats: AdminUsageStatsResponse | null }>()

const { t } = useI18n()

const formatDuration = (ms: number) =>
  ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`

const formatTokens = (value: number) => {
  if (value >= 1e9) return (value / 1e9).toFixed(2) + 'B'
  if (value >= 1e6) return (value / 1e6).toFixed(2) + 'M'
  if (value >= 1e3) return (value / 1e3).toFixed(2) + 'K'
  return value.toLocaleString()
}
</script>

<style scoped>
.usage-stats-cards__card {
  gap: var(--theme-usage-stats-card-gap);
  padding: var(--theme-usage-stats-card-padding);
}

.usage-stats-cards__icon-shell {
  @apply flex items-center justify-center;
  border-radius: var(--theme-usage-stats-icon-radius);
  padding: var(--theme-usage-stats-icon-padding);
}

.usage-stats-cards__icon-shell--info {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.usage-stats-cards__icon-shell--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.usage-stats-cards__icon-shell--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.usage-stats-cards__icon-shell--brand {
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.usage-stats-cards__label,
.usage-stats-cards__subtle {
  color: var(--theme-page-muted);
}

.usage-stats-cards__value {
  color: var(--theme-page-text);
}

.usage-stats-cards__value--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.usage-stats-cards__muted-value {
  color: color-mix(in srgb, var(--theme-page-muted) 62%, var(--theme-surface));
}
</style>
