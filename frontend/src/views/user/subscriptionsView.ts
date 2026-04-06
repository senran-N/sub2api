import type { UserSubscription } from '@/types'
import { formatDateOnly } from '@/utils/format'

type Translate = (key: string, params?: Record<string, unknown>) => string

export function buildSubscriptionProgressWidth(
  used: number | undefined,
  limit: number | null | undefined
): string {
  if (!limit || limit === 0) {
    return '0%'
  }

  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

export function buildSubscriptionProgressBarClass(
  used: number | undefined,
  limit: number | null | undefined
): string {
  if (!limit || limit === 0) {
    return 'subscription-usage-card__progress-bar subscription-usage-card__progress-bar--neutral'
  }

  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) {
    return 'subscription-usage-card__progress-bar subscription-usage-card__progress-bar--danger'
  }
  if (percentage >= 70) {
    return 'subscription-usage-card__progress-bar subscription-usage-card__progress-bar--warning'
  }

  return 'subscription-usage-card__progress-bar subscription-usage-card__progress-bar--success'
}

export function formatSubscriptionExpirationDate(
  expiresAt: string,
  now: Date,
  t: Translate
): string {
  const expires = new Date(expiresAt)
  const todayStart = new Date(now)
  todayStart.setHours(0, 0, 0, 0)
  const expiresDayStart = new Date(expires)
  expiresDayStart.setHours(0, 0, 0, 0)
  const days = Math.round((expiresDayStart.getTime() - todayStart.getTime()) / 86400000)

  if (days < 0) {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateOnly(expires)
  if (days === 0) {
    return `${dateStr} (Today)`
  }
  if (days === 1) {
    return `${dateStr} (Tomorrow)`
  }

  return `${t('userSubscriptions.daysRemaining', { days })} (${dateStr})`
}

export function resolveSubscriptionExpirationClass(expiresAt: string, now: Date): string {
  const expires = new Date(expiresAt)
  const todayStart = new Date(now)
  todayStart.setHours(0, 0, 0, 0)
  const expiresDayStart = new Date(expires)
  expiresDayStart.setHours(0, 0, 0, 0)
  const days = Math.round((expiresDayStart.getTime() - todayStart.getTime()) / 86400000)

  if (days <= 0) {
    return 'subscription-usage-card__expiration subscription-usage-card__expiration--expired'
  }
  if (days <= 3) {
    return 'subscription-usage-card__expiration subscription-usage-card__expiration--urgent'
  }
  if (days <= 7) {
    return 'subscription-usage-card__expiration subscription-usage-card__expiration--warning'
  }

  return 'subscription-usage-card__expiration subscription-usage-card__expiration--default'
}

export function formatSubscriptionResetTime(
  windowStart: string | null,
  windowHours: number,
  now: Date,
  t: Translate
): string {
  if (!windowStart) {
    return t('userSubscriptions.windowNotActive')
  }

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 3600000)
  const diff = end.getTime() - now.getTime()

  if (diff <= 0) {
    return t('userSubscriptions.windowNotActive')
  }

  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)

  if (hours > 24) {
    const days = Math.floor(hours / 24)
    const remainingHours = hours % 24
    return `${days}d ${remainingHours}h`
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  return `${minutes}m`
}

export function hasSubscriptionLimits(subscription: UserSubscription): boolean {
  return Boolean(
    subscription.group?.daily_limit_usd ||
      subscription.group?.weekly_limit_usd ||
      subscription.group?.monthly_limit_usd
  )
}
