<template>
  <div v-if="hasActiveSubscriptions" class="subscription-progress-mini relative" ref="containerRef">
    <!-- Mini Progress Display -->
    <button
      @click="toggleTooltip"
      class="subscription-progress-mini__trigger subscription-progress-mini__trigger-spacing flex cursor-pointer items-center gap-2 transition-colors"
      :title="t('subscriptionProgress.viewDetails')"
    >
      <Icon name="creditCard" size="sm" class="subscription-progress-mini__icon" />
      <div class="flex items-center gap-1.5">
        <!-- Combined progress indicator -->
        <div class="flex items-center gap-0.5">
          <div
            v-for="(sub, index) in displaySubscriptions.slice(0, 3)"
            :key="index"
            class="subscription-progress-mini__state-dot"
            :class="getProgressDotClass(sub)"
          ></div>
        </div>
        <span class="subscription-progress-mini__count text-xs font-medium">
          {{ activeSubscriptions.length }}
        </span>
      </div>
    </button>

    <!-- Hover/Click Tooltip -->
    <transition name="dropdown">
      <div
        v-if="tooltipOpen"
        class="subscription-progress-mini__panel absolute right-0 z-50 overflow-hidden"
      >
        <div class="subscription-progress-mini__header subscription-progress-mini__header-spacing">
          <h3 class="subscription-progress-mini__title text-sm font-semibold">
            {{ t('subscriptionProgress.title') }}
          </h3>
          <p class="subscription-progress-mini__meta mt-0.5 text-xs">
            {{ t('subscriptionProgress.activeCount', { count: activeSubscriptions.length }) }}
          </p>
        </div>

        <div class="subscription-progress-mini__list overflow-y-auto">
          <div
            v-for="subscription in displaySubscriptions"
            :key="subscription.id"
            class="subscription-progress-mini__item subscription-progress-mini__item-spacing last:border-b-0"
          >
            <div class="mb-2 flex items-center justify-between">
              <span class="subscription-progress-mini__group text-sm font-medium">
                {{ subscription.group?.name || `Group #${subscription.group_id}` }}
              </span>
              <span
                v-if="subscription.expires_at"
                class="subscription-progress-mini__expiry text-xs"
                :class="getDaysRemainingClass(subscription.expires_at)"
              >
                {{ formatDaysRemaining(subscription.expires_at) }}
              </span>
            </div>

            <!-- Progress bars or Unlimited badge -->
            <div class="space-y-1.5">
              <!-- Unlimited subscription badge -->
              <div
                v-if="isUnlimited(subscription)"
                class="subscription-progress-mini__unlimited subscription-progress-mini__unlimited-spacing flex items-center gap-2"
              >
                <span class="subscription-progress-mini__unlimited-symbol text-lg">∞</span>
                <span class="subscription-progress-mini__unlimited-label text-xs font-medium">
                  {{ t('subscriptionProgress.unlimited') }}
                </span>
              </div>

              <!-- Progress bars for limited subscriptions -->
              <template v-else>
                <div v-if="subscription.group?.daily_limit_usd" class="flex items-center gap-2">
                  <span class="subscription-progress-mini__metric-label">{{
                    t('subscriptionProgress.daily')
                  }}</span>
                  <div class="subscription-progress-mini__progress-track">
                    <div
                      class="subscription-progress-mini__progress-fill transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.daily_usage_usd,
                          subscription.group?.daily_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.daily_usage_usd,
                          subscription.group?.daily_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="subscription-progress-mini__metric-value">
                    {{
                      formatUsage(subscription.daily_usage_usd, subscription.group?.daily_limit_usd)
                    }}
                  </span>
                </div>

                <div v-if="subscription.group?.weekly_limit_usd" class="flex items-center gap-2">
                  <span class="subscription-progress-mini__metric-label">{{
                    t('subscriptionProgress.weekly')
                  }}</span>
                  <div class="subscription-progress-mini__progress-track">
                    <div
                      class="subscription-progress-mini__progress-fill transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.weekly_usage_usd,
                          subscription.group?.weekly_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.weekly_usage_usd,
                          subscription.group?.weekly_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="subscription-progress-mini__metric-value">
                    {{
                      formatUsage(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd)
                    }}
                  </span>
                </div>

                <div v-if="subscription.group?.monthly_limit_usd" class="flex items-center gap-2">
                  <span class="subscription-progress-mini__metric-label">{{
                    t('subscriptionProgress.monthly')
                  }}</span>
                  <div class="subscription-progress-mini__progress-track">
                    <div
                      class="subscription-progress-mini__progress-fill transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.monthly_usage_usd,
                          subscription.group?.monthly_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.monthly_usage_usd,
                          subscription.group?.monthly_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="subscription-progress-mini__metric-value">
                    {{
                      formatUsage(
                        subscription.monthly_usage_usd,
                        subscription.group?.monthly_limit_usd
                      )
                    }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="subscription-progress-mini__footer subscription-progress-mini__footer-spacing">
          <router-link
            to="/subscriptions"
            @click="closeTooltip"
            class="subscription-progress-mini__footer-link subscription-progress-mini__footer-link-spacing block w-full text-center text-xs"
          >
            {{ t('subscriptionProgress.viewAll') }}
          </router-link>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useSubscriptionStore } from '@/stores'
