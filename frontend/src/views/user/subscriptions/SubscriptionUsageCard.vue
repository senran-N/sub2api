<template>
  <div class="card overflow-hidden">
    <div class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700">
      <div class="flex items-center gap-3">
        <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-purple-100 dark:bg-purple-900/30">
          <Icon name="creditCard" size="md" class="text-purple-600 dark:text-purple-400" />
        </div>
        <div>
          <h3 class="font-semibold text-gray-900 dark:text-white">
            {{ subscription.group?.name || `Group #${subscription.group_id}` }}
          </h3>
          <p class="text-xs text-gray-500 dark:text-dark-400">
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

    <div class="space-y-4 p-4">
      <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
        <span class="text-gray-500 dark:text-dark-400">{{ t('userSubscriptions.expires') }}</span>
        <span :class="resolveSubscriptionExpirationClass(subscription.expires_at, now)">
          {{ formatSubscriptionExpirationDate(subscription.expires_at, now, t) }}
        </span>
      </div>
      <div v-else class="flex items-center justify-between text-sm">
        <span class="text-gray-500 dark:text-dark-400">{{ t('userSubscriptions.expires') }}</span>
        <span class="text-gray-700 dark:text-gray-300">{{ t('userSubscriptions.noExpiration') }}</span>
      </div>

      <div v-if="subscription.group?.daily_limit_usd" class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('userSubscriptions.daily') }}
          </span>
          <span class="text-sm text-gray-500 dark:text-dark-400">
            ${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{
              subscription.group.daily_limit_usd.toFixed(2)
            }}
          </span>
        </div>
        <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
          <div
            class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
            :class="buildSubscriptionProgressBarClass(subscription.daily_usage_usd, subscription.group.daily_limit_usd)"
            :style="{ width: buildSubscriptionProgressWidth(subscription.daily_usage_usd, subscription.group.daily_limit_usd) }"
          ></div>
        </div>
        <p v-if="subscription.daily_window_start" class="text-xs text-gray-500 dark:text-dark-400">
          {{ t('userSubscriptions.resetIn', { time: formatSubscriptionResetTime(subscription.daily_window_start, 24, now, t) }) }}
        </p>
      </div>

      <div v-if="subscription.group?.weekly_limit_usd" class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('userSubscriptions.weekly') }}
          </span>
          <span class="text-sm text-gray-500 dark:text-dark-400">
            ${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{
              subscription.group.weekly_limit_usd.toFixed(2)
            }}
          </span>
        </div>
        <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
          <div
            class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
            :class="buildSubscriptionProgressBarClass(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd)"
            :style="{ width: buildSubscriptionProgressWidth(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd) }"
          ></div>
        </div>
        <p v-if="subscription.weekly_window_start" class="text-xs text-gray-500 dark:text-dark-400">
          {{ t('userSubscriptions.resetIn', { time: formatSubscriptionResetTime(subscription.weekly_window_start, 168, now, t) }) }}
        </p>
      </div>

      <div v-if="subscription.group?.monthly_limit_usd" class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('userSubscriptions.monthly') }}
          </span>
          <span class="text-sm text-gray-500 dark:text-dark-400">
            ${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{
              subscription.group.monthly_limit_usd.toFixed(2)
            }}
          </span>
        </div>
        <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
          <div
            class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
            :class="buildSubscriptionProgressBarClass(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd)"
            :style="{ width: buildSubscriptionProgressWidth(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd) }"
          ></div>
        </div>
        <p v-if="subscription.monthly_window_start" class="text-xs text-gray-500 dark:text-dark-400">
          {{ t('userSubscriptions.resetIn', { time: formatSubscriptionResetTime(subscription.monthly_window_start, 720, now, t) }) }}
        </p>
      </div>

      <div
        v-if="!hasSubscriptionLimits(subscription)"
        class="flex items-center justify-center rounded-xl bg-gradient-to-r from-emerald-50 to-teal-50 py-6 dark:from-emerald-900/20 dark:to-teal-900/20"
      >
        <div class="flex items-center gap-3">
          <span class="text-4xl text-emerald-600 dark:text-emerald-400">∞</span>
          <div>
            <p class="text-sm font-medium text-emerald-700 dark:text-emerald-300">
              {{ t('userSubscriptions.unlimited') }}
            </p>
            <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">
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
