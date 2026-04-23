<template>
  <div class="payment-result-view">
    <div class="payment-result-view__shell">
      <!-- Loading -->
      <div v-if="loading" class="payment-result-view__loading">
        <div class="payment-result-view__spinner"></div>
      </div>
      <template v-else>
        <!-- Status Icon -->
        <div class="payment-result-view__hero">
          <div v-if="isSuccess"
            class="payment-result-view__hero-icon payment-result-view__hero-icon--success">
            <svg class="payment-result-view__hero-symbol payment-result-view__hero-symbol--success" fill="none" viewBox="0 0 24 24" stroke="currentColor"
              stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <div v-else-if="isPending" class="payment-result-view__hero-icon">
            <div class="payment-result-view__spinner"></div>
          </div>
          <div v-else
            class="payment-result-view__hero-icon payment-result-view__hero-icon--danger">
            <svg class="payment-result-view__hero-symbol payment-result-view__hero-symbol--danger" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 class="payment-result-view__title">
            {{ statusTitle }}
          </h2>
        </div>
        <!-- Order Info -->
        <div v-if="order" class="payment-result-view__card">
          <div class="payment-result-view__details">
            <div class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.orderId') }}</span>
              <span class="payment-result-view__value">#{{ order.id }}</span>
            </div>
            <div v-if="order.out_trade_no" class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.orderNo') }}</span>
              <span class="payment-result-view__value">{{ order.out_trade_no }}</span>
            </div>
            <div class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.baseAmount') }}</span>
              <span class="payment-result-view__value">&#165;{{ baseAmount.toFixed(2) }}</span>
            </div>
            <div v-if="order.fee_rate > 0" class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.fee') }} ({{ order.fee_rate }}%)</span>
              <span class="payment-result-view__value">&#165;{{ feeAmount.toFixed(2) }}</span>
            </div>
            <div class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.payAmount') }}</span>
              <span class="payment-result-view__value payment-result-view__value--accent">&#165;{{ order.pay_amount.toFixed(2) }}</span>
            </div>
            <div v-if="order.amount !== order.pay_amount" class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.creditedAmount') }}</span>
              <span class="payment-result-view__value">{{ order.order_type === 'balance' ? '$' : '¥' }}{{ order.amount.toFixed(2) }}</span>
            </div>
            <div class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.paymentMethod') }}</span>
              <span class="payment-result-view__value">{{ t('payment.methods.' + order.payment_type, order.payment_type) }}</span>
            </div>
            <div class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.status') }}</span>
              <OrderStatusBadge :status="order.status" />
            </div>
          </div>
        </div>
        <!-- EasyPay return info (when no order loaded) -->
        <div v-else-if="returnInfo" class="payment-result-view__card">
          <div class="payment-result-view__details">
            <div v-if="returnInfo.outTradeNo" class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.orderId') }}</span>
              <span class="payment-result-view__value">{{ returnInfo.outTradeNo }}</span>
            </div>
            <div v-if="returnInfo.money" class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.payAmount') }}</span>
              <span class="payment-result-view__value">&#165;{{ returnInfo.money }}</span>
            </div>
            <div v-if="returnInfo.type" class="payment-result-view__row">
              <span class="payment-result-view__label">{{ t('payment.orders.paymentMethod') }}</span>
              <span class="payment-result-view__value">{{ t('payment.methods.' + returnInfo.type, returnInfo.type) }}</span>
            </div>
          </div>
        </div>
        <!-- Actions -->
        <div class="payment-result-view__actions">
          <button class="btn btn-secondary flex-1" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
          <button class="btn btn-primary flex-1" @click="router.push('/orders')">{{ t('payment.result.viewOrders') }}</button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import { usePaymentStore } from '@/stores/payment'
import { paymentAPI } from '@/api/payment'
import type { PaymentOrder } from '@/types/payment'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const paymentStore = usePaymentStore()

const order = ref<PaymentOrder | null>(null)
const loading = ref(true)

interface ReturnInfo {
  outTradeNo: string
  money: string
  type: string
  tradeStatus: string
}
const returnInfo = ref<ReturnInfo | null>(null)

const SUCCESS_STATUSES = new Set(['COMPLETED', 'PAID', 'RECHARGING'])
const PENDING_STATUSES = new Set(['PENDING', 'CREATED', 'WAITING', 'PROCESSING'])
const STATUS_REFRESH_INTERVAL_MS = 2000
const STATUS_REFRESH_MAX_ATTEMPTS = 15

let statusRefreshTimer: ReturnType<typeof setTimeout> | null = null
const refreshAttempts = ref(0)

/** 充值金额 = pay_amount / (1 + fee_rate/100)，fee_rate=0 时等于 pay_amount */
const baseAmount = computed(() => {
  if (!order.value || order.value.fee_rate <= 0) return order.value?.pay_amount ?? 0
  return Math.round((order.value.pay_amount / (1 + order.value.fee_rate / 100)) * 100) / 100
})

/** 手续费 = pay_amount - baseAmount */
const feeAmount = computed(() => {
  if (!order.value || order.value.fee_rate <= 0) return 0
  return Math.round((order.value.pay_amount - baseAmount.value) * 100) / 100
})

const isSuccess = computed(() => {
  if (order.value) {
    return SUCCESS_STATUSES.has(order.value.status)
  }
  if (route.query.status === 'success') return true
  if (route.query.trade_status === 'TRADE_SUCCESS') return true
  return false
})

