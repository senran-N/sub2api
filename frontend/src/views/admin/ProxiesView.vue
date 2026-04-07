<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <ProxyFilterFields
            v-model:searchQuery="searchQuery"
            v-model:protocol="filters.protocol"
            v-model:status="filters.status"
            :protocol-options="protocolOptions"
            :status-options="statusOptions"
            @search-input="handleSearch"
            @protocol-change="loadProxies"
            @status-change="loadProxies"
          />

          <ProxyActionToolbar
            :loading="loading"
            :batch-testing="batchTesting"
            :batch-quality-checking="batchQualityChecking"
            :selected-count="selectedCount"
            @refresh="loadProxies"
            @batch-test="handleBatchTest"
            @batch-quality-check="handleBatchQualityCheck"
            @batch-delete="openBatchDelete"
            @import="showImportData = true"
            @export="showExportDataDialog = true"
            @create="showCreateModal = true"
          />
        </div>
      </template>

      <template #table>
        <div ref="proxyTableRef" class="flex min-h-0 flex-1 flex-col overflow-hidden">
        <DataTable :columns="columns" :data="proxies" :loading="loading">
          <template #header-select>
            <ProxySelectionCheckbox
              :checked="allVisibleSelected"
              @change="toggleSelectAllVisible"
            />
          </template>

          <template #cell-select="{ row }">
            <ProxySelectionCheckbox
              :checked="selectedProxyIds.has(row.id)"
              @change="toggleSelectRow(row.id, $event)"
            />
          </template>

          <template #cell-name="{ value }">
            <ProxyNameCell :value="value" />
          </template>

          <template #cell-protocol="{ value }">
            <ProxyProtocolBadge :protocol="value" />
          </template>

          <template #cell-address="{ row }">
            <ProxyAddressCell
              :proxy="row"
              :copy-menu-open="copyMenuProxyId === row.id"
              :copy-formats="getCopyFormats(row)"
              @copy-url="copyProxyUrl"
              @toggle-copy-menu="toggleCopyMenu"
              @copy-format="copyFormat"
            />
          </template>

          <template #cell-auth="{ row }">
            <ProxyAuthCell
              :proxy="row"
              :password-visible="visiblePasswordIds.has(row.id)"
              @toggle-password="togglePasswordVisibility"
            />
          </template>

          <template #cell-location="{ row }">
            <ProxyLocationCell :proxy="row" />
          </template>

          <template #cell-account_count="{ row }">
            <ProxyAccountCountCell :proxy="row" @accounts="openAccountsModal" />
          </template>

          <template #cell-latency="{ row }">
            <ProxyLatencyCell :proxy="row" />
          </template>

          <template #cell-status="{ value }">
            <ProxyStatusBadge :status="value" />
          </template>

          <template #cell-actions="{ row }">
            <ProxyActionsCell
              :proxy="row"
              :testing="testingProxyIds.has(row.id)"
              :quality-checking="qualityCheckingProxyIds.has(row.id)"
              @test="handleTestConnection"
              @quality-check="handleQualityCheck"
              @edit="handleEdit"
              @delete="handleDelete"
            />
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.proxies.noProxiesYet')"
              :description="t('admin.proxies.createFirstProxy')"
              :action-text="t('admin.proxies.createProxy')"
              @action="showCreateModal = true"
            />
          </template>
        </DataTable>
        </div>
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

    <!-- Create Proxy Modal -->
    <BaseDialog
      :show="showCreateModal"
      :title="t('admin.proxies.createProxy')"
      width="normal"
      @close="closeCreateModal"
    >
      <ProxyCreateModeTabs v-model="createMode" />

      <!-- Standard Add Form -->
      <form
        v-if="createMode === 'standard'"
        id="create-proxy-form"
        @submit.prevent="handleCreateProxy"
        class="space-y-5"
      >
        <ProxyFormFieldsSection
          :form="createForm"
          :protocol-options="protocolSelectOptions"
          :password-visible="createPasswordVisible"
          :name-placeholder="t('admin.proxies.enterProxyName')"
          :host-placeholder="t('admin.proxies.form.hostPlaceholder')"
          :port-placeholder="t('admin.proxies.form.portPlaceholder')"
          :username-placeholder="t('admin.proxies.optionalAuth')"
          :password-placeholder="t('admin.proxies.optionalAuth')"
          @toggle-password-visibility="createPasswordVisible = !createPasswordVisible"
        />
      </form>

      <ProxyBatchInputSection
        v-else
        v-model="batchInput"
        :summary="batchParseResult"
        @input="parseBatchInput"
      />

      <template #footer>
        <ProxyCreateDialogFooter
          :mode="createMode"
          :submitting="submitting"
          :valid-count="batchParseResult.valid"
          @close="closeCreateModal"
          @batch-create="handleBatchCreate"
        />
      </template>
    </BaseDialog>

    <!-- Edit Proxy Modal -->
    <BaseDialog
      :show="showEditModal"
      :title="t('admin.proxies.editProxy')"
      width="normal"
      @close="closeEditModal"
    >
      <form
        v-if="editingProxy"
        id="edit-proxy-form"
        @submit.prevent="handleUpdateProxy"
        class="space-y-5"
      >
        <ProxyFormFieldsSection
          :form="editForm"
          :protocol-options="protocolSelectOptions"
          :password-visible="editPasswordVisible"
          :password-placeholder="t('admin.proxies.leaveEmptyToKeep')"
          :show-status="true"
          :status-options="editStatusOptions"
          @toggle-password-visibility="editPasswordVisible = !editPasswordVisible"
          @password-input="editPasswordDirty = true"
        />
      </form>

      <template #footer>
        <ProxyEditDialogFooter
          :show-submit="!!editingProxy"
          :submitting="submitting"
          @close="closeEditModal"
        />
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.proxies.deleteProxy')"
      :message="t('admin.proxies.deleteConfirm', { name: deletingProxy?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Batch Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showBatchDeleteDialog"
      :title="t('admin.proxies.batchDelete')"
      :message="t('admin.proxies.batchDeleteConfirm', { count: selectedCount })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmBatchDelete"
      @cancel="showBatchDeleteDialog = false"
    />
    <ConfirmDialog
      :show="showExportDataDialog"
      :title="t('admin.proxies.dataExport')"
      :message="t('admin.proxies.dataExportConfirmMessage')"
      :confirm-text="t('admin.proxies.dataExportConfirm')"
      :cancel-text="t('common.cancel')"
      @confirm="handleExportData"
      @cancel="showExportDataDialog = false"
    />

    <ImportDataModal
      :show="showImportData"
      @close="showImportData = false"
      @imported="handleDataImported"
    />

    <ProxyQualityReportDialog
      :show="showQualityReportDialog"
      :proxy-name="qualityReportProxy?.name"
      :report="qualityReport"
      @close="closeQualityReportDialog"
    />

    <ProxyAccountsDialog
      :show="showAccountsModal"
      :proxy-name="accountsProxy?.name"
      :loading="accountsLoading"
      :accounts="proxyAccounts"
      @close="closeAccountsModal"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import type { Proxy } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ImportDataModal from '@/components/admin/proxy/ImportDataModal.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useSwipeSelect } from '@/composables/useSwipeSelect'
