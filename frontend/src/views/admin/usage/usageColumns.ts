import type { Column } from '@/components/common/types'

export const ALWAYS_VISIBLE_USAGE_COLUMNS = ['user', 'created_at'] as const
export const DEFAULT_HIDDEN_USAGE_COLUMNS = ['reasoning_effort', 'user_agent'] as const
export const USAGE_HIDDEN_COLUMNS_STORAGE_KEY = 'usage-hidden-columns'

export function isUsageColumnAlwaysVisible(key: string) {
  return ALWAYS_VISIBLE_USAGE_COLUMNS.includes(
    key as (typeof ALWAYS_VISIBLE_USAGE_COLUMNS)[number]
  )
}

export function buildToggleableUsageColumns(columns: Column[]) {
  return columns.filter((column) => !isUsageColumnAlwaysVisible(column.key))
}

export function buildVisibleUsageColumns(columns: Column[], hiddenColumns: Set<string>) {
  return columns.filter(
    (column) => isUsageColumnAlwaysVisible(column.key) || !hiddenColumns.has(column.key)
  )
}

export function toggleUsageHiddenColumn(hiddenColumns: Set<string>, key: string) {
  if (hiddenColumns.has(key)) {
    hiddenColumns.delete(key)
    return false
  }

  hiddenColumns.add(key)
  return true
}

export function saveUsageHiddenColumns(hiddenColumns: Set<string>) {
  localStorage.setItem(
    USAGE_HIDDEN_COLUMNS_STORAGE_KEY,
    JSON.stringify(Array.from(hiddenColumns))
  )
}

export function loadUsageHiddenColumns(): Set<string> {
  const fallback = new Set<string>(DEFAULT_HIDDEN_USAGE_COLUMNS)

  try {
    const saved = localStorage.getItem(USAGE_HIDDEN_COLUMNS_STORAGE_KEY)
    if (!saved) {
      return fallback
    }

    return new Set(JSON.parse(saved) as string[])
  } catch {
    return fallback
  }
}