import type { UserSubscription } from '@/types'

const { t } = useI18n()

const subscriptionStore = useSubscriptionStore()

const containerRef = ref<HTMLElement | null>(null)
const tooltipOpen = ref(false)

// Use store data instead of local state
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)
const hasActiveSubscriptions = computed(() => subscriptionStore.hasActiveSubscriptions)

const displaySubscriptions = computed(() => {
  // Sort by most usage (highest percentage first)
  return [...activeSubscriptions.value].sort((a, b) => {
    const aMax = getMaxUsagePercentage(a)
    const bMax = getMaxUsagePercentage(b)
    return bMax - aMax
  })
})

function getMaxUsagePercentage(sub: UserSubscription): number {
  const percentages: number[] = []
  if (sub.group?.daily_limit_usd) {
    percentages.push(((sub.daily_usage_usd || 0) / sub.group.daily_limit_usd) * 100)
  }
  if (sub.group?.weekly_limit_usd) {
    percentages.push(((sub.weekly_usage_usd || 0) / sub.group.weekly_limit_usd) * 100)
  }
  if (sub.group?.monthly_limit_usd) {
    percentages.push(((sub.monthly_usage_usd || 0) / sub.group.monthly_limit_usd) * 100)
  }
  return percentages.length > 0 ? Math.max(...percentages) : 0
}

function isUnlimited(sub: UserSubscription): boolean {
  return (
    !sub.group?.daily_limit_usd &&
    !sub.group?.weekly_limit_usd &&
    !sub.group?.monthly_limit_usd
  )
}

function getProgressDotClass(sub: UserSubscription): string {
  // Unlimited subscriptions get a special color
  if (isUnlimited(sub)) {
    return 'subscription-progress-mini__state-dot--unlimited'
  }
  const maxPercentage = getMaxUsagePercentage(sub)
  if (maxPercentage >= 90) return 'subscription-progress-mini__state-dot--critical'
  if (maxPercentage >= 70) return 'subscription-progress-mini__state-dot--warning'
  return 'subscription-progress-mini__state-dot--healthy'
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'subscription-progress-mini__progress-fill--muted'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'subscription-progress-mini__progress-fill--critical'
  if (percentage >= 70) return 'subscription-progress-mini__progress-fill--warning'
  return 'subscription-progress-mini__progress-fill--healthy'
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function formatUsage(used: number | undefined, limit: number | null | undefined): string {
  const usedValue = (used || 0).toFixed(2)
  const limitValue = limit?.toFixed(2) || '∞'
  return `$${usedValue}/$${limitValue}`
}

function formatDaysRemaining(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  if (diff < 0) return t('subscriptionProgress.expired')
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return t('subscriptionProgress.expiresToday')
  if (days === 1) return t('subscriptionProgress.expiresTomorrow')
  return t('subscriptionProgress.daysRemaining', { days })
}

function getDaysRemainingClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days <= 3) return 'subscription-progress-mini__expiry--critical'
  if (days <= 7) return 'subscription-progress-mini__expiry--warning'
  return 'subscription-progress-mini__expiry--normal'
}

function toggleTooltip() {
  tooltipOpen.value = !tooltipOpen.value
}

function closeTooltip() {
  tooltipOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    closeTooltip()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  // Trigger initial fetch if not already loaded
  // The actual data loading is handled by App.vue globally
  subscriptionStore.fetchActiveSubscriptions().catch((error) => {
    console.error('Failed to load subscriptions in SubscriptionProgressMini:', error)
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}

.subscription-progress-mini {
  --subscription-progress-mini-danger-rgb: var(--theme-danger-rgb);
  --subscription-progress-mini-warning-rgb: var(--theme-warning-rgb);
  --subscription-progress-mini-success-rgb: var(--theme-success-rgb);
}

.subscription-progress-mini__trigger,
.subscription-progress-mini__panel,
.subscription-progress-mini__unlimited,
.subscription-progress-mini__footer-link {
  border-radius: var(--theme-subscription-panel-radius);
}

.subscription-progress-mini__trigger {
  background: color-mix(in srgb, var(--theme-accent-soft) 74%, var(--theme-surface));
}

.subscription-progress-mini__trigger-spacing {
  padding: var(--theme-subscription-mini-trigger-padding-y)
    var(--theme-subscription-mini-trigger-padding-x);
}

.subscription-progress-mini__trigger:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 92%, var(--theme-surface));
}

