<template>
  <component :is="isPopup ? 'div' : AppLayout" :class="isPopup ? 'payment-popup-surface' : ''">
    <div class="payment-page payment-page--medium" :class="isPopup ? '' : 'py-8'">
      <div v-if="loading" class="payment-loading">
        <div class="payment-spinner"></div>
      </div>
      <div v-else-if="initError" class="payment-panel payment-panel--center">
        <div class="payment-status-icon payment-status-icon--danger mx-auto mb-4">
          <Icon name="exclamationCircle" size="xl" />
        </div>
        <h3 class="payment-title">{{ t('payment.stripeLoadFailed') }}</h3>
        <p class="payment-status-description mt-2">{{ initError }}</p>
        <button class="btn btn-primary mt-6" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
      </div>
      <template v-else>
        <!-- Amount header -->
        <div v-if="order" class="payment-panel payment-panel--tight overflow-hidden">
          <div class="payment-hero payment-hero--stripe">
            <p class="payment-hero__label">{{ t('payment.actualPay') }}</p>
            <p class="payment-hero__value">&#165;{{ order.pay_amount.toFixed(2) }}</p>
          </div>
        </div>

        <!-- WeChat QR Code display -->
        <template v-if="wechatQrUrl">
          <div class="payment-panel">
            <div class="payment-status-block">
              <p class="payment-title">{{ t('payment.qr.scanWxpay') }}</p>
              <div class="payment-qr-shell payment-qr-shell--wxpay">
                <img :src="wechatQrUrl" alt="WeChat Pay QR" class="payment-qr-image h-56 w-56" />
                <div class="pointer-events-none absolute inset-0 flex items-center justify-center">
                  <span class="payment-qr-shell__logo payment-qr-shell--wxpay">
                    <svg class="h-5 w-5 text-white" viewBox="0 0 24 24" fill="currentColor"><path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm3.636 4.35c-2.084 0-3.993.672-5.363 1.844-1.188.982-2.004 2.308-2.004 3.862 0 1.207.546 2.355 1.483 3.285.114.113.238.213.358.321l-.105.42c-.021.084-.042.17-.042.253 0 .168.126.258.282.258.065 0 .126-.025.18-.058l1.27-.765a.69.69 0 0 1 .58-.086c.96.282 1.99.437 3.043.437 2.633 0 5.03-.972 6.4-2.5.782-.87 1.258-1.901 1.258-3.006 0-3.328-3.325-6.006-7.34-6.006zm-3.21 3.09c.52 0 .94.429.94.957a.949.949 0 0 1-.94.955.949.949 0 0 1-.94-.955c0-.528.42-.957.94-.957zm4.739 0c.52 0 .94.429.94.957a.949.949 0 0 1-.94.955.949.949 0 0 1-.94-.955c0-.528.42-.957.94-.957z"/></svg>
                  </span>
                </div>
              </div>
              <p class="payment-status-description">{{ t('payment.qr.scanWxpayHint') }}</p>
            </div>
          </div>
          <div class="payment-panel payment-panel--tight payment-panel--center">
            <p class="payment-countdown__label">{{ t('payment.qr.waitingPayment') }}</p>
          </div>
        </template>

        <!-- Alipay redirecting state -->
        <template v-else-if="redirecting">
          <div class="payment-panel">
            <div class="payment-status-block">
              <div class="payment-spinner payment-spinner--md"></div>
              <p class="payment-status-description">{{ t('payment.qr.payInNewWindowHint') }}</p>
            </div>
          </div>
        </template>

        <!-- Success state -->
        <template v-else-if="stripeSuccess">
          <div class="payment-panel payment-panel--center">
            <div class="payment-status-block">
              <div class="payment-status-icon payment-status-icon--success">
                <Icon name="check" size="lg" />
              </div>
              <p class="payment-status-title">{{ t('payment.result.success') }}</p>
              <p class="payment-status-description">{{ t('payment.stripeSuccessProcessing') }}</p>
            </div>
          </div>
        </template>

        <!-- Fallback: full Payment Element (no method param or unknown method) -->
        <template v-else-if="showPaymentElement">
          <div class="payment-panel">
            <div id="stripe-payment-element" class="min-h-[200px]"></div>
            <p v-if="stripeError" class="payment-feedback payment-feedback--danger mt-4">{{ stripeError }}</p>
            <button class="btn payment-submit-button payment-submit-button--stripe mt-6 w-full py-3 text-base" :disabled="stripeSubmitting || !stripeReady" @click="handleGenericPay">
              <span v-if="stripeSubmitting" class="flex items-center justify-center gap-2">
                <span class="payment-spinner payment-spinner--sm"></span>
                {{ t('common.processing') }}
              </span>
              <span v-else>{{ t('payment.stripePay') }}</span>
            </button>
          </div>
          <div class="text-center">
            <button class="btn btn-secondary" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
          </div>
        </template>

        <!-- Error -->
        <div v-if="stripeError && !showPaymentElement" class="payment-panel payment-panel--tight">
          <p class="payment-feedback payment-feedback--danger">{{ stripeError }}</p>
          <button class="btn btn-secondary mt-3 w-full" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
        </div>
      </template>
    </div>
  </component>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { usePaymentStore } from '@/stores/payment'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import type { PaymentOrder } from '@/types/payment'
import type { Stripe, StripeElements } from '@stripe/stripe-js'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import '@/components/payment/paymentTheme.css'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const paymentStore = usePaymentStore()

// Popup mode: skip AppLayout when opened with a specific method (alipay/wechat_pay)
const isPopup = computed(() => !!route.query.method)

