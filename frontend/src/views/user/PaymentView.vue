<template>
  <AppLayout>
    <div class="payment-page payment-page--wide">
      <div v-if="loading" class="payment-loading">
        <div class="payment-spinner"></div>
      </div>
      <div v-else-if="paymentUnavailable" class="payment-panel payment-panel--center">
        <Icon name="creditCard" size="xl" class="mx-auto mb-3 payment-muted" />
        <p class="payment-title">{{ t('purchase.notEnabledTitle') }}</p>
        <p class="payment-muted mt-2 text-center text-sm">{{ paymentUnavailableMessage }}</p>
      </div>
      <template v-else>
        <!-- Tab Switcher (hide during payment and subscription confirm) -->
        <div v-if="tabs.length > 1 && paymentPhase === 'select' && !selectedPlan" class="payment-tab-list">
          <button v-for="tab in tabs" :key="tab.key"
            class="payment-tab"
            :class="{ 'payment-tab--active': activeTab === tab.key }"
            @click="activeTab = tab.key">{{ tab.label }}</button>
        </div>
        <!-- Payment in progress (shared by recharge and subscription) -->
        <template v-if="paymentPhase === 'paying'">
          <PaymentStatusPanel
            :order-id="paymentState.orderId"
            :qr-code="paymentState.qrCode"
            :expires-at="paymentState.expiresAt"
            :payment-type="paymentState.paymentType"
            :pay-url="paymentState.payUrl"
            :order-type="paymentState.orderType"
            @done="onPaymentDone"
            @success="onPaymentSuccess"
          />
        </template>
        <template v-else-if="paymentPhase === 'stripe'">
          <StripePaymentInline
            :order-id="paymentState.orderId"
            :amount="paymentState.amount"
            :client-secret="paymentState.clientSecret"
            :order-type="paymentState.orderType || undefined"
            :publishable-key="checkout.stripe_publishable_key"
            :pay-amount="paymentState.payAmount"
            @success="onPaymentSuccess"
            @done="onStripeDone"
            @back="resetPayment"
            @redirect="onStripeRedirect"
          />
        </template>
        <!-- Tab content (select phase) -->
        <template v-else>
          <!-- Top-up Tab -->
          <template v-if="activeTab === 'recharge'">
            <!-- Recharge Account Card -->
            <div class="payment-panel">
              <p class="payment-section-label">{{ t('payment.rechargeAccount') }}</p>
              <p class="payment-title mt-1">{{ user?.username || '' }}</p>
              <p class="payment-status-description mt-0.5">{{ t('payment.currentBalance') }}: {{ user?.balance?.toFixed(2) || '0.00' }}</p>
            </div>
            <div v-if="enabledMethods.length === 0" class="payment-panel payment-panel--center">
              <p class="payment-muted">{{ t('payment.notAvailable') }}</p>
            </div>
            <template v-else>
            <div class="payment-panel">
              <AmountInput
                v-model="amount"
                :amounts="[10, 20, 50, 100, 200, 500, 1000, 2000, 5000]"
                :min="globalMinAmount"
                :max="globalMaxAmount"
              />
              <p v-if="amountError" class="payment-feedback payment-feedback--warning mt-3">{{ amountError }}</p>
            </div>
            <div v-if="enabledMethods.length >= 1" class="payment-panel">
              <PaymentMethodSelector
                :methods="methodOptions"
                :selected="selectedMethod"
                @select="selectedMethod = $event"
              />
            </div>
            <div v-if="validAmount > 0" class="payment-panel">
              <div class="payment-detail-list">
                <div class="payment-detail-row">
                  <span class="payment-detail-label">{{ t('payment.paymentAmount') }}</span>
                  <span class="payment-detail-value">¥{{ validAmount.toFixed(2) }}</span>
                </div>
                <div v-if="feeRate > 0" class="payment-detail-row">
                  <span class="payment-detail-label">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                  <span class="payment-detail-value">¥{{ feeAmount.toFixed(2) }}</span>
                </div>
                <div v-if="feeRate > 0" class="payment-detail-row payment-detail-row--divided">
                  <span class="payment-detail-value">{{ t('payment.actualPay') }}</span>
                  <span class="payment-detail-value payment-detail-value--accent">¥{{ totalAmount.toFixed(2) }}</span>
                </div>
                <div v-if="balanceRechargeMultiplier !== 1" class="payment-detail-row" :class="{ 'payment-detail-row--divided': feeRate <= 0 }">
                  <span class="payment-detail-label">{{ t('payment.creditedBalance') }}</span>
                  <span class="payment-detail-value">${{ creditedAmount.toFixed(2) }}</span>
                </div>
                <p v-if="balanceRechargeMultiplier !== 1" class="payment-detail-note">
                  {{ t('payment.rechargeRatePreview', { usd: balanceRechargeMultiplier.toFixed(2) }) }}
                </p>
              </div>
            </div>
            <button :class="['btn w-full py-3 text-base font-medium payment-submit-button', paymentButtonClass]" :disabled="!canSubmit || submitting" @click="handleSubmitRecharge">
              <span v-if="submitting" class="flex items-center justify-center gap-2">
                <span class="payment-spinner payment-spinner--sm"></span>
                {{ t('common.processing') }}
              </span>
              <span v-else>{{ t('payment.createOrder') }} ¥{{ totalAmount.toFixed(2) }}</span>
            </button>
            <div v-if="errorMessage" class="payment-feedback payment-feedback--danger">
              <p>{{ errorMessage }}</p>
            </div>
            </template>
          </template>
          <!-- Subscribe Tab -->
          <template v-else-if="activeTab === 'subscription'">
            <!-- Subscription confirm (inline, replaces plan list) -->
            <template v-if="selectedPlan">
              <div :class="['payment-panel', planToneClassName]">
                <!-- Header: platform badge + plan name -->
                <div class="mb-3 flex flex-wrap items-center gap-2">
                  <span :class="['theme-chip theme-chip--regular', planBadgeClass]">
                    {{ platformLabel(selectedPlan.group_platform || '') }}
                  </span>
                  <h3 class="payment-title">{{ selectedPlan.name }}</h3>
                </div>
                <!-- Price -->
                <div class="flex items-baseline gap-2">
                  <span v-if="selectedPlan.original_price" class="payment-muted text-sm line-through">
                    ¥{{ selectedPlan.original_price }}
                  </span>
                  <span class="payment-plan-card__price text-3xl font-bold">¥{{ selectedPlan.price }}</span>
                  <span class="payment-muted text-sm">/ {{ planValiditySuffix }}</span>
                </div>
                <!-- Description -->
                <p v-if="selectedPlan.description" class="payment-muted mt-2 text-sm leading-relaxed">
                  {{ selectedPlan.description }}
                </p>
                <!-- Rate + Limits grid -->
                <div class="payment-plan-card__quota mt-3">
                  <div>
                    <span class="payment-plan-card__quota-label">{{ t('payment.planCard.rate') }}</span>
                    <div class="flex items-baseline">
                      <span class="payment-plan-card__price text-lg font-bold">×{{ selectedPlan.rate_multiplier ?? 1 }}</span>
                    </div>
                  </div>
                  <div v-if="selectedPlan.daily_limit_usd != null">
                    <span class="payment-plan-card__quota-label">{{ t('payment.planCard.dailyLimit') }}</span>
                    <div class="payment-plan-card__quota-value text-lg font-semibold">${{ selectedPlan.daily_limit_usd }}</div>
                  </div>
                  <div v-if="selectedPlan.weekly_limit_usd != null">
                    <span class="payment-plan-card__quota-label">{{ t('payment.planCard.weeklyLimit') }}</span>
                    <div class="payment-plan-card__quota-value text-lg font-semibold">${{ selectedPlan.weekly_limit_usd }}</div>
                  </div>
                  <div v-if="selectedPlan.monthly_limit_usd != null">
                    <span class="payment-plan-card__quota-label">{{ t('payment.planCard.monthlyLimit') }}</span>
                    <div class="payment-plan-card__quota-value text-lg font-semibold">${{ selectedPlan.monthly_limit_usd }}</div>
                  </div>
                  <div v-if="selectedPlan.daily_limit_usd == null && selectedPlan.weekly_limit_usd == null && selectedPlan.monthly_limit_usd == null">
                    <span class="payment-plan-card__quota-label">{{ t('payment.planCard.quota') }}</span>
                    <div class="payment-plan-card__quota-value text-lg font-semibold">{{ t('payment.planCard.unlimited') }}</div>
                  </div>
                </div>
              </div>
              <div v-if="enabledMethods.length >= 1" class="payment-panel">
                <PaymentMethodSelector
                  :methods="subMethodOptions"
                  :selected="selectedMethod"
                  @select="selectedMethod = $event"
                />
              </div>
              <div v-if="feeRate > 0 && selectedPlan.price > 0" class="payment-panel">
                <div class="payment-detail-list">
                  <div class="payment-detail-row">
                    <span class="payment-detail-label">{{ t('payment.amountLabel') }}</span>
                    <span class="payment-detail-value">¥{{ selectedPlan.price.toFixed(2) }}</span>
                  </div>
                  <div class="payment-detail-row">
                    <span class="payment-detail-label">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                    <span class="payment-detail-value">¥{{ subFeeAmount.toFixed(2) }}</span>
                  </div>
                  <div class="payment-detail-row payment-detail-row--divided">
                    <span class="payment-detail-value">{{ t('payment.actualPay') }}</span>
                    <span class="payment-detail-value payment-detail-value--accent">¥{{ subTotalAmount.toFixed(2) }}</span>
                  </div>
                </div>
              </div>
              <button :class="['btn w-full py-3 text-base font-medium payment-submit-button', paymentButtonClass]" :disabled="!canSubmitSubscription || submitting" @click="confirmSubscribe">
                <span v-if="submitting" class="flex items-center justify-center gap-2">
                  <span class="payment-spinner payment-spinner--sm"></span>
                  {{ t('common.processing') }}
                </span>
                <span v-else>{{ t('payment.createOrder') }} ¥{{ (feeRate > 0 ? subTotalAmount : selectedPlan.price).toFixed(2) }}</span>
              </button>
              <button class="btn btn-secondary w-full" @click="selectedPlan = null">{{ t('common.cancel') }}</button>
              <div v-if="errorMessage" class="payment-feedback payment-feedback--danger">
                <p>{{ errorMessage }}</p>
              </div>
            </template>
            <!-- Plan list -->
            <template v-else>
              <div v-if="checkout.plans.length === 0" class="payment-panel payment-panel--center">
                <Icon name="gift" size="xl" class="mx-auto mb-3 payment-muted" />
                <p class="payment-muted">{{ t('payment.noPlans') }}</p>
              </div>
              <div v-else :class="planGridClass">
                <SubscriptionPlanCard v-for="plan in checkout.plans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlan" />
              </div>
              <!-- Active subscriptions (compact, below plan list) -->
              <div v-if="activeSubscriptions.length > 0">
                <p class="payment-section-label mb-2">{{ t('payment.activeSubscription') }}</p>
                <div class="space-y-2">
                  <div v-for="sub in activeSubscriptions" :key="sub.id" :class="['payment-subscription-strip', planToneClass(sub.group?.platform || '')]">
                    <div class="payment-subscription-strip__accent" />
                    <div class="payment-subscription-strip__body">
                      <div class="flex items-center gap-1.5">
                        <span class="payment-subscription-strip__title">{{ sub.group?.name || `Group #${sub.group_id}` }}</span>
                        <span :class="['theme-chip theme-chip--compact', platformChipClass(sub.group?.platform || '')]">{{ platformLabel(sub.group?.platform || '') }}</span>
                      </div>
                      <div class="payment-subscription-strip__meta">
                        <span>{{ t('payment.planCard.rate') }}: ×{{ sub.group?.rate_multiplier ?? 1 }}</span>
                        <span v-if="sub.group?.daily_limit_usd == null && sub.group?.weekly_limit_usd == null && sub.group?.monthly_limit_usd == null">{{ t('payment.planCard.quota') }}: {{ t('payment.planCard.unlimited') }}</span>
                        <span v-if="sub.expires_at">{{ t('userSubscriptions.daysRemaining', { days: getDaysRemaining(sub.expires_at) }) }}</span>
                        <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
                      </div>
                    </div>
                    <span class="badge badge-success shrink-0 text-[10px]">{{ t('userSubscriptions.status.active') }}</span>
                  </div>
                </div>
              </div>
            </template>
          </template>
        </template>
        <div v-if="(checkout.help_text || checkout.help_image_url) && paymentPhase === 'select' && !selectedPlan" class="payment-panel payment-panel--tight">
          <div class="flex flex-col items-center gap-3">
            <img v-if="checkout.help_image_url" :src="checkout.help_image_url" alt=""
              class="payment-help-image"
              @click="previewImage = checkout.help_image_url" />
            <p v-if="checkout.help_text" class="payment-muted text-center text-sm">{{ checkout.help_text }}</p>
          </div>
        </div>
      </template>
    </div>
    <!-- Renewal Plan Selection Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenewalModal" class="payment-modal-shell" @click.self="closeRenewalModal">
          <div class="payment-panel payment-modal-card">
            <!-- Close button -->
            <button class="btn btn-secondary btn-icon absolute right-4 top-4" @click="closeRenewalModal">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
            <h3 class="mb-4 text-lg font-semibold">{{ t('payment.selectPlan') }}</h3>
            <div class="space-y-4">
              <SubscriptionPlanCard v-for="plan in renewalPlans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlanFromModal" />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- Image Preview Overlay -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewImage" class="payment-preview-overlay" @click="previewImage = ''">
          <img :src="previewImage" alt="" class="payment-preview-image" />
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import type { SubscriptionPlan, CheckoutInfoResponse, OrderType } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import AmountInput from '@/components/payment/AmountInput.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import { METHOD_ORDER, getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import StripePaymentInline from '@/components/payment/StripePaymentInline.vue'
import Icon from '@/components/icons/Icon.vue'
import type { PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import {
  buildCreateOrderPayload,
  clearPaymentSnapshot,
  createPaymentSnapshot,
  interpretWeChatJSAPIResult,
  invokeWeChatJSAPI,
  persistPaymentSnapshot,
  readPaymentSnapshot,
  resolvePaymentLaunch,
  type PaymentResumeInput,
} from '@/components/payment/paymentFlow'
import { hasResponseStatus } from '@/utils/requestError'
import {
  resolvePaymentWechatResumeIntent,
  stripPaymentWechatResumeQuery,
} from '@/views/user/paymentWechatResume'
import '@/components/payment/paymentTheme.css'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)

