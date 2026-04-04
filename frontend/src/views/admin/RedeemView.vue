<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <RedeemToolbar
          :search-query="searchQuery"
          :type="filters.type"
          :status="filters.status"
          :type-options="filterTypeOptions"
          :status-options="filterStatusOptions"
          :loading="loading"
          @update:search-query="searchQuery = $event"
          @update:type="filters.type = $event"
          @update:status="filters.status = $event"
          @search="handleSearch"
          @refresh="loadCodes"
          @export="handleExportCodes"
          @generate="showGenerateDialog = true"
        />
      </template>

      <template #table>
        <DataTable :columns="columns" :data="codes" :loading="loading">
          <template #cell-code="{ value }">
            <RedeemCodeCell
              :code="value"
              :copied="copiedCode === value"
              @copy="copyCodeToClipboard(value)"
            />
          </template>

          <template #cell-type="{ value }">
            <RedeemTypeBadge :type="value" />
          </template>

          <template #cell-value="{ row }">
            <RedeemValueCell :code="row" />
          </template>

          <template #cell-status="{ value }">
            <RedeemStatusBadge :status="value" />
          </template>

          <template #cell-used_by="{ value, row }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ row.user?.email || (value ? t('admin.redeem.userPrefix', { id: value }) : '-') }}
            </span>
          </template>

          <template #cell-used_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{
              value ? formatDateTime(value) : '-'
            }}</span>
          </template>

          <template #cell-actions="{ row }">
            <RedeemActionsCell
              :show-delete="row.status === 'unused'"
              @delete="handleDelete(row)"
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

        <div v-if="filters.status === 'unused'" class="flex justify-end">
          <button @click="showDeleteUnusedDialog = true" class="btn btn-danger">
            {{ t('admin.redeem.deleteAllUnused') }}
          </button>
        </div>
      </template>
    </TablePageLayout>

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.redeem.deleteCode')"
      :message="t('admin.redeem.deleteCodeConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <ConfirmDialog
      :show="showDeleteUnusedDialog"
      :title="t('admin.redeem.deleteAllUnused')"
      :message="t('admin.redeem.deleteAllUnusedConfirm')"
      :confirm-text="t('admin.redeem.deleteAll')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDeleteUnused"
      @cancel="showDeleteUnusedDialog = false"
    />

    <RedeemGenerateDialog
      :show="showGenerateDialog"
      :form="generateForm"
      :type-options="typeOptions"
      :subscription-group-options="subscriptionGroupOptions"
      :submitting="generating"
      @close="showGenerateDialog = false"
      @submit="handleGenerateCodes"
    />

    <RedeemGeneratedResultDialog
      :show="showResultDialog"
      :count="generatedCodes.length"
      :codes-text="generatedCodesText"
      :textarea-height="textareaHeight"
      :copied-all="copiedAll"
      @close="closeResultDialog"
      @copy="copyGeneratedCodes"
      @download="downloadGeneratedCodes"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateTime } from '@/utils/format'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import RedeemActionsCell from './redeem/RedeemActionsCell.vue'
import RedeemCodeCell from './redeem/RedeemCodeCell.vue'
import RedeemGeneratedResultDialog from './redeem/RedeemGeneratedResultDialog.vue'
import RedeemGenerateDialog from './redeem/RedeemGenerateDialog.vue'
import RedeemStatusBadge from './redeem/RedeemStatusBadge.vue'
import RedeemToolbar from './redeem/RedeemToolbar.vue'
import RedeemTypeBadge from './redeem/RedeemTypeBadge.vue'
import RedeemValueCell from './redeem/RedeemValueCell.vue'
import { useRedeemGeneration } from './useRedeemGeneration'
import { useRedeemViewData } from './useRedeemViewData'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

const columns = computed<Column[]>(() => [
  { key: 'code', label: t('admin.redeem.columns.code') },
  { key: 'type', label: t('admin.redeem.columns.type'), sortable: true },
  { key: 'value', label: t('admin.redeem.columns.value'), sortable: true },
  { key: 'status', label: t('admin.redeem.columns.status'), sortable: true },
  { key: 'used_by', label: t('admin.redeem.columns.usedBy') },
  { key: 'used_at', label: t('admin.redeem.columns.usedAt'), sortable: true },
  { key: 'actions', label: t('admin.redeem.columns.actions') }
])

const typeOptions = computed(() => [
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') },
  { value: 'subscription', label: t('admin.redeem.subscription') },
  { value: 'invitation', label: t('admin.redeem.invitation') }
])

const filterTypeOptions = computed(() => [
  { value: '', label: t('admin.redeem.allTypes') },
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') },
  { value: 'subscription', label: t('admin.redeem.subscription') },
  { value: 'invitation', label: t('admin.redeem.invitation') }
])

const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.redeem.allStatus') },
  { value: 'unused', label: t('admin.redeem.unused') },
  { value: 'used', label: t('admin.redeem.used') },
  { value: 'expired', label: t('admin.redeem.status.expired') }
])

const {
  codes,
  loading,
  searchQuery,
  filters,
  pagination,
  showDeleteDialog,
  showDeleteUnusedDialog,
  copiedCode,
  loadCodes,
  handleSearch,
  handlePageChange,
  handlePageSizeChange,
  handleExportCodes,
  copyCodeToClipboard,
  handleDelete,
  confirmDelete,
  confirmDeleteUnused,
  dispose: disposeRedeemViewData
} = useRedeemViewData({
  t,
  showError: appStore.showError,
  showInfo: appStore.showInfo,
  showSuccess: appStore.showSuccess,
  copyToClipboard: clipboardCopy
})

const {
  showGenerateDialog,
  showResultDialog,
  generating,
  copiedAll,
  generatedCodes,
  generateForm,
  subscriptionGroupOptions,
  generatedCodesText,
  textareaHeight,
  loadSubscriptionGroups,
  handleGenerateCodes,
  closeResultDialog,
  copyGeneratedCodes,
  downloadGeneratedCodes,
  dispose: disposeRedeemGeneration
} = useRedeemGeneration({
  t,
  showError: appStore.showError,
  copyToClipboard: clipboardCopy,
  reloadCodes: loadCodes
})

onMounted(() => {
  void loadCodes()
  void loadSubscriptionGroups()
})

onUnmounted(() => {
  disposeRedeemViewData()
  disposeRedeemGeneration()
})
</script>
