import type { Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Proxy } from '@/types'
import { isAbortError, resolveRequestErrorMessage } from '@/utils/requestError'
import {
  applyProxyPageChange,
  applyProxyPageSizeChange,
  buildProxyListFilters,
  resetProxyListPage,
  type ProxyListFiltersState,
  type ProxyPaginationState
} from './proxyList'

interface ProxyListDataOptions {
  proxies: Ref<Proxy[]>
  loading: Ref<boolean>
  searchQuery: Ref<string>
  filters: ProxyListFiltersState
  pagination: ProxyPaginationState
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
}

export function useProxyListData(options: ProxyListDataOptions) {
  let abortController: AbortController | null = null
  let searchTimeout: ReturnType<typeof setTimeout> | null = null

  const loadProxies = async () => {
    abortController?.abort()

    const currentAbortController = new AbortController()
    abortController = currentAbortController
    options.loading.value = true

    try {
      const response = await adminAPI.proxies.list(
        options.pagination.page,
        options.pagination.page_size,
        buildProxyListFilters(options.filters, options.searchQuery.value),
        { signal: currentAbortController.signal }
      )

      if (currentAbortController.signal.aborted || abortController !== currentAbortController) {
        return
      }

      options.proxies.value = response.items
      options.pagination.total = response.total
      options.pagination.pages = response.pages
    } catch (error) {
      if (isAbortError(error)) {
        return
      }

      options.showError(resolveRequestErrorMessage(error, options.t('admin.proxies.failedToLoad')))
      console.error('Error loading proxies:', error)
    } finally {
      if (abortController === currentAbortController) {
        options.loading.value = false
        abortController = null
      }
    }
  }

  const handleSearch = () => {
    if (searchTimeout) {
      clearTimeout(searchTimeout)
    }

    searchTimeout = setTimeout(() => {
      resetProxyListPage(options.pagination)
      loadProxies()
    }, 300)
  }

  const handlePageChange = (page: number) => {
    applyProxyPageChange(options.pagination, page)
    loadProxies()
  }

  const handlePageSizeChange = (pageSize: number) => {
    applyProxyPageSizeChange(options.pagination, pageSize)
    loadProxies()
  }

  const cleanup = () => {
    if (searchTimeout) {
      clearTimeout(searchTimeout)
      searchTimeout = null
    }

    abortController?.abort()
  }

  return {
    cleanup,
    handlePageChange,
    handlePageSizeChange,
    handleSearch,
    loadProxies
  }
}
