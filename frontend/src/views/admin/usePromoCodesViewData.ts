import { reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import type { PromoCode } from '@/types'
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
  const codes = ref<PromoCode[]>([])
  const loading = ref(false)
  const searchQuery = ref('')
  const copiedCode = ref<string | null>(null)
  const filters = reactive<PromoCodeFiltersState>({
    status: ''
  })
  const pagination = reactive({
    page: 1,
    page_size: getPersistedPageSize(),
    total: 0
  })

  let abortController: AbortController | null = null
  let searchTimeout: ReturnType<typeof setTimeout> | null = null
  let copiedCodeTimeout: ReturnType<typeof setTimeout> | null = null

  const loadCodes = async () => {
    abortController?.abort()

    const currentController = new AbortController()
    abortController = currentController
    loading.value = true

    try {
      const response = await adminAPI.promo.list(
        pagination.page,
        pagination.page_size,
        buildPromoCodeListFilters(filters, searchQuery.value),
        {
          signal: currentController.signal
        }
      )
      if (currentController.signal.aborted || abortController !== currentController) {
        return
      }

      codes.value = response.items
      pagination.total = response.total
      pagination.page = response.page
      pagination.page_size = response.page_size
    } catch (error: any) {
      if (
        currentController.signal.aborted ||
        error?.name === 'AbortError' ||
        error?.code === 'ERR_CANCELED'
      ) {
        return
      }
      options.showError(options.t('admin.promo.failedToLoad'))
      console.error('Error loading promo codes:', error)
    } finally {
      if (abortController === currentController) {
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
