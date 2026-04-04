<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <AnnouncementsToolbar
          :search-query="searchQuery"
          :status="filters.status"
          :status-options="statusFilterOptions"
          :loading="loading"
          @update:search-query="searchQuery = $event"
          @update:status="filters.status = $event"
          @search="handleSearch"
          @status-change="handleStatusChange"
          @refresh="loadAnnouncements"
          @create="openCreateDialog"
        />
      </template>

      <template #table>
        <DataTable :columns="columns" :data="announcements" :loading="loading">
          <template #cell-title="{ value, row }">
            <AnnouncementTitleCell :id="row.id" :title="value" :created-at="row.created_at" />
          </template>

          <template #cell-status="{ value }">
            <AnnouncementStatusBadge :status="value" />
          </template>

          <template #cell-notifyMode="{ row }">
            <AnnouncementNotifyModeBadge :notify-mode="row.notify_mode" />
          </template>

          <template #cell-targeting="{ row }">
            <AnnouncementTargetingCell :targeting="row.targeting" />
          </template>

          <template #cell-timeRange="{ row }">
            <AnnouncementTimeRangeCell :starts-at="row.starts_at" :ends-at="row.ends_at" />
          </template>

          <template #cell-createdAt="{ value }">
            <AnnouncementCreatedAtCell :value="value" />
          </template>

          <template #cell-actions="{ row }">
            <AnnouncementActionsCell
              @read-status="openReadStatus(row)"
              @edit="openEditDialog(row)"
              @delete="handleDelete(row)"
            />
          </template>

          <template #empty>
            <EmptyState
              :title="t('empty.noData')"
              :description="t('admin.announcements.failedToLoad')"
              :action-text="t('admin.announcements.createAnnouncement')"
              @action="openCreateDialog"
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

    <AnnouncementEditDialog
      :show="showEditDialog"
      :editing="isEditing"
      :saving="saving"
      :form="form"
      :subscription-groups="subscriptionGroups"
      :status-options="statusOptions"
      :notify-mode-options="notifyModeOptions"
      @close="closeEdit"
      @submit="handleSave"
    />

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.announcements.deleteAnnouncement')"
      :message="t('admin.announcements.deleteConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <AnnouncementReadStatusDialog
      :show="showReadStatusDialog"
      :announcement-id="readStatusAnnouncementId"
      @close="showReadStatusDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { Announcement } from '@/types'
import type { Column } from '@/components/common/types'

import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'

import AnnouncementReadStatusDialog from '@/components/admin/announcements/AnnouncementReadStatusDialog.vue'
import AnnouncementActionsCell from './announcements/AnnouncementActionsCell.vue'
import AnnouncementCreatedAtCell from './announcements/AnnouncementCreatedAtCell.vue'
import AnnouncementEditDialog from './announcements/AnnouncementEditDialog.vue'
import AnnouncementNotifyModeBadge from './announcements/AnnouncementNotifyModeBadge.vue'
import AnnouncementStatusBadge from './announcements/AnnouncementStatusBadge.vue'
import AnnouncementTargetingCell from './announcements/AnnouncementTargetingCell.vue'
import AnnouncementTimeRangeCell from './announcements/AnnouncementTimeRangeCell.vue'
import AnnouncementTitleCell from './announcements/AnnouncementTitleCell.vue'
import AnnouncementsToolbar from './announcements/AnnouncementsToolbar.vue'
import { useAnnouncementsViewData } from './useAnnouncementsViewData'
import { useAnnouncementsViewEditor } from './useAnnouncementsViewEditor'

const { t } = useI18n()
const appStore = useAppStore()

const statusFilterOptions = computed(() => [
  { value: '', label: t('admin.announcements.allStatus') },
  { value: 'draft', label: t('admin.announcements.statusLabels.draft') },
  { value: 'active', label: t('admin.announcements.statusLabels.active') },
  { value: 'archived', label: t('admin.announcements.statusLabels.archived') }
])

const statusOptions = computed<Array<{ value: Announcement['status']; label: string }>>(() => [
  { value: 'draft', label: t('admin.announcements.statusLabels.draft') },
  { value: 'active', label: t('admin.announcements.statusLabels.active') },
  { value: 'archived', label: t('admin.announcements.statusLabels.archived') }
])

const notifyModeOptions = computed<Array<{ value: NonNullable<Announcement['notify_mode']>; label: string }>>(() => [
  { value: 'silent', label: t('admin.announcements.notifyModeLabels.silent') },
  { value: 'popup', label: t('admin.announcements.notifyModeLabels.popup') }
])

const columns = computed<Column[]>(() => [
  { key: 'title', label: t('admin.announcements.columns.title') },
  { key: 'status', label: t('admin.announcements.columns.status') },
  { key: 'notifyMode', label: t('admin.announcements.columns.notifyMode') },
  { key: 'targeting', label: t('admin.announcements.columns.targeting') },
  { key: 'timeRange', label: t('admin.announcements.columns.timeRange') },
  { key: 'createdAt', label: t('admin.announcements.columns.createdAt') },
  { key: 'actions', label: t('admin.announcements.columns.actions') }
])

const {
  announcements,
  loading,
  filters,
  searchQuery,
  pagination,
  loadAnnouncements,
  handlePageChange,
  handlePageSizeChange,
  handleStatusChange,
  handleSearch,
  dispose
} = useAnnouncementsViewData({
  t,
  showError: appStore.showError
})

const {
  showEditDialog,
  saving,
  isEditing,
  form,
  subscriptionGroups,
  loadSubscriptionGroups,
  openCreateDialog,
  openEditDialog,
  closeEdit,
  handleSave
} = useAnnouncementsViewEditor({
  t,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError,
  reloadAnnouncements: loadAnnouncements
})

// ===== Delete =====
const showDeleteDialog = ref(false)
const deletingAnnouncement = ref<Announcement | null>(null)

function handleDelete(row: Announcement) {
  deletingAnnouncement.value = row
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!deletingAnnouncement.value) return

  try {
    await adminAPI.announcements.delete(deletingAnnouncement.value.id)
    appStore.showSuccess(t('common.success'))
    showDeleteDialog.value = false
    deletingAnnouncement.value = null
    await loadAnnouncements()
  } catch (error: any) {
    console.error('Failed to delete announcement:', error)
    appStore.showError(error.response?.data?.detail || t('admin.announcements.failedToDelete'))
  }
}

// ===== Read status =====
const showReadStatusDialog = ref(false)
const readStatusAnnouncementId = ref<number | null>(null)

function openReadStatus(row: Announcement) {
  readStatusAnnouncementId.value = row.id
  showReadStatusDialog.value = true
}

onMounted(async () => {
  await loadSubscriptionGroups()
  await loadAnnouncements()
})

onUnmounted(() => {
  dispose()
})
</script>
