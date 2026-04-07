import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { useTableLoader } from '@/composables/useTableLoader'
import type { PromoCode } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  buildPromoCodeListFilters,
  type PromoCodeFiltersState
} from './promoCodeForm'

interface PromoCodesViewDataOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  copyToClipboard: (text: string, successMessage?: string) => Promise<boolean>
}

export function usePromoCodesViewData(options: PromoCodesViewDataOptions) {
  const searchQuery = ref('')
  const copiedCode = ref<string | null>(null)
  const filters = reactive<PromoCodeFiltersState>({
    status: ''
  })
  const {
    items: codes,
    loading,
    pagination,
    load: loadCodes,
    debouncedReload,
    handlePageChange,
    handlePageSizeChange,
    dispose: disposeLoader
  } = useTableLoader<PromoCode, Record<string, never>>({
    fetchFn: (page, pageSize, _params, requestOptions) =>
      adminAPI.promo.list(
        page,
        pageSize,
        buildPromoCodeListFilters(filters, searchQuery.value),
        requestOptions
      ),
    onError: (error) => {
      options.showError(resolveRequestErrorMessage(error, options.t('admin.promo.failedToLoad')))
    },
    syncPaginationFromResponse: true,
    clampPageChange: false
  })
  let copiedCodeTimeout: ReturnType<typeof setTimeout> | null = null

  const handleSearch = () => {
    void debouncedReload()
  }

  const handleCopyCode = async (text: string) => {
    const success = await options.copyToClipboard(text, options.t('admin.promo.copied'))
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

  const dispose = () => {
    disposeLoader()
    if (copiedCodeTimeout) {
      clearTimeout(copiedCodeTimeout)
    }
  }

  return {
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
  }
}
