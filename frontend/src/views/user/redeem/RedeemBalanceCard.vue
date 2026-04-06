<template>
  <div class="card redeem-balance-card">
    <div class="redeem-balance-card__hero">
      <div class="redeem-balance-card__icon">
        <Icon name="creditCard" size="xl" class="redeem-balance-card__icon-symbol" />
      </div>
      <p class="redeem-balance-card__label">{{ currentBalanceLabel }}</p>
      <p class="redeem-balance-card__value">{{ balanceText }}</p>
      <p class="redeem-balance-card__meta">
        {{ concurrencyLabel }}: {{ concurrency }} {{ requestsLabel }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  balanceText: string
  concurrency: number
  concurrencyLabel: string
  currentBalanceLabel: string
  requestsLabel: string
}>()
</script>

<style scoped>
.redeem-balance-card {
  overflow: hidden;
}

.redeem-balance-card__hero {
  padding: calc(var(--theme-table-mobile-card-padding) * 1.5) calc(var(--theme-table-mobile-card-padding) * 1.5) calc(var(--theme-table-mobile-card-padding) * 2);
  text-align: center;
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent) 76%, rgb(var(--theme-info-rgb)))
  );
}

.redeem-balance-card__icon {
  margin-bottom: calc(var(--theme-table-mobile-card-padding) * 0.75);
  display: inline-flex;
  height: var(--theme-auth-logo-size);
  width: var(--theme-auth-logo-size);
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-auth-logo-radius);
  background: color-mix(in srgb, var(--theme-filled-text) 18%, transparent);
  backdrop-filter: blur(10px);
}

.redeem-balance-card__icon-symbol,
.redeem-balance-card__value {
  color: var(--theme-filled-text);
}

.redeem-balance-card__label,
.redeem-balance-card__meta {
  font-size: 0.875rem;
}

.redeem-balance-card__label {
  font-weight: 500;
}

.redeem-balance-card__value {
  margin-top: 0.5rem;
  font-size: clamp(2rem, 3.5vw, 2.25rem);
  font-weight: 700;
  line-height: 1.1;
}

.redeem-balance-card__meta {
  margin-top: 0.5rem;
}

.redeem-balance-card__label,
.redeem-balance-card__meta {
  color: color-mix(in srgb, var(--theme-filled-text) 84%, transparent);
}
</style>
