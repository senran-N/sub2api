<template>
  <div class="payment-status-panel">
    <!-- ═══ Terminal States: show result, user clicks to return ═══ -->

    <!-- Success -->
    <template v-if="outcome === 'success'">
      <div class="payment-status-panel__card">
        <div class="payment-status-panel__state">
          <div class="payment-status-panel__state-icon-shell payment-status-panel__state-icon-shell--success">
            <Icon name="check" size="lg" class="payment-status-panel__state-icon payment-status-panel__state-icon--success" />
          </div>
          <p class="payment-status-panel__state-title">{{ props.orderType === 'subscription' ? t('payment.result.subscriptionSuccess') : t('payment.result.success') }}</p>
          <div v-if="paidOrder" class="payment-status-panel__summary">
            <div class="payment-status-panel__summary-list">
              <div class="payment-status-panel__summary-row">
                <span class="payment-status-panel__summary-label">{{ t('payment.orders.orderId') }}</span>
                <span class="payment-status-panel__summary-value">#{{ paidOrder.id }}</span>
              </div>
              <div v-if="paidOrder.out_trade_no" class="payment-status-panel__summary-row">
                <span class="payment-status-panel__summary-label">{{ t('payment.orders.orderNo') }}</span>
                <span class="payment-status-panel__summary-value">{{ paidOrder.out_trade_no }}</span>
              </div>
              <div class="payment-status-panel__summary-row">
                <span class="payment-status-panel__summary-label">{{ t('payment.orders.amount') }}</span>
                <span class="payment-status-panel__summary-value">{{ paidOrder.order_type === 'balance' ? '$' : '¥' }}{{ paidOrder.amount.toFixed(2) }}</span>
              </div>
              <div class="payment-status-panel__summary-row">
                <span class="payment-status-panel__summary-label">{{ t('payment.orders.payAmount') }}</span>
                <span class="payment-status-panel__summary-value">¥{{ paidOrder.pay_amount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Cancelled -->
    <template v-else-if="outcome === 'cancelled'">
      <div class="payment-status-panel__card">
        <div class="payment-status-panel__state">
          <div class="payment-status-panel__state-icon-shell payment-status-panel__state-icon-shell--muted">
            <svg class="payment-status-panel__state-icon payment-status-panel__state-icon--muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <p class="payment-status-panel__state-title">{{ t('payment.qr.cancelled') }}</p>
          <p class="payment-status-panel__state-copy">{{ t('payment.qr.cancelledDesc') }}</p>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Expired / Failed -->
    <template v-else-if="outcome === 'expired'">
      <div class="payment-status-panel__card">
        <div class="payment-status-panel__state">
          <div class="payment-status-panel__state-icon-shell payment-status-panel__state-icon-shell--warning">
            <svg class="payment-status-panel__state-icon payment-status-panel__state-icon--warning" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p class="payment-status-panel__state-title">{{ t('payment.qr.expired') }}</p>
          <p class="payment-status-panel__state-copy">{{ t('payment.qr.expiredDesc') }}</p>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- ═══ Active States: QR or Popup waiting ═══ -->

    <!-- QR Code Mode -->
    <template v-else-if="qrUrl">
      <div class="payment-status-panel__card">
        <div class="payment-status-panel__state">
          <p class="payment-status-panel__state-title payment-status-panel__state-title--compact">{{ scanTitle }}</p>
          <div :class="['payment-status-panel__qr-shell', qrBorderClass]">
            <canvas ref="qrCanvas" class="mx-auto"></canvas>
            <!-- Brand logo overlay -->
            <div class="pointer-events-none absolute inset-0 flex items-center justify-center">
              <span :class="['payment-status-panel__qr-logo', qrLogoBgClass]">
                <img :src="isAlipay ? alipayIcon : wxpayIcon" alt="" class="h-5 w-5 brightness-0 invert" />
              </span>
            </div>
          </div>
          <p v-if="scanHint" class="payment-status-panel__state-copy payment-status-panel__state-copy--center">{{ scanHint }}</p>
        </div>
      </div>
      <div class="payment-status-panel__countdown">
        <p class="payment-status-panel__countdown-label">{{ t('payment.qr.expiresIn') }}</p>
        <p class="payment-status-panel__countdown-value">{{ countdownDisplay }}</p>
        <p class="payment-status-panel__countdown-hint">{{ t('payment.qr.waitingPayment') }}</p>
      </div>
      <button class="btn btn-secondary w-full" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>

    <!-- Waiting for Popup/Redirect Mode -->
    <template v-else>
      <div class="payment-status-panel__card">
        <div class="payment-status-panel__state">
          <div class="payment-status-panel__spinner"></div>
          <p class="payment-status-panel__state-copy payment-status-panel__state-copy--center">{{ t('payment.qr.payInNewWindowHint') }}</p>
          <button v-if="payUrl" class="btn btn-secondary text-sm" @click="reopenPopup">
            {{ t('payment.qr.openPayWindow') }}
          </button>
        </div>
      </div>
      <div class="payment-status-panel__countdown">
        <p class="payment-status-panel__countdown-value">{{ countdownDisplay }}</p>
        <p class="payment-status-panel__countdown-hint">{{ t('payment.qr.waitingPayment') }}</p>
      </div>
      <button class="btn btn-secondary w-full" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePaymentStore } from '@/stores/payment'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import type { PaymentOrder } from '@/types/payment'
import Icon from '@/components/icons/Icon.vue'
import QRCode from 'qrcode'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'

const props = defineProps<{
  orderId: number
  qrCode: string
  expiresAt: string
  paymentType: string
  payUrl?: string
  orderType?: string
}>()

const emit = defineEmits<{ done: []; success: [] }>()

const { t } = useI18n()
const paymentStore = usePaymentStore()
const appStore = useAppStore()

const qrCanvas = ref<HTMLCanvasElement | null>(null)
const qrUrl = ref('')
const remainingSeconds = ref(0)
const cancelling = ref(false)
const paidOrder = ref<PaymentOrder | null>(null)

// Terminal outcome: null = still active, 'success' | 'cancelled' | 'expired'
const outcome = ref<'success' | 'cancelled' | 'expired' | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null

const isAlipay = computed(() => props.paymentType.includes('alipay'))
const isWxpay = computed(() => props.paymentType.includes('wxpay'))

const qrBorderClass = computed(() => {
  if (isAlipay.value) return 'payment-status-panel__qr-shell--alipay'
  if (isWxpay.value) return 'payment-status-panel__qr-shell--wxpay'
  return 'payment-status-panel__qr-shell--default'
})

const qrLogoBgClass = computed(() => {
  if (isAlipay.value) return 'payment-status-panel__qr-logo--alipay'
  if (isWxpay.value) return 'payment-status-panel__qr-logo--wxpay'
  return 'payment-status-panel__qr-logo--default'
})

const scanTitle = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipay')
  if (isWxpay.value) return t('payment.qr.scanWxpay')
  return t('payment.qr.scanToPay')
})

