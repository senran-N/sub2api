import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import type { RedeemCode } from '@/types'
import { isAbortError, resolveRequestErrorMessage } from '@/utils/requestError'
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
  const codes = ref<RedeemCode[]>([])
  const loading = ref(false)
  const searchQuery = ref('')
  const filters = reactive(createDefaultRedeemFilters())
  const pagination = reactive({
    page: 1,
    page_size: getPersistedPageSize(),
    total: 0,
    pages: 0
  })

  const showDeleteDialog = ref(false)
  const showDeleteUnusedDialog = ref(false)
  const deletingCode = ref<RedeemCode | null>(null)
  const copiedCode = ref<string | null>(null)

  let abortController: AbortController | null = null
  let searchTimeout: ReturnType<typeof setTimeout> | null = null
  let copiedCodeTimeout: ReturnType<typeof setTimeout> | null = null

  const loadCodes = async () => {
    abortController?.abort()

    const requestController = new AbortController()
    abortController = requestController
    loading.value = true

    try {
      const response = await adminAPI.redeem.list(
        pagination.page,
        pagination.page_size,
        buildRedeemListFilters(filters, searchQuery.value),
        {
          signal: requestController.signal
        }
      )

      if (requestController.signal.aborted || abortController !== requestController) {
        return
      }

      codes.value = response.items
      pagination.total = response.total
      pagination.pages = response.pages
      pagination.page = response.page
      pagination.page_size = response.page_size
    } catch (error) {
      if (requestController.signal.aborted || isAbortError(error)) {
        return
      }

      options.showError(resolveRequestErrorMessage(error, options.t('admin.redeem.failedToLoad')))
      console.error('Error loading redeem codes:', error)
    } finally {
      if (abortController === requestController) {
        loading.value = false
        abortController = null
      }
    }
  }

  const handleSearch = () => {
    if (searchTimeout) {
      clearTimeout(searchTimeout)
    }

    searchTimeout = setTimeout(() => {
      pagination.page = 1
      void loadCodes()
    }, 300)
  }

  const handlePageChange = (page: number) => {
    pagination.page = page
    void loadCodes()
  }

  const handlePageSizeChange = (pageSize: number) => {
    pagination.page_size = pageSize
    pagination.page = 1
    void loadCodes()
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
    abortController?.abort()
    if (searchTimeout) {
      clearTimeout(searchTimeout)
    }
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
