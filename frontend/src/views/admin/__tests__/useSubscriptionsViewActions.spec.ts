import { beforeEach, describe, expect, it, vi } from 'vitest'
import { reactive, ref } from 'vue'
import type { UserSubscription } from '@/types'
import {
  createDefaultAssignSubscriptionForm,
  createDefaultExtendSubscriptionForm
} from '../subscriptions/subscriptionForm'
import { useSubscriptionsViewActions } from '../subscriptions/useSubscriptionsViewActions'

const { assignSubscription, extendSubscription, revokeSubscription, resetQuota } = vi.hoisted(
  () => ({
    assignSubscription: vi.fn(),
    extendSubscription: vi.fn(),
    revokeSubscription: vi.fn(),
    resetQuota: vi.fn()
  })
)

vi.mock('@/api/admin', () => ({
  adminAPI: {
    subscriptions: {
      assign: assignSubscription,
      extend: extendSubscription,
      revoke: revokeSubscription,
      resetQuota
    }
  }
}))

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

function createActions() {
  const assignForm = reactive(createDefaultAssignSubscriptionForm())
  const extendForm = reactive(createDefaultExtendSubscriptionForm())
  const extendingSubscription = ref<UserSubscription | null>(null)
  const revokingSubscription = ref<UserSubscription | null>(null)
  const resettingSubscription = ref<UserSubscription | null>(null)
  const openExtendModal = vi.fn((subscription: UserSubscription) => {
    extendingSubscription.value = subscription
  })
  const openRevokeDialog = vi.fn((subscription: UserSubscription) => {
    revokingSubscription.value = subscription
  })
  const openResetQuotaConfirm = vi.fn((subscription: UserSubscription) => {
    resettingSubscription.value = subscription
  })
  const closeAssignModal = vi.fn()
  const closeExtendModal = vi.fn(() => {
    extendingSubscription.value = null
  })
  const closeRevokeDialog = vi.fn(() => {
    revokingSubscription.value = null
  })
  const closeResetQuotaConfirm = vi.fn(() => {
    resettingSubscription.value = null
  })
  const reloadSubscriptions = vi.fn()
  const showSuccess = vi.fn()
  const showError = vi.fn()

  const actions = useSubscriptionsViewActions({
    assignForm,
    extendForm,
    extendingSubscription,
    revokingSubscription,
    resettingSubscription,
    openExtendModal,
    openRevokeDialog,
    openResetQuotaConfirm,
    closeAssignModal,
    closeExtendModal,
    closeRevokeDialog,
    closeResetQuotaConfirm,
    reloadSubscriptions,
    t: (key: string) => key,
    showSuccess,
    showError
  })

  return {
    actions,
    assignForm,
    extendForm,
    extendingSubscription,
    revokingSubscription,
    resettingSubscription,
    openExtendModal,
    openRevokeDialog,
    openResetQuotaConfirm,
    closeAssignModal,
    closeExtendModal,
    closeRevokeDialog,
    closeResetQuotaConfirm,
    reloadSubscriptions,
    showSuccess,
    showError
  }
}

describe('useSubscriptionsViewActions', () => {
  beforeEach(() => {
    assignSubscription.mockReset()
    extendSubscription.mockReset()
    revokeSubscription.mockReset()
    resetQuota.mockReset()
  })

  it('validates and assigns subscriptions', async () => {
    const setup = createActions()

    await setup.actions.handleAssignSubscription()
    expect(setup.showError).toHaveBeenCalledWith('admin.subscriptions.pleaseSelectUser')

    setup.assignForm.user_id = 2
    await setup.actions.handleAssignSubscription()
    expect(setup.showError).toHaveBeenCalledWith('admin.subscriptions.pleaseSelectGroup')

    setup.assignForm.group_id = 3
    setup.assignForm.validity_days = 30
    assignSubscription.mockResolvedValue(createSubscription())

    await setup.actions.handleAssignSubscription()

    expect(assignSubscription).toHaveBeenCalledWith({
      user_id: 2,
      group_id: 3,
      validity_days: 30
    })
    expect(setup.closeAssignModal).toHaveBeenCalledTimes(1)
    expect(setup.reloadSubscriptions).toHaveBeenCalledTimes(1)
    expect(setup.showSuccess).toHaveBeenCalledWith('admin.subscriptions.subscriptionAssigned')
  })

  it('opens and confirms extend/revoke/reset quota actions', async () => {
    const setup = createActions()
    const subscription = createSubscription()

    setup.actions.handleExtend(subscription)
    expect(setup.openExtendModal).toHaveBeenCalledWith(subscription)
    extendSubscription.mockResolvedValue(subscription)
    await setup.actions.handleExtendSubscription()
    expect(extendSubscription).toHaveBeenCalledWith(1, { days: 30 })
    expect(setup.closeExtendModal).toHaveBeenCalledTimes(1)

    setup.actions.handleRevoke(subscription)
    expect(setup.openRevokeDialog).toHaveBeenCalledWith(subscription)
    revokeSubscription.mockResolvedValue({ message: 'ok' })
    await setup.actions.confirmRevoke()
    expect(revokeSubscription).toHaveBeenCalledWith(1)
    expect(setup.closeRevokeDialog).toHaveBeenCalledTimes(1)

    setup.actions.handleResetQuota(subscription)
    expect(setup.openResetQuotaConfirm).toHaveBeenCalledWith(subscription)
    resetQuota.mockResolvedValue(subscription)
    await setup.actions.confirmResetQuota()
    expect(resetQuota).toHaveBeenCalledWith(1, {
      daily: true,
      weekly: true,
      monthly: true
    })
    expect(setup.closeResetQuotaConfirm).toHaveBeenCalledTimes(1)
  })
})
