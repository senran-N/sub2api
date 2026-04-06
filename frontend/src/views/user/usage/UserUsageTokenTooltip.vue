<template>
  <Teleport to="body">
    <div
      v-if="visible && log"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="tooltipStyle"
    >
      <div class="user-usage-token-tooltip__panel whitespace-nowrap text-xs shadow-xl">
        <div class="space-y-1.5">
          <div>
            <div class="user-usage-token-tooltip__title mb-1 text-xs font-semibold">{{ t('usage.tokenDetails') }}</div>
            <div v-if="log.input_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-token-tooltip__label">{{ t('admin.usage.inputTokens') }}</span>
              <span class="user-usage-token-tooltip__value font-medium">{{ log.input_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="log.output_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-token-tooltip__label">{{ t('admin.usage.outputTokens') }}</span>
              <span class="user-usage-token-tooltip__value font-medium">{{ log.output_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="log.cache_creation_tokens > 0">
              <template v-if="hasUserUsageCacheCreationBreakdown(log)">
                <div v-if="log.cache_creation_5m_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="user-usage-token-tooltip__label flex items-center gap-1.5">
                    {{ t('admin.usage.cacheCreation5mTokens') }}
                    <span class="user-usage-token-tooltip__badge user-usage-token-tooltip__badge--warning inline-flex items-center text-[10px] font-medium leading-tight">
                      5m
                    </span>
                  </span>
                  <span class="user-usage-token-tooltip__value font-medium">{{ log.cache_creation_5m_tokens.toLocaleString() }}</span>
                </div>
                <div v-if="log.cache_creation_1h_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="user-usage-token-tooltip__label flex items-center gap-1.5">
                    {{ t('admin.usage.cacheCreation1hTokens') }}
                    <span class="user-usage-token-tooltip__badge user-usage-token-tooltip__badge--brand inline-flex items-center text-[10px] font-medium leading-tight">
                      1h
                    </span>
                  </span>
                  <span class="user-usage-token-tooltip__value font-medium">{{ log.cache_creation_1h_tokens.toLocaleString() }}</span>
                </div>
              </template>
              <div v-else class="flex items-center justify-between gap-4">
                <span class="user-usage-token-tooltip__label">{{ t('admin.usage.cacheCreationTokens') }}</span>
                <span class="user-usage-token-tooltip__value font-medium">{{ log.cache_creation_tokens.toLocaleString() }}</span>
              </div>
            </div>
            <div v-if="log.cache_ttl_overridden" class="flex items-center justify-between gap-4">
              <span class="user-usage-token-tooltip__label flex items-center gap-1.5">
                {{ t('usage.cacheTtlOverriddenLabel') }}
                <span class="user-usage-token-tooltip__badge user-usage-token-tooltip__badge--danger inline-flex items-center text-[10px] font-medium leading-tight">
                  {{ getUserUsageCacheOverrideBadgeText(log) }}
                </span>
              </span>
              <span class="user-usage-token-tooltip__value user-usage-token-tooltip__value--danger font-medium">{{ t(getUserUsageCacheOverrideLabelKey(log)) }}</span>
            </div>
            <div v-if="log.cache_read_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-token-tooltip__label">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="user-usage-token-tooltip__value font-medium">{{ log.cache_read_tokens.toLocaleString() }}</span>
            </div>
          </div>
          <div class="user-usage-token-tooltip__total-row flex items-center justify-between gap-6 pt-1.5">
            <span class="user-usage-token-tooltip__label">{{ t('usage.totalTokens') }}</span>
            <span class="user-usage-token-tooltip__value user-usage-token-tooltip__value--info font-semibold">
              {{ getUserUsageTotalTokens(log).toLocaleString() }}
            </span>
          </div>
        </div>
        <div class="user-usage-token-tooltip__arrow absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-t-transparent"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UsageLog } from '@/types'
import type { UserUsageHoverTooltipPosition } from './useUserUsageHoverTooltip'
import {
  getUserUsageCacheOverrideBadgeText,
  getUserUsageCacheOverrideLabelKey,
  getUserUsageTotalTokens,
  hasUserUsageCacheCreationBreakdown
} from '../userUsageView'

const props = defineProps<{
  visible: boolean
  position: UserUsageHoverTooltipPosition
  log: UsageLog | null
}>()

const { t } = useI18n()

const tooltipStyle = computed(() => ({
  left: `${props.position.x}px`,
  top: `${props.position.y}px`
}))
</script>

<style scoped>
.user-usage-token-tooltip__panel {
  position: relative;
  border-radius: var(--theme-user-usage-token-tooltip-radius);
  padding: var(--theme-user-usage-token-tooltip-padding-y)
    var(--theme-user-usage-token-tooltip-padding-x);
  border: 1px solid color-mix(in srgb, var(--theme-surface-contrast) 16%, transparent);
  background: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
  color: var(--theme-surface-contrast-text);
}

.user-usage-token-tooltip__title,
.user-usage-token-tooltip__value {
  color: var(--theme-surface-contrast-text);
}

.user-usage-token-tooltip__label {
  color: color-mix(in srgb, var(--theme-surface-contrast-text) 62%, transparent);
}

.user-usage-token-tooltip__badge {
  border-radius: var(--theme-user-usage-token-tooltip-badge-radius);
  padding: var(--theme-user-usage-token-tooltip-badge-padding-y)
    var(--theme-user-usage-token-tooltip-badge-padding-x);
  border: 1px solid transparent;
}

.user-usage-token-tooltip__badge--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 18%, transparent);
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 74%, var(--theme-surface-contrast-text));
  border-color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 28%, transparent);
}

.user-usage-token-tooltip__badge--brand {
  background: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 18%, transparent);
  color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 74%, var(--theme-surface-contrast-text));
  border-color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 28%, transparent);
}

.user-usage-token-tooltip__badge--danger,
.user-usage-token-tooltip__value--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 18%, transparent);
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 74%, var(--theme-surface-contrast-text));
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 28%, transparent);
}

.user-usage-token-tooltip__value--danger {
  background: transparent;
  border-color: transparent;
}

.user-usage-token-tooltip__value--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 74%, var(--theme-surface-contrast-text));
}

.user-usage-token-tooltip__total-row {
  border-top: 1px solid color-mix(in srgb, var(--theme-surface-contrast-text) 16%, transparent);
}

.user-usage-token-tooltip__arrow {
  border-right-color: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
}
</style>
