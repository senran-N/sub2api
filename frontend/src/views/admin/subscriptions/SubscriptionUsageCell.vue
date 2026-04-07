<template>
  <div class="subscription-usage-cell space-y-2">
    <div v-if="subscription.group?.daily_limit_usd" class="usage-row">
      <div class="flex items-center gap-2">
        <span class="usage-label">{{ t('admin.subscriptions.daily') }}</span>
        <div class="subscription-usage-cell__track h-1.5 flex-1 rounded-full">
          <div
            class="theme-progress-fill h-1.5"
            :class="getUsageProgressClass(subscription.daily_usage_usd, subscription.group?.daily_limit_usd)"
            :style="{ width: getUsageProgressWidth(subscription.daily_usage_usd, subscription.group?.daily_limit_usd) }"
          ></div>
        </div>
        <span class="usage-amount">
          ${{ subscription.daily_usage_usd?.toFixed(2) || '0.00' }}
          <span class="subscription-usage-cell__separator">/</span>
          ${{ subscription.group?.daily_limit_usd?.toFixed(2) }}
        </span>
      </div>
      <div v-if="subscription.daily_window_start" class="reset-info">
        <svg
          class="h-3 w-3"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{{ formatResetTime(subscription.daily_window_start, 'daily') }}</span>
      </div>
    </div>

    <div v-if="subscription.group?.weekly_limit_usd" class="usage-row">
      <div class="flex items-center gap-2">
        <span class="usage-label">{{ t('admin.subscriptions.weekly') }}</span>
        <div class="subscription-usage-cell__track h-1.5 flex-1 rounded-full">
          <div
            class="theme-progress-fill h-1.5"
            :class="getUsageProgressClass(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd)"
            :style="{ width: getUsageProgressWidth(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd) }"
          ></div>
        </div>
        <span class="usage-amount">
          ${{ subscription.weekly_usage_usd?.toFixed(2) || '0.00' }}
          <span class="subscription-usage-cell__separator">/</span>
          ${{ subscription.group?.weekly_limit_usd?.toFixed(2) }}
        </span>
      </div>
      <div v-if="subscription.weekly_window_start" class="reset-info">
        <svg
          class="h-3 w-3"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{{ formatResetTime(subscription.weekly_window_start, 'weekly') }}</span>
      </div>
    </div>

    <div v-if="subscription.group?.monthly_limit_usd" class="usage-row">
      <div class="flex items-center gap-2">
        <span class="usage-label">{{ t('admin.subscriptions.monthly') }}</span>
        <div class="subscription-usage-cell__track h-1.5 flex-1 rounded-full">
          <div
            class="theme-progress-fill h-1.5"
            :class="getUsageProgressClass(subscription.monthly_usage_usd, subscription.group?.monthly_limit_usd)"
            :style="{ width: getUsageProgressWidth(subscription.monthly_usage_usd, subscription.group?.monthly_limit_usd) }"
          ></div>
        </div>
        <span class="usage-amount">
          ${{ subscription.monthly_usage_usd?.toFixed(2) || '0.00' }}
          <span class="subscription-usage-cell__separator">/</span>
          ${{ subscription.group?.monthly_limit_usd?.toFixed(2) }}
        </span>
      </div>
      <div v-if="subscription.monthly_window_start" class="reset-info">
        <svg
          class="h-3 w-3"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{{ formatResetTime(subscription.monthly_window_start, 'monthly') }}</span>
      </div>
    </div>

    <div
      v-if="!subscription.group?.daily_limit_usd && !subscription.group?.weekly_limit_usd && !subscription.group?.monthly_limit_usd"
      class="subscription-usage-cell__unlimited flex items-center gap-2"
    >
      <span class="subscription-usage-cell__unlimited-icon text-lg">∞</span>
      <span class="subscription-usage-cell__unlimited-text text-xs font-medium">
        {{ t('admin.subscriptions.unlimited') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { UserSubscription } from '@/types'
import {
  getResetWindowMessage,
  getUsageProgressClass,
  getUsageProgressWidth,
  type ResetWindowPeriod
} from './subscriptionForm'

defineProps<{
  subscription: UserSubscription
}>()

const { t } = useI18n()

const formatResetTime = (windowStart: string, period: ResetWindowPeriod): string => {
  const message = getResetWindowMessage(windowStart, period)
  return message.params ? t(message.key, message.params) : t(message.key)
}
</script>

<style scoped>
.subscription-usage-cell {
  min-width: var(--theme-subscription-usage-min-width);
}

.usage-row {
  @apply space-y-1;
}

.usage-label {
  @apply w-10 flex-shrink-0 text-xs font-medium;
  color: var(--theme-page-muted);
}

.usage-amount {
  @apply whitespace-nowrap text-xs tabular-nums;
  color: var(--theme-page-text);
}

.reset-info {
  @apply flex items-center gap-1 pl-12 text-[10px];
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.subscription-usage-cell__track {
  background: color-mix(in srgb, var(--theme-page-border) 78%, var(--theme-surface));
}

.subscription-usage-cell__separator {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, var(--theme-surface));
}

.subscription-usage-cell__unlimited {
  padding: calc(var(--theme-markdown-block-padding) * 0.5) calc(var(--theme-markdown-block-padding) * 0.75);
  border-radius: var(--theme-subscription-panel-radius);
  background: linear-gradient(
    90deg,
    color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface)),
    color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface))
  );
}

.subscription-usage-cell__unlimited-icon,
.subscription-usage-cell__unlimited-text {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}
</style>