const scanHint = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipayHint')
  if (isWxpay.value) return t('payment.qr.scanWxpayHint')
  return ''
})

const countdownDisplay = computed(() => {
  const m = Math.floor(remainingSeconds.value / 60)
  const s = remainingSeconds.value % 60
  return m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0')
})

function reopenPopup() {
  if (props.payUrl) {
    window.open(props.payUrl, 'paymentPopup', getPaymentPopupFeatures())
  }
}

async function renderQR() {
  await nextTick()
  if (!qrCanvas.value || !qrUrl.value) return
  await QRCode.toCanvas(qrCanvas.value, qrUrl.value, {
    width: 220, margin: 2,
    errorCorrectionLevel: 'M',
  })
}

async function pollStatus() {
  if (!props.orderId || outcome.value) return
  const order = await paymentStore.pollOrderStatus(props.orderId)
  if (!order) return
  if (order.status === 'COMPLETED' || order.status === 'PAID') {
    cleanup()
    paidOrder.value = order
    outcome.value = 'success'
    emit('success')
  } else if (order.status === 'CANCELLED') {
    cleanup()
    outcome.value = 'cancelled'
  } else if (order.status === 'EXPIRED' || order.status === 'FAILED') {
    cleanup()
    outcome.value = 'expired'
  }
}

