<template>
  <div class="flex flex-col gap-1.5">
    <div class="flex items-center gap-1.5">
      <span :class="concurrencyClass">
        <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
        </svg>
        <span class="font-mono">{{ currentConcurrency }}</span>
        <span class="account-capacity-cell__separator">/</span>
        <span class="font-mono">{{ account.concurrency }}</span>
      </span>
    </div>

    <div v-if="showWindowCost" class="flex items-center gap-1">
      <span :class="windowCostClass" :title="windowCostTooltip">
        <svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="font-mono">${{ formatCost(currentWindowCost) }}</span>
        <span class="account-capacity-cell__separator">/</span>
        <span class="font-mono">${{ formatCost(account.window_cost_limit) }}</span>
      </span>
    </div>

    <div v-if="showSessionLimit" class="flex items-center gap-1">
      <span :class="sessionLimitClass" :title="sessionLimitTooltip">
        <svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
        </svg>
        <span class="font-mono">{{ activeSessions }}</span>
        <span class="account-capacity-cell__separator">/</span>
        <span class="font-mono">{{ account.max_sessions }}</span>
      </span>
    </div>

    <div v-if="showRpmLimit" class="flex items-center gap-1">
      <span :class="rpmClass" :title="rpmTooltip">
        <svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
        </svg>
        <span class="font-mono">{{ currentRPM }}</span>
        <span class="account-capacity-cell__separator">/</span>
        <span class="font-mono">{{ account.base_rpm }}</span>
        <span class="account-capacity-cell__strategy-tag text-[9px]">{{ rpmStrategyTag }}</span>
      </span>
    </div>

    <QuotaBadge v-if="showDailyQuota" :used="account.quota_daily_used ?? 0" :limit="account.quota_daily_limit!" label="D" />
    <QuotaBadge v-if="showWeeklyQuota" :used="account.quota_weekly_used ?? 0" :limit="account.quota_weekly_limit!" label="W" />
    <QuotaBadge v-if="showTotalQuota" :used="account.quota_used ?? 0" :limit="account.quota_limit!" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'
import QuotaBadge from './QuotaBadge.vue'

type CapacityTone = 'neutral' | 'success' | 'warning' | 'brand' | 'danger'

const props = defineProps<{
  account: Account
}>()

const { t } = useI18n()

const buildMetricBadgeClass = (tone: CapacityTone, compact = true) => [
  'theme-chip',
  compact ? 'theme-chip--compact' : 'theme-chip--regular',
  'account-capacity-cell__metric-badge',
  `account-capacity-cell__metric-badge--${tone}`
]

const currentConcurrency = computed(() => props.account.current_concurrency ?? 0)

const isAnthropicOAuthOrSetupToken = computed(() => {
  return (
    props.account.platform === 'anthropic' &&
    (props.account.type === 'oauth' || props.account.type === 'setup-token')
  )
})

const showWindowCost = computed(() => {
  return (
    isAnthropicOAuthOrSetupToken.value &&
    props.account.window_cost_limit !== undefined &&
    props.account.window_cost_limit !== null &&
    props.account.window_cost_limit > 0
  )
})

const currentWindowCost = computed(() => props.account.current_window_cost ?? 0)

const showSessionLimit = computed(() => {
  return (
    isAnthropicOAuthOrSetupToken.value &&
    props.account.max_sessions !== undefined &&
    props.account.max_sessions !== null &&
    props.account.max_sessions > 0
  )
})

const activeSessions = computed(() => props.account.active_sessions ?? 0)

const concurrencyClass = computed(() => {
  const current = currentConcurrency.value
  const max = props.account.concurrency

  if (current >= max) {
    return buildMetricBadgeClass('danger', false)
  }
  if (current > 0) {
    return buildMetricBadgeClass('warning', false)
  }
  return buildMetricBadgeClass('neutral', false)
})

const windowCostClass = computed(() => {
  if (!showWindowCost.value) return []

  const current = currentWindowCost.value
  const limit = props.account.window_cost_limit || 0
  const reserve = props.account.window_cost_sticky_reserve || 10

  if (current >= limit + reserve) {
    return buildMetricBadgeClass('danger')
  }
  if (current >= limit) {
    return buildMetricBadgeClass('brand')
  }
  if (current >= limit * 0.8) {
    return buildMetricBadgeClass('warning')
  }
  return buildMetricBadgeClass('success')
})

const windowCostTooltip = computed(() => {
  if (!showWindowCost.value) return ''

  const current = currentWindowCost.value
  const limit = props.account.window_cost_limit || 0
  const reserve = props.account.window_cost_sticky_reserve || 10

  if (current >= limit + reserve) {
    return t('admin.accounts.capacity.windowCost.blocked')
  }
  if (current >= limit) {
    return t('admin.accounts.capacity.windowCost.stickyOnly')
  }
  return t('admin.accounts.capacity.windowCost.normal')
})

