<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <KeysFiltersBar
          :search="filterSearch"
          :group-id="filterGroupId"
          :status="filterStatus"
          :group-options="groupFilterOptions"
          :status-options="statusFilterOptions"
          :public-settings="publicSettings"
          @update:search="filterSearch = $event"
          @update:group-id="onGroupFilterChange"
          @update:status="onStatusFilterChange"
          @apply="onFilterChange"
        />
      </template>

      <template #actions>
        <KeysActionsBar
          :loading="loading"
          @refresh="loadApiKeys"
          @create="showCreateModal = true"
        />
      </template>

      <template #table>
        <DataTable :columns="columns" :data="apiKeys" :loading="loading">
          <template #cell-key="{ value, row }">
            <KeysKeyCell
              :value="value"
              :row-id="row.id"
              :copied-key-id="copiedKeyId"
              @copy="copyToClipboard"
            />
          </template>

          <template #cell-name="{ value, row }">
            <KeysNameCell :value="value" :row="row" />
          </template>

          <template #cell-group="{ row }">
            <KeysGroupCell
              :row="row"
              :user-group-rates="userGroupRates"
              :click-to-change-title="t('keys.clickToChangeGroup')"
              :no-group-label="t('keys.noGroup')"
              :select-group-label="t('keys.selectGroup')"
              :button-ref="(el) => setGroupButtonRef(row.id, el)"
              @open-selector="openGroupSelector"
            />
          </template>

          <template #cell-usage="{ row }">
            <KeysUsageCell :row="row" :stats="usageStats[row.id]" />
          </template>

          <template #cell-rate_limit="{ row }">
            <KeysRateLimitCell
              :row="row"
              :format-reset-time="formatResetTime"
              @reset="confirmResetRateLimitFromTable"
            />
          </template>

          <template #cell-expires_at="{ value }">
            <KeysExpirationCell :value="value" />
          </template>

          <template #cell-status="{ value }">
            <KeysStatusBadge :status="value" />
          </template>

          <template #cell-last_used_at="{ value }">
            <span v-if="value" class="theme-text-muted text-sm">
              {{ formatDateTime(value) }}
            </span>
            <span v-else class="theme-text-subtle text-sm">-</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="theme-text-muted text-sm">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <KeysRowActions
              :row="row"
              :button-ref="(el) => setMoreMenuRef(row.id, el)"
              @use="openUseKeyModal"
              @edit="editKey"
              @toggle-more="toggleMoreMenu"
            />
          </template>

          <template #empty>
            <EmptyState
              :title="t('keys.noKeysYet')"
              :description="t('keys.createFirstKey')"
              :action-text="t('keys.createKey')"
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

    <KeysDialogs
      :show-create-modal="showCreateModal"
      :show-edit-modal="showEditModal"
      :show-delete-dialog="showDeleteDialog"
      :show-reset-quota-dialog="showResetQuotaDialog"
      :show-reset-rate-limit-dialog="showResetRateLimitDialog"
      :show-use-key-modal="showUseKeyModal"
      :show-ccs-client-select="showCcsClientSelect"
      :create-title="t('keys.createKey')"
      :edit-title="t('keys.editKey')"
      :delete-title="t('keys.deleteKey')"
      :delete-message="t('keys.deleteConfirmMessage', { name: selectedKey?.name })"
      :delete-confirm-text="t('common.delete')"
      :reset-quota-title="t('keys.resetQuotaTitle')"
      :reset-quota-message="t('keys.resetQuotaConfirmMessage', { name: selectedKey?.name, used: selectedKey?.quota_used?.toFixed(4) })"
      :reset-rate-limit-title="t('keys.resetRateLimitTitle')"
      :reset-rate-limit-message="t('keys.resetRateLimitConfirmMessage', { name: selectedKey?.name })"
      :reset-text="t('keys.reset')"
      :cancel-text="t('common.cancel')"
      :ccs-client-select-title="t('keys.ccsClientSelect.title')"
      :ccs-client-select-description="t('keys.ccsClientSelect.description')"
      :claude-label="t('keys.ccsClientSelect.claudeCode')"
      :claude-description="t('keys.ccsClientSelect.claudeCodeDesc')"
      :gemini-label="t('keys.ccsClientSelect.geminiCli')"
      :gemini-description="t('keys.ccsClientSelect.geminiCliDesc')"
      :form-data="formData"
      :group-options="groupOptions"
      :status-options="statusOptions"
      :custom-key-error="customKeyError"
      :selected-key="selectedKey"
      :submitting="submitting"
      :public-settings="publicSettings"
      @close-modals="closeModals"
      @submit="handleSubmit"
      @confirm-reset-quota="confirmResetQuota"
      @confirm-reset-rate-limit="confirmResetRateLimit"
      @reset-quota="resetQuotaUsed"
      @reset-rate-limit="resetRateLimitUsage"
      @set-expiration-days="setExpirationDays"
      @delete="handleDelete"
      @update:showDeleteDialog="showDeleteDialog = $event"
      @update:showResetQuotaDialog="showResetQuotaDialog = $event"
      @update:showResetRateLimitDialog="showResetRateLimitDialog = $event"
      @close-use-key-modal="closeUseKeyModal"
      @close-ccs-client-select="closeCcsClientSelect"
      @select-ccs-client="handleCcsClientSelect"
    />

    <KeysOverlayMenus
      :show-more-menu="moreMenuKeyId !== null && moreMenuPosition !== null"
      :more-menu-position="moreMenuPosition"
      :more-menu-row="moreMenuRow"
      :hide-ccs-import-button="Boolean(publicSettings?.hide_ccs_import_button)"
      :show-group-selector="groupSelectorKeyId !== null && dropdownPosition !== null"
      :dropdown-position="dropdownPosition"
      :group-search-query="groupSearchQuery"
      :filtered-group-options="filteredGroupOptions"
      :selected-key-for-group="selectedKeyForGroup"
      @close-more-menu="closeMoreMenu"
      @import-to-ccswitch="importToCcswitch"
      @toggle-key-status="toggleKeyStatus"
      @confirm-delete="confirmDelete"
      @close-group-selector="closeGroupSelector"
      @update:groupSearchQuery="groupSearchQuery = $event"
      @change-group="changeGroup"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import KeysFiltersBar from './keys/KeysFiltersBar.vue'
