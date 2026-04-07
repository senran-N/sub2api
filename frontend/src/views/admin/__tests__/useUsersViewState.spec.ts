import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { UserAttributeDefinition } from '@/types'
import {
  USER_FILTER_VALUES_KEY,
  USER_HIDDEN_COLUMNS_KEY,
  USER_VISIBLE_FILTERS_KEY,
  useUsersViewState
} from '../users/useUsersViewState'

function createAttributeDefinition(
  overrides: Partial<UserAttributeDefinition> = {}
): UserAttributeDefinition {
  return {
    id: 1,
    key: 'department',
    name: 'Department',
    description: '',
    type: 'select',
    options: [],
    required: false,
    validation: {},
    placeholder: '',
    display_order: 0,
    enabled: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

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

describe('useUsersViewState', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('restores persisted column and filter state and triggers groups lazy-load when needed', () => {
    const localStorageMock = createStorage()
    localStorageMock.setItem(
      USER_HIDDEN_COLUMNS_KEY,
      JSON.stringify(['usage', 'subscriptions'])
    )
    localStorageMock.setItem(USER_VISIBLE_FILTERS_KEY, JSON.stringify(['group', 'attr_7']))
    localStorageMock.setItem(
      USER_FILTER_VALUES_KEY,
      JSON.stringify({
        role: 'admin',
        status: 'disabled',
        group: 'VIP',
        attributes: { 7: 'north' }
      })
    )
    vi.stubGlobal('localStorage', localStorageMock)

    const loadUsers = vi.fn()
    const loadGroups = vi.fn()

    const state = useUsersViewState({
      attributeDefinitions: ref([createAttributeDefinition({ id: 7 })]),
      initialPageSize: 20,
      loadUsers,
      loadGroups,
      loadSecondaryData: vi.fn(),
      resetSecondaryData: vi.fn()
    })

    state.initializePersistedState()

    expect([...state.hiddenColumns]).toEqual(['usage', 'subscriptions'])
    expect([...state.visibleFilters]).toEqual(['group', 'attr_7'])
    expect(state.filters).toEqual({
      role: 'admin',
      status: 'disabled',
      group: 'VIP'
    })
    expect(state.activeAttributeFilters).toEqual({ 7: 'north' })
    expect(loadGroups).toHaveBeenCalledTimes(1)
    expect(loadUsers).not.toHaveBeenCalled()
  })

  it('falls back to defaults when persisted state is corrupted', () => {
    const localStorageMock = createStorage()
    localStorageMock.getItem.mockImplementation((key: string) => {
      if (key === USER_HIDDEN_COLUMNS_KEY) return 'not-json'
      if (key === USER_VISIBLE_FILTERS_KEY) return 'still-bad'
      return null
    })
    vi.stubGlobal('localStorage', localStorageMock)
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const state = useUsersViewState({
      attributeDefinitions: ref([]),
      initialPageSize: 20,
      loadUsers: vi.fn(),
      loadGroups: vi.fn(),
      loadSecondaryData: vi.fn(),
      resetSecondaryData: vi.fn()
    })

    state.initializePersistedState()

    expect(state.hiddenColumns.has('notes')).toBe(true)
    expect(state.hiddenColumns.has('groups')).toBe(true)
    expect(state.visibleFilters.size).toBe(0)
    expect(state.filters).toEqual({
      role: '',
      status: '',
      group: ''
    })
    expect(errorSpy).toHaveBeenCalled()
  })

  it('debounces search and applies column toggle side effects', () => {
    const localStorageMock = createStorage()
    vi.stubGlobal('localStorage', localStorageMock)

    const loadUsers = vi.fn()
    const loadGroups = vi.fn()
    const loadSecondaryData = vi.fn()

    const state = useUsersViewState({
      attributeDefinitions: ref([createAttributeDefinition({ id: 9 })]),
      initialPageSize: 20,
      loadUsers,
      loadGroups,
      loadSecondaryData,
      resetSecondaryData: vi.fn()
    })

    state.initializePersistedState()
    state.pagination.page = 4
    state.handleSearch()
    state.handleSearch()
    vi.advanceTimersByTime(299)
    expect(loadUsers).not.toHaveBeenCalled()
    vi.advanceTimersByTime(1)
    expect(state.pagination.page).toBe(1)
    expect(loadUsers).toHaveBeenCalledTimes(1)

    state.setCurrentUserIds([1, 2])
    state.toggleColumn('usage')
    expect(loadSecondaryData).toHaveBeenCalledWith([1, 2], undefined, 1)

    state.toggleColumn('subscriptions')
    expect(loadUsers).toHaveBeenCalledTimes(2)

    state.toggleColumn('groups')
    expect(loadGroups).toHaveBeenCalledTimes(1)
  })

  it('coordinates secondary data scheduling and filter mutations', () => {
    const localStorageMock = createStorage()
    vi.stubGlobal('localStorage', localStorageMock)

    const loadUsers = vi.fn()
    const loadGroups = vi.fn()
    const loadSecondaryData = vi.fn()
    const resetSecondaryData = vi.fn()
    const signal = new AbortController().signal

    const state = useUsersViewState({
      attributeDefinitions: ref([createAttributeDefinition({ id: 5 })]),
      initialPageSize: 20,
      loadUsers,
      loadGroups,
      loadSecondaryData,
      resetSecondaryData
    })

    state.setCurrentUserIds([3, 4])
    state.resetSecondaryDataState()
    state.scheduleUsersSecondaryDataLoad(signal)
    state.scheduleUsersSecondaryDataLoad(signal)

    expect(resetSecondaryData).toHaveBeenCalledTimes(1)
    vi.advanceTimersByTime(49)
    expect(loadSecondaryData).not.toHaveBeenCalled()
    vi.advanceTimersByTime(1)
    expect(loadSecondaryData).toHaveBeenCalledTimes(1)
    expect(loadSecondaryData).toHaveBeenCalledWith([3, 4], signal, 1)
    expect(state.isSecondaryDataRequestCurrent(1)).toBe(true)
    expect(state.isSecondaryDataRequestCurrent(2)).toBe(false)

    state.toggleBuiltInFilter('group')
    expect(loadGroups).toHaveBeenCalledTimes(1)
    expect(loadUsers).toHaveBeenCalledTimes(1)

    state.toggleAttributeFilter(5)
    expect(state.visibleFilters.has('attr_5')).toBe(true)
    expect(state.activeAttributeFilters[5]).toBe('')
    expect(loadUsers).toHaveBeenCalledTimes(2)

    state.updateAttributeFilter(5, 'vip')
    state.applyFilter()
    expect(localStorageMock.setItem).toHaveBeenCalled()
    expect(loadUsers).toHaveBeenCalledTimes(3)
  })
})
