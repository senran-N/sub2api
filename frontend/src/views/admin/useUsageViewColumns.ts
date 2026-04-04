import { computed, reactive, type ComputedRef } from 'vue'
import type { Column } from '@/components/common/types'
import {
  buildToggleableUsageColumns,
  buildVisibleUsageColumns,
  loadUsageHiddenColumns,
  saveUsageHiddenColumns,
  toggleUsageHiddenColumn
} from './usageColumns'

interface UsageViewColumnsOptions {
  allColumns: ComputedRef<Column[]>
}

export function useUsageViewColumns(options: UsageViewColumnsOptions) {
  const hiddenColumns = reactive<Set<string>>(loadUsageHiddenColumns())

  const toggleableColumns = computed(() =>
    buildToggleableUsageColumns(options.allColumns.value)
  )

  const visibleColumns = computed(() =>
    buildVisibleUsageColumns(options.allColumns.value, hiddenColumns)
  )

  const isColumnVisible = (key: string) => !hiddenColumns.has(key)

  const toggleColumn = (key: string) => {
    try {
      toggleUsageHiddenColumn(hiddenColumns, key)
      saveUsageHiddenColumns(hiddenColumns)
    } catch (error) {
      console.error('Failed to save columns:', error)
    }
  }

  return {
    hiddenColumns,
    toggleableColumns,
    visibleColumns,
    isColumnVisible,
    toggleColumn
  }
}
