import { ref } from 'vue'
import type { AdminUser } from '@/types'

interface UsageViewDialogsOptions {
  fetchUserById: (userId: number) => Promise<AdminUser>
  showLoadUserError: () => void
}

export function useUsageViewDialogs(options: UsageViewDialogsOptions) {
  const cleanupDialogVisible = ref(false)
  const showBalanceHistoryModal = ref(false)
  const balanceHistoryUser = ref<AdminUser | null>(null)

  const openCleanupDialog = () => {
    cleanupDialogVisible.value = true
  }

  const closeCleanupDialog = () => {
    cleanupDialogVisible.value = false
  }

  const openBalanceHistory = (user: AdminUser) => {
    balanceHistoryUser.value = user
    showBalanceHistoryModal.value = true
  }

  const closeBalanceHistoryModal = () => {
    showBalanceHistoryModal.value = false
    balanceHistoryUser.value = null
  }

  const handleUserClick = async (userId: number) => {
    try {
      const user = await options.fetchUserById(userId)
      openBalanceHistory(user)
    } catch {
      options.showLoadUserError()
    }
  }

  return {
    cleanupDialogVisible,
    showBalanceHistoryModal,
    balanceHistoryUser,
    openCleanupDialog,
    closeCleanupDialog,
    openBalanceHistory,
    closeBalanceHistoryModal,
    handleUserClick
  }
}
