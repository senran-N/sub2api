<template>
  <div class="flex items-center gap-2">
    <div v-if="isRateLimited" class="flex flex-col items-center gap-1">
      <span class="badge text-xs badge-warning">{{ t('admin.accounts.status.rateLimited') }}</span>
      <span class="account-status-indicator__countdown text-[11px]">{{ rateLimitResumeText }}</span>
    </div>

    <div v-else-if="isOverloaded" class="flex flex-col items-center gap-1">
      <span class="badge text-xs badge-danger">{{ t('admin.accounts.status.overloaded') }}</span>
      <span class="account-status-indicator__countdown text-[11px]">{{ overloadCountdown }}</span>
    </div>

    <template v-else>
      <button
        v-if="isTempUnschedulable"
        type="button"
        :class="['badge text-xs', statusClass, 'cursor-pointer']"
        :title="t('admin.accounts.status.viewTempUnschedDetails')"
        @click="handleTempUnschedClick"
      >
        {{ statusText }}
      </button>
      <span v-else :class="['badge text-xs', statusClass]">
        {{ statusText }}
      </span>
    </template>

    <div v-if="hasError && account.error_message" class="group/error relative">
      <svg
        class="account-status-indicator__error-icon h-4 w-4 cursor-help transition-colors"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z"
        />
      </svg>
      <div
        class="account-status-indicator__tooltip invisible absolute left-0 top-full z-[100] text-xs opacity-0 shadow-xl transition-all duration-200 group-hover/error:visible group-hover/error:opacity-100"
      >
        <div class="account-status-indicator__tooltip-body whitespace-pre-wrap break-words leading-relaxed">
          {{ account.error_message }}
        </div>
        <div
          class="account-status-indicator__tooltip-arrow absolute bottom-full left-3 border-[6px] border-transparent"
        ></div>
      </div>
    </div>

    <div v-if="isRateLimited" class="group relative">
      <span :class="getSmallIndicatorClass('warning')">
        <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
        429
      </span>
      <div
        class="account-status-indicator__floating-tooltip pointer-events-none absolute bottom-full left-1/2 z-50 -translate-x-1/2 whitespace-normal text-center text-xs leading-relaxed opacity-0 transition-opacity group-hover:opacity-100"
      >
        {{ t('admin.accounts.status.rateLimitedUntil', { time: formatDateTime(account.rate_limit_reset_at) }) }}
        <div
          class="account-status-indicator__floating-tooltip-arrow absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent"
        ></div>
      </div>
    </div>

    <div
      v-if="activeModelStatuses.length > 0"
      :class="[
        activeModelStatuses.length <= 4
          ? 'flex flex-col gap-1'
          : activeModelStatuses.length <= 8
            ? 'columns-2 gap-x-2'
            : 'columns-3 gap-x-2'
      ]"
    >
      <div v-for="item in activeModelStatuses" :key="`${item.kind}-${item.model}`" class="group relative mb-1 break-inside-avoid">
        <span
          v-if="item.kind === 'credits_exhausted'"
          :class="getSmallIndicatorClass('danger')"
        >
          <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
          {{ t('admin.accounts.status.creditsExhausted') }}
          <span class="account-status-indicator__indicator-meta">{{ formatModelResetTime(item.reset_at) }}</span>
        </span>
        <span
          v-else-if="item.kind === 'credits_active'"
          :class="getSmallIndicatorClass('warning')"
        >
          <span>⚡</span>
          {{ formatScopeName(item.model) }}
          <span class="account-status-indicator__indicator-meta">{{ formatModelResetTime(item.reset_at) }}</span>
        </span>
        <span
          v-else
          :class="getSmallIndicatorClass('brand')"
        >
          <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
          {{ formatScopeName(item.model) }}
          <span class="account-status-indicator__indicator-meta">{{ formatModelResetTime(item.reset_at) }}</span>
        </span>
        <div
          class="account-status-indicator__floating-tooltip pointer-events-none absolute bottom-full left-1/2 z-50 -translate-x-1/2 whitespace-normal text-center text-xs leading-relaxed opacity-0 transition-opacity group-hover:opacity-100"
        >
          {{
            item.kind === 'credits_exhausted'
              ? t('admin.accounts.status.creditsExhaustedUntil', { time: formatTime(item.reset_at) })
              : item.kind === 'credits_active'
                ? t('admin.accounts.status.modelCreditOveragesUntil', { model: formatScopeName(item.model), time: formatTime(item.reset_at) })
                : t('admin.accounts.status.modelRateLimitedUntil', { model: formatScopeName(item.model), time: formatTime(item.reset_at) })
          }}
          <div
            class="account-status-indicator__floating-tooltip-arrow absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent"
          ></div>
        </div>
      </div>
    </div>

    <div v-if="isOverloaded" class="group relative">
      <span :class="getSmallIndicatorClass('danger')">
        <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
        529
      </span>
      <div
        class="account-status-indicator__floating-tooltip pointer-events-none absolute bottom-full left-1/2 z-50 -translate-x-1/2 whitespace-normal text-center text-xs leading-relaxed opacity-0 transition-opacity group-hover:opacity-100"
      >
        {{ t('admin.accounts.status.overloadedUntil', { time: formatTime(account.overload_until) }) }}
        <div
          class="account-status-indicator__floating-tooltip-arrow absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { Account } from '@/types'
