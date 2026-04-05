<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <UserUsageStatsCards :stats="usageStats" />
      </template>

      <template #filters>
        <UserUsageFiltersBar
          :api-key-id="filters.api_key_id"
          :api-key-options="apiKeyOptions"
          :start-date="startDate"
          :end-date="endDate"
          :loading="loading"
          :exporting="exporting"
          @update:api-key-id="filters.api_key_id = $event"
          @update:start-date="startDate = $event"
          @update:end-date="endDate = $event"
          @date-range-change="onDateRangeChange"
          @apply-filters="applyFilters"
          @reset="resetFilters"
          @export="exportToCSV"
        />
      </template>

      <template #table>
        <DataTable :columns="columns" :data="usageLogs" :loading="loading">
          <template #cell-api_key="{ row }">
            <span class="text-sm text-gray-900 dark:text-white">{{
              row.api_key?.name || '-'
            }}</span>
          </template>

          <template #cell-model="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
          </template>

          <template #cell-reasoning_effort="{ row }">
            <span class="text-sm text-gray-900 dark:text-white">
              {{ formatReasoningEffort(row.reasoning_effort) }}
            </span>
          </template>

          <template #cell-endpoint="{ row }">
            <UserUsageEndpointCell :endpoint="row.inbound_endpoint" />
          </template>

          <template #cell-stream="{ row }"><UserUsageRequestTypeBadge :log="row" /></template>

          <template #cell-tokens="{ row }">
            <UserUsageTokenCell
              :row="row"
              @show-details="showTokenTooltip"
              @hide-details="hideTokenTooltip"
            />
          </template>

          <template #cell-cost="{ row }">
            <UserUsageCostCell
              :row="row"
              @show-details="showTooltip"
              @hide-details="hideTooltip"
            />
          </template>

          <template #cell-first_token="{ row }">
            <UserUsageDurationCell :value="row.first_token_ms" />
          </template>

          <template #cell-duration="{ row }">
            <UserUsageDurationCell :value="row.duration_ms" />
          </template>

          <template #cell-created_at="{ value }">
            <UserUsageDateTimeCell :value="value" />
          </template>

          <template #cell-user_agent="{ row }">
            <UserUsageUserAgentCell :value="row.user_agent" />
          </template>

          <template #empty>
            <EmptyState :message="t('usage.noRecords')" />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>
  </AppLayout>

  <UserUsageHoverOverlays
    :token-tooltip-visible="tokenTooltipVisible"
    :token-tooltip-position="tokenTooltipPosition"
    :token-tooltip-data="tokenTooltipData"
    :tooltip-visible="tooltipVisible"
    :tooltip-position="tooltipPosition"
    :tooltip-data="tooltipData"
  />
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { formatReasoningEffort } from '@/utils/format'
import UserUsageFiltersBar from './usage/UserUsageFiltersBar.vue'
import UserUsageStatsCards from './usage/UserUsageStatsCards.vue'
import UserUsageEndpointCell from './usage/UserUsageEndpointCell.vue'
import UserUsageRequestTypeBadge from './usage/UserUsageRequestTypeBadge.vue'
import UserUsageTokenCell from './usage/UserUsageTokenCell.vue'
import UserUsageCostCell from './usage/UserUsageCostCell.vue'
import UserUsageDurationCell from './usage/UserUsageDurationCell.vue'
import UserUsageDateTimeCell from './usage/UserUsageDateTimeCell.vue'
import UserUsageUserAgentCell from './usage/UserUsageUserAgentCell.vue'
import UserUsageHoverOverlays from './usage/UserUsageHoverOverlays.vue'
import { useUserUsagePageViewModel } from './usage/useUserUsagePageViewModel'

const { t } = useI18n()
const {
  columns,
  usageStats,
  usageLogs,
  loading,
  exporting,
  apiKeyOptions,
  startDate,
  endDate,
  filters,
  pagination,
  tooltipVisible,
  tooltipPosition,
  tooltipData,
  showTooltip,
  hideTooltip,
  tokenTooltipVisible,
  tokenTooltipPosition,
  tokenTooltipData,
  showTokenTooltip,
  hideTokenTooltip,
  onDateRangeChange,
  applyFilters,
  resetFilters,
  handlePageChange,
  handlePageSizeChange,
  exportToCSV
} = useUserUsagePageViewModel()
</script>
