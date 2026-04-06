import type { Column } from '@/components/common/types'
import { formatDateTime } from '@/utils/format'
import type { Account } from '@/types'
import type { AccountListFilters } from './accountsList'

export function buildAccountAutoRefreshIntervalLabel(
  seconds: number,
  t: (key: string) => string
): string {
  if (seconds === 5) return t('admin.accounts.refreshInterval5s')
  if (seconds === 10) return t('admin.accounts.refreshInterval10s')
  if (seconds === 15) return t('admin.accounts.refreshInterval15s')
  if (seconds === 30) return t('admin.accounts.refreshInterval30s')
  return `${seconds}s`
}

export function buildAccountTableColumns(
  isSimpleMode: boolean,
  t: (key: string) => string
): Column[] {
  const columns: Column[] = [
    { key: 'select', label: '', sortable: false },
    { key: 'name', label: t('admin.accounts.columns.name'), sortable: true },
    { key: 'platform_type', label: t('admin.accounts.columns.platformType'), sortable: false },
    { key: 'capacity', label: t('admin.accounts.columns.capacity'), sortable: false },
    { key: 'status', label: t('admin.accounts.columns.status'), sortable: true },
    { key: 'schedulable', label: t('admin.accounts.columns.schedulable'), sortable: true },
    { key: 'today_stats', label: t('admin.accounts.columns.todayStats'), sortable: false }
  ]

  if (!isSimpleMode) {
    columns.push({ key: 'groups', label: t('admin.accounts.columns.groups'), sortable: false })
  }

  columns.push(
    { key: 'usage', label: t('admin.accounts.columns.usageWindows'), sortable: false },
    { key: 'proxy', label: t('admin.accounts.columns.proxy'), sortable: false },
    { key: 'priority', label: t('admin.accounts.columns.priority'), sortable: true },
    {
      key: 'rate_multiplier',
      label: t('admin.accounts.columns.billingRateMultiplier'),
      sortable: true
    },
    { key: 'last_used_at', label: t('admin.accounts.columns.lastUsed'), sortable: true },
    { key: 'expires_at', label: t('admin.accounts.columns.expiresAt'), sortable: true },
    { key: 'notes', label: t('admin.accounts.columns.notes'), sortable: false },
    { key: 'actions', label: t('admin.accounts.columns.actions'), sortable: false }
  )

  return columns
}

export function formatAccountExportTimestamp(date: Date = new Date()): string {
  const pad2 = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}${pad2(date.getMonth() + 1)}${pad2(date.getDate())}${pad2(date.getHours())}${pad2(date.getMinutes())}${pad2(date.getSeconds())}`
}

export function buildAccountExportFilename(date: Date = new Date()): string {
  return `sub2api-account-${formatAccountExportTimestamp(date)}.json`
}

export function buildAccountExportRequest(
  selectedIds: number[],
  includeProxies: boolean,
  filters: Pick<AccountListFilters, 'platform' | 'type' | 'status' | 'search'>
):
  | { ids: number[]; includeProxies: boolean }
  | {
      includeProxies: boolean
      filters: Pick<AccountListFilters, 'platform' | 'type' | 'status' | 'search'>
    } {
  if (selectedIds.length > 0) {
    return {
      ids: selectedIds,
      includeProxies
    }
  }

  return {
    includeProxies,
    filters
  }
}

function readAccountExtraRecord(account: Pick<Account, 'platform' | 'extra'>): Record<string, unknown> | null {
  if (account.platform !== 'antigravity') {
    return null
  }

  const extra = account.extra
  if (!extra || typeof extra !== 'object') {
    return null
  }

  return extra as Record<string, unknown>
}

export function getAccountAntigravityTier(account: Pick<Account, 'platform' | 'extra'>): string | null {
  const extra = readAccountExtraRecord(account)
  if (!extra) {
    return null
  }

  const loadCodeAssist = extra.load_code_assist
  if (!loadCodeAssist || typeof loadCodeAssist !== 'object') {
    return null
  }

  const paidTier = (loadCodeAssist as Record<string, unknown>).paidTier
  if (paidTier && typeof paidTier === 'object' && typeof (paidTier as Record<string, unknown>).id === 'string') {
    return (paidTier as Record<string, unknown>).id as string
  }

  const currentTier = (loadCodeAssist as Record<string, unknown>).currentTier
  if (
    currentTier &&
    typeof currentTier === 'object' &&
    typeof (currentTier as Record<string, unknown>).id === 'string'
  ) {
    return (currentTier as Record<string, unknown>).id as string
  }

  return null
}

export function getAccountAntigravityTierLabel(
  account: Pick<Account, 'platform' | 'extra'>,
  t: (key: string) => string
): string | null {
  const tier = getAccountAntigravityTier(account)
  switch (tier) {
    case 'free-tier':
      return t('admin.accounts.tier.free')
    case 'g1-pro-tier':
      return t('admin.accounts.tier.pro')
    case 'g1-ultra-tier':
      return t('admin.accounts.tier.ultra')
    default:
      return null
  }
}

export function getAccountAntigravityTierClass(account: Pick<Account, 'platform' | 'extra'>): string {
  const tier = getAccountAntigravityTier(account)
  switch (tier) {
    case 'free-tier':
      return 'theme-chip--neutral'
    case 'g1-pro-tier':
      return 'theme-chip--info'
    case 'g1-ultra-tier':
      return 'theme-chip--brand-purple'
    default:
      return ''
  }
}

export function formatAccountExpiresAt(value: number | null): string {
  if (!value) {
    return '-'
  }

  return formatDateTime(
    new Date(value * 1000),
    {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    },
    'sv-SE'
  )
}

export function isAccountExpired(value: number | null, nowMs: number = Date.now()): boolean {
  if (!value) {
    return false
  }

  return value * 1000 <= nowMs
}
