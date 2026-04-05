import { computed, nextTick, onMounted, onUnmounted, ref, type Ref } from 'vue'
import {
  buildKeyUsageDateParams,
  buildKeyUsageDateRanges,
  buildKeyUsageDetailRows,
  buildKeyUsageRequestUrl,
  buildKeyUsageRingGridClass,
  buildKeyUsageRingItems,
  buildKeyUsageStatusInfo,
  buildKeyUsageUsageStatCells,
  formatKeyUsageDate,
  formatKeyUsageNumber,
  formatKeyUsageResetTime,
  formatKeyUsageUsd,
  resolveKeyUsageQueryErrorMessage,
  type KeyUsageDateRangeKey,
  type KeyUsageModelStat,
  type KeyUsageResult,
  type KeyUsageRingItem
} from './keyUsageView'

interface KeyUsageViewModelOptions {
  isDark: Ref<boolean>
  locale: Ref<string>
  showError: (message: string) => void
  showInfo: (message: string) => void
  showSuccess: (message: string) => void
  t: (key: string, values?: Record<string, unknown>) => string
}

export function useKeyUsageViewModel(options: KeyUsageViewModelOptions) {
  const apiKey = ref('')
  const keyVisible = ref(false)
  const isQuerying = ref(false)
  const showResults = ref(false)
  const showLoading = ref(false)
  const showDatePicker = ref(false)
  const resultData = ref<KeyUsageResult | null>(null)
  const now = ref(new Date())
  const currentRange = ref<KeyUsageDateRangeKey>('today')
  const customStartDate = ref('')
  const customEndDate = ref('')
  const ringAnimated = ref(false)
  const displayPcts = ref<number[]>([])
  let resetTimer: ReturnType<typeof setInterval> | null = null

  const dateRanges = computed(() => buildKeyUsageDateRanges(options.t))
  const ringTrackColor = computed(() => (options.isDark.value ? '#222222' : '#F0F0EE'))

  const usd = (value: number | null | undefined) => formatKeyUsageUsd(value)
  const fmtNum = (value: number | null | undefined) => formatKeyUsageNumber(value)
  const formatDate = (iso: string | null | undefined) =>
    formatKeyUsageDate(iso, options.locale.value)
  const formatResetTime = (resetAt: string | null | undefined) =>
    formatKeyUsageResetTime(resetAt, now.value, options.t('keyUsage.resetNow'))

  const statusInfo = computed(() => buildKeyUsageStatusInfo(resultData.value, options.t))

  const ringItems = computed<KeyUsageRingItem[]>(() =>
    buildKeyUsageRingItems(resultData.value, {
      t: options.t,
      usd
    })
  )

  const ringGridClass = computed(() => buildKeyUsageRingGridClass(ringItems.value.length))

  const detailRows = computed(() =>
    buildKeyUsageDetailRows(resultData.value, {
      t: options.t,
      locale: options.locale.value,
      usd,
      formatDate,
      formatResetTime
    })
  )

  const usageStatCells = computed(() =>
    buildKeyUsageUsageStatCells(resultData.value, {
      t: options.t,
      fmtNum,
      usd
    })
  )

  const modelStats = computed<KeyUsageModelStat[]>(() => resultData.value?.model_stats || [])

  const modelStatsLabels = computed(() => ({
    model: options.t('keyUsage.model'),
    requests: options.t('keyUsage.requests'),
    inputTokens: options.t('keyUsage.inputTokens'),
    outputTokens: options.t('keyUsage.outputTokens'),
    cacheCreationTokens: options.t('keyUsage.cacheCreationTokens'),
    cacheReadTokens: options.t('keyUsage.cacheReadTokens'),
    totalTokens: options.t('keyUsage.totalTokens'),
    cost: options.t('keyUsage.cost')
  }))

  const getDateParams = () =>
    buildKeyUsageDateParams({
      range: currentRange.value,
      customStartDate: customStartDate.value,
      customEndDate: customEndDate.value
    })

  const triggerRingAnimation = (items: KeyUsageRingItem[]) => {
    ringAnimated.value = false
    displayPcts.value = items.map(() => 0)

    requestAnimationFrame(() => {
      window.setTimeout(() => {
        ringAnimated.value = true

        const duration = 1000
        const startTime = performance.now()
        const targets = items.map((item) => (item.isBalance ? 0 : item.pct))

        const tick = () => {
          const elapsed = performance.now() - startTime
          const progress = Math.min(elapsed / duration, 1)
          const ease = 1 - Math.pow(1 - progress, 3)
          displayPcts.value = targets.map((target) => Math.round(ease * target))

          if (progress < 1) {
            requestAnimationFrame(tick)
          }
        }

        requestAnimationFrame(tick)
      }, 50)
    })
  }

  const fetchUsage = async (key: string): Promise<KeyUsageResult> => {
    const response = await fetch(buildKeyUsageRequestUrl(getDateParams()), {
      headers: { Authorization: `Bearer ${key}` }
    })

    if (!response.ok) {
      const body = await response.json().catch(() => null)
      throw new Error(
        resolveKeyUsageQueryErrorMessage(body, response.status, options.t('keyUsage.queryFailed'))
      )
    }

    return (await response.json()) as KeyUsageResult
  }

  const queryKey = async () => {
    if (isQuerying.value) {
      return
    }

    const trimmedKey = apiKey.value.trim()
    if (!trimmedKey) {
      options.showInfo(options.t('keyUsage.enterApiKey'))
      return
    }

    isQuerying.value = true
    showResults.value = true
    showLoading.value = true
    resultData.value = null

    try {
      resultData.value = await fetchUsage(trimmedKey)
      showLoading.value = false
      showDatePicker.value = true

      void nextTick(() => {
        triggerRingAnimation(ringItems.value)
      })

      options.showSuccess(options.t('keyUsage.querySuccess'))
    } catch (error) {
      showResults.value = false
      showLoading.value = false
      options.showError((error as Error).message || options.t('keyUsage.queryFailedRetry'))
    } finally {
      isQuerying.value = false
    }
  }

  const setDateRange = (range: KeyUsageDateRangeKey) => {
    currentRange.value = range
    if (range !== 'custom') {
      void queryKey()
    }
  }

  onMounted(() => {
    resetTimer = setInterval(() => {
      now.value = new Date()
    }, 60000)
  })

  onUnmounted(() => {
    if (resetTimer) {
      clearInterval(resetTimer)
    }
  })

  return {
    apiKey,
    currentRange,
    customEndDate,
    customStartDate,
    dateRanges,
    detailRows,
    displayPcts,
    fmtNum,
    formatResetTime,
    isQuerying,
    keyVisible,
    modelStats,
    modelStatsLabels,
    queryKey,
    resultData,
    ringAnimated,
    ringGridClass,
    ringItems,
    ringTrackColor,
    setDateRange,
    showDatePicker,
    showLoading,
    showResults,
    statusInfo,
    usageStatCells,
    usd
  }
}
