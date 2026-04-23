import type { CreateOrderRequest, CreateOrderResult, OrderType } from '@/types/payment'

const PAYMENT_SNAPSHOT_STORAGE_KEY = 'sub2api.payment.snapshot'
const PAYMENT_SNAPSHOT_TTL_MS = 30 * 60 * 1000

export interface PaymentSnapshot {
  amount: number
  orderType: OrderType
  paymentType: string
  planId?: number
  updatedAt: number
}

export interface PaymentResumeInput {
  wechatResumeToken?: string
  openid?: string
}

export interface PaymentViewState {
  orderId: number
  amount: number
  qrCode: string
  expiresAt: string
  paymentType: string
  payUrl: string
  clientSecret: string
  payAmount: number
  orderType: OrderType | ''
}

export interface ResolvePaymentLaunchOptions {
  paymentType: string
  orderType: OrderType
  isMobile: boolean
}

export interface PaymentLaunchDecision {
  kind: 'stripe' | 'oauth_redirect' | 'jsapi' | 'mobile_redirect' | 'qr' | 'popup' | 'error'
  paymentState: PaymentViewState
  redirectUrl?: string
  jsapiParams?: Record<string, string>
  reason?: string
}

interface WeChatBridgeInvokeResult {
  err_msg?: string
  errMsg?: string
  [key: string]: unknown
}

interface WeChatBridge {
  invoke: (
    name: string,
    params: Record<string, string>,
    callback: (result: WeChatBridgeInvokeResult) => void,
  ) => void
}

declare global {
  interface Window {
    WeixinJSBridge?: WeChatBridge
  }
}

export function createPaymentSnapshot(input: {
  amount: number
  orderType: OrderType
  paymentType: string
  planId?: number
}): PaymentSnapshot {
  return {
    amount: input.amount,
    orderType: input.orderType,
    paymentType: input.paymentType,
    planId: input.planId,
    updatedAt: Date.now(),
  }
}

export function persistPaymentSnapshot(snapshot: PaymentSnapshot) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(PAYMENT_SNAPSHOT_STORAGE_KEY, JSON.stringify(snapshot))
  } catch (_err: unknown) {
    // Ignore storage failures and continue with the in-memory flow.
  }
}

export function readPaymentSnapshot(): PaymentSnapshot | null {
  if (typeof window === 'undefined') return null
  try {
    const raw = window.localStorage.getItem(PAYMENT_SNAPSHOT_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<PaymentSnapshot>
    if (
      typeof parsed.amount !== 'number'
      || (parsed.orderType !== 'balance' && parsed.orderType !== 'subscription')
      || typeof parsed.paymentType !== 'string'
      || typeof parsed.updatedAt !== 'number'
    ) {
      clearPaymentSnapshot()
      return null
    }
    if (Date.now() - parsed.updatedAt > PAYMENT_SNAPSHOT_TTL_MS) {
      clearPaymentSnapshot()
      return null
    }
    return {
      amount: parsed.amount,
      orderType: parsed.orderType,
      paymentType: parsed.paymentType,
      planId: typeof parsed.planId === 'number' ? parsed.planId : undefined,
      updatedAt: parsed.updatedAt,
    }
  } catch (_err: unknown) {
    clearPaymentSnapshot()
    return null
  }
}

export function clearPaymentSnapshot() {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.removeItem(PAYMENT_SNAPSHOT_STORAGE_KEY)
  } catch (_err: unknown) {
    // Ignore storage failures and continue with the in-memory flow.
  }
}

export function buildCreateOrderPayload(input: {
  amount: number
  orderType: OrderType
  paymentType: string
  planId?: number
  isMobile: boolean
  resume?: PaymentResumeInput | null
}): CreateOrderRequest {
  const returnUrl = typeof window !== 'undefined'
    ? `${window.location.origin}/payment/result`
    : undefined

  return {
    amount: input.amount,
    payment_type: input.paymentType,
    order_type: input.orderType,
    plan_id: input.planId,
    is_mobile: input.isMobile,
    return_url: returnUrl,
    payment_source: input.resume?.wechatResumeToken ? 'wechat_in_app_resume' : undefined,
    wechat_resume_token: input.resume?.wechatResumeToken,
    openid: input.resume?.openid,
  }
}

