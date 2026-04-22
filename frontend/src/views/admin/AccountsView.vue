<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap-reverse items-start justify-between gap-3">
          <AccountTableFilters
            v-model:searchQuery="params.search"
            :filters="params"
            :groups="groups"
            @update:filters="(newFilters) => Object.assign(params, newFilters)"
            @change="debouncedReload"
            @update:searchQuery="debouncedReload"
          />
          <AccountTableActions
            :loading="loading"
            @refresh="handleManualRefresh"
            @sync="showSync = true"
            @create="showCreate = true"
          >
            <template #after>
              <AccountToolbarControls
                :auto-refresh-enabled="autoRefreshEnabled"
                :auto-refresh-countdown="autoRefreshCountdown"
                :auto-refresh-intervals="autoRefreshIntervals"
                :auto-refresh-interval-seconds="autoRefreshIntervalSeconds"
                :auto-refresh-interval-label="autoRefreshIntervalLabel"
                :toggleable-columns="toggleableColumns"
                :is-column-visible="isColumnVisible"
                @toggle-auto-refresh-enabled="setAutoRefreshEnabled(!autoRefreshEnabled)"
                @set-auto-refresh-interval="setAutoRefreshInterval"
                @toggle-column="toggleColumn"
                @error-passthrough="showErrorPassthrough = true"
                @tls-profiles="showTLSFingerprintProfiles = true"
              />
            </template>
            <template #beforeCreate>
              <AccountSecondaryActions
                :selected-count="selIds.length"
                @import="showImportData = true"
                @export="openExportDataDialog"
              />
            </template>
          </AccountTableActions>
        </div>
        <AccountPendingSyncBanner
          v-if="hasPendingListSync"
          @sync="syncPendingListChanges"
        />
      </template>
      <template #table>
        <AccountBulkActionsBar :selected-ids="selIds" @delete="handleBulkDelete" @reset-status="handleBulkResetStatus" @refresh-token="handleBulkRefreshToken" @edit="showBulkEdit = true" @clear="clearSelection" @select-page="selectPage" @toggle-schedulable="handleBulkToggleSchedulable" />
        <div ref="accountTableRef" class="flex min-h-0 flex-1 flex-col overflow-hidden">
          <DataTable
            :columns="cols"
            :data="accounts"
            :loading="loading"
            row-key="id"
            default-sort-key="name"
            default-sort-order="asc"
            :sort-storage-key="ACCOUNT_SORT_STORAGE_KEY"
          >
            <template #header-select>
              <AccountSelectionCheckbox
                id="accounts-select-all-visible"
                :checked="allVisibleSelected"
                name="select_all_visible_accounts"
                :aria-label="t('admin.accounts.bulkActions.selectCurrentPage')"
                @change="toggleSelectAllVisible"
              />
            </template>
            <template #cell-select="{ row }">
              <AccountSelectionCheckbox
                :id="`account-select-${row.id}`"
                :checked="selIds.includes(row.id)"
                :name="`account_select_${row.id}`"
                :aria-label="t('admin.accounts.selectAccount', { name: row.name || `#${row.id}` })"
                @change="toggleSel(row.id)"
              />
            </template>
            <template #cell-name="{ row }">
              <AccountNameCell :account="row" />
            </template>
            <template #cell-notes="{ value }">
              <AccountNotesCell :notes="value" />
            </template>
            <template #cell-platform_type="{ row }">
              <AccountPlatformTypeCell :account="row" />
            </template>
            <template #cell-capacity="{ row }">
              <AccountCapacityCell :account="row" />
            </template>
            <template #cell-status="{ row }">
              <AccountStatusIndicator :account="row" @show-temp-unsched="handleShowTempUnsched" />
            </template>
            <template #cell-schedulable="{ row }">
              <AccountSchedulableToggle
                :account="row"
                :loading="togglingSchedulable === row.id"
                @toggle="handleToggleSchedulable"
              />
            </template>
            <template #cell-today_stats="{ row }">
              <AccountTodayStatsCell
                :stats="todayStatsByAccountId[String(row.id)] ?? null"
                :loading="todayStatsLoading"
                :error="todayStatsError"
              />
            </template>
            <template #cell-groups="{ row }">
              <AccountGroupsCell :groups="row.groups" :max-display="4" />
            </template>
            <template #cell-usage="{ row }">
              <AccountUsageCell
                :account="row"
                :today-stats="todayStatsByAccountId[String(row.id)] ?? null"
                :today-stats-loading="todayStatsLoading"
                :manual-refresh-token="usageManualRefreshToken"
              />
            </template>
            <template #cell-proxy="{ row }">
              <AccountProxyCell :proxy="row.proxy" />
            </template>
            <template #cell-rate_multiplier="{ row }">
              <AccountRateMultiplierCell :rate-multiplier="row.rate_multiplier" />
            </template>
            <template #cell-priority="{ value }">
              <AccountPriorityCell :value="value" />
            </template>
            <template #cell-last_used_at="{ value }">
              <AccountLastUsedCell :value="value" />
            </template>
            <template #cell-expires_at="{ row, value }">
              <AccountExpiresCell :account="row" :value="value" />
            </template>
            <template #cell-actions="{ row }">
              <AccountActionsCell
                :account="row"
                @edit="handleEdit"
                @delete="handleDelete"
                @open-menu="openMenu"
              />
            </template>
          </DataTable>
        </div>
      </template>
      <template #pagination><Pagination v-if="pagination.total > 0" :page="pagination.page" :total="pagination.total" :page-size="pagination.page_size" @update:page="handlePageChange" @update:pageSize="handlePageSizeChange" /></template>
    </TablePageLayout>
    <CreateAccountModal v-if="showCreate" :show="showCreate" :proxies="proxies" :groups="groups" @close="showCreate = false" @created="reload" />
    <EditAccountModal v-if="showEdit" :show="showEdit" :account="edAcc" :proxies="proxies" :groups="groups" @close="showEdit = false" @updated="handleAccountUpdated" />
    <ReAuthAccountModal v-if="showReAuth" :show="showReAuth" :account="reAuthAcc" @close="closeReAuthModal" @reauthorized="handleAccountUpdated" />
    <AccountTestModal v-if="showTest" :show="showTest" :account="testingAcc" @close="closeTestModal" />
    <AccountStatsModal v-if="showStats" :show="showStats" :account="statsAcc" @close="closeStatsModal" />
    <ScheduledTestsPanel v-if="showSchedulePanel" :show="showSchedulePanel" :account-id="scheduleAcc?.id ?? null" :model-options="scheduleModelOptions" @close="closeSchedulePanel" />
    <AccountActionMenu :show="menu.show" :account="menu.acc" :position="menu.pos" @close="menu.show = false" @test="handleTest" @stats="handleViewStats" @schedule="handleSchedule" @reauth="handleReAuth" @refresh-token="handleRefresh" @recover-state="handleRecoverState" @reset-quota="handleResetQuota" @set-privacy="handleSetPrivacy" />
    <SyncFromCrsModal v-if="showSync" :show="showSync" @close="showSync = false" @synced="reload" />
    <ImportDataModal v-if="showImportData" :show="showImportData" @close="showImportData = false" @imported="handleDataImported" />
    <BulkEditAccountModal v-if="showBulkEdit" :show="showBulkEdit" :account-ids="selIds" :selected-platforms="selPlatforms" :selected-types="selTypes" :proxies="proxies" :groups="groups" @close="showBulkEdit = false" @updated="handleBulkUpdated" />
    <TempUnschedStatusModal v-if="showTempUnsched" :show="showTempUnsched" :account="tempUnschedAcc" @close="showTempUnsched = false" @reset="handleTempUnschedReset" />
    <ConfirmDialog :show="showDeleteDialog" :title="t('admin.accounts.deleteAccount')" :message="t('admin.accounts.deleteConfirm', { name: deletingAcc?.name })" :confirm-text="t('common.delete')" :cancel-text="t('common.cancel')" :danger="true" @confirm="confirmDelete" @cancel="showDeleteDialog = false" />
    <ConfirmDialog :show="showExportDataDialog" :title="t('admin.accounts.dataExport')" :message="t('admin.accounts.dataExportConfirmMessage')" :confirm-text="t('admin.accounts.dataExportConfirm')" :cancel-text="t('common.cancel')" @confirm="handleExportData" @cancel="showExportDataDialog = false">
      <AccountExportDialogOptions v-model="includeProxyOnExport" />
    </ConfirmDialog>
    <ErrorPassthroughRulesModal v-if="showErrorPassthrough" :show="showErrorPassthrough" @close="showErrorPassthrough = false" />
    <TLSFingerprintProfilesModal v-if="showTLSFingerprintProfiles" :show="showTLSFingerprintProfiles" @close="showTLSFingerprintProfiles = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { adminAPI } from '@/api/admin'
