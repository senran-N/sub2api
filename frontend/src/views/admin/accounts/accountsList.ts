import { buildOpenAIUsageRefreshKey } from '@/utils/accountUsageRefresh'
import type { Account, WindowStats } from '@/types'

export const DEFAULT_ACCOUNT_HIDDEN_COLUMNS = [
  'today_stats',
  'proxy',
  'notes',
  'priority',
  'rate_multiplier'
] as const

export const ACCOUNT_HIDDEN_COLUMNS_KEY = 'account-hidden-columns'
export const ACCOUNT_SORT_STORAGE_KEY = 'account-table-sort'
export const ACCOUNT_AUTO_REFRESH_STORAGE_KEY = 'account-auto-refresh'
export const ACCOUNT_AUTO_REFRESH_INTERVALS = [5, 10, 15, 30] as const
export const ACCOUNT_AUTO_REFRESH_SILENT_WINDOW_MS = 15000

export interface BulkSchedulableResult {
  success?: number
  failed?: number
  success_ids?: number[]
  failed_ids?: number[]
  results?: Array<{ account_id: number; success: boolean }>
}

export interface NormalizedBulkSchedulableResult {
  successIds: number[]
  failedIds: number[]
  successCount: number
  failedCount: number
  hasIds: boolean
  hasCounts: boolean
}

export interface IncrementalAccountMergeResult {
  rows: Account[]
  changed: boolean
}

export interface AccountListFilters {
  platform?: string
  type?: string
  status?: string
  privacy_mode?: string
  group?: string
  search?: string
}

export type AccountListQuery = Required<AccountListFilters>

export interface AccountListPagination {
  page: number
  page_size: number
  total: number
  pages: number
}

export interface PatchedAccountListResult {
  accounts: Account[]
  pagination: Pick<AccountListPagination, 'page' | 'total' | 'pages'>
  hasPendingListSync: boolean
  removedAccountId: number | null
  patchedAccount: Account | null
  shouldCloseMenu: boolean
}

export function buildDefaultTodayStats(): WindowStats {
  return {
    requests: 0,
    tokens: 0,
    cost: 0,
    standard_cost: 0,
    user_cost: 0
  }
}

export function buildAccountTodayStatsMap(
  accountIds: number[],
  serverStats: Record<string, WindowStats> = {}
): Record<string, WindowStats> {
  const nextStats: Record<string, WindowStats> = {}

  for (const accountId of accountIds) {
    const key = String(accountId)
    nextStats[key] = serverStats[key] ?? buildDefaultTodayStats()
  }

  return nextStats
}

export function shouldReplaceAutoRefreshAccountRow(current: Account, next: Account): boolean {
  return (
    current.updated_at !== next.updated_at ||
    current.current_concurrency !== next.current_concurrency ||
    current.current_window_cost !== next.current_window_cost ||
    current.active_sessions !== next.active_sessions ||
    current.schedulable !== next.schedulable ||
    current.status !== next.status ||
    current.rate_limit_reset_at !== next.rate_limit_reset_at ||
    current.overload_until !== next.overload_until ||
    current.temp_unschedulable_until !== next.temp_unschedulable_until ||
    buildOpenAIUsageRefreshKey(current) !== buildOpenAIUsageRefreshKey(next)
  )
}

export function mergeIncrementalAccountRows(
  currentRows: Account[],
  nextRows: Account[],
  onReplaced?: (nextAccount: Account) => void
): IncrementalAccountMergeResult {
  const currentById = new Map(currentRows.map((row) => [row.id, row]))
  let changed = nextRows.length !== currentRows.length

  const rows = nextRows.map((nextRow) => {
    const currentRow = currentById.get(nextRow.id)
    if (!currentRow) {
      changed = true
      return nextRow
    }

    if (shouldReplaceAutoRefreshAccountRow(currentRow, nextRow)) {
      changed = true
      onReplaced?.(nextRow)
      return nextRow
    }

    return currentRow
  })

  if (!changed) {
    for (let index = 0; index < rows.length; index += 1) {
      if (rows[index].id !== currentRows[index]?.id) {
        changed = true
        break
      }
    }
  }

  return {
    rows: changed ? rows : currentRows,
    changed
  }
}

export function normalizeBulkSchedulableResult(
  result: BulkSchedulableResult,
  accountIds: number[]
): NormalizedBulkSchedulableResult {
  const responseSuccessIds = Array.isArray(result.success_ids) ? result.success_ids : []
  const responseFailedIds = Array.isArray(result.failed_ids) ? result.failed_ids : []
  if (responseSuccessIds.length > 0 || responseFailedIds.length > 0) {
    return {
      successIds: responseSuccessIds,
      failedIds: responseFailedIds,
      successCount:
        typeof result.success === 'number' ? result.success : responseSuccessIds.length,
      failedCount: typeof result.failed === 'number' ? result.failed : responseFailedIds.length,
      hasIds: true,
      hasCounts: true
    }
  }

  const resultEntries = Array.isArray(result.results) ? result.results : []
  if (resultEntries.length > 0) {
    const successIds = resultEntries
      .filter((entry) => entry.success)
      .map((entry) => entry.account_id)
    const failedIds = resultEntries
      .filter((entry) => !entry.success)
      .map((entry) => entry.account_id)

    return {
      successIds,
      failedIds,
      successCount: typeof result.success === 'number' ? result.success : successIds.length,
      failedCount: typeof result.failed === 'number' ? result.failed : failedIds.length,
      hasIds: true,
      hasCounts: true
    }
  }

  const hasExplicitCounts =
    typeof result.success === 'number' || typeof result.failed === 'number'
  const successCount = typeof result.success === 'number' ? result.success : 0
  const failedCount = typeof result.failed === 'number' ? result.failed : 0

  if (
    hasExplicitCounts &&
    failedCount === 0 &&
    successCount === accountIds.length &&
    accountIds.length > 0
  ) {
    return {
      successIds: accountIds,
      failedIds: [],
      successCount,
      failedCount,
      hasIds: true,
      hasCounts: true
    }
  }

  return {
    successIds: [],
    failedIds: [],
    successCount,
    failedCount,
    hasIds: false,
    hasCounts: hasExplicitCounts
  }
}

