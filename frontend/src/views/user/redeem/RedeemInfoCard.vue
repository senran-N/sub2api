<template>
  <div class="card redeem-info-card">
    <div class="redeem-info-card__content">
      <div class="redeem-info-card__row">
        <div class="redeem-info-card__icon">
          <Icon name="infoCircle" size="md" class="redeem-info-card__icon-symbol" />
        </div>
        <div class="redeem-info-card__body">
          <h3 class="redeem-info-card__title">
            {{ aboutCodesLabel }}
          </h3>
          <ul class="redeem-info-card__list">
            <li>{{ codeRule1 }}</li>
            <li>{{ codeRule2 }}</li>
            <li>
              {{ codeRule3 }}
              <span v-if="contactInfo" class="redeem-info-card__contact">
                {{ contactInfo }}
              </span>
            </li>
            <li>{{ codeRule4 }}</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  aboutCodesLabel: string
  codeRule1: string
  codeRule2: string
  codeRule3: string
  codeRule4: string
  contactInfo: string
}>()
</script>

<style scoped>
.redeem-info-card {
  border-color: color-mix(in srgb, var(--theme-accent) 18%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
}

.redeem-info-card__content {
  padding: calc(var(--theme-table-mobile-card-padding) * 1.5);
}

.redeem-info-card__row {
  display: flex;
  align-items: flex-start;
  gap: var(--theme-table-mobile-card-padding);
}

.redeem-info-card__icon {
  display: flex;
  height: calc(var(--theme-header-avatar-size) + 8px);
  width: calc(var(--theme-header-avatar-size) + 8px);
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background: color-mix(in srgb, var(--theme-accent-soft) 88%, var(--theme-surface));
}

.redeem-info-card__body {
  flex: 1;
}

.redeem-info-card__icon-symbol {
  color: var(--theme-accent);
}

.redeem-info-card__title {
  color: color-mix(in srgb, var(--theme-accent) 72%, var(--theme-page-text));
  font-size: 0.875rem;
  font-weight: 600;
}

.redeem-info-card__list {
  margin-top: 0.5rem;
  list-style: disc inside;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.875rem;
  color: color-mix(in srgb, var(--theme-accent) 58%, var(--theme-page-text));
}

.redeem-info-card__contact {
  margin-left: 0.375rem;
  display: inline-flex;
  align-items: center;
  border-radius: var(--theme-button-radius);
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  background: color-mix(in srgb, var(--theme-accent-soft) 62%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 76%, var(--theme-page-text));
}
</style>
