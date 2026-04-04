import type { SubscriptionListFilters } from '@/api/admin/subscriptions'
import type { SelectOption } from '@/components/common/Select.vue'
import type {
  AssignSubscriptionRequest,
  ExtendSubscriptionRequest,
  Group,
  GroupPlatform,
  SubscriptionType,
  UserSubscription
} from '@/types'

export type SubscriptionStatusFilter = '' | 'active' | 'expired' | 'revoked'
export type ResetWindowPeriod = 'daily' | 'weekly' | 'monthly'

export interface SubscriptionFiltersState {
  status: SubscriptionStatusFilter
  group_id: string
  platform: string
  user_id: number | null
}

export interface SubscriptionSortState {
  sort_by: string
  sort_order: 'asc' | 'desc'
}

export interface AssignSubscriptionForm {
  user_id: number | null
  group_id: number | null
  validity_days: number
}

export interface ExtendSubscriptionForm {
  days: number
}

export interface SubscriptionGroupOption extends SelectOption {
  value: number
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

export interface ResetWindowMessage {
  key: string
  params?: {
    days?: number
    hours?: number
    minutes?: number
  }
}

const DAY_IN_MS = 24 * 60 * 60 * 1000
const WEEK_IN_MS = 7 * DAY_IN_MS
const THIRTY_DAYS_IN_MS = 30 * DAY_IN_MS

export function createDefaultSubscriptionFilters(): SubscriptionFiltersState {
  return {
    status: 'active',
    group_id: '',
    platform: '',
    user_id: null
  }
}

export function createDefaultAssignSubscriptionForm(): AssignSubscriptionForm {
  return {
    user_id: null,
    group_id: null,
    validity_days: 30
  }
}

export function createDefaultExtendSubscriptionForm(): ExtendSubscriptionForm {
  return {
    days: 30
  }
}

export function resetAssignSubscriptionForm(form: AssignSubscriptionForm): void {
  Object.assign(form, createDefaultAssignSubscriptionForm())
}

export function resetExtendSubscriptionForm(form: ExtendSubscriptionForm): void {
  Object.assign(form, createDefaultExtendSubscriptionForm())
}

export function buildSubscriptionListFilters(
  filters: SubscriptionFiltersState,
  sortState: SubscriptionSortState
): SubscriptionListFilters {
  return {
    status: filters.status || undefined,
    group_id: filters.group_id ? Number.parseInt(filters.group_id, 10) : undefined,
    platform: filters.platform || undefined,
    user_id: filters.user_id || undefined,
    sort_by: sortState.sort_by,
    sort_order: sortState.sort_order
  }
}

export function buildAssignSubscriptionRequest(
  form: AssignSubscriptionForm
): AssignSubscriptionRequest {
  return {
    user_id: form.user_id as number,
    group_id: form.group_id as number,
    validity_days: form.validity_days
  }
}

export function buildExtendSubscriptionRequest(
  form: ExtendSubscriptionForm
): ExtendSubscriptionRequest {
  return {
    days: form.days
  }
}

export function buildSubscriptionGroupOptions(groups: Group[]): SubscriptionGroupOption[] {
  return groups
    .filter((group) => group.subscription_type === 'subscription' && group.status === 'active')
    .map((group) => ({
      value: group.id,
      label: group.name,
      description: group.description,
      platform: group.platform,
      subscriptionType: group.subscription_type,
      rate: group.rate_multiplier
    }))
}

export function getSubscriptionDaysRemaining(
  expiresAt: string | null | undefined,
  now: Date = new Date()
): number | null {
  if (!expiresAt) {
    return null
  }

  const diff = new Date(expiresAt).getTime() - now.getTime()
  if (diff < 0) {
    return null
  }

  return Math.ceil(diff / DAY_IN_MS)
}

export function isSubscriptionExpiringSoon(
  expiresAt: string | null | undefined,
  now?: Date
): boolean {
  const daysRemaining = getSubscriptionDaysRemaining(expiresAt, now)
  return daysRemaining !== null && daysRemaining <= 7
}

export function willSubscriptionAdjustmentRemainActive(
  subscription: Pick<UserSubscription, 'expires_at'>,
  days: number,
  now: Date = new Date()
): boolean {
  if (!subscription.expires_at) {
    return true
  }

  const newExpiresAt = new Date(subscription.expires_at).getTime() + days * DAY_IN_MS
  return newExpiresAt > now.getTime()
}

export function getUsageProgressWidth(
  used: number | null | undefined,
  limit: number | null
): string {
  if (!limit || limit === 0) {
    return '0%'
  }

  const percentage = Math.min(((used ?? 0) / limit) * 100, 100)
  return `${percentage}%`
}

export function getUsageProgressClass(
  used: number | null | undefined,
  limit: number | null
): string {
  if (!limit || limit === 0) {
    return 'bg-gray-400'
  }

  const percentage = ((used ?? 0) / limit) * 100
  if (percentage >= 90) {
    return 'bg-red-500'
  }
  if (percentage >= 70) {
    return 'bg-orange-500'
  }
  return 'bg-green-500'
}

export function getResetWindowMessage(
  windowStart: string | null | undefined,
  period: ResetWindowPeriod,
  now: Date = new Date()
): ResetWindowMessage {
  if (!windowStart) {
    return { key: 'admin.subscriptions.windowNotActive' }
  }

  const startTime = new Date(windowStart).getTime()
  const duration =
    period === 'daily' ? DAY_IN_MS : period === 'weekly' ? WEEK_IN_MS : THIRTY_DAYS_IN_MS
  const diffSeconds = Math.floor((startTime + duration - now.getTime()) / 1000)

  if (diffSeconds <= 0) {
    return { key: 'admin.subscriptions.windowNotActive' }
  }

  const days = Math.floor(diffSeconds / 86400)
  const hours = Math.floor((diffSeconds % 86400) / 3600)
  const minutes = Math.floor((diffSeconds % 3600) / 60)

  if (days > 0) {
    return {
      key: 'admin.subscriptions.resetInDaysHours',
      params: { days, hours }
    }
  }

  if (hours > 0) {
    return {
      key: 'admin.subscriptions.resetInHoursMinutes',
      params: { hours, minutes }
    }
  }

  return {
    key: 'admin.subscriptions.resetInMinutes',
    params: { minutes }
  }
}
