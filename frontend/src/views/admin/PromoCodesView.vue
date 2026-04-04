<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <PromoCodesToolbar
          :search-query="searchQuery"
          :status="filters.status"
          :status-options="filterStatusOptions"
          :loading="loading"
          @update:search-query="searchQuery = $event"
          @update:status="filters.status = $event"
          @search="handleSearch"
          @refresh="loadCodes"
          @create="showCreateDialog = true"
        />
      </template>

      <template #table>
        <DataTable :columns="columns" :data="codes" :loading="loading">
          <template #cell-code="{ value }">
            <PromoCodeCodeCell
              :code="value"
              :copied="copiedCode === value"
              @copy="handleCopyCode(value)"
            />
          </template>

          <template #cell-bonus_amount="{ value }">
            <PromoCodeAmountCell :amount="value" />
          </template>

          <template #cell-usage="{ row }">
            <PromoCodeUsageCell :used-count="row.used_count" :max-uses="row.max_uses" />
          </template>

          <template #cell-status="{ row }">
            <PromoCodeStatusBadge :code="row" />
          </template>

          <template #cell-expires_at="{ value }">
            <PromoCodeDateCell :value="value" :fallback-text="t('admin.promo.neverExpires')" />
          </template>

          <template #cell-created_at="{ value }">
            <PromoCodeDateCell :value="value" />
          </template>

          <template #cell-actions="{ row }">
            <PromoCodeActionsCell
              @copy-link="copyRegisterLink(row)"
              @view-usages="handleViewUsages(row)"
              @edit="handleEdit(row)"
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
      </template>
    </TablePageLayout>

    <PromoCodeCreateDialog
      :show="showCreateDialog"
      :form="createForm"
      :submitting="creating"
      @close="closeCreateDialog"
      @submit="handleCreate"
    />

    <PromoCodeEditDialog
      :show="showEditDialog"
      :form="editForm"
      :status-options="statusOptions"
      :submitting="updating"
      @close="closeEditDialog"
      @submit="handleUpdate"
    />

    <PromoCodeUsagesDialog
      :show="showUsagesDialog"
      :loading="usagesLoading"
      :usages="usages"
      :page="usagesPage"
      :page-size="usagesPageSize"
      :total="usagesTotal"
      @close="closeUsagesDialog"
      @update:page="handleUsagesPageChange"
      @update:page-size="handleUsagesPageSizeChange"
    />

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.promo.deleteCode')"
      :message="t('admin.promo.deleteCodeConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import PromoCodeActionsCell from './promocodes/PromoCodeActionsCell.vue'
import PromoCodeAmountCell from './promocodes/PromoCodeAmountCell.vue'
import PromoCodeCodeCell from './promocodes/PromoCodeCodeCell.vue'
import PromoCodeCreateDialog from './promocodes/PromoCodeCreateDialog.vue'
import PromoCodeDateCell from './promocodes/PromoCodeDateCell.vue'
import PromoCodeEditDialog from './promocodes/PromoCodeEditDialog.vue'
import PromoCodeStatusBadge from './promocodes/PromoCodeStatusBadge.vue'
import PromoCodeUsageCell from './promocodes/PromoCodeUsageCell.vue'
import PromoCodeUsagesDialog from './promocodes/PromoCodeUsagesDialog.vue'
import PromoCodesToolbar from './promocodes/PromoCodesToolbar.vue'
import { usePromoCodesViewActions } from './usePromoCodesViewActions'
import { usePromoCodesViewData } from './usePromoCodesViewData'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

// Options
const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.promo.allStatus') },
  { value: 'active', label: t('admin.promo.statusActive') },
  { value: 'disabled', label: t('admin.promo.statusDisabled') }
])

const statusOptions = computed(() => [
  { value: 'active', label: t('admin.promo.statusActive') },
  { value: 'disabled', label: t('admin.promo.statusDisabled') }
])

const columns = computed<Column[]>(() => [
  { key: 'code', label: t('admin.promo.columns.code') },
  { key: 'bonus_amount', label: t('admin.promo.columns.bonusAmount'), sortable: true },
  { key: 'usage', label: t('admin.promo.columns.usage') },
  { key: 'status', label: t('admin.promo.columns.status'), sortable: true },
  { key: 'expires_at', label: t('admin.promo.columns.expiresAt'), sortable: true },
  { key: 'created_at', label: t('admin.promo.columns.createdAt'), sortable: true },
  { key: 'actions', label: t('admin.promo.columns.actions') }
])

const {
  codes,
  loading,
  searchQuery,
  copiedCode,
  filters,
  pagination,
  loadCodes,
  handleSearch,
  handlePageChange,
  handlePageSizeChange,
  handleCopyCode,
  dispose
} = usePromoCodesViewData({
  t,
  showError: appStore.showError,
  copyToClipboard: clipboardCopy
})

const {
  creating,
  updating,
  showCreateDialog,
  showEditDialog,
  showDeleteDialog,
  showUsagesDialog,
  createForm,
  editForm,
  usages,
  usagesLoading,
  usagesPage,
  usagesPageSize,
  usagesTotal,
  handleCreate,
  closeCreateDialog,
  handleEdit,
  closeEditDialog,
  handleUpdate,
  copyRegisterLink,
  handleDelete,
  confirmDelete,
  handleViewUsages,
  closeUsagesDialog,
  handleUsagesPageChange,
  handleUsagesPageSizeChange
} = usePromoCodesViewActions({
  origin: window.location.origin,
  t,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError,
  copyToClipboard: clipboardCopy,
  reloadCodes: loadCodes
})

onMounted(() => {
  void loadCodes()
})

onUnmounted(() => {
  dispose()
})
</script>
