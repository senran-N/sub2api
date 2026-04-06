<template>
  <div class="user-dashboard-stats__grid">
    <div v-if="!isSimple" class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--success">
          <svg
            class="user-dashboard-stats__icon-symbol h-5 w-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
            />
          </svg>
        </div>
        <div class="min-w-0">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.balance') }}
          </p>
          <p class="user-dashboard-stats__value user-dashboard-stats__value--success text-xl font-bold">
            ${{ formatBalance(balance) }}
          </p>
          <p class="user-dashboard-stats__muted text-xs">{{ t('common.available') }}</p>
        </div>
      </div>
    </div>

    <div class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--info">
          <Icon name="key" size="md" class="user-dashboard-stats__icon-symbol" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.apiKeys') }}
          </p>
          <p class="user-dashboard-stats__value text-xl font-bold">
            {{ stats?.total_api_keys || 0 }}
          </p>
          <p class="user-dashboard-stats__meta user-dashboard-stats__meta--success text-xs">
            {{ stats?.active_api_keys || 0 }} {{ t('common.active') }}
          </p>
        </div>
      </div>
    </div>

    <div class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--success">
          <Icon name="chart" size="md" class="user-dashboard-stats__icon-symbol" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.todayRequests') }}
          </p>
          <p class="user-dashboard-stats__value text-xl font-bold">
            {{ stats?.today_requests || 0 }}
          </p>
          <p class="user-dashboard-stats__muted text-xs">
            {{ t('common.total') }}: {{ formatNumber(stats?.total_requests || 0) }}
          </p>
        </div>
      </div>
    </div>

    <div class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--purple">
          <Icon name="dollar" size="md" class="user-dashboard-stats__icon-symbol" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.todayCost') }}
          </p>
          <p class="user-dashboard-stats__value text-xl font-bold">
            <span class="user-dashboard-stats__meta user-dashboard-stats__meta--purple" :title="t('dashboard.actual')">
              ${{ formatCost(stats?.today_actual_cost || 0) }}
            </span>
            <span class="user-dashboard-stats__meta user-dashboard-stats__meta--muted text-sm font-normal" :title="t('dashboard.standard')">
              / ${{ formatCost(stats?.today_cost || 0) }}
            </span>
          </p>
          <p class="text-xs">
            <span class="user-dashboard-stats__muted">{{ t('common.total') }}: </span>
            <span class="user-dashboard-stats__meta user-dashboard-stats__meta--purple" :title="t('dashboard.actual')">
              ${{ formatCost(stats?.total_actual_cost || 0) }}
            </span>
            <span class="user-dashboard-stats__meta user-dashboard-stats__meta--muted" :title="t('dashboard.standard')">
              / ${{ formatCost(stats?.total_cost || 0) }}
            </span>
          </p>
        </div>
      </div>
    </div>
  </div>

  <div class="user-dashboard-stats__grid">
    <div class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--warning">
          <Icon name="cube" size="md" class="user-dashboard-stats__icon-symbol" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.todayTokens') }}
          </p>
          <p class="user-dashboard-stats__value text-xl font-bold">
            {{ formatTokens(stats?.today_tokens || 0) }}
          </p>
          <p
            class="user-dashboard-stats__muted truncate text-xs"
            :title="`${t('dashboard.input')}: ${formatTokens(stats?.today_input_tokens || 0)} / ${t('dashboard.output')}: ${formatTokens(stats?.today_output_tokens || 0)}`"
          >
            {{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} /
            {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }}
          </p>
        </div>
      </div>
    </div>

    <div class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--info">
          <Icon
            name="database"
            size="md"
            class="user-dashboard-stats__icon-symbol"
            :stroke-width="2"
          />
        </div>
        <div class="min-w-0">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.totalTokens') }}
          </p>
          <p class="user-dashboard-stats__value text-xl font-bold">
            {{ formatTokens(stats?.total_tokens || 0) }}
          </p>
          <p
            class="user-dashboard-stats__muted truncate text-xs"
            :title="`${t('dashboard.input')}: ${formatTokens(stats?.total_input_tokens || 0)} / ${t('dashboard.output')}: ${formatTokens(stats?.total_output_tokens || 0)}`"
          >
            {{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} /
            {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}
          </p>
        </div>
      </div>
    </div>

    <div class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--purple">
          <Icon name="bolt" size="md" class="user-dashboard-stats__icon-symbol" :stroke-width="2" />
        </div>
        <div class="flex-1">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.performance') }}
          </p>
          <div class="flex items-baseline gap-2">
            <p class="user-dashboard-stats__value text-xl font-bold">
              {{ formatTokens(stats?.rpm || 0) }}
            </p>
            <span class="user-dashboard-stats__muted text-xs">RPM</span>
          </div>
          <div class="flex items-baseline gap-2">
            <p class="user-dashboard-stats__meta user-dashboard-stats__meta--purple text-sm font-semibold">
              {{ formatTokens(stats?.tpm || 0) }}
            </p>
            <span class="user-dashboard-stats__muted text-xs">TPM</span>
          </div>
        </div>
      </div>
    </div>

    <div class="user-dashboard-stats__card card">
      <div class="user-dashboard-stats__card-content">
        <div class="user-dashboard-stats__icon user-dashboard-stats__icon--rose">
          <Icon name="clock" size="md" class="user-dashboard-stats__icon-symbol" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="user-dashboard-stats__label text-xs font-medium">
            {{ t('dashboard.avgResponse') }}
          </p>
          <p class="user-dashboard-stats__value text-xl font-bold">
            {{ formatDuration(stats?.average_duration_ms || 0) }}
          </p>
          <p class="user-dashboard-stats__muted text-xs">{{ t('dashboard.averageTime') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
}>()

const { t } = useI18n()

const formatBalance = (balanceValue: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(balanceValue)

const formatNumber = (value: number) => value.toLocaleString()
const formatCost = (value: number) => value.toFixed(4)
const formatTokens = (value: number) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toString()
}
const formatDuration = (ms: number) => (ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`)
</script>

<style scoped>
.user-dashboard-stats__grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: var(--theme-table-layout-gap);
}

.user-dashboard-stats__card {
  padding: var(--theme-stat-card-padding);
}

.user-dashboard-stats__card-content {
  display: flex;
  align-items: center;
  gap: var(--theme-stat-card-gap);
}

.user-dashboard-stats__label,
.user-dashboard-stats__muted,
.user-dashboard-stats__meta--muted {
  color: var(--theme-page-muted);
}

.user-dashboard-stats__value {
  color: var(--theme-page-text);
}

.user-dashboard-stats__value--success,
.user-dashboard-stats__meta--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-dashboard-stats__meta--purple {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.user-dashboard-stats__icon {
  --user-dashboard-tone-rgb: var(--theme-info-rgb);
  display: flex;
  width: var(--theme-stat-icon-size);
  height: var(--theme-stat-icon-size);
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: var(--theme-stat-icon-radius);
  background: color-mix(in srgb, rgb(var(--user-dashboard-tone-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--user-dashboard-tone-rgb)) 84%, var(--theme-page-text));
}

.user-dashboard-stats__icon--success {
  --user-dashboard-tone-rgb: var(--theme-success-rgb);
}

.user-dashboard-stats__icon--info {
  --user-dashboard-tone-rgb: var(--theme-info-rgb);
}

.user-dashboard-stats__icon--warning {
  --user-dashboard-tone-rgb: var(--theme-warning-rgb);
}

.user-dashboard-stats__icon--purple {
  --user-dashboard-tone-rgb: var(--theme-brand-purple-rgb);
}

.user-dashboard-stats__icon--rose {
  --user-dashboard-tone-rgb: var(--theme-brand-rose-rgb);
}

@media (min-width: 640px) {
  .user-dashboard-stats__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .user-dashboard-stats__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
