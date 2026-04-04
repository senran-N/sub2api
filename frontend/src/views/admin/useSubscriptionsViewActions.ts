import { ref, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { UserSubscription } from '@/types'
import {
  buildAssignSubscriptionRequest,
  buildExtendSubscriptionRequest,
  willSubscriptionAdjustmentRemainActive,
  type AssignSubscriptionForm,
  type ExtendSubscriptionForm
} from './subscriptionForm'

interface SubscriptionsViewActionsOptions {
  assignForm: AssignSubscriptionForm
  extendForm: ExtendSubscriptionForm
  extendingSubscription: Ref<UserSubscription | null>
  revokingSubscription: Ref<UserSubscription | null>
  resettingSubscription: Ref<UserSubscription | null>
  openExtendModal: (subscription: UserSubscription) => void
  openRevokeDialog: (subscription: UserSubscription) => void
  openResetQuotaConfirm: (subscription: UserSubscription) => void
  closeAssignModal: () => void
  closeExtendModal: () => void
  closeRevokeDialog: () => void
  closeResetQuotaConfirm: () => void
  reloadSubscriptions: () => void | Promise<void>
  t: (key: string) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
}

export function useSubscriptionsViewActions(options: SubscriptionsViewActionsOptions) {
  const submitting = ref(false)
  const resettingQuota = ref(false)

  const handleAssignSubscription = async () => {
    if (!options.assignForm.user_id) {
      options.showError(options.t('admin.subscriptions.pleaseSelectUser'))
      return
    }
    if (!options.assignForm.group_id) {
      options.showError(options.t('admin.subscriptions.pleaseSelectGroup'))
      return
    }
    if (!options.assignForm.validity_days || options.assignForm.validity_days < 1) {
      options.showError(options.t('admin.subscriptions.validityDaysRequired'))
      return
    }

    submitting.value = true
    try {
      await adminAPI.subscriptions.assign(buildAssignSubscriptionRequest(options.assignForm))
      options.showSuccess(options.t('admin.subscriptions.subscriptionAssigned'))
      options.closeAssignModal()
      await options.reloadSubscriptions()
    } catch (error: any) {
      options.showError(error.response?.data?.detail || options.t('admin.subscriptions.failedToAssign'))
      console.error('Error assigning subscription:', error)
    } finally {
      submitting.value = false
    }
  }

  const handleExtend = (subscription: UserSubscription) => {
    options.openExtendModal(subscription)
  }

  const handleExtendSubscription = async () => {
    if (!options.extendingSubscription.value) {
      return
    }

    if (
      !willSubscriptionAdjustmentRemainActive(
        options.extendingSubscription.value,
        options.extendForm.days
      )
    ) {
      options.showError(options.t('admin.subscriptions.adjustWouldExpire'))
      return
    }

    submitting.value = true
    try {
      await adminAPI.subscriptions.extend(
        options.extendingSubscription.value.id,
        buildExtendSubscriptionRequest(options.extendForm)
      )
      options.showSuccess(options.t('admin.subscriptions.subscriptionAdjusted'))
      options.closeExtendModal()
      await options.reloadSubscriptions()
    } catch (error: any) {
      options.showError(error.response?.data?.detail || options.t('admin.subscriptions.failedToAdjust'))
      console.error('Error adjusting subscription:', error)
    } finally {
      submitting.value = false
    }
  }

  const handleRevoke = (subscription: UserSubscription) => {
    options.openRevokeDialog(subscription)
  }

  const confirmRevoke = async () => {
    if (!options.revokingSubscription.value) {
      return
    }

    try {
      await adminAPI.subscriptions.revoke(options.revokingSubscription.value.id)
      options.showSuccess(options.t('admin.subscriptions.subscriptionRevoked'))
      options.closeRevokeDialog()
      await options.reloadSubscriptions()
    } catch (error: any) {
      options.showError(error.response?.data?.detail || options.t('admin.subscriptions.failedToRevoke'))
      console.error('Error revoking subscription:', error)
    }
  }

  const handleResetQuota = (subscription: UserSubscription) => {
    options.openResetQuotaConfirm(subscription)
  }

  const confirmResetQuota = async () => {
    if (!options.resettingSubscription.value || resettingQuota.value) {
      return
    }

    resettingQuota.value = true
    try {
      await adminAPI.subscriptions.resetQuota(options.resettingSubscription.value.id, {
        daily: true,
        weekly: true,
        monthly: true
      })
      options.showSuccess(options.t('admin.subscriptions.quotaResetSuccess'))
      options.closeResetQuotaConfirm()
      await options.reloadSubscriptions()
    } catch (error: any) {
      options.showError(
        error.response?.data?.detail || options.t('admin.subscriptions.failedToResetQuota')
      )
      console.error('Error resetting quota:', error)
    } finally {
      resettingQuota.value = false
    }
  }

  return {
    submitting,
    resettingQuota,
    handleAssignSubscription,
    handleExtend,
    handleExtendSubscription,
    handleRevoke,
    confirmRevoke,
    handleResetQuota,
    confirmResetQuota
  }
}