function startCountdown(seconds: number) {
  remainingSeconds.value = Math.max(0, seconds)
  if (remainingSeconds.value <= 0) { outcome.value = 'expired'; return }
  countdownTimer = setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) { outcome.value = 'expired'; cleanup() }
  }, 1000)
}

async function handleCancel() {
  if (!props.orderId || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(props.orderId)
    cleanup()
    outcome.value = 'cancelled'
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}

function handleDone() { cleanup(); emit('done') }

function cleanup() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

// Initialize on mount
qrUrl.value = props.qrCode
let seconds = 30 * 60
if (props.expiresAt) {
  seconds = Math.floor((new Date(props.expiresAt).getTime() - Date.now()) / 1000)
}
startCountdown(seconds)
pollTimer = setInterval(pollStatus, 3000)
renderQR()

watch(() => qrUrl.value, () => renderQR())
onUnmounted(() => cleanup())
</script>

<style scoped>
.payment-status-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.payment-status-panel__card,
.payment-status-panel__countdown {
  border: var(--theme-card-border-width) solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 8px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.payment-status-panel__card {
  padding: 1.5rem;
}

.payment-status-panel__state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding-block: 1rem;
}

.payment-status-panel__state-icon-shell {
  display: flex;
  height: 4rem;
  width: 4rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
}

.payment-status-panel__state-icon-shell--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
}

.payment-status-panel__state-icon-shell--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 12%, var(--theme-surface));
}

.payment-status-panel__state-icon-shell--muted {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.payment-status-panel__state-icon {
  height: 2rem;
  width: 2rem;
}

.payment-status-panel__state-icon--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.payment-status-panel__state-icon--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.payment-status-panel__state-icon--muted {
  color: var(--theme-page-muted);
}

.payment-status-panel__state-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--theme-page-text);
}

.payment-status-panel__state-title--compact {
  font-weight: 600;
}

.payment-status-panel__state-copy,
.payment-status-panel__summary-label,
.payment-status-panel__countdown-label,
.payment-status-panel__countdown-hint {
  color: var(--theme-page-muted);
}

.payment-status-panel__state-copy {
  font-size: 0.875rem;
}

.payment-status-panel__state-copy--center {
  text-align: center;
}

.payment-status-panel__summary {
  width: 100%;
  border-radius: calc(var(--theme-surface-radius) + 6px);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
  padding: 1rem;
}

.payment-status-panel__summary-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.payment-status-panel__summary-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.payment-status-panel__summary-value {
  font-weight: 600;
  color: var(--theme-page-text);
}

.payment-status-panel__qr-shell {
  position: relative;
  border: 2px solid color-mix(in srgb, var(--theme-input-border) 88%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 6px);
  background: var(--theme-surface);
  padding: 1rem;
}

.payment-status-panel__qr-shell--alipay {
  border-color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 40%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.payment-status-panel__qr-shell--wxpay {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 40%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.payment-status-panel__qr-logo {
  border-radius: 999px;
  padding: 0.5rem;
  box-shadow: var(--theme-card-shadow);
  outline: 2px solid var(--theme-surface);
}

.payment-status-panel__qr-logo--alipay {
  background: rgb(var(--theme-info-rgb));
}

.payment-status-panel__qr-logo--wxpay {
  background: rgb(var(--theme-success-rgb));
}

.payment-status-panel__qr-logo--default {
  background: color-mix(in srgb, var(--theme-page-muted) 72%, var(--theme-page-text));
}

.payment-status-panel__countdown {
  padding: 1rem;
  text-align: center;
}

.payment-status-panel__spinner {
  height: 2.5rem;
  width: 2.5rem;
  animation: payment-status-panel-spin 1s linear infinite;
  border: 4px solid rgb(var(--theme-accent-rgb));
  border-top-color: transparent;
  border-radius: 999px;
}

@keyframes payment-status-panel-spin {
  to {
    transform: rotate(360deg);
  }
}

.payment-status-panel__countdown-value {
  margin-top: 0.25rem;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--theme-page-text);
  font-variant-numeric: tabular-nums;
}

.payment-status-panel__countdown-hint {
  margin-top: 0.25rem;
  font-size: 0.75rem;
}
</style>
