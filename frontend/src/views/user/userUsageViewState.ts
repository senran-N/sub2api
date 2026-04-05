import type { UsageQueryParams } from '@/types'

export interface UserUsageDateRangeSelection {
  startDate: string
  endDate: string
}

export interface UserUsagePaginationState {
  page: number
  page_size: number
}

export function formatUserUsageLocalDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function getLast7DaysUserUsageRange(
  now: Date = new Date()
): UserUsageDateRangeSelection {
  const start = new Date(now)
  start.setDate(start.getDate() - 6)

  return {
    startDate: formatUserUsageLocalDate(start),
    endDate: formatUserUsageLocalDate(now)
  }
}

export function buildDefaultUserUsageFilters(
  range: UserUsageDateRangeSelection
): UsageQueryParams {
  return {
    api_key_id: undefined,
    start_date: range.startDate,
    end_date: range.endDate
  }
}

export function applyUserUsageDateRange(
  filters: UsageQueryParams,
  range: UserUsageDateRangeSelection
): UsageQueryParams {
  return {
    ...filters,
    start_date: range.startDate,
    end_date: range.endDate
  }
}

export function buildResetUserUsageState(now: Date = new Date()): {
  filters: UsageQueryParams
  range: UserUsageDateRangeSelection
} {
  const range = getLast7DaysUserUsageRange(now)

  return {
    filters: buildDefaultUserUsageFilters(range),
    range
  }
}

export function applyUserUsagePageChange(
  pagination: UserUsagePaginationState,
  page: number
) {
  pagination.page = page
}

export function applyUserUsagePageSizeChange(
  pagination: UserUsagePaginationState,
  pageSize: number
) {
  pagination.page_size = pageSize
  pagination.page = 1
}

export function resetUserUsagePaginationPage(
  pagination: UserUsagePaginationState
) {
  pagination.page = 1
}
