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
          <div v-else
            class="payment-result-view__hero-icon payment-result-view__hero-icon--danger">
            <svg class="payment-result-view__hero-symbol payment-result-view__hero-symbol--danger" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 class="payment-result-view__title">
            {{ isSuccess ? t('payment.result.success') : t('payment.result.failed') }}
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
import { ref, computed, onMounted } from 'vue'
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
  // Always prioritize actual order status from backend
  if (order.value) {
    return SUCCESS_STATUSES.has(order.value.status)
  }
  // Fallback only when order not loaded
  if (route.query.status === 'success') return true
  if (route.query.trade_status === 'TRADE_SUCCESS') return true
  return false
})

/** Extract numeric order ID from out_trade_no like "sub2_46" → 46 */
function parseOutTradeNo(outTradeNo: string): number {
  const match = outTradeNo.match(/_(\d+)$/)
  return match ? Number(match[1]) : 0
}

onMounted(async () => {
  // Try order_id first (internal navigation from QRCode/Stripe pages)
  let orderId = Number(route.query.order_id) || 0
  const outTradeNo = String(route.query.out_trade_no || '')

  // Fallback: EasyPay return URL with out_trade_no
  if (!orderId && outTradeNo) {
    orderId = parseOutTradeNo(outTradeNo)
    // Store return info for display when order lookup fails
    returnInfo.value = {
      outTradeNo,
      money: String(route.query.money || ''),
      type: String(route.query.type || ''),
      tradeStatus: String(route.query.trade_status || ''),
    }
  }

  // Verify payment via public endpoint (works without login)
  if (outTradeNo) {
    try {
      const result = await paymentAPI.verifyOrderPublic(outTradeNo)
      order.value = result.data
    } catch (_err: unknown) {
      // Public verify failed, try authenticated endpoint if logged in
      try {
        const result = await paymentAPI.verifyOrder(outTradeNo)
        order.value = result.data
      } catch (_e: unknown) { /* fall through */ }
    }
  }

  // Normal order lookup by ID (if verify didn't load the order)
  if (!order.value && orderId) {
    try {
      order.value = await paymentStore.pollOrderStatus(orderId)
    } catch (_err: unknown) {
      // Order lookup failed, will show returnInfo fallback
    }
  }
  loading.value = false
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
