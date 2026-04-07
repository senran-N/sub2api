<template>
  <AppLayout>
    <TablePageLayout>
      <!-- Single Row: Search, Filters, and Actions -->
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <UserFilterFields
            :search-query="searchQuery"
            :filters="filters"
            :visible-filters="visibleFilters"
            :group-filter-options="groupFilterOptions"
            :active-attribute-filters="activeAttributeFilters"
            :get-attribute-definition="getAttributeDefinition"
            :get-attribute-definition-name="getAttributeDefinitionName"
            :update-attribute-filter="updateAttributeFilter"
            :apply-filter="applyFilter"
            @update:search-query="searchQuery = $event"
            @search-input="handleSearch"
          />
          <UserToolbarActions
            :loading="loading"
            :visible-filters="visibleFilters"
            :built-in-filters="builtInFilters"
            :filterable-attributes="filterableAttributes"
            :toggleable-columns="toggleableColumns"
            :is-column-visible="isColumnVisible"
            @refresh="loadUsers"
            @toggle-built-in-filter="toggleBuiltInFilter"
            @toggle-attribute-filter="toggleAttributeFilter"
            @toggle-column="toggleColumn"
            @open-attributes="openAttributesModal"
            @create="openCreateModal"
          />
        </div>
      </template>

      <!-- Users Table -->
      <template #table>
        <DataTable :columns="columns" :data="users" :loading="loading" :actions-count="7">
          <template #cell-email="{ value }">
            <UserEmailCell :email="value" />
          </template>

          <template #cell-username="{ value }">
            <UserUsernameCell :value="value" />
          </template>

          <template #cell-notes="{ value }">
            <UserNotesCell :notes="value" />
          </template>

          <!-- Dynamic attribute columns -->
          <template
            v-for="def in attributeDefinitions.filter(d => d.enabled)"
            :key="def.id"
            #[`cell-attr_${def.id}`]="{ row }"
          >
            <UserAttributeValueCell :value="getAttributeValue(row.id, def.id)" />
          </template>

          <template #cell-role="{ value }">
            <UserRoleCell :value="value" />
          </template>

          <template #cell-groups="{ row }">
            <UserGroupsCell
              :user="row"
              :has-groups-data="allGroups.length > 0"
              :expanded="expandedGroupUserId === row.id"
              :summary="getUserGroups(row)"
              @toggle-expanded="toggleExpandedGroup"
              @replace-group="openGroupReplace"
            />
          </template>

          <template #cell-subscriptions="{ row }">
            <UserSubscriptionsCell :user="row" />
          </template>

          <template #cell-balance="{ value, row }">
            <UserBalanceCell
              :user="{ ...row, balance: value }"
              @history="handleBalanceHistory"
              @deposit="handleDeposit"
            />
          </template>

          <template #cell-usage="{ row }">
            <UserUsageCell :usage="usageStats[row.id]" />
          </template>

          <template #cell-concurrency="{ row }">
            <UserConcurrencyCell
              :current="row.current_concurrency ?? 0"
              :max="row.concurrency"
            />
          </template>

          <template #cell-status="{ value }">
            <UserStatusCell :status="value" />
          </template>

          <template #cell-created_at="{ value }">
            <UserCreatedAtCell :value="value" />
          </template>

          <template #cell-actions="{ row }">
            <UserActionsCell
              :user="row"
              :menu-open="activeMenuId === row.id"
              @edit="handleEdit"
              @toggle-status="handleToggleStatus"
              @open-menu="openActionMenu"
            />
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.users.noUsersYet')"
              :description="t('admin.users.createFirstUser')"
              :action-text="t('admin.users.createUser')"
              @action="openCreateModal"
            />
          </template>
        </DataTable>
      </template>

      <!-- Pagination -->
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

    <UserActionMenu
      :user="activeMenuUser"
      :position="menuPosition"
      @close="closeActionMenu"
      @api-keys="handleViewApiKeys"
      @groups="handleAllowedGroups"
      @deposit="handleDeposit"
      @withdraw="handleWithdraw"
      @history="handleBalanceHistory"
      @delete="handleDelete"
    />

    <ConfirmDialog :show="showDeleteDialog" :title="t('admin.users.deleteUser')" :message="t('admin.users.deleteConfirm', { email: deletingUser?.email })" :danger="true" @confirm="confirmDelete" @cancel="closeDeleteDialog" />
    <UserCreateModal v-if="showCreateModal" :show="showCreateModal" @close="closeCreateModal" @success="loadUsers" />
    <UserEditModal v-if="showEditModal" :show="showEditModal" :user="editingUser" @close="closeEditModal" @success="loadUsers" />
    <UserApiKeysModal v-if="showApiKeysModal" :show="showApiKeysModal" :user="viewingUser" @close="closeApiKeysModal" />
    <UserAllowedGroupsModal v-if="showAllowedGroupsModal" :show="showAllowedGroupsModal" :user="allowedGroupsUser" @close="closeAllowedGroupsModal" @success="loadUsers" />
    <UserBalanceModal v-if="showBalanceModal" :show="showBalanceModal" :user="balanceUser" :operation="balanceOperation" @close="closeBalanceModal" @success="loadUsers" />
    <UserBalanceHistoryModal v-if="showBalanceHistoryModal" :show="showBalanceHistoryModal" :user="balanceHistoryUser" @close="closeBalanceHistoryModal" @deposit="handleDepositFromHistory" @withdraw="handleWithdrawFromHistory" />
    <GroupReplaceModal v-if="showGroupReplaceModal" :show="showGroupReplaceModal" :user="groupReplaceUser" :old-group="groupReplaceOldGroup" :all-groups="allGroups" @close="closeGroupReplaceModal" @success="loadUsers" />
    <UserAttributesConfigModal v-if="showAttributesModal" :show="showAttributesModal" @close="handleAttributesModalClose" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()
