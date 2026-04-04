import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { computed } from 'vue'
import { useUsageViewColumns } from '../useUsageViewColumns'

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

describe('useUsageViewColumns', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('restores persisted hidden columns and derives toggleable state', () => {
    const localStorageMock = createStorage()
    localStorageMock.setItem('usage-hidden-columns', JSON.stringify(['model']))
    vi.stubGlobal('localStorage', localStorageMock)

    const state = useUsageViewColumns({
      allColumns: computed(() => [
        { key: 'user', label: 'User', sortable: false },
        { key: 'model', label: 'Model', sortable: false },
        { key: 'created_at', label: 'Time', sortable: false }
      ])
    })

    expect([...state.hiddenColumns]).toEqual(['model'])
    expect(state.toggleableColumns.value.map((column) => column.key)).toEqual(['model'])
    expect(state.visibleColumns.value.map((column) => column.key)).toEqual(['user', 'created_at'])
  })

  it('toggles persisted columns without mixing in dropdown state', () => {
    vi.stubGlobal('localStorage', createStorage())

    const state = useUsageViewColumns({
      allColumns: computed(() => [
        { key: 'user', label: 'User', sortable: false },
        { key: 'model', label: 'Model', sortable: false },
        { key: 'created_at', label: 'Time', sortable: false }
      ])
    })

    state.toggleColumn('model')
    expect(state.hiddenColumns.has('model')).toBe(true)
    expect(state.isColumnVisible('model')).toBe(false)
    expect(localStorage.setItem).toHaveBeenCalled()

    state.toggleColumn('model')
    expect(state.hiddenColumns.has('model')).toBe(false)
    expect(state.isColumnVisible('model')).toBe(true)
  })
})
