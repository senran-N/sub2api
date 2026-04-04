import { describe, expect, it, vi } from 'vitest'
import type { AdminUser } from '@/types'
import { useUsageViewDialogs } from '../useUsageViewDialogs'

function createUser(overrides: Partial<AdminUser> = {}): AdminUser {
  return {
    id: 7,
    username: 'demo',
    email: 'demo@example.com',
    role: 'user',
    balance: 1,
    concurrency: 1,
    status: 'active',
    allowed_groups: null,
    notes: '',
    sora_storage_quota_bytes: 0,
    sora_storage_used_bytes: 0,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

describe('useUsageViewDialogs', () => {
  it('opens and closes dialog state explicitly', () => {
    const state = useUsageViewDialogs({
      fetchUserById: vi.fn(),
      showLoadUserError: vi.fn()
    })

    state.openCleanupDialog()
    expect(state.cleanupDialogVisible.value).toBe(true)
    state.closeCleanupDialog()
    expect(state.cleanupDialogVisible.value).toBe(false)

    state.openBalanceHistory(createUser())
    expect(state.showBalanceHistoryModal.value).toBe(true)
    expect(state.balanceHistoryUser.value?.id).toBe(7)
    state.closeBalanceHistoryModal()
    expect(state.showBalanceHistoryModal.value).toBe(false)
    expect(state.balanceHistoryUser.value).toBeNull()
  })

  it('loads the clicked user and reports load failures', async () => {
    const fetchedUser = createUser({ id: 9 })
    const fetchUserById = vi.fn().mockResolvedValue(fetchedUser)
    const showLoadUserError = vi.fn()

    const state = useUsageViewDialogs({
      fetchUserById,
      showLoadUserError
    })

    await state.handleUserClick(9)
    expect(fetchUserById).toHaveBeenCalledWith(9)
    expect(state.showBalanceHistoryModal.value).toBe(true)
    expect(state.balanceHistoryUser.value?.id).toBe(9)

    fetchUserById.mockRejectedValueOnce(new Error('boom'))
    await state.handleUserClick(10)
    expect(showLoadUserError).toHaveBeenCalledTimes(1)
  })
})
