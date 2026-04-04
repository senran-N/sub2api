import { afterEach, describe, expect, it, vi } from 'vitest'
import { computed } from 'vue'
import {
  SUBSCRIPTION_HIDDEN_COLUMNS_KEY,
  SUBSCRIPTION_USER_COLUMN_MODE_KEY,
  useSubscriptionsViewColumns
} from '../useSubscriptionsViewColumns'

function createStorage() {
  const store = new Map<string, string>()

  return {
    getItem: vi.fn((key: string) => store.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store.set(key, value)
    }),
    removeItem: vi.fn((key: string) => {
      store.delete(key)
    }),
    clear: vi.fn(() => {
      store.clear()
    })
  }
}

describe('useSubscriptionsViewColumns', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('restores persisted hidden columns and user column mode', () => {
    const localStorageMock = createStorage()
    localStorageMock.setItem(SUBSCRIPTION_HIDDEN_COLUMNS_KEY, JSON.stringify(['group']))
    localStorageMock.setItem(SUBSCRIPTION_USER_COLUMN_MODE_KEY, 'username')
    vi.stubGlobal('localStorage', localStorageMock)

    const state = useSubscriptionsViewColumns({
      allColumns: computed(() => [
        { key: 'user', label: 'User', sortable: false },
        { key: 'group', label: 'Group', sortable: false },
        { key: 'status', label: 'Status', sortable: false },
        { key: 'actions', label: 'Actions', sortable: false }
      ])
    })

    expect(state.userColumnMode.value).toBe('username')
    expect([...state.hiddenColumns]).toEqual(['group'])
    expect(state.toggleableColumns.value.map((column) => column.key)).toEqual([
      'group',
      'status'
    ])
    expect(state.columns.value.map((column) => column.key)).toEqual([
      'user',
      'status',
      'actions'
    ])
  })

  it('persists mode and column changes without dropdown state', () => {
    vi.stubGlobal('localStorage', createStorage())

    const state = useSubscriptionsViewColumns({
      allColumns: computed(() => [
        { key: 'user', label: 'User', sortable: false },
        { key: 'group', label: 'Group', sortable: false },
        { key: 'status', label: 'Status', sortable: false },
        { key: 'actions', label: 'Actions', sortable: false }
      ])
    })

    state.setUserColumnMode('username')
    expect(localStorage.setItem).toHaveBeenCalledWith(
      SUBSCRIPTION_USER_COLUMN_MODE_KEY,
      'username'
    )

    state.toggleColumn('group')
    expect(localStorage.setItem).toHaveBeenCalledWith(
      SUBSCRIPTION_HIDDEN_COLUMNS_KEY,
      JSON.stringify(['group'])
    )
  })
})