import { useTableLoader } from '@/composables/useTableLoader'
import { useSwipeSelect } from '@/composables/useSwipeSelect'
import { useTableSelection } from '@/composables/useTableSelection'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import AccountTableActions from '@/components/admin/account/AccountTableActions.vue'
import AccountTableFilters from '@/components/admin/account/AccountTableFilters.vue'
import AccountBulkActionsBar from '@/components/admin/account/AccountBulkActionsBar.vue'
import AccountActionMenu from '@/components/admin/account/AccountActionMenu.vue'
import AccountStatusIndicator from '@/components/account/AccountStatusIndicator.vue'
import AccountUsageCell from '@/components/account/AccountUsageCell.vue'
import AccountTodayStatsCell from '@/components/account/AccountTodayStatsCell.vue'
import AccountGroupsCell from '@/components/account/AccountGroupsCell.vue'
import AccountCapacityCell from '@/components/account/AccountCapacityCell.vue'
import type { Account } from '@/types'
import AccountActionsCell from './accounts/AccountActionsCell.vue'
import AccountExpiresCell from './accounts/AccountExpiresCell.vue'
import AccountExportDialogOptions from './accounts/AccountExportDialogOptions.vue'
import AccountLastUsedCell from './accounts/AccountLastUsedCell.vue'
import AccountNameCell from './accounts/AccountNameCell.vue'
import AccountNotesCell from './accounts/AccountNotesCell.vue'
import AccountPendingSyncBanner from './accounts/AccountPendingSyncBanner.vue'
import AccountPlatformTypeCell from './accounts/AccountPlatformTypeCell.vue'
import AccountPriorityCell from './accounts/AccountPriorityCell.vue'
import AccountProxyCell from './accounts/AccountProxyCell.vue'
import AccountRateMultiplierCell from './accounts/AccountRateMultiplierCell.vue'
import AccountSecondaryActions from './accounts/AccountSecondaryActions.vue'
import AccountSelectionCheckbox from './accounts/AccountSelectionCheckbox.vue'
import AccountSchedulableToggle from './accounts/AccountSchedulableToggle.vue'
import AccountToolbarControls from './accounts/AccountToolbarControls.vue'
import {
  ACCOUNT_SORT_STORAGE_KEY,
  type AccountListQuery
} from './accounts/accountsList'
import { useAccountsViewColumns } from './accounts/useAccountsViewColumns'
import { useAccountsViewActions } from './accounts/useAccountsViewActions'
import { useAccountsViewBootstrap } from './accounts/useAccountsViewBootstrap'
import { useAccountsViewDialogs } from './accounts/useAccountsViewDialogs'
import {
  downloadAccountsExportJson,
  useAccountsViewExport
} from './accounts/useAccountsViewExport'
import { useAccountsViewRefresh } from './accounts/useAccountsViewRefresh'
import { useAccountsViewState } from './accounts/useAccountsViewState'

