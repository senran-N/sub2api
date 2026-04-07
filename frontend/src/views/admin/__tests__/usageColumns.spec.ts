import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Column } from '@/components/common/types'
import {
  ALWAYS_VISIBLE_USAGE_COLUMNS,
  DEFAULT_HIDDEN_USAGE_COLUMNS,
  buildToggleableUsageColumns,
  buildVisibleUsageColumns,
  isUsageColumnAlwaysVisible,
  loadUsageHiddenColumns,
  saveUsageHiddenColumns,
  toggleUsageHiddenColumn
} from '../usage/usageColumns'

describe('usageColumns', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('recognizes always visible columns and derives toggleable/visible subsets', () => {
    expect(ALWAYS_VISIBLE_USAGE_COLUMNS).toEqual(['user', 'created_at'])
    expect(DEFAULT_HIDDEN_USAGE_COLUMNS).toEqual(['reasoning_effort', 'user_agent'])
    expect(isUsageColumnAlwaysVisible('user')).toBe(true)
    expect(isUsageColumnAlwaysVisible('model')).toBe(false)

    const columns: Column[] = [
      { key: 'user', label: 'User', sortable: false },
      { key: 'model', label: 'Model', sortable: false },
      { key: 'created_at', label: 'Time', sortable: false }
    ]

    expect(buildToggleableUsageColumns(columns).map((column) => column.key)).toEqual(['model'])
    expect(
      buildVisibleUsageColumns(columns, new Set(['model'])).map((column) => column.key)
    ).toEqual(['user', 'created_at'])
  })

  it('toggles hidden columns and persists them', () => {
    const hiddenColumns = new Set<string>(['reasoning_effort'])

    expect(toggleUsageHiddenColumn(hiddenColumns, 'model')).toBe(true)
    expect(hiddenColumns.has('model')).toBe(true)

    expect(toggleUsageHiddenColumn(hiddenColumns, 'reasoning_effort')).toBe(false)
    expect(hiddenColumns.has('reasoning_effort')).toBe(false)

    saveUsageHiddenColumns(hiddenColumns)
    expect(loadUsageHiddenColumns()).toEqual(hiddenColumns)
  })

  it('falls back to defaults when local storage is missing or invalid', () => {
    expect(loadUsageHiddenColumns()).toEqual(new Set(DEFAULT_HIDDEN_USAGE_COLUMNS))

    const getItem = vi.spyOn(Storage.prototype, 'getItem').mockReturnValue('not-json')
    expect(loadUsageHiddenColumns()).toEqual(new Set(DEFAULT_HIDDEN_USAGE_COLUMNS))
    getItem.mockRestore()
  })
})
