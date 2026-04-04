<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <GroupFilterFields
            v-model:searchQuery="searchQuery"
            v-model:platform="filters.platform"
            v-model:status="filters.status"
            v-model:isExclusive="filters.is_exclusive"
            :platform-options="platformFilterOptions"
            :status-options="statusOptions"
            :exclusive-options="exclusiveOptions"
            @search-input="handleSearch"
            @platform-change="loadGroups"
            @status-change="loadGroups"
            @exclusive-change="loadGroups"
          />

          <GroupActionToolbar
            :loading="loading"
            @refresh="loadGroups"
            @sort-order="openSortModal"
            @create="showCreateModal = true"
          />
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="groups" :loading="loading">
          <template #cell-name="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
          </template>

          <template #cell-platform="{ value }">
            <GroupPlatformBadge :platform="value" />
          </template>

          <template #cell-billing_type="{ row }">
            <GroupBillingTypeCell :group="row" />
          </template>

          <template #cell-rate_multiplier="{ value }">
            <GroupRateMultiplierCell :rate-multiplier="value" />
          </template>

          <template #cell-is_exclusive="{ value }">
            <GroupExclusivityBadge :exclusive="value" />
          </template>

          <template #cell-account_count="{ row }">
            <GroupAccountCountCell :group="row" />
          </template>

          <template #cell-capacity="{ row }">
            <GroupCapacityCell :capacity="capacityMap.get(row.id)" />
          </template>

          <template #cell-usage="{ row }">
            <GroupUsageCell
              :loading="usageLoading"
              :summary="usageMap.get(row.id)"
            />
          </template>

          <template #cell-status="{ value }">
            <GroupStatusBadge :status="value" />
          </template>

          <template #cell-actions="{ row }">
            <GroupActionsCell
              :group="row"
              @edit="handleEdit"
              @rate-multipliers="handleRateMultipliers"
              @delete="handleDelete"
            />
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.groups.noGroupsYet')"
              :description="t('admin.groups.createFirstGroup')"
              :action-text="t('admin.groups.createGroup')"
              @action="showCreateModal = true"
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

    <GroupCreateDialog
      :show="showCreateModal"
      :submitting="submitting"
      :form="createForm"
      :platform-options="platformOptions"
      :copy-accounts-group-options="copyAccountsGroupOptions"
      :subscription-type-options="subscriptionTypeOptions"
      :fallback-group-options="fallbackGroupOptions"
      :invalid-request-fallback-options="invalidRequestFallbackOptions"
      :rules="createModelRoutingRules"
      :account-search-keyword="createAccountSearchKeyword"
      :account-search-results="createAccountSearchResults"
      :show-account-dropdown="createShowAccountDropdown"
      :get-rule-render-key="getCreateRuleRenderKey"
      :get-rule-search-key="getCreateRuleSearchKey"
      :search-accounts-by-rule="searchCreateAccountsByRule"
      :select-account="selectCreateAccount"
      :remove-selected-account="removeCreateSelectedAccount"
      :on-account-search-focus="onCreateAccountSearchFocus"
      :add-routing-rule="addCreateRoutingRule"
      :remove-routing-rule="removeCreateRoutingRule"
      @close="closeCreateModal"
      @submit="handleCreateGroup"
      @add-copy-group="handleCreateCopyAccountsGroupAdd"
      @remove-copy-group="handleCreateCopyAccountsGroupRemove"
      @toggle-scope="toggleCreateScope"
    />

    <GroupEditDialog
      :show="showEditModal"
      :submitting="submitting"
      :editing-group="editingGroup"
      :form="editForm"
      :platform-options="platformOptions"
      :copy-accounts-group-options="copyAccountsGroupOptionsForEdit"
      :edit-status-options="editStatusOptions"
      :subscription-type-options="subscriptionTypeOptions"
      :fallback-group-options="fallbackGroupOptionsForEdit"
      :invalid-request-fallback-options="invalidRequestFallbackOptionsForEdit"
      :rules="editModelRoutingRules"
      :account-search-keyword="editAccountSearchKeyword"
      :account-search-results="editAccountSearchResults"
      :show-account-dropdown="editShowAccountDropdown"
      :get-rule-render-key="getEditRuleRenderKey"
      :get-rule-search-key="getEditRuleSearchKey"
      :search-accounts-by-rule="searchEditAccountsByRule"
      :select-account="selectEditAccount"
      :remove-selected-account="removeEditSelectedAccount"
      :on-account-search-focus="onEditAccountSearchFocus"
      :add-routing-rule="addEditRoutingRule"
      :remove-routing-rule="removeEditRoutingRule"
      @close="closeEditModal"
      @submit="handleUpdateGroup"
      @add-copy-group="handleEditCopyAccountsGroupAdd"
      @remove-copy-group="handleEditCopyAccountsGroupRemove"
      @toggle-scope="toggleEditScope"
    />

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.groups.deleteGroup')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <GroupSortOrderDialog
      :show="showSortModal"
      :groups="sortableGroups"
      :submitting="sortSubmitting"
      @update:groups="sortableGroups = $event"
      @close="closeSortModal"
      @save="saveSortOrder"
    />

    <!-- Group Rate Multipliers Modal -->
    <GroupRateMultipliersModal
      :show="showRateMultipliersModal"
      :group="rateMultipliersGroup"
      @close="showRateMultipliersModal = false"
      @success="loadGroups"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useOnboardingStore } from '@/stores/onboarding'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import GroupRateMultipliersModal from '@/components/admin/group/GroupRateMultipliersModal.vue'
