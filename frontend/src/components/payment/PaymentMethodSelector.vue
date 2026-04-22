<template>
  <div class="payment-method-selector">
    <label class="payment-method-selector__label">
      {{ t('payment.paymentMethod') }}
    </label>
    <div class="payment-method-selector__grid">
      <button
        v-for="method in sortedMethods"
        :key="method.type"
        type="button"
        :disabled="!method.available"
        :class="[
          'payment-method-selector__option',
          !method.available
            ? 'payment-method-selector__option--disabled'
            : selected === method.type
              ? methodSelectedClass(method.type)
              : 'payment-method-selector__option--idle',
        ]"
        @click="method.available && emit('select', method.type)"
      >
        <span class="payment-method-selector__content">
          <img :src="methodIcon(method.type)" :alt="t(`payment.methods.${method.type}`)" class="h-7 w-7" />
          <span class="payment-method-selector__copy">
            <span class="payment-method-selector__title">{{ t(`payment.methods.${method.type}`) }}</span>
            <span
              v-if="method.fee_rate > 0"
              class="payment-method-selector__fee"
            >
              {{ t('payment.fee') }} {{ method.fee_rate }}%
            </span>
          </span>
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { METHOD_ORDER } from './providerConfig'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'

export interface PaymentMethodOption {
  type: string
  fee_rate: number
  available: boolean
}

const props = defineProps<{
  methods: PaymentMethodOption[]
  selected: string
}>()

const emit = defineEmits<{
  select: [type: string]
}>()

const { t } = useI18n()

const METHOD_ICONS: Record<string, string> = {
  alipay: alipayIcon,
  wxpay: wxpayIcon,
  stripe: stripeIcon,
}

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

function methodIcon(type: string): string {
  if (type.includes('alipay')) return METHOD_ICONS.alipay
  if (type.includes('wxpay')) return METHOD_ICONS.wxpay
  return METHOD_ICONS[type] || alipayIcon
}

function methodSelectedClass(type: string): string {
  if (type.includes('alipay')) return 'payment-method-selector__option--alipay'
  if (type.includes('wxpay')) return 'payment-method-selector__option--wxpay'
  if (type === 'stripe') return 'payment-method-selector__option--stripe'
  return 'payment-method-selector__option--default'
}
</script>

<style scoped>
.payment-method-selector {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.payment-method-selector__label {
  display: block;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--theme-page-text);
}

.payment-method-selector__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.payment-method-selector__option {
  display: flex;
  min-height: 60px;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--theme-input-border) 88%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 6px);
  background: color-mix(in srgb, var(--theme-surface) 92%, transparent);
  color: var(--theme-page-text);
  padding: 0.75rem;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease,
    background 0.2s ease;
}

.payment-method-selector__option:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--theme-accent) 24%, var(--theme-card-border));
  transform: translateY(-1px);
}

.payment-method-selector__option--idle {
  box-shadow: var(--theme-card-shadow);
}

.payment-method-selector__option--disabled {
  cursor: not-allowed;
  background: var(--theme-disabled-surface);
  border-color: var(--theme-disabled-border);
  color: var(--theme-disabled-text);
  opacity: 0.7;
  box-shadow: none;
}

.payment-method-selector__option--alipay,
.payment-method-selector__option--wxpay,
.payment-method-selector__option--stripe,
.payment-method-selector__option--default {
  box-shadow: var(--theme-card-shadow);
}

.payment-method-selector__option--alipay {
  border-color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 32%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.payment-method-selector__option--wxpay {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 32%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.payment-method-selector__option--stripe {
  border-color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 32%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 10%, var(--theme-surface));
}

.payment-method-selector__option--default {
  border-color: color-mix(in srgb, var(--theme-accent) 28%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
}

.payment-method-selector__content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.payment-method-selector__copy {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.125rem;
  line-height: 1;
}

.payment-method-selector__title {
  font-size: 1rem;
  font-weight: 700;
}

.payment-method-selector__fee {
  font-size: 0.625rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--theme-page-muted);
}

@media (min-width: 640px) {
  .payment-method-selector__grid {
    display: flex;
  }

  .payment-method-selector__option {
    flex: 1;
  }
}
</style>
