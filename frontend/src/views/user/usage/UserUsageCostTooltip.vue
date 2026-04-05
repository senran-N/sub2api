<template>
  <Teleport to="body">
    <div
      v-if="visible && log"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="tooltipStyle"
    >
      <div
        class="whitespace-nowrap rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800"
      >
        <div class="space-y-1.5">
          <div class="mb-2 border-b border-gray-700 pb-1.5">
            <div class="mb-1 text-xs font-semibold text-gray-300">{{ t('usage.costDetails') }}</div>
            <div v-if="log.input_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.inputCost') }}</span>
              <span class="font-medium text-white">${{ log.input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="log.output_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.outputCost') }}</span>
              <span class="font-medium text-white">${{ log.output_cost.toFixed(6) }}</span>
            </div>
            <div v-if="log.input_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.inputTokenPrice') }}</span>
              <span class="font-medium text-sky-300">
                {{ formatTokenPricePerMillion(log.input_cost, log.input_tokens) }}
                {{ t('usage.perMillionTokens') }}
              </span>
            </div>
            <div v-if="log.output_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.outputTokenPrice') }}</span>
              <span class="font-medium text-violet-300">
                {{ formatTokenPricePerMillion(log.output_cost, log.output_tokens) }}
                {{ t('usage.perMillionTokens') }}
              </span>
            </div>
            <div v-if="log.cache_creation_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="font-medium text-white">${{ log.cache_creation_cost.toFixed(6) }}</span>
            </div>
            <div v-if="log.cache_read_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="font-medium text-white">${{ log.cache_read_cost.toFixed(6) }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.serviceTier') }}</span>
            <span class="font-semibold text-cyan-300">
              {{ getUsageServiceTierLabel(log.service_tier, t) }}
            </span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.rate') }}</span>
            <span class="font-semibold text-blue-400">{{ log.rate_multiplier.toFixed(2) }}x</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.original') }}</span>
            <span class="font-medium text-white">${{ log.total_cost.toFixed(6) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.billed') }}</span>
            <span class="font-semibold text-green-400">${{ log.actual_cost.toFixed(6) }}</span>
          </div>
        </div>
        <div
          class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-gray-900 border-t-transparent dark:border-r-gray-800"
        ></div>
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
