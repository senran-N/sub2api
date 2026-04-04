import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { reactive } from 'vue'
import {
  createDefaultAssignSubscriptionForm,
  createDefaultSubscriptionFilters
} from '../subscriptionForm'
import { useSubscriptionsViewUserSearches } from '../useSubscriptionsViewUserSearches'

describe('useSubscriptionsViewUserSearches', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('searches and selects filter users with debounce', async () => {
    const filters = reactive(createDefaultSubscriptionFilters())
    const assignForm = reactive(createDefaultAssignSubscriptionForm())
    const applyFilters = vi.fn()
    const searchUsers = vi.fn().mockResolvedValue([{ id: 7, email: 'demo@example.com' }])

    const state = useSubscriptionsViewUserSearches({
      filters,
      assignForm,
      applyFilters,
      searchUsers
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
    expect(filters.user_id).toBe(7)
    expect(applyFilters).toHaveBeenCalledTimes(1)

    state.clearFilterUser()
    expect(filters.user_id).toBeNull()
    expect(applyFilters).toHaveBeenCalledTimes(2)
  })

  it('searches assign users and resets assign state', async () => {
    const filters = reactive(createDefaultSubscriptionFilters())
    const assignForm = reactive(createDefaultAssignSubscriptionForm())
    const searchUsers = vi.fn().mockResolvedValue([{ id: 9, email: 'assign@example.com' }])

    const state = useSubscriptionsViewUserSearches({
      filters,
      assignForm,
      applyFilters: vi.fn(),
      searchUsers
    })

    state.userSearchKeyword.value = 'assign'
    state.debounceSearchUsers()
    vi.advanceTimersByTime(300)
    await Promise.resolve()

    expect(searchUsers).toHaveBeenCalledWith('assign')
    state.selectUser({ id: 9, email: 'assign@example.com' })
    expect(assignForm.user_id).toBe(9)

    state.resetAssignSearch()
    expect(assignForm.user_id).toBeNull()
    expect(state.userSearchKeyword.value).toBe('')
    expect(state.showUserDropdown.value).toBe(false)
  })
})
