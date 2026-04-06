<template>
  <div class="card subscription-usage-card overflow-hidden">
    <div class="subscription-usage-card__header">
      <div class="flex items-center gap-3">
        <div class="subscription-usage-card__icon-shell">
          <Icon name="creditCard" size="md" class="subscription-usage-card__icon" />
        </div>
        <div>
          <h3 class="subscription-usage-card__title">
            {{ subscription.group?.name || `Group #${subscription.group_id}` }}
          </h3>
          <p class="subscription-usage-card__description">
            {{ subscription.group?.description || '' }}
          </p>
        </div>
      </div>
      <span
        :class="[
          'badge',
          subscription.status === 'active'
            ? 'badge-success'
            : subscription.status === 'expired'
              ? 'badge-warning'
              : 'badge-danger'
        ]"
      >
        {{ t(`userSubscriptions.status.${subscription.status}`) }}
      </span>
    </div>

    <div class="subscription-usage-card__body space-y-4">
      <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
        <span class="subscription-usage-card__meta">{{ t('userSubscriptions.expires') }}</span>
        <span :class="resolveSubscriptionExpirationClass(subscription.expires_at, now)">
          {{ formatSubscriptionExpirationDate(subscription.expires_at, now, t) }}
        </span>
      </div>
      <div v-else class="flex items-center justify-between text-sm">
        <span class="subscription-usage-card__meta">{{ t('userSubscriptions.expires') }}</span>
        <span class="subscription-usage-card__value">{{ t('userSubscriptions.noExpiration') }}</span>
      </div>

      <div v-if="subscription.group?.daily_limit_usd" class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="subscription-usage-card__label">
            {{ t('userSubscriptions.daily') }}
          </span>
          <span class="subscription-usage-card__meta">
            ${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{
              subscription.group.daily_limit_usd.toFixed(2)
            }}
          </span>
        </div>
        <div class="subscription-usage-card__progress-track">
          <div
            :class="buildSubscriptionProgressBarClass(subscription.daily_usage_usd, subscription.group.daily_limit_usd)"
            :style="{ width: buildSubscriptionProgressWidth(subscription.daily_usage_usd, subscription.group.daily_limit_usd) }"
          ></div>
        </div>
        <p v-if="subscription.daily_window_start" class="subscription-usage-card__meta subscription-usage-card__meta--small">
          {{ t('userSubscriptions.resetIn', { time: formatSubscriptionResetTime(subscription.daily_window_start, 24, now, t) }) }}
        </p>
      </div>

      <div v-if="subscription.group?.weekly_limit_usd" class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="subscription-usage-card__label">
            {{ t('userSubscriptions.weekly') }}
          </span>
          <span class="subscription-usage-card__meta">
            ${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{
              subscription.group.weekly_limit_usd.toFixed(2)
            }}
          </span>
        </div>
        <div class="subscription-usage-card__progress-track">
          <div
            :class="buildSubscriptionProgressBarClass(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd)"
            :style="{ width: buildSubscriptionProgressWidth(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd) }"
          ></div>
        </div>
        <p v-if="subscription.weekly_window_start" class="subscription-usage-card__meta subscription-usage-card__meta--small">
          {{ t('userSubscriptions.resetIn', { time: formatSubscriptionResetTime(subscription.weekly_window_start, 168, now, t) }) }}
        </p>
      </div>

      <div v-if="subscription.group?.monthly_limit_usd" class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="subscription-usage-card__label">
            {{ t('userSubscriptions.monthly') }}
          </span>
          <span class="subscription-usage-card__meta">
            ${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{
              subscription.group.monthly_limit_usd.toFixed(2)
            }}
          </span>
        </div>
        <div class="subscription-usage-card__progress-track">
          <div
            :class="buildSubscriptionProgressBarClass(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd)"
            :style="{ width: buildSubscriptionProgressWidth(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd) }"
          ></div>
        </div>
        <p v-if="subscription.monthly_window_start" class="subscription-usage-card__meta subscription-usage-card__meta--small">
          {{ t('userSubscriptions.resetIn', { time: formatSubscriptionResetTime(subscription.monthly_window_start, 720, now, t) }) }}
        </p>
      </div>

      <div
        v-if="!hasSubscriptionLimits(subscription)"
        class="subscription-usage-card__unlimited"
      >
        <div class="flex items-center gap-3">
          <span class="subscription-usage-card__unlimited-icon">∞</span>
          <div>
            <p class="subscription-usage-card__unlimited-title">
              {{ t('userSubscriptions.unlimited') }}
            </p>
            <p class="subscription-usage-card__unlimited-description">
              {{ t('userSubscriptions.unlimitedDesc') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { UserSubscription } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import {
  buildSubscriptionProgressBarClass,
  buildSubscriptionProgressWidth,
  formatSubscriptionExpirationDate,
  formatSubscriptionResetTime,
  hasSubscriptionLimits,
  resolveSubscriptionExpirationClass
} from '../subscriptionsView'

defineProps<{
  now: Date
  subscription: UserSubscription
}>()

const { t } = useI18n()
</script>

<style scoped>
.subscription-usage-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  padding: var(--theme-settings-card-panel-padding);
}

.subscription-usage-card__body {
  padding: var(--theme-settings-card-panel-padding);
}

.subscription-usage-card__icon-shell {
  display: flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: calc(var(--theme-surface-radius) - 2px);
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 12%, var(--theme-surface));
}

