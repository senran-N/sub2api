export type KeyUsageDateRangeKey = 'today' | '7d' | '30d' | 'custom'

export interface KeyUsageQuota {
  limit: number
  used: number
  remaining: number
}

export interface KeyUsageRateLimit {
  window: string
  used: number
  limit: number
  reset_at?: string | null
}

export interface KeyUsageSubscription {
  daily_usage_usd: number
  daily_limit_usd: number
  weekly_usage_usd: number
  weekly_limit_usd: number
  monthly_usage_usd: number
  monthly_limit_usd: number
  expires_at?: string | null
}

export interface KeyUsageUsageSummary {
  requests?: number
  input_tokens?: number
  output_tokens?: number
  total_tokens?: number
  cache_creation_tokens?: number
  cache_read_tokens?: number
  actual_cost?: number
}

export interface KeyUsageUsage {
  today?: KeyUsageUsageSummary
  total?: KeyUsageUsageSummary
  rpm?: number
  tpm?: number
  average_duration_ms?: number | null
}

export interface KeyUsageModelStat {
  model?: string | null
  requests?: number
  input_tokens?: number
  output_tokens?: number
  cache_creation_tokens?: number
  cache_read_tokens?: number
  total_tokens?: number
  actual_cost?: number | null
  cost?: number | null
}

export interface KeyUsageResult {
  mode?: string
  isValid?: boolean
  status?: string
  planName?: string
  quota?: KeyUsageQuota | null
  rate_limits?: KeyUsageRateLimit[] | null
  subscription?: KeyUsageSubscription | null
  balance?: number | null
  remaining?: number | null
  expires_at?: string | null
  days_until_expiry?: number | null
  usage?: KeyUsageUsage | null
  model_stats?: KeyUsageModelStat[] | null
}

export interface KeyUsageStatusInfo {
  label: string
  statusText: string
  isActive: boolean
}

export interface KeyUsageRingItem {
  title: string
  pct: number
  amount: string
  isBalance?: boolean
  iconType: 'clock' | 'calendar' | 'dollar'
  resetAt?: string | null
}

export interface KeyUsageDetailRow {
  iconBg: string
  iconColor: string
  iconSvg: string
  label: string
  value: string
  valueClass: string
}

export interface KeyUsageStatCell {
  label: string
  value: string
}

export interface KeyUsageDateRangeOption {
  key: KeyUsageDateRangeKey
  label: string
}

export const KEY_USAGE_RING_CIRCUMFERENCE = 2 * Math.PI * 68

export const KEY_USAGE_RING_GRADIENTS = [
  { from: '#14b8a6', to: '#5eead4' },
  { from: '#6366F1', to: '#A5B4FC' },
  { from: '#10B981', to: '#6EE7B7' },
  { from: '#F59E0B', to: '#FCD34D' }
] as const

const DETAIL_ICON_SHIELD =
  '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>'
const DETAIL_ICON_CALENDAR =
  '<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>'
const DETAIL_ICON_DOLLAR =
  '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>'
const DETAIL_ICON_CHECK = '<polyline points="20 6 9 17 4 12"/>'

export function buildKeyUsageDateRanges(
  t: (key: string) => string
): KeyUsageDateRangeOption[] {
  return [
    { key: 'today', label: t('keyUsage.dateRangeToday') },
    { key: '7d', label: t('keyUsage.dateRange7d') },
    { key: '30d', label: t('keyUsage.dateRange30d') },
    { key: 'custom', label: t('keyUsage.dateRangeCustom') }
  ]
}

function formatIsoDate(date: Date): string {
  return date.toISOString().split('T')[0]
}

export function buildKeyUsageDateParams(options: {
  range: KeyUsageDateRangeKey
  customStartDate: string
  customEndDate: string
  now?: Date
}): string {
  const now = options.now ?? new Date()

  if (options.range === 'custom') {
    if (options.customStartDate && options.customEndDate) {
      return `start_date=${options.customStartDate}&end_date=${options.customEndDate}`
    }

    return ''
  }

  const end = formatIsoDate(now)
  let start = end

  switch (options.range) {
    case '7d':
      start = formatIsoDate(new Date(now.getTime() - 7 * 86400000))
      break
    case '30d':
      start = formatIsoDate(new Date(now.getTime() - 30 * 86400000))
      break
    case 'today':
    default:
      start = end
      break
  }

  return `start_date=${start}&end_date=${end}`
}

