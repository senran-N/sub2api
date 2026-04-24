import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { useTableLoader } from '@/composables/useTableLoader'
import type { RedeemCode } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import { buildRedeemExportFilters, buildRedeemListFilters, createDefaultRedeemFilters } from './redeemForm'

interface RedeemViewDataOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showInfo: (message: string) => void
  showSuccess: (message: string) => void
  copyToClipboard: (text: string, successMessage?: string) => Promise<boolean>
}

function downloadBlob(blob: Blob, filename: string): void {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

export function useRedeemViewData(options: RedeemViewDataOptions) {
  const searchQuery = ref('')
  const filters = reactive(createDefaultRedeemFilters())
  const {
    items: codes,
    loading,
    pagination,
    load: loadCodes,
    debouncedReload,
    handlePageChange,
    handlePageSizeChange,
    dispose: disposeTableLoader
  } = useTableLoader<RedeemCode, Record<string, never>>({
    fetchFn: (page, pageSize, _params, requestOptions) =>
      adminAPI.redeem.list(
        page,
        pageSize,
        buildRedeemListFilters(filters, searchQuery.value),
        requestOptions
      ),
    onError: (error) => {
      options.showError(resolveRequestErrorMessage(error, options.t('admin.redeem.failedToLoad')))
    },
    syncPaginationFromResponse: true,
    clampPageChange: false
  })

  const showDeleteDialog = ref(false)
  const showDeleteUnusedDialog = ref(false)
  const deletingCode = ref<RedeemCode | null>(null)
  const copiedCode = ref<string | null>(null)

  let copiedCodeTimeout: ReturnType<typeof setTimeout> | null = null

  const handleSearch = () => {
    void debouncedReload()
  }

  const handleExportCodes = async () => {
    try {
      const blob = await adminAPI.redeem.exportCodes(buildRedeemExportFilters(filters))
      const date = new Date().toISOString().split('T')[0]
      downloadBlob(blob, `redeem-codes-${date}.csv`)
      options.showSuccess(options.t('admin.redeem.codesExported'))
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('admin.redeem.failedToExport')))
      console.error('Error exporting codes:', error)
    }
  }

  const copyCodeToClipboard = async (text: string) => {
    const success = await options.copyToClipboard(text, options.t('admin.redeem.copied'))
    if (!success) {
      return
    }

    copiedCode.value = text
    if (copiedCodeTimeout) {
      clearTimeout(copiedCodeTimeout)
    }
    copiedCodeTimeout = setTimeout(() => {
      copiedCode.value = null
    }, 2000)
  }

  const handleDelete = (code: RedeemCode) => {
    deletingCode.value = code
    showDeleteDialog.value = true
  }

  const confirmDelete = async () => {
    if (!deletingCode.value) {
      return
    }

    try {
      await adminAPI.redeem.delete(deletingCode.value.id)
      options.showSuccess(options.t('admin.redeem.codeDeleted'))
      showDeleteDialog.value = false
      deletingCode.value = null
      await loadCodes()
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('admin.redeem.failedToDelete')))
      console.error('Error deleting code:', error)
    }
  }

  const confirmDeleteUnused = async () => {
    try {
      const unusedCodesResponse = await adminAPI.redeem.list(1, 1000, { status: 'unused' })
      const unusedCodeIds = unusedCodesResponse.items.map((code) => code.id)

      if (unusedCodeIds.length === 0) {
        options.showInfo(options.t('admin.redeem.noUnusedCodes'))
        showDeleteUnusedDialog.value = false
        return
      }

      const result = await adminAPI.redeem.batchDelete(unusedCodeIds)
      options.showSuccess(options.t('admin.redeem.codesDeleted', { count: result.deleted }))
      showDeleteUnusedDialog.value = false
      await loadCodes()
    } catch (error) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.redeem.failedToDeleteUnused'))
      )
      console.error('Error deleting unused codes:', error)
    }
  }

  const dispose = () => {
    disposeTableLoader()
    if (copiedCodeTimeout) {
      clearTimeout(copiedCodeTimeout)
    }
  }

  return {
    codes,
    loading,
    searchQuery,
    filters,
    pagination,
    showDeleteDialog,
    showDeleteUnusedDialog,
    deletingCode,
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
    dispose
  }
}
