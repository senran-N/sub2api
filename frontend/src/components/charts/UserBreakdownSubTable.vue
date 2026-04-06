<template>
  <div class="user-breakdown-sub-table">
    <div v-if="loading" class="user-breakdown-sub-table__loading flex items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="items.length === 0" class="user-breakdown-sub-table__empty user-breakdown-sub-table__empty-spacing text-center text-xs">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
    <table v-else class="w-full text-xs">
      <tbody>
        <tr
          v-for="user in items"
          :key="user.user_id"
          class="user-breakdown-sub-table__row border-t"
        >
          <td class="user-breakdown-sub-table__user user-breakdown-sub-table__user-cell truncate" :title="user.email">
            {{ user.email || `User #${user.user_id}` }}
          </td>
          <td class="user-breakdown-sub-table__muted user-breakdown-sub-table__cell text-right">
            {{ user.requests.toLocaleString() }}
          </td>
          <td class="user-breakdown-sub-table__muted user-breakdown-sub-table__cell text-right">
            {{ formatTokens(user.total_tokens) }}
          </td>
          <td class="user-breakdown-sub-table__actual user-breakdown-sub-table__cell text-right">
            ${{ formatCost(user.actual_cost) }}
          </td>
          <td class="user-breakdown-sub-table__standard user-breakdown-sub-table__cell user-breakdown-sub-table__cell--last text-right">
            ${{ formatCost(user.cost) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { UserBreakdownItem } from '@/types'

const { t } = useI18n()

defineProps<{
  items: UserBreakdownItem[]
  loading?: boolean
}>()

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}

const formatCost = (value: number): string => {
  if (value >= 1000) return (value / 1000).toFixed(2) + 'K'
  if (value >= 1) return value.toFixed(2)
  if (value >= 0.01) return value.toFixed(3)
  return value.toFixed(4)
}
</script>

<style scoped>
.user-breakdown-sub-table {
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}

.user-breakdown-sub-table__loading {
  padding-block: var(--theme-breakdown-loading-padding-y);
}

.user-breakdown-sub-table__row {
  border-color: color-mix(in srgb, var(--theme-card-border) 60%, transparent);
}

.user-breakdown-sub-table__user {
  color: color-mix(in srgb, var(--theme-page-text) 88%, transparent);
}

.user-breakdown-sub-table__user-cell {
  max-width: var(--theme-breakdown-user-max-width);
  padding-block: var(--theme-breakdown-cell-padding-y);
  padding-inline-start: var(--theme-breakdown-user-padding-start);
}

.user-breakdown-sub-table__muted,
.user-breakdown-sub-table__standard,
.user-breakdown-sub-table__empty {
  color: var(--theme-page-muted);
}

.user-breakdown-sub-table__empty-spacing {
  padding-block: var(--theme-breakdown-empty-padding-y);
}

.user-breakdown-sub-table__cell {
  padding-block: var(--theme-breakdown-cell-padding-y);
}

.user-breakdown-sub-table__cell--last {
  padding-inline-end: var(--theme-breakdown-last-cell-padding-end);
}

.user-breakdown-sub-table__actual {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}
</style>
