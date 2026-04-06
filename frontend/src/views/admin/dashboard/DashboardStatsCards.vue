<template>
  <div class="dashboard-stats">
    <div class="dashboard-stats__grid">
      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--info">
            <Icon name="key" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.apiKeys') }}
            </p>
            <p class="dashboard-stats__value text-xl font-bold">
              {{ stats.total_api_keys }}
            </p>
            <p class="dashboard-stats__meta dashboard-stats__meta--success text-xs">
              {{ stats.active_api_keys }} {{ t('common.active') }}
            </p>
          </div>
        </div>
      </div>

      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--purple">
            <Icon name="server" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.accounts') }}
            </p>
            <p class="dashboard-stats__value text-xl font-bold">
              {{ stats.total_accounts }}
            </p>
            <p class="text-xs">
              <span class="dashboard-stats__meta dashboard-stats__meta--success">
                {{ stats.normal_accounts }} {{ t('common.active') }}
              </span>
              <span v-if="stats.error_accounts > 0" class="dashboard-stats__meta dashboard-stats__meta--danger ml-1">
                {{ stats.error_accounts }} {{ t('common.error') }}
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--success">
            <Icon name="chart" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.todayRequests') }}
            </p>
            <p class="dashboard-stats__value text-xl font-bold">
              {{ stats.today_requests }}
            </p>
            <p class="dashboard-stats__muted text-xs">
              {{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}
            </p>
          </div>
        </div>
      </div>

      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--success">
            <Icon name="userPlus" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.users') }}
            </p>
            <p class="dashboard-stats__value dashboard-stats__value--success text-xl font-bold">
              +{{ stats.today_new_users }}
            </p>
            <p class="dashboard-stats__muted text-xs">
              {{ t('common.total') }}: {{ formatNumber(stats.total_users) }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <div class="dashboard-stats__grid">
      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--warning">
            <Icon name="cube" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.todayTokens') }}
            </p>
            <p class="dashboard-stats__value text-xl font-bold">
              {{ formatTokens(stats.today_tokens) }}
            </p>
            <p class="text-xs">
              <span class="dashboard-stats__meta dashboard-stats__meta--warning" :title="t('admin.dashboard.actual')">
                ${{ formatCost(stats.today_actual_cost) }}
              </span>
              <span class="dashboard-stats__meta dashboard-stats__meta--muted" :title="t('admin.dashboard.standard')">
                / ${{ formatCost(stats.today_cost) }}
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--info">
            <Icon name="database" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.totalTokens') }}
            </p>
            <p class="dashboard-stats__value text-xl font-bold">
              {{ formatTokens(stats.total_tokens) }}
            </p>
            <p class="text-xs">
              <span class="dashboard-stats__meta dashboard-stats__meta--info" :title="t('admin.dashboard.actual')">
                ${{ formatCost(stats.total_actual_cost) }}
              </span>
              <span class="dashboard-stats__meta dashboard-stats__meta--muted" :title="t('admin.dashboard.standard')">
                / ${{ formatCost(stats.total_cost) }}
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--purple">
            <Icon name="bolt" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="flex-1">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.performance') }}
            </p>
            <div class="flex items-baseline gap-2">
              <p class="dashboard-stats__value text-xl font-bold">
                {{ formatTokens(stats.rpm) }}
              </p>
              <span class="dashboard-stats__muted text-xs">RPM</span>
            </div>
            <div class="flex items-baseline gap-2">
              <p class="dashboard-stats__meta dashboard-stats__meta--purple text-sm font-semibold">
                {{ formatTokens(stats.tpm) }}
              </p>
              <span class="dashboard-stats__muted text-xs">TPM</span>
            </div>
          </div>
        </div>
      </div>

      <div class="dashboard-stats__card card">
        <div class="dashboard-stats__card-content">
          <div class="dashboard-stats__icon dashboard-stats__icon--rose">
            <Icon name="clock" size="md" class="dashboard-stats__icon-symbol" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="dashboard-stats__label text-xs font-medium">
              {{ t('admin.dashboard.avgResponse') }}
            </p>
            <p class="dashboard-stats__value text-xl font-bold">
              {{ formatDuration(stats.average_duration_ms) }}
            </p>
            <p class="dashboard-stats__muted text-xs">
              {{ stats.active_users }} {{ t('admin.dashboard.activeUsers') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { DashboardStats } from '@/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  stats: DashboardStats
  formatTokens: (value: number | undefined | null) => string
  formatNumber: (value: number) => string
  formatCost: (value: number) => string
  formatDuration: (value: number) => string
}>()

const { t } = useI18n()
</script>

<style scoped>
.dashboard-stats {
  display: flex;
  flex-direction: column;
  gap: var(--theme-table-layout-gap);
}

.dashboard-stats__grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: var(--theme-table-layout-gap);
}

.dashboard-stats__card {
  padding: var(--theme-stat-card-padding);
}

.dashboard-stats__card-content {
  display: flex;
  align-items: center;
  gap: var(--theme-stat-card-gap);
}

.dashboard-stats__label,
.dashboard-stats__muted,
.dashboard-stats__meta--muted {
  color: var(--theme-page-muted);
}

.dashboard-stats__value {
  color: var(--theme-page-text);
}

.dashboard-stats__value--success,
.dashboard-stats__meta--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.dashboard-stats__meta--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.dashboard-stats__meta--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.dashboard-stats__meta--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.dashboard-stats__meta--purple {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.dashboard-stats__icon {
  --dashboard-stats-tone-rgb: var(--theme-info-rgb);
  display: flex;
  width: var(--theme-stat-icon-size);
  height: var(--theme-stat-icon-size);
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: var(--theme-stat-icon-radius);
  background: color-mix(in srgb, rgb(var(--dashboard-stats-tone-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--dashboard-stats-tone-rgb)) 84%, var(--theme-page-text));
}

.dashboard-stats__icon--info {
  --dashboard-stats-tone-rgb: var(--theme-info-rgb);
}

.dashboard-stats__icon--success {
  --dashboard-stats-tone-rgb: var(--theme-success-rgb);
}

.dashboard-stats__icon--warning {
  --dashboard-stats-tone-rgb: var(--theme-warning-rgb);
}

.dashboard-stats__icon--purple {
  --dashboard-stats-tone-rgb: var(--theme-brand-purple-rgb);
}

.dashboard-stats__icon--rose {
  --dashboard-stats-tone-rgb: var(--theme-brand-rose-rgb);
}

@media (min-width: 640px) {
  .dashboard-stats__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .dashboard-stats__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
