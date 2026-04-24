import { ref, reactive, getCurrentScope, onScopeDispose, toRaw } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import type { BasePaginationResponse, FetchOptions } from '@/types'
import { getPersistedPageSize, setPersistedPageSize } from './usePersistedPageSize'
import { isAbortError } from '@/utils/requestError'

interface PaginationState {
  page: number
  page_size: number
  total: number
  pages: number
}

interface PaginationBinding {
  page: number
  page_size: number
  total: number
  pages?: number
}

interface TableLoaderOptions<T, P> {
  fetchFn: (page: number, pageSize: number, params: P, options?: FetchOptions) => Promise<BasePaginationResponse<T>>
  initialParams?: P
  pagination?: PaginationBinding
  pageSize?: number
  debounceMs?: number
  onError?: (error: unknown) => void
  onLoaded?: (response: BasePaginationResponse<T>) => void
  syncPaginationFromResponse?: boolean
  clampPageChange?: boolean
}

/**
 * 通用表格数据加载 Composable
 * 统一处理分页、筛选、搜索防抖和请求取消
 */
export function useTableLoader<T, P extends Record<string, unknown>>(
  options: TableLoaderOptions<T, P>
) {
  const { fetchFn, initialParams, pageSize, debounceMs = 300 } = options
  const shouldClampPageChange = options.clampPageChange ?? true

  const items = ref<T[]>([])
  const loading = ref(false)
  const params = reactive<P>({ ...(initialParams || {}) } as P)
  const pagination = (options.pagination ?? reactive<PaginationState>({
    page: 1,
    page_size: pageSize ?? getPersistedPageSize(),
    total: 0,
    pages: 0
  })) as PaginationState
  if (typeof pagination.pages !== 'number') {
    pagination.pages = 0
  }

  let abortController: AbortController | null = null

  const load = async () => {
    if (abortController) {
      abortController.abort()
    }
    const currentController = new AbortController()
    abortController = currentController
    loading.value = true

    try {
      const response = await fetchFn(
        pagination.page,
        pagination.page_size,
        toRaw(params) as P,
        { signal: currentController.signal }
      )

      if (abortController !== currentController || currentController.signal.aborted) {
        return
      }

      items.value = response.items || []
      pagination.total = response.total || 0
      pagination.pages = response.pages || 0
      if (options.syncPaginationFromResponse) {
        pagination.page = response.page || pagination.page
        pagination.page_size = response.page_size || pagination.page_size
      }
      options.onLoaded?.(response)
    } catch (error) {
      if (abortController !== currentController || currentController.signal.aborted) {
        return
      }
      if (!isAbortError(error)) {
        console.error('Table load error:', error)
        if (options.onError) {
          options.onError(error)
          return
        }
        throw error
      }
    } finally {
      if (abortController === currentController) {
        loading.value = false
        abortController = null
      }
    }
  }

  const reload = () => {
    pagination.page = 1
    return load()
  }

  const debouncedReload = useDebounceFn(reload, debounceMs)

  const handlePageChange = (page: number) => {
    pagination.page = shouldClampPageChange
      ? Math.max(1, Math.min(page, pagination.pages || 1))
      : Math.max(1, page)
    return load()
  }

  const handlePageSizeChange = (size: number) => {
    pagination.page_size = size
    pagination.page = 1
    setPersistedPageSize(size)
    return load()
  }

  const cancelDebouncedReload = () => {
    const cancelableDebouncedReload = debouncedReload as typeof debouncedReload & {
      cancel?: () => void
    }
    cancelableDebouncedReload.cancel?.()
  }

  if (getCurrentScope()) {
    onScopeDispose(() => {
      abortController?.abort()
      cancelDebouncedReload()
    })
  }

  const dispose = () => {
    abortController?.abort()
    abortController = null
    cancelDebouncedReload()
  }

  return {
    items,
    loading,
    params,
    pagination,
    load,
    reload,
    debouncedReload,
    handlePageChange,
    handlePageSizeChange,
    dispose
  }
}