import type { AdminUser, UserAttributeDefinition } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import UserConcurrencyCell from '@/components/user/UserConcurrencyCell.vue'
import { useUsersViewAdminActions } from './users/useUsersViewAdminActions'
import { useUsersViewData } from './users/useUsersViewData'
import { useUsersViewDialogs } from './users/useUsersViewDialogs'
import { useUsersViewState } from './users/useUsersViewState'
import UserActionMenu from './users/UserActionMenu.vue'
import UserActionsCell from './users/UserActionsCell.vue'
import UserBalanceCell from './users/UserBalanceCell.vue'
import UserCreatedAtCell from './users/UserCreatedAtCell.vue'
import UserAttributeValueCell from './users/UserAttributeValueCell.vue'
import UserEmailCell from './users/UserEmailCell.vue'
import UserFilterFields from './users/UserFilterFields.vue'
import UserGroupsCell from './users/UserGroupsCell.vue'
import UserNotesCell from './users/UserNotesCell.vue'
import UserRoleCell from './users/UserRoleCell.vue'
import UserStatusCell from './users/UserStatusCell.vue'
import UserSubscriptionsCell from './users/UserSubscriptionsCell.vue'
import UserToolbarActions from './users/UserToolbarActions.vue'
import UserUsernameCell from './users/UserUsernameCell.vue'
import UserUsageCell from './users/UserUsageCell.vue'
import {
  buildUserAttributeColumns,
  type BuiltInUserFilterKey,
  buildUserGroupFilterOptions,
  buildUserTableColumns,
  filterVisibleUserColumns,
  formatUserAttributeValue,
  getAttributeDefinitionName as getUserAttributeDefinitionName,
  getUserGroupsSummary
} from './users/usersTable'

const UserAttributesConfigModal = defineAsyncComponent(() => import('@/components/user/UserAttributesConfigModal.vue'))
const UserCreateModal = defineAsyncComponent(() => import('@/components/admin/user/UserCreateModal.vue'))
const UserEditModal = defineAsyncComponent(() => import('@/components/admin/user/UserEditModal.vue'))
const UserApiKeysModal = defineAsyncComponent(() => import('@/components/admin/user/UserApiKeysModal.vue'))
const UserAllowedGroupsModal = defineAsyncComponent(() => import('@/components/admin/user/UserAllowedGroupsModal.vue'))
const UserBalanceModal = defineAsyncComponent(() => import('@/components/admin/user/UserBalanceModal.vue'))
const UserBalanceHistoryModal = defineAsyncComponent(() => import('@/components/admin/user/UserBalanceHistoryModal.vue'))
const GroupReplaceModal = defineAsyncComponent(() => import('@/components/admin/user/GroupReplaceModal.vue'))

