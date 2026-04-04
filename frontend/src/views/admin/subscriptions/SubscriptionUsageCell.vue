<template>
  <div class="min-w-[280px] space-y-2">
    <div v-if="subscription.group?.daily_limit_usd" class="usage-row">
      <div class="flex items-center gap-2">
        <span class="usage-label">{{ t('admin.subscriptions.daily') }}</span>
        <div class="h-1.5 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
          <div
            class="h-1.5 rounded-full transition-all"
            :class="getUsageProgressClass(subscription.daily_usage_usd, subscription.group?.daily_limit_usd)"
            :style="{ width: getUsageProgressWidth(subscription.daily_usage_usd, subscription.group?.daily_limit_usd) }"
          ></div>
        </div>
        <span class="usage-amount">
          ${{ subscription.daily_usage_usd?.toFixed(2) || '0.00' }}
          <span class="text-gray-400">/</span>
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
        <div class="h-1.5 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
          <div
            class="h-1.5 rounded-full transition-all"
            :class="getUsageProgressClass(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd)"
            :style="{ width: getUsageProgressWidth(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd) }"
          ></div>
        </div>
        <span class="usage-amount">
          ${{ subscription.weekly_usage_usd?.toFixed(2) || '0.00' }}
          <span class="text-gray-400">/</span>
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
        <div class="h-1.5 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
          <div
            class="h-1.5 rounded-full transition-all"
            :class="getUsageProgressClass(subscription.monthly_usage_usd, subscription.group?.monthly_limit_usd)"
            :style="{ width: getUsageProgressWidth(subscription.monthly_usage_usd, subscription.group?.monthly_limit_usd) }"
          ></div>
        </div>
        <span class="usage-amount">
          ${{ subscription.monthly_usage_usd?.toFixed(2) || '0.00' }}
          <span class="text-gray-400">/</span>
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
      class="flex items-center gap-2 rounded-lg bg-gradient-to-r from-emerald-50 to-teal-50 px-3 py-2 dark:from-emerald-900/20 dark:to-teal-900/20"
    >
      <span class="text-lg text-emerald-600 dark:text-emerald-400">∞</span>
      <span class="text-xs font-medium text-emerald-700 dark:text-emerald-300">
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
} from '../subscriptionForm'

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
.usage-row {
  @apply space-y-1;
}

.usage-label {
  @apply w-10 flex-shrink-0 text-xs font-medium text-gray-500 dark:text-gray-400;
}

.usage-amount {
  @apply whitespace-nowrap text-xs tabular-nums text-gray-600 dark:text-gray-300;
}

.reset-info {
  @apply flex items-center gap-1 pl-12 text-[10px] text-blue-600 dark:text-blue-400;
}
</style>
