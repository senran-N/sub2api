import { describe, expect, it } from 'vitest'
import type { Group, UserSubscription } from '@/types'
import {
  buildAssignSubscriptionRequest,
  buildExtendSubscriptionRequest,
  buildSubscriptionGroupOptions,
  buildSubscriptionListFilters,
  createDefaultAssignSubscriptionForm,
  createDefaultExtendSubscriptionForm,
  createDefaultSubscriptionFilters,
  getResetWindowMessage,
  getSubscriptionDaysRemaining,
  getUsageProgressClass,
  getUsageProgressWidth,
  isSubscriptionExpiringSoon,
  resetAssignSubscriptionForm,
  resetExtendSubscriptionForm,
  validateAssignSubscriptionForm,
  validateExtendSubscriptionAdjustment,
  willSubscriptionAdjustmentRemainActive
} from '../subscriptions/subscriptionForm'

function createGroup(overrides: Partial<Group> = {}): Group {
  return {
    id: 1,
    name: 'Pro',
    description: 'subscription plan',
    platform: 'openai',
    rate_multiplier: 1.5,
    status: 'active',
    subscription_type: 'subscription',
    ...overrides
  } as Group
}

function createSubscription(
  overrides: Partial<Pick<UserSubscription, 'expires_at'>> = {}
): Pick<UserSubscription, 'expires_at'> {
  return {
    expires_at: '2026-04-10T00:00:00Z',
    ...overrides
  }
}

describe('subscriptionForm helpers', () => {
  it('creates, resets, and serializes assign and extend forms', () => {
    const assignForm = createDefaultAssignSubscriptionForm()
    assignForm.user_id = 7
    assignForm.group_id = 11
    assignForm.validity_days = 45

    expect(buildAssignSubscriptionRequest(assignForm)).toEqual({
      user_id: 7,
      group_id: 11,
      validity_days: 45
    })

    assignForm.validity_days = 1
    resetAssignSubscriptionForm(assignForm)
    expect(assignForm).toEqual(createDefaultAssignSubscriptionForm())

    const extendForm = createDefaultExtendSubscriptionForm()
    extendForm.days = -5
    expect(buildExtendSubscriptionRequest(extendForm)).toEqual({ days: -5 })

    resetExtendSubscriptionForm(extendForm)
    expect(extendForm).toEqual(createDefaultExtendSubscriptionForm())
  })

  it('normalizes list filters and maps active subscription groups into select options', () => {
    const filters = createDefaultSubscriptionFilters()
    filters.status = 'expired'
    filters.group_id = '12'
    filters.platform = 'anthropic'
    filters.user_id = 99

    expect(
      buildSubscriptionListFilters(filters, {
        sort_by: 'expires_at',
        sort_order: 'asc'
      })
    ).toEqual({
      status: 'expired',
      group_id: 12,
      platform: 'anthropic',
      user_id: 99,
      sort_by: 'expires_at',
      sort_order: 'asc'
    })

    expect(
      buildSubscriptionGroupOptions([
        createGroup(),
        createGroup({ id: 2, name: 'Disabled', status: 'disabled' }),
        createGroup({ id: 3, name: 'Standard', subscription_type: 'standard' })
      ])
    ).toEqual([
      {
        value: 1,
        label: 'Pro',
        description: 'subscription plan',
        platform: 'openai',
        subscriptionType: 'subscription',
        rate: 1.5
      }
    ])
  })

  it('computes expiration state and validates adjustment windows against a fixed now', () => {
    const now = new Date('2026-04-04T00:00:00Z')

    expect(getSubscriptionDaysRemaining('2026-04-10T00:00:00Z', now)).toBe(6)
    expect(getSubscriptionDaysRemaining('2026-04-03T23:59:59Z', now)).toBeNull()
    expect(isSubscriptionExpiringSoon('2026-04-10T00:00:00Z', now)).toBe(true)
    expect(isSubscriptionExpiringSoon('2026-04-20T00:00:00Z', now)).toBe(false)

    expect(
      willSubscriptionAdjustmentRemainActive(
        createSubscription({ expires_at: '2026-04-05T00:00:00Z' }),
        -2,
        now
      )
    ).toBe(false)
    expect(
      willSubscriptionAdjustmentRemainActive(
        createSubscription({ expires_at: '2026-04-05T00:00:00Z' }),
        5,
        now
      )
    ).toBe(true)
    expect(
      willSubscriptionAdjustmentRemainActive(
        createSubscription({ expires_at: null }),
        -30,
        now
      )
    ).toBe(true)
  })

  it('derives usage bar state and reset countdown message descriptors', () => {
    const now = new Date('2026-04-04T00:00:00Z')

    expect(getUsageProgressWidth(15, 10)).toBe('100%')
    expect(getUsageProgressWidth(2.5, 10)).toBe('25%')
    expect(getUsageProgressClass(9, 10)).toBe('theme-progress-fill--danger')
    expect(getUsageProgressClass(7.5, 10)).toBe('theme-progress-fill--warning')
    expect(getUsageProgressClass(3, 10)).toBe('theme-progress-fill--success')
    expect(getUsageProgressClass(3, null)).toBe('theme-progress-fill--muted')

    expect(getResetWindowMessage(null, 'daily', now)).toEqual({
      key: 'admin.subscriptions.windowNotActive'
    })
    expect(getResetWindowMessage('2026-04-03T00:00:00Z', 'daily', now)).toEqual({
      key: 'admin.subscriptions.windowNotActive'
    })
    expect(getResetWindowMessage('2026-04-03T20:00:00Z', 'daily', now)).toEqual({
      key: 'admin.subscriptions.resetInHoursMinutes',
      params: { hours: 20, minutes: 0 }
    })
    expect(getResetWindowMessage('2026-03-30T00:00:00Z', 'weekly', now)).toEqual({
      key: 'admin.subscriptions.resetInDaysHours',
      params: { days: 2, hours: 0 }
    })
    expect(getResetWindowMessage('2026-04-03T00:15:00Z', 'daily', now)).toEqual({
      key: 'admin.subscriptions.resetInMinutes',
      params: { minutes: 15 }
    })
  })

  it('validates assign payloads and extend adjustments before actions run', () => {
    expect(validateAssignSubscriptionForm(createDefaultAssignSubscriptionForm())).toBe(
      'admin.subscriptions.pleaseSelectUser'
    )
    expect(
      validateAssignSubscriptionForm({
        user_id: 2,
        group_id: null,
        validity_days: 30
      })
    ).toBe('admin.subscriptions.pleaseSelectGroup')
    expect(
      validateAssignSubscriptionForm({
        user_id: 2,
        group_id: 3,
        validity_days: 0
      })
    ).toBe('admin.subscriptions.validityDaysRequired')
    expect(
      validateAssignSubscriptionForm({
        user_id: 2,
        group_id: 3,
        validity_days: 30
      })
    ).toBeNull()

    expect(
      validateExtendSubscriptionAdjustment(
        createSubscription({ expires_at: '2026-04-01T00:00:00Z' }),
        { days: -10 },
        new Date('2026-04-10T00:00:00Z')
      )
    ).toBe('admin.subscriptions.adjustWouldExpire')
    expect(
      validateExtendSubscriptionAdjustment(
        createSubscription({ expires_at: '2026-04-20T00:00:00Z' }),
        { days: 5 },
        new Date('2026-04-10T00:00:00Z')
      )
    ).toBeNull()
  })
})
