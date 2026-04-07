import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useSubscriptionsViewUserSearches } from '../subscriptions/useSubscriptionsViewUserSearches'

describe('useSubscriptionsViewUserSearches', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('searches and selects filter users with debounce', async () => {
    const applyFilters = vi.fn()
    const searchUsers = vi.fn().mockResolvedValue([{ id: 7, email: 'demo@example.com' }])
    const selectFilterUser = vi.fn()
    const clearFilterUser = vi.fn()

    const state = useSubscriptionsViewUserSearches({
      applyFilters,
      searchUsers,
      selectFilterUser,
      clearFilterUser,
      selectAssignUser: vi.fn(),
      clearAssignUser: vi.fn()
    })

    state.filterUserKeyword.value = 'demo'
    state.debounceSearchFilterUsers()
    vi.advanceTimersByTime(299)
    expect(searchUsers).not.toHaveBeenCalled()
    vi.advanceTimersByTime(1)
    await Promise.resolve()

    expect(searchUsers).toHaveBeenCalledWith('demo')
    expect(state.filterUserResults.value).toEqual([{ id: 7, email: 'demo@example.com' }])

    state.selectFilterUser({ id: 7, email: 'demo@example.com' })
    expect(selectFilterUser).toHaveBeenCalledWith(7)
    expect(applyFilters).toHaveBeenCalledTimes(1)

    state.clearFilterUser()
    expect(clearFilterUser).toHaveBeenCalledTimes(1)
    expect(applyFilters).toHaveBeenCalledTimes(2)
  })

  it('searches assign users and resets assign state', async () => {
    const searchUsers = vi.fn().mockResolvedValue([{ id: 9, email: 'assign@example.com' }])
    const selectAssignUser = vi.fn()
    const clearAssignUser = vi.fn()

    const state = useSubscriptionsViewUserSearches({
      applyFilters: vi.fn(),
      searchUsers,
      selectFilterUser: vi.fn(),
      clearFilterUser: vi.fn(),
      selectAssignUser,
      clearAssignUser
    })

    state.userSearchKeyword.value = 'assign'
    state.debounceSearchUsers()
    vi.advanceTimersByTime(300)
    await Promise.resolve()

    expect(searchUsers).toHaveBeenCalledWith('assign')
    state.selectUser({ id: 9, email: 'assign@example.com' })
    expect(selectAssignUser).toHaveBeenCalledWith(9)

    state.resetAssignSearch()
    expect(clearAssignUser).toHaveBeenCalledTimes(1)
    expect(state.userSearchKeyword.value).toBe('')
    expect(state.showUserDropdown.value).toBe(false)
  })
})