import {
  addCopyAccountsGroupSelection,
  buildCopyAccountsGroupOptions,
  buildFallbackGroupOptions,
  buildInvalidRequestFallbackOptions,
  removeCopyAccountsGroupSelection
} from './groupsForm'
import GroupAccountCountCell from './groups/GroupAccountCountCell.vue'
import GroupActionsCell from './groups/GroupActionsCell.vue'
import GroupActionToolbar from './groups/GroupActionToolbar.vue'
import GroupBillingTypeCell from './groups/GroupBillingTypeCell.vue'
import GroupCapacityCell from './groups/GroupCapacityCell.vue'
import GroupCreateDialog from './groups/GroupCreateDialog.vue'
import GroupEditDialog from './groups/GroupEditDialog.vue'
import GroupExclusivityBadge from './groups/GroupExclusivityBadge.vue'
import GroupFilterFields from './groups/GroupFilterFields.vue'
import GroupPlatformBadge from './groups/GroupPlatformBadge.vue'
import GroupRateMultiplierCell from './groups/GroupRateMultiplierCell.vue'
import GroupSortOrderDialog from './groups/GroupSortOrderDialog.vue'
import GroupStatusBadge from './groups/GroupStatusBadge.vue'
import GroupUsageCell from './groups/GroupUsageCell.vue'
import { useGroupsViewData } from './useGroupsViewData'
import { useGroupsViewManagement } from './useGroupsViewManagement'

const { t } = useI18n()
const appStore = useAppStore()
const onboardingStore = useOnboardingStore()

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.groups.columns.name'), sortable: true },
  { key: 'platform', label: t('admin.groups.columns.platform'), sortable: true },
  { key: 'billing_type', label: t('admin.groups.columns.billingType'), sortable: true },
  { key: 'rate_multiplier', label: t('admin.groups.columns.rateMultiplier'), sortable: true },
  { key: 'is_exclusive', label: t('admin.groups.columns.type'), sortable: true },
  { key: 'account_count', label: t('admin.groups.columns.accounts'), sortable: true },
  { key: 'capacity', label: t('admin.groups.columns.capacity'), sortable: false },
  { key: 'usage', label: t('admin.groups.columns.usage'), sortable: false },
  { key: 'status', label: t('admin.groups.columns.status'), sortable: true },
  { key: 'actions', label: t('admin.groups.columns.actions'), sortable: false }
])

// Filter options
const statusOptions = computed(() => [
  { value: '', label: t('admin.groups.allStatus') },
  { value: 'active', label: t('admin.accounts.status.active') },
  { value: 'inactive', label: t('admin.accounts.status.inactive') }
])

const exclusiveOptions = computed(() => [
  { value: '', label: t('admin.groups.allGroups') },
  { value: 'true', label: t('admin.groups.exclusive') },
  { value: 'false', label: t('admin.groups.nonExclusive') }
])

const platformOptions = computed(() => [
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' },
  { value: 'sora', label: 'Sora' }
])

const platformFilterOptions = computed(() => [
  { value: '', label: t('admin.groups.allPlatforms') },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' },
  { value: 'sora', label: 'Sora' }
])

const editStatusOptions = computed(() => [
  { value: 'active', label: t('admin.accounts.status.active') },
  { value: 'inactive', label: t('admin.accounts.status.inactive') }
])

const subscriptionTypeOptions = computed(() => [
  { value: 'standard', label: t('admin.groups.subscription.standard') },
  { value: 'subscription', label: t('admin.groups.subscription.subscription') }
])

