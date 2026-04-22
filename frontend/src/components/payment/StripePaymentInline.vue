<template>
  <div class="payment-page">
    <div v-if="loading" class="payment-loading">
      <div class="payment-spinner"></div>
    </div>
    <div v-else-if="initError" class="payment-panel payment-panel--center">
      <div class="payment-status-block !py-0">
        <div class="payment-status-icon payment-status-icon--danger">
          <Icon name="x" size="lg" />
        </div>
        <p class="payment-status-description">{{ initError }}</p>
      </div>
      <button class="btn btn-secondary mt-4" @click="$emit('back')">{{ t('payment.result.backToRecharge') }}</button>
    </div>
    <template v-else-if="success">
      <div class="payment-panel">
        <div class="payment-status-block">
          <div class="payment-status-icon payment-status-icon--success">
            <Icon name="check" size="lg" />
          </div>
          <p class="payment-status-title">{{ t('payment.result.success') }}</p>
          <div class="payment-panel payment-panel--soft w-full">
            <div class="payment-detail-list">
              <div class="payment-detail-row">
                <span class="payment-detail-label">{{ t('payment.orders.orderId') }}</span>
                <span class="payment-detail-value">#{{ orderId }}</span>
              </div>
              <div v-if="amount > 0" class="payment-detail-row">
                <span class="payment-detail-label">{{ t('payment.orders.amount') }}</span>
                <span class="payment-detail-value">{{ orderType === 'balance' ? '$' : '¥' }}{{ amount.toFixed(2) }}</span>
              </div>
              <div class="payment-detail-row">
                <span class="payment-detail-label">{{ t('payment.orders.payAmount') }}</span>
                <span class="payment-detail-value payment-detail-value--strong">¥{{ payAmount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
          <button class="btn btn-primary" @click="$emit('done')">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>
    <template v-else>
      <div class="payment-panel overflow-hidden p-0">
        <div class="payment-hero payment-hero--stripe">
          <p class="payment-hero__label">{{ t('payment.actualPay') }}</p>
          <p class="payment-hero__value">¥{{ payAmount.toFixed(2) }}</p>
        </div>
      </div>
      <div class="payment-panel">
        <div ref="stripeMount" class="min-h-[200px]"></div>
        <p v-if="error" class="payment-feedback payment-feedback--danger mt-4">{{ error }}</p>
        <button class="btn payment-submit-button payment-submit-button--stripe mt-6 w-full py-3 text-base" :disabled="submitting || !ready" @click="handlePay">
          <span v-if="submitting" class="flex items-center justify-center gap-2">
            <span class="payment-spinner payment-spinner--sm"></span>
            {{ t('common.processing') }}
          </span>
          <span v-else>{{ t('payment.stripePay') }}</span>
        </button>
      </div>
      <!-- Cancel order -->
      <button class="btn btn-secondary w-full" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import type { Stripe, StripeElements } from '@stripe/stripe-js'
import Icon from '@/components/icons/Icon.vue'
import '@/components/payment/paymentTheme.css'

// Stripe payment methods that open a popup (redirect or QR code)
const POPUP_METHODS = new Set(['alipay', 'wechat_pay'])

const props = defineProps<{
  orderId: number
  amount: number
  clientSecret: string
  orderType?: 'balance' | 'subscription'
  publishableKey: string
  payAmount: number
}>()

const emit = defineEmits<{ success: []; done: []; back: []; redirect: [orderId: number, payUrl: string] }>()

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const stripeMount = ref<HTMLElement | null>(null)
const loading = ref(true)
const initError = ref('')
const error = ref('')
const submitting = ref(false)
const cancelling = ref(false)
const success = ref(false)
const ready = ref(false)
const selectedType = ref('')

let stripeInstance: Stripe | null = null
let elementsInstance: StripeElements | null = null

onMounted(async () => {
  try {
    const { loadStripe } = await import('@stripe/stripe-js')
    const stripe = await loadStripe(props.publishableKey)
    if (!stripe) { initError.value = t('payment.stripeLoadFailed'); return }

    stripeInstance = stripe
    loading.value = false
    await nextTick()
    if (!stripeMount.value) return

    const isDark = document.documentElement.classList.contains('dark')
    const elements = stripe.elements({
      clientSecret: props.clientSecret,
      appearance: { theme: isDark ? 'night' : 'stripe', variables: { borderRadius: '8px' } },
    })
    elementsInstance = elements
    const paymentElement = elements.create('payment', {
      layout: 'tabs',
      paymentMethodOrder: ['alipay', 'wechat_pay', 'card', 'link'],
    } as Record<string, unknown>)
    paymentElement.mount(stripeMount.value)
    paymentElement.on('ready', () => { ready.value = true })
    paymentElement.on('change', (event: { value: { type: string } }) => {
      selectedType.value = event.value.type
    })
  } catch (err: unknown) {
    initError.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.stripeLoadFailed'))
  } finally {
    loading.value = false
  }
})

async function handlePay() {
  if (!stripeInstance || !elementsInstance || submitting.value) return

  // Alipay / WeChat Pay: open popup for redirect or QR display
  if (POPUP_METHODS.has(selectedType.value)) {
    const popupUrl = router.resolve({
      path: '/payment/stripe-popup',
      query: {
        order_id: String(props.orderId),
        method: selectedType.value,
        amount: String(props.payAmount),
      },
    }).href
    const popup = window.open(popupUrl, 'paymentPopup', getPaymentPopupFeatures())

    const onReady = (event: MessageEvent) => {
      if (event.source !== popup || event.data?.type !== 'STRIPE_POPUP_READY') return
      window.removeEventListener('message', onReady)
      popup?.postMessage({
        type: 'STRIPE_POPUP_INIT',
        clientSecret: props.clientSecret,
        publishableKey: props.publishableKey,
      }, window.location.origin)
    }
    window.addEventListener('message', onReady)

    emit('redirect', props.orderId, popupUrl)
    return
  }

  // Card / Link: confirm inline
  submitting.value = true
  error.value = ''
  try {
    const { error: stripeError } = await stripeInstance.confirmPayment({
      elements: elementsInstance,
      confirmParams: {
        return_url: window.location.origin + '/payment/result?order_id=' + props.orderId + '&status=success',
      },
      redirect: 'if_required',
    })
    if (stripeError) {
      error.value = stripeError.message || t('payment.result.failed')
    } else {
      success.value = true
      emit('success')
    }
  } catch (err: unknown) {
    error.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.result.failed'))
  } finally {
    submitting.value = false
  }
}

async function handleCancel() {
  if (!props.orderId || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(props.orderId)
    emit('back')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}
</script>