export function buildKeyUsageRequestUrl(dateParams: string): string {
  return `/v1/usage${dateParams ? `?${dateParams}` : ''}`
}

export function formatKeyUsageUsd(value: number | null | undefined): string {
  if (value == null || value < 0) {
    return '-'
  }

  return `$${Number(value).toFixed(2)}`
}

export function formatKeyUsageNumber(value: number | null | undefined): string {
  if (value == null) {
    return '-'
  }

  return value.toLocaleString()
}

export function formatKeyUsageDate(
  iso: string | null | undefined,
  locale: string
): string {
  if (!iso) {
    return '-'
  }

  const date = new Date(iso)
  const targetLocale = locale === 'zh' ? 'zh-CN' : 'en-US'

  return date.toLocaleDateString(targetLocale, {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

export function formatKeyUsageResetTime(
  resetAt: string | null | undefined,
  now: Date,
  resetNowLabel: string
): string {
  if (!resetAt) {
    return ''
  }

  const diff = new Date(resetAt).getTime() - now.getTime()
  if (diff <= 0) {
    return resetNowLabel
  }

  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)

  if (days > 0) {
    return `${days}d ${hours}h`
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  return `${minutes}m`
}

export function resolveKeyUsageQueryErrorMessage(
  body: { error?: { message?: string }; message?: string } | null,
  status: number,
  fallback: string
): string {
  return body?.error?.message || body?.message || `${fallback} (${status})`
}

export function buildKeyUsageStatusInfo(
  data: KeyUsageResult | null,
  t: (key: string) => string
): KeyUsageStatusInfo | null {
  if (data == null) {
    return null
  }

  if (data.mode === 'quota_limited') {
    const isValid = data.isValid !== false
    const statusMap: Record<string, string> = {
      active: 'Active',
      quota_exhausted: 'Quota Exhausted',
      expired: 'Expired'
    }

    return {
      label: t('keyUsage.quotaMode'),
      statusText: statusMap[data.status ?? ''] || data.status || 'Unknown',
      isActive: isValid && data.status === 'active'
    }
  }

  return {
    label: data.planName || t('keyUsage.walletBalance'),
    statusText: 'Active',
    isActive: true
  }
}

export function buildKeyUsageRingItems(
  data: KeyUsageResult | null,
  options: {
    t: (key: string) => string
    usd: (value: number | null | undefined) => string
  }
): KeyUsageRingItem[] {
  if (data == null) {
    return []
  }

  const items: KeyUsageRingItem[] = []

  if (data.mode === 'quota_limited') {
    if (data.quota) {
      const pct =
        data.quota.limit > 0
          ? Math.min(Math.round((data.quota.used / data.quota.limit) * 100), 100)
          : 0
      items.push({
        title: options.t('keyUsage.totalQuota'),
        pct,
        amount: `${options.usd(data.quota.used)} / ${options.usd(data.quota.limit)}`,
        iconType: 'dollar'
      })
    }

    if (data.rate_limits) {
      const windowLabels: Record<string, string> = {
        '5h': options.t('keyUsage.limit5h'),
        '1d': options.t('keyUsage.limitDaily'),
        '7d': options.t('keyUsage.limit7d')
      }
      const windowIcons: Record<string, 'clock' | 'calendar'> = {
        '5h': 'clock',
        '1d': 'calendar',
        '7d': 'calendar'
      }

      for (const rateLimit of data.rate_limits) {
        const pct =
          rateLimit.limit > 0
            ? Math.min(Math.round((rateLimit.used / rateLimit.limit) * 100), 100)
            : 0
        items.push({
          title: windowLabels[rateLimit.window] || rateLimit.window,
          pct,
          amount: `${options.usd(rateLimit.used)} / ${options.usd(rateLimit.limit)}`,
          iconType: windowIcons[rateLimit.window] || 'clock',
          resetAt: rateLimit.reset_at
        })
      }
    }

    return items
  }

  if (data.subscription) {
    const limits = [
      {
        label: options.t('keyUsage.limitDaily'),
        usage: data.subscription.daily_usage_usd,
        limit: data.subscription.daily_limit_usd
      },
      {
        label: options.t('keyUsage.limitWeekly'),
        usage: data.subscription.weekly_usage_usd,
        limit: data.subscription.weekly_limit_usd
      },
      {
        label: options.t('keyUsage.limitMonthly'),
        usage: data.subscription.monthly_usage_usd,
        limit: data.subscription.monthly_limit_usd
      }
    ]

    for (const limit of limits) {
      if (limit.limit != null && limit.limit > 0) {
        items.push({
          title: limit.label,
          pct: Math.min(Math.round((limit.usage / limit.limit) * 100), 100),
          amount: `${options.usd(limit.usage)} / ${options.usd(limit.limit)}`,
          iconType: 'calendar'
        })
      }
    }
  }

  if (!data.subscription && data.balance != null) {
    items.push({
      title: options.t('keyUsage.walletBalance'),
      pct: 0,
      amount: options.usd(data.balance),
      isBalance: true,
      iconType: 'dollar'
    })
  }

  return items
}

export function buildKeyUsageRingGridClass(length: number): string {
  if (length === 1) {
    return 'grid grid-cols-1 max-w-md mx-auto gap-6'
  }
  if (length === 2) {
    return 'grid grid-cols-1 md:grid-cols-2 gap-6'
  }

  return 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6'
}

function getKeyUsageColor(pct: number): string {
  if (pct > 90) {
    return 'text-rose-500'
  }
  if (pct > 70) {
    return 'text-amber-500'
  }

  return 'text-emerald-500'
}

export function buildKeyUsageDetailRows(
  data: KeyUsageResult | null,
  options: {
    t: (key: string, values?: Record<string, unknown>) => string
    locale: string
    usd: (value: number | null | undefined) => string
    formatDate: (iso: string | null | undefined) => string
    formatResetTime: (resetAt: string | null | undefined) => string
  }
): KeyUsageDetailRow[] {
  if (data == null) {
    return []
  }

  const rows: KeyUsageDetailRow[] = []

  if (data.mode === 'quota_limited') {
    if (data.quota) {
      const remainColor =
        data.quota.remaining <= 0
          ? 'text-rose-500'
          : data.quota.remaining < data.quota.limit * 0.1
            ? 'text-amber-500'
            : 'text-emerald-500'
      rows.push({
        iconBg: 'bg-emerald-500/10',
        iconColor: 'text-emerald-500',
        iconSvg: DETAIL_ICON_SHIELD,
        label: options.t('keyUsage.remainingQuota'),
        value: options.usd(data.quota.remaining),
        valueClass: remainColor
      })
    }

    if (data.expires_at) {
      const daysLeft = data.days_until_expiry
      let expiryText = options.formatDate(data.expires_at)

      if (daysLeft != null) {
        expiryText +=
          daysLeft > 0
            ? ` ${options.t('keyUsage.daysLeft', { days: daysLeft })}`
            : daysLeft === 0
              ? ` ${options.t('keyUsage.todayExpires')}`
              : ''
      }

      rows.push({
        iconBg: 'bg-amber-500/10',
        iconColor: 'text-amber-500',
        iconSvg: DETAIL_ICON_CALENDAR,
        label: options.t('keyUsage.expiresAt'),
        value: expiryText,
        valueClass: ''
      })
    }

    if (data.rate_limits) {
      const windowMap: Record<string, string> = {
        '5h': '5H',
        '1d': options.locale === 'zh' ? '日' : 'D',
        '7d': '7D'
      }

      for (const rateLimit of data.rate_limits) {
        const pct =
          rateLimit.limit > 0 ? (rateLimit.used / rateLimit.limit) * 100 : 0
        let value = `${options.usd(rateLimit.used)} / ${options.usd(rateLimit.limit)}`
        const resetTime = options.formatResetTime(rateLimit.reset_at)

        if (resetTime) {
          value += ` (⟳ ${resetTime})`
        }

        rows.push({
          iconBg: 'bg-primary-500/10',
          iconColor: 'text-primary-500',
          iconSvg: DETAIL_ICON_DOLLAR,
          label: `${options.t('keyUsage.usedQuota')} (${windowMap[rateLimit.window] || rateLimit.window})`,
          value,
          valueClass: getKeyUsageColor(pct)
        })
      }
    }

    return rows
  }

  rows.push({
    iconBg: 'bg-emerald-500/10',
    iconColor: 'text-emerald-500',
    iconSvg: DETAIL_ICON_CHECK,
    label: options.t('keyUsage.subscriptionType'),
    value: data.planName || options.t('keyUsage.walletBalance'),
    valueClass: ''
  })

  if (data.subscription) {
    const subscriptionRows = [
      {
        suffix: options.locale === 'zh' ? '日' : 'D',
        usage: data.subscription.daily_usage_usd,
        limit: data.subscription.daily_limit_usd,
        iconBg: 'bg-primary-500/10',
        iconColor: 'text-primary-500'
      },
      {
        suffix: options.locale === 'zh' ? '周' : 'W',
        usage: data.subscription.weekly_usage_usd,
        limit: data.subscription.weekly_limit_usd,
        iconBg: 'bg-indigo-500/10',
        iconColor: 'text-indigo-500'
      },
      {
        suffix: options.locale === 'zh' ? '月' : 'M',
        usage: data.subscription.monthly_usage_usd,
        limit: data.subscription.monthly_limit_usd,
        iconBg: 'bg-emerald-500/10',
        iconColor: 'text-emerald-500'
      }
    ]

    for (const row of subscriptionRows) {
      if (row.limit > 0) {
        const pct = (row.usage / row.limit) * 100
        rows.push({
          iconBg: row.iconBg,
          iconColor: row.iconColor,
          iconSvg: DETAIL_ICON_DOLLAR,
          label: `${options.t('keyUsage.usedQuota')} (${row.suffix})`,
          value: `${options.usd(row.usage)} / ${options.usd(row.limit)}`,
          valueClass: getKeyUsageColor(pct)
        })
      }
    }

    if (data.subscription.expires_at) {
      rows.push({
        iconBg: 'bg-amber-500/10',
        iconColor: 'text-amber-500',
        iconSvg: DETAIL_ICON_CALENDAR,
        label: options.t('keyUsage.subscriptionExpires'),
        value: options.formatDate(data.subscription.expires_at),
        valueClass: ''
      })
    }
  }

  const remainColor =
    data.remaining != null
      ? data.remaining <= 0
        ? 'text-rose-500'
        : data.remaining < 10
          ? 'text-amber-500'
          : 'text-emerald-500'
      : ''

  rows.push({
    iconBg: 'bg-emerald-500/10',
    iconColor: 'text-emerald-500',
    iconSvg: DETAIL_ICON_SHIELD,
    label: options.t('keyUsage.remainingQuota'),
    value: data.remaining != null ? options.usd(data.remaining) : '-',
    valueClass: remainColor
  })

  return rows
}

export function buildKeyUsageUsageStatCells(
  data: KeyUsageResult | null,
  options: {
    t: (key: string) => string
    fmtNum: (value: number | null | undefined) => string
    usd: (value: number | null | undefined) => string
  }
): KeyUsageStatCell[] {
  const usage = data?.usage
  if (!usage) {
    return []
  }

  const today = usage.today || {}
  const total = usage.total || {}

  return [
    { label: options.t('keyUsage.todayRequests'), value: options.fmtNum(today.requests) },
    { label: options.t('keyUsage.todayInputTokens'), value: options.fmtNum(today.input_tokens) },
    { label: options.t('keyUsage.todayOutputTokens'), value: options.fmtNum(today.output_tokens) },
    { label: options.t('keyUsage.todayTokens'), value: options.fmtNum(today.total_tokens) },
    { label: options.t('keyUsage.todayCacheCreation'), value: options.fmtNum(today.cache_creation_tokens) },
    { label: options.t('keyUsage.todayCacheRead'), value: options.fmtNum(today.cache_read_tokens) },
    { label: options.t('keyUsage.todayCost'), value: options.usd(today.actual_cost) },
    { label: options.t('keyUsage.rpmTpm'), value: `${usage.rpm || 0} / ${usage.tpm || 0}` },
    { label: options.t('keyUsage.totalRequests'), value: options.fmtNum(total.requests) },
    { label: options.t('keyUsage.totalInputTokens'), value: options.fmtNum(total.input_tokens) },
    { label: options.t('keyUsage.totalOutputTokens'), value: options.fmtNum(total.output_tokens) },
    { label: options.t('keyUsage.totalTokensLabel'), value: options.fmtNum(total.total_tokens) },
    { label: options.t('keyUsage.totalCacheCreation'), value: options.fmtNum(total.cache_creation_tokens) },
    { label: options.t('keyUsage.totalCacheRead'), value: options.fmtNum(total.cache_read_tokens) },
    { label: options.t('keyUsage.totalCost'), value: options.usd(total.actual_cost) },
    {
      label: options.t('keyUsage.avgDuration'),
      value:
        usage.average_duration_ms != null
          ? `${Math.round(usage.average_duration_ms)} ms`
          : '-'
    }
  ]
}
