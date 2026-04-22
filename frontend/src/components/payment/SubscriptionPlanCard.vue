<template>
  <div :class="['payment-plan-card', toneClass]">
    <div class="payment-plan-card__accent" />

    <div class="payment-plan-card__body">
      <div class="mb-3 flex items-start justify-between gap-2">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <h3 class="truncate text-base font-bold">{{ plan.name }}</h3>
            <span class="theme-chip theme-chip--compact payment-plan-card__platform-badge">
              {{ platformLabelText }}
            </span>
          </div>
          <p v-if="plan.description" class="mt-0.5 line-clamp-2 text-xs leading-relaxed payment-muted">
            {{ plan.description }}
          </p>
        </div>
        <div class="shrink-0 text-right">
          <div class="flex items-baseline gap-1">
            <span class="payment-table-subtext text-xs">$</span>
            <span class="payment-plan-card__price text-2xl font-extrabold tracking-tight">{{ plan.price }}</span>
          </div>
          <span class="payment-table-subtext text-[11px]">/ {{ validitySuffix }}</span>
          <div v-if="plan.original_price" class="mt-0.5 flex items-center justify-end gap-1.5">
            <span class="payment-table-subtext text-xs line-through">${{ plan.original_price }}</span>
            <span class="payment-plan-card__discount">{{ discountText }}</span>
          </div>
        </div>
      </div>

      <div class="payment-plan-card__quota">
        <div class="flex items-center justify-between">
          <span class="payment-plan-card__quota-label">{{ t('payment.planCard.rate') }}</span>
          <span class="payment-plan-card__quota-value">{{ rateDisplay }}</span>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="flex items-center justify-between">
          <span class="payment-plan-card__quota-label">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="payment-plan-card__quota-value">${{ plan.daily_limit_usd }}</span>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="flex items-center justify-between">
          <span class="payment-plan-card__quota-label">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="payment-plan-card__quota-value">${{ plan.weekly_limit_usd }}</span>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="flex items-center justify-between">
          <span class="payment-plan-card__quota-label">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="payment-plan-card__quota-value">${{ plan.monthly_limit_usd }}</span>
        </div>
        <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null" class="flex items-center justify-between">
          <span class="payment-plan-card__quota-label">{{ t('payment.planCard.quota') }}</span>
          <span class="payment-plan-card__quota-value">{{ t('payment.planCard.unlimited') }}</span>
        </div>
        <div v-if="modelScopeLabels.length > 0" class="col-span-2 flex items-center justify-between">
          <span class="payment-plan-card__quota-label">{{ t('payment.planCard.models') }}</span>
          <div class="flex flex-wrap justify-end gap-1">
            <span v-for="scope in modelScopeLabels" :key="scope" class="payment-plan-card__scope-chip">
              {{ scope }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="plan.features.length > 0" class="mb-3 space-y-1">
        <div v-for="feature in plan.features" :key="feature" class="flex items-start gap-1.5">
          <svg class="payment-plan-card__feature-icon mt-0.5 h-3.5 w-3.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="payment-plan-card__feature-text">{{ feature }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <button type="button" class="payment-plan-card__button" @click="emit('select', plan)">
        {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import '@/components/payment/paymentTheme.css'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

const toneClass = computed(() => {
  switch (platform.value) {
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
})

const platformLabelText = computed(() => {
  switch (platform.value) {
    case 'anthropic':
      return 'Anthropic'
    case 'openai':
      return 'OpenAI'
    case 'antigravity':
      return 'Antigravity'
    case 'gemini':
      return 'Gemini'
    default:
      return platform.value || 'API'
  }
})

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const validitySuffix = computed(() => {
  const u = props.plan.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${props.plan.validity_days}${t('payment.days')}`
})
</script>