export function buildPaymentState(
  result: CreateOrderResult,
  paymentType: string,
  orderType: OrderType,
): PaymentViewState {
  return {
    orderId: result.order_id,
    amount: result.amount,
    qrCode: result.qr_code || '',
    expiresAt: result.expires_at || '',
    paymentType,
    payUrl: result.pay_url || '',
    clientSecret: result.client_secret || '',
    payAmount: result.pay_amount,
    orderType,
  }
}

export function resolvePaymentLaunch(
  result: CreateOrderResult,
  options: ResolvePaymentLaunchOptions,
): PaymentLaunchDecision {
  const paymentState = buildPaymentState(result, options.paymentType, options.orderType)
  const resultType = result.result_type || 'standard'
  const redirectUrl = result.redirect_url || result.pay_url

  if (resultType === 'oauth_required') {
    return {
      kind: redirectUrl ? 'oauth_redirect' : 'error',
      paymentState,
      redirectUrl,
      reason: redirectUrl ? undefined : 'missing_oauth_redirect',
    }
  }

  if (result.client_secret) {
    return { kind: 'stripe', paymentState }
  }

  if (resultType === 'jsapi_ready') {
    if (canUseWeChatJSAPI() && result.jsapi_params) {
      return {
        kind: 'jsapi',
        paymentState,
        jsapiParams: result.jsapi_params,
      }
    }
    if (paymentState.qrCode) {
      return {
        kind: 'qr',
        paymentState,
        reason: 'jsapi_unavailable',
      }
    }
    if (paymentState.payUrl) {
      return {
        kind: options.isMobile ? 'mobile_redirect' : 'popup',
        paymentState,
        redirectUrl: paymentState.payUrl,
        reason: 'jsapi_unavailable',
      }
    }
    return {
      kind: 'error',
      paymentState,
      reason: 'jsapi_unavailable',
    }
  }

  if (options.isMobile && paymentState.payUrl) {
    return {
      kind: 'mobile_redirect',
      paymentState,
      redirectUrl: paymentState.payUrl,
    }
  }

  if (paymentState.qrCode) {
    return {
      kind: 'qr',
      paymentState,
    }
  }

  if (paymentState.payUrl) {
    return {
      kind: 'popup',
      paymentState,
      redirectUrl: paymentState.payUrl,
    }
  }

  return {
    kind: 'error',
    paymentState,
    reason: 'missing_payment_launch_data',
  }
}

export function isWechatEmbeddedBrowser(): boolean {
  if (typeof navigator === 'undefined') return false
  return /MicroMessenger/i.test(navigator.userAgent)
}

export function canUseWeChatJSAPI(): boolean {
  if (typeof window === 'undefined') return false
  return isWechatEmbeddedBrowser()
}

async function waitForWeChatBridge(timeoutMs = 3000): Promise<WeChatBridge> {
  if (typeof window === 'undefined') {
    throw new Error('window_unavailable')
  }
  if (window.WeixinJSBridge) {
    return window.WeixinJSBridge
  }
  return await new Promise<WeChatBridge>((resolve, reject) => {
    let timer: ReturnType<typeof setTimeout> | null = null

    const cleanup = () => {
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      if (timer) {
        clearTimeout(timer)
      }
    }

    const handleReady = () => {
      if (!window.WeixinJSBridge) {
        return
      }
      cleanup()
      resolve(window.WeixinJSBridge)
    }

    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    timer = setTimeout(() => {
      cleanup()
      reject(new Error('wechat_jsapi_timeout'))
    }, timeoutMs)
  })
}

export async function invokeWeChatJSAPI(
  params: Record<string, string>,
): Promise<WeChatBridgeInvokeResult> {
  const bridge = await waitForWeChatBridge()
  return await new Promise<WeChatBridgeInvokeResult>((resolve) => {
    bridge.invoke('getBrandWCPayRequest', params, (result) => {
      resolve(result)
    })
  })
}

export function interpretWeChatJSAPIResult(
  result: WeChatBridgeInvokeResult,
): 'success' | 'cancel' | 'fail' | 'unknown' {
  const raw = String(result.err_msg || result.errMsg || '').toLowerCase()
  if (!raw) return 'unknown'
  if (raw.includes(':ok')) return 'success'
  if (raw.includes(':cancel')) return 'cancel'
  if (raw.includes(':fail')) return 'fail'
  return 'unknown'
}
