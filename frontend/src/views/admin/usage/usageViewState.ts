import type { AdminUsageQueryParams } from '@/api/admin/usage'

export type UsageGranularity = 'day' | 'hour'

export interface UsageDateRangeSelection {
  startDate: string
  endDate: string
}

export interface UsagePaginationState {
  page: number
  page_size: number
}

type UsageQueryValue = string | null | Array<string | null> | undefined

export interface UsageRouteQueryLike {
  start_date?: UsageQueryValue
  end_date?: UsageQueryValue
  user_id?: UsageQueryValue
}

export function formatUsageLocalDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function getLast24HoursUsageRange(now: Date = new Date()): UsageDateRangeSelection {
  const start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
  return {
    startDate: formatUsageLocalDate(start),
    endDate: formatUsageLocalDate(now)
  }
}

export function getUsageGranularityForRange(
  startDate: string,
  endDate: string
): UsageGranularity {
  const startTime = new Date(`${startDate}T00:00:00`).getTime()
  const endTime = new Date(`${endDate}T00:00:00`).getTime()
  const daysDiff = Math.ceil((endTime - startTime) / (1000 * 60 * 60 * 24))
  return daysDiff <= 1 ? 'hour' : 'day'
}

export function getUsageQueryStringValue(value: UsageQueryValue): string | undefined {
  if (Array.isArray(value)) {
    return value.find((item): item is string => typeof item === 'string' && item.length > 0)
  }

  return typeof value === 'string' && value.length > 0 ? value : undefined
}

export function getUsageQueryNumberValue(value: UsageQueryValue): number | undefined {
  const raw = getUsageQueryStringValue(value)
  if (!raw) {
    return undefined
  }

  const parsed = Number(raw)
  return Number.isFinite(parsed) ? parsed : undefined
}

export function buildDefaultUsageFilters(
  range: UsageDateRangeSelection
): AdminUsageQueryParams {
  return {
    user_id: undefined,
    model: undefined,
    group_id: undefined,
    request_type: undefined,
    billing_type: null,
    start_date: range.startDate,
    end_date: range.endDate
  }
}

export function applyUsageRouteQueryState(
  query: UsageRouteQueryLike,
  filters: AdminUsageQueryParams,
  range: UsageDateRangeSelection
): {
  filters: AdminUsageQueryParams
  range: UsageDateRangeSelection
  granularity: UsageGranularity
} {
  const nextRange = {
    startDate: getUsageQueryStringValue(query.start_date) || range.startDate,
    endDate: getUsageQueryStringValue(query.end_date) || range.endDate
  }

  return {
    filters: {
      ...filters,
      user_id: getUsageQueryNumberValue(query.user_id),
      start_date: nextRange.startDate,
      end_date: nextRange.endDate
    },
    range: nextRange,
    granularity: getUsageGranularityForRange(nextRange.startDate, nextRange.endDate)
  }
}

export function applyUsageDateRangeState(
  range: UsageDateRangeSelection,
  filters: AdminUsageQueryParams
): {
  filters: AdminUsageQueryParams
  range: UsageDateRangeSelection
  granularity: UsageGranularity
} {
  return {
    filters: {
      ...filters,
      start_date: range.startDate,
      end_date: range.endDate
    },
    range,
    granularity: getUsageGranularityForRange(range.startDate, range.endDate)
  }
}

export function buildResetUsageState(now: Date = new Date()): {
  filters: AdminUsageQueryParams
  range: UsageDateRangeSelection
  granularity: UsageGranularity
} {
  const range = getLast24HoursUsageRange(now)
  return {
    filters: {
      start_date: range.startDate,
      end_date: range.endDate,
      request_type: undefined,
      billing_type: null
    },
    range,
    granularity: getUsageGranularityForRange(range.startDate, range.endDate)
  }
}

export function applyUsagePageChange(pagination: UsagePaginationState, page: number) {
  pagination.page = page
}

export function applyUsagePageSizeChange(
  pagination: UsagePaginationState,
  pageSize: number
) {
  pagination.page_size = pageSize
  pagination.page = 1
}

export function resetUsagePaginationPage(pagination: UsagePaginationState) {
  pagination.page = 1
}
