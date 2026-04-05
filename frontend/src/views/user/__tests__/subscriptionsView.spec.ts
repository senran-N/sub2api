import { describe, expect, it } from 'vitest'
import {
  buildSubscriptionProgressBarClass,
  buildSubscriptionProgressWidth,
  formatSubscriptionExpirationDate,
  formatSubscriptionResetTime,
  hasSubscriptionLimits,
  resolveSubscriptionExpirationClass
} from '../subscriptionsView'

const t = (key: string, params?: Record<string, unknown>) =>
  params ? `${key}:${JSON.stringify(params)}` : key

describe('subscriptionsView helpers', () => {
  it('builds progress widths and colors from usage ratios', () => {
    expect(buildSubscriptionProgressWidth(5, 10)).toBe('50%')
    expect(buildSubscriptionProgressWidth(5, 0)).toBe('0%')
    expect(buildSubscriptionProgressBarClass(95, 100)).toBe('bg-red-500')
    expect(buildSubscriptionProgressBarClass(75, 100)).toBe('bg-orange-500')
    expect(buildSubscriptionProgressBarClass(20, 100)).toBe('bg-green-500')
  })

  it('formats expiration labels and styles from remaining time', () => {
    const now = new Date('2026-04-05T00:00:00Z')

    expect(formatSubscriptionExpirationDate('2026-04-05T12:00:00Z', now, t)).toContain('(Today)')
    expect(formatSubscriptionExpirationDate('2026-04-06T12:00:00Z', now, t)).toContain(
      '(Tomorrow)'
    )
    expect(formatSubscriptionExpirationDate('2026-04-10T12:00:00Z', now, t)).toContain(
      'userSubscriptions.daysRemaining'
    )
    expect(resolveSubscriptionExpirationClass('2026-04-05T12:00:00Z', now)).toContain('text-red-600')
  })

  it('formats reset windows and detects unlimited subscriptions', () => {
    const now = new Date('2026-04-05T10:00:00Z')

    expect(
      formatSubscriptionResetTime('2026-04-05T00:00:00Z', 24, now, t)
    ).toBe('14h 0m')
    expect(formatSubscriptionResetTime(null, 24, now, t)).toBe('userSubscriptions.windowNotActive')

    expect(
      hasSubscriptionLimits({
        group: {
          daily_limit_usd: 10,
          weekly_limit_usd: null,
          monthly_limit_usd: null
        }
      } as never)
    ).toBe(true)

    expect(
      hasSubscriptionLimits({
        group: {
          daily_limit_usd: null,
          weekly_limit_usd: null,
          monthly_limit_usd: null
        }
      } as never)
    ).toBe(false)
  })
})
