import type { LocationQuery, LocationQueryRaw } from 'vue-router'
import type { OrderType } from '@/types/payment'
import type { PaymentSnapshot } from '@/components/payment/paymentFlow'

const PAYMENT_WECHAT_RESUME_QUERY_KEYS = [
  'wechat_resume',
  'wechat_resume_token',
  'openid',
  'state',
  'scope',
  'payment_type',
  'amount',
  'order_type',
  'plan_id',
] as const

export interface PaymentWechatResumeQuery {
  active: boolean
  wechatResumeToken?: string
  openid?: string
  state?: string
  scope?: string
  paymentType?: string
  amount?: number
  orderType?: OrderType
  planId?: number
}

export interface PaymentWechatResumeIntent {
  shouldResume: boolean
  resume: {
    wechatResumeToken?: string
    openid?: string
  }
  paymentType?: string
  amount?: number
  orderType: OrderType
  planId?: number
}

function readQueryString(query: LocationQuery, key: string): string {
  const value = query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function parsePositiveNumber(value: string): number | undefined {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return undefined
  }
  return parsed
}

function normalizeOrderType(value: string, fallback: OrderType): OrderType {
  return value === 'subscription' ? 'subscription' : fallback
}

export function readPaymentWechatResumeQuery(query: LocationQuery): PaymentWechatResumeQuery {
  const wechatResume = readQueryString(query, 'wechat_resume')
  const wechatResumeToken = readQueryString(query, 'wechat_resume_token')
  const openid = readQueryString(query, 'openid')
  const rawOrderType = readQueryString(query, 'order_type')

  return {
    active: wechatResume === '1' || Boolean(wechatResumeToken) || Boolean(openid),
    wechatResumeToken: wechatResumeToken || undefined,
    openid: openid || undefined,
    state: readQueryString(query, 'state') || undefined,
    scope: readQueryString(query, 'scope') || undefined,
    paymentType: readQueryString(query, 'payment_type') || undefined,
    amount: parsePositiveNumber(readQueryString(query, 'amount')),
    orderType: rawOrderType === 'balance' || rawOrderType === 'subscription'
      ? rawOrderType
      : undefined,
    planId: parsePositiveNumber(readQueryString(query, 'plan_id')),
  }
}

export function stripPaymentWechatResumeQuery(query: LocationQuery): LocationQueryRaw {
  const stripped: LocationQueryRaw = {}

  for (const [key, value] of Object.entries(query)) {
    if (PAYMENT_WECHAT_RESUME_QUERY_KEYS.includes(key as typeof PAYMENT_WECHAT_RESUME_QUERY_KEYS[number])) {
      continue
    }
    stripped[key] = value
  }

  return stripped
}

export function resolvePaymentWechatResumeIntent(
  query: LocationQuery,
  snapshot: PaymentSnapshot | null,
): PaymentWechatResumeIntent | null {
  const parsed = readPaymentWechatResumeQuery(query)
  if (!parsed.active) {
    return null
  }

  const snapshotOrderType = snapshot?.orderType === 'subscription' ? 'subscription' : 'balance'
  const orderType = normalizeOrderType(parsed.orderType || '', snapshotOrderType)

  return {
    shouldResume: true,
    resume: {
      wechatResumeToken: parsed.wechatResumeToken,
      openid: parsed.openid,
    },
    paymentType: parsed.paymentType || snapshot?.paymentType || undefined,
    amount: parsed.amount ?? (orderType === 'balance' ? snapshot?.amount : undefined),
    orderType,
    planId: parsed.planId ?? (orderType === 'subscription' ? snapshot?.planId : undefined),
  }
}
