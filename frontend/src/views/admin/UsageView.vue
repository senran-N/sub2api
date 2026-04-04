<template>
  <AppLayout>
    <div class="space-y-6">
      <UsageStatsCards :stats="usageStats" />
      <div class="space-y-4">
        <UsageChartsToolbar
          :start-date="startDate"
          :end-date="endDate"
          :granularity="granularity"
          :granularity-options="granularityOptions"
          @update:start-date="startDate = $event"
          @update:end-date="endDate = $event"
          @update:granularity="granularity = $event"
          @date-range-change="onDateRangeChange"
          @granularity-change="loadChartData"
        />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <ModelDistributionChart
            v-model:source="modelDistributionSource"
            v-model:metric="modelDistributionMetric"
            :model-stats="requestedModelStats"
            :upstream-model-stats="upstreamModelStats"
            :mapping-model-stats="mappingModelStats"
            :loading="modelStatsLoading"
            :show-source-toggle="true"
            :show-metric-toggle="true"
            :start-date="startDate"
            :end-date="endDate"
          />
          <GroupDistributionChart
            v-model:metric="groupDistributionMetric"
            :group-stats="groupStats"
            :loading="chartsLoading"
            :show-metric-toggle="true"
            :start-date="startDate"
            :end-date="endDate"
          />
        </div>
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <EndpointDistributionChart
            v-model:source="endpointDistributionSource"
            v-model:metric="endpointDistributionMetric"
            :endpoint-stats="inboundEndpointStats"
            :upstream-endpoint-stats="upstreamEndpointStats"
            :endpoint-path-stats="endpointPathStats"
            :loading="endpointStatsLoading"
            :show-source-toggle="true"
            :show-metric-toggle="true"
            :title="t('usage.endpointDistribution')"
            :start-date="startDate"
            :end-date="endDate"
          />
          <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
        </div>
      </div>
      <UsageFilters v-model="filters" :start-date="startDate" :end-date="endDate" :exporting="exporting" @change="applyFilters" @refresh="refreshData" @reset="resetFilters" @cleanup="openCleanupDialog" @export="exportToExcel">
        <template #after-reset>
          <UsageColumnSettingsControl
            :toggleable-columns="toggleableColumns"
            :is-column-visible="isColumnVisible"
            @toggle-column="toggleColumn"
          />
        </template>
      </UsageFilters>
      <UsageTable :data="usageLogs" :loading="loading" :columns="visibleColumns" @userClick="handleUserClick" />
      <Pagination v-if="pagination.total > 0" :page="pagination.page" :total="pagination.total" :page-size="pagination.page_size" @update:page="handlePageChange" @update:pageSize="handlePageSizeChange" />
    </div>
  </AppLayout>
  <UsageExportProgress :show="exportProgress.show" :progress="exportProgress.progress" :current="exportProgress.current" :total="exportProgress.total" :estimated-time="exportProgress.estimatedTime" @cancel="cancelExport" />
  <UsageCleanupDialog
    :show="cleanupDialogVisible"
    :filters="filters"
    :start-date="startDate"
    :end-date="endDate"
    @close="closeCleanupDialog"
  />
  <!-- Balance history modal triggered from usage table user click -->
  <UserBalanceHistoryModal
    :show="showBalanceHistoryModal"
    :user="balanceHistoryUser"
    :hide-actions="true"
    @close="closeBalanceHistoryModal"
  />
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'
import UsageFilters from '@/components/admin/usage/UsageFilters.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import UsageExportProgress from '@/components/admin/usage/UsageExportProgress.vue'
import UsageCleanupDialog from '@/components/admin/usage/UsageCleanupDialog.vue'
import UserBalanceHistoryModal from '@/components/admin/user/UserBalanceHistoryModal.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import GroupDistributionChart from '@/components/charts/GroupDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import type { AdminUsageQueryParams } from '@/api/admin/usage'
import {
  applyUsageDateRangeState,
  applyUsageRouteQueryState,
  buildDefaultUsageFilters,
  buildResetUsageState,
  getLast24HoursUsageRange,
} from './usageViewState'
import {
  useUsageViewData,
  type ModelDistributionSource
} from './useUsageViewData'
import { useUsageViewColumns } from './useUsageViewColumns'
import { useUsageViewDialogs } from './useUsageViewDialogs'
import UsageChartsToolbar from './usage/UsageChartsToolbar.vue'
import UsageColumnSettingsControl from './usage/UsageColumnSettingsControl.vue'