export function updateSchedulableAccounts(
  accounts: Account[],
  accountIds: number[],
  schedulable: boolean
): Account[] {
  if (accountIds.length === 0) {
    return accounts
  }

  const idSet = new Set(accountIds)
  return accounts.map((account) =>
    idSet.has(account.id) ? { ...account, schedulable } : account
  )
}

export function accountMatchesCurrentFilters(
  account: Account,
  filters: AccountListFilters,
  nowMs: number = Date.now()
): boolean {
  if (filters.platform && account.platform !== filters.platform) {
    return false
  }

  if (filters.type && account.type !== filters.type) {
    return false
  }

  if (filters.status) {
    if (filters.status === 'rate_limited') {
      if (!account.rate_limit_reset_at) {
        return false
      }

      const resetAt = new Date(account.rate_limit_reset_at).getTime()
      if (!Number.isFinite(resetAt) || resetAt <= nowMs) {
        return false
      }
    } else if (account.status !== filters.status) {
      return false
    }
  }

  if (filters.privacy_mode) {
    const privacyMode = String(account.extra?.privacy_mode ?? '')
    if (filters.privacy_mode === '__unset__') {
      if (privacyMode) {
        return false
      }
    } else if (privacyMode !== filters.privacy_mode) {
      return false
    }
  }

  if (filters.group) {
    const groupIDs = Array.isArray(account.group_ids)
      ? account.group_ids
      : Array.isArray(account.groups)
        ? account.groups.map((group) => group.id)
        : []

    if (filters.group === 'ungrouped') {
      if (groupIDs.length > 0) {
        return false
      }
    } else {
      const targetGroupID = Number.parseInt(filters.group, 10)
      if (!Number.isInteger(targetGroupID) || !groupIDs.includes(targetGroupID)) {
        return false
      }
    }
  }

  const search = String(filters.search || '').trim().toLowerCase()
  if (search) {
    const matchesName = account.name.toLowerCase().includes(search)
    const matchesID = String(account.id) === search
    if (!matchesName && !matchesID) {
      return false
    }
  }

  return true
}

export function mergeAccountRuntimeFields(
  currentAccount: Account,
  updatedAccount: Account
): Account {
  return {
    ...updatedAccount,
    current_concurrency:
      updatedAccount.current_concurrency ?? currentAccount.current_concurrency,
    current_window_cost:
      updatedAccount.current_window_cost ?? currentAccount.current_window_cost,
    active_sessions: updatedAccount.active_sessions ?? currentAccount.active_sessions
  }
}

export function patchAccountList(
  accounts: Account[],
  updatedAccount: Account,
  filters: AccountListFilters,
  pagination: AccountListPagination,
  hasPendingListSync: boolean,
  menuAccountId?: number | null,
  nowMs: number = Date.now()
): PatchedAccountListResult {
  const accountIndex = accounts.findIndex((account) => account.id === updatedAccount.id)

  if (accountIndex === -1) {
    return {
      accounts,
      pagination: {
        page: pagination.page,
        total: pagination.total,
        pages: pagination.pages
      },
      hasPendingListSync,
      removedAccountId: null,
      patchedAccount: null,
      shouldCloseMenu: false
    }
  }

  const mergedAccount = mergeAccountRuntimeFields(accounts[accountIndex], updatedAccount)
  if (!accountMatchesCurrentFilters(mergedAccount, filters, nowMs)) {
    const remainingAccounts = accounts.filter((account) => account.id !== mergedAccount.id)
    const nextTotal = Math.max(0, pagination.total - 1)
    const nextPages = nextTotal > 0 ? Math.ceil(nextTotal / pagination.page_size) : 0
    const maxPage = Math.max(1, nextPages || 1)

    return {
      accounts: remainingAccounts,
      pagination: {
        total: nextTotal,
        pages: nextPages,
        page: pagination.page > maxPage ? maxPage : pagination.page
      },
      hasPendingListSync: nextTotal > 0,
      removedAccountId: mergedAccount.id,
      patchedAccount: null,
      shouldCloseMenu: menuAccountId === mergedAccount.id
    }
  }

  const nextAccounts = [...accounts]
  nextAccounts[accountIndex] = mergedAccount

  return {
    accounts: nextAccounts,
    pagination: {
      page: pagination.page,
      total: pagination.total,
      pages: pagination.pages
    },
    hasPendingListSync,
    removedAccountId: null,
    patchedAccount: mergedAccount,
    shouldCloseMenu: false
  }
}
