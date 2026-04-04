import type { UserUsageTrendPoint } from '@/types'

const USER_TREND_COLORS = [
  '#3b82f6',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#ec4899',
  '#14b8a6',
  '#f97316',
  '#6366f1',
  '#84cc16',
  '#06b6d4',
  '#a855f7'
] as const

export function formatDashboardLocalDate(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function getDashboardLast24HoursRangeDates(now: Date = new Date()): {
  start: string
  end: string
} {
  const start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
  return {
    start: formatDashboardLocalDate(start),
    end: formatDashboardLocalDate(now)
  }
}

export function getDashboardGranularityForRange(
  startDate: string,
  endDate: string
): 'day' | 'hour' {
  const start = new Date(startDate)
  const end = new Date(endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))
  return daysDiff <= 1 ? 'hour' : 'day'
}

export function formatDashboardTokens(value: number | undefined | null): string {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  }
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

export function formatDashboardNumber(value: number): string {
  return value.toLocaleString()
}

export function formatDashboardCost(value: number): string {
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)}K`
  }
  if (value >= 1) {
    return value.toFixed(2)
  }
  if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}

export function formatDashboardDuration(ms: number): string {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

export function buildDashboardUserTrendChartData(
  userTrend: UserUsageTrendPoint[],
  t: (key: string, params?: Record<string, unknown>) => string
): {
  labels: string[]
  datasets: Array<{
    label: string
    data: number[]
    borderColor: string
    backgroundColor: string
    fill: boolean
    tension: number
  }>
} | null {
  if (!userTrend.length) {
    return null
  }

  const getDisplayName = (point: UserUsageTrendPoint): string => {
    const username = point.username?.trim()
    if (username) {
      return username
    }

    const email = point.email?.trim()
    if (email) {
      return email
    }

    return t('admin.redeem.userPrefix', { id: point.user_id })
  }

  const userGroups = new Map<number, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  userTrend.forEach((point) => {
    allDates.add(point.date)
    if (!userGroups.has(point.user_id)) {
      userGroups.set(point.user_id, {
        name: getDisplayName(point),
        data: new Map()
      })
    }
    userGroups.get(point.user_id)!.data.set(point.date, point.tokens)
  })

  const sortedDates = Array.from(allDates).sort()
  const datasets = Array.from(userGroups.values()).map((group, index) => ({
    label: group.name,
    data: sortedDates.map((date) => group.data.get(date) || 0),
    borderColor: USER_TREND_COLORS[index % USER_TREND_COLORS.length],
    backgroundColor: `${USER_TREND_COLORS[index % USER_TREND_COLORS.length]}20`,
    fill: false,
    tension: 0.3
  }))

  return {
    labels: sortedDates,
    datasets
  }
}