function getDaysRemaining(expiresAt: string): number {
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const paymentUnavailable = ref(false)
const paymentUnavailableMessage = ref('')
const activeTab = ref<'recharge' | 'subscription'>('recharge')
const amount = ref<number | null>(null)
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const previewImage = ref('')
const wechatResumeHandled = ref(false)

// Payment phase: 'select' → 'paying' (QR/redirect) or 'stripe' (inline Stripe)
const paymentPhase = ref<'select' | 'paying' | 'stripe'>('select')
const paymentState = ref<{
  orderId: number
  amount: number
  qrCode: string
  expiresAt: string
  paymentType: string
  payUrl: string
  clientSecret: string
  payAmount: number
  orderType: OrderType | ''
}>({ orderId: 0, amount: 0, qrCode: '', expiresAt: '', paymentType: '', payUrl: '', clientSecret: '', payAmount: 0, orderType: '' })

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = { orderId: 0, amount: 0, qrCode: '', expiresAt: '', paymentType: '', payUrl: '', clientSecret: '', payAmount: 0, orderType: '' }
  clearPaymentSnapshot()
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSuccess() {
  clearPaymentSnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onStripeDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onStripeRedirect(orderId: number, payUrl: string) {
  paymentState.value = { ...paymentState.value, orderId, payUrl, qrCode: '' }
  paymentPhase.value = 'paying'
}

// All checkout data from single API call
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], balance_disabled: false, balance_recharge_multiplier: 1, recharge_fee_rate: 0, help_text: '', help_image_url: '', stripe_publishable_key: '',
})

