<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <SubscriptionFiltersBar
          :filter-user-keyword="filterUserKeyword"
          :filter-user-results="filterUserResults"
          :filter-user-loading="filterUserLoading"
          :show-filter-user-dropdown="showFilterUserDropdown"
          :selected-filter-user="selectedFilterUser"
          :status="filters.status"
          :group-id="filters.group_id"
          :platform="filters.platform"
          :status-options="statusOptions"
          :group-options="groupOptions"
          :platform-filter-options="platformFilterOptions"
          :loading="loading"
          :user-column-mode="userColumnMode"
          :toggleable-columns="toggleableColumns"
          :is-column-visible="isColumnVisible"
          @update:filter-user-keyword="filterUserKeyword = $event"
          @search-filter-users="debounceSearchFilterUsers"
          @show-filter-user-dropdown="showFilterUserDropdown = true"
          @select-filter-user="selectFilterUser"
          @clear-filter-user="clearFilterUser"
          @update:status="setFilterStatus"
          @update:group-id="setFilterGroupId"
          @update:platform="setFilterPlatform"
          @apply-filters="applyFilters"
          @refresh="loadSubscriptions"
          @set-user-mode="setUserColumnMode"
          @toggle-column="toggleColumn"
          @guide="openGuideModal"
          @assign="openAssignModal"
        />
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="subscriptions"
          :loading="loading"
          :server-side-sort="true"
          @sort="handleSort"
        >
          <template #cell-user="{ row }">
            <SubscriptionUserCell :subscription="row" :mode="userColumnMode" />
          </template>

          <template #cell-group="{ row }">
            <GroupBadge
              v-if="row.group"
              :name="row.group.name"
              :platform="row.group.platform"
              :subscription-type="row.group.subscription_type"
              :rate-multiplier="row.group.rate_multiplier"
              :show-rate="false"
            />
            <span v-else class="theme-text-subtle text-sm">-</span>
          </template>

          <template #cell-usage="{ row }">
            <SubscriptionUsageCell :subscription="row" />
          </template>

          <template #cell-expires_at="{ value }">
            <SubscriptionExpirationCell :expires-at="value" />
          </template>

          <template #cell-status="{ value }">
            <SubscriptionStatusBadge :status="value" />
          </template>

          <template #cell-actions="{ row }">
            <SubscriptionActionsCell
              :subscription="row"
              :resetting="resettingQuota && resettingSubscription?.id === row.id"
              @adjust="handleExtend"
              @reset-quota="handleResetQuota"
              @revoke="handleRevoke"
            />
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.subscriptions.noSubscriptionsYet')"
              :description="t('admin.subscriptions.assignFirstSubscription')"
              :action-text="t('admin.subscriptions.assignSubscription')"
              @action="openAssignModal"
            />
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

    <SubscriptionAssignDialog
      :show="showAssignModal"
      :form="assignForm"
      :user-keyword="userSearchKeyword"
      :user-results="userSearchResults"
      :user-loading="userSearchLoading"
      :show-user-dropdown="showUserDropdown"
      :selected-user="selectedUser"
      :group-options="subscriptionGroupOptions"
      :submitting="submitting"
      @close="closeAssignModal"
      @submit="handleAssignSubscription"
      @update:user-keyword="userSearchKeyword = $event"
      @show-user-dropdown="showUserDropdown = true"
      @search-users="debounceSearchUsers"
      @select-user="selectUser"
      @clear-user="clearUserSelection"
    />

    <SubscriptionExtendDialog
      :show="showExtendModal"
      :subscription="extendingSubscription"
      :form="extendForm"
      :submitting="submitting"
      @close="closeExtendModal"
      @submit="handleExtendSubscription"
    />

    <ConfirmDialog
      :show="showRevokeDialog"
      :title="t('admin.subscriptions.revokeSubscription')"
      :message="t('admin.subscriptions.revokeConfirm', { user: revokingSubscription?.user?.email })"
      :confirm-text="t('admin.subscriptions.revoke')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmRevoke"
      @cancel="closeRevokeDialog"
    />

    <ConfirmDialog
      :show="showResetQuotaConfirm"
      :title="t('admin.subscriptions.resetQuotaTitle')"
      :message="t('admin.subscriptions.resetQuotaConfirm', { user: resettingSubscription?.user?.email })"
      :confirm-text="t('admin.subscriptions.resetQuota')"
      :cancel-text="t('common.cancel')"
      @confirm="confirmResetQuota"
      @cancel="closeResetQuotaConfirm"
    />
    <SubscriptionGuideModal :show="showGuideModal" @close="closeGuideModal" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { buildAdminPlatformOptions } from '@/components/admin/platformOptions'
import GroupBadge from '@/components/common/GroupBadge.vue'
import {
  buildSubscriptionGroupOptions
} from './subscriptions/subscriptionForm'
import SubscriptionActionsCell from './subscriptions/SubscriptionActionsCell.vue'
import SubscriptionAssignDialog from './subscriptions/SubscriptionAssignDialog.vue'
import SubscriptionExpirationCell from './subscriptions/SubscriptionExpirationCell.vue'
import SubscriptionFiltersBar from './subscriptions/SubscriptionFiltersBar.vue'
import SubscriptionExtendDialog from './subscriptions/SubscriptionExtendDialog.vue'
import SubscriptionGuideModal from './subscriptions/SubscriptionGuideModal.vue'
import SubscriptionStatusBadge from './subscriptions/SubscriptionStatusBadge.vue'
import SubscriptionUsageCell from './subscriptions/SubscriptionUsageCell.vue'
import SubscriptionUserCell from './subscriptions/SubscriptionUserCell.vue'
import { useSubscriptionsViewFormState } from './subscriptions/useSubscriptionsViewFormState'
import { useSubscriptionsViewColumns } from './subscriptions/useSubscriptionsViewColumns'
import { useSubscriptionsViewActions } from './subscriptions/useSubscriptionsViewActions'
import { useSubscriptionsViewData } from './subscriptions/useSubscriptionsViewData'
import { useSubscriptionsViewDialogs } from './subscriptions/useSubscriptionsViewDialogs'
import { useSubscriptionsViewUserSearches } from './subscriptions/useSubscriptionsViewUserSearches'

