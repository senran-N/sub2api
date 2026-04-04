import { describe, expect, it, vi } from 'vitest'
import type { UserSubscription } from '@/types'
import { useSubscriptionsViewDialogs } from '../useSubscriptionsViewDialogs'

function createSubscription(overrides: Partial<UserSubscription> = {}): UserSubscription {
  return {
    id: 1,
    user_id: 2,
    group_id: 3,
    quota_limit: 0,
    quota_used: 0,
    expires_at: '2026-04-10T00:00:00Z',
    status: 'active',
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    ...overrides
  } as UserSubscription
}

describe('useSubscriptionsViewDialogs', () => {
  it('opens and closes modal states while resetting dependent form state', () => {
    const resetAssignState = vi.fn()
    const resetExtendState = vi.fn()
    const state = useSubscriptionsViewDialogs({
      resetAssignState,
      resetExtendState
    })

    state.openGuideModal()
    expect(state.showGuideModal.value).toBe(true)
    state.closeGuideModal()
    expect(state.showGuideModal.value).toBe(false)

    state.openAssignModal()
    expect(state.showAssignModal.value).toBe(true)
    state.closeAssignModal()
    expect(state.showAssignModal.value).toBe(false)
    expect(resetAssignState).toHaveBeenCalledTimes(1)

    const subscription = createSubscription()
    state.openExtendModal(subscription)
    expect(state.showExtendModal.value).toBe(true)
    expect(state.extendingSubscription.value?.id).toBe(1)
    expect(resetExtendState).toHaveBeenCalledTimes(1)
    state.closeExtendModal()
    expect(state.showExtendModal.value).toBe(false)
    expect(state.extendingSubscription.value).toBeNull()
    expect(resetExtendState).toHaveBeenCalledTimes(2)
  })

  it('tracks revoke/reset confirmation targets', () => {
    const state = useSubscriptionsViewDialogs({
      resetAssignState: vi.fn(),
      resetExtendState: vi.fn()
    })
    const subscription = createSubscription({ id: 9 })

    state.openRevokeDialog(subscription)
    expect(state.showRevokeDialog.value).toBe(true)
    expect(state.revokingSubscription.value?.id).toBe(9)
    state.closeRevokeDialog()
    expect(state.showRevokeDialog.value).toBe(false)
    expect(state.revokingSubscription.value).toBeNull()

    state.openResetQuotaConfirm(subscription)
    expect(state.showResetQuotaConfirm.value).toBe(true)
    expect(state.resettingSubscription.value?.id).toBe(9)
    state.closeResetQuotaConfirm()
    expect(state.showResetQuotaConfirm.value).toBe(false)
    expect(state.resettingSubscription.value).toBeNull()
  })
})