const loading = ref(true)
const initError = ref('')
const stripeError = ref('')
const stripeSubmitting = ref(false)
const stripeSuccess = ref(false)
const stripeReady = ref(false)
const order = ref<PaymentOrder | null>(null)
const wechatQrUrl = ref('')
const redirecting = ref(false)
const showPaymentElement = ref(false)

let stripeInstance: Stripe | null = null
let elementsInstance: StripeElements | null = null
let redirectTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  const orderId = Number(route.query.order_id)
  const clientSecret = String(route.query.client_secret || '')
  const method = String(route.query.method || '')

  if (!orderId || !clientSecret) {
    loading.value = false
    initError.value = t('payment.stripeMissingParams')
    return
  }

  try {
    const res = await paymentAPI.getOrder(orderId)
    order.value = res.data

    await paymentStore.fetchConfig()
    const publishableKey = paymentStore.config?.stripe_publishable_key
    if (!publishableKey) { initError.value = t('payment.stripeNotConfigured'); return }

    const { loadStripe } = await import('@stripe/stripe-js')
    const stripe = await loadStripe(publishableKey)
    if (!stripe) { initError.value = t('payment.stripeLoadFailed'); return }

    stripeInstance = stripe
    loading.value = false

    // Direct confirm for specific methods (no Payment Element needed)
    if (method === 'alipay') {
      await confirmAlipay(stripe, clientSecret, orderId)
    } else if (method === 'wechat_pay') {
      await confirmWechatPay(stripe, clientSecret)
    } else {
      // Fallback: render full Payment Element
      showPaymentElement.value = true
      await nextTick()
      mountPaymentElement(stripe, clientSecret)
    }
  } catch (err: unknown) {
    initError.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.stripeLoadFailed'))
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (redirectTimer) clearTimeout(redirectTimer)
})

async function confirmAlipay(stripe: Stripe, clientSecret: string, orderId: number) {
  redirecting.value = true
  const returnUrl = window.location.origin + '/payment/result?order_id=' + orderId + '&status=success'
  const { error } = await stripe.confirmAlipayPayment(clientSecret, { return_url: returnUrl })
  if (error) {
    redirecting.value = false
    stripeError.value = error.message || t('payment.result.failed')
  }
  // If no error, Stripe redirects automatically — nothing else to do
}

async function confirmWechatPay(stripe: Stripe, clientSecret: string) {
  const { paymentIntent, error } = await (stripe as Stripe & {
    confirmWechatPayPayment: (cs: string, opts: Record<string, unknown>) => Promise<{ paymentIntent?: { status: string; next_action?: { wechat_pay_display_qr_code?: { image_data_url?: string } } }; error?: { message?: string } }>
  }).confirmWechatPayPayment(clientSecret, {
    payment_method_options: { wechat_pay: { client: isMobileDevice() ? 'mobile_web' : 'web' } },
  })

  if (error) {
    stripeError.value = error.message || t('payment.result.failed')
    return
  }

  // Extract QR code image from next_action
  const qrData = paymentIntent?.next_action?.wechat_pay_display_qr_code?.image_data_url
  if (qrData) {
    wechatQrUrl.value = qrData
    // Poll for completion
    startPolling()
  } else if (paymentIntent?.status === 'succeeded') {
    stripeSuccess.value = true
    scheduleClose()
  } else {
    stripeError.value = t('payment.result.failed')
  }
}

function mountPaymentElement(stripe: Stripe, clientSecret: string) {
  const isDark = document.documentElement.classList.contains('dark')
  const elements = stripe.elements({
    clientSecret,
    appearance: { theme: isDark ? 'night' : 'stripe', variables: { borderRadius: '8px' } },
  })
  elementsInstance = elements
  const paymentElement = elements.create('payment', {
    layout: 'tabs',
    paymentMethodOrder: ['alipay', 'wechat_pay', 'card', 'link'],
  } as Record<string, unknown>)
  paymentElement.mount('#stripe-payment-element')
  paymentElement.on('ready', () => { stripeReady.value = true })
}

async function handleGenericPay() {
  if (!stripeInstance || !elementsInstance || stripeSubmitting.value) return
  stripeSubmitting.value = true
  stripeError.value = ''
  try {
    const { error } = await stripeInstance.confirmPayment({
      elements: elementsInstance,
      confirmParams: {
        return_url: window.location.origin + '/payment/result?order_id=' + route.query.order_id + '&status=success',
      },
      redirect: 'if_required',
    })
    if (error) {
      stripeError.value = error.message || t('payment.result.failed')
    } else {
      stripeSuccess.value = true
      scheduleClose()
    }
  } catch (err: unknown) {
    stripeError.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.result.failed'))
  } finally {
    stripeSubmitting.value = false
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null

function startPolling() {
  const orderId = Number(route.query.order_id)
  if (!orderId) return
  pollTimer = setInterval(async () => {
    const o = await paymentStore.pollOrderStatus(orderId)
    if (!o) return
    if (o.status === 'COMPLETED' || o.status === 'PAID') {
      if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
      stripeSuccess.value = true
      wechatQrUrl.value = ''
      scheduleClose()
    }
  }, 3000)
}

function scheduleClose() {
  if (window.opener) {
    redirectTimer = setTimeout(() => { window.close() }, 2000)
  } else {
    redirectTimer = setTimeout(() => {
      router.push({ path: '/payment/result', query: { order_id: String(route.query.order_id || ''), status: 'success' } })
    }, 2000)
  }
}

onUnmounted(() => {
  if (redirectTimer) clearTimeout(redirectTimer)
  if (pollTimer) clearInterval(pollTimer)
})
</script>
