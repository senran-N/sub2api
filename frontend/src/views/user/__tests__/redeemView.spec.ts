import { describe, expect, it } from 'vitest'
import type { RedeemHistoryItem } from '@/api'
import {
  buildRedeemHistoryPresentation,
  formatRedeemBalance,
  formatRedeemHistoryValue,
  isAdminAdjustmentRedeemType,
  isBalanceRedeemType,
  isSubscriptionRedeemType,
  resolveRedeemErrorMessage,
  resolveRedeemHistoryTitle
} from '../redeem/redeemView'

const t = (key: string) => key

const createHistoryItem = (overrides: Partial<RedeemHistoryItem>): RedeemHistoryItem => ({
  id: 1,
  code: 'ABCDEFGH12345678',
  type: 'balance',
  value: 10,
  status: 'used',
  used_at: '2026-04-05T00:00:00Z',
  created_at: '2026-04-05T00:00:00Z',
  ...overrides
})

describe('redeemView helpers', () => {
  it('formats balance display consistently', () => {
    expect(formatRedeemBalance(12.5)).toBe('$12.50')
    expect(formatRedeemBalance(undefined)).toBe('$0.00')
  })

  it('detects redeem history item categories', () => {
    expect(isBalanceRedeemType('balance')).toBe(true)
    expect(isBalanceRedeemType('admin_balance')).toBe(true)
    expect(isSubscriptionRedeemType('subscription')).toBe(true)
    expect(isAdminAdjustmentRedeemType('admin_concurrency')).toBe(true)
    expect(isAdminAdjustmentRedeemType('concurrency')).toBe(false)
  })

  it('builds history titles, values, and presentation tones', () => {
    const balanceItem = createHistoryItem({ type: 'admin_balance', value: -2.5 })
    expect(resolveRedeemHistoryTitle(balanceItem, t)).toBe('redeem.balanceDeductedAdmin')
    expect(formatRedeemHistoryValue(balanceItem, t)).toBe('-$2.50')
    expect(buildRedeemHistoryPresentation(balanceItem).iconName).toBe('dollar')

    const subscriptionItem = createHistoryItem({
      type: 'subscription',
      value: 30,
      validity_days: 30,
      group: { id: 2, name: 'Pro' }
    })
    expect(resolveRedeemHistoryTitle(subscriptionItem, t)).toBe('redeem.subscriptionAssigned')
    expect(formatRedeemHistoryValue(subscriptionItem, t)).toBe('30redeem.days - Pro')
    expect(buildRedeemHistoryPresentation(subscriptionItem).iconName).toBe('badge')

    const concurrencyItem = createHistoryItem({ type: 'concurrency', value: 5 })
    expect(formatRedeemHistoryValue(concurrencyItem, t)).toBe('+5 redeem.requests')
    expect(buildRedeemHistoryPresentation(concurrencyItem).iconName).toBe('bolt')
  })

  it('resolves redeem API errors from backend detail first', () => {
    expect(
      resolveRedeemErrorMessage({ response: { data: { detail: 'detail message' } } }, 'fallback')
    ).toBe('detail message')
    expect(resolveRedeemErrorMessage({}, 'fallback')).toBe('fallback')
  })
})
