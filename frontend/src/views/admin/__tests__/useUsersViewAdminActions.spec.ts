import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { AdminUser } from '@/types'
import { useUsersViewAdminActions } from '../users/useUsersViewAdminActions'

const { toggleStatus, deleteUser } = vi.hoisted(() => ({
  toggleStatus: vi.fn(),
  deleteUser: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      toggleStatus,
      delete: deleteUser
    }
  }
}))

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

describe('useUsersViewAdminActions', () => {
  const reloadUsers = vi.fn()
  const reloadAttributeDefinitions = vi.fn()
  const showSuccess = vi.fn()
  const showError = vi.fn()
  const t = (key: string) => key

  beforeEach(() => {
    reloadUsers.mockReset()
    reloadAttributeDefinitions.mockReset()
    showSuccess.mockReset()
    showError.mockReset()
    toggleStatus.mockReset()
    deleteUser.mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('manages modal visibility for create, edit, and attributes dialogs', async () => {
    const actions = useUsersViewAdminActions({
      reloadUsers,
      reloadAttributeDefinitions,
      showSuccess,
      showError,
      t
    })
    const user = createAdminUser()

    actions.openCreateModal()
    expect(actions.showCreateModal.value).toBe(true)
    actions.closeCreateModal()
    expect(actions.showCreateModal.value).toBe(false)

    actions.handleEdit(user)
    expect(actions.showEditModal.value).toBe(true)
    expect(actions.editingUser.value?.id).toBe(1)
    actions.closeEditModal()
    expect(actions.editingUser.value).toBeNull()

    actions.openAttributesModal()
    expect(actions.showAttributesModal.value).toBe(true)
    await actions.handleAttributesModalClose()
    expect(actions.showAttributesModal.value).toBe(false)
    expect(reloadAttributeDefinitions).toHaveBeenCalledTimes(1)
    expect(reloadUsers).toHaveBeenCalledTimes(1)
  })

  it('toggles user status and reloads the list on success', async () => {
    toggleStatus.mockResolvedValue({})
    const actions = useUsersViewAdminActions({
      reloadUsers,
      reloadAttributeDefinitions,
      showSuccess,
      showError,
      t
    })

    await actions.handleToggleStatus(createAdminUser({ id: 7, status: 'disabled' }))

    expect(toggleStatus).toHaveBeenCalledWith(7, 'active')
    expect(showSuccess).toHaveBeenCalledWith('admin.users.userEnabled')
    expect(reloadUsers).toHaveBeenCalledTimes(1)
  })

  it('surfaces toggle errors without reloading', async () => {
    const error = {
      response: {
        data: {
          detail: 'toggle failed'
        }
      }
    }
    toggleStatus.mockRejectedValue(error)
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const actions = useUsersViewAdminActions({
      reloadUsers,
      reloadAttributeDefinitions,
      showSuccess,
      showError,
      t
    })

    await actions.handleToggleStatus(createAdminUser({ id: 9 }))

    expect(showError).toHaveBeenCalledWith('toggle failed')
    expect(reloadUsers).not.toHaveBeenCalled()
    expect(consoleSpy).toHaveBeenCalled()
  })

  it('confirms delete, closes the dialog, and reloads on success', async () => {
    deleteUser.mockResolvedValue({})
    const actions = useUsersViewAdminActions({
      reloadUsers,
      reloadAttributeDefinitions,
      showSuccess,
      showError,
      t
    })

    actions.handleDelete(createAdminUser({ id: 4 }))
    expect(actions.showDeleteDialog.value).toBe(true)
    expect(actions.deletingUser.value?.id).toBe(4)

    await actions.confirmDelete()

    expect(deleteUser).toHaveBeenCalledWith(4)
    expect(actions.showDeleteDialog.value).toBe(false)
    expect(actions.deletingUser.value).toBeNull()
    expect(showSuccess).toHaveBeenCalledWith('common.success')
    expect(reloadUsers).toHaveBeenCalledTimes(1)
  })

  it('falls back to plain error messages for toggle and delete failures', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const actions = useUsersViewAdminActions({
      reloadUsers,
      reloadAttributeDefinitions,
      showSuccess,
      showError,
      t
    })

    toggleStatus.mockRejectedValueOnce(new Error('network down'))
    await actions.handleToggleStatus(createAdminUser({ id: 10 }))

    actions.handleDelete(createAdminUser({ id: 11 }))
    deleteUser.mockRejectedValueOnce(new Error('delete unavailable'))
    await actions.confirmDelete()

    expect(showError).toHaveBeenNthCalledWith(1, 'network down')
    expect(showError).toHaveBeenNthCalledWith(2, 'delete unavailable')
    expect(consoleSpy).toHaveBeenCalledTimes(2)
  })
})