const CreateAccountModal = defineAsyncComponent(() => import('@/components/account/CreateAccountModal.vue'))
const EditAccountModal = defineAsyncComponent(() => import('@/components/account/EditAccountModal.vue'))
const BulkEditAccountModal = defineAsyncComponent(() => import('@/components/account/BulkEditAccountModal.vue'))
const SyncFromCrsModal = defineAsyncComponent(() => import('@/components/account/SyncFromCrsModal.vue'))
const TempUnschedStatusModal = defineAsyncComponent(() => import('@/components/account/TempUnschedStatusModal.vue'))
const ImportDataModal = defineAsyncComponent(() => import('@/components/admin/account/ImportDataModal.vue'))
const ReAuthAccountModal = defineAsyncComponent(() => import('@/components/admin/account/ReAuthAccountModal.vue'))
const AccountTestModal = defineAsyncComponent(() => import('@/components/admin/account/AccountTestModal.vue'))
const AccountStatsModal = defineAsyncComponent(() => import('@/components/admin/account/AccountStatsModal.vue'))
const ScheduledTestsPanel = defineAsyncComponent(() => import('@/components/admin/account/ScheduledTestsPanel.vue'))
const ErrorPassthroughRulesModal = defineAsyncComponent(() => import('@/components/admin/ErrorPassthroughRulesModal.vue'))
const TLSFingerprintProfilesModal = defineAsyncComponent(() => import('@/components/admin/TLSFingerprintProfilesModal.vue'))

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const accountTableRef = ref<HTMLElement | null>(null)