.subscription-progress-mini__icon,
.subscription-progress-mini__count {
  color: var(--theme-accent);
}

.subscription-progress-mini__panel {
  margin-top: var(--theme-subscription-mini-panel-offset);
  width: min(calc(100vw - 2rem), var(--theme-subscription-panel-width));
  border: 1px solid var(--theme-dropdown-border);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.subscription-progress-mini__header,
.subscription-progress-mini__footer {
  border-color: var(--theme-page-border);
}

.subscription-progress-mini__header {
  border-bottom: 1px solid var(--theme-page-border);
}

.subscription-progress-mini__header-spacing {
  padding: var(--theme-subscription-mini-header-padding);
}

.subscription-progress-mini__footer {
  border-top: 1px solid var(--theme-page-border);
}

.subscription-progress-mini__title,
.subscription-progress-mini__group {
  color: var(--theme-page-text);
}

.subscription-progress-mini__meta,
.subscription-progress-mini__metric-label,
.subscription-progress-mini__metric-value,
.subscription-progress-mini__expiry--normal {
  color: var(--theme-page-muted);
}

.subscription-progress-mini__item {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 66%, transparent);
}

.subscription-progress-mini__item-spacing {
  padding: var(--theme-subscription-mini-item-padding);
}

.subscription-progress-mini__unlimited {
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, rgb(var(--subscription-progress-mini-success-rgb)) 12%, var(--theme-surface)),
      color-mix(in srgb, rgb(var(--subscription-progress-mini-success-rgb)) 7%, var(--theme-surface-soft))
    );
}

.subscription-progress-mini__unlimited-spacing {
  padding: var(--theme-subscription-mini-unlimited-padding-y)
    var(--theme-subscription-mini-unlimited-padding-x);
}

.subscription-progress-mini__unlimited-symbol,
.subscription-progress-mini__unlimited-label {
  color: rgb(var(--subscription-progress-mini-success-rgb));
}

.subscription-progress-mini__progress-track {
  flex: 1 1 0;
  min-width: 0;
  height: var(--theme-subscription-mini-progress-height);
  border-radius: 999px;
  background: color-mix(in srgb, var(--theme-page-border) 84%, transparent);
}

.subscription-progress-mini__progress-fill {
  height: var(--theme-subscription-mini-progress-height);
  border-radius: 999px;
}

.subscription-progress-mini__state-dot {
  width: var(--theme-subscription-mini-dot-size);
  height: var(--theme-subscription-mini-dot-size);
  border-radius: 999px;
}

.subscription-progress-mini__list {
  max-height: var(--theme-subscription-mini-list-max-height);
}

.subscription-progress-mini__metric-label {
  width: var(--theme-subscription-mini-label-width);
  flex-shrink: 0;
  font-size: var(--theme-subscription-mini-metric-font-size);
}

.subscription-progress-mini__metric-value {
  width: var(--theme-subscription-mini-value-width);
  flex-shrink: 0;
  text-align: right;
  font-size: var(--theme-subscription-mini-metric-font-size);
}

.subscription-progress-mini__progress-fill--muted {
  background: color-mix(in srgb, var(--theme-page-muted) 60%, transparent);
}

.subscription-progress-mini__state-dot--critical,
.subscription-progress-mini__progress-fill--critical,
.subscription-progress-mini__expiry--critical {
  color: rgb(var(--subscription-progress-mini-danger-rgb));
  background: rgb(var(--subscription-progress-mini-danger-rgb));
}

.subscription-progress-mini__state-dot--warning,
.subscription-progress-mini__progress-fill--warning,
.subscription-progress-mini__expiry--warning {
  color: rgb(var(--subscription-progress-mini-warning-rgb));
  background: rgb(var(--subscription-progress-mini-warning-rgb));
}

.subscription-progress-mini__state-dot--healthy,
.subscription-progress-mini__progress-fill--healthy,
.subscription-progress-mini__state-dot--unlimited {
  background: rgb(var(--subscription-progress-mini-success-rgb));
}

.subscription-progress-mini__footer-link {
  color: var(--theme-accent);
}

.subscription-progress-mini__footer-spacing {
  padding: var(--theme-subscription-mini-footer-padding);
}

.subscription-progress-mini__footer-link-spacing {
  padding-block: var(--theme-subscription-mini-footer-link-padding-y);
}

.subscription-progress-mini__footer-link:hover {
  background: var(--theme-dropdown-item-hover-bg);
}
</style>
