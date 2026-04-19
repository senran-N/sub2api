<template>
  <div>
    <div
      v-if="windowStats && (windowStats.requests > 0 || windowStats.tokens > 0)"
      class="mb-0.5 flex items-center"
    >
      <div class="usage-progress-bar__stats-row flex items-center gap-1.5 text-[9px]">
        <span class="usage-progress-bar__stat-chip usage-progress-bar__stat-chip-spacing rounded">
          {{ formatRequests }} req
        </span>
        <span class="usage-progress-bar__stat-chip usage-progress-bar__stat-chip-spacing rounded">
          {{ formatTokens }}
        </span>
        <span
          class="usage-progress-bar__stat-chip usage-progress-bar__stat-chip-spacing rounded"
          :title="t('usage.accountBilled')"
        >
          A ${{ formatAccountCost }}
        </span>
        <span
          v-if="windowStats?.user_cost != null"
          class="usage-progress-bar__stat-chip usage-progress-bar__stat-chip-spacing rounded"
          :title="t('usage.userBilled')"
        >
          U ${{ formatUserCost }}
        </span>
      </div>
    </div>

    <div class="flex items-center gap-1">
      <span :class="labelClass">
        {{ label }}
      </span>

      <div class="usage-progress-bar__track h-1.5 w-8 shrink-0 overflow-hidden rounded-full">
        <div :class="barClass" :style="{ width: barWidth }"></div>
      </div>

      <span :class="textClass">
        {{ displayPercent }}
      </span>

      <span v-if="shouldShowResetTime" class="usage-progress-bar__reset-time shrink-0 text-[10px]">
        {{ formatResetTime }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import type { WindowStats } from '@/types'
import { formatCompactNumber } from '@/utils/format'

type WindowColor = 'indigo' | 'emerald' | 'purple' | 'amber'
type UtilizationTone = 'neutral' | 'success' | 'warning' | 'danger'

const props = defineProps<{
  label: string
  utilization: number
  resetsAt?: string | null
  color: WindowColor
  windowStats?: WindowStats | null
  showNowWhenIdle?: boolean
}>()

const { t } = useI18n()

const now = ref(new Date())
const { pause: pauseClock, resume: resumeClock } = useIntervalFn(
  () => {
    now.value = new Date()
  },
  60_000,
  { immediate: false },
)

const resetTimestamp = computed(() => {
  if (!props.resetsAt) {
    return null
  }

  const timestamp = new Date(props.resetsAt).getTime()
  return Number.isFinite(timestamp) ? timestamp : null
})

const effectiveUtilization = computed(() => {
  if (resetTimestamp.value != null && resetTimestamp.value <= now.value.getTime()) {
    return 0
  }
  return props.utilization
})

if (resetTimestamp.value != null) {
  resumeClock()
}

watch(
  resetTimestamp,
  (value) => {
    if (value != null) {
      now.value = new Date()
      resumeClock()
      return
    }
    pauseClock()
  },
)

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getUtilizationTone = (): UtilizationTone => {
  if (effectiveUtilization.value >= 100) {
    return 'danger'
  }
  if (effectiveUtilization.value >= 80) {
    return 'warning'
  }
  return 'success'
}

const labelClass = computed(() => {
  return joinClassNames(
    'usage-progress-bar__label usage-progress-bar__label-layout shrink-0 rounded text-center text-[10px] font-medium',
    `usage-progress-bar__label--${props.color}`
  )
})

const barClass = computed(() => {
  return joinClassNames(
    'usage-progress-bar__fill h-full transition-all duration-300',
    `usage-progress-bar__fill--${getUtilizationTone()}`
  )
})

const textClass = computed(() => {
  const tone = effectiveUtilization.value >= 100
    ? 'danger'
    : effectiveUtilization.value >= 80
      ? 'warning'
      : 'neutral'

  return joinClassNames(
    'usage-progress-bar__percent usage-progress-bar__percent-layout shrink-0 text-right text-[10px] font-medium',
    `usage-progress-bar__percent--${tone}`
  )
})

const barWidth = computed(() => `${Math.min(effectiveUtilization.value, 100)}%`)

const displayPercent = computed(() => {
  const percent = Math.round(effectiveUtilization.value)
  return percent > 999 ? '>999%' : `${percent}%`
})

const shouldShowResetTime = computed(() => {
  if (props.resetsAt) {
    return true
  }
  return Boolean(props.showNowWhenIdle && effectiveUtilization.value <= 0)
})

const formatResetTime = computed(() => {
  if (props.showNowWhenIdle && effectiveUtilization.value <= 0) {
    return '现在'
  }

  if (resetTimestamp.value == null) {
    return '-'
  }

  const diffMs = resetTimestamp.value - now.value.getTime()

  if (diffMs <= 0) {
    return '现在'
  }

  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))

  if (diffHours >= 24) {
    const days = Math.floor(diffHours / 24)
    return `${days}d ${diffHours % 24}h`
  }
  if (diffHours > 0) {
    return `${diffHours}h ${diffMins}m`
  }
  return `${diffMins}m`
})

const formatRequests = computed(() => {
  if (!props.windowStats) {
    return ''
  }
  return formatCompactNumber(props.windowStats.requests, { allowBillions: false })
})

const formatTokens = computed(() => {
  if (!props.windowStats) {
    return ''
  }
  return formatCompactNumber(props.windowStats.tokens)
})

const formatAccountCost = computed(() => {
  if (!props.windowStats) {
    return '0.00'
  }
  return props.windowStats.cost.toFixed(2)
})

const formatUserCost = computed(() => {
  if (!props.windowStats || props.windowStats.user_cost == null) {
    return '0.00'
  }
  return props.windowStats.user_cost.toFixed(2)
})
</script>

<style scoped>
.usage-progress-bar__stats-row,
.usage-progress-bar__reset-time,
.usage-progress-bar__percent--neutral {
  color: var(--theme-page-muted);
}

.usage-progress-bar__stat-chip {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.usage-progress-bar__stat-chip-spacing {
  padding: var(--theme-usage-progress-chip-padding-y) var(--theme-usage-progress-chip-padding-x);
}

.usage-progress-bar__label {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.usage-progress-bar__label-layout {
  width: var(--theme-usage-progress-label-width);
  padding-inline: var(--theme-usage-progress-label-padding-x);
}

.usage-progress-bar__label--indigo {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.usage-progress-bar__label--emerald {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.usage-progress-bar__label--purple {
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.usage-progress-bar__label--amber {
  background: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 84%, var(--theme-page-text));
}

.usage-progress-bar__track {
  background: color-mix(in srgb, var(--theme-page-border) 78%, var(--theme-surface));
}

.usage-progress-bar__fill--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 88%, var(--theme-page-text));
}

.usage-progress-bar__fill--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 90%, var(--theme-page-text));
}

.usage-progress-bar__fill--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 90%, var(--theme-page-text));
}

.usage-progress-bar__percent--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.usage-progress-bar__percent-layout {
  width: var(--theme-usage-progress-percent-width);
}

.usage-progress-bar__percent--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