.subscription-usage-card__icon {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.subscription-usage-card__title,
.subscription-usage-card__label,
.subscription-usage-card__value,
.subscription-usage-card__unlimited-title {
  color: var(--theme-page-text);
}

.subscription-usage-card__title {
  font-weight: 600;
}

.subscription-usage-card__description,
.subscription-usage-card__meta,
.subscription-usage-card__unlimited-description {
  color: var(--theme-page-muted);
}

.subscription-usage-card__description,
.subscription-usage-card__meta--small,
.subscription-usage-card__unlimited-description {
  font-size: 0.75rem;
}

.subscription-usage-card__label {
  font-size: 0.875rem;
  font-weight: 500;
}

.subscription-usage-card__progress-track {
  position: relative;
  height: 0.5rem;
  overflow: hidden;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-page-border) 76%, var(--theme-surface));
}

.subscription-usage-card__progress-bar {
  position: absolute;
  inset: 0 auto 0 0;
  border-radius: 9999px;
  transition: width 0.3s ease;
}

.subscription-usage-card__progress-bar--neutral {
  background: color-mix(in srgb, var(--theme-page-muted) 62%, var(--theme-surface));
}

.subscription-usage-card__progress-bar--danger {
  background: rgb(var(--theme-danger-rgb));
}

.subscription-usage-card__progress-bar--warning {
  background: rgb(var(--theme-warning-rgb));
}

.subscription-usage-card__progress-bar--success {
  background: rgb(var(--theme-success-rgb));
}

.subscription-usage-card__expiration {
  color: var(--theme-page-text);
}

.subscription-usage-card__expiration--default {
  color: var(--theme-page-text);
}

.subscription-usage-card__expiration--warning {
  color: rgb(var(--theme-warning-rgb));
}

.subscription-usage-card__expiration--urgent,
.subscription-usage-card__expiration--expired {
  color: rgb(var(--theme-danger-rgb));
}

.subscription-usage-card__expiration--expired {
  font-weight: 600;
}

.subscription-usage-card__unlimited {
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background:
    linear-gradient(
      90deg,
      color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface)) 0%,
      color-mix(in srgb, rgb(var(--theme-info-rgb)) 8%, var(--theme-surface)) 100%
    );
  padding: var(--theme-settings-card-body-padding) 0;
}

.subscription-usage-card__unlimited-icon {
  color: rgb(var(--theme-success-rgb));
  font-size: 2.25rem;
}

.subscription-usage-card__unlimited-title {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 78%, var(--theme-page-text));
  font-size: 0.875rem;
  font-weight: 600;
}

.subscription-usage-card__unlimited-description {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 68%, var(--theme-page-muted));
}
</style>
