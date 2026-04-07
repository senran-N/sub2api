import { ref } from 'vue'
import type { UserSubscription } from '@/types'

interface SubscriptionsViewDialogsOptions {
  resetAssignState: () => void
  resetExtendState: () => void
}

export function useSubscriptionsViewDialogs(options: SubscriptionsViewDialogsOptions) {
  const showGuideModal = ref(false)
  const showAssignModal = ref(false)
  const showExtendModal = ref(false)
  const showRevokeDialog = ref(false)
  const showResetQuotaConfirm = ref(false)
  const extendingSubscription = ref<UserSubscription | null>(null)
  const revokingSubscription = ref<UserSubscription | null>(null)
  const resettingSubscription = ref<UserSubscription | null>(null)

  const openGuideModal = () => {
    showGuideModal.value = true
  }

  const closeGuideModal = () => {
    showGuideModal.value = false
  }

  const openAssignModal = () => {
    showAssignModal.value = true
  }

  const closeAssignModal = () => {
    showAssignModal.value = false
    options.resetAssignState()
  }

  const openExtendModal = (subscription: UserSubscription) => {
    extendingSubscription.value = subscription
    showExtendModal.value = true
    options.resetExtendState()
  }

  const closeExtendModal = () => {
    showExtendModal.value = false
    extendingSubscription.value = null
    options.resetExtendState()
  }

  const openRevokeDialog = (subscription: UserSubscription) => {
    revokingSubscription.value = subscription
    showRevokeDialog.value = true
  }

  const closeRevokeDialog = () => {
    showRevokeDialog.value = false
    revokingSubscription.value = null
  }

  const openResetQuotaConfirm = (subscription: UserSubscription) => {
    resettingSubscription.value = subscription
    showResetQuotaConfirm.value = true
  }

  const closeResetQuotaConfirm = () => {
    showResetQuotaConfirm.value = false
    resettingSubscription.value = null
  }

  return {
    showGuideModal,
    showAssignModal,
    showExtendModal,
    showRevokeDialog,
    showResetQuotaConfirm,
    extendingSubscription,
    revokingSubscription,
    resettingSubscription,
    openGuideModal,
    closeGuideModal,
    openAssignModal,
    closeAssignModal,
    openExtendModal,
    closeExtendModal,
    openRevokeDialog,
    closeRevokeDialog,
    openResetQuotaConfirm,
    closeResetQuotaConfirm
  }
}