const sessionLimitClass = computed(() => {
  if (!showSessionLimit.value) return []

  const current = activeSessions.value
  const max = props.account.max_sessions || 0

  if (current >= max) {
    return buildMetricBadgeClass('danger')
  }
  if (current >= max * 0.8) {
    return buildMetricBadgeClass('warning')
  }
  return buildMetricBadgeClass('success')
})

const sessionLimitTooltip = computed(() => {
  if (!showSessionLimit.value) return ''

  const current = activeSessions.value
  const max = props.account.max_sessions || 0
  const idle = props.account.session_idle_timeout_minutes || 5

  if (current >= max) {
    return t('admin.accounts.capacity.sessions.full', { idle })
  }
  return t('admin.accounts.capacity.sessions.normal', { idle })
})

const showRpmLimit = computed(() => {
  return (
    isAnthropicOAuthOrSetupToken.value &&
    props.account.base_rpm !== undefined &&
    props.account.base_rpm !== null &&
    props.account.base_rpm > 0
  )
})

const currentRPM = computed(() => props.account.current_rpm ?? 0)

const rpmStrategy = computed(() => props.account.rpm_strategy || 'tiered')

const rpmStrategyTag = computed(() => {
  return rpmStrategy.value === 'sticky_exempt' ? '[S]' : '[T]'
})

const rpmBuffer = computed(() => {
  const base = props.account.base_rpm || 0
  return props.account.rpm_sticky_buffer ?? (base > 0 ? Math.max(1, Math.floor(base / 5)) : 0)
})

const rpmClass = computed(() => {
  if (!showRpmLimit.value) return []

  const current = currentRPM.value
  const base = props.account.base_rpm ?? 0
  const buffer = rpmBuffer.value

  if (rpmStrategy.value === 'tiered') {
    if (current >= base + buffer) {
      return buildMetricBadgeClass('danger')
    }
    if (current >= base) {
      return buildMetricBadgeClass('brand')
    }
  } else {
    if (current >= base) {
      return buildMetricBadgeClass('brand')
    }
  }
  if (current >= base * 0.8) {
    return buildMetricBadgeClass('warning')
  }
  return buildMetricBadgeClass('success')
})

const rpmTooltip = computed(() => {
  if (!showRpmLimit.value) return ''

  const current = currentRPM.value
  const base = props.account.base_rpm ?? 0
  const buffer = rpmBuffer.value

  if (rpmStrategy.value === 'tiered') {
    if (current >= base + buffer) {
      return t('admin.accounts.capacity.rpm.tieredBlocked', { buffer })
    }
    if (current >= base) {
      return t('admin.accounts.capacity.rpm.tieredStickyOnly', { buffer })
    }
    if (current >= base * 0.8) {
      return t('admin.accounts.capacity.rpm.tieredWarning')
    }
    return t('admin.accounts.capacity.rpm.tieredNormal')
  } else {
    if (current >= base) {
      return t('admin.accounts.capacity.rpm.stickyExemptOver')
    }
    if (current >= base * 0.8) {
      return t('admin.accounts.capacity.rpm.stickyExemptWarning')
    }
    return t('admin.accounts.capacity.rpm.stickyExemptNormal')
  }
})

const isQuotaEligible = computed(() => props.account.type === 'apikey' || props.account.type === 'bedrock')

const showDailyQuota = computed(() => {
  return isQuotaEligible.value && (props.account.quota_daily_limit ?? 0) > 0
})

const showWeeklyQuota = computed(() => {
  return isQuotaEligible.value && (props.account.quota_weekly_limit ?? 0) > 0
})

const showTotalQuota = computed(() => {
  return isQuotaEligible.value && (props.account.quota_limit ?? 0) > 0
})

// 格式化费用显示
const formatCost = (value: number | null | undefined) => {
  if (value === null || value === undefined) return '0'
  return value.toFixed(2)
}
</script>

<style scoped>
.account-capacity-cell__separator,
.account-capacity-cell__strategy-tag {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, var(--theme-surface));
}

.account-capacity-cell__metric-badge {
  min-height: 1.25rem;
}

.account-capacity-cell__metric-badge--neutral {
  --theme-chip-bg: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  --theme-chip-fg: var(--theme-page-muted);
}

.account-capacity-cell__metric-badge--success {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-success-rgb)) 18%, var(--theme-card-border));
}

.account-capacity-cell__metric-badge--warning {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 18%, var(--theme-card-border));
}

.account-capacity-cell__metric-badge--brand {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 18%, var(--theme-card-border));
}

.account-capacity-cell__metric-badge--danger {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 18%, var(--theme-card-border));
}
</style>
