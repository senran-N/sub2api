import { ref, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Proxy, ProxyAccountSummary } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import { buildProxyUrl } from './proxyUtils'

interface ProxyExportFilters {
  protocol: string
  status: string
}

interface ProxyViewInteractionsOptions {
  selectedCount: Readonly<Ref<number>>
  selectedProxyIds: Ref<Set<number>>
  filters: ProxyExportFilters
  searchQuery: Ref<string>
  copyToClipboard: (value: string, message: string) => void
  clearSelectedProxies: () => void
  removeSelectedProxies: (ids: number[]) => void
  loadProxies: () => void | Promise<void>
  t: (key: string, params?: Record<string, unknown>) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
  showInfo: (message: string) => void
}

function formatExportTimestamp(now: Date = new Date()) {
  const pad2 = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}${pad2(now.getMonth() + 1)}${pad2(now.getDate())}${pad2(now.getHours())}${pad2(now.getMinutes())}${pad2(now.getSeconds())}`
}

export function useProxyViewInteractions(options: ProxyViewInteractionsOptions) {
  const copyMenuProxyId = ref<number | null>(null)
  const showDeleteDialog = ref(false)
  const showBatchDeleteDialog = ref(false)
  const showExportDataDialog = ref(false)
  const showAccountsModal = ref(false)
  const exportingData = ref(false)
  const accountsProxy = ref<Proxy | null>(null)
  const proxyAccounts = ref<ProxyAccountSummary[]>([])
  const accountsLoading = ref(false)
  const deletingProxy = ref<Proxy | null>(null)
  const accountsRequestSeq = ref(0)

  const handleExportData = async () => {
    if (exportingData.value) {
      return
    }

    exportingData.value = true
    try {
      const dataPayload = await adminAPI.proxies.exportData(
        options.selectedCount.value > 0
          ? { ids: Array.from(options.selectedProxyIds.value) }
          : {
              filters: {
                protocol: options.filters.protocol || undefined,
                status: (options.filters.status || undefined) as
                  | 'active'
                  | 'inactive'
                  | undefined,
                search: options.searchQuery.value || undefined
              }
            }
      )

      const timestamp = formatExportTimestamp()
      const filename = `sub2api-proxy-${timestamp}.json`
      const blob = new Blob([JSON.stringify(dataPayload, null, 2)], {
        type: 'application/json'
      })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      link.click()
      URL.revokeObjectURL(url)
      options.showSuccess(options.t('admin.proxies.dataExported'))
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.dataExportFailed'))
      )
    } finally {
      exportingData.value = false
      showExportDataDialog.value = false
    }
  }

  const handleDelete = (proxy: Proxy) => {
    if ((proxy.account_count || 0) > 0) {
      options.showError(options.t('admin.proxies.deleteBlockedInUse'))
      return
    }

    deletingProxy.value = proxy
    showDeleteDialog.value = true
  }

  const openBatchDelete = () => {
    if (options.selectedCount.value === 0) {
      return
    }

    showBatchDeleteDialog.value = true
  }

  const confirmDelete = async () => {
    if (!deletingProxy.value) {
      return
    }

    try {
      await adminAPI.proxies.delete(deletingProxy.value.id)
      options.showSuccess(options.t('admin.proxies.proxyDeleted'))
      showDeleteDialog.value = false
      options.removeSelectedProxies([deletingProxy.value.id])
      deletingProxy.value = null
      await options.loadProxies()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.failedToDelete'))
      )
      console.error('Error deleting proxy:', error)
    }
  }

  const confirmBatchDelete = async () => {
    const ids = Array.from(options.selectedProxyIds.value)
    if (ids.length === 0) {
      showBatchDeleteDialog.value = false
      return
    }

    try {
      const result = await adminAPI.proxies.batchDelete(ids)
      const deleted = result.deleted_ids?.length || 0
      const skipped = result.skipped?.length || 0

      if (deleted > 0) {
        options.showSuccess(options.t('admin.proxies.batchDeleteDone', { deleted, skipped }))
      } else if (skipped > 0) {
        options.showInfo(options.t('admin.proxies.batchDeleteSkipped', { skipped }))
      }

      options.clearSelectedProxies()
      showBatchDeleteDialog.value = false
      await options.loadProxies()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.batchDeleteFailed'))
      )
      console.error('Error batch deleting proxies:', error)
    }
  }

  const openAccountsModal = async (proxy: Proxy) => {
    const requestSeq = accountsRequestSeq.value + 1
    accountsRequestSeq.value = requestSeq
    accountsProxy.value = proxy
    proxyAccounts.value = []
    accountsLoading.value = true
    showAccountsModal.value = true

    try {
      const accounts = await adminAPI.proxies.getProxyAccounts(proxy.id)
      if (requestSeq !== accountsRequestSeq.value) {
        return
      }

      proxyAccounts.value = accounts
    } catch (error: unknown) {
      if (requestSeq !== accountsRequestSeq.value) {
        return
      }

      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.accountsFailed'))
      )
      console.error('Error loading proxy accounts:', error)
    } finally {
      if (requestSeq === accountsRequestSeq.value) {
        accountsLoading.value = false
      }
    }
  }

  const closeAccountsModal = () => {
    accountsRequestSeq.value += 1
    showAccountsModal.value = false
    accountsProxy.value = null
    proxyAccounts.value = []
    accountsLoading.value = false
  }

  const copyProxyUrl = (proxy: Proxy) => {
    options.copyToClipboard(buildProxyUrl(proxy), options.t('admin.proxies.urlCopied'))
    copyMenuProxyId.value = null
  }

  const toggleCopyMenu = (proxyId: number) => {
    copyMenuProxyId.value = copyMenuProxyId.value === proxyId ? null : proxyId
  }

  const copyFormat = (value: string) => {
    options.copyToClipboard(value, options.t('admin.proxies.urlCopied'))
    copyMenuProxyId.value = null
  }

  const closeCopyMenu = () => {
    copyMenuProxyId.value = null
  }

  return {
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
    exportingData,
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
  }
}