const tabs = computed(() => {
  const result: { key: 'recharge' | 'subscription'; label: string }[] = []
  if (!checkout.value.balance_disabled) result.push({ key: 'recharge', label: t('payment.tabTopUp') })
  result.push({ key: 'subscription', label: t('payment.tabSubscribe') })
  return result
})

const enabledMethods = computed(() => Object.keys(checkout.value.methods))
const validAmount = computed(() => amount.value ?? 0)
const balanceRechargeMultiplier = computed(() => {
  const multiplier = checkout.value.balance_recharge_multiplier
  return multiplier > 0 ? multiplier : 1
})
const creditedAmount = computed(() => Math.round((validAmount.value * balanceRechargeMultiplier.value) * 100) / 100)

// Adaptive grid: center single card, 2-col for 2 plans, 3-col for 3+
const planGridClass = computed(() => {
  const n = checkout.value.plans.length
  if (n <= 2) return 'grid grid-cols-1 gap-5 sm:grid-cols-2'
  return 'grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3'
})

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  const ml = checkout.value.methods[methodType]
  if (!ml) return false
  if (ml.single_min > 0 && amt < ml.single_min) return false
  if (ml.single_max > 0 && amt > ml.single_max) return false
  return true
}

// Global range for AmountInput (union of all methods, precomputed by backend)
const globalMinAmount = computed(() => checkout.value.global_min)
const globalMaxAmount = computed(() => checkout.value.global_max)

// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() => checkout.value.methods[selectedMethod.value])

const methodOptions = computed<PaymentMethodOption[]>(() =>
  enabledMethods.value.map((type) => {
    const ml = checkout.value.methods[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)
const feeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const totalAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + feeAmount.value) * 100) / 100
    : validAmount.value
)

const amountError = computed(() => {
  if (validAmount.value <= 0) return ''
  // No method can handle this amount
  if (!enabledMethods.value.some((m) => amountFitsMethod(validAmount.value, m))) {
    return t('payment.amountNoMethod')
  }
  // Selected method can't handle this amount (but others can)
  const ml = selectedLimit.value
  if (ml) {
    if (ml.single_min > 0 && validAmount.value < ml.single_min) return t('payment.amountTooLow', { min: ml.single_min })
    if (ml.single_max > 0 && validAmount.value > ml.single_max) return t('payment.amountTooHigh', { max: ml.single_max })
  }
  return ''
})

const canSubmit = computed(() =>
  validAmount.value > 0
    && amountFitsMethod(validAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Subscription-specific: method options based on plan price
const subMethodOptions = computed<PaymentMethodOption[]>(() => {
  const planPrice = selectedPlan.value?.price ?? 0
  return enabledMethods.value.map((type) => {
    const ml = checkout.value.methods[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(planPrice, type),
    }
  })
})

const subFeeAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return 0
  return Math.ceil(((price * feeRate.value) / 100) * 100) / 100
})

const subTotalAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return price
  return Math.round((price + subFeeAmount.value) * 100) / 100
})

const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && amountFitsMethod(selectedPlan.value.price, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedMethod.value] as const, ([amt, method]) => {
  if (amt <= 0 || amountFitsMethod(amt, method)) return
  const available = enabledMethods.value.find((m) => amountFitsMethod(amt, m))
  if (available) selectedMethod.value = available
})

// Payment button class: follows selected payment method color
const paymentButtonClass = computed(() => {
  const m = selectedMethod.value
  if (m.includes('alipay')) return 'payment-submit-button--alipay'
  if (m.includes('wxpay')) return 'payment-submit-button--wxpay'
  if (m === 'stripe') return 'payment-submit-button--stripe'
  return ''
})

function planToneClass(platform: string): string {
  switch (platform) {
    case 'anthropic':
      return 'payment-plan-card--anthropic'
    case 'openai':
      return 'payment-plan-card--openai'
    case 'antigravity':
      return 'payment-plan-card--antigravity'
    case 'gemini':
      return 'payment-plan-card--gemini'
    default:
      return ''
  }
}

function platformChipClass(platform: string): string {
  switch (platform) {
    case 'anthropic':
      return 'theme-chip--brand-orange'
    case 'openai':
      return 'theme-chip--success'
    case 'antigravity':
      return 'theme-chip--brand-purple'
    case 'gemini':
      return 'theme-chip--info'
    default:
      return 'theme-chip--accent'
  }
}