import { formatCountdown, formatDateTime, formatCountdownWithSuffix, formatTime } from '@/utils/format'

const { t } = useI18n()

const props = defineProps<{
  account: Account
}>()

const emit = defineEmits<{
  (e: 'show-temp-unsched', account: Account): void
}>()

type IndicatorTone = 'warning' | 'danger' | 'brand'

// Computed: is rate limited (429)
const isRateLimited = computed(() => {
  if (!props.account.rate_limit_reset_at) return false
  return new Date(props.account.rate_limit_reset_at) > new Date()
})

type AccountModelStatusItem = {
  kind: 'rate_limit' | 'credits_exhausted' | 'credits_active'
  model: string
  reset_at: string
}

// Computed: active model statuses (普通模型限流 + 积分耗尽 + 走积分中)
const activeModelStatuses = computed<AccountModelStatusItem[]>(() => {
  const extra = props.account.extra as Record<string, unknown> | undefined
  const modelLimits = extra?.model_rate_limits as
    | Record<string, { rate_limited_at: string; rate_limit_reset_at: string }>
    | undefined
  const now = new Date()
  const items: AccountModelStatusItem[] = []

  if (!modelLimits) return items

  // 检查 AICredits key 是否生效（积分是否耗尽）
  const aiCreditsEntry = modelLimits['AICredits']
  const hasActiveAICredits = aiCreditsEntry && new Date(aiCreditsEntry.rate_limit_reset_at) > now
  const allowOverages = !!(extra?.allow_overages)

  for (const [model, info] of Object.entries(modelLimits)) {
    if (new Date(info.rate_limit_reset_at) <= now) continue

    if (model === 'AICredits') {
      // AICredits key → 积分已用尽
      items.push({ kind: 'credits_exhausted', model, reset_at: info.rate_limit_reset_at })
    } else if (allowOverages && !hasActiveAICredits) {
      // 普通模型限流 + overages 启用 + 积分可用 → 正在走积分
      items.push({ kind: 'credits_active', model, reset_at: info.rate_limit_reset_at })
    } else {
      // 普通模型限流
      items.push({ kind: 'rate_limit', model, reset_at: info.rate_limit_reset_at })
    }
  }

  return items
})

const formatScopeName = (scope: string): string => {
  const aliases: Record<string, string> = {
    // Claude 系列
    'claude-opus-4-6': 'COpus46',
    'claude-opus-4-6-thinking': 'COpus46T',
    'claude-sonnet-4-6': 'CSon46',
    'claude-sonnet-4-5': 'CSon45',
    'claude-sonnet-4-5-thinking': 'CSon45T',
    // Gemini 2.5 系列
    'gemini-2.5-flash': 'G25F',
    'gemini-2.5-flash-lite': 'G25FL',
    'gemini-2.5-flash-thinking': 'G25FT',
    'gemini-2.5-pro': 'G25P',
    'gemini-2.5-flash-image': 'G25I',
    // Gemini 3 系列
    'gemini-3-flash': 'G3F',
    'gemini-3.1-pro-high': 'G3PH',
    'gemini-3.1-pro-low': 'G3PL',
    'gemini-3-pro-image': 'G3PI',
    'gemini-3.1-flash-image': 'G31FI',
    // 其他
    'gpt-oss-120b-medium': 'GPT120',
    'tab_flash_lite_preview': 'TabFL',
    // 旧版 scope 别名（兼容）
    claude: 'Claude',
    claude_sonnet: 'CSon',
    claude_opus: 'COpus',
    claude_haiku: 'CHaiku',
    gemini_text: 'Gemini',
    gemini_image: 'GImg',
    gemini_flash: 'GFlash',
    gemini_pro: 'GPro',
  }
  return aliases[scope] || scope
}

const formatModelResetTime = (resetAt: string): string => {
  const date = new Date(resetAt)
  const now = new Date()
  const diffMs = date.getTime() - now.getTime()
  if (diffMs <= 0) return ''
  const totalSecs = Math.floor(diffMs / 1000)
  const h = Math.floor(totalSecs / 3600)
  const m = Math.floor((totalSecs % 3600) / 60)
  const s = totalSecs % 60
  if (h > 0) return `${h}h${m}m`
  if (m > 0) return `${m}m${s}s`
  return `${s}s`
}