import { useTableSelection } from '@/composables/useTableSelection'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { buildProxyCopyFormats } from './proxies/proxyUtils'
import {
  createDefaultProxyBatchParseState,
  createDefaultProxyCreateForm,
  createDefaultProxyEditForm
} from './proxies/proxyForm'
import { useProxyFormActions } from './proxies/useProxyFormActions'
import { useProxyListData } from './proxies/useProxyListData'
import ProxyAccountCountCell from './proxies/ProxyAccountCountCell.vue'
import ProxyAccountsDialog from './proxies/ProxyAccountsDialog.vue'
import ProxyActionToolbar from './proxies/ProxyActionToolbar.vue'
import ProxyActionsCell from './proxies/ProxyActionsCell.vue'
import ProxyAddressCell from './proxies/ProxyAddressCell.vue'
import ProxyAuthCell from './proxies/ProxyAuthCell.vue'
import ProxyBatchInputSection from './proxies/ProxyBatchInputSection.vue'
import ProxyCreateDialogFooter from './proxies/ProxyCreateDialogFooter.vue'
import ProxyCreateModeTabs from './proxies/ProxyCreateModeTabs.vue'
import ProxyEditDialogFooter from './proxies/ProxyEditDialogFooter.vue'
import ProxyFilterFields from './proxies/ProxyFilterFields.vue'
import ProxyFormFieldsSection from './proxies/ProxyFormFieldsSection.vue'
import ProxyLatencyCell from './proxies/ProxyLatencyCell.vue'
import ProxyLocationCell from './proxies/ProxyLocationCell.vue'
import ProxyNameCell from './proxies/ProxyNameCell.vue'
import ProxyProtocolBadge from './proxies/ProxyProtocolBadge.vue'
import ProxyQualityReportDialog from './proxies/ProxyQualityReportDialog.vue'
import ProxySelectionCheckbox from './proxies/ProxySelectionCheckbox.vue'
import ProxyStatusBadge from './proxies/ProxyStatusBadge.vue'
import { useProxyTestingActions } from './proxies/useProxyTestingActions'
import { useProxyViewInteractions } from './proxies/useProxyViewInteractions'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const columns = computed<Column[]>(() => [
  { key: 'select', label: '', sortable: false },
  { key: 'name', label: t('admin.proxies.columns.name'), sortable: true },
  { key: 'protocol', label: t('admin.proxies.columns.protocol'), sortable: true },
  { key: 'address', label: t('admin.proxies.columns.address'), sortable: false },
  { key: 'auth', label: t('admin.proxies.columns.auth'), sortable: false },
  { key: 'location', label: t('admin.proxies.columns.location'), sortable: false },
  { key: 'account_count', label: t('admin.proxies.columns.accounts'), sortable: true },
  { key: 'latency', label: t('admin.proxies.columns.latency'), sortable: false },
  { key: 'status', label: t('admin.proxies.columns.status'), sortable: true },
  { key: 'actions', label: t('admin.proxies.columns.actions'), sortable: false }
])