// 降级分组选项（创建时）- 仅包含 anthropic 平台且未启用 claude_code_only 的分组
const fallbackGroupOptions = computed(() => {
  return buildFallbackGroupOptions(
    groups.value,
    t('admin.groups.claudeCode.noFallback')
  )
})

// 降级分组选项（编辑时）- 排除自身
const fallbackGroupOptionsForEdit = computed(() => {
  return buildFallbackGroupOptions(
    groups.value,
    t('admin.groups.claudeCode.noFallback'),
    editingGroup.value?.id
  )
})

// 无效请求兜底分组选项（创建时）- 仅包含 anthropic 平台、非订阅且未配置兜底的分组
const invalidRequestFallbackOptions = computed(() => {
  return buildInvalidRequestFallbackOptions(
    groups.value,
    t('admin.groups.invalidRequestFallback.noFallback')
  )
})

// 无效请求兜底分组选项（编辑时）- 排除自身
const invalidRequestFallbackOptionsForEdit = computed(() => {
  return buildInvalidRequestFallbackOptions(
    groups.value,
    t('admin.groups.invalidRequestFallback.noFallback'),
    editingGroup.value?.id
  )
})

// 复制账号的源分组选项（创建时）- 仅包含相同平台且有账号的分组
const copyAccountsGroupOptions = computed(() => {
  return buildCopyAccountsGroupOptions(groups.value, createForm.platform)
})

// 复制账号的源分组选项（编辑时）- 仅包含相同平台且有账号的分组，排除自身
const copyAccountsGroupOptionsForEdit = computed(() => {
  return buildCopyAccountsGroupOptions(
    groups.value,
    editForm.platform,
    editingGroup.value?.id
  )
})

const {
  groups,
  loading,
  usageMap,
  usageLoading,
  capacityMap,
  searchQuery,
  filters,
  pagination,
  showSortModal,
  sortSubmitting,
  sortableGroups,
  loadGroups,
  handleSearch,
  handlePageChange,
  handlePageSizeChange,
  openSortModal,
  closeSortModal,
  saveSortOrder,
  dispose: disposeGroupsViewData
} = useGroupsViewData({
  t,
  showError: appStore.showError,
  showSuccess: appStore.showSuccess
})

const {
  showCreateModal,
  showEditModal,
  showDeleteDialog,
  showRateMultipliersModal,
  submitting,
  editingGroup,
  rateMultipliersGroup,
  createForm,
  editForm,
  createModelRoutingRules,
  createAccountSearchKeyword,
  createAccountSearchResults,
  createShowAccountDropdown,
  getCreateRuleRenderKey,
  getCreateRuleSearchKey,
  searchCreateAccountsByRule,
  selectCreateAccount,
  removeCreateSelectedAccount,
  onCreateAccountSearchFocus,
  addCreateRoutingRule,
  removeCreateRoutingRule,
  editModelRoutingRules,
  editAccountSearchKeyword,
  editAccountSearchResults,
  editShowAccountDropdown,
  getEditRuleRenderKey,
  getEditRuleSearchKey,
  searchEditAccountsByRule,
  selectEditAccount,
  removeEditSelectedAccount,
  onEditAccountSearchFocus,
  addEditRoutingRule,
  removeEditRoutingRule,
  deleteConfirmMessage,
  toggleCreateScope,
  toggleEditScope,
  closeCreateModal,
  handleCreateGroup,
  handleEdit,
  closeEditModal,
  handleUpdateGroup,
  handleRateMultipliers,
  handleDelete,
  confirmDelete,
  handleClickOutside
} = useGroupsViewManagement({
  t,
  showError: appStore.showError,
  showSuccess: appStore.showSuccess,
  loadGroups,
  isCurrentOnboardingStep: onboardingStore.isCurrentStep,
  advanceOnboarding: onboardingStore.nextStep
})

function handleCreateCopyAccountsGroupAdd(groupId: number): void {
  addCopyAccountsGroupSelection(createForm.copy_accounts_from_group_ids, groupId)
}

function handleCreateCopyAccountsGroupRemove(groupId: number): void {
  removeCopyAccountsGroupSelection(createForm.copy_accounts_from_group_ids, groupId)
}

function handleEditCopyAccountsGroupAdd(groupId: number): void {
  addCopyAccountsGroupSelection(editForm.copy_accounts_from_group_ids, groupId)
}

function handleEditCopyAccountsGroupRemove(groupId: number): void {
  removeCopyAccountsGroupSelection(editForm.copy_accounts_from_group_ids, groupId)
}

onMounted(() => {
  void loadGroups()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  disposeGroupsViewData()
})
</script>
