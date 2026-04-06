import type { RedeemHistoryItem } from '@/api'

type Translate = (key: string, params?: Record<string, unknown>) => string

export interface RedeemResultData {
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
}

export interface RedeemHistoryPresentation {
  iconName: 'dollar' | 'badge' | 'bolt'
  tone: 'success' | 'danger' | 'brand' | 'info' | 'warning'
}

interface RedeemErrorLike {
  response?: {
    data?: {
      detail?: string
    }
  }
}

export function formatRedeemBalance(balance: number | null | undefined): string {
  return `$${Number(balance || 0).toFixed(2)}`
}

export function isBalanceRedeemType(type: string): boolean {
  return type === 'balance' || type === 'admin_balance'
}

export function isSubscriptionRedeemType(type: string): boolean {
  return type === 'subscription'
}

export function isAdminAdjustmentRedeemType(type: string): boolean {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

export function resolveRedeemHistoryTitle(
  item: RedeemHistoryItem,
  t: Translate
): string {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  }
  if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  }
  if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  }
  if (item.type === 'admin_concurrency') {
    return item.value >= 0
      ? t('redeem.concurrencyAddedAdmin')
      : t('redeem.concurrencyReducedAdmin')
  }
  if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }

  return t('common.unknown')
}

export function formatRedeemHistoryValue(
  item: RedeemHistoryItem,
  t: Translate
): string {
  if (isBalanceRedeemType(item.type)) {
    const sign = item.value >= 0 ? '+' : '-'
    return `${sign}$${Math.abs(item.value).toFixed(2)}`
  }

  if (isSubscriptionRedeemType(item.type)) {
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  }

  const sign = item.value >= 0 ? '+' : '-'
  return `${sign}${Math.abs(item.value)} ${t('redeem.requests')}`
}

export function buildRedeemHistoryPresentation(item: RedeemHistoryItem): RedeemHistoryPresentation {
  if (isBalanceRedeemType(item.type)) {
    return item.value >= 0
      ? {
          iconName: 'dollar',
          tone: 'success'
        }
      : {
          iconName: 'dollar',
          tone: 'danger'
        }
  }

  if (isSubscriptionRedeemType(item.type)) {
    return {
      iconName: 'badge',
      tone: 'brand'
    }
  }

  return item.value >= 0
    ? {
        iconName: 'bolt',
        tone: 'info'
      }
    : {
        iconName: 'bolt',
        tone: 'warning'
      }
}

export function resolveRedeemErrorMessage(
  error: unknown,
  fallback: string
): string {
  const redeemError = error as RedeemErrorLike | null
  return redeemError?.response?.data?.detail || fallback
}
