import { describe, expect, it } from 'vitest'
import type { Group, RedeemCode } from '@/types'
import {
  buildGeneratedRedeemCodesText,
  buildRedeemExportFilters,
  buildRedeemListFilters,
  buildRedeemSubscriptionGroupOptions,
  createDefaultRedeemFilters,
  createDefaultRedeemGenerationForm,
  getGeneratedRedeemTextareaHeight,
  resetRedeemGenerationSubscriptionFields,
  syncRedeemGenerationFormValue
} from '../redeem/redeemForm'

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

function createCode(code: string): RedeemCode {
  return {
    id: Number(code.length),
    code,
    type: 'balance',
    value: 10,
    status: 'unused',
    used_by: null,
    used_at: null,
    created_at: '2026-04-04T00:00:00Z'
  }
}

describe('redeemForm helpers', () => {
  it('creates and serializes filters', () => {
    const filters = createDefaultRedeemFilters()
    filters.type = 'subscription'
    filters.status = 'unused'

    expect(buildRedeemListFilters(filters, 'vip')).toEqual({
      type: 'subscription',
      status: 'unused',
      search: 'vip'
    })
    expect(buildRedeemExportFilters(filters)).toEqual({
      type: 'subscription',
      status: 'unused'
    })
  })

  it('syncs and resets generation form state', () => {
    const form = createDefaultRedeemGenerationForm()

    form.type = 'invitation'
    syncRedeemGenerationFormValue(form)
    expect(form.value).toBe(0)

    form.type = 'balance'
    syncRedeemGenerationFormValue(form)
    expect(form.value).toBe(10)

    form.group_id = 7
    form.validity_days = 90
    resetRedeemGenerationSubscriptionFields(form)
    expect(form.group_id).toBeNull()
    expect(form.validity_days).toBe(30)
  })

  it('maps subscription groups and derives generated code presentation', () => {
    expect(
      buildRedeemSubscriptionGroupOptions([
        createGroup(),
        createGroup({ id: 2, name: 'Standard', subscription_type: 'standard' }),
        createGroup({ id: 3, name: 'Anthropic', platform: 'anthropic' })
      ])
    ).toEqual([
      {
        value: 1,
        label: 'Pro',
        description: 'subscription plan',
        platform: 'openai',
        subscriptionType: 'subscription',
        rate: 1.5
      },
      {
        value: 3,
        label: 'Anthropic',
        description: 'subscription plan',
        platform: 'anthropic',
        subscriptionType: 'subscription',
        rate: 1.5
      }
    ])

    expect(buildGeneratedRedeemCodesText([createCode('AAA'), createCode('BBB')])).toBe('AAA\nBBB')
    expect(getGeneratedRedeemTextareaHeight(0)).toBe(60)
    expect(getGeneratedRedeemTextareaHeight(2)).toBe(72)
    expect(getGeneratedRedeemTextareaHeight(20)).toBe(240)
  })
})
