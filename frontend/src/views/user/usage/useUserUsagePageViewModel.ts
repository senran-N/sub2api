import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import type { UsageLog, UsageQueryParams } from '@/types'
import { buildUserUsageApiKeyOptions } from '../userUsageView'
import {
  applyUserUsageDateRange,
  buildDefaultUserUsageFilters,
  buildResetUserUsageState,
  getLast7DaysUserUsageRange
} from '../userUsageViewState'
import { useUserUsageViewData } from '../useUserUsageViewData'
import { useUserUsageHoverTooltip } from './useUserUsageHoverTooltip'

export function useUserUsagePageViewModel() {
  const { t } = useI18n()
  const appStore = useAppStore()

  const {
    visible: tooltipVisible,
    position: tooltipPosition,
    data: tooltipData,
    show: showTooltip,
    hide: hideTooltip
  } = useUserUsageHoverTooltip<UsageLog>()
  const {
    visible: tokenTooltipVisible,
    position: tokenTooltipPosition,
    data: tokenTooltipData,
    show: showTokenTooltip,
    hide: hideTokenTooltip
  } = useUserUsageHoverTooltip<UsageLog>()

  const columns = computed<Column[]>(() => [
    { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
    { key: 'model', label: t('usage.model'), sortable: true },
    { key: 'reasoning_effort', label: t('usage.reasoningEffort'), sortable: false },
    { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
    { key: 'stream', label: t('usage.type'), sortable: false },
    { key: 'tokens', label: t('usage.tokens'), sortable: false },
    { key: 'cost', label: t('usage.cost'), sortable: false },
    { key: 'first_token', label: t('usage.firstToken'), sortable: false },
    { key: 'duration', label: t('usage.duration'), sortable: false },
    { key: 'created_at', label: t('usage.time'), sortable: true },
    { key: 'user_agent', label: t('usage.userAgent'), sortable: false }
  ])

  const defaultRange = getLast7DaysUserUsageRange()
  const startDate = ref(defaultRange.startDate)
  const endDate = ref(defaultRange.endDate)
  const filters = ref<UsageQueryParams>(buildDefaultUserUsageFilters(defaultRange))

  const pagination = reactive({
    page: 1,
    page_size: getPersistedPageSize(),
    total: 0,
    pages: 0
  })

  const {
    usageStats,
    usageLogs,
    apiKeys,
    loading,
    exporting,
    applyFilters,
    handlePageChange,
    handlePageSizeChange,
    exportToCSV,
    loadInitialData,
    dispose
  } = useUserUsageViewData({
    filters,
    startDate,
    endDate,
    pagination,
    showError: appStore.showError,
    showWarning: appStore.showWarning,
    showSuccess: appStore.showSuccess,
    showInfo: appStore.showInfo,
    t
  })

  const apiKeyOptions = computed(() =>
    buildUserUsageApiKeyOptions(apiKeys.value, t('usage.allApiKeys'))
  )

  function onDateRangeChange(range: {
    startDate: string
    endDate: string
    preset: string | null
  }) {
    startDate.value = range.startDate
    endDate.value = range.endDate
    filters.value = applyUserUsageDateRange(filters.value, {
      startDate: range.startDate,
      endDate: range.endDate
    })
    applyFilters()
  }

  function resetFilters() {
    const nextState = buildResetUserUsageState()
    startDate.value = nextState.range.startDate
    endDate.value = nextState.range.endDate
    filters.value = nextState.filters
    applyFilters()
  }

  onMounted(() => {
    loadInitialData()
  })

  onUnmounted(() => {
    dispose()
  })

  return {
    columns,
    usageStats,
    usageLogs,
    loading,
    exporting,
    apiKeyOptions,
    startDate,
    endDate,
    filters,
    pagination,
    tooltipVisible,
    tooltipPosition,
    tooltipData,
    showTooltip,
    hideTooltip,
    tokenTooltipVisible,
    tokenTooltipPosition,
    tokenTooltipData,
    showTokenTooltip,
    hideTokenTooltip,
    onDateRangeChange,
    applyFilters,
    resetFilters,
    handlePageChange,
    handlePageSizeChange,
    exportToCSV
  }
}
