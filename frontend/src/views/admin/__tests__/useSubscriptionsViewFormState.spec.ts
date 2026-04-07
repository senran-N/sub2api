import { describe, expect, it } from 'vitest'
import { useSubscriptionsViewFormState } from '../subscriptions/useSubscriptionsViewFormState'

describe('useSubscriptionsViewFormState', () => {
  it('updates semantic filter and user state without leaking raw mutations to the view', () => {
    const state = useSubscriptionsViewFormState()

    state.setFilterStatus('expired')
    state.setFilterGroupId('7')
    state.setFilterPlatform('openai')
    state.selectFilterUser(9)
    state.selectAssignUser(11)

    expect(state.filters.status).toBe('expired')
    expect(state.filters.group_id).toBe('7')
    expect(state.filters.platform).toBe('openai')
    expect(state.filters.user_id).toBe(9)
    expect(state.assignForm.user_id).toBe(11)

    state.clearFilterUser()
    state.clearAssignUser()
    expect(state.filters.user_id).toBeNull()
    expect(state.assignForm.user_id).toBeNull()
  })

  it('resets assign and extend form state without touching list filters', () => {
    const state = useSubscriptionsViewFormState()

    state.assignForm.user_id = 3
    state.assignForm.group_id = 4
    state.assignForm.validity_days = 60
    state.extendForm.days = 90
    state.filters.status = 'revoked'

    state.resetAssignFormState()
    state.resetExtendFormState()

    expect(state.assignForm).toEqual({
      user_id: null,
      group_id: null,
      validity_days: 30
    })
    expect(state.extendForm).toEqual({ days: 30 })
    expect(state.filters.status).toBe('revoked')
  })
})
