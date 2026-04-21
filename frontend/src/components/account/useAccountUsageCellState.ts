import { computed, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Account, AccountUsageInfo, GeminiCredentials, WindowStats } from '@/types'
import { buildAccountUsageRefreshKey } from '@/utils/accountUsageRefresh'
import { formatCompactNumber } from '@/utils/format'
import { getGrokAccountRuntime } from '@/utils/grokAccountRuntime'

export interface AccountUsageCellProps {
  account: Account
  todayStats?: WindowStats | null
  todayStatsLoading?: boolean
  manualRefreshToken?: number
}

interface AntigravityUsageResult {
  utilization: number
  resetTime: string | null
}

interface QuotaBarInfo {
  utilization: number
  resetsAt: string | null
}

type Translate = (key: string) => string

interface AccountUsageCellStateOptions {
  rootRef?: Ref<HTMLElement | null>
}

export function useAccountUsageCellState(
  props: Required<AccountUsageCellProps>,
  t: Translate,
  options: AccountUsageCellStateOptions = {}
) {
  const desktopViewportQuery = '(min-width: 1024px)'
  const loading = ref(false)
  const activeQueryLoading = ref(false)
  const error = ref<string | null>(null)
  const usageInfo = ref<AccountUsageInfo | null>(null)
  const linkCopied = ref(false)
  const isDesktopViewport = ref(
    typeof window === 'undefined' || typeof window.matchMedia !== 'function'
      ? true
      : window.matchMedia(desktopViewportQuery).matches
  )
  const hasEnteredViewport = ref(isDesktopViewport.value)
  const pendingAutoLoad = ref(false)
  const pendingAutoLoadSource = ref<'passive' | 'active' | undefined>(undefined)

  let desktopViewportMediaQuery: MediaQueryList | null = null
  let desktopViewportListener: ((event: MediaQueryListEvent) => void) | null = null
  let visibilityObserver: IntersectionObserver | null = null
  let latestUsageWriteSeq = 0

  const showUsageWindows = computed(() => {
    if (props.account.platform === 'gemini') return true
    if (props.account.platform === 'grok') return props.account.type === 'session'
    return props.account.type === 'oauth' || props.account.type === 'setup-token'
  })

  const shouldFetchUsage = computed(() => {
    if (props.account.platform === 'anthropic') {
      return props.account.type === 'oauth' || props.account.type === 'setup-token'
    }
    if (props.account.platform === 'gemini') {
      return true
    }
    if (props.account.platform === 'antigravity') {
      return props.account.type === 'oauth'
    }
    if (props.account.platform === 'openai') {
      return props.account.type === 'oauth'
    }
    if (props.account.platform === 'grok') {
      return props.account.type === 'session'
    }
    return false
  })

  const grokUsageBars = computed(() => {
    if (props.account.platform !== 'grok' || props.account.type !== 'session') return []

    const bars: Array<{
      key: string
      name: string
      utilization: number
      resetsAt: string | null
      windowStats?: WindowStats | null
      color: 'indigo' | 'emerald' | 'purple' | 'amber'
    }> = []

    const quotaWindows = usageInfo.value?.grok_quota_windows
    const orderedNames = ['auto', 'fast', 'expert', 'heavy']

    if (quotaWindows && Object.keys(quotaWindows).length > 0) {
      for (const name of orderedNames) {
        const window = quotaWindows[name]
        if (!window) continue

        bars.push({
          key: name,
          name,
          utilization: window.utilization,
          resetsAt: window.resets_at,
          windowStats: window.window_stats,
          color:
            name === 'auto'
              ? 'indigo'
              : name === 'fast'
                ? 'emerald'
                : name === 'expert'
                  ? 'purple'
                  : 'amber'
        })
      }
      return bars
    }

    const grokRuntime = getGrokAccountRuntime(props.account)
    const runtimeWindows = grokRuntime?.quotaWindows ?? []
    for (const name of orderedNames) {
      const window = runtimeWindows.find((item) => item.name === name)
      if (!window || (!window.hasSignal && window.total <= 0)) continue

      const total = window.total > 0 ? window.total : 0
      const remaining = Math.max(window.remaining, 0)
      const used = total > remaining ? total - remaining : 0
      const utilization = total > 0 ? (used / total) * 100 : 0

      bars.push({
        key: name,
        name,
        utilization,
        resetsAt: window.resetAt,
        color:
          name === 'auto'
            ? 'indigo'
            : name === 'fast'
              ? 'emerald'
              : name === 'expert'
                ? 'purple'
                : 'amber'
      })
    }

    return bars
  })

  const geminiUsageAvailable = computed(() => {
    return (
      !!usageInfo.value?.gemini_shared_daily ||
      !!usageInfo.value?.gemini_pro_daily ||
      !!usageInfo.value?.gemini_flash_daily ||
      !!usageInfo.value?.gemini_shared_minute ||
      !!usageInfo.value?.gemini_pro_minute ||
      !!usageInfo.value?.gemini_flash_minute
    )
  })

  const hasOpenAIUsageFallback = computed(() => {
    if (props.account.platform !== 'openai' || props.account.type !== 'oauth') return false
    return !!usageInfo.value?.five_hour || !!usageInfo.value?.seven_day
  })

  const accountUsageRefreshKey = computed(() => buildAccountUsageRefreshKey(props.account))

  const shouldAutoLoadUsageOnMount = computed(() => shouldFetchUsage.value)
  const shouldLazyLoadOnMobile = computed(() => shouldFetchUsage.value && !isDesktopViewport.value)

  const hasAntigravityQuotaFromAPI = computed(() => {
    return (
      !!usageInfo.value?.antigravity_quota &&
      Object.keys(usageInfo.value.antigravity_quota).length > 0
    )
  })

  const getAntigravityUsageFromAPI = (modelNames: string[]): AntigravityUsageResult | null => {
    const quota = usageInfo.value?.antigravity_quota
    if (!quota) return null

    let maxUtilization = 0
    let earliestReset: string | null = null

    for (const model of modelNames) {
      const modelQuota = quota[model]
      if (!modelQuota) continue

      if (modelQuota.utilization > maxUtilization) {
        maxUtilization = modelQuota.utilization
      }
      if (modelQuota.reset_time) {
        if (!earliestReset || modelQuota.reset_time < earliestReset) {
          earliestReset = modelQuota.reset_time
        }
      }
    }

    if (maxUtilization === 0 && earliestReset === null) {
      const hasAnyData = modelNames.some((model) => quota[model])
      if (!hasAnyData) return null
    }

    return {
      utilization: maxUtilization,
      resetTime: earliestReset
    }
  }

  const antigravity3ProUsageFromAPI = computed(() =>
    getAntigravityUsageFromAPI(['gemini-3-pro-low', 'gemini-3-pro-high', 'gemini-3-pro-preview'])
  )

  const antigravity3FlashUsageFromAPI = computed(() =>
    getAntigravityUsageFromAPI(['gemini-3-flash'])
  )

  const antigravity3ImageUsageFromAPI = computed(() =>
    getAntigravityUsageFromAPI([
      'gemini-2.5-flash-image',
      'gemini-3.1-flash-image',
      'gemini-3-pro-image'
    ])
  )

  const antigravityClaudeUsageFromAPI = computed(() =>
    getAntigravityUsageFromAPI([
      'claude-sonnet-4-5',
      'claude-opus-4-5-thinking',
      'claude-sonnet-4-6',
      'claude-opus-4-6',
      'claude-opus-4-6-thinking'
    ])
  )

  const aiCreditsDisplay = computed(() => {
    const credits = usageInfo.value?.ai_credits
    if (!credits || credits.length === 0) return null
    const total = credits.reduce((sum, credit) => sum + (credit.amount ?? 0), 0)
    if (total <= 0) return null
    return total.toFixed(0)
  })

  const antigravityTier = computed(() => {
    const extra = props.account.extra as Record<string, unknown> | undefined
    if (!extra) return null

    const loadCodeAssist = extra.load_code_assist as Record<string, unknown> | undefined
    if (!loadCodeAssist) return null

    const paidTier = loadCodeAssist.paidTier as Record<string, unknown> | undefined
    if (paidTier && typeof paidTier.id === 'string') {
      return paidTier.id
    }

    const currentTier = loadCodeAssist.currentTier as Record<string, unknown> | undefined
    if (currentTier && typeof currentTier.id === 'string') {
      return currentTier.id
    }

    return null
  })

  const geminiTier = computed(() => {
    if (props.account.platform !== 'gemini') return null
    const creds = props.account.credentials as GeminiCredentials | undefined
    return creds?.tier_id || null
  })

  const geminiOAuthType = computed(() => {
    if (props.account.platform !== 'gemini') return null
    const creds = props.account.credentials as GeminiCredentials | undefined
    return (creds?.oauth_type || '').trim() || null
  })

  const isGeminiCodeAssist = computed(() => {
    if (props.account.platform !== 'gemini') return false
    const creds = props.account.credentials as GeminiCredentials | undefined
    return creds?.oauth_type === 'code_assist' || (!creds?.oauth_type && !!creds?.project_id)
  })

  const geminiChannelShort = computed((): 'ai studio' | 'gcp' | 'google one' | 'client' | null => {
    if (props.account.platform !== 'gemini') return null

    if (props.account.type === 'apikey') return 'ai studio'
    if (geminiOAuthType.value === 'google_one') return 'google one'
    if (isGeminiCodeAssist.value) return 'gcp'
    if (geminiOAuthType.value === 'ai_studio') return 'client'

    return 'ai studio'
  })

  const geminiUserLevel = computed((): string | null => {
    if (props.account.platform !== 'gemini') return null

    const tier = (geminiTier.value || '').toString().trim()
    const tierLower = tier.toLowerCase()
    const tierUpper = tier.toUpperCase()

    if (geminiOAuthType.value === 'google_one') {
      if (tierLower === 'google_one_free') return 'free'
      if (tierLower === 'google_ai_pro') return 'pro'
      if (tierLower === 'google_ai_ultra') return 'ultra'
      if (
        tierUpper === 'AI_PREMIUM' ||
        tierUpper === 'GOOGLE_ONE_STANDARD'
      ) {
        return 'pro'
      }
      if (tierUpper === 'GOOGLE_ONE_UNLIMITED') return 'ultra'
      if (
        tierUpper === 'FREE' ||
        tierUpper === 'GOOGLE_ONE_BASIC' ||
        tierUpper === 'GOOGLE_ONE_UNKNOWN' ||
        tierUpper === ''
      ) {
        return 'free'
      }

      return null
    }

    if (isGeminiCodeAssist.value) {
      if (tierLower === 'gcp_enterprise') return 'enterprise'
      if (tierLower === 'gcp_standard') return 'standard'
      if (tierUpper.includes('ULTRA') || tierUpper.includes('ENTERPRISE')) return 'enterprise'
      return 'standard'
    }

    if (props.account.type === 'apikey' || geminiOAuthType.value === 'ai_studio') {
      if (tierLower === 'aistudio_paid') return 'paid'
      if (tierLower === 'aistudio_free') return 'free'
      if (tierUpper.includes('PAID') || tierUpper.includes('PAYG') || tierUpper.includes('PAY')) {
        return 'paid'
      }
      if (tierUpper.includes('FREE')) return 'free'
      if (props.account.type === 'apikey') return 'free'
      return null
    }

    return null
  })

  const geminiAuthTypeLabel = computed(() => {
    if (props.account.platform !== 'gemini') return null
    if (!geminiChannelShort.value) return null
    return geminiUserLevel.value
      ? `${geminiChannelShort.value} ${geminiUserLevel.value}`
      : geminiChannelShort.value
  })

  const geminiTierClass = computed(() => {
    const channel = geminiChannelShort.value
    const level = geminiUserLevel.value

    if (channel === 'client' || channel === 'ai studio') {
      return 'theme-chip theme-chip--compact theme-chip--info'
    }

    if (channel === 'google one') {
      if (level === 'ultra') {
        return 'theme-chip theme-chip--compact theme-chip--brand-purple'
      }
      if (level === 'pro') {
        return 'theme-chip theme-chip--compact theme-chip--info'
      }
      return 'theme-chip theme-chip--compact theme-chip--neutral'
    }

    if (channel === 'gcp') {
      if (level === 'enterprise') {
        return 'theme-chip theme-chip--compact theme-chip--brand-purple'
      }
      return 'theme-chip theme-chip--compact theme-chip--info'
    }

    return ''
  })

  const geminiQuotaPolicyChannel = computed(() => {
    if (geminiOAuthType.value === 'google_one') {
      return t('admin.accounts.gemini.quotaPolicy.rows.googleOne.channel')
    }
    if (isGeminiCodeAssist.value) {
      return t('admin.accounts.gemini.quotaPolicy.rows.gcp.channel')
    }
    return t('admin.accounts.gemini.quotaPolicy.rows.aiStudio.channel')
  })

  const geminiQuotaPolicyLimits = computed(() => {
    const tierLower = (geminiTier.value || '').toString().trim().toLowerCase()

    if (geminiOAuthType.value === 'google_one') {
      if (tierLower === 'google_ai_ultra' || geminiUserLevel.value === 'ultra') {
        return t('admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsUltra')
      }
      if (tierLower === 'google_ai_pro' || geminiUserLevel.value === 'pro') {
        return t('admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsPro')
      }
      return t('admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsFree')
    }

    if (isGeminiCodeAssist.value) {
      if (tierLower === 'gcp_enterprise' || geminiUserLevel.value === 'enterprise') {
        return t('admin.accounts.gemini.quotaPolicy.rows.gcp.limitsEnterprise')
      }
      return t('admin.accounts.gemini.quotaPolicy.rows.gcp.limitsStandard')
    }

    if (tierLower === 'aistudio_paid' || geminiUserLevel.value === 'paid') {
      return t('admin.accounts.gemini.quotaPolicy.rows.aiStudio.limitsPaid')
    }
    return t('admin.accounts.gemini.quotaPolicy.rows.aiStudio.limitsFree')
  })

  const geminiQuotaPolicyDocsUrl = computed(() => {
    if (geminiOAuthType.value === 'google_one' || isGeminiCodeAssist.value) {
      return 'https://developers.google.com/gemini-code-assist/resources/quotas'
    }
    return 'https://ai.google.dev/pricing'
  })

  const geminiUsesSharedDaily = computed(() => {
    if (props.account.platform !== 'gemini') return false
    return (
      !!usageInfo.value?.gemini_shared_daily ||
      !!usageInfo.value?.gemini_shared_minute ||
      geminiOAuthType.value === 'google_one' ||
      isGeminiCodeAssist.value
    )
  })

  const geminiUsageBars = computed(() => {
    if (props.account.platform !== 'gemini') return []
    if (!usageInfo.value) return []

    const bars: Array<{
      key: string
      label: string
      utilization: number
      resetsAt: string | null
      windowStats?: WindowStats | null
      color: 'indigo' | 'emerald'
    }> = []

    if (geminiUsesSharedDaily.value) {
      const sharedDaily = usageInfo.value.gemini_shared_daily
      if (sharedDaily) {
        bars.push({
          key: 'shared_daily',
          label: '1d',
          utilization: sharedDaily.utilization,
          resetsAt: sharedDaily.resets_at,
          windowStats: sharedDaily.window_stats,
          color: 'indigo'
        })
      }
      return bars
    }

    const pro = usageInfo.value.gemini_pro_daily
    if (pro) {
      bars.push({
        key: 'pro_daily',
        label: 'pro',
        utilization: pro.utilization,
        resetsAt: pro.resets_at,
        windowStats: pro.window_stats,
        color: 'indigo'
      })
    }

    const flash = usageInfo.value.gemini_flash_daily
    if (flash) {
      bars.push({
        key: 'flash_daily',
        label: 'flash',
        utilization: flash.utilization,
        resetsAt: flash.resets_at,
        windowStats: flash.window_stats,
        color: 'emerald'
      })
    }

    return bars
  })

  const antigravityTierLabel = computed(() => {
    switch (antigravityTier.value) {
      case 'free-tier':
        return t('admin.accounts.tier.free')
      case 'g1-pro-tier':
        return t('admin.accounts.tier.pro')
      case 'g1-ultra-tier':
        return t('admin.accounts.tier.ultra')
      default:
        return null
    }
  })

  const antigravityTierClass = computed(() => {
    switch (antigravityTier.value) {
      case 'free-tier':
        return 'theme-chip theme-chip--compact theme-chip--neutral'
      case 'g1-pro-tier':
        return 'theme-chip theme-chip--compact theme-chip--info'
      case 'g1-ultra-tier':
        return 'theme-chip theme-chip--compact theme-chip--brand-purple'
      default:
        return ''
    }
  })

  const hasIneligibleTiers = computed(() => {
    const extra = props.account.extra as Record<string, unknown> | undefined
    if (!extra) return false

    const loadCodeAssist = extra.load_code_assist as Record<string, unknown> | undefined
    if (!loadCodeAssist) return false

    const ineligibleTiers = loadCodeAssist.ineligibleTiers as unknown[] | undefined
    return Array.isArray(ineligibleTiers) && ineligibleTiers.length > 0
  })

  const isForbidden = computed(() => !!usageInfo.value?.is_forbidden)
  const forbiddenType = computed(() => usageInfo.value?.forbidden_type || 'forbidden')
  const validationURL = computed(() => usageInfo.value?.validation_url || '')
  const needsReauth = computed(() => !!usageInfo.value?.needs_reauth)

  const usageErrorLabel = computed(() => {
    if (usageInfo.value?.error_code === 'rate_limited') return t('admin.accounts.rateLimited')
    return t('admin.accounts.usageError')
  })

  const forbiddenLabel = computed(() => {
    switch (forbiddenType.value) {
      case 'validation':
        return t('admin.accounts.forbiddenValidation')
      case 'violation':
        return t('admin.accounts.forbiddenViolation')
      default:
        return t('admin.accounts.forbidden')
    }
  })

  const forbiddenBadgeClass = computed(() => {
    if (forbiddenType.value === 'validation') {
      return 'theme-chip theme-chip--compact theme-chip--warning'
    }
    return 'theme-chip theme-chip--compact theme-chip--danger'
  })

  const copyValidationURL = async () => {
    if (!validationURL.value) return

    try {
      await navigator.clipboard.writeText(validationURL.value)
      linkCopied.value = true
      setTimeout(() => {
        linkCopied.value = false
      }, 2000)
    } catch {
      // Ignore clipboard failures to avoid adding fallback branches.
    }
  }

  const isAnthropicOAuthOrSetupToken = computed(() => {
    return (
      props.account.platform === 'anthropic' &&
      (props.account.type === 'oauth' || props.account.type === 'setup-token')
    )
  })

  const beginUsageWrite = () => {
    latestUsageWriteSeq += 1
    return latestUsageWriteSeq
  }

  const isLatestUsageWrite = (writeSeq: number, accountId: Account['id']) => {
    return writeSeq === latestUsageWriteSeq && props.account.id === accountId
  }

  const loadUsage = async (source?: 'passive' | 'active') => {
    if (!shouldFetchUsage.value) return

    const writeSeq = beginUsageWrite()
    const accountId = props.account.id
    loading.value = true
    error.value = null

    try {
      const nextUsageInfo = source
        ? await adminAPI.accounts.getUsage(props.account.id, source)
        : await adminAPI.accounts.getUsage(props.account.id)
      if (!isLatestUsageWrite(writeSeq, accountId)) {
        return
      }
      usageInfo.value = nextUsageInfo
    } catch (caughtError) {
      if (!isLatestUsageWrite(writeSeq, accountId)) {
        return
      }
      error.value = t('common.error')
      console.error('Failed to load usage:', caughtError)
    } finally {
      if (props.account.id === accountId) {
        loading.value = false
      }
    }
  }

  const loadActiveUsage = async () => {
    const writeSeq = beginUsageWrite()
    const accountId = props.account.id
    activeQueryLoading.value = true

    try {
      const nextUsageInfo = await adminAPI.accounts.getUsage(props.account.id, 'active')
      if (!isLatestUsageWrite(writeSeq, accountId)) {
        return
      }
      usageInfo.value = nextUsageInfo
    } catch (error) {
      if (!isLatestUsageWrite(writeSeq, accountId)) {
        return
      }
      console.error('Failed to load active usage:', error)
    } finally {
      if (props.account.id === accountId) {
        activeQueryLoading.value = false
      }
    }
  }

  const flushPendingAutoLoad = () => {
    if (!pendingAutoLoad.value) return
    const source = pendingAutoLoadSource.value
    pendingAutoLoad.value = false
    pendingAutoLoadSource.value = undefined
    void loadUsage(source).catch((caughtError) => {
      console.error('Failed to load deferred usage:', caughtError)
    })
  }

  const requestAutoLoad = (source?: 'passive' | 'active') => {
    if (!shouldFetchUsage.value) return
    if (shouldLazyLoadOnMobile.value && !hasEnteredViewport.value) {
      pendingAutoLoad.value = true
      pendingAutoLoadSource.value = source
      return
    }
    void loadUsage(source).catch((caughtError) => {
      console.error('Failed to auto load usage:', caughtError)
    })
  }

  const detachVisibilityObserver = () => {
    visibilityObserver?.disconnect()
    visibilityObserver = null
  }

  const attachVisibilityObserver = () => {
    detachVisibilityObserver()
    if (!shouldLazyLoadOnMobile.value || hasEnteredViewport.value) return

    if (typeof window === 'undefined' || typeof IntersectionObserver === 'undefined') {
      hasEnteredViewport.value = true
      flushPendingAutoLoad()
      return
    }

    const rootEl = options.rootRef?.value
    if (!rootEl) return

    visibilityObserver = new IntersectionObserver((entries) => {
      if (!entries.some((entry) => entry.isIntersecting)) return
      hasEnteredViewport.value = true
      detachVisibilityObserver()
      flushPendingAutoLoad()
    }, {
      root: null,
      rootMargin: '200px 0px',
      threshold: 0.01
    })
    visibilityObserver.observe(rootEl)
  }

  const makeQuotaBar = (used: number, limit: number, startKey?: string): QuotaBarInfo => {
    const utilization = limit > 0 ? (used / limit) * 100 : 0
    let resetsAt: string | null = null

    if (startKey) {
      const extra = props.account.extra as Record<string, unknown> | undefined
      const isDaily = startKey.includes('daily')
      const mode = isDaily
        ? ((extra?.quota_daily_reset_mode as string) || 'rolling')
        : ((extra?.quota_weekly_reset_mode as string) || 'rolling')

      if (mode === 'fixed') {
        const resetAtKey = isDaily ? 'quota_daily_reset_at' : 'quota_weekly_reset_at'
        resetsAt = (extra?.[resetAtKey] as string) || null
      } else {
        const start = extra?.[startKey] as string | undefined
        if (start) {
          const startDate = new Date(start)
          const periodMs = isDaily ? 86400000 : 7 * 86400000
          resetsAt = new Date(startDate.getTime() + periodMs).toISOString()
        }
      }
    }

    return { utilization, resetsAt }
  }

  const hasApiKeyQuota = computed(() => {
    if (props.account.type !== 'apikey' && props.account.type !== 'bedrock') return false
    return (
      (props.account.quota_daily_limit ?? 0) > 0 ||
      (props.account.quota_weekly_limit ?? 0) > 0 ||
      (props.account.quota_limit ?? 0) > 0
    )
  })

  const quotaDailyBar = computed((): QuotaBarInfo | null => {
    const limit = props.account.quota_daily_limit ?? 0
    if (limit <= 0) return null
    return makeQuotaBar(props.account.quota_daily_used ?? 0, limit, 'quota_daily_start')
  })

  const quotaWeeklyBar = computed((): QuotaBarInfo | null => {
    const limit = props.account.quota_weekly_limit ?? 0
    if (limit <= 0) return null
    return makeQuotaBar(props.account.quota_weekly_used ?? 0, limit, 'quota_weekly_start')
  })

  const quotaTotalBar = computed((): QuotaBarInfo | null => {
    const limit = props.account.quota_limit ?? 0
    if (limit <= 0) return null
    return makeQuotaBar(props.account.quota_used ?? 0, limit)
  })

  const formatKeyRequests = computed(() => {
    if (!props.todayStats) return ''
    return formatCompactNumber(props.todayStats.requests, { allowBillions: false })
  })

  const formatKeyTokens = computed(() => {
    if (!props.todayStats) return ''
    return formatCompactNumber(props.todayStats.tokens)
  })

  const formatKeyCost = computed(() => {
    if (!props.todayStats) return '0.00'
    return props.todayStats.cost.toFixed(2)
  })

  const formatKeyUserCost = computed(() => {
    if (!props.todayStats || props.todayStats.user_cost == null) return '0.00'
    return props.todayStats.user_cost.toFixed(2)
  })

  onMounted(() => {
    if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
      desktopViewportMediaQuery = window.matchMedia(desktopViewportQuery)
      isDesktopViewport.value = desktopViewportMediaQuery.matches
      hasEnteredViewport.value = desktopViewportMediaQuery.matches
      desktopViewportListener = (event: MediaQueryListEvent) => {
        isDesktopViewport.value = event.matches
      }
      if (typeof desktopViewportMediaQuery.addEventListener === 'function') {
        desktopViewportMediaQuery.addEventListener('change', desktopViewportListener)
      } else {
        desktopViewportMediaQuery.addListener(desktopViewportListener)
      }
    }

    if (!shouldAutoLoadUsageOnMount.value) return

    const source = isAnthropicOAuthOrSetupToken.value ? 'passive' : undefined
    requestAutoLoad(source)
  })

  watch(accountUsageRefreshKey, (nextKey, prevKey) => {
    if (!prevKey || nextKey === prevKey) return
    if (!shouldFetchUsage.value) return

    const source = isAnthropicOAuthOrSetupToken.value ? 'passive' : undefined
    requestAutoLoad(source)
  })

  watch(
    () => props.manualRefreshToken,
    (nextToken, prevToken) => {
      if (nextToken === prevToken) return
      if (!shouldFetchUsage.value) return

      const source = isAnthropicOAuthOrSetupToken.value ? 'passive' : undefined
      requestAutoLoad(source)
    }
  )

  watch(
    [() => options.rootRef?.value ?? null, shouldLazyLoadOnMobile],
    () => {
      if (shouldLazyLoadOnMobile.value) {
        attachVisibilityObserver()
        return
      }
      detachVisibilityObserver()
    },
    { immediate: true, flush: 'post' }
  )

  watch(isDesktopViewport, (isDesktop) => {
    if (isDesktop) {
      detachVisibilityObserver()
      hasEnteredViewport.value = true
      flushPendingAutoLoad()
      return
    }
    hasEnteredViewport.value = false
    attachVisibilityObserver()
  })

  onUnmounted(() => {
    detachVisibilityObserver()
    if (desktopViewportMediaQuery && desktopViewportListener) {
      if (typeof desktopViewportMediaQuery.removeEventListener === 'function') {
        desktopViewportMediaQuery.removeEventListener('change', desktopViewportListener)
      } else {
        desktopViewportMediaQuery.removeListener(desktopViewportListener)
      }
    }
    desktopViewportListener = null
    desktopViewportMediaQuery = null
  })

  return {
    activeQueryLoading,
    aiCreditsDisplay,
    antigravity3FlashUsageFromAPI,
    antigravity3ImageUsageFromAPI,
    antigravity3ProUsageFromAPI,
    antigravityClaudeUsageFromAPI,
    antigravityTierClass,
    antigravityTierLabel,
    copyValidationURL,
    error,
    forbiddenBadgeClass,
    forbiddenLabel,
    formatKeyCost,
    formatKeyRequests,
    formatKeyTokens,
    formatKeyUserCost,
    geminiAuthTypeLabel,
    geminiQuotaPolicyChannel,
    geminiQuotaPolicyDocsUrl,
    geminiQuotaPolicyLimits,
    geminiTierClass,
    geminiUsageAvailable,
    geminiUsageBars,
    grokUsageBars,
    hasAntigravityQuotaFromAPI,
    hasApiKeyQuota,
    hasIneligibleTiers,
    hasOpenAIUsageFallback,
    isForbidden,
    linkCopied,
    loadActiveUsage,
    loading,
    needsReauth,
    quotaDailyBar,
    quotaTotalBar,
    quotaWeeklyBar,
    showUsageWindows,
    usageErrorLabel,
    usageInfo,
    validationURL
  }
}