const appStore = useAppStore()
const usersViewDialogs = useUsersViewDialogs()
let usersViewData: ReturnType<typeof useUsersViewData> | null = null

const requireUsersViewData = (): ReturnType<typeof useUsersViewData> => {
  if (!usersViewData) {
    throw new Error('Users view data is not initialized')
  }

  return usersViewData
}

const loadUsers = () => requireUsersViewData().loadUsers()
const loadAllGroups = () => requireUsersViewData().loadAllGroups()
const loadAttributeDefinitions = () => requireUsersViewData().loadAttributeDefinitions()
const loadUsersSecondaryData = (
  userIds: number[],
  signal?: AbortSignal,
  expectedSeq?: number
) => requireUsersViewData().loadUsersSecondaryData(userIds, signal, expectedSeq)
const resetUsersSecondaryData = () => requireUsersViewData().resetSecondaryData()
const disposeUsersViewData = () => requireUsersViewData().dispose()

const usersViewState = useUsersViewState({
  attributeDefinitions: computed(() => requireUsersViewData().attributeDefinitions.value),
  initialPageSize: getPersistedPageSize(),
  loadUsers,
  loadGroups: loadAllGroups,
  loadSecondaryData: loadUsersSecondaryData,
  resetSecondaryData: resetUsersSecondaryData
})

const {
  activeMenuId,
  menuPosition,
  showApiKeysModal,
  viewingUser,
  showAllowedGroupsModal,
  allowedGroupsUser,
  expandedGroupUserId,
  showGroupReplaceModal,
  groupReplaceUser,
  groupReplaceOldGroup,
  showBalanceModal,
  balanceUser,
  balanceOperation,
  showBalanceHistoryModal,
  balanceHistoryUser,
  openActionMenu,
  closeActionMenu,
  openViewApiKeys,
  closeApiKeysModal,
  openAllowedGroups,
  closeAllowedGroupsModal,
  toggleExpandedGroup,
  closeExpandedGroup,
  openGroupReplace: openGroupReplaceDialog,
  closeGroupReplaceModal,
  openBalanceModal,
  closeBalanceModal,
  openBalanceHistory,
  closeBalanceHistoryModal,
  reopenBalanceFromHistory
} = usersViewDialogs

const {
  hiddenColumns,
  filters,
  activeAttributeFilters,
  visibleFilters,
  searchQuery,
  pagination,
  hasVisibleUsageColumn,
  hasVisibleSubscriptionsColumn,
  hasVisibleAttributeColumns,
  initializePersistedState,
  isColumnVisible,
  toggleColumn,
  handleSearch,
  handlePageChange,
  handlePageSizeChange,
  toggleBuiltInFilter,
  toggleAttributeFilter: toggleAttributeFilterById,
  updateAttributeFilter,
  applyFilter,
  setCurrentUserIds,
  resetSecondaryDataState,
  scheduleUsersSecondaryDataLoad,
  isSecondaryDataRequestCurrent,
  dispose: disposeUsersViewState
} = usersViewState

usersViewData = useUsersViewData({
  t,
  showError: appStore.showError,
  filters,
  activeAttributeFilters,
  searchQuery,
  pagination,
  hasVisibleUsageColumn,
  hasVisibleSubscriptionsColumn,
  hasVisibleAttributeColumns,
  isSecondaryDataRequestCurrent,
  setCurrentUserIds,
  resetSecondaryDataState,
  scheduleUsersSecondaryDataLoad
})

const {
  usageStats,
  attributeDefinitions,
  userAttributeValues,
  users,
  loading,
  allGroups
} = requireUsersViewData()

const usersViewAdminActions = useUsersViewAdminActions({
  reloadUsers: loadUsers,
  reloadAttributeDefinitions: loadAttributeDefinitions,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError,
  t
})

const {
  showCreateModal,
  showEditModal,
  showDeleteDialog,
  showAttributesModal,
  editingUser,
  deletingUser,
  openCreateModal,
  closeCreateModal,
  handleEdit,
  closeEditModal,
  openAttributesModal,
  handleAttributesModalClose,
  handleToggleStatus,
  handleDelete,
  closeDeleteDialog,
  confirmDelete
} = usersViewAdminActions

// Generate dynamic attribute columns from enabled definitions
const attributeColumns = computed<Column[]>(() =>
  buildUserAttributeColumns(attributeDefinitions.value)
)

