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
          <div>
            <div class="mb-1 text-xs font-semibold text-gray-300">{{ t('usage.tokenDetails') }}</div>
            <div v-if="log.input_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
              <span class="font-medium text-white">{{ log.input_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="log.output_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
              <span class="font-medium text-white">{{ log.output_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="log.cache_creation_tokens > 0">
              <template v-if="hasUserUsageCacheCreationBreakdown(log)">
                <div v-if="log.cache_creation_5m_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="flex items-center gap-1.5 text-gray-400">
                    {{ t('admin.usage.cacheCreation5mTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-amber-500/20 text-amber-400 ring-1 ring-inset ring-amber-500/30">
                      5m
                    </span>
                  </span>
                  <span class="font-medium text-white">{{ log.cache_creation_5m_tokens.toLocaleString() }}</span>
                </div>
                <div v-if="log.cache_creation_1h_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="flex items-center gap-1.5 text-gray-400">
                    {{ t('admin.usage.cacheCreation1hTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-orange-500/20 text-orange-400 ring-1 ring-inset ring-orange-500/30">
                      1h
                    </span>
                  </span>
                  <span class="font-medium text-white">{{ log.cache_creation_1h_tokens.toLocaleString() }}</span>
                </div>
              </template>
              <div v-else class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('admin.usage.cacheCreationTokens') }}</span>
                <span class="font-medium text-white">{{ log.cache_creation_tokens.toLocaleString() }}</span>
              </div>
            </div>
            <div v-if="log.cache_ttl_overridden" class="flex items-center justify-between gap-4">
              <span class="flex items-center gap-1.5 text-gray-400">
                {{ t('usage.cacheTtlOverriddenLabel') }}
                <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-500/20 text-rose-400 ring-1 ring-inset ring-rose-500/30">
                  {{ getUserUsageCacheOverrideBadgeText(log) }}
                </span>
              </span>
              <span class="font-medium text-rose-400">{{ t(getUserUsageCacheOverrideLabelKey(log)) }}</span>
            </div>
            <div v-if="log.cache_read_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="font-medium text-white">{{ log.cache_read_tokens.toLocaleString() }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-6 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-blue-400">
              {{ getUserUsageTotalTokens(log).toLocaleString() }}
            </span>
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
