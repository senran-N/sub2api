import { ref, type Ref } from 'vue'
import { buildAccountExportFilename, buildAccountExportRequest } from './accountsView'
import type { AccountListFilters } from './accountsList'

interface AccountsViewExportOptions {
  t: (key: string) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
  selectedIds: Ref<number[]>
  includeProxyOnExport: Ref<boolean>
  params: Pick<AccountListFilters, 'platform' | 'type' | 'status' | 'search'>
  showExportDataDialog: Ref<boolean>
  exportData: (
    payload:
      | { ids: number[]; includeProxies: boolean }
      | { includeProxies: boolean; filters: Pick<AccountListFilters, 'platform' | 'type' | 'status' | 'search'> }
  ) => Promise<unknown>
  downloadJson: (data: unknown, filename: string) => void
}

export function downloadAccountsExportJson(data: unknown, filename: string): void {
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export function useAccountsViewExport(options: AccountsViewExportOptions) {
  const exportingData = ref(false)

  const handleExportData = async () => {
    if (exportingData.value) {
      return
    }

    exportingData.value = true
    try {
      const payload = buildAccountExportRequest(
        options.selectedIds.value,
        options.includeProxyOnExport.value,
        options.params
      )
      const data = await options.exportData(payload)
      options.downloadJson(data, buildAccountExportFilename())
      options.showSuccess(options.t('admin.accounts.dataExported'))
    } catch (error) {
      const message =
        typeof error === 'object' &&
        error !== null &&
        'message' in error &&
        typeof error.message === 'string' &&
        error.message
          ? error.message
          : options.t('admin.accounts.dataExportFailed')
      options.showError(message)
    } finally {
      exportingData.value = false
      options.showExportDataDialog.value = false
    }
  }

  return {
    exportingData,
    handleExportData
  }
}