import KeysActionsBar from './keys/KeysActionsBar.vue'
import KeysDialogs from './keys/KeysDialogs.vue'
import KeysExpirationCell from './keys/KeysExpirationCell.vue'
import KeysGroupCell from './keys/KeysGroupCell.vue'
import KeysKeyCell from './keys/KeysKeyCell.vue'
import KeysNameCell from './keys/KeysNameCell.vue'
import KeysOverlayMenus from './keys/KeysOverlayMenus.vue'
import KeysUsageCell from './keys/KeysUsageCell.vue'
import KeysRateLimitCell from './keys/KeysRateLimitCell.vue'
import KeysStatusBadge from './keys/KeysStatusBadge.vue'
import KeysRowActions from './keys/KeysRowActions.vue'
import { formatDateTime } from '@/utils/format'
import { useKeysViewModel } from './keys/useKeysViewModel'

const {
  columns,
  apiKeys,
  loading,
  submitting,
  usageStats,
  userGroupRates,
  pagination,
  filterSearch,
  filterStatus,
  filterGroupId,
  showCreateModal,
  showEditModal,
  showDeleteDialog,
  showResetQuotaDialog,
  showResetRateLimitDialog,
  showUseKeyModal,
  showCcsClientSelect,
  selectedKey,
  copiedKeyId,
  groupSelectorKeyId,
  publicSettings,
  dropdownPosition,
  moreMenuKeyId,
  moreMenuPosition,
  moreMenuRow,
  selectedKeyForGroup,
  formData,
  customKeyError,
  statusOptions,
  groupFilterOptions,
  statusFilterOptions,
  groupOptions,
  groupSearchQuery,
  filteredGroupOptions,
  loadApiKeys,
  onFilterChange,
  onGroupFilterChange,
  onStatusFilterChange,
  copyToClipboard,
  setMoreMenuRef,
  toggleMoreMenu,
  closeMoreMenu,
  setGroupButtonRef,
  openGroupSelector,
  closeGroupSelector,
  openUseKeyModal,
  closeUseKeyModal,
  handlePageChange,
  handlePageSizeChange,
  editKey,
  toggleKeyStatus,
  changeGroup,
  confirmDelete,
  handleSubmit,
  handleDelete,
  closeModals,
  confirmResetQuota,
  setExpirationDays,
  resetQuotaUsed,
  confirmResetRateLimit,
  confirmResetRateLimitFromTable,
  resetRateLimitUsage,
  importToCcswitch,
  handleCcsClientSelect,
  closeCcsClientSelect,
  formatResetTime
} = useKeysViewModel()
</script>