const {
  showCreate,
  showSync,
  showImportData,
  showExportDataDialog,
  includeProxyOnExport,
  showBulkEdit,
  showErrorPassthrough,
  showTLSFingerprintProfiles,
  menu,
  closeActionMenu,
  syncMenuAccount,
  openMenu,
  openExportDataDialog
} = useAccountsViewDialogs()

const {
  items: accounts,
  loading,
  params,
  pagination,
  load: baseLoad,
  reload: baseReload,
  debouncedReload: baseDebouncedReload,
  handlePageChange: baseHandlePageChange,
  handlePageSizeChange: baseHandlePageSizeChange
} = useTableLoader<Account, AccountListQuery>({
  fetchFn: adminAPI.accounts.list,
  initialParams: { platform: '', type: '', status: '', privacy_mode: '', group: '', search: '' }
})

const {
  selectedIds: selIds,
  allVisibleSelected,
  isSelected,
  setSelectedIds,
  select,
  deselect,
  toggle: toggleSel,
  clear: clearSelection,
  removeMany: removeSelectedAccounts,
  toggleVisible,
  selectVisible: selectPage
} = useTableSelection<Account>({
  rows: accounts,
  getId: (account) => account.id
})

useSwipeSelect(accountTableRef, {
  isSelected,
  select,
  deselect
})

const {
  selPlatforms,
  selTypes,
  showEdit,
  showTempUnsched,
  showDeleteDialog,
  showReAuth,
  showTest,
  showStats,
  showSchedulePanel,
  edAcc,
  tempUnschedAcc,
  deletingAcc,
  reAuthAcc,
  testingAcc,
  statsAcc,
  scheduleAcc,
  scheduleModelOptions,
  togglingSchedulable,
  isAnyModalOpen,
  syncAccountRefs,
  toggleSelectAllVisible,
  updateSchedulableInList,
  handleBulkUpdated,
  handleDataImported,
  patchAccountInList
} = useAccountsViewState({
  accounts,
  isSelected,
  toggleVisible,
  clearSelection,
  reload: () => reload(),
  params,
  pagination,
  getHasPendingListSync: () => hasPendingListSync.value,
  setHasPendingListSync: (value) => {
    hasPendingListSync.value = value
  },
  removeSelectedAccounts,
  menu,
  syncMenuAccount,
  showCreate,
  showSync,
  showImportData,
  showExportDataDialog,
  showBulkEdit,
  showErrorPassthrough
})

const isActionMenuOpen = computed(() => menu.show)

