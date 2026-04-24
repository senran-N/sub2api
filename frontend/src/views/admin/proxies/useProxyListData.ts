import { watch, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Proxy } from '@/types'
import { useTableLoader } from '@/composables/useTableLoader'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  buildProxyListFilters,
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
  const {
    debouncedReload,
    dispose,
    handlePageChange,
    handlePageSizeChange,
    load: loadProxies,
    loading
  } = useTableLoader<Proxy, Record<string, never>>({
    pagination: options.pagination,
    clampPageChange: false,
    fetchFn: (page, pageSize, _params, fetchOptions) =>
      adminAPI.proxies.list(
        page,
        pageSize,
        buildProxyListFilters(options.filters, options.searchQuery.value),
        fetchOptions
      ),
    onLoaded: (response) => {
      options.proxies.value = response.items
    },
    onError: (error) => {
      options.showError(resolveRequestErrorMessage(error, options.t('admin.proxies.failedToLoad')))
      console.error('Error loading proxies:', error)
    }
  })

  watch(loading, (value) => {
    options.loading.value = value
  }, { immediate: true })

  const cleanup = () => {
    dispose()
  }

  return {
    cleanup,
    handlePageChange,
    handlePageSizeChange,
    handleSearch: debouncedReload,
    loadProxies
  }
}