// Filter options
const protocolOptions = computed(() => [
  { value: '', label: t('admin.proxies.allProtocols') },
  { value: 'http', label: 'HTTP' },
  { value: 'https', label: 'HTTPS' },
  { value: 'socks5', label: 'SOCKS5' },
  { value: 'socks5h', label: 'SOCKS5H' }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.proxies.allStatus') },
  { value: 'active', label: t('admin.accounts.status.active') },
  { value: 'inactive', label: t('admin.accounts.status.inactive') }
])

// Form options
const protocolSelectOptions = computed(() => [
  { value: 'http', label: t('admin.proxies.protocols.http') },
  { value: 'https', label: t('admin.proxies.protocols.https') },
  { value: 'socks5', label: t('admin.proxies.protocols.socks5') },
  { value: 'socks5h', label: t('admin.proxies.protocols.socks5h') }
])

const editStatusOptions = computed(() => [
  { value: 'active', label: t('admin.accounts.status.active') },
  { value: 'inactive', label: t('admin.accounts.status.inactive') }
])

const proxies = ref<Proxy[]>([])
const visiblePasswordIds = reactive(new Set<number>())
const loading = ref(false)
const searchQuery = ref('')
const filters = reactive({
  protocol: '',
  status: ''
})
const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})

const showCreateModal = ref(false)
const createPasswordVisible = ref(false)
const showEditModal = ref(false)
const editPasswordVisible = ref(false)
const editPasswordDirty = ref(false)
const showImportData = ref(false)
const submitting = ref(false)
const proxyTableRef = ref<HTMLElement | null>(null)
const {
  selectedSet: selectedProxyIds,
  selectedCount,
  allVisibleSelected,
  isSelected,
  select,
  deselect,
  clear: clearSelectedProxies,
  removeMany: removeSelectedProxies,
  toggleVisible
} = useTableSelection<Proxy>({
  rows: proxies,
  getId: (proxy) => proxy.id
})
useSwipeSelect(proxyTableRef, {
  isSelected,
  select,
  deselect
})
const editingProxy = ref<Proxy | null>(null)