let refreshUsageColumnStats = () => Promise.resolve()

const {
  hiddenColumns,
  toggleableColumns,
  cols,
  isColumnVisible,
  toggleColumn,
  autoRefreshIntervalLabel
} = useAccountsViewColumns({
  t,
  isSimpleMode: computed(() => authStore.isSimpleMode),
  onUsageColumnShown: () => refreshUsageColumnStats()
})

const {
  autoRefreshIntervals,
  autoRefreshEnabled,
  autoRefreshIntervalSeconds,
  autoRefreshCountdown,
  hasPendingListSync,
  todayStatsByAccountId,
  todayStatsLoading,
  todayStatsError,
  usageManualRefreshToken,
  bumpUsageManualRefreshToken,
  refreshTodayStatsBatch,
  load,
  reload,
  debouncedReload,
  handlePageChange,
  handlePageSizeChange,
  handleManualRefresh,
  syncPendingListChanges,
  setAutoRefreshEnabled,
  setAutoRefreshInterval,
  enterAutoRefreshSilentWindow,
  initializeAutoRefresh,
  dispose: disposeAccountsViewRefresh
} = useAccountsViewRefresh({
  accounts,
  loading,
  params,
  pagination,
  hiddenColumns,
  isAnyModalOpen,
  isActionMenuOpen,
  loadBase: baseLoad,
  reloadBase: baseReload,
  debouncedReloadBase: baseDebouncedReload,
  handlePageChangeBase: baseHandlePageChange,
  handlePageSizeChangeBase: baseHandlePageSizeChange,
  fetchTodayStats: adminAPI.accounts.getBatchTodayStats,
  fetchAccountsIncrementally: adminAPI.accounts.listWithEtag,
  syncAccountRefs
})
refreshUsageColumnStats = refreshTodayStatsBatch

const {
  handleEdit,
  handleBulkDelete,
  handleBulkResetStatus,
  handleBulkRefreshToken,
  handleBulkToggleSchedulable,
  handleAccountUpdated,
  closeTestModal,
  closeStatsModal,
  closeReAuthModal,
  handleTest,
  handleViewStats,
  handleSchedule,
  closeSchedulePanel,
  handleReAuth,
  handleRefresh,
  handleRecoverState,
  handleResetQuota,
  handleSetPrivacy,
  handleDelete,
  confirmDelete,
  handleToggleSchedulable,
  handleShowTempUnsched,
  handleTempUnschedReset
} = useAccountsViewActions({
  showEdit,
  showTempUnsched,
  showDeleteDialog,
  showReAuth,
  showTest,
  showStats,
  showSchedulePanel,
  edAcc,
  tempUnschedAcc,
  deletingAcc,
  reAuthAcc,
  testingAcc,
  statsAcc,
  scheduleAcc,
  scheduleModelOptions,
  togglingSchedulable,
  getSelectedIds: () => selIds.value,
  confirmAction: () => confirm(t('common.confirm')),
  clearSelection,
  setSelectedIds,
  load,
  reload,
  patchAccountInList,
  updateSchedulableInList,
  enterAutoRefreshSilentWindow,
  refreshUsageCells: bumpUsageManualRefreshToken,
  t,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError
})

const { handleExportData } = useAccountsViewExport({
  t,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError,
  selectedIds: selIds,
  includeProxyOnExport,
  params,
  showExportDataDialog,
  exportData: adminAPI.accounts.exportData,
  downloadJson: downloadAccountsExportJson
})

const { proxies, groups, initialize, dispose } = useAccountsViewBootstrap({
  t,
  showError: appStore.showError,
  load,
  fetchProxies: adminAPI.proxies.getAll,
  fetchGroups: adminAPI.groups.getAll,
  closeActionMenu,
  initializeAutoRefresh,
  disposeAutoRefresh: disposeAccountsViewRefresh
})

onMounted(() => {
  initialize()
})

onUnmounted(() => {
  dispose()
})
</script>
