<template>
  <div class="card">
    <div class="user-dashboard-recent-usage__header flex items-center justify-between border-b">
      <h2 class="user-dashboard-recent-usage__title text-lg font-semibold">{{ t('dashboard.recentUsage') }}</h2>
      <span class="badge badge-gray">{{ t('dashboard.last7Days') }}</span>
    </div>
    <div class="user-dashboard-recent-usage__body">
      <div v-if="loading" class="user-dashboard-recent-usage__loading flex items-center justify-center">
        <LoadingSpinner size="lg" />
      </div>
      <div v-else-if="data.length === 0" class="user-dashboard-recent-usage__empty">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <div v-else class="space-y-3">
        <div v-for="log in data" :key="log.id" class="user-dashboard-recent-usage__item flex items-center justify-between transition-colors">
          <div class="flex items-center gap-4">
            <div class="user-dashboard-recent-usage__icon-shell flex items-center justify-center">
              <Icon name="beaker" size="md" class="user-dashboard-recent-usage__icon" />
            </div>
            <div>
              <p class="user-dashboard-recent-usage__title text-sm font-medium">{{ log.model }}</p>
              <p class="user-dashboard-recent-usage__meta text-xs">{{ formatDateTime(log.created_at) }}</p>
            </div>
          </div>
          <div class="text-right">
            <p class="text-sm font-semibold">
              <span class="user-dashboard-recent-usage__actual" :title="t('dashboard.actual')">${{ formatCost(log.actual_cost) }}</span>
              <span class="user-dashboard-recent-usage__meta font-normal" :title="t('dashboard.standard')"> / ${{ formatCost(log.total_cost) }}</span>
            </p>
            <p class="user-dashboard-recent-usage__meta text-xs">{{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens</p>
          </div>
        </div>

        <router-link to="/usage" class="user-dashboard-recent-usage__link flex items-center justify-center gap-2 text-sm font-medium transition-colors">
          {{ t('dashboard.viewAllUsage') }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
}>()
const { t } = useI18n()
const formatCost = (c: number) => c.toFixed(4)
</script>

<style scoped>
.user-dashboard-recent-usage__header {
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  padding: var(--theme-user-recent-usage-header-padding-y)
    var(--theme-user-recent-usage-header-padding-x);
}

.user-dashboard-recent-usage__body {
  padding: var(--theme-user-recent-usage-body-padding);
}

.user-dashboard-recent-usage__loading {
  padding-block: var(--theme-user-recent-usage-loading-padding-y);
}

.user-dashboard-recent-usage__empty {
  padding-block: var(--theme-user-recent-usage-empty-padding-y);
}

.user-dashboard-recent-usage__title {
  color: var(--theme-page-text);
}

.user-dashboard-recent-usage__item {
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
  border-radius: var(--theme-user-recent-usage-item-radius);
  padding: var(--theme-user-recent-usage-item-padding);
}

.user-dashboard-recent-usage__item:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 72%, var(--theme-surface));
}

.user-dashboard-recent-usage__icon-shell {
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
  width: var(--theme-user-recent-usage-icon-size);
  height: var(--theme-user-recent-usage-icon-size);
  border-radius: var(--theme-user-recent-usage-item-radius);
}

.user-dashboard-recent-usage__icon,
.user-dashboard-recent-usage__link {
  color: var(--theme-accent);
}

.user-dashboard-recent-usage__meta {
  color: var(--theme-page-muted);
}

.user-dashboard-recent-usage__actual {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-dashboard-recent-usage__link {
  padding-block: var(--theme-user-recent-usage-link-padding-y);
}

.user-dashboard-recent-usage__link:hover {
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}
</style>
