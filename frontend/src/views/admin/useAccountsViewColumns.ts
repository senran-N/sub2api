import { computed, reactive, type ComputedRef } from 'vue'
import type { Column } from '@/components/common/types'
import {
  ACCOUNT_HIDDEN_COLUMNS_KEY,
  DEFAULT_ACCOUNT_HIDDEN_COLUMNS
} from './accountsList'
import {
  buildAccountAutoRefreshIntervalLabel,
  buildAccountTableColumns
} from './accountsView'

interface AccountsViewColumnsOptions {
  t: (key: string) => string
  isSimpleMode: ComputedRef<boolean>
  onUsageColumnShown: () => void | Promise<void>
}

function loadSavedAccountColumns(hiddenColumns: Set<string>) {
  try {
    const saved = localStorage.getItem(ACCOUNT_HIDDEN_COLUMNS_KEY)
    if (saved) {
      const parsed = JSON.parse(saved) as string[]
      parsed.forEach((key) => {
        hiddenColumns.add(key)
      })
      return
    }
  } catch (error) {
    console.error('Failed to load saved columns:', error)
  }

  DEFAULT_ACCOUNT_HIDDEN_COLUMNS.forEach((key) => {
    hiddenColumns.add(key)
  })
}

function saveAccountColumns(hiddenColumns: Set<string>) {
  try {
    localStorage.setItem(ACCOUNT_HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
  } catch (error) {
    console.error('Failed to save columns:', error)
  }
}

export function useAccountsViewColumns(options: AccountsViewColumnsOptions) {
  const hiddenColumns = reactive<Set<string>>(new Set())

  if (typeof window !== 'undefined') {
    loadSavedAccountColumns(hiddenColumns)
  }

  const allColumns = computed<Column[]>(() =>
    buildAccountTableColumns(options.isSimpleMode.value, options.t)
  )

  const toggleableColumns = computed(() =>
    allColumns.value.filter(
      (column) =>
        column.key !== 'select' && column.key !== 'name' && column.key !== 'actions'
    )
  )

  const cols = computed(() =>
    allColumns.value.filter(
      (column) =>
        column.key === 'select' ||
        column.key === 'name' ||
        column.key === 'actions' ||
        !hiddenColumns.has(column.key)
    )
  )

  const isColumnVisible = (key: string) => !hiddenColumns.has(key)

  const toggleColumn = (key: string) => {
    const wasHidden = hiddenColumns.has(key)
    if (wasHidden) {
      hiddenColumns.delete(key)
    } else {
      hiddenColumns.add(key)
    }
    saveAccountColumns(hiddenColumns)

    if ((key === 'today_stats' || key === 'usage') && wasHidden) {
      Promise.resolve(options.onUsageColumnShown()).catch((error) => {
        console.error('Failed to load account today stats after showing column:', error)
      })
    }
  }

  const autoRefreshIntervalLabel = (seconds: number) =>
    buildAccountAutoRefreshIntervalLabel(seconds, options.t)

  return {
    hiddenColumns,
    allColumns,
    toggleableColumns,
    cols,
    isColumnVisible,
    toggleColumn,
    autoRefreshIntervalLabel
  }
}