// Batch import state
const createMode = ref<'standard' | 'batch'>('standard')
const batchInput = ref('')
const batchParseResult = reactive({
  ...createDefaultProxyBatchParseState()
})

const createForm = reactive(createDefaultProxyCreateForm())

const editForm = reactive(createDefaultProxyEditForm())

const togglePasswordVisibility = (proxyId: number) => {
  if (visiblePasswordIds.has(proxyId)) {
    visiblePasswordIds.delete(proxyId)
    return
  }

  visiblePasswordIds.add(proxyId)
}

const toggleSelectRow = (id: number, event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.checked) {
    select(id)
    return
  }
  deselect(id)
}

const toggleSelectAllVisible = (event: Event) => {
  const target = event.target as HTMLInputElement
  toggleVisible(target.checked)
}

const { cleanup, handlePageChange, handlePageSizeChange, handleSearch, loadProxies } =
  useProxyListData({
    proxies,
    loading,
    searchQuery,
    filters,
    pagination,
    t,
    showError: (message) => appStore.showError(message)
  })

const {
  batchQualityChecking,
  batchTesting,
  closeQualityReportDialog,
  handleBatchQualityCheck,
  handleBatchTest,
  handleQualityCheck,
  handleTestConnection,
  qualityCheckingProxyIds,
  qualityReport,
  qualityReportProxy,
  showQualityReportDialog,
  testingProxyIds
} = useProxyTestingActions({
  proxies,
  selectedProxyIds,
  selectedCount,
  loadProxies,
  getBatchFilters: () => ({
    protocol: filters.protocol || undefined,
    status: (filters.status || undefined) as 'active' | 'inactive' | undefined,
    search: searchQuery.value || undefined
  }),
  t,
  showSuccess: (message) => appStore.showSuccess(message),
  showError: (message) => appStore.showError(message),
  showInfo: (message) => appStore.showInfo(message)
})

const {
  accountsLoading,
  accountsProxy,
  closeAccountsModal,
  closeCopyMenu,
  confirmBatchDelete,
  confirmDelete,
  copyFormat,
  copyMenuProxyId,
  copyProxyUrl,
  deletingProxy,
  handleDelete,
  handleExportData,
  openAccountsModal,
  openBatchDelete,
  proxyAccounts,
  showAccountsModal,
  showBatchDeleteDialog,
  showDeleteDialog,
  showExportDataDialog,
  toggleCopyMenu
} = useProxyViewInteractions({
  selectedCount,
  selectedProxyIds,
  filters,
  searchQuery,
  copyToClipboard,
  clearSelectedProxies,
  removeSelectedProxies,
  loadProxies,
  t,
  showSuccess: (message) => appStore.showSuccess(message),
  showError: (message) => appStore.showError(message),
  showInfo: (message) => appStore.showInfo(message)
})

const {
  closeCreateModal,
  closeEditModal,
  handleBatchCreate,
  handleCreateProxy,
  handleDataImported,
  handleEdit,
  handleUpdateProxy,
  parseBatchInput
} = useProxyFormActions({
  showCreateModal,
  createMode,
  createForm,
  createPasswordVisible,
  batchInput,
  batchParseResult,
  showImportData,
  editingProxy,
  editForm,
  showEditModal,
  editPasswordVisible,
  editPasswordDirty,
  submitting,
  loadProxies,
  t,
  showSuccess: (message) => appStore.showSuccess(message),
  showError: (message) => appStore.showError(message),
  showInfo: (message) => appStore.showInfo(message)
})

// ── Proxy URL copy ──
const getCopyFormats = (row: Proxy) => buildProxyCopyFormats(row)

onMounted(() => {
  loadProxies()
  document.addEventListener('click', closeCopyMenu)
})

onUnmounted(() => {
  cleanup()
  document.removeEventListener('click', closeCopyMenu)
})
</script>