// Get formatted attribute value for display in table
const getAttributeValue = (userId: number, attrId: number): string => {
  return formatUserAttributeValue(
    userAttributeValues.value,
    attributeDefinitions.value,
    userId,
    attrId
  )
}

// All possible columns (for column settings)
const allColumns = computed<Column[]>(() =>
  buildUserTableColumns(attributeColumns.value, {
    user: t('admin.users.columns.user'),
    id: 'ID',
    username: t('admin.users.columns.username'),
    notes: t('admin.users.columns.notes'),
    role: t('admin.users.columns.role'),
    groups: t('admin.users.columns.groups'),
    subscriptions: t('admin.users.columns.subscriptions'),
    balance: t('admin.users.columns.balance'),
    usage: t('admin.users.columns.usage'),
    concurrency: t('admin.users.columns.concurrency'),
    status: t('admin.users.columns.status'),
    created: t('admin.users.columns.created'),
    actions: t('admin.users.columns.actions')
  })
)

// Columns that can be toggled (exclude email and actions which are always visible)
const toggleableColumns = computed(() =>
  allColumns.value.filter(col => col.key !== 'email' && col.key !== 'actions')
)

// Filtered columns based on visibility
const columns = computed<Column[]>(() =>
  filterVisibleUserColumns(allColumns.value, hiddenColumns)
)
// Resolve user's accessible groups: exclusive groups first, then public groups
const getUserGroups = (user: AdminUser) => getUserGroupsSummary(allGroups.value, user)
const activeMenuUser = computed(() => {
  if (activeMenuId.value === null) {
    return null
  }

  return users.value.find((user) => user.id === activeMenuId.value) ?? null
})

// Group filter options: "All Groups" + active exclusive groups (value = group name for fuzzy match)
const groupFilterOptions = computed(() =>
  buildUserGroupFilterOptions(allGroups.value, t('admin.users.allGroups'))
)

// All filterable attribute definitions (enabled attributes)
const filterableAttributes = computed(() =>
  attributeDefinitions.value.filter(def => def.enabled)
)

// Built-in filter definitions
const builtInFilters = computed<{ key: BuiltInUserFilterKey; name: string; type: 'select' }[]>(() => [
  { key: 'role', name: t('admin.users.columns.role'), type: 'select' as const },
  { key: 'status', name: t('admin.users.columns.status'), type: 'select' as const },
  { key: 'group', name: t('admin.users.columns.groups'), type: 'select' as const }
])

// Get attribute definition by ID
const getAttributeDefinition = (attrId: number): UserAttributeDefinition | undefined => {
  return attributeDefinitions.value.find(d => d.id === attrId)
}
const getAttributeDefinitionName = (attrId: number): string =>
  getUserAttributeDefinitionName(attributeDefinitions.value, attrId)

// Close menu when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.action-menu-trigger') && !target.closest('.action-menu-content')) {
    closeActionMenu()
  }
  // Close expanded group dropdown when clicking outside
  if (expandedGroupUserId.value !== null) {
    closeExpandedGroup()
  }
}

const toggleAttributeFilter = (attr: UserAttributeDefinition) => {
  toggleAttributeFilterById(attr.id)
}

const handleViewApiKeys = (user: AdminUser) => {
  openViewApiKeys(user)
}

const handleAllowedGroups = (user: AdminUser) => {
  openAllowedGroups(user)
}

const openGroupReplace = (user: AdminUser, group: { id: number; name: string }) => {
  openGroupReplaceDialog(user, group)
}

const handleDeposit = (user: AdminUser) => {
  openBalanceModal(user, 'add')
}

const handleWithdraw = (user: AdminUser) => {
  openBalanceModal(user, 'subtract')
}

const handleBalanceHistory = (user: AdminUser) => {
  openBalanceHistory(user)
}

// Handle deposit from balance history modal
const handleDepositFromHistory = () => {
  reopenBalanceFromHistory('add')
}

// Handle withdraw from balance history modal
const handleWithdrawFromHistory = () => {
  reopenBalanceFromHistory('subtract')
}

// 滚动时关闭菜单
const handleScroll = () => {
  closeActionMenu()
}

onMounted(async () => {
  await loadAttributeDefinitions()
  initializePersistedState()
  loadUsers()
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('scroll', handleScroll, true)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', handleScroll, true)
  disposeUsersViewState()
  disposeUsersViewData()
})
</script>