const isPending = computed(() => {
  if (order.value) {
    return PENDING_STATUSES.has(order.value.status)
  }
  return false
})

const statusTitle = computed(() => {
  if (isSuccess.value) return t('payment.result.success')
  if (isPending.value) return t('payment.result.processing')
  return t('payment.result.failed')
})

function readQueryString(key: string): string {
  const value = route.query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function parseOutTradeNo(outTradeNo: string): number {
  const match = outTradeNo.match(/_(\d+)$/)
  return match ? Number(match[1]) : 0
}

async function resolveOrderFromResumeToken(resumeToken: string): Promise<PaymentOrder | null> {
  try {
    const result = await paymentAPI.resolveOrderPublicByResumeToken(resumeToken)
    return result.data
  } catch (_err: unknown) {
    return null
  }
}

async function resolveOrderFromOutTradeNo(outTradeNo: string): Promise<PaymentOrder | null> {
  try {
    const result = await paymentAPI.verifyOrderPublic(outTradeNo)
    return result.data
  } catch (_err: unknown) {
    try {
      const result = await paymentAPI.verifyOrder(outTradeNo)
      return result.data
    } catch (_verifyErr: unknown) {
      return null
    }
  }
}

async function resolveOrderFromID(orderId: number): Promise<PaymentOrder | null> {
  try {
    return await paymentStore.pollOrderStatus(orderId)
  } catch (_err: unknown) {
    return null
  }
}

function clearStatusRefreshTimer() {
  if (statusRefreshTimer !== null) {
    clearTimeout(statusRefreshTimer)
    statusRefreshTimer = null
  }
}

function scheduleStatusRefresh(refreshOrder: (() => Promise<PaymentOrder | null>) | null) {
  clearStatusRefreshTimer()
  if (!refreshOrder || !isPending.value || refreshAttempts.value >= STATUS_REFRESH_MAX_ATTEMPTS) {
    return
  }
  statusRefreshTimer = setTimeout(async () => {
    refreshAttempts.value += 1
    const refreshedOrder = await refreshOrder()
    if (refreshedOrder) {
      order.value = refreshedOrder
    }
    scheduleStatusRefresh(refreshOrder)
  }, STATUS_REFRESH_INTERVAL_MS)
}

onMounted(async () => {
  let orderId = Number(readQueryString('order_id')) || 0
  const resumeToken = readQueryString('resume_token') || readQueryString('wechat_resume_token')
  const outTradeNo = readQueryString('out_trade_no')
  let refreshOrder: (() => Promise<PaymentOrder | null>) | null = null

  if (!orderId && outTradeNo) {
    orderId = parseOutTradeNo(outTradeNo)
    returnInfo.value = {
      outTradeNo,
      money: readQueryString('money'),
      type: readQueryString('type'),
      tradeStatus: readQueryString('trade_status'),
    }
  }

  if (resumeToken) {
    order.value = await resolveOrderFromResumeToken(resumeToken)
    refreshOrder = () => resolveOrderFromResumeToken(resumeToken)
  }
  if (!order.value && outTradeNo) {
    order.value = await resolveOrderFromOutTradeNo(outTradeNo)
    refreshOrder = () => resolveOrderFromOutTradeNo(outTradeNo)
  }
  if (!order.value && orderId) {
    order.value = await resolveOrderFromID(orderId)
    refreshOrder = () => resolveOrderFromID(orderId)
  }
  if (order.value) {
    scheduleStatusRefresh(refreshOrder)
  }
  loading.value = false
})

onBeforeUnmount(() => {
  clearStatusRefreshTimer()
})
</script>

<style scoped>
.payment-result-view {
  display: flex;
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--theme-page-bg) 92%, var(--theme-surface-soft)),
      var(--theme-page-bg)
    );
  padding: 1rem;
}

.payment-result-view__shell {
  width: 100%;
  max-width: 28rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.payment-result-view__loading {
  display: flex;
  justify-content: center;
  padding-block: 5rem;
}

.payment-result-view__spinner {
  height: 2rem;
  width: 2rem;
  animation: payment-result-view-spin 1s linear infinite;
  border: 4px solid rgb(var(--theme-accent-rgb));
  border-top-color: transparent;
  border-radius: 999px;
}

.payment-result-view__hero {
  text-align: center;
}

@keyframes payment-result-view-spin {
  to {
    transform: rotate(360deg);
  }
}

.payment-result-view__hero-icon {
  margin: 0 auto;
  display: flex;
  height: 5rem;
  width: 5rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
}

.payment-result-view__hero-icon--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
}

.payment-result-view__hero-icon--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 12%, var(--theme-surface));
}

.payment-result-view__hero-symbol {
  height: 2.5rem;
  width: 2.5rem;
}

.payment-result-view__hero-symbol--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.payment-result-view__hero-symbol--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.payment-result-view__title {
  margin-top: 1rem;
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--theme-page-text);
}

.payment-result-view__card {
  border: var(--theme-card-border-width) solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 10px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
  padding: 1.25rem;
}

.payment-result-view__details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  font-size: 0.875rem;
}

.payment-result-view__row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
}

.payment-result-view__label {
  color: var(--theme-page-muted);
}

.payment-result-view__value {
  font-weight: 600;
  color: var(--theme-page-text);
}

.payment-result-view__value--accent {
  color: var(--theme-accent);
}

.payment-result-view__actions {
  display: flex;
  gap: 0.75rem;
}
</style>
