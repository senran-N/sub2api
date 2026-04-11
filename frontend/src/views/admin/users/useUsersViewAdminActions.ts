import { ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'

interface UsersViewAdminActionsOptions {
  reloadUsers: () => void | Promise<void>
  reloadAttributeDefinitions: () => void | Promise<void>
  showSuccess: (message: string) => void
  showError: (message: string) => void
  t: (key: string, params?: Record<string, unknown>) => string
}

export function useUsersViewAdminActions(options: UsersViewAdminActionsOptions) {
  const showCreateModal = ref(false)
  const showEditModal = ref(false)
  const showDeleteDialog = ref(false)
  const showAttributesModal = ref(false)
  const editingUser = ref<AdminUser | null>(null)
  const deletingUser = ref<AdminUser | null>(null)

  const openCreateModal = () => {
    showCreateModal.value = true
  }

  const closeCreateModal = () => {
    showCreateModal.value = false
  }

  const handleEdit = (user: AdminUser) => {
    editingUser.value = user
    showEditModal.value = true
  }

  const closeEditModal = () => {
    showEditModal.value = false
    editingUser.value = null
  }

  const openAttributesModal = () => {
    showAttributesModal.value = true
  }

  const handleAttributesModalClose = async () => {
    showAttributesModal.value = false
    await options.reloadAttributeDefinitions()
    await options.reloadUsers()
  }

  const handleToggleStatus = async (user: AdminUser) => {
    const newStatus = user.status === 'active' ? 'disabled' : 'active'
    try {
      await adminAPI.users.toggleStatus(user.id, newStatus)
      options.showSuccess(
        newStatus === 'active'
          ? options.t('admin.users.userEnabled')
          : options.t('admin.users.userDisabled')
      )
      await options.reloadUsers()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.users.failedToToggle'))
      )
      console.error('Error toggling user status:', error)
    }
  }

  const handleDelete = (user: AdminUser) => {
    deletingUser.value = user
    showDeleteDialog.value = true
  }

  const closeDeleteDialog = () => {
    showDeleteDialog.value = false
  }

  const confirmDelete = async () => {
    if (!deletingUser.value) {
      return
    }

    try {
      await adminAPI.users.delete(deletingUser.value.id)
      options.showSuccess(options.t('common.success'))
      closeDeleteDialog()
      deletingUser.value = null
      await options.reloadUsers()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.users.failedToDelete'))
      )
      console.error('Error deleting user:', error)
    }
  }

  return {
    showCreateModal,
    showEditModal,
    showDeleteDialog,
    showAttributesModal,
    editingUser,
    deletingUser,
    openCreateModal,
    closeCreateModal,
    handleEdit,
    closeEditModal,
    openAttributesModal,
    handleAttributesModalClose,
    handleToggleStatus,
    handleDelete,
    closeDeleteDialog,
    confirmDelete
  }
}
