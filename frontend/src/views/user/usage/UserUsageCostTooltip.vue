<template>
  <Teleport to="body">
    <div
      v-if="visible && log"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="tooltipStyle"
    >
      <div class="user-usage-cost-tooltip__panel whitespace-nowrap text-xs shadow-xl">
        <div class="space-y-1.5">
          <div class="user-usage-cost-tooltip__section mb-2 pb-1.5">
            <div class="user-usage-cost-tooltip__title mb-1 text-xs font-semibold">{{ t('usage.costDetails') }}</div>
            <div v-if="log.input_cost > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-cost-tooltip__label">{{ t('admin.usage.inputCost') }}</span>
              <span class="user-usage-cost-tooltip__value font-medium">${{ log.input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="log.output_cost > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-cost-tooltip__label">{{ t('admin.usage.outputCost') }}</span>
              <span class="user-usage-cost-tooltip__value font-medium">${{ log.output_cost.toFixed(6) }}</span>
            </div>
            <div v-if="log.input_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-cost-tooltip__label">{{ t('usage.inputTokenPrice') }}</span>
              <span class="user-usage-cost-tooltip__value user-usage-cost-tooltip__value--info font-medium">
                {{ formatTokenPricePerMillion(log.input_cost, log.input_tokens) }}
                {{ t('usage.perMillionTokens') }}
              </span>
            </div>
            <div v-if="log.output_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-cost-tooltip__label">{{ t('usage.outputTokenPrice') }}</span>
              <span class="user-usage-cost-tooltip__value user-usage-cost-tooltip__value--brand font-medium">
                {{ formatTokenPricePerMillion(log.output_cost, log.output_tokens) }}
                {{ t('usage.perMillionTokens') }}
              </span>
            </div>
            <div v-if="log.cache_creation_cost > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-cost-tooltip__label">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="user-usage-cost-tooltip__value font-medium">${{ log.cache_creation_cost.toFixed(6) }}</span>
            </div>
            <div v-if="log.cache_read_cost > 0" class="flex items-center justify-between gap-4">
              <span class="user-usage-cost-tooltip__label">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="user-usage-cost-tooltip__value font-medium">${{ log.cache_read_cost.toFixed(6) }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="user-usage-cost-tooltip__label">{{ t('usage.serviceTier') }}</span>
            <span class="user-usage-cost-tooltip__value user-usage-cost-tooltip__value--info font-semibold">
              {{ getUsageServiceTierLabel(log.service_tier, t) }}
            </span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="user-usage-cost-tooltip__label">{{ t('usage.rate') }}</span>
            <span class="user-usage-cost-tooltip__value user-usage-cost-tooltip__value--info font-semibold">{{ log.rate_multiplier.toFixed(2) }}x</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="user-usage-cost-tooltip__label">{{ t('usage.original') }}</span>
            <span class="user-usage-cost-tooltip__value font-medium">${{ log.total_cost.toFixed(6) }}</span>
          </div>
          <div class="user-usage-cost-tooltip__total-row flex items-center justify-between gap-6 pt-1.5">
            <span class="user-usage-cost-tooltip__label">{{ t('usage.billed') }}</span>
            <span class="user-usage-cost-tooltip__value user-usage-cost-tooltip__value--success font-semibold">${{ log.actual_cost.toFixed(6) }}</span>
          </div>
        </div>
        <div class="user-usage-cost-tooltip__arrow absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-t-transparent"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UsageLog } from '@/types'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import type { UserUsageHoverTooltipPosition } from './useUserUsageHoverTooltip'

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
.user-usage-cost-tooltip__panel {
  position: relative;
  border-radius: var(--theme-user-usage-token-tooltip-radius);
  padding:
    var(--theme-user-usage-token-tooltip-padding-y)
    var(--theme-user-usage-token-tooltip-padding-x);
  border: 1px solid color-mix(in srgb, var(--theme-surface-contrast) 16%, transparent);
  background: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
  color: var(--theme-surface-contrast-text);
}

.user-usage-cost-tooltip__section,
.user-usage-cost-tooltip__total-row {
  border-top: 1px solid color-mix(in srgb, var(--theme-surface-contrast-text) 16%, transparent);
}

.user-usage-cost-tooltip__section {
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-surface-contrast-text) 16%, transparent);
}

.user-usage-cost-tooltip__title,
.user-usage-cost-tooltip__value {
  color: var(--theme-surface-contrast-text);
}

.user-usage-cost-tooltip__label {
  color: color-mix(in srgb, var(--theme-surface-contrast-text) 62%, transparent);
}

.user-usage-cost-tooltip__value--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 74%, var(--theme-surface-contrast-text));
}

.user-usage-cost-tooltip__value--brand {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 74%, var(--theme-surface-contrast-text));
}

.user-usage-cost-tooltip__value--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 74%, var(--theme-surface-contrast-text));
}

.user-usage-cost-tooltip__arrow {
  border-right-color: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
}
</style>