const { t } = useI18n()
const appStore = useAppStore()

// All available columns
const allColumns = computed<Column[]>(() => [
  {
    key: 'user',
    label: userColumnMode.value === 'email'
      ? t('admin.subscriptions.columns.user')
      : t('admin.users.columns.username'),
    sortable: false
  },
  { key: 'group', label: t('admin.subscriptions.columns.group'), sortable: false },
  { key: 'usage', label: t('admin.subscriptions.columns.usage'), sortable: false },
  { key: 'expires_at', label: t('admin.subscriptions.columns.expires'), sortable: true },
  { key: 'status', label: t('admin.subscriptions.columns.status'), sortable: true },
  { key: 'actions', label: t('admin.subscriptions.columns.actions'), sortable: false }
])
const {
  userColumnMode,
  toggleableColumns,
  columns,
  isColumnVisible,
  toggleColumn,
  setUserColumnMode
} = useSubscriptionsViewColumns({
  allColumns
})

// Filter options
const statusOptions = computed(() => [
  { value: '', label: t('admin.subscriptions.allStatus') },
  { value: 'active', label: t('admin.subscriptions.status.active') },
  { value: 'expired', label: t('admin.subscriptions.status.expired') },
  { value: 'revoked', label: t('admin.subscriptions.status.revoked') }
])

const {
  filters,
  assignForm,
  extendForm,
  setFilterStatus,
  setFilterGroupId,
  setFilterPlatform,
  selectFilterUser: applySelectedFilterUser,
  clearFilterUser: clearSelectedFilterUser,
  selectAssignUser: applySelectedAssignUser,
  clearAssignUser: clearSelectedAssignUser,
  resetAssignFormState,
  resetExtendFormState
} = useSubscriptionsViewFormState()
const {
  subscriptions,
  groups,
  loading,
  pagination,
  loadSubscriptions,
  applyFilters,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
  loadInitialData,
  dispose: disposeDataState
} = useSubscriptionsViewData({
  showLoadError: (message) => appStore.showError(message),
  t
})

const resetAssignModalState = () => {
  resetAssignFormState()
  resetAssignSearch()
}

const resetExtendModalState = () => {
  resetExtendFormState()
}

const {
  showGuideModal,
  showAssignModal,
  showExtendModal,
  showRevokeDialog,
  showResetQuotaConfirm,
  extendingSubscription,
  revokingSubscription,
  resettingSubscription,
  openGuideModal,
  closeGuideModal,
  openAssignModal,
  closeAssignModal,
  openExtendModal,
  closeExtendModal,
  openRevokeDialog,
  closeRevokeDialog,
  openResetQuotaConfirm,
  closeResetQuotaConfirm
} = useSubscriptionsViewDialogs({
  resetAssignState: resetAssignModalState,
  resetExtendState: resetExtendModalState
})
const {
  filterUserKeyword,
  filterUserResults,
  filterUserLoading,
  showFilterUserDropdown,
  selectedFilterUser,
  userSearchKeyword,
  userSearchResults,
  userSearchLoading,
  showUserDropdown,
  selectedUser,
  debounceSearchFilterUsers,
  selectFilterUser,
  clearFilterUser,
  debounceSearchUsers,
  selectUser,
  clearUserSelection,
  resetAssignSearch,
  initialize: initializeUserSearches,
  dispose: disposeUserSearches
} = useSubscriptionsViewUserSearches({
  applyFilters,
  searchUsers: (keyword) => adminAPI.usage.searchUsers(keyword),
  selectFilterUser: applySelectedFilterUser,
  clearFilterUser: clearSelectedFilterUser,
  selectAssignUser: applySelectedAssignUser,
  clearAssignUser: clearSelectedAssignUser
})

// Group options for filter (all groups)
const groupOptions = computed(() => [
  { value: '', label: t('admin.subscriptions.allGroups') },
  ...groups.value.map((g) => ({ value: g.id.toString(), label: g.name }))
])

const platformFilterOptions = computed(() =>
  buildAdminPlatformOptions(t, {
    allLabel: t('admin.subscriptions.allPlatforms')
  })
)

// Group options for assign (only subscription type groups)
const subscriptionGroupOptions = computed(() => buildSubscriptionGroupOptions(groups.value))
const {
  submitting,
  resettingQuota,
  handleAssignSubscription,
  handleExtend,
  handleExtendSubscription,
  handleRevoke,
  confirmRevoke,
  handleResetQuota,
  confirmResetQuota
} = useSubscriptionsViewActions({
  assignForm,
  extendForm,
  extendingSubscription,
  revokingSubscription,
  resettingSubscription,
  openExtendModal,
  openRevokeDialog,
  openResetQuotaConfirm,
  closeAssignModal,
  closeExtendModal,
  closeRevokeDialog,
  closeResetQuotaConfirm,
  reloadSubscriptions: loadSubscriptions,
  t,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError
})

onMounted(() => {
  initializeUserSearches()
  loadInitialData()
})

onUnmounted(() => {
  disposeUserSearches()
  disposeDataState()
})
</script>