const { t } = useI18n()
const appStore = useAppStore()
type DistributionMetric = 'tokens' | 'actual_cost'
type EndpointSource = 'inbound' | 'upstream' | 'path'
const route = useRoute()
const granularity = ref<'day' | 'hour'>('hour')
const modelDistributionMetric = ref<DistributionMetric>('tokens')
const modelDistributionSource = ref<ModelDistributionSource>('requested')
const groupDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionSource = ref<EndpointSource>('inbound')

const granularityOptions = computed<Array<{ value: 'day' | 'hour'; label: string }>>(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])
const defaultRange = getLast24HoursUsageRange()
const startDate = ref(defaultRange.startDate)
const endDate = ref(defaultRange.endDate)
const filters = ref<AdminUsageQueryParams>(buildDefaultUsageFilters(defaultRange))
const pagination = reactive({ page: 1, page_size: getPersistedPageSize(), total: 0 })
const {
  cleanupDialogVisible,
  showBalanceHistoryModal,
  balanceHistoryUser,
  openCleanupDialog,
  closeCleanupDialog,
  closeBalanceHistoryModal,
  handleUserClick
} = useUsageViewDialogs({
  fetchUserById: (userId) => adminAPI.users.getById(userId),
  showLoadUserError: () => appStore.showError(t('admin.usage.failedToLoadUser'))
})
const {
  usageStats,
  usageLogs,
  loading,
  exporting,
  trendData,
  requestedModelStats,
  upstreamModelStats,
  mappingModelStats,
  groupStats,
  chartsLoading,
  modelStatsLoading,
  inboundEndpointStats,
  upstreamEndpointStats,
  endpointPathStats,
  endpointStatsLoading,
  exportProgress,
  applyFilters,
  refreshData,
  loadModelStats,
  loadChartData,
  loadInitialData,
  handlePageChange,
  handlePageSizeChange,
  cancelExport,
  exportToExcel,
  dispose
} = useUsageViewData({
  filters,
  startDate,
  endDate,
  granularity,
  modelDistributionSource,
  pagination,
  t,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError
})

const applyRouteQueryFilters = () => {
  const nextState = applyUsageRouteQueryState(route.query, filters.value, {
    startDate: startDate.value,
    endDate: endDate.value
  })
  startDate.value = nextState.range.startDate
  endDate.value = nextState.range.endDate
  filters.value = nextState.filters
  granularity.value = nextState.granularity
}

const onDateRangeChange = (range: { startDate: string; endDate: string; preset: string | null }) => {
  const nextState = applyUsageDateRangeState(
    {
      startDate: range.startDate,
      endDate: range.endDate
    },
    filters.value
  )
  startDate.value = nextState.range.startDate
  endDate.value = nextState.range.endDate
  filters.value = nextState.filters
  granularity.value = nextState.granularity
  applyFilters()
}
const resetFilters = () => {
  const nextState = buildResetUsageState()
  startDate.value = nextState.range.startDate
  endDate.value = nextState.range.endDate
  filters.value = nextState.filters
  granularity.value = nextState.granularity
  applyFilters()
}

// Column visibility
const allColumns = computed(() => [
  { key: 'user', label: t('admin.usage.user'), sortable: false },
  { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
  { key: 'account', label: t('admin.usage.account'), sortable: false },
  { key: 'model', label: t('usage.model'), sortable: true },
  { key: 'reasoning_effort', label: t('usage.reasoningEffort'), sortable: false },
  { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
  { key: 'group', label: t('admin.usage.group'), sortable: false },
  { key: 'stream', label: t('usage.type'), sortable: false },
  { key: 'tokens', label: t('usage.tokens'), sortable: false },
  { key: 'cost', label: t('usage.cost'), sortable: false },
  { key: 'first_token', label: t('usage.firstToken'), sortable: false },
  { key: 'duration', label: t('usage.duration'), sortable: false },
  { key: 'created_at', label: t('usage.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent'), sortable: false },
  { key: 'ip_address', label: t('admin.usage.ipAddress'), sortable: false }
])
const {
  toggleableColumns,
  visibleColumns,
  isColumnVisible,
  toggleColumn
} = useUsageViewColumns({
  allColumns
})

onMounted(() => {
  applyRouteQueryFilters()
  loadInitialData()
})
onUnmounted(() => {
  dispose()
})

watch(modelDistributionSource, (source) => {
  void loadModelStats(source)
})
</script>
