<template>
  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
    <div class="card user-usage-stats-cards__card">
      <div class="user-usage-stats-cards__card-content">
        <div class="user-usage-stats-cards__icon-shell user-usage-stats-cards__icon-shell--info">
          <Icon name="document" size="md" class="user-usage-stats-cards__icon user-usage-stats-cards__icon--info" />
        </div>
        <div>
          <p class="user-usage-stats-cards__label">
            {{ t('usage.totalRequests') }}
          </p>
          <p class="user-usage-stats-cards__value">
            {{ stats?.total_requests?.toLocaleString() || '0' }}
          </p>
          <p class="user-usage-stats-cards__meta">
            {{ t('usage.inSelectedRange') }}
          </p>
        </div>
      </div>
    </div>

    <div class="card user-usage-stats-cards__card">
      <div class="user-usage-stats-cards__card-content">
        <div class="user-usage-stats-cards__icon-shell user-usage-stats-cards__icon-shell--warning">
          <Icon name="cube" size="md" class="user-usage-stats-cards__icon user-usage-stats-cards__icon--warning" />
        </div>
        <div>
          <p class="user-usage-stats-cards__label">
            {{ t('usage.totalTokens') }}
          </p>
          <p class="user-usage-stats-cards__value">
            {{ formatUserUsageTokens(stats?.total_tokens || 0) }}
          </p>
          <p class="user-usage-stats-cards__meta">
            {{ t('usage.in') }}: {{ formatUserUsageTokens(stats?.total_input_tokens || 0) }} /
            {{ t('usage.out') }}: {{ formatUserUsageTokens(stats?.total_output_tokens || 0) }}
          </p>
        </div>
      </div>
    </div>

    <div class="card user-usage-stats-cards__card">
      <div class="user-usage-stats-cards__card-content">
        <div class="user-usage-stats-cards__icon-shell user-usage-stats-cards__icon-shell--success">
          <Icon name="dollar" size="md" class="user-usage-stats-cards__icon user-usage-stats-cards__icon--success" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="user-usage-stats-cards__label">
            {{ t('usage.totalCost') }}
          </p>
          <p class="user-usage-stats-cards__value user-usage-stats-cards__value--success">
            ${{ (stats?.total_actual_cost || 0).toFixed(4) }}
          </p>
          <p class="user-usage-stats-cards__meta">
            {{ t('usage.actualCost') }} /
            <span class="user-usage-stats-cards__meta-strike">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
            {{ t('usage.standardCost') }}
          </p>
        </div>
      </div>
    </div>

    <div class="card user-usage-stats-cards__card">
      <div class="user-usage-stats-cards__card-content">
        <div class="user-usage-stats-cards__icon-shell user-usage-stats-cards__icon-shell--purple">
          <Icon name="clock" size="md" class="user-usage-stats-cards__icon user-usage-stats-cards__icon--purple" />
        </div>
        <div>
          <p class="user-usage-stats-cards__label">
            {{ t('usage.avgDuration') }}
          </p>
          <p class="user-usage-stats-cards__value">
            {{ formatUserUsageDuration(stats?.average_duration_ms || 0) }}
          </p>
          <p class="user-usage-stats-cards__meta">{{ t('usage.perRequest') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UsageStatsResponse } from '@/types'
import {
  formatUserUsageDuration,
  formatUserUsageTokens
} from '../userUsageView'

defineProps<{
  stats: UsageStatsResponse | null
}>()

const { t } = useI18n()
</script>

<style scoped>
.user-usage-stats-cards__card {
  padding: var(--theme-usage-stats-card-padding);
  background: color-mix(in srgb, var(--theme-surface) 92%, transparent);
}

.user-usage-stats-cards__card-content {
  @apply flex items-center;
  gap: var(--theme-usage-stats-card-gap);
}

.user-usage-stats-cards__icon-shell {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: var(--theme-usage-stats-icon-padding);
}

.user-usage-stats-cards__icon-shell--info {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 12%, var(--theme-surface));
}

.user-usage-stats-cards__icon-shell--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 12%, var(--theme-surface));
}

.user-usage-stats-cards__icon-shell--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
}

.user-usage-stats-cards__icon-shell--purple {
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 12%, var(--theme-surface));
}

.user-usage-stats-cards__icon--info {
  color: rgb(var(--theme-info-rgb));
}

.user-usage-stats-cards__icon--warning {
  color: rgb(var(--theme-warning-rgb));
}

.user-usage-stats-cards__icon--success {
  color: rgb(var(--theme-success-rgb));
}

.user-usage-stats-cards__icon--purple {
  color: rgb(var(--theme-brand-purple-rgb));
}

.user-usage-stats-cards__label,
.user-usage-stats-cards__meta {
  color: var(--theme-page-muted);
}

.user-usage-stats-cards__label {
  font-size: 0.75rem;
  font-weight: 500;
}

.user-usage-stats-cards__meta {
  font-size: 0.75rem;
}

.user-usage-stats-cards__value {
  color: var(--theme-page-text);
  font-size: 1.25rem;
  font-weight: 700;
}

.user-usage-stats-cards__value--success {
  color: rgb(var(--theme-success-rgb));
}

.user-usage-stats-cards__meta-strike {
  text-decoration: line-through;
}
</style>
