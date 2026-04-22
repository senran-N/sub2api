<template>
  <div class="payment-popup-surface stripe-popup-view" :class="methodToneClass">
    <div class="payment-page payment-page--compact payment-page--centered">
      <div class="payment-panel stripe-popup-view__card">
      <!-- Amount + Order ID -->
        <div v-if="amount" class="stripe-popup-view__header">
          <p class="stripe-popup-view__amount">¥{{ amount }}</p>
          <p v-if="orderId" class="payment-muted mt-1 text-sm">
          {{ t('payment.orders.orderId') }}: {{ orderId }}
          </p>
        </div>

        <!-- Error -->
        <div v-if="error" class="space-y-3">
          <div class="payment-feedback payment-feedback--danger">
            {{ error }}
          </div>
          <button
            class="stripe-popup-view__action"
            @click="closeWindow"
          >
            {{ t('common.close') }}
          </button>
        </div>

        <!-- Success -->
        <div v-else-if="success" class="payment-status-block">
          <div class="payment-status-icon payment-status-icon--success stripe-popup-view__success-icon">✓</div>
          <p class="payment-status-description">{{ t('payment.result.success') }}</p>
          <button
            class="stripe-popup-view__action"
            @click="closeWindow"
          >
            {{ t('common.close') }}
          </button>
        </div>

        <!-- Loading / Redirecting -->
        <div v-else class="stripe-popup-view__loading">
          <div class="payment-spinner stripe-popup-view__spinner"></div>
          <span class="payment-muted text-sm">{{ hint }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import '@/components/payment/paymentTheme.css'

interface StripeWithWechatPay {
  confirmWechatPayPayment(clientSecret: string, options: Record<string, unknown>): Promise<{ error?: { message?: string }; paymentIntent?: { status: string } }>
}

const { t } = useI18n()
const route = useRoute()

const orderId = String(route.query.order_id || '')
const method = String(route.query.method || 'alipay')
const amount = String(route.query.amount || '')

const methodToneClass = computed(() => {
  if (method === 'alipay') return 'stripe-popup-view--alipay'
  if (method === 'wechat_pay') return 'stripe-popup-view--wechat'
  return 'stripe-popup-view--default'
})

const error = ref('')
const success = ref(false)
const hint = ref(t('payment.stripePopup.redirecting'))

let pollTimer: ReturnType<typeof setInterval> | null = null

function closeWindow() { window.close() }

onMounted(() => {
  const handler = (event: MessageEvent) => {
    if (event.origin !== window.location.origin) return
    if (event.data?.type !== 'STRIPE_POPUP_INIT') return
    window.removeEventListener('message', handler)
    initStripe(event.data.clientSecret, event.data.publishableKey)
  }
  window.addEventListener('message', handler)

  if (window.opener) {
    window.opener.postMessage({ type: 'STRIPE_POPUP_READY' }, window.location.origin)
  }

  setTimeout(() => {
    if (!error.value && !success.value) {
      error.value = t('payment.stripePopup.timeout')
    }
  }, 15000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

async function initStripe(clientSecret: string, publishableKey: string) {
  if (!clientSecret || !publishableKey) {
    error.value = t('payment.stripeMissingParams')
    return
  }
  try {
    const { loadStripe } = await import('@stripe/stripe-js')
    const stripe = await loadStripe(publishableKey)
    if (!stripe) { error.value = t('payment.stripeLoadFailed'); return }

    const returnUrl = window.location.origin + '/payment/result?order_id=' + orderId + '&status=success'

    if (method === 'alipay') {
      // Alipay: redirect this popup to Alipay payment page
      const { error: err } = await stripe.confirmAlipayPayment(clientSecret, { return_url: returnUrl })
      if (err) error.value = err.message || t('payment.result.failed')
    } else if (method === 'wechat_pay') {
      // WeChat: Stripe shows its built-in QR dialog, user scans, promise resolves
      hint.value = t('payment.stripePopup.loadingQr')
      const result = await (stripe as unknown as StripeWithWechatPay).confirmWechatPayPayment(clientSecret, {
        payment_method_options: { wechat_pay: { client: isMobileDevice() ? 'mobile_web' : 'web' } },
      })
      if (result.error) {
        error.value = result.error.message || t('payment.result.failed')
      } else if (result.paymentIntent?.status === 'succeeded') {
        success.value = true
        setTimeout(closeWindow, 2000)
      } else {
        // Payment not completed (user closed QR dialog)
        startPolling()
      }
    }
  } catch (err: unknown) {
    error.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.stripeLoadFailed'))
  }
}

function startPolling() {
  pollTimer = setInterval(async () => {
    try {
      const token = document.cookie.split('; ').find(c => c.startsWith('token='))?.split('=')[1]
        || localStorage.getItem('token') || ''
      const res = await fetch('/api/v1/payment/orders/' + orderId, {
        headers: token ? { Authorization: 'Bearer ' + token } : {},
        credentials: 'include',
      })
      if (!res.ok) return
      const data = await res.json()
      const status = data?.data?.status
      if (status === 'COMPLETED' || status === 'PAID') {
        if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
        success.value = true
        setTimeout(closeWindow, 2000)
      }
    } catch { /* ignore */ }
  }, 3000)
}
</script>

<style scoped>
.stripe-popup-view--alipay {
  --stripe-popup-view-tone: rgb(var(--theme-info-rgb));
}

.stripe-popup-view--wechat {
  --stripe-popup-view-tone: rgb(var(--theme-success-rgb));
}

.stripe-popup-view--default {
  --stripe-popup-view-tone: var(--theme-accent);
}

.stripe-popup-view__card {
  width: 100%;
  max-width: 28rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.stripe-popup-view__header {
  text-align: center;
}

.stripe-popup-view__amount {
  font-size: 1.875rem;
  font-weight: 700;
  color: var(--stripe-popup-view-tone);
}

.stripe-popup-view__action {
  width: 100%;
  font-size: 0.875rem;
  text-decoration: underline;
  color: var(--stripe-popup-view-tone);
}

.stripe-popup-view__success-icon {
  font-size: 2rem;
  font-weight: 700;
}

.stripe-popup-view__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding-block: 2rem;
}

.stripe-popup-view__spinner {
  color: var(--stripe-popup-view-tone);
}
</style>