// Computed: is overloaded (529)
const isOverloaded = computed(() => {
  if (!props.account.overload_until) return false
  return new Date(props.account.overload_until) > new Date()
})

// Computed: is temp unschedulable
const isTempUnschedulable = computed(() => {
  if (!props.account.temp_unschedulable_until) return false
  return new Date(props.account.temp_unschedulable_until) > new Date()
})

// Computed: has error status
const hasError = computed(() => {
  return props.account.status === 'error'
})

const isQuotaExceeded = computed(() => {
  const exceeded = (used?: number | null, limit?: number | null) =>
    typeof limit === 'number' && limit > 0 && typeof used === 'number' && used >= limit

  return (
    exceeded(props.account.quota_used, props.account.quota_limit) ||
    exceeded(props.account.quota_daily_used, props.account.quota_daily_limit) ||
    exceeded(props.account.quota_weekly_used, props.account.quota_weekly_limit)
  )
})

// Computed: countdown text for rate limit (429)
const rateLimitCountdown = computed(() => {
  return formatCountdown(props.account.rate_limit_reset_at)
})

const rateLimitResumeText = computed(() => {
  if (!rateLimitCountdown.value) return ''
  return t('admin.accounts.status.rateLimitedAutoResume', { time: rateLimitCountdown.value })
})

// Computed: countdown text for overload (529)
const overloadCountdown = computed(() => {
  return formatCountdownWithSuffix(props.account.overload_until)
})

// Computed: status badge class
const statusClass = computed(() => {
  if (hasError.value) {
    return 'badge-danger'
  }
  if (isTempUnschedulable.value) {
    return 'badge-warning'
  }
  if (props.account.status !== 'active') {
    return props.account.status === 'error' ? 'badge-danger' : 'badge-gray'
  }
  if (isQuotaExceeded.value) {
    return 'badge-warning'
  }
  if (!props.account.schedulable) {
    return 'badge-gray'
  }
  return 'badge-success'
})

// Computed: status text
const statusText = computed(() => {
  if (hasError.value) {
    return t('admin.accounts.status.error')
  }
  if (isTempUnschedulable.value) {
    return t('admin.accounts.status.tempUnschedulable')
  }
  if (props.account.status !== 'active') {
    return t(`admin.accounts.status.${props.account.status}`)
  }
  if (isQuotaExceeded.value) {
    return t('admin.accounts.status.quotaExceeded')
  }
  if (!props.account.schedulable) {
    return t('admin.accounts.status.paused')
  }
  return t(`admin.accounts.status.${props.account.status}`)
})

const handleTempUnschedClick = () => {
  if (!isTempUnschedulable.value) return
  emit('show-temp-unsched', props.account)
}

const getSmallIndicatorClass = (tone: IndicatorTone) => [
  'theme-chip',
  'theme-chip--compact',
  'account-status-indicator__small-indicator',
  `account-status-indicator__small-indicator--${tone}`
]
</script>

<style scoped>
.account-status-indicator__countdown,
.account-status-indicator__indicator-meta {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, var(--theme-surface));
}

.account-status-indicator__indicator-meta {
  font-size: var(--theme-account-status-meta-font-size);
}

.account-status-indicator__error-icon {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 78%, var(--theme-page-text));
}

.account-status-indicator__error-icon:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 92%, var(--theme-page-text));
}

.account-status-indicator__tooltip,
.account-status-indicator__floating-tooltip {
  border-radius: var(--theme-account-status-tooltip-radius);
  padding: var(--theme-account-status-tooltip-padding-y)
    var(--theme-account-status-tooltip-padding-x);
  background: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
  color: var(--theme-surface-contrast-text);
}

.account-status-indicator__tooltip {
  margin-top: var(--theme-account-status-tooltip-margin-top);
  min-width: var(--theme-account-status-tooltip-min-width);
  max-width: var(--theme-account-status-tooltip-max-width);
}

.account-status-indicator__floating-tooltip {
  margin-bottom: var(--theme-account-status-tooltip-offset);
  width: var(--theme-account-status-tooltip-width);
}

.account-status-indicator__tooltip-body {
  color: color-mix(in srgb, var(--theme-surface-contrast-text) 70%, transparent);
}

.account-status-indicator__tooltip-arrow {
  border-bottom-color: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
}

.account-status-indicator__floating-tooltip-arrow {
  border-top-color: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
}

.account-status-indicator__small-indicator--warning {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 18%, var(--theme-card-border));
}

.account-status-indicator__small-indicator--danger {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 18%, var(--theme-card-border));
}

.account-status-indicator__small-indicator--brand {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 18%, var(--theme-card-border));
}
</style>
