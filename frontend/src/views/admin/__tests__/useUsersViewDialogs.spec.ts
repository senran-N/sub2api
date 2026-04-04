import { describe, expect, it, vi, afterEach } from 'vitest'
import type { AdminUser } from '@/types'
import {
  calculateUserActionMenuPosition,
  useUsersViewDialogs
} from '../useUsersViewDialogs'

function createAdminUser(overrides: Partial<AdminUser> = {}): AdminUser {
  return {
    id: 1,
    email: 'user@example.com',
    username: 'user',
    role: 'user',
    balance: 0,
    status: 'active',
    allowed_groups: [],
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    notes: '',
    group_rates: {},
    current_concurrency: 0,
    sora_storage_quota_bytes: 0,
    sora_storage_used_bytes: 0,
    concurrency: 1,
    ...overrides
  } as AdminUser
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('useUsersViewDialogs', () => {
  it('calculates menu positions for mobile and desktop layouts', () => {
    expect(
      calculateUserActionMenuPosition({
        rect: { left: 120, top: 220, bottom: 260, width: 40 },
        pointerX: 220,
        pointerY: 250,
        viewportWidth: 390,
        viewportHeight: 844
      })
    ).toEqual({
      left: 40,
      top: 264
    })

    expect(
      calculateUserActionMenuPosition({
        rect: { left: 600, top: 700, bottom: 730, width: 32 },
        pointerX: 950,
        pointerY: 760,
        viewportWidth: 1280,
        viewportHeight: 800
      })
    ).toEqual({
      left: 750,
      top: 552
    })
  })

  it('opens and closes the floating action menu using the current trigger element', () => {
    vi.stubGlobal('window', {
      innerWidth: 1280,
      innerHeight: 800
    })

    const dialogs = useUsersViewDialogs()
    const user = createAdminUser({ id: 7 })
    const trigger = {
      getBoundingClientRect: () => ({
        left: 900,
        top: 200,
        bottom: 240,
        width: 40
      })
    } as HTMLElement

    dialogs.openActionMenu(user, {
      currentTarget: trigger,
      clientX: 950,
      clientY: 300
    } as MouseEvent)
    expect(dialogs.activeMenuId.value).toBe(7)
    expect(dialogs.menuPosition.value).toEqual({
      left: 750,
      top: 300
    })

    dialogs.openActionMenu(user, {
      currentTarget: trigger,
      clientX: 950,
      clientY: 300
    } as MouseEvent)
    expect(dialogs.activeMenuId.value).toBeNull()
    expect(dialogs.menuPosition.value).toBeNull()
  })

  it('tracks modal state transitions for API keys, groups, balance, and history', () => {
    const dialogs = useUsersViewDialogs()
    const user = createAdminUser({ id: 11 })

    dialogs.openViewApiKeys(user)
    expect(dialogs.showApiKeysModal.value).toBe(true)
    expect(dialogs.viewingUser.value?.id).toBe(11)
    dialogs.closeApiKeysModal()
    expect(dialogs.viewingUser.value).toBeNull()

    dialogs.openAllowedGroups(user)
    expect(dialogs.showAllowedGroupsModal.value).toBe(true)
    dialogs.closeAllowedGroupsModal()
    expect(dialogs.allowedGroupsUser.value).toBeNull()

    dialogs.openBalanceModal(user, 'subtract')
    expect(dialogs.showBalanceModal.value).toBe(true)
    expect(dialogs.balanceOperation.value).toBe('subtract')
    dialogs.closeBalanceModal()
    expect(dialogs.balanceUser.value).toBeNull()

    dialogs.openBalanceHistory(user)
    dialogs.reopenBalanceFromHistory('add')
    expect(dialogs.balanceUser.value?.id).toBe(11)
    expect(dialogs.balanceOperation.value).toBe('add')
    dialogs.closeBalanceHistoryModal()
    expect(dialogs.balanceHistoryUser.value).toBeNull()
  })

  it('toggles expanded groups and replaces selected exclusive groups', () => {
    const dialogs = useUsersViewDialogs()
    const user = createAdminUser({ id: 5 })

    dialogs.toggleExpandedGroup(5)
    expect(dialogs.expandedGroupUserId.value).toBe(5)
    dialogs.toggleExpandedGroup(5)
    expect(dialogs.expandedGroupUserId.value).toBeNull()

    dialogs.toggleExpandedGroup(5)
    dialogs.openGroupReplace(user, { id: 9, name: 'VIP' })
    expect(dialogs.expandedGroupUserId.value).toBeNull()
    expect(dialogs.showGroupReplaceModal.value).toBe(true)
    expect(dialogs.groupReplaceUser.value?.id).toBe(5)
    expect(dialogs.groupReplaceOldGroup.value).toEqual({ id: 9, name: 'VIP' })

    dialogs.closeGroupReplaceModal()
    expect(dialogs.groupReplaceUser.value).toBeNull()
    expect(dialogs.groupReplaceOldGroup.value).toBeNull()
  })
})
