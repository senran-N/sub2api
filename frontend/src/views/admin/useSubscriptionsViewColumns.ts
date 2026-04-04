import { computed, reactive, ref, type ComputedRef } from 'vue'
import type { Column } from '@/components/common/types'

export const SUBSCRIPTION_HIDDEN_COLUMNS_KEY = 'subscription-hidden-columns'
export const SUBSCRIPTION_USER_COLUMN_MODE_KEY = 'subscription-user-column-mode'

type UserColumnMode = 'email' | 'username'

interface SubscriptionsViewColumnsOptions {
  allColumns: ComputedRef<Column[]>
}

function loadSubscriptionHiddenColumns(): Set<string> {
  try {
    const saved = localStorage.getItem(SUBSCRIPTION_HIDDEN_COLUMNS_KEY)
    if (!saved) {
      return new Set<string>()
    }

    return new Set(JSON.parse(saved) as string[])
  } catch (error) {
    console.error('Failed to load saved columns:', error)
    return new Set<string>()
  }
}

function loadSubscriptionUserColumnMode(): UserColumnMode {
  try {
    const saved = localStorage.getItem(SUBSCRIPTION_USER_COLUMN_MODE_KEY)
    if (saved === 'email' || saved === 'username') {
      return saved
    }
  } catch (error) {
    console.error('Failed to load user column mode:', error)
  }

  return 'email'
}

export function useSubscriptionsViewColumns(options: SubscriptionsViewColumnsOptions) {
  const hiddenColumns = reactive<Set<string>>(loadSubscriptionHiddenColumns())
  const userColumnMode = ref<UserColumnMode>(loadSubscriptionUserColumnMode())

  const toggleableColumns = computed(() =>
    options.allColumns.value.filter((column) => column.key !== 'user' && column.key !== 'actions')
  )

  const columns = computed<Column[]>(() =>
    options.allColumns.value.filter(
      (column) =>
        column.key === 'user' || column.key === 'actions' || !hiddenColumns.has(column.key)
    )
  )

  const saveColumnsToStorage = () => {
    try {
      localStorage.setItem(
        SUBSCRIPTION_HIDDEN_COLUMNS_KEY,
        JSON.stringify(Array.from(hiddenColumns))
      )
    } catch (error) {
      console.error('Failed to save columns:', error)
    }
  }

  const isColumnVisible = (key: string) => !hiddenColumns.has(key)

  const toggleColumn = (key: string) => {
    if (hiddenColumns.has(key)) {
      hiddenColumns.delete(key)
    } else {
      hiddenColumns.add(key)
    }

    saveColumnsToStorage()
  }

  const setUserColumnMode = (mode: UserColumnMode) => {
    userColumnMode.value = mode

    try {
      localStorage.setItem(SUBSCRIPTION_USER_COLUMN_MODE_KEY, mode)
    } catch (error) {
      console.error('Failed to save user column mode:', error)
    }
  }

  return {
    hiddenColumns,
    userColumnMode,
    toggleableColumns,
    columns,
    isColumnVisible,
    toggleColumn,
    setUserColumnMode
  }
}