function platformLabel(platform: string): string {
  switch (platform) {
    case 'anthropic':
      return 'Anthropic'
    case 'openai':
      return 'OpenAI'
    case 'antigravity':
      return 'Antigravity'
    case 'gemini':
      return 'Gemini'
    default:
      return platform || 'API'
  }
}

const planToneClassName = computed(() => planToneClass(selectedPlan.value?.group_platform || ''))
const planBadgeClass = computed(() => platformChipClass(selectedPlan.value?.group_platform || ''))

// Renewal modal state
const showRenewalModal = ref(false)
const renewGroupId = ref<number | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

const planValiditySuffix = computed(() => {
  if (!selectedPlan.value) return ''
  const u = selectedPlan.value.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${selectedPlan.value.validity_days}${t('payment.days')}`
})

function selectPlan(plan: SubscriptionPlan) {
  selectedPlan.value = plan
  errorMessage.value = ''
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  selectedPlan.value = plan
  errorMessage.value = ''
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

function openWindow(url: string) {
  const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
  if (!win || win.closed) {
    window.location.href = url
  }
}

async function stripConsumedWechatResumeQuery() {
  const strippedQuery = stripPaymentWechatResumeQuery(route.query)
  const currentKeys = Object.keys(route.query)
  const strippedKeys = Object.keys(strippedQuery)
  if (currentKeys.length === strippedKeys.length) {
    return
  }
  await router.replace({
    path: route.path,
    query: strippedQuery,
  })
}

function applyWechatResumeSelection(intent: {
  paymentType?: string
  amount?: number
  orderType: OrderType
  planId?: number
}) {
  if (intent.paymentType && enabledMethods.value.includes(intent.paymentType)) {
    selectedMethod.value = intent.paymentType
  }

  if (intent.orderType === 'subscription') {
    activeTab.value = 'subscription'
    if (intent.planId) {
      const matchedPlan = checkout.value.plans.find((plan) => plan.id === intent.planId)
      if (matchedPlan) {
        selectedPlan.value = matchedPlan
      }
    }
    return
  }

  activeTab.value = 'recharge'
  if (intent.amount && intent.amount > 0) {
    amount.value = intent.amount
  }
}

async function maybeResumeWechatPayment() {
  if (wechatResumeHandled.value) {
    return
  }

  const intent = resolvePaymentWechatResumeIntent(route.query, readPaymentSnapshot())
  if (!intent?.shouldResume) {
    return
  }

  wechatResumeHandled.value = true
  applyWechatResumeSelection(intent)
  await stripConsumedWechatResumeQuery()

  if (intent.orderType === 'subscription') {
    if (!intent.planId) {
      return
    }
    const matchedPlan = checkout.value.plans.find((plan) => plan.id === intent.planId)
    if (!matchedPlan) {
      return
    }
    selectedPlan.value = matchedPlan
    await createOrder(matchedPlan.price, 'subscription', matchedPlan.id, intent.resume)
    return
  }

  if (intent.amount && intent.amount > 0) {
    await createOrder(intent.amount, 'balance', undefined, intent.resume)
  }
}

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  await createOrder(validAmount.value, 'balance')
}

async function confirmSubscribe() {
  if (!selectedPlan.value || submitting.value) return
  await createOrder(selectedPlan.value.price, 'subscription', selectedPlan.value.id)
}

async function createOrder(
  orderAmount: number,
  orderType: OrderType,
  planId?: number,
  resume?: PaymentResumeInput | null,
) {
  submitting.value = true
  errorMessage.value = ''
  try {
    const isMobile = isMobileDevice()
    persistPaymentSnapshot(createPaymentSnapshot({
      amount: orderAmount,
      orderType,
      paymentType: selectedMethod.value,
      planId,
    }))

    const result = await paymentStore.createOrder(buildCreateOrderPayload({
      amount: orderAmount,
      orderType,
      paymentType: selectedMethod.value,
      planId,
      isMobile,
      resume,
    }))

    const launch = resolvePaymentLaunch(result, {
      paymentType: selectedMethod.value,
      orderType,
      isMobile,
    })

    if (launch.kind === 'oauth_redirect') {
      if (!launch.redirectUrl) {
        throw new Error('missing_oauth_redirect')
      }
      window.location.href = launch.redirectUrl
      return
    }

    if (launch.kind === 'stripe') {
      paymentState.value = launch.paymentState
      paymentPhase.value = 'stripe'
      return
    }

    if (launch.kind === 'jsapi') {
      paymentState.value = {
        ...launch.paymentState,
        qrCode: '',
      }
      paymentPhase.value = 'paying'

      try {
        const jsapiResult = await invokeWeChatJSAPI(launch.jsapiParams || {})
        const jsapiOutcome = interpretWeChatJSAPIResult(jsapiResult)

        if ((jsapiOutcome === 'cancel' || jsapiOutcome === 'fail') && launch.paymentState.qrCode) {
          paymentState.value = launch.paymentState
        } else if (jsapiOutcome === 'fail' && launch.paymentState.payUrl) {
          paymentState.value = launch.paymentState
          window.location.href = launch.paymentState.payUrl
          return
        }
      } catch (_err: unknown) {
        if (launch.paymentState.qrCode) {
          paymentState.value = launch.paymentState
        } else if (launch.paymentState.payUrl) {
          paymentState.value = launch.paymentState
          window.location.href = launch.paymentState.payUrl
          return
        } else {
          errorMessage.value = t('payment.result.failed')
          appStore.showError(errorMessage.value)
        }
      }
      return
    }

    if (launch.kind === 'mobile_redirect') {
      if (!launch.redirectUrl) {
        throw new Error('missing_mobile_redirect')
      }
      paymentState.value = launch.paymentState
      paymentPhase.value = 'paying'
      window.location.href = launch.redirectUrl
      return
    }

    if (launch.kind === 'qr') {
      paymentState.value = launch.paymentState
      paymentPhase.value = 'paying'
      return
    }

    if (launch.kind === 'popup') {
      if (!launch.redirectUrl) {
        throw new Error('missing_popup_redirect')
      }
      openWindow(launch.redirectUrl)
      paymentState.value = launch.paymentState
      paymentPhase.value = 'paying'
      return
    }

    errorMessage.value = t('payment.result.failed')
    appStore.showError(errorMessage.value)
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
    } else {
      errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.result.failed'))
    }
    appStore.showError(errorMessage.value)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  try {
    const res = await paymentAPI.getCheckoutInfo()
    checkout.value = res.data
    if (enabledMethods.value.length) {
      const order: readonly string[] = METHOD_ORDER
      const sorted = [...enabledMethods.value].sort((a, b) => {
        const ai = order.indexOf(a)
        const bi = order.indexOf(b)
        return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
      })
      selectedMethod.value = sorted[0]
    }
    if (checkout.value.balance_disabled) {
      activeTab.value = 'subscription'
    }
    // Handle renewal navigation: ?tab=subscription&group=123
    if (route.query.tab === 'subscription') {
      activeTab.value = 'subscription'
      if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          selectedPlan.value = groupPlans[0]
        } else if (groupPlans.length > 1) {
          renewGroupId.value = groupId
          showRenewalModal.value = true
        }
      }
    }
    await maybeResumeWechatPayment()
  } catch (err: unknown) {
    if (hasResponseStatus(err, 404)) {
      paymentUnavailable.value = true
      paymentUnavailableMessage.value = t('purchase.notEnabledDesc')
    } else {
      appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
    }
  }
  finally { loading.value = false }
  // Fetch active subscriptions (uses cache, non-blocking)
  subscriptionStore.fetchActiveSubscriptions().catch(() => {})
})
</script>
